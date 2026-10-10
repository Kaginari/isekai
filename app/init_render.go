package app

import (
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// initReport is `init`'s whole answer, gathered once: a terminal renders it styled
// (renderInit), pipes and bench read the wire form printed by cmdInit — byte-stable,
// tests and scripts depend on it.
type initReport struct {
	dist, root, level string
	made              []string // scaffold writes from Init
	setupPath         string   // config written by the setup step ("" when none/skipped)
	boardPort         int      // the board port the written config chose (0 when nothing written)
	medium            *mediumRep
	bench             bool
	housekept         int // chain-law repairs sealed at found (0 = whole at birth)
	leftover          int // findings still wanting an operator
}

type mediumRep struct {
	members                    []plan
	wrote                      []string
	kept, covered              int
	elves, orcs, slimes, kijin int
	lanes                      []string          // "elf-x: orc-a (3), orc-b (2)"
	prefix                     map[string]string // the lexicon's name prefixes ("" map: canonical)
}

// census is the colony's head count, spelled out: "1 elf · 3 orcs · 9 slimes · 1 kijin".
func (m *mediumRep) census() string {
	part := func(n int, one, many string) string {
		if n == 1 {
			return fmt.Sprintf("%d %s", n, one)
		}
		return fmt.Sprintf("%d %s", n, many)
	}
	out := []string{part(m.elves, "elf", "elves"), part(m.orcs, "orc", "orcs"), part(m.slimes, "slime", "slimes")}
	if m.kijin > 0 {
		out = append(out, part(m.kijin, "kijin", "kijin"))
	}
	return strings.Join(out, " · ")
}

// name is a member's full creature id under the lexicon's prefixes.
func (m *mediumRep) name(p plan) string {
	if pre, ok := m.prefix[p.Rank]; ok {
		return pre + p.Short
	}
	return p.Rank + "-" + p.Short
}

// initWords is the world's voice for the init face.
type initWords struct {
	header       string
	scaffoldNote string
	teamWord     string
	keptWord     string
	complexNote  string
	footer       string
}

func defaultInitWords() initWords {
	return initWords{
		header:       "founding a world at",
		scaffoldNote: "the law · the instruments · memory tiers",
		teamWord:     "a colony sketched from the tree",
		keptWord:     "kept as they were",
		complexNote:  "the founding session starts: the model reads all the code and founds the colony",
		footer:       "human gate on · guts on",
	}
}

// styledInit answers whether Out deserves the formatted face: a real terminal with colour.
// The same check the tui's palette uses (tui/theme.go) — lipgloss itself strips the escapes
// on an ASCII profile, but the wire form is the better pipe answer, so pipes never render.
func styledInit(w io.Writer, plain bool) bool {
	if plain {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	p := colorprofile.Detect(f, os.Environ())
	return p != colorprofile.Ascii && p != colorprofile.NoTTY
}

// initStyles mirrors the tui dark palette (the board tokens in tui/theme.go). A light-bg
// variant comes when the tui theme exposes one for commands.
type initStyles struct {
	title, section, okGlow, dim, count, file lipgloss.Style
	rank                                     map[string]lipgloss.Style
}

func newInitStyles() initStyles {
	c := lipgloss.Color
	return initStyles{
		title:   lipgloss.NewStyle().Foreground(c("#d4af37")).Bold(true),
		section: lipgloss.NewStyle().Foreground(c("#9fb0c3")).Bold(true),
		okGlow:  lipgloss.NewStyle().Foreground(c("#5fbf7a")).Bold(true),
		dim:     lipgloss.NewStyle().Foreground(c("#7f8a99")),
		count:   lipgloss.NewStyle().Foreground(c("#bd8c24")),
		file:    lipgloss.NewStyle().Foreground(c("#9fb0c3")),
		rank: map[string]lipgloss.Style{
			"elf":     lipgloss.NewStyle().Foreground(c("#bd8c24")),
			"orc":     lipgloss.NewStyle().Foreground(c("#7c5cd6")),
			"slime":   lipgloss.NewStyle().Foreground(c("#2695bd")),
			"kijin":   lipgloss.NewStyle().Foreground(c("#c53d34")),
			"coord":   lipgloss.NewStyle().Foreground(c("#bd8c24")),
			"domain":  lipgloss.NewStyle().Foreground(c("#7c5cd6")),
			"zone":    lipgloss.NewStyle().Foreground(c("#2695bd")),
			"service": lipgloss.NewStyle().Foreground(c("#c53d34")),
		},
	}
}

// renderInit draws the init answer for a terminal: sectioned, spelled out, no wire tags.
func renderInit(w io.Writer, r initReport, words initWords) {
	st := newInitStyles()
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n\n", st.title.Render("◆ "+r.dist), st.dim.Render("· "+words.header+" "+r.root))
	writeInitSection(&b, st, "scaffold", words.scaffoldNote, r.made, len(r.made), "already founded — nothing touched")
	if r.setupPath != "" {
		writeInitSection(&b, st, "setup", fmt.Sprintf("config written · board on http://127.0.0.1:%d", r.boardPort), []string{r.setupPath}, 1, "")
	} else {
		b.WriteString(st.dim.Render("  board → http://127.0.0.1:7411 · the schema default (no world config written; set ui.board.port to move it)") + "\n")
	}
	if r.medium != nil {
		m := r.medium
		head := st.okGlow.Render("  ✓ ") + st.section.Render(words.teamWord) + st.dim.Render("  ·  ") +
			st.count.Render(fmt.Sprintf("%d written", len(m.wrote)))
		if m.kept > 0 {
			head += st.dim.Render(fmt.Sprintf(" · %d %s", m.kept, words.keptWord))
		}
		if m.covered > 0 {
			head += st.dim.Render(fmt.Sprintf(" · %d already held by the world's creatures", m.covered))
		}
		b.WriteString(head + "\n")
		if m.elves+m.orcs+m.slimes > 0 {
			b.WriteString(st.dim.Render("    "+m.census()) + "\n")
		}
		// the tree: each creature under the one it reports to, rimuru's children at the top
		kids := map[string][]plan{}
		for _, p := range m.members {
			kids[p.Parent] = append(kids[p.Parent], p)
		}
		var draw func(parent string, depth int)
		draw = func(parent string, depth int) {
			for _, p := range kids[parent] {
				rank := st.rank[p.Rank]
				line := rank.Render(strings.Repeat("  ", depth+2) + m.name(p))
				if p.Rank == "slime" || p.Rank == "kijin" || (p.Rank == "elf" && len(kids[m.name(p)]) == 0) {
					line += st.dim.Render("  · " + strings.Join(p.Territory, " "))
				}
				if p.Rank == "slime" && len(p.Verify) == 0 {
					line += st.dim.Render("  (no tests seen)")
				}
				b.WriteString(line + "\n")
				if depth < 4 {
					draw(m.name(p), depth+1)
				}
			}
		}
		draw("rimuru", 0)
	}
	if r.level == "complex" {
		b.WriteString(st.okGlow.Render("  ✦ ") + st.section.Render("founding session") + st.dim.Render("  "+words.complexNote) + "\n")
	}
	if r.housekept > 0 || r.leftover > 0 {
		b.WriteString(st.okGlow.Render("  ✓ ") + st.section.Render("housekeep") + st.dim.Render(fmt.Sprintf("  · %d sealed by the chain law · %d want an operator", r.housekept, r.leftover)) + "\n")
	}
	if r.bench {
		b.WriteString(st.dim.Render("  · bench: config from "+strings.ToUpper(strings.ReplaceAll(r.dist, "-", "_"))+"_CONFIG_CONTENT; the gate needs a TTY or a file-layer pre-approval") + "\n")
	}
	fmt.Fprintf(&b, "\n%s\n  %s %s\n", st.dim.Render("── next ──────────────────────────"), st.section.Render(r.dist+" run \"<ask>\""), st.dim.Render("· "+words.footer))
	fmt.Fprint(w, b.String())
}

func writeInitSection(b *strings.Builder, st initStyles, name, note string, files []string, n int, emptyNote string) {
	fmt.Fprintf(b, "%s%s", st.okGlow.Render("  ✓ "), st.section.Render(name))
	if note != "" {
		b.WriteString(st.dim.Render("  ·  " + note))
	}
	if n > 0 {
		b.WriteString(st.count.Render(fmt.Sprintf("  %d written", n)) + "\n")
		for _, m := range files {
			b.WriteString(st.file.Render("    + "+m) + "\n")
		}
	} else if emptyNote != "" {
		b.WriteString(st.dim.Render("  ·  "+emptyNote) + "\n")
	} else {
		b.WriteString("\n")
	}
}

// padTo right-pads a word with spaces to at least n columns (ASCII-only content here).
func padTo(s string, n int) string {
	for len(s) < n {
		s += " "
	}
	return s
}
