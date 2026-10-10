package world

// Cross-session end-gate attribution commissioning (sealed 2026-10-03, per-rule6
// attribution gate, adapted from body 1's design header). Prior session wrote a
// slime's doc pair; next session, a FRESH engine (new shell REPL in the same
// World process — new Engine.Store / runtime.Mem) re-consults the same throne.
// The sealed text requires the consult to hit a WORLD-RESIDENT memory keyed by
// creature name so consults pair cross-session. Redis-rollback covers the
// ignored-soil kit; the world lane lives in World memory, so it survives
// Close(engine) and is cleared only by a NewWorld.
//
// Recording is driven through w.courtGated — the exact block the after-hook
// fires (court.go after-hook, gated:429); driving the full dispatch loop would
// need a live LLM provider, which the fixture does not carry. Consult is
// exercised through w.courtGatedTask (court.go:640), the per-owner-doc consult
// half that ownWrites consults. Engine "slime-x" is a valid Court per
// lawCheck, so the hook path records under the creature name.

import (
	"context"
	"fmt"
	"testing"

	"github.com/Kaginari/isekai/loop"
	"github.com/Kaginari/isekai/wire"

	"os"
	"path/filepath"
)

// TestGateCrossSessionCourtPairingSurvivesClose proves arm 2's two halves:
// lane B written at courtGated survives Close(engine), and a consult through
// a FRESH engine in the same World pairs via the creature name carried in
// w.sessions — no cross-creature bleed, bound respected.
func TestGateCrossSessionCourtPairingSurvivesClose(t *testing.T) {
	w := open(t, fixture(t)) // fixture has no .git — the ignored-soil case

	// (a) Drive the REAL recording block the after-hook calls, then Close.
	slimeE := w.Engine("slime-auth", Build{})
	w.courtGated(slimeE, "slime-auth", []string{".isekai/slime/auth/README.md"})
	w.Close(slimeE)

	rel := normRel(".isekai/slime/auth/README.md")
	if got := w.courtWrites["slime-auth"][rel].Writer; got == "" {
		t.Fatalf("lane B lost after Close: courtWrites[slime-auth][%q] empty", rel)
	}

	// (b) Court-side consult through a FRESH engine pairs via s.Engine.As.
	fresh := w.Engine("slime-auth", Build{})
	if fresh == slimeE {
		t.Fatalf("fresh engine aliases the closed one — not a fresh consult")
	}
	// w.sessions only fills via the hook track() on a live turn (needs a real
	// provider); courtGatedTask consults s.Engine directly, so a minimal
	// session shell over the fresh engine exercises the consult faithfully.
	freshSess := &loop.Session{Engine: fresh}
	if !w.courtGatedTask(freshSess, ".isekai/slime/auth/README.md") {
		t.Fatalf("fresh-engine consult did not pair cross-session for slime-auth doc")
	}

	// (c) A DIFFERENT creature's doc is present in NO lane (no bleed).
	if w.courtGatedTask(freshSess, ".isekai/slime/api/README.md") {
		t.Fatalf("bleed: slime-api doc paired for slime-auth consult")
	}
	if got := w.courtWrites["slime-api"][normRel(".isekai/slime/api/README.md")].Writer; got != "" {
		t.Fatalf("bleed in world lane: slime-api has Writer %q", got)
	}
}

// TestGateCrossSessionCourtWritesBound proves the world lane caps at
// courtWritesBound=8 per creature with the oldest evicted first.
func TestGateCrossSessionCourtWritesBound(t *testing.T) {
	w := open(t, fixture(t))
	slimeE := w.Engine("slime-auth", Build{})

	for i := 0; i < courtWritesBound+3; i++ {
		p := fmt.Sprintf(".isekai/slime/auth/doc-%02d.md", i)
		w.courtGated(slimeE, "slime-auth", []string{p})
	}
	m := w.courtWrites["slime-auth"]
	if got := len(m); got > courtWritesBound {
		t.Fatalf("courtWrites over bound: got %d, bound %d", got, courtWritesBound)
	}
	if got := m[normRel(".isekai/slime/auth/doc-00.md")].Writer; got != "" {
		t.Fatalf("oldest doc-00 not evicted; Writer=%q", got)
	}
	last := fmt.Sprintf(".isekai/slime/auth/doc-%02d.md", courtWritesBound+2)
	if got := m[normRel(last)].Writer; got == "" {
		t.Fatalf("newest %q missing after bound evict", last)
	}
}

