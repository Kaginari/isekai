package app

import (
	"strings"
	"testing"
)

// The never-list restates the code: the territory policy and the gate leave a Slime its own nest
// in the world's scratch (world.ZoneOfPath, SoilScratch), so the Slime's never-line names that
// carve-out, and the gate section says what the scratch is.
func TestLawNeverListNamesTheScratch(t *testing.T) {
	law := LawText()
	i := strings.Index(law, "- **A Slime** never writes outside its zone")
	if i < 0 {
		t.Fatal("the Slime's never-line is gone")
	}
	line := law[i:]
	if j := strings.Index(line[2:], "\n- "); j >= 0 {
		line = line[:j+2]
	}
	if !strings.Contains(line, "`.isekai/tmp/`") {
		t.Errorf("the Slime's never-line contradicts the scratch carve-out: %q", line)
	}
	gate := law[strings.Index(law, "## The gate"):]
	gate = gate[:strings.Index(gate, "\n## ")]
	if !strings.Contains(gate, "The world's scratch (`.isekai/tmp/`") {
		t.Error("the gate section does not say what the scratch is")
	}
}
