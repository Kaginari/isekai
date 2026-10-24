package app

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kaginari/isekai/board"
	"github.com/Kaginari/isekai/onto"
)

// The law, embedded so `init` can found a world anywhere (a bench container, a fresh
// checkout). A test keeps it equal to the world's own copy (Vitality). Agent-one's law ships
// with the agent-one binary, from its own repository: this binary carries no twin to drift.
//
//go:embed law/isekai.md
var lawIsekai string

// LawText is the embedded law.
func LawText() string { return lawIsekai }

// Init seeds a minimal world under root: <dist-dir>/ with the law, an empty log, the
// instrument dirs and the memory tiers' folders, and a portraits/ with the rank PNGs
// (the tui shows them in the pane as the mascot). Idempotent: an existing file is never
// touched; nothing outside the world dir is written. It returns what it created.
func Init(dist, root string) ([]string, error) {
	d, err := parseDist(dist)
	if err != nil {
		return nil, err
	}
	if d.Name != "isekai" {
		return nil, fmt.Errorf("%s's law ships with the %s binary — found the %s with it", d.Name, d.Name, worldWord(d.Name))
	}
	base := filepath.Join(root, d.WorldDir)
	var made []string
	dirs := []string{"", "instruments", "instruments/loop", "instruments/usage", "instruments/desk", "instruments/toolbox", "memory", "memory/short", "memory/long", "memory/shared", "toolbox", "ontology", "ontology/graph", "tmp", "base-skills", "crafted-skills", "portraits"}
	for _, sub := range dirs {
		p := filepath.Join(base, sub)
		if _, err := os.Stat(p); err == nil {
			continue
		}
		if err := os.MkdirAll(p, 0o755); err != nil {
			return made, err
		}
		made = append(made, relOrAbs(root, p)+"/")
	}
	files := map[string]string{
		d.LawFile:                   LawText(),
		"log.md":                    "# Chronicle — change log\n\nAppend-only. Newest entries at the bottom. One entry per change.\n\n---\n",
		"name":                      filepath.Base(root) + "\n",
		"memory/shared/notes.jsonl": "",
		"ontology/schema.ttl":       onto.DefaultSchema,
		".gitignore":                "instruments/loop/\ninstruments/usage/\ninstruments/desk/\ninstruments/toolbox/\nmemory/short/\nmemory/long/\ntoolbox/registry.json\ntmp/\nconfig.local.*\n*.local.yaml\ninstruments/stream/\n",
	}
	for name, body := range files {
		p := filepath.Join(base, name)
		if _, err := os.Stat(p); err == nil {
			continue
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			return made, fmt.Errorf("%s: %w", relOrAbs(root, p), err)
		}
		made = append(made, relOrAbs(root, p))
	}
	// the rank portraits ride the board's embed; init stages them into the world's
	// portraits/ so the tui mascot can prefer the pixel source over the hand-drawn sprite,
	// world-local law: the elf owns portraits/ after init.
	for _, name := range board.PortraitNames() {
		p := filepath.Join(base, "portraits", name+".png")
		if _, err := os.Stat(p); err == nil {
			continue
		}
		b, err := board.Portrait(name)
		if err != nil {
			return made, fmt.Errorf("portrait %s: %w", name, err)
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			return made, fmt.Errorf("%s: %w", relOrAbs(root, p), err)
		}
		made = append(made, relOrAbs(root, p))
	}
	return made, nil
}
