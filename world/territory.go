package world

import (
	"fmt"
	"strings"

	"github.com/Kaginari/isekai/tool"
)

// TerritoryOptions is law.territory: an authoring body's write outside its territory is
// refused, not warned (Law 2). Which ranks are held to it is the table's Authors column; Ranks
// narrows it to a list when set.
type TerritoryOptions struct {
	Enabled            bool
	Ranks              []string
	ThroneNeverAuthors bool
}

func (o TerritoryOptions) held(w *World, c *Creature) bool {
	if len(o.Ranks) > 0 {
		return contains(o.Ranks, c.Rank)
	}
	r, ok := w.Ranks.Get(c.Rank)
	return ok && r.Authors
}

// Policy is the tool.Policy for a body: every path a write/edit touches must be inside its
// territory. A refusal carries the escalation hint — one hop up, never sideways. The throne is
// bound too under the hard law (Veldora decree 2026-09-30): rimuru writing creature soil is
// refused with a dispatch-redirection — dispatching through the layer chain is the coded
// default; only a creature's own README (co-touches pass so pairing stays possible) and the
// throne's own seat files are exempt.
func (w *World) Policy(body string, opt TerritoryOptions) tool.Policy {
	if !opt.Enabled {
		return nil
	}
	return func(a tool.Access) error {
		if a.Class < tool.Write || len(a.Paths) == 0 {
			return nil
		}
		// engine soil: journals, reports, ledgers under instruments/loop and instruments/usage
		// land from the harness only. Every body's write there is refused — Law 2 grows a
		// sub-case: a body's report speaks through the wire, the engine files it, nobody edits.
		eng1 := normRel(w.Lex.WorldDir + "/instruments/loop")
		eng2 := normRel(w.Lex.WorldDir + "/instruments/usage")
		for _, p := range a.Paths {
			rel := normRel(w.Rel(p))
			if under(rel, eng1) || under(rel, eng2) {
				return fmt.Errorf("territory: %s is engine soil (journals and reports land there from the harness only) — a body never writes under %s; speak on the wire and the engine files it", rel, eng1)
			}
		}
		// the throne, bound (pure hard law): a rimuru write inside another creature's soil is
		// refused with the chain named — dispatch + the pairing gate are the lawful shapes.
		if body == "" || body == Rimuru || w.RankOf(body).Name == Rimuru {
			var seat1, seat2 = normRel(w.Lex.WorldDir + "/dimension.yaml"), normRel(w.Lex.WorldDir + "/" + w.Lex.Law)
			for _, p := range a.Paths {
				rel := normRel(w.Rel(p))
				if rel == seat1 || rel == seat2 {
					continue // the throne's own seat files
				}
				if opt.ThroneNeverAuthors && rel != seat2 {
					first := w.Ranks.firstChildRank(Rimuru)
					childHint := "elf"
					if first != "" {
						childHint = first
					}
					return fmt.Errorf("territory: %s — Law 7 + Veldora decree 2026-10-01: the throne authors nothing but the law (.isekai/isekai.md); dispatch the chain (elf -> first rung: `%s`)", rel, childHint)
				}
				owner := w.Owner(rel)
				if owner == nil {
					continue // no-creature soil stays lawful for the throne
				}
				if owner.Doc != "" && rel == normRel(owner.Doc) {
					continue // touching the owner's doc itself is the pairing act; allowed
				}
				first := w.Ranks.firstChildRank(Rimuru)
				childHint := "elf"
				if first != "" {
					childHint = first
				}
				return fmt.Errorf("territory: %s belongs to %s — Law 2 plus the hard-law decree: rimuru does not author another creature's soil; dispatch the chain (elf -> orc -> slime; first rung: `%s`), or patch the config's law.territory to loosen this",
					rel, owner.Name, childHint)
			}
			return nil
		}
		c := w.Creature(body)
		if c == nil || !opt.held(w, c) || (len(c.Territory) == 0 && c.Dir == "") {
			return nil
		}
		for _, p := range a.Paths {
			rel := w.Rel(p)
			if c.InTerritory(rel) {
				continue
			}
			owner := w.Owner(rel)
			if owner != nil && owner.Name != c.Name && owner.Rank == c.Rank {
				// sideways soil of the same rank IS licensed — dispatch the owner directly.
				myRank, _ := w.Ranks.Get(c.Rank)
				if myRank.Sideways {
					return fmt.Errorf("territory: %s is outside %s's territory — it belongs to %s, your rank-mate. Dispatch %s directly (sideways licensed across %s); escalation one hop up to %s only if the sideways rung fails",
						rel, c.Name, owner.Name, owner.Name, owner.Rank, or(c.Parent, "the dispatcher"))
				}
			}
			up := c.Parent
			if up == "" {
				up = "the dispatcher"
			}
			// the consult card (2026-10-09 human law): a body on foreign soil does not
			// re-derive who owns what — the refusal itself names the owner and the three
			// arms (widen / birth / route), so the parent decides without a drain
			ownerLine := " — the soil is unrouted (a Genesis trigger when it recurs)"
			if owner != nil {
				ownerLine = fmt.Sprintf(" — it belongs to %s (%s, domain %s)", owner.Name, owner.Rank, or(owner.Parent, "rimuru"))
			}
			return fmt.Errorf("territory: %s is outside %s's territory%s — Law 2: do not write there; consult one hop up to %s with `@? %s`. Arms for %s: widen (grant %s the claim on %s — its README Territory grows, the doc pairs in-turn), birth (mint a creature for that zone), route (dispatch the owner's line)",
				rel, c.Name, ownerLine, up, oneLine("needs "+rel+" which "+c.Name+" does not own"), up, c.Name, rel)
		}
		return nil
	}
}

// DispatchCard is the dispatcher's one-read answer for a path: the owner, its rank,
// its domain, its strain reading, and the consult arms (widen / birth / route) — so an
// orc or elf grounds its do-it / birth / route decision on live state instead of
// re-deriving who owns what (the 2026-10-09 audit: re-derivation chains burned a day).
// The `onto` tool serves it on path queries.
func (w *World) DispatchCard(rel string) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	rel = normRel(rel)
	c := w.Owner(rel)
	if c == nil {
		return rel + " is unrouted — no creature owns it (a Genesis trigger when it recurs). Arms: widen an adjacent creature's claim, birth a creature for this zone, or leave it open."
	}
	strain := ""
	if s := w.stress[c.Name]; s > 0 {
		strain = fmt.Sprintf(", strain %d", s)
	}
	dom := or(c.Parent, "rimuru")
	return fmt.Sprintf("%s is owned by %s (%s%s) — doc %s; domain %s; territory %s. Arms for the dispatcher: widen (grant another creature the claim — its README Territory grows, Vitality pairs in-turn), birth (mint a creature for a slice of this zone), or route (dispatch %s through %s).",
		rel, c.Name, c.Rank, strain, or(c.Doc, "none"), dom, strings.Join(c.Territory, ", "), c.Name, dom)
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}
