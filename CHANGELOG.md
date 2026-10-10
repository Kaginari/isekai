# Changelog

All notable changes to Isekai are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow
[Semantic Versioning](https://semver.org).

## [Unreleased]

### Added

- **The dashboard's overview is a cockpit.** The neural net is the hero, the console docks beside
  it (drag the splitter to resize, fold it to a rail — both remembered), and a timeline of Courts
  and verdicts sits under the net (`/api/activity` now carries the newest runs, their wire headers
  dropped); a row focuses its neuron. The header gains the plan's progress. The folder tree leaves
  the grid for a drawer over every view — the files button, `t`, or Alt+T from the prompt; Esc or a
  click away closes it — and hovering a path lights the neuron that owns it. Every action pulses
  on its own neuron in 2D and 3D: reads ripple softly, writes pulse in the rank's colour, outward or
  destructive steps and a waiting gate ring amber, verdicts flare green or red; the throne breathes
  while the session runs.

- **The cockpit, reviewed.** The net fills exactly the box its panel gives it with the legend in
  view; each subtree under the throne reads as a lane band, the kijin and an elf with no orcs sit
  as satellites just under the throne, every upper layer is centred, rows and tiles tighten before
  the slimes wrap, and no label is cut below six ems; the minds become a strip of brains under the
  layers. A "now" strip over the net carries the session's state, its ask, the live bodies with
  their context use and a wait on you. The timeline keeps its history (a journaled run folds into
  its live row only when it is the same run) and names the throne by the lexicon. A plain turn's
  end pings calmly in the rank's colour — green and red are for real verdicts — and a burst of
  steps is one breathing neuron. The 3D stage stands as layers (spaced planes, rank captions,
  depth fog, smaller halos on the light theme), fills the panel, walks the tree from the keyboard,
  has a "reset view" and a first-use hint, and idles at 12 frames a second, its turn stopping
  after an idle minute. The tree drawer is a modal dialog; the phone header is two rows; sideways
  rows fade at the edge that has more. `/api/colony` answers an empty world with lists, never
  null; `/api/activity`'s recent runs carry their id.

- **The neural net, rebuilt.** The dashboard's net draws the hierarchy as layers — the throne, then
  each rank in the order work flows down, the minds beside — every neuron a rank-coloured tile
  whose 28 cells are its last four weeks of work (runs and tokens from the journals; blank where no
  instrument saw any), every tree bond weighted by the commissions it carried. Hovering or focusing
  a neuron lights its path to the throne and its subtree; arrow keys walk the tree; the side card
  reads its doc, territory, minds, bonds and work. A Court's dispatch travels down the tree as a
  spark, its report comes back up (green when DONE, amber or red otherwise), and a running body
  breathes. A **3D** stage (hand-written WebGL2, no library) stands the same tree as glowing planes
  with particles drifting down the bonds; drag orbits, the wheel or a pinch zooms, a click focuses.
  It animates only while seen, honours reduced motion, follows the theme's tokens, and falls back
  to 2D without WebGL2. Narrow screens get the tree as an outline.

- **Select and copy in the TUI's panes.** A drag in the chat pane selects (clamped to the chat,
  scrolling at its edges; a double click takes a word, a triple a line) and the release copies
  the plain text — OSC 52, the primary selection, and off SSH the desktop's `wl-copy`/`xclip`/
  `xsel`/`pbcopy` — with a toast. `ctrl+y` and `/copy` copy the last answer's markdown; `/mouse`
  hands the mouse to the terminal (a `mouse: terminal` footer badge) and back.

- **The law inside Claude Code.** `isekai install --claude` wires `isekai hook` into Claude Code's
  hooks: the crest and the log's last entries at every session start (and after `/clear` and
  compaction), recall on every prompt, the guard and the human gate before every tool (it only
  tightens), and the gate when a turn ends — a territory changed without its doc sends the turn
  back with the reasons, and the verdict lands in `log.md`. Claude Code runs the model on your plan;
  the hooks never call one and need no key.

