package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// /forge is the desk: it lists what courts carried home (@FORGE → the colony ledger) and mints
// by the desk's yes — a skill lands as a world note, a tool as a tools.custom patch that
// reloads live. Minted drafts mark themselves; nothing mints twice.
func TestForgeDeskListsAndMints(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.script(".isekai/tmp/s.json", text("nothing"))
	w.write(".isekai/config.yaml", mockCfg(".isekai/tmp/s.json", ""))
	a := w.open()
	// the colony ledger, as filed by record beats
	notes := `{"at":"2026-09-28T21:00:00Z","by":"slime-auth","world":"x","tag":"forge","kind":"colony","text":"tool:ui-wrangler | rewrite bare /v1 strings | ./wrangle {{path}}"}
{"at":"2026-09-28T21:01:00Z","by":"slime-auth","world":"x","tag":"forge","kind":"colony","text":"skill:ttl-craft | one hour is the shelf life"}
{"at":"2026-09-28T21:02:00Z","by":"slime-auth","world":"x","tag":"forge-minted","kind":"colony","text":"skill:ttl-craft"}
`
	path := filepath.Join(w.root, ".isekai", "memory", "shared")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "notes.jsonl"), []byte(notes), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errw bytes.Buffer
	a.forgeCmd(&out, &errw, "")
	if s := out.String(); !strings.Contains(s, "1. slime-auth  tool:ui-wrangler") || strings.Contains(s, "ttl-craft") {
		t.Fatalf("the desk lists only the unminted:\n%s", s)
	}
	out.Reset()
	// a tool mint patches tools.custom and marks the draft
	a.forgeCmd(&out, &errw, "mint 1")
	if !strings.Contains(out.String(), "minted tool:ui-wrangler") {
		t.Fatalf("mint: %s / err: %s", out.String(), errw.String())
	}
	y, err := os.ReadFile(filepath.Join(w.root, ".isekai", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if c := string(y); !strings.Contains(c, "ui-wrangler") || !strings.Contains(c, "- ./wrangle") || !strings.Contains(c, `"{{path}}"`) || !strings.Contains(c, "path:") || !strings.Contains(c, "crafted from slime-auth's run") {
		t.Fatalf("the mint's config patch:\n%s", c)
	}
	out.Reset()
	a.forgeCmd(&out, &errw, "")
	if s := out.String(); !strings.Contains(s, "the forge ledger is quiet") {
		t.Fatalf("minted things leave the list:\n%s", s)
	}
	// a skill draft mints as a note riding kind colony
	out.Reset()
	w2 := newTestWorld(t, "isekai", isekaiCreatures())
	w2.script(".isekai/tmp/s.json", text("nothing"))
	w2.write(".isekai/config.yaml", mockCfg(".isekai/tmp/s.json", ""))
	a2 := w2.open()
	if err := os.MkdirAll(filepath.Join(w2.root, ".isekai", "memory", "shared"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(w2.root, ".isekai", "memory", "shared", "notes.jsonl"), []byte(`{"at":"2026-09-28T21:00:00Z","by":"slime-auth","world":"x","tag":"forge","kind":"colony","text":"skill:ttl-craft | one hour is the shelf life"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	a2.forgeCmd(&out, &errw, "mint 1")
	if !strings.Contains(out.String(), "minted skill:ttl-craft") {
		t.Fatalf("skill mint: %s / err: %s", out.String(), errw.String())
	}
	b, err2 := os.ReadFile(filepath.Join(w2.root, ".isekai", "memory", "shared", "notes.jsonl"))
	if err2 != nil {
		t.Fatal(err2)
	}
	if got := string(b); !strings.Contains(got, `"tag":"skill"`) || !strings.Contains(got, `"tag":"forge-minted"`) {
		t.Fatalf("the skill lands as a note, marked minted:\n%s", got)
	}
}
