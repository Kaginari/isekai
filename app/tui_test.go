package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Kaginari/isekai/loop"
	"github.com/Kaginari/isekai/tool"
	"github.com/Kaginari/isekai/tui"
	"github.com/Kaginari/isekai/wire"
	"github.com/Kaginari/isekai/world"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"
)

// The junction of the terminal UI with the real engine: the app's host under teatest, the mock
// provider scripted, real tools and the real classifier behind the gate.

type tmSender struct{ tm *teatest.TestModel }

func (s tmSender) Send(m tea.Msg) { s.tm.Send(m) }

type screen struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *screen) follow(r io.Reader) {
	go func() {
		b := make([]byte, 4096)
		for {
			n, err := r.Read(b)
			if n > 0 {
				s.mu.Lock()
				s.buf.Write(b[:n])
				s.mu.Unlock()
			}
			if err == io.EOF {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			if err != nil {
				return
			}
		}
	}()
}

func (s *screen) text() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return ansi.Strip(s.buf.String())
}

func (s *screen) wait(t testing.TB, want string) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for !strings.Contains(s.text(), want) {
		if time.Now().After(deadline) {
			t.Fatalf("waited for %q; the screen so far:\n%s", want, s.text())
		}
		time.Sleep(15 * time.Millisecond)
	}
}

// tuiSessionHost is tuiSession keeping the host — tests that drive the host's own seams
// (Queue, Interrupt) rather than only keys need it.
func tuiSessionHost(t *testing.T, w *testWorld) (*tuiHost, *teatest.TestModel, *screen) {
	t.Helper()
	a := w.open()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	h, err := a.tuiHost(ctx)
	if err != nil {
		t.Fatal(err)
	}
	m := tui.New(h, tui.NewTheme(true), h.words)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 30))
	m.Attach(tm.Send)
	h.attach(tmSender{tm})
	sc := &screen{}
	sc.follow(tm.Output())
	sc.wait(t, "✦ isekai")
	t.Cleanup(func() {
		tm.Send(tui.EvQuit{})
		tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
	})
	return h, tm, sc
}

// tuiSession opens the app on a test world and drives its host under teatest.
func tuiSession(t *testing.T, w *testWorld) (*App, *teatest.TestModel, *screen) {
	t.Helper()
	a := w.open()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	h, err := a.tuiHost(ctx)
	if err != nil {
		t.Fatal(err)
	}
	m := tui.New(h, tui.NewTheme(true), h.words)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 30))
	m.Attach(tm.Send)
	h.attach(tmSender{tm})
	sc := &screen{}
	sc.follow(tm.Output())
	sc.wait(t, "✦ isekai")
	t.Cleanup(func() {
		tm.Send(tui.EvQuit{})
		tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
	})
	return a, tm, sc
}

func enter(tm *teatest.TestModel) { tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter}) }

