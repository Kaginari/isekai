package app

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Kaginari/isekai/instrument"
	"github.com/Kaginari/isekai/loop"
	"github.com/Kaginari/isekai/wire"
)

// LiveBody is one body the court can see: rank, office, model, state, elapsed, context, spend.
type LiveBody struct {
	Name    string
	Rank    string
	Office  string
	Model   string
	State   string // thinking · tool · waiting on gate · done · queued
	Exit    string // the run journal's t:end verdict (DONE · FAIL · ESCALATE · DENIED · INTERRUPTED), lifted off the wire report at Finish — the court card's face
	Failed  bool   // Exit names a failed verdict
	Started time.Time
	Ended   time.Time
	Session *loop.Session
	Depth   int
	// two honest lanes (freeze review 2026-10-07): Send lines ride the running turn's next
	// tool batch; Enqueue lines sit until the turn drains. Before, both shared one inbox —
	// Enqueue's "never mid-turn" contract was a lie under loop/loop.go's Inbox hook.
	mid    []string
	next   []string
	report string
	orders []wire.Enqueue // @$ lines it asked for; drained as dispatches at the turn boundary
	done   chan struct{}
}

// Court is the live court (binary.md §Live session): every body that runs or ran in this
// session, its state, its inbox for /send, and the reports of asynchronous Courts waiting to
// wake their dispatcher.
type Court struct {
	app    *App
	mu     sync.Mutex
	bodies map[string]*LiveBody
	order  []string
	wake   []string // reports of finished background Courts, for the session's next turn
	onWake func()
	// onFinish sees a background Court's report as it lands (the TUI draws its block).
	onFinish func(name, report string)
	// relays holds the @$ orders a dispatched body's commission declared it must return in its
	// report. Expectations register at dispatch-observe time (the ask carried @$ lines), clear
	// when the named body is tracked (the relay booked), and close at Finish with either the
	// echoed order or an unbooked-relay wake for the throne.
	relays map[string][]wire.Enqueue
}

func newCourt(a *App) *Court {
	c := &Court{app: a, bodies: map[string]*LiveBody{}}
	prev := a.OnState
	a.OnState = func(body, state string) {
		c.setState(body, state)
		if prev != nil {
			prev(body, state)
		}
	}
	return c
}

// Track registers a body's session as it starts.
func (c *Court) Track(name, rank, office, model string, s *loop.Session, depth int) *LiveBody {
	c.mu.Lock()
	defer c.mu.Unlock()
	b := c.bodies[name]
	if b == nil {
		b = &LiveBody{Name: name, done: make(chan struct{})}
		c.bodies[name] = b
		c.order = append(c.order, name)
	}
	b.Rank, b.Office, b.Model, b.Session, b.Depth = rank, office, model, s, depth
	b.Started, b.Ended, b.State = time.Now(), time.Time{}, "queued"
	if b.done == nil {
		b.done = make(chan struct{})
	}
	// the relay booked: a body named in someone's outstanding relay expectation is now running
	for holder, ords := range c.relays {
		keep := ords[:0]
		for _, o := range ords {
			if o.Body != name {
				keep = append(keep, o)
			}
		}
		if len(keep) == 0 {
			delete(c.relays, holder)
		} else {
			c.relays[holder] = keep
		}
	}
	return b
}

// ExpectRelay records the orders a commission's @$ lines ask a dispatched body to echo back in
// its report. A relay not carried by the report and not booked meanwhile surfaces as a wake at
// the body's Finish — the throne hears the chain landed or that it owes the booking.
func (c *Court) ExpectRelay(body string, orders []wire.Enqueue) {
	if len(orders) == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.relays == nil {
		c.relays = map[string][]wire.Enqueue{}
	}
	c.relays[strings.ToLower(body)] = append(c.relays[strings.ToLower(body)], orders...)
}

func (c *Court) setState(body, state string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	b := c.bodies[body]
	if b == nil {
		b = &LiveBody{Name: body, Started: time.Now(), done: make(chan struct{})}
		c.bodies[body] = b
		c.order = append(c.order, body)
	}
	b.State = state
	if state == "done" {
		b.Ended = time.Now()
	}
}

// observe is the loop's State seam: the body's state and its live session.
func (c *Court) observe(s *loop.Session, state string) {
	name := bodyName(s)
	c.setState(name, state)
	c.mu.Lock()
	if b := c.bodies[name]; b != nil && b.Session == nil {
		b.Session = s
	}
	c.mu.Unlock()
}

