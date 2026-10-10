# Verdict — the dash redesign: tron theme, world-tree column, three-column cockpit

Decided 2026-10-02 by elf-isekai (cross-rules voice), the tui-board territory read in
full (`board/web/**`, `board/board.go`, `board/sources.go`, `board/live.go`, `app/ui.go`).
Every rule cites file:line. Recommendation per section, then the landing list.

## (A) THEME — commit to **tron**, a new named palette; do not evolve aurora

Reason: the palette system is *swap by attribute* — `<html data-palette="tron">`
(board/web/palettes.css:16) — and aurora already serves the quiet dashboard voice
("live dashboards and graphs", palettes.css:162-163). Editing aurora borrows every
existing dash page's look onto the new cockpit; a new block is one appended palette,
zero risk to the old. The header contract decides the member list
(palettes.css:4-16): all ten `--n-*`, `--p-1`, `--p-1-dark`, both inks, `--p-2…--p-5`,
all oklch, `--p-1-dark` ≥4.5:1 on `--n-9` (palettes.css:9).

Exact values, appended after the aurora block (palettes.css:172), comment style
mirroring the neighbours:

```css
  /* tron — near-black blue, luminous cyan accent, violet counter; the cockpit: machine
     glow on deep night, signal colours for instruments, dark-first */
  [data-palette="tron"] {
    --p-1: oklch(0.50 0.12 225); --p-1-dark: oklch(0.88 0.145 190);
    --p-1-ink: oklch(0.986 0.005 262); --p-1-dark-ink: oklch(0.155 0.024 262);
    --p-2: oklch(0.78 0.15 222); --p-3: oklch(0.64 0.21 330); --p-4: oklch(0.34 0.09 258); --p-5: oklch(0.94 0.03 210);
    --n-0: oklch(0.986 0.005 262); --n-1: oklch(0.955 0.008 262); --n-2: oklch(0.91 0.010 262);
    --n-3: oklch(0.83 0.012 262);  --n-4: oklch(0.70 0.016 262);  --n-5: oklch(0.58 0.018 262);
    --n-6: oklch(0.47 0.020 262);  --n-7: oklch(0.35 0.020 262);  --n-8: oklch(0.236 0.012 262);
    --n-9: oklch(0.157 0.013 262);
  }
```

Contrast arithmetic (oklch → sRGB → WCAG, computed):
- `--p-1-dark (0.88 0.145 190)` on `--n-9 (0.157 0.013 262)`: **14.35:1** — passes ≥4.5:1
  with triple headroom; the L=0.88 lift carries the cyan far above the line so the chroma
  can stay luminous (this is the aurora trick at palettes.css:165, pushed further).
- `--n-0` on `--n-9`: 18.7:1 — dark-theme `--ink` (tokens.css:104) is safe on the
  near-black base; `--n-0` on `--n-8` (dark `--surface-2`, tokens.css:100): 16.0:1.
- `--p-3 (0.64 0.21 330)` violet counter on `--n-9`: 5.2:1 — fine for its job
  (second series / tags at hero lightness, palettes.css:13; never body text).
