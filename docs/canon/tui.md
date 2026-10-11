# The terminal UI — Claude Code style

The live session is an inline terminal UI at the level of Claude Code: the conversation flows into
the terminal's own scrollback, a bordered input sits at the bottom, and every event of the loop is
drawn as a readable block, not a raw line. Built on the Charm libraries, v2 (`charm.land/…`: Bubble Tea, Bubbles, Lip Gloss, Glamour) — the
one place the binary takes third-party code; versions pinned in go.sum.

## Layout

- **Inline, not full-screen.** Finished blocks are printed above the live area and stay in the
  normal scrollback (copyable, searchable, survive exit); only the bottom region is live: the
  spinner line, the input box, the footer.
- **A width change reprints.** A terminal rewraps the live area's full-width lines when it
  narrows, and the renderer, counting the rows it drew, then leaves copies behind. So every
  finished block is kept as a render at a width — rendered once per width and theme, then
  cached, so a frame, a new block or a reprint never runs Glamour over the transcript again; once a drag settles (~120 ms) the screen and the
  scrollback are cleared and the transcript is printed again at the new width — Claude Code's
  answer too. A height-only change reprints nothing; `ctrl+l` reprints on demand; `/clear` empties
  the transcript. The welcome waits for the first size, so it is never drawn at a guessed width.
- **The intro** — on a real terminal the mascot (isekai's slime, agent-one's agent: half-block
  sprites, `tui/mascot.go`) drops in on a Harmonica spring, squashes as it lands, blinks, and the
  welcome box then prints around the very place it landed. About a second; `<PREFIX>NO_INTRO=1`
  skips it; tests never run it.
- **Colour in motion** — the title on the mascot's gradient; the spinner's verb shimmers (a soft
  highlight sweeps across it with each spinner frame); the footer's context is a meter, each cell
  coloured along green → amber → red.
- **Welcome** — one card at start: the mascot beside the name on its gradient, the tagline and the
  lane legend as chips; then a label/value grid (small-caps labels) — the world (its tail kept
  when cut), the model, the offices folded to one line (`3 · great-sage raphael → flash · ciel →
  ultra`; `/offices` has them whole), the session, the board URL, a waiting handoff, the off-list
  count; what the engine could not settle, one `▲` line each; one row of where to start (`⏎ ask
  · / commands · ctrl+k palette · ? keys`). It reads at 80 columns.
- **Your message** — a bar (`▌`) in the human's accent down its left edge, the words beside it.
- **The answer** — the turn's first piece wears the primary's `✦` in the gutter; the text sits
  two cells in, so asks, answers and cards each own a column-0 marker.
- **Assistant text** — streamed, rendered as markdown (Glamour, a theme that follows the terminal's
  dark/light background), code blocks highlighted.
- **Tool calls** — one compact line each while running and after: `● bash  go test ./...` with the
  class as a coloured tag (read · write · outward · destructive), then a status line under it
  (`⎿ ok · 1.2s · 34 lines` or the error). Output is collapsed to its first lines with
  `… +N lines (ctrl+o to expand)`; `write`/`edit`/`multiedit`/`patch` show a coloured diff
  (± lines, line numbers, two lines of context, a gap mark between hunks), collapsed past ~20
  lines. The diff is the file before and after the act, so a shell-made change to the same path
  is not shown — only the path tools diff. The blocks are drawn from the loop's `Observe` seam
  (a step at "start" once its class is settled, at "end" once its record is final); `dispatch`
  and `ask` draw no tool block — the Court block and the question block are theirs. A Court's
  own tool steps are not drawn; its block's state says `tool`.
- **Courts** — a dispatched Court is one block with its rank, office and model, a live state, and
  on return its report rendered (status, findings, holes, the unsaid) — not the raw wire text.
- **The gate verdict** — one quiet line after a turn that wrote: a chip in the verdict's colour
  (`✓ gate · n/a (no orcs)`), then the log it went to, dim.
- **Spinner line** — while a turn runs: a spinner, a verb for the current beat (`Thinking…`,
  `Running bash…`, `Waiting for approval…`), elapsed time, tokens so far, `esc to interrupt`.
- **Footer** — under the input: the model · context % · cost (or "no calls yet") · live courts ·
  the world dir; `? for shortcuts`. Under the frame it is the status bar (below).

