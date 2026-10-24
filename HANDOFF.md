# Handoff — 2026-10-09 (wave 2) — zone width at birth + the single-prefix tree split

Who: opencode session, on Veldora's word: "implement all" (workstream B of `../todo-isekai.md`). Trigger: the human's zone-sizing law — "a slime holds a zone; it should not be too big, otherwise it will stress when new files need to be created."

## What landed

- **`app/found.md` (the founding brief)** — the Zone width law joins the team-deciding section: size a territory to what its keeper actually works — not so wide that every new file's creation lands on one owner, not so narrow that one change always crosses zones; prefer several precise claims over one wide dir. The complex level's founding session now carries the law.
- **`app/found.go` (sketchColony)** — a big flat dir (no sub-areas) no longer births ONE slime holding the whole tree: direct source files chunk into bounded file-set slimes (8 per chunk, at most 6, the sixth absorbs the rest — `slime-<area>-a/b/…`), each reporting to the area's orc. Small flat dirs (≤8 files) keep the single-slime birth. `dirFiles` is the chunker's input (direct source files, sorted).
- **`world/relief.go` — the single-prefix tree split**: until now a one-prefix zone (`app/` and nothing else) could never split — the council's relief ladder split only ≥2-prefix zones, so a wide single-dir zone could only RETIRE. Now `splitByTree` halves the prefix by what actually lives under it (sub-dirs + direct files, junk-dirs skipped), each entry becoming a young one's claim — nothing under the prefix `.

## Verified, not changed (the audit items that read the world's own lanes)

- New-file ownership: `Owner()` matches by territory prefix, so a new file under a claimed zone routes to its holder at the gate — the sizing law above is what keeps that one keeper from being everyone's bottleneck.
- The init chain law (`app/cli.go:362`): `onto.Housekeep` runs after the sketch write for medium/complex (light is exempt — law only, nothing to chain) and heals the mechanical census classes at birth.

## Caveat — the derive lane rides the world's own commission

The roster parses Territory through `onto/world.go`'s derive, which is mid-surgery (the census class-(f) seat-fix slice sits dirty on the tree; its `inZone` admits `.isekai`/`src` roots and seat-eponymous claims). The tree-split test therefore pins the mechanic with a `src/`-rooted zone (admitted under every genome state landed so far); an `app`-rooted single-prefix zone re-verifies when the world's class-(f) /a app/b]` / `[app/root.go]`).

## Tests

`app/found_test.go` — `TestSketchColonyChunksFlatDirs` (a 12-file flat dir chunks 8+4 into `slime-flat-a/b` reporting to `orc-flat`; a 2-file dir stays one zone). `world/relief_test.go` — `TestCouncilReliefTreeSplit` (single-prefix zone with a real tree splits by it, halves' union is the whole tree, ownership rides with the young, ledger `split-tree` row; the existing retire test unchanged — a zone with nothin nly the pre-existing `TestDashGateAnswerOverHTTP` commission.

Wave 2 of `../todo-isekai.md`. All three waves now landed; the todo's remaining derive-side item is the world's own census commission.

---

# Handoff — 2026-10-09 (wave 3) — the territory consult arms + the dispatch card

Who: opencode session, on Veldora's word: "continue" (waves C+D of `../todo-isekai.md`). Trigger: the audit's finding #1 — bodies re-derive "who owns what" from scratch on every boundary hit; the Oct-8-9 chain (drain → arbitration → 3 dispatches) burned a day on exactly that, and `ask` reaches the human only.

## What landed

- **`world/territory.go` — the consult arms**: the default out-of-territory refusal now names the foreign soil's owner (any rank — `(name, rank, domain orc)`) or the unrouted Genesis line, and carries the three arms for the parent: **widen** (grant the claim — the README Territory grows, the doc pairs in-turn), **birth** (mint a creature for that zone), **route** (dispatch the owner's line). The refusal IS the The sideways rank-mate branch (dispatch the owner directly) is unchanged; the throne branch unchanged.
- **`world/territory.go` — `DispatchCard(rel)`**: the dispatcher's one-read answer — owner, rank, **strain from the ledger shelf** (live state, not re-derivation), domain, territory, and the arms. Unrouted soil answers with the Genesis trigger line.
- **`app/worldtools.go`**: the `onto` tool's path query serves `DispatchCard` (was: owner + doc only) — elves and orcs ground do-it / birth / route decisions on one read; description updated.

## Tests

