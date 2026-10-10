# Consolidated relaunch brief — files-activity panel (wave 1)

**From:** elf-isekai (per Veldora slab vote) · **To:** orc-world (relay to slime-world-core) · **Date:** 2026-10-02
**World root:** `/Users/j00000/Workspace/agentic/operation-room-litellm/isekai` — sandbox forbids `cd`; run everything from here.

This brief consolidates the daemon-side order (`.isekai/slab/2026-10-02-relaunch-files-panel.md:18`), the FE panel contract (`docs/commissions/2026-10-02-files-panel-brief.md`, still binding for wave 2) and the Go skeleton spec (`docs/commissions/2026-10-02-files-changes-skeleton-world-core-brief.md`, EDITS written to the letter) into one launch point.

---

## (a) Toolchain recipe — ONE-BREATH (binding on every `go` act)

Shell acts are stateless (env dies between acts). Every `go` verify MUST run in a single breath:

```
export PATH="$PWD/.isekai/tmp/gosdk/sdk/go1.26.8/golang.org/toolchain@v0.0.1-go1.26.8.darwin-arm64/bin:$PATH" GOTOOLCHAIN=local GOCACHE="$PWD/.isekai/tmp/gocache" && <cmd>
```

**Warning:** never probe machine `~/sdk` (e.g. `~/sdk/go1.26.8/bin/go version`). Bundle D tmp-redirect (`app/guard_tmp.go`, `loop/loop.go` Guard-hook) rewrites those reads into per-run scratch `.isekai/tmp/scratch/<runid>/` — a probe there is a false negative. The in-world staged SDK (`.isekai/tmp/gosdk/`, staged 18:15Z) is the only source of truth.

## (b) Current on-disk state (probe-verified 2026-10-02 evening, HEAD `a8799db`)

| Item | State |
|---|---|
| `board/src/*.ts` (dash/dom/events/files/filesview/files-treemap, 5 files) | UNTRACKED ride-under from a timed-out court — **UNVERIFIED**, wave-2 material |
| `board/web/components/files/files.css` | UNTRACKED ride-under — **UNVERIFIED**, wave-2 material |
| `board/web/files.css` | does NOT exist (only the components/files path above) |
| `board/web/dash.js` | 811 lines, hand-written, tracked — **the live spec** the TS port must not regress (panel-brief :52) |
| `board/sources.go` | **MODIFIED, uncommitted** — world-core brief EDITS (a)-(d) already applied verbatim on disk (`FileChange` type :296-303, `Files` field :324-326, default :372, `withDefaults` :419-421); gofmt clean |
| `board/tree.go` | UNTOUCHED — world-core brief EDIT 1 (`Kind` field on `TreeEntry` :12-17) is still missing |
| `.isekai/slime/world-core/README.md` | not yet moved — pairing law still owed for the sources.go batch |
| HEAD `a8799db` | builds green with the staged toolchain |
| `stash@{0}` (WIP on 5f9f7e5: board.go, events.go, board/src/*, files.css copies) | Go spares therein are **superseded** by the world-core brief's forward-schema version — do NOT pop |

## (c) Wave split

- **Wave 1 (this turn):** world-core Go skeleton only — `board/tree.go` EDIT 1 + confirm `board/sources.go` EDITS (a)-(d) + pairing README move. Full spec: `docs/commissions/2026-10-02-files-changes-skeleton-world-core-brief.md`.
- **Wave 2 (later turn, NOT now):** FE panel (`board/src/*` TS → bundled `board/web/dash.js`, dash.html trident, files.css) + app wiring (`app/board.go` `Files:` feed, `app/web.go` toolJSON) via orc-host. Contract: `docs/commissions/2026-10-02-files-panel-brief.md`.

## (d) Pairing law (Vitality — gate-enforced at world/gate.go)

- **slime-world-core:** every `board/` write batch moves `.isekai/slime/world-core/README.md` in the SAME turn (dated line noting the new `FileChange` type / `Sources.Files` feed / `TreeEntry.Kind` field).
- Wave-2 slimes, for the record: slime-tui-board pairs `.isekai/slime/tui-board/README.md`; slime-app-commands pairs `.isekai/slime/app-commands/README.md`.

## (e) Verify bar

Build + vet clean **under the one-breath recipe**; no NEW test failures beyond the known baseline of **12** (11 in `board/` — TestLexiconSwapChangesLabels, TestUsagePageRangesAndChartData, TestColonyPageEmbedsGraph, TestLogSearchAndPartial, TestSSEEmitsOnUsageAppend, TestUsageReadAndRollup, TestLogParse, TestLoopRuns, TestColonyGraph, TestMemoryAndToolboxViews, TestDocIsFenced — + flaky `TestTUIInterruptAndQueue` in `app/`).

---

## Wave-1 commission to orc-world — exact asks

1. Relay `docs/commissions/2026-10-02-files-changes-skeleton-world-core-brief.md` to slime-world-core **verbatim**, with the toolchain-recipe correction below.
2. State the disk prelude explicitly so the slime does not redo work: **EDITS 2a-2d (sources.go) are already applied on disk** — verify them against the brief hunks character-by-character and gofmt-cleanliness; **EDIT 1 (tree.go `Kind` field) is NOT applied** — apply it exactly as written, then `gofmt -w board/tree.go board/sources.go`.
3. Pairing: `.isekai/slime/world-core/README.md` must move in the same turn (dated line: `TreeEntry.Kind` + `FileChange` type + `Sources.Files` feed).
4. Do not touch anything else: no `board/web/**`, `board/src/**`, `board/board.go`, `board/events.go`, `app/**`, `Makefile`, `package.json`; leave the untracked ride-unders and `stash@{0}` alone.

