package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Kaginari/isekai/wire"
)

// Commission is what the dispatch tool hands its runner: a body to mint and the wire
// commission it carries (@ROOT @SCOPE @ASK @CAP).
type Commission struct {
	Body   string
	Ask    string
	Scope  string
	Root   string
	Unsaid bool
	// UnsaidExplicit marks a demand made on purpose (the caller passed unsaid:true), vs the
	// law's ambient default. Only explicit demands make the close-out bounce for a missing @U.
	UnsaidExplicit bool
	Cap            int
	Depth          int    // the Court's nesting depth (1 for a body dispatched by the session)
	Office         string // the office the ask names: findings → great-sage, verdict → raphael, draft → ciel
}

// OfficeOf reads the office off an @ASK word (the triad: perceive, judge, speak); "" if none.
func OfficeOf(ask string) string {
	for _, w := range strings.Fields(strings.ToLower(ask)) {
		switch w {
		case "findings":
			return "great-sage"
		case "verdict":
			return "raphael"
		case "draft":
			return "ciel"
		}
	}
	return ""
}

// String renders the commission on the wire.
func (c Commission) String() string {
	var b strings.Builder
	if c.Root != "" {
		fmt.Fprintf(&b, "@ROOT %s\n", c.Root)
	}
	if c.Scope != "" {
		fmt.Fprintf(&b, "@SCOPE %s\n", c.Scope)
	}
	ask := "findings"
	switch c.Office {
	case "raphael":
		ask = "verdict"
	case "ciel":
		ask = "draft"
	}
	if c.Unsaid {
		ask += " +unsaid"
	}
	fmt.Fprintf(&b, "@ASK %s\n", ask)
	if c.Cap > 0 {
		fmt.Fprintf(&b, "@CAP %d\n", c.Cap)
	}
	b.WriteString(strings.TrimSpace(c.Ask) + "\n")
	return b.String()
}

// DispatchReport is what a runner brings back: the Court's wire report (under @CAP) and the
// readings the dispatcher needs.
type DispatchReport struct {
	Text    string   // the report, ready for the model
	Status  string   // the run's status word
	Unsaid  int      // @U lines carried
	Failed  bool     // the run failed (the act failed)
	Holes   []string // the engine's own holes
	Wrote   []string // paths the Court wrote (gated on its own account)
	Journal string
	// ReportPath/ReportBytes/ReportHash are the courier envelope: the harness filed the
	// report on the instruments shelf, and the dispatcher's context carries the digest
	// (@R path · size · @H fnv64) with the wrote list (@W) inline. The parent reads the
	// file when the verdict needs the flesh — relaying prose up the rungs is the tax this
	// closes.
	ReportPath  string
	ReportBytes int
	ReportHash  string
	// Forges are the run's drafts: skills or tools the body proved on the wire (@FORGE). They
	// ride up unminted — the layer curates, the desk mints.
	Forges []wire.Forge
}

// Dispatcher runs one commission as a Court Body: a fresh loop with its own context, tools cut
// to rank and territory, one wire report back. The world package implements it; this package
// cannot import the loop.
type Dispatcher func(ctx context.Context, env Env, c Commission) (DispatchReport, error)

// DispatchOptions is tools.dispatch.
type DispatchOptions struct {
	Enabled    bool
	Depth      int      // the depth of the body holding this tool (0 for the session)
	MaxDepth   int      // 0 = 1: a Court may not dispatch
	Cap        int      // default @CAP for a Court's report (0 = the wire's default)
	Unsaid     bool     // commission +unsaid by default
	Bodies     []string // allowed body names; nil = any
	Background bool     // `background: true` may run the Court asynchronously (tools.dispatch.background)
	// BackgroundTimeout bounds a background Court's run (0: 30m). The provider's own
	// timeout bounds each model call; nothing else stopped a background run from
	// outliving everything around it.
	BackgroundTimeout time.Duration
	// SessionBackground: at the session's own depth a commission defaults to background — the
	// throne stays free; the Court rides on and its report wakes the session. A commission may
	// still name background:false to ride in the foreground.
	SessionBackground bool
	// Wake receives a background Court's report when it lands; nil: background is refused.
	Wake func(body, report string, failed bool)
}

