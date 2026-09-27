# UI — how pages are built

Two kinds of page, one discipline: every page is built from one design system, so a reader who
learned one page can read them all, and a global change is one edit.

## Pages the agent builds for an app — the ui system

The app's own ui dir (`ui/`, `src/ui/`, `web/ui/` …) is the source of truth; the world keeps only
the derived legend. The built-in `ui` Mind carries the method and the craft; `isekai ui` is the
instrument.

- **Tokens in three tiers**, read in one direction: raw (palettes.css `--n-*` `--p-*`, the space,
  type, radius and time scales) → meaning (`--surface --ink --accent --space-m --step-1 …`, per
  theme) → component (a component's own `--card-pad`, aliasing meaning tokens). A retheme, dark
  mode, a density or type change, a radius change: one tier, one file.
- **Cascade layers** `reset, tokens, layout, components, utilities` — a global rule never loses a
  specificity fight.
- **Layout primitives** (layout.css) instead of per-component layout: `.page` (the grid with a
  full-bleed escape), `.stack` `.cluster` `.grid` (auto-fit, no breakpoints) `.cols` (12 columns by
  container width — the Bootstrap idea, container-driven) `.sidebar` `.switcher` `.center` `.cover`
  `.frame`. Components respond to the space they get (container queries), not the screen.
- **One component = one dir** `components/<name>/<name>.{css,html}`, one root class, variants as
  `data-*`, states as ARIA; its css starts `/* @component <name> — tokens: … */`.
- **The legend is derived**: `ui scan` writes `<world>/ui-assets/{manifest.json,legend.md,catalogue.html}`.
  Nobody edits it; the agent reads legend.md before any ui change.
- **Proof**: `ui check` lints (no literal colour or px outside the token files; every class
  declared; every token defined; headers true) and renders the catalogue at 360/768/1280 in both
  themes, failing a component that scrolls sideways. The gate runs the lints on every turn that
  touched the ui dir (`law.gate.ui`); the screenshots are read by the agent before "done".
- **An existing system wins**: a project on Tailwind, Bootstrap or a design-system package keeps
  it; the principles apply inside it.

## The world's own pages — the board

The board and the world's reports follow the rules below.

## Layout: the Bootstrap grid

- **One layout system: Bootstrap 5.** Pages use its container → row → column grid, 12 columns,
  and its breakpoints (`sm` 576 · `md` 768 · `lg` 992 · `xl` 1200 · `xxl` 1400). A panel says how
  wide it is at each breakpoint (`col-12 col-md-6 col-xl-4`); nothing is sized in fixed pixels
  wider than a phone.
- **Mobile first.** The base class is the phone layout; wider breakpoints only add columns. A
  page never scrolls sideways at 360px.
- **Components before custom markup.** Navbar, nav tabs, cards, tables (`table-responsive`),
  badges, progress bars, list groups, modals, offcanvas — Bootstrap's, restyled by the theme,
  never rebuilt.
- **Pages, not one endless scroll.** A page shows one question well (what is running, what did it
  cost, who owns this path); navigation moves between questions.

## Theme: isekai over Bootstrap

- The theme is CSS custom properties layered over Bootstrap's (`--bs-*`), one token per rank and
  mind lane (`--r-slime`, `--r-orc`, `--r-elf`, …, `--l-zone`, `--l-verdict`, `--l-global`,
  `--l-shared`), in a dark and a light variant, both checked with the `palette-audit` Mind
  (contrast, CVD separation) — never the same hex reused for both.
- `data-bs-theme="dark|light"` switches the whole page; the choice follows the system unless the
  reader picks one.
- Motion respects `prefers-reduced-motion`.

## Data on the page

- Charts and graphs are inline SVG sized by their grid column (`viewBox` + `width:100%`), redrawn
  on resize — never a fixed-width canvas. The `dataviz` Mind governs chart form and colour.
- Numbers are instruments: every figure on a page says where it came from (file, time), and a
  silent or stale instrument is shown as such, never as zero.
- Assets are embedded or served locally; a page works offline.

## Proving a page

The ladder (`isekai.md` §The loop) holds for pages too, with one rung more: a page is not done
when its handler tests pass, but when it has been **rendered** — screenshots at 360px and at a
desktop width, in both themes — and looked at. Markup tests pass on pages that still sort lanes
alphabetically, print a pointer instead of a value, or miss a label; only the render shows it.