func TestTUIJunctionTurn(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.script(".isekai/tmp/s.json",
		when("say hi", map[string]interface{}{"text": "Let me run it.\n\n", "calls": []interface{}{map[string]interface{}{"name": "bash", "input": map[string]string{"command": "echo hi; echo there"}}}}),
		when("there", call("edit", map[string]string{"path": "README.md", "old": "hello", "new": "hello\n\nthe login zone"})),
		when("edited", text("All **done**: `hi` was said and the file edited.\n\n```go\npackage auth\n```\n")),
	)
	// Law 7 (throne-never-authors) is default-on; this rendering test's mock throne edit of
	// no-creature soil (README.md) must stay lawful — the fixture opts the flag off.
	w.write(".isekai/config.yaml", mockCfg(".isekai/tmp/s.json", "law: {territory: {throneNeverAuthors: false}}\n"))
	_, tm, sc := tuiSession(t, w)
	tm.Type("say hi")
	enter(tm)
	sc.wait(t, "> say hi")
	sc.wait(t, "● bash  echo hi; echo there  read")
	sc.wait(t, "⎿ ok · ")
	sc.wait(t, "there")
	sc.wait(t, "● edit  README.md  write")
	sc.wait(t, " ⎿ ok · ") // the diff is ctrl+o territory now, never ambient
	tm.Send(tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl})
	sc.wait(t, "+ the login zone")
	sc.wait(t, "hi was said and the file edited")
	sc.wait(t, "✓ gate · ")
	s := sc.text()
	if strings.Index(s, "Let me run it.") > strings.Index(s, "● bash") {
		t.Fatalf("the streamed text lands before the tool block:\n%s", s)
	}
	if !strings.Contains(s, "+2 −0") {
		t.Fatalf("the diff stat is missing:\n%s", s)
	}
	if !strings.Contains(w.read(".isekai/log.md"), "Gate") {
		t.Fatal("the gate verdict did not reach log.md")
	}
	// the footer reads the instruments
	sc.wait(t, "mock/m · ctx ")
	// a slash command through the menu
	tm.Type("/stat")
	sc.wait(t, "❯ /status")
	enter(tm)
	sc.wait(t, "sandbox: none")
}

func TestTUIApprovalYes(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.script(".isekai/tmp/s.json",
		when("push it", call("bash", map[string]string{"command": "git push origin main"})),
		text("push attempted"),
	)
	w.write(".isekai/config.yaml", mockCfg(".isekai/tmp/s.json", ""))
	_, tm, sc := tuiSession(t, w)
	tm.Type("push it")
	enter(tm)
	sc.wait(t, "1 · Yes")
	if s := sc.text(); !strings.Contains(s, "outward  bash  git push origin main") || !strings.Contains(s, "don't ask again for bash:git push*") {
		t.Fatalf("approval block:\n%s", s)
	}
	enter(tm)
	sc.wait(t, "✓ Yes")
	sc.wait(t, "push attempted")
	if j := w.journalText(".isekai"); !strings.Contains(j, `"by":"tty"`) || !strings.Contains(j, `"decision":"approved"`) {
		t.Fatalf("the journal does not carry the human's approval:\n%s", j)
	}
	if _, err := readFile(w, ".isekai/config.local.yaml"); err == nil {
		t.Fatal("a plain yes must not write a rule")
	}
}

func TestTUIApprovalDontAskAgain(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.script(".isekai/tmp/s.json",
		when("push it", call("bash", map[string]string{"command": "git push origin main"})),
		call("bash", map[string]string{"command": "git push origin main"}),
		text("pushed twice"),
	)
	w.write(".isekai/config.yaml", mockCfg(".isekai/tmp/s.json", ""))
	w.write(".isekai/config.local.yaml", "permissions:\n  rules:\n    - {match: \"bash:make *\", action: allow}\n")
	_, tm, sc := tuiSession(t, w)
	tm.Type("push it")
	enter(tm)
	sc.wait(t, "1 · Yes")
	tm.Type("2")
	sc.wait(t, "rule bash:git push* → allow written to .isekai/config.local.yaml")
	sc.wait(t, "pushed twice")
	local, _ := readFile(w, ".isekai/config.local.yaml")
	if !strings.Contains(local, `bash:git push*`) || !strings.Contains(local, `bash:make *`) {
		t.Fatalf("the local layer lost a rule or gained none:\n%s", local)
	}
	j := w.journalText(".isekai")
	if strings.Count(j, `"by":"tty"`) != 1 || !strings.Contains(j, `"by":"rule"`) {
		t.Fatalf("the second push must pass by the rule, live:\n%s", j)
	}
}

