package onto

// Housekeeping — the workspace's own chain law, enforced by code:
// no rank without its chain (a coord without domains, a domain without zones,
// a zone without its truth edge is broken), no creature without a home, no
// overlap left silent. `check` reports; `heal` seals the mechanical classes
// (mint homes, rewire missing rank edges to their law parent, seat uncovered
// hot dirs as draft zones); overlaps and dangling prose bonds are reported,
// never silently rewritten — those want an operator.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Repair is one sealed mechanical fix (file + what).
type Repair struct {
	Path, What string
}

// HouseReport is one housekeeping run: what the machine sealed, what still wants an operator.
type HouseReport struct {
	Repairs   []Repair
	Leftover  []string
	Healed    int
	Spawned   []string
	DeadBonds []string // seats whose Reports-to names a name with no live seat — healable mechanically
	Notes     []string // the after-pass holes (@?)
}

// Housekeep audits the workspace at root and, when heal is set, seals the mechanical classes.
// With spawn it also seats draft zone homes for tracked top-level dirs no creature covers.
func Housekeep(root string, heal, spawn bool) (*HouseReport, []Finding, error) {
	rep := &HouseReport{}
	w, err := Load(root)
	if err != nil {
		return nil, nil, err
	}
	l := w.Layout.norm()
	rep.probeDeadBonds(w, l, heal)
	if !heal && !spawn {
		rep.Notes = w.Notes
		return rep, w.Validate(), nil
	}
	if heal {
		for _, n := range w.Notes {
			if strings.Contains(n, " has no doc in ") {
				if r, ok := healMissingDoc(w, l, n); ok {
					rep.Repairs = append(rep.Repairs, r)
					continue
				}
			}
			rep.Leftover = append(rep.Leftover, n)
		}
		for _, f := range w.Validate() {
			switch {
			case strings.Contains(f.Msg, "wants exactly 1"):
				if r, ok := healMissingEdge(w, l, f); ok {
					rep.Repairs = append(rep.Repairs, r)
				} else {
					rep.Leftover = append(rep.Leftover, f.String())
				}
			default:
				rep.Leftover = append(rep.Leftover, f.String())
			}
		}
	}
	if spawn {
		rep.Spawned = seatUncovered(w, l, root)
	}
	rep.Healed = len(rep.Repairs) + len(rep.Spawned)
	if rep.Healed > 0 {
		journal(root, l, rep)
	}
	w2, err2 := Load(root)
	if err2 != nil {
		return rep, nil, err2
	}
	rep.Notes = w2.Notes
	return rep, w2.Validate(), nil
}

// healMissingDoc mints a creature home scaffold for a note like
// "<ws>/<rank>/<dir>/:1: <creature> has no doc in <ws>/<rank>/<dir>/".
func healMissingDoc(w *World, l Layout, note string) (Repair, bool) {
	i := strings.Index(note, " has no doc in ")
	if i < 0 {
		return Repair{}, false
	}
	dir := strings.TrimSuffix(strings.TrimSpace(note[i+len(" has no doc in "):]), "/")
	head := note[:i]
	if j := strings.Index(head, ": "); j >= 0 {
		head = head[j+2:]
	}
	name := strings.TrimSpace(head)
	subj := Is(name)
	rank := rankOf(w, subj)
	if rank == "" {
		// the creature is a dir-shape ghost the graph never derived; the dir names the rank
		parts := strings.Split(filepath.ToSlash(dir), "/")
		if len(parts) >= 2 {
			rank = parts[len(parts)-2]
		}
	}
	parent := parentCreature(w, l, rank, "")
	if parent == "" {
		return Repair{}, false
	}
	path := filepath.Join(w.Root, dir, "README.md")
	body := fmt.Sprintf("# %s\n\n- **Rank:** %s\n- **Reports to:** %s\n- **Purpose:** seated by housekeep %s; describe the desk and rewrite this line.\n\n## Traits\n\n1. This doc was minted by the chain law; its traits and verify lines wait on their first keeper.\n\n## Verify\n\n- `test -f %s`\n\n## Working notes\n\n- %s — home minted by housekeep (missing creature doc).\n",
		name, displayRank(rank), parent, today(), filepath.ToSlash(path), today())
	if err := os.MkdirAll(filepath.Join(w.Root, dir), 0o755); err != nil {
		return Repair{}, false
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return Repair{}, false
	}
	return Repair{Path: dir + "/", What: "minted a creature home (" + rank + ", reports to " + parent + ")"}, true
}

