# slime-toolbox

- **Rank:** Slime
- **Territory:** `toolbox/`
- **Reports to:** orc-world
- **Minds:** slime-go-proving, slime-wire-and-journals
- **Purpose:** the ground truth of the two-level toolbox: a priced registry of Minds, commands, tools, Bodies and externals; level 1 the @T manifest that fits the ask and a budget; level 2 the journaled load — the Go port of toolbox.js.

## Traits
- RegV 1, BudgetDefault 1500 tokens, TOK(s) = ceil(len/4); Kinds are mind, command, tool, body, external; laneOfRace maps a mind's name prefix to its lane (slime → zone, orc → verdict, elf and kijin → global) — a skill with no rank prefix is listed as a shared host tool.
- Sources are memory's lists, never a second copy: Minds from memory.MindBases(dir) (discover.SkillBases: the world's crafted-skills and base-skills, then .claude/skills, .opencode/skills, .opencode/skill, .agents/skills), commands from memory.CommandBases, Bodies from .claude/agents and .opencode/agent(s), externals from extra.jsonl plus config's registry.tools (World.Extra).
- Frontmatter() reads the YAML-ish front matter (name, description, triggers as a list or string); triggers are weighted frontmatter .8, quoted .8, name .5, sentence .4 (triggerW) when picking for an ask; FitWords 2.
- Files under the world dir: toolbox/registry.json (derived, rebuildable), toolbox/extra.jsonl (hand-kept externals), instruments/toolbox/loads.jsonl (the loads journal). LoadRegistry reports Live or stale; IsStale triggers a rebuild in the world's Recall hook.
- Brief(q, PickOpts{As, Budget}) returns level-1 `@T` lines under the budget with holes ("registry older", "no registry", "registry empty" name the world's state, not faults — the empty-registry hole names the world dir); Load(name, LoadOpts) returns a body whole or by section anchor and journals it; Status reports loads against offers; Explain shows one entry's pricing.
- CLI: index, brief (alias pick), load, status, explain, selftest — the same words and output as toolbox.js; the interop test skips when node or the JS tool is absent.

## Verify
- `go test -count=1 -run 'TestSelftest|TestFrontmatterAndHeader|TestAPI|TestEmptyRegistryHoleNamesTheWorldDir|TestInteropWithToolboxJS' ./toolbox/`

## Thoughts
- (none yet)