`world/territory_consult_test.go` — `TestTerritoryRefusalCarriesConsultArms` (coder-writer foreign soil names the owner + the three arms; unrouted soil carries the Genesis line), `TestDispatchCard` (strain row replayed from the ledger shelf appears in the card; unrouted card). Existing refusals stay compatible: `TestTerritoryRefusal`, `TestCustomRanksReplace` (their pinned substrings — "outside …'s territory", " fails only the pre-existing `TestDashGateAnswerOverHTTP` commission.

Wave 3 of `../todo-isekai.md` (C: consult arms, D: dispatch card on the onto tool). Wave 2 (birth/zone-sizing) stays sequenced behind the world's own onto class-(f) slice. The running world gains this at its next restart — the human's act.

---

# Handoff — 2026-10-09 — the guts' bench-thrash arm + the `pulse` tool + the self-state prompt law

Who: opencode session, on Veldora's word: "can you implement the todo?" (workstream A of `../todo-isekai.md`, the plan staged from the audit `instruments/audits/2026-10-09-slime-onto-ownership-time-loss.md`). Trigger: the 084230 drain — a slime hit one bench 63 times across 20.9 min (34 red, 18 piped-unclassifiable), and every health guard stayed silent: the red-pin exemption scored each visit neutral (`loop/gut heir status, and doomLoop — default-disabled in `config/defaults.go` — never armed.

## What landed

- **`loop/guts.go`**: `Guts` grows `BenchThrashVisits` (default 6, `<0` off) and `DigestEvery` (default 10, `≤0` off). Every bench visit (`go test`/`go vet` through a shell tool — `gutsBenchCall`, which also detects the pipe that mutes an outcome) lands in a per-bench tally `Session.gutsBenchHits` (`benchHit{visits, red, green, muted}` — `gutsBenchOutcome`: red beats green beats muted; a piped bench with no surv bench costs ≥1 with why `bench thrash — N hits (R red / G green / M muted)` — independent of the echo penalty's write gate (the exact 53ca blindspot). The red pin keeps early visits neutral (TDD's signal — `TestGutsRedPin` semantics unchanged). A spoke note whose act carries no why of its own now names the run's most recent losing ground (`gutsLastWhy`) — the floor-crossing write between thrash hits no longer spea
- **`loop/guts.go` + `loop/loop.go`**: `Session.Pulse()` — the acting body's own run state (steps · turns · wall · ctx zone · momentum · benches with red/green/muted · writes · failed acts); `digest` rides one `[pulse: …]` line on a tool result every `DigestEvery` steps. `env.Self = s.Pulse` wired at the act mint (`loop/loop.go`).
- **`tool/pulse.go`**: the `pulse` tool (class read) — serves `Env.Self`; nil-Self contexts (dispatcher previews) degrade to a note, never an error. Registered in `app/worldtools.go` + seated in `app/shelf.go` `worldToolNames` + `config` (`Tools.Pulse`, `Builtins["pulse"]`, defaults on).
- **`world/prompt.go`**: `PromptOptions.SelfState` (default on; `app/world.go` wires it to `law.guts.enabled`) — the prompt block names the pulse tool, the guts/doomLoop note semantics, and the check-loop law ("repeated reds on one bench with no green is a check-loop, not TDD; never pipe a bench's output through a filter that hides its exit status").
- **`config`**: `GutsCfg` grows `benchThrashVisits`/`digestEvery` (law.yaml ladder — the world tunes the sense as it evolves); `LawGuts` maps them; **doomLoop default flips to enabled, threshold 3** (H4 — new worlds inherit the armed ladder; the world's own law.yaml already enabled it).

## Tests

`loop/guts_test.go` — `TestGutsBenchThrashSpeaks` (the strict 53ca shape: a write between every bench pair so the echo penalty provably stays silent; visits 1-5 neutral red pins, 6+ priced, the run speaks and the body hears `[guts: … bench thrash …]` inside the tool result), `TestGutsPipedBenchCounted` (muted never green; FAIL surviving the pipe still red; read tool never a bench), `TestGutsDigestRidesEveryNSteps — digest served verbatim, nil-Self degrades. `world/prompt_selfstate_test.go` — the block on/off. Verified: `go test ./loop ./tool ./config ./world` green; `go test ./app` fails ONLY `TestDashGateAnswerOverHTTP` — the pre-existing open commission in `.isekai/slab/gate-ui-approval-brief.md` (403 + the predicted TempDir flake), not this change. `gofmt` clean on touched files.

Companion doc routings (no code): `.isekai/orc/world/README.md` re-dispatch law (elf-verified disk-truth envelope, no live reasoning inside briefs) + `.isekai/slime/onto/README.md` probe law (never pipe a bench through a status-hiding filter; instrument before loop-cutting; pulse before re-running a bench hit ≥4 times) — both creature docs ride on disk (gitignored / the world's own dirty slice), never in this co

Wave 1 of `../todo-isekai.md`. Waves 2-3 (birth/zone-sizing audit, slime→orc dispatch consult, dispatcher digest) staged there. The running world gains all of this at its next restart — the human's act.

---

# Handoff — 2026-10-05 — boot heals the long-memory index (`memory.long.indexOnBoot`)

Who: opencode session, on Veldora's word: "start implementing it" (the top fix routed by the audit `2026-10-05-overnight-slice3-family-audit.md`, F1). Trigger: the 2026-10-04-05 family ran a full 20 h on a 24 h-stale index — 8,564 stale-index nag rows across 4,295 recalls of the family; the 10-02 escalation (floor-up: automate the reindex) had never landed as code.

## What landed

- **`config`**: `MemoryLong.IndexOnBoot` (`indexOnBoot`, default `true` — defaults.go). Older binaries ignore the key; worlds can opt out with `false` in a `memory.yaml` layer.
- **`app/app.go` (`New`)**: when long memory is enabled and this is a session-carrying boot (`!Quiet`, world present), the boot loads the index and, on any staleness hole (`LoadIndex` why or `IndexHoles`), rebuilds it once, off the recall beat, before any body runs — success/failure both surface as hole notes (`boot re-indexed N memories` / `boot re-index failed`). Recall stays rebuild-free (`rebuildOnStale:false emantics; staleness now dies at the boundary instead of nagging all run).
- **Test**: `app/boot_index_test.go` `TestBootHealsLongIndex` — boot 1 builds a missing index + notes it, source rewrite + future Chtimes rots it, boot 2 heals staleness clean (0.10 s).
- **World config truth** (F2, comment-only): `.isekai/config.local.yaml`'s `contextWindow: 100000` comment rewritten — the pin is shadowed by per-model rows in `~/.config/isekai/providers.yaml`, the operative budget is law `budget.contextTokens 200000 / stressTokens 180000` (rimuru rode 166.5k NEAR with the pin in place); explicit numbers everywhere stay so the drain's fraction arm never rides `auto`.

Verified: `TestBootHealsLongIndex` green; `go test ./config ./memory ./world ./app` green (the `TestDashGateAnswerFlow` full-suite flake fired once and passed isolated — the documented pre-existing cascade). `gofmt` clean on touched files; pre-existing drift elsewhere untouched.

Binary re-installed via `make install` → `~/.local/bin/isekai`; the relaunch boots on a fresh index (manual reindex run as the wave's last act, per the mind's rule 3).

---

# Handoff — 2026-10-02 — body cards law: the world carries its own — `.claude/agents/` mint retired (both distributions: isekai + agent-one)

Who: opencode session, on Veldora's word: "no it should be inside isekai not there — change stp".

## What landed

- **Law change** (`ISEKAI.md`, `app/law/isekai.md` embed, this world's `.isekai/isekai.md`): the creature doc IS the body card — `{{WORLD}}/<rank dir>/<name>/README.md` is the body. The `.claude/agents/<name>.md` mint line is gone (Veldora decree 2026-10-02 cited in the text).
- **`app/found.md` (the founding prompt, embedded)** follows: nothing written outside the world; bodies are no longer a separate `.claude` card, keepers are "the creature whose rank is Keeper", not "a file with mode: all".
- **Stale cards removed locally**: `.claude/agents/{elf-isekai,kijin-release,orc-bench,orc-engine,orc-host,orc-world}.md` (minted 2026-09-30 under the old law, gitignored host surface) — deleted; `.claude/` dir left for future foreign-body imports (`shelf_test.go` still pins import-parity read).
- Discovery unchanged: `discover.Agents` still scans `.claude/agents` etc. because import-parity reads it for foreign worlds that DID mint there — reading stays, minting stops.
- Binary rebuilt + installed.

Verified: `go test ./app -run GateAnswer -count=3` green; board's 11 fails + app's `TestTUIInterruptAndQueue` are the chronic pre-existing reds (confirmed against d471e7a worktree earlier today), uncoupled from this change.

---

# Handoff — 2026-10-02 — tron-cockpit drift cured (UI-system gate back to green)

Who: opencode session, on Veldora's word: "go on correct please". Trigger: after the tron wave (33e9c13), `TestDashboardHoldsTheUISystem` failed — tree.css header drifted from its own reads, `#tree-note` styled an id, and `board/web/palettes.css` had the tron values while the foundation copy was still aurora.

## What landed

- **`ui/assets/foundation/palettes.css` promoted to the tron copy** (the law is foundation→boards, not the reverse — the new values live in both now; `ui.Scan/Lint` test passes).
- **`tree.css` header re-synced** to the tokens the styles actually read (--surface/--surface-2 out; --accent-hero, --focus, --focus-width, --radius-s in).
- **`#tree-note` id styling → `[data-note="tree"]`** (dash.html attr + dash.js selector follow).
- **TempDir flake cured per the commission's own contingency** (`TestDashGateAnswerFlow` ×5 stable): the in-process flow test returned while the turn still wrote into `.isekai/instruments/` — the test now waits for the turn to finish before yielding to teardown. NOTE: the commission guessed "close the board in the test"; the actual flake lives in the in-process flow test (NoBoard), so the wait is the correct seam,

Verified: `go test ./app/ -run GateAnswer -count=10` green; `go test ./ui ./board` — no NEW failures against the d471e7a baseline (11 board fails + TestTUIInterruptAndQueue pre-exist, untouched).

Binary re-installed via `make install` → `~/.local/bin/isekai` (embed.FS carries the sync'd palettes + fixed tree.css).

---

# Handoff — 2026-10-02 — degenerate-query recall silence

Who: opencode session, on Veldora's word: "rebuild binary and commit so i restart". Trigger: throne session showed `recall: memory/toolbox … no content words after stop-word removal` on menu-answer turns (`yo`, `1`, `B`).

## What landed

- **Recall beat exits silent on zero-content-word asks** (world/hooks.go): after the existing `!Enabled || trimspace-empty` guard, a `len(memory.Tokens(ask)) == 0` early-return skips the memory+toolbox ranking pair. No anchors, no tools, no holes — the journals stop carrying two "nothing to rank" lines per greeting/menu turn. Root cause notebook: `team-example/isekai/.isekai/instruments
- Pin test covers `yo`,`1`,`B`,`!` running fully silent (world/world_test.go).
- Behavior on real asks unchanged; stale-index nag intact (fix for junk anchors remains `isekai memory index` at idle — still pending from morning audit).
- Binary rebuilt + installed: `make install` → `~/.local/bin/isekai`.

Committed surgical: only `world/hooks.go` + `world/world_test.go` hunks of this change — the surrounding WIP (board/app pair, gate signature thread) untouched. agent-one mirroring NOT done (same standing decision as the 10-01 wave — offered, awaiting word).

---

# Handoff — 2026-10-01 — the audit wave: machine config + three latency cures, committed beside a live WIP

Who: opencode session (workspace elf frame), on Veldora's word: "audit it" → "fix other things, validate isekai" → "commit, and leave word for the next launch". Committed in the commit that carries this entry; agent-one mirroring is NOT done (offered, awaiting word).

## What landed

1. **The recall beat now honors `memory.long.rebuildOnStale`** (world/hooks.go, app/world.go, DefaultHooks true): in a writing session the index goes stale nearly every model call (the end-gate appends to log.md, an indexed source), and the stale path rebuilt the WHOLE long-term index inline on the call's critical path — the visible "remembering" stall. With the knob false, recall rides the last g . (Cause of the orc Veldora watched thinking-and-remembering per paragraph, 2026-10-01.)
2. **`law.gate.verifyTimeout` knob** (config/schema.go Gate + defaults "120s" — zero behavior change; wired into GateOptions.Timeout in app/world.go): per-creature verify commands were a hardcoded 120s wall clock each, sequential — one slow verify held a write turn's tail for minutes.
3. **The tree stamp never walks dependency/build dirs** (loop/snapshot.go + loop/snapshot_test.go): node_modules/.venv/venv/__pycache__/target/dist/.next/.turbo/.cache skipped at any depth (lstat storms per turn-start/per-write/turn-end stamp); `vendor` and friends stay watched — tests pin both sides of the law.

## Machine config wave (NOT committed — ~/.config/isekai/, backups at *.bak-2026-10-01)

- law.yaml: `stress { escalateAt: 3, maxConsults: 2 }` — a struggling court's worst-case council fan-out drops ~9→4 child runs before the human is asked.
- tools.yaml: `output.maxBytes` 51200→24576 — tool-result history inflation halved; overflow still dumped full to tmp/tool-out (dumpDir, keepDays 7).
- Untouched by law: budget steps/minutes stay -1 (Veldora's 2026-09-28 order, uncapped); providers.yaml contextWindow stays pinned 100000 — effective drain is min(trigger.tokens, 0.85×window) = 85k (compact.Trigger.Threshold math).
- `guts` needed nothing: on by default, no layer disables it.

## Earned lessons (Nature 3 — read before touching this tree again)

- **This tree carries a MID-FLIGHT WIP that is NOT this session's** — the web-approve bundle (webBus.answerable from birth, Answer first-wins race, dash wiring removed; app/web.go, app/dash.go, untracked app/zz_gate_flow_test.go) and the hardlaw-decree bundle (world/gate.go +92, territory.go throneNeverAuthors, rank.go, court.go, world.go + decree/pair/seat tests). NEVER `git add .` or `git add -A` hunk, gofmt, add, restore).
- app+board package reds are PRE-EXISTING WIP drift (stash-verified both ways): board fixtures still expect old world seeding ("no creature on disk"), TestDashGateAnswerFlow passes isolated (0.09s) but 10s-times-out inside full-app runs (shared-state interference). TestTUIInterruptAndQueue remains the known timing flake.
- **Activation law**: `law.gate.verifyTimeout` must NOT be written into any live config until a binary built after this commit boots — the strict loader rejects unknown keys. `memory.long.rebuildOnStale: false` is safe to write any time (old binaries know the key; only the shelf tool honored it before this commit).
- Audit facts worth keeping: guard/classify/pattern gate = pure regex, zero LLM per tool call; MCP is persistent (boot cost only); TUI re-renders the growing tail paragraph through Glamour per token (cosmetic churn, NOT fixed); openai-provider retry backoffs (1s/2s/4s) and the interactive session's in-flight wall clock stay hardcoded (deferred, listed for future waves).
- gofmt: the WIP left schema.go drifted tree-wide; only the files this session touched were normalized.

## Verification

`go build ./...` clean; gofmt clean on touched files; `go test`: loop/world/config/memory/tool green; the snapshot regression pair green; stash-isolation proved all app+board reds pre-existing.

---

# Handoff — 2026-09-28 (eve) — waves 3–8: the court answers back, stress, the container's
trust, background rides, the pane pipes, and the forge

Who: rimuru (the throne, machine-global), on bench-forge, work mirrored 1:1 into agent-one.

## Landed this session (top first)

- **The forge (wave 8).** Wire verb `@FORGE tool:<name> | <summary> | <argv>` /
  `skill:<name> | <text>`; drafts ride reports up (`tool.DispatchReport.Forges`), land at
  the record beat as shared notes (kind colony, tag forge), mint only at the desk:
  `/forge` lists, `/forge mint <n>` — a skill becomes a world note, a tool a live-reloaded
  `tools.custom` patch; minted marks are append-only. Tests: wire round trips both trees,
  world/council_test.go, app/forge_test.go.
- **Bodies pane pipes (wave 6):** ctrl+t → i types into a body's inbox (FIFO at its next
  turn boundary); paste batching forwards to the pipe's input now.
- **sessionBackground (wave 5):** `tools.dispatch.background: true` + per-call override;
  enabled on this machine (`~/.config/isekai/tools.yaml`).
- **trustContainer (wave 4):** `law.humanGate.trustContainer: true` — inside
  `TEMPESTCONTAINERED=1` every class pre-approves; guard denylist still refuses first.
- **The council (wave 3):** `law.stress.escalateAt`/`maxConsults`; warrants @INSTRUCT/
  @SPLIT/@ESCALATE; verdicts read off the wire, never the loop status.
- **Zombie court cards fixed** (observe hoist in `app/tui.go` + turn-end sweep +
  frozen elapsed in `tui/model.go`), mirrored in agent-one's panes.
- **Pin law swapped (wave 2):** `creature → rank → office → task → default`; office hats
  ride a body's office (findings→great-sage, verdict→raphael, draft→ciel); a creature's
  pin overrides its office.
- **Fallback ride on 404:** `retryable()` treats HTTP 404 as missed-roll (gateway model
  group names rot); live model IDs for the gateway: glm-5.**3**,
  kimi-k3, muse-glimmer-30b.

## Open

- **`TestTUIInterruptAndQueue` is timing-flaky in `app/`** (the cancel-vs-queue leg; the
  backing law test `TestCancelDraftsQueue` in `tui/` passes standalone).
- The bench gate rule-write trace at line 474 still ungraded ("edit produced an unreadable
  document" on `.isekai/config.local.yaml`) — repro then harden.
- **Vocabulary guard** on the sibling tree: never write isekai-words (colony/court/world,
  Tempest names) into agent-one files — `TestNoForkedVocabularyLeaks` (needs `-tags leak`)
  watches.

---

# Handoff — 2026-09-28 — the config plumbing wave

Who: rimuru (the throne, machine-global), on Veldora's word, worked from bench-forge and
landed here too. What a fresh session in this dir needs to know.

## What changed today (this repo)

1. **The local layer learned one file per section** — `config/load.go`:
   `localSections()` folds `<section>.local.yaml|.yml|.json` (beside `config.local.yaml`,
   each file = the section's VALUE, no wrapper key; `guards.local.*` aliases `guard`)
   into the local layer, after `config.local.yaml`, so it wins there. Mint-if-absent.
   Duplicated section in the local layer = load error ("one file per section"), same
   rule as section files. `config.local.yaml` itself is exempt by name.
2. **`isekai config scaffold [--check]`** — `config/scaffold.go` + `config/cli.go`:
   mints a world's machine-local set (config.local.yaml + providers/models/law/budgets/
   permissions/guards `.local.yaml`, defaults + a comment per key), never overwrites,
   appends `*.local.yaml` to the world-dir's `.gitignore`; `--check` writes nothing and
   names drift (unknown sections/keys as `@?`, un-minted sections as `@F`), exit 1 on
   drift — the sync-to-canon knob: re-run after upgrades.
3. **Docs**: `README.md` (scaffold paragraph under Configuration; layer table names the
   per-section local files), `docs/canon/config.md` (one-file-per-section ↔ local mirror).
4. **Tests**: `config/config_test.go` `TestLocalSectionFiles`; scaffold coverage lives in
   the CLI selftest block (mint → loads → kept → drift → clean), `config/cli.go`.

## Earned lessons (Nature 3 — read before touching config again)

- A **section file holds the section's VALUE, bare** (`providers.yaml` starts at `gateway:`);
  a **full config file keeps the nested key**. Mixing them fails validation loudly
  (`providers.providers.type`).
- `--check`'s known-key set is `defaults keys ∪ Config struct json tags` — legal keys like
  `rankSet`/`ranks` ship no default; flagging them is a false drift.
- `resolveScaffoldTarget` must default Env/Home/Cwd exactly like LoadWith, else a world
  entered *through* its world-dir mints into the void.
- Merge law: an empty list/map NEVER clobbers a lower layer — inert scaffolds are safe.
- Pre-existing env-flaky tests on this machine (not ours): `tool` TestBashOnShell /
  TestClassifyCommand, `tui` board label tests — fail standalone AND with stash, both ways.

## Verification

`go vet ./config/` clean · `go test -count=1 ./config/` green · binary installed
(`make install` → ~/.local/bin/isekai) · live: bench-forge (board seat 7412 pinned in
`.isekai/config.yaml`; drain-before-wall compaction; `config check` PASS, origins attribute
per file) · agent-one world (scaffold minted + permissions moved to permissions.local.yaml;
`config check` PASS; `--check` `@E 0 drift`) · global layer split into
~/.config/isekai/{providers,models,law,permissions,guards}.yaml (git push denied via
permissions.yaml) with a `config.yaml.pre-split-2026-09-28` snapshot kept.

## Open

- agent-one's rank policy (`rankSet: extend`, founder mounts, divine bonds) sits in a
  machine-LOCAL file — world canon in a machine layer; Veldora to say whether it promotes
  to that world's tracked `config.yaml`.
- `~/.isekai/` shared notes carry today's entries (tags: harness-era, harness-config).
- Pre-existing env-flaky suites above would love an owner one day.

## Wave 2 (same day) — the plan deck

Veldora's order (the opencode todo pattern, native to the rank chain): new world tool `plan`
(app/worldtools.go; per-body deck under `<world>/tmp/plans/<body>.json`; items carry `who`).
Seats opened through the whole spine, in order: `config/schema.go` Builtins+Tools struct field,
`config/defaults.go` tools row, `app/shelf.go` worldToolNames, `world/rank.go` AllTools += plan
+ elf Job (write deck → map slices to domains → dispatch, orcs in parallel by background) +
orc Job (verify written before the slime is ordered). app/plan_test.go covers the deck file
round-trip. The earned lesson: a builtin tool needs its seat in FOUR places — Builtins map,
Tools struct (reflection: missing field = nil deref at checkTools, learned by panic), defaults
row, shelf names list. world's Job text is where rank BEHAVIOR for tools is coded.

## Wave 3 (same evening) — the zombie Court cards, killed in code

Veldora's live chain test (rimuru → elf-colony → orc-experience → slime-ui-design, planted
anti-cheat held end-to-end) exposed the card flow's hole: the orc card showed `DONE · 15m0s` —
the budget clock's own axe — and the bodies pane still read "elf · thinking" over finished
work. Root cause: `tuiHost.observe` seeded/landed Court cards only for the SESSION's dispatch
steps; a Court's own steps arrive here as its PARENT's steps, so a nested Court closed "done"
at the loop with no report ever reaching its card. The card stayed in `m.courts`, `Render`
recomputed its elapsed every frame, and the body pane kept the last mid-run EvState forever
(a Court's terminal "done" never rides EvState — `h.state` routes it to EvCourt or drops it).

Coded fixes (never prompted):
- `app/tui.go` observe(): the dispatch seed (start → courtAsk) and land (end → courtReport)
  hoist above the session-only branch — every body's dispatch step now closes its child's card.
- `tui/model.go` Render(): a done Court's elapsed is frozen (no per-frame recompute).
- `tui/model.go` finish(): the turn-end sweep lands every done-but-unlanded card exactly once
  (covers failed/unparseable reports too — `wire.ParseReport` at the step end is the gate).
- `tui/model.go` EvCourt: the body pane's log state closes with the card (both branches).

Tests: `tui.TestCourtZombieCardLandsFrozenAtTurnEnd` (frozen clock + one tombstone + pane says
done + no second tombstone), `app.TestTUINestedDispatchStepsLandTheirCourts` (seed/land at
depth, parent log intact, rimuru-level shape unchanged, no generic tool blocks for dispatch).

Verified: `tui` and `app` suites green — EXCEPT the pre-existing break below; binaries
re-installed (`make install` → ~/.local/bin/isekai).

### Open (same evening)

- **`app.TestTUIInterruptAndQueue` red at wave-state, green at HEAD (ef837d0)** — diagnosed as
  this wave's regression, NOT today's: with the queue fix pending, an esc-interrupted turn whose
  bash step survives cancel long enough runs one more drain and eats the queued line into the
  dying context; the turn-end drain then finds rimuru's inbox empty. Failing shape: footer stuck
  "1 queued", verb stuck "interrupting…", "↻ continuing" never prints. Owner: next session —
  the cancel-vs-queue invariant needs its own care.
- agent-one mirror of everything above landed in its own HANDOFF; its app/tui/workspace suites
  are green (its interrupt test passes — different assertion shape: it re-types the next line);
  its HEAD-identical env set still fails on this machine (board/usage/team/log pages, SSE).

## Wave 4 (same evening) — the pin law + global model routing

Veldora's words: "if they are set in config it forces the creature to use that model, otherwise
it chooses the best hat for the task." The old resolver ran creature → office → RANK → … with
the office ahead of the rank — a race pin could never force. Fixed in code:
`config.ResolveModel` now resolves creature → rank → office → task → default (mirrored in
agent-one as member → rank → role → task → default). A collision row in `TestModelsResolution`
locks it. Global routing configured per Veldora's matching: the cheap class reads+writes
(muse-glimmer-30b = great-sage hat + slime/zone pins), verdicts ride the mid court
(glm-5 = raphael/judge + orc/domain/kijin/service/auditor pins), the colony's thinker
is the strongest (kimi-k3 = ciel/drafter + elf/coord pins), and the binary's own tasks
match: gate+drain on glm-5, log+bench on muse-glimmer-30b — all fall back up the ladder.
Files: ~/.config/isekai/models.yaml · ~/.config/agent-one/models.yaml (machine layer; nothing
world-tracked moved). Validated live: `isekai status` in bench-forge lists every rank pin with
file:line origins; agent-one's ledger via `config explain` (its bare-world status prints hold
today). Binaries re-installed. Leftovers, deliberately untouched: the shared-agents agent-one
world's config.local.yaml `ranks.principal_coordinator.ascendsWhen` finding (wave-1 open: the
underscored key misses the builtin's promotesFrom — ascension law is Veldora's), and the rank→
jobs→actions→office schema idea Veldora sketched for a future leg. Also: bench-forge's world
rides the new globals with zero world-local models (its models.local.yaml is `{}` by canon).

## Wave 5 (same evening) — the council: Veldora's stress law, coded

Veldora: "more than five thoughts and the body is stressed — the hook asks its manager whether
context is missing; the manager rewrites, splits, or escalates (≥3 consults then to the human)."
Landed:
- `loop.Result.Strains` — the turn's governance sent-backs (failed acts standing + gate +
  wire re-asks) surfaced for the ledger.
- `world.councilLedger` — per-body strain ledger in the dispatcher; Fail/Checkpoint runs and
  self-verdicts FAIL/ESCALATE accrue; DENIED/INTERRUPTED never consult (the human's word and
  the clock already answered).
- the warrant: `@? council warrant — …` stamped into the failing Court's report Text, verbs
  @INSTRUCT / @SPLIT / @ESCALATE with sideways counsel named lawful; consults bounded
  (maxConsults, 3); past it rep.Status hardens @S ESCALATE + Failed. A clean landing clears
  the ledger.
- config: `law.stress.escalateAt` (default 5; 0/-1 = the council off) + `law.stress.maxConsults`
  (0→3); schema → defaults → plumbing (CourtOptions ← cfg) all seated; the scaffold's
  law.local.yaml body carries the rows; canon config.md has the row; test:
  `world/council_test.go` — eleven dispatches: silent below 5, warrants 1/2/3, ESCALATE at the
  fourth, clean-run reset, silent after.
Earned lesson (second docking): loop Status ≠ the body's verdict — a Court reporting @S FAIL
ran its lifecycle DONE; when judging a court, read the wire's first @S, never r.Status.
agent-one mirror: `workspace.boardLedger` ("@? board warrant"), `policy.stress` (same keys),
workspace/council_test.go, repo docs/agent-one/config.md row.

## Wave 6 (same evening) — the throne stays free; the pane pipes

Veldora: "when rimuru asks the elf, rimuru should stay available; I can re-ask something else;
ctrl+t to watch, push a line in its pipe." Landed:
- `tools.dispatch.sessionBackground` (default false in canon; true on this machine via
  ~/.config/<dist>/tools.yaml both harnesses). A session-level dispatch without an explicit
  background word rides background (Background must be allowed first — a default that refuses
  is no default). The tool's `background` arg is now tri-state (*bool) so a commission may
  still force the foreground. A Court's own dispatches stay foreground (depth law).
- the bodies/agents pane: `i` pipes a line into the selected body's inbox (Enqueue — next turn
  boundary, never mid-turn, the B contract); done bodies refuse at the host (Pipe on the Host
  interface; the "added Host method" seat lands on every tuiHost + fakeHost).
- `law/policy.humanGate.trustContainer` (default false): inside the harness's own runtime
  container every class is pre-approved (sandbox = membrane); the guard denylist still refuses
  first. Mirror in agent-one incl. its no-leak guard repointing (the court word stayed out).
Tests: tool/dispatch_stamp_test.go (background default), app/container_test.go (trust),
tui/model_test.go (the pane pipe x2 trees).
Open from Veldora's last live log: glm-5 404s on the gateway (provider group —
gateway-side, outside creature territory; the global pins ride it for raphael/judge-class work:
watch or repin); the gate's "don't ask again" rule-write failed once mid-run ("edit produced an
unreadable document" on config.local.yaml) — a rule-write leg to own next; isekai's
TestTUIInterruptAndQueue break stands red (cancel-vs-queue leg).

---

# Handoff — 2026-10-02 (late) — the Wave landing: harness laws died-in and the world carries itself

Who: same opencode session continuing through the audit arc's wind-down. Veldora: "go on".

The whole session's findings convert to silhouette before the ribbon:

## New lands (evening wave)

- **`court_vita.go` + Dispatcher wiring causes pre-empts** — stress-deaths memo (one-time, advisory only… a soft `Veldora-cassette.txt` line rides the dispatch wire: "X burned its clock on lane Y; next-lfi ladder says Z"), plus stratTargets refusal-listing Law-2. Audit coverage
- **Timeout messages name the corpse** (018) — sole property of your dawn-of-Verdict period: diesByClock jots `model openai/<m> on provider <p> did not answer within N-minute court budget`.
- **Territory hard-law flipped to true in-world** (019): `law/territory.enabled: true, throneNeverAuthors: true` — the territory gate fires live, not just at sealing.
- **The world carries its config** — `.isekai/{models,law,memory,guards,permissions,tools}.yaml` _(was `~/.config/isekai/*.yaml`)_; la couche globale archive.
- **Model lanes in-world annonce** (Veldora swap) : tous glimmer-glm-latest, log/bench-kimi-latest — voir `.isekai/models.yaml`, borning source. The slime does not spawn on glimmer again.
- **Go 1.26.8 install via Artifactory+netrc** — `~/.netrc` (tamponné 600), download manuel du toolchain zip (`toolchain/@v/list` worked; `dl.google.com` bloc-proxy), binaire final installé à `~/.local/bin/isekai`.

## Never landed but the WIP is captured

- The slime's TS files (board/src + board/web/components/files) remain **untracked at working tree** — stash@{0} is the safe copy. Its dash.js was never committed elsewhere; the current tree's dash.js is still the 49425-byte (index) version.
- 010-011 (zone gating + slime split/retire), 017 (TUI court list ledger), 021 (read memo) — route card remains at `.isekai/instruments/audits/2026-10-02-afternoon-wire-audit.md`.
- 012 (junk recall anchors) is un-fixed until next session's `isekai memory index`.


---

# Handoff — 2026-10-03 — background dispatch unlocked at depth ≥1 + TUI marker hardening

Who: opencode session, on Veldora's words: "make possible background dispatch since each has his own territory but i think its blocked by config" + "the Maintenance/Evolution marker messages fuck up the tui". UNCOMMITTED — human commits.

## What landed

- **Background Courts at any depth** (was dead canon): config already said `tools.dispatch.background: true`; the block was code — `world/court.go` gated `Background && b.Depth == 0` and the only wake channel was the session's queue, so a deeper body's background child would have orphaned its report on the throne (24 refusals in one elf run this morning). Now: `Build.backgroundWake(a tcher's own mid-turn inbox (`app.Court.Send`, the `/send` channel), `CourtOptions.Noted` does session bookkeeping (@U seats → world holes, court card closes) at every depth; Deliver falls through to the session queue when the dispatcher already closed. Refusal message now honest (`or no inbox delivery wired at depth N`).
- **TUI marker hardening**: `tui.textSanitize` strips U+FE0F at both render boundaries (`Theme.Markdown`, `wrap`) — x/ansi counts ⚙️ at 2 cells, terminals on text-presentation fallback draw 1; every width-padded marker line then lied. Renders pin to measurable cells now.
- **Speech law note** in `.isekai/memory/shared/notes.jsonl` (tag `speech`, kind `law`): the room's outer-report markers never ride in-world speech. Index reseeded (261 memories).

## Evidence

- Tests: `TestBackgroundWakeDepthSemantics` (world), `TestMarkdownStripsVS16` + `TestWrapStripsVS16` (tui). Suites: `./world`, `./tool`, `./tui` green; `./app` red ONLY on `TestTUIInterruptAndQueue` (stash-verified identical on clean baseline — chronic env-flake).
- Files: `world/court.go` (CourtOptions +Deliver/Noted, backgroundWake), `app/app.go` (noted/deliver wiring), `tool/dispatch.go` (refusal message), `tui/{blocks.go,markdown.go}` (textSanitize), new tests, notes.jsonl + reindex.

## Watch next session

- An elf actually dispatching `background: true` — report must arrive as `[court <child> reported]` inside the ELF's own turn, never the throne's queue, and the spawned card must close (`Noted`).
- Marker echoes in throne speech should die out as the law note is recalled; if they persist, scrub the two 2026-10-01 entries in `.isekai/memory/short/rimuru.jsonl`.

---

# Handoff — 2026-10-03 — memory instrument: `clean` cure arm + `review` three-tier lens

Who: opencode session, on Veldora's words: "make it possible to clean memory via cmd … when it review its memory it's a path to genesis or minds or craft a skill … we have 3 types of memory, each one has different data, from each one we can get different intelligence".

## What landed

- **`isekai memory clean [--apply]`** (`memory/scrub.go`): walks the memory surfaces with a dated cure deck (today: the outer-room report markers + U+FE0F squib). Dry-run measures; `--apply` writes through byte-exact backups under `.isekai/tmp/scrub/<ts>/`, drops only cache lines whose question survives with zero content tokens (caches are free; the log and the notes are records — La — no scrub hits anywhere`).
- **`isekai memory review`** (`memory/review.go`): per-tier lens on the tier model (*Memory-tiers*, isekai.md):
  - **short — attention**: cross-body question stems (≥2 bodies on the same stem → Nature 4 current: genesis-or-skill route), one-body research loops (≥3 → craft skill), junk-anchor share (starved recall → reseed/skill). Live-fire today: 414 junk answers of 707; six slimes on one wire-shape stem; rimuru on a 25× founding-desk stem + "yo" ×8.
  - **shared — lessons**: restated lessons (same token head twice) → the first cure died in practice → `[law]` consolidate into genome, or the second naming.
  - **long — genomic**: index freshness (`IndexHoles`), kind balance (procedural == 0 → teach the how via skills), friction families in the episode stream (same deck as clean).
- Every review finding carries its route — `[genesis]` birth a creature · `[skill]` craft under `.isekai/crafted-skills/` · `[law]` consolidate · `[harness]` instrument fix. Review decides nothing — births/law wait on Veldora out loud (Containment, Law 6).
- Live cure today: 5 marker stains scrubbed (notes.jsonl ×2, rimuru caches ×3), index reseeded, the speech-law note re-worded so it names the forbidden forms descriptively.

## Evidence

- Tests: `TestScrubDryVsApply` (dry-touches-nothing, backups, drop-on-empty-question, reseed, idempotence), `TestReviewTiers` (per-tier signal+route presence). Full suites green: world, tool, tui, memory, loop, wire, config.
- Files: memory/{scrub,review}{,_test}.go, memory/cli.go (cases), bin rebuilt.

## Watch next session

- A marker recurrence = second naming → the genesis/skill decision belongs to Veldora's call: speech keeper creature vs `crafted-skills/speech-register`.
- Junk-anchor share should sag now that the index is fresh post-wave; if not, the short tier is starved of procedural truth — the route is [skill], not another reindex.

---

# Handoff — 2026-10-03 — plan decks are session-aware: stale = memory, never a gate-holder

Who: opencode session, on Veldora's report: relaunched TUI shows yesterday's rimuru plan and asks why it never updates / what "continue on last plan" sees. COMMITTED this wave.

## What landed

- **Root**: decks live as global `tmp/plans/<body>.json` with an `updated` stamp but nothing session-scoped, so the rail re-showed 2026-10-02's rimuru deck as live, and the `PlanOpen` end-gate (loop.go:994) judged the new session by yesterday's open items.
- **Fix (no deletion — decks keep continuity)**: deck stamps compare against the session's birth. Interactive modes stamp it exactly (`App.planBoot` in REPL + TUI start), single-shot runs ride `processBoot`. A stale deck:
  - never holds `PlanOpen` (`planOpenReason` takes boot; stale → ""),
  - wears a banner in `plan read` ("stale deck — memory of last session's intent … re-pen to continue"),
  - shows `· stale (penned in an earlier session)` on the TUI rail card (`RailPlan.Stale`),
  - Re-penning refreshes the stamp → current again. Unparsable stamps fail open.
- Files: app/worldtools.go, app/app.go, app/repl.go, app/tui.go, app/tuiboard.go, tui/blocks.go (+ TestPlanSessionSemantics in app/plan_test.go), gofmt normalized on the touched files.

## Verified

- `TestPlanSessionSemantics`: stale≠gate-holder, stale banner present, re-pen cures, unparsable fail-open.
- Suites: tui green; app green except chronic `TestTUIInterruptAndQueue` (pre-existing, stash-verified this morning).
- NOTE first wave before this one (9d37db0..36241c6) already rode an old binary in your live session — today's plan fix lands when the next launch starts, exactly like the dispatch one.

---

# Handoff — 2026-10-03 — the file kit (tool/delete + upgraded read/write/edit), branch feat/filekit

Who: opencode session, on Veldora's orders: "do a good tool optimised (read/write/edit/delete) for LLM", "delete inside the project actually archives in old session not delete", "write of file anchor it to slime ... helps the harness", "execute freely inside the workdir, outside needs approval", "review also read if you can do better".

## What landed (commit 035bb03 on feat/filekit)

- **delete — archives, never destroys**: moves to `.isekai/attic/<date>/<rel>` + ledger line in `attic/attic.jsonl`; records (Law 4) refuse outright; floor=Write so in-world archive is free like its siblings, outside rises to Outward (human approves), records to Destructive.
- **write — drift receipts + zone anchor**: overwrite names old/new bytes (+same-bytes recognised), result ends ", zone of slime-x" when a creature authors the zone.
- **edit — teachable errors**: 0-match answers name the closest line; multi-match names the matching line numbers; success carries the zone anchor.
- **read — tails (`limit:-N`), did-you-mean sibling hints on missing paths, long lines clamp at 2000, batch entries may pin per-path windows (`path:start[:end]`), one-file reads name the zone's keeper**.
- **ZoneOwner wiring**: `tool.Env.ZoneOwner` ← `loop.Engine.ZoneOwner` ← `w.Owner(rel)` (the same source the gate reads).
- **Surface**: rank tools whitelist `*` — writing ranks get delete auto; dark-elf stays read-only by its explicit list. Config schema + permissions + defaults + app shelf extended; TestRegistry grown to 8 builtins.

## Collision log (the live world twice swept this WIP, both times preserved)

1. slime-app-commands git-stash-pushed foreign WIP to clear its own gate ("foreign-WIP tool/ round-2") — recovered from stash.
2. a later hygiene pass `reset`+`clean`+dropped stash — recovered from dangling commits via `git fsck --unreachable` (3fe4819, fd355b8). Lost-but-rewritten: loop/world ZoneOwner wiring (never stashed, re-authored).

Lesson the audit will want: foreign-WIP in the worktree trips the live world's own gates. This wave rode a feature branch with early commits.

## Verified

- tool/filekit_test.go: TestWriteDriftAndZone, TestEditHints, TestReadKitUpgrades, TestDeleteArchives. Suites: tool, config, world, loop green; app green but for chronic TestTUIInterruptAndQueue (stash-verified pre-existing).

## For merge later (when the live session closes)

`git checkout main && git merge feat/filekit` — replay test expectations if the world touched shared tests meanwhile.

---

# Handoff — 2026-10-03 (evening) — file kit merged to main, binary installed, memory reseeded

Live world stopped by Veldora's hand; merge landed (`e2c3bcb`, --no-ff): tool kit on main. Full suites green (tool/config/world/loop/memory). Installed `~/.local/bin/isekai`. Closing rites: index reseeded (278 memories; built @ merged main), markers stay clean (`@S CLEAN — no scrub hits anywhere`). Stale decks from the wiped stash saga are gone from pose; `plansBoot` anchors ru ssion-aware plan decks, and now the LLM file kit.

---

# Handoff — 2026-10-03 (night) — housekeep --attribute + choice modal (recommend / pin / ledger)

Who: opencode session, on Veldora's orders: "audit is better", "simpler has more gain", "in other harness recommendations when choosing", "the popup text floats when I move choice to choice and the format is not the best", "the recommendation could have a value function that adapts and evolves", "all what's not said is intelligence".

## What landed (on main)

- **`isekai housekeep --attribute`** (`cmdAttribute`, commit a7b2981): audits the WHOLE file inventory (tracked + untracked + root-level ignored surfaces like log-git.md) against creature territories; every unowned file gets the strongest suggestion the law's census allows — same-dir majority → ancestor territory → journal-class (elf/isekai). Reports, never assigns: territory real map of the root-journal debt the elf was choking on).
- **Choice recommendations** (opencode-parity): `tool.Question` gains 1-based `recommended`; the TUI tags the advised option `◆ recommended`, pre-seats the cursor there — NEVER auto-picks, and approvals over outward/destructive acts stay silent **on purpose** (the absence is load-bearing; what's not said is intelligence).
- **Float cure**: the modal is growth-pinned (`pinChoice`) — cursor moves and the typing row stop reflowing the centred card; goldens rewritten (`-goldens` flag lawfully) — approval cards now render compact single-row options with inline notes.
- **Choice ledger**: answered choices publish `EvChoiceAnswered` over the bus → the Bundle-C tap lands them in `.isekai/instruments/stream/<day>.jsonl` as `choice-answered` rows. This is the adaptation's honest home: evidence for "you always allow class X → graduate a rule", never live cursor steering.
- **Tap hardening**: the instruments tap now only WRITES to the already-open file — publish paths never create dirs (dash teardown no longer races the tap; was the failing TestDashGateAnswerFlow).

## Verified

`TestAuditOwnership`, `TestChoiceRecommendedTag`, `TestChoicePinStabilisesCard`, `TestChoiceAnsweredLandsOnTheTap`; dash/choice 5×; suites green but for chronic stash-verified `TestTUIInterruptAndQueue`.

## Deferred by design

- No adaptive value function in the gate — the doctrine (recorded for review): live nudges optimise approval reflexes (bad at safety surfaces); the evidence lane feeds rule graduation instead (Nature 3).

---

# Handoff — 2026-10-03 (deep night) — the scenario funnel (scenario-3-2-1)

On Veldora's decree: value function = scorecard read off instruments (friction, verdicts, holes), never online weights, never a human inside the funnel.

## Landed

- **`.isekai/crafted-skills/scenario-3-2-1/SKILL.md`** — first citizen of crafted-skills; carries the ritual: rimuru 3 scored cards → elf keeps 2 plans → orc commands 1; annulled cards ride the wire as provenance; score rows quoted or named a hole. Law note on the shared bus.
- **`plan` deck** gains `score`/`why` on items; `renderPlan` prints `◆ score (why)` and names the hole on a naked number. Test: TestPlanScoreCards.
- **`EvPlanOutcome`**: at turn close, every deck updated inside the turn rides the verdict out over the bus tap (`plan-outcome` rows in instruments/stream). Test: TestChoiceAnsweredLandsOnTheTap (ledger arm) + wire live.
- **`memory review` fourth tier "value — decisions"**: today's per-body verdicts/gate fails, deck scored-adoption, tap rows. Live-fire lit correctly: 12 bodies, decks unscored (the adoption hole surfaced on day 0, as the ritual teaches).

## Verified

tool/tui/memory/world/loop/config suites green; app suite green but the chronic stash-verified flake set (dash goes red ~in-suite only, standalone 10/10 — documented class).

## Watch next funnel run

- the first rimuru order ×3 cards → elf ×2 → orc ×1: does the ritual hold the shape, and does the first `plan-outcome` row land in the tap (the learning ground).
- naked-score path is loudly marked (@?). If you SEE one, the ritual is being half-worn.

---

# Handoff — 2026-10-03 (dawn) — race law: the crest stays global, each race boots its own page

On Veldora's redesign call: "keep Nature law global, then law per race — better harmony, low context, keep room for useful stuff. In harmony, simpler, and smarter."

## Landed

- **`.isekai/law/race/{rimuru,elf,orc,slime,kijin}.md`** — five thin pages; each opens with the consult rule ("read the crest daily; consult the code on the day a question crosses ranks") and names only what that race answers for. Nothing duplicated from the universal code (Nature 2 protected).
- **Prompt hook**: `w.Prompt` now renders `Law of your race` right after the crest. `raceLawPath` refuses any body not on the roster (RankOf's rimuru fallback was a rules-filter default, never a warrant for the throne's law) — tested no-leak on all five plus a rankless probe.
- **Load economics**: slime boots crest + slime page only. The universal code stays consult-on-demand via the `law` tool — unchanged.

## Verified

- `TestRaceLawNoLeaking` green marker-per-race + rankless probe; world suite green; golden shots untouched.

## World-impact note

First prompt-builds post-wave will carry race pages — context math improves immediately; leaks are test-locked. The funnel's per-race duties now live in the right pages too (rimuru: 3 cards; elf: keep 2 plan 2; orc: hold 2 command 1; slime: no crowds).

---

# Handoff — 2026-10-03 (afternoon, post-relaunch) — the pre-armed doom & the resume door

Two live-wire bugs from the 14:52Z session (the first to ride today's waves), both sealed:

## 1. The pre-armed doom (loop)

Doom checkpointed rimuru at 15:24Z (guts spoke streak 3 ≥ threshold — bench re-echoes during an
audit-style prowl keep momentum at floor; the guts were honest: those reads *weren't* earning).
The relief then **left the session armed**: `gutsSpoke` and the valve never reset. The wake turn
at 15:35Z ran one act, momentum still at floor from before, and insta-checkpointed. The world
looked alive and could not work. Fix: the doom branch now spends the streak and clears the
valve (`loop/loop.go`), tested in `TestDoomLoopReliefLadder` (a session that slides again after
relief must earn its doom afresh; a relief must not hand the next turn a primed one).
law.yaml's comment corrected: threshold counts guts spoke-at-floor streaks, not failed acts.

## 2. The resume door (cli)

Bare `isekai resume` errored `needs a session id` while the checkpoint hint urged "resume with …" —
the command said one thing and failed when followed. Fix: bare resume now continues the newest
session and says so on stderr. (Commit `c9de101`.)

## Note

Your session `20261003-145249-f21b6d` is intact, 334 messages, resumable. Kill the old process,
relaunch (rides `48737bc`), then `isekai resume` bare.

---

# Handoff — 2026-10-03 (evening) — the sed ban + kin-first recall, after the orc ate the first attempt

Two fixes your wire earned today:

## The sed ban (guard, not doc)

An in-place `sed -i` edit silences both failure and attribution: wrong anchor → exit 0, file
untouched, no res.Wrote. Today's journals held ~216 sed-backed acts against 80 `edit` calls; the
verification re-reads behind unverified writes fed the guts' floor (doom feeders). `app/guard_sed.go`
joins the guard chain — refuses anchored `sed -i` / `sed --in-place` / `-i.bak` / `-i ''` through bash
at `PreTool`, names the `edit` tool in the refusal. Streams stay free: `sed -n N,Mp` reads and pipe
shaping untouched.

## Kin-first recall

Dispatched bodies were re-deriving context from log.md because the index didn't know each race's
`thoughts/*.md` and the chronicle won ties on pure term mass. Now: creature thoughts files are
indexed as `procedural/thoughts` sources; a body's own creature mems (+0.25) and race mems (+0.10)
top the RelationBoost ladder; episodic log.md takes a soft −25% demote. Elf beats should now boot
warm from the prequel's thoughts instead of re-reading the chronicle.

## The porcelain law

`orc-host` at 17:41:50Z ran `git checkout --` over tracked edits and `rm -f`'d three of my untracked
fix files because it was policing porcelain it didn't author. Bounded by the new orc-page line:
discipline never deletes the foreign; stash-verification sweeps keep stash/pop only.

---

# Handoff — 2026-10-03 (evening, second wave) — the onto hooks ride the write path

Veldora's ontology order landed as the advisory pair, honest by construction:

## What landed

- **The `Note` seam** (`loop/loop.go`): post-act advisory channel that rides the tool result the
  provider receives — ground truth stays on the step, the wire learns the note. Merged through
  `MergeHooks` so courts get the app's onto hooks automatically.
- **`app/onto_hook.go`**: every write-class act journals an `onto` ledger row (`path`, `owner`,
  `as`) — the attribution census counts itself from now on. First unowned write of a run rides
  one advisory line back to the body: claim the soil in a creature's Territory field or journal
  provenance. One nudge per run by law; claimed soil never nags; reads never count.
- Switch: `law.territory.enabled` — ownership intervention rides the same gate that names it.
  Foreign-owner refusal stays unbuilt deliberately: warn first, refuse when the census is complete.

## Evidence embedded

The advisory text names the day it came from (2026-10-03: sed no-op → bench echoes → doom feeders).
Tests: `TestNoteHookRidesResult` (wire-position proof via mock requests), `TestOntoNote`
(ledger rows, once-per-run cadence, owned soil silent), `TestMergeHooks` (note concatenation).

---

# Handoff — 2026-10-03 (night) — the relay speaks the existing wire

The Arm A stall: the desk narrated a relay; the @ASK spoke findings-only; nobody was booked.
Root cause read through app/court.go: the court ALREADY had order-book plumbing (`@$ body: ask`
lines in a report drained into dispatches at turn end, fail-closed). The day's gap was grammar,
not machinery: nobody wrote the @$.

Commissions now speak it twice-duplex:

- **rimuru page**: relay commissions carry `@$ <body>: <ask>` literally inside the ask,
- **elf page** (mirror): findings reports echo the @$ lines back or name the refusal at @?,
- **orc page** (gate-side): expecting @$ in, @$ out; unbooked = hole to name.
- **harness**: court records every commission's @$ as an expectation; a body's report that
  closes without echoing its @$ wakes the throne with `[relay unbooked] <body>: <ask> …` —
  one per missing line; a relay booked meanwhile (tracked body) rides clean. Wire grammar
  exported as `wire.EnqueuesOf`.

---

# Handoff — 2026-10-03 (night cap) — the veil-exit row: the hand's name rides the journal

Audit question: "you stopped it — did the instruments count it?" Standing answer before: they
saw the done-end, but could not tell earned end from human stop. Now they can:

- **cause on end rows**: a turn closed mid-drive by a cancelled context carries
  `"cause": "<the lexicon's word for the human>"` (veldora here; "the human" when the world
  carries no lexicon). Earned closes journal no cause.
- **exit rows between turns**: the TUI and REPL write `{"t":"exit","cause":"veldora","ms":…}`
  when the hand reaches stop between turns — separate from instrument checkpoints.

