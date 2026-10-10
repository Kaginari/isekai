package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// The alt+n "current" switch: a stale deck stays browsable, the banner names the hint, and
// alt+n walks the view to the current deck (the throne default the elf pick uses). Plain c
// keeps typing; another stale view switches back.
func TestPlanCurrentSwitch(t *testing.T) {
	host := &fakeHost{plans: []RailPlan{
		{Body: "slime-tui-board", Rank: "slime", Office: "p2", Stale: true,
			Items: []RailPlanItem{{Text: "yesterday's slice", Status: "pending"}}},
		{Body: "elf-colony", Rank: "elf", Office: "p1",
			Items: []RailPlanItem{{Text: "slice now", Status: "in_progress"}}},
		{Body: "rimuru", Rank: "rimuru", Office: "p1",
			Items: []RailPlanItem{{Text: "the throne's slice", Status: "in_progress"}}},
	}}
	m := New(host, NewTheme(true), DefaultWords())
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	if got, ok := mm.(*Model); ok {
		m = got
	}
	m.Update(evPlans{plans: m.host.Plans()}) // the tick's read, landed
	if len(m.plans) == 0 {
		t.Fatal("evPlans must populate m.plans")
	}
	if got := m.plans[m.planIdx]; got.Body != "rimuru" || got.Archived {
		t.Fatalf("the throne deck leads by default: %+v", m.plans[m.planIdx])
	}

	// browse the stale deck: ctrl+p cycles the rail (input.go:45)
	m.Update(tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}) // rimuru(idx 2) → slime-tui-board(idx 0) [wrap]
	if got := m.plans[m.planIdx]; got.Body != "slime-tui-board" || !got.Stale {
		t.Fatalf("the stale deck is now in view: %+v", got)
	}
	view := ansi.Strip(m.Render())
	if !strings.Contains(view, "yesterday's slice") {
		t.Fatal("the stale deck stays browsable")
	}
	if strings.Contains(view, "the throne's slice") {
		t.Fatal("the live deck is not the view")
	}
	if !strings.Contains(view, "ᴀʟᴛ+ɴ —") || !strings.Contains(view, "ᴄᴜʀʀᴇɴᴛ") {
		t.Fatalf("the stale banner names the hint:\n%s", view)
	}

	// alt+n switches to the current deck and frees the sticky pick
	m.key(tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt})
	if m.planSticky {
		t.Fatal("the switch hands the default back")
	}
	if got := m.plans[m.planIdx]; got.Body != "rimuru" || got.Archived {
		t.Fatalf("the current deck is in view: %+v", got)
	}
	if !strings.Contains(ansi.Strip(m.Render()), "the throne's slice") {
		t.Fatal("the current deck renders")
	}

	// plain c always types
	m.key(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if m.input.Value() != "c" {
		t.Fatalf("plain c keeps typing on a live view: %q", m.input.Value())
	}
	m.setInput("")

	// and from yet another stale view the same key walks back to that live one
	m.Update(tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	if !m.plans[m.planIdx].Stale {
		t.Fatalf("setup: second stale view wanted, got %+v", m.plans[m.planIdx])
	}
	m.key(tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt})
	if got := m.plans[m.planIdx]; got.Body != "rimuru" {
		t.Fatalf("the switch is not per-deck: %s", got.Body)
	}

	// an archived deck answers alt+n the same way — it is a stale view once removed
	// (the archived pin rides the Plans-host side; here we assert the switch on it).
	// Flip the stale slime deck to archived, put the view back on it, press alt+n → rimuru.
	m.plans[0].Archived = true
	m.planIdx = 0
	m.key(tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt})
	if got := m.plans[m.planIdx]; got.Body != "rimuru" {
		t.Fatalf("an archived view switches too: %s", got.Body)
	}
}

// 2026-10-13 relay (throne→elf-isekai→orc-host, Card A): a session booting onto a STALE
// default deck must never eat plain 'c' runes.
func TestPlanCurrentStaleDefaultTypesC(t *testing.T) {
	host := &fakeHost{plans: []RailPlan{
		{Body: "rimuru", Rank: "rimuru", Office: "p1", Stale: true,
			Items: []RailPlanItem{{Text: "last session's intent", Status: "pending"}}},
	}}
	m := New(host, NewTheme(true), DefaultWords())
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	if got, ok := mm.(*Model); ok {
		m = got
	}
	m.Update(evPlans{plans: m.host.Plans()})
	if len(m.plans) == 0 {
		t.Fatal("evPlans must populate m.plans")
	}
	if got := m.plans[m.planIdx]; !got.Stale {
		t.Fatalf("setup: the only deck is stale, the default view faces it (fallback preserved): %+v", got)
	}
	if m.planSticky {
		t.Fatal("setup: the boot-default view is not a deliberate pin")
	}

	// every lowercase 'c' press types; nothing snaps, no banner fires
	for want := 1; want <= 3; want++ {
		m.key(tea.KeyPressMsg{Code: 'c', Text: "c"})
		if got := m.input.Value(); got != strings.Repeat("c", want) {
			t.Fatalf("press %d: plain c falls through to the input line on a boot-default stale view: %q", want, got)
		}
	}
	m.setInput("")
}

