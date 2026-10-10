# slime-compact

- **Rank:** Slime
- **Territory:** `compact/`
- **Reports to:** orc-engine
- **Minds:** slime-go-proving, loop-core, slime-wire-and-journals
- **Purpose:** the ground truth of the drain: when the context reading crosses the threshold, each piece of context moves to its home and the context is rebuilt by pointer.

## Traits
- Two strategies: "drain" (the default — pointerize → trimSpent → one model call for the unsaid and the desk → episode, each pass behind its own Switch in Options.Passes) and "summary" (one summarising call, the generic fallback). Verify runs when its Switch is on and always for "summary".
- Trigger.Threshold(stress) is Trigger.Tokens (0 = the stress line), lowered to Window × Fraction when both are set — the lower wins. Drainer.Perceive lowers the reading's stress line (and re-zones it) to that threshold; Drainer.Hooks() hands the loop Perceive and Drain only.
- cutIndex keeps message 0 (the ask) always and the assistant turns of the last KeepTurns exchanges (0 = 4); fewer than one finished exchange is "nothing to drain" — an abort.
- Any failed pass restores the old messages, journals `drain-abort` with the report and returns an error — the loop then checkpoints. A landed drain journals `drain-report` and sets Session.Last to the rebuilt size, so the next reading is the drained one.
- The rebuilt head is the ask plus @DESK, @D lines, @OPEN, @POINTERS (a pointer whose digest changed is flagged, not trusted) and @DRAINED. The desk file is <world>/instruments/desk/<run>.md; with no model desk a mechanical one is written. More than memory.DeskLimit thoughts is a stress flag in Report.Flags, never a truncation.
- Drained turns are appended to <world>/memory/short/<run>.drain.jsonl (numbered on from the last N) so the `drained` tool can recall them; landed writes call Homes.Episode (the world's log entry) when the Episode pass is on.
- Tokens(msgs) is an estimate (about four bytes a token), used only for the report's Before/After; the trigger reads the provider's usage.

## Verify
- `go test -count=1 -run 'TestDrain|TestVerifyAbortsAndKeeps|TestPassesOffAndSummary|TestTriggerAndPerceive' ./compact/`

## Thoughts
- (none yet)