// healMissingEdge sets a creature's Reports-to to its law parent when its chain edge is absent.
func healMissingEdge(w *World, l Layout, f Finding) (Repair, bool) {
	rank := rankOf(w, f.Subject)
	if rank == "" {
		return Repair{}, false
	}
	doc := strings.SplitN(f.Doc, ":", 2)[0]
	parent := parentCreature(w, l, rank, ownsOf(w, f.Subject))
	if parent == "" {
		return Repair{}, false
	}
	b, err := os.ReadFile(filepath.Join(w.Root, doc))
	if err != nil {
		return Repair{}, false
	}
	lines := strings.Split(string(b), "\n")
	set := false
	for i, s := range lines {
		if strings.HasPrefix(s, "- **Reports to:**") {
			lines[i] = "- **Reports to:** " + parent
			set = true
			break
		}
	}
	if !set {
		for i, s := range lines {
			if strings.HasPrefix(s, "- **Rank:**") {
				lines = append(lines[:i+1], append([]string{"- **Reports to:** " + parent}, lines[i+1:]...)...)
				set = true
				break
			}
		}
	}
	if !set {
		return Repair{}, false
	}
	if err := os.WriteFile(filepath.Join(w.Root, doc), []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return Repair{}, false
	}
	return Repair{Path: doc, What: "rewired " + f.Subject.Local() + "'s chain edge to " + parent}, true
}

// seatUncovered mints draft zone homes for tracked top-level dirs no creature covers.
func seatUncovered(w *World, l Layout, root string) []string {
	zrank, zdir := "", ""
	for _, r := range l.Ranks {
		if ClassOf(r.Name).Local() == "Slime" {
			zrank, zdir = r.Name, r.Dir
			break
		}
	}
	var out []string
	for _, d := range trackedDirs(root) {
		covered := false
		for _, own := range allOwns(w) {
			if own == d || strings.HasPrefix(own, d+"/") || strings.HasPrefix(d, own+"/") {
				covered = true
				break
			}
		}
		if covered {
			continue
		}
		base := strings.TrimSuffix(d, "/")
		dir := filepath.Join(w.Root, l.WorldDir, zdir, base)
		if _, err := os.Stat(filepath.Join(dir, "README.md")); err == nil {
			continue
		}
		parent := parentCreature(w, l, zrank, base)
		if parent == "" {
			continue
		}
		name := zrank + "-" + base
		body := fmt.Sprintf("# %s\n\n- **Rank:** %s\n- **Territory:** `%s/`\n- **Reports to:** %s\n- **Purpose:** seated by housekeep --spawn %s (uncovered dir); its keeper refines this desk.\n\n## Traits\n\n1. Draft seat; territory and traits wait on their first keeper's proof.\n\n## Working notes\n\n- %s — seated by housekeep (no creature covered this dir).\n",
			name, displayRank(zrank), base, parent, today(), today())
		if err := os.MkdirAll(dir, 0o755); err != nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(body), 0o644); err != nil {
			continue
		}
		out = append(out, name+" → "+parent)
	}
	return out
}

// — helpers

// rankOf answers a creature's rank word from its graph class ("Slime" → "zone").
func rankOf(w *World, m Term) string {
	for _, t := range w.Graph.Objects(m, rdfType) {
		cl := t.Local()
		for _, r := range w.Layout.norm().Ranks {
			if ClassOf(r.Name).Local() == cl {
				return r.Name
			}
		}
	}
	return ""
}

// ownsOf answers the creature's first asserted owns path (used for parent matching).
func ownsOf(w *World, m Term) string {
	for _, t := range w.Graph.Objects(m, pOwns) {
		if t.Kind == Literal {
			return t.Value
		}
	}
	return ""
}

// allOwns lists every asserted owns path.
func allOwns(w *World) []string {
	var out []string
	for _, tr := range w.Graph.All() {
		if tr.P == pOwns && tr.O.Kind == Literal {
			out = append(out, tr.O.Value)
		}
	}
	sort.Strings(out)
	return out
}

// parentCreature picks the concrete parent creature for a rank (containment match wins, else first by name).
func parentCreature(w *World, l Layout, rank, childOwns string) string {
	want := ""
	for _, r := range l.Ranks {
		if r.Name == rank {
			want = r.Parent
			break
		}
	}
	if want == "" {
		return ""
	}
	if want == "rimuru" {
		return "rimuru"
	}
	prefix := want + "-"
	var cands []string
	best := ""
	for _, tr := range w.Graph.All() {
		if tr.P == pName && tr.O.Kind == Literal && strings.HasPrefix(tr.O.Value, prefix) {
			cands = append(cands, tr.O.Value)
			if childOwns != "" && ownsHas(w, tr.S, childOwns) {
				best = tr.O.Value
			}
		}
	}
	if best != "" {
		return best
	}
	if len(cands) == 0 {
		return ""
	}
	sort.Strings(cands)
	return cands[0]
}

// ownsHas: the creature owns a path containing child (prefix either way).
func ownsHas(w *World, m Term, child string) bool {
	for _, tr := range w.Graph.All() {
		if tr.S == m && tr.P == pOwns && tr.O.Kind == Literal {
			s := tr.O.Value
			if s == child || strings.HasPrefix(child, s+"/") || strings.HasPrefix(s, child+"/") {
				return true
			}
		}
	}
	return false
}

