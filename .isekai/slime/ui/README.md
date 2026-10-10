# slime-ui

- **Rank:** Slime
- **Territory:** `ui/`
- **Reports to:** orc-interfaces
- **Minds:** slime-go-proving, slime-creature-docs
- **Purpose:** the ground truth of the web UI system: the embedded foundation (tokens, palettes, layout, example components) `ui init` lays down, the scan that derives the legend, the lints `ui check` and the gate run, the catalogue screenshots, and the ui Mind the binary ships.

## Traits
- Foundation (`ui/assets/foundation`, embedded): tokens.css in three tiers (raw scales and palettes → meaning tokens like --surface, --ink, --accent, --space-m, --step-1 → a component's own --<name>-* aliases), palettes.css with 13 OKLCH palettes selected by `data-palette` (aurora calm earth forest ink mono neon ocean pastel retro sunset trust warm) in light and dark, layout.css primitives, the cascade layers `reset, tokens, layout, components, utilities`, and two example components (button, card). Init never overwrites an existing file and returns what it created.
- FoundationFile is the byte source for any page that carries its own copy: board/web's tokens.css, palettes.css and layout.css must stay equal to it (a board test enforces it), so a foundation change is copied there in the same change.
- FindDir tries Candidates in order (ui, src/ui, web/ui, app/ui, static/ui, public/ui, frontend/ui) for a dir holding tokens.css or components/. Scan(root, dir) builds the Manifest (globals, tokens, components with the classes they own, variants, consumed tokens, preludes); Write renders manifest.json, legend.md and catalogue.html under <worldDir>/ui-assets (AssetsDir) and returns only the files it changed; Stale reports drift. The legend is derived, never hand-edited.
- Lint: literal colours and lengths only in TokenFiles (tokens.css, palettes.css); every component has its own css whose first line is the header `/* @component <name> — tokens: <list>. … */` naming exactly the tokens it reads; every token read is defined; every class the HTML uses is declared in the dir's css; a styled #id is a warning. Errors(findings) counts the failing ones; the shipped foundation must lint clean.
- The gate calls this through app's uiGate when law.gate.ui is on (default): a turn that wrote under the ui dir regenerates the legend and must lint without errors.
- Shoot renders the catalogue at Widths 360/768/1280 × Themes light/dark in a headless browser; Browser(env) reads UI_CHROMIUM or CHROME, else looks for chromium, chromium-browser, google-chrome(-stable), chrome, headless-shell on PATH; a snap wrapper gets a staging dir under $HOME; the page's data-overflow attribute names components that scroll sideways. No browser is a reported hole in `ui check`, never a silent pass.
- Skill() is the ui Mind (assets/SKILL.md + craft.md) the app serves as a builtin skill with {{BIN}} and {{WORLD}} filled.
- ui_test runs without a browser; TestShoot runs only with UI_SHOTS=1 and a Chromium.

## Verify
- `go test -count=1 -run 'TestFoundationIsClean|TestLintCatchesDrift|TestLegendIsDerivedAndStaleness|TestInputsAndDataClass|TestParseHeader' ./ui/`
- `go test -count=1 -run TestDashboardHoldsTheUISystem ./board/`

## Thoughts
- (none yet)
