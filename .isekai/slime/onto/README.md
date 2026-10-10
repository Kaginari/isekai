# slime-onto

- **Rank:** Slime
- **Territory:** `onto/`
- **Reports to:** orc-world
- **Minds:** slime-go-proving, slime-creature-docs
- **Purpose:** the ground truth of the ontology: the Turtle-subset schema and parser, the triple store, the forward-chaining reasoner, the shapes `onto check` enforces, the derivation of creatures and claims from their docs, the flow of facts along bonds, the @ONTO projection, and housekeep's chain law.

## Traits
- One prefix, `is:` = isekai:. schema.ttl declares the classes (Creature, Rimuru, Elf, Orc, Slime, Kijin, Mind, Doc, Fact with Law/Colony/Territory), the bonds (above transitive; truth, verdict, reports subPropertyOf above; wears/wornBy, owns, knows/knownBy, doc, shared) and four shapes: SlimeTruth (exactly one truth edge), SlimeTerritory (owns disjoint), MindWorn (≥ 1 wornBy unless shared), BondTarget (every above-target has a doc). The world's copy of the schema must equal onto.DefaultSchema (TestSchemaOnDiskMatchesBuiltIn).
- Overlap (validate.go `overlap`): two creatures asserting the same FILE is a lawful split, not a finding; the same DIRECTORY claimed twice collides, and so does a strict ancestor (src/auth vs src/auth/login). Two claims of one creature never collide with each other.
- derive() never writes: creature id = `<rank dir>-<folder lowercased>`; the doc is README.md, else `<folder>.md`, else SKILL.md, else the first .md; fields are read leniently from `- **Key:** value` lines (reports to / orc / elf / parent; territory / zone / owns; minds / mind / wears / hats), first wins, and a `## Territory` heading's backticked bullets count too. The parent falls back to the first creature of the expected rank the doc names, or rimuru.
- Claims: the structured Territory field is the authoritative claim list — a bare entry (a top-level dir, a file name) derives; a slashed entry derives only when rooted in the seat's own name (or `.isekai` / `src`), otherwise it is a cross-reference and mints nothing. Prose backticks derive owns only when they self-name the creature (selfNamesSpan) and are rooted in its seat or `.isekai`; a foreign file path, bare .isekai, anything under .isekai/tmp, URI templates and REST routes never derive.
- Minds come from .opencode/skills, .opencode/skill and .claude/skills (mindDirs); a skill whose name prefix is no rank dir is `shared`. A creature wears a mind its Minds field names or whose name appears as a whole word in its doc.
- Housekeep(root, heal, spawn): check-only reports and writes nothing; heal mints a missing creature home, rewires a missing chain edge, and rewires a dead bond (a `Reports to:` naming no live seat) to the rank's law parent — a live bond, law-shaped or not, stays the operator's. spawn seats draft homes for tracked top-level dirs nobody covers. Overlaps are reported, never rewritten. app runs heal at every world-touching CLI start and at init, so a heal must stay idempotent.
- CLI: `onto check` prints `@S PASS|FAIL n findings · creatures · minds · facts · triples` (non-zero on findings), plus show, project, owner (ownerOf mirrors world.Owner: longest claim, authoring rank on a tie) and assert. Project(creature, budget) renders @ONTO by token budget (Tokens = runes/4, rounded up); Visible walks the bonds.

## Verify
- `go test -count=1 -run 'TestHousekeep|TestDeriveOwns|TestDeriveSkipsDeadSeatPrefixes' ./onto/`
- `go test -count=1 -run 'TestDeriveAndValidate|TestValidateBroken|TestDanglingBond|TestSchemaOnDiskMatchesBuiltIn|TestFlowAndProjection|TestCLI|TestSelftest' ./onto/`

## Thoughts
- (none yet)
