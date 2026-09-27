package app

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Kaginari/isekai/config"
)

// The three ways to found a world (`init --level`):
//   - light (soft): the law, the log, the instruments and the memory tiers — no creature.
//   - medium: light, and a colony sketched from the tree with no model: an elf, an orc per
//     source area, a slime per sub-area, a kijin when there is CI, verify lines for the language.
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

// sketch is a colony read from the tree alone.
type sketch struct {
	Project string
	Lang    string
	Elf     plan
	Orcs    []plan
	Kijin   *plan
}

type plan struct {
	Rank, Short, Purpose string
	Territory            []string
	Parent               string
	Verify               []string
	Slimes               []plan
}

var (
	sourceExt = map[string]string{".go": "go", ".ts": "node", ".tsx": "node", ".js": "node", ".jsx": "node", ".mjs": "node", ".vue": "node", ".svelte": "node",
		".py": "python", ".rs": "rust", ".java": "jvm", ".kt": "jvm", ".scala": "jvm", ".rb": "ruby", ".php": "php", ".cs": "dotnet", ".swift": "swift",
		".c": "c", ".cc": "c", ".cpp": "c", ".h": "c", ".hpp": "c", ".ex": "elixir", ".exs": "elixir"}
	skipDir = map[string]bool{"node_modules": true, "vendor": true, "dist": true, "build": true, "target": true, "out": true, "bin": true,
		"testdata": true, "coverage": true, "__pycache__": true, ".venv": true, "venv": true}
	reName = regexp.MustCompile(`[^a-z0-9]+`)
)

func slug(s string) string {
	return strings.Trim(reName.ReplaceAllString(strings.ToLower(strings.TrimLeft(s, "_.")), "-"), "-")
}

// sourceFiles counts the source files under dir (recursively), by language.
func sourceFiles(dir string) (int, map[string]int) {
	n, langs := 0, map[string]int{}
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if p != dir && (strings.HasPrefix(d.Name(), ".") || skipDir[d.Name()]) {
				return filepath.SkipDir
			}
			return nil
		}
		if l, ok := sourceExt[filepath.Ext(p)]; ok {
			n++
			langs[l]++
		}
		return nil
	})
	return n, langs
}

