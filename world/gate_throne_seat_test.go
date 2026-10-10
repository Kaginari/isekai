package world

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/wire"
)

func writeFixtureTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for p, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(p))
		os.MkdirAll(filepath.Dir(full), 0o755)
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// The throne's own seat files — the dimension manifest (.isekai/dimension.yaml) and the law
// (.isekai/isekai.md) — are rimuru's work product, of one class with a creature's own doc. When
// the throne authors one, Owner must resolve to the throne (rimuru), not fall to the gate
// holder (orc): before the fix the manifest fell to orc-core by Owner's gate-holder fallback, so
// rimuru's lawful relocate of it failed both right-author and Vitality (demanding orc's README).
func TestGateThroneSeatFilesAreRimurus(t *testing.T) {
	dir := fixture(t)
	w := open(t, dir)
	opt := DefaultHooks().Gate
	opt.TraitsHold = false // the fixture's verify lines are not at issue here
	// DocTruthful STAYS ON — the live gate's Vitality check is half the live failure.
	for _, seat := range []string{".isekai/dimension.yaml", ".isekai/isekai.md"} {
		v := w.Gate(context.Background(), nil, opt, Rimuru, []string{seat}, &wire.Report{Status: "DONE"}, true, "relocate manifest")
		if strings.Contains(v.String(), "Right slime authored") {
			t.Errorf("%s: right-author fired on a throne seat file written by the throne: %s", seat, v)
		}
		if strings.Contains(v.String(), "Doc truthful") {
			t.Errorf("%s: Vitality demanded a body's README for the throne's own seat file: %s", seat, v)
		}
	}
}

// The throne-seat gate-fail reproduced live in the chronicle: `.isekai/dimension.yaml belongs
// to orc-core, written by rimuru`. The existing fixture worlds lack a gate-holder whose
// territory reaches the manifest, so it was invisible there. This test fixture supplies the
// missing shape: an orc gate-holder whose territory prefix is `.isekai/` — exactly the live
// throne seat that fired.
func TestGateThroneSeatLiveShape(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		".isekai/isekai.md":          lawText,
		".isekai/log.md":             "# Chronicle\n\nAppend-only.\n\n---\n",
		".isekai/orc/core/README.md": "# orc-core\n\n- **Rank:** Orc\n- **Territory:** `.isekai/`\n- **Verify:** `true`\n",
		// no elf, no slime: the live throne's orc seat alone is enough to fire the bug
	}
	writeFixtureTree(t, dir, files)
	w, err := Open(dir, Isekai(), Options{Ontology: true})
	if err != nil {
		t.Fatal(err)
	}
	opt := DefaultHooks().Gate
	opt.TraitsHold = false
	for _, seat := range []string{".isekai/dimension.yaml", ".isekai/isekai.md"} {
		v := w.Gate(context.Background(), nil, opt, Rimuru, []string{seat}, &wire.Report{Status: "DONE"}, true, "relocate manifest")
		if strings.Contains(v.String(), "Right slime authored") {
			t.Errorf("[live-shape] %s: right-author fired on a throne seat file written by the throne: %s", seat, v)
		}
		if strings.Contains(v.String(), "Doc truthful") {
			t.Errorf("[live-shape] %s: Vitality demanded a body's README for the throne's own seat file: %s", seat, v)
		}
	}
}

// The exemption is the throne authoring its own seat, not a blanket pass on the manifest: a
// non-throne body writing the throne seat file still fails right-author.
func TestGateThroneSeatStillPolicesNonThrone(t *testing.T) {
	dir := fixture(t)
	w := open(t, dir)
	opt := DefaultHooks().Gate
	opt.TraitsHold = false
	v := w.Gate(context.Background(), nil, opt, "slime-auth", []string{".isekai/dimension.yaml"}, &wire.Report{Status: "DONE"}, true, "x")
	if v.Word != "fail" || !strings.Contains(v.String(), "Right slime authored") {
		t.Errorf("a non-throne body writing the throne seat file must fail right-author: %+v", v)
	}
}
