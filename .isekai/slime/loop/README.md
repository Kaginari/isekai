# slime-loop

- **Rank:** Slime
- **Territory:** `loop/`, `instrument/`
- **Reports to:** orc-engine
- **Minds:** slime-go-proving, slime-class-contract, slime-wire-and-journals
- **Purpose:** the ground truth of the turn engine (perceive → recall → plan → act → verify → record per tool step, budgets as readings, the journal, resume) and of the context-occupancy instrument.

## Traits
- Session.step runs in this order: step and wall-clock budgets (CHECKPOINT with a resume hint) → a missing tool (Hooks.Missing; the name `_malformed` carries a parse fault from a text-tool-call provider) → t.Settle → Hooks.Guard (status `refused`, by guard) → env.Policy (territory; `refused`, by policy) → Hooks.Decide (deny/allow/ask; allow marks the gate answer By "rule") → gate.Ask → dry-run (`would-ask` / `would-run`) → Hooks.PreTool → t.Run → Hooks.PostTool → verify (res.Err is a failed act) → record. Consecutive failures past Budget.retries() end the run ESCALATE, one hop up.
- Every beat is journaled to .isekai/instruments/loop/<run-id>.jsonl as events keyed "t" (perceive, gate, act, verify, record, hook, drain-report, drain-abort); loop.State and loop.Runs rebuild a run from the journal, and Engine.Resume re-gates an interrupted outward act instead of replaying it.
- Writes are detected by stamping the tree (snapshot.go): a shell command never names what it wrote, so any Write-class act diffs the tree before and after. Past SnapshotCap (200000 files) the stamp is skipped with the reason and writes go unseen — a hazard on huge trees. IsTestFile decides what the gate's Tests-intact check watches.
- A failed end gate (Hooks.EndGate) goes back to the model as the next message up to Hooks.GateRetries times in the same turn (errGateRetry), then the turn fails.
- Hooks.Guard nil means no guard at all: the app must wire it (app/guard.go); a bare Engine in a test is unguarded by design.
- Hooks.Stream and Hooks.Observe see every body's text and steps, including Courts'; Hooks.Inbox rides lines typed mid-turn on the next tool result.
- instrument: DefaultBudget 200000, DefaultStress 180000; zones OK, NEAR (past 75% of stress), STRESS, UNREAD. UNREAD is a finding, never a zero; the reading is provider-reported usage (Input + CacheRead + CacheWrite), not an estimate.
- Result.Emit(cap) renders the wire report under @CAP; a cut is named as a hole.

## Verify
- `go test ./loop/... ./instrument/...`

## Thoughts
