package world

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Kaginari/isekai/loop"
)

// Rule is one injected rule text (config supplies them). Scope is `all`, `rank:<race>` or
// `creature:<name>`.
type Rule struct {
	Text  string
	Scope string
}

// Applies reports whether the rule reaches a body.
func (r Rule) Applies(as, race string) bool {
	s := strings.ToLower(strings.TrimSpace(r.Scope))
	switch {
	case s == "" || s == "all":
		return true
	case strings.HasPrefix(s, "rank:"):
		return strings.TrimPrefix(s, "rank:") == race || (race == "" && strings.TrimPrefix(s, "rank:") == "rimuru")
	case strings.HasPrefix(s, "creature:"):
		return strings.TrimPrefix(s, "creature:") == strings.ToLower(as)
	}
	return false
}

// OntologyOptions is the projection's switch and budget (tokens).
type OntologyOptions struct {
	Enabled bool
	Budget  int
}

// PromptOptions is the system-prompt builder's switchboard.
type PromptOptions struct {
	Crest        bool // law.crest: the nine breaths, first and always
	Identity     bool // who the body is, the classes, the gate, the wire (the engine's default text)
	Rules        []Rule
	Ontology     OntologyOptions
	Instructions bool // project instruction files (AGENTS.md / CLAUDE.md)
	Creature     bool // the body's own card: territory, parent, minds, verify
	Memory       bool // one line on how the tiers are reached (the manifest itself rides the Recall hook)
	SelfState    bool // the loop's own instruments named: the pulse tool, guts notes, the check-loop law
}

// DefaultPrompt is everything on.
func DefaultPrompt() PromptOptions {
	return PromptOptions{Crest: true, Identity: true, Ontology: OntologyOptions{Enabled: true, Budget: 800}, Instructions: true, Creature: true, Memory: true, SelfState: true}
}

// System is the loop's System hook: crest · identity · injected rules · the body's card ·
// ontology projection · instructions · memory line. The recall manifest and the toolbox brief
// are appended by the loop from the Recall hook, so a provider call sees one prompt.
func (w *World) System(opt PromptOptions) func(s *loop.Session) string {
	return func(s *loop.Session) string {
		as := s.Engine.As
		if as == "" {
			as = s.Engine.Lexicon.Body
		}
		return w.Prompt(as, opt, s.Engine)
	}
}