func displayRank(rank string) string {
	// Class labels are lexicon truth; the doc's Rank field is "Domain owner" style.
	s := ClassOf(rank).Local()
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte(' ')
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}

// trackedDirs answers the repo's tracked top-level dirs (git ls-files).
func trackedDirs(root string) []string {
	out, err := exec.Command("git", "-C", root, "ls-files").Output()
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var dirs []string
	for _, ln := range strings.Split(string(out), "\n") {
		i := strings.IndexByte(ln, '/')
		if i <= 0 {
			continue
		}
		d := ln[:i]
		if strings.HasPrefix(d, ".") || seen[d] {
			continue
		}
		seen[d] = true
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return dirs
}

func today() string { return time.Now().Format("2006-01-02") }

// journal appends the run to the world chronicle, house shape.
func journal(root string, l Layout, rep *HouseReport) {
	p := filepath.Join(root, l.WorldDir, "log.md")
	var lines []string
	for _, r := range rep.Repairs {
		lines = append(lines, "  - "+r.Path+" — "+r.What)
	}
	for _, s := range rep.Spawned {
		lines = append(lines, "  - "+s+" (draft seat)")
	}
	entry := "### [" + time.Now().Format("2006-01-02 15:04") + "] housekeep — chain repairs\n\n- Repairs:\n" + strings.Join(lines, "\n")
	if len(rep.Leftover) > 0 {
		entry += "\n- Leftover (operator):\n  - " + strings.Join(rep.Leftover, "\n  - ")
	}
	entry += "\n"
	if _, err := os.Stat(p); err != nil {
		entry = "# Log\n\n" + entry
	}
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprint(f, "\n"+entry)
}

// probeDeadBonds flags every creature whose Reports-to names a name with no
// live seat; a dead bond mechanically heals to the seat's law parent; a live
// one — whatever the law says — is the operator's bond.
func (rep *HouseReport) probeDeadBonds(w *World, l Layout, heal bool) {
	worldDir := filepath.Join(w.Root, l.WorldDir)
	live := map[string]bool{}
	filepath.WalkDir(worldDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "README.md" {
			return nil
		}
		rel, _ := filepath.Rel(worldDir, path)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) == 3 {
			live[parts[1]] = true // <rank>/<name>/README.md -> name
		}
		return nil
	})
	filepath.WalkDir(worldDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "README.md" {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		body := string(b)
		bond := bondOf(body, "Reports to")
		if bond == "" || live[bond] {
			return nil // empty, or a live seat — including a live non-law target, the operator's bond
		}
		rank, name := creatureRankName(body, worldDir, path)
		parent := parentCreature(w, l, rank, ownsLine(body))
		if parent == "" || parent == bond {
			return nil // no law parent for this rank, or already on it — nothing mechanical to seal
		}
		rep.DeadBonds = append(rep.DeadBonds, name)
		if heal {
			lines := strings.Split(body, "\n")
			for i, ln := range lines {
				if strings.HasPrefix(ln, "- **Reports to:** ") {
					lines[i] = "- **Reports to:** " + parent
					break
				}
			}
			if os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644) == nil {
				rep.Repairs = append(rep.Repairs, Repair{Path: l.WorldDir + "/" + rank + "/" + name + "/README.md", What: "rewired " + name + "'s dead bond (" + bond + ") to law parent " + parent})
			} else {
				rep.Notes = append(rep.Notes, "@? "+name+" dead bond unhealed")
			}
		}
		return nil
	})
}

// creatureRankName derives a seat's rank dir and name from its doc — Rank line first, path fallback.
func creatureRankName(body, worldDir, path string) (rank, name string) {
	rel, _ := filepath.Rel(worldDir, path)
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) == 3 {
		rank, name = parts[0], parts[1]
	}
	if r := bondOf(body, "Rank"); r != "" {
		rank = strings.ToLower(r)
	}
	return rank, name
}

// ownsLine reads a doc's `- **Territory:** `x/“ first own, or "" — feeds the law-parent map.
func ownsLine(body string) string {
	for ln := range strings.Lines(body) {
		if strings.HasPrefix(ln, "- **Territory:**") {
			s := strings.TrimSpace(strings.TrimPrefix(ln, "- **Territory:**"))
			return strings.TrimSuffix(strings.Trim(strings.TrimSpace(strings.Split(s, ",")[0]), "`"), "/")
		}
	}
	return ""
}

// bondOf reads `- **<field>:** value` from a creature doc, or "".
func bondOf(body, field string) string {
	for ln := range strings.Lines(body) {
		if strings.HasPrefix(ln, "- **"+field+":** ") {
			return strings.TrimSpace(strings.TrimPrefix(ln, "- **"+field+":** "))
		}
	}
	return ""
}
