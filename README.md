<p align="center"><img src="portraits/rimuru.png" width="96" alt="Rimuru"></p>

<h1 align="center">Isekai</h1>

<p align="center"><b>An agent harness where the law is the harness.</b><br>
One Go binary. The convention isn't a prompt the model may skip — the gate, the log, the human
gate, the sandbox and the memory are code paths it cannot route around.</p>

---

## Why

Coding agents read their project's rules and then, sometimes, don't follow them. Isekai turns a
directory into a **world** governed by [`ISEKAI.md`](ISEKAI.md) and runs the agent inside that
law:

- **The gate** — every turn that changed the world is checked before it's reported done: the right
  creature authored it, its invariants hold, its duties are done, its doc changed with it. The
  verdict lands in an append-only `log.md`.
- **The human gate** — every action is classified (`read` · `write` · `outward` · `destructive`)
  before it runs. Outward and destructive acts ask you; per-command rules (`bash:git push*`) can
  allow, ask or deny. The model's claim about a command can only make it stricter.
- **The sandbox** — `bash` runs in a persistent shell under `bwrap`: the filesystem read-only
  except the world, no network unless the gate let the act out, secrets scrubbed from the env.
- **Memory, native** — recall before every step, remember after; knowledge the agent held but never
  wrote down (*the unsaid*) is surfaced and routed to where it belongs. When context fills, it is
  **drained** into those homes by pointer, not summarized away.
- **Ranks and offices** — the session (Rimuru) dispatches Court Bodies to Elves, Orcs and Slimes;
  each office (Great Sage reads, Raphael judges, Ciel writes) can run on its own model.
- **Everything on, config takes away** — every tool, law feature and instrument is switchable in
  YAML or JSON, and a switched-off law is always shown, never silent.

## Install

```sh
curl -fsSL -o install.sh https://raw.githubusercontent.com/Kaginari/isekai/main/install.sh
sh install.sh                     # downloads the release for your OS and checks it against checksums.txt
# or
go install github.com/Kaginari/isekai/cmd/isekai@latest
```

Or build it: `CGO_ENABLED=0 go build -o isekai ./cmd/isekai` (Go ≥ 1.26).

Release archives for linux/macOS × amd64/arm64 ship with `checksums.txt` and build provenance:
`gh attestation verify <archive> --repo Kaginari/isekai`. A container image to try the binary:
`docker run --rm -it -e ANTHROPIC_API_KEY -v "$PWD:/work" ghcr.io/kaginari/isekai:latest` — the
binary alone on Debian slim (no bubblewrap, git or toolchains); for real work in a container use
`--containered`, which builds a runtime with them.

**Requirements.** Linux or macOS. `bubblewrap` (`bwrap`) for the sandboxed shell on Linux — without it
the shell runs unsandboxed and `status` says so; macOS has no bwrap. Docker for `--containered`.

**Data boundary.** No telemetry. The binary contacts only the providers you configure (and the
package or container registries you name); any other network act is `outward` and asks you, and the
sandboxed shell has no network unless an `outward` act was approved.

## Quick start

```sh
cd your-project
isekai init                      # found a world here: .isekai/ with the law and a config
export ANTHROPIC_API_KEY=…        # or any OpenAI-compatible endpoint (vLLM, Ollama, OpenRouter)
isekai                           # the live session — the board opens at http://127.0.0.1:7411
isekai run "add a health check endpoint and its test"
```

## The terminal

A terminal UI in the class of Claude Code, built on the Charm libraries: the conversation flows into
your scrollback, and every event of the loop is drawn — thinking, tool calls, diffs, Courts, the gate.

![welcome](docs/screens/isekai-welcome.png)

The model's thinking streams while it arrives and folds when the answer starts; each rank has its own
face and verbs beside the spinner:

![thinking](docs/screens/isekai-thinking.png)

When a turn wrote files, the orc weighs the verdict before the turn is done:

![the gate](docs/screens/isekai-gating.png)

Tool calls are cards edged in their class's colour; diffs keep the code's own colours:

![tool cards and a diff](docs/screens/isekai-tools.png)

`ctrl+t` watches every body — the session and each Court, its thinking and its steps, live:

![every body](docs/screens/isekai-bodies.png)

`/board` shows the world full screen — the agents, the reasoned ontology as a graph, the offices, the
usage — and `ctrl+k` opens a fuzzy command palette:

