# Tron cockpit — landing note

Decided 2026-10-02 in `2026-10-02-dash-redesign-verdict.md`; landed the same day across two
dispatch passes (orc-host raphael → slime-tui-board hands), then verified by elf-isekai. Every
claim cites file:line. This note is the design rationale the README Thoughts point at.

## What landed

| Layer | File | What |
|---|---|---|
| Palette | `board/web/palettes.css:166-173` | aurora evolved in place to the tron values (`--p-1-dark 0.82·0.16·195`, `--p-2 0.70·0.19·205`, `--p-3 0.76·0.17·330`); a tron comment lives at `:163`. No separate `[data-palette="tron"]` block — see §Evolve-vs-new-palette. |
| Component | `board/web/components/tree/tree.css` | new `@component tree`; heat by act class, per-rank body tint, ancestor bubble, reduced-motion variant. |
| Markup | `board/web/dash.html:14,54,112-117` | tree.css link, 6th tab `data-view="tree"`, tree panel section last. |
| Grid | `board/web/components/shell/shell.css:29-34` | trident `@container (min-width: 96rem)`: `"net console tree"` ×2 + `"metrics relations relations"`, columns `3fr / 4fr / 3fr`. |
| Paint loop | `board/web/dash.js:703-808` | one `api/tree` fetch, SSE heat (tool/bodystep/state), rAF-batched, 3.4 s decay, truncated tail note. |
| Feed | `board/tree.go:19-72` | `WalkTree` — flat parent-before-child walk, cap 4096, skip set from app/found.go:60 ∪ loop/snapshot.go:38. |
| Wiring | `board/sources.go:313,358,402-404` + `board/board.go:233,577-587` | `Sources.Tree` field, FileSources fills it, `/api/tree` route + handler. |

## api/tree contract (as shipped)

`GET /api/tree` → `{live: bool, root: string, tree: [{path, name, dir, depth}], truncated: bool}`.
Flat, root-relative slash paths, `depth` top-level = 1, parent before child; `live:false` + empty
tree when the integrator unwires `Sources.Tree` (board.go:578-580). The handler never errors on a
walk failure — entries-so-far + truncated is the honest answer (Perception).

## Layout: container 96rem, three columns

The shell is a container query (`shell.css:4` — an element never queries itself); trident binds
only at or above **96rem**. Below 96rem the existing 72rem twocol (`shell.css:23-28`) takes over,
then onecol under that — the collapse path is trident→twocol→onecol with no panel ever dropped in
overview. Console center is widest (4fr vs 3fr flanks); net left, tree right. All six tabs and
`view()` stay generic (dash.js:67) — the tree tab needs no view() edit.

Deviation from the verdict accepted by the commission brief: verdict §C suggested `5fr/7fr/5fr`
at a re-used 72rem threshold; the brief chose 96rem + `3fr/4fr/3fr` and that is what landed — the
verdict's own dissent ("if the render shows the center console starved, the threshold moves, not
the column weights — decision left to the render read") covers the threshold; the weights are
proportionally equivalent.

## File-heat rules

Acts, not verdicts — the dashboard's second color vocabulary (dash.js:163 already types toolcards
by `data-class`):

- **read** → `--series-1` (tree.css:29); **write/gate/outward** → `--warn` (tree.css:30);
  **destructive** → `--danger` (tree.css:31). `--ok/--warn/--danger` stay semantic elsewhere
  (console tone, net stress) — the tree's `--warn` reuse is an act-lane choice, named in the
  verdict's dissent as a taste call.
- **rank tint** rides `data-rank` mirroring net.css's map (tree.css:21-27): slime→series-1,
  orc/highorc→series-2, kijin→series-3, elf/highelf→series-4, darkelf→series-5,
  rimuru/throne→series-6.
- **ancestry**: a touched path's ancestors glow softer (`data-bubble`, tree.css:40) so a deep
  write still lifts its folder chain.
- **decay**: `--glow-opacity` stepless fade over `HEAT_LIFE = 3400 ms`, one rAF ticker,
  batch-painted (dash.js:770-785).
