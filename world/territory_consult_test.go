package world

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/provider/mock"
)

// The 2026-10-09 human law: a slime on foreign soil checks with its orc — the gate
// refusal is the consult trigger, and it must carry the three arms (widen / birth /
// route) plus the owner's name, so the parent decides without a re-derivation drain.
func TestTerritoryRefusalCarriesConsultArms(t *testing.T) {
	// the custom-ranks world: coder-auth (coder) writing docs/new.md — writer-docs
	// (writer) authors it, a DIFFERENT rank, so the default branch names the owner
	_, w := customWorld(t)
	m := mock.New(writeCall("1", "docs/new.md", "x"), mock.Text("@S DONE\n@? docs/new.md needed\n@U colony c\n@E 0"))
	e := w.Engine("coder-auth", build(m))
	r, err := e.Run(context.Background(), "x")
	if err != nil || r.Steps[0].Status != "refused" {
		t.Fatalf("%v %+v", err, r)
	}
	res := m.Requests[1].Messages[2].ToolResults[0]
	for _, want := range []string{"outside coder-auth's territory", "belongs to writer-docs", "one hop up", "keeper-core", "widen", "birth", "route"} {
		if !strings.Contains(res.Content, want) {
			t.Fatalf("the consult refusal must carry %q:\n%s", want, res.Content)
		}
	}
	// unrouted soil: the Genesis line names itself (standard fixture, no author above src/)
	w2 := open(t, fixture(t))
	m2 := mock.New(writeCall("1", "docs/new.md", "x"), mock.Text("@S DONE\n@? docs/new.md needed\n@U colony c\n@E 0"))
	e2 := w2.Engine("slime-auth", build(m2))
	if r2, _ := e2.Run(context.Background(), "x"); r2.Steps[0].Status != "refused" {
		t.Fatalf("unrouted soil: %+v", r2.Steps[0])
	}
	res2 := m2.Requests[1].Messages[2].ToolResults[0]
	for _, want := range []string{"unrouted", "Genesis trigger", "widen", "birth", "route"} {
		if !strings.Contains(res2.Content, want) {
			t.Fatalf("the unrouted refusal must carry %q:\n%s", want, res2.Content)
		}
	}
}

// The dispatch card is the dispatcher's one-read ground truth (the onto tool serves
// it): owner + rank + domain + strain from the ledger shelf + the arms.
func TestDispatchCard(t *testing.T) {
	dir := fixture(t)
	// a strain row rides the ledger shelf — the card reads live state, not re-derivation
	if err := os.MkdirAll(filepath.Join(dir, ".isekai", "instruments"), 0o755); err != nil {
		t.Fatal(err)
	}
	row := `{"kind":"stress","body":"slime-auth","value":3,"at":"2026-10-09T09:00:00Z"}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, ".isekai", "instruments", "ledgers.jsonl"), []byte(row), 0o644); err != nil {
		t.Fatal(err)
	}
	w := open(t, dir)
	card := w.DispatchCard("src/auth/login.go")
	for _, want := range []string{"slime-auth", "slime", "strain 3", "domain orc-security", "doc .isekai/slime/auth/README.md", "widen", "birth", "route"} {
		if !strings.Contains(card, want) {
			t.Fatalf("the dispatch card must carry %q:\n%s", want, card)
		}
	}
	card2 := w.DispatchCard("docs/x.md")
	for _, want := range []string{"unrouted", "Genesis trigger"} {
		if !strings.Contains(card2, want) {
			t.Fatalf("the unrouted card must carry %q:\n%s", want, card2)
		}
	}
}