func TestTUIApprovalNoWithReason(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.script(".isekai/tmp/s.json",
		when("push it", call("bash", map[string]string{"command": "git push origin main"})),
		when("run the tests first", text("Understood: I will run the tests before any push.")),
	)
	w.write(".isekai/config.yaml", mockCfg(".isekai/tmp/s.json", ""))
	_, tm, sc := tuiSession(t, w)
	tm.Type("push it")
	enter(tm)
	sc.wait(t, "1 · Yes")
	tm.Type("3")
	sc.wait(t, "why not?")
	tm.Type("run the tests first")
	enter(tm)
	sc.wait(t, "✓ No, and tell the model why: run the tests first")
	sc.wait(t, "denied at the gate")
	sc.wait(t, "> run the tests first")
	sc.wait(t, "I will run the tests before any push")
	j := w.journalText(".isekai")
	if !strings.Contains(j, `"by":"tty"`) || !strings.Contains(j, `"decision":"denied"`) {
		t.Fatalf("the journal does not carry the denial:\n%s", j)
	}
	if !strings.Contains(w.read(".isekai/instruments/loop/"+firstJournal(w)), "the human said: run the tests first") {
		t.Fatal("the reason did not reach the model's tool result")
	}
}

func TestTUICourtsAndQuestion(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.script(".isekai/tmp/session.json",
		when("look", call("dispatch", map[string]interface{}{"body": "slime-auth", "ask": "findings: look at the login zone"})),
		when("@F src/auth/login.go:1 fine", call("dispatch", map[string]interface{}{"body": "slime-api", "ask": "findings: look at the api", "background": true})),
		when("started in the background", call("ask", map[string]interface{}{"question": "Which zone next?", "options": []string{"auth", "api"}})),
		when("api", text("The api zone it is.")),
		when("[court slime-api reported]", text("Both courts reported.")),
	)
	w.script(".isekai/tmp/court.json",
		when("login zone", text("@S PASS\n@F src/auth/login.go:1 fine\n@? no test for refresh\n@U colony courts can run behind\n@E 80")),
		when("the api", call("bash", map[string]string{"command": "sleep 1; echo slept"})),
		when("slept", text("@S PASS\n@F src/api/server.go:1 fine too\n@U territory the api zone serves json\n@E 70")),
	)
	w.write(".isekai/config.yaml", `
models: {default: mock/m, offices: {great-sage: court/c}}
providers:
  mock: {enabled: true, script: .isekai/tmp/session.json}
  court: {type: mock, script: .isekai/tmp/court.json}
tools: {bash: {sandbox: none}, dispatch: {background: true, strictFlow: false}}
ui: {board: {autostart: false}}
`)
	_, tm, sc := tuiSession(t, w)
	tm.Type("look")
	enter(tm)
	sc.wait(t, "→ court slime-auth started")
	sc.wait(t, "◆ court slime-auth · slime · great-sage · court/c")
	sc.wait(t, "⎿ PASS")
	sc.wait(t, "● src/auth/login.go:1 fine")
	sc.wait(t, "∴ colony: courts can run behind")
	sc.wait(t, "1 · auth")
	tm.Type("2")
	sc.wait(t, "The api zone it is.")
	sc.wait(t, "◆ court slime-api")
	sc.wait(t, "● src/api/server.go:1 fine too")
	sc.wait(t, "↻ continuing with what arrived")
	sc.wait(t, "Both courts reported.")
}

