package world

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/loop"
)

// A lineage refusal names the lawful dispatch path, each hop proven against the dispatch law:
// the nearest creature of the target's line the dispatcher may order itself, else up with @? to
// where the lines meet and down from there.
func TestRouteVerification(t *testing.T) {
	_, w := twoLines(t)
	strict := CourtOptions{StrictFlow: true}
	refuse := func(from, to string) string {
		t.Helper()
		err := w.mayDispatch(from, w.Creature(to), strict)
		if err == nil {
			t.Fatalf("%s → %s went through", from, to)
		}
		return err.Error()
	}
	if got := refuse("orc-a", "slime-b"); !strings.Contains(got, "lawful route: commission orc-b, which orders down (orc-b → slime-b)") {
		t.Errorf("sideways route: %s", got)
	}
	if got := refuse("elf-a", "orc-b"); !strings.Contains(got, "lawful route: carry it up with @? (elf-a → rimuru), then rimuru → elf-b → orc-b") {
		t.Errorf("up-and-down route: %s", got)
	}
}

// A Reports-to hop resolves against the living chain: a hop nothing living holds ends the line,
// is never offered as the one who commissions, and no route is guessed through it.
func TestLineageLivingChain(t *testing.T) {
	dir, w := twoLines(t)
	stray := filepath.Join(dir, ".isekai/slime/stray/README.md")
	os.MkdirAll(filepath.Dir(stray), 0o755)
	os.WriteFile(stray, []byte("# slime-stray\n\n- **Rank:** Slime\n- **Reports to:** orc-ghost\n- **Verify:** `true`\n"), 0o644)
	if err := w.Reload(); err != nil {
		t.Fatal(err)
	}
	if line, broken := w.lineage("slime-stray"); len(line) != 0 || broken != "orc-ghost" {
		t.Fatalf("lineage: %v broken at %q", line, broken)
	}
	err := w.mayDispatch("orc-a", w.Creature("slime-stray"), CourtOptions{})
	if err == nil || !strings.Contains(err.Error(), "its line breaks at orc-ghost") || strings.Contains(err.Error(), "orc-ghost commissions it") {
		t.Fatalf("a dead hop: %v", err)
	}
	if err := w.mayDispatch(Rimuru, w.Creature("slime-stray"), CourtOptions{}); err != nil {
		t.Fatalf("rimuru tops every chain: %v", err)
	}
}

// The lawful-target list is exactly what preflight accepts: a slime with no coded verdict is not
// offered to its gate holder, and a body is named once.
func TestStratTargetsFaithful(t *testing.T) {
	dir, w := twoLines(t)
	bare := filepath.Join(dir, ".isekai/slime/bare/README.md")
	os.MkdirAll(filepath.Dir(bare), 0o755)
	os.WriteFile(bare, []byte("# slime-bare\n\n- **Rank:** Slime\n- **Territory:** `a/`\n- **Reports to:** orc-a\n"), 0o644)
	w.AddForeign(Creature{Name: "slime-a", Rank: "slime"})
	if err := w.Reload(); err != nil {
		t.Fatal(err)
	}
	got := w.stratTargets(&loop.Engine{As: "orc-a"}, CourtOptions{StrictFlow: true})
	if strings.Contains(got, "slime-bare") || strings.Count(got, "slime-a") != 1 {
		t.Fatalf("orc-a's lawful targets: %q", got)
	}
	if err := w.MayOrder("orc-a", "slime-bare", CourtOptions{StrictFlow: true}); err == nil {
		t.Fatal("preflight took an order with no coded verdict")
	}
}
