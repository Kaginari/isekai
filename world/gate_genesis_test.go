package world

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/wire"
)

// genesisWorld is the fixture under git, committed, with an unclaimed src/pay/ on the tree.
func genesisWorld(t *testing.T) (string, *World) {
	t.Helper()
	dir := fixture(t)
	if err := os.MkdirAll(filepath.Join(dir, "src/pay"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src/pay/pay.go"), []byte("package pay\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "init", "-b", "main")
	git(t, dir, "add", "-A")
	git(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "world")
	return dir, open(t, dir)
}

// A zone-parent may mint its own child's README: new since the last commit, reporting to the
// writer, claiming live soil that is the writer's or nobody's. Each condition is load-bearing.
func TestGateGenesisExemption(t *testing.T) {
	dir, w := genesisWorld(t)
	g := DefaultHooks().Gate
	rep := &wire.Report{Status: "DONE"}
	mint := func(as, rel, doc string) Verdict {
		t.Helper()
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(doc), 0o644); err != nil {
			t.Fatal(err)
		}
		v := w.Gate(context.Background(), nil, g, as, []string{rel}, rep, true, "x")
		os.RemoveAll(filepath.Dir(full))
		_ = w.Reload()
		return v
	}
	child := func(name, parent, terr string) string {
		return "# " + name + "\n\n- **Rank:** Slime\n- **Territory:** `" + terr + "`\n- **Reports to:** " + parent + "\n- **Verify:** `true`\n"
	}
	if v := mint("orc-security", ".isekai/slime/pay/README.md", child("slime-pay", "orc-security", "src/pay/")); v.Word != "pass" {
		t.Errorf("the parent's genesis of its child: %s", v)
	}
	if v := mint(Rimuru, ".isekai/elf/ops/README.md", "# elf-ops\n\n- **Rank:** Elf\n- **Reports to:** rimuru\n"); v.Word != "pass" {
		t.Errorf("the throne founding an elf: %s", v)
	}
	for why, c := range map[string][2]string{
		"another parent's child":   {".isekai/slime/pay/README.md", child("slime-pay", "orc-other", "src/pay/")},
		"a third body's soil":      {".isekai/slime/pay/README.md", child("slime-pay", "orc-security", "src/auth/")},
		"a claim on nothing":       {".isekai/slime/pay/README.md", child("slime-pay", "orc-security", "src/ghost/")},
		"a claim over a third one": {".isekai/slime/pay/README.md", child("slime-pay", "orc-security", "src/")},
	} {
		if v := mint("orc-security", c[0], c[1]); v.Word != "fail" || !strings.Contains(v.String(), "Right slime authored") {
			t.Errorf("%s passed as genesis: %s", why, v)
		}
	}
	// one commit window: once the child's doc is committed, its re-authoring is the child's
	rel := ".isekai/slime/pay/README.md"
	full := filepath.Join(dir, rel)
	os.MkdirAll(filepath.Dir(full), 0o755)
	os.WriteFile(full, []byte(child("slime-pay", "orc-security", "src/pay/")), 0o644)
	git(t, dir, "add", "-A")
	git(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "born")
	os.WriteFile(full, []byte(child("slime-pay", "orc-security", "src/pay/")+"- more\n"), 0o644)
	_ = w.Reload()
	if v := w.Gate(context.Background(), nil, g, "orc-security", []string{rel}, rep, true, "x"); v.Word != "fail" {
		t.Errorf("genesis outlived its commit window: %s", v)
	}
}