func TestTUIInterruptAndQueue(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	// "a note" must NEVER appear as a turn prompt in the running "slow" turn (no
	// mid-turn injection). It runs as its own turn after esc pivots to it.
	w.script(".isekai/tmp/s.json",
		when("slow", call("bash", map[string]string{"command": "sleep 5; echo late"})),
		when("late", text("late came")),
		when("a note", text("the note ran as its own turn")),
	)
	w.write(".isekai/config.yaml", mockCfg(".isekai/tmp/s.json", ""))
	_, tm, sc := tuiSession(t, w)
	tm.Type("slow")
	enter(tm)
	sc.wait(t, "Running bash…")
	tm.Type("a note")
	enter(tm)
	sc.wait(t, "⏎ queued: a note")
	// esc pivots: the running turn aborts and the queued line fires as the next turn.
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEsc})
	sc.wait(t, "■ interrupted")
	sc.wait(t, "the note ran as its own turn")
	// the FIFO contract: "a note" waited for "slow" to end, then ran as its OWN turn —
	// never a mid-turn injection into the running turn's messages. Assert via the queue's
	// observable ordering: the queued echo precedes the interrupt, and the note's turn
	// output follows it. A regression that injects mid-turn would show the note reaching
	// the model while "slow" was still running — the mock script has no step answering a
	// mid-turn "a note", so any injection yields no turn and this wait fails:
	if !strings.Contains(sc.text(), "⏎ queued: a note") {
		t.Fatalf("the queued echo vanished:\n%s", sc.text())
	}
}

func TestTUIGateCardFreedByInterrupt(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	// Freeze-law regression (2026-10-07): a gate card left unanswered wedged the whole
	// world in h.approve's session-ctx select — esc could not free it, queued asks never
	// drained. The card's wait now selects on the turn ctx too: esc resolves the card as
	// an interruption-denial AND the queued ask runs as the next turn.
	w.script(".isekai/tmp/s.json",
		when("push it", call("bash", map[string]string{"command": "git push origin main"})),
		when("hello", text("hello back")),
	)
	w.write(".isekai/config.yaml", mockCfg(".isekai/tmp/s.json", ""))
	h, tm, sc := tuiSessionHost(t, w) // keeps the host: the queue entry point rides it directly
	tm.Type("push it")
	enter(tm)
	sc.wait(t, "1 · Yes") // the card is up; nothing answers it
	// the human's voice arrives the DASHBOARD way while the card holds the gate
	// (the 2026-10-07 freeze's real buffer: Queue → queuedBuf → drained only at finish)
	if why := h.Queue("hello"); !strings.Contains(why, "queued") {
		t.Fatalf("the queue word did not read as queued: %q", why)
	}
	if strings.Contains(sc.text(), "hello back") {
		t.Fatal("the world answered while its card was open — the queue must wait")
	}
	// the dashboard's interrupt endpoint (what a human hits while the modal holds the
	// world) calls exactly this: Interrupt → turn ctx fires → the card's wait resolves
	h.Interrupt()
	sc.wait(t, "■ interrupted") // freed — the world was never stuck on a modal
	sc.wait(t, "hello back")    // and the queued ask ran as the very next turn
	j := w.journalText(".isekai")
	if !strings.Contains(j, `"decision":"denied"`) || !strings.Contains(j, "interrupted while waiting") {
		t.Fatalf("the freed card did not journal as an interruption-denial:\n%s", j)
	}
}

func TestWantsTUI(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.script(".isekai/tmp/s.json", text("nothing"))
	w.write(".isekai/config.yaml", mockCfg(".isekai/tmp/s.json", ""))
	a := w.open()
	if a.wantsTUI(false) {
		t.Fatal("a pipe must keep the line REPL")
	}
	if a.wantsTUI(true) {
		t.Fatal("--plain must keep the line REPL")
	}
}

func readFile(w *testWorld, p string) (string, error) {
	b, err := os.ReadFile(filepath.Join(w.root, p))
	return string(b), err
}

func firstJournal(w *testWorld) string {
	ents, _ := os.ReadDir(filepath.Join(w.root, ".isekai/instruments/loop"))
	for _, e := range ents {
		if strings.HasSuffix(e.Name(), ".jsonl") {
			return e.Name()
		}
	}
	return ""
}

func TestTUIStartFailureReachesTheTerminal(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.write(".isekai/config.yaml", "providers:\n  vllm: {type: openai, baseURL: http://127.0.0.1:1/v1, apiKeyEnv: VLLM_API_KEY}\nmodels: {default: vllm/fake/model}\n")
	a := w.open()
	errw := a.Opt.Err
	if code := a.TUI(context.Background()); code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if got := errw.(*strings.Builder).String(); !strings.Contains(got, "VLLM_API_KEY is not set") {
		t.Fatalf("a TUI that cannot start must say why on stderr; got %q", got)
	}
}

