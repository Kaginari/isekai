package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderInitStyledText(t *testing.T) {
	var buf bytes.Buffer
	rep := initReport{
		dist:  "isekai",
		root:  "/tmp/demo",
		level: "medium",
		made:  []string{".isekai/ISEKAI.md"},
		medium: &mediumRep{
			members: []plan{
				{Rank: "elf", Short: "core", Parent: "rimuru", Territory: []string{"go.mod"}},
				{Rank: "orc", Short: "app", Parent: "elf-core", Territory: []string{"app/"}},
				{Rank: "slime", Short: "app", Parent: "orc-app", Territory: []string{"app/"}, Verify: []string{"go test ./app/..."}},
				{Rank: "slime", Short: "cli", Parent: "orc-app", Territory: []string{"cli/"}},
				{Rank: "kijin", Short: "ci", Parent: "rimuru", Territory: []string{".github/"}},
			},
			wrote: []string{".isekai/elf/core/README.md"},
			kept:  1,
			elves: 1, orcs: 1, slimes: 2, kijin: 1,
			prefix: map[string]string{"elf": "elf-", "orc": "orc-", "slime": "slime-", "kijin": "kijin-"},
		},
	}
	renderInit(&buf, rep, defaultInitWords())
	out := buf.String()
	for _, want := range []string{"isekai", "scaffold", "ISEKAI.md", "colony", "1 written", "1 kept", "1 elf · 1 orc · 2 slimes · 1 kijin",
		"    elf-core", "      orc-app", "        slime-app", "slime-cli", "cli/", "(no tests seen)", "    kijin-ci", "next", "run \"<ask>\""} {
		if !strings.Contains(out, want) {
			t.Errorf("styled init output missing %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "@S") || strings.Contains(out, "@F") {
		t.Errorf("styled output must not leak wire tags, got:\n%s", out)
	}
}

func TestStyledInitGate(t *testing.T) {
	var buf bytes.Buffer
	if styledInit(&buf, false) {
		t.Error("a buffer is not a terminal — styled gate must close")
	}
	if styledInit(nil, true) {
		t.Error("--plain closes the styled face")
	}
}

func TestPadTo(t *testing.T) {
	if got := padTo("zone", 8); got != "zone    " {
		t.Errorf("padTo: got %q", got)
	}
	if got := padTo("orchestrator", 4); got != "orchestrator" {
		t.Errorf("padTo must not truncate: got %q", got)
	}
}