- **The TUI rail scrolls and folds, and the footer counts its tokens.** The side rail gains
  its own scroll offset — the wheel over the rail column walks it clamped to the content,
  with ▲▼ edge markers when more rides off-screen (f2ce295) — and the plan/courts/files rail
  cards collapse to a one-line `+<glyph> <title> (<count>)` header via alt+p/c/f or a click
  on the header border, the session card exempt; all three bindings join the Shortcuts
  overlay (0fdfb9f). The footer reads `ctx ▮▮▮▯ 42% · 84k/200k`: FooterView gains
  `CtxTokens`/`CtxLimit`, the humanized context token counts gated at width ≥ 88 while the
  meter keeps its ≥ 72 gate, and zero values render byte-identical to before (cbb181d);
  `app/tui.go` hoists `readingOf` out of the footer closure and fills both fields from
  `r.Tokens`/`r.Limit` (4d08c4c).

- **The web answers gates in every run mode (webBus.answerable from birth).** The browser
  dashboard now mints a choice id for every gate ask whatever the front-end — dash-only,
  TUI with the board up, or the container — so approvals answer from the web everywhere.
  First answer wins across surfaces: whichever of the terminal or the browser answers first
  lands, and the loser gets the existing "answered elsewhere" refusal instead of a hang.

- **Open plans block a clean close (plan conformance, the closing check's second leg).** The loop
  gains `PlanOpen`, wired to the body's deck: the closing check's re-ask names the open item, and a
  body that closes anyway gets the hole on the report — `@? closed with an open plan: …`.

- **The closing check (`law.closingCheck`).** When enabled, a body's run that closes with no tool
  call is sent back once ("is the ask applied end-to-end?") before it may close — the bench-class
  failure it catches is a mid-plan close ("let me merge…" with the merge never run). Off by
  default; the bench config enables it (`bench/configs/gateway.yaml`).

- **The forge (`@FORGE` + `/forge`).** A run that proves a repeating need may carry its draft on
  the report's wire — `@FORGE tool:<name> | <summary> | <argv>` (argv holes `{{param}}` declare
  as string params) or `@FORGE skill:<name> | <text>`. Drafts ride up unminted and land as
  shared notes tagged `forge` on the record beat; the desk lists them (`/forge`) and mints by
  index (`/forge mint <n>`) — a skill lands as a world note, a tool lands as a `tools.custom`
  patch that reloads live. Minted drafts mark themselves (`forge-minted`); nothing mints twice,
  and an existing `tools.custom` name is never overwritten.

- **The container's trust (law.humanGate.trustContainer).** Inside the harness's own runtime
  container — `<PREFIX>CONTAINERED=1`, set only by the harness's own container argv, spoofable by
  no body — every action class reads as pre-approved at the Gate: the sandbox is the membrane,
  so outward and destructive stop asking. The guard denylist still refuses first; the world dir
  mounts read-write inside the sandbox or out.

- **The throne stays free (tools.dispatch.sessionBackground).** With `tools.dispatch.background`
  on, the session's own dispatches default to background — rimuru keeps talking while Courts
  ride on; each report wakes the session. A commission keeps its word: `background: false`
  still rides the foreground. One layer down, a Court's own dispatches stay foreground by canon.

- **The bodies pane pipes (ctrl+t → i).** A body selected in the pane takes a typed line into
  its inbox — it lands at the body's next turn boundary (the FIFO contract, never mid-turn) and
  a done body refuses: its context died with its task. esc cancels a pipe in the making.

- **The council (law.stress).** Governance strain is now counted and answered: the loop's
  turn-end surface carries `Strains` (failed acts standing + gate re-asks + wire re-asks), and
  the dispatcher keeps a strain ledger per body. A Court that never lands — loop Fail or
  Checkpoint, or its own wire verdict reads FAIL/ESCALATE — accrues strain, and past
  `law.stress.escalateAt` (default 5) its report carries a **council warrant** the dispatcher
  answers as its manager: `@INSTRUCT` folds context into the re-commissioned ask, `@SPLIT`
  narrows to one commission per slice, `@ESCALATE` rides the hole up — sideways counsel is
  named lawful in the warrant's text. The warrants a body carries are bounded by
  `law.stress.maxConsults` (default 3); past the bound the report hardens to `@S ESCALATE` and
  rises to the human. A clean landing clears the ledger. The body's own wire verdict drives —
  a court that lands its run cleanly but says `@S FAIL` is counted as failing (the Law-9
  reading, coded end-to-end now: the journal's status is lifecycle, the body's @S is verdict).

