// Package tui is the live session's terminal UI (canon/tui.md): the conversation flows into the
// terminal's own scrollback as readable blocks, a bordered input sits at the bottom, and every
// event of the loop — a streamed answer, a tool step, a diff, a Court, the gate, an approval —
// is drawn, not printed raw. Built on the Charm libraries; the one place the binary takes
// third-party code.
//
// The package is layered by the ladder: blocks.go is the view model (pure functions from views
// to styled text), model.go the Bubble Tea program over it, and the app wires the loop's
// events into both.
package tui

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// Theme is the palette: the board's rank and lane tokens (board.css) carried to the terminal,
// a dark and a light variant, and nothing at all when the terminal has no colour (NO_COLOR,
// a pipe): lipgloss drops every escape on an ASCII profile, so a plain run is the same text.
type Theme struct {
	Dark  bool
	Color bool
	// Forced is true when the environment chose dark or light: the terminal's own answer
	// about its background is then not asked for.
	Forced bool

	dim, text, accent, user, ok, bad, warn, border, tag lipgloss.Style
	// the typographic system's two extra roles: label is the secondary heading (a card's
	// title, a key in a key-value grid) one step brighter than dim; magenta is the HUD's second
	// accent — the human's own marks (the ask's bar, the waiting chip), never a second accent
	// in a region that already wears the primary cyan
	label, magenta            lipgloss.Style
	classes                   map[string]lipgloss.Style
	ranks                     map[string]lipgloss.Style
	lanes                     map[string]color.Color
	add, del, ctxLine, lineNo lipgloss.Style
	sel                       lipgloss.Style // a drag selection in the chat pane
}

// Words are the labels a distribution renames (agent-one says Subagent, Operator, Workspace).
type Words struct {
	Court     string // "court" | "subagent"
	Courts    string // plural
	World     string // "world" | "workspace"
	Human     string // "the human" | "Operator"
	Gate      string // "gate"
	Dist      string // binary name
	ConfigRel string // where a don't-ask-again rule lands, world-relative
	// SessionName is the session body's name ("rimuru" | "orchestrator"): its events drive
	// the spinner, a Court's its block.
	SessionName string
}

// DefaultWords is the isekai vocabulary.
func DefaultWords() Words {
	return Words{Court: "court", Courts: "courts", World: "world", Human: "the human", Gate: "gate", Dist: "isekai", ConfigRel: ".isekai/config.local.yaml", SessionName: "rimuru"}
}