// DispatchTool is `dispatch`: mint a named body for one commission and return its report. A
// report with no @U line is flagged to the dispatcher (the unsaid).
func DispatchTool(opt DispatchOptions, run Dispatcher) *Tool {
	return &Tool{
		Name:        "dispatch",
		Description: "Dispatch a Court Body: a named body (slime-<zone>, orc-<domain>, elf-<name>) works one commission in a fresh context, tools cut to its rank and territory, and answers one wire report. Its context dies with the task.",
		Schema:      json.RawMessage(`{"type":"object","properties":{"body":{"type":"string"},"ask":{"type":"string"},"scope":{"type":"string"},"office":{"type":"string","enum":["great-sage","raphael","ciel"],"description":"findings (great-sage), verdict (raphael) or draft (ciel); default by the ask"},"cap":{"type":"integer"},"unsaid":{"type":"boolean"},"background":{"type":"boolean","description":"run the Court in the background; its report wakes you when it lands"}},"required":["body","ask"]}`),
		Class:       Write,
		Classify: func(env Env, in json.RawMessage) Classification {
			var a struct{ Body string }
			_ = decode(in, &a)
			return Classification{Class: Write, Why: "dispatches " + a.Body}
		},
		Run: func(ctx context.Context, env Env, in json.RawMessage) Result {
			var a struct {
				Body, Ask, Scope, Office string
				Cap                      int
				Unsaid                   *bool
				Background               *bool
			}
			if err := decode(in, &a); err != nil {
				return fail("dispatch: %v", err)
			}
			if strings.TrimSpace(a.Body) == "" || strings.TrimSpace(a.Ask) == "" {
				return fail("dispatch: body and ask are required")
			}
			if !opt.Enabled {
				return fail("dispatch is off (tools.dispatch.enabled: false)")
			}
			max := opt.MaxDepth
			if max <= 0 {
				max = 1
			}
			if opt.Depth >= max {
				// the world floors its own ceiling at the rank chain's length (world.Engine), so this
				// refusal only answers at the floor rank or under a human-set MaxDepth below it
				return fail("dispatch: nesting limit reached (depth %d of %d) — a Court at this depth may not dispatch; answer with @? one hop up instead", opt.Depth, max)
			}
			if opt.Bodies != nil && !inList(opt.Bodies, a.Body) {
				return fail("dispatch: %q is not a body this session may dispatch (%s)", a.Body, strings.Join(opt.Bodies, ", "))
			}
			if run == nil {
				return fail("dispatch: no runner wired")
			}
			c := Commission{Body: strings.TrimSpace(a.Body), Ask: a.Ask, Scope: a.Scope, Root: env.Root, Cap: a.Cap, Depth: opt.Depth + 1, Unsaid: opt.Unsaid, Office: a.Office}
			if c.Office == "" {
				c.Office = OfficeOf(a.Ask)
			}
			if c.Office == "" {
				c.Office = "great-sage"
			}
			if c.Cap == 0 {
				c.Cap = opt.Cap
			}
			if a.Unsaid != nil {
				c.Unsaid = *a.Unsaid
				c.UnsaidExplicit = true // the caller named +unsaid in words — not the ambient law
			}
			background := opt.Background && opt.SessionBackground && opt.Depth == 0
			if a.Background != nil {
				background = *a.Background
			}
			render := func(rep DispatchReport) string {
				// court-card-exit-verdict law (harness leg 2b): the exit verdict is the run
				// journal's t:end status (rep.Status), stamped as the payload's first line so
				// any renderer that reads only the returned string still shows DONE/FAIL/
				// ESCALATE — never a collapsed lifecycle "done". The body's own @S follows.
				var b strings.Builder
				if rep.Status != "" {
					b.WriteString("@S " + rep.Status + "\n")
				}
				text := strings.TrimSpace(rep.Text)
				if rep.ReportPath != "" {
					// the courier envelope: the report is filed on the instruments shelf; the
					// parent's context buys only the digest — @R where it lives, @H its
					// integrity, @W what the court wrote, a head-digest of the first lines.
					fmt.Fprintf(&b, "@R %s — the report is filed there (%d bytes); read it when the verdict needs the flesh\n", rep.ReportPath, rep.ReportBytes)
					fmt.Fprintf(&b, "@H fnv64:%s\n", rep.ReportHash)
					for _, p := range rep.Wrote {
						b.WriteString("@W " + p + "\n")
					}
					if text != "" {
						head := text
						if ls := strings.Split(text, "\n"); len(ls) > 8 {
							head = strings.Join(ls[:8], "\n") + "\n…"
						}
						b.WriteString("digest:\n" + head + "\n")
					}
				} else {
					b.WriteString(text)
				}
				out := b.String()
				if rep.Unsaid == 0 {
					out += fmt.Sprintf("\n@? dispatch %s: report carries no @U line — nothing unsaid, or the duty failed; ask again with +unsaid", c.Body)
				}
				return out
			}
			if background {
				if !opt.Background || opt.Wake == nil {
					return fail("dispatch: background Courts are off here (tools.dispatch.background: false, or no inbox delivery wired at depth %d) — dispatch %s in the foreground", opt.Depth, c.Body)
				}
				dispatchedAt := time.Now().UnixNano()
				go spawnWatch(env, c.Body, dispatchedAt, opt.Wake)
				go func() {
					defer func() {
						// a background runner panicking must never take the process down
						if rec := recover(); rec != nil {
							debug.PrintStack()
							opt.Wake(c.Body, fmt.Sprintf("@S FAIL\n@? dispatch %s: panic: %v\n@E 0", c.Body, rec), true)
						}
					}()
					t := opt.BackgroundTimeout
					if t <= 0 {
						t = 30 * time.Minute
					}
					bctx, cancel := context.WithTimeout(context.Background(), t)
					defer cancel()
					rep, err := run(bctx, env, c)
					if err != nil {
						opt.Wake(c.Body, fmt.Sprintf("@S FAIL\n@? dispatch %s: %v\n@E 0", c.Body, err), true)
						return
					}
					opt.Wake(c.Body, render(rep), rep.Failed)
				}()
				return Result{Output: fmt.Sprintf("court %s started in the background on `%s` — keep working; its report wakes you when it lands (or /send %s <text> to reach it)", c.Body, oneLineAsk(c.Ask), c.Body)}
			}
			rep, err := run(ctx, env, c)
			if err != nil {
				return fail("dispatch %s: %v", c.Body, err)
			}
			// the Court's writes are named so the dispatcher's record is honest; they were gated
			// on the Court's own account and its dispatcher's gate leaves them alone
			return Result{Output: render(rep), Err: rep.Failed, Wrote: rep.Wrote}
		},
	}
}

