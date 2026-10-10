<p align="center"><img src="portraits/rimuru.png" width="96" alt="Rimuru"></p>

<h1 align="center">Isekai</h1>

<p align="center"><b>An agent harness where the law is the harness — and the world learns from its own runs.</b><br>
One Go binary. The convention isn't a prompt the model may skip — the gate, the log, the human
gate, the sandbox and the memory are code paths it cannot route around; every run's per-step
readings (`guts`) cluster into what `lessons` turns into skill candidates.</p>

---

## Why

Coding agents read their project's rules and then, sometimes, don't follow them. Isekai turns a
directory into a **world** governed by [`ISEKAI.md`](ISEKAI.md) and runs the agent inside that
law:

- **Learns from its own runs** — the guts score every act mid-court and journal the reading;
  `isekai lessons` clusters what kept costing ground across runs, so the next body crystallizes
  the top class into a skill — a run stops waiting for the floor's verdict to learn.
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

From source (the estate way — Go modules resolve through the Artifactory proxy):

```sh
export GOPROXY='https://<user>:<token>@artifactory.example.com/artifactory/api/go/go-remote'
git clone https://gitlab.example.com/team-example/isekai.git
cd isekai && CGO_ENABLED=0 go build -o isekai ./cmd/isekai && ./isekai version
```

Or the Docker image (built by `.gitlab-ci.yml`, listed in the portal):

```sh
docker login registry.example.com
docker run --rm -it -e GATEWAY_API_KEY -v "$PWD:/work" \
  registry.example.com/team-example/isekai:<tag>   # current tags listed in the portal
```

**Requirements.** Linux or macOS. Inside the image: bash (the shell tool) and ca-certificates incl.
the corp CA; git and curl install only in CI kaniko builds (the estate apt helper ships there) —
local images go without. On a
host, `bubblewrap` (`bwrap`) for the sandboxed shell on Linux; without it the shell runs unsandboxed
and `status` says so (macOS has no bwrap). Docker for `--containered`. Against the estate gateway set
the provider to OpenAI-compatible (`examples/world-rules.yaml` carries a working config with the
bench-earned laws); the API key travels as an env-var *name*, never a value in a file.

**Estate wiring, pinned.** Models answer on the conduit front-door: `baseURL:
https://gateway.example.com/v1/`. Package registries are Artifactory remotes on
`artifactory.example.com` (npm / pypi / go). Container images always pass by a docker vhost:
`registry.example.com/<image>:<tag>` for hub proxies,
`registry.example.com/team-example/<image>:<tag>` for estate builds — `docker login
registry.example.com` first.

**Data boundary.** No telemetry. The binary contacts only the providers you configure (and the
package or container registries you name); any other network act is `outward` and asks you, and the
sandboxed shell has no network unless an `outward` act was approved.

## Quick start

First, plug a model. On the estate the gateway is an OpenAI-compatible endpoint serving **conduit**
model ids directly — no proxy hop:

```yaml
# .isekai/config.local.yaml
providers:
  gateway:
    enabled: true
    type: openai                       # OpenAI-compatible: the conduit ids answer on it natively
    baseURL: https://gateway.example.com/v1/   # the conduit front-door (hprd; not a local lab proxy)
    apiKeyEnv: GATEWAY_API_KEY         # preferred: name the env var
    # apiKey: "<value>"                # fallback when an env var can't be set (headless containers);
                                       # a key literal never belongs in a committed file
    toolCalls: native
    contextWindow: auto
models:
  default: gateway/kimi-latest         # or glm-latest, glm-5-3-llmd, muse-glimmer-30b, …
```

`examples/world-rules.yaml` is the same shape with the bench-earned laws already in. (Full config
language below.)

### without Docker

```sh
cd your-project
isekai init --level complex           # reincarnate a world here (light: the law only · medium: a
                                      # colony sketched from the tree · complex: the model reads
                                      # the code and founds the colony)
export GATEWAY_API_KEY=…              # your key in the env, never a file
isekai                                # the live session — board at http://127.0.0.1:7411
isekai run "add a health check endpoint and its test"
```

### with Docker

