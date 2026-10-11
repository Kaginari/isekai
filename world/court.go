package world

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kaginari/isekai/gate"
	"github.com/Kaginari/isekai/loop"
	"github.com/Kaginari/isekai/provider"
	"github.com/Kaginari/isekai/tool"
	"github.com/Kaginari/isekai/value"
	"github.com/Kaginari/isekai/wire"
)

// reportEnvelopeFloor is the envelope threshold: a report under it rides the parent's context
// whole (courier overhead would exceed relay savings); over it the court files it on the
// instruments shelf and sends the digest. Small relay answers stay inline on purpose.
const reportEnvelopeFloor = 600

// fileReport lands a Court's report text beside its journal on the instruments shelf
// (engine soil — territory law refuses creature writes under .isekai/instruments/, so the
// file lands from the harness, never through the body's hands) and answers the envelope fields.
func (w *World) fileReport(body, text string) (rel string, bytes int, hash string) {
	dir := filepath.Join(w.Root, w.Lex.WorldDir) + "/instruments/loop"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", 0, ""
	}
	name := fmt.Sprintf("%s-%s.report.md", body, time.Now().UTC().Format("20060102030405"))
	path := filepath.Join(dir, name)
	h := fnv.New64a()
	h.Write([]byte(text))
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return "", 0, ""
	}
	return filepath.ToSlash(filepath.Join(w.Lex.WorldDir, "instruments", "loop", name)), len(text), fmt.Sprintf("%x", h.Sum64())
}

// CourtOptions is tools.dispatch as the world applies it: who may dispatch, how deep, what a
// Court's report is capped at. Which ranks may dispatch, and whom, is the rank table's: a body
// dispatches ranks below its own (or its own when the rank is Sideways); the shelf is the
// rank's Tools.
type CourtOptions struct {
	Enabled    bool
	MaxDepth   int       // 0 = 1
	Cap        int       // 0 = the wire's default
	Unsaid     bool      // commission +unsaid by default (a report with no @U is flagged either way)
	Bodies     []string  // names a session may dispatch; nil = any creature on the roster
	Trace      io.Writer // a Court's step trace; nil is quiet
	Background bool      // background Courts allowed (tools.dispatch.background)
	// BackgroundTimeout bounds a background run; 0 falls back to the dispatch tool's default
	BackgroundTimeout time.Duration
	Wake              func(body, report string, failed bool) // a background Court's report lands here (depth 0: the session's queue)
	// Deliver drops a line into a RUNNING body's mid-turn inbox (the app court's Send path —
	// the same channel /send rides). Wired at depth ≥1, a background Court's report wakes its
	// own dispatcher there instead of the session's queue; nil there keeps background refused
	// at that depth (an orphaned report would wake the wrong body). Territory law already
	// fences parallel Courts' writes to their own ground, so async children never collide.
	Deliver func(body, line string)
	// Noted is the session's bookkeeping for a finished background Court: its @U seats land as
	// world holes and its court card closes. Every depth calls it; the session queue (Wake) is
	// stacked on top of it at depth 0, Deliver at depth ≥1.
	Noted      func(body, report string, failed bool)
	StrictFlow bool // one hop per layer: a body dispatches only the rank directly below it (no skip-level); neighbors allowed per the rank's Sideways
	// OpenLineage (tools.dispatch.lineage: false) lifts the creature-level law: by default a body
	// commissions only creatures whose Reports-to chain reaches it — an orc its own slimes, an
	// elf its own orcs; open, the rank law alone decides (any orc may order any slime).
	OpenLineage bool
	// SalvageTokens caps the journalise turn a Clock-court runs in its grace window — the body is
	// resumed from its own transcript with "spend this on @U/@F only" so its knowledge lands before
	// the axe. 0 = the loop default; negative = off, the body dies silent.
	SalvageTokens int
	// StressEscalateAt is the body's strain ledger reading that draws a council warrant into
	// the Court's report (0 = the council is off). StressMaxConsults bounds the warrants one
	// body may carry before the run hardens @S ESCALATE (0 reads as 3). StressRelieveAt is
	// the council's act threshold: strain past it calls relieve() — the zone itself changes
	// (split a wide one, retire a narrow one), and the ledger shelf keeps the birth/death.
	StressEscalateAt, StressMaxConsults, StressRelieveAt int
	// SessionBackground (tools.dispatch.sessionBackground): the session's own dispatches default
	// to background — rimuru stays free; the Court rides on and its report wakes the session.
	SessionBackground bool
}

// Route is what the integrator's model router sees for one engine: the office the ask names
// (findings → great-sage, verdict → raphael, draft → ciel), the rank and the creature.
type Route struct {
	Office   string
	Rank     string
	Creature string
	Task     string // dispatch | session
}

