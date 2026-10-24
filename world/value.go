package world

import (
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/Kaginari/isekai/loop"
	"github.com/Kaginari/isekai/rank"
	"github.com/Kaginari/isekai/value"
)

// The value law, coded (races.md §Value). Every Court's report and every throne turn folds one
// row into the value ledger (`<worldDir>/shared/value.jsonl`), always — a world with no
// ascension still measures its bodies. A gate verdict credits the body that held the gate; a
// pass a later fail on the same paths undoes is an overturn. The ledger is replayed at Open; the
// ascension condition, the council's reading, the board's leaderboard and the meters all read
// the one fold (value.Fold) under the one definition of success (law.value).

// ValueOptions is law.value as the world reads it. The zero value keeps the ledger with the
// shipped definition over the whole record.
type ValueOptions struct {
	Off         bool          // law.value.enabled: false — no row is written, nothing is measured
	Def         value.Def     // pass / fail / neutral words; zero = value.DefaultDef
	Window      time.Duration // the fold's window; 0 = the whole ledger
	WindowText  string        // the window as config wrote it (28d)
	ReopenTurns int           // a throne pass rewritten within this many turns reopens
	MinRecord   int           // reports below which a score is low-confidence; 0 = value.DefaultMinRecord
	EscalateAt  int           // law.stress.escalateAt: the strain that reads 100%
	Meters      map[string]value.MeterDef
}

// valueState is the ledger in memory, guarded by its own lock.
type valueState struct {
	mu       sync.Mutex
	rows     []value.Row
	passRefs map[string]value.Row // path → the standing pass verdict on it
	gates    map[string]gateNote  // body → its last end-gate verdict, taken by the court's fold
	scores   map[string]*value.Expr
}

// gateNote is one end-gate verdict held for the fold.
type gateNote struct {
	Word      string
	Paths     []string
	VerifyRun int
	VerifyRed int
}

func (w *World) valuePath() string {
	return filepath.Join(w.Dir(), filepath.FromSlash(value.Shelf))
}

// openValue seeds the ledger once from the legacy ascension tally, replays it, rebuilds the
// standing pass verdicts and marks the bodies already ascended. An unreadable ledger opens the
// world with an empty record, never blocks it.
func (w *World) openValue() {
	w.val = &valueState{passRefs: map[string]value.Row{}, gates: map[string]gateNote{}, scores: map[string]*value.Expr{}}
	if w.valOpt.Off {
		return
	}
	// a world that never kept a ledger is seeded once; one whose ledger is still on the old
	// shelf (its move failed) already carries its seed
	if len(w.shelfReadPaths(value.Shelf)) == 0 {
		if n, err := value.Seed(w.Dir()); err != nil {
			w.Notes = append(w.Notes, "value ledger: "+err.Error())
		} else if n > 0 {
			w.Notes = append(w.Notes, "value ledger seeded from "+value.LegacyShelf)
		}
	}
	var homes [][]value.Row
	for _, p := range w.shelfReadPaths(value.Shelf) {
		rs, err := value.Read(p)
		if err != nil {
			w.Notes = append(w.Notes, "value ledger: "+err.Error())
		}
		homes = append(homes, rs)
	}
	rows := mergeHomes(homes, func(r value.Row) time.Time { return r.At }, func(r value.Row) bool { return r.Kind == value.KindSeed })
	w.val.rows = rows
	for _, r := range rows {
		switch r.Kind {
		case value.KindVerdict:
			w.val.track(r)
		case value.KindAscend:
			if w.ascend != nil {
				w.ascend.done[r.Body] = true
			}
		}
	}
}

// track keeps the standing pass verdicts: a pass stands on its paths until a fail overturns it
// (callers hold mu). It answers the passes this verdict overturns.
func (v *valueState) track(r value.Row) []value.Row {
	var undone []value.Row
	switch r.Status {
	case "pass":
		for _, p := range r.Paths {
			v.passRefs[p] = r
		}
	case "fail":
		seen := map[string]bool{}
		for _, p := range r.Paths {
			ref, ok := v.passRefs[p]
			if !ok {
				continue
			}
			delete(v.passRefs, p)
			id := ref.At.Format(time.RFC3339Nano) + "\x00" + ref.Body
			if !seen[id] {
				seen[id] = true
				undone = append(undone, ref)
			}
		}
		// a pass overturned once is gone from every path it stood on
		for p, ref := range v.passRefs {
			for _, u := range undone {
				if ref.Body == u.Body && ref.At.Equal(u.At) {
					delete(v.passRefs, p)
				}
			}
		}
	}
	return undone
}

