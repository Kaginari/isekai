# The canon

The specs the binary is built to. [`ISEKAI.md`](../../ISEKAI.md) is the law every creature reads;
these files are what it points to when a day asks a specific question — how the loop runs, what
config holds, what the terminal draws. Where a spec and the code disagree, the code is wrong or
the spec is stale, and either is a Vitality failure to fix in the same change.

Some files carry the convention's longer history — the numbered rule set, the ascended races,
the founding phases — as source material for the law, not as operative specs; they say so at
their top.

## What's here

- [`rules-chapters.md`](rules-chapters.md) — the E1–E6 protocols and Chapters I–VII (rules 1–21)
- [`ascended-races.md`](ascended-races.md) — fuller prose on high orc and dark elf
- [`rimuru-throne-body.md`](rimuru-throne-body.md) — the OpenCode-side global "throne body" agent spec
- [`isekai-phases.md`](isekai-phases.md) — the `/isekai` command's real Phase 2a (audit-only)
  and Phase 2b (populate) logic — richer than what's installed
- [`templates.md`](templates.md) — the `AGENTS.md` colony-block template Phase 2b appends
- [`instruments-tempest.md`](instruments-tempest.md) — the full `tempest.js` instrument-board
  spec — in the binary: `isekai status`, the board
- [`memory-tiers.md`](memory-tiers.md) — the three memories every creature has (short · long ·
  shared), the files-are-truth boundary, relation-ranked recall, and the path to a DB tier —
  in the binary: `memory/`, `isekai memory`
- [`agent-loop.md`](agent-loop.md) — the six-beat loop every Court Body runs (perceive → recall →
  plan → act → verify → record): budgets as readings, four step classes with the human gate,
  journaled resume, one-hop escalation — in the binary: `loop/`
- [`toolbox.md`](toolbox.md) — the two-level toolbox: a priced registry of Minds, commands, tools,
  Bodies and externals; level 1 injects only the manifest that fits the turn and the budget, level 2
  loads a body on the creature's own decision — in the binary: `toolbox/`, `isekai toolbox`
- [`binary.md`](binary.md) — the `isekai` Go binary: the law as a harness — native ranks and
  Court dispatch, the gate and Vitality in code, the human gate, the wire, and the ontology that
  passes knowledge along creature relations — the source is this repository
- [`harness-parity.md`](harness-parity.md) — what Claude Code and OpenCode ship, what the binary
  ships in v0, what it defers, what it refuses, and how each capability meets the law
- [`config.md`](config.md) — the binary's config: locations, precedence, the full schema with every
  feature switchable, per-distribution defaults, and the rule that a switched-off law is a finding
- [`tui.md`](tui.md) — the terminal UI, Claude Code style: inline scrollback, streamed markdown,
  tool blocks with diffs, inline approvals, the slash menu, the spinner and footer
- [`ui.md`](ui.md) — how the world draws its pages: the Bootstrap 5 grid, mobile first,
  components before custom markup, an isekai theme over Bootstrap's tokens, SVG sized by column
- [`runtime.md`](runtime.md) — the v0.2 design: the container runtime (permissions follow what is at
  risk), enterprise registries, secrets in two steps, the corporate proxy, worlds and dimensions

## Gaps in the source material

The history files were transcribed from a fuller text than survived; what is missing:

- The full text of rules 12–13 (Chapter V — Tracking) — only the tail of 13 survived.
- The Kijin paragraph in the ascended-races section (High orc and Dark elf are complete; Kijin
  is not).
- Rules 1–3 of Raphael's own office laws (only 4–6 and its Wire mouth section survived).
- Great Sage's own Wire mouth commissions (only its office laws survived).
- The `.opencode/skills/` seed template that Phase 2b step 3 refers to ("seed template
  below") was never itself photographed — only the reference to it.

Don't fill these gaps by inference — if you find the rest of this source, transcribe it
directly; otherwise leave the gap marked.
