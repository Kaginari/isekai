package onto

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// World is a loaded ontology: the schema, the asserted graph files, the
// triples derived from the creatures on disk, and the reasoner's closure.
type World struct {
	Root     string // the directory that holds .isekai/
	Graph    *Graph
	Prefixes map[string]string
	Notes    []string // holes (@?), e.g. a missing schema file
	Docs     map[Term]*DocRef
	Layout   Layout
}

// DocRef points findings back at the doc a creature was derived from.
type DocRef struct {
	Path  string         // relative to Root
	Lines map[string]int // field → line number (1-based)
}

// RankDir is one rank as the derivation needs it: its canonical name, the folder its
// creatures live in (also their id prefix, `<dir>-<name>`) and the rank it reports to
// ("rimuru" at the top). Ranks are data (binary.md §Ranks and bodies): a name the built-in
// schema knows (elf, orc, slime, kijin) maps to its class; any other becomes a derived class
// ⊂ Creature, and its bond to its parent is the generic `above` unless it is truth- or
// verdict-shaped (slime ⇒ orc, orc ⇒ elf) or reports to rimuru.
type RankDir struct {
	Name   string
	Dir    string
	Parent string
}

// Layout is what a distribution renames on disk: the world directory, the law file and the
// rank table.
type Layout struct {
	WorldDir string    // ".isekai"
	Law      string    // "isekai.md"
	Ranks    []RankDir // nil = DefaultLayout's
}

// DefaultLayout is the isekai distribution's: the law's own ranks.
func DefaultLayout() Layout {
	return Layout{WorldDir: ".isekai", Law: "isekai.md", Ranks: []RankDir{
		{"elf", "elf", "rimuru"}, {"orc", "orc", "elf"}, {"slime", "slime", "orc"}, {"kijin", "kijin", "rimuru"},
	}}
}

func (l Layout) norm() Layout {
	d := DefaultLayout()
	if l.WorldDir != "" {
		d.WorldDir = l.WorldDir
	}
	if l.Law != "" {
		d.Law = l.Law
	}
	if l.Ranks != nil {
		d.Ranks = l.Ranks
	}
	return d
}

// ClassOf is the class term of a rank: the schema's for the law's ranks, a derived one
// (CamelCase of the name, ⊂ Creature) for any other.
func ClassOf(rank string) Term {
	rank = strings.ToLower(strings.TrimSpace(rank))
	if c, ok := raceClass[rank]; ok {
		return c
	}
	var b strings.Builder
	for _, part := range strings.FieldsFunc(rank, func(r rune) bool { return r == '_' || r == '-' || r == ' ' }) {
		b.WriteString(strings.ToUpper(part[:1]) + part[1:])
	}
	if b.Len() == 0 {
		return cCreature
	}
	return Is(b.String())
}

// bondFor is the bond a child rank draws to its parent rank.
func bondFor(child, parent string) Term {
	switch {
	case child == "slime" && parent == "orc":
		return pTruth
	case child == "orc" && parent == "elf":
		return pVerdict
	case parent == "rimuru":
		return pReports
	}
	return pAbove
}

// namePattern is the creature-id pattern for this layout, e.g. (?:zone|domain|coord)-x.
func (l Layout) namePattern() string {
	var ds []string
	seen := map[string]bool{}
	for _, r := range l.Ranks {
		if r.Dir != "" && !seen[r.Dir] {
			seen[r.Dir] = true
			ds = append(ds, regexp.QuoteMeta(r.Dir))
		}
	}
	if len(ds) == 0 {
		ds = []string{"never-matches-anything"}
	}
	return `(?:` + strings.Join(ds, "|") + `)-[a-z0-9][a-z0-9-]*`
}

// FindRoot walks up from dir to the nearest directory holding .isekai/.
func FindRoot(dir string) (string, error) { return FindRootIn(dir, DefaultLayout()) }

// FindRootIn walks up from dir to the nearest directory holding the layout's world dir.
func FindRootIn(dir string, l Layout) (string, error) {
	l = l.norm()
	d, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if st, err := os.Stat(filepath.Join(d, l.WorldDir)); err == nil && st.IsDir() {
			return d, nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			return "", errors.New("no " + l.WorldDir + "/ found here or above")
		}
		d = parent
	}
}