- **The model resolver's pin law.** A pin set in config now FORCES: `config.ResolveModel`
  resolves creature → rank → office → task → default (was: office ahead of rank), so
  `models.ranks.<race>` or `ranks.<race>.model` outrules the office hat the commission
  named. Only an unpinned body takes the lens of the ask — and the binary's own tasks
  (drain/gate/log/bench) sit below offices, routable per task. Doubly-locked by
  `TestModelsResolution`'s collision row: a pinned race rides its pin even on a differently
  lensed commission.

- **Nested dispatch steps land their Courts, coded.** A dispatch step's start seeds the
  Court card's ask and its end lands the card with the wire report at EVERY depth — a Court
  dispatched by a Court is no longer invisible to the card flow: the host sees the child's
  steps only as its parent's, so `tuiHost.observe` carries the seed/land handling at the top,
  before the session-only branch. The parent's own log lines ride along unchanged.

- **Court-card exit-verdict law (Leg 2).** A court card's face now prints the run journal's
  exit verdict — `FAIL · 2m35s` under a failed journal — never the bare lifecycle word `done`.
  `LiveBody.Exit` is populated at `Court.Finish` from the wire report's first `@S` token; the
  dispatch tool re-stamps `rep.Status` as the payload's first `@S` line so a renderer that
  reads only the returned string can never show a collapsed "done" over a FAILed journal. The
  lifecycle word moves to the ask-line tooltip. (`nil` derailed the 2026-09-28 session: Law 9.)

### Security

- **The gate runs no command the model wrote this turn.** A verify line added to a creature doc in
  the same turn is a hole for the gate holder, not a reading; every verify line and injected check
  runs under the sandbox (no network, secrets scrubbed, its process group killed at the timeout).
- **A command is its name however it is called.** `/usr/bin/git push`, `/bin/rm -rf`, `\rm`,
  `xargs /bin/rm` and `sudo -u root env -i /bin/rm` classify and hit the guard like `git push`
  and `rm`; `>|` into a system path is outward.
- **An allow never carries a chained act.** `bash:npm test*` allows `npm test`, not
  `npm test && curl …`, `npm test & curl …` or `npm test $(curl …)`: each outward or destructive
  part needs its own allow. An allow in the tracked project config never pre-approves an outward
  or destructive act — put it in `config.local.yaml` or the global config.
- **`/dev/tcp` is the network, system directories are read-only.** A redirect or write into
  `/dev/tcp|udp`, `/etc`, `/usr`, `/opt`… is outward; `/library` is no longer read as `/lib`.
- **The tracked project config may not loosen the human gate.** `law.humanGate.enabled: false` and
  `approve: [outward|destructive]` are refused from `.isekai/config.yaml` (they travel with every
  clone) and honored from `config.local.yaml` or the global config; every standing approval is in
  the off-list.
- **More secrets scrubbed**: names holding `PRIVATE_KEY`, `ACCESS_KEY`, `CREDENTIAL` (so
  `AWS_SECRET_ACCESS_KEY`), and URLs carrying a password (`DATABASE_URL`).

### Changed

- **Ascension happens.** The high elf and the high orc ascend by default on a long, clean record —
  `reports >= 20 && passRate >= 90% && holeRate <= 10%` (`ranks.high-elf.ascendsWhen`,
  `ranks.high-orc.ascendsWhen`; `""` turns it off). An ascended rank lives in its own dir and
  prefix (`.isekai/high-orc/<name>/`, `high-orc-<name>`; agent-one `principal-domain-`), so the
  mint no longer collides with its base. The body rises whole: its doc, territory, traits and
  verify are carried into the new seat, the base's seat retires, and the creatures that reported to
  it now report to its ascendant, which commands them. Each ascension is appended to `log.md` with
  the metrics that earned it and announced on the dispatcher's report; a failed mint is logged and
  reaches the dispatcher as `@?`. The tally is an append-only ledger
  (`.isekai/instruments/ascension.jsonl`) replayed at start, so a record counts across sessions,
  and a body's own `@S FAIL` is never counted as a pass.
- **Dispatch follows the line.** An orc orders only its own slimes, an elf only its own orcs —
  the creatures whose `Reports to` chain reaches it; a commission to another line's creature is
  refused naming the parent that commissions it. `tools.dispatch.lineage: false` restores the
  rank-only law.
- **The kijin holds no dispatch.** A rank no other rank reports to (and not sideways) is never
  handed the dispatch tool, whatever its tools list says.
