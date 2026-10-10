# The isekai binary — the law as a harness

`isekai` is a single Go binary that runs an agent session on a world. The convention was written
for harnesses that only *read* it (Claude Code, OpenCode); a model can read a law and still skip
it. Here the law is the harness: the gate, Vitality, the log, the human gate, the instruments and
the wire are code paths the model cannot route around, and every creature is a native type.

Source: this repository (Go module `github.com/Kaginari/isekai`; the standard library plus the
Charm libraries and their companions — chroma for highlighting, harmonica, colorprofile — for the
terminal UI, `tui.md`). Build with `go build -o bin/isekai ./cmd/isekai`. Toolchain: the `go`
line of `go.mod` (Go ≥ 1.26).

## Principles of the build

- **Minimal.** Standard library only — except the terminal UI, built on the Charm libraries
  (`canon/tui.md`). No LSP, no plugin system. A feature enters when a session names the need twice
  (Nature 4).
- **Files are truth.** The binary reads and writes the same files the JS instruments do
  (`log.md`, `memory/`, `toolbox/`, `instruments/`) in the same formats. It never keeps a private
  store the world cannot read without it.
- **The law is enforced, not recited.** Every rule the binary can check, it checks in code; what
  it cannot check it puts in front of the model at the moment it applies, not once at start.
- **Containment by construction.** Nothing reaches outside the world except the model provider
  call itself and actions that passed the human gate.

## Packages

