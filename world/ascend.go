package world

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Kaginari/isekai/value"
)

// Ascension is the coded answer to "an orc with experience evolves": a rank declares
// ascendsWhen (ASL), the world keeps the value ledger every Court report folds into, and
// after each fold the world evaluates the divine bonds. A base body whose snapshot crosses the
// condition ascends: the harness mints the ascended body in the ascended rank's own dir under
// the base's short name (orc-engine → high-orc-engine), carrying the base's doc whole — every
// trait, territory, verify line and Mind (ISEKAI.md Nature 5: "an ascended creature keeps every
// trait and territory it held before") — and retires the base's seat, so one body holds the
// soil and territories stay disjoint. The retired seat keeps its graph memory; creatures that
// reported to it now report to its ascendant (the roster resolves the succession).
//
// Each report folds into the value ledger (`<worldDir>/shared/value.jsonl`, world/value.go)
// replayed at Open, so a record earned over many sessions counts. Each ascension is appended
// to log.md with the reading that earned it; each failure is a hole, logged and surfaced to
// the dispatcher as `@?`.
//
// Nothing here reads a prompt. Evaluation is the ASL condition over the tally, deterministic;
// the mint and the doc write are the world's own.

// ascender is one resolved divine bond the world watches.
type ascender struct {
	To  Rank // the ascended rank (holds AscendsWhen)
	ASL ASL  // its parsed condition
}

// ascendState is World's ascension machinery, guarded by its own lock. The record it reads is
// the value ledger (world/value.go); it keeps only the watch list and what it minted.
type ascendState struct {
	mu     sync.Mutex
	watch  map[string]ascender // base rank → what ascends from it
	done   map[string]bool     // base bodies already ascended (replayed from the ledger)
	Events []AscendEvent       // the ascensions of this session
	news   []AscendEvent       // minted, not yet announced to a dispatcher
	holes  []string            // mint failures, drained by AscendHoles
}

// AscendEvent is one minted ascension.
type AscendEvent struct {
	Base  string    `json:"base"` // the ranking body (orc-auth)
	New   string    `json:"new"`  // the ascended body (high-orc-auth)
	Rank  string    `json:"rank"` // the ascended rank (high-orc)
	When  string    `json:"when"` // the ASL condition, as config wrote it
	At    time.Time `json:"at"`
	Doc   string    `json:"doc"` // the doc the harness wrote for the new body
	MetAt string    `json:"met"` // the snapshot line that crossed the condition
}

// newAscendState builds the watch list from the rank table, or nil when no rank ascends.
func newAscendState(rs Ranks) *ascendState {
	as := rs.Ascensions()
	if len(as) == 0 {
		return nil
	}
	st := &ascendState{watch: map[string]ascender{}, done: map[string]bool{}}
	for _, a := range as {
		cond, err := ParseASL(a.When)
		if err != nil { // Validate has already refused this; belt and braces
			continue
		}
		to, _ := rs.Get(a.To)
		st.watch[a.From] = ascender{To: to, ASL: cond}
	}
	return st
}