// OntologyDir is where the world keeps its schema and asserted graph.
func OntologyDir(root string) string { return OntologyDirIn(root, DefaultLayout()) }

// OntologyDirIn is OntologyDir under a layout's world dir.
func OntologyDirIn(root string, l Layout) string {
	return filepath.Join(root, l.norm().WorldDir, "ontology")
}

// Load reads schema + graph/*.ttl, derives the world's creatures, and reasons
// to a fixpoint.
func Load(root string) (*World, error) { return LoadLayout(root, DefaultLayout()) }

// LoadLayout is Load under a distribution's layout.
func LoadLayout(root string, l Layout) (*World, error) {
	l = l.norm()
	w := &World{Root: root, Graph: New(), Prefixes: DefaultPrefixes(), Docs: map[Term]*DocRef{}, Layout: l}
	dir := OntologyDirIn(root, l)
	schema := filepath.Join(dir, "schema.ttl")
	if b, err := os.ReadFile(schema); err == nil {
		if w.Prefixes, err = Parse(rel(root, schema), string(b), w.Graph, w.Prefixes); err != nil {
			return nil, err
		}
	} else {
		w.Notes = append(w.Notes, rel(root, schema)+" missing; using the built-in schema")
		if w.Prefixes, err = Parse("schema.ttl", DefaultSchema, w.Graph, w.Prefixes); err != nil {
			return nil, err
		}
	}
	files, _ := filepath.Glob(filepath.Join(dir, "graph", "*.ttl"))
	sort.Strings(files)
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		if w.Prefixes, err = Parse(rel(root, f), string(b), w.Graph, w.Prefixes); err != nil {
			return nil, err
		}
	}
	if err := w.derive(); err != nil {
		return nil, err
	}
	Infer(w.Graph)
	return w, nil
}

func rel(root, p string) string {
	if r, err := filepath.Rel(root, p); err == nil {
		return filepath.ToSlash(r)
	}
	return p
}

var (
	reField   = regexp.MustCompile(`(?i)^\s*[-*]?\s*(?:\*\*)?([A-Za-z][A-Za-z ]*?)(?::\*\*|\*\*:|:)\s*(.+?)\s*$`)
	reTick    = regexp.MustCompile("`([^`\n\\s]*/[^`\n\\s]*)`") // a path never holds a space; a command does
	rePath    = regexp.MustCompile(`[\w.\-*]+(?:/[\w.\-*]*)+`)
	reWord    = regexp.MustCompile(`[a-z0-9][a-z0-9-]*`)
	fieldKeys = map[string]string{"reports to": "parent", "orc": "parent", "elf": "parent", "parent": "parent",
		"territory": "territory", "zone": "territory", "owns": "territory", "minds": "minds", "mind": "minds", "wears": "minds", "hats": "minds"}
)

// fields reads `- **Key:** value` lines leniently; keys are normalized.
// A `## Territory` (or `**Territory:**`) HEADING ZONE is also honored: any
// `- `code“ bullets under it are joined into a comma-list so lane filters
// run on one closet shape in both markups (class (f) fixture pins this
// bullet form that earlier regex never matched, minting no owns from it).
func fields(doc string) (map[string]string, map[string]int) {
	vals, lines := map[string]string{}, map[string]int{}
	var headingKey string // seat field being filled by `- `_`` bullets
	for i, line := range strings.Split(doc, "\n") {
		t := strings.TrimSpace(line)
		// end any open bullet-accumulation when a new heading begins
		if strings.HasPrefix(t, "#") {
			if !strings.HasPrefix(t, "## ") && !strings.HasPrefix(t, "##\t") {
				headingKey = "" // top-level headings reset the accumulator to silence
				continue
			}
			h := strings.ToLower(strings.TrimSpace(strings.TrimLeft(t, "# ")))
			h = strings.TrimSuffix(h, ":")
			if k, ok := fieldKeys[h]; ok {
				headingKey = k
			} else {
				headingKey = ""
			}
			continue
		}
		if headingKey != "" && strings.HasPrefix(t, "- ") {
			// bullets under the heading — keep just the backticked code spans,
			// falling back to whole-line only when no code marks the row (a
			// `# heading` may precede or follow them).
			body := ""
			if idx := strings.Index(t, "`"); idx >= 0 {
				rest := t[idx+1:]
				if end := strings.Index(rest, "`"); end >= 0 {
					body = rest[:end]
				}
			}
			if body == "" {
				body = strings.TrimSpace(strings.TrimPrefix(t, "- "))
			}
			if _, seen := vals[headingKey]; !seen {
				vals[headingKey] = body
			} else if !strings.Contains(vals[headingKey], body) {
				vals[headingKey] += ", " + body
			}
			if _, seen := lines[headingKey]; !seen {
				lines[headingKey] = i + 1
			}
			continue
		}
		m := reField.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(m[1]))
		k, ok := fieldKeys[key]
		if !ok {
			continue
		}
		if _, seen := vals[k]; seen {
			continue
		}
		vals[k] = strings.TrimSpace(m[2])
		lines[k] = i + 1
	}
	return vals, lines
}

