---
name: loop-core
description: The six-beat agent run (provider call, tool act, wire read/write, drain, verify/court, compact+usages) — worn by the creatures of loop/, wire/, compact/, instrument/, and app-court/app-engine when the run itself is the ask. Back-ticked file/module reads anchor the facts.
---

# loop-core

- A turn in this code is the same six beats everywhere: provider → tool act → wire → drain → verify/court → compact+usages. Each beat is one function family in one package; a seventh beat does not exist.
- Beat 1 (provider): provider/provider.go FanOut → provider/env.go env reads → provider/openai/openai.go or provider/anthropic/anthropic.go Request. The mock provider (provider/mock/mock.go Return=fixed text stream) is the test stand-in.
- Beat 2 (tool act): app/worldtools.go tools run first; missed names descend app/shelf.go (registry -> custom -> tool/missing.go doom marker). tool/dispatch.go is the walk for dispatch of a task.
- Beat 3 (wire): wire/wire.go is THE wire — `Emit` frame, `Unsaid` flag, report lines `@S/@E/@F/@?`; byte-stable literal tag constants are here, never formatted elsewhere.
- Beat 4 (drain): world/drain.go + compact/compact.go — Recall emplaces memory, drain drops on turn limit (config'd `TurnBudget` and `MaxInMemory`).
- Beat 5 (verify/court): app/court.go + world/gate.go — Court blocks collect per-snare doc/beats; Verdict lands in wire + bench/runs.jsonl.
- Beat 6 (compact+usages): compact/compact.go Compress + instrument/instrument.go Measure; tokens recorded to board/usage.go.
- The wire package tag constants are the contract. Never generate them by string concat — always `wire.Emit`/`wire.Unsaid`. `wire_test.go` covers multi-line frame packing; run it after any wire touch.
- Dispatch is NOT free: tool/dispatch.go spawns a Court (fresh context, tools cut) — minds with big context body live there, not inline.

**Wall contract (commit-per-slice + tree-sync-before-build)** — seat of the wall-contract commission (orc-engine routing, 2026-10-04, doom-2026-10-04-dispatch-wall). Commit-per-slice: when a commission lands slices, seal each slice on landing before the next breath — never accumulate a sack of uncommitted work across slices. Slime ranks have no git-write seat; their slices die with their context — armor ride is required: the answering orc seals the slime's landed slices in the same turn the slime reports DONE. Tree-sync-before-build: no build/act that consumes the tree runs against a dirty tree unless the dirt is the commission's own mid-flight state; the acting body seals or checkpoints the delta first. Wall-hit (checkpoint) protocol: on wall, restate deltas already sealed, name what remains uncommitted, CHECKPOINT with the delta named. The mechanical seat at tool/dispatch.go:237 (wall-fires path = loop/loop.go checkpoint trigger) remains doc-law-only unless an honest minimal hook exists — the slime-loop reading found none without rewriting the loop's beat structure.
- The loop stands — provider stands — guard stands. Dispatch yourself a review loop in app/court.go to self-heal an invalid doc; loop/loop.go's close-on-invalid-doc walks you there.