// Finish marks a body done with its report. Before the report goes out the Court holds its
// @$ lines on the world's shared queue (app/wake_held.go; the body's order book when there is
// no world) — claimed as dispatches by the wake at the turn boundary, where the chain gate
// re-checks each one (fail-closed; a refused order surfaces as a hole).
func (c *Court) Finish(name, report string) {
	var held []wire.Enqueue
	hold := c.app != nil && c.app.World != nil
	c.mu.Lock()
	b := c.bodies[name]
	if b == nil {
		b = &LiveBody{Name: name, done: make(chan struct{})}
		c.bodies[name] = b
		c.order = append(c.order, name)
	}
	if rep, ok := wire.ParseReport(report); ok {
		if hold {
			held = rep.Enqueues
		} else {
			b.orders = append(b.orders, rep.Enqueues...)
		}
		// court-card-exit-verdict law (leg 2a): the exit word is the first @S token — the loop
		// stamps a non-DONE exit as "FAIL — …", so token one is the journal's verdict.
		if rep.Status != "" {
			tok := strings.ToUpper(strings.Fields(rep.Status)[0])
			switch tok {
			case "DONE", "FAIL", "ESCALATE", "DENIED", "INTERRUPTED":
				b.Exit = tok
			}
		}
	}
	b.Failed = b.Exit == "FAIL" || b.Exit == "DENIED" || b.Exit == "ESCALATE"
	b.State, b.Ended, b.report = "done", time.Now(), report
	// the relay ledger: for each @$ the body's commission declared, its own report must carry the
	// echoed order toward that body, or the throne hears "[relay unbooked]"
	if exp := c.relays[strings.ToLower(name)]; len(exp) > 0 {
		delete(c.relays, strings.ToLower(name))
		var echoed []string
		if rep, ok := wire.ParseReport(report); ok {
			for _, o := range rep.Enqueues {
				echoed = append(echoed, o.Body)
			}
		}
		for _, e := range exp {
			found := false
			for _, eb := range echoed {
				if eb == e.Body {
					found = true
					break
				}
			}
			foundTracked := false
			if tb := c.bodies[e.Body]; tb != nil && tb.State != "done" {
				foundTracked = true
			}
			if !found && !foundTracked {
				c.wake = append(c.wake, fmt.Sprintf("[relay unbooked] %s closed and its report left '%s' unbooked (its commission carried '@$ %s: %s') — the relay needs the throne's hand", name, e.Body, e.Body, e.Ask))
			}
		}
	}
	select {
	case <-b.done:
	default:
		close(b.done)
	}
	f := c.onFinish
	c.mu.Unlock()
	if len(held) > 0 {
		c.app.holdOrders(name, held)
	}
	if f != nil && report != "" {
		f(name, report)
	}
}

// reportOf is a finished body's last report, "" while it runs or before it ever reported.
func (c *Court) reportOf(name string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if b := c.bodies[name]; b != nil && b.State == "done" {
		return b.report
	}
	return ""
}

// Orders drains a body's @$ order book — the dispatches it asked for at the boundary.
func (c *Court) Orders(name string) []wire.Enqueue {
	c.mu.Lock()
	defer c.mu.Unlock()
	b := c.bodies[name]
	if b == nil || len(b.orders) == 0 {
		return nil
	}
	out := b.orders
	b.orders = nil
	return out
}

// OrderBook drains every queued @$ order and returns each one with the report of any same-named
// Court that already ran and finished — so the REPL's turn-done drain can mint the new dispatch
// with thread ("the same body's earlier Court reported: @S …"). Orders are flat (no per-body
// queue); the order book refills as reports land.
func (c *Court) OrderBook() []struct {
	Body, Ask, Report string
} {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []struct {
		Body, Ask, Report string
	}
	for _, b := range c.bodies {
		for _, o := range b.orders {
			re := struct{ Body, Ask, Report string }{Body: o.Body, Ask: o.Ask}
			if done := c.bodies[o.Body]; done != nil && done.State == "done" && done.report != "" {
				re.Report = done.report
			}
			out = append(out, re)
		}
		b.orders = nil
	}
	return out
}

// Send queues a line for a running body (delivered at its next tool step).
// Mid-turn steering only — a human's queued line must NOT use this: it rides the
// running turn's context and breaks task isolation. Use Enqueue for FIFO steering
// (the B contract: complete the current task, then hear the next line).
func (c *Court) Send(name, text string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	b := c.bodies[strings.ToLower(name)]
	if b == nil {
		return fmt.Errorf("no body named %q in the court (%s)", name, strings.Join(c.order, ", "))
	}
	if b.State == "done" {
		return fmt.Errorf("%s is done — its context died with its task; dispatch it again", name)
	}
	b.mid = append(b.mid, text)
	return nil
}

