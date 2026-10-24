package world

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/tool"
	"github.com/Kaginari/isekai/wire"
)

// keeperWorld is the fixture with a world-keeper slime whose Territory claims `.isekai` whole —
// the broad claim every seat, shelf and scratch path sits under.
func keeperWorld(t *testing.T) (string, *World) {
	t.Helper()
	dir := fixture(t)
	if err := os.MkdirAll(filepath.Join(dir, ".isekai/slime/keeper"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".isekai/slime/keeper/README.md"), []byte("# slime-keeper\n\n- **Rank:** Slime\n- **Territory:** `.isekai`\n- **Reports to:** orc-security\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, open(t, dir)
}

// ZoneOfPath is one reading for every ownership question: the harness's shelves and the scratch
// are nobody's, a seat is its creature's even under a broader claim, then territory.
func TestZoneOfPath(t *testing.T) {
	_, w := keeperWorld(t)
	for _, c := range []struct {
		rel, owner string
		soil       Soil
	}{
		{".isekai/shared/queue.jsonl", "", SoilHarness},
		{".isekai/instruments/loop/r1.jsonl", "", SoilHarness},
		{".isekai/tmp/scratch/r1/x.txt", "", SoilScratch},
		{".isekai/tmp/plans/slime-auth.json", "", SoilScratch},
		{".isekai/slime/auth/README.md", "slime-auth", SoilSeat},
		{".isekai/slime/auth/notes.md", "slime-auth", SoilSeat},
		{".isekai/log.md", "slime-keeper", SoilTerritory},
		{"src/auth/login.go", "slime-auth", SoilTerritory},
		{"README.md", "", SoilOpen},
	} {
		z := w.ZoneOfPath(c.rel)
		got := ""
		if z.Creature != nil {
			got = z.Creature.Name
		}
		if z.Soil != c.soil || got != c.owner {
			t.Errorf("%s: %s/%q, want %s/%q", c.rel, z.Soil, got, c.soil, c.owner)
		}
		if o := w.Owner(c.rel); (o == nil) != (c.owner == "") || (o != nil && o.Name != c.owner) {
			t.Errorf("Owner(%s) disagrees with ZoneOfPath: %v", c.rel, o)
		}
	}
}

// A creature's seat leaves the pairing set: its README and seat files never bill the body whose
// claim covers `.isekai`; the scratch is the writer's own nest; the shelves bill no creature.
func TestGateSeatAndScratchSoil(t *testing.T) {
	_, w := keeperWorld(t)
	g := DefaultHooks().Gate
	rep := &wire.Report{Status: "DONE"}
	gate := func(as string, wrote ...string) Verdict {
		t.Helper()
		return w.Gate(context.Background(), nil, g, as, wrote, rep, true, "x")
	}
	for _, rel := range []string{".isekai/slime/auth/README.md", ".isekai/slime/auth/notes.md", ".isekai/tmp/scratch/r1/x.txt", ".isekai/tmp/plans/slime-auth.json"} {
		if v := gate("slime-auth", rel); v.Word != "pass" {
			t.Errorf("slime-auth writing %s: %s", rel, v)
		}
	}
	if v := gate("slime-auth", ".isekai/tmp/plans/slime-api.json"); v.Word != "fail" || !strings.Contains(v.String(), "is slime-api's plan deck, written by slime-auth") {
		t.Errorf("another body's deck: %s", v)
	}
	if v := gate("slime-auth", ".isekai/shared/queue.jsonl"); strings.Contains(v.String(), "slime-keeper") {
		t.Errorf("the shared shelf billed the keeper: %s", v)
	}
	if v := gate("slime-api", ".isekai/slime/auth/notes.md"); v.Word != "fail" || !strings.Contains(v.String(), "belongs to slime-auth, written by slime-api") {
		t.Errorf("a foreign body on another's seat: %s", v)
	}
}

// The territory policy admits the scratch to every body — its own nest — but never another
// body's plan deck.
func TestPolicyScratchNest(t *testing.T) {
	dir, w := keeperWorld(t)
	pol := w.Policy("slime-api", TerritoryOptions{Enabled: true})
	acc := func(rel string) tool.Access {
		return tool.Access{Class: tool.Write, Paths: []string{filepath.Join(dir, filepath.FromSlash(rel))}}
	}
	if err := pol(acc(".isekai/tmp/scratch/r1/out.txt")); err != nil {
		t.Errorf("own scratch refused: %v", err)
	}
	if err := pol(acc(".isekai/tmp/plans/slime-api.json")); err != nil {
		t.Errorf("own deck refused: %v", err)
	}
	if err := pol(acc(".isekai/tmp/plans/slime-auth.json")); err == nil || !strings.Contains(err.Error(), "slime-auth's plan deck") {
		t.Errorf("another body's deck: %v", err)
	}
	if err := pol(acc("src/auth/login.go")); err == nil {
		t.Error("foreign territory admitted")
	}
}