// NewTheme builds the palette for a dark or light background.
func NewTheme(dark bool) Theme {
	p := colorprofile.Detect(os.Stdout, os.Environ())
	t := Theme{Dark: dark, Color: p != colorprofile.Ascii && p != colorprofile.NoTTY}
	c := func(dark, light string) color.Color {
		if t.Dark {
			return lipgloss.Color(dark)
		}
		return lipgloss.Color(light)
	}
	// the palette, by role (the dashboard's Neon HUD carried to the cell grid): the ground is
	// the terminal's own; borders sit quieter than any text; text, then label, then dim carry
	// the hierarchy; cyan is the primary (brand, focus, the live state), magenta the human's
	// accent; ok/warn/bad say verdicts. Dark wears the neons softened to read as text; light
	// wears their deep-ink twins (every text role ≥ 4.5:1 on its ground).
	t.dim = lipgloss.NewStyle().Foreground(c("#7f8a99", "#5c6672"))
	t.label = lipgloss.NewStyle().Foreground(c("#b4bfcc", "#36404b"))
	t.text = lipgloss.NewStyle()
	t.accent = lipgloss.NewStyle().Foreground(c("#4fe3ee", "#00707c"))  // neon cyan, softened / deep teal
	t.magenta = lipgloss.NewStyle().Foreground(c("#ef6fe6", "#a0168f")) // neon magenta, softened / deep plum
	t.user = lipgloss.NewStyle().Foreground(c("#d5dde6", "#1f2730"))
	t.ok = lipgloss.NewStyle().Foreground(c("#5fcf86", "#1f7a3a"))
	t.bad = lipgloss.NewStyle().Foreground(c("#f0716a", "#b8341f"))
	t.warn = lipgloss.NewStyle().Foreground(c("#e8b955", "#8a5a00"))
	t.border = lipgloss.NewStyle().Foreground(c("#3d4a5a", "#b3bdc8"))
	t.tag = lipgloss.NewStyle().Bold(true)
	t.add = lipgloss.NewStyle().Foreground(c("#5fcf86", "#1f7a3a"))
	t.del = lipgloss.NewStyle().Foreground(c("#f0716a", "#b8341f"))
	t.ctxLine = lipgloss.NewStyle().Foreground(c("#8a94a3", "#5c6672"))
	t.lineNo = lipgloss.NewStyle().Foreground(c("#5b6675", "#8a94a3"))
	t.sel = lipgloss.NewStyle().Background(c("#1d4a5a", "#bfeef2")).Foreground(c("#eaffff", "#0b1424"))
	t.classes = map[string]lipgloss.Style{
		"read":        lipgloss.NewStyle().Foreground(c("#4db8dd", "#1a7fa3")),
		"write":       lipgloss.NewStyle().Foreground(c("#9a86e8", "#5c3fc9")),
		"outward":     lipgloss.NewStyle().Foreground(c("#e0b04f", "#8a6600")).Bold(true),
		"destructive": lipgloss.NewStyle().Foreground(c("#e06c5f", "#b8341f")).Bold(true),
	}
	t.ranks = map[string]lipgloss.Style{
		"slime":    lipgloss.NewStyle().Foreground(c("#2695bd", "#1a7fa3")),
		"orc":      lipgloss.NewStyle().Foreground(c("#7c5cd6", "#5c3fc9")),
		"elf":      lipgloss.NewStyle().Foreground(c("#bd8c24", "#8a6600")),
		"dark-elf": lipgloss.NewStyle().Foreground(c("#d6402a", "#b8341f")),
		"high-orc": lipgloss.NewStyle().Foreground(c("#8a7300", "#6b7400")),
		"kijin":    lipgloss.NewStyle().Foreground(c("#c53d34", "#a12e26")),
		"rimuru":   lipgloss.NewStyle().Foreground(c("#eaf2ff", "#0b1424")),
		// agent-one's ranks share the lanes
		"zone": lipgloss.NewStyle().Foreground(c("#2695bd", "#1a7fa3")), "domain": lipgloss.NewStyle().Foreground(c("#7c5cd6", "#5c3fc9")),
		"coord": lipgloss.NewStyle().Foreground(c("#bd8c24", "#8a6600")), "service": lipgloss.NewStyle().Foreground(c("#c53d34", "#a12e26")),
		"auditor": lipgloss.NewStyle().Foreground(c("#d6402a", "#b8341f")),
	}
	// the race pulse lanes (N16 slice 1): the colour a live Court row's motion wears — the
	// shimmer's crest, the tool spinner's glyph, the gate-wait bob, the landing flash. Distinct
	// from the rank token that colours the ◆ and the name: the lane says "alive", the token says
	// "who".
	t.lanes = map[string]color.Color{
		"rimuru":   c("#e8b545", "#b88b1a"), // golden
		"elf":      c("#e88030", "#b25c18"), // orange
		"orc":      c("#6fae5a", "#3f7f33"), // green
		"slime":    c("#58b7f0", "#1f78b8"), // blue
		"kijin":    c("#c53d34", "#a12e26"), // red
		"dark-elf": c("#9a86e8", "#5c3fc9"), // violet
		// agent-one's ranks on the same lanes
		"orchestrator": c("#e8b545", "#b88b1a"), "coord": c("#e88030", "#b25c18"),
		"domain": c("#6fae5a", "#3f7f33"), "zone": c("#58b7f0", "#1f78b8"),
		// the office hats ride lane colours too — the plan card and per-card titles wear them:
		// great-sage (findings/reads) is the slime-shared blue of the ground truth, raphael
		// (verdicts) the orc's gate greens, ciel (drafts) the elf's cross-domain orange
		"great-sage": c("#58b7f0", "#1f78b8"), "analyst": c("#58b7f0", "#1f78b8"),
		"raphael": c("#6fae5a", "#3f7f33"), "judge": c("#6fae5a", "#3f7f33"),
		"ciel": c("#e88030", "#b25c18"), "drafter": c("#e88030", "#b25c18"),
	}
	return t
}

// lane is a race's pulse colour; nil when the rank is unknown (callers fall back to the accent).
func (t Theme) lane(rank string) color.Color {
	if c, ok := t.lanes[strings.ToLower(rank)]; ok {
		return c
	}
	return nil
}

// Lane exposes lane() to the host (the rail plan card paints a deck in its author's race
// lane); the paper lane for "" is nil.
func (t Theme) Lane(rank string) color.Color { return t.lane(rank) }