## The frame

A window of at least 60×26 is owned by one box-drawn grid (`tui/frame.go`) under a blank first
row — the headroom VS Code's sticky scroll paints the launching command over:

- **Header** — the brand on its gradient (the primary's cyan through the plan's violet into the
  accent's rose) and the version, the world, the model; on the right the session's short id and
  its state chip (`○ idle`, `● thinking 12s` in the primary, `◆ waiting for you` in the accent).
  Items yield right to left as the window narrows.
- **Conversation** — the main column, titled in its top edge: the transcript in a viewport
  (pgup/pgdn, shift+↑/↓, the wheel), the live rows under it. Scrolled up, the edge says `↓ N more`.
  While the transcript overflows the pane a slim scrollbar rides its right padding column: a quiet
  `▕` track, a thumb in the primary sized to the visible fraction and placed by the offset at
  half-row precision (`▐`, its ends `▗`/`▝`); a click on it jumps there, a drag walks it. It is
  drawing, never text: a selection never reaches it.
- **Ask** — the input lives inside the main column, under the conversation, its own rule titled
  `ask`. Its well starts three rows tall with a placeholder (`ask anything · / for commands · @
  for files`), soft-wraps a long line inside the box (never a sideways scroll), grows with its
  rows up to eight (or 35% of the window), then scrolls inside; the conversation gives the rows
  up and the frame stays rectangle-exact. The rule's right end says the draft's size (`3 lines ·
  342 chars`), then what enter does (`enter sends`, `enter queues for rimuru` while a turn runs)
  and `shift+enter newline`, yielding right to left. It never reaches under the sidebar. The slash
  menu, `@` completions and the `?` keys sheet rise from its top edge as cards over the
  conversation's last rows — the frame never reflows for them.
- **Sidebar** — the right third, full height from header to status bar, titled `session`: a
  cockpit column of tiles, each its title in its own hue. The header tile: the world's sprite
  beside (or above, when narrow) the world's name, the model and the session's state chip. The
  vitals: a label/value grid — the context meter (each cell in the band its position reads under
  the ctx meter's law, `cfg.Meters()`: ok to the NEAR line, warn to the stress line, danger past
  it — never a number of the terminal's own), the percentage, the cap; the tokens, the cost; `val`
  the throne's own score (its meter, a bar in its band); `asc` the body nearest its ascension
  (`orc-engine · reports 17/20 to ascend`, or how many conditions are left and the nearest); the
  board's link. The races (law.value, `docs/canon/races.md` §Value): each race's dot and name in
  its lane (a custom race in its family's: root the throne's, planner the elf's, gate the orc's,
  author the slime's, steward the kijin's, auditor the dark-elf's), its meter's label, a bar in its
  band's colour and the reading; confident races first by their mean score, a race with no
  confident record dimmed and unranked (`low`), the session's tokens at the row's end on a wide
  column. While the law has no record (off, a young world) the tile is each race's share of the
  session's tokens in its lane. The courts: a roster, the live ones first, each state a chip on
  the right edge (the lane's while it runs, the verdict's once landed), the body's reading beside
  it in its band (dimmed under its record). The plan: a progress bar of the deck's done steps in
  the plan's violet and the count in the plan meter's band once a step is done, the current step
  lit in the plan's violet, the rest under it. The files touched: a Lip Gloss tree. The races,
  courts, plan and files fold to their title: `alt+c` / `alt+p` / `alt+f` or a click on it; a
  click on a roster row shows that body's plan. Taller than its column, the sidebar scrolls with
  the wheel and wears the same scrollbar on its right border (the border its track, a heavy `┃`
  thumb, half-row ends `╹`/`╻`), draggable. `alt+s` or `/sidebar` hides it — the conversation
  takes the whole frame.
- **Status bar** — the context meter (and raw counts on wide windows), the session's tokens, the
  cost, live bodies, queued lines, a mode badge; on the right a transient message or the key hints.

Boxes share their edges: one line where two meet, `┬ ┤ ┼ ┴` where three do, one border tone. Every
row is exactly the window's width and the frame is exactly the window's height less the headroom
— the inline renderer counts rows, and a frame that lies leaves ghosts. The geometry is computed
once per frame and the mouse reads it back: the selection's origin, the sidebar's hit rows, where
the wheel goes — nothing scans glyphs. Below 60×26 the session flows inline as above.

## Type and colour

Inside the cell grid type is weight, case and tone: bold is for titles only; region and section
titles are small caps (`ᴄᴏɴᴠᴇʀꜱᴀᴛɪᴏɴ`), as are the welcome's labels; text, then label, then dim
carry the hierarchy. Titles wear colour, each region its own: the conversation on a gradient of
the primary into the slime's blue, the session the throne's amber into the accent's rose, the ask
the rose into the plan's violet; each sidebar tile's title its own hue (vitals the primary, races
amber, courts rose, plan violet, files blue); a tool card's name its class's hue, a Court's its
race's lane; a panel's title the primary into the violet.

The colours come from one palette — the binary's, shared with the dashboard (package `theme`;
`tui/palette.go` only names it in the terminal's words): a named value, role → hex for a dark and a
light variant, read by one constructor (`NewThemeFrom`) — nothing else names a colour of the
scheme. Config picks it (`ui.theme`: cyberpunk, neon-dystopia, corporate-rebellion, rain-city,
synthwave, glitch, chromecore, realm) and its variant (`ui.mode`: dark, light, or system — the
terminal's answer to the background query; `<PREFIX>_THEME=dark|light` in the environment still
wins). `/theme <name> [dark|light|system]` wears another at once (alone it flips dark and light;
an unknown name lists them), `/settings` picks both; the host writes the choice to the world's
`config.local.yaml` and the dashboard follows, and a switch made on the dashboard reaches the
terminal as `EvTheme` and is worn the same way. The default is omoro.io's **Cyberpunk** (Void `#0D1117`, Smog Purple `#1A0A2E`, Danger
Red `#FF003C`, Amber Haze `#FF9900`, Scanner Cyan `#00FFFF`): the ground is Void, the surfaces
step from it toward Smog Purple; Scanner Cyan is the primary (brand, focus, the live state,
links); a rose from Danger Red is the human's accent (the ask's spine, the waiting chip); the
verdicts are a mint (the palette has no green), Amber Haze and Danger Red lifted to read; the plan
is Smog Purple's violet at text lightness; the race lanes are rimuru amber, elf magenta, orc acid
lime, slime blue, kijin red, dark-elf indigo. Light is the same scheme as deep ink on a
smog-tinted paper. In every palette, both variants, every text role, lane, tool class and the
plan's hue holds 4.5:1 on every surface it can sit on, every row of every gradient and every chip
included (`TestDepthContrast`, run over `theme.All`);
where a raw swatch fails, the palette's role is the swatch retuned, and the palette's comment
says which.

## Depth

On a truecolor terminal the frame is drawn as layered planes (`tui/depth.go`), in colour and the
block elements alone — no animation, no glyph that is not drawing:

- **One tonal scale.** Every surface is a step of the palette's tonal anchors — the palette owns
  the ground (the depth paints every cell of the frame; the terminal's answer to the background
  query only tells dark from light). Elevation `-1` the input well, `0` the ground (the borders
  sit on it), `1` the planes (the conversation, the sidebar, the status bar), `2` the cards, `3`
  the floats. On the dark variant a closer surface is lighter; on the light one the planes are a
  soft lavender grey and the cards and floats paper-white over them. Each surface leans toward
  the hue that says what it is: the human's ask the accent, the answer the primary, a tool its
  class, a Court its race's lane, the plan its violet, a toast its verdict.
- **Raised, with volume.** A card's face is a gradient — each row a step from lit at the top to
  shaded at the bottom (an answer stays flat: its pieces join into one slab); a glossy half row
  `▄` a tone lighter than the lit face runs under its top edge; its left bevel is crisp and lit
  (`▏`, or a coloured spine `▌`), its right bevel two tones deep (a half-tone face under a shaded
  `▕`), its bottom edge `▝▀…` shaded. It casts a shadow two cells right and half a row down: the
  near column dark, the far one a half-width fade (`▌`), its corners quarter blocks (`▖`, `▘`, a
  half-tone `▀` where it begins) so it falls soft; a float's shadow is the deepest, with a faint
  halo of the primary. Text sits inside a cell of padding on every side.
- **Glow.** An active surface lights its rim in its hue instead of a shade: the running Court in
  its lane (its gloss and lit edge lean into the hue, its shadow becomes the hue's spill), a
  fresh toast in its verdict, the ask's well in the primary while the human types.
- **Sunk.** The input sits in a well: its top and left edges in shadow, its right and bottom edges
  lit — the inverse of a card; its first row sits in the lip's shade. It costs the conversation
  two rows.
- **The blocks.** The ask and each answer are slabs with a spine (a turn's answer pieces join into
  one slab, a code fence sunk into its own inset inside it); a tool and a Court are smaller chips,
  their head on the lit band — a Court's chip tinted deeper in its lane than a tool's, its band
  deeper still, its state a chip of its own (the lane's while it runs, the verdict's once landed);
  the plan tool's card wears the plan's violet and shows the deck's progress and the step in hand;
  the welcome is the hero slab, its face drifting from the primary's
  tint to the accent's row by row; a gate verdict is a pill with half-block caps. The header is a
  raised bar on a horizontal drift of the same two tints. The sidebar's sections are stacked tiles
  — the header, the vitals, the races, the courts, the plan, the files — each its own tint, its
  title on the lit band (a click on a foldable one folds the tile). The meters are slim half-row
  bars (`▄`, a half cell `▖` at the fill's end) over a tint of their hue. A toast is a float over the conversation's top-right corner; the ask
  card floats over the frame sunk flat beneath it; the palette and the popups are floats.
- **Text first.** Every text role holds 4.5:1 on every tint and every row of every gradient, dark
  and light (`TestDepthContrast`).
  A drag selection copies the text — the spines, bevels and shadows copy as nothing, and a slab's
  padding is taken off the copy.
- **Fallbacks.** A 256-colour terminal (its tonal steps would band) and a colourless one draw the
  flat frame; so does the scrollback a narrow window prints into.

## Interaction

- **Input** — multiline (`shift+enter` where the terminal reports it apart from `enter` — Bubble
  Tea v2 asks for key disambiguation — and everywhere `\` + enter, `alt+enter`, `ctrl+j`); the box
  soft-wraps and grows with its visual rows from three (one inline) up to eight, then scrolls
  inside — the text itself is never capped;
  history with ↑/↓ on the first/last row, a paste of three lines or more collapsed to
  `[pasted N lines]` and restored on send, `@path` completes files in the world (tab).
- **Slash commands** — typing `/` opens an inline menu of built-in and discovered commands with
  their descriptions, filtered as you type; tab completes, enter runs. `/help` prints the menu
  and the shortcuts; `/quit` ends the session.
- **Mid-turn input** — typing while a turn runs queues the line (shown as queued) and it reaches the
  body at its next tool step, as the loop already does.
- **Approvals** — the human gate is an inline choice block: what, its class, why it needs you
  (`outward: git push origin main`), and options `1 Yes · 2 Yes, and don't ask again for
  bash:git push* (writes an allow rule to .isekai/config.local.yaml and reloads the config live,
  so the next step already reads it) · 3 No, and tell the model why`; arrow keys + enter or the
  digit. A denial ends the turn (the law: a denial stops the run there); the reason reaches the
  model twice — in the denied step's result, and as your own next message, which opens the next
  turn at once. A question the model asks (`ask`) is the same block with its options and a free
  line. The proposed rule is the tool and the act's first two words (`bash:git push*`), the host
  for a URL, the subject itself otherwise; it lands in the local layer, never the project's.
- **Keys** — `esc` closes a menu, else interrupts the turn, else clears the input; `ctrl+c`
  clears the input, else interrupts the turn, else (idle, empty) arms the exit and a second
  within two seconds exits — a mashed `ctrl+c` stops a turn, never the session. A cleared draft
  is one ↑ away, the newest history item. `ctrl+o` re-prints the last collapsed block in full,
  `ctrl+l` redraws, `?` on an empty input shows the shortcuts; `ctrl+p` walks the plan decks and
  `alt+n` snaps back to the current one (no plain letter is ever a chord). Under the panes
  `pgup`/`pgdn` page the chat, `shift+↑/↓` walk it a line, the wheel scrolls it; `ctrl+u`/`ctrl+d`
  stay the input's.
- **Markdown** — `<` and `>` outside code are text, never HTML: `Vec<T>` reads `Vec<T>`.
- **Errors** wrap to the width; nothing is clipped.
- **Non-interactive** — without a TTY on both stdin and stdout (pipes, CI, `run`, `TERM=dumb`) the
  plain line output stays: the TUI is only for a terminal. `--plain` (or `<PREFIX>PLAIN=1`)
  forces the line REPL anywhere.

## The cast — faces, voices, thinking, every body

- **Thinking, shown.** The model's reasoning (Anthropic's thinking, OpenRouter's `reasoning`,
  `reasoning_content` elsewhere) streams dim under `∴ thinking` while it arrives and folds into
  `∴ Thought for Ns` when the answer, a tool or the turn's end comes; `ctrl+o` expands the last one.
  Shown, never replayed from here.
- **Faces.** Each rank has a two-frame half-block icon beside the spinner — the slime (the session
  and the slimes), the orc with its tusks, the elf, the horned kijin, the dark elf; agent-one's robot
  in each rank's colour. It bobs and blinks with the spinner.
- **Voices.** The verb is the rank's own, like Claude Code's whimsy but on the law: the session
  ponders and consults the Great Sage, the slime gathers ground truth, and while the end-of-turn gate
  runs the line wears the orc — "Weighing the verdict…", "Guarding the gate…" — with "orc · the gate
  weighs the turn's writes" under it. A tool or an approval keeps the plain word.
- **Every body — `ctrl+t`.** A full-screen view with a tab per body: the session and every Court
  seen. A Court's tab is its own log as it happens — its thinking, its text, its tool steps as cards —
  under its face, rank, state and ask; the session's tab is its transcript. The live area names the
  running Courts with the key to watch them. `tab` switches, `↑↓` scrolls, `esc` returns; `i`
  pipes a line to the selected body's inbox, borrowing the input line and giving the draft back.

## Overlays

Drawn as Lip Gloss v2 layers over the live area (the scrollback above is the terminal's, never
overdrawn):
- **The palette — `ctrl+k`.** Every command, the UI's own (board, theme, settings, sidebar,
  offices, help, clear, quit) and
  the host's, fuzzy-matched as you type: a subsequence, word starts and runs scoring higher, the
  matched letters lit. Enter runs a command that takes nothing at once; one that takes arguments
  lands in the input to be finished. `theme` flips dark and light in the palette worn and
  reprints the transcript.
- **Toasts.** A Court landing (✓ done / ✗ failed) and the gate's verdict raise a one-line pill at
  the live area's top right (under the depth a floating card, the highest elevation); it slides in
  on a spring and leaves after four seconds, three at most.
  The block it echoes is in the scrollback already — a toast is a glance, never the record.
- **Settings — `/settings`.** A huh form floating over the frame: the palette (every one the
  binary ships), its mode (dark, light, system), who owns the mouse, the sidebar — inline selects (`←→` change, `tab` next, `enter` saves, `esc` closes unchanged).
  The form is a Bubble Tea model inside the session's: while it is open every message reaches it
  first.

## Blocks as cards

- **Tool cards** — every line of a tool block carries an edge (`▎`) in its class's colour: read
  blue, write magenta, outward amber, destructive red — the class is seen before it is read; the
  plan tool's edge is the plan's violet. Under
  the depth the edge is the raised chip's spine and the chip is tinted in the same hue.
- **Diffs in colour** — added and removed lines keep their `+`/`-` and line numbers, their code in
  its language's colours (chroma, per line; catppuccin-mocha dark, github light) over a green or
  red tint to the card's edge; context lines coloured, untinted. Unknown language or no colour: the
  plain `+`/`-` lines.
- **Board tables** — Agents and Offices are Lip Gloss tables (rounded, dim bold headers, numbers
  right-aligned); the selected row carries a gold `›` and a bold name, not a background (cells keep
  their own colours).

## The terminal around it

- **Window title** — `<dist> · <world> · <state>` (idle, thinking, running bash, waiting for you).
- **Tab progress** — while a turn runs the terminal's tab shows an indeterminate progress mark
  (OSC 9;4, where the terminal draws it); an open approval shows a warning mark.
- **Links** — a tool block's path is an OSC 8 hyperlink to the file: a click opens it where the
  terminal supports links, plain text elsewhere.
- **Mouse on the board** — a click on a tab switches to it, on a row selects it (again: its
  detail), on a graph node selects it; the wheel scrolls.
- **Mouse under the panes** — on a window wide enough for the sidebar the transcript lives in the
  chat pane, so the program takes the mouse: the wheel scrolls the chat (the sidebar over its own
  column), a click on a sidebar section's rule or a roster row acts on it, a click or a drag on a
  scrollbar walks its pane. A drag selects in the
  conversation only — never a border, the input or the sidebar — scrolling when it rests on the
  pane's edge; a double
  click takes a word, a triple click a line; `esc` clears it. The release copies the plain text
  (escapes stripped, each line's pad trimmed) by OSC 52, to the primary selection too, and —
  off SSH — through the desktop's own tool (`wl-copy`, `xclip`, `xsel`, `pbcopy`), as some
  terminals ignore OSC 52; a toast says how many lines. `ctrl+y` or `/copy` copies the last
  answer's markdown. `/mouse` gives the mouse back to the terminal — its own selection, a
  `mouse: terminal` badge in the footer — until `/mouse` again; `shift+drag` does the same for
  one drag on most terminals. On quit the transcript the panes held prints to the scrollback.
  A narrow window never captures the mouse: the terminal selects as always.

## The board — `/board`

The web board's feeds drawn full screen in the terminal (alt screen); `esc` gives the session back
with its scrollback untouched. Blocks that finish while it is open wait and land once when it
closes; a width change meanwhile reprints the transcript on the way out. Four pages, `tab` or `1`–`4`:

- **Agents** — every live body: state glyph, rank, office, state, age, a context bar, tokens,
  cost, model (columns drop by priority on a narrow screen); `enter` opens the body's detail.
- **Graph** — the reasoned ontology as a layered graph: the session on top, a hop down per row,
  each one-hop bond (truth · verdict · reports · above — the web board's own edges) drawn from the
  child up to its parent and colored by bond; `◇` marks a creature with shape findings. Arrows walk
  it. The selected creature's knowledge card sits beside it (below when narrow): its classes, the
  inferred `above` chain, what is under it, doc, territory, minds worn, facts it knows and sees by
  the flow rules, its findings — the graph answering, not a document.
- **Offices** — the triad as config resolves it: what each does, its model and fallback, where
  that was set, the ranks it serves, live bodies in it, tokens spent.
- **Usage** — the usage journal by range (`r` cycles 24h · 7d · 30d · all): totals, a by-day
  sparkline, bars by body, model and office.

## Voice

Both distributions share the TUI; every label comes from the lexicon (agent-one says Subagents,
Roles, Workspace). Colours: the board's rank and lane tokens, adapted to 256-colour and truecolor
terminals; no colour when `NO_COLOR` is set (an ASCII colour profile drops every escape, so a
plain run is the same text; Lip Gloss v2 styles in full colour and the program's writer downsamples
to the terminal's profile). `<PREFIX>THEME=light|dark` or `COLORFGBG` decide the theme when set;
otherwise the terminal is asked for its background colour and the welcome waits for the answer at
most 150 ms — a terminal that never answers keeps the dark theme, so a start is never held.

## Proving it

The ladder applies, with the render rung of `ui.md`. Rung 1 is the view model (`tui/blocks.go`:
pure functions from a view to styled text, golden tests at 80 and 120 columns with the escapes
stripped, `-goldens` rewrites them). Rung 2 is the Bubble Tea program under teatest with a fake
host. Rungs 3–4 are the junction with the real engine (`app/tui.go` under teatest: the mock
provider scripted, real tools, the real classifier behind the gate — the three approval paths,
streaming, diffs, Courts foreground and background, interrupt, queueing). Rung 5 runs the built
binary in a pseudo-terminal against a scripted world (`.isekai/tmp/tui-shots/shoot.py`), feeds
keystrokes, and renders the ANSI stream through `pyte` at 80×24 and 120×40 into plain-text
screens — welcome, a streamed markdown answer with a code block, a bash tool block, an edit with
its diff, a collapsed output and its expansion, the approval prompt and its three paths, a Court
block with its report, the slash menu, the shortcuts, `/agents`, `@path` completion — looked at
before it is called done. Finished blocks leave the program through one FIFO printer, so they
land in the scrollback in the order the loop produced them.