```sh
docker pull registry.example.com/team-example/isekai:<tag>   # current tags listed in the portal
docker run --rm -it -e GATEWAY_API_KEY -v "$PWD:/work" \
  registry.example.com/team-example/isekai:<tag>   # current tags listed in the portal
```

The image carries only what the CLI needs: bash (the shell tool), git, ca-certificates,

### the dashboard through Docker

`dash` is the session in the browser: the chat console, the neural net, metrics — and approvals get
answered there. The board binds `127.0.0.1` by design (and rejects non-loopback Host headers), so
`docker run -p` cannot reach it — run on the host network instead:

```sh
docker run --rm --network host \
  -e GATEWAY_API_KEY -v "$PWD":/work -w /work \
  registry.example.com/team-example/isekai:<tag> dash --open
# → http://127.0.0.1:7411/dash — ctrl-c stops
```

The `-e GATEWAY_API_KEY` line rides only if the key comes from your shell env. If your gitignored
`.isekai/config.local.yaml` already holds `providers.gateway.apiKey` as a literal (the harness's
designated escape for machine-local files), the key travels inside the mounted world and the `-e`
line can go.

Host networking is native on Linux; on Docker Desktop enable it under Settings → Resources →
Network. A local image for the same flow: `docker build --build-arg GOPROXY="$GOPROXY" -t
isekai:local .` (GOPROXY from your `.env`). Pulling
from a private registry a world can reach is `registry.containers.usernameEnv/tokenEnv` (or
`username` / `token` inline when a body runs headless with no env).

## The CLI, seen

A working session: the ask up top, every classified step drawn as it lands:

![the session](docs/screens/isekai-welcome.png)

## Architecture

**Wrapped layers.** A model on its own is text in, text out — no hands, no memory, no check on
itself. A *body* is that model inside a loop with tools. A *world* is everything around the loop.
Isekai is one static Go binary at the outermost shell — the law runs code:

<table>
<tr><th>layer</th><th>wraps</th><th>is</th></tr>
<tr><td><b>LLM</b></td><td>—</td><td>the bounded mind — text in, text out; no hands, no memory, no check on itself</td></tr>
<tr><td><b>body</b></td><td>the LLM</td><td>the model inside a loop with tools — the model walks</td></tr>
<tr><td><b>world</b></td><td>the body</td><td>loop + context/memory + permissions/sandbox — the code protects</td></tr>
<tr><td><b>Isekai</b></td><td>the world</td><td>one static Go binary at the outermost shell — the law runs code</td></tr>
</table>

**Six beats per step, one model call.** The model lives only in the plan-and-act beats; perceive,
recall, verify and record are code — no tokens, milliseconds. `law.guts` scores each act at verify
and speaks once inside a slide; the witness demand runs at close when files were written.