// /board's view is the real world's: the reasoned ontology as a graph with its bonds and
// knowledge, the triad as config resolves it.
func TestTUIBoardReadsTheWorld(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.script(".isekai/tmp/s.json", text("nothing"))
	w.write(".isekai/config.yaml", mockCfg(".isekai/tmp/s.json", ""))
	a := w.open()
	h, err := a.tuiHost(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	v := h.Board("24h")
	byID := map[string]tui.GraphNode{}
	for _, n := range v.Graph.Nodes {
		byID[n.ID] = n
	}
	slime, ok := byID["slime-auth"]
	if !ok || len(byID) != 5 {
		t.Fatalf("graph nodes %v, want rimuru, elf-core, orc-security, slime-auth, slime-api", v.Graph.Nodes)
	}
	if len(slime.Up) != 1 || slime.Up[0].To != "orc-security" || slime.Up[0].Bond != "truth" || slime.Rank != "slime" {
		t.Fatalf("slime-auth answers %v (rank %q), want orc-security by truth", slime.Up, slime.Rank)
	}
	if byID["rimuru"].Level != 0 || byID["elf-core"].Level != 1 || byID["orc-security"].Level != 2 || slime.Level != 3 {
		t.Fatalf("levels: %+v", byID)
	}
	know := ""
	for _, kv := range slime.Knowledge {
		know += kv.Key + "=" + kv.Value + "\n"
	}
	if !strings.Contains(know, "chain=") || !strings.Contains(know, "rimuru") || !strings.Contains(know, "owns=src/auth") {
		t.Fatalf("slime-auth's knowledge misses the inferred chain or its territory:\n%s", know)
	}
	if !strings.Contains(v.Graph.Summary, "5 creatures") || !strings.Contains(v.Graph.Summary, "reasoned") {
		t.Fatalf("summary %q", v.Graph.Summary)
	}
	if len(v.Offices) != 3 || v.Offices[0].Name != "great-sage" || v.Offices[0].Model != "mock/m" || len(v.Offices[0].Ranks) == 0 {
		t.Fatalf("offices %+v", v.Offices)
	}
}

// collector is a test sender that records the events a tuiHost emits.
type collector struct {
	mu   sync.Mutex
	msgs []tea.Msg
}

func (c *collector) Send(m tea.Msg) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.msgs = append(c.msgs, m)
}

func (c *collector) events() []tea.Msg {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]tea.Msg(nil), c.msgs...)
}

func (c *collector) views() (courts []tui.CourtView, states []tui.EvState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, m := range c.msgs {
		switch e := m.(type) {
		case tui.EvCourt:
			courts = append(courts, e.Court)
		case tui.EvState:
			states = append(states, e)
		}
	}
	return
}

