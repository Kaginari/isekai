package app

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Kaginari/isekai/config"
)

// The three ways to found a world (`init --level`):
//   - light (soft): the law, the log, the instruments and the memory tiers — no creature.
//   - medium: light, and a colony sketched from the tree with no model, sized to the project:
//     a slime per meaningful package or module dir, orcs grouping related slimes by imports,
//     role and naming, an elf per three orcs (lane elves and a canon elf on a large project),
//     a kijin when there is CI, release or infra; verify lines only where tests exist.
//   - complex: light, then a founding session — the model reads all the code and decides the
//     team: how many elves, orcs, slimes, kijin, which minds and which bodies, the relations,
//     verify lines that pass; the sketch is its first hint. Like hiring a team of experts.

// Levels are the init levels, in order.
var Levels = []string{"light", "medium", "complex"}

// normLevel reads a level name ("soft" is light); "" when unknown.
func normLevel(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "light", "soft":
		return "light"
	case "medium", "mid":
		return "medium"
	case "complex", "full", "deep":
		return "complex"
	}
	return ""
}

// sketch is a colony read from the tree alone (found_survey.go reads, found_cut.go cuts).
type sketch struct {
	Project string
	Lang    string
	Elves   []plan // the canon elf (or the one elf) first, then the lane elves
	Orcs    []plan // each carries its slimes
	Kijin   *plan
}

type plan struct {
	Rank, Short, Purpose string
	Territory            []string
	Parent               string
	Minds                []string
	Traits               []string // the sketch's reading, each marked "sketch:"
	Verify               []string
	Slimes               []plan
	Orcs                 []string // an elf's carried creatures, by name
}

var reName = regexp.MustCompile(`[^a-z0-9]+`)

func slug(s string) string {
	return strings.Trim(reName.ReplaceAllString(strings.ToLower(strings.TrimLeft(s, "_.@")), "-"), "-")
}

// sketchColony reads the tree into a colony, without a model.
func sketchColony(root string, d config.Dist) sketch {
	project := slug(filepath.Base(root))
	if project == "" {
		project = "core"
	}
	lex := lexiconFor(d.Name)
	return readSurvey(root, d.WorldDir).cut(project, lex.Prefixes, readMinds(root, d.WorldDir))
}

// doc is a creature's README in the shape the ontology reads.
func (p plan) doc(prefix, display map[string]string, level string) string {
	var b strings.Builder
	name := prefix[p.Rank] + p.Short
	fmt.Fprintf(&b, "# %s\n\n", name)
	rank := display[p.Rank]
	if rank == "" {
		rank = strings.ToUpper(p.Rank[:1]) + p.Rank[1:]
	}
	fmt.Fprintf(&b, "- **Rank:** %s\n", rank)
	if len(p.Territory) > 0 {
		t := make([]string, len(p.Territory))
		for i, x := range p.Territory {
			t[i] = "`" + x + "`"
		}
		fmt.Fprintf(&b, "- **Territory:** %s\n", strings.Join(t, ", "))
	}
	fmt.Fprintf(&b, "- **Reports to:** %s\n", p.Parent)
	if len(p.Minds) > 0 {
		fmt.Fprintf(&b, "- **Minds:** %s\n", strings.Join(p.Minds, ", "))
	} else {
		b.WriteString("- **Minds:** none\n")
	}
	fmt.Fprintf(&b, "- **Purpose:** %s\n", p.Purpose)
	fmt.Fprintf(&b, "- **Founded:** `init --level %s`, from the tree alone — the traits marked sketch are the founding sketch's reading; refine them and the verify lines from the code\n", level)
	b.WriteString("\n## Traits\n")
	if len(p.Traits) == 0 {
		b.WriteString("- (none recorded yet: the first session that works here writes what it learns)\n")
	}
	for _, t := range p.Traits {
		fmt.Fprintf(&b, "- %s\n", t)
	}
	if len(p.Verify) > 0 {
		b.WriteString("\n## Verify\n")
		for _, v := range p.Verify {
			fmt.Fprintf(&b, "- `%s`\n", v)
		}
	}
	b.WriteString("\n## Thoughts\n")
	return b.String()
}

// all is every creature of the sketch, parents first.
func (s sketch) all() []plan {
	out := append([]plan(nil), s.Elves...)
	out = append(out, s.Orcs...)
	for _, o := range s.Orcs {
		out = append(out, o.Slimes...)
	}
	if s.Kijin != nil {
		out = append(out, *s.Kijin)
	}
	return out
}

// count is the sketch's census by rank.
func (s sketch) count(rank string) int {
	n := 0
	for _, p := range s.all() {
		if p.Rank == rank {
			n++
		}
	}
	return n
}

// lanes names each elf with the orcs it carries — the report's summary.
func (s sketch) lanes(prefix map[string]string) []string {
	var out []string
	for _, e := range s.Elves {
		var orcs []string
		for _, o := range s.Orcs {
			if o.Parent == prefix["elf"]+e.Short {
				orcs = append(orcs, fmt.Sprintf("%s (%d)", prefix["orc"]+o.Short, len(o.Slimes)))
			}
		}
		if len(orcs) == 0 {
			orcs = append(orcs, "the project-wide files")
		}
		out = append(out, prefix["elf"]+e.Short+": "+strings.Join(orcs, ", "))
	}
	return out
}

// sketchWrite is what writeSketch did: the docs written, the ones kept as they were (a doc of
// that name exists), the ones left out because a creature already in the world covers them.
type sketchWrite struct {
	wrote, kept, covered []string
}

