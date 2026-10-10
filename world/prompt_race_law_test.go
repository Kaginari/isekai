package world

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The race-law slot (Veldora 2026-10-03): each body boots with the crest + its race's page —
// one race's words must never leak into another's context (context economy with intelligence).
func TestRaceLawNoLeaking(t *testing.T) {
	dir := fixture(t)
	rl := filepath.Join(dir, ".isekai", "law", "race")
	if err := os.MkdirAll(rl, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(rl, "slime.md"), []byte("# Race law — slime\n\n- SLIME-ONLY-MARKER the orc commands you one slice.\n"), 0o644)
	os.WriteFile(filepath.Join(rl, "elf.md"), []byte("# Race law — elf\n\n- ELF-ONLY-MARKER hold the 3 keep 2.\n"), 0o644)
	os.WriteFile(filepath.Join(rl, "rimuru.md"), []byte("# Race law — rimuru\n\n- THRONE-ONLY-MARKER author three cards.\n"), 0o644)

	w := open(t, dir)
	for _, tc := range []struct {
		as         string
		mustWear   string
		mustNotSee []string
	}{
		{"slime-auth", "SLIME-ONLY-MARKER", []string{"ELF-ONLY-MARKER", "THRONE-ONLY-MARKER"}},
		{"slime-api", "SLIME-ONLY-MARKER", []string{"ELF-ONLY-MARKER", "THRONE-ONLY-MARKER"}},
		{"elf-core", "ELF-ONLY-MARKER", []string{"SLIME-ONLY-MARKER", "THRONE-ONLY-MARKER"}},
	} {
		p := w.Prompt(tc.as, DefaultPrompt(), nil)
		if !strings.Contains(p, tc.mustWear) {
			t.Fatalf("%s: its own race's page is missing\n%s", tc.as, p[:400])
		}
		for _, bad := range tc.mustNotSee {
			if strings.Contains(p, bad) {
				t.Fatalf("%s sees another race's law (%s) — a leak", tc.as, bad)
			}
		}
	}
	// the throne never blanks its page — rimuru belongs to a race too
	p := w.Prompt("rimuru", DefaultPrompt(), nil)
	if !strings.Contains(p, "THRONE-ONLY-MARKER") {
		t.Fatal("rimuru's page never reached the throne")
	}
	// a rank with no page yet gets NO phantom section — silent, never a lie
	if q := w.Prompt("nobody-nowhere", DefaultPrompt(), nil); strings.Contains(q, "Law of your race") {
		t.Fatalf("a rankless body must not wear a law section:\n%s", q[:300])
	}
}