- **A rank's job reaches its body.** The `job` line rides the creature card in the system prompt.
  `ranks.<r>.body` is gone from the rank table — every body runs as a Court — and a configured
  value is reported as an ignored key.

- **The law reads the same on every host.** `ISEKAI.md` names instruments (the context reading,
  memory, toolbox, the loop) and maps them once to each host — the binary's commands, or the JS
  tools and `/don`·`/mint` of a Claude Code or OpenCode world. Three contradictions resolved: the
  desk lives in the worn Mind, or the creature's own doc when it wore none; the call to the
  configured model is Containment's one standing outward act; Law 2 names the dimension's
  Elf ⇄ Elf channel. The board's drawing rules moved to `canon/binary.md`; memory tiers, the loop
  and the toolbox keep their law here and point to their canon for the spec.
- **`init --level complex` checks the model first**: with none reachable it founds the light world,
  names the missing key and the rerun, and does not start a founding session it cannot finish.

### Fixed

- **The orc's verify check reads the slime's doc fresh.** A verify line the orc wrote earlier in
  the same turn now counts when it dispatches the slime.

- **The dashboard's folder tree loads again.** It read the root level's entries under `nodes`
  while `/api/tree/children` answers `children`, so the tree always said "no tree source".

- **The neural net draws the throne and its tree.** `/api/colony` always carries the throne node
  (the session body, under the lexicon's word for it) and names every creature's `parent`; a
  creature with no bond up to a body on disk hangs under the throne by an `implied` reports edge
  and a note, so the tree is connected and the gap is said. The ontology reads the world's own
  `crafted-skills/` and `base-skills/` as minds, so a creature's Minds field becomes `wears` bonds.
  `/api/activity` serves each body's last 28 days (runs from the loop journals, tokens from the
  usage journal) — what the net's activity tiles and bond weights read.

- **The TUI's input and keys.** A capital `C` (or `c` on a pinned stale deck) no longer vanishes
  from a typed word: the plan snap moved to `alt+n`. The input box grows with a long line's wraps,
  not only its newlines. `ctrl+c` clears a draft first and interrupts a busy turn without arming
  the exit; `esc` and `ctrl+c` keep a cleared draft one ↑ away. Under the panes `ctrl+u`/`ctrl+d`
  are the input's again — `pgup`/`pgdn` (now matched), `shift+↑/↓` and the wheel scroll the chat —
  and the shortcuts list says so. The bodies view's pipe gives the draft back.
- **The TUI stays smooth on a long session.** Each block renders once per width and theme
  (cached; a new block costs its own render, not the transcript's), the chat viewport is fed only
  when the transcript or the live rows changed, and the rail's plan decks load in a `tea.Cmd`,
  read again only when a deck file's size or mtime (or a live body's office) moved — no disk
  read on the UI goroutine five times a second.
- **The TUI's transcript.** Quitting in pane mode prints the conversation to the scrollback, so it
  stays in the terminal after exit. A block that finished under the board prints once, not twice.
  `<` and `>` in an answer survive Glamour (`Vec<T>`, `providers.<name>.script`); long errors wrap.

- **`--help`, `-h` and `--version` answer at once**, outside a world too, and `<command> --help`
  shows that command; an unknown command is named unknown before any world lookup.
- **Hints name the binary's own commands** (`isekai memory index`), not `memory.js`, in a world
  that does not ship the JS instruments.
- **The setup form no longer offers to keep a global config that does not exist.**
- **The README's Quick start sets the key before `init --level complex`**; Chromium is listed for
  `ui check`; `install.sh` no longer advertises piping into `sh`.
- **The canon describes this repository**: its index, the toolchain (Go ≥ 1.26), one `cmd/`, and
  `init` as what creates a world.
- **Anthropic streams.** A tool with no arguments arrives as `{}`, not as a malformed call;
  redacted thinking keeps its data and replays as it came; a stream cut before `message_stop` is
  an error, never a partial answer taken for a whole one.
- **429 and 529 are retried on Anthropic too** — one retry policy for both clients (Retry-After,
  else 1s/2s/4s, three times, never past the context).
