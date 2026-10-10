package app

import (
	"testing"

	"github.com/Kaginari/isekai/wire"
)

// TestOrderBookDrains: a body's queued @$ orders come back flat and drain; a body that already
// ran and finished contributes its report so the next dispatch keeps the thread.
func TestOrderBookDrains(t *testing.T) {
	a := &App{}
	c := newCourt(a)

	// elf-think finishes with a report; rimuru's book orders an orc follow-up.
	c.Finish("elf-think", "@S DONE\n@F loop.go:42 the gate fired twice\n@E 9")
	c.Finish("rimuru", "@S DONE\n@$ orc-auth: now gate the retry path\n@$ slime-ledger: count the retries\n@? a note with no order\n@E 40")

	got := c.OrderBook()
	if len(got) != 2 {
		t.Fatalf("order book: %v", got)
	}
	if got[0].Body != "orc-auth" || got[0].Ask != "now gate the retry path" {
		t.Fatalf("first order: %+v", got[0])
	}
	if got[1].Body != "slime-ledger" || got[1].Ask != "count the retries" {
		t.Fatalf("second order: %+v", got[1])
	}
	// a drained book stays drained.
	if again := c.OrderBook(); len(again) != 0 {
		t.Fatalf("drained the book twice: %v", again)
	}

	// the finished body's report rides as thread when it exists *before* the order is drained.
	c.Finish("orc-auth", "@S DONE\n@F loop.go:88 retry gated\n@E 8")
	c.Finish("rimuru", "@S DONE\n@$ orc-auth: gate the abort path too\n@$ elf-plan: plan the next layer\n@E 30")
	got = c.OrderBook()
	if len(got) != 2 {
		t.Fatalf("second drain: %v", got)
	}
	byBody := map[string]string{}
	for _, o := range got {
		byBody[o.Body] = o.Report
	}
	if byBody["orc-auth"] == "" || byBody["orc-auth"][:2] != "@S" {
		t.Fatalf("the finished Court's report rides as thread: %+v", got)
	}
	if byBody["elf-plan"] != "" {
		t.Fatalf("a fresh order carries no thread: %+v", got)
	}
}

// TestDispatchOrdersPrompt: the prompt a drained order becomes names the dispatch and carries
// any thread.
func TestDispatchOrdersPrompt(t *testing.T) {
	a := &App{}
	got := a.dispatchOrders([]struct{ Body, Ask, Report string }{
		{Body: "orc-a", Ask: "gate the retry path", Report: "@S DONE\n@F loop.go:88 fired\n"},
		{Body: "slime-b", Ask: "count holes"},
	})
	if len(got) != 2 {
		t.Fatalf("dispatchOrders: %v", got)
	}
	if got[0] != "dispatch orc-a: gate the retry path\n\n(the same body's earlier Court reported — the thread so far:)\n@S DONE\n@F loop.go:88 fired\n" {
		t.Fatalf("with thread: %q", got[0])
	}
	if got[1] != "dispatch slime-b: count holes" {
		t.Fatalf("without thread: %q", got[1])
	}
}

// silence the wire import — the test leans on the wire's Parse via Finish.
var _ = wire.Enqueue{}
