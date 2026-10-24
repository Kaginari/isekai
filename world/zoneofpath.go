package world

import (
	"path"
	"strings"
)

// Soil is what kind of ground a root-relative path stands on — the one reading every
// ownership question asks first (the gate, the territory policy, the dispatch card).
type Soil int

const (
	// SoilOpen is ground no living creature claims.
	SoilOpen Soil = iota
	// SoilTerritory is a creature's claimed soil (its Territory prefixes).
	SoilTerritory
	// SoilSeat is a creature's own seat: its directory under <worldDir>/<rank>/, where its doc
	// lives. The seat is the creature's, whatever another creature's broad claim covers.
	SoilSeat
	// SoilHarness is the harness's own bookkeeping — the shared shelf (queue, ledgers, value),
	// the instruments shelf (journals, reports, usage) and the attic (old docs, births). It lands from the harness only and
	// never bills a body: no creature owns it, so no doc is owed for it.
	SoilHarness
	// SoilScratch is the world's scratch (<worldDir>/tmp/): a body's own nest — its redirected
	// /tmp, its plan deck. Lawful for every body, owned by none, never a Vitality debt.
	SoilScratch
)

func (s Soil) String() string {
	return [...]string{"open ground", "territory", "seat", "harness soil", "scratch"}[s]
}

// Zone is ZoneOfPath's answer: the soil and, on a seat or territory, the creature it is.
type Zone struct {
	Soil     Soil
	Creature *Creature
}

// harnessShelves and scratchShelf are the world-dir children ZoneOfPath reads as harness soil
// and scratch. The attic keeps what the harness set aside: old docs, the births ledger.
var harnessShelves = []string{sharedShelf, "instruments", "attic"}

const scratchShelf = "tmp"

// ZoneOfPath is the one ownership truth for a root-relative path, in order: the harness's
// shelves and the scratch (nobody's), a creature's seat (the seat's, even under a broader
// claim — a creature's README is never another body's pairing debt), then the longest
// territory claim of an authoring creature, then of a gate holder. Owner is its creature.
func (w *World) ZoneOfPath(rel string) Zone { return w.zoneAmong(w.Creatures, rel) }

// zoneAmong is ZoneOfPath over a given roster (the trusted one: world/claims.go).
func (w *World) zoneAmong(cs []Creature, rel string) Zone {
	rel = normRel(rel)
	wd := normRel(w.Lex.WorldDir)
	for _, s := range harnessShelves {
		if under(rel, wd+"/"+s) {
			return Zone{Soil: SoilHarness}
		}
	}
	if under(rel, wd+"/"+scratchShelf) {
		return Zone{Soil: SoilScratch}
	}
	var seat *Creature
	for i := range cs {
		c := &cs[i]
		// a foreign body (another harness's agent file) has a doc and no dir: its doc is its seat
		in := (c.Dir != "" && under(rel, c.Dir)) || (c.Doc != "" && normRel(c.Doc) == rel)
		if in && (seat == nil || len(c.Dir) > len(seat.Dir)) {
			seat = c
		}
	}
	if seat != nil {
		return Zone{Soil: SoilSeat, Creature: seat}
	}
	for _, pick := range []func(Rank) bool{func(r Rank) bool { return r.Authors }, func(r Rank) bool { return r.HoldsGate && !r.Authors }} {
		var best *Creature
		bestLen := -1
		for i := range cs {
			c := &cs[i]
			if r, ok := w.Ranks.Get(c.Rank); !ok || !pick(r) {
				continue
			}
			for _, t := range c.Territory {
				if under(rel, t) && len(t) > bestLen {
					best, bestLen = c, len(t)
				}
			}
		}
		if best != nil {
			return Zone{Soil: SoilTerritory, Creature: best}
		}
	}
	return Zone{Soil: SoilOpen}
}

// plansShelf is where the plan tool pens each body's deck (<worldDir>/tmp/plans/<body>.json).
const plansShelf = scratchShelf + "/plans"

// foreignDeck names the body whose plan deck rel is, when it is not `as`'s own: the scratch is
// every body's nest, but a deck speaks for its body — "" when rel is no deck or `as`'s own.
func (w *World) foreignDeck(as, rel string) string {
	dir := normRel(w.Lex.WorldDir + "/" + plansShelf)
	rel = normRel(rel)
	if !under(rel, dir) || strings.Contains(strings.TrimPrefix(rel, dir+"/"), "/") {
		return "" // the archive and anything nested below it is the plan tool's own sweep
	}
	deck := strings.TrimSuffix(path.Base(rel), path.Ext(rel))
	if deck == "" || deck == strings.ToLower(or(as, Rimuru)) {
		return ""
	}
	return deck
}