// ascendMaybe evaluates the divine bond over the body's rank. A body ascends once: the shelf
// and the roster (a live creature naming it as its base) both remember.
func (w *World) ascendMaybe(body string) {
	if w.ascend == nil {
		return
	}
	rank := w.RankOf(body).Name
	if w.successorOf(body) != "" {
		return
	}
	w.ascend.mu.Lock()
	watcher, ok := w.ascend.watch[rank]
	if !ok || w.ascend.done[body] {
		w.ascend.mu.Unlock()
		return
	}
	w.ascend.mu.Unlock()
	m := w.Metrics(body)
	if !watcher.ASL.Cond.Eval(m) {
		return
	}
	w.ascend.mu.Lock()
	if w.ascend.done[body] {
		w.ascend.mu.Unlock()
		return
	}
	// the crossing is real: mark before the write so a reload cannot double-mint
	w.ascend.done[body] = true
	w.ascend.mu.Unlock()

	ev, err := w.mintAscended(body, watcher, m)
	if err != nil {
		w.ascendHole(body, watcher, fmt.Sprintf("ascend %s: %v", body, err))
		return
	}
	w.ascend.mu.Lock()
	w.ascend.Events = append(w.ascend.Events, ev)
	w.ascend.news = append(w.ascend.news, ev)
	w.ascend.mu.Unlock()
	if w.val != nil {
		w.val.mu.Lock()
		w.appendValue(value.Row{Kind: value.KindAscend, Body: body, New: ev.New, At: ev.At})
		w.val.mu.Unlock()
	}
	if err := w.AppendLog(Entry{At: ev.At, Author: ev.New, Title: fmt.Sprintf("ascension — %s rose as %s", body, ev.New),
		Task:    "ascendsWhen `" + ev.When + "`",
		Files:   []string{ev.Doc, w.docOf(body)},
		Gate:    "n/a (minted by the harness; no body wrote it)",
		Result:  fmt.Sprintf("%s (%s) holds %s's territory, traits and verify; %s's seat retired", ev.New, ev.Rank, body, body),
		Learned: []string{"earned: " + ev.MetAt}}); err != nil {
		w.ascendHole(body, watcher, fmt.Sprintf("ascend %s: log.md: %v", body, err))
	}
	// the ascended body must be dispatchable on the next commission and the base must leave
	// the roster: re-derive it from the docs
	if err := w.Reload(); err != nil {
		w.ascendHole(body, watcher, fmt.Sprintf("reload after %s: %v", ev.New, err))
	}
}

// ascendHole records one failure: drained by AscendHoles for the dispatcher, and appended to
// log.md so the record names it even when no dispatcher reads it.
func (w *World) ascendHole(body string, a ascender, msg string) {
	w.ascend.mu.Lock()
	w.ascend.holes = append(w.ascend.holes, msg)
	w.ascend.mu.Unlock()
	_ = w.AppendLog(Entry{Author: body, Title: "ascension hole — " + body, Task: "ascendsWhen `" + a.ASL.String() + "`",
		Gate: "n/a (the harness's own act)", Result: "hole: " + msg})
}

// docOf is a creature's doc path, "" when it is not on the roster.
func (w *World) docOf(name string) string {
	if c := w.Creature(name); c != nil {
		return c.Doc
	}
	return ""
}

// successorOf names the live creature that ascended from `base`, "" when none.
func (w *World) successorOf(base string) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, c := range w.Creatures {
		if c.AscendantOf == base {
			return c.Name
		}
	}
	return ""
}

// mintAscended writes the ascended body's doc in the ascended rank's own dir — the base's doc
// carried whole under the new name, with the bond and the reading that earned it — then retires
// the base's seat. The base retires only once the new doc stands, so a failed mint never
// orphans the territory.
func (w *World) mintAscended(base string, a ascender, m Metrics) (AscendEvent, error) {
	baseC := w.Creature(base)
	if baseC == nil {
		return AscendEvent{}, fmt.Errorf("no creature %q on the roster", base)
	}
	if baseC.Doc == "" {
		return AscendEvent{}, fmt.Errorf("%s carries no doc to ascend with", base)
	}
	baseRank, _ := w.Ranks.Get(baseC.Rank)
	if a.To.Dir == "" || a.To.Prefix == "" || a.To.Dir == baseRank.Dir || a.To.Prefix == baseRank.Prefix {
		return AscendEvent{}, fmt.Errorf("rank %s needs a dir and prefix of its own (ranks.%s.dir / .prefix) — on %s's the ascended name would be the base's", a.To.Name, strings.ReplaceAll(a.To.Name, "_", "-"), baseRank.Name)
	}
	short := strings.TrimPrefix(base, baseRank.Prefix)
	newName := a.To.Prefix + short
	if w.Creature(newName) != nil {
		return AscendEvent{}, fmt.Errorf("%q is already on the roster", newName)
	}
	docRel := filepath.ToSlash(filepath.Join(w.Lex.WorldDir, a.To.Dir, short, "README.md"))
	abs := filepath.Join(w.Root, filepath.FromSlash(docRel))
	if _, err := os.Stat(abs); err == nil {
		return AscendEvent{}, fmt.Errorf("%s already stands on disk", docRel)
	}
	text, err := os.ReadFile(filepath.Join(w.Root, filepath.FromSlash(baseC.Doc)))
	if err != nil {
		return AscendEvent{}, err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return AscendEvent{}, err
	}
	now := time.Now().UTC()
	date := now.Format("2006-01-02")
	metAt := fmt.Sprintf("reports %.0f, passes %.0f, fails %.0f, holes %.0f, passRate %.0f%%, holeRate %.0f%%, score %.0f",
		m["reports"], m["passes"], m["fails"], m["holes"], m["passRate"], m["holeRate"], m["score"])
	when := a.ASL.String()
	if err := os.WriteFile(abs, []byte(ascendDoc(string(text), newName, a.To, base, when, metAt, date)), 0o644); err != nil {
		return AscendEvent{}, err
	}
	if err := w.retireDocWhy(baseC, "ascended "+date+": the body rose as "+newName+", its territory and traits with it"); err != nil {
		_ = os.Remove(abs) // the base still holds the soil: the half-minted ascendant goes
		return AscendEvent{}, fmt.Errorf("retiring %s: %v", base, err)
	}
	return AscendEvent{Base: base, New: newName, Rank: a.To.Name, When: when, At: now, Doc: docRel, MetAt: metAt}, nil
}