![the ontology graph](docs/screens/isekai-board-graph.png)

![the palette and a toast](docs/screens/isekai-overlay.png)

| Key | |
|---|---|
| `enter` · `shift+enter` | send · newline |
| `/` · `ctrl+k` | commands · the palette |
| `ctrl+t` | every body, live |
| `ctrl+o` | expand the last folded block |
| `esc` | interrupt the turn |

## The CLI

| Command | What it does |
|---|---|
| `isekai` / `repl` | the live session: talk while Courts work; `/agents`, `/send`, `/usage`, `/status`, `/compact` |
| `run "<ask>"` | one ask to its end, non-interactive; `--json` for machines |
| `resume <id>` · `sessions` | continue or list sessions |
| `status` | every live reading (models, sandbox, guard, memory) and every law feature switched off, with where — gaps print as `@?` lines |
| `config show\|explain\|check\|path\|patch` | the effective config with the origin of every value |
| `memory` · `toolbox` · `onto` | the three memories, the two-level toolbox, the creature graph |
| `usage` | tokens and cost by body, office, rank, model and day |
| `bench` | a fixed task set on every configured model |
| `goal --validate "<cmd>" "<objective>"` | work turn after turn until the command passes; the binary runs it, the model cannot skip it |
| `review [range]` | two reviewers on two models in parallel, one merged shortlist; nothing fixed before you approve |
| `handoff [focus]` | a handoff note for a fresh session (`/handoff read` picks it up) |
| `guard check\|test\|show\|export\|hook\|install` | the global dangerous-command guard; `install --yes` wires it into Claude Code and OpenCode |
| `ui init\|scan\|check` | web pages built from one token system: lay the foundation, regenerate the legend, lint and screenshot at 360/768/1280 in both themes |
| `board [--ssh [addr]]` | the board without a session — on the web, and over SSH (keys in `~/.ssh/authorized_keys` only) |
| `--containered` | the whole binary in a Docker container: the world read-write, the rest read-only |
| `selftest` · `version` · `init` | |

## Configuration

Configuration is YAML (or JSON), layered — each layer overrides the one before it:

| Layer | Where | Kept in git |
|---|---|---|
| built-in defaults | in the binary — everything on | — |
| global | `~/.config/isekai/` | no |
| project | `.isekai/` in the world | yes |
| project local | `.isekai/config.local.yaml` | no (machine-only) |
| environment | `ISEKAI_CONFIG=<file>`, `ISEKAI_CONFIG_CONTENT=<yaml>`, `ISEKAI_MODEL`, `ISEKAI_DISABLE_PROJECT_CONFIG=1` (skip the project layers), `ISEKAI_PROVIDER_<NAME>_API_KEY` (built-in provider names only; a custom provider uses `apiKeyEnv`) | — |
| flags | `--model`, `--set key=value`, `--approve outward`, `--dry-run`, `--no-<feature>` | — |

`isekai config explain` shows every effective value and the file and line it came from;
`isekai status` shows every model, the guard, the sandbox and anything switched off.

### One file per part

A layer can be one `config.yaml`, or split: beside it, a file named after a section holds just that
section. The example in [`examples/gateway/`](examples/gateway/) — isekai on vLLM-hosted models
behind a gateway — is laid out this way:

```
.isekai/
├── config.yaml        # the world's own settings
├── providers.yaml     # model endpoints
├── registry.yaml      # models, tools, package and container registries
├── models.yaml        # who runs on what
├── guards.yaml        # catastrophic commands, refused before any approval
├── rules.yaml         # rules every creature reads, optionally checked at the gate
└── permissions.yaml   # ask / allow / deny per command
```

**`providers.yaml`** — where the models are. Keys are named, never written:

```yaml
gateway:
  enabled: true
  type: openai                       # any OpenAI-compatible endpoint: vLLM, a gateway, OpenRouter, Ollama
  baseURL: https://llm-gateway.example.corp/v1
  apiKeyEnv: GATEWAY_API_KEY         # the variable's name; export the key in your shell
  toolCalls: native                  # vLLM with --enable-auto-tool-choice; "text" for tags in the text
  contextWindow: auto                # read from /v1/models (max_model_len)
  timeout: 10m
  # headers: { X-Tenant: platform }        # a routing header; a credential here is refused
  # tls: { caFile: /etc/ssl/certs/corp-ca.pem }
```

