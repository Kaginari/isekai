package world

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Kaginari/isekai/discover"
	"github.com/Kaginari/isekai/memory"
	"github.com/Kaginari/isekai/value"
)

// The three arms' drafts (world/evolve.go weighs, proves and lands them). Each reads the docs as
// they stand and the code as it is, and writes nothing: it fills the plan.

func skillBases(worldDir string) []string { return discover.SkillBases(worldDir) }

// --- reading a creature doc ---

func fieldRe(key string) *regexp.Regexp {
	return regexp.MustCompile(`(?mi)^[ \t]*[-*][ \t]*\*\*` + regexp.QuoteMeta(key) + `:?\*\*:?[ \t]*(.*)$`)
}

// docField is the value of a doc's `- **Key:** value` line, "" when it has none.
func docField(text, key string) string {
	if m := fieldRe(key).FindStringSubmatch(text); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// setField rewrites a doc's `- **Key:** value` line, or adds it after the doc's last header field.
func setField(text, key, val string) string {
	line := "- **" + key + ":** " + val
	re := fieldRe(key)
	if loc := re.FindStringIndex(text); loc != nil {
		return text[:loc[0]] + line + text[loc[1]:]
	}
	for _, after := range []string{"Reports to", "Territory", "Rank"} {
		if loc := fieldRe(after).FindStringIndex(text); loc != nil {
			return text[:loc[1]] + "\n" + line + text[loc[1]:]
		}
	}
	if loc := reH1.FindStringIndex(text); loc != nil {
		return text[:loc[1]] + "\n\n" + line + text[loc[1]:]
	}
	return line + "\n" + text
}

// sectionSpan is the body of a doc's `## <heading>` section (after its heading line, up to the
// next # or ## heading); ok false when the doc has none.
func sectionSpan(text, heading string) (start, end int, ok bool) {
	re := regexp.MustCompile(`(?mi)^##[ \t]+` + regexp.QuoteMeta(heading) + `[ \t]*$`)
	loc := re.FindStringIndex(text)
	if loc == nil {
		return 0, 0, false
	}
	start = loc[1]
	if start < len(text) && text[start] == '\n' {
		start++
	}
	end = len(text)
	if m := regexp.MustCompile(`(?m)^##?[ \t]`).FindStringIndex(text[start:]); m != nil {
		end = start + m[0]
	}
	return start, end, true
}

// sectionBullets are the `- ` lines of a section.
func sectionBullets(text, heading string) []string {
	s, e, ok := sectionSpan(text, heading)
	if !ok {
		return nil
	}
	var out []string
	for _, l := range strings.Split(text[s:e], "\n") {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") {
			out = append(out, strings.TrimSpace(t[2:]))
		}
	}
	return out
}