<table>
<tr><th>beat</th><th>runs on</th><th>spend</th></tr>
<tr><td>1 · perceive — context read, budgets before any call</td><td>code</td><td>no tokens</td></tr>
<tr><td>2 · recall — anchors rebuilt per call, never cached</td><td>code</td><td>no tokens</td></tr>
<tr><td>3 · plan — <b>the one model call</b></td><td>model</td><td>tokens</td></tr>
<tr><td>4 · act — the tool calls the plan ordered</td><td>model directs, code runs</td><td>tokens (the plan's)</td></tr>
<tr><td>5 · verify — checks run; <code>law.guts</code> scores the act</td><td>code</td><td>no tokens</td></tr>
<tr><td>6 · record — journal, memory, roster state</td><td>code</td><td>no tokens</td></tr>
</table>

**A task, end to end** — tools act inside the world; nothing closes on belief.

<ol>
<li>the task lands — a prompt, or a commission a court carries</li>
<li>perceive: context read, budgets checked before any call</li>
<li>recall: anchors + the world's law rebuilt for this call</li>
<li>one model call — the plan
<ol type="a">
<li>each tool call it orders is classified first (read · write · outward · destructive)</li>
<li>the guard refuses catastrophic patterns before any ask</li>
<li>outward and destructive ride to the human gate; a denial comes back as a message the model reads</li>
<li>an approved act runs sandboxed — timeout, kill tree, fs under the root</li>
<li>the step lands: observe marks it, record journals it — and while tool calls remain, the loop turns again</li>
</ol></li>
<li>the model stops calling tools: the turn-end gate runs its checks — wrote files? an orc holds the gate</li>
<li>the wire rides up: status, findings, holes, unsaid — the context dies with the task</li>
</ol>

## What lands where

Three tiers keep the law clean: rules teach, code enforces, instruments remember.

<table>
<tr><th>tier</th><th>duty</th><th>homes</th></tr>
<tr><td><b>rules teach</b></td><td>steer the model</td><td><code>ISEKAI.md</code> · <code>rules:</code> entries — injected into the system prompt, advice never law</td></tr>
<tr><td><b>code enforces</b></td><td>protect the world</td><td>the gate · permissions · the guard · budgets · territory ownership</td></tr>
<tr><td><b>instruments remember</b></td><td>make runs accountable</td><td>the wire formats (<code>@S</code>/<code>@F</code>/<code>@?</code>/<code>@U</code>) · the journal · memory anchors</td></tr>
</table>

**Terminology, once and for all.** A *prompt* is what you type — the ask. A *system prompt* is the
machine's standing instruction, auto-composed at every turn before the ask rides in. *Code* is
everything the world enforces without asking a body to believe anything.

**The system prompt, composed per turn, contains:**

- identity + standing: who this body is, the world it belongs to, its ranks and offices
- the law: `ISEKAI.md` in full when `law.principles` is on (default), plus every `rules:` entry
  verbatim
- the four action classes and who asks before a privileged one runs
- the wire it must answer in (`@S` / `@F` / `@U` / `@D` / `@E`)
- instrument readings it may claim: budgets read per step; the guts' announced behavior
- docs-as-code duty and the colony map for this world

**In per-step context, on top of the ask:**

- memory anchors: notes matched to this ask (short / long / shared tiers, budgeted)
- the toolbox manifest — only the tools that fit this step's question
- working notes, journal tails, the last tool results (the step's own evidence)
- the witness demands when files were written and the body's close is being checked

**Never in any prompt — coded and enforced:**

- the guard (dangerous commands refused before any approval) and the gate (who may approve)
- the sandbox knot (bwrap / `--containered` envelope)
- budgets read as instruments: context window % readings, stress-zone behavior, session wallets
- the closing witness's admissibility rule (machine-check or independent method)
- the guts' scoring, the echo penalty, and where every reading journals
- wire parsing, reports, sessions, undo, hooks — surface, not advice

- **`rules:` are converted to the system prompt.** Each `text:` line is law text the binary reads
  into every turn's system message — shape culture here, never private muscle (omitted from a
  subagent's prompt unless `raw` is on; an entry can be a file path instead of inline text).
- **The ask, memory anchors, the toolbox manifest and witness evidence go in per-step context** —
  what this step needs, nothing it doesn't.
- **`law:` is enforced in code, never prompted** — the guard and gate classes, the sandbox,
  budgets read as instruments, the closing witness demand, the guts' scoring. A body can only
  tighten a law, never loosen it; `status` names every switched-off law aloud.

## Deep: the terminal

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
| `config scaffold [--check]` | mint (never overwrite) the machine-local config set — `config.local.yaml` + `<section>.local.yaml`, commented and default-valued — or drift-check it against this binary |
| `memory` · `toolbox` · `onto` | the three memories, the two-level toolbox, the creature graph |
| `usage` | tokens and cost by body, office, rank, model and day |
| `bench` | a fixed task set on every configured model |
| `goal --validate "<cmd>" "<objective>"` | work turn after turn until the command passes; the binary runs it, the model cannot skip it |
| `review [range]` | two reviewers on two models in parallel, one merged shortlist; nothing fixed before you approve |
| `handoff [focus]` | a handoff note for a fresh session (`/handoff read` picks it up) |
| `guard check\|test\|show\|export\|hook\|install` | the global dangerous-command guard; `install --yes` wires it into Claude Code and OpenCode |
| `ui init\|scan\|check` | web pages built from one token system: lay the foundation, regenerate the legend, lint and screenshot at 360/768/1280 in both themes |
| `dash [--open]` | the session in the browser: the console, the neural net, metrics, relations; approvals answered there |
| `board` | the board and a read-only dashboard without a session, on 127.0.0.1 |
| `--containered` | the whole binary in a Docker container: the world read-write, the rest read-only |
| `init [--level light\|medium\|complex]` | found a world: the law only; a colony sketched from the tree (no model); or a founding session that reads all the code and decides the elves, orcs, slimes, kijin, minds, bodies and relations |
| `selftest` · `version` | |

## Configuration

Configuration is YAML (or JSON), layered — each layer overrides the one before it:

| Layer | Where | Kept in git |
|---|---|---|
| built-in defaults | in the binary — everything on | — |
| global | `~/.config/isekai/` | no |
| project | `.isekai/` in the world | yes |
| project local | `.isekai/config.local.yaml`, and per-section `.isekai/<section>.local.yaml` | no (machine-only) |
| environment | `ISEKAI_CONFIG=<file>`, `ISEKAI_CONFIG_CONTENT=<yaml>`, `ISEKAI_MODEL`, `ISEKAI_DISABLE_PROJECT_CONFIG=1` (skip the project layers), `ISEKAI_PROVIDER_<NAME>_API_KEY` (built-in provider names only; a custom provider uses `apiKeyEnv`) | — |
| flags | `--model`, `--set key=value`, `--approve outward`, `--dry-run`, `--no-<feature>` | — |

`isekai config explain` shows every effective value and the file and line it came from;
`isekai status` shows every model, the guard, the sandbox and anything switched off.

### Keys and registry credentials

Providers prefer `apiKeyEnv` (name the variable, never the value). Where an env var cannot exist —
the estate's headless containers — the designated escape is `providers.<name>.apiKey` holding the
literal; every other field still refuses credentials in a file (headers, urls, unknown keys).

`--containered`'s registry pulls authenticate the same way: env-first, literal fallback —
`registry.containers: { usernameEnv, tokenEnv }` or `{ username, token }`. Docker receives the
token on stdin, never argv.

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
  baseURL: https://gateway.example.com/v1/   # the conduit front-door (hprd)
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
  gateway/muse-glimmer-30b: { tier: 1, contextWindow: 131072, price: { input: 0, output: 0 } }
  gateway/glm-latest:       { tier: 2 }
  gateway/kimi-latest:      { tier: 3 }
tools:                                      # external tools the toolbox indexes
  - { name: kubectl, description: "the cluster CLI", triggers: [kube, pods, deploy] }
packages:                                   # set in every shell command and in the container
  npm: https://artifactory.example.com/artifactory/api/npm/npm-remote/
  pip: https://artifactory.example.com/artifactory/api/pypi/pypi-remote/simple
  go:  https://artifactory.example.com/artifactory/api/go/go-remote
  tokenEnv: ARTIFACTORY_TOKEN               # let through by name — the agent can read it: use a read-only token
  env: { GONOSUMDB: example.corp }
containers:                                          # for --containered — images always pass by a docker vhost
  base: registry.example.com/debian:bookworm-slim   # the hub proxy (docker-remote)
  apt:  https://artifactory.example.com/artifactory/debian-remote
  # image: registry.example.com/team-example/isekai-runtime:1   # our local repo: pull a prebuilt runtime instead
```

For a private container registry, `docker login registry.example.com` first: docker's own login
answers for the pull.

**`models.yaml`** — the session's model and the three offices (Great Sage reads, Raphael gives verdicts,
Ciel drafts). A call is retried on 429 and 5xx (honouring `Retry-After`); when it still fails, or
fails outright on a transport error, the fallback is tried — never on a refusal. An *office* is what a
Court Body does for a call; a *rank* is its place in the tree (elf → orc → slime); a *creature* is
one named body — each can be given its own model:

```yaml
default: { model: gateway/kimi-latest,      fallback: gateway/glm-latest }
offices:
  great-sage: { model: gateway/muse-glimmer-30b, fallback: gateway/glm-latest }
  raphael:    { model: gateway/glm-latest,       fallback: gateway/kimi-latest }
  ciel:       { model: gateway/kimi-latest,      fallback: gateway/glm-latest }
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

`isekai config scaffold` mints a world's machine-local config set for you — `config.local.yaml`
plus one `<section>.local.yaml` per section, each with defaults and a comment per key; it never
overwrites an existing file, and `--check` answers what drifted between a world's local files
and this binary's canon (so a harness upgrade adapts, never surprises).

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
CI runs a no-cost smoke benchmark (`bench/smoke.sh`) on every commit against a fake
OpenAI-compatible server. A real run, isekai against OpenCode on the same gateway models, is
`bench/harbor.sh` — see [`bench/HARBOR.md`](bench/HARBOR.md).

**Reading honestly, first.** The two columns below are not a same-day A/B: on the left, the world's
newest grid (run `tb2-final2-024306`); on the right, OpenCode's worst reading tonight on the same
tasks. The pairing shows how far the loop lifts a floor the stock loop dropped — not a fairness
claim. The control's best tonight was 0.92 with its stock loop, lifted by fixes around it, never by
a loop change (terminal-bench-2 slice, 6 tasks × 2 attempts, kimi on the estate gateway, same gateway
all night).

| task | isekai · `tb2-final2-024306` (witness on · guts on · echo applied) | opencode (stock loop) — its weakest grid tonight |
|---|---|---|
| bn-fit-modify | 2/2 | 2/2 |
| constraints-scheduling | 2/2 | 2/2 |
| feal-differential-cryptanalysis | 2/2 | 2/2 |
| filter-js-from-html | **2/2** ← flipped | 0/2 |
| fix-git | 2/2 | 2/2 |
| gcode-to-text | 0/2 | 0/2 |
| **score** | **10/12 = 0.83** | **8/12 = 0.67** |

Per task, the two grids (mean per trial of 2 attempts; the guts' spoke count rides beside):

| task | world in/out | loops | guts spoke | OpenCode in/out | loops |
|---|---|---|---|---|---|
| bn-fit-modify | 0.11M / 14k | 15 | 3 | 0.12M / 3k | 13 |
| constraints-scheduling | 0.07M / 6k | 8 | 1 | 0.06M / 2k | 6 |
| feal-differential-cryptanalysis | 0.09M / 15k | 10 | 3 | 0.17M / 4k | 12 |
| filter-js-from-html | 0.71M / 47k | 26 | 1 | 1.31M / 15k | 28 |
| fix-git | 0.05M / 7k | 14 | 1 | 0.10M / 1k | 11 |
| gcode-to-text | 1.87M / 33k | 62 | 18 | 1.30M / 13k | 49 |

### Experimental loop features (off-switches, measured)

| feature | default | off with |
|---|---|---|
| `law.guts` | on | `--set law.guts.enabled=false` |
| `law.closingCheck` | on under bench, off live | `--set law.closingCheck.enabled=false` |
| echo penalty + witness admissibility | inside guts / closingCheck | ride their toggles |

Same 6 tasks × 2 attempts, kimi on the estate gateway, same cells:

| run | loop state | score | input / out |
|---|---|---|---|
| `tb2-final-165052` | everything off | 8/12 (.67) | 3.1M / 0.25M |
| `tb2-fixed-220431` | laws + ceilings off | 7/12 (.58) | 2.4M / 0.18M |
| `tb2-clean-010744` | + witness close + guts | **10/12 (.83)** | 12.2M / 0.32M |
| `tb2-final2-024306` | + echo + admissibility | **10/12 (.83)** | **5.8M** / 0.25M |

On tokens: the winning harness grid spent 5.8M in / 0.25M out against the control's 6.1M in / 78k
out.

agent-one's current loop needs **23% less input per point of score** than the control's weakest
day — same tasks, same model. (For honesty: the control's best reading tonight was 0.92 — its loop
never changed; lifts came only from fixes around it. gcode-to-text sits at the model's glyph
precision edge for every harness: probe runs read the flag twice and miss a 0/O glyph; refusals
are honest, not confused.)

The culture that came from the bench is in [`examples/world-rules.yaml`](examples/world-rules.yaml)
— two laws earned, kept as principles, plus the close-time falsification witness (`law.closingCheck`)
and the per-step value sense (`law.guts`) they motivated.

## License

MIT — see [LICENSE](LICENSE). Bootstrap 5 (embedded in the board) is MIT.
