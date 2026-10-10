# slime-compact

- **Rank:** Slime
- **Territory:** `compact/`
- **Reports to:** orc-engine
- **Minds:** slime-go-proving, slime-wire-and-journals
- **Purpose:** the ground truth of the drain: when the context reading crosses the threshold, each piece of context moves to its home and the context is rebuilt by pointer.

## Traits
- Two strategies: "drain" (passes pointerize → trimSpent → one model call for the unsaid and the desk → episode → verify, each behind its own Switch) and "summary" (one summarising call, the generic fallback). Options.Trigger.Threshold(stress) decides when.
- cutIndex keeps message 0 (the ask) always and the last KeepTurns finished exchanges; fewer than one finished exchange means nothing to drain.
- A failed pass restores the old messages, journals drain-abort and returns an error — the run then checkpoints (loop.Hooks.Drain false). Verify is mandatory for the summary strategy.
- The desk written at .isekai/instruments/desk/<run>.md is capped by memory.DeskLimit (5); more is flagged as a stress reading in Report.Flags, not truncated.
- Drained turns are indexed into the short memory tier so they stay recallable; landed changes with writes call Homes.Episode (the world's log entry).
- Drainer.Hooks() supplies Perceive and Drain to the loop; Perceive overrides the reading when the drainer knows better. config refuses compaction.trigger.tokens ≥ law.budget.contextTokens (a drain that never fires is a disabled drain).

## Verify
- `go test ./compact/...`

## Thoughts
