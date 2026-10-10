package world

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Ascension is the coded answer to "an orc with experience evolves": a rank declares
// ascendsWhen (ASL), the world keeps a MetricsStore that every Court report folds into, and
// after each fold the world evaluates the divine bonds. A base body whose snapshot crosses the
// condition is ascended: the harness mints the ascended body — its doc lands on disk under the
// base's dir, the roster gains it — without a prompt, without the body asking. rankSet: replace
// still governs which ranks exist; ascendsWhen is what config adds to an ascended rank so the
// world knows the bond is lived, not just declared.
//
// Nothing here reads a prompt. Evaluation is the ASL condition over the tally, deterministic;
// the mint and the doc write are the world's own.

// ascender is one resolved divine bond the world watches.
type ascender struct {
	To   Rank            // the ascended rank (holds AscendsWhen)
	ASL  ASL             // its parsed condition
	Done map[string]bool // bodies already ascended this session — mint once
}

// ascendState is World's ascension machinery, guarded by the world lock.
type ascendState struct {
	mu      sync.Mutex
	metrics *MetricsStore
	watch   map[string]ascender // base rank → what ascends from it
	Events  []AscendEvent       // the ledger of ascensions this session
	holes   []string            // mint failures, surfaced at Close as an error
}

// AscendEvent is one minted ascension, journaled so the log can say what happened.
type AscendEvent struct {
	Base  string    // the ranking body (orc-auth)
	New   string    // the ascended body (high-orc-auth)
	Rank  string    // the ascended rank (high_orc)
	When  string    // the ASL condition, as config wrote it
	At    time.Time `json:"at"`
	Doc   string    // the doc the harness wrote for the new body
	MetAt string    // the snapshot line that crossed the condition
}

// ascendState builds the watch list from the rank table, or nil when no rank ascends.
func newAscendState(rs Ranks) *ascendState {
	as := rs.Ascensions()
	if len(as) == 0 {
		return nil
	}
	st := &ascendState{metrics: NewMetricsStore(), watch: map[string]ascender{}}
	for _, a := range as {
		cond, err := ParseASL(a.When)
		if err != nil { // Validate has already refused this; belt and braces
			continue
		}
		to, _ := rs.Get(a.To)
		st.watch[a.From] = ascender{To: to, ASL: cond, Done: map[string]bool{}}
	}
	return st
}

// FoldMetrics counts one Court report and runs the ascension pass. The parent body is who gets
// the tally (the ranking body is the dispatcher's *target* — but the target filed the report,
// so its own record grows; ascension reads the target's rank, never the dispatcher's).
func (w *World) FoldMetrics(body, status string, holes int) {
	if w.ascend == nil {
		return
	}
	w.ascend.mu.Lock()
	w.ascend.metrics.Fold(body, status, holes)
	w.ascend.mu.Unlock()
	w.ascendMaybe(body)
}

// ascendMaybe evaluates every divine bond against the body's rank. It mints at most one
// ascension per base body per session; the event is journaled and the roster reloaded.
func (w *World) ascendMaybe(body string) {
	if w.ascend == nil {
		return
	}
	w.ascend.mu.Lock()
	rank := w.RankOf(body).Name
	watcher, ok := w.ascend.watch[rank]
	if !ok || watcher.Done[body] {
		w.ascend.mu.Unlock()
		return
	}
	m := w.ascend.metrics.Metrics(body)
	if !watcher.ASL.Cond.Eval(m) {
		w.ascend.mu.Unlock()
		return
	}
	// the crossing is real: mark before the write so a reload cannot double-mint
	watcher.Done[body] = true
	w.ascend.watch[rank] = watcher
	w.ascend.mu.Unlock()

	newName, ev, err := w.mintAscended(body, watcher, m)
	if err != nil {
		w.ascend.mu.Lock()
		w.ascend.holes = append(w.ascend.holes, fmt.Sprintf("ascend %s: %v", body, err))
		w.ascend.mu.Unlock()
		return
	}
	w.ascend.mu.Lock()
	ev.At = time.Now()
	w.ascend.Events = append(w.ascend.Events, ev)
	w.ascend.mu.Unlock()
	// the ascended body must be dispatchable on the next commission: re-derive the roster so
	// creaturesOf reads the minted doc through the ontology, body consistent with the others.
	if err := w.Reload(); err != nil {
		w.ascend.mu.Lock()
		w.ascend.holes = append(w.ascend.holes, fmt.Sprintf("reload %s: %v", newName, err))
		w.ascend.mu.Unlock()
	}
}