// Enqueue FIFOs a line for a body with no mid-turn delivery: it sits in the inbox
// until an end-of-turn drain (finish/woke/repl turn-done) starts it as the next turn.
// Unlike Send, it never rides the running turn's next tool step — a human's queued
// line uses this so the current task completes before the new one is heard.
func (c *Court) Enqueue(name, text string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	b := c.bodies[strings.ToLower(name)]
	if b == nil {
		return fmt.Errorf("no body named %q in the court (%s)", name, strings.Join(c.order, ", "))
	}
	b.next = append(b.next, text)
	return nil
}

// Inbox drains a body's MID-TURN lines (the Send lane) — the loop's tool-batch seam reads
// exactly this; Enqueue lines never surface here.
func (c *Court) Inbox(name string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	b := c.bodies[name]
	if b == nil || len(b.mid) == 0 {
		return nil
	}
	lines := b.mid
	b.mid = nil
	return lines
}

// DrainAll empties both lanes (mid leftovers then the FIFO) — the turn-end seam's word:
// "finish with everything that arrived while you worked".
func (c *Court) DrainAll(name string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	b := c.bodies[name]
	if b == nil || len(b.mid)+len(b.next) == 0 {
		return nil
	}
	lines := append(b.mid, b.next...)
	b.mid, b.next = nil, nil
	return lines
}

// QueueDepth reports a body's unread line count (both lanes) — the dashboard's wedge meter.
func (c *Court) QueueDepth(name string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	b := c.bodies[name]
	if b == nil {
		return 0
	}
	return len(b.mid) + len(b.next)
}

// Wake queues a finished background Court's report for the dispatcher's next turn.
func (c *Court) Wake(report string) {
	c.mu.Lock()
	c.wake = append(c.wake, report)
	f := c.onWake
	c.mu.Unlock()
	if f != nil {
		f()
	}
}

// TakeWake returns the queued reports, emptying the queue.
func (c *Court) TakeWake() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := c.wake
	c.wake = nil
	return out
}

// Bodies lists the court in start order.
func (c *Court) Bodies() []LiveBody {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []LiveBody
	for _, n := range c.order {
		out = append(out, *c.bodies[n])
	}
	return out
}

// Live counts the bodies not done.
func (c *Court) Live() int {
	n := 0
	for _, b := range c.Bodies() {
		if b.State != "done" && b.Depth > 0 { // the session itself is not a dispatched agent
			n++
		}
	}
	return n
}

// Lines renders /agents: rank, office, model, state, elapsed, context %, tokens, cost.
func (c *Court) Lines() []string {
	bodies := c.Bodies()
	if len(bodies) == 0 {
		return []string{"the court is empty"}
	}
	sort.SliceStable(bodies, func(i, j int) bool { return bodies[i].Started.Before(bodies[j].Started) })
	var out []string
	for _, b := range bodies {
		end := time.Now()
		if !b.Ended.IsZero() {
			end = b.Ended
		}
		ctx := "ctx —"
		if b.Session != nil && b.Session.Reading().Available {
			ctx = fmt.Sprintf("ctx %d%%", b.Session.Reading().Percent())
		}
		sp := c.app.Journal.Body(b.Name)
		cost := "unpriced"
		if sp.Calls > 0 && sp.Unpriced == 0 {
			cost = fmt.Sprintf("$%.4f", sp.USD)
		}
		out = append(out, fmt.Sprintf("%-20s %-8s %-10s %-32s %-15s %6s  %s  %6d tok  %s", b.Name, orStr(b.Rank, "rimuru"), orStr(b.Office, "—"), orStr(b.Model, "—"), b.State, end.Sub(b.Started).Round(time.Second), ctx, sp.Tokens(), cost))
	}
	return out
}

// StatusLine is the one line above the prompt.
func (c *Court) StatusLine(s *loop.Session) string {
	tot := c.app.Journal.Total()
	cost := "unpriced"
	switch {
	case tot.Calls == 0:
		cost = "no calls yet"
	case tot.Unpriced == 0:
		cost = fmt.Sprintf("$%.4f", tot.USD)
	}
	ctx := "ctx —"
	if r := readingOf(s); r.Available {
		ctx = fmt.Sprintf("ctx %d%% (%s)", r.Percent(), r.Zone)
	}
	live := c.Live()
	courts := ""
	if live > 0 {
		courts = fmt.Sprintf(" · %d live", live)
	}
	return fmt.Sprintf("%s · %s · %s · %d tok · %s%s", c.app.Cfg.Dist.Name, c.app.mountModel.Ref.Model, ctx, tot.Tokens(), cost, courts)
}

// Close marks every body done.
func (c *Court) Close() {
	for _, b := range c.Bodies() {
		if b.State != "done" {
			c.Finish(b.Name, "")
		}
	}
}

func readingOf(s *loop.Session) instrument.Context {
	if s == nil {
		return instrument.Context{}
	}
	return s.Reading()
}
