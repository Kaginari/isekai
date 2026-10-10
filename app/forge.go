package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Kaginari/isekai/config"
	"github.com/Kaginari/isekai/wire"
)

// --- the desk's forge ---
//
// Drafts ride every Court's report as `@FORGE <kind>:<name> | <summary> | <argv>`, and the
// record beat files them home as shared notes of kind "forge" — never minted by the wire. The
// desk (this command) is where the curation meets law: list, then mint by index. A minted
// skill lands as a world note (the anchors it names ride recall); a minted tool lands as a
// config patch on tools.custom (the Genesis ceremony owns the file). Minted drafts are marked
// append-only: a "forge-minted" note names them, never a line struck.

type noteLine struct {
	By   string `json:"by"`
	Kind string `json:"kind"`
	Text string `json:"text"`
	Tag  string `json:"tag"`
}

func (a *App) notesPath() string {
	return filepath.Join(a.Root, a.Cfg.Dist.WorldDir, "memory", "shared", "notes.jsonl")
}

// forgeDrafts reads the forge ledger: notes tagged forge, minus any a forge-minted note marked.
// Two passes: the marks are append-only anywhere in the ledger, so they must gather first.
func (a *App) forgeDrafts() ([]wire.Forge, []string, error) {
	b, err := os.ReadFile(a.notesPath())
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var lines []noteLine
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var n noteLine
		if json.Unmarshal([]byte(line), &n) != nil {
			continue
		}
		lines = append(lines, n)
	}
	minted := map[string]bool{}
	for _, n := range lines {
		if n.Tag == "forge-minted" {
			minted[strings.TrimSpace(n.Text)] = true
		}
	}
	var drafts []wire.Forge
	var bodies []string
	for _, n := range lines {
		if n.Tag != "forge" {
			continue
		}
		f, ok := wireForze(n.Text)
		if !ok {
			f = wire.Forge{Kind: "skill", Name: firstWord(n.Text), Text: n.Text}
		}
		if !minted[f.Kind+":"+f.Name] {
			drafts = append(drafts, f)
			bodies = append(bodies, n.By)
		}
	}
	return drafts, bodies, nil
}

// wireForze re-parses a filed forge note back into a Forge.
func wireForze(text string) (wire.Forge, bool) {
	i := strings.Index(text, ":")
	if i <= 0 {
		return wire.Forge{}, false
	}
	kind := strings.ToLower(strings.TrimSpace(text[:i]))
	if kind != "tool" && kind != "skill" {
		return wire.Forge{}, false
	}
	payload := strings.TrimSpace(text[i+1:])
	name := payload
	rest := ""
	if j := strings.Index(payload, "|"); j >= 0 {
		name, rest = strings.TrimSpace(payload[:j]), strings.TrimSpace(payload[j+1:])
	}
	if name == "" {
		return wire.Forge{}, false
	}
	return wire.Forge{Kind: kind, Name: name, Text: rest}, true
}

// forgeCmd is /forge [mint <n>].
func (a *App) forgeCmd(out, errw io.Writer, rest string) {
	drafts, bodies, err := a.forgeDrafts()
	if err != nil {
		fmt.Fprintln(errw, "forge: "+err.Error())
		return
	}
	if rest == "" {
		if len(drafts) == 0 {
			fmt.Fprintln(out, "the forge ledger is quiet — courts carry drafts home as @FORGE lines")
			return
		}
		for i, f := range drafts {
			fmt.Fprintf(out, "%d. %s  %s:%s", i+1, bodies[i], f.Kind, f.Name)
			if f.Text != "" {
				fmt.Fprintf(out, "  — %s", oneLineText(f.Text))
			}
			fmt.Fprintln(out)
		}
		fmt.Fprintln(out, "/forge mint <n> — a skill lands as a world note; a tool lands in tools.custom")
		return
	}
	n, ok := 0, false
	if strings.HasPrefix(rest, "mint ") {
		_, serr := fmt.Sscanf(strings.TrimPrefix(rest, "mint "), "%d", &n)
		ok = serr == nil && n >= 1 && n <= len(drafts)
	}
	if !ok {
		fmt.Fprintln(errw, "usage: /forge mint <n> (1.."+itoaI(len(drafts))+")")
		return
	}
	f := drafts[n-1]
	if err := a.forgeMint(bodies[n-1], f); err != nil {
		fmt.Fprintln(errw, "forge mint: "+err.Error())
		return
	}
	fmt.Fprintf(out, "minted %s:%s\n", f.Kind, f.Name)
}