func normPath(p string) string {
	p = strings.TrimPrefix(p, "./")
	for _, suf := range []string{"/**/*", "/**", "/*", "**", "*"} {
		p = strings.TrimSuffix(p, suf)
	}
	return strings.ToLower(strings.TrimRight(p, "/"))
}

// derive adds the structural triples the world's files imply. Nothing here is
// ever written back to disk.
func (w *World) derive() error {
	g := w.Graph
	l := w.Layout.norm()
	reName := regexp.MustCompile(`(?i)\b(?:rimuru|` + l.namePattern() + `)`)
	// the first rank on a dir is the class its creatures take (an ascended rank shares its base's dir)
	rankOfDir := map[string]RankDir{}
	var dirs []string
	for _, r := range l.Ranks {
		if r.Dir == "" {
			continue
		}
		if _, dup := rankOfDir[r.Dir]; !dup {
			rankOfDir[r.Dir] = r
			dirs = append(dirs, r.Dir)
		}
	}
	rankByName := map[string]RankDir{}
	for _, r := range l.Ranks {
		rankByName[r.Name] = r
	}
	add := func(s, p, o Term) { g.AddDerived(Triple{s, p, o}) }
	for _, d := range dirs {
		c := ClassOf(rankOfDir[d].Name)
		if _, builtin := raceClass[rankOfDir[d].Name]; !builtin {
			add(c, rdfType, Is("Class"))
			add(c, pSubClassOf, cCreature)
		}
	}
	docNode := func(owner Term, path string) {
		d := Is("doc-" + owner.Local())
		add(owner, pDoc, d)
		add(d, rdfType, cDoc)
		add(d, pPath, L(path))
	}
	add(tRimuru, rdfType, cRimuru)
	add(tRimuru, pName, L("rimuru"))
	docNode(tRimuru, l.WorldDir+"/"+l.Law)

	// minds first, so docs can be matched against their names
	minds := map[string]string{}
	// the world's own skill dirs first (crafted, then base — the dist names its world dir), then
	// the hosts': a creature's Minds field names them, so the graph must know them to bond `wears`
	bases := append([]string{l.WorldDir + "/crafted-skills", l.WorldDir + "/base-skills"}, mindDirs...)
	for _, base := range bases {
		ents, err := os.ReadDir(filepath.Join(w.Root, base))
		if err != nil {
			continue
		}
		for _, e := range ents {
			if !e.IsDir() {
				continue
			}
			p := filepath.Join(w.Root, base, e.Name(), "SKILL.md")
			if _, err := os.Stat(p); err == nil {
				if _, dup := minds[e.Name()]; !dup {
					minds[e.Name()] = rel(w.Root, p)
				}
			}
		}
	}
	mindNames := make([]string, 0, len(minds))
	for n := range minds {
		mindNames = append(mindNames, n)
	}
	sort.Strings(mindNames)
	for _, n := range mindNames {
		m := Is("mind-" + n)
		add(m, rdfType, cMind)
		add(m, pName, L(n))
		docNode(m, minds[n])
		// a skill with no race prefix is a shared host tool (the law's shared-skills row), not a
		// creature's mind: it is nobody's to wear, so MindWorn does not apply to it
		if _, ranked := rankOfDir[strings.SplitN(n, "-", 2)[0]]; !ranked {
			add(m, pShared, Bool(true))
		}
	}

	type creature struct {
		race, dir, id, doc string
		text               string
	}
	var cs []creature
	for _, rd := range dirs {
		race := rankOfDir[rd].Name
		d := filepath.Join(w.Root, l.WorldDir, rd)
		ents, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, e := range ents {
			if !e.IsDir() {
				continue
			}
			c := creature{race: race, dir: e.Name(), id: rd + "-" + strings.ToLower(e.Name())}
			files, _ := os.ReadDir(filepath.Join(d, e.Name()))
			var mds []string
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
					mds = append(mds, f.Name())
				}
			}
			sort.Strings(mds)
			for _, want := range []string{"README.md", e.Name() + ".md", "SKILL.md"} {
				for _, f := range mds {
					if f == want && c.doc == "" {
						c.doc = f
					}
				}
			}
			if c.doc == "" && len(mds) > 0 {
				c.doc = mds[0]
			}
			if c.doc != "" {
				p := filepath.Join(d, e.Name(), c.doc)
				b, err := os.ReadFile(p)
				if err != nil {
					return err
				}
				c.text = string(b)
				c.doc = rel(w.Root, p)
			}
			cs = append(cs, c)
		}
	}
	known := map[string]bool{"rimuru": true}
	for _, c := range cs {
		known[c.id] = true
	}
	for _, c := range cs {
		self := Is(c.id)
		add(self, rdfType, ClassOf(c.race))
		add(self, pName, L(c.id))
		rdir := rankByName[c.race].Dir
		if c.doc == "" {
			w.Notes = append(w.Notes, c.id+" has no doc in "+l.WorldDir+"/"+rdir+"/"+c.dir+"/")
			w.Docs[self] = &DocRef{Path: l.WorldDir + "/" + rdir + "/" + c.dir + "/", Lines: map[string]int{}}
			continue
		}
		docNode(self, c.doc)
		vals, lines := fields(c.text)
		w.Docs[self] = &DocRef{Path: c.doc, Lines: lines}

		// parent: the declared field, else the first creature of the expected rank the doc names
		parent := ""
		if v, ok := vals["parent"]; ok {
			if m := reName.FindString(v); m != "" {
				parent = strings.ToLower(m)
			}
		}
		prank := rankByName[c.race].Parent
		if parent == "" {
			wantRace := ""
			if pr, ok := rankByName[prank]; ok && pr.Dir != "" {
				wantRace = pr.Dir + "-"
			}
			for _, m := range reName.FindAllString(c.text, -1) {
				m = strings.ToLower(m)
				if wantRace != "" && strings.HasPrefix(m, wantRace) && m != c.id {
					parent = m
					break
				}
			}
			if parent == "" && prank == "rimuru" {
				parent = "rimuru"
			}
		}
		if parent != "" && parent != c.id {
			prace := "rimuru"
			if parent != "rimuru" {
				prace = rankOfDir[strings.SplitN(parent, "-", 2)[0]].Name
			}
			add(self, bondFor(c.race, prace), Is(parent))
			if !known[parent] {
				add(Is(parent), pName, L(parent))
			}
		}

		// Creature identity for the prose gate: the name the census trusts
		// (frontmatter, falling back to the seat id) and the zone dir base.
		// Both are what "self-naming" means for the boundary test below.
		name := normName(vals["name"])
		if name == "" {
			name = c.id
		}
		zone := strings.TrimSuffix(c.dir, "/") // seat dir — self-naming + blob claims
		// territory: STRUCTURED field is authoritative; prose is supplement.
		seen := map[string]bool{}
		// The structured Territory field is a claim list, consumed exactly
		// the way the write gate consumes it (world.World.Owner,
		// world/world.go:418 chain): split on commas, trim spaces, strip
		// wrapping backticks, normPath, keep case.
		//
		// Ruling (2026-10-11, census class (f) derive heal — supersedes the
		// 2026-10-07 class-(e) zone-root-disjoint rule): the carve-out to
		// "every entry derives" — a structured entry whose resolved path is
		// NOT contained by the creature's seat (`c.dir`, the directory under
		// `.isekai/<race>/` where its doc lives) is a cross-ref, never owns
		// (census schema mints no owns: onto/schema.ttl has no crossref
		// predicate → silent drop). Containment is seat-scoped so disjoint
		// seats sharing a leading root token (`app-commands` → `app/`) do NOT
		// guard: `app/app.go` under seat `slime-app-commands` drops outright.
		// Containsment is root-first: the blob's root token must equal the
		// seat's leaf name OR live under it (`bench/tasks-tb2` under seat
		// `slime-bench-tasks`: root `bench` = `base(bench-tasks)` → derives).
		// slime-owna declares `.isekai/slime/owna/...` under seat `slime-owna`:
		// root `.isekai` ≠ `owna`, ∉ owna/ → drop. slime-auth declares
		// `src/auth/` under seat `slime-auth`: root `src` ≠ `auth`, ∉ auth/
		// → drop (verbatim shape guard, census class (e) pin — kept live per
		// TestDeriveOwns_StructuredBareFilename and the census-heal ruling).
		// Bare tokens (`local.go`, `app`) are creature-local (world_test.go:
		// 387, TestDeriveOwns_StructuredBareFilename).
		entries := []string{}
		for _, e := range strings.Split(vals["territory"], ",") {
			e = strings.Trim(strings.TrimSpace(e), "`")
			if p := normPath(e); p != "" && !seen[p] {
				seen[p] = true
				entries = append(entries, p)
			}
		}
		slug := filepath.Base(zone)
		// class-(f) 2026-10-11: README ピコで stage unreliable: hyphen-span slug
		// head-vs-tail separation is the razor. headFrag keeps only the first
		// token; tailFrags keeps anything after first hyphen. Hoisted so both
		// the structured lane (below) and the diagnostic banner (Load)
		// inspect the same fragments.
		headFrag := slug
		if i := strings.Index(slug, "-"); i > 0 {
			headFrag = slug[:i]
		}
		tailFrags := ""
		if i := strings.Index(slug, "-"); i > 0 {
			tailFrags = slug[i+1:]
		}
		inZone := func(p string) bool {
			if i := strings.Index(p, "/"); i > 0 {
				slashFirst := p[:i]
				// hyphen-stem: the seat slug must split (e.g. bench-tasks); if
				// the slug is unhyphenated, it never admits hyphen-adjacent roots
				// like "app-commands". Then the leading path root must equal
				// headFrag, and the tailFrags must appear as a later segment in
				// the path (`bench/tasks-tb2` under seat slime-bench-tasks;
				// `app/app.go` under seat slime-app-commands has no "commands"
				// segment → drops).
				if tailFrags != "" && slashFirst == headFrag &&
					strings.Contains(strings.TrimPrefix(p, slashFirst)+"/", tailFrags) {
					return true
				}
				// class-(f) fixture pin (world_test.go:564-575): bare `.isekai`
				// stays mintable by EVERY creature — dedupe is Validate's job.
				// class-(e): rank-law claims the law path; shared `src/` trunk
				// is shared by every creature and never binds one seat.
				if slashFirst == ".isekai" || slashFirst == "src" {
					return true
				}
				// own-name arm: name-eponymous seat traces its own root
				// (slime-onto → `onto/…`, slime-owna → `owna/…`).
				// For a hyphenated slug (slime-app-commands): after hyphen-stem
				// failed (no later segment matching the tail), root == headFrag
				// is the foreign-shared lookalike — silent drop so
				// `app/app.go` under seat slime-app-commands does NOT mint.
				if tailFrags == "" && slashFirst == headFrag {
					return true
				}
				// foreign disjoint — silent drop (onto/schema.ttl: no crossref).
				return false
			}
			if p == ".isekai" {
				return true
			}
			// bare filename, creature-local (world_test.go:387)
			return true
		}
		for _, p := range entries {
			if !inZone(p) {
				continue // zone-foreign structured claim: crossref, never owns
			}
			add(self, pOwns, L(p))
		}
		// Prose supplement: backticked path spans in the doc body — but only
		// spans that SELF-NAME the creature (its name or zone dir adjoining
		// the span) derive is:owns. Structured Territory stays the one
		// authoritative claim list; a backticked cross-reference to foreign
		// soil (`world/world.go` in another creature's doc) is a path, not
		// a claim. Slashes hold the filesystem — plain words never derive.
		for _, m := range reTick.FindAllStringSubmatchIndex(c.text, -1) {
			span := c.text[m[2]:m[3]]
			if !selfNamesSpan(c.text, m[2], m[3], name, zone) {
				continue
			}
			p := normPath(span)
			// census class (f) 2026-10-11: prose claims must cite the creature's
			// seat-root; those that do not are foreign territorials minted by
			// selfNamesSpan whole-word adjacency from the SAME comment naming the
			// creature's doc (class (f) commission: "slime-onto's Territory claims
			// bench/tasks-tb2 as own ... prose mention minted under TWO seats
			// simultaneously"). Real fixture: auditor-eye mints its `.isekai/log.md`
			// EXACTLY BECAUSE Fixture auth also claims it — but the structured
			// Territory field is the only legitimate channel for that; markdown
			// a creature reads (or humans annotate into its doc) is never itself
			// a claim the THRONE wrote.
			if i := strings.Index(p, "/"); i > 0 {
				switch p[:i] {
				case ".isekai":
					// World-shared prefix — the SELF-NAME lane already requires
					// adjacency to the creature's name.
				case zone:
					// Seat-rooted claim, e.g. bench/tasks-tb2 under slime-bench-tasks.
				default:
					// class-(f) sharpen: prose mints only what the THRONE wrote.
					// smoke's Territory authoritatively lists `.isekai`; prose under
					// seat smoke naming `bench/tasks-tb2` (root=bench) is a third-party
					// commentary, never a claim — the prose lane admits only roots
					// that are the seat's own slug or a hyphen-stem of it.
					continue
				}
			}
			// The seat-root prefixes are historical artifacts (dead `.opencode/skills`,
			// moved `.isekai/minds`) that bodies mention in Thoughts when recording
			// their own corrections — they are never a claim of ownership. Drop them,
			// root or anything nested under them.
			if p == ".opencode/skills" || strings.HasPrefix(p, ".opencode/skills/") ||
				p == ".isekai/minds" || strings.HasPrefix(p, ".isekai/minds/") {
				continue
			}
			// census-heal 2026-10-11: a prose line that names the creature may
			// still mint owns the world grammar reserves as shared soil — bare
			// `.isekai`, anything under `.isekai/tmp/` — or may mint URI
			// templates (query placeholders like `api/plan?body=<id>`) as if
			// they were filesystem paths. Neither is ownable; the structured
			// Territory field is the only legitimate channel for those claims.
			if p == ".isekai" || p == ".isekai/" {
				continue
			}
			if p == ".isekai/tmp" || strings.HasPrefix(p, ".isekai/tmp/") {
				continue
			}
			if strings.ContainsAny(p, "?<>") {
				continue
			}
			// class (d) 2026-10-07: a prose-narrated span with a leading
			// slash and NO dot-extension (`/api/tree`, `/api/tree/children`)
			// is a REST route, never filesystem soil — normPath kept the
			// leading slash (normalising it merely because it holds a
			// slash), so the route arrives here intact. Mirror of the
			// committed census-heal refusal above: structured Territory
			// already owns the claim channel, so the prose-mint lane
			// refuses the route. Dot-bearing code paths survive.
			if strings.HasPrefix(p, "/") && !strings.Contains(p, ".") {
				continue
			}
			if p == "" || seen[p] {
				continue
			}
			seen[p] = true
			add(self, pOwns, L(p))
		}

		// worn minds: named in a Minds field, or any mind's name in the doc as a whole word
		lower := strings.ToLower(c.text)
		for _, n := range mindNames {
			if wordIn(lower, strings.ToLower(n)) {
				add(self, pWears, Is("mind-"+n))
			}
		}
		if v, ok := vals["minds"]; ok {
			for _, n := range reWord.FindAllString(strings.ToLower(v), -1) {
				if _, exists := minds[n]; exists {
					add(self, pWears, Is("mind-"+n))
				}
			}
		}
	}
	return nil
}

