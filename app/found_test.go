package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tree(t *testing.T, files ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, f := range files {
		p := filepath.Join(root, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestSketchColony(t *testing.T) {
	root := tree(t, "go.mod", "README.md", "main.go", ".github/workflows/ci.yml",
		"cmd/a/x.go", "cmd/a/y.go", "cmd/_b/z.go", "cmd/root.go",
		"lib/p.go", "lib/q.go", "lib/r.go",
		"tiny/one.go", "node_modules/m/i.js", ".isekai/log.md")
	d, _ := parseDist("isekai")
	s := sketchColony(root, d)
	if s.Lang != "go" || s.Elf.Short == "" {
		t.Fatalf("lang %q elf %q", s.Lang, s.Elf.Short)
	}
	names := map[string]plan{}
	for _, p := range s.all() {
		names[p.Rank+"-"+p.Short] = p
	}
	for _, want := range []string{"orc-cmd", "orc-lib", "orc-core", "slime-a", "slime-b", "slime-tiny", "slime-entry", "kijin-release"} {
		if _, ok := names[want]; !ok {
			t.Errorf("missing %s in %v", want, s.table(lexiconFor("isekai").Prefixes))
		}
	}
	if _, ok := names["orc-node-modules"]; ok {
		t.Error("node_modules is never a domain")
	}
	if v := names["slime-b"].Verify; len(v) != 1 || v[0] != "go test ./cmd/_b/" {
		t.Errorf("an underscore dir is tested by explicit path: %v", v)
	}
	if p := names["slime-tiny"]; p.Parent != "orc-core" {
		t.Errorf("a thin area merges into the core orc: %+v", p)
	}
}

func TestMediumWritesOnceAndFoundingBrief(t *testing.T) {
	root := tree(t, "go.mod", "svc/a/x.go", "svc/a/y.go", "svc/b/z.go")
	d, _ := parseDist("isekai")
	lex := lexiconFor("isekai")
	s := sketchColony(root, d)
	made, err := writeSketch(root, d, s, lex.Ranks, lex.Prefixes, lex.Display)
	if err != nil || len(made) == 0 {
		t.Fatalf("made %v, %v", made, err)
	}
	doc, _ := os.ReadFile(filepath.Join(root, ".isekai", "orc", "svc", "README.md"))
	for _, want := range []string{"# orc-svc", "- **Rank:** Orc", "- **Territory:** `svc/`", "- **Reports to:** elf-", "## Verify\n- `go test ./svc/...`", "## Thoughts"} {
		if !strings.Contains(string(doc), want) {
			t.Errorf("orc doc lacks %q:\n%s", want, doc)
		}
	}
	os.WriteFile(filepath.Join(root, ".isekai", "orc", "svc", "README.md"), []byte("# orc-svc\nmine\n"), 0o644)
	again, _ := writeSketch(root, d, s, lex.Ranks, lex.Prefixes, lex.Display)
	if len(again) != 0 {
		t.Fatalf("medium rewrote %v", again)
	}
	ask := foundingAsk(root, d, s, lex.Ranks, lex.Prefixes, lex.Display)
	if strings.Contains(ask, "{{") || !strings.Contains(ask, "orc-svc → elf-") || !strings.Contains(ask, "isekai onto check") {
		t.Fatalf("the founding brief is not filled in:\n%s", ask)
	}
	a1 := lexiconFor("agent-one")
	d1, _ := parseDist("agent-one")
	if ask1 := foundingAsk(root, d1, sketchColony(root, d1), a1.Ranks, a1.Prefixes, a1.Display); !strings.Contains(ask1, "domain-svc") || !strings.Contains(ask1, "agent-one onto check") || !strings.Contains(ask1, "**Rank:** Zone worker") {
		t.Fatalf("the brief speaks the distribution's words")
	}
}

func TestNormLevel(t *testing.T) {
	for in, want := range map[string]string{"": "light", "soft": "light", "Light": "light", "medium": "medium", "complex": "complex", "huge": ""} {
		if got := normLevel(in); got != want {
			t.Errorf("normLevel(%q) = %q, want %q", in, got, want)
		}
	}
}