// A body dispatched down the ladder by a drained @$ order never lands a courtReport — its
// terminal state arrives only through the loop's State hook ("done"). The banner must close
// there, exactly once, instead of staying "thinking" with a running timer (the stale-court bug
// Veldora saw: banners reading thinking·44m after the loop journal already said DONE).
func TestTUICourtBannerClosesOnDoneState(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.write(".isekai/config.yaml", `
models: {default: mock/m, offices: {great-sage: court/c}}
providers:
  mock: {enabled: true}
  court: {type: mock}
ui: {board: {autostart: false}}
`)
	a := w.open()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h, err := a.tuiHost(ctx)
	if err != nil {
		t.Fatal(err)
	}
	c := &collector{}
	h.attach(c)

	// the Court knows the body (as Track would have registered it at dispatch)
	h.a.Court.Track("slime-ledger", "slime", "great-sage", "court/c", nil, 1)

	h.state("slime-ledger", "thinking") // opens the banner
	h.state("slime-ledger", "tool")     // mid-run updates ride EvState
	h.state("slime-ledger", "done")     // the drained body's only terminal signal
	h.state("slime-ledger", "done")     // idempotent: no second banner, no panic

	courts, states := c.views()
	var opens, closes int
	for _, v := range courts {
		if v.Name != "slime-ledger" {
			continue
		}
		switch v.State {
		case "thinking":
			opens++
		case "done":
			closes++
			if v.Elapsed < 0 {
				t.Fatalf("closing banner carries a negative elapsed: %v", v.Elapsed)
			}
		}
	}
	if opens != 1 || closes != 1 {
		t.Fatalf("banners: %d opens, %d closes — want 1/1; all courts: %+v", opens, closes, courts)
	}
	for _, s := range states {
		if s.Body == "slime-ledger" && s.State == "done" {
			t.Fatalf("a bare done EvState escaped — the banner would stay open: %+v", states)
		}
	}
	h.mu.Lock()
	seenLeft := h.courtSeen["slime-ledger"]
	h.mu.Unlock()
	if seenLeft {
		t.Fatal("courtSeen still holds the body after done — the banner is still live")
	}
}

// A court's dispatch step seeds and lands its Court's card at EVERY depth (Veldora's chain
// run, 2026-09-28): rimuru → elf-colony → orc-experience closed DOWN the loop, and the orc's
// card sat in the live area reading DONE long after the body's clock axed it — the host saw
// the orc's steps only as elf-colony's steps, so the report never reached the card. Now any
// body's dispatch step end lands the child's card (its parent's log line rides along), and the
// session's own dispatch steps behave exactly as before.
func TestTUINestedDispatchStepsLandTheirCourts(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.script(".isekai/tmp/s.json", text("nothing"))
	w.write(".isekai/config.yaml", mockCfg(".isekai/tmp/s.json", ""))
	a := w.open()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h, err := a.tuiHost(ctx)
	if err != nil {
		t.Fatal(err)
	}
	c := &collector{}
	h.attach(c)

	parent := &loop.Session{Engine: &loop.Engine{As: "elf-colony"}}
	rimuru := &loop.Session{Engine: &loop.Engine{As: world.Rimuru}}
	input, _ := json.Marshal(map[string]string{"body": "orc-x", "ask": "verdict on the trials"})

	// a court dispatches: the start seeds the child's ask; the parent's log still steps
	h.observe(parent, &loop.StepRecord{Tool: "dispatch", Input: input, Status: "running"}, "start")
	h.mu.Lock()
	seeded := h.courtAsk["orc-x"]
	h.mu.Unlock()
	if seeded != "verdict on the trials" {
		t.Fatalf("the nested dispatch's start must seed the card's ask, got %q", seeded)
	}
	h.a.Court.Track("orc-x", "orc", "raphael", "mock/r", nil, 2)
	h.state("orc-x", "thinking")
	h.state("orc-x", "done")
	// the step end carries the child's wire report: the card gets it
	h.observe(parent, &loop.StepRecord{Tool: "dispatch", Input: input, Status: "done",
		Result: tool.Result{Output: "@S DONE\n@F src/x.go:1 ok\n@U colony nested courts land their cards\n@E 20"}}, "end")
	// a failed dispatch's output is no wire report: no card report, the parent's log still closes
	h.observe(parent, &loop.StepRecord{Tool: "dispatch", Input: input, Status: "failed",
		Result: tool.Result{Output: "dispatch orc-x: court ran out of its 15.0-minute budget", Err: true}}, "end")

	// the session's own dispatch step: unchanged shape — no tool block, no body-log lines
	h.observe(rimuru, &loop.StepRecord{Tool: "dispatch", Input: input, Status: "running"}, "start")
	h.observe(rimuru, &loop.StepRecord{Tool: "dispatch", Input: input, Status: "done",
		Result: tool.Result{Output: "@S DONE\n@F src/x.go:1 ok too\n@E 20"}}, "end")

	courts, _ := c.views()
	var opens, reports int
	for _, v := range courts {
		if v.Name != "orc-x" {
			continue
		}
		if v.State == "thinking" {
			opens++
			if v.Ask != "verdict on the trials" || v.Model != "mock/r" || v.Rank != "orc" || v.Office != "raphael" {
				t.Fatalf("the nested card opened unseeded: %+v", v)
			}
		}
		if v.State == "done" && v.Report != nil {
			reports++
			if v.Exit != "PASSED" {
				t.Fatalf("a DONE journal verdict must read PASSED on the card face, got %q", v.Exit)
			}
		}
	}
	if opens != 1 || reports != 2 {
		t.Fatalf("opens %d, landed reports %d — want 1 open, 2 reports (parent's + session's): %+v", opens, reports, courts)
	}
	var parentLogs, toolBlocks int
	for _, m := range c.events() {
		switch e := m.(type) {
		case tui.EvBodyStep:
			if e.Body == "elf-colony" && e.End {
				parentLogs++
			}
		case tui.EvToolStart, tui.EvToolEnd:
			toolBlocks++
		}
	}
	if parentLogs != 2 {
		t.Fatalf("the parent's own log lost steps to the child's card: %d step ends, want 2 (the reported one and the failed one)", parentLogs)
	}
	if toolBlocks != 0 {
		t.Fatalf("a dispatch step must never draw a generic tool block: %d", toolBlocks)
	}
}

