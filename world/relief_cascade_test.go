package world

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeWorld(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for p, s := range files {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// A split walks the retiree's children: each one's Reports-to line moves to the young whose half
// holds its claims, the retiree names its heir, and every doc the act wrote is handed back.
func TestReliefCascadeSplit(t *testing.T) {
	dir := writeWorld(t, map[string]string{
		".isekai/isekai.md":          lawText,
		".isekai/log.md":             "# Chronicle\n",
		".isekai/elf/e/README.md":    "# elf-e\n\n- **Rank:** Elf\n- **Reports to:** rimuru\n",
		".isekai/orc/wide/README.md": "# orc-wide\n\n- **Rank:** Orc\n- **Territory:** `a/`, `b/`\n- **Reports to:** elf-e\n",
		".isekai/slime/a/README.md":  "# slime-a\n\n- **Rank:** Slime\n- **Territory:** `a/x/`\n- **Reports to:** orc-wide\n",
		".isekai/slime/b/README.md":  "# slime-b\n\n- **Rank:** Slime\n- **Territory:** `b/y/`\n- **Orc:** orc-wide\n",
		"a/x/a.go":                   "package x\n",
		"b/y/b.go":                   "package y\n",
	})
	w := open(t, dir)
	act, wrote, ok := w.relieve("orc-wide")
	if !ok || !strings.Contains(act, "split into orc-wide-a + orc-wide-b") || !strings.Contains(act, "re-seated slime-a, slime-b") {
		t.Fatalf("split: %v %q", ok, act)
	}
	for kid, heir := range map[string]string{"slime-a": "orc-wide-a", "slime-b": "orc-wide-b"} {
		if c := w.Creature(kid); c == nil || c.Parent != heir {
			t.Errorf("%s reports to %v, want %s", kid, c, heir)
		}
	}
	for _, doc := range []string{".isekai/orc/wide-a/README.md", ".isekai/orc/wide-b/README.md", ".isekai/orc/wide/README.md", ".isekai/slime/a/README.md", ".isekai/slime/b/README.md"} {
		if !contains(wrote, doc) {
			t.Errorf("%s missing from the act's writes %v", doc, wrote)
		}
	}
	if b, _ := os.ReadFile(filepath.Join(dir, ".isekai/orc/wide/README.md")); !strings.Contains(string(b), "- **Heir:** orc-wide-a") {
		t.Errorf("the retiree names no heir:\n%s", b)
	}
	if line := w.lineOf("slime-b"); strings.Join(line, ",") != "orc-wide-b,elf-e,rimuru" {
		t.Errorf("the living chain: %v", line)
	}
}

// A retirement re-seats the children one hop up; a child the cascade never reached still
// resolves through the retiree's Heir mark.
func TestReliefCascadeRetire(t *testing.T) {
	dir := writeWorld(t, map[string]string{
		".isekai/isekai.md":         lawText,
		".isekai/log.md":            "# Chronicle\n",
		".isekai/elf/e/README.md":   "# elf-e\n\n- **Rank:** Elf\n- **Reports to:** rimuru\n",
		".isekai/orc/mid/README.md": "# orc-mid\n\n- **Rank:** Orc\n- **Territory:** `zone/`\n- **Reports to:** elf-e\n",
		".isekai/slime/n/README.md": "# slime-n\n\n- **Rank:** Slime\n- **Territory:** `zone/n/`\n- **Reports to:** orc-mid\n",
		"zone/n/n.go":               "package n\n",
	})
	w := open(t, dir)
	act, wrote, ok := w.relieve("orc-mid")
	if !ok || !strings.Contains(act, "retired orc-mid") || !contains(wrote, ".isekai/slime/n/README.md") {
		t.Fatalf("retire: %v %q %v", ok, act, wrote)
	}
	if c := w.Creature("slime-n"); c == nil || c.Parent != "elf-e" {
		t.Fatalf("slime-n reports to %+v, want elf-e", c)
	}
	// a child born naming the retired seat (the cascade never saw it) resolves through the Heir
	os.MkdirAll(filepath.Join(dir, ".isekai/slime/late"), 0o755)
	os.WriteFile(filepath.Join(dir, ".isekai/slime/late/README.md"), []byte("# slime-late\n\n- **Rank:** Slime\n- **Reports to:** orc-mid\n"), 0o644)
	if err := w.Reload(); err != nil {
		t.Fatal(err)
	}
	if c := w.Creature("slime-late"); c == nil || c.Parent != "elf-e" {
		t.Fatalf("slime-late reports to %+v, want the heir elf-e", c)
	}
}
