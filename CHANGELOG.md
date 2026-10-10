# Changelog

All notable changes to Isekai are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow
[Semantic Versioning](https://semver.org).

## [Unreleased]

### Added

- **The law inside Claude Code.** `isekai install --claude` wires `isekai hook` into Claude Code's
  hooks: the crest and the log's last entries at every session start (and after `/clear` and
  compaction), recall on every prompt, the guard and the human gate before every tool (it only
  tightens), and the gate when a turn ends — a territory changed without its doc sends the turn
  back with the reasons, and the verdict lands in `log.md`. Claude Code runs the model on your plan;
  the hooks never call one and need no key.

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

### Removed

- The SSH board (`board --ssh`): the board is local only — 127.0.0.1, or the host's loopback from
  `--containered`.

### Security

- **The gate runs no command the model wrote this turn.** A verify line added to a creature doc in
  the same turn is a hole for the gate holder, not a reading; every verify line and injected check
  runs under the sandbox (no network, secrets scrubbed, its process group killed at the timeout).
- **A command is its name however it is called.** `/usr/bin/git push`, `/bin/rm -rf`, `\rm` and
  `xargs /bin/rm` classify and hit the guard like `git push` and `rm`.
- **An allow never carries a chained act.** `bash:npm test*` allows `npm test`, not
  `npm test && curl …` or `npm test $(curl …)`: each outward or destructive part needs its own allow.
- **`/dev/tcp` is the network, system directories are read-only.** A redirect or write into
  `/dev/tcp|udp`, `/etc`, `/usr`, `/opt`… is outward; `/library` is no longer read as `/lib`.
- **The tracked project config may not loosen the human gate.** `law.humanGate.enabled: false` and
  `approve: [outward|destructive]` are refused from `.isekai/config.yaml` (they travel with every
  clone) and honored from `config.local.yaml` or the global config; every standing approval is in
  the off-list.
- **More secrets scrubbed**: names holding `PRIVATE_KEY`, `ACCESS_KEY`, `CREDENTIAL` (so
  `AWS_SECRET_ACCESS_KEY`), and URLs carrying a password (`DATABASE_URL`).

### Changed

- **The law reads the same on every host.** `ISEKAI.md` names instruments (the context reading,
  memory, toolbox, the loop) and maps them once to each host — the binary's commands, or the JS
  tools and `/don`·`/mint` of a Claude Code or OpenCode world. Three contradictions resolved: the
  desk lives in the worn Mind, or the creature's own doc when it wore none; the call to the
  configured model is Containment's one standing outward act; Law 2 names the dimension's
  Elf ⇄ Elf channel. The board's drawing rules moved to `canon/binary.md`; memory tiers, the loop
  and the toolbox keep their law here and point to their canon for the spec.
- **`init` talks to a person in prose** on a terminal — what was founded and the next step (the
  model to set, the session to open); the wire lines stay for pipes and `--format`.
- **`init --level complex` checks the model first**: with none reachable it founds the light world,
  names the missing key and the rerun, and does not start a founding session it cannot finish.

### Fixed

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