// Snapping from a stale view whose session has NO live deck says so — the snap never lands on
// another stale deck.
func TestPlanCurrentNoLiveDeck(t *testing.T) {
	host := &fakeHost{plans: []RailPlan{
		{Body: "slime-a", Rank: "slime", Stale: true, Archived: true,
			Items: []RailPlanItem{{Text: "stale slice a", Status: "pending"}}},
		{Body: "elf-b", Rank: "elf", Archived: true,
			Items: []RailPlanItem{{Text: "archived slice b", Status: "pending"}}},
	}}
	m := New(host, NewTheme(true), DefaultWords())
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	if got, ok := mm.(*Model); ok {
		m = got
	}
	m.Update(evPlans{plans: m.host.Plans()})
	if len(m.plans) == 0 {
		t.Fatal("evPlans must populate m.plans")
	}

	// pin the stale deck, then press alt+n: nothing live to snap to…
	m.key(tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}) // flips planSticky on
	m.planIdx = 0                                       // view the stale deck itself
	if !m.planSticky || !m.plans[m.planIdx].Stale {
		t.Fatalf("setup: ctrl+p pins the stale deck, got sticky=%v deck=%+v", m.planSticky, m.plans[m.planIdx])
	}
	before := m.planIdx
	m.key(tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt})
	if m.planIdx != before {
		t.Fatalf("no live deck exists: the view stays put, got %d (was %d)", m.planIdx, before)
	}
	if m.message != "plan · no live deck" {
		t.Fatalf("the snap says 'plan · no live deck' instead of looping onto another stale deck, got %q", m.message)
	}
	if got := m.input.Value(); got != "" {
		t.Fatalf("the chord never types: %q", got)
	}
	// …and plain c types, pinned or not
	m.key(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if got := m.input.Value(); got != "c" {
		t.Fatalf("plain c types on a pinned stale deck: %q", got)
	}
	m.setInput("")
}

// alt+n works from any view — a stale one, NOT sticky — and lands on the live throne deck;
// shift+C is a typed capital.
func TestPlanCurrentAltN(t *testing.T) {
	host := &fakeHost{plans: []RailPlan{
		{Body: "slime-tui-board", Rank: "slime", Stale: true,
			Items: []RailPlanItem{{Text: "yesterday's slice", Status: "pending"}}},
		{Body: "rimuru", Rank: "rimuru",
			Items: []RailPlanItem{{Text: "the throne's slice", Status: "in_progress"}}},
	}}
	m := New(host, NewTheme(true), DefaultWords())
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	if got, ok := mm.(*Model); ok {
		m = got
	}
	m.Update(evPlans{plans: m.host.Plans()})
	if len(m.plans) == 0 {
		t.Fatal("evPlans must populate m.plans")
	}
	m.planIdx = 0 // view the stale slime deck WITHOUT a sticky pin (boot-style view)
	m.planSticky = false
	if !m.plans[m.planIdx].Stale {
		t.Fatalf("setup: stale deck wanted at idx 0, got %+v", m.plans[m.planIdx])
	}

	m.key(tea.KeyPressMsg{Code: 'C', Text: "C", Mod: tea.ModShift})
	if got := m.input.Value(); got != "C" || m.planIdx != 0 {
		t.Fatalf("shift+C types and leaves the view: %q idx %d", got, m.planIdx)
	}
	m.setInput("")
	m.key(tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt})
	if got := m.plans[m.planIdx]; got.Body != "rimuru" || got.Stale || got.Archived {
		t.Fatalf("alt+n snaps to the live throne deck from any view: %+v", got)
	}
	liveBanner := "plan · rimuru (2/2) · current"
	if m.message != liveBanner {
		t.Fatalf("the alt+n banner fires on a non-sticky view: want %q, got %q", liveBanner, m.message)
	}
	if got := m.input.Value(); got != "" {
		t.Fatalf("alt+n never types: %q", got)
	}
}
