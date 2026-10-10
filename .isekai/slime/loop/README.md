# slime-loop

- **Rank:** Slime
- **Territory:** `loop/`, `instrument/`, `compact/`, `discover/`
- **Reports to:** orc-bench
- **Minds:** loop-core
- **Purpose:** the ground truth of the six-beat run (loop/), the instruments that measure it (instrument/), and the two memory halves (compact/ + discover/).

## Traits
- Session.step runs in this order: step and wall-clock budgets (CHECKPOINT with a resume hint) → a missing tool (Hooks.Missing; the name `_malformed` carries a parse fault from a text-tool-call provider) → t.Settle → Hooks.Guard (status `refused`, by guard) → env.Policy (territory; `refused`, by policy) → Hooks.Decide (deny/allow/ask; allow marks the gate answer By "rule") → gate.Ask → dry-run (`would-ask` / `would-run`) → Hooks.PreTool → t.Run → Hooks.PostTool → verify (res.Err is a failed act) → record. Consecutive failures past Budget.retries() end the run ESCALATE, one hop up.
- loop/loop.go is the run loop; dirty reads live in loop/guts.go, snapshot in loop/snapshot.go, the journal in loop/journal.go. Tests: loop/loop_test.go asserts the budget-test reading.
- instrument/instrument.go measures callers by id from `.isekai/instruments/*.jsonl`. TestReading proves it.
- compact/compact.go Compress writes one summary per turn; drain (world/drain.go) empties on TurnBudget. Both live in this slime's territory because both own the context's shape.
//@@MISSING g1677
- discover/discover.go finds worlds (`Discover` walking up) and their instructions/skills/commands/agents — the pieces a new session dons first.

- loop/loop.go:874-877 — gutsNote reads every step's output and classifies it; the note rides the tool result back to the model (`Content` patch at 874: `res.Content += gutsNote`). Guts speak inline in the result stream; they never write a parallel channel.
- loop/loop.go:878 — the landing-write reset (`s.gutsSpoke = 0`) fires only when `len(res.Wrote) > 0` or `cls.Class >= tool.Write` AND `!res.Err`. A failed write does not reset the streak; momentum keeps counting.
- loop/loop.go:883-893 — Doom fires when `s.gutsSpoke >= e.Doom`: the loop auto-checkpoints with resume hint (`isekai resume <runid>`), logs `t:budget why:doomLoop:guts` with streak+threshold, then zeroes `gutsSpoke` and `gutsVals` inside the same act that crossed the line (relief is atomic, not deferred to a later beat).
- loop/loop.go:888 — the Doom reason enumerates three arms for the dispatcher in one sentence: retire, split, genome-shift. The checkpoint is the only mechanical act; the dispatch-side choreography is advisory text.
- compact/compact.go:193 — `compactReplay` rewrites `res.Output` to fold consecutive user messages before any gate or context math reads them; the loop's later `perceive()` then measures post-fold tokens. The measurement pipe and the mutation pipe are the same bytes.
- discover/discover.go:242 — `Discover` walks `up` from cwd until it finds a `.isekai` dir or hits filesystem root; no upward bound exists. Discovery trusts directory ancestry, never host config.
- instrument/instrument.go:77-90 — `Measure` iterates `.isekai/instruments/*.jsonl` in sorted filename order (the `os.ReadDir` contract), so caller aggregation is deterministic across runs; the sort is a side-effect of the FS call, not an explicit guarantee.

## Verify
- `go test -count=1 -run 'TestBudgetsAreReadings' ./loop`
- `go test -count=1 -run 'TestReading' ./instrument`
- `go test -count=1 -run 'TestDrain' ./compact`
- `go test -count=1 -run 'TestInstructions' ./discover`
- `go test -count=1 -run 'TestDoomLoopReliefLadder' ./loop`

## Thoughts
- 2026-10-03 census run 3 (brief .isekai/tmp/slime-loop/2026-10-03-loop-census-brief.md): working findings land in `census-working.md` in this dir (both the brief's `.isekai/instruments/loop/` target and `.isekai/tmp/slime-loop/` were refused — engine soil / outside territory); final findings cross on the wire.
- 2026-10-03 (Vitality pairing, orc-engine): the brief at `.isekai/tmp/slime-loop/2026-10-03-loop-census-brief.md` is this slime's commission artefact — planted by the owning orc per the corrected file-brief pattern (the elf never plants in the slime's land; the owning orc plants the working copy). Recorded here in the same turn as the plant.
--- re-authored/adopted by slime-loop 2026-10-03 (loop census pairing)



