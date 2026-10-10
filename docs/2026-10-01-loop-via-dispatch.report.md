# Loop census — relayed report for rimuru

Commission: rimuru → elf-isekai → orc-engine → slime-loop (file-brief pattern), 2026-10-01.
Scope: loop. Office: findings (great-sage). Cap 2048 consumed; three dispatch attempts.

Chain artifact trail:
- slime-loop working census (settled Q1-Q3, drifted Q4/Q5): `.isekai/slime/loop/census-working.md`
- orc-engine relay fail-log: `.isekai/instruments/loop/orc-engine-202610014424.report.md`
- elf-isekai closure reads: `loop/loop.go`, `loop/journal.go`, `app/worldtools.go`, `board/sources.go`, `world/territory.go`

## Q1 — the six beats

Named in the spoken protocol, `.isekai/base-skills/loop-core/SKILL.md`
(load=525tok). The code never strings them together; beats 2-5 are real
functions in `loop/loop.go`, beat 6 (compact+usages) fires only via the
302 summoning comment in the same file. Detailed per-beat table:
`.isekai/slime/loop/census-working.md` Q1.

## Q2 — enforcement: both, asymmetric per beat

Code-enforced: budget readings, checkpointing, summaries (loop.go),
drain (@S/@FORGE, summoned at loop.go:302). Protocol-only: the wire
contract (loop.go:594); .isekai/log.md surfaces FAIL example at lines
160/177/185/195. Per-beat detail: census-working.md Q2.

## Q3 — budget readings (settled at elf layer, correcting the orc's partial claim)

Budget struct + the three reading fields: `loop/loop.go:59-90`.
Per-step gate (`step()`): **steps** checked at `loop/loop.go:724`,
**wall clock** (time budget) at `loop/loop.go:728` → both fire
checkpoint+stop per step. **Context window** is checked **per turn**, not
per step, at `loop/loop.go:589`:
`if b := s.Budget; b != nil && b.Context > 0 && s.Context >= b.Context`.

Correction to orc-engine's relay: its "only turn-level → law aspirational
at step level" escape overstated the gap. Slime-loop's census verdict is
the accurate one: two-thirds true at step level (steps + wall), one-third
turn-level (context). The crest wording ("budgets are checked before every
step") is fully true for steps/wall, aspirational for context-within-a-turn.

## Q4 — journal sinks (beats → sinks)

Sinks on disk under `.isekai/instruments/`: desk, plans, gate, loop,
usage, audits, stream, toolbox (orc enumeration; verified against
`app/embed.go:41` seeding list and `app/embed.go:58` .gitignore).

Beat-to-sink, verified at elf layer:
- **run journals** (`instruments/loop/run-*.jsonl`) — `Loop.journalPath`
  `loop/loop.go:174`; writer `loop/journal.go:2`; turn events logged at
  `loop/loop.go:647` (t:turn, usage, context, calls).
- **usage** (`instruments/usage/<session>.jsonl`) — no in-tree writer
  (grep repo-wide finds only test readers + board `board/sources.go:82`);
  the usage lines confirm loop.go:647 writes usage fields into the run
  journal; the separate usage file's producer is **not in this world —
  hole** (likely host harness side, same shape as desk below).
- **desk** (`instruments/desk/<session>.md`) — host tool, not the loop:
  `app/worldtools.go:537` (desk tool; "over the limit is a stress
  reading, said out loud").
- **stream / toolbox / plans / audits / gate** — directories seeded by
  `app/embed.go:41`; writers not re-derived here (cut at census), carried
  from orc enumeration. Stream per-turn chunks, toolbox the @T manifest,
  plans the plan tool, gate the orc decision journal, audits the
  run-audit verdicts.

Journal record dual shape: human-readable (`loop/journal.go:13`) and
JSONL-ish one-event-per-line (`loop/journal.go:14`).

## Q5 — loop.js vs the by-hand protocol

Settled shape, open location. The Go engine **mirrors** loop.js — wire and
data shapes follow the scripted tool: `Loop.Journal` docstring
`loop/loop.go:171-173` ("Journal is the run's journal — the loop.js
script's journal, mirrored"); the dual record shape journal.go:13-14 is
that mirror. **loop.js file not found** under the world root or the
guessed `scripts/` path (probed: absent). Its on-disk home is **the one
open question** — likely the host harness (opencode-side) rather than this
world's Go tree.

Difference: loop.js = the same six beats scripted as an executable harness;
the by-hand protocol = the same beats spoken/followed by a body when no
script drives (loop-core SKILL.md). The Go engine is a third, mirror
implementation that journals loop.js's shapes.

## Unsaid (from both lower layers + elf)

- orc-engine: budget >30% consumed on one greppable question — dispatch +
  door + filing = deductible from remaining alone; multi-question loops
  need per-question FALLBACK pre-wired into the brief.
- orc-engine: the **minimal payload** (brief + greps only, answers cut)
  completed ~2× faster than the full-payload tries — corroborates the
  second-strike note; the fix that works is trimming the *payload*, not
  adding pre-stamps.
- orc-engine: `step()` gated to budget-check is asymmetric to the crest's
  declaration (partially corrected above — only the context reading is
  turn-level).
- Verification of the door: `grep -c '^## Q' census = 5` passed, but one
  2-k `$|\\d+\\.` (working-prefix sentinel) to slime-loop's own README
  failed — orc's exit greps were brittle against the file-brief pattern's
  working-file prefix. Future file-briefs must either name the final file
  in the greps or make greps prefix-tolerant.
- elf-isekai (territory): dispatch-timeouts consumed without slime/README
  doc-pairing on my side because all writes landed in other creatures'
  territories (slime-loop's land, instruments/loop). My relay report is
  the doc.
- elf-isekai (colony): the FAIL stamp on the relay did not mean no
  content — always read the filed report before concluding a dispatch
  yielded nothing.