- **A malformed native tool call is named to the model** ("the arguments to read are not valid
  JSON") instead of running the tool on an input it misreads.
- **A turn that stops early keeps its writes** — the ones a read-classified command made too — in
  the transcript the resume and the gate read.
- **The write watch skips `node_modules`, `.venv` and `__pycache__`**: no more walking a
  dependency tree after every write.
- **A deadline is a deadline.** Gate readings, hooks, rule checks, `goal --validate` and custom
  tools kill their whole process group at the timeout and stop waiting on pipes a backgrounded
  child kept open.
- **The zombie Court card.** A Court that closed "done" without its wire report ever reaching
  the model (dispatched by a Court, a drained `@%`, an unparseable or failed report) kept its
  card in the live area forever, the clock recomputed on every frame — an orc card read
  `DONE · 15m0s` when its own budget clock had axed it. Coded at the seams, never prompted:
  `Render` stops recomputing a done Court's elapsed; the turn's end sweeps every
  done-but-unlanded card into exactly one tombstone; and a Court event closes the body pane's
  state too, so no pane outlives its card reading "thinking".

- `TestREPLLive`, `world.TestDispatch`, and `tui.TestBlocksGolden` walk again — the goldens
  now carry the new face stamps; the dispatch asserts expect the stamped `@S DONE`.

- **Per-section local config.** The local layer mirrors one-file-per-section: beside (or
  instead of) `.isekai/config.local.yaml`, a `<section>.local.yaml` (`.yml`/`.json`) holds one
  top-level section's value, loads after `config.local.yaml`, wins there, and keeps its own
  file for `config explain`; two local files for one section are a load error, mirroring the
  section-file rule. Machine-local by canon: `*.local.yaml` belongs in the world's gitignore.
- **`isekai config scaffold [--check]`** — the local-layer mint: writes (never overwrites)
  `config.local.yaml` and one commented, default-valued `<section>.local.yaml` per section
  into the world dir, appends `*.local.yaml` to the world's `.gitignore`; `--check` writes
  nothing and answers on the wire what drifted: unknown local sections/keys as `@?`, defaults
  sections with no local file yet as `@F`. File names and keys come from the running binary's
  own defaults, so re-running the command syncs the scaffold when canon moves.

- **The plan deck is native.** A new world tool, `plan`: every planner-level body (rimuru, elf,
  orc, kijin, ascended) keeps a todo deck under the plan dir — one line per slice with the rank or
  zone it belongs to (`who`), exactly one `in_progress`; `tick` moves a line, `read` reads another
  body's deck, all machine-local under tmp. The elf's rank text now codes the chain: write the
  deck, map lines to owning domains, dispatch — orcs in parallel (`background: true`) when the
  slices are independent; the orc's text codes the verify-first: the slice's checks are written
  before its slime is ordered.

## [0.3.0] - 2026-09-28

### Highlights

**The world versions and defends its own seat.** A world root bearing `.isekai/dimension.yaml`
outside any work-tree is versioned `git init -b main` once at open — a world is born under git, by
the harness itself. And the gate now attributes the throne's own seat files (the dimension manifest
and the law doc) to Rimuru: before, a gate-holder whose territory reached `.isekai/` silently owned
them through Owner's fallback, so Rimuru's lawful relocation of the manifest failed right-author and
Vitality — demanding an orc's README for the throne's own file. The exemption mirrors the creature
doc rule and polices non-throne bodies exactly as before.

### Added

- **Web UIs from one token system.** A built-in `ui` Mind (method + design craft, offline), and
  `isekai ui init|scan|check`: a foundation (tokens in three tiers, 13 OKLCH palettes checked for
  WCAG AA, layout primitives, cascade layers, two example components), a derived legend under
  `.isekai/ui-assets/` (manifest, legend.md, a live catalogue), lints (tokens only, declared
  classes, defined tokens, true headers) and screenshots at 360/768/1280 in both themes with
  sideways-scroll detection. The gate runs the lints on a turn that touched the ui dir
  (`law.gate.ui`, on by default). The container image ships Chromium.
- **The dashboard** at `/dash` on the board, and `isekai dash` to run a session in the browser: a
  console like the CLI (streamed answers, thinking, tool and Court cards with diffs, approvals),
  the world as a live neural net with knowledge cards, metrics, and the relations matrix with its
  dimensions — built on the ui system and held to `ui check`. Acts carry a per-run token and a
  loopback/same-origin check.

- **Three ways to found a world**: `init --level light|medium|complex` (a terminal asks). Medium
  sketches a colony from the tree with no model; complex runs a founding session that reads all the
  code and hires the team — elves, orcs, slimes, a kijin, minds and bodies, relations, verify lines
  that pass — until `onto check` is clean.