| Package | Holds | Law it carries |
|---|---|---|
| `provider` | one `Provider` interface (messages + tool calls + usage); `anthropic` (Messages API), `openai` (chat-completions — also OpenRouter, Ollama, any compatible endpoint), `mock` (scripted, for tests) | the provider call is the one standing outward act |
| `tool` | `Tool{Name, Description, Schema, Class, Run}`; built-ins `read`, `write`, `edit`, `bash`, `glob`, `grep`, `dispatch` | every tool declares a class; `bash` is classified per command (port of `loop.js`'s classifier) — a declared class only tightens |
| `gate` | the human gate: the TUI's choice block or the TTY prompt (`Gate.Answer` is the seam), `--approve <class>` pre-approval, `--dry-run` | Nature 7, Law 6: outward and destructive always ask; nothing auto-approved; a denial stops the turn |
| `loop` | the turn engine: perceive → recall → plan → act → verify → record per tool step; journal to `.isekai/instruments/loop/<run-id>.jsonl` | §The loop — budgets are readings; past one, checkpoint and stop honestly |
| `instrument` | context occupancy from provider-reported usage, 200k budget / 180k stress zone; `status` board | Nature 9, §Instruments — Rimuru's stress is a reading |
| `memory` | port of `memory.js`: status, index, recall (meaning + relation rank), remember (`--kind law|colony|territory`) | §Memory tiers — same files, same formats |
| `toolbox` | port of `toolbox.js`: registry build, `brief` (level-1 `@T` manifest under a budget), `load` (level 2, journaled) | §Minds & Bodies — the toolbox |
| `wire` | parse and emit the envelope (`@S @F @V @? @U @FORGE @E`, `@ROOT @SCOPE @ASK @CAP @DUMP`); `@CAP` enforced | Absolute Rule II, Nature 8 |
| `world` | world discovery, the law loader (crest always; code sections on demand), ranks and bodies from `.isekai/{elf,orc,slime}/`, `log.md` append-only writer, the Orc gate and the Vitality check, Court dispatch | Natures 1–4, the gate, Laws 2–4 |
| `onto` | the ontology: schema, graph, reasoning, validation, projection | §The ontology below |
| `config` | the switchboard (`canon/config.md`): layers, origins, refusals, permission rules, models per office/rank/task, ranks as data | the honesty rule; a disabled law is a finding |
| `compact` | the drain: pointerize, trim, the unsaid, the desk, the episode, verify | §Compaction below |
| `sandbox` · `shell` | bwrap (Linux) / Seatbelt (macOS) containment and env scrubbing; one persistent bash per body with background jobs | Nature 7, §Bash below |
| `mcp` · `discover` | the MCP client (stdio, streamable HTTP); what other harnesses wrote (instructions, Minds, commands, Bodies, MCP imports) | §MCP below; `harness-parity.md` |
| `board` | the world, seen: an HTTP handler over the same instruments | §The board below |
| `app` | the one engine behind both binaries: config → world → providers, shelf, MCP, discoveries, hooks; sessions (JSONL), the live session (the TUI on a terminal, the line REPL on a pipe or `--plain`), `run`, `bench`, `selftest`, `init`, `board` | everything above, wired |
| `tui` | the terminal UI (`canon/tui.md`): the view model (blocks from the loop's events), the Bubble Tea program, the choice block the gate and the `ask` tool answer through | the one siphon that never narrows — the human's |
| `cmd/isekai` | the distribution's `main`: the name it hands `app.Main` decides the lexicon, world dir, env prefix and law file (agent-one is the same engine under another name, built from its own repository) | — |

## Bash — the model's shell, the world's rules

- **Anthropic-defined tools on Anthropic models.** With an `anthropic` provider the binary declares
  Claude's own trained tools — `{"type":"bash_20250124","name":"bash"}` and
  `{"type":"text_editor_20250728","name":"str_replace_based_edit_tool"}` — schema-less, so the model
  drives the shell and editor it was trained on. With an `openai` provider (OpenRouter, Ollama,
  vLLM) the same executors sit behind custom-schema `bash` / `edit` tools. The model chooses the
  words; the binary always executes.
- **One persistent session per body.** The shell keeps its cwd and exported variables between
  calls; `{"restart": true}` resets it. A Court Body gets its own session and loses it with its task.
- **Sandboxed by default.** The fence is the platform's, chosen by the OS and probed once
  (`sandbox.probeFor`). On Linux each command runs under `bwrap`: the filesystem is read-only except
  the world root and a private `/tmp`, no network unless the command passed the human gate as
  `outward`, its own pid/ipc/uts namespaces (the host's processes are out of sight), and the home's
  credential paths (`sandbox.secretHomes`: `~/.ssh`, `~/.gnupg`, `~/.aws`, `~/.kube`, `~/.docker`,
  `~/.netrc`, `~/.config/gcloud`, `~/.config/gh`, keyrings, browser profiles …) masked — an empty
  tmpfs over a directory, `/dev/null` over a file. On macOS it runs under Seatbelt:
  `/usr/bin/sandbox-exec -p <profile> -DKEY=path… -- argv` (`sandbox/seatbelt.go`), every path a
  `-D` parameter read by `(param "KEY")`, never spliced into the SBPL. The profile is
  `(allow default)` then fences: `(deny file-write*)` but the world root, the extra read-write
  paths, `/private/tmp`, `/private/var/folders` and the terminal/null devices (symlinks resolved —
  Seatbelt matches real paths); `(deny network*)` but loopback and unix sockets, mDNSResponder
  closed, unless outward; `process-info*` and `signal` only within the sandbox (no pid namespace:
  other processes stay listed); Apple Events, LaunchServices and the pasteboard denied; and, last
  so nothing re-allows them, reads and writes of `secretHomes` plus `macSecretHomes` (keychains,
  cookies, browser profiles), each as written and as resolved. Allow-default, not deny-default:
  isekai runs whatever the world builds with, and a deny-default list breaks toolchains it cannot
  foresee; mach IPC stays open, so it fences mistakes, not a hostile binary (`--docker-agent` is
  that boundary). The probe runs `/usr/bin/true` under the very profile, so a macOS that refuses it
  reads `sandbox-exec refused: …`. No `bwrap`, a kernel that refuses user namespaces, a refused
  profile or another OS → `sandbox: none`, reported as `@?` in `status`.
- **The whole binary in a container — `--docker-agent`.** The binary re-runs itself under Docker
  (`docker run --rm --init`): the world mounted read-write at its own path, this binary and what a
  session reads from the host mounted read-only (`~/.config/<dist>`, `~/.<dist>`, the Claude and
  OpenCode instruction/skill/command/agent dirs, git's identity — only those that exist), the
  session store `~/.local/share/<dist>` read-write (sessions must save), `$HOME` otherwise an empty
  tmpfs (no `~/.ssh`, no other secrets), the caller's uid, the host network (the model API and the
  board). Keys cross by name (`-e NAME`), never by value: the terminal's variables, the dist's own
  `<PREFIX>*`, every `*_API_KEY` / `*_BASE_URL`. The image is built on first use from the Dockerfile
  embedded in the binary (debian slim + bash, git, ripgrep, python3, jq, make, curl…), tagged by the
  Dockerfile's hash so a changed runtime is rebuilt; `--image <ref>` runs a richer one. Inside,
  `<PREFIX>DOCKER_AGENT=1` makes the flag a no-op, and `status` says `container: docker <image>` in
  place of the sandbox line: the container is the boundary, bwrap does not nest in it.
- **The guard, before everything.** A denylist of the catastrophic and irreversible — disk wipes,
  `rm -rf` of `/`, home or a system dir (by `~`, `$HOME` or the home's own path), force-pushes and
  remote deletions, repository and secret deletion, history purges, secret-store reads, a script
  from the network piped into a shell — refuses a bash command before the policy and the gate: no
  approval runs it. The built-in list (`guard/patterns.txt`) is always on; the machine-wide
  `~/.agents/hooks/dangerous-patterns.txt` and `guard.files` add to it. `guard test` runs the
  corpus (151 commands, from davidondrej/skills, MIT) — every block blocked, every allow allowed —
  and the classifier is held to the same corpus: every command the guard blocks must reach the
  human (outward or destructive) even with the guard off. `guard install` wires the same list into
  Claude Code (a PreToolUse hook) and OpenCode (a plugin), showing the change first — one denylist
  for every agent on the machine. It stops accidents, not a determined agent (`python -c` slips past
  any regex); the sandbox and the gate stay the containment.
- **Classified before it runs.** The classifier settles the class; permission rules and the human
  gate decide; the model's own claim only tightens. `git` is read through its global options
  (`git -C . push`, `git -c k=v push`, `git --no-pager push` are the verb's class). A path is
  read for where it points: a symlink out of the world, and a file yet to be created under
  one, are outward for `read`, `write`, `edit` and `patch` alike; the editor refuses them.
- **Secrets never cross.** The command's environment drops every name holding `API_KEY`, `TOKEN`,
  `SECRET`, `PASSWORD`, `PRIVATE_KEY`, `ACCESS_KEY` or `CREDENTIAL` as a whole word, every value that
  is a URL with a password in it, and every provider's `apiKeyEnv`; config holds an allow-list.
- **Nothing outlives its call unannounced.** Each command runs in its own process group; a timeout
  kills the group. `background: true` starts a named, logged background job with its own output
  file; jobs are listed in `status` and killed when the session ends.

## The default toolset — everything on, config takes away

The binary ships the tools a daily-driver harness needs, all enabled by default (profile `max`);
config disables, never has to discover:

| Tool | Class | Does |
|---|---|---|
| `bash` | classified per command | persistent sandboxed shell; `background` jobs |
| `read` · `ls` · `glob` · `grep` | read | files, directories, patterns, content search |
| `write` · `edit` · `multiedit` · `patch` | write | create, `str_replace`, several replaces atomically, apply a unified diff |
| `git` | classified per subcommand | `status`/`diff`/`log`/`show` read; `commit` write; `push` outward; `reset --hard` destructive |
| `webfetch` · `websearch` | outward | fetch a URL as text; search via a configured backend |
| `ask` | read | ask the human a question with options, answer returned to the model |
| `dispatch` | per rank | a Court Body (§Ranks and bodies) |
| `recall` · `remember` | read · write | the memory tiers; `remember` refuses the forge's tags (`forge`, `forge-minted`, `forge-rejected`, `skill`) — only the harness files drafts |
| `toolbox` | read | search the registry, load a skill/command/tool body (level 2) |
| `onto` | read | query the creature graph: who owns a path, what a creature may see |
| `skill` | read | don a Mind by name, from the one registry: `crafted-skills` > `base-skills` > the host's project dirs > its global dirs > the binary's own |
| `desk` · `plan` | write | the body's working memory (≤5 thoughts) · its plan deck — the todo list written before dispatching, one line per slice with its owner |
| MCP tools | declared by server, classified like any other | every configured MCP server |

Custom tools from config (`tools.custom`) join the same table on the same terms.

**The forge.** A `@FORGE` draft is filed by the harness from the report it holds — the
dispatcher's own result, before `@CAP` cuts it or the courier envelope folds it to a digest — as
a shared note tagged `forge` carrying the drafting body, its run and a content hash; a hash
already in the ledger is not filed again, so a relayed line lands once, as its drafter's. The
throne's own final answer, the salvage turn of a clock-killed Court and the Claude Code Stop
hook's final text file their drafts the same way. The desk (`/forge`, `isekai forge`) lists the
open drafts by hash and mints by the human's word: a skill writes
`<world>/crafted-skills/<name>/SKILL.md` (Claude Code's frontmatter plus `version`, `status`,
`by`, `born`, `minted`; never an overwrite — `--version` lands v(n+1) and keeps the old file
under `versions/`); a tool writes `tools.custom.<name>` at the classifier's class with
`crafted:` provenance. Marks (`forge-minted`, `forge-rejected`) are append-only notes naming
the hash.

## MCP

The binary is an MCP client (JSON-RPC 2.0, stdlib): **stdio** and **streamable HTTP** transports;
OAuth for remote servers comes later — until then a remote server authenticates by header from
an env var.
- **Servers** come from `mcp.servers` in config, and — each behind its own `mcp.import` switch —
  from Claude Code's `.mcp.json` and OpenCode's `opencode.json(c)` `mcp` block, read as-is.
- **Tools** are named `mcp__<server>__<tool>` and enter the toolbox registry as level-1 entries:
  the model sees names and one-line descriptions under the budget, never fifty full schemas; a
  schema is sent to the model once the toolbox picks it for the turn.
- **Classes.** A server is `outward` unless config marks it `inward` (Nature 7 — it is a mouth
  outside the world). Within that, MCP tool annotations only tighten: `destructiveHint` →
  destructive, `openWorldHint` → outward; `readOnlyHint` lowers nothing on its own — only config
  may (`mcp.servers.<s>.tools.<t>.class`). Every call passes the same permission rules and gate.
- **Containment.** A stdio server runs under the same sandbox and env scrubbing as `bash`, with
  its own `env` / `envAllow` and a `network: true|false`; a server that fails to start or dies is
  a `@?` in `status`, and its tools report the fault instead of vanishing.
- **Resources and prompts.** Resources are readable through the `read` tool as
  `mcp://<server>/<uri>`; prompts become slash commands `/<server>:<prompt>`.

## When a tool is missing

The model never meets a silent failure, and never loops on one:
1. **Unknown or disabled tool called** → one error result naming why (`disabled by
   tools.webfetch.enabled in .isekai/config.yaml:12` or `no such tool`) and the closest enabled
   alternatives. The same missing call twice in a turn stops the turn (doom-loop guard).
2. **A capability the body doesn't hold** → `toolbox` search first: the registry covers disabled
   tools, MCP servers, skills and commands too, so the model learns what exists, what is off, and
   what loading it costs.
3. **Composable from what is on** → it is composed (e.g. `ls` via `bash` when `ls` is off), under
   the same classification and gate — a disabled tool is never re-enabled by the back door: a
   `bash` call that reproduces a disabled *outward* tool is still outward and still asks.
4. **Not composable** → the need is named: a `@?` hole to the dispatcher, one hop (Absolute Rule
   III), and a shared note of kind `colony` recording the missing capability.
5. **Named twice** (Nature 4 — Genesis) → the binary proposes the tool: a drafted `tools.custom`
   entry or an `enabled: true` patch, shown to the human as a config diff. Config is the human's
   written word, so it lands only on their yes; then it reloads live.

## Open models

Open-weight models (served by vLLM, Ollama, SGLang, OpenRouter) differ from Claude in four ways
the binary adapts to, per provider in config:
- **Tool shape.** A `minimal` tool profile: few tools, flat schemas, short descriptions. Editing is
  the `str_replace` family (`view` / `create` / `str_replace` / `insert`) as a custom tool — the
  shape most open coding models met in training through SWE-agent/OpenHands-style scaffolds.
- **Tool-call transport.** `toolCalls: native` uses the server's function calling (on vLLM it
  needs `--enable-auto-tool-choice` and the model's `--tool-call-parser`); `toolCalls: text`
  asks for calls as tagged text and parses them in the binary, for servers or models whose native
  calling is broken. A malformed call gets one error result naming the fault, never a guess.
- **Wire by grammar.** Where the server supports guided decoding (`response_format` with a JSON
  schema), a Court Body's report is constrained to the envelope's schema and rendered to the wire
  by the binary — a small model cannot drop `@S` or `@U`. Elsewhere the wire is parsed and a
  malformed report is re-asked once.
- **The real window.** `contextWindow: auto` reads the served maximum (vLLM's `/v1/models`
  `max_model_len`); budgets, the stress zone and the drain trigger scale to it, so a 32k model
  drains early instead of overflowing.

## Models per office, rank and task

The throne never chooses its horse: the session (Rimuru) runs on the model the human picked
(`models.default`, or `--model`). Every other call is routed by what the task demands
(`canon/model-assignments.md`):
- **The triad is chosen by the ask.** The wire already names the office: `@ASK findings` is
  Great Sage's (perceive), `@ASK verdict` Raphael's (judge), `@ASK draft` Ciel's (speak). A
  `dispatch` runs on `models.offices.<office>`; the binary's own calls are routed the same way —
  the end-of-turn gate verdict to Raphael, log entries and the drain's desk to Ciel, recall
  summaries to Great Sage.
- **Resolution, most specific first:** creature (`models.creatures.<name>`) → office → rank
  (`models.ranks.<race>`) → task (`models.tasks.<drain|gate|log|bench>`) → `models.default`. Each
  entry is `provider/model` with optional `effort` and `fallback`. `status` shows the resolved
  model for every office and rank, with the origin of each.
- **The lineage holds.** Great Sage → Raphael → Ciel is one escalating chain: Ciel's model must
  never sit below Raphael's, nor Raphael's below Great Sage's. Providers declare a `tier` per
  model; config that breaks the order is refused at load, and a model with no declared tier is a
  `@?`, never a guess.
- **A script beats a model.** Anything deterministic (journal lines, verdict records, pointer
  tables) is code, and never reaches any office.

## Ranks and bodies, native

- The session is **Rimuru**. A creature is a record read from its doc under
  `.isekai/<race>/<name>/`: race, territory (path globs), traits, worn minds, orc (for a slime).
- A **Court Body** is `dispatch`: a fresh loop with its own context, a commission in the wire
  (`@ROOT @SCOPE @ASK @CAP`), a tool set cut to its rank and territory, its own shell, and one
  wire report back. Its context dies with the call; only what it wrote to disk and its report
  survive. A Court's writes are gated by the Court itself — the end-of-turn gate runs on its
  own account and its verdict is recorded under its own name; nothing propagates upward for a
  second gate. A report with no `@U` line is flagged to the dispatcher (the unsaid). The
  `@U` lines of a background Court are landed by the dispatcher when the report wakes it.
- **Ranks are data, the law's table is the default.** The binary carries the ranks of
  `isekai.md` §The world (Elf, Orc, Slime; Kijin; High Elf, High Orc, Dark Elf) as its built-in
  table: parent, job, whether the rank authors, whether it holds a gate, body mode, default office,
  tools, model. Config (`ranks.<name>`) overrides a field of a built-in rank or adds a new rank; with
  `rankSet: replace` the ranks listed in config are the whole hierarchy and the built-in table is
  not loaded (the default, `extend`, merges over it).
  The law's shape is checked at load and a config that breaks it is refused: the hierarchy is a
  tree rooted at Rimuru; a rank that authors has a gate-holding rank above it (Law 3); a path has
  exactly one gate holder; every rank has an escalation parent (Absolute Rule III); ascended ranks
  keep everything their base rank held. A rank from config becomes an ontology class with its
  bonds, so knowledge flows along it. `status` lists every rank that differs from the law.
- **Territory is enforced.** A Slime's `write`/`edit` outside its territory is refused, not
  warned (Law 2). Escalation is a `@?` to its dispatcher, one hop (Absolute Rule III).

## Live session — talk while the court works

- **Two faces, one engine.** On a terminal the session is the TUI (`canon/tui.md`); on a pipe,
  under `TERM=dumb`, or with `--plain`, it is the line REPL. Both run the same engine, sessions,
  Courts and gate.
- **The human is never locked out.** A turn runs in the background. A line typed mid-turn is
  queued and delivered to the running body at its next tool step (the way Claude Code does);
  `esc` (TUI) or `Ctrl-C` interrupts the turn, a second `Ctrl-C` exits.
- **A stopped turn is not a dead session.** When a turn ends on its tool results (denied,
  escalated, checkpointed), the next ask rides that user message rather than following it — two
  user messages in a row is a shape no provider takes.
- **Courts run in the background.** `dispatch` may be asynchronous: the dispatcher keeps working
  and is woken by the Court's report. A running Court stays addressable (`/send <court> <text>`)
  until its dispatcher accepts the report; then its context dies (§Minds & Bodies — the task is
  over when the report is taken).
- **The court is visible.** `/agents` (and a status line above the prompt) shows every live body:
  rank, office, model, state (thinking · tool · waiting on gate · done), elapsed, context
  occupancy against its window, tokens and cost so far. Starts and reports are announced between
  prompts, one line each (the TUI draws each Court as a block with its report rendered).
- **Consumption is an instrument** (Nature 9). Every provider call's usage (input, output, cache
  read, cache write) is journaled to `.isekai/instruments/usage/<session>.jsonl`, priced from the
  provider's per-model `price` in config (unpriced models show tokens, never a guessed cost), and
  rolled up per body, office, rank, model and session. `/usage` and `isekai usage [--session
  <id>|--since <date>]` read it; `tempest.js` can read the same file. The session's own calls
  carry no office — the throne is not an office; an office labels a dispatched Court and the
  binary's own routed calls (verdict → raphael, drain and log → ciel, recall → great-sage).
- **Budgets are readings too.** `budgets.session.tokens|usd` and `budgets.court.tokens|usd` stop a
  body honestly at the line (checkpoint, report, resume command), never mid-write.

## The gate and Vitality, in code

At the end of any turn that wrote files, before the turn is reported done:
1. **Right slime authored** — every written path maps to the territory of the body that wrote it.
2. **Traits hold** — each touched creature's `verify` commands (from its doc) run; exit codes are
   the reading. The lines that run are the ones the doc held *before* the turn: Vitality makes a
   body edit its own doc in the same turn, so the post-turn lines would let it rewrite the check it
   is judged by. A pre-turn line the turn removed or changed still runs; a line the turn added is
   the model's own command and does not run — both are holes for the gate holder to confirm. Every
   verify line and injected check runs under the sandbox: no network, secrets scrubbed, the process
   group killed at its timeout.
- **Tests intact** (`law.gate.testsIntact`) — a turn may not pass by making the tests easier. The
   loop keeps the test files' text as the turn opens (Go, JS/TS, Python; bounded); a touched test
   file that lost tests, gained a skip or an `.only`, or was deleted fails the gate. Only the human
   decides a test goes.
- **UI system** (`law.gate.ui`) — a turn that wrote under the app's ui dir (`ui/`, `src/ui/`, …)
   regenerates the legend (`<world>/ui-assets/`, derived, never the model's to write) and must pass
   `ui check`'s lints: no literal colour or pixel length outside tokens.css and palettes.css, every
   class the html uses declared, every token read defined, every component header true. The
   screenshots are not taken at the gate; the Mind has the agent take and read them.
3. **Duties done** — the commission's `@ASK` is answered (`@S` present, holes named as `@?`).
4. **Doc truthful** — a change under a territory with no change to its owning doc fails
   (Nature 1). The owning doc is the Slime's doc, else the Orc's.

The verdict (pass / fail + reason) is appended to `log.md` by the binary. A world with no orcs
records `Gate: n/a (no orcs)` and still runs check 4 against any doc it can find.

**A fail goes back once.** A failed verdict is first sent back to the model, its reasons as the next
message, and the same turn continues — the whole turn's writes are gated again — up to
`law.gate.retries` times (default 1; 0 fails at once). Every attempt's verdict is in `log.md`, and
the answer carries a hole naming each send-back; past the retries the turn fails.

**Every write is seen, whatever tool made it.** A tool that knows its paths reports them
(`write`, `edit`, `patch`); the shell, a custom tool and an MCP server do not. So the engine
stamps the world tree — size and mtime per file, `.git` never walked, bounded at 200k files — at
the start of a turn, again after every write-class step (the difference is that step's writes
when it named none), and once more at the turn's end (a write a read-classified command slipped
in, `xargs touch`). The world dir's own instruments and records are not the model's writes and
are not stamped: `instruments/`, `tmp/`, `memory/`, `toolbox/registry.json`,
`ontology/graph/unsaid.ttl`, `log.md`. A tree past the cap is not watched and the hole is named
in the report. A turn's writes are gated once; the next turn of the same session gates only its
own. The `dispatch` step names what its Court wrote, and the dispatcher's gate leaves those paths
alone — one verdict per landing, the Court's. A background Court still writing when its
dispatcher's turn ends is the one case the dispatcher's gate may see a Court's write.

## The ontology

Knowledge passes between creatures along a typed graph of their relations. The graph is
formal (so it can be checked) and projected to terse text (so a model reads it cheaply).

- **Schema** — `.isekai/ontology/schema.ttl`, a Turtle subset: classes (`Creature` ⊃ `Rimuru`,
  `Elf`, `Orc`, `Slime`, `Kijin`; `Mind`; `Doc`; `Fact` ⊃ `Law`, `Colony`, `Territory`), properties
  with domain, range, `subPropertyOf`, `inverseOf`, transitivity, and cardinality shapes. The
  bonds are the world's own: `truth` (slime ⇒ orc), `verdict` (orc ⇒ elf), `wears` (body ⇌ mind),
  `owns` (creature → path), `knows` (creature → fact), `about` (fact → creature | path).
- **Graph** — asserted triples in `.isekai/ontology/graph/*.ttl`, plus triples *derived* at load
  from the world itself (creatures, territories, worn minds, orcs). Derived triples are never
  written back; files are truth.
- **Reasoning** — forward-chaining to a fixpoint over RDFS-style rules (subclass, subproperty,
  domain/range typing, inverse, transitive). Small, total, deterministic.
- **Validation** — shapes: every Slime has exactly one `truth` edge; every Mind is worn, or is
  shared (a host skill with no race prefix); no two Slimes own an overlapping path (Nature 2). A violation is a finding.
- **Flow** — analysis flows up: a `Fact` a Slime `knows` is visible to every creature up its
  `truth`/`verdict` chain. Wisdom flows down: a `Law` known above is visible below. A Court's `@U`
  lines become asserted `Fact`s, typed by their kind, attached to the reporting body.
- **Projection** — `isekai onto project <creature> --budget N` walks what that creature may see
  and emits one plain line per fact, nearest first, under the budget:
  `slime-auth ⇒truth orc-security · token TTL is 15m (territory)`. This is what a dispatched Court
  receives; the Turtle never enters a prompt (Tokens, not eyes — no IRIs, no prefixes).

## Compaction — drain, don't summarize

A generic harness compacts by asking the model to summarize the whole conversation into one blob:
lossy, blind to what is already on disk, and the unsaid dies with the context. Here the memory
tiers already say where every kind of knowledge lives, so compaction is a **drain**: each piece of
context moves to its home, and the new context is rebuilt from those homes by pointer.

**Trigger — a reading, never an overflow.** The context instrument crosses the configured
threshold (default: the 180k stress zone of the 200k budget, or a fraction of the model's real
window, whichever is lower). Never on a provider "context too long" error — by then it is too late
to drain honestly.

**The drain, in order** — mechanical passes first, the model only where judgment is needed:
1. **Point, don't carry** (no model). Every tool result that is a file read, grep or glob is
   replaced by a pointer: `path:line-range · sha256-prefix`. A file already on disk is never kept
   twice. Duplicate reads collapse to the latest.
2. **Trim the spent** (no model). Tool outputs of finished steps shrink to their journal line
   (exit code, verdict, tail) — the full output is already in `.isekai/instruments/loop/`.
3. **Surface the unsaid** (model, one call). The model is commissioned `@ASK findings +unsaid`
   over the turns being drained and answers in the wire. Each `@U law|colony|territory` goes home:
   a shared note (`memory remember --kind`), an ontology Fact attached to the current creature, and
   for territory a proposed edit to the Slime's doc (gated like any write). `@F` lines worth
   keeping land as shared notes.
4. **Write the desk** (same call). The working memory — at most ~5 dated thoughts: the goal, the
   plan's position, decisions taken and why, open holes (`@?`), the next step. This is the
   hippocampal buffer, not a summary of history.
5. **Record the episode** (no model). If the drained turns landed changes, their `log.md` entry is
   appended now, not at the end of the session.

**The rebuilt context**, in order: the crest · the commission (the human's ask, verbatim) · the
desk · the ontology projection for the current creature · the recall manifest for the current ask
(memory hits + toolbox `@T` lines, pointers only) · the pointer table · the last N turns verbatim.

**Nothing is lost that can be recalled.** The drained turns are indexed into the short tier's
semantic cache for this session; `recall` can bring any of them back on demand, and any pointer
can be re-read. A pointer whose file changed since (digest mismatch) is flagged, not trusted
(Nature 9).

**Verify the drain.** After rebuilding, the binary checks: the ask is present verbatim; every `@?`
open before is still present; every pointer resolves; tokens after < tokens before. A failed check
aborts the compaction and keeps the old context. Every drain is journaled with before/after tokens,
pointers made, notes and facts landed.

**Dispatch first, drain second.** In the stress zone the cheapest compaction is not to accumulate:
heavy work goes to a Court Body whose context dies with its task (§Minds & Bodies). Compaction is
for what dispatch could not avoid.

Every pass is switchable in config (`compaction.passes.*`); a `summary` strategy exists as the
generic fallback, and `status` shows which strategy and passes are live.

## Handoff — what a fresh session needs

`/handoff [focus]` (or `<dist> handoff [focus]`) is a turn: the binary gathers what it knows — the
session's first ask, the uncommitted changes, the recent commits, the last log.md entries, the previous
handoff to carry forward — and the model writes `<world>/handoffs/<time>.md` from a fixed template:
goal, why, state (done · partial · not started — state, not orders), decisions and why, traps and dead
ends, pointers (by path, never copied), open work. Secrets by location only. The next session's welcome
names a handoff under two weeks old; `/handoff read [path]` hands it to the model with one rule: read
every listed file, trust no claim unverified, then wait for the human. The context stress zone points
at it. (After davidondrej/skills' handoff, MIT.)

## Goals — work until a command proves it

`<dist> goal --validate "<cmd>" [--read <files>] [--constraints <text>] [--max-turns N] "<objective>"`
runs one session to a contract: objective, what to read first, what must not change, the validation,
the stop condition. The binary runs the validation itself after every turn (bash in the world root,
15 min) — the model can neither skip nor edit it — and hands a failure back as the next turn's ask with
the exit code and the output's tail. It stops when the validation passes (exit 0), when the model ends
an answer with `@? human: <what it needs>` (exit 3), at the turn ceiling (default 12, exit 1), or when a
turn fails or checkpoints. The contract forbids weakening tests; the gate's tests-intact enforces it on
every turn. (After davidondrej/skills' goal-loop, MIT.)

## Review — two reviewers, one shortlist

`/review [range]` (or `<dist> review [range]`) runs two independent reviewers in parallel, on two
offices' models (raphael and ciel: two models when config gives them two), each with the same neutral
brief — read the diff, the changed code in full, the code around it and its tests; report serious or
critical issues with file:line, why and the fix; separate verified from suspected; say whether it is
ready to merge. The range is the argument, else the uncommitted changes, else the last commit. A
reviewer's shelf is read-only: anything above a read is refused, never asked. A reviewer that does not
finish fails the review — a partial review is not a review. Both reports go to the session, which
merges them: deduplicated, judged (agreement alone does not make an issue real), a numbered shortlist
marked [both] / [raphael] / [ciel] with the ones both found first, the count dropped as overthinking,
and a request for the human's approval — nothing is fixed before it. In a session the reviewers run in
the background and the merge arrives as a turn. A world's own `/review` command wins over the built-in one. (After davidondrej/skills' total-review, MIT.)

## The first run — a setup form

`init --level` founds the world at one of three depths (a terminal asks which when no level is named):
**light** (or soft) writes the law, the log, the instruments and the memory tiers — creatures come
later; **medium** also founds a colony from the tree with no model, sized to the project. It reads
the build files (go.mod and nested modules, package.json workspaces, pyproject, Cargo workspaces,
Maven and Gradle modules), counts files and lines, tells source from tests, skips vendored, built,
generated and fixture dirs, and reads imports where that is cheap (Go, Python, JS/TS, JVM, the
workspace dependencies). A slime per meaningful package or module dir: namespace dirs (`src/`,
`internal/`, `pkg/`, `packages/`, …) and giants are opened into their sub-dirs, a tiny dir folds into
what it imports; territories are directories (the root's own source files are the one file list)
and every name is chosen so the ontology derives each claim. Orcs group related slimes by imports,
shared neighbours, role (entry points, interfaces, config, tests) and path; one elf carries up to
three orcs, and a larger colony gets lane elves plus `elf-canon` for the manifests and docs; a kijin
holds CI, release and infra (`.github/`, `Dockerfile*`, `deploy/`, …). Each doc carries its purpose
(package doc, package description or README line), the traits the sketch observed marked `sketch:`
(sizes, exported names, entry points, imports, missing tests), minds from the world's base-skills
where they fit, and verify lines only where tests exist and a runner is known (`go test ./x/...`, a
workspace's own test script, `pytest`, `cargo test -p`, …). It never overwrites a doc, leaves out a
creature whose ground the world's own creatures already hold, and reports the census and the lanes;
**complex** runs a founding session: the model reads
every source file and decides the team the code needs — how many elves, orcs, slimes, a kijin,
which minds (zone, verdict, global) and which bodies (court, keeper) — writes the docs, skills and
agents in the shapes the ontology reads, keeps only verify lines that pass now, and runs
`onto check` until it is clean; the medium sketch is its first hint (`app/found.md` is the brief).
The founding turn writes every creature's doc as the session, so its gate lifts "right author" for
that turn only; every other check holds.

`init` founds the world; on a terminal it then asks, in a Huh form, which model the world runs on —
keep the global config, OpenRouter's free models, Anthropic, OpenAI, Ollama, or any OpenAI-compatible
server — and writes the world's `config.yaml` (the key's *name* only; the key never lands in a file),
saying so when that variable is not set in the shell. A pipe, `--plain`, a world that already has a
config, or `<PREFIX>NO_SETUP` skips it.

## Claude Code — the law as hooks

A host can run the loop and leave the law to the binary. Under Claude Code the model, the tools
and the session are Claude Code's; `<dist> hook <event>` answers its hooks from the hook's JSON
on stdin, and `<dist> install --claude` wires them into `.claude/settings.local.json` (the
machine's own, like `config.local`) or, with `--global`, `~/.claude/settings.json`.
- **No model, no key.** A hook reads config, the world and the instruments, and answers; it never
  calls a provider. Outside a world it answers nothing, so a machine-wide install is free
  elsewhere. A broken config in a world is a non-blocking error the human sees.
- **SessionStart** injects the crest, the log's last entries and a note on how the law runs here —
  on `startup`, `resume`, `clear` and `compact` alike, so a fresh context reconnects to what was
  written down.
- **UserPromptSubmit** stamps the tree (the turn's baseline, the same watch a session keeps),
  takes the verify lines as the turn opens (each hook is a fresh process, so the gate cannot read
  them later — a line the model adds in the turn is a hole, never run), and answers with the
  recall beat: memory anchors and `@T` lines, pointers only.
- **PreToolUse only tightens.** The guard denies; the world's permission rules deny or ask; with no
  rule, `outward` and `destructive` ask (and `write` under `strict`). An `allow` — a rule's, or a
  class's default — answers nothing, so Claude Code's own permissions still decide.
- **Stop is the gate.** What the turn changed — the file tools' paths and everything the stamps
  show differs, shell writes included — goes through the gate with Rimuru as the author: Vitality,
  verify lines (under the sandbox), injected checks. Duties-done is off (Claude Code answers in
  prose, not on the wire); right-author is a hole, not a failure (the session is Rimuru, and the
  hole names whose body should have authored). A failure answers `decision: block` with the
  reasons, up to `law.gate.retries` times a turn; then the turn may end. The final verdict is
  appended to `log.md` and shown to the human.
- The turn's state lives in `<dist-dir>/tmp/claude/<session>.json`. `install` gives the Stop hook
  a 600 s timeout: the gate runs verify lines and checks, and the host's default is shorter.

## The board — the world, seen

`isekai` lights the board when it starts (`ui.board.autostart`, default on; `isekai board` alone
serves it without a session). It is served by the binary itself on `127.0.0.1` (`ui.board.port`),
reads the same instruments the session writes, and updates live over server-sent events. It
supersedes `tempest.js` once it shows everything tempest shows.
- **One page, at the root; a tab per question.** The board serves one page — the dashboard, at
  `/` — and everything else it serves is an API, the event stream, or the page's own assets. The
  tabs: Overview (the cockpit) · Console · Neural net · Metrics · Court · Memory · Toolbox ·
  Relations · Tree · Config; a deep link opens one (`/#memory`, `/#config`, `/#court/<run id>`),
  and the root opens on the Overview. The old addresses keep working: `/dash` redirects to `/`
  for good, and `/court`, `/memory`, `/toolbox`, `/config`, `/usage`, `/colony` redirect to their
  tab.
- **The world's tabs, read-only.** Court is the history the cockpit only glimpses: every
  journaled run (filter by body, by how it ended, by words), and one run whole — its commission,
  every step with the gate's class and reason, how long it took, what it left unsaid; the bodies
  live this minute head the tab. Memory is the three tiers as instruments (short: the context
  window, the desks, the semantic caches · long: the index and its freshness · shared: the notes
  and the unsaid by kind), a recall over them that writes no cache line, the desks and the notes,
  and the memory's own review on demand (each tier read with its lens, every signal routed).
  Toolbox is the registry (every mind, command, tool and body it harvested, filtered by kind and
  state, with each entry's resident and full cost, its offers and loads), the skills the session
  can wear, the minted bodies, and the loads journal newest first. Config is the resolved config
  as a tree, each value with the layer that wrote it (a filter, "only what a file set", the
  layers counted), credentials shown as set and never their value, and the off-list — every law
  feature this world switched off, with its origin. The JSON behind them: `/api/runs`, `/api/run`,
  `/api/memory`, `/api/memory/recall`, `/api/memory/review`, `/api/toolbox`, `/api/config` — each
  answers only a loopback Host.
- **Ten tabs on any width.** The tabs come in two groups, the session's views and then the
  world's; whatever does not fit the row folds, last first, into a "more" menu at its end (the
  chosen view always stays in the row). Arrows walk the tabs, Home and End jump; the menu is a
  menu (arrows, Esc returns to its button).

**The dashboard — `/`, and `dash`.** The page is built on the world's own ui system (`board/web/`: tokens, palettes, layout primitives,
components — held to `ui check` by a test). It wears its own theme, "realm" (`board/web/theme.css`,
over the foundation's copies): a neon HUD — a near-black ground, neon cyan as the primary and
magenta as the accent, thin hairlines with L-shaped corner brackets on panels and the primary
action, Orbitron for display type and tracked capital labels, Rajdhani for reading, JetBrains Mono
for numbers and code (all SIL OFL, vendored under `board/web/fonts/` and served by the board, so
the self-only CSP holds), a faint slow scanline and a one-time glitch on the brand — every motion
off under `prefers-reduced-motion`. Its colour roles (ground, panel, sunken, border, text, muted,
primary, accent, ok/warn/danger, a hue per rank) are named the same in the terminal ui; dark is
the default, light is a real light mode in the same language (deep cyan and magenta inks on a
cool paper ground), and "system" follows the reader's system; every text colour clears 4.5:1 in
both. Its overview is a cockpit: the living neural net is the
hero; the session's console docks beside it (the transcript read like a page, tool calls, Courts
and verdicts as folded cards, approvals and questions as choice cards, one input: Enter sends, ↑
recalls, Esc interrupts, `/` runs a command) — a splitter resizes it, a button folds it to a rail,
both remembered; over the net a "now" strip shows the session's state, the ask it is on, every
live body as a chip with its context use, and a wait on the human as an inline banner; under the
net a timeline lists the Courts and verdicts, newest first (the journals' recent runs and what the
page saw live, merged by time — a journaled run is folded into its live row only when it is the
same run, started within seconds of it; the throne is named by the lexicon; a row focuses its
neuron, and a pinned neuron's card takes the same band; in 3D the band folds to two rows). The
header carries the vitals: model, state, context, the plan's progress,
cost, live bodies, a wait on the human. The folder tree is a drawer over every view (the files
button, `t`, or Alt+T from the prompt; Esc or a click away closes it); hovering a path lights the
neuron that owns it; open, the drawer is a modal dialog that keeps the focus inside it. Narrow,
the cockpit stacks: net, timeline, console; the header's vitals become one row that scrolls, and
the vitals' row fades at the edge that has more. The net, metrics (tokens,
cost, calls, context, cache; a leaderboard of the creatures by their work — rank, runs, pass rate,
tokens, live or idle; by day, model and body), relations (the bond matrix and its
dimensions: by rank, lane and bond) and the tree keep their own full-size views.
With the terminal open the dashboard mirrors the session and can ask into it; approvals stay in the
terminal. `<dist> dash [--open]` runs the session headless with the dashboard as its only ui, and
approvals are answered in the browser. The board listens on 127.0.0.1 only (with `--docker-agent`
the container shares the host network, so it is still the host's loopback); every act (ask,
answer, interrupt) carries a per-run token the page embeds and must name a loopback Host with a
same-origin Origin, so another site cannot drive the session and a rebound DNS name cannot reach
it. The page's CSP allows only its own scripts.

**How the colony is drawn.** The dashboard's net is the hierarchy as layers, read the way work
flows: the throne (the session body, named by the lexicon) first, then each rank below it — the
elves and the kijin, the orcs, the slimes — and the minds as a side layer, joined to their wearers
by anima-threads that stay hairlines until a focus lights them. `/api/colony` guarantees the shape:
the throne is always a node, every creature names its `parent` by its first upward bond (truth,
verdict, reports, above), and a creature with no way up hangs under the throne by an `implied`
reports edge and a note — the tree is connected and the gap is said (a cycle is broken where it
is found; an empty world is the throne alone, every list a list). Wide, the layers are columns
and the tree is tidy (a parent sits at the middle of its children, each upper layer centred in its
column); each subtree under the throne is a lane — an elf with its orcs and their slimes — set
apart by a gap and drawn as one band; a leaf under the throne (the kijin, an elf with no orcs) is a
satellite stacked just under the throne. In the cockpit the net fills exactly the box its panel
gives it, the legend always in view: rows tighten and tiles shrink first, then the deepest layer
wraps into staggered sub-columns, and no label is cut below six ems — a world too large for the
box grows and its stage scrolls. The minds keep out of the way as a strip of brains under the
layers (their names on hover or focus), a column only when that alone lets the layers fit.
Narrow, it is an outline, one row per neuron, indented by depth. `board/testdata/netlayout`
holds a ~60-creature fixture and a check that lays it out at the boxes 1280, 1440 and 1920 wide
give the net (`TestNetLayoutHolds` runs it when node is on the PATH). Each neuron is a rank-coloured tile whose 28 cells are its last
four weeks (`/api/activity`: runs from the loop journals, tokens from the usage journal — blank
where no instrument saw any, never a made-up zero); each tree bond is as thick as the commissions
it carried. Hover or keyboard focus lights a neuron's path to the throne and its subtree, the rest
steps back; a click pins it and the side card reads its doc, territory, minds, bonds and work. Every action pulses on its own neuron — a session step on the throne, a Court's step on its
creature: a read is a soft accent ripple, a write and a plain turn's end the rank's own pulse, an
outward or destructive act and a waiting gate an amber ring; green and red are for real verdicts
only — a passing one flares green, a failure or denial red. Bursts coalesce: within a burst (acts
under a second and a half apart) a neuron ripples again only when something louder arrives and
otherwise just breathes, so a twenty-step Court is one breathing neuron. A Court on the stream sends a spark down the tree from its dispatcher (the nearest running ancestor,
else the throne) and its report back up, green for DONE, amber or red otherwise; a running body
breathes. The **3D** stage (`net3d.js`, WebGL2 written here, no library) draws the same tree as a
brain. The throne sits at the centre, the thalamus every signal relays through; each rank is a shell
further out (the first rank inner, the last on the outer "cortex"; a first-rank creature with
nobody below it, like the kijin, close to the throne), so rank reads as depth from the centre. Each
lane — a first-rank creature's subtree — is a region of the brain, tinted by a faint glow and
captioned with its head's name; the tint can instead show each region's work in four weeks
(heat) or the share of its recent runs that passed (health). Inside a region a seeded force
layout on two hemispheres holds a creature near its parent and pulls related creatures together:
the colony's `relates` (`board/relate.go`: source files of one territory importing another's, Go
through `go/parser`'s imports-only pass and Python and JS/TS by their import lines, plus minds both
wear; the deepest owner of a path counts, ancestor pairs are the tree's business, each creature
keeps its strongest four), days worked together and runs commissioned in the same minute. The
layout is fixed for a world's shape and settles once, outward from the throne. The minds are a
crown above, each over its wearers, its `wears` filaments falling in the mind's own colour to the
creatures that wear it. Every rank is its own lit solid, one instanced draw per shape: the throne a
stellated icosahedron inside two orbit rings, an elf a slender bipyramid, an orc a hexagonal
prism, a slime a soft wobbling blob, the kijin a tetrahedron, an ascended rank its base with a
crown ring (the dark elf on six sides), a mind a torus; a key in the corner names them in the
lexicon's rank words. Synapses taper from the parent to the child; a dispatch travels as a bright
pulse with a trail that lights each synapse it crosses, the report comes back the same way, a
firing neuron flashes and rings; particles drift down the tree, and down a mind's filament while
its wearer works; faint dendrites branch off the outer neurons and slow motes give depth. Hovering
a relates bond says why it holds (imports, shared minds, days together); "replay" plays the last
turn's signals again on a scrubber (a turn ends after fifteen quiet seconds; pauses inside it
shorten to a second and a half). Drag orbits, the wheel or a pinch zooms, a click focuses (the
camera eases there and the rest of the brain darkens), arrows walk the tree with the camera
following (Enter pins, Shift+arrows turn), "reset view" comes home to a view that fits the brain
and its crown. Labels show for the throne, the first rank and the minds, on hover and focus, and
for every neuron once 24 or fewer are in view. It falls back to 2D without WebGL2 and rebuilds
after a lost context. Both animate only while the net is seen, at a frame budget (the 3D stage
drops to 12 frames a second after ten idle seconds and stops turning and spinning after an idle
minute), and stand still under `prefers-reduced-motion` (no particles, no spin, no settling);
every colour is a token (the canvas reads them through probe elements, its fonts once per theme or
size).

The command's own output — the board, each session — is a Charm Log: levelled, coloured, timed.

## Tests

Every package carries Go tests; providers are exercised through `mock`. `isekai selftest` runs
the in-binary checks the way the JS instruments' `selftest` does and answers on the wire
(`@S PASS n checks`). Interop tests run the binary and the JS tools against the same files.
