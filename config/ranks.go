package config

import (
	"fmt"
	"github.com/Kaginari/isekai/value"
	"sort"
	"strings"
)

// RankDef is a rank as config writes it; every field is optional and overrides the built-in
// table's field of the same name (rankSet: extend) or defines a new rank outright.
type RankDef struct {
	ReportsTo   string     `json:"reportsTo"`
	Job         string     `json:"job"`
	Authors     *bool      `json:"authors"`
	HoldsGate   *bool      `json:"holdsGate"`
	Body        string     `json:"body"` // read by nothing: every dispatched body is a Court (a set value is a hole)
	Office      string     `json:"office"`
	Tools       []string   `json:"tools"` // globs over tool names
	Model       ModelRef   `json:"model"`
	Dir         string     `json:"dir"`
	Prefix      string     `json:"prefix"`
	AscendsFrom string     `json:"ascendsFrom"`
	Sideways    *bool      `json:"sideways"`    // may speak to its own rank (orc ↔ orc) and name its layer-mates
	Layer       int        `json:"layer"`       // a rung, 0 when unset; two ranks share a layer iff equal and nonzero
	DivineRules []string   `json:"divineRules"` // an ascended rank's absolute rules, laced into the base rank's prompt
	AscendsWhen string     `json:"ascendsWhen"` // the ASL condition under which a body of this rank ascends ("" = never)
	Family      string     `json:"family"`      // the value family (rank.Families); unset = derived from the flags
	Portrait    string     `json:"portrait"`    // a portrait file, relative to the world root, before the embedded one
	Value       *RankValue `json:"value"`       // the race's own value law over its family's
}

// Rank is one resolved rank with the origin of each field ("law" for the built-in table).
type Rank struct {
	Name        string
	ReportsTo   string
	Job         string
	Authors     bool
	HoldsGate   bool
	Office      string
	Tools       []string
	Model       ModelRef
	Dir         string
	Prefix      string
	AscendsFrom string
	Sideways    bool
	Layer       int
	DivineRules []string
	AscendsWhen string
	Family      string
	Portrait    string
	// Value is the race's resolved value law: its own score/meter/minRecord over its family's.
	Value   value.Spec
	Origins map[string]string // field → origin (value.score, value.meter, value.minRecord, family included)
	Builtin bool
}

// RootRank is the session's rank: the tree's root, never configured.
const RootRank = "rimuru"

// The default ascension conditions: a long record (twenty reports), nearly all landed (a pass rate
// of 90% or more), nearly none leaving a hole (at most one per ten reports). world.DefaultRanks
// carries the same strings; `ranks.high-elf.ascendsWhen` / `ranks.high-orc.ascendsWhen` override
// them, and an empty string turns ascension off.
const (
	DefaultAscendElf = "ascend elf when reports >= 20 && passRate >= 90% && holeRate <= 10%"
	DefaultAscendOrc = "ascend orc when reports >= 20 && passRate >= 90% && holeRate <= 10%"
)

// RenderAscend is the shipped ascension sentence for a base rank with law.ascend's green-floor
// knobs folded in: a landRate floor and a tenure, each clause only when its knob is set — zero
// knobs render the shipped sentence itself.
func RenderAscend(base string, k AscendLaw) string {
	s := "ascend " + base + " when reports >= 20 && passRate >= 90% && holeRate <= 10%"
	if k.GreenFloor > 0 {
		s += fmt.Sprintf(" && landRate >= %d%%", k.GreenFloor)
	}
	if k.Tenure > 0 {
		s += fmt.Sprintf(" && tenure >= %d", k.Tenure)
	}
	return s
}

// applyAscendKnobs folds law.ascend into every ascended rank still carrying the shipped
// sentence (a world that wrote its own ascendsWhen keeps it whole).
func (c *Config) applyAscendKnobs(byName map[string]*Rank, order []string) error {
	k := c.Law.Ascend
	if k.GreenFloor < 0 || k.GreenFloor > 100 {
		return fmt.Errorf("%s: law.ascend.greenFloor %d is a percent, 0..100", c.Where("law.ascend.greenFloor"), k.GreenFloor)
	}
	if k.Tenure < 0 {
		return fmt.Errorf("%s: law.ascend.tenure %d is a count of days, 0 or more", c.Where("law.ascend.tenure"), k.Tenure)
	}
	if k == (AscendLaw{}) {
		return nil
	}
	where := c.Where("law.ascend.greenFloor")
	if k.GreenFloor == 0 {
		where = c.Where("law.ascend.tenure")
	}
	for _, name := range order {
		r := byName[name]
		if r.AscendsFrom == "" || r.Origins["ascendsWhen"] != "law" || r.AscendsWhen != RenderAscend(r.AscendsFrom, AscendLaw{}) {
			continue
		}
		r.AscendsWhen = RenderAscend(r.AscendsFrom, k)
		r.Origins["ascendsWhen"] = where
	}
	return nil
}