// sketchColony reads the tree into a colony, without a model.
func sketchColony(root string, d config.Dist) sketch {
	s := sketch{Project: slug(filepath.Base(root))}
	if s.Project == "" {
		s.Project = "core"
	}
	has := func(p string) bool { _, err := os.Stat(filepath.Join(root, p)); return err == nil }
	_, langs := sourceFiles(root)
	best := 0
	for l, n := range langs {
		if n > best || (n == best && l < s.Lang) {
			s.Lang, best = l, n
		}
	}
	test := func(dir string) []string {
		switch {
		case s.Lang == "go" && has("go.mod"):
			if dir == "" {
				return []string{"go build ./...", "go vet ./..."}
			}
			if strings.HasPrefix(filepath.Base(dir), "_") {
				return []string{"go test ./" + dir + "/"} // ./... skips a leading underscore
			}
			return []string{"go test ./" + dir + "/..."}
		case s.Lang == "python" && dir != "":
			return []string{"python -m pytest -q " + dir}
		case s.Lang == "rust" && dir == "":
			return []string{"cargo test"}
		case s.Lang == "node" && dir == "" && has("package.json"):
			return []string{"npm test"}
		}
		return nil
	}
	var manifests []string
	for _, m := range []string{"README.md", "go.mod", "package.json", "pyproject.toml", "setup.py", "Cargo.toml", "pom.xml", "build.gradle"} {
		if has(m) {
			manifests = append(manifests, m)
		}
	}
	s.Elf = plan{Rank: "elf", Short: s.Project, Parent: "rimuru", Territory: manifests, Verify: test(""),
		Purpose: "the shared mind and voice of " + filepath.Base(root) + ": keeps the cross-domain rules, routes work to the orcs, speaks back up."}
	// areas: top-level dirs holding source, largest first
	type area struct {
		name string
		n    int
	}
	var areas []area
	ents, _ := os.ReadDir(root)
	var rootFiles []string
	for _, e := range ents {
		name := e.Name()
		if strings.HasPrefix(name, ".") || skipDir[name] || name == d.WorldDir {
			continue
		}
		if !e.IsDir() {
			if _, ok := sourceExt[filepath.Ext(name)]; ok {
				rootFiles = append(rootFiles, name)
			}
			continue
		}
		if n, _ := sourceFiles(filepath.Join(root, name)); n > 0 {
			areas = append(areas, area{name, n})
		}
	}
	sort.SliceStable(areas, func(i, j int) bool { return areas[i].n > areas[j].n })
	var core *plan
	coreOrc := func() *plan {
		if core == nil {
			s.Orcs = append(s.Orcs, plan{Rank: "orc", Short: "core", Parent: "elf-" + s.Project, Verify: test(""),
				Purpose: "rules the small areas and the entry point, until one of them grows its own domain."})
			core = &s.Orcs[len(s.Orcs)-1]
		}
		return core
	}
	for i, a := range areas {
		if a.n < 3 || i >= 6 {
			c := coreOrc()
			c.Territory = append(c.Territory, a.name+"/")
			c.Slimes = append(c.Slimes, plan{Rank: "slime", Short: slug(a.name), Parent: "orc-core", Territory: []string{a.name + "/"}, Verify: test(a.name),
				Purpose: "the ground truth of `" + a.name + "/`."})
			continue
		}
		o := plan{Rank: "orc", Short: slug(a.name), Parent: "elf-" + s.Project, Territory: []string{a.name + "/"}, Verify: test(a.name),
			Purpose: "rules the `" + a.name + "/` domain and holds its gate."}
		subs, _ := os.ReadDir(filepath.Join(root, a.name))
		for _, sub := range subs {
			if !sub.IsDir() || strings.HasPrefix(sub.Name(), ".") || skipDir[sub.Name()] || len(o.Slimes) >= 6 {
				continue
			}
			rel := a.name + "/" + sub.Name()
			if n, _ := sourceFiles(filepath.Join(root, rel)); n > 0 {
				o.Slimes = append(o.Slimes, plan{Rank: "slime", Short: slug(sub.Name()), Parent: "orc-" + o.Short, Territory: []string{rel + "/"}, Verify: test(rel),
					Purpose: "the ground truth of `" + rel + "/`."})
			}
		}
		if len(o.Slimes) == 0 {
			o.Slimes = append(o.Slimes, plan{Rank: "slime", Short: slug(a.name), Parent: "orc-" + o.Short, Territory: []string{a.name + "/"}, Verify: test(a.name),
				Purpose: "the ground truth of `" + a.name + "/`."})
		}
		s.Orcs = append(s.Orcs, o)
	}
	if len(rootFiles) > 0 {
		c := coreOrc()
		c.Territory = append(c.Territory, rootFiles...)
		c.Slimes = append(c.Slimes, plan{Rank: "slime", Short: "entry", Parent: "orc-core", Territory: rootFiles, Purpose: "the ground truth of the entry point (" + strings.Join(rootFiles, ", ") + ")."})
	}
	for _, ci := range []string{".github/workflows", ".gitlab-ci.yml", ".circleci", "Jenkinsfile", ".drone.yml"} {
		if has(ci) {
			t := ci
			if st, err := os.Stat(filepath.Join(root, ci)); err == nil && st.IsDir() {
				t += "/"
			}
			k := plan{Rank: "kijin", Short: "release", Parent: "rimuru", Territory: []string{t}, Verify: test(""),
				Purpose: "owns CI and releases end to end, across sessions; anything that tags, publishes or provisions waits for the human."}
			s.Kijin = &k
			break
		}
	}
	return s
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
	fmt.Fprintf(&b, "- **Purpose:** %s\n", p.Purpose)
	fmt.Fprintf(&b, "- **Founded:** `init --level %s`, from the tree alone — refine the traits and the verify lines from the code\n", level)
	b.WriteString("\n## Traits\n- (none recorded yet: the first session that works here writes what it learns)\n")
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
	out := []plan{s.Elf}
	for _, o := range s.Orcs {
		out = append(out, o)
	}
	for _, o := range s.Orcs {
		out = append(out, o.Slimes...)
	}
	if s.Kijin != nil {
		out = append(out, *s.Kijin)
	}
	return out
}

// writeSketch writes the sketch's docs under the world dir; an existing doc is never touched.
func writeSketch(root string, d config.Dist, s sketch, dirs, prefix, display map[string]string) ([]string, error) {
	var made []string
	for _, p := range s.all() {
		rel := filepath.Join(d.WorldDir, dirs[p.Rank], p.Short, "README.md")
		path := filepath.Join(root, rel)
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return made, err
		}
		if err := os.WriteFile(path, []byte(p.doc(prefix, display, "medium")), 0o644); err != nil {
			return made, err
		}
		made = append(made, filepath.ToSlash(rel))
	}
	return made, nil
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
