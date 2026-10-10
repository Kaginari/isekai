package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kaginari/isekai/config"
	"github.com/Kaginari/isekai/tool"
)

func mustDefaults(t *testing.T) *config.Config {
	t.Helper()
	c, err := config.Defaults("isekai")
	if err != nil {
		t.Fatalf("defaults: %v", err)
	}
	c.Plan.Dir = "tmp/plans"
	return c
}

// marshal a deck penned at a fixed time; helper for the archive tests. Updated is the string
// stamp planStale judges on; writePlan restamps, so tests that pin a time marshal by hand.
func writePlanFile(t *testing.T, dir, name string, at time.Time, done bool) string {
	t.Helper()
	status := "in_progress"
	if done {
		status = "done"
	}
	d := planDeck{Body: name, Updated: at.UTC().Format("2006-01-02T15:04:05Z"), Items: []planItem{{ID: "q1", Text: "sealed card", Status: status}}}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	f := planFile(dir, name)
	if err := os.WriteFile(f, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestPlanArchiveSweep(t *testing.T) {
	dir := t.TempDir()
	boot := time.Now()
	session := "slime-x"

	staleDone := writePlanFile(t, dir, "orc-host", boot.Add(-3*time.Hour), true)
	staleLive := writePlanFile(t, dir, "elf-isekai", boot.Add(-time.Hour), false)
	fresh := writePlanFile(t, dir, "slime-a", boot.Add(time.Hour), true)
	own := writePlanFile(t, dir, session, boot.Add(-2*time.Hour), true)
	corrupt := planFile(dir, "slime-b")
	if err := os.WriteFile(corrupt, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	moved, err := planArchive(dir, session, boot)
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if moved != 2 {
		t.Fatalf("sweep moved = %d, want 2 (stale decks of other bodies)", moved)
	}

	archive := planArchiveDir(dir)
	for _, f := range []string{staleDone, staleLive} {
		got, err := os.ReadFile(filepath.Join(archive, filepath.Base(f)))
		if err != nil {
			t.Fatalf("archived deck unreadable: %s: %v", f, err)
		}
		want, _ := os.ReadFile(planFile(dir, filepath.Base(f)))
		_ = want // live copy is gone; comparing below
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Errorf("stale deck remained in live dir: %s", f)
		}
		if len(got) == 0 {
			t.Errorf("archived deck is empty: %s", f)
		}
	}
	// live dir keeps only current decks: fresh, own (never archived), corrupt (fail-open), plus the archive dir itself
	entries, _ := os.ReadDir(dir)
	var liveLeft []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		liveLeft = append(liveLeft, e.Name())
	}
	if len(liveLeft) != 3 {
		t.Errorf("live dir entries after sweep = %v, want exactly 3 (fresh, own, corrupt)", liveLeft)
	}
	for _, f := range []string{fresh, own, corrupt} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("current deck missing after sweep: %s: %v", f, err)
		}
	}
	// append-only, never delete: a pre-existing archived twin is left untouched and the rename is deferred
	preStale := writePlanFile(t, dir, "slime-c", boot.Add(-time.Hour), true)
	if err := os.MkdirAll(archive, 0o755); err != nil {
		t.Fatal(err)
	}
	twin := planFile(archive, "slime-c")
	if err := os.WriteFile(twin, []byte(`{"body":"slime-c","updated":"older"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(twin)
	moved2, err := planArchive(dir, session, boot)
	if err != nil {
		t.Errorf("sweep with an archive twin must not error (append-only defers the rename): %v", err)
	}
	if moved2 != 0 {
		t.Errorf("sweep moved = %d, want 0 (the twin's rename is deferred, not forced)", moved2)
	}
	after, _ := os.ReadFile(twin)
	if string(before) != string(after) {
		t.Errorf("archived twin mutated: the archive must be append-only")
	}
	if _, err := os.Stat(preStale); err != nil {
		t.Errorf("stale live deck removed when its rename was deferred: %v", err)
	}
}

func TestPlanReadFallThrough(t *testing.T) {
	dir := t.TempDir()
	boot := time.Now()
	f := writePlanFile(t, dir, "slime-x", boot, true)

	live := planRead(dir, "slime-x")
	if live.Body != "slime-x" || len(live.Items) != 1 {
		t.Fatalf("planRead live: %+v", live)
	}
	if err := os.MkdirAll(planArchiveDir(dir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(f, planFile(planArchiveDir(dir), "slime-x")); err != nil {
		t.Fatal(err)
	}
	arch := planRead(dir, "slime-x")
	if arch.Body != "slime-x" || len(arch.Items) != 1 || string(arch.Items[0].Status) != "done" {
		t.Fatalf("planRead archived fall-through: %+v", arch)
	}
	missing := planRead(dir, "slime-nope")
	if missing.Body != "" || len(missing.Items) != 0 {
		t.Fatalf("planRead missing: want zero deck, got %+v", missing)
	}
}

func TestPlanGateUnchangedByArchive(t *testing.T) {
	dir := t.TempDir()
	boot := time.Now()
	// a stale deck archived away must NOT yield an open reason — the gate passes live decks only,
	// and planStale already silences a stale live deck; the sweep only moves the file.
	writePlanFile(t, dir, "slime-x", boot.Add(-time.Hour), true)     // stale + all-done
	writePlanFile(t, dir, "elf-isekai", boot.Add(-time.Hour), false) // stale + open (pre-boot pending)

	if reason := planOpenReason(dir, "elf-isekai", boot); reason != "" {
		t.Errorf("stale open deck yields reason %q — a stale deck's open item must not gate", reason)
	}
	if _, err := planArchive(dir, "slime-y", boot); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	// both stale decks are archived now; live dir holds neither. The open item must still not gate
	// (gate behavior for live decks is unchanged by the archive existing).
	if reason := planOpenReason(dir, "elf-isekai", boot); reason != "" {
		t.Errorf("archived stale open deck yields reason %q after sweep", reason)
	}
	// a live post-boot open deck gates exactly as before the sweep existed
	writePlanFile(t, dir, "orc-host", boot.Add(time.Hour), false)
	if reason := planOpenReason(dir, "orc-host", boot); reason == "" {
		t.Errorf("live open deck post-sweep yields no reason — live gate behavior must be unchanged")
	}
}

func TestPlanToolBootSweepRun(t *testing.T) {
	dir := t.TempDir()
	cfg := mustDefaults(t)
	a := &App{Root: dir, Cfg: cfg, planBoot: time.Now()}
	// world dir under Root: planDir resolves to <Root>/tmp/plans on a bare world
	pdir := a.planDir()
	if err := os.MkdirAll(pdir, 0o755); err != nil {
		t.Fatal(err)
	}
	boot := a.plansBoot()
	writePlanFile(t, pdir, "orc-host", boot.Add(-time.Hour), true)
	writePlanFile(t, pdir, "slime-y", boot.Add(time.Hour), false)

	pl := a.planTool("slime-y")
	if res := pl.Run(t.Context(), tool.Env{Root: dir}, []byte(`{"items":[{"id":"q1","text":"landed","status":"pending"}]}`)); res.Err {
		t.Fatalf("plan run: %s", res.Output)
	}
	if _, err := os.Stat(planFile(planArchiveDir(pdir), "orc-host")); err != nil {
		t.Errorf("boot sweep on first plan run did not archive the stale deck: %v", err)
	}
	if _, err := os.Stat(planFile(pdir, "slime-y")); err != nil {
		t.Errorf("session's own deck missing after sweep+write: %v", err)
	}
	if _, err := os.Stat(planFile(pdir, "orc-host")); !os.IsNotExist(err) {
		t.Errorf("stale deck still live after boot sweep")
	}
	// second run must not re-sweep (append-only, idempotent — the archived deck is not touched again)
	arcBefore, _ := os.ReadFile(planFile(planArchiveDir(pdir), "orc-host"))
	if res := pl.Run(t.Context(), tool.Env{Root: dir}, []byte(`{"tick":"q1","status":"done"}`)); res.Err {
		t.Fatalf("plan tick: %s", res.Output)
	}
	arcAfter, _ := os.ReadFile(planFile(planArchiveDir(pdir), "orc-host"))
	if string(arcBefore) != string(arcAfter) {
		t.Errorf("archived deck mutated by a later plan act")
	}
}
