package world

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/loop"
	"github.com/Kaginari/isekai/tool"
)

// Hard law (Veldora decree 2026-09-30): rimuru writing another creature's soil is refused at
// the policy level with a dispatch redirection; only the owner's own README, the throne seat
// files, and no-creature soil stay lawful.
func TestPolicyThroneCrossTerritoryRefused(t *testing.T) {
	dir := fixture(t)
	w := open(t, dir)
	pol := w.Policy(Rimuru, TerritoryOptions{Enabled: true})
	if pol == nil {
		t.Fatal("policy nil")
	}

	call := func(paths ...string) error {
		abs := make([]string, 0, len(paths))
		for _, p := range paths {
			abs = append(abs, w.Root+"/"+p)
		}
		return pol(tool.Access{Tool: "write", Class: tool.Write, Paths: abs})
	}

	// 1. refused on slime-owned soil
	if err := call("src/auth/login.go"); err == nil || !strings.Contains(err.Error(), "dispatch") {
		t.Errorf("rimuru on slime soil must refuse with a dispatch redirection: %v", err)
	}
	// 2. allowed on the owner's own doc — that's the pairing lever
	if err := call(".isekai/slime/auth/README.md"); err != nil {
		t.Errorf("rimuru writing the slime's own README is lawful (Vitality pairing): %v", err)
	}
	// 3. allowed on the throne seat files
	for _, seat := range []string{".isekai/dimension.yaml", ".isekai/isekai.md"} {
		if err := call(seat); err != nil {
			t.Errorf("rimuru on throne seat %s is the throne's own work: %v", seat, err)
		}
	}
	// 4. allowed on no-creature soil
	if err := call("notes/freeform.txt"); err != nil {
		t.Errorf("rimuru on no-creature soil is lawful: %v", err)
	}
	// 5. rank-less name (a mount under rimuru but not the literal string) also held
	pol2 := w.Policy("", TerritoryOptions{Enabled: true})
	if err := pol2(tool.Access{Tool: "write", Class: tool.Write, Paths: []string{w.Root + "/src/auth/login.go"}}); err == nil {
		t.Error("rank-less mount under the throne must be held to the hard-law refusal too")
	}
}

// Orc-TDD as code (Law 3 + Veldora decree 2026-09-30): a gate-holding rank dispatching an
// authoring slime whose doc carries no Verify line is refused at dispatch time.
func TestDispatchVerifyPrecondition(t *testing.T) {
	// slime-no-verify fixture: an authoring slime whose README carries no Verify lines
	dir := t.TempDir()
	writeDocHL(t, dir, ".isekai/isekai.md", lawText)
	writeDocHL(t, dir, ".isekai/log.md", "# Chronicle\n\nAppend-only.\n\n---\n")
	writeDocHL(t, dir, ".isekai/orc/core/README.md", "# orc-core\n\n- **Rank:** Orc\n- **Territory:** `world/`\n- **Verify:** `true`\n")
	writeDocHL(t, dir, ".isekai/slime/dirty/README.md", "# slime-dirty\n\n- **Rank:** Slime\n- **Territory:** `world/dirty/`\n- **Reports to:** orc-core\n")

	w, err := Open(dir, Isekai(), Options{Ontology: true})
	if err != nil {
		t.Fatal(err)
	}

	// Locate the two creatures
	orc := w.Creature("orc-core")
	slime := w.Creature("slime-dirty")
	if orc == nil || slime == nil {
		t.Fatalf("fixture: orc=%v slime=%v", orc != nil, slime != nil)
	}
	if slime.VerifyCoded() {
		t.Fatal("slime-dirty must start verify-less for this test")
	}

	// Build a parent engine belonging to orc-core. The dispatcher reads only parent.As and
	// reaches for session/budget only when non-nil — w.Engine wires the lexicon and hooks.
	parentBuild := Build{Tools: tool.Builtins()}
	pe := w.Engine("orc-core", parentBuild)
	_ = loop.Engine{}

	disp := w.Dispatcher(pe, parentBuild)
	_, err = disp(context.Background(), tool.Env{Root: dir}, tool.Commission{Body: "slime-dirty", Ask: "implement X"})
	if err == nil {
		t.Fatal("dispatch without slime verify must be refused")
	}
	if !strings.Contains(err.Error(), "no Verify line") || !strings.Contains(err.Error(), "TDD") {
		t.Errorf("refusal must name the missing verify line + TDD rule: %v", err)
	}

	// With the verify line coded, the dispatch is allowed (gate still may fail on the slime's
	// writing, but the *dispatch gate* must not answer with the verify-precondition error).
	writeDocHL(t, dir, slime.Doc, "# slime-dirty\n\n- **Rank:** Slime\n- **Territory:** `world/dirty/`\n- **Reports to:** orc-core\n\n## Verify\n- `test -d world/dirty`\n")
	// reload so the roster re-derives
	if err := w.Reload(); err != nil {
		t.Fatal(err)
	}
	slime = w.Creature("slime-dirty")
	if !slime.VerifyCoded() {
		t.Fatal("slime-dirty must now read VerifyCoded == true")
	}
	// The dispatch now proceeds; we don't care about outcome — we assert the refusal error text
	// is no longer the verify-precondition one. A run failure (budget exhausted etc.) is fine
	// and out of this test's scope; we only check the dispatch gate doesn't answer with TDD.
	_, err2 := disp(context.Background(), tool.Env{Root: dir}, tool.Commission{Body: "slime-dirty", Ask: "implement X"})
	if err2 != nil && strings.Contains(err2.Error(), "no Verify line") && strings.Contains(err2.Error(), "TDD") {
		t.Errorf("verify-precondition must not fire once the slime is coded: %v", err2)
	}
}

