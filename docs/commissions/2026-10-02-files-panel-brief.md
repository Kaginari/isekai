# Commission — Dashboard files-activity panel

**From:** rimuru · **Through:** elf-isekai · **To:** orc-host (relay to slime-tui-board + slime-app-commands)
**Date:** 2026-10-02 · **World root:** `/Users/j00000/Workspace/agentic/operation-room-litellm/isekai` (run everything from here; sandbox forbids `cd`).

Orc-host: read this whole file, then dispatch the two slimes. You own them; elf-isekai cannot reach them. One slime NEVER edits another creature's land — it does the writes in its own territory and emits **exact unified-diff hunks** for every neighbour file (never cross-edit). Each hunk is anchored to the file:line it is based on. Brand every handoff `HANDOFF-<rank>-<file>`.

---

## Corrected facts (probe-verified by elf-isekai — override the raw commission wording)

1. `scripts/` does **not** exist → creating `scripts/build-dash.mjs` is genesis; ownership is unregistered.
2. The dash.js embed is at **`board/board.go:28-29`** (`//go:embed assets/... all:web`) — **slime-tui-board land**, NOT app/. Any `//go:generate` must be anchored there (tui-board's own edit), not in app/.
3. **`ui/lint.go` lints CSS only** (`cssFiles()` — no .js/.ts). The premise "lint.go lints the built dash.js" is **false against the code**; no built-bundle exemption is needed. See `ui/` ownership note below, then drop without action.
4. **Pre-existing test failures = 12, not 11** — 11 in `board/` (incl. TestLexiconSwapChangesLabels, TestUsagePageRangesAndChartData, TestColonyPageEmbedsGraph, TestLogSearchAndPartial, TestSSEEmitsOnUsageAppend, TestUsageReadAndRollup, TestLogParse, TestLoopRuns, TestColonyGraph, TestMemoryAndToolboxViews, TestDocIsFenced) + flaky `TestTUIInterruptAndQueue` in `app/`. Gate = **no NEW failures** beyond these 12.
5. dash.html **has no metrics top-rail** — `main` (dash.html:56) holds the `.panel` sections directly. "Collapse the top rail to tabs" reads as: the trident replaces `main`'s layout; under the container query it falls back to the existing one-panel-at-a-time tab behaviour.
6. Dashboard session SSE is served by **app/web.go** (`webEvent` :205, `toolJSON` :244) — the live feed the panel self-feeds on.

## Ownership / no-cross-edit map (probe-verified via onto)

| Path | Owner (rank) | Consequence |
|---|---|---|
| `board/src/**` (new), `board/web/dash.js`, `board/web/dash.html`, `board/web/files.css` (new), `board/web/components/files/files.css`, `board/board.go`, `board/events.go`, `tui/` | **slime-tui-board** | A-slime writes directly (tree, `files` SSE event, `/api/files`, embed+generate here are its own) |
| `app/board.go` (Sources :42-63, opts :65), `app/web.go` (toolJSON :244, webEvent :205) | **slime-app-commands** | B-slime writes directly |
| `board/tree.go` | **slime-world-core** | diff-only (`Kind` type) — NOT orc-host's rung; route UP |
| `board/sources.go` | **slime-world-core** | diff-only (`Sources.Files` field `func() []board.FileChange` :296) |
| `scripts/` (new), `Makefile`, `log.md`, `CHANGELOG.md`, `ui/lint.go` | **no registered owner** | diff-only / hand to elf-isekai |
| `board/web/tokens.css`, `board/web/palettes.css` | slime-tui-board | avoid edits; alias existing tokens only |

## Available tokens (grep-verified — use; define NOTHING new)

`--ok`, `--danger`, `--ink-muted`, `--glow-opacity`, `--time`, `--ease`, `--time-fast`, `--time-slow`, `--radius-pill/-s/-m/-l`, `--step--2/--step--1/--step-0`, `--space-3xs..--space-l`, `--surface`, `--surface-2`, `--surface-sunken`, `--line`, `--hairline`, `--font-code`, `--font-body`, `--focus-width`, `--measure`, `--leading`, `--weight-bold`, `--accent`, `--accent-hero`, `--accent-ink`, `--accent-wash`.
`prefers-reduced-motion` pattern to mirror: `board/web/components/tree/tree.css:47`. `@layer components` is the component tier (see any component css header). Colour mix precedent: `toolcard.css:27-29` (`color-mix(in oklch, var(--ok) 16%, transparent)`).

## SHARED CONTRACT (binds both slimes)

- Session SSE (already live): `{type:'tool'|'bodystep', phase?, tool:{id,name,class,status,wrote?:string[],diff?:{path,add,del,lines:[{k:'+'|'-'|'~',o?,n?,t?}]}}}` — shape set by `app/web.go:205-263` (`toolJSON`). `status:'end'|...` carries outcome; `wrote` = touched paths; `diff` = per-file add/del.
- **New** `GET /api/files` → `{live:bool, root:string, files:[{path:string,kind:'write'|'delete'|'read'|'create',adds:int,dels:int,at:string(RFC3339)}], truncated:bool}`. (`kind` carries 'create' so the UI can mark new files distinctly from generic writes; both treated green/`add`.)
- **New** board SSE event `files` broadcast when the tracked set changes out-of-band (deleted-file purge, truncation roll-off), payload re-shaped like `/api/files` so the panel can resync idempotently.
- The panel must fully function off session SSE alone when `/api/files` is absent or empty (self-feed); `/api/files` + `files` events are an enhancement layer for hydration + cross-session persistence.

---

## SLICE A — slime-tui-board (board/, tui/ land; it OWNS board.go+events.go here)

**Build a writes-only, lazy, low-lag activity panel in TypeScript → `board/web/dash.js`.**

Toolchain (in-world only — NO npm install, NO `cd`):
- node v26.8.2 on PATH; `.isekai/tmp/ts-bin/node_modules/.bin/esbuild` @0.28.2, `tsc` @5.9.3.
- Bundle: `.isekai/tmp/ts-bin/node_modules/.bin/esbuild board/src/dash.ts --bundle --outfile=board/web/dash.js --format=iife --target=es2022`

New `board/src/*.ts` (compile to the bundle; keep every existing behaviour byte-for-byte — the existing `board/web/dash.js` (811 lines) is hand-written JS; treat it as the spec to port into TS modules, NOT to regress):
- `dom.ts` — typed `el`/`svgEl` mirroring current dash.js element helpers.
- `events.ts` — typed SSE payloads mirroring `app/web.go` `toolJSON`/webEvent shapes above.
- `files/treemap.ts` — trie store `{name,path,dir,children:Map, own tally{touches,creates,deletes,adds,dels}, rolled, rolls[] newest-first cap 8, full subtree tally incremented along the ancestor chain}`. Touch = O(depth), NEVER a DOM scan.
- `files/filesview.ts` — renderer into `#files`: `Map path→row`, event delegation `closest('[data-path]')`, rAF-batched rebuild of visible levels only from `rolls` order, deleted rows persist (red), virtualize older/collapsed rows.
- `dash.ts` — entry; wire SSE `tool`/`bodystep` → `treemap.touch`; lazy first render on Files tab; hydrate from `/api/files`; ADDITIONAL panel alongside every current tab (Tree stays).

`board/web/dash.html` (its land):
- New `<button class="tab" data-view="files">Files</button>` beside Tree (:54) + `<section class="panel" data-panel="files">…<div id="files">…</div></section>` in `main`.
- `main` (:56) becomes the layout.css trident (`.cols`/`.sidebar`/`container-type`): left rail `#files-rail` (Σ totals + active dirs), centre = active panel, right rail = `live-bodies`+`kpis` (relocated from the metrics panel — keep their ids so dash.js finds them). Under the container query it collapses back to one-panel-per-tab.
- Add `<link rel="stylesheet" href="web/components/files/files.css">` with the others (:9-28).

`board/web/components/files/files.css` (new component, match the house style):
- `/* @component files — tokens: --ok, --danger, --ink-muted, --surface, --line, --glow-opacity, ... */` header (lint-consumes must match reads).
- `@layer components`; `data-kind="add"`→green `var(--ok)`, `data-kind="del"`→red `var(--danger)`, other→gray `var(--ink-muted)`; heat via `data-hot` + `--glow-opacity` (mirror tree.css :8,33,37,40); reduced-motion swap mirroring tree.css:47; pin `.panel-note` size.

`scripts/build-dash.mjs` (genesis — into elf-isekai/handoff land): esbuild **JS-API** bundler with binary-spawn fallback, IIFE es2022, dev sourcemap, `minify` on `NODE_ENV=production`, log output size. Emit it as a `HANDOFF-elf-isekai-scripts-build-dash.mjs` ready file; elf-isekai will birth `scripts/`.

`tui/` UI rules **of the panel** (apply inside board land): only CHANGED files show; a dir shows only while a file under it changed; each row a `↻n` update-count badge; added/new GREEN, deleted RED, read/other GRAY; NO per-event querySelector (Map path→row + trie rolls; <1 ms/event); lazy render + virtualized old rows; `prefers-reduced-motion`.

`board/board.go` (its own land — DO the writes here): add `mux.HandleFunc("/api/files", b.apiFiles)`; embed stays :28-29; add a `//go:generate` line for the bundle **anchored at the :28-29 embed** (e.g. `build-dash.mjs` invocation) since the embed reads the built dash.js; `apiFiles` reads `b.src.Files()`.

`board/events.go` (its own land — DO): broadcast `files` events when the tracked set changes (hook from the feed/watcher; keep the existing hub/watch semantics :116-145 — `files` is an out-of-band signal, not a polled filesystem walk).

## SLICE B — slime-app-commands (app/ land; it OWNS app/board.go+app/web.go)

**Wire the Go feed + produce the neighbour diffs. Writes stay inside app/ only.**

In `app/board.go` (its own): at the `src := board.Sources{...}` assembly (:42-63), wire a `Files:` func that feeds the source (backing store per HANDOFF-world-core tree.go `Kind`/store). `board.Options` return at :65.
In `app/web.go` (its own): confirm/extend `toolJSON` (:244) so the session-SSE tool events carry exactly the shared-contract shape the panel consumes; no shape breakage of existing fields.

**Handoffs (diff-only, unified hunks, anchored):**
- `HANDOFF-world-core-board/tree.go` — add a file-change `Kind` + store type. NOTE tree.go today is a pure `WalkTree` scanner (:19-72; `_ = os.PathSeparator` :70 is filler) — the diff should introduce a small append-only `Files` accumulator the board polls, NOT turn tree.go into an event bus. Anchor hunks to :12-17 (TreeEntry) / :19.
- `HANDOFF-world-core-board/sources.go` — add `Files func() []board.FileChange` to `Sources` (:296-…), + a `FileChange` type {Path,Kind,Adds,Dels,At}; default in `FileSources`/`withDefaults` (:332,:363). Anchor to :296-301.
- `HANDOFF-tui-board-board/board.go` + `HANDOFF-tui-board-board/events.go` — only if app-commands spots a needed hook tui-board must add (else tui-board owns these and this handoff is empty).
- `HANDOFF-elf-isekai-Makefile` — exact `dash-js` target lines (Makefile is unowned land; elf-isekai applies).
- `HANDOFF-elf-isekai-scripts-build-dash.mjs` — from Slice A.

## INTEGRATION + VERIFY (orc-host does LAST, after collecting both slime handoffs)

Run the bundler, then verify. Show RAW outputs in the report:
```
.isekai/tmp/ts-bin/node_modules/.bin/esbuild board/src/dash.ts --bundle --outfile=board/web/dash.js --format=iife --target=es2022
node --check board/web/dash.js
GOCACHE=$PWD/.isekai/tmp/gocache go build ./app/ && GOCACHE=$PWD/.isekai/tmp/gocache go vet ./app/
GOCACHE=$PWD/.isekai/tmp/gocache go build ./...
GOCACHE=$PWD/.isekai/tmp/gocache go test -short ./board/ ./app/ ./ui/
```
Pass bar: build+vet clean; `node --check` ok; **no NEW test failures beyond the 12 known** (Corrected fact #4). `go:generate` optional to run (needs node at generate time) — building dash.js via the esbuild CLI above satisfies it.

## DOCS (orc-host delegates to elf-isekai — unowned land)

Journal the change in `/log.md` (append, newest-bottom) + a D3 landing card in `/CHANGELOG.md` (surface / consumers / regression-guard=no-new-failures+`node --check` / rollout=`isekai dash` or `make docker-dash` / choices+why / escalation floor). Orc-host: collect slice facts and pass to elf-isekai; elf-isekai births both + registers `scripts/` genesis + applies the Makefile handoff.

## Pairing law (Vitality — gate-enforced at world/gate.go; violating it REFUSES Policy)

- **slime-tui-board:** no write inside `board/` or `tui/` lands unless **`.isekai/slime/tui-board/README.md` moves in the SAME turn** (hard law, human decree 2026-09-30; Law 7 2026-10-01 folds tree panes in too — `board/tree.go` edits refuse Policy absent the pair). Pair the README with every board/tui write batch.
- **slime-app-commands:** pair `.isekai/slime/app-commands/README.md` with its `app/` writes in the same turn.
- Neighbour land (`board/tree.go`, `board/sources.go`, `scripts/`, `Makefile`, `log.md`, `CHANGELOG.md`, `ui/lint.go`) is NEVER written by these slimes — diff-only handoffs, elf-isekai applies.

## Report protocol (wire)

Orc-host: write the consolidated report file to `docs/commissions/2026-10-02-files-panel-report.md` (elf-isekai land — included here is your licence) and answer on the wire:
- `@S done|blocked` then one `@F <path:line> <fact>` per finding.
- **Per-slice authored files**: `@F board/src/files/treemap.ts:1 · ...`
- **Handoff lists grouped by owning rank** (world-core / tui-board / elf-isekai / kijin-...) as `@F <brief-path>#handoff-<rank> · ...`.
- **Holes**: `@? ...` — never guess.
- **Unsaid**: `@U ...` for anything true that disk does not say.
- Raw verify outputs as `@F` facts or in the referenced report file.
- `@E <bytes>` to close.

