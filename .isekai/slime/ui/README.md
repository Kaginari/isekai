# slime-ui

- **Rank:** Slime
- **Territory:** `ui/`
- **Reports to:** orc-interfaces
- **Minds:** slime-go-proving
- **Purpose:** the ground truth of the web ui system: the embedded foundation `ui init` lays down, the scan that derives the legend, the lints, and the screenshots.

## Traits
- Foundation (embedded, written by Init into the app's ui dir; FindDir tries ui, src/ui, web/ui, app/ui, static/ui, public/ui, frontend/ui): tokens.css in three tiers, palettes.css with 13 OKLCH palettes checked for WCAG AA, layout.css primitives, cascade layers, components/<name>/{<name>.html,<name>.css} with a header comment naming classes and variants. Skill() is the built-in ui Mind (skill + craft texts) the app serves as a builtin discovered skill.
- Scan(root, dir) builds the Manifest (globals, components, declared classes, variants, preludes); Write renders manifest.json, legend.md and catalogue.html under <world>/ui-assets/ (AssetsDir); Stale compares what is on disk with what scan would write — the legend is never hand-edited.
- Lint: literals allowed only in TokenFiles (tokens.css, palettes.css); every class a component's HTML uses must be declared in its CSS; every token used must be defined; every header must be true. Errors(findings) counts the failing ones; the gate calls these through app.uiGate.
- Shoot renders the catalogue at Widths 360/768/1280 × Themes light/dark with a headless browser; Browser(env) reads UI_CHROMIUM, else finds chromium/chromium-browser/google-chrome on PATH; a snap wrapper needs a staging dir under $HOME (isSnapWrapper); the page's data-overflow attribute reports sideways scroll.
- ui_test runs without a browser; screenshots are exercised by `isekai ui check` and are skipped when no browser is found (reported, not hidden).

## Verify
- `go test ./ui/...`

## Thoughts
