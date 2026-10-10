# Brief — TUI UX program, two parts (routed elf-isekai → orc-host → slimes)

Commissioned by: elf-isekai, 2026-10-03. The ask below came from the throne; twice it failed there (layer-gate refusal direct to slime, then spawn death). Route it through your own chain so each body authors only its own territory.

## Plan

Two work packages; PART 2 compiles against PART 1, so sequence them. Both targets sit under your tree (`orc-host`).

- PART 1 → `tui/` — owned by slime-tui-board
- PART 2 → `app/tui.go` — owned by slime-app-commands

Dispatch slime-tui-board first with PART 1 only. When it reports green and its creature doc + CHANGELOG have moved, dispatch slime-app-commands with PART 2 only. Each slime edits only its own territory plus its own creature doc (Vitality pairing).

## PART 1 — three changes inside `tui/` (slime-tui-board)



1. **RAIL SCROLL** — the rail gets its own scroll offset on the Model (reset/clamped on resize and window-mode change); wheel scrolls the rail when the pointer is over the rail column (hit-test near `tui/model.go:888`), the pane otherwise as today; offset clamped to content height; ▲/▼ markers when content is clipped.
2. **COLLAPSIBLE CARDS** — plan / courts / files cards collapse to a one-line header: expanded header shows `▾`, collapsed shows `+` (session card exempt). Toggle state lives on the tui Model, not app's RailView. Toggle via mouse click on the header line and keyboard (pick free keys, e.g. `alt+p`/`alt+c`/`alt+f` — verify no collision; alt+enter, alt+j are taken). All new bindings join the Shortcuts overlay.
3. **FOOTER TOKEN COUNT** — FooterView gains raw token fields (e.g. `CtxTokens int`, `CtxLimit int`; presentation stays in tui). Footer renders `ctx ▮▮▮▯▯▯▯▯ 42% · 84k/200k` — compact humanized (84k not 84,213). Gate the `· 84k/200k` segment on width (keep when width >= ~88; existing `>=72` gate for the meter unchanged). Zero values (PART 2 not yet landed) must render exactly as today.

Constraints: `charm.land/bubbletea/v2` + `lipgloss/v2` idioms (`t.dim`, `t.accent`, railCard lane colors). No changes outside `tui/`. Tests for scroll clamp + indicators, collapse toggle render (the `+`/`▾` markers), footer token formatting + zero-value fallback; run `go test ./tui/...` and `go vet ./tui/...`.

## PART 2 — `app/tui.go` fill (slime-app-commands)

`Footer()` at `app/tui.go:229-242` already calls `readingOf` (defined `app/court.go:325`) which returns raw `r.Tokens` and `r.Limit` — fill the new FooterView fields from that reading. No other file changes; keep app tests green (`go test ./app/...`).

## Both parts — Vitality

Update the owning creature docs (`.isekai/slime/tui-board/`, `.isekai/slime/app-commands/`) and add a CHANGELOG entry in the same change (Vitality law). If a slime may not write `CHANGELOG.md` at the root, pair its creature doc and name the CHANGELOG line in the report so elf-isekai pairs it.

## Report shape wanted back

Which bodies staffed the work; `file:line` anchors for every edit; keys chosen (and the collision check); `go test`/`go vet` output; and the doc/CHANGELOG updates.