func wordIn(text, word string) bool {
	for i := 0; ; {
		j := strings.Index(text[i:], word)
		if j < 0 {
			return false
		}
		s, e := i+j, i+j+len(word)
		before := s == 0 || !isWordByte(text[s-1])
		after := e == len(text) || !isWordByte(text[e])
		if before && after {
			return true
		}
		i = s + 1
	}
}

func isWordByte(b byte) bool {
	return b == '-' || b == '_' || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b >= 0x80
}

// normName normalizes a frontmatter name into the census-stable lower form:
// lower-cased, trimmed, internal whitespace collapsed. May return "".
func normName(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}

// selfNamesSpan reports whether the backticked span doc[start:end] self-names
// the creature — the prose gate for the census owns-supplement. The creature
// itself is its zone dir base and its frontmatter name stripped of the rank
// prefix (a rank holds no claim). A span self-names when the creature heads
// the path (`owna/…`) or bases the file (`leaf/owna.go`), or when the name
// adjoins the span as a whole word. A bare span fails by reTick's slash
// before ever reaching here; the equals-guards are the belt to its braces.
// `world/world.go` is a foreign cross-ref (head world ≠ owna) and `owna` a
// zone word, not a file — neither derives; OWNERSHIP.md and child owna pass.
func selfNamesSpan(doc string, start, end int, name, zone string) bool {
	span := doc[start:end]
	self := zone
	if rest, ok := strings.CutPrefix(name, "-"+zone); zone != "" && ok && rest != "" {
		self = rest // frontmatter name stripped of its rank prefix (slime-owna → owna)
	}
	if span == zone || (self != "" && span == self) {
		return false // bare zone/name word, no file shape
	}
	if self != "" && span == self+"/" {
		return false // zone prefix alone is a name, not a path — the trailing
		// slash holds no file, and normPath would collapse it to the bare zone.
	}
	if self != "" && (strings.HasPrefix(span, self+"/") ||
		strings.HasSuffix(span, "/"+self) || baseOf(span) == self) {
		return true // the path heads or bases on the creature itself
	}
	if dot := strings.LastIndexByte(span, '.'); dot > 0 && dot < len(span)-1 {
		return false // a foreign file path (world/world.go) is a cross-ref, not a claim
	}
	if name == "" {
		return false
	}
	// A name-like span (OWNERSHIP.md, dir/owna, owna/child) self-names when the
	// creature's whole name adjoins it as a word, backticks ignored.
	lo, hi := start, end
	if lo > 0 && doc[lo-1] == '`' {
		lo--
	}
	if hi < len(doc) && doc[hi] == '`' {
		hi++
	}
	return (lo > 0 && wordIn(doc[:lo], name)) || (hi < len(doc) && wordIn(doc[hi:], name))
}

func baseOf(p string) string {
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[i+1:]
	}
	return p
}

// Creature resolves a name like slime-auth (or auth, when unambiguous) to its term.
func (w *World) Creature(name string) (Term, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "rimuru" {
		return tRimuru, nil
	}
	if w.Graph.Has(Triple{Is(name), rdfType, cCreature}) {
		return Is(name), nil
	}
	var hits []Term
	for _, c := range w.Graph.Instances(cCreature) {
		if strings.HasSuffix(c.Local(), "-"+name) {
			hits = append(hits, c)
		}
	}
	if len(hits) == 1 {
		return hits[0], nil
	}
	if len(hits) > 1 {
		return Term{}, fmt.Errorf("ambiguous creature %q: %s", name, joinLocal(hits))
	}
	return Term{}, fmt.Errorf("unknown creature %q", name)
}

func joinLocal(ts []Term) string {
	s := make([]string, len(ts))
	for i, t := range ts {
		s[i] = t.Local()
	}
	return strings.Join(s, ", ")
}
