package app

import (
	"strings"
	"testing"

	"github.com/Kaginari/isekai/wire"
)

// relay bookkeeping: expect the commission's @$ echo at Finish, one wake per unbooked order.
func TestCourtRelayBookkeeping(t *testing.T) {
	c := newCourt(testTmpApp(t, t.TempDir()))
	c.ExpectRelay("elf-isekai", wire.EnqueuesOf("…work the findings…\n@$ slime-tui-board: land the Arm A fixtures per the fence\n@E 0"))
	c.Track("elf-isekai", "elf", "great-sage", "kimi", nil, 1)
	if len(c.relays["elf-isekai"]) != 1 {
		t.Fatal("expectation not registered: ", c.relays)
	}
	// truthful echo: report carries the @$ → no wake
	c.Finish("elf-isekai", "@S DONE\n@$ slime-tui-board: land the Arm A fixtures per the fence\n@E 0")
	if woke := c.TakeWake(); len(woke) != 0 {
		t.Fatalf("echoed relays wake nobody: %v", woke)
	}
	// no echo, body never tracked → one unbooked wake naming body + ask
	c.ExpectRelay("elf-isekai", wire.EnqueuesOf("@$ ring-wraith: cross the ford with the findings\n@E 0"))
	c.Finish("elf-isekai", "@S DONE\n@F strong\n@E 0")
	woke := c.TakeWake()
	if len(woke) == 0 || !strings.Contains(woke[0], "[relay unbooked]") || !strings.Contains(woke[0], "ring-wraith") || !strings.Contains(woke[0], "cross the ford") {
		t.Fatalf("relay unbooked must be named with body and ask: %v", woke)
	}
	// booking by tracking: the body sits on the board running → no wake
	c.ExpectRelay("elf-isekai", wire.EnqueuesOf("@$ slime-memory: chair the file kit beacon\n@E 0"))
	c.Track("slime-memory", "slime", "great-sage", "kimi", nil, 2)
	if len(c.relays["elf-isekai"]) != 0 {
		t.Error("tracking clears the expectation")
	}
	c.Finish("elf-isekai", "@S DONE\n@E 0")
	if woke := c.TakeWake(); len(woke) != 0 {
		t.Fatalf("tracked relays booked: %v", woke)
	}
}