- **reduced motion**: `@media (prefers-reduced-motion: reduce)` kills the pulse; the static
  highlight stays until decay (tree.css:47-49).

## Contrast (measured, oklch → sRGB → WCAG, against `--n-9 = oklch(0.145 0.022 272)`)

The palettes.css:9 contract is series vs `--n-9` ≥ 4.5:1 on the dark scheme. Measured this run
(the brief's estimates were ~8.5/6.0/7.2/1.9 — all conservative):

| token | role | contrast |
|---|---|---|
| `--p-1-dark` (0.82) | accent | **12.40:1** ✅ |
| `--p-2` (0.70) = series-1/accent-hero | hero/charts | **8.67:1** ✅ |
| `--p-3` (0.76) = series-2 | second series | **8.60:1** ✅ |
| `--p-4` (0.30) | accent-wash surfaces | 1.56:1 — surfaces/borders, **not text**; contract allows (brief :72-74) |
| `--warn` dark (0.8) = series-4 | heat: write/gate lane | **10.46:1** ✅ |
| `--n-3` ink-muted / `--n-0` ink | text | 11.72:1 / 19.03:1 ✅ |

All dark-scheme text/chart lanes PASS. Light theme over aurora remains as before this change.

## OPEN DEBT — s1 vs s6 under CVD

`--series-1` cyan `oklch(0.70 0.19 205)` vs `--series-6` = `--accent` = `--p-1-dark`
`oklch(0.82 0.16 195)`: hue 205° vs 195° — **10° apart**, separated only by lightness
(0.70 vs 0.82); mutual contrast 1.43:1. A deutan (and, on the tree, *any* reader in a dim room)
cannot tell a slime-heat glow from a rimuru/crown tint. Do **not** fix in-place (the series map
is net.css canon too — one change must serve both components). Route: palette-audit; candidate
moves are a hue shift on series-6 (toward 260-280 violet where n-9's own hue lives) or a
chroma/marker distinction (icon or dash-pattern on the crown tint instead of hue alone).

## Evolve-auroa-vs-new-palette — why in-place won

The verdict recommended a separate `[data-palette="tron"]` block. The commission overrode:
**aurora's only consumer is dash.html:2** (grep under board/ proves it — other worlds carry
their own board build), so in-place token mutation has a blast radius of one page. That keeps
twelve palettes, not thirteen; the "tron" name survives as the comment at palettes.css:163, and
`dash.html:2` still reads `data-palette="aurora"`. If a second dash consumer ever wants the old
quiet aurora, this evolve must be undone and tron minted as its own block — the verdict §A holds
the ready-made values.

## Debt & leftovers (reported, not fixed)

- **`live:false` unreachable in shipped wiring** — `FileSources` fills `Tree` (sources.go:358)
  before the nil-check, so the off-state is only reachable through an integrator that explicitly
  nils it *and* bypasses `withDefaults`. The two-return `(entries, truncated)` signature also
  means a partial-walk error degrades silently. Acceptable now; if an embeddable read-only board
  mode appears, gate `FileSources`'s Tree fill on an Options flag.
- **`board.Board.Token()` exported accessor** (board.go:130-135) was landed by the same court
  pass though not in the brief — it backs `TestDashGateAnswerOverHTTP`; standing, paired in the
  tui-board README Thought of 15:07.
- **tree.go's stale trait-bullet history**: a phantom "files-pane watcher" trait from the dropped
  `wip-law7-tree` stash was corrected in-turn; the bullet at tui-board README:13 now describes the
  *actual* walker.
- **Pre-existing red, untouched by design**: `board/testdata/world/.isekai` fixture hole (git
  won't track an empty `.isekai`; tests that need it skip). All three verification lanes green:
  `go build ./board/ ./app/`; `TestEveryPageRendersOnFixtureAndEmptyWorld`; `-run 'Trident|Tree'`
  (5 tests).
- **CVD debt** per the prior section — OPEN, routed to palette-audit.

Pairing receipts: `.isekai/slime/tui-board/README.md` ## Thoughts (two entries, this date) and
`.isekai/slime/world-core/README.md` tail entry of 2026-10-02 carry the in-turn witnesses.