// existingClaims reads the territory claims of the creatures already in the world, by name.
func existingClaims(root string, d config.Dist, dirs map[string]string) map[string][]string {
	out := map[string][]string{}
	seen := map[string]bool{}
	for rank, dir := range dirs {
		if dir == "" || seen[dir] || (rank != "slime" && rank != "kijin") {
			continue
		}
		seen[dir] = true
		ents, _ := os.ReadDir(filepath.Join(root, d.WorldDir, dir))
		for _, e := range ents {
			b, err := os.ReadFile(filepath.Join(root, d.WorldDir, dir, e.Name(), "README.md"))
			if err != nil {
				continue
			}
			for _, l := range strings.Split(string(b), "\n") {
				if !strings.Contains(l, "**Territory:**") {
					continue
				}
				for _, m := range reTick.FindAllStringSubmatch(l, -1) {
					out[dir+"/"+e.Name()] = append(out[dir+"/"+e.Name()], strings.TrimSuffix(strings.TrimPrefix(m[1], "./"), "/"))
				}
			}
		}
	}
	return out
}

var reTick = regexp.MustCompile("`([^`]+)`")

func claimsOverlap(a, b string) bool {
	return a == b || strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/")
}

// writeSketch writes the sketch's docs under the world dir. An existing doc is never touched
// (kept); a slime or kijin whose territory a creature already in the world claims is left out
// (covered) — a re-run over a refined colony adds only what nobody holds; an orc or elf is
// written only when something below it is (or it is the canon elf of a world with no elf).
func writeSketch(root string, d config.Dist, s sketch, dirs, prefix, display map[string]string) (sketchWrite, error) {
	var res sketchWrite
	held := existingClaims(root, d, dirs)
	docPath := func(p plan) string { return filepath.Join(d.WorldDir, dirs[p.Rank], p.Short, "README.md") }
	have := func(p plan) bool { return exists(filepath.Join(root, docPath(p))) }
	covered := func(p plan) bool {
		for owner, cs := range held {
			if strings.HasSuffix(owner, "/"+p.Short) && strings.HasPrefix(owner, dirs[p.Rank]+"/") {
				continue
			}
			for _, c := range cs {
				for _, t := range p.Territory {
					if claimsOverlap(c, strings.TrimSuffix(t, "/")) {
						return true
					}
				}
			}
		}
		return false
	}
	write := func(p plan) error {
		rel := docPath(p)
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(p.doc(prefix, display, "medium")), 0o644); err != nil {
			return err
		}
		res.wrote = append(res.wrote, filepath.ToSlash(rel))
		return nil
	}
	name := func(p plan) string { return prefix[p.Rank] + p.Short }
	needed := map[string]bool{} // parents something new reports to
	decide := func(p plan) (bool, error) {
		switch {
		case have(p):
			res.kept = append(res.kept, name(p))
			return false, nil
		case covered(p):
			res.covered = append(res.covered, name(p))
			return false, nil
		}
		needed[p.Parent] = true
		return true, write(p)
	}
	for _, o := range s.Orcs {
		for _, sl := range o.Slimes {
			if _, err := decide(sl); err != nil {
				return res, err
			}
		}
	}
	if s.Kijin != nil {
		if _, err := decide(*s.Kijin); err != nil {
			return res, err
		}
	}
	anyElf := false
	if ents, err := os.ReadDir(filepath.Join(root, d.WorldDir, dirs["elf"])); err == nil && len(ents) > 0 {
		anyElf = true
	}
	for _, o := range s.Orcs {
		if have(o) {
			res.kept = append(res.kept, name(o))
			continue
		}
		if needed[name(o)] {
			needed[o.Parent] = true
			if err := write(o); err != nil {
				return res, err
			}
		}
	}
	for i, e := range s.Elves {
		if have(e) {
			res.kept = append(res.kept, name(e))
			continue
		}
		if needed[name(e)] || (i == 0 && !anyElf) {
			if err := write(e); err != nil {
				return res, err
			}
		}
	}
	return res, nil
}

// table is the sketch as the founding brief shows it.
func (s sketch) table(prefix map[string]string) string {
	var b strings.Builder
	for _, p := range s.all() {
		fmt.Fprintf(&b, "- %s%s → %s · %s\n", prefix[p.Rank], p.Short, p.Parent, strings.Join(p.Territory, ", "))
	}
	return b.String()
}

//go:embed found.md
var foundBrief string

// foundingAsk is the complex level's turn: the brief, the words of this distribution, the sketch.
func foundingAsk(root string, d config.Dist, s sketch, dirs, prefix, display map[string]string) string {
	r := strings.NewReplacer(
		"{{BIN}}", d.Name, "{{WORLD}}", d.WorldDir, "{{PROJECT}}", filepath.Base(root), "{{LANG}}", or(s.Lang, "unknown"),
		"{{VOICE_DIR}}", dirs["elf"], "{{GATE_DIR}}", dirs["orc"], "{{TRUTH_DIR}}", dirs["slime"], "{{LEAD_DIR}}", dirs["kijin"],
		"{{VOICE}}", prefix["elf"], "{{GATE}}", prefix["orc"], "{{TRUTH}}", prefix["slime"], "{{LEAD}}", prefix["kijin"],
		"{{TRUTH_RANK}}", or(display["slime"], "Slime"),
		"{{SKETCH}}", s.table(prefix))
	return r.Replace(foundBrief)
}

func or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