// Build is what an engine for a body is made of: the provider and gate it shares with its
// dispatcher, the base shelf, budgets, and every switch.
type Build struct {
	Provider  provider.Provider
	Models    func(Route) provider.Provider // nil = Provider for every route; the integrator routes models.offices/ranks/creatures here
	Gate      *gate.Gate
	Tools     *tool.Registry                                      // the base shelf; nil = tool.Builtins()
	Shelf     func(as string, depth int) (*tool.Registry, func()) // per-body shelf (its own shell); nil = Tools; the closer runs when the body's engine is closed
	Budget    loop.Budget
	Hooks     HookOptions
	Territory TerritoryOptions
	Court     CourtOptions
	Journal   string
	Trace     io.Writer
	Depth     int // 0 for the session; a dispatched body is Depth+1
	Extra     loop.Hooks
	Office    string     // the office this engine serves (set by the dispatcher from @ASK; "" = the session's)
	Guts      *loop.Guts // the per-step value sense the throne already carries — children inherit it here
	Doom      int        // thoughts-stress auto-relief threshold (law.budget doomLoop); 0 = off
	// ForgeDemandAt is law.forge.demandAtSteps: a run this long is asked once, at its close, for
	// its @FORGE draft (or why the path is one-off). 0 = off.
	ForgeDemandAt int
}

// providerFor resolves the provider for a route.
func (b Build) providerFor(r Route) provider.Provider {
	if b.Models != nil {
		if p := b.Models(r); p != nil {
			return p
		}
	}
	return b.Provider
}

// Engine builds a full engine for a body: tools cut to its rank and territory (plus dispatch
// when its rank and depth allow), the territory policy, the world's hooks, the lexicon.
func (w *World) Engine(as string, b Build) *loop.Engine {
	base := b.Tools
	var closer func()
	if b.Shelf != nil {
		base, closer = b.Shelf(as, b.Depth)
	}
	if base == nil {
		base = tool.Builtins()
	}
	shelf := tool.NewRegistry()
	for _, n := range base.Names() {
		if t, ok := base.Get(n); ok {
			shelf.Add(t)
		}
	}
	shelf.Add(w.LawTool())
	rank := w.RankOf(as)
	names := ExpandTools(w.Ranks.Tools(rank.Name), append(shelf.Names(), "dispatch"))
	office := b.Office
	if office == "" {
		office = rank.Office
	}
	task := "session"
	if b.Depth > 0 {
		task = "dispatch"
	}
	e := &loop.Engine{
		Provider: b.providerFor(Route{Office: office, Rank: rank.Name, Creature: as, Task: task}), Gate: b.Gate, Root: w.Root, As: as, Budget: b.Budget, Journal: b.Journal, Trace: b.Trace,
		Lexicon: w.Lex.Loop(), Policy: w.Policy(as, b.Territory), Cap: b.Court.Cap,
		Guts: b.Guts, Doom: b.Doom, ForgeDemandAt: b.ForgeDemandAt,
		// every write/edit/delete answers naming its zone's keeper — the territory truth
		// quoted at the moment of the landing, out of the same w.Owner the gate reads
		ZoneOwner: func(rel string) string {
			if c := w.Owner(rel); c != nil {
				return c.Name
			}
			return ""
		},
	}
	if e.Cap == 0 {
		e.Cap = wire.DefaultCap
	}
	max := b.Court.MaxDepth
	if max <= 0 {
		max = 1
	}
	// the depth ceiling never hides the rung a body still owes: the longest rank chain is the
	// configured floor — an elf always reaches its orc unless the human set a lower MaxDepth
	if cd := w.Ranks.ChainDepth(); max < cd {
		max = cd
	}
	// a rank that commands no one holds no dispatch, whatever its tools list (Ranks.Commands)
	if b.Court.Enabled && contains(names, "dispatch") && b.Depth < max && w.Ranks.Commands(rank.Name) {
		bg, wake := b.backgroundWake(as)
		shelf.Add(tool.DispatchTool(tool.DispatchOptions{Enabled: true, Depth: b.Depth, MaxDepth: max, Cap: b.Court.Cap, Unsaid: b.Court.Unsaid, Bodies: b.Court.Bodies, Background: bg, BackgroundTimeout: b.Court.BackgroundTimeout, Wake: wake, SessionBackground: b.Court.SessionBackground}, w.Dispatcher(e, b)))
	}
	e.Tools = shelf.Only(names...)
	e.Hooks = MergeHooks(w.Hooks(b.Hooks), b.Extra)
	if b.Depth == 0 {
		// the throne's turns are the root's record: every human ask folds into the value ledger
		e.Hooks = MergeHooks(e.Hooks, loop.Hooks{Ended: w.turnEnded(as)})
	}
	if closer != nil {
		w.mu.Lock()
		w.closers[e] = closer
		w.mu.Unlock()
	}
	return e
}