// paneBg is the sidebar's background: one honest solid step off the session background —
// painted cell by cell, never a transparency story.
func (t Theme) paneBg() color.Color {
	if t.Dark {
		return lipgloss.Color("#12161d")
	}
	return lipgloss.Color("#eef1f4")
}

// paneArt re-wears the pane's background behind the mascot's foreground pixels: lipgloss'
// per-glyph full resets (\x1b[m / \x1b[0m) would otherwise drop the background mid-line, so
// they are rewritten into bg-restores.
func (t Theme) paneArt(art string) string {
	r, g, b, _ := t.paneBg().RGBA()
	sgr := fmt.Sprintf("48;2;%d;%d;%d", int(r>>8), int(g>>8), int(b>>8))
	art = strings.ReplaceAll(art, "\x1b[0m", "\x1b["+sgr+"m")
	return strings.ReplaceAll(art, "\x1b[m", "\x1b["+sgr+"m")
}

// onAccent is the face a fill in the primary wears: deep ink on the dark theme's neon, paper
// on the light theme's deep teal.
func (t Theme) onAccent() color.Color {
	if t.Dark {
		return lipgloss.Color("#06121a")
	}
	return lipgloss.Color("#ffffff")
}

// chip is a small tag worn in a hue: the hue as its face over a tint of itself sunk into the
// panel — lit, never a neon slab. A colourless terminal reads the same word in brackets-free
// plain text.
func (t Theme) chip(hue color.Color, text string) string {
	if hue == nil {
		hue = t.dim.GetForeground()
	}
	if !t.Color {
		return text
	}
	bg := lipgloss.Blend1D(5, t.paneBg(), hue)[1]
	fg := hue
	if !t.Dark {
		// a lane's light-theme hue is a mid tone: its chip face sinks a third toward ink so the
		// word holds 4.5:1 on its own tint
		fg = lipgloss.Blend1D(4, hue, lipgloss.Color("#000000"))[1]
	}
	return lipgloss.NewStyle().Foreground(fg).Background(bg).Padding(0, 1).Render(text)
}

// laneLegend is the welcome's foreshadow of the pulse lanes: each race a chip in its own hue —
// rimuru golden, elf orange, orc green, slime blue (agent-one's vocabulary follows its four
// mapped ranks).
func (t Theme) laneLegend(dist string) string {
	order := []string{"rimuru", "elf", "orc", "slime"}
	if dist == "agent-one" {
		order = []string{"orchestrator", "coord", "domain", "zone"}
	}
	var parts []string
	for _, r := range order {
		parts = append(parts, t.chip(t.lane(r), r))
	}
	return strings.Join(parts, " ")
}

// Brand is the distribution's name on its gradient, tuned to the background: isekai wears the
// HUD's cyan → magenta (deep-ink twins on a light ground); another distribution keeps its
// sprite's own stops.
func (t Theme) Brand(dist string) string {
	if !t.Color {
		return "✦ " + dist
	}
	if dist == "isekai" {
		stops := []color.Color{lipgloss.Color("#4fe3ee"), lipgloss.Color("#8fb4ff"), lipgloss.Color("#ef6fe6")}
		if !t.Dark {
			stops = []color.Color{lipgloss.Color("#00707c"), lipgloss.Color("#3f4fb0"), lipgloss.Color("#a0168f")}
		}
		return Gradient("✦ "+dist, true, stops...)
	}
	return Title(dist)
}

// DetectTheme picks dark unless the environment says the background is light (COLORFGBG's
// background code, or <PREFIX>_THEME=light); the terminal is never queried, so a start is never
// held by an unanswered escape.
func DetectTheme(env func(string) string, prefix string) Theme {
	if env == nil {
		env = os.Getenv
	}
	forced := func(dark bool) Theme {
		t := NewTheme(dark)
		t.Forced = true
		return t
	}
	if v := strings.ToLower(env(prefix + "THEME")); v == "light" {
		return forced(false)
	} else if v == "dark" {
		return forced(true)
	}
	if v := env("COLORFGBG"); v != "" {
		parts := strings.Split(v, ";")
		bg := parts[len(parts)-1]
		return forced(!(bg == "7" || bg == "15"))
	}
	return NewTheme(true) // until the terminal answers the background query
}

func (t Theme) class(name string) lipgloss.Style {
	if s, ok := t.classes[name]; ok {
		return s
	}
	return t.dim
}

func (t Theme) rank(name string) lipgloss.Style {
	if s, ok := t.ranks[strings.ToLower(name)]; ok {
		return s
	}
	return t.dim
}