// forgeMint lands a draft. A skill writes a world note whose anchors ride recall; a tool
// patches tools.custom.<name> with its run template and reloads live (the Genesis file law).
func (a *App) forgeMint(by string, f wire.Forge) error {
	if f.Kind == "skill" {
		if err := a.World.Remember(by, "colony", f.Name+" — "+f.Text, "skill"); err != nil {
			return err
		}
		return a.World.Remember(by, "colony", f.Kind+":"+f.Name, "forge-minted")
	}
	// a tool draft: "<summary> | <argv template>"
	if _, exists := a.Cfg.Tools.Custom[f.Name]; exists {
		return fmt.Errorf("tools.custom.%s already declares %s — refuse to overwrite the desk's law; rename the draft", f.Name, f.Name)
	}
	summary, argv := f.Text, ""
	if i := strings.Index(f.Text, "|"); i >= 0 {
		summary, argv = strings.TrimSpace(f.Text[:i]), strings.TrimSpace(f.Text[i+1:])
	}
	if argv == "" {
		return fmt.Errorf("a tool draft needs an argv template: @FORGE tool:%s | <summary> | <argv>", f.Name)
	}
	var args []string
	params := map[string]bool{}
	for _, piece := range strings.Fields(argv) {
		args = append(args, yamlQuote(piece))
		for _, hole := range forgeParams(piece) {
			params[hole] = true
		}
	}
	parts := []string{
		"description: " + yamlQuote(summary+" — crafted from "+by+"'s run"),
		"run: [" + strings.Join(args, ", ") + "]",
		"class: write",
	}
	if len(params) > 0 {
		var ps []string
		for p := range params {
			ps = append(ps, p+": {type: string, required: true}")
		}
		sort.Strings(ps)
		parts = append(parts, "params: {"+strings.Join(ps, ", ")+"}")
	}
	file := a.Cfg.ProjectFile()
	ops := []config.Op{{Path: "tools.custom." + f.Name, Value: "{" + strings.Join(parts, ", ") + "}"}}
	if err := a.Cfg.ApplyPatch(file, ops); err != nil {
		return fmt.Errorf("the mint failed: %w", err)
	}
	next, changes, err := a.Cfg.Reload()
	if err != nil {
		return fmt.Errorf("written, but the reload failed: %w", err)
	}
	a.reload(next, changes)
	a.Missing.SetConfig(next)
	if err := a.World.Remember(by, "colony", f.Kind+":"+f.Name, "forge-minted"); err != nil {
		a.hole("forge-minted note: " + err.Error())
	}
	return nil
}

// forgeParams collects the {{name}} holes of one argv piece (each becomes a declared param).
func forgeParams(piece string) []string {
	var out []string
	rest := piece
	for {
		i := strings.Index(rest, "{{")
		if i < 0 {
			return out
		}
		rest = rest[i+2:]
		j := strings.Index(rest, "}}")
		if j <= 0 {
			return out
		}
		name := strings.TrimSpace(rest[:j])
		valid := name != ""
		for _, r := range name {
			if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
				valid = false
				break
			}
		}
		if valid {
			out = append(out, name)
		}
		rest = rest[j+2:]
	}
}

func firstWord(s string) string {
	f := strings.Fields(s)
	if len(f) == 0 {
		return ""
	}
	return f[0]
}

func itoaI(n int) string {
	if n == 0 {
		return "0"
	}
	out := ""
	for n > 0 {
		out = string(rune('0'+n%10)) + out
		n /= 10
	}
	return out
}