- **Dimension self-init.** `isekai/world/dimension.go`: `LoadDimension` parses
  `.isekai/dimension.yaml`, and `GitSelfInit` versions a manifest-bearing root outside any
  work-tree (`git init -b main`, once, at open) — malformed manifests are a note, never a failed
  open; the harness never commits nor pushes.

### Fixed

- **Throne-seat attribution at the gate.** The gate's right-author and Vitality checks exempt the
  throne's seat files (`.isekai/dimension.yaml`, `.isekai/isekai.md`) when the throne authors them —
  Owner's gate-holder fallback no longer lets an orc whose territory is `.isekai/` claim the throne's
  own files. A non-throne body writing them still fails right-author.

### Removed

- The SSH board (`board --ssh`): the board is local only — 127.0.0.1, or the host's loopback from
  `--containered`.

## [0.2.0] - 2026-09-27

### Highlights

**A terminal in the class of Claude Code**, on Charm v2: the mascot drops in; the model's thinking
streams and folds; each rank has a face and its own verbs, and the orc weighs the verdict at the gate;
tool calls are class-coloured cards and diffs keep the code's colours; `ctrl+t` watches every body,
live; `/board` full screen with the reasoned ontology as a graph; a `ctrl+k` palette and toasts; the
window title and tab progress follow the session; the transcript reprints on resize.

**Safety in code**: a global dangerous-command guard refused before any gate (151-case corpus;
`guard install` wires it into Claude Code and OpenCode); the gate runs the pre-turn verify lines, keeps
the tests intact, and sends a failed gate back to the model once.

**New commands**: `goal`, `review`, `handoff`, `board --ssh`, `--containered`, a setup form at `init`;
config one file per section (`guards.yaml`, `rules.yaml`…).

**Fixes**: an upstream `finish_reason: "error"` is a failed (retryable) call, not an answer; a TUI that
cannot start says why; a piped slash command's turn can no longer be skipped by the next line.

## [0.1.5] - 2026-09-27

### Highlights

Works with Google Gemini's OpenAI-compatible API, free tier included: each tool call's `thought_signature` travels back on the next turn (multi-step tool use failed before), 429 and 5xx answers are retried with backoff (honouring `Retry-After`), and Gemini's array-form errors are shown instead of "unreadable body". The session answers you in plain language; only dispatched Court Bodies answer on the wire.

## [0.1.4] - 2026-09-27

### Highlights

`run "<ask>"` no longer reads stdin — an inherited pipe that never closes (CI, another agent, cron) could hang it; stdin is read only for `run` with no ask or `run -`. Background jobs killed by a timeout no longer print a stray "Killed" line into the next command's output.

## [0.1.3] - 2026-09-27

### Highlights

The board's off-list now names switched-off tools (it said "everything on" while a tool was off), and long origins wrap inside their card.

## [0.1.2] - 2026-09-27

### Highlights

A patch release for two defects found by running the released binary in a plain world.

### Fixed
- A machine-wide agent file named after the throne (e.g. OpenCode's `~/.config/opencode/agents/rimuru.md`) was imported as a creature with rank `elf`, so the session ran with the elf's office, model route and tool shelf. The session's name is now reserved.
- The toolbox's level-2 hint named `node .isekai/tools/toolbox.js`, which releases do not ship; it names `isekai toolbox` unless the world ships the JS instrument.

## [0.1.1] - 2026-09-27

### Highlights

A patch release: the default model is `anthropic/claude-opus-5` (0.1.0 shipped an outdated default),
and the live REPL no longer races the session's context reading while a turn perceives it.

### Fixed
- Default model `anthropic/claude-opus-5`.
- A data race between the status line / board and the running turn's context reading.
- The shared Harbor smoke task uses neutral wording.

## [0.1.0] - 2026-09-27

### Highlights

First release: the isekai convention as a Go binary. The law is enforced by the harness — the end-of-turn gate with its verdict in `log.md`, the human gate with per-command permission rules, a `bwrap`-sandboxed persistent shell, and native memory with compaction by drain. Ranks and Court dispatch are native, each office can run its own model, and every feature is switchable in YAML or JSON with switched-off law always reported. Ships the board (eight responsive pages), MCP (stdio and HTTP), Claude Code / OpenCode compatibility for instructions, skills, commands and agents, and a Harbor benchmark adapter.