`contextWindow: auto` reads the served maximum from `/v1/models`; a gateway that does not report it
needs `contextWindow: <n>` here or per model in `registry.yaml`. (A provider named `vllm` needs no
`type`: it is known.)

**`registry.yaml`** — where things come from: the models a gateway serves, the tools the agent can
discover, a private Artifactory, the container registry (the example ships `packages` and `containers`
commented out):

```yaml
models:                                     # "<provider>/<id>" — tier orders the offices; price in USD per 1M tokens
  gateway/muse-glimmer: { tier: 1, contextWindow: 131072, price: { input: 0, output: 0 } }
  gateway/glm-5-3:      { tier: 2 }
  gateway/kimi-k3:      { tier: 3 }
tools:                                      # external tools the toolbox indexes
  - { name: kubectl, description: "the cluster CLI", triggers: [kube, pods, deploy] }
packages:                                   # set in every shell command and in the container
  npm: https://artifactory.example.corp/artifactory/api/npm/npm-remote/
  pip: https://artifactory.example.corp/artifactory/api/pypi/pypi-remote/simple
  go:  https://artifactory.example.corp/artifactory/api/go/go-remote
  tokenEnv: ARTIFACTORY_TOKEN               # let through by name — the agent can read it: use a read-only token
  env: { GONOSUMDB: example.corp }
containers:                                 # for --containered
  base: artifactory.example.corp/docker-remote/debian:bookworm-slim
  apt:  https://artifactory.example.corp/artifactory/debian-remote
  # image: artifactory.example.corp/docker-local/isekai-runtime:1   # pull a prebuilt runtime instead
```

For a private container registry, `docker login artifactory.example.corp` first: docker's own login
answers for the pull.

**`models.yaml`** — the session's model and the three offices (Great Sage reads, Raphael gives verdicts,
Ciel drafts). A call is retried on 429 and 5xx (honouring `Retry-After`); when it still fails, or
fails outright on a transport error, the fallback is tried — never on a refusal. An *office* is what a
Court Body does for a call; a *rank* is its place in the tree (elf → orc → slime); a *creature* is
one named body — each can be given its own model:

```yaml
default: { model: gateway/kimi-k3, fallback: gateway/glm-5-3 }
offices:
  great-sage: { model: gateway/muse-glimmer, fallback: gateway/glm-5-3 }
  raphael:    { model: gateway/glm-5-3,      fallback: gateway/kimi-k3 }
  ciel:       { model: gateway/kimi-k3,      fallback: gateway/glm-5-3 }
# ranks: { slime: …, orc: … }   creatures: { slime-auth: … }   tasks: { drain: …, gate: … }
```

**`guards.yaml`** — added to the built-in denylist (disk wipes, `rm -rf ~`, force-pushes, repository and
secret deletion, history purges, secret-store reads, `curl … | sh`). A match is refused before any
approval — no rule or `--approve` can run it:

```yaml
- '(^|[[:space:]])terraform[[:space:]]+destroy'
- 'kubectl[[:space:]]+delete[[:space:]]+(ns|namespace|node)'
- 'helm[[:space:]]+uninstall'
```

**`rules.yaml`** and **`permissions.yaml`**:

```yaml
# rules.yaml — every creature reads them; a check runs at the gate
- { text: "Tests must pass before a change lands.", check: "make test" }
- { text: "No secret, key or token is ever written to a file." }
```

```yaml
# permissions.yaml — ask / allow / deny per command. An allow on bash, git or webfetch is a
# loosening that `config check` lists; a bare bash:* allow is refused.
rules:
  - { match: "bash:git push*", action: ask }
  - { match: "bash:kubectl get*", action: allow }
  - { match: "bash:kubectl*", action: ask }
```

### Every other section

| Section | What it controls | A taste |
|---|---|---|
| `tools` | the builtin tools, their limits, custom tools, the bash sandbox | `bash: { sandbox: bwrap, timeout: 2m }` · `custom: { kube-pods: { run: [kubectl, get, pods], class: read } }` |
| `law` | the gate and its checks, the human gate, budgets per turn | `gate: { retries: 1, testsIntact: true }` · `humanGate: { approve: [outward] }` |
| `memory` · `compaction` | recall and notes; draining the context when it fills | `compaction: { trigger: { fraction: 0.85 } }` |
| `mcp` | MCP servers (Claude Code's `.mcp.json` and OpenCode's are imported) | `servers: { gh: { command: [gh-mcp] } }` |
| `hooks` | shell hooks at preTool, postTool, sessionStart, preCompact, stop, userPrompt | `preTool: [{ match: "bash", command: "./check.sh" }]` |
| `discovery` | where instructions, skills, commands and agents are found (Claude Code and OpenCode dirs included) | `skills: { paths: [.claude/skills] }` |
| `budgets` | spend ceilings for the session and each Court | `session: { tokens: 400000, usd: 5 }` |
| `ui` · `output` | the dashboard, the status line, output format | `board: { autostart: false }` |
| `sessions` · `undo` | where sessions and undo snapshots live, how long | `sessions: { keepDays: 30 }` |