// builtinRanks is isekai.md §The world as data.
var builtinRanks = []Rank{
	{Name: "elf", ReportsTo: "rimuru", Job: "the shared mind and voice: thinks across domains, routes work to orcs, speaks back up", Office: "ciel", Tools: []string{"*"}, Dir: "elf", Prefix: "elf-"},
	{Name: "orc", ReportsTo: "elf", Job: "rules its domain, holds the gate, commands and validates its slimes", HoldsGate: true, Sideways: true, Office: "raphael", Tools: []string{"*"}, Dir: "orc", Prefix: "orc-"},
	{Name: "slime", ReportsTo: "orc", Job: "the ground truth of its zone; authors changes there", Authors: true, Office: "great-sage", Tools: []string{"*"}, Dir: "slime", Prefix: "slime-"},
	{Name: "kijin", ReportsTo: "rimuru", Job: "a standing domain lead owning one subsystem end-to-end across sessions", Authors: true, HoldsGate: true, Office: "raphael", Tools: []string{"*"}, Dir: "kijin", Prefix: "kijin-"},
	{Name: "high-elf", AscendsFrom: "elf", Job: "ascended elf: holds the accumulated cross-domain traits", Dir: "high-elf", Prefix: "high-elf-", AscendsWhen: DefaultAscendElf},
	{Name: "high-orc", AscendsFrom: "orc", Job: "ascended orc: holds the accumulated domain traits", Dir: "high-orc", Prefix: "high-orc-", AscendsWhen: DefaultAscendOrc},
	{Name: "dark-elf", ReportsTo: "rimuru", Job: "the auditor: reviews gate verdicts and log.md for law violations; never authors", Office: "raphael", Tools: []string{"read", "glob", "grep", "ls", "recall", "onto", "toolbox", "git", "ask"}, Dir: "dark-elf", Prefix: "dark-elf-"},
}

// RankNames lists the built-in ranks in law order.
func RankNames() []string {
	var out []string
	for _, r := range builtinRanks {
		out = append(out, r.Name)
	}
	return out
}