// appendValue writes one row to the ledger and the memory (callers hold val.mu).
func (w *World) appendValue(r value.Row) {
	if w.valOpt.Off || w.val == nil {
		return
	}
	if r.At.IsZero() {
		r.At = time.Now().UTC()
	}
	if err := value.Append(w.valuePath(), r); err != nil {
		w.mu.Lock()
		w.Notes = append(w.Notes, "value ledger: "+err.Error())
		w.mu.Unlock()
	}
	w.val.rows = append(w.val.rows, r)
}

// RunFold is one run as the value ledger reads it.
type RunFold struct {
	Body      string
	By        string // the dispatcher; "" for a throne turn
	Loop      string // the loop's status
	Word      string // the body's own @S word
	Gate      string // pass | fail | off | "" (nothing gated)
	GateTries int
	VerifyRun int
	VerifyRed int
	Holes     value.Holes
	Findings  int
	Strain    int
	Warrant   bool
	Guts      loop.GutsSum
	Tokens    int
	Paths     []string
	Turn      bool
}

// FoldRun appends one run to the ledger, then runs the ascension pass for its body. The record
// belongs to the body that ran; ascension reads its rank, never the dispatcher's.
func (w *World) FoldRun(f RunFold) {
	if w.val == nil {
		return
	}
	row := value.Row{Kind: value.KindRun, Body: f.Body, Rank: w.RankOf(f.Body).Name, By: f.By, Loop: value.Word(f.Loop), Word: value.Word(f.Word),
		Status: value.RunWord(f.Loop, f.Word), Gate: f.Gate, GateTries: f.GateTries, VerifyRun: f.VerifyRun, VerifyRed: f.VerifyRed,
		Holes: f.Holes, Findings: f.Findings, Strain: f.Strain, Warrant: f.Warrant, Thrash: f.Guts.Thrash, Momentum: f.Guts.Momentum,
		Guts: f.Guts.Read, Tokens: f.Tokens, Paths: f.Paths, Turn: f.Turn}
	w.val.mu.Lock()
	w.appendValue(row)
	w.val.mu.Unlock()
	w.ascendMaybe(f.Body)
}

// FoldMetrics counts one report by its status word and harness-hole count — the bare fold a
// caller with no run in hand makes. A loop word (FAIL, CHECKPOINT, DENIED …) stands for the
// loop's status; any other word is the body's own on a run the loop closed DONE.
func (w *World) FoldMetrics(body, status string, holes int) {
	loopStatus := "DONE"
	switch value.Word(status) {
	case loop.Fail, loop.Checkpoint, loop.Escalate, loop.Denied, loop.Dry, loop.Suspended, "INTERRUPTED":
		loopStatus = value.Word(status)
	}
	w.FoldRun(RunFold{Body: body, Loop: loopStatus, Word: status, Holes: value.Holes{Harness: holes}})
}

// FoldVerdict credits a gate verdict to the body that held the gate; a fail overturns every
// standing pass on the same paths, each an overturn row against the holder that issued it.
func (w *World) FoldVerdict(holder, of, word string, paths []string) {
	if w.val == nil || holder == "" || (word != "pass" && word != "fail") {
		return
	}
	row := value.Row{Kind: value.KindVerdict, Body: holder, Rank: w.RankOf(holder).Name, Of: of, Status: word, Paths: append([]string(nil), paths...)}
	w.val.mu.Lock()
	defer w.val.mu.Unlock()
	if w.valOpt.Off {
		return
	}
	row.At = time.Now().UTC()
	w.appendValue(row)
	for _, ref := range w.val.track(row) {
		w.appendValue(value.Row{Kind: value.KindOverturn, Body: ref.Body, Rank: ref.Rank, Of: ref.Of, Ref: ref.At.Format(time.RFC3339Nano), Paths: ref.Paths})
	}
}

// noteGate holds a body's end-gate verdict for the fold that takes its report.
func (w *World) noteGate(as string, v Verdict, paths []string) {
	if w.val == nil {
		return
	}
	word := v.Word
	if word != "pass" && word != "fail" && word != "off" {
		word = "" // n/a: no gate holder — nothing was judged
	}
	w.val.mu.Lock()
	w.val.gates[as] = gateNote{Word: word, Paths: append([]string(nil), paths...), VerifyRun: v.VerifyRun, VerifyRed: v.VerifyRed}
	w.val.mu.Unlock()
}

// takeGate drains a body's held verdict.
func (w *World) takeGate(as string) (gateNote, bool) {
	if w.val == nil {
		return gateNote{}, false
	}
	w.val.mu.Lock()
	defer w.val.mu.Unlock()
	g, ok := w.val.gates[as]
	delete(w.val.gates, as)
	return g, ok
}

