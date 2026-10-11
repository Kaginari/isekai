package app

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/tui"
	"github.com/Kaginari/isekai/world"
)

// law.evolution through the app: config arms the world (mode, desk limit, the founding sketch's
// survey); the run-start sweep weighs a stressed body; propose mode queues; the desk lists,
// shows and applies; status and the dashboard's stream carry it.
func TestEvolutionThroughTheApp(t *testing.T) {
	files := isekaiCreatures()
	files[".isekai/slime/auth/README.md"] = "# slime-auth\n\n- **Rank:** Slime\n- **Territory:** `src/auth/`\n- **Reports to:** orc-security\n- **Verify:** `test -d src/auth`\n\n## Traits\n- tokens expire after one hour\n\n## Thoughts\n" +
		"- 2026-10-01 refresh: always run the token test first\n- 2026-10-02 refresh races the logout\n- 2026-10-03 refresh: use the clock helper\n- 2026-10-04 refresh window is 5m\n"
	files["go.mod"] = "module example.com/w\n\ngo 1.21\n"
	files["src/auth/login_test.go"] = "package auth\n\nimport \"testing\"\n\nfunc TestLogin(t *testing.T) {}\n"
	w := newTestWorld(t, "isekai", files)
	w.write(".isekai/config.yaml", "models: {default: mock/m}\nproviders: {mock: {enabled: true}}\ntools: {bash: {sandbox: none}}\nui: {board: {autostart: false}}\nlaw: {evolution: {mode: propose, thoughtsMax: 3, failStreak: 2}}\n")
	a := w.open()
	o := a.World.Evolution()
	if !o.Enabled || o.Mode != "propose" || o.ThoughtsMax != 3 || o.StressAt != 5 || o.Survey == nil || o.On == nil {
		t.Fatalf("config arms the world: %+v", o)
	}
	sv := a.surveyTerritory([]string{"src/auth/"})
	if sv.Files != 2 || len(sv.Verify) != 1 || !strings.Contains(sv.Verify[0], "go test ./src/auth/") || len(sv.Traits) == 0 || !strings.HasPrefix(sv.Traits[0], "sketch: Go package auth") {
		t.Fatalf("the founding sketch reads the territory: %+v", sv)
	}
	var heard []tui.EvEvolution
	a.evoSink = func(e tui.EvEvolution) { heard = append(heard, e) }
	for i := 0; i < 2; i++ {
		a.World.FoldRun(world.RunFold{Body: "slime-auth", Loop: "DONE", Word: "FAIL", Gate: "fail"})
	}
	a.evolveSweep()
	if len(heard) != 1 || heard[0].Status != "proposed" || heard[0].Arm != "mutate" || heard[0].Body != "slime-auth" {
		t.Fatalf("the sweep proposes and the surfaces hear it: %+v", heard)
	}
	id := heard[0].ID
	var out, errw bytes.Buffer
	io := IO{Out: &out, Err: &errw}
	if code := a.cmdEvolve([]string{"list"}, io); code != 0 || !strings.Contains(out.String(), id) || !strings.Contains(out.String(), "1 open proposal") {
		t.Fatalf("evolve list: %d %s %s", code, out.String(), errw.String())
	}
	out.Reset()
	if code := a.cmdEvolve([]string{"show", id[:6]}, io); code != 0 || !strings.Contains(out.String(), "+- refresh — ") {
		t.Fatalf("evolve show carries the diff: %d\n%s", code, out.String())
	}
	out.Reset()
	if code := a.cmdEvolve([]string{"apply", id}, io); code != 0 || !strings.Contains(out.String(), "@S OK slime-auth mutated its genome") {
		t.Fatalf("evolve apply: %d %s %s", code, out.String(), errw.String())
	}
	doc := w.read(".isekai/slime/auth/README.md")
	if !strings.Contains(doc, "- refresh — ") || !strings.Contains(doc, "slime-auth-refresh") {
		t.Fatalf("applied — the trait and the minted how:\n%s", doc)
	}
	out.Reset()
	if code := a.cmdStatus(io); code != 0 || !strings.Contains(out.String(), "evolution: on · mode propose") || !strings.Contains(out.String(), "slime-auth applied: mutate") {
		t.Fatalf("status shows the evolution:\n%s", out.String())
	}
	ev := webEvent(heard[len(heard)-1])
	if ev["type"] != "evolution" || ev["body"] != "slime-auth" || ev["status"] != "applied" {
		t.Fatalf("the dashboard's stream carries it: %v", ev)
	}
	if code := a.cmdEvolve([]string{"reject", id, "late"}, io); code == 0 {
		t.Fatal("an applied proposal is not rejected")
	}
}

// The court seam end to end: the throne dispatches a slime, the court reports FAIL, the strain
// crosses stressAt, its desk is past the limit — it evolves inside the dispatch, the report says
// so, and the throne's own gate never takes the harness's writes for its own.
func TestEvolutionAtTheCourtSeam(t *testing.T) {
	files := isekaiCreatures()
	var desk []string
	for i := 1; i <= 8; i++ {
		desk = append(desk, fmt.Sprintf("- 2026-10-%02d session: run the token test before the refresh (%d)", i, i))
	}
	files[".isekai/slime/auth/README.md"] = "# slime-auth\n\n- **Rank:** Slime\n- **Territory:** `src/auth/`\n- **Reports to:** orc-security\n- **Verify:** `test -d src/auth`\n\n## Thoughts\n" + strings.Join(desk, "\n") + "\n"
	w := newTestWorld(t, "isekai", files)
	w.script(".isekai/tmp/s.json",
		when("go", map[string]interface{}{"calls": []interface{}{map[string]interface{}{"name": "dispatch", "input": map[string]string{"body": "slime-auth", "ask": "settle the login"}}}}),
		when("settle the login", text("@S FAIL\n@? the login would not settle\n@U colony nothing new\n@E 10")),
		text("noted"),
	)
	w.write(".isekai/config.yaml", "models: {default: mock/m}\nproviders: {mock: {enabled: true, script: .isekai/tmp/s.json}}\ntools: {bash: {sandbox: none}, dispatch: {strictFlow: false, lineage: false}}\nui: {board: {autostart: false}}\nlaw: {evolution: {stressAt: 1}}\n")
	a := w.open()
	e, err := a.Engine()
	if err != nil {
		t.Fatal(err)
	}
	r, err := e.Run(context.Background(), "go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Steps[0].Result.Output, "[evolution] slime-auth mutated its genome") {
		t.Fatalf("the court's report carries the evolution:\n%s", r.Steps[0].Result.Output)
	}
	for _, h := range r.Holes {
		if strings.Contains(h, "gate") {
			t.Fatalf("the throne's gate took the harness's writes for its own: %v", r.Holes)
		}
	}
	if doc := w.read(".isekai/slime/auth/README.md"); !strings.Contains(doc, "- refresh — ") || !strings.Contains(doc, "slime-auth-refresh") {
		t.Fatalf("the genome mutated:\n%s", doc)
	}
	if !strings.Contains(w.read(".isekai/log.md"), "evolution — slime-auth mutated its genome") {
		t.Fatal("logged")
	}
}