// mintAscended names the ascended body, writes its doc beside the base's, and reloads the
// roster so the new creature is dispatchable. The doc is the harness's own — a record of the
// ascension, never the base's doc edited.
func (w *World) mintAscended(base string, a ascender, m Metrics) (string, AscendEvent, error) {
	baseC := w.Creature(base)
	if baseC == nil {
		return "", AscendEvent{}, fmt.Errorf("no creature %q on the roster", base)
	}
	// the ascended id: the ascended rank's prefix + the base's short name, in its dir
	short := base
	for _, p := range []string{baseC.Rank + "-", a.To.Base + "-"} {
		short = strings.TrimPrefix(short, p)
	}
	prefix := a.To.Prefix
	if prefix == "" && a.To.Base != "" {
		if br, ok := w.Ranks.Get(a.To.Base); ok {
			prefix = br.Prefix
		}
	}
	newName := prefix + short
	if w.Creature(newName) != nil {
		return "", AscendEvent{}, fmt.Errorf("%q is already on the roster", newName)
	}
	dir := a.To.Dir
	if dir == "" {
		dir = baseC.Dir
	}
	// the folder is the base's short name; in the ascended rank's own dir there is no
	// collision. When the ascended shares the base's dir, the README would land on the base's
	// own doc — the stat below refuses rather than overwrite.
	docRel := filepath.ToSlash(filepath.Join(w.Lex.WorldDir, dir, short, "README.md"))
	abs := filepath.Join(w.Root, filepath.FromSlash(docRel))
	if _, err := os.Stat(abs); err == nil {
		return "", AscendEvent{}, fmt.Errorf("%s already stands on disk", docRel)
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", AscendEvent{}, err
	}
	metAt := fmt.Sprintf("reports %d passes %d fails %d holes %d passRate %.0f%%",
		m.Reports, m.Passes, m.Fails, m.Holes, m.rate(m.Passes))
	doc := ascendDoc(newName, a.To, base, watcherWhen(a), metAt, rankParentOf(w.Ranks, a.To.Name, base))
	if err := os.WriteFile(abs, []byte(doc), 0o644); err != nil {
		return "", AscendEvent{}, err
	}
	ev := AscendEvent{Base: base, New: newName, Rank: a.To.Name, When: watcherWhen(a), Doc: docRel, MetAt: metAt}
	return newName, ev, nil
}

// ascendDoc is the body card the harness mints for an ascension — the world's words, carrying
// the rank, the territory it inherits, the bond to its base, and why it rose.
func ascendDoc(name string, to Rank, base, when, metAt, parent string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", name)
	fmt.Fprintf(&b, "- **Rank:** %s (ascended %s)\n", to.Name, to.Base)
	fmt.Fprintf(&b, "- **Territory:** inherits the base's — the ascended keeps everything %s held\n", to.Base)
	fmt.Fprintf(&b, "- **Ascendant of:** `%s`\n", base)
	fmt.Fprintf(&b, "- **Purpose:** %s\n", or(to.Job, "the distilled shape of its base's experience"))
	fmt.Fprintf(&b, "\n## Ascension\n\n")
	fmt.Fprintf(&b, "- **Condition:** `%s`\n- **Crossed at:** %s\n", when, metAt)
	if parent != "" {
		fmt.Fprintf(&b, "\n## Bond\n\n- **Above:** `%s`\n", parent)
	}
	return b.String()
}

// rankParentOf names the ascended body's one-hop-up — the ascended rank's own reportsTo,
// falling back to the base body it ascended from when config left the rank's parent blank.
func rankParentOf(rs Ranks, ascendedRank, base string) string {
	if r, ok := rs.Get(ascendedRank); ok && r.ReportsTo != "" {
		return r.ReportsTo
	}
	return base
}

func watcherWhen(a ascender) string { return a.ASL.String() }

// Ascensions is the session's ledger of minted ascensions (the log sink reads it).
func (w *World) Ascensions() []AscendEvent {
	if w.ascend == nil {
		return nil
	}
	w.ascend.mu.Lock()
	defer w.ascend.mu.Unlock()
	out := append([]AscendEvent(nil), w.ascend.Events...)
	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out
}

// AscendHoles drains the ascension failures so a whole-world error surface can report them.
func (w *World) AscendHoles() []string {
	if w.ascend == nil {
		return nil
	}
	w.ascend.mu.Lock()
	defer w.ascend.mu.Unlock()
	out := append([]string(nil), w.ascend.holes...)
	w.ascend.holes = nil
	return out
}

// Metrics reads one body's ascension snapshot (for the world's self-measure and tests).
func (w *World) Metrics(body string) Metrics {
	if w.ascend == nil {
		return Metrics{}
	}
	w.ascend.mu.Lock()
	defer w.ascend.mu.Unlock()
	return w.ascend.metrics.Metrics(body)
}
