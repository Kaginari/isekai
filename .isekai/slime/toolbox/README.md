# slime-toolbox

- **Rank:** Slime
- **Territory:** `toolbox/`
- **Reports to:** orc-world
- **Minds:** slime-go-proving, slime-wire-and-journals
- **Purpose:** the ground truth of the two-level toolbox: a priced registry of Minds, commands, tools, Bodies and externals; level 1 the @T manifest that fits the ask and a budget; level 2 the journaled load — the Go port of toolbox.js.

## Traits
- RegV 1, BudgetDefault 1500 tokens, TOK(s) = ceil(len/4); Kinds are mind, command, tool, body, external; laneOfRace maps a mind's name prefix to its lane (slime → zone, orc → verdict, elf and kijin → global) — a skill with no rank prefix is listed as a shared host tool.
- Frontmatter() reads the YAML-ish front matter (name, description, triggers as a list or string); triggers are weighted frontmatter .8, quoted .8, name .5, sentence .4 (triggerW) when picking for an ask; FitWords 2.
- Files: toolbox/registry.json (derived, rebuildable, gitignored by init), toolbox/extra.jsonl (hand-kept externals), the loads journal under instruments/toolbox. LoadRegistry reports Live/stale; IsStale triggers a rebuild in the world's Recall hook.
- Brief(q, PickOpts{As, Budget}) returns level-1 `@T` lines under the budget with holes ("registry older", "no registry", "registry empty" are the world's state, not faults); Load(name, LoadOpts) returns a body whole or by section anchor and journals it; Status reports loads against offers; Explain shows one entry's pricing.
- CLI: index, brief, pick, load, status, explain — same words and output as toolbox.js; selftest runs the CLI against a fixture.

## Verify
- `go test ./toolbox/...`

## Thoughts