// backgroundWake resolves the dispatch tool's background switch and wake channel for one
// engine's depth. The session (depth 0) keeps the session queue as always. Deeper, an
// explicit `background: true` was dead canon — the wake default-queued into the SESSION,
// orphaning the report from its real dispatcher (the body that asked). With Deliver wired
// (the app court's Send inbox channel), a background Court's report wakes its own dispatcher
// mid-turn at any depth; without Deliver, background stays refused at that depth rather than
// landing on the wrong creature.
func (b Build) backgroundWake(as string) (bool, func(body, report string, failed bool)) {
	if b.Depth == 0 {
		return b.Court.Background, b.Court.Wake
	}
	if b.Court.Deliver == nil {
		return false, nil
	}
	parent, deliver, noted := as, b.Court.Deliver, b.Court.Noted
	return b.Court.Background, func(child, report string, failed bool) {
		if noted != nil {
			noted(child, report, failed)
		}
		deliver(parent, fmt.Sprintf("[court %s reported]\n%s", child, report))
	}
}

// Close ends what an engine's body held (its shell, its jobs). A dispatcher closes a Court's
// engine when the report is taken; the session closes its own at exit.
func (w *World) Close(e *loop.Engine) {
	w.mu.Lock()
	c := w.closers[e]
	delete(w.closers, e)
	delete(w.sessions, e)
	// courtWrites is arm 2's cross-session memory — Close must not clear it:
	// gate.go courtPairedDoc (OUT-2) consults lane B past the engine's Close.
	delete(w.courtWrote, e)
	w.mu.Unlock()
	if c != nil {
		c()
	}
}

// Dispatcher runs a commission as a Court Body under a parent engine: a fresh loop with its
// own context, the parent's provider and gate, budgets cut to what the parent has left, one
// wire report back. The Court's spend is added to the parent's.
// stratTargets lists the creatures this engine may lawfully dispatch right now — kept short,
// one name per creature that passes rank law + strict-flow (caller is either just above it in
// the tree, or sideways for ranks that ride side-to-side). The law's refusal prints it instead
// of a dead "Law 2" wall: refusal + known-good list in the same breath (07 of the 2026-10-02
// audit — the "throne flails 3 refused hops" shape).
func (w *World) stratTargets(parent *loop.Engine, opt CourtOptions) string {
	var ok []string
	for _, c := range append(append([]Creature{}, w.Creatures...), w.Foreign...) {
		c := c
		if c.Name == parent.As {
			continue
		}
		if w.mayDispatch(parent.As, &c, opt) == nil {
			ok = append(ok, c.Name)
		}
	}
	if len(ok) == 0 {
		return "(none — try @? to escalate)"
	}
	return strings.Join(ok, ", ")
}

// lineOf is a creature's Reports-to chain, nearest first, ending at rimuru (or where the chain
// breaks: a parent off the roster, a doc naming none).
func (w *World) lineOf(name string) []string {
	var line []string
	seen := map[string]bool{name: true}
	for cur := w.Creature(name); cur != nil && cur.Parent != "" && !seen[cur.Parent]; cur = w.Creature(cur.Parent) {
		seen[cur.Parent] = true
		line = append(line, cur.Parent)
		if cur.Parent == Rimuru {
			break
		}
	}
	return line
}