func (c *Config) buildRanks() error {
	byName := map[string]*Rank{}
	var order []string
	if c.RankSet == "" {
		c.RankSet = "extend"
	}
	if c.RankSet != "extend" && c.RankSet != "replace" {
		return fmt.Errorf("%s: rankSet: %q is not extend or replace", c.Where("rankSet"), c.RankSet)
	}
	if c.RankSet == "extend" {
		for _, r := range builtinRanks {
			rr := r
			rr.Builtin = true
			rr.Origins = map[string]string{}
			for _, f := range rankFields {
				rr.Origins[f] = "law"
			}
			byName[rr.Name] = &rr
			order = append(order, rr.Name)
		}
	}
	for _, name := range sortedKeys(c.RankDefs) {
		if strings.EqualFold(name, RootRank) {
			return fmt.Errorf("%s: ranks.%s: the root rank is the session and is not configured", c.Where("ranks."+name), name)
		}
		def := c.RankDefs[name]
		r, ok := byName[name]
		if !ok {
			r = &Rank{Name: name, Origins: map[string]string{}}
			byName[name] = r
			order = append(order, name)
		}
		set := func(field string, apply func()) {
			path := "ranks." + name + "." + field
			if _, ok := c.Origins[path]; ok {
				apply()
				r.Origins[field] = c.Where(path)
			}
		}
		set("reportsTo", func() { r.ReportsTo = canonicalRank(def.ReportsTo) })
		set("job", func() { r.Job = def.Job })
		set("authors", func() { r.Authors = *def.Authors })
		set("holdsGate", func() { r.HoldsGate = *def.HoldsGate })
		if _, ok := c.Origins["ranks."+name+".body"]; ok {
			// a persistent body does not exist: every dispatched body is a Court with a fresh
			// context, a kijin's included — the field is accepted and named a hole, never obeyed
			c.Holes = append(c.Holes, fmt.Sprintf("ranks.%s.body is ignored — every dispatched body runs as a Court with a fresh context (%s)", name, c.Where("ranks."+name+".body")))
		}
		set("office", func() { r.Office = def.Office })
		set("tools", func() { r.Tools = def.Tools })
		set("model", func() { r.Model = def.Model })
		set("dir", func() { r.Dir = def.Dir })
		set("prefix", func() { r.Prefix = def.Prefix })
		set("ascendsFrom", func() { r.AscendsFrom = canonicalRank(def.AscendsFrom) })
		set("sideways", func() { r.Sideways = *def.Sideways })
		set("layer", func() { r.Layer = def.Layer })
		set("divineRules", func() { r.DivineRules = def.DivineRules })
		set("ascendsWhen", func() { r.AscendsWhen = def.AscendsWhen })
		set("family", func() { r.Family = strings.ToLower(strings.TrimSpace(def.Family)) })
		set("portrait", func() { r.Portrait = def.Portrait })
		// models.ranks.<r> and ranks.<r>.model are one slot.
		if mr, ok := c.Models.Ranks[name]; ok && def.Model.Model != "" && mr.Model != def.Model.Model {
			return fmt.Errorf("%s: ranks.%s.model %q conflicts with models.ranks.%s %q (%s) — one slot, one value",
				c.Where("ranks."+name+".model"), name, def.Model.Model, name, mr.Model, c.Where("models.ranks."+name))
		}
	}
	// Ascended ranks inherit every base field they did not set; dropping one is refused.
	for _, name := range order {
		r := byName[name]
		if r.AscendsFrom == "" {
			continue
		}
		base, ok := byName[r.AscendsFrom]
		if !ok {
			return fmt.Errorf("ranks.%s.ascendsFrom: no rank %q (%s)", name, r.AscendsFrom, c.rankWhere(name, "ascendsFrom"))
		}
		if base.AscendsFrom != "" {
			return fmt.Errorf("ranks.%s.ascendsFrom: %q is itself ascended", name, r.AscendsFrom)
		}
		// A field config did not set is inherited; a held capability config drops is refused.
		inherit := func(field string, apply func()) {
			if o := r.Origins[field]; o == "" || o == "law" {
				apply()
				r.Origins[field] = "ascends:" + base.Name
			}
		}
		inherit("reportsTo", func() { r.ReportsTo = base.ReportsTo })
		inherit("authors", func() { r.Authors = base.Authors })
		inherit("holdsGate", func() { r.HoldsGate = base.HoldsGate })
		inherit("sideways", func() { r.Sideways = base.Sideways })
		inherit("office", func() { r.Office = base.Office })
		inherit("tools", func() { r.Tools = base.Tools })
		inherit("model", func() { r.Model = base.Model })
		inherit("layer", func() { r.Layer = base.Layer })
		for f, dropped := range map[string]bool{"authors": base.Authors && !r.Authors, "holdsGate": base.HoldsGate && !r.HoldsGate, "sideways": base.Sideways && !r.Sideways} {
			if dropped {
				return fmt.Errorf("ranks.%s.%s: an ascended rank keeps everything %s held (%s)", name, f, base.Name, r.Origins[f])
			}
		}
		// divineRules and ascendsWhen are the ascended rank's own — never inherited from a base
		// (a base does not carry its overload's rules or its evolution condition). So are dir and
		// prefix: the ascended body lives in its own dir (default: the rank's name), or its minted
		// name would be its base's.
	}
	if err := c.applyAscendKnobs(byName, order); err != nil {
		return err
	}
	// Validations that need every name resolved (ascendsWhen parses and must name its base).
	for _, name := range order {
		r := byName[name]
		if r.AscendsWhen != "" {
			if r.AscendsFrom == "" {
				return fmt.Errorf("ranks.%s.ascendsWhen: an ascended rank's field — %s ascends from nothing", name, name)
			}
			// the condition's shape is checked at decode; the world parses it again before use.
			fields := strings.Fields(r.AscendsWhen)
			if len(fields) < 3 || fields[0] != "ascend" {
				return fmt.Errorf("ranks.%s.ascendsWhen %q: opens with `ascend <rank> when …`", name, r.AscendsWhen)
			}
			if fields[2] != "when" {
				return fmt.Errorf("ranks.%s.ascendsWhen %q: missing `when` after the ascended rank", name, r.AscendsWhen)
			}
			if fields[1] != r.AscendsFrom {
				return fmt.Errorf("ranks.%s.ascendsWhen: ascends %s, not its base %s", name, fields[1], r.AscendsFrom)
			}
		}
		if len(r.DivineRules) > 0 && r.AscendsFrom == "" {
			return fmt.Errorf("ranks.%s.divineRules: an ascended rank's field — %s ascends from nothing", name, name)
		}
	}
	// Defaults for new ranks.
	for _, name := range order {
		r := byName[name]
		if len(r.Tools) == 0 {
			r.Tools = []string{"*"}
		}
		if r.Dir == "" {
			r.Dir = name
		}
		if r.Prefix == "" {
			r.Prefix = name + "-"
		}
	}
	// The law's shape.
	for _, name := range order {
		r := byName[name]
		if r.ReportsTo == "" {
			return fmt.Errorf("ranks.%s.reportsTo: every rank has an escalation parent (Absolute Rule III)", name)
		}
		if r.ReportsTo != RootRank {
			if _, ok := byName[r.ReportsTo]; !ok {
				return fmt.Errorf("ranks.%s.reportsTo: no rank %q (%s)", name, r.ReportsTo, c.rankWhere(name, "reportsTo"))
			}
		}
	}
	for _, name := range order {
		seen := map[string]bool{name: true}
		gateAbove := false
		for cur := byName[name]; cur.ReportsTo != RootRank; {
			next := byName[cur.ReportsTo]
			if seen[next.Name] {
				return fmt.Errorf("ranks: a cycle through %s — the hierarchy is a tree rooted at %s", next.Name, RootRank)
			}
			seen[next.Name] = true
			if next.HoldsGate {
				gateAbove = true
			}
			cur = next
		}
		r := byName[name]
		if r.Authors && !gateAbove && !r.HoldsGate {
			return fmt.Errorf("ranks.%s authors changes but no rank above it holds a gate (Law 3: nothing lands without the gate)", name)
		}
	}
	c.ranks = make([]Rank, 0, len(order))
	for _, name := range order {
		c.ranks = append(c.ranks, *byName[name])
	}
	return nil
}