// Prompt builds the system prompt for a body outside a session (a dispatcher's preview, the
// drain's rebuilt context). engine may be nil.
//
// w.mu is held over the whole build, not just the lookup below: the roster
// loop at 129 reads w.Creatures, raceLawPath → RankOf → Creature's locked
// tail reads w.Creatures, and promptOntoSection projects w.Onto — all three
// race Creature's mutation of the same two fields when a court-rendered
// foreign body lands mid-fanout (the app-package shutdown-concurrency race,
// 2026-10-03). One existing mutex, widened — no new fine-grained lock and no
// lock-order policy invented (adding onto's own mutex would order the two
// locks; everything Prompt reaches under it — rankLawText, creaturesOf's
// memo/Rd readers, engine hooks — takes no world lock of its own).
func (w *World) Prompt(as string, opt PromptOptions, engine *loop.Engine) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	var b strings.Builder
	race := w.rankOf(as).Name
	if race == Rimuru {
		race = ""
	}
	if opt.Crest {
		if w.Law != nil && w.Law.Crest != "" {
			fmt.Fprintf(&b, "%s — %s\n\n%s\n\n", w.Law.Path, or(w.Lex.Crest, "the crest"), w.Law.Crest)
		} else {
			fmt.Fprintf(&b, "@? no crest: %s/%s is missing or has no crest section\n\n", w.Lex.WorldDir, w.Lex.Law)
		}
	}
	// the race's own page of law (Veldora 2026-10-03: global natures stay the crest's; the
	// per-race page is thin on purpose — it never restates the crest, names only what the race
	// answers for; every other arm of the code stays consult-on-demand).
	if rp := raceLawPath(w, as); rp != "" {
		if data, err := os.ReadFile(rp); err == nil && strings.TrimSpace(string(data)) != "" {
			fmt.Fprintf(&b, "Law of your race (%s):\n%s\n\n", w.Rel(rp), strings.TrimSpace(string(data)))
		}
	}
	if opt.Identity {
		if engine != nil {
			b.WriteString(strings.TrimSpace(engine.DefaultSystem(nil)) + "\n\n")
		} else {
			fmt.Fprintf(&b, "You are %s, a body working inside the world rooted at %s.\n\n", as, w.Root)
		}
	}
	if len(opt.Rules) > 0 {
		var lines []string
		for _, r := range opt.Rules {
			if r.Applies(as, race) && strings.TrimSpace(r.Text) != "" {
				lines = append(lines, "- "+oneLine(r.Text))
			}
		}
		if len(lines) > 0 {
			b.WriteString("Rules in force for you:\n" + strings.Join(lines, "\n") + "\n\n")
		}
	}
	if opt.Creature {
		if c := w.creature(as); c != nil {
			r, _ := w.Ranks.Get(c.Rank)
			fmt.Fprintf(&b, "You are %s (%s%s). Doc: %s. Territory: %s.", c.Name, c.Rank, jobOf(r), or(c.Doc, "none"), or(strings.Join(c.Territory, ", "), "none declared"))
			if c.Parent != "" {
				fmt.Fprintf(&b, " One hop up: %s.", c.Parent)
			}
			if len(c.Minds) > 0 {
				fmt.Fprintf(&b, " Minds worn: %s.", strings.Join(c.Minds, ", "))
			}
			if len(c.Verify) > 0 {
				fmt.Fprintf(&b, " Verify: %s.", strings.Join(c.Verify, " · "))
			}
			b.WriteString(" A write outside your territory is refused; a change under it changes its doc in the same turn (Vitality).\n\n")
			// neighbourKnows: a sideways rank's layer-mates are named to it, coded — the
			// harness reads the bonds, never the body asking. Divine rules of the ascension
			// rank above a base reach it as absolute rules (coded, not a prompt rule).
			if nb := w.Ranks.NeighborRanks(c.Rank); len(nb) > 0 {
				var mates []string
				for _, other := range w.Creatures {
					for _, nr := range nb {
						if other.Rank == nr && other.Name != c.Name {
							mates = append(mates, other.Name)
							break
						}
					}
				}
				if len(mates) > 0 {
					sort.Strings(mates)
					fmt.Fprintf(&b, "Neighbors in your layer (%s): %s — work may reach them sideways, but only through dispatch.\n\n", strings.Join(nb, ", "), strings.Join(mates, ", "))
				}
			}
			if dr := w.Ranks.DivineRulesFor(c.Rank); len(dr) > 0 {
				b.WriteString("Absolute rules from the ascended above you (divine, never overridden):\n")
				for _, d := range dr {
					b.WriteString("- " + oneLine(d) + "\n")
				}
				b.WriteString("\n")
			}
		}
	}
	if opt.SelfState {
		b.WriteString("Self-state (the loop's own instruments, yours to read):\n" +
			"- `pulse` (tool) returns your live run state: steps, context, momentum, and every bench you have hit with its red/green/muted counts. Read it before re-running a bench you have already struck.\n" +
			"- A `[guts: …]` note inside a tool result means the run is losing ground; a `[doomLoop: …]` note means it checkpointed. Both are readings, not verdicts.\n" +
			"- Repeated reds on one bench with no green is a check-loop, not TDD. If the counts say thrash, stop re-running: change the approach, instrument the probe once, or answer on the wire. Never pipe a bench's output through a filter that hides its exit status.\n\n")
	}
	if opt.Ontology.Enabled && w.Onto != nil {
		budget := opt.Ontology.Budget
		if budget == 0 {
			budget = 800
		}
		name := as
		if w.creature(as) == nil {
			name = "rimuru"
		}
		if lines, err := w.Onto.Project(name, budget); err == nil && len(lines) > 0 {
			b.WriteString("@ONTO\n" + strings.Join(lines, "\n") + "\n\n")
		}
	}
	if opt.Instructions {
		for _, in := range w.Instructions {
			fmt.Fprintf(&b, "Instructions (%s, %s):\n%s\n\n", in.Scope, w.Rel(in.Path), strings.TrimSpace(in.Text))
		}
	}
	if opt.Memory {
		fmt.Fprintf(&b, "Memory: @RECALL below carries anchors (path#section) into the tiers and @TOOLS the level-1 manifest; load a section or a Mind only on your own decision, never pre-emptively. What you learned that nothing on disk says goes out as @U <%s|%s|%s>; the law's code is read with the `law` tool by heading.\n",
			w.Lex.Token("law"), w.Lex.Token("colony"), w.Lex.Token("territory"))
	}
	return strings.TrimSpace(b.String())
}

// raceLawPath is <worldDir>/law/race/<rank>.md — "" when the world hasn't authored this race's
// page. No leak law: a body's page rides its own creature card or the throne — a name the
// roster doesn't know reads NO race page (RankOf's rimuru fallback is a default for rules
// filtering, never a warrant to read the throne's law).
// Callers must hold w.mu — Prompt locks over the whole build and this fans into the
// unlocked Creature/RankOf bodies through creature / rankOf.
func raceLawPath(w *World, as string) string {
	if as != Rimuru && w.creature(as) == nil {
		return ""
	}
	rank := w.rankOf(as).Name
	if rank == "" {
		return ""
	}
	p := filepath.Join(w.Root, w.Lex.WorldDir, "law", "race", rank+".md")
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return p
	}
	return ""
}

func jobOf(r Rank) string {
	var tags []string
	if r.Authors {
		tags = append(tags, "authors")
	}
	if r.HoldsGate {
		tags = append(tags, "holds the gate")
	}
	if len(tags) == 0 {
		return ""
	}
	return ": " + strings.Join(tags, ", ")
}