// mayDispatch is the dispatch law for one commission, nil when `from` may order `cr`:
//
//  1. rank law (Law 2): work goes down the ranks — or sideways within a Sideways rank; an
//     ascended rank commands what its base commanded;
//  2. strict flow (tools.dispatch.strictFlow): one rung at a time — the target's rank sits
//     directly under the dispatcher's (rimuru: its direct children);
//  3. lineage (tools.dispatch.lineage, default on): the target's own Reports-to chain reaches
//     the dispatcher — an orc orders its own slimes, an elf its own orcs; under strict flow the
//     dispatcher is the target's one hop up. Rimuru tops every chain.
func (w *World) mayDispatch(from string, cr *Creature, opt CourtOptions) error {
	me := w.RankOf(from)
	if me.Name == Rimuru {
		if opt.StrictFlow && !w.Ranks.DirectChild(Rimuru, cr.Rank) {
			return fmt.Errorf("%s (%s) may not dispatch %s (%s) under strict layer flow: work crosses one layer at a time — hand it to the rank directly below (%s), and the neighbor knows its own", from, me.Name, cr.Name, cr.Rank, w.firstChildLine(me.Name))
		}
		return nil
	}
	side := me.Sideways && cr.Rank == me.Name
	if !side && !w.Ranks.BelowAs(cr.Rank, me.Name) {
		return fmt.Errorf("%s (%s) may not dispatch %s (%s): work goes down the ranks, never up or sideways (Law 2) — escalate with @? instead", from, me.Name, cr.Name, cr.Rank)
	}
	if side {
		return nil
	}
	if opt.StrictFlow && !w.Ranks.DirectChildAs(me.Name, cr.Rank) {
		return fmt.Errorf("%s (%s) may not dispatch %s (%s) under strict layer flow: work crosses one layer at a time — hand it to the rank directly below (%s), and the neighbor knows its own", from, me.Name, cr.Name, cr.Rank, w.firstChildLine(me.Name))
	}
	if opt.OpenLineage {
		return nil
	}
	line := w.lineOf(cr.Name)
	at := -1
	for i, p := range line {
		if p == from {
			at = i
			break
		}
	}
	if len(line) == 0 {
		return fmt.Errorf("%s may not dispatch %s: its doc (%s) names no Reports to, so its line is unknown — only rimuru may commission it until its doc names its one hop up", from, cr.Name, or(cr.Doc, "none"))
	}
	if at < 0 {
		return fmt.Errorf("%s is not in your line: it reports to %s (%s → %s) — %s commissions it; carry the need up with @? so the line above routes it there (tools.dispatch.lineage)", cr.Name, line[0], cr.Name, strings.Join(line, " → "), line[0])
	}
	if opt.StrictFlow && at > 0 {
		return fmt.Errorf("%s reports to %s, not to you: strict flow crosses one rung at a time — commission %s and let it order %s", cr.Name, line[0], line[0], cr.Name)
	}
	return nil
}

// verifyCodedNow reads the creature's doc fresh off disk: a verify line written earlier in this
// very turn counts at the TDD check (the roster's copy is as old as its last reload).
func (w *World) verifyCodedNow(cr *Creature) bool {
	if cr.Doc != "" {
		if b, err := os.ReadFile(filepath.Join(w.Root, filepath.FromSlash(cr.Doc))); err == nil {
			fresh := Creature{Verify: VerifyOf(string(b))}
			return fresh.VerifyCoded()
		}
	}
	return cr.VerifyCoded()
}

