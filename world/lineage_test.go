package world

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/tool"
)

// twoLines is a world with two lanes and a kijin: elf-a → orc-a → slime-a, elf-b → orc-b →
// slime-b, kijin-rel under rimuru. Every slime's verify is coded, so only the dispatch law speaks.
func twoLines(t *testing.T) (string, *World) {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		".isekai/isekai.md":           lawText,
		".isekai/log.md":              "# Chronicle\n\nAppend-only.\n\n---\n",
		".isekai/elf/a/README.md":     "# elf-a\n\n- **Rank:** Elf\n- **Reports to:** rimuru\n",
		".isekai/elf/b/README.md":     "# elf-b\n\n- **Rank:** Elf\n- **Reports to:** rimuru\n",
		".isekai/orc/a/README.md":     "# orc-a\n\n- **Rank:** Orc\n- **Territory:** `a/`\n- **Reports to:** elf-a\n",
		".isekai/orc/b/README.md":     "# orc-b\n\n- **Rank:** Orc\n- **Territory:** `b/`\n- **Reports to:** elf-b\n",
		".isekai/slime/a/README.md":   "# slime-a\n\n- **Rank:** Slime\n- **Territory:** `a/`\n- **Reports to:** orc-a\n- **Verify:** `test -d a`\n",
		".isekai/slime/b/README.md":   "# slime-b\n\n- **Rank:** Slime\n- **Territory:** `b/`\n- **Reports to:** orc-b\n- **Verify:** `test -d b`\n",
		".isekai/kijin/rel/README.md": "# kijin-rel\n\n- **Rank:** Kijin\n- **Territory:** `.github/`\n- **Reports to:** rimuru\n- **Verify:** `test -d .github`\n",
		"a/a.go":                      "package a\n",
		"b/b.go":                      "package b\n",
		".github/ci.yml":              "on: push\n",
	}
	for p, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir, open(t, dir)
}

// The dispatch law is creature-level: an orc orders its own slimes, an elf its own orcs; a
// commission to another line's creature is refused naming the parent that commissions it.
func TestDispatchFollowsTheLine(t *testing.T) {
	_, w := twoLines(t)
	strict := CourtOptions{StrictFlow: true}
	may := func(from, to string, opt CourtOptions) error {
		t.Helper()
		c := w.Creature(to)
		if c == nil {
			t.Fatalf("no %s on the roster", to)
		}
		return w.mayDispatch(from, c, opt)
	}
	for _, ok := range [][2]string{{"orc-a", "slime-a"}, {"orc-b", "slime-b"}, {"elf-a", "orc-a"}, {"elf-b", "orc-b"}, {Rimuru, "elf-a"}, {Rimuru, "kijin-rel"}, {"orc-a", "orc-b"}} {
		if err := may(ok[0], ok[1], strict); err != nil {
			t.Errorf("%s → %s is its own line (or a sideways neighbour): %v", ok[0], ok[1], err)
		}
	}
	err := may("orc-a", "slime-b", strict)
	if err == nil || !strings.Contains(err.Error(), "not in your line") || !strings.Contains(err.Error(), "orc-b commissions it") {
		t.Fatalf("orc-a → slime-b must be refused naming orc-b: %v", err)
	}
	if err := may("elf-a", "orc-b", strict); err == nil || !strings.Contains(err.Error(), "elf-b commissions it") {
		t.Fatalf("elf-a → orc-b must be refused naming elf-b: %v", err)
	}
	// without strict flow an elf may reach down its own line two rungs, never into another's
	open := CourtOptions{}
	if err := may("elf-a", "slime-a", open); err != nil {
		t.Errorf("elf-a → slime-a is its own line (strict flow off): %v", err)
	}
	if err := may("elf-a", "slime-b", open); err == nil {
		t.Error("elf-a → slime-b crosses lines even with strict flow off")
	}
	// strict flow still holds one rung at a time
	if err := may("elf-a", "slime-a", strict); err == nil || !strings.Contains(err.Error(), "strict layer flow") {
		t.Errorf("elf-a → slime-a skips a rung under strict flow: %v", err)
	}
	// the escape (tools.dispatch.lineage: false): the rank law alone decides
	if err := may("orc-a", "slime-b", CourtOptions{StrictFlow: true, OpenLineage: true}); err != nil {
		t.Errorf("lineage off: any orc may order any slime: %v", err)
	}
	// the refusal list a dispatcher reads names only its own line
	pe := w.Engine("orc-a", Build{Tools: tool.Builtins()})
	if got := w.stratTargets(pe, strict); !strings.Contains(got, "slime-a") || strings.Contains(got, "slime-b") {
		t.Errorf("orc-a's lawful targets: %q", got)
	}
	// and the dispatcher itself refuses with the hint
	b := Build{Tools: tool.Builtins(), Court: strict}
	disp := w.Dispatcher(w.Engine("orc-a", b), b)
	if _, err := disp(context.Background(), tool.Env{Root: w.Root}, tool.Commission{Body: "slime-b", Ask: "do it"}); err == nil || !strings.Contains(err.Error(), "orc-b commissions it") {
		t.Errorf("the dispatcher must refuse another line's slime: %v", err)
	}
}

