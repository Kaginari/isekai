# Brief — Walking tour of the loop package (slime-loop via orc-engine)

Commissioned by: elf-isekai (relaying the human's question, routed down the tree after a direct throne->slime dispatch was refused by layer flow).

Deliverable: a condensed walking-tour report of slime-loop's territory. Findings must bubble back up through orc-engine to elf-isekai, condensed with `file:line` anchors — this feeds an answer to the human, not a doc.

## Scope — the tour route

1. **The six beats as they run in `loop/loop.go`** — the run loop's phases named in creature speech: perceive / recall / drive / step / finish. Where each beat lands in code, in what order they fire, what each one reads/writes.
2. **Budgets as readings** — how steps / wall-clock / context window are checked before every step, and the checkpoint + resume flow (`loop/snapshot.go`, `loop/checkpoint*.go`, `resume` path in loop.go).
3. **The compact replay-fold** (`compact/`, `loop/compact.go`) — how compaction folds history, what a compact event looks like in the wire, usage accumulation after fold.
4. **Guts scoring + the doom-loop relief ladder** (`loop/guts.go` and related) — how acts are scored, what the relief ladder's rungs are, how it speaks inside tool results.
5. **The journal event taxonomy** (`loop/journal.go`) — the event kinds the journal defines, with their names and what each records.
6. **Test-pinned invariants** (`loop/loop_test.go` and sibling `*_test.go`, incl. `loop/checkpoint_resume_test.go`) — pin these five, with the test names + anchors that pin them:
   - budgets-as-readings (checkpoint+stop, not hard kill)
   - witness demands
   - gate-once-per-turn
   - unsaid bounce
   - doom-loop ladder

Territory of slime-loop: `loop/`, `compact/`, `discover/`, `instrument/`, `.isekai/instruments/*.jsonl` (orc-engine owns `.isekai/instruments` itself).

## Exit greps (sanity anchors — do not stop at these; read around them)

- Six beats: `grep -n "perceive\|recall\|drive\|beat" loop/loop.go | head -40`
- Budgets: `grep -n "Budget\|budget\|checkpoint" loop/loop.go | head -40`
- Compact fold: `grep -n "compact\|Compact" loop/loop.go compact/*.go | head -40`
- Guts: `grep -n "guts\|Guts\|doom\|ladder" loop/guts.go loop/loop.go | head -40`
- Journal taxonomy: `grep -n "EventKind\|event kind\|Kind =" loop/journal.go | head -40`
- Invariants: `grep -n "func Test" loop/loop_test.go loop/checkpoint_resume_test.go | head -60`

## Report shape wanted back

- One section per tour stop (1-6), bullets with `file:line` anchors.
- A short "flow spine": the run loop's beat order in one numbered list.
- 2-3 sentence summary at the top suitable for quoting to the human.
- Holes flagged, never guessed.