func (w *World) Dispatcher(parent *loop.Engine, b Build) tool.Dispatcher {
	return func(ctx context.Context, env tool.Env, c tool.Commission) (tool.DispatchReport, error) {
		body := strings.ToLower(strings.TrimSpace(c.Body))
		cr := w.Creature(body)
		if cr == nil {
			return tool.DispatchReport{}, fmt.Errorf("no body named %q on the roster (%s/{%s}/<name>/) — a Court is minted for a rank, never generic. Dispatchable from here: %s", body, w.Lex.WorldDir, strings.Join(rankDirs(w.Ranks), ","), w.stratTargets(parent, b.Court))
		}
		me := w.RankOf(parent.As)
		// the rank law, strict flow and the lineage (mayDispatch); a refusal names who may
		// lawfully be ordered from here
		if err := w.mayDispatch(parent.As, cr, b.Court); err != nil {
			return tool.DispatchReport{}, fmt.Errorf("%v. Dispatchable from you: %s", err, w.stratTargets(parent, b.Court))
		}
		// Orc TDD as code (Law 3 + Veldora decree 2026-09-30): a gate holder dispatching an
		// authoring slime must first have coded that slime's verify lines on its README — the
		// verdict the gate will run exists BEFORE the work is ordered, never written in the same
		// breath. Fires when the dispatcher's rank HoldsGate and the target's rank Authors.
		{
			targetRank, _ := w.Ranks.Get(cr.Rank)
			meRank, _ := w.Ranks.Get(me.Name)
			if targetRank.Authors && meRank.HoldsGate && !w.verifyCodedNow(cr) {
				return tool.DispatchReport{}, fmt.Errorf("%s (%s) has no Verify line coded — the orc writes the slime's verdict BEFORE the order (TDD per Law 3, Veldora decree 2026-09-30). Add a non-trivial line in %s (front-matter verify:, or a `- `<cmd>`` bullet under ## Verify), then dispatch again", cr.Name, cr.Rank, cr.Doc)
			}
		}
		child := b
		child.Depth = b.Depth + 1
		child.Office = OfficeOf(c.Ask)
		if c.Office != "" {
			child.Office = c.Office
		}
		child.Trace = b.Court.Trace
		child.Budget = b.Budget
		if ps := w.SessionOf(parent); ps != nil {
			if steps := parent.Budget.Steps; steps >= 0 {
				if steps == 0 {
					steps = loop.DefaultSteps
				}
				left := steps - ps.Steps
				if left <= 0 {
					return tool.DispatchReport{}, fmt.Errorf("parent step budget spent (%d of %d) — nothing left for a Court", ps.Steps, steps)
				}
				child.Budget.Steps = left
			}
			if m := parent.Budget.Minutes; m >= 0 {
				if m == 0 {
					m = loop.DefaultMinutes
				}
				left := m - time.Since(ps.Started).Minutes()
				if left <= 0 {
					return tool.DispatchReport{}, fmt.Errorf("parent wall-clock budget spent (%g min) — nothing left for a Court", m)
				}
				child.Budget.Minutes = left
			}
		}
		e := w.Engine(cr.Name, child)
		defer w.Close(e) // the Court's context dies with the task
		e.Unsaid = c.Unsaid
		e.UnsaidExplicit = c.UnsaidExplicit
		e.Wire = true
		if c.Cap > 0 {
			e.Cap = c.Cap
		}
		c.Cap = e.Cap // the ceiling in force rides the commission
		// a body that never reports holds its own clock: a hung model call can be cancelled
		// mid-turn, not just between turns — the parent waits forever otherwise (the stuck
		// court Veldora saw: `thinking · 17m`, then `thinking · 22m` on a streamed Completion).
		// The uncapped-order (law.budget.minutes: -1) outlaws the wall-clock for the SESSION's
		// own turns — a dispatched body is not the session: it always carries a hard clock, the
		// parent's ceiling when one is armed, loop.DefaultMinutes when the session runs uncapped.
		courtMinutes := child.Budget.Minutes
		if courtMinutes <= 0 {
			courtMinutes = loop.DefaultMinutes
		}
		// two clocks: a grace deadline, then the axe. When the grace fires the worker journalizes
		// its own transcript (the resume pack) so a clock-killed court dies WITH its knowledge
		// on disk — a next run can `isekai resume <runID>` and inherit it, instead of starting cold.
		total := time.Duration(courtMinutes * float64(time.Minute))
		grace := total / 10
		if grace < 2*time.Second {
			grace = 2 * time.Second
		}
		if grace > time.Minute {
			grace = time.Minute
		}
		// the worker keeps the full ceiling unless the grace is spendable: with salvage off (negative)
		// there is nothing to do in the window, so the court runs to the axe exactly as it always has —
		// a parent whose own budget.minutes matches the court's then dies by its own clock first.
		salvageArmed := b.Court.SalvageTokens >= 0
		workerTotal := total
		if salvageArmed {
			workerTotal = total - grace
		}
		inner, cancelWorker := context.WithTimeout(ctx, workerTotal)
		ctx2, cancel := context.WithTimeout(ctx, total)
		if e.JournalDir() != "" {
			go func() {
				<-inner.Done()
				if ctx2.Err() != nil {
					return // the axe fell first, or the parent died — nothing to prepare
				}
				s := w.SessionOf(e)
				if s == nil {
					return
				}
				tp := strings.TrimSuffix(s.Journal.Path, ".jsonl") + ".transcript.json"
				if b := s.Transcript(); b != nil {
					_ = os.WriteFile(tp, b, 0o644)
				}
			}()
		}
		defer cancel()
		defer cancelWorker()
		r, err := e.Run(inner, c.String())
		// the clock fired inside the run when the worker context died by deadline. A hung child comes
		// back err == nil — the loop folds the cancelled Complete into a Fail status via end(), which
		// returns a nil Go error — so the trigger is the worker deadline with a non-Done result, never
		// err != nil. That is the shape of the stuck courts Veldora saw (`thinking · 17m`).
		clockFired := inner.Err() == context.DeadlineExceeded || ctx2.Err() == context.DeadlineExceeded
		diedByClock := clockFired && (r == nil || r.Status != loop.Done)
		memoLine := ""
		if diedByClock {
			// Name it: the lane, the provider, the budget — never a mule "the model never answered"
			// (018 of the 2026-10-02 audit: the user reads this line on the wire).
			lane := "unknown-model"
			if e.Provider != nil {
				lane = e.Provider.Name() // "openai/<model>"
			}
			provider, _, _ := strings.Cut(lane, "/")
			timeoutWhy := fmt.Sprintf("model %s on provider %q did not answer within the %g-minute court budget", lane, provider, courtMinutes)
			if err != nil && !errors.Is(err, context.DeadlineExceeded) {
				err = fmt.Errorf("%s: court ran out of its %g-minute budget — %s (escalate with @? or re-dispatch): %w", cr.Name, courtMinutes, timeoutWhy, err)
			} else if err == nil {
				err = fmt.Errorf("%s: court ran out of its %g-minute budget — %s (escalate with @? or re-dispatch)", cr.Name, courtMinutes, timeoutWhy)
			}
			// genome memo (world/court_vita.go): the stress ledger keeps the lane of the burn; the next-fit
			// ladder rides the wire as advice — never a silent mutation (Veldora law 2026-10-02).
			w.mu.Lock()
			w.stressDeaths[body] = stressEnd{Lane: lane, At: time.Now().UTC()}
			w.ledgerNote(ledgerRow{Kind: "death", Body: body, Lane: lane, At: w.stressDeaths[body].At})
			w.mu.Unlock()
			if up := laneUp(lane); up != "" {
				memoLine = fmt.Sprintf("\n[court genome memo] %s burned its court clock on lane %s — next-fit ladder says %s (or split the ask). Advisory only: nobody's DNA moves until the parent or human signs off.", cr.Name, lane, up)
			}
			// then spend the grace on a journalise turn: the body resumes from its own transcript,
			// hard-capped at salvageTokens, under the axe's remaining window — told to land @U/@F only,
			// so a clock-killed court dies with its knowledge on the wire and the shared notes, not in
			// a transcript nobody reads.
			if st := b.Court.SalvageTokens; st >= 0 {
				if st == 0 {
					st = loop.DefaultSalvageTokens
				}
				se := w.Engine(cr.Name, child)
				se.Unsaid, se.Wire = c.Unsaid, true
				se.Budget.MaxTokens, se.Budget.Steps = st, 1
				se.Provider = b.providerFor(Route{Office: OfficeOf(c.Ask), Rank: w.RankOf(cr.Name).Name, Creature: cr.Name, Task: "salvage"})
				if sr, serr := se.Resume(ctx2, e.LastRunID(), "You are out of time — the grace window is all there is. Do NOT continue the task. Answer on the wire now: `@S HOLE`, one `@F` per finding already certain, every `@U` you owe (@U law/colony/territory), `@?` for what you could not settle, `@E`. Plain notes next."); serr == nil && sr != nil && sr.IsWire {
					if rep, ok := wire.ParseReport(sr.Text); ok && len(rep.Unsaid) > 0 {
						holes := w.Land(cr.Name, rep)
						_ = holes
					}
					// the salvage turn's drafts are the dying court's too: filed, never lost
					_ = w.FileForges(cr.Name, e.LastRunID(), forgesOfResult(sr))
				}
				w.Close(se)
			}
		}
		if err != nil {
			return tool.DispatchReport{}, err
		}
		if ps := w.SessionOf(parent); ps != nil {
			ps.Spend = ps.Spend.Add(r.Usage)
		}
		// the court's drafts are filed here, from the report the harness holds — before @CAP cuts
		// it and before the courier envelope folds it to a digest. A relayed line hashes the same
		// one hop up and is not filed twice: the first filer, the drafting body, wins.
		r.Holes = append(r.Holes, w.FileForges(cr.Name, r.RunID, forgesOfResult(r))...)
		rep := tool.DispatchReport{Text: r.Emit(e.Cap), Status: r.Status, Unsaid: len(r.Report.Unsaid), Failed: r.Status == loop.Fail, Holes: r.Holes, Wrote: wroteOf(r), Journal: r.Journal, Forges: r.Report.Forges}
		rep.Text += memoLine
		// the courier law: substantive reports are filed on the instruments shelf, and what
		// rides the parent's context is the envelope — the path, the size, the checksum. The
		// parent that needs the flesh buys it with one read; relaying prose up five rungs is
		// the relay tax this closes.
		if strings.TrimSpace(rep.Text) != "" && len(rep.Text) > reportEnvelopeFloor {
			rep.ReportPath, rep.ReportBytes, rep.ReportHash = w.fileReport(body, rep.Text)
		}
		warrant := false
		if stamp, escalate := w.councilLedger(body, verdictOf(r), r, b.Court); stamp != "" {
			rep.Text += stamp
			warrant = true
			if escalate {
				rep.Status, rep.Failed = "ESCALATE", true
			}
		}
		w.courtGated(parent, body, rep.Wrote)
		// the value ledger grows before the return: the ranking body's report is its record, and
		// an ascension may mint off it. One definition of success (law.value): the run's word is
		// the loop's when the loop did not close DONE, else the body's own @S — its FAIL /
		// ESCALATE overrides a clean loop status. The gate verdict credits the body that held it.
		w.mu.Lock()
		strainNow := w.stress[body]
		w.mu.Unlock()
		w.foldResult(cr.Name, parent.As, r, strainNow, warrant, false) // the roster's name: a short or old name folds into one record
		lines := w.ascendLines()
		// law.evolution: a stressed body whose desk overflowed heals itself in the same wave —
		// split, mutate or rebirth, proven before it lands (world/evolve.go)
		evolved, evoFiles := w.evolveAfterRun(cr.Name)
		lines = append(lines, evolved...)
		// the harness wrote them on the body's behalf, proven before the swap: the dispatcher's
		// end gate leaves them alone, as it does a court's own gated writes
		w.courtGated(parent, body, evoFiles)
		// the relief arm (law.stress.relieveAt): the warrant talked, now the zone changes —
		// outside the council's lock; relieve rereads safely, and both the act and the ledger
		// shall witness it. A body that just evolved has had its zone changed already.
		if rel := b.Court.StressRelieveAt; rel > 0 && len(evolved) == 0 {
			w.mu.Lock()
			strainNow := w.stress[body]
			w.mu.Unlock()
			if strainNow >= rel && w.Creature(body) != nil {
				if act, ok := w.relieve(body); ok {
					rep.Text += "\n@? council relief — " + body + " crosses relieveAt: " + act
				}
			}
		}
		if len(lines) > 0 {
			note := strings.Join(lines, "\n")
			if rep.ReportPath != "" {
				rep.Text = note + "\n" + rep.Text // the envelope's digest shows the head: lead with it
			} else {
				rep.Text += "\n" + note
			}
		}
		return rep, nil
	}
}