// The kijin reports to rimuru and nothing reports to it: it is never handed dispatch, whatever
// its tools list says; an orc, which commands slimes, is.
func TestKijinHoldsNoDispatch(t *testing.T) {
	_, w := twoLines(t)
	b := Build{Tools: tool.Builtins(), Court: CourtOptions{Enabled: true, MaxDepth: 3}, Depth: 1}
	has := func(as string) bool {
		e := w.Engine(as, b)
		defer w.Close(e)
		_, ok := e.Tools.Get("dispatch")
		return ok
	}
	if has("kijin-rel") {
		t.Error("the kijin commands no one: no dispatch on its shelf")
	}
	if !has("orc-a") {
		t.Error("an orc commands its slimes: dispatch on its shelf")
	}
	if Ranks(DefaultRanks(Isekai())).Commands("kijin") || !Ranks(DefaultRanks(Isekai())).Commands("high-orc") {
		t.Error("Commands: kijin none; a high orc commands its base's slimes")
	}
}

// The TDD precondition reads the slime's doc fresh: a verify line the orc wrote earlier in the
// same turn (no reload since) is seen at the check.
func TestDispatchVerifyReadsTheDocFresh(t *testing.T) {
	dir, w := twoLines(t)
	bare := "# slime-a\n\n- **Rank:** Slime\n- **Territory:** `a/`\n- **Reports to:** orc-a\n"
	if err := os.WriteFile(filepath.Join(dir, ".isekai/slime/a/README.md"), []byte(bare), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := w.Reload(); err != nil {
		t.Fatal(err)
	}
	b := Build{Tools: tool.Builtins(), Court: CourtOptions{StrictFlow: true}}
	disp := w.Dispatcher(w.Engine("orc-a", b), b)
	if _, err := disp(context.Background(), tool.Env{Root: dir}, tool.Commission{Body: "slime-a", Ask: "x"}); err == nil || !strings.Contains(err.Error(), "no Verify line") {
		t.Fatalf("an uncoded slime is refused: %v", err)
	}
	// the orc codes the verify this turn — the roster is not reloaded
	if err := os.WriteFile(filepath.Join(dir, ".isekai/slime/a/README.md"), []byte(bare+"- **Verify:** `test -d a`\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if c := w.Creature("slime-a"); c.VerifyCoded() {
		t.Fatal("fixture: the roster's copy must still be stale")
	}
	if _, err := disp(context.Background(), tool.Env{Root: dir}, tool.Commission{Body: "slime-a", Ask: "x"}); err != nil && strings.Contains(err.Error(), "no Verify line") {
		t.Fatalf("a verify written this turn must count: %v", err)
	}
}

// A rank's job reaches the body's system prompt, in the distribution's word for the rank.
func TestPromptCarriesTheRanksJob(t *testing.T) {
	_, w := twoLines(t)
	p := w.Prompt("orc-a", PromptOptions{Creature: true}, nil)
	orc, _ := w.Ranks.Get("orc")
	if !strings.Contains(p, "Your rank's job (Orc): "+strings.TrimSuffix(orc.Job, ".")) {
		t.Fatalf("the orc's job must ride its card:\n%s", p)
	}
}