func writeDocHL(t *testing.T, dir, rel, body string) {
	t.Helper()
	full := dir + "/" + rel
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestPolicyThroneNeverAuthors — Law 7 decree 2026-10-01 flag-gated via TerritoryOptions.ThroneNeverAuthors.
// When enabled, the throne authors nothing but the law; pairing exemption for owner's README dies under Law 7.
func TestPolicyThroneNeverAuthors(t *testing.T) {
	dir := fixture(t)
	w := open(t, dir)
	// helper call mimics the existing pattern
	makeCall := func(opt TerritoryOptions) func(paths ...string) error {
		pol := w.Policy(Rimuru, opt)
		if pol == nil {
			return func(paths ...string) error { return nil }
		}
		return func(paths ...string) error {
			abs := make([]string, 0, len(paths))
			for _, p := range paths {
				abs = append(abs, w.Root+"/"+p)
			}
			return pol(tool.Access{Tool: "write", Class: tool.Write, Paths: abs})
		}
	}

	// 1. Flag on: throneNeverAuthors true
	callOn := makeCall(TerritoryOptions{Enabled: true, ThroneNeverAuthors: true})
	// a) freeform no-creature soil must REFUSE with Law 7 + dispatch
	if err := callOn("notes/freeform.txt"); err == nil || !strings.Contains(err.Error(), "Law 7") || !strings.Contains(err.Error(), "dispatch") {
		t.Errorf("ThroneNeverAuthors on: notes/freeform.txt must refuse with Law 7 + dispatch, got %v", err)
	}
	// b) law file passes
	if err := callOn(".isekai/isekai.md"); err != nil {
		t.Errorf("ThroneNeverAuthors on: .isekai/isekai.md must pass, got %v", err)
	}
	// c) owner's README is now refused (pairing exemption dies)
	if err := callOn(".isekai/slime/auth/README.md"); err == nil || !strings.Contains(err.Error(), "Law 7") {
		t.Errorf("ThroneNeverAuthors on: slime README must refuse with Law 7, got %v", err)
	}

	// 2. Flag off (zero value): legacy behavior — no-creature soil passes
	callOff := makeCall(TerritoryOptions{Enabled: true})
	if err := callOff("notes/freeform.txt"); err != nil {
		t.Errorf("ThroneNeverAuthors off: notes/freeform.txt must pass (legacy), got %v", err)
	}
}