// firstChildLine names the layer directly below a rank for the strict-flow refusal text.
func (w *World) firstChildLine(rank string) string {
	if rank == Rimuru {
		var kids []string
		for _, r := range w.Ranks {
			if r.ReportsTo == Rimuru {
				kids = append(kids, r.Name)
			}
		}
		if len(kids) > 0 {
			return "rimuru's children: " + strings.Join(kids, ", ")
		}
		return "rimuru has no children"
	}
	r, ok := w.Ranks.Get(rank)
	if !ok {
		return "the layer below"
	}
	var kids []string
	for _, o := range w.Ranks {
		if o.ReportsTo == r.Name {
			kids = append(kids, o.Name)
		}
	}
	if len(kids) > 0 {
		return "the one beneath: " + strings.Join(kids, ", ")
	}
	if r.Sideways {
		return "sideways within the layer: " + strings.Join(w.Ranks.NeighborRanks(r.Name), ", ")
	}
	return "no child and no neighbor — escalate with @?"
}

// verdictOf is the body's own word from its wire: the first token of its @S line. The loop's
// lifecycle status is not the verdict — a court that reports "@S FAIL" landed its run cleanly
// (the Law-9 lesson); the journal's status says DONE where the body says FAIL.
func verdictOf(r *loop.Result) string {
	if r.Report.Status != "" {
		return value.Word(r.Report.Status)
	}
	return ""
}

