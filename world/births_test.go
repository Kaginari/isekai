package world

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kaginari/isekai/loop"
	"github.com/Kaginari/isekai/value"
	"github.com/Kaginari/isekai/wire"
)

// Every harness birth lands one line in the body's own births file: a relief's young, and a
// genesis that passed its gate — once, at the body's first.
func TestBirthsLedger(t *testing.T) {
	dir := writeWorld(t, map[string]string{
		".isekai/isekai.md":          lawText,
		".isekai/log.md":             "# Chronicle\n",
		".isekai/elf/e/README.md":    "# elf-e\n\n- **Rank:** Elf\n- **Reports to:** rimuru\n",
		".isekai/orc/wide/README.md": "# orc-wide\n\n- **Rank:** Orc\n- **Territory:** `a/`, `b/`\n- **Reports to:** elf-e\n",
		"a/a.go":                     "package a\n",
		"b/b.go":                     "package b\n",
	})
	w := open(t, dir)
	if _, _, ok := w.relieve("orc-wide"); !ok {
		t.Fatal("no relief")
	}
	for _, y := range []string{"orc-wide-a", "orc-wide-b"} {
		bs := w.Births(y)
		if len(bs) != 1 || bs[0].Kind != "split" || bs[0].From != "orc-wide" || bs[0].Doc == "" {
			t.Errorf("%s births: %+v", y, bs)
		}
	}
	// a genesis that passes is logged — once, at the body's first, however often it passes
	gdir, gw := genesisWorld(t)
	doc := ".isekai/slime/pay/README.md"
	os.MkdirAll(filepath.Join(gdir, ".isekai/slime/pay"), 0o755)
	os.WriteFile(filepath.Join(gdir, doc), []byte("# slime-pay\n\n- **Rank:** Slime\n- **Territory:** `src/pay/`\n- **Reports to:** orc-security\n- **Verify:** `true`\n"), 0o644)
	gate := gw.EndGate(DefaultHooks().Gate)
	for i := 0; i < 2; i++ {
		s := &loop.Session{Engine: &loop.Engine{As: "orc-security"}, Wrote: []string{doc}}
		if _, _, err := gate(context.Background(), s, &loop.Result{Report: wire.Report{Status: "DONE"}, IsWire: true}); err != nil {
			t.Fatalf("genesis gate %d: %v", i, err)
		}
	}
	if bs := gw.Births("slime-pay"); len(bs) != 1 || bs[0].Kind != "genesis" || bs[0].By != "orc-security" {
		t.Fatalf("slime-pay births: %+v", bs)
	}
}

// Tenure is whole days since the first birth; a body with no births line is grandparented by
// the first commit of its doc; a body neither speaks for reads zero.
func TestTenureSignal(t *testing.T) {
	dir := fixture(t)
	git(t, dir, "init", "-b", "main")
	git(t, dir, "add", "-A")
	old := time.Now().Add(-40 * 24 * time.Hour).Format(time.RFC3339)
	c := exec.Command("git", "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "world")
	c.Dir, c.Env = dir, append(os.Environ(), "GIT_AUTHOR_DATE="+old, "GIT_COMMITTER_DATE="+old)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("commit: %v %s", err, out)
	}
	w := open(t, dir)
	w.logBirth(birthRecord{At: time.Now().Add(-10*24*time.Hour - time.Hour), Body: "slime-auth", Kind: "genesis"})
	w.val.mu.Lock()
	for _, b := range []string{"slime-auth", "slime-api", "slime-ghost"} {
		w.appendValue(value.Row{Kind: value.KindRun, Body: b, Status: "DONE", Word: "DONE"})
	}
	w.val.mu.Unlock()
	all := w.AllSignals()
	for body, want := range map[string]float64{"slime-auth": 10, "slime-api": 40, "slime-ghost": 0} {
		if got := all[body]["tenure"]; got != want {
			t.Errorf("%s tenure %v, want %v", body, got, want)
		}
	}
}