var (
	reH1     = regexp.MustCompile(`(?m)^#[ \t]+.*$`)
	reRankLn = regexp.MustCompile(`(?mi)^[ \t]*[-*][ \t]*\*\*rank:?\*\*:?.*$`)
	reAscOf  = regexp.MustCompile("(?mi)^[ \\t]*[-*][ \\t]*\\*\\*ascendant of:?\\*\\*:?[ \\t]*`?([a-z0-9][a-z0-9_-]*)`?")
)

// ascendDoc is the ascended body's card: the base's doc carried whole — identity renamed, the
// rank line naming the ascended rank, the bond to the base — and an Ascension section with the
// condition and the reading that crossed it.
func ascendDoc(base, name string, to Rank, from, when, metAt, date string) string {
	doc := base
	if loc := reH1.FindStringIndex(doc); loc != nil {
		doc = doc[:loc[0]] + "# " + name + doc[loc[1]:]
	} else {
		doc = "# " + name + "\n\n" + doc
	}
	bond := fmt.Sprintf("- **Rank:** %s\n- **Ascendant of:** `%s` — ascended %s; its territory, traits and verify carried whole", to.Name, from, date)
	if loc := reRankLn.FindStringIndex(doc); loc != nil {
		doc = doc[:loc[0]] + bond + doc[loc[1]:]
	} else if loc := reH1.FindStringIndex(doc); loc != nil {
		doc = doc[:loc[1]] + "\n\n" + bond + doc[loc[1]:]
	}
	if !strings.HasSuffix(doc, "\n") {
		doc += "\n"
	}
	doc += fmt.Sprintf("\n## Ascension\n\n- **Condition:** `%s`\n- **Crossed at:** %s\n- **Date:** %s\n", when, metAt, date)
	return doc
}

// ascendantOf reads the base a creature's doc says it ascended from ("" for none).
func ascendantOf(doc string) string {
	if m := reAscOf.FindStringSubmatch(doc); m != nil {
		return strings.ToLower(m[1])
	}
	return ""
}

// Ascensions is the session's ledger of minted ascensions.
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

// AscendNews drains the ascensions minted since the last call — the dispatcher announces them.
func (w *World) AscendNews() []AscendEvent {
	if w.ascend == nil {
		return nil
	}
	w.ascend.mu.Lock()
	defer w.ascend.mu.Unlock()
	out := w.ascend.news
	w.ascend.news = nil
	return out
}

// AscendHoles drains the ascension failures — the dispatcher surfaces them as `@?`.
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

// ascendLines renders what the last fold did for the dispatcher's report: each ascension
// announced, each hole as `@?`.
func (w *World) ascendLines() []string {
	var out []string
	for _, ev := range w.AscendNews() {
		out = append(out, fmt.Sprintf("[ascension] %s rose as %s (%s): %s — its territory, traits and verify carried whole; %s's seat retired; dispatch %s from now on (log.md)", ev.Base, ev.New, ev.Rank, ev.MetAt, ev.Base, ev.New))
	}
	for _, h := range w.AscendHoles() {
		out = append(out, "@? ascension hole — "+h)
	}
	return out
}