var rankFields = []string{"reportsTo", "job", "authors", "holdsGate", "office", "tools", "model", "dir", "prefix", "ascendsFrom", "sideways", "layer", "divineRules", "ascendsWhen", "family", "portrait"}

func (c *Config) rankWhere(name, field string) string { return c.Where("ranks." + name + "." + field) }

// Ranks returns the resolved hierarchy in law order then config order; Rimuru is implicit.
func (c *Config) Ranks() []Rank { return append([]Rank(nil), c.ranks...) }

// Rank finds one rank by any spelling of its name (rank.Canon).
func (c *Config) Rank(name string) (Rank, bool) {
	name = canonicalRank(name)
	for _, r := range c.ranks {
		if r.Name == name {
			return r, true
		}
	}
	return Rank{}, false
}

// Deviations lists every rank field whose value did not come from the law's table, and every
// rank the law does not have; empty when the world runs on the law as written.
func (c *Config) Deviations() []string {
	var out []string
	if c.RankSet == "replace" {
		out = append(out, "rankSet: replace — the law's table is not loaded ("+c.Where("rankSet")+")")
	}
	for _, r := range c.ranks {
		if !r.Builtin {
			out = append(out, fmt.Sprintf("rank %s: not in the law (%s)", c.Dist.Word(r.Name), c.Where("ranks."+r.Name)))
			continue
		}
		fields := sortedKeys(r.Origins)
		for _, f := range fields {
			o := r.Origins[f]
			// a derived family and a family's own law are the law's; a race's own value is not
			if o != "law" && !strings.HasPrefix(o, "ascends:") && !strings.HasPrefix(o, "derived:") && !(strings.HasPrefix(o, "family:") && !strings.Contains(o, "(")) {
				out = append(out, fmt.Sprintf("rank %s.%s: %s", c.Dist.Word(r.Name), f, o))
			}
		}
	}
	return out
}

// RankTree renders the hierarchy as an indented tree under rimuru, in the distribution's words.
func (c *Config) RankTree() string {
	word := c.Dist.Word
	children := map[string][]string{}
	for _, r := range c.ranks {
		children[r.ReportsTo] = append(children[r.ReportsTo], r.Name)
	}
	for k := range children {
		sort.Strings(children[k])
	}
	var b strings.Builder
	var walk func(name, pad string)
	walk = func(name, pad string) {
		for _, ch := range children[name] {
			r, _ := c.Rank(ch)
			var marks []string
			if r.Authors {
				marks = append(marks, "authors")
			}
			if r.HoldsGate {
				marks = append(marks, "gate")
			}
			if r.Sideways {
				marks = append(marks, "sideways")
			}
			if r.Layer != 0 {
				marks = append(marks, fmt.Sprintf("layer %d", r.Layer))
			}
			if r.AscendsFrom != "" {
				marks = append(marks, "ascends "+word(r.AscendsFrom))
			}
			if len(r.DivineRules) > 0 {
				marks = append(marks, fmt.Sprintf("%d divine rules", len(r.DivineRules)))
			}
			if r.AscendsWhen != "" {
				// the condition only: "ascend <base>" is canonical words, and the base is named above
				cond := r.AscendsWhen
				if _, after, ok := strings.Cut(cond, " when "); ok {
					cond = after
				}
				marks = append(marks, "when `"+strings.TrimSpace(cond)+"`")
			}
			if r.Office != "" {
				marks = append(marks, "office "+word(r.Office))
			}
			if len(marks) == 0 {
				fmt.Fprintf(&b, "%s%s\n", pad, word(ch))
			} else {
				fmt.Fprintf(&b, "%s%s (%s)\n", pad, word(ch), strings.Join(marks, ", "))
			}
			walk(ch, pad+"  ")
		}
	}
	fmt.Fprintf(&b, "%s [rankSet: %s]\n", word(RootRank), c.RankSet)
	walk(RootRank, "  ")
	return b.String()
}
