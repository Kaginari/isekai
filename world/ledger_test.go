package world

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Kaginari/isekai/loop"
)

// TestLedgersSurviveRestarts covers the shelf law (freeze review 2026-10-07): council
// strain, consult counts, stress-death memos and the court-write notebook are appended to
// instruments/ledgers.jsonl as they mutate, and a fresh Open replays them. A world that
// lost those at every restart relearned nothing.
func TestLedgersSurviveRestarts(t *testing.T) {
	dir := fixture(t)
	w := open(t, dir)
	// strain: a failed court under escalate-at 1 books a warrant ledger row
	w.councilLedger("orc-security", "FAIL", &loop.Result{Status: loop.Fail}, CourtOptions{StressEscalateAt: 1, StressMaxConsults: 3})
	if w.stress["orc-security"] == 0 {
		t.Fatal("the strain did not land in the ledger")
	}
	// death memo + write attribution: noted the same way the real mutations do
	w.mu.Lock()
	w.ledgerNote(ledgerRow{Kind: "death", Body: "slime-auth", Lane: "openai/z-ai/glm-5.3"})
	w.ledgerNote(ledgerRow{Kind: "courtWrite", Body: "slime-auth", Path: ".isekai/slime/auth/README.md", Writer: "slime-auth"})
	w.mu.Unlock()

	// reincarnation: a fresh Open on the same world must carry the scars
	w2 := open(t, dir)
	func() {
		w2.mu.Lock()
		defer w2.mu.Unlock()
		if w2.stress["orc-security"] == 0 || w2.consults["orc-security"] == 0 {
			t.Fatalf("strain/consults did not replay: %+v %+v", w2.stress, w2.consults)
		}
		if w2.stressDeaths["slime-auth"].Lane != "openai/z-ai/glm-5.3" {
			t.Fatalf("the death memo did not replay: %+v", w2.stressDeaths)
		}
		if cw := w2.courtWrites["slime-auth"][normRel(".isekai/slime/auth/README.md")]; cw.Writer != "slime-auth" {
			t.Fatalf("the write attribution did not replay: %+v", w2.courtWrites)
		}
	}()

	// the clean-landing law replays too: strain booked, then cleared, stays cleared
	w2.councilLedger("orc-security", "FAIL", &loop.Result{Status: loop.Fail}, CourtOptions{StressEscalateAt: 1})
	w2.councilLedger("orc-security", "DONE", &loop.Result{Status: loop.Done}, CourtOptions{StressEscalateAt: 1})
	w3 := open(t, dir)
	w3.mu.Lock()
	defer w3.mu.Unlock()
	if _, ok := w3.stress["orc-security"]; ok {
		t.Fatalf("a clean landing must clear the ledger forever: %+v", w3.stress)
	}
}

// TestLedgerTornTailSwallowsNothing: a crash that left half a row at the ledger shelf's end does
// not swallow the next note — it starts on a line of its own and replays.
func TestLedgerTornTailSwallowsNothing(t *testing.T) {
	dir := fixture(t)
	w := open(t, dir)
	p := filepath.Join(w.Dir(), filepath.FromSlash(ledgerShelf))
	os.MkdirAll(filepath.Dir(p), 0o755)
	os.WriteFile(p, []byte(`{"kind":"stress","body":"slime-auth","val`), 0o600)
	w.mu.Lock()
	w.ledgerNote(ledgerRow{Kind: "death", Body: "slime-auth", Lane: "lane-a"})
	w.mu.Unlock()
	if w2 := open(t, dir); w2.stressDeaths["slime-auth"].Lane != "lane-a" {
		t.Fatalf("the note after a torn tail replays: %+v", w2.stressDeaths)
	}
}