// Court-card exit-verdict law (harness leg 2c): a court whose run journal ended FAIL — the
// wire's @S carries the loop's status — must render "FAIL · Ns" on the card face, styled bad,
// with lifecycle "done" demoted to the dim ask line. Lifecycle returned ≠ verdict; the desk
// must never print a bare "done · 2m35s" over a FAILED journal (the 2026-09-28 Law-9 break).
func TestTUICourtCardRendersExitVerdict(t *testing.T) {
	theme := tui.NewTheme(true)
	rep := wire.Report{Status: "FAIL — PASSED", Findings: []string{"src/x.go:1 broke"}}
	v := tui.CourtView{
		Name: "orc-trials", Rank: "orc", Office: "raphael", Model: "court/c",
		State: "done", Exit: exitVerdict(rep), Failed: true,
		Elapsed: 155 * time.Second, Ask: "rule on the trials",
		Report: reportView(rep), Word: "court",
	}
	if v.Exit != "FAIL" {
		t.Fatalf("exitVerdict(%q) = %q, want FAIL", rep.Status, v.Exit)
	}
	card := ansi.Strip(theme.Court(v, 100))
	if !strings.Contains(card, "FAIL · 2m35s") {
		t.Fatalf("the card face must read the journal verdict before the duration; got:\n%s", card)
	}
	if strings.Contains(card, "DONE · 2m35s") || strings.Contains(card, "done · 2m35s") {
		t.Fatalf("lifecycle done leaked onto the face over a FAIL verdict:\n%s", card)
	}
	// lifecycle survives demoted: the ask line is the tooltip tier
	if !strings.Contains(card, "rule on the trials · done") {
		t.Fatalf("lifecycle word must demote to the ask line; got:\n%s", card)
	}

	verdicts := map[string]string{
		"DONE": "PASSED", "ESCALATE": "ESCALATE", "DENIED": "DENIED", "INTERRUPTED": "INTERRUPTED", "FAIL": "FAIL",
		"": "", "PASSED": "", "SHINY": "", // legacy/junk: no verdict surface
	}
	for status, expect := range verdicts {
		if got := exitVerdict(wire.Report{Status: status}); got != expect {
			t.Fatalf("exitVerdict(%q) = %q, want %q", status, got, expect)
		}
	}
}