// gateHolderFor is the body a verdict on `author`'s writes credits: the dispatcher when its rank
// holds the gate, else the author itself when it holds its own (a kijin), else no one.
func (w *World) gateHolderFor(dispatcher, author string) string {
	if dispatcher != "" && w.RankOf(dispatcher).HoldsGate {
		return dispatcher
	}
	if w.RankOf(author).HoldsGate {
		return author
	}
	return ""
}

// foldResult folds a closed run: its words, its holes split by kind, its gate, its gut, its spend.
func (w *World) foldResult(body, by string, r *loop.Result, strain int, warrant, turn bool) {
	infra, harness, self := r.HoleCount()
	f := RunFold{Body: body, By: by, Loop: r.Status, Word: r.Report.Status, GateTries: r.GateTries,
		Holes: value.Holes{Infra: infra, Harness: harness, Self: self}, Strain: strain, Warrant: warrant,
		Guts: r.Guts, Tokens: r.Usage.Input + r.Usage.Output + r.Usage.CacheRead + r.Usage.CacheWrite, Paths: r.Wrote, Turn: turn}
	if r.Interrupted {
		f.Loop = "INTERRUPTED"
	}
	if r.IsWire {
		f.Findings = len(r.Report.Findings)
	}
	if g, ok := w.takeGate(body); ok {
		f.Gate, f.VerifyRun, f.VerifyRed = g.Word, g.VerifyRun, g.VerifyRed
		if len(g.Paths) > 0 {
			f.Paths = g.Paths
		}
		if holder := w.gateHolderFor(by, body); holder != "" && !turn {
			w.FoldVerdict(holder, body, g.Word, g.Paths)
		}
	}
	w.FoldRun(f)
}

// turnEnded is the throne's Ended hook: every human turn is a row of the root's record.
func (w *World) turnEnded(as string) func(s *loop.Session, r *loop.Result) {
	return func(s *loop.Session, r *loop.Result) {
		w.mu.Lock()
		strain := w.stress[as]
		w.mu.Unlock()
		w.foldResult(as, "", r, strain, false, true)
	}
}

// Ledger is a copy of the value ledger's rows (for the board and tests).
func (w *World) Ledger() []value.Row {
	if w.val == nil {
		return nil
	}
	w.val.mu.Lock()
	defer w.val.mu.Unlock()
	return append([]value.Row(nil), w.val.rows...)
}

// ValueOpts is the fold's tuning as the world reads it.
func (w *World) ValueOpts() value.Opts {
	def := w.valOpt.Def
	if len(def.Pass)+len(def.Fail)+len(def.Neutral) == 0 {
		def = value.DefaultDef()
	}
	return value.Opts{Def: def, Window: w.valOpt.Window, Now: time.Now(), EscalateAt: w.valOpt.EscalateAt, ReopenTurns: w.valOpt.ReopenTurns}
}

// Spec is a rank's resolved value law: its own score and meter, else its family's.
func (w *World) Spec(rankName string) value.Spec {
	fam := w.Ranks.FamilyOf(rankName)
	sp := value.FamilyDefault(fam)
	if r, ok := w.Ranks.Get(rankName); ok && rankName != Rimuru {
		if r.Score != "" {
			sp.Score = r.Score
		}
		if r.Meter.Signal != "" {
			sp.Meter = r.Meter
		}
		if r.MinRecord > 0 {
			sp.MinRecord = r.MinRecord
		}
	}
	if sp.MinRecord <= 0 {
		sp.MinRecord = w.valOpt.MinRecord
	}
	if sp.MinRecord <= 0 {
		sp.MinRecord = value.DefaultMinRecord
	}
	return sp
}

// scoreExpr parses a rank's score once.
func (w *World) scoreExpr(rankName string) *value.Expr {
	w.val.mu.Lock()
	defer w.val.mu.Unlock()
	if e, ok := w.val.scores[rankName]; ok {
		return e
	}
	e, err := value.ParseNumber(w.Spec(rankName).Score, value.IsSignal)
	if err != nil {
		e = nil // config refused this at load; a broken score reads zero, never panics
	}
	w.val.scores[rankName] = e
	return e
}

// AllSignals folds the ledger over the window into every body's signals, each with its race's
// score (clamped 0..100).
func (w *World) AllSignals() map[string]value.Signals {
	if w.val == nil {
		return map[string]value.Signals{}
	}
	out := value.Fold(w.Ledger(), w.ValueOpts())
	now := time.Now()
	for body, s := range out {
		s["tenure"] = 0
		if born, ok := w.bornAt(body); ok {
			s["tenure"] = value.TenureDays(born, now)
		}
		s["score"] = value.Clamp(w.scoreExpr(w.RankOf(body).Name).Eval(s))
	}
	return out
}