// TestGateArm2CrosssessionConsultThroughGate exercises the w.Gate end-gate
// level (deadlock close-out, slime law-pair lane): the prior session sealed a
// doc court record via the real recording block arm A rides, Close dropped the
// engine-keyed ledger, and a FRESH session pairs the write through courtPairedDoc
// arm 2 — the world-resident lane keyed by creature name. Negative control: a
// doc with no court record anywhere fails arm 2's `rec.Writer == ""` engine-As
// short-circuit, so the same Gate shape FAILs Vitality. gitToWorld sets a doc
// baseline; the test bodies untracked write-dir rewrites drift from it.
func TestGateArm2CrosssessionConsultThroughGate(t *testing.T) {
	dir := initBaselineWorld(t)
	writeFixtureFile(t, dir, ".isekai/slime/auth/README.md", "# auth\n")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-m", "auth doc baseline")
	w := open(t, dir)

	// Arm A: prior session commissions slime-auth's doc, then Closes.
	slimeE := w.Engine("slime-auth", Build{})
	w.courtGated(slimeE, "slime-auth", []string{".isekai/slime/auth/README.md"})
	w.Close(slimeE)

	// Arm B: FRESH engine embedded in a session shell — Gate consults
	// s.Engine.As directly, so track() is not needed for arm 2 to find lane B.
	fresh := w.Engine("slime-auth", Build{})
	sess := &loop.Session{Engine: fresh}
	writeFixtureFile(t, dir, ".isekai/slime/auth/README.md", "# auth\ntouch\n")
	v := w.Gate(context.Background(), sess, GateOptions{Enabled: true, DocTruthful: true}, Rimuru, []string{".isekai/slime/auth/README.md"}, &wire.Report{}, false, "reshoot")
	if v.Word != "pass" {
		t.Fatalf("arm-2 cross-session consult at w.Gate level must PASS, got %q reasons=%v", v.Word, v.Reasons)
	}

	// Negative control: same drifted state, doc with no court record in any
	// lane, and no live session for the owner — every arm falls through.
	// RightAuthor: true is where the discrimination lives: the creature-doc arm
	// fails a throne-authored re-authoring with no pairing evidence. DocTruthful
	// by design only pairs NON-doc writes (its own-doc skip at the Vitality lane
	// exempts owner docs), so it is on here only to keep the negative control's
	// options faithful to the real end-gate shape.
	neg := initBaselineWorld(t)
	writeFixtureFile(t, neg, ".isekai/slime/api/README.md", "# api\n")
	git(t, neg, "add", "-A")
	git(t, neg, "commit", "-m", "api doc baseline")
	wn := open(t, neg)
	writeFixtureFile(t, neg, ".isekai/slime/api/README.md", "# api\ntouch\n")
	// The live gate resolves as via s.Engine.As (gate.go:260), so the negative
	// control's session carries the throne creed — the scene it pins is
	// throne-authored re-authoring with no pairing evidence.
	n := wn.Gate(context.Background(), &loop.Session{Engine: wn.Engine(Rimuru, Build{})}, GateOptions{Enabled: true, RightAuthor: true, DocTruthful: true}, Rimuru, []string{".isekai/slime/api/README.md"}, &wire.Report{}, false, "reshoot")
	if n.Word == "pass" {
		t.Fatalf("no-record doc passed pairing — arm-2 consult did not discriminate")
	}
}

// initBaselineWorld builds a real git work-tree from the census fixture so the
// Vitality lane has a HEAD baseline to diff against (gitToWorld lives in the
// dimension tests, not this file). The doc baseline commit follows in-test.
func initBaselineWorld(t *testing.T) string {
	t.Helper()
	root := fixture(t)
	git(t, root, "init")
	git(t, root, "config", "user.email", "gate-test@isekai.local")
	git(t, root, "config", "user.name", "gate-test")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-m", "fixture baseline")
	return root
}

// writeFixtureFile writes a file inside the fixture world.
func writeFixtureFile(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
