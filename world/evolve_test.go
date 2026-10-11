package world

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/memory"
	"github.com/Kaginari/isekai/onto"
	"github.com/Kaginari/isekai/value"
)

// evoWorld writes a fixture world: an orc over one slime whose zone is src/pay/ (two
// sub-directories with code), the slime's doc carrying the given Thoughts.
func evoWorld(t *testing.T, thoughts []string, extra map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	desk := strings.Join(thoughts, "\n")
	files := map[string]string{
		".isekai/isekai.md":           lawText,
		".isekai/log.md":              "# Chronicle\n",
		".isekai/elf/core/README.md":  "# elf-core\n\n- **Rank:** Elf\n- **Territory:** `docs/`\n- **Reports to:** rimuru\n",
		".isekai/orc/pay/README.md":   "# orc-pay\n\n- **Rank:** Orc\n- **Territory:** `src/`\n- **Reports to:** elf-core\n",
		".isekai/slime/pay/README.md": "# slime-pay\n\n- **Rank:** Slime\n- **Territory:** `src/pay/`\n- **Reports to:** orc-pay\n- **Minds:** none\n- **Purpose:** the payments zone\n\n## Traits\n- charges are idempotent\n\n## Verify\n- `true`\n\n## Thoughts\n" + desk + "\n",
		"docs/README.md":              "docs\n",
		"src/pay/api/handler.go":      "package api\n\nfunc Handle() {}\n",
		"src/pay/api/routes.go":       "package api\n\nfunc Routes() {}\n",
		"src/pay/core/ledger.go":      "package core\n\nfunc Post() {}\n",
		"src/pay/core/ledger_test.go": "package core\n\nimport \"testing\"\n\nfunc TestPost(t *testing.T) {}\n",
	}
	for p, s := range extra {
		files[p] = s
	}
	for p, s := range files {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func evoOpen(t *testing.T, dir string, mut func(*EvolutionOptions)) *World {
	t.Helper()
	o := EvolutionOptions{Enabled: true, Mode: "auto", ThoughtsMax: 7, StressAt: 5, FailStreak: 3, CooldownRuns: 10, Arms: []string{"split", "mutate", "rebirth"}, SplitLines: 0, SplitFiles: 100, SkillStatus: "trial"}
	if mut != nil {
		mut(&o)
	}
	w, err := Open(dir, Isekai(), Options{Ontology: true, Evolution: o})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func stress(w *World, body string, n int) {
	w.mu.Lock()
	w.stress[body] = n
	w.mu.Unlock()
}

func ontoClean(t *testing.T, w *World) {
	t.Helper()
	ow, err := onto.LoadLayout(w.Root, w.Ranks.Layout(w.Lex))
	if err != nil {
		t.Fatal(err)
	}
	if fs := ow.Validate(); len(fs) > 0 {
		t.Fatalf("onto check must stay at 0 findings, got %v", fs)
	}
}

func read(t *testing.T, dir, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// eight Thoughts with no repeated lesson
func loose(n int) []string {
	words := []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet"}
	var out []string
	for i := 0; i < n; i++ {
		out = append(out, fmt.Sprintf("- 2026-10-%02d %s", i+1, words[i]))
	}
	return out
}

func TestEvolveTriggerNeedsStressAndAFullDesk(t *testing.T) {
	// seven Thoughts, stressed: the desk is not past thoughtsMax
	w := evoOpen(t, evoWorld(t, loose(7), nil), nil)
	stress(w, "slime-pay", 9)
	if ev, ok := w.EvolveMaybe("slime-pay"); ok {
		t.Fatalf("a desk at thoughtsMax must not evolve: %+v", ev)
	}
	// eight Thoughts, calm: not stressed
	w = evoOpen(t, evoWorld(t, loose(8), nil), nil)
	if ev, ok := w.EvolveMaybe("slime-pay"); ok {
		t.Fatalf("an unstressed body must not evolve: %+v", ev)
	}
	if st := w.StressOf("slime-pay"); st.Stressed {
		t.Fatalf("no reading says stressed: %+v", st)
	}
	// eight Thoughts AND stressed: it evolves
	stress(w, "slime-pay", 5)
	ev, ok := w.EvolveMaybe("slime-pay")
	if !ok || ev.Status != "applied" {
		t.Fatalf("stressed with 8 Thoughts must evolve: %+v ok=%v", ev, ok)
	}
	// with no repetition and no other evidence, the desk is still distilled (Nature 5)
	if ev.Arm != "mutate" || !strings.Contains(ev.Reason, "Nature 5") {
		t.Fatalf("the fallback is the plain distillation: %+v", ev)
	}
	if n := w.DeskOf("slime-pay"); n > 1 {
		t.Fatalf("the desk must be emptied but its dated line, got %d", n)
	}
	log := read(t, w.Root, ".isekai/log.md")
	if !strings.Contains(log, "evolution — slime-pay mutated its genome") || !strings.Contains(log, "wrap-up: alpha") {
		t.Fatalf("the evolution and its wrap-up are logged:\n%s", log)
	}
	ontoClean(t, w)
}

func TestEvolveStressReadings(t *testing.T) {
	w := evoOpen(t, evoWorld(t, loose(8), nil), nil)
	for i := 0; i < 3; i++ {
		w.FoldRun(RunFold{Body: "slime-pay", Loop: "DONE", Word: "FAIL", Gate: "fail"})
	}
	st := w.StressOf("slime-pay")
	if !st.Stressed || st.Streak != 3 {
		t.Fatalf("three failed verdicts in a row read stressed: %+v", st)
	}
	w2 := evoOpen(t, evoWorld(t, loose(8), nil), nil)
	w2.FoldRun(RunFold{Body: "slime-pay", Loop: "CHECKPOINT"})
	if st := w2.StressOf("slime-pay"); !st.Stressed || !st.Doom {
		t.Fatalf("a doom checkpoint reads stressed: %+v", st)
	}
}

func splitThoughts() []string {
	return []string{
		"- 2026-10-01 `src/pay/api/handler.go` returns 500 on a bad currency",
		"- 2026-10-02 `src/pay/api/routes.go` mounts the refund route twice",
		"- 2026-10-03 api/handler.go swallows the timeout",
		"- 2026-10-04 `src/pay/api/handler.go` leaks the idempotency key",
		"- 2026-10-05 `src/pay/core/ledger.go` double posts on retry",
		"- 2026-10-06 core/ledger.go rounds half-down",
		"- 2026-10-07 `src/pay/core/ledger.go` lacks a lock",
		"- 2026-10-08 the zone feels wide",
	}
}

func TestEvolveSplitDividesTheZoneAndTheDesk(t *testing.T) {
	w := evoOpen(t, evoWorld(t, splitThoughts(), nil), func(o *EvolutionOptions) { o.SplitFiles = 3 })
	stress(w, "slime-pay", 6)
	ev, ok := w.EvolveMaybe("slime-pay")
	if !ok || ev.Status != "applied" || ev.Arm != "split" {
		t.Fatalf("a large clustered zone splits: %+v", ev)
	}
	if strings.Join(ev.Into, ",") != "slime-pay-api,slime-pay-core" {
		t.Fatalf("the young are named for their halves: %v", ev.Into)
	}
	if w.Creature("slime-pay") != nil {
		t.Fatal("the parent retires off the roster")
	}
	a, c := w.Creature("slime-pay-api"), w.Creature("slime-pay-core")
	if a == nil || c == nil {
		t.Fatalf("both young stand: %v %v", a, c)
	}
	if own := w.Owner("src/pay/api/handler.go"); own == nil || own.Name != "slime-pay-api" {
		t.Fatalf("api soil rides with slime-pay-api: %+v", own)
	}
	if own := w.Owner("src/pay/core/ledger.go"); own == nil || own.Name != "slime-pay-core" {
		t.Fatalf("core soil rides with slime-pay-core: %+v", own)
	}
	if a.Parent != "orc-pay" || c.Parent != "orc-pay" {
		t.Fatalf("the young report where the parent did: %s %s", a.Parent, c.Parent)
	}
	da, dc := read(t, w.Root, a.Doc), read(t, w.Root, c.Doc)
	if !strings.Contains(da, "routes.go") || strings.Contains(da, "ledger.go") {
		t.Fatalf("the api half carries the api Thoughts only:\n%s", da)
	}
	if !strings.Contains(dc, "lacks a lock") || strings.Contains(dc, "refund route") {
		t.Fatalf("the core half carries the core Thoughts only:\n%s", dc)
	}
	na, nc := len(memory.ThoughtsOf(da)), len(memory.ThoughtsOf(dc))
	if na+nc != 8 {
		t.Fatalf("every Thought goes to one half: %d + %d", na, nc)
	}
	if !strings.Contains(dc, "go test") && !strings.Contains(dc, "`true`") {
		t.Fatalf("the young carry a verify line:\n%s", dc)
	}
	parent := read(t, w.Root, ".isekai/slime/pay/README.md")
	if !docRetired(parent) || !strings.Contains(parent, "split into slime-pay-api + slime-pay-core") {
		t.Fatalf("the parent's doc carries the retired mark:\n%s", parent)
	}
	if _, err := os.Stat(filepath.Join(w.Root, ".isekai", "attic")); err != nil {
		t.Fatal("the parent's genome rests in the attic")
	}
	ontoClean(t, w)
	var row *value.Row
	for _, r := range w.Ledger() {
		if r.Kind == value.KindEvolve {
			r := r
			row = &r
		}
	}
	if row == nil || row.Arm != "split" || row.Status != "applied" || len(row.Into) != 2 {
		t.Fatalf("the ledger notes the evolution: %+v", row)
	}
	if !strings.Contains(read(t, w.Root, ".isekai/log.md"), "slime-pay evolved: split → slime-pay-api, slime-pay-core") {
		t.Fatal("the split is logged")
	}
}

func TestEvolveSplitAnOrcDividesItsSlimes(t *testing.T) {
	dir := evoWorld(t, nil, map[string]string{
		".isekai/orc/pay/README.md":   "# orc-pay\n\n- **Rank:** Orc\n- **Territory:** `src/`\n- **Reports to:** elf-core\n\n## Thoughts\n" + strings.Join([]string{"- 2026-10-01 `src/pay/` slow", "- 2026-10-02 `src/pay/` again", "- 2026-10-03 `src/web/` broke", "- 2026-10-04 `src/web/` css", "- 2026-10-05 `src/pay/` x", "- 2026-10-06 `src/web/` y", "- 2026-10-07 `src/pay/` z", "- 2026-10-08 `src/web/` w"}, "\n") + "\n",
		".isekai/slime/web/README.md": "# slime-web\n\n- **Rank:** Slime\n- **Territory:** `src/web/`\n- **Reports to:** orc-pay\n- **Verify:** `true`\n",
		"src/web/app.go":              "package web\n",
	})
	w := evoOpen(t, dir, func(o *EvolutionOptions) { o.SplitFiles = 3 })
	stress(w, "orc-pay", 6)
	ev, ok := w.EvolveMaybe("orc-pay")
	if !ok || ev.Status != "applied" || ev.Arm != "split" {
		t.Fatalf("an orc splits: %+v", ev)
	}
	if w.Creature("orc-pay-pay") == nil || w.Creature("orc-pay-web") == nil {
		t.Fatalf("two orcs stand: %v", ev.Into)
	}
	if p := w.Creature("slime-pay").Parent; p != "orc-pay-pay" {
		t.Fatalf("slime-pay follows its soil: %s", p)
	}
	if p := w.Creature("slime-web").Parent; p != "orc-pay-web" {
		t.Fatalf("slime-web follows its soil: %s", p)
	}
	ontoClean(t, w)
}

func mutateThoughts() []string {
	return []string{
		"- 2026-10-01 retry: always run `go test ./src/pay/core/` before touching the retry loop",
		"- 2026-10-02 the retry loop needs the jitter first",
		"- 2026-10-03 retry storms: use the backoff helper, never sleep",
		"- 2026-10-04 retry counts must reset after a success",
		"- 2026-10-05 retry: run the ledger test then the api test",
		"- 2026-10-06 retry window comes from config",
		"- 2026-10-07 retry in `src/pay/core/ledger.go` must be idempotent",
		"- 2026-10-08 refunds settle in two days",
	}
}

func TestEvolveMutateDistilsAndMintsTheHow(t *testing.T) {
	w := evoOpen(t, evoWorld(t, mutateThoughts(), nil), func(o *EvolutionOptions) {
		o.Survey = func(p []string) TerritorySurvey {
			sv := TerritorySurvey{Files: 1, Lines: 3}
			if len(p) == 1 && strings.HasPrefix(p[0], "src/pay/core") {
				sv.Verify = []string{"test -f src/pay/core/ledger_test.go"}
			}
			return sv
		}
	})
	stress(w, "slime-pay", 5)
	ev, ok := w.EvolveMaybe("slime-pay")
	if !ok || ev.Status != "applied" || ev.Arm != "mutate" {
		t.Fatalf("a repeated lesson mutates: %+v", ev)
	}
	if !strings.Contains(ev.Reason, `name "retry"`) {
		t.Fatalf("the reason names the lesson: %s", ev.Reason)
	}
	doc := read(t, w.Root, ".isekai/slime/pay/README.md")
	if !strings.Contains(doc, "- retry — ") || !strings.Contains(doc, "from 7 thoughts") {
		t.Fatalf("the lesson is a trait:\n%s", doc)
	}
	if !strings.Contains(doc, "- `test -f src/pay/core/ledger_test.go`") {
		t.Fatalf("a verify line where tests cover the lesson's paths:\n%s", doc)
	}
	if ev.Skill != "slime-pay-retry" || !strings.Contains(doc, "- **Minds:** slime-pay-retry") {
		t.Fatalf("the how is minted and worn: %q\n%s", ev.Skill, doc)
	}
	skill := read(t, w.Root, ".isekai/crafted-skills/slime-pay-retry/SKILL.md")
	if !strings.Contains(skill, "status: trial") || !strings.Contains(skill, "name: slime-pay-retry") || !strings.Contains(skill, "backoff helper") {
		t.Fatalf("the skill file:\n%s", skill)
	}
	ts := memory.ThoughtsOf(doc)
	if len(ts) != 1 || !strings.Contains(ts[0].Text, "distilled into") {
		t.Fatalf("the desk keeps only its dated distilled line: %+v", ts)
	}
	if c := w.Creature("slime-pay"); c == nil || !contains(c.Minds, "slime-pay-retry") || !sameSoil(c.Territory, []string{"src/pay/"}) {
		t.Fatalf("same name, same territory, the new hat worn: %+v", c)
	}
	if !strings.Contains(read(t, w.Root, ".isekai/log.md"), "wrap-up: refunds settle in two days") {
		t.Fatal("the single goes to the log's wrap-up")
	}
	ontoClean(t, w)
}

func TestEvolveMutateDistilsAWornMindsDesk(t *testing.T) {
	mind := "---\nname: slime-pay-wire\ndescription: the wire\n---\n# wire\n\n## Thoughts\n" + strings.Join([]string{"- 2026-10-01 envelope cap trims findings", "- 2026-10-02 envelope cap hides holes", "- 2026-10-03 envelope cap again"}, "\n") + "\n"
	dir := evoWorld(t, loose(5), map[string]string{".isekai/crafted-skills/slime-pay-wire/SKILL.md": mind})
	doc := read(t, dir, ".isekai/slime/pay/README.md")
	_ = os.WriteFile(filepath.Join(dir, ".isekai/slime/pay/README.md"), []byte(strings.Replace(doc, "- **Minds:** none", "- **Minds:** slime-pay-wire", 1)), 0o644)
	w := evoOpen(t, dir, nil)
	if n := w.DeskOf("slime-pay"); n != 8 {
		t.Fatalf("the desk counts the doc's and the worn mind's: %d", n)
	}
	stress(w, "slime-pay", 5)
	if ev, ok := w.EvolveMaybe("slime-pay"); !ok || ev.Status != "applied" {
		t.Fatalf("it evolves: %+v", ev)
	}
	m := read(t, dir, ".isekai/crafted-skills/slime-pay-wire/SKILL.md")
	if ts := memory.ThoughtsOf(m); len(ts) != 1 || !strings.Contains(ts[0].Text, "distilled into slime-pay's genome") {
		t.Fatalf("the mind's desk is distilled too:\n%s", m)
	}
	if !strings.Contains(read(t, dir, ".isekai/slime/pay/README.md"), "envelope") {
		t.Fatal("the mind's lesson lands in the body's traits")
	}
	ontoClean(t, w)
}

func TestEvolveRebirthWritesAFreshGenome(t *testing.T) {
	dir := evoWorld(t, loose(8), nil)
	doc := read(t, dir, ".isekai/slime/pay/README.md")
	_ = os.WriteFile(filepath.Join(dir, ".isekai/slime/pay/README.md"), []byte(strings.Replace(doc, "- charges are idempotent", "- charges post through `src/pay/old/charge.go`", 1)), 0o644)
	w := evoOpen(t, dir, func(o *EvolutionOptions) {
		o.Survey = func(p []string) TerritorySurvey {
			return TerritorySurvey{Files: 4, Lines: 12, Traits: []string{"sketch: Go packages api, core: 3 source files."}, Verify: []string{"test -d src/pay"}}
		}
	})
	for i := 0; i < 4; i++ {
		w.FoldRun(RunFold{Body: "slime-pay", Loop: "DONE", Word: "FAIL", Gate: "fail", VerifyRun: 1, VerifyRed: 1})
	}
	if m := w.Metrics("slime-pay"); m["runs"] != 4 {
		t.Fatalf("the record holds 4 runs: %v", m["runs"])
	}
	ev, ok := w.EvolveMaybe("slime-pay")
	if !ok || ev.Status != "applied" || ev.Arm != "rebirth" {
		t.Fatalf("a stale genome is reborn: %+v", ev)
	}
	if !strings.Contains(ev.Reason, "its gate failed on its own verify lines") || !strings.Contains(ev.Reason, "src/pay/old/charge.go") {
		t.Fatalf("the reason names the evidence: %s", ev.Reason)
	}
	fresh := read(t, w.Root, ".isekai/slime/pay/README.md")
	for _, want := range []string{"- **Reborn:**", "sketch: Go packages api, core", "- `test -d src/pay`", "- **Reports to:** orc-pay", "- **Territory:** `src/pay/`"} {
		if !strings.Contains(fresh, want) {
			t.Fatalf("the fresh doc carries %q:\n%s", want, fresh)
		}
	}
	if strings.Contains(strings.Join(sectionBullets(fresh, "Traits"), "\n"), "src/pay/old/charge.go") {
		t.Fatal("the stale trait does not come back")
	}
	atts, _ := filepath.Glob(filepath.Join(w.Root, ".isekai", "attic", "slime-pay@*.md"))
	if len(atts) != 1 || !strings.Contains(read(t, w.Root, w.Rel(atts[0])), "src/pay/old/charge.go") {
		t.Fatalf("the old genome rests in the attic: %v", atts)
	}
	if m := w.Metrics("slime-pay"); m["runs"] != 0 || m["fails"] != 0 {
		t.Fatalf("the value window starts again: runs %v fails %v", m["runs"], m["fails"])
	}
	reborn := false
	for _, r := range w.Ledger() {
		reborn = reborn || (r.Kind == value.KindRebirth && r.Body == "slime-pay")
	}
	if !reborn {
		t.Fatal("a rebirth row is written")
	}
	ontoClean(t, w)
}

func TestEvolveFailedVerifyKeepsTheOldDoc(t *testing.T) {
	dir := evoWorld(t, loose(8), nil)
	w := evoOpen(t, dir, func(o *EvolutionOptions) {
		o.Survey = func(p []string) TerritorySurvey { return TerritorySurvey{Verify: []string{"exit 3"}} }
	})
	w.FoldRun(RunFold{Body: "slime-pay", Loop: "DONE", Word: "FAIL", Gate: "fail", VerifyRun: 1, VerifyRed: 1})
	stress(w, "slime-pay", 5)
	before := read(t, dir, ".isekai/slime/pay/README.md")
	ev, ok := w.EvolveMaybe("slime-pay")
	if !ok || ev.Status != "hole" || !strings.Contains(ev.Hole, "exit 3") {
		t.Fatalf("a red verify is a hole: %+v", ev)
	}
	if after := read(t, dir, ".isekai/slime/pay/README.md"); after != before {
		t.Fatal("the old doc stands untouched")
	}
	if _, err := os.Stat(filepath.Join(dir, ".isekai", "attic")); err == nil {
		t.Fatal("nothing was swapped in, not even the attic copy")
	}
	if !strings.Contains(read(t, dir, ".isekai/log.md"), "evolution hole — slime-pay") {
		t.Fatal("the hole is logged")
	}
	if cd := w.CooldownOf("slime-pay"); cd.Left == 0 {
		t.Fatal("a failed try starts the cooldown too (no redraw on every landing)")
	}
	ontoClean(t, w)
}

func TestEvolveCooldown(t *testing.T) {
	dir := evoWorld(t, loose(8), nil)
	w := evoOpen(t, dir, func(o *EvolutionOptions) { o.CooldownRuns = 2 })
	stress(w, "slime-pay", 5)
	if _, ok := w.EvolveMaybe("slime-pay"); !ok {
		t.Fatal("first evolution")
	}
	// the desk fills again and the strain returns
	doc := read(t, dir, ".isekai/slime/pay/README.md")
	_ = os.WriteFile(filepath.Join(dir, ".isekai/slime/pay/README.md"), []byte(doc+strings.Join(loose(8), "\n")+"\n"), 0o644)
	stress(w, "slime-pay", 5)
	if ev, ok := w.EvolveMaybe("slime-pay"); ok {
		t.Fatalf("inside the cooldown nothing evolves: %+v", ev)
	}
	w.FoldRun(RunFold{Body: "slime-pay", Loop: "DONE", Word: "FAIL"})
	w.FoldRun(RunFold{Body: "slime-pay", Loop: "DONE", Word: "FAIL"})
	if cd := w.CooldownOf("slime-pay"); cd.Left != 0 || cd.RunsSince != 2 {
		t.Fatalf("two runs later the cooldown is spent: %+v", cd)
	}
	// the cooldown is persisted: a reopened world remembers it
	w2 := evoOpen(t, dir, func(o *EvolutionOptions) { o.CooldownRuns = 3 })
	if cd := w2.CooldownOf("slime-pay"); cd.Last == nil || cd.Left != 1 {
		t.Fatalf("the ledger carries the cooldown across a restart: %+v", cd)
	}
	if ev, ok := w.EvolveMaybe("slime-pay"); !ok || ev.Status != "applied" {
		t.Fatalf("out of the cooldown it evolves again: %+v", ev)
	}
}

func TestEvolveProposeQueuesInsteadOfApplying(t *testing.T) {
	dir := evoWorld(t, mutateThoughts(), nil)
	var heard []EvolveEvent
	w := evoOpen(t, dir, func(o *EvolutionOptions) {
		o.Mode = "propose"
		o.On = func(e EvolveEvent) { heard = append(heard, e) }
	})
	stress(w, "slime-pay", 5)
	before := read(t, dir, ".isekai/slime/pay/README.md")
	ev, ok := w.EvolveMaybe("slime-pay")
	if !ok || ev.Status != "proposed" || ev.ID == "" {
		t.Fatalf("propose mode queues: %+v", ev)
	}
	if read(t, dir, ".isekai/slime/pay/README.md") != before {
		t.Fatal("nothing is applied in propose mode")
	}
	if ev2, ok := w.EvolveMaybe("slime-pay"); ok {
		t.Fatalf("an open proposal is not proposed again: %+v", ev2)
	}
	p, err := w.Proposal(ev.ID[:5])
	if err != nil || p.Status != "open" || !strings.Contains(p.Diff(), "+- retry — ") {
		t.Fatalf("the queue holds the diff: %v\n%s", err, p.Diff())
	}
	got, err := w.ApplyProposal(context.Background(), ev.ID)
	if err != nil || got.Status != "applied" {
		t.Fatalf("apply lands it through the same proof: %v %+v", err, got)
	}
	if !strings.Contains(read(t, dir, ".isekai/slime/pay/README.md"), "- retry — ") {
		t.Fatal("applied")
	}
	if p, _ := w.Proposal(ev.ID); p.Status != "applied" {
		t.Fatalf("the proposal is marked applied: %s", p.Status)
	}
	if len(heard) != 2 || heard[0].Status != "proposed" || heard[1].Status != "applied" {
		t.Fatalf("the surfaces hear both: %+v", heard)
	}
	ontoClean(t, w)
	// a second proposal, rejected
	dir2 := evoWorld(t, loose(8), nil)
	w2 := evoOpen(t, dir2, func(o *EvolutionOptions) { o.Mode = "propose" })
	stress(w2, "slime-pay", 5)
	ev, _ = w2.EvolveMaybe("slime-pay")
	if _, err := w2.RejectProposal(ev.ID, "not now"); err != nil {
		t.Fatal(err)
	}
	if _, err := w2.ApplyProposal(context.Background(), ev.ID); err == nil {
		t.Fatal("a rejected proposal never applies")
	}
}

func TestEvolveProposalRefusesAMovedWorld(t *testing.T) {
	dir := evoWorld(t, loose(8), nil)
	w := evoOpen(t, dir, func(o *EvolutionOptions) { o.Mode = "propose" })
	stress(w, "slime-pay", 5)
	ev, _ := w.EvolveMaybe("slime-pay")
	p := filepath.Join(dir, ".isekai/slime/pay/README.md")
	b, _ := os.ReadFile(p)
	_ = os.WriteFile(p, append(b, []byte("- 2026-10-11 a human wrote here\n")...), 0o644)
	if _, err := w.ApplyProposal(context.Background(), ev.ID); err == nil || !strings.Contains(err.Error(), "changed since the plan") {
		t.Fatalf("a doc that moved under the plan refuses it: %v", err)
	}
}

func TestEvolveOffAndSweep(t *testing.T) {
	dir := evoWorld(t, loose(8), nil)
	w := evoOpen(t, dir, func(o *EvolutionOptions) { o.ThoughtsMax = 0 })
	stress(w, "slime-pay", 9)
	if evs := w.EvolveSweep(); len(evs) != 0 {
		t.Fatalf("thoughtsMax 0 is off: %+v", evs)
	}
	w = evoOpen(t, dir, nil)
	stress(w, "slime-pay", 9)
	evs := w.EvolveSweep()
	if len(evs) != 1 || evs[0].Body != "slime-pay" {
		t.Fatalf("the run-start sweep weighs every body: %+v", evs)
	}
	if lines := w.evolveLines(); len(lines) != 1 || !strings.HasPrefix(lines[0], "[evolution] slime-pay mutated its genome") {
		t.Fatalf("the dispatcher's report carries it: %v", lines)
	}
	rep := w.ValueReport()
	if len(rep.Evolution.Recent) != 1 || rep.Evolution.ThoughtsMax != 7 {
		t.Fatalf("/api/value carries the evolution: %+v", rep.Evolution)
	}
	for _, b := range rep.Bodies {
		if b.Body == "slime-pay" && (b.Evolution == nil || b.Evolution.Last == nil || b.Evolution.CooldownLeft != 10) {
			t.Fatalf("the body's cooldown and last evolution: %+v", b.Evolution)
		}
	}
}

func TestThoughtsOfMatchesTheDeskCount(t *testing.T) {
	doc := "# x\n\n## Thoughts\npreamble\n- 2026-10-01 one\n  more of one\n### [2026-10-02]\nprose of two\n- not dated\n\n## Next\n- 2026-10-03 outside\n"
	ts := memory.ThoughtsOf(doc)
	if len(ts) != 2 || memory.CountThoughts(doc) != 2 {
		t.Fatalf("two dated entries: %+v", ts)
	}
	if !strings.Contains(ts[0].Text, "more of one") || !strings.Contains(ts[1].Text, "- not dated") {
		t.Fatalf("an entry carries its continuation lines: %+v", ts)
	}
	cut := cutDesk(doc, ts, "- 2026-10-04 distilled into 1 trait (evolution)")
	if memory.CountThoughts(cut) != 1 || !strings.Contains(cut, "preamble") || !strings.Contains(cut, "## Next\n- 2026-10-03 outside") {
		t.Fatalf("the cut leaves the rest whole:\n%s", cut)
	}
}