// councilLedger is the world's stress law, coded. A Court whose work did not land — the loop
// ended Fail or Checkpoint, or the body's own verdict is FAIL/ESCALATE — adds strain to its
// body's ledger: the run's governance sent-backs (failed acts standing at the end, gate
// re-asks, wire re-asks) plus one for the landing itself. Past law.stress.escalateAt a warrant
// rides the report back to the dispatcher — a body under strain needs its MANAGER, and the
// manager answers in kind: rewrite the spec (@INSTRUCT), narrow it (@SPLIT), take counsel
// sideways, or carry the hole up (@ESCALATE). The warrants a body may carry are bounded
// (law.stress.maxConsults); past the bound the warrant hardens: the report ends @S ESCALATE
// and rises to the human. A clean landing clears the ledger.
//
// Landed and struggled are law.value's one definition (value.Def): a pass word lands, a fail
// word struggles, a neutral word — the human's DENIED, an INTERRUPTED, a DRY run — never
// consults: the human's word and the clock's law already answered.
func (w *World) councilLedger(body string, verdict string, r *loop.Result, court CourtOptions) (stamp string, escalate bool) {
	word := value.RunWord(r.Status, verdict)
	if r.Interrupted {
		word = "INTERRUPTED"
	}
	outcome := w.ValueOpts().Def.Outcome(word)
	w.mu.Lock()
	defer w.mu.Unlock()
	if outcome == value.Pass { // a clean landing clears the ledger — management by exception
		delete(w.stress, body)
		delete(w.consults, body)
		w.ledgerNote(ledgerRow{Kind: "stressClear", Body: body})
		return "", false
	}
	if outcome != value.Fail {
		return "", false // neutral: another law already spoke
	}
	at := court.StressEscalateAt
	if at <= 0 {
		return "", false // the council is off
	}
	w.stress[body] += max(1, r.Strains+1)
	w.ledgerNote(ledgerRow{Kind: "stress", Body: body, Value: w.stress[body]})
	if w.stress[body] < at {
		return "", false
	}
	w.consults[body]++
	w.ledgerNote(ledgerRow{Kind: "consults", Body: body, Value: w.consults[body]})
	maxConsults := court.StressMaxConsults
	if maxConsults <= 0 {
		maxConsults = 3
	}
	head := fmt.Sprintf("\n@? council warrant — %s carries strain %d (escalate-at %d): its governance thought-line broke %d time%s this run and never landed.", body, w.stress[body], at, r.Strains+1, plural(r.Strains+1))
	if w.consults[body] > maxConsults {
		return head + fmt.Sprintf("\n@? council exhausted: %d consults carried and the strain holds. This Court ends @S ESCALATE — the warrant rises to the human.", maxConsults), true
	}
	return head + fmt.Sprintf(`
@? Answer it on your next step, as its manager — a strain the worker carries is a spec the manager rewrites:
@?     @INSTRUCT <text>    re-commission this body with the instruction folded into the ask (context rides the spec, never the worker)
@?     @SPLIT <asks>       narrow it: one commission per slice, same body
@?     @ESCALATE           carry @? up — your layer cannot settle it (or take counsel sideways first: layer-mates know the domain)
@? This warrant is consult %d of %d; a clean landing clears the ledger.`, w.consults[body], maxConsults), false
}