// appendSection adds lines at the end of a section, creating it (before ## Thoughts, else at the
// end) when the doc has none.
func appendSection(text, heading string, lines []string) string {
	if len(lines) == 0 {
		return text
	}
	add := strings.Join(lines, "\n") + "\n"
	if s, e, ok := sectionSpan(text, heading); ok {
		body := strings.TrimRight(text[s:e], "\n")
		if body != "" {
			body += "\n"
		}
		return text[:s] + body + add + "\n" + text[e:]
	}
	block := "## " + heading + "\n" + add + "\n"
	if s, _, ok := sectionSpan(text, "Thoughts"); ok {
		head := regexp.MustCompile(`(?mi)^##[ \t]+Thoughts[ \t]*$`).FindStringIndex(text[:s])
		if head != nil {
			return text[:head[0]] + block + text[head[0]:]
		}
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return text + "\n" + block
}

// cutDesk takes Thoughts out of a doc's desk and leaves one dated line saying where they went.
func cutDesk(text string, drop []memory.Thought, line string) string {
	sort.Slice(drop, func(i, j int) bool { return drop[i].Start > drop[j].Start })
	for _, t := range drop {
		if t.Start >= 0 && t.End <= len(text) && t.Start <= t.End {
			text = text[:t.Start] + text[t.End:]
		}
	}
	if line == "" {
		return text
	}
	if _, end, ok := memory.DeskSpan(text); ok {
		pre := text[:end]
		trimmed := strings.TrimRight(pre, "\n")
		rest := text[end:]
		sep := "\n"
		if rest != "" {
			sep = "\n\n"
		}
		return trimmed + "\n" + line + sep + strings.TrimLeft(rest, "\n")
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return text + "\n## Thoughts\n" + line + "\n"
}

var rePathy = regexp.MustCompile(`[A-Za-z0-9_.\-]+(?:/[A-Za-z0-9_.\-]+)+/?`)

// pathsIn are the path-shaped spans of a text.
func pathsIn(text string) []string {
	var out []string
	seen := map[string]bool{}
	for _, p := range rePathy.FindAllString(text, -1) {
		p = strings.Trim(p, ".")
		if p == "" || seen[p] || strings.Count(p, "/") == 0 {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

// entryOf is the zone entry a path falls under (tried as written, then under each territory
// prefix: a Thought may name `api/x.go` for `src/pay/api/x.go`); "" when none.
func entryOf(path string, entries, terr []string) string {
	cands := []string{path}
	for _, t := range terr {
		cands = append(cands, strings.TrimSuffix(t, "/")+"/"+path)
	}
	best := ""
	for _, c := range cands {
		rel := normRel(c)
		for _, e := range entries {
			e2 := strings.TrimSuffix(e, "/")
			if under(rel, e2) && len(e2) > len(best) {
				best = e
			}
		}
		if best != "" {
			return best
		}
	}
	return ""
}

// --- the survey ---

// survey reads a territory: the integrator's sketch when wired, else files and lines alone.
func (w *World) survey(prefixes []string) TerritorySurvey {
	if s := w.Evolution().Survey; s != nil {
		return s(prefixes)
	}
	var sv TerritorySurvey
	for _, p := range prefixes {
		root := filepath.Join(w.Root, filepath.FromSlash(strings.TrimSuffix(p, "/")))
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if path != root && (strings.HasPrefix(d.Name(), ".") || skipDirRelief[d.Name()]) {
					return filepath.SkipDir
				}
				return nil
			}
			sv.Files++
			if fi, err := d.Info(); err == nil && fi.Size() < 1<<20 {
				if f, err := os.Open(path); err == nil {
					sc := bufio.NewScanner(f)
					sc.Buffer(make([]byte, 64<<10), 1<<20)
					for sc.Scan() {
						sv.Lines++
					}
					f.Close()
				}
			}
			return nil
		})
	}
	sv.Traits = []string{fmt.Sprintf("sketch: %d file%s, %d lines.", sv.Files, plural(sv.Files), sv.Lines)}
	return sv
}

// --- split ---

// zoneEntries are the cuttable parts of a territory: its prefixes when it holds two or more, else
// what lives under its one prefix.
func (w *World) zoneEntries(c *Creature) []string {
	terr := cleanPrefixes(c.Territory)
	if len(terr) >= 2 {
		return terr
	}
	if len(terr) == 1 {
		if a, b, ok := w.splitByTree(terr[0]); ok {
			out := append(append([]string(nil), a...), b...)
			sort.Strings(out)
			return out
		}
	}
	return nil
}

// clusters counts, per zone entry, the creature's own Thoughts and its failed runs' paths that
// fall under it.
func (w *World) clusters(c *Creature, desk []deskEntry, entries []string) map[string]int {
	terr := cleanPrefixes(c.Territory)
	n := map[string]int{}
	for _, d := range desk {
		if d.Mind != "" {
			continue
		}
		hit := map[string]bool{}
		for _, p := range pathsIn(d.Text) {
			if e := entryOf(p, entries, terr); e != "" {
				hit[e] = true
			}
		}
		for e := range hit {
			n[e]++
		}
	}
	def := w.ValueOpts().Def
	for _, r := range w.ownRuns(c.Name) {
		if def.Outcome(r.Status) != value.Fail && r.Gate != "fail" {
			continue
		}
		hit := map[string]bool{}
		for _, p := range r.Paths {
			if e := entryOf(p, entries, terr); e != "" {
				hit[e] = true
			}
		}
		for e := range hit {
			n[e]++
		}
	}
	return n
}

// mayCut is whether a rank splits: an author (slime) or a gate that commands (orc).
func (w *World) mayCut(c *Creature) (string, bool) {
	fam := w.Ranks.FamilyOf(c.Rank)
	return fam, fam == "author" || fam == "gate"
}

func (w *World) splitEvidence(c *Creature, desk []deskEntry) (string, bool) {
	fam, ok := w.mayCut(c)
	if !ok {
		return "its rank does not split (family " + fam + ")", false
	}
	entries := w.zoneEntries(c)
	if len(entries) < 2 {
		return "its zone holds nothing to cut in two", false
	}
	o := w.Evolution()
	sv := w.survey(cleanPrefixes(c.Territory))
	large := (o.SplitLines > 0 && sv.Lines >= o.SplitLines) || (o.SplitFiles > 0 && sv.Files >= o.SplitFiles)
	if !large {
		return fmt.Sprintf("its territory is not large (%d lines, %d files; splitLines %d, splitFiles %d)", sv.Lines, sv.Files, o.SplitLines, o.SplitFiles), false
	}
	n := w.clusters(c, desk, entries)
	var hot []string
	for _, e := range entries {
		if n[e] > 0 {
			hot = append(hot, fmt.Sprintf("%s ×%d", e, n[e]))
		}
	}
	if len(hot) < 2 {
		return fmt.Sprintf("its territory is large (%d lines, %d files) but its failures and Thoughts do not cluster in two sub-directories", sv.Lines, sv.Files), false
	}
	return fmt.Sprintf("its territory is large (%d lines, %d files) and its failures and Thoughts cluster in %d sub-directories (%s)", sv.Lines, sv.Files, len(hot), strings.Join(hot, ", ")), true
}

var reSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slugOf(s string) string {
	return strings.Trim(reSlug.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func (w *World) planSplit(p *EvolvePlan, c *Creature, desk []deskEntry) error {
	entries := w.zoneEntries(c)
	terr := cleanPrefixes(c.Territory)
	n := w.clusters(c, desk, entries)
	ranked := append([]string(nil), entries...)
	sort.SliceStable(ranked, func(i, j int) bool {
		if n[ranked[i]] != n[ranked[j]] {
			return n[ranked[i]] > n[ranked[j]]
		}
		return ranked[i] < ranked[j]
	})
	halves := [2][]string{{ranked[0]}, {ranked[1]}}
	lines := [2]int{w.survey([]string{ranked[0]}).Lines, w.survey([]string{ranked[1]}).Lines}
	rest := append([]string(nil), ranked[2:]...)
	sort.Strings(rest)
	for _, e := range rest {
		l := w.survey([]string{e}).Lines
		i := 0
		if lines[1] < lines[0] || (lines[1] == lines[0] && len(halves[1]) < len(halves[0])) {
			i = 1
		}
		halves[i] = append(halves[i], e)
		lines[i] += l
	}
	sort.Strings(halves[0])
	sort.Strings(halves[1])
	rk, _ := w.Ranks.Get(c.Rank)
	short := filepath.Base(c.Dir)
	rankDir := filepath.ToSlash(filepath.Dir(c.Dir))
	free := func(sfx [2]string) bool {
		if sfx[0] == "" || sfx[1] == "" || sfx[0] == sfx[1] {
			return false
		}
		for _, s := range sfx {
			if w.Creature(rk.Prefix+short+"-"+s) != nil {
				return false
			}
			if _, err := os.Stat(filepath.Join(w.Root, filepath.FromSlash(rankDir+"/"+short+"-"+s))); err == nil {
				return false
			}
		}
		return true
	}
	sfx := [2]string{"a", "b"}
	if len(halves[0]) == 1 && len(halves[1]) == 1 {
		leaf := [2]string{slugOf(filepath.Base(strings.TrimSuffix(halves[0][0], "/"))), slugOf(filepath.Base(strings.TrimSuffix(halves[1][0], "/")))}
		if free(leaf) {
			sfx = leaf
		}
	}
	if !free(sfx) {
		return fmt.Errorf("the young's names (%s%s-%s, %s%s-%s) are taken", rk.Prefix, short, sfx[0], rk.Prefix, short, sfx[1])
	}
	names := [2]string{rk.Prefix + short + "-" + sfx[0], rk.Prefix + short + "-" + sfx[1]}
	// the Thoughts follow their soil; the ones that name none go to the half carrying fewer
	var thoughts [2][]deskEntry
	var loose []deskEntry
	for _, d := range desk {
		if d.Mind != "" {
			continue // a worn mind's desk stays in the mind both young wear
		}
		half := -1
		for _, pth := range pathsIn(d.Text) {
			if e := entryOf(pth, entries, terr); e != "" {
				for i := range halves {
					if contains(halves[i], e) {
						half = i
					}
				}
				break
			}
		}
		if half < 0 {
			loose = append(loose, d)
			continue
		}
		thoughts[half] = append(thoughts[half], d)
	}
	for _, d := range loose {
		i := 0
		if len(thoughts[1]) < len(thoughts[0]) {
			i = 1
		}
		thoughts[i] = append(thoughts[i], d)
	}
	old := w.planned(p, c.Doc)
	attic := w.atticPath(p, c.Name)
	w.planWrite(p, attic, old)
	date := time.Now().UTC().Format("2006-01-02")
	for i := range names {
		sv := w.survey(halves[i])
		doc := splitChildDoc(old, names[i], names[1-i], c.Name, halves[i], sv, thoughts[i], p.Reason, attic, date)
		rel := rankDir + "/" + short + "-" + sfx[i] + "/README.md"
		w.planWrite(p, rel, doc)
		p.Into = append(p.Into, names[i])
		p.Territory[names[i]] = halves[i]
		p.Verify[names[i]] = VerifyOf(doc)
	}
	// an orc's slimes are divided with it: each reports to the young whose half holds its soil
	if fam, _ := w.mayCut(c); fam == "gate" {
		var kids []Creature
		for _, k := range w.Roster() {
			if k.Parent == c.Name && k.Doc != "" {
				kids = append(kids, k)
			}
		}
		sort.Slice(kids, func(i, j int) bool { return kids[i].Name < kids[j].Name })
		count := [2]int{}
		for _, k := range kids {
			half := -1
			for i := range halves {
				all := len(k.Territory) > 0
				for _, t := range k.Territory {
					in := false
					for _, h := range halves[i] {
						if under(normRel(t), strings.TrimSuffix(h, "/")) {
							in = true
						}
					}
					all = all && in
				}
				if all {
					half = i
				}
			}
			if half < 0 {
				half = 0
				if count[1] < count[0] {
					half = 1
				}
			}
			count[half]++
			w.planWrite(p, k.Doc, setField(w.planned(p, k.Doc), "Reports to", names[half]))
			p.Territory[k.Name] = k.Territory
		}
	}
	retire := w.planned(p, c.Doc)
	if !strings.HasSuffix(retire, "\n") {
		retire += "\n"
	}
	retire += retiredMark + " — evolved " + date + ": split into " + names[0] + " + " + names[1] + " (" + p.Reason + "); the genome rests in the attic\n"
	w.planWrite(p, c.Doc, retire)
	return nil
}

// splitChildDoc is a young one's card: the parent's identity fields, its half of the soil, the
// code's reading of that half, the parent's traits that fit it, its verify, its Thoughts.
func splitChildDoc(parent, name, twin, from string, half []string, sv TerritorySurvey, thoughts []deskEntry, reason, attic, date string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", name)
	fmt.Fprintf(&b, "- **Rank:** %s\n", or(docField(parent, "Rank"), "Slime"))
	ticks := make([]string, len(half))
	for i, h := range half {
		ticks[i] = "`" + h + "`"
	}
	fmt.Fprintf(&b, "- **Territory:** %s\n", strings.Join(ticks, ", "))
	if r := docField(parent, "Reports to"); r != "" {
		fmt.Fprintf(&b, "- **Reports to:** %s\n", r)
	}
	fmt.Fprintf(&b, "- **Minds:** %s\n", or(docField(parent, "Minds"), "none"))
	purpose := strings.TrimSpace(sv.Purpose)
	if purpose == "" {
		purpose = or(docField(parent, "Purpose"), "the ground truth of its zone")
	}
	fmt.Fprintf(&b, "- **Purpose:** %s\n", purpose)
	fmt.Fprintf(&b, "- **Born:** %s — evolved from %s by a split (%s); twin %s; the parent's genome rests at %s\n", date, from, reason, twin, attic)
	b.WriteString("\n## Traits\n")
	n := 0
	for _, t := range sv.Traits {
		fmt.Fprintf(&b, "- %s\n", t)
		n++
	}
	for _, t := range sectionBullets(parent, "Traits") {
		if strings.HasPrefix(t, "sketch:") {
			continue
		}
		fits := true
		for _, pth := range pathsIn(t) {
			in := false
			for _, h := range half {
				if under(normRel(pth), strings.TrimSuffix(h, "/")) {
					in = true
				}
			}
			fits = fits && in
		}
		if fits {
			fmt.Fprintf(&b, "- %s\n", t)
			n++
		}
	}
	if n == 0 {
		b.WriteString("- (none recorded yet: the first session that works here writes what it learns)\n")
	}
	verify := sv.Verify
	if len(verify) == 0 {
		verify = VerifyOf(parent)
	}
	if len(verify) > 0 {
		b.WriteString("\n## Verify\n")
		for _, v := range verify {
			fmt.Fprintf(&b, "- `%s`\n", v)
		}
	}
	b.WriteString("\n## Thoughts\n")
	for _, t := range thoughts {
		b.WriteString(t.Text + "\n")
	}
	return b.String()
}

// --- rebirth ---

func (w *World) rebirthEvidence(c *Creature) (string, bool) {
	var why []string
	runs := w.ownRuns(c.Name)
	for i := len(runs) - 1; i >= 0; i-- {
		g := runs[i].Gate
		if g != "pass" && g != "fail" {
			continue
		}
		if g == "fail" && runs[i].VerifyRed > 0 {
			why = append(why, fmt.Sprintf("its gate failed on its own verify lines (%d of %d red at its last verdict)", runs[i].VerifyRed, runs[i].VerifyRun))
		}
		break
	}
	var gone []string
	for _, t := range cleanPrefixes(c.Territory) {
		if _, err := os.Stat(filepath.Join(w.Root, filepath.FromSlash(strings.TrimSuffix(t, "/")))); err != nil {
			gone = append(gone, t)
		}
	}
	text := w.docText(c)
	for _, t := range sectionBullets(text, "Traits") {
		for _, pth := range pathsIn(t) {
			top := strings.SplitN(pth, "/", 2)[0]
			if fi, err := os.Stat(filepath.Join(w.Root, top)); err != nil || !fi.IsDir() {
				continue // not a path of this tree
			}
			if _, err := os.Stat(filepath.Join(w.Root, filepath.FromSlash(strings.TrimSuffix(pth, "/")))); err != nil && !contains(gone, pth) {
				gone = append(gone, pth)
			}
		}
	}
	if len(gone) > 0 {
		why = append(why, "its doc names soil that is gone ("+strings.Join(gone, ", ")+")")
	}
	if len(why) == 0 {
		return "its genome holds (its verify was green at its last verdict, every path its doc names stands)", false
	}
	return strings.Join(why, "; "), true
}

func (w *World) docText(c *Creature) string {
	b, _ := os.ReadFile(filepath.Join(w.Root, filepath.FromSlash(c.Doc)))
	return string(b)
}

func (w *World) planRebirth(p *EvolvePlan, c *Creature, desk []deskEntry) error {
	old := w.planned(p, c.Doc)
	attic := w.atticPath(p, c.Name)
	w.planWrite(p, attic, old)
	terr := cleanPrefixes(c.Territory)
	sv := w.survey(terr)
	fams, singles := families(desk, c)
	date := time.Now().UTC().Format("2006-01-02")
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", c.Name)
	fmt.Fprintf(&b, "- **Rank:** %s\n", or(docField(old, "Rank"), c.Rank))
	if a := docField(old, "Ascendant of"); a != "" {
		fmt.Fprintf(&b, "- **Ascendant of:** %s\n", a)
	}
	if t := docField(old, "Territory"); t != "" {
		fmt.Fprintf(&b, "- **Territory:** %s\n", t)
	}
	if r := docField(old, "Reports to"); r != "" {
		fmt.Fprintf(&b, "- **Reports to:** %s\n", r)
	}
	fmt.Fprintf(&b, "- **Minds:** %s\n", or(docField(old, "Minds"), "none"))
	fmt.Fprintf(&b, "- **Purpose:** %s\n", or(docField(old, "Purpose"), or(sv.Purpose, "the ground truth of its zone")))
	fmt.Fprintf(&b, "- **Reborn:** %s — its old genome stood stale (%s); it rests at %s\n", date, p.Reason, attic)
	b.WriteString("\n## Traits\n")
	n := 0
	for _, t := range sv.Traits {
		fmt.Fprintf(&b, "- %s\n", t)
		n++
	}
	for _, f := range fams {
		fmt.Fprintf(&b, "- %s\n", f.trait(date))
		n++
	}
	if n == 0 {
		b.WriteString("- (none recorded yet: the first session that works here writes what it learns)\n")
	}
	verify := sv.Verify
	if len(verify) == 0 {
		verify = VerifyOf(old)
	}
	if len(verify) > 0 {
		b.WriteString("\n## Verify\n")
		for _, v := range verify {
			fmt.Fprintf(&b, "- `%s`\n", v)
		}
	}
	distilled := len(desk) - countMind(desk)
	b.WriteString("\n## Thoughts\n")
	fmt.Fprintf(&b, "- %s reborn: %d thought%s distilled into the traits above and the log's wrap-up (evolution)\n", date, distilled, plural(distilled))
	doc := b.String()
	w.planWrite(p, c.Doc, doc)
	w.cutMindDesks(p, c, desk, date)
	p.Rebirth = true
	p.Into = []string{c.Name}
	p.Territory[c.Name] = c.Territory
	p.Verify[c.Name] = VerifyOf(doc)
	for _, s := range singles {
		p.WrapUp = append(p.WrapUp, cleanThought(s.Text))
	}
	return nil
}

func countMind(desk []deskEntry) int {
	n := 0
	for _, d := range desk {
		if d.Mind != "" {
			n++
		}
	}
	return n
}

// cutMindDesks empties the worn minds' desks of what the plan distilled, each keeping one dated
// line naming where its Thoughts went.
func (w *World) cutMindDesks(p *EvolvePlan, c *Creature, desk []deskEntry, date string) {
	by := map[string][]memory.Thought{}
	var srcs []string
	for _, d := range desk {
		if d.Mind == "" {
			continue
		}
		if _, ok := by[d.Src]; !ok {
			srcs = append(srcs, d.Src)
		}
		by[d.Src] = append(by[d.Src], d.Thought)
	}
	sort.Strings(srcs)
	for _, src := range srcs {
		line := fmt.Sprintf("- %s distilled into %s's genome: %d thought%s (evolution)", date, c.Name, len(by[src]), plural(len(by[src])))
		w.planWrite(p, src, cutDesk(w.planned(p, src), by[src], line))
	}
}

// --- mutate ---

// family is a lesson the desk repeats: the Thoughts that share a term (a path, or a word).
type family struct {
	Term    string
	Path    bool
	Members []deskEntry
	How     bool // most of its Thoughts say how to do something
}

var (
	reDeskHead  = regexp.MustCompile(`^\s*(?:[-*]|###)\s*(?:\[\s*)?\d{4}-\d{2}-\d{2}(?:[T ][0-9:.Z+-]+)?(?:\s*\])?\s*[—:–-]?\s*`)
	reDates     = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	reBackCmd   = regexp.MustCompile("`[^`\n]*\\s[^`\n]*`")
	howWords    = regexp.MustCompile(`(?i)\b(run|use|call|always|never|first|then|before|after|step|must|how to|instead|pass|set)\b`)
	termSkip    = map[string]bool{"thought": true, "thoughts": true, "note": true, "learned": true, "today": true, "still": true, "need": true, "needs": true, "found": true, "always": true, "never": true, "again": true, "instead": true, "must": true, "use": true, "run": true, "first": true, "step": true, "set": true, "pass": true, "call": true}
	harnessLine = regexp.MustCompile(`(?i)distilled into|\breborn:|born of the split|\(evolution\)`)
)

// cleanThought is a Thought's text without its bullet and date head, on one line.
func cleanThought(text string) string {
	first, rest, _ := strings.Cut(text, "\n")
	first = reDeskHead.ReplaceAllString(first, "")
	return oneLine(strings.TrimSpace(first + " " + rest))
}

func termsOf(text string, skip map[string]bool) (words, paths []string) {
	clean := reDates.ReplaceAllString(cleanThought(text), " ")
	seen := map[string]bool{}
	for _, p := range pathsIn(clean) {
		p = strings.TrimSuffix(strings.ToLower(p), "/")
		if !seen[p] {
			seen[p] = true
			paths = append(paths, p)
		}
	}
	for _, t := range memory.Tokens(clean) {
		if len(t) < 3 || skip[t] || termSkip[t] || seen[t] || strings.Contains(t, "/") {
			continue
		}
		seen[t] = true
		words = append(words, t)
	}
	return words, paths
}

// families reads the desk's lessons: greedily, the term the most Thoughts share (a path before a
// word, a longer term before a shorter, then the alphabet), its Thoughts one family, again over
// what is left, while two or more still share one. The rest are singles (one naming is an
// observation, Nature 4): they go to the log's wrap-up. A harness line (a former distillation)
// is neither.
func families(desk []deskEntry, c *Creature) ([]family, []deskEntry) {
	skip := map[string]bool{}
	for _, part := range strings.Split(c.Name, "-") {
		skip[part] = true
	}
	skip[c.Name] = true
	type item struct {
		d     deskEntry
		terms map[string]bool
		paths map[string]bool
	}
	var left []item
	for _, d := range desk {
		if harnessLine.MatchString(d.Text) {
			continue
		}
		ws, ps := termsOf(d.Text, skip)
		it := item{d: d, terms: map[string]bool{}, paths: map[string]bool{}}
		for _, x := range ws {
			it.terms[x] = true
		}
		for _, x := range ps {
			it.terms[x] = true
			it.paths[x] = true
		}
		left = append(left, it)
	}
	var out []family
	for {
		count := map[string]int{}
		isPath := map[string]bool{}
		for _, it := range left {
			for t := range it.terms {
				count[t]++
				if it.paths[t] {
					isPath[t] = true
				}
			}
		}
		best := ""
		for t, n := range count {
			if n < 2 {
				continue
			}
			switch {
			case best == "":
				best = t
			case n != count[best]:
				if n > count[best] {
					best = t
				}
			case isPath[t] != isPath[best]:
				if isPath[t] {
					best = t
				}
			case len(t) != len(best):
				if len(t) > len(best) {
					best = t
				}
			case t < best:
				best = t
			}
		}
		if best == "" {
			break
		}
		f := family{Term: best, Path: isPath[best]}
		var rest []item
		how := 0
		for _, it := range left {
			if it.terms[best] {
				f.Members = append(f.Members, it.d)
				if reBackCmd.MatchString(it.d.Text) || howWords.MatchString(cleanThought(it.d.Text)) {
					how++
				}
			} else {
				rest = append(rest, it)
			}
		}
		f.How = how*2 >= len(f.Members)
		out = append(out, f)
		left = rest
	}
	var singles []deskEntry
	for _, it := range left {
		singles = append(singles, it.d)
	}
	return out, singles
}

// trait is a family distilled: its term, its latest wording, its record.
func (f family) trait(date string) string {
	latest := f.Members[0]
	for _, m := range f.Members {
		if m.Date >= latest.Date {
			latest = m
		}
	}
	text := cleanThought(latest.Text)
	if r := []rune(text); len(r) > 240 {
		text = string(r[:239]) + "…"
	}
	var dates []string
	for _, m := range f.Members {
		dates = append(dates, m.Date)
	}
	sort.Strings(dates)
	term := f.Term
	if f.Path {
		term = "`" + term + "`"
	}
	return fmt.Sprintf("%s — %s (distilled %s from %d thoughts, %s…%s)", term, text, date, len(f.Members), dates[0], dates[len(dates)-1])
}

func (w *World) planMutate(p *EvolvePlan, c *Creature, desk []deskEntry) error {
	o := w.Evolution()
	fams, singles := families(desk, c)
	date := time.Now().UTC().Format("2006-01-02")
	text := w.planned(p, c.Doc)
	var own []memory.Thought
	for _, d := range desk {
		if d.Mind == "" {
			own = append(own, d.Thought)
		}
	}
	var traits, targets []string
	for _, f := range fams {
		traits = append(traits, "- "+f.trait(date))
	}
	if len(traits) > 0 {
		targets = append(targets, fmt.Sprintf("%d trait%s", len(traits), plural(len(traits))))
	}
	// a verify line when tests cover the paths a lesson names (and the doc does not run them yet)
	have := VerifyOf(text)
	var added []string
	terr := cleanPrefixes(c.Territory)
	for _, f := range fams {
		for _, m := range f.Members {
			for _, pth := range pathsIn(m.Text) {
				dir := pth
				if filepath.Ext(dir) != "" {
					dir = filepath.ToSlash(filepath.Dir(dir))
				}
				dir = strings.TrimSuffix(dir, "/")
				full := ""
				for _, cand := range append([]string{dir}, prefixed(terr, dir)...) {
					if fi, err := os.Stat(filepath.Join(w.Root, filepath.FromSlash(cand))); err == nil && fi.IsDir() && w.ownsPath(c, cand) {
						full = cand
						break
					}
				}
				if full == "" {
					continue
				}
				for _, v := range w.survey([]string{full + "/"}).Verify {
					if !contains(have, v) && !contains(added, v) && len(added) < 2 {
						added = append(added, v)
					}
				}
			}
		}
	}
	// a recurring how becomes a crafted skill the creature wears
	minds := docField(text, "Minds")
	for _, f := range fams {
		if !f.How || len(f.Members) < 2 {
			continue
		}
		name := skillNameFor(c.Name, f.Term)
		if name == "" || w.mindDoc(name) != "" {
			continue
		}
		rel := filepath.ToSlash(filepath.Join(w.Lex.WorldDir, "crafted-skills", name, "SKILL.md"))
		w.planWrite(p, rel, evolvedSkill(name, c, f, or(o.SkillStatus, "trial"), p.ID, date))
		if minds == "" || strings.EqualFold(minds, "none") {
			minds = name
		} else {
			minds += ", " + name
		}
		p.Skill = name
		targets = append(targets, "the crafted skill "+name)
		break
	}
	var vlines []string
	for _, v := range added {
		vlines = append(vlines, "- `"+v+"`")
	}
	if len(added) > 0 {
		targets = append(targets, fmt.Sprintf("%d verify line%s", len(added), plural(len(added))))
	}
	if len(singles) > 0 || len(targets) == 0 {
		targets = append(targets, "the log's wrap-up")
	}
	// the desk's distilled entries leave (cut first: the offsets are the doc's as read); one dated
	// line says where they went; then the genome takes the traits, the verify, the worn skill
	cut := cutDesk(text, own, fmt.Sprintf("- %s distilled into %s: %d thought%s (evolution)", date, strings.Join(targets, ", "), len(desk), plural(len(desk))))
	cut = appendSection(cut, "Traits", traits)
	cut = appendSection(cut, "Verify", vlines)
	if p.Skill != "" {
		cut = setField(cut, "Minds", minds)
	}
	w.planWrite(p, c.Doc, cut)
	w.cutMindDesks(p, c, desk, date)
	for _, s := range singles {
		p.WrapUp = append(p.WrapUp, cleanThought(s.Text))
	}
	p.Into = []string{c.Name}
	p.Territory[c.Name] = c.Territory
	if len(added) > 0 {
		p.Verify[c.Name] = added
	}
	return nil
}

// prefixed is dir under each territory prefix (a Thought names `api/` for `src/pay/api/`).
func prefixed(terr []string, dir string) []string {
	var out []string
	for _, t := range terr {
		out = append(out, strings.TrimSuffix(t, "/")+"/"+dir)
	}
	return out
}

func (w *World) ownsPath(c *Creature, rel string) bool {
	for _, t := range c.Territory {
		if under(normRel(rel), strings.TrimSuffix(strings.Trim(t, "` "), "/")) {
			return true
		}
	}
	return false
}

var reSkillName = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)

// skillNameFor is the crafted skill a lesson becomes: the creature's name and the lesson's term —
// a race-prefixed Mind, the body's own hat.
func skillNameFor(body, term string) string {
	t := slugOf(strings.TrimSuffix(filepath.Base(term), filepath.Ext(term)))
	if t == "" {
		return ""
	}
	name := body + "-" + t
	if len(name) > 64 {
		name = strings.Trim(name[:64], "-")
	}
	if !reSkillName.MatchString(name) {
		return ""
	}
	return name
}

// evolvedSkill is the SKILL.md the mutate arm mints: the host's frontmatter (name, description)
// and the world's fields (version, status, by, born, minted), the lesson's steps as its body.
func evolvedSkill(name string, c *Creature, f family, status, id, date string) string {
	terr := strings.Join(cleanPrefixes(c.Territory), ", ")
	desc := fmt.Sprintf("Working on %s in %s: the steps %s learned the hard way (%d times)", f.Term, terr, c.Name, len(f.Members))
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: " + name + "\n")
	b.WriteString("description: " + strconv.Quote(desc) + "\n")
	b.WriteString("version: 1\n")
	b.WriteString("status: " + status + "\n")
	b.WriteString("by: " + strconv.Quote(c.Name) + "\n")
	b.WriteString("born: " + strconv.Quote("evolution "+id) + "\n")
	b.WriteString("minted: {by: \"law.evolution\", at: " + strconv.Quote(time.Now().UTC().Format(time.RFC3339)) + "}\n")
	b.WriteString("---\n\n")
	b.WriteString("# " + name + "\n\n")
	fmt.Fprintf(&b, "Distilled by evolution on %s from %d Thoughts of %s that kept naming %s.\n\n## Steps\n", date, len(f.Members), c.Name, f.Term)
	ms := append([]deskEntry(nil), f.Members...)
	sort.SliceStable(ms, func(i, j int) bool { return ms[i].Date < ms[j].Date })
	for _, m := range ms {
		fmt.Fprintf(&b, "- %s\n", cleanThought(m.Text))
	}
	b.WriteString("\n## Thoughts\n")
	return b.String()
}
