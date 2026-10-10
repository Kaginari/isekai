package world

import (
	"strings"
	"testing"
)

// Pins both legs of the genome-mutation law (Veldora decree 2026-10-02): a burned lane
// climbs, an out-of-ladder or topmost lane holds (no silent mutation, no invention).
func TestLaneUp(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"openai/muse-glimmer-30b", "openai/kimi-3-conduit"},
		{"openai/kimi-3-conduit", "openai/glm-5-3-conduit"},
		{"openai/glm-5-3-conduit", ""},                       // peak: split, don't mutate past it
		{"openai/some-future-model", ""},                     // off-ladder: advice belongs to config, not to the ladder
		{"litellm/kimi-3-conduit", "openai/glm-5-3-conduit"}, // machine-config naming resolves
	} {
		if got := laneUp(c.in); got != c.want {
			t.Errorf("laneUp(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// stratTargets on the fixture world: rimuru may dispatch the elf only (Law 2 + strict
// layer flow — slime/orc are deeper) and an elf may dispatch the orcs it parents.
func TestStratTargets(t *testing.T) {
	dir := fixture(t)
	w := open(t, dir)
	thrn := w.Engine("rimuru", Build{Tools: nil})
	got := w.stratTargets(thrn, CourtOptions{StrictFlow: true})
	if !strings.Contains(got, "elf-core") || strings.Contains(got, "slime-auth") {
		t.Fatalf("throne may reach only its child rank (elf-core here): %q", got)
	}
	elf := w.Engine("elf-core", Build{Tools: nil})
	got = w.stratTargets(elf, CourtOptions{StrictFlow: true})
	if !strings.Contains(got, "orc-security") || strings.Contains(got, "slime-auth") {
		t.Fatalf("elf-core may reach its orcs, never the slimes directly: %q", got)
	}
}
