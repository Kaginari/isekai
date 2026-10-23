package app

import (
	"strings"
	"testing"

	"github.com/Kaginari/isekai/queue"
)

// TestRelayHoldWaitsForTheWake: a background report's @$ orders are held on the shared queue —
// nothing dispatches them, the in-memory book stays empty — until the wake claims them, once,
// each in its body's name; the body's DONE report closes its claim.
func TestRelayHoldWaitsForTheWake(t *testing.T) {
	_, a := forgeWorld(t)
	defer a.Close()
	a.Court.Finish("orc-security", "@S DONE\n@$ slime-auth: tighten the ttl\n@$ slime-api: page the routes\n@E 0")
	q := a.World.Queue()
	held, _ := q.Held()
	if len(held) != 2 || held[0].For != "slime-auth" || held[0].From != "orc-security" || a.heldCount() != 2 {
		t.Fatalf("the orders are held on the shared queue: %+v", held)
	}
	if book := a.Court.OrderBook(); len(book) != 0 {
		t.Fatalf("a held order also sat in the memory book: %+v", book)
	}
	// another session's wake took slime-api's first: this wake leaves it alone
	if _, err := queue.Open(q.Path(), nil).Claim("q2", "slime-api"); err != nil {
		t.Fatal(err)
	}
	lines := a.heldWakeLines()
	if len(lines) != 1 || !strings.HasPrefix(lines[0], "dispatch slime-auth: tighten the ttl\n(queue q1, claimed for slime-auth") {
		t.Fatalf("the wake's prompts: %q", lines)
	}
	if again := a.heldWakeLines(); len(again) != 0 || a.heldCount() != 0 {
		t.Fatalf("a row claimed twice: %q", again)
	}
	a.settleHeld("slime-api", "court slime-api started in the background")
	a.settleHeld("slime-api", "@S FAIL — tests red\n@E 0")
	a.settleHeld("slime-auth", "@S DONE\n@E 0")
	one, _, _ := q.Get("q1")
	two, _, _ := q.Get("q2")
	if one.State != queue.Done || two.State != queue.Claimed || !strings.Contains(two.Note, "reported FAIL") {
		t.Fatalf("settled: %+v / %+v", one, two)
	}
}

// TestRelayHoldWithoutAWorld: a court with no world keeps its orders in the memory book, and the
// wake still drains them.
func TestRelayHoldWithoutAWorld(t *testing.T) {
	a := &App{}
	a.Court = newCourt(a)
	a.Court.Finish("orc-x", "@S DONE\n@$ slime-y: count the retries\n@E 0")
	if lines := a.heldWakeLines(); len(lines) != 1 || lines[0] != "dispatch slime-y: count the retries" {
		t.Fatalf("%q", lines)
	}
}
