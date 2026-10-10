package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tierOf(t *testing.T, r *ReviewResult, name string) ReviewTier {
	t.Helper()
	for _, tier := range r.Tiers {
		if tier.Name == name {
			return tier
		}
	}
	t.Fatalf("no tier %q in %v", name, r.Tiers)
	return ReviewTier{}
}

func findingsContain(tier ReviewTier, signal, route string) bool {
	for _, f := range tier.Findings {
		if strings.Contains(f.Signal, signal) && strings.Contains(f.Route, route) {
			return true
		}
	}
	return false
}

// TestReviewTiers: each tier must be read with its own lens — the short tier must name the
// cross-body question stem (attention), the shared tier the restated lesson (lessons), and a
// world with no long index must say so out loud with the friction family it finds in the
// episode stream (genomic). Every finding must carry its evolution route.
func TestReviewTiers(t *testing.T) {
	world, home := fixture(t)
	w, err := Open(world)
	if err != nil {
		t.Fatal(err)
	}
	w.Home = home

	shortDir := filepath.Join(world, ".isekai", "memory", "short")
	os.MkdirAll(shortDir, 0o755)
	os.WriteFile(filepath.Join(shortDir, "rimuru.jsonl"), []byte(
		`{"q":"how does the dispatch gate rank the flow","results":[{"src":".isekai/isekai.md"}]}`+"\n"+
			`{"q":"how does the dispatch gate rank the flow","results":[{"src":".isekai/log.md"},{"src":".isekai/memory/shared/notes.jsonl"}]}`+"\n"), 0o644)
	os.WriteFile(filepath.Join(shortDir, "elf-isekai.jsonl"), []byte(
		`{"q":"how does the dispatch gate rank the flow","results":[{"src":".isekai/isekai.md"}]}`+"\n"), 0o644)

	sharedDir := filepath.Join(world, ".isekai", "memory", "shared")
	os.MkdirAll(sharedDir, 0o755)
	dup := `{"at":"2026-10-02T00:00:00.000Z","by":"rimuru","kind":"law","text":"reindex last, after the writing wave lands, never mid-wave"}` + "\n"
	os.WriteFile(filepath.Join(sharedDir, "notes.jsonl"), []byte(
		dup+strings.Replace(dup, "rimuru", "elf-isekai", 1)), 0o644)

	log := filepath.Join(world, ".isekai", "log.md")
	raw, _ := os.ReadFile(log)
	os.WriteFile(log, []byte(string(raw)+"\n- **Maintenance:** ⚙️ echo from before the law note\n"), 0o644)

	r := w.Review()
	if len(r.Tiers) != 4 {
		t.Fatalf("tiers = %d, want 4", len(r.Tiers))
	}

	short := tierOf(t, r, "short — attention")
	if !findingsContain(short, "cross-body question stem", "[genesis]") {
		t.Fatalf("short tier missed the cross-body stem or its genesis route: %+v", short.Findings)
	}
	if !findingsContain(short, "junk-anchor", "[harness]") {
		t.Fatalf("short tier missed the junk share: %+v", short.Findings)
	}

	shared := tierOf(t, r, "shared — lessons")
	if !findingsContain(shared, "lesson restated", "[law]") {
		t.Fatalf("shared tier missed the restated lesson: %+v", shared.Findings)
	}

	long := tierOf(t, r, "long — genomic")
	if !findingsContain(long, "no long-term index", "[harness]") {
		t.Fatalf("long tier missed the absent index: %+v", long.Findings)
	}
	if !findingsContain(long, "friction family maintenance-marker", "") {
		t.Fatalf("long tier missed the marker family in log.md: %+v", long.Findings)
	}

	value := tierOf(t, r, "value — decisions")
	if !findingsContain(value, "no plan-outcome rows yet", "[harness]") {
		t.Fatalf("value tier must say the learning loop is empty Ground Zero, got: %+v", value.Findings)
	}
}