- The n-scale keeps L .986→.157 equispaced with a *flatter* 8→9 drop than aurora
  (.236/.157 vs aurora's .21/.145 at palettes.css:170-171), giving the deep base the
  ask wants while the pairwise ladder stays monotone.

**data-theme pairing**: tron is *dark-first* but a legal full palette, so it must be
served with `data-theme="dark"` (or the reader's system dark, tokens.css:116-117 — the
html already sets `<meta name="color-scheme" content="light dark">`, dash.html:7).
Light theme under tron still works (`--p-1` ≥4.5:1 on `--n-0`/`--n-1` per
palettes.css:8 — verify in the palette-audit before landing).

**ZERO-EDIT answer: yes.** With tron present, `<html data-palette="tron" data-theme="dark">`
restyles every surface/ink/accent/series with **zero edits to tokens.css**. Proof: the dark
block (tokens.css:97-136) reads *only* palette members — `--surface: var(--n-9)`,
`--ink: var(--n-0)`, `--accent: var(--p-1-dark)`, `--accent-ink: var(--p-1-dark-ink)`,
`--accent-wash: var(--p-4)`, `--accent-hero/--second/--focus: var(--p-2/--p-3)`,
`--series-1..6` from `--p-2/--p-3/--ok/--warn/--danger/--accent` (tokens.css:84-89) —
never a literal. The only literals in the block are the ramped `--ok/--warn/--danger`
(tokens.css:111-113) and `--shadow` (114), which are theme-tier decisions by design
and stay inside tokens.css for every palette today. Components consume meaning tokens
plus `--series-*` and `--ok/--warn/--danger` (net.css:1's token list), both delivered
by the meaning tier. The `tron` name never needs to appear in tokens.css — the selector
is `[data-palette="tron"]` in palettes.css only, exactly as `aurora` does.

## (B) TREE COLORS — commit to meaning-state tokens for heat, series tokens for rank

File hot states (the running task's act on a path, mirroring the tool `class` — the
dash already renders that class on toolcards, dash.js:163 `data-class: t.class`):

| state | token | where it comes from |
|---|---|---|
| idle (untouched file) | `--n-6` (via `--line`, dark: `--n-6` tokens.css:102) | raw ramp, meaning alias |
| **read** | `--accent-hero` (`var(--p-2)`, tokens.css:82) | per ui.md:12-15 component tier; dash.js sets only `--<component>-*` inputs and types heat via `color-mix` in css |
| **write** | `--series-2` (`var(--p-3)`, tokens.css:85) — heat lands as dashing/density via `stroke-dasharray`/`opacity` (net.css:36, 45 pattern), keeps `--ok` semantic for its existing pass/converge meaning | series tier, cascade-order safe |
| **outward-gate** | `--danger` (highest) / breaching `--warn` | tokens.css:113, 112 dark ramp |
| a second hot axis if ever needed | `--series-4/--series-5` (`--warn/--danger`) still free | tokens.css:87-88 |

The choice deliberately does *not* put ok/warn/danger on read/write: `--ok/--warn/--danger`
are already the console's tone channels (console.css:19-20, dash.js:158's map) and the
net's stress stroke (net.css:25-29). Read/write/gate are *acts*, not verdicts; series
tokens keep the two vocabularies separable on one page.

Per-body rank color: **yes, reuse net.css's pattern wholesale** — `data-rank` → `--node`
custom property (net.css:7-18: slime→series-1, orc/highorc→series-2, kijin→series-3,
elf/highelf→series-4, darkelf→series-5, rimuru→series-6, minds mind:zone/verdict/global
and mind:shared→ink-muted). In tree.css define the same map scoped to the tree root:
`.tree [data-rank="slime"] { --node: var(--series-1); }` etc. Tree files glow with the
*rank of the body touching them*, and the declared token list in the header names
`--series-1..6, --accent-hero, --ok, --warn, --danger` — same discipline as net.css:1.
Duplicating the data-rank block in tree.css is the correct cost (each component owns
its tokens per ui.md:23); never import net.css into tree.css — components stay
self-contained (ui.md:22).

## (C) LAYOUT — yes: container query, tabs collapse in only below a threshold

Commit: keep the existing tabs as the *narrow* navigation; add the three-column grid
only when `.shell` is wide. The mechanism already exists and is the canon's own:

- `.shell` is already the container: `container-type: inline-size` (shell.css:4) — an
  element never queries itself, `.shell-main` is the grid it rearranges (shell.css:1
  comment). No new container needed.
- The overview uses `@container (min-width: 72rem)` for its two-column split
  (shell.css:22-27). The cockpit mirrors this with **three** areas:
  `grid-template-areas: "net console tree"` (net left, console center, tree right) —
  grid-area bindings already exist per panel (shell.css:18-21); add
  `.shell-main > [data-panel="tree"] { grid-area: tree; }`.
- Balance: keep console widest. Suggest `minmax(0, 5fr) minmax(0, 7fr) minmax(0, 5fr)`
  — console holds the conversation, net and tree read as flanking instruments; wider
  than 72rem threshold is optional but **48rem is too narrow for three columns**;
  reuse **72rem** to stay consistent with the existing query (shell.css:22) unless the
  render proves otherwise.
- Tabs: `tabs.css` scrolls horizontally when narrow (tabs.css:3) — that stays. The
  collapse is: at `max-width: 72rem` (container query on `.shell`, view = landing),
  show tabs and stack console/net/tree single-column; at ≥72rem, hide the tab row (or
  mark the three visible ones `aria-selected` together) since all three live side by
  side. The "tabs vanish only when wide" reading: tabs are the *narrow-mode* nav,
  exactly as console's own `@container (max-width: 40rem)` mirrors (console.css:26,39-41).
- Canon bases: components respond to the space they get, container queries not screen
  breakpoints (ui.md:18-21); a page never scrolls sideways at 360px (ui.md:44) — the
  collapse must land at the column level (single-column), not by squeezing three.

## (D) TREE SOURCE — wiring, exact convention

1. **Sources struct**: append `Tree func() []TreeNode` (or a `*TreeView` struct with a
   `Reading` — mirror `Sources.Memory/Toolbox`, sources.go:296-311; the Reading pattern,
   `Reading{Src, Lit, Read, Why}` at board.go:530-535's usage in config, is the honest
   "silent instrument" contract per ui.md:66). File counterpart field name `Tree`.
2. **FileSources constructor**: add `Tree: f.tree` beside `Doc: f.doc` (sources.go:354);
   a nil-able integrator slot, since FileSources' integrator-owned feeds answer empty
   (Court/Config/Off/Session/Bodies/Skills nil closures, sources.go:341-357) — the tree
   of *files* the world walks is file-readable, so fileSources can own it; a *live* heat
   overlay (which body, which act) is integrator-owned Session/Loop data and rides the
   dash event stream, not the tree feed.
3. **withDefaults**: add the nil-merge stanza beside `s.Doc` (sources.go:380-419):
   `if s.Tree == nil { s.Tree = d.Tree }`.
4. **Route**: in `routes()` add `b.mux.HandleFunc("/api/tree", b.apiTree)` beside the
   other `/api/*` handlers (board.go:219-233; put it next to `/api/doc` at board.go:227,
   the other file-reading endpoint). Handler shape mirrors apiCourt (board.go:544-551):
   ```go
   func (b *Board) apiTree(w http.ResponseWriter, r *http.Request) {
       t := b.src.Tree()
       if t == nil { t = []TreeNode{} }
       writeJSON(w, map[string]any{"tree": t, "at": b.opt.Now()})
   }
   ```
   using the existing `writeJSON` helper (board.go:552-556: sets application/json +
   no-store, encodes the value) — no new helper. The hot-file map rides
   `/dash/stream` events and is joined client-side by path, as the toolcard `data-class`
   join already works (dash.js:163).
5. **Embed**: **yes, a new `board/web/components/tree/tree.css` is picked up
   automatically** — the directive is `//go:embed … all:web` (board.go:28); `all:web`
   includes the whole `web/` tree including any new subdir. Add the
   `<link rel="stylesheet" href="web/components/tree/tree.css">` line to dash.html
   beside the other component links (dash.html:9-28) — that's an html edit, not a build
   change. Static serving of `/web/*` is already wired (live.go:72-78).
6. **Dir-walk to reuse**: there is **no existing dir-walk utility in board/** for this —
   sources.go reads fixed journals (`instruments/usage/*.jsonl`,
   `instruments/loop/*.jsonl`, log.md — sources.go:417-419) via `os.ReadDir`
   (sources.go:429-438), not recursion. Two walk-with-skip precedents exist elsewhere:
   `filepath.WalkDir` with `filepath.SkipDir` at app/tui.go:392-399 and app/found.go:72-78 —
   reuse that *pattern* (SkipDir on `.git`, `node_modules`, the world dir's volatile
   instruments) inside a new `func (f *fileSources) tree()`, do not import/depend across
   packages (board imports onto only, board.go:27). Walk from `f.root`, store paths
   `f.rel(p)`-rewritten (sources.go:407-412) so the client joins on repo-relative paths.
7. **Inside-world refusal**: mirror `doc`'s discipline (sources.go:833-841 —
   `errOutside` sentinel, `filepath.Clean("/"+p)`, allowlist of world-dir/skills/agents
   prefixes, `.md` suffix). For the tree the refusal is inverted: the walk enumerates
   *inside* the root and refuses to ascend (`filepath.Rel` must not escape —
   `f.rel` already keeps results repo-relative or passes through unchanged,
   sources.go:407-412; add the guard: if `filepath.Rel(f.root, p)` yields `..`, skip).
   The tree is read-only by construction; no user path input reaches the walk, so the
   apiDoc-style user-input refusal (board.go:415-421 — 404 "not a readable doc inside
   the world") needs no counterpart in apiTree.
8. **@component header**: `tree.css` line 1 must be a header in net.css:1's exact form —
   `/* @component tree — tokens: <every meaning/series token it reads>. <what it is>. */`,
   The lint parses the regex at ui/scan.go:68 (`/\*\s*@component\s+([\w-]+)(.*?)\*/`),
   then checks the header *lists* every token the styles read and *contains no* token the
   styles don't read (ui/lint.go:62,77,80 — declares-missing and declares-dead both
   error). The format in one line, mirroring net.css:1's shape:
   `/* @component tree — tokens: --surface, --surface-2, --ink, --ink-muted, --line, --accent-hero, --second, --series-1, --series-2, --ok, --warn, --danger, --hairline, --radius-m, --space-3xs, --space-2xs, --space-xs, --step--2, --step--1, --font-code, --time, --time-slow, --ease. The world's files as a lit tree: the run's reads in accent glow, writes in series colour, gates in danger. */`
   (final token list must equal exactly what the styles read — the lint's two-way check
   enforces it; the list above is the intended working set from §B).
9. **Gate enforcement**: `app/ui.go`'s `uiGate` fires only when a turn wrote under the
   **app's** ui dir found by `ui.FindDir` (Candidates: `ui, src/ui, web/ui, app/ui,
   static/ui, public/ui, frontend/ui` — ui/scan.go:52-55, used at app/ui.go:154-162).
   `board/web/` is **not** in Candidates, so tree.css/the dash edits will *not* trip the
   in-loop gate; the discipline still holds via the canon (docs/canon/binary.md:264-268
   names the same lint set, and the render-read requirement is the agent's own rung —
   ui.md:28-29, binary.md:267-268) plus the world-level tree walk you approved. Run
   `isekai ui check --dir board/web` manually (the `--dir` flag exists, app/ui.go:73)
   after landing — that is the instrument's documented escape hatch for non-candidate dirs.

## Landing list (order)

1. `board/web/palettes.css` +`[data-palette="tron"]` block (values in §A).
2. `board/web/dash.html` `data-palette="aurora"` → `"tron"`, add `data-theme="dark"`
   (dash.html:2), add the tree.css link.
3. `board/web/components/tree/{tree.css,tree.html}` new (header per §D.8, markup mirrors
   net.html/console.html's one-dir-one-component shape).
4. `board/web/components/shell/shell.css` — add the `[data-panel="tree"]` area + the
   72rem container query for `"net console tree"`.
5. `board/sources.go` — `Sources.Tree`, `FileSources` wiring, `withDefaults` stanza,
   `fileSources.tree()` walk w/ SkipDir pattern.
6. `board/board.go` — `routes()` +`b.mux.HandleFunc("/api/tree", b.apiTree)`, handler per §D.4.
7. `make verify` (or `go test ./board/ ./ui/`) + `isekai ui check --dir board/web` and a
   rendered read of the catalogue page under tron before "done" (ui.md:69-74).
8. Doc pairing in the same turn: `board/README.md` (the tree component + Sources.Tree
   feed) and `docs/canon/ui.md`'s palette header mention stay truthful themselves;
   this doc is the verdict the pairing cites.

## Dissents and holes

- `--ok/--warn/--danger` stay semantic (console tone, net stress). If the human prefers
  gate-danger to read as *the* danger on the page, the tree can alias
  `gate → --danger` with no cost — but the ok=converge / danger=gate collision on one
  screen was weighed and series-token wins. Flag for the human: this is a taste call.
- Threshold 72rem mirrors the existing query; if the render shows the center console
  starved at 72rem, the threshold moves, not the column weights — decision left to the
  render read.
- `TreeNode` shape (name/path/kind/children, heat joined client-side) is the one shape
  not pinned here; it will be pinned by tree.html's consumption in the implementing turn.
- No reuse of `app/tui.go`'s walk: two unrelated packages, the pattern is small;
  importing app→board or a new shared package for ~10 lines of SkipDir is worse than
  duplication. If a third walker appears, it graduates to a helper then.