// courtWrite is the court-write ledger's value: which body wrote the path, and when its report
// was taken (the path is the key). The ledger is an attribution record, not a mere set — the
// throne's second pairing arm (world/gate.go, decree 2026-10-01) reads it to prove a doc on
// disk is this session's own court work, merely a turn apart from the code that lands under it.
type courtWrite struct {
	Writer string    // the Court's own body name
	At     time.Time // when its report was taken
}

// courtWritesBound caps arm 2's cross-session memory per creature.
const courtWritesBound = 8

// courtGated records what a Court of the parent engine wrote (and gated itself): path, writer,
// moment. The record rides the PARENT's key at court-report time, and the parent's key survives
// the Court's own Close (Close deletes the court's key only), so the attribution holds across
// the dispatcher's later turns and resumes within the session.
func (w *World) courtGated(parent *loop.Engine, body string, wrote []string) {
	if len(wrote) == 0 {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	m := w.courtWrote[parent]
	if m == nil {
		m = map[string]courtWrite{}
		w.courtWrote[parent] = m
	}
	now := time.Now()
	for _, p := range wrote {
		m[normRel(p)] = courtWrite{Writer: body, At: now}
	}
	// Lane B: double-write into the world-resident record keyed by the creature name,
	// so a throne-side consult through a fresh engine still pairs cross-session.
	bm := w.courtWrites[body]
	if bm == nil {
		bm = map[string]courtWrite{}
		w.courtWrites[body] = bm
	}
	for _, p := range wrote {
		bm[normRel(p)] = courtWrite{Writer: body, At: now}
		w.ledgerNote(ledgerRow{Kind: "courtWrite", Body: body, Path: normRel(p), Writer: body, At: now})
	}
	// Evict oldest-bound arm 2's cross-session memory per creature.
	for len(bm) > courtWritesBound {
		oldest := ""
		for k, cw := range bm {
			if oldest == "" || cw.At.Before(bm[oldest].At) {
				oldest = k
			}
		}
		delete(bm, oldest)
	}
}

// ownWrites is a session's writes less what its Courts wrote and gated on their own account.
func (w *World) ownWrites(s *loop.Session) []string {
	w.mu.Lock()
	m := w.courtWrote[s.Engine]
	w.mu.Unlock()
	if len(m) == 0 {
		return s.Wrote
	}
	var out []string
	for _, p := range s.Wrote {
		if _, court := m[normRel(p)]; !court {
			out = append(out, p)
		}
	}
	return out
}

// courtGatedTask reports whether doc was written by a Court of s's engine —
// lane A (this session's per-engine ledger) or, lane B, the world-resident
// cross-session ledger keyed by the creature name carried in w.sessions, so a
// throne re-consult through a fresh engine still pairs end-gate. Lane B is live
// at the gate: gate.go courtPairedDoc (OUT-2/arm 2) consults it directly since
// 2026-10-03; this helper predates that wiring and now serves the test suite.
func (w *World) courtGatedTask(s *loop.Session, doc string) bool {
	if doc == "" {
		return false
	}
	rel := normRel(doc)
	w.mu.Lock()
	defer w.mu.Unlock()
	if m := w.courtWrote[s.Engine]; m != nil && m[rel].Writer != "" {
		return true
	}
	if name := s.Engine.As; name != "" {
		if m := w.courtWrites[name]; m != nil && m[rel].Writer != "" {
			return true
		}
	}
	return false
}

func rankDirs(rs Ranks) []string {
	var out []string
	seen := map[string]bool{}
	for _, r := range rs {
		if r.Dir != "" && !seen[r.Dir] {
			seen[r.Dir] = true
			out = append(out, r.Dir)
		}
	}
	return out
}

func wroteOf(r *loop.Result) []string {
	if len(r.Wrote) > 0 {
		return r.Wrote // every write the Court's gate saw, whatever tool made it
	}
	var out []string
	for _, st := range r.Steps {
		out = append(out, st.Wrote...)
	}
	return out
}