func oneLineAsk(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 80 {
		s = s[:80] + "…"
	}
	return s
}

// spawnWatch is the phantom-dispatch guard (audit N5, 2026-10-02 evening): an acked background
// dispatch once returned ok while the child never minted its journal — the parent closed DONE
// believing the slice was staffed and the work was nobody's. Within spawnGrace of the ack, the
// child's journal must exist under <root>/<worldDir>/instruments/loop/<body>-<ts>-<id>.jsonl
// and have a first event. No journal, no first event → the dispatcher is woken with a
// spawn-failed FAIL on the wire, naming what to trust instead.
const spawnGrace = 10 * time.Second

func spawnWatch(env Env, body string, dispatchedAt int64, wake func(body, report string, failed bool)) {
	if wake == nil || body == "" {
		return
	}
	dir := env.Root + "/" + env.worldDir() + "/instruments/loop"
	deadline := time.Now().Add(spawnGrace)
	for time.Now().Before(deadline) {
		ents, err := os.ReadDir(dir)
		if err == nil {
			for _, ent := range ents {
				if ent.IsDir() || !strings.HasPrefix(ent.Name(), body+"-") || !strings.HasSuffix(ent.Name(), ".jsonl") {
					continue
				}
				info, err := ent.Info()
				if err != nil {
					continue
				}
				if info.ModTime().UnixNano() < dispatchedAt {
					continue // a stale journal from an earlier run of the same body
				}
				if info.Size() == 0 {
					continue // created but no first event yet
				}
				return // alive and writing
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	wake(body, fmt.Sprintf("@S FAIL\n@? dispatch %s: spawn failed — %s passed and no journal with a first event appeared in %s. The slice was never staffed; re-dispatch or escalate with @?, do not assume the child is alive.\n@E 0", body, dir, spawnGrace), true)
}

func inList(xs []string, x string) bool {
	for _, y := range xs {
		if strings.EqualFold(y, x) {
			return true
		}
	}
	return false
}
