package world

import (
	"sort"
	"strings"
)

// Claims and trust: a body's README names its territory, and the body edits that README itself
// (Vitality). Read straight off the docs, that would let a body widen its own claim onto a
// rank-mate's soil and land there in the same turn — the longest prefix would make it the owner.
// So the claims a body's writes are judged by are the claims the gate trusts (claimBase), the
// same discipline as the verify lines (selfheal.go):
//
//   - at Open, and at a harness mint (relief, ascension, evolution), every claim on disk is
//     trusted;
//   - a claim a body added since is trusted at once only when it takes nothing from another
//     living body — no owning body's trusted claim or seat sits at, under or over it; open
//     ground is a body's to claim;
//   - a claim that does take another living body's soil is held back until a gate other than
//     the claimant's — its parent's line or the throne, carrying the claimant's README — passes;
//     the claimant's own gate fails it, and a reload never trusts it on its own.

// trustedClaimsLocked is c's claims as the gate trusts them; w.mu held.
func (w *World) trustedClaimsLocked(c *Creature) []string {
	if w.claimBase == nil {
		return c.Territory
	}
	base := w.claimBase[c.Name]
	var out []string
	for _, t := range c.Territory {
		if containsRel(base, t) || w.claimTakesLocked(c.Name, t) == "" {
			out = append(out, t)
		}
	}
	return out
}

// claimTakesLocked names the living body whose trusted soil a claim t of body takes ("" when it
// takes none): an owning body's (one that authors or holds a gate) trusted claim or seat at,
// under or over t. w.mu held.
func (w *World) claimTakesLocked(body, t string) string {
	t = normRel(t)
	for i := range w.Creatures {
		o := &w.Creatures[i]
		if o.Name == body {
			continue
		}
		if r, ok := w.Ranks.Get(o.Rank); !ok || !(r.Authors || r.HoldsGate) {
			continue
		}
		if o.Dir != "" && (under(t, o.Dir) || under(normRel(o.Dir), t)) {
			return o.Name
		}
		claims := o.Territory
		if base, known := w.claimBase[o.Name]; known {
			claims = base
		}
		for _, p := range claims {
			if under(t, p) || under(normRel(p), t) {
				return o.Name
			}
		}
	}
	return ""
}

// heldClaims is the claims of body the gate does not trust yet, each with the living body whose
// soil it takes.
func (w *World) heldClaims(body string) map[string]string {
	w.mu.Lock()
	defer w.mu.Unlock()
	c := w.creatureByName(body)
	if c == nil || w.claimBase == nil {
		return nil
	}
	out := map[string]string{}
	for _, t := range c.Territory {
		if containsRel(w.claimBase[c.Name], t) {
			continue
		}
		if o := w.claimTakesLocked(c.Name, t); o != "" {
			out[t] = o
		}
	}
	return out
}

// trustedRosterLocked is the roster with each body's territory cut to its trusted claims.
func (w *World) trustedRosterLocked() []Creature {
	out := make([]Creature, len(w.Creatures))
	for i := range w.Creatures {
		out[i] = w.Creatures[i]
		out[i].Territory = w.trustedClaimsLocked(&w.Creatures[i])
	}
	return out
}

// trustedZone is ZoneOfPath over the trusted claims, and the body named on it as the trusted
// roster reads it (nil when the body is not on the roster).
func (w *World) trustedZone(rel, body string) (Zone, *Creature) {
	w.mu.Lock()
	defer w.mu.Unlock()
	cs := w.trustedRosterLocked()
	var self *Creature
	for i := range cs {
		if cs[i].Name == body {
			self = &cs[i]
		}
	}
	return w.zoneAmong(cs, rel), self
}

// rebaseClaimsLocked takes the claims the gate trusts after a fresh read: everything when all
// is set (Open, a harness mint), else what the last base held plus what takes no living body's
// soil. w.mu held, the fresh roster on w.Creatures.
func (w *World) rebaseClaimsLocked(all bool) {
	next := map[string][]string{}
	for i := range w.Creatures {
		c := &w.Creatures[i]
		if all || w.claimBase == nil {
			next[c.Name] = append([]string(nil), c.Territory...)
			continue
		}
		next[c.Name] = append([]string(nil), w.trustedClaimsLocked(c)...)
	}
	w.claimBase = next
}

// grantClaims trusts every claim of the bodies whose README a passing gate carried for a writer
// in their line above (the parent, its line, the throne) — never the claimant's own gate.
func (w *World) grantClaims(as string, wrote []string) {
	throne := as == "" || as == Rimuru || w.RankOf(as).Name == Rimuru
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.claimBase == nil {
		return
	}
	writer := strings.ToLower(or(as, Rimuru))
	for i := range w.Creatures {
		c := &w.Creatures[i]
		if c.Doc == "" || c.Name == writer || !containsRel(wrote, c.Doc) {
			continue
		}
		above := throne
		for cur, seen := c, map[string]bool{}; !above && cur != nil && cur.Parent != "" && !seen[cur.Parent]; cur = w.creatureByName(cur.Parent) {
			seen[cur.Parent] = true
			above = cur.Parent == writer
		}
		if above {
			w.claimBase[c.Name] = append([]string(nil), c.Territory...)
		}
	}
}

func containsRel(xs []string, x string) bool {
	x = normRel(x)
	for _, y := range xs {
		if normRel(y) == x {
			return true
		}
	}
	return false
}

// heldAdmits reports whether rel is admitted to body only by a claim the gate holds back.
func (w *World) heldAdmits(rel, body string) bool {
	_, tc := w.trustedZone(rel, body)
	return tc != nil && !tc.InTerritory(rel)
}

// heldList renders held claims for a refusal: "src/x/ (slime-y's soil), ...".
func heldList(held map[string]string) string {
	ts := make([]string, 0, len(held))
	for t := range held {
		ts = append(ts, t)
	}
	sort.Strings(ts)
	for i, t := range ts {
		ts[i] = t + " (" + held[t] + "'s soil)"
	}
	return strings.Join(ts, ", ")
}

func creatureName(c *Creature) string {
	if c == nil {
		return ""
	}
	return c.Name
}