| `toolbox` · `ontology` · `instruments` | the searchable tool registry, the creature graph, the journals | `toolbox: { budgetTokens: 1500 }` |
| `mode` · `smallModel` · `logLevel` | `build` or `plan` (plan writes nothing); a small model for small tasks | `mode: plan` |
| `tools.profile` · `tools.missing` | the toolset shape (`max`, `anthropic`, `openai`, `minimal`); what happens when a tool is missing | `profile: minimal` |
| `permissions.import` · `mcp.import` | take Claude Code's and OpenCode's permissions and MCP servers | `import: { claudeCode: { enabled: false } }` |
| `guard` | more denylist files, patterns inline | `files: [~/corp-guards.txt]` |

Everything is on by default; switching something off is always shown in `status`, never silent.
`isekai config explain` shows every value with the file and line it came from. The full schema:
[`docs/canon/config.md`](docs/canon/config.md) · the design: [`docs/canon/binary.md`](docs/canon/binary.md) · the terminal:
[`docs/canon/tui.md`](docs/canon/tui.md).

## Web UIs

When the agent builds a page, it wears the built-in **`ui` Mind** (a project skill of the same name
overrides it). The app's own `ui/` dir is the source of truth; nothing is copied into the world:

```
ui/tokens.css      raw scales → meaning tokens (--surface --ink --accent --space-m --step-1 …), both themes, density
ui/palettes.css    thirteen palettes in OKLCH, contrast-checked; <html data-palette="forest"> swaps every colour
ui/layout.css      the page grid and primitives: .page .stack .cluster .grid .cols .sidebar .switcher .center .cover .frame
ui/components/<name>/<name>.{css,html}   one component; its css starts /* @component <name> — tokens: … */
```

A global or radical change is one edit in one tier (a palette, `--density`, `--type-ratio`, a
primitive); no component is touched. `isekai ui scan` derives the **legend** —
`.isekai/ui-assets/{manifest.json,legend.md,catalogue.html}`: every component, the classes it owns,
its variants, the tokens it reads. The agent reads it before it changes anything. `isekai ui check`
lints the system (no literal colour or pixel length outside the token files, every class declared,
every token defined, every header true) and renders the catalogue with headless Chromium at 360, 768
and 1280px in light and dark, failing any component that scrolls sideways. The gate runs the lints on
every turn that touched the ui dir (`law.gate.ui`). A project already on Tailwind, Bootstrap or a
design-system package keeps it: the Mind follows the existing system.

## Safety, in code

Beyond the gate, the human gate and the sandbox described above:

- **The guard** — the catastrophic and irreversible are refused before any approval, held to a
  151-case test corpus; `isekai guard install --yes` wires the same list into Claude Code and
  OpenCode (without `--yes` it only shows the change).
- **The gate, hardened** — it runs the checks as they were before the turn, fails a turn that
  deleted or skipped tests, and sends a failure back to the model once, then to you.
- **Dry-run and previews** — `--dry-run` plans and shows what it would run or ask, running nothing;
  every edit shows its diff; `/board` and `ctrl+t` show every agent live.
- **The sandbox** (`bwrap`) or **the container** (`--containered`: the world read-write, the rest
  read-only, pulled or built from your private registry).

## Benchmarks

Benchmarks run under [Harbor](https://github.com/harbor-framework/harbor) in Docker — one container
per task, the agent inside it, the task's own tests deciding. Every launch is recorded in
[`bench/runs.jsonl`](bench/runs.jsonl); [`bench/RESULTS.md`](bench/RESULTS.md) is generated from it.
The €0 smoke (`bench/smoke.sh`) runs on every commit in CI against a fake OpenAI-compatible server.

## License

MIT — see [LICENSE](LICENSE). Bootstrap 5 (embedded in the board) is MIT.
