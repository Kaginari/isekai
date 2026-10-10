# slime-onto

- **Rank:** Slime
- **Territory:** `onto/`
- **Reports to:** orc-world
- **Minds:** slime-go-proving, slime-creature-docs
- **Purpose:** the ground truth of the ontology: the Turtle-subset schema and parser, the triple store, the forward-chaining reasoner, the shapes `onto check` enforces, the derivation of creatures and minds from files, the flow of facts along bonds, and the @ONTO projection.

## Traits
- One prefix, `is:` = isekai:. schema.ttl declares classes (Creature, Rimuru, Elf, Orc, Slime, Kijin, Mind, Doc, Fact with Law/Colony/Territory), the bonds (above transitive; truth Slime→Orc, verdict Orc→Elf, reports Creature→Rimuru all subPropertyOf above; wears/wornBy, owns, knows/knownBy, doc, shared) and the shapes. The world's copy at .isekai/ontology/schema.ttl is written by init from onto.DefaultSchema.
- Shapes (validate.go): SlimeTruth — a slime has exactly one truth edge; SlimeTerritory — slimes' owns literals are disjoint (overlap = equal, or prefix with a `/`); MindWorn — a mind has ≥ 1 wornBy unless `shared`; BondTarget — every above-target has a doc (a parent named but absent from the roster is a finding at the child's parent line).
- derive() never writes: creature id = <rank dir>-<folder lowercased>; the doc is README.md, else <folder>.md, else SKILL.md, else the first .md; fields are read leniently from `- **Key:** value` lines (keys reports to / orc / elf / parent; territory / zone / owns; minds / mind / wears / hats), first occurrence wins; the parent falls back to the first name of the expected rank in the doc, or rimuru.
- Territory is the Territory field PLUS every back-ticked token containing a slash and no whitespace anywhere in the doc — a foreign path in backticks becomes an owns edge and an overlap finding. Paths are normalised (./ and trailing /, *, ** stripped, lower-cased).
- Minds come from .opencode/skills, .opencode/skill and .claude/skills (mindDirs); a skill whose name prefix is not a rank dir is `shared` — the host's tool, nobody's mind. A creature wears a mind when the Minds field names it or the mind's name appears as a whole word anywhere in its doc.
- Rimuru's doc is the law file; Project(creature, budget) renders the @ONTO block by token budget (Tokens = runes/4, rounded up); Visible walks the bonds to say what a viewer may see.
- CLI: `onto check` prints `@S PASS n findings · creatures · minds · facts · triples` and exits non-zero on findings; also project and query. Selftest builds a fixture world, sound and broken.

## Verify
- `go test ./onto/...`

## Thoughts