// Metrics reads one body's signals (zero for a body with no record).
func (w *World) Metrics(body string) Metrics {
	if s, ok := w.AllSignals()[body]; ok {
		return s
	}
	return Metrics{}
}

// BodyValue is one body's value as every surface reads it.
type BodyValue struct {
	Body       string          `json:"body"`
	Rank       string          `json:"rank"`
	Family     string          `json:"family"`
	Signals    value.Signals   `json:"signals"`
	Score      float64         `json:"score"`
	Confidence string          `json:"confidence"` // ok | low (reports under the race's minRecord: never ranked)
	Meter      value.Meter     `json:"meter"`
	Strain     int             `json:"strain"`
	Ascension  *AscendProgress `json:"ascension,omitempty"`
	Evolution  *BodyEvolution  `json:"evolution,omitempty"`
}

// BodyEvolution is where a body stands against law.evolution: its desk, its stress, its cooldown
// and its last evolution.
type BodyEvolution struct {
	Thoughts     int        `json:"thoughts"`
	ThoughtsMax  int        `json:"thoughtsMax"`
	Stressed     bool       `json:"stressed"`
	Why          []string   `json:"why,omitempty"`
	RunsSince    int        `json:"runsSince"`
	CooldownLeft int        `json:"cooldownLeft"`
	Last         *value.Row `json:"last,omitempty"`
}

// EvolutionReport is law.evolution as armed, and the latest evolutions the ledger holds.
type EvolutionReport struct {
	Enabled      bool        `json:"enabled"`
	Mode         string      `json:"mode"`
	ThoughtsMax  int         `json:"thoughtsMax"`
	StressAt     int         `json:"stressAt"`
	FailStreak   int         `json:"failStreak"`
	CooldownRuns int         `json:"cooldownRuns"`
	Arms         []string    `json:"arms"`
	Recent       []value.Row `json:"recent"`
	Open         int         `json:"open"` // proposals waiting on the human
}

// EvolutionView reads law.evolution and the ledger's evolution rows, newest first (at most n).
func (w *World) EvolutionView(n int) EvolutionReport {
	o := w.Evolution()
	rep := EvolutionReport{Enabled: o.on(), Mode: o.Mode, ThoughtsMax: o.ThoughtsMax, StressAt: o.StressAt, FailStreak: o.FailStreak, CooldownRuns: o.CooldownRuns, Arms: o.Arms, Recent: []value.Row{}}
	rows := w.Ledger()
	for i := len(rows) - 1; i >= 0 && len(rep.Recent) < n; i-- {
		if rows[i].Kind == value.KindEvolve {
			rep.Recent = append(rep.Recent, rows[i])
		}
	}
	for _, p := range w.Proposals() {
		if p.Status == "open" {
			rep.Open++
		}
	}
	return rep
}

// AscendProgress is a body's distance to its ascension: each condition leaf's reading.
type AscendProgress struct {
	To     string       `json:"to"`
	When   string       `json:"when"`
	Met    bool         `json:"met"`
	Leaves []value.Leaf `json:"leaves"`
}

// RankValue is one rank's value law and its bodies' standing.
type RankValue struct {
	Rank      string         `json:"rank"`
	Family    string         `json:"family"`
	Score     string         `json:"score"`
	Meter     value.MeterDef `json:"meter"`
	MinRecord int            `json:"minRecord"`
	Bodies    int            `json:"bodies"`
	Confident int            `json:"confident"`
	Mean      float64        `json:"mean"` // the mean score of its confident bodies
	Reading   value.Meter    `json:"reading"`
}

// ValueReport is the whole value picture: what /api/value serves.
type ValueReport struct {
	Window    string                    `json:"window"`
	Def       value.Def                 `json:"def"`
	Ledger    string                    `json:"ledger"`
	Rows      int                       `json:"rows"`
	Off       bool                      `json:"off"`
	Bodies    []BodyValue               `json:"bodies"`
	Ranks     []RankValue               `json:"ranks"`
	Meters    map[string]value.MeterDef `json:"meters"`
	Evolution EvolutionReport           `json:"evolution"`
}

// valueOf renders one body's value from its signals.
func (w *World) valueOf(body string, s value.Signals) BodyValue {
	rk := w.RankOf(body).Name
	sp := w.Spec(rk)
	bv := BodyValue{Body: body, Rank: rk, Family: w.Ranks.FamilyOf(rk), Signals: s, Score: s["score"], Strain: int(s["strain"])}
	bv.Confidence = "ok"
	if s["reports"] < float64(sp.MinRecord) {
		bv.Confidence = "low"
	}
	bv.Meter = value.Render(sp.Meter, s[sp.Meter.Signal], s["reports"] > 0)
	if w.ascend != nil {
		w.ascend.mu.Lock()
		watcher, ok := w.ascend.watch[rk]
		w.ascend.mu.Unlock()
		if ok {
			bv.Ascension = &AscendProgress{To: watcher.To.Name, When: watcher.ASL.Cond.String(), Met: watcher.ASL.Cond.Eval(s), Leaves: watcher.ASL.Cond.Leaves(s)}
		}
	}
	if c := w.Creature(body); c != nil && c.Doc != "" && c.Name == body {
		st, cd := w.StressOf(body), w.CooldownOf(body)
		bv.Evolution = &BodyEvolution{Thoughts: w.DeskOf(body), ThoughtsMax: w.Evolution().ThoughtsMax, Stressed: st.Stressed, Why: st.Why, RunsSince: cd.RunsSince, CooldownLeft: cd.Left, Last: cd.Last}
	}
	return bv
}

// Value is one body's value (a body with no record reads zeros, low-confidence).
func (w *World) Value(body string) BodyValue {
	s, ok := w.AllSignals()[body]
	if !ok {
		s = value.Signals{}
	}
	return w.valueOf(body, s)
}

// Meter is one body's meter, rendered.
func (w *World) Meter(body string) value.Meter { return w.Value(body).Meter }

// ValueReport folds the ledger once into every body's and every rank's value.
func (w *World) ValueReport() ValueReport {
	rep := ValueReport{Window: w.valOpt.WindowText, Def: w.ValueOpts().Def, Ledger: filepath.ToSlash(filepath.Join(w.Lex.WorldDir, value.Shelf)), Off: w.valOpt.Off, Meters: map[string]value.MeterDef{}}
	if rep.Window == "" {
		rep.Window = "all"
	}
	for k, m := range w.valOpt.Meters {
		rep.Meters[k] = m
	}
	all := w.AllSignals()
	rep.Rows = len(w.Ledger())
	rep.Evolution = w.EvolutionView(10)
	// every body on the roster appears, a body with no record as a low-confidence zero
	names := map[string]bool{}
	for b := range all {
		names[b] = true
	}
	for _, c := range w.Roster() {
		names[c.Name] = true
	}
	var sorted []string
	for b := range names {
		sorted = append(sorted, b)
	}
	sort.Strings(sorted)
	byRank := map[string][]BodyValue{}
	for _, b := range sorted {
		s, ok := all[b]
		if !ok {
			s = value.Signals{}
		}
		bv := w.valueOf(b, s)
		rep.Bodies = append(rep.Bodies, bv)
		byRank[bv.Rank] = append(byRank[bv.Rank], bv)
	}
	ranks := []string{Rimuru}
	for _, r := range w.Ranks {
		ranks = append(ranks, r.Name)
	}
	for _, rk := range ranks {
		sp := w.Spec(rk)
		rv := RankValue{Rank: rk, Family: w.Ranks.FamilyOf(rk), Score: sp.Score, Meter: sp.Meter, MinRecord: sp.MinRecord}
		sum, meterSum := 0.0, 0.0
		for _, bv := range byRank[rk] {
			rv.Bodies++
			if bv.Confidence == "ok" {
				rv.Confident++
				sum += bv.Score
				meterSum += bv.Signals[sp.Meter.Signal]
			}
		}
		if rv.Confident > 0 {
			rv.Mean = sum / float64(rv.Confident)
			rv.Reading = value.Render(sp.Meter, meterSum/float64(rv.Confident), true)
		} else {
			rv.Reading = value.Render(sp.Meter, 0, false)
		}
		rep.Ranks = append(rep.Ranks, rv)
	}
	return rep
}

// Roster is the creatures on the roster, foreign bodies included.
func (w *World) Roster() []Creature {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append(append([]Creature(nil), w.Creatures...), w.Foreign...)
}

// CanonRank is the one spelling of a rank name (rank.Canon), re-exported for callers that hold
// the world only.
func CanonRank(name string) string { return rank.Canon(name) }

// RankMeters is every rank's resolved meter, keyed by rank (the board's /api/meters).
func (r ValueReport) RankMeters() map[string]value.MeterDef {
	out := map[string]value.MeterDef{}
	for _, rv := range r.Ranks {
		out[rv.Rank] = rv.Meter
	}
	return out
}
