# The races' loops

How each race of a world runs its loop, as the binary really runs it: who commissions a body and
with what, what its system prompt carries, what it may touch, the beats of one turn, how the turn
ends, what stress does to it, and how it ascends. [`ISEKAI.md`](../../ISEKAI.md) says the same
law in prose; where the two read differently, this page says what the code does, and the last
section lists the places.

The examples use this repository's own colony (`.isekai/`): four elves (`elf-engine`,
`elf-world`, `elf-host`, `elf-canon`), six orcs (`orc-engine`, `orc-safety`, `orc-world`,
`orc-config`, `orc-app`, `orc-interfaces`), twenty slimes (`slime-wire`, `slime-loop`,
`slime-gate`, …) and one kijin (`kijin-release`).

## Two tables, one law

A rank is a row of data, not code (`world/rank.go` `Rank`): `ReportsTo`, `Job`, `Authors`,
`HoldsGate`, `Sideways`, `Office`, `Tools`, `Dir`/`Prefix`, `Base`, `Layer`, `DivineRules`,
`AscendsWhen`. Nothing in the gate, the territory policy or the dispatcher names elf, orc or
slime; they read these columns.

The binary takes its table from config (`config/ranks.go` `builtinRanks`, folded into world ranks
by `app/world.go` `ranksFor`); `world.DefaultRanks` is the table a world opened without config
uses. They agree on the tree, the flags, the ascended ranks' own dirs and the ascension
conditions; they differ in two places:

| column | the binary (`config/ranks.go`) | `world.DefaultRanks` |
|---|---|---|
| `Tools` | `*` for elf, orc, slime, kijin | elf/orc/kijin every tool; slime `read write edit bash glob grep law` (no `dispatch`, no `plan`) |
| dark-elf | `read glob grep ls recall onto toolbox git ask` | `read bash glob grep law` |

A rank's `Tools` list is the most it may hold, not a promise: a rank that commands no one — no
rank reports to it (an ascended rank counts its base's children) and it is not `Sideways` — is
never handed `dispatch` (`Ranks.Commands`, checked in `world/court.go` `Engine`). So the kijin,
whose table row says `*`, holds no dispatch.

The `Job` line rides the body's creature card in its prompt, in the distribution's word for the
rank: "Your rank's job (Orc): rules its domain, holds the gate, …" (`world/prompt.go`). There is
no `Body` column: every dispatched body, a kijin's included, is a Court with a fresh context, so
config's `ranks.<r>.body` is accepted and reported as a hole ("ignored"), never obeyed.

| rank | reports to | authors | holds gate | sideways |
|---|---|---|---|---|
| elf | rimuru | — | — | — |
| orc | elf | — | yes | yes |
| slime | orc | yes | — | — |
| kijin | rimuru | yes | yes | — |
| high_elf | rimuru | as elf | as elf | as elf |
| high_orc | elf | as orc | yes | yes |

An ascended rank inherits every field of its base except four that are its own: `Dir` and
`Prefix` (`high-elf/`, `high-orc/`; agent-one `principal-coord/`, `principal-domain/`),
`DivineRules` and `AscendsWhen`.
| dark_elf | rimuru | — | — | — |

A rank's `Office` is a fallback nothing dispatched reaches: a Court's office always comes from
its commission (`world/court.go` `Dispatcher`), and the throne holds none.

`Ranks.Validate` refuses a table where an authoring rank has no gate holder at or above it (Law 3)
or a rank does not reach rimuru.

## The loop every body shares

`loop/loop.go` runs one turn: the model plans, each tool call is a step, the model answers.

- **Perceive** (before each model call, `drive`): wall clock, the budget hook, the context
  reading. In the stress zone (`law.budget.stressTokens`, default 180000 of `contextTokens`
  200000) the compaction drain runs; if it cannot drain, the run ends `CHECKPOINT`.
- **Recall** (per model call): memory anchors (`path#section`) and the toolbox's level-1 manifest
  for the current ask ride the prompt tail as `@RECALL`/`@TOOLS` (`world/hooks.go` `Recall`).
- **Plan**: the one model call.
- **Act**, per tool call (`step`): step budget and clock; the tool's class is settled (`read`,
  `write`, `outward`, `destructive`, `tool/classify.go`); then the guard, the territory policy
  (`world/territory.go`), permission rules, and the human gate, which asks for `outward` and
  `destructive` always and for `write` too under `law.humanGate.strict` (`gate/gate.go` `Needs`).
  The gate is the session's: a Court's outward act is put to the same human.
- **Verify**: the tool's own result; an error is a failed act. The guts score the act
  (`loop/guts.go`; window 5, floor −2, `law.guts`) and speak once inside the tool result when the
  window slides past the floor. When they have spoken `law.doomLoop.threshold` (3) windows running
  with no recovery, the run checkpoints with the retire / split / genome-shift arms for the parent.
- **Record**: the journal line; `@U` lines inside a tool result (a Court's report) land as an
  ontology fact and a shared note (`world/hooks.go` `Record`, `Land`). `@FORGE` drafts are not
  read off the tool result: the dispatcher files them from the Court's own report, whole, as a
  colony note tagged `forge` with the body, run and hash (`world/court.go`, `world/forge.go`).

More than `law.escalation.maxRetries` (2) consecutive failed acts end the turn `ESCALATE`, "one
hop up: the dispatcher decides".

**Closing** (`finish`). A dispatched body must answer on the wire (`@S` opens): an answer off the
wire is sent back once; with `+unsaid` in force (`law.wire.requireUnsaid`, default on) a report
with no `@U` is sent back once. Then, if the turn wrote files, the turn-end gate runs
(`world/gate.go` `EndGate`, below). A fail goes back to the model with its reasons
`law.gate.retries` (1) times — the writes stand, the fix is a new step — then the turn ends `FAIL`.

**The gate** runs on the body's own writes, less what its Courts wrote and gated on their own
account (`world/court.go` `ownWrites`):

1. *Right author* (only while some creature holds a gate): a path inside another creature's
   territory, written by a body that neither owns it nor holds it in its own territory, fails.
   A creature's own `README.md` is exempt for that creature.
2. *Traits hold*: every touched creature's verify lines run, sandboxed, `law.gate.verifyTimeout`
   (120s) each. The lines are those that stood **before** the turn; a line added or changed this
   turn is reported as a hole and runs from the next turn.
3. *Duties done*: a commissioned turn must answer with a non-empty `@S`.
4. *Doc truthful* (Vitality): a change under a territory needs the owner's doc to move in the
   same wave, or an earlier paired wave (a Court's recorded doc write, or a pass-minted pair mark).

Then *tests intact*, the UI lints when the turn touched the UI dir, and the injected `law` checks
whose scope reaches the body. The verdict is appended to `.isekai/log.md`. The same body failing
on the same file-set three times running hardens the fail text into the lawful next acts
(rewrite the commission, `@SPLIT`, escalate) — the circuit-breaker, `gateBreakerAt`.

## Rimuru — the throne

**Commissioned by** the human: a typed line, a queued line, a background Court's report waking
the session, or a drained `@$` order. Rimuru is the session engine at depth 0
(`app/app.go` `Engine`), on the mounted model, with no office.

**Its prompt** (`world/prompt.go` `Prompt`): the crest from `.isekai/isekai.md`, the race page
`.isekai/law/race/rimuru.md` when one exists, the identity text in its human form ("you speak to
the human directly… the wire is for the reports of bodies you dispatch"), config `rules` scoped
to it, the self-state lines (the `pulse` tool, guts and doom notes), the `@ONTO` projection for
`rimuru`, `AGENTS.md`/`CLAUDE.md`, and the memory line. No creature card: the throne owns no
territory.

**What it may do.** Every tool. Under `law.territory.throneNeverAuthors` (default on) every write
is refused except the law file and `.isekai/dimension.yaml`, with a refusal naming the first rung
to dispatch; with it off, the throne may write unowned soil and a creature's own doc, and is
refused on a creature's soil otherwise. It dispatches; under `tools.dispatch.strictFlow`
(default on) only the ranks directly under it — `elf-*`, `kijin-release`, and any high elf or
dark elf. Every creature's line reaches the throne, so the lineage law below never binds it. Dispatches run in the foreground; with `tools.dispatch.background` on, a commission may
ask `background: true` (`sessionBackground` makes it the default) and the report wakes the session.

**One turn.** Perceive and recall as above; the model reads the ask, may write its own plan deck
(`plan` tool, `.isekai/tmp/plans/rimuru.json`), and calls `dispatch` on an elf. Each report comes
back as the dispatch step's result (below); its `@U` lines are landed by the record beat. At the
end, the throne's own writes (usually none) meet the gate, and a deck with an open item is
reported as a hole ("closed with an open plan"). `law.closingCheck` (default off) adds one more
re-ask: is the ask applied end to end, with a witness for every behaviour the writes require.
`law.forge.demandAtSteps` (12; 0 = off) asks any body whose turn ran that many steps, once, for
its `@FORGE` draft or why the path is one-off — with or without the closing check. The throne's
own final answer files its `@FORGE` lines as drafts by `rimuru`.

**Ends** with a plain answer to the human — never capped (`Cap = 0`). Stress for the throne is the
context reading: the drain, then the checkpoint. The council does not watch it; it is no Court.

## Elf — the shared mind and voice

**Commissioned by** rimuru (or, through `@$`, by any body whose order the throne re-dispatches).
What arrives: a fresh context and one user message, the commission rendered by the dispatch tool
(`tool/dispatch.go` `Commission.String`):

```
@ROOT /home/…/isekai
@ASK draft +unsaid
@CAP 2048
<the ask in prose>
```

The `@ASK` word is the office (`findings` → great-sage, `verdict` → raphael, `draft` → ciel;
default great-sage) and routes the model through `models.offices` / `models.ranks` /
`models.creatures`.

**Its prompt**: as the throne's, but the identity is the wire form (answer `@S … @E`, `@U` for
the unsaid, `@FORGE` for a proven repeating need, and — because it holds `dispatch` — "order work
one layer down with `@$ <body>: <ask>`"), plus its creature card: name, rank, doc path, territory
(for `elf-engine`: `loop/ … tool/`, the lane's routing map), one hop up, the Minds worn, its
verify lines, the Vitality rule, its rank's job, and the divine rules any high elf declares. Its layer is every
rank under rimuru, so the card names `elf-canon`, `elf-host`, `elf-world` and `kijin-release` as
neighbors — knowledge only: an elf is not `Sideways` and may not dispatch them. Its doc is named,
not inlined; it reads it and dons a Mind (`skill`) on its own decision.

**What it may do.** No authoring flag and no gate: the territory policy does not bind it at the
step, but its writes meet the gate at turn end, where a path a slime owns fails right-author (an
elf may not move the slime's doc) or, inside its lane, Vitality. It dispatches its own orcs: the rank law
(`Ranks.BelowAs`, `DirectChildAs`) says an elf orders orcs, and the lineage law
(`World.mayDispatch`, `tools.dispatch.lineage`, default on) says which — those whose `Reports to`
chain reaches it. `elf-engine` may order `orc-engine` and `orc-safety`, never `orc-world`; the
refusal names the elf that commissions it. Under strict flow it never orders a slime directly. It
does not dispatch another elf (elf is not `Sideways`).

**One turn.** Its doc's own trait and the `plan` tool's description set the order: write the
plan deck first — one item per slice, each `who` naming the orc whose domain owns it
(`app/worldtools.go` `planTool`, `.isekai/tmp/plans/elf-engine.json`) — then dispatch. The `onto`
tool answers who owns a path with the dispatch card (owner, strain, domain, and the widen / birth
/ route arms, `world/territory.go` `DispatchCard`). Several dispatch calls in one model message
run one after another; true parallel orcs need `background: true`, which is refused unless
`tools.dispatch.background` is on (its report then lands in the elf's mid-turn inbox). Each
orc report returns; the elf ticks its deck and speaks the lane's verdict back up.

**Ends** with its wire report to rimuru. The gate runs only on its own writes, usually none.

## Orc — the domain and its gate

**Commissioned by** its elf (the one its `Reports to` names), or sideways by another orc
(`Sideways`: `orc-engine` may dispatch `orc-safety`). Same commission shape as the elf's.

**Its prompt**: the wire identity with `@$`; its card (`orc-engine`: `loop/ instrument/ provider/
wire/ compact/`, one hop up `elf-engine`, Minds `orc-gate-verdict, review-core, loop-core`); and
"Neighbors in your layer (orc, high_orc): orc-app, orc-config, …" — the ranks sharing its
parent, named so it may reach them sideways through dispatch (`Ranks.NeighborRanks`).

**What it may do.** It holds the gate, which in code means four things:

- its existence turns on right-author for the whole world (`GateHolders`);
- a path no authoring creature covers falls to it as owner, so its doc is the one Vitality asks
  for (`World.Owner`);
- **before** it may dispatch a slime, the slime's doc must carry a non-trivial verify line
  (`VerifyCoded`: not empty, `true` or `exit 0`) — the TDD precondition in `World.Dispatcher`,
  read from the slime's doc on disk at the check, so a line the orc wrote earlier in the same
  turn counts;
- the verify lines it writes on its slimes' docs are what the gate runs on their writes.

It may not author on a slime's soil and pass: a code change there needs the slime's doc to move,
and the orc writing that doc fails right-author. An orc's own edit of a slime's doc is lawful at
its gate only when the slime's Court also writes that doc in its turn, which records the path as
court work and subtracts it from the orc's bucket.

**One turn.** Read the slice; ask `onto` who owns each path; write its own deck if the slice is
several; code each slime's verify; dispatch its slimes (only those whose `Reports to` names it —
another orc's slime is refused with that orc's name); read each report. A report filed by the courier (over 600
bytes) arrives as an envelope — `@R <path>` on `.isekai/instruments/loop/`, `@H fnv64:…`, one
`@W` per path written, an eight-line digest — and the orc reads the file when its verdict needs
the flesh. It judges the landing against the verify it wrote and answers `@S` with `@V` verdicts.

**Ends** with its wire report to the elf.

## Slime — the ground truth of a zone

**Commissioned by** its own orc — the one its `Reports to` names (strict flow keeps the elf above
from reaching it; the lineage law keeps every other orc out).
It runs at depth 3 — the floor of the default chain, so the dispatch tool is not on its shelf
(`world/court.go` `Engine`: `Depth < max`, where max is `tools.dispatch.maxDepth` floored at
`Ranks.ChainDepth`, both 3).

**Its prompt**: the wire identity without `@$`; its card (`slime-wire`: territory `wire/`, one hop
up `orc-engine`, Minds `slime-go-proving, slime-wire-and-journals`, verify
`go test -count=1 ./wire/...`), "a write outside your territory is refused; a change under it
changes its doc in the same turn"; and its neighbors, the other slimes, reachable only through
dispatch.

**What it may do.** It authors, so the territory policy binds it at every write step
(`World.Policy`): a path outside `wire/` and its own doc dir is refused before it runs, and the
refusal names the owner and the consult — `@? needs <path>` one hop up to `orc-engine`, with the
widen / birth / route arms for the orc to choose. Writes under `.isekai/instruments/loop/` and
`usage/` are refused for every body: engine soil.

**One turn.** Read the zone and its doc; write the change in `wire/`; edit
`.isekai/slime/wire/README.md` in the same wave (Vitality); run its own bench. From the sixth
hit on one bench (`law.guts.benchThrashVisits`) the guts price every further hit as thrash. At
the end the gate runs the verify lines that stood before the turn — the orc's — and the slime
does not pass by rewriting them: a changed line is a hole, and a deleted or skipped test fails
*tests intact*.

**Ends** with `@S`, one `@F <file:line>` per finding, `@?` for what it could not settle, `@U`
for what nothing on disk says (law / colony / territory), `@E`. Over `@CAP` (`law.wire.cap`, 2048)
the report drops `Other`, `@T`, `@FORGE`, `@F`, `@V` in that order; never `@S`, `@?` or `@U`.

## Kijin — the standing lane

**Commissioned by** rimuru directly (it reports to the throne, past the elf). `kijin-release`
owns `.github/`, `.goreleaser.yaml`, `Dockerfile.release`, `install.sh`, `Makefile`,
`RELEASING.md`, `CHANGELOG.md`.

**Its prompt** carries its card, its job and Minds (`kijin-release-pipeline, slime-go-proving`)
like any Court; "standing across sessions" lives in its doc's traits, not in the runtime.

**What it may do.** It both authors and holds a gate: the policy binds its writes to its
territory, and the gate runs its verify lines (`bash -n install.sh && go build …`, `gofmt`, the
`## [Unreleased]` check) on its own writes. No rank reports to kijin, so it commands no one and
its shelf carries no `dispatch`: the release lane is its own hands' work, and what it cannot
settle goes up to rimuru as `@?`.

**One turn** is the release lane: change the workflow, the changelog or the installer inside its
territory, prove it with its verify lines, report. A local `git tag` is a write; what leaves the
world — `git push` of a branch or a tag, a registry push — classifies `outward` (`--force`
`destructive`) and stops at the human gate (`tool/git.go`, `tool/classify.go`); a denial ends the
turn `DENIED`.

## Born in time — high elf, high orc, dark elf

- **High elf / high orc** run their base's loop: an ascended rank keeps every flag its base held
  (config refuses dropping one) and commands what its base commanded — a high orc orders the
  slimes that reported to the orc it rose from — and its `DivineRules` reach the base rank's
  prompt as "absolute rules from the ascended above you". A high elf reports to rimuru; a high orc
  to the elf, so its elf may dispatch it. Its body lives in its own dir (`.isekai/high-orc/engine/`)
  under the base's short name.
- **Dark elf** — the auditor, reporting to rimuru. Its shelf carries no `write`, `edit` or
  `dispatch`; a turn that writes nothing meets no gate. It answers with `@V` verdicts on `log.md` and
  the gate's verdicts.

## The return path

What the dispatcher does with a report (`world/court.go` `Dispatcher`, `tool/dispatch.go`
`render`):

1. The Court's step budget is what the parent has left (`law.budget.steps`, default 40), its
   clock the parent's remaining minutes (`law.budget.minutes`, 60), or 30 minutes when the
   session runs uncapped. A tenth of it (2s–1m) is a grace window: on the clock's death the body
   is resumed once, capped at `law.budget.salvageTokens` (4096), to land its `@U`/`@F` before the
   axe; the death and the model lane are stamped on the stress ledger with a next-lane memo.
2. The report rides the dispatch step's result under an exit line `@S <DONE|FAIL|…>`; a report
   with no `@U` gets a `@?` asking for `+unsaid`. A `FAIL` run returns as a failed act, so three
   failed dispatches running escalate the dispatcher's own turn.
3. Writes the Court reported are recorded against the dispatcher, so its gate does not re-judge
   them (`courtGated`).
4. The council reads the run (below); the tally that ascension reads grows, and an ascension or
   an ascension hole that fold caused rides the report (below).

`@$ <body>: <ask>` lines in a report are orders, never Courts minted mid-turn: a background
Court's orders go to the order book (`app/court.go` `Finish`), and at the throne's turn boundary
each becomes a `dispatch <body>: <ask>` line for rimuru's next turn (`app/repl.go`), where the
dispatcher's rank and strict-flow checks apply again.

## Stress — the council, the doom checkpoint, relief, evolution

All four are coded; none asks the body.

- **Council warrant** (`world/court.go` `councilLedger`). A Court whose run's word is a fail
  under `law.value` (§Value: the loop's status when it did not close `DONE`, else its own `@S`;
  `FAIL`, `ESCALATE`, `CHECKPOINT` by default) adds strain:
  `failed acts + gate re-asks + wire re-asks + 1`, accumulating across re-dispatches. At `law.stress.escalateAt` (5) a warrant rides
  the report to its dispatcher: answer as its manager with `@INSTRUCT` (re-commission with the
  instruction folded in), `@SPLIT` (one commission per slice) or `@ESCALATE`. Past
  `law.stress.maxConsults` (3) warrants the report ends `@S ESCALATE`, toward the human. A clean
  landing (a pass word) clears the ledger. A neutral word — the human's `DENIED`, an
  `INTERRUPTED` or `DRY` run, a `SUSPENDED` question — never counts.
- **Relief** (`world/relief.go` `relieve`), at `law.stress.relieveAt` (7): a creature with two
  or more territory prefixes splits into two same-race young (`slime-loop-a` holding `instrument/`,
  `slime-loop-b` holding `loop/`), each inheriting the doc, verify and Minds; a one-prefix zone
  splits by the tree under it when it holds two or more entries; otherwise the creature retires
  (`- **Retired:** true` in its doc: the seat keeps its memory, the zone becomes open ground). The
  dispatcher reads `@? council relief — …`. Any rank with a doc can be relieved, an orc or an elf
  included.
- **Doom checkpoint** (`loop/loop.go`): the in-run counterpart — the guts spoke
  `law.doomLoop.threshold` windows running, the run checkpoints, and the parent receives the
  same arms (retire, split, genome-shift).
- **Evolution** (`world/evolve.go`, `law.evolution`), weighed after every Court's report folds
  and once at run start: a creature that is **stressed** — strain at `stressAt` (0 reads
  `law.stress.escalateAt`), its last `failStreak` (3) gate verdicts failed, or its last run a doom
  checkpoint — **and** whose desks hold more than `thoughtsMax` (7) Thoughts (its own doc's plus
  every worn Mind's under the root, the creed's split) heals itself in the same wave. Relief does
  not also fire on a body that just evolved. See §Evolution.

Escalation is always one hop: a slime's `@?` reaches its orc in the report; the orc settles it or
carries it up in its own; `ESCALATE` from a failed-act streak or an exhausted council ends the
turn and the dispatcher decides.

## Evolution

A stressed creature whose desk overflowed is a genome that no longer fits its work (Natures 3–5,
the adaptation loop). `law.evolution` (`enabled: true`, `mode: auto | propose`, `thoughtsMax: 7`,
`stressAt`, `failStreak: 3`, `cooldownRuns: 10`, `arms: [split, mutate, rebirth]`, `splitLines`,
`splitFiles`) picks one arm from evidence, in this order, and writes the reason in `log.md`:

1. **split** — the rank may split (an author: slime; a gate that commands: orc), its territory
   reads large (`splitLines` source lines or `splitFiles` files in the founding sketch's survey),
   and its own Thoughts and its failed runs' paths fall in two or more of its sub-directories (its
   prefixes, else the entries under its one prefix). The two busiest entries seed two halves, the
   rest go to the lighter half. Two young stand in the rank dir — named for their halves
   (`slime-pay-api`, `slime-pay-core`) when each half is one entry, else `-a`/`-b` — each with the
   parent's identity fields, its half, the survey's traits and verify for it (else the parent's
   verify), the parent's traits that name only its soil, and the Thoughts that name its soil (the
   rest to the half carrying fewer). A worn Mind's desk stays in the Mind both wear. An orc's
   slimes are divided with it: each reports to the young whose half holds its soil. The parent
   retires (`- **Retired:** true — evolved …`).
2. **rebirth** — its genome is stale: its gate failed on its own verify lines at its last verdict,
   or its doc names soil that is gone (a territory prefix, or a path in its traits under a tree
   that exists). A fresh doc is written from the code as it is now (the survey's traits and verify)
   plus every distilled lesson, keeping name, rank, territory, parent and Minds, with a
   `- **Reborn:**` line; the old doc rests in the attic, and a `rebirth` row starts the body's value
   record again (its own runs, verdicts and strain before it are not read; a dispatcher's counts of
   it are).
3. **mutate** — its Thoughts repeat a lesson: the term (a path before a word) three or more share.
   Every family of two or more becomes a trait (`<term> — <latest wording> (distilled <date> from
   N thoughts, first…last)`); a verify line is added where tests cover the paths a lesson names;
   a family most of whose Thoughts say how (a backticked command, run / use / always / before …)
   is minted as a crafted skill `<body>-<term>` at `law.forge.evolvedStatus` (trial) and added to
   the body's Minds. Same name, same territory. When no arm's evidence holds the desk is still
   distilled (Nature 5) as a mutate. Thoughts with no family go to the log's wrap-up.

Every arm empties the desks it distilled into the genome, each keeping one dated `distilled into …`
line. **Safety:** the arm is a plan of writes, staged under `<world>/tmp/evolution/<id>/`, proven
before the swap — the new verify lines run green (the gate's sandbox and clock), and the ontology
read over an overlay of the planned docs carries no finding the world did not already carry and
gives every planned body exactly the soil the plan meant. A plan that fails its proof, or meets a
doc changed since it was drawn, swaps nothing: the old docs stand and a hole is logged. The old
doc always rests in the attic (`<world>/attic/<name>@<date>.md`). A retired seat keeps its graph
memory but claims no soil, so the young never overlap it.

Every evolution is logged through the append-only writer (the trigger evidence, the arm, the
reason, the files, the wrap-up), noted as an `evolution` row in the value ledger (`status`
applied | proposed | hole; the cooldown counts the body's runs after its latest row, so a world
reopened remembers it), clears the body's strain, rides the dispatcher's report as
`[evolution] …`, and reaches the surfaces: an `evolution` event on the session stream, the
terminal's toast and sidebar line, the dashboard's timeline row and a ring in the arm's hue on the
brain net (split, mutate, rebirth each their own), `isekai status`, `/api/value` (each body's desk,
stress, cooldown and last evolution). The harness's writes are the body's: the dispatcher's gate
leaves them alone. `mode: propose` writes the plan and its diff to `<world>/evolution/<id>.json`
instead; `isekai evolve list | show <id> | apply <id> | reject <id> <why>` is the desk, and apply
runs the same proof. Evolution never moves a body's model lane (the genome-shift arm stays
advisory). `isekai config explain evolution` prints the law with its origins.

## Value

Every race is measured by its own job, by one law (`law.value`), from one ledger
(`<world>/shared/value.jsonl`). Nothing here asks a model: every number is a fold over rows
the harness wrote (`value/`, `world/value.go`).

**One definition of success.** A run's word is the loop's status when the loop did not close
`DONE` (the clock, the human or the gate spoke), else the body's own `@S` word. `law.value.pass`
(`DONE`, `PASS`), `.fail` (`FAIL`, `ESCALATE`, `CHECKPOINT`) and `.neutral` (`DENIED`,
`INTERRUPTED`, `DRY`, `SUSPENDED`) classify it; a word on no list is neutral. The ascension fold,
the council, the dashboard's leaderboard and activity, and the loop journal's `end` line (its
`word`) all read it the same way. A neutral run counts in `runs`, never in `reports`, so no rate
reads the human's word as the body's failure.

**The ledger.** One append-only line per row, always kept — a world that ascends nobody is still
measured — and replayed at `Open`. Kinds:

| kind | written when | carries |
|---|---|---|
| `run` | a Court's report is taken; a throne turn closes (the `Ended` hook) | the loop status, the body's word, the run's word, the gate's verdict and retries, verify lines run / red, holes by kind, `@F` findings, strain after the council, a warrant, the guts' closing momentum and thrash count, tokens, the paths written, the dispatcher (`by`) |
| `verdict` | an end gate judged a Court's writes | the body that held the gate (the dispatcher when its rank holds the gate, else the author when it holds its own), the author (`of`), pass or fail, the paths |
| `overturn` | a fail lands on paths a standing pass covered | the holder whose pass it undid |
| `ascend` | an ascension is minted | the base and the risen body |
| `seed` | the ledger was first written from the old `instruments/ascension.jsonl` | the shelf it read (the old report rows become run rows, their holes harness holes) |
| `evolution` | a stressed body evolved, proposed to, or failed its proof (§Evolution) | the arm, the reason, the young, the status, the paths |
| `rebirth` | a body was reborn | nothing more: the body's record starts again from it |

**Holes by kind.** The harness raises holes of two kinds (`loop.HoleKind`): *harness* — it caught
the body's discipline (an answer off the wire, a missing `@U`, a gate sent back, a plan left open,
a failed-act streak); *infra* — the world's fault or a switch (the provider refused, a cut at
`max_tokens`, a drain or recall failure, `law.gate` off, a note that could not be filed). The
body's own `@?` lines are *self*. Only harness holes cost: `holes` counts them, `holeRate` is the
share of reports that carried one (never above 100%).

**Signals.** The fold over the window (`law.value.window`, default `28d`) gives each body:
`runs`, `reports`, `passes`, `fails`, `neutral`, `holes`, `passRate`, `failRate`, `holeRate`,
`landRate` (the loop closed `DONE`, the word is no fail, no gate failed), `intelligence`
(passRate less holeRate, floored at zero), `gateRate` (first-try gate passes, authors),
`verifyGreenRate` (verify lines green at the gate, authors), `gateRight` (verdicts held and not
overturned, gate holders), `verdicts`, `overturned`, `slimeHealth` (100 less the mean strain of the
bodies it dispatched, strain read as a percent of `law.stress.escalateAt`), `planLandRate` and
`reaskRate` (its dispatches that landed / that re-commissioned a body after a council warrant,
commanders), `askDoneRate` (the throne: human turns closed with a pass and not reopened — the same
paths rewritten — within `law.value.reopenTurns` turns), `upheldRate` (reports with findings that
landed), `findings`, `thrashRate`, `momentum`, `strain`, `strainPct`, `tokens`, `tenure` (whole
days since the body's first birth in the births ledger, `attic/births/<body>.jsonl`; a body born
before the ledger is grandparented — the commit that added its doc stands in), and `score`. A
rate on no basis reads zero.

**Families.** A rank's flags pick its family, in order: the throne → `root`; authors that hold
their own gate under the throne → `steward` (kijin); authors with no gate → `author` (slime); a
gate that commands → `gate` (orc, high-orc); a commander with no gate → `planner` (elf, high-elf);
a read-only shelf → `auditor` (dark-elf); else `generic`. `ranks.<r>.family` names one outright.

**Scores and meters.** A race's score is an arithmetic expression over signals (`+ - * /`,
numbers, parens), clamped 0..100: its own `ranks.<r>.value.score`, else its family's
(`law.value.families.<f>.score`, else the binary's):

| family | score | meter (ok · warn) |
|---|---|---|
| root | `askDoneRate` | Asks done 80 · 50 |
| planner | `0.6*planLandRate - 0.2*reaskRate + 0.2*landRate` | Plans landed 80 · 50 |
| gate | `0.5*gateRight + 0.3*slimeHealth + 0.2*landRate` | Gate right 90 · 70 |
| author | `0.6*landRate + 0.3*gateRate - 0.1*holeRate` | Landed 80 · 50 |
| steward | `0.7*verifyGreenRate + 0.3*landRate` | Verify green 90 · 70 |
| auditor | `0.7*upheldRate + 0.3*passRate` | Upheld 80 · 50 |
| generic | `0.8*passRate - 0.2*holeRate` | Pass 90 · 50 |

A body with fewer reports than its race's `minRecord` (`law.value.minRecord`, 5) is
low-confidence: shown, never ranked. A score, a meter or an `ascendsWhen` that reads a signal the
race can never produce (a read-only race scored on `gateRight`) loads with a hole naming it. The
meter is `{signal, label, scale: percent|count|momentum, bands: {ok, warn}, direction: up|down}`;
one renderer (`world.Meter`, `value.Render`) bands it for every surface. `isekai config explain
value` prints each race's resolved law with the origin of every part.

## Ascension

After every Court report the world folds the value ledger and evaluates each ascended rank's
`ascendsWhen` (`world/ascend.go`, `world/rank_eval.go`) against the base body's signals — every
signal above, `score` included, with arithmetic inside the comparisons.

The default band lives in config alone (`config.DefaultAscendElf`, `config.DefaultAscendOrc`;
`ranks.high-elf.ascendsWhen`, `ranks.high-orc.ascendsWhen`; an empty string turns ascension off,
and `status` lists that switch):

```
ascend orc when reports >= 20 && passRate >= 90% && holeRate <= 10%
```

— a long record, nearly all landed, at most one report in ten carrying a harness hole.

The record is the value ledger, replayed at `Open`, so a record earned over many sessions counts.

A crossing mints the ascended body once (the ledger's `ascend` row and the risen body on the
roster both remember):

1. **The body rises whole.** `orc-engine` becomes `high-orc-engine`: its doc is carried into
   `.isekai/high-orc/engine/README.md` — every trait, territory, verify line and Mind (Nature 5:
   "an ascended creature keeps every trait and territory it held before") — renamed, its rank line
   `high_orc`, an `Ascendant of: orc-engine` line, and an Ascension section with the condition and
   the reading that crossed it.
2. **The base's seat retires** (`- **Retired:** true — ascended <date>: the body rose as
   high-orc-engine …`): the roster drops it, the ontology keeps it as memory. One body holds the
   soil, so territories stay disjoint. A creature whose `Reports to` names the retired seat
   reports to its ascendant (the roster resolves the succession), so `slime-wire` now answers to
   `high-orc-engine`, which may order it.
3. **It is logged** in `log.md` through the append-only writer (Law 4), authored by the risen
   body: the condition, both docs, and `earned: reports 20, passes 19, …`.
4. **It is announced**: the dispatcher's report carries `[ascension] orc-engine rose as
   high-orc-engine …`; the roster is reloaded so the next commission reaches the risen body.

A mint that cannot land (the ascended rank configured onto its base's dir or prefix, a doc
already standing, a write failing) is a hole: it is appended to `log.md` and rides the report as
`@? ascension hole — …`. The base is retired only once the new doc stands.

## One order, down and back

The ask to rimuru: *"the wire should read `@W <path>` lines into `Report.Wrote`."*

```
human ── ask ──► RIMURU (session, depth 0)
                   │ dispatch elf-engine  "@ASK draft +unsaid …"
                   ▼
                 ELF-ENGINE (court, depth 1)
                   │ plan: [1 wire parses @W → orc-engine]
                   │ dispatch orc-engine  "@ASK verdict +unsaid …"
                   ▼
                 ORC-ENGINE (court, depth 2)       slime-wire's doc already holds
                   │ dispatch slime-wire            `go test -count=1 ./wire/...`
                   ▼
                 SLIME-WIRE (court, depth 3)
                   writes wire/wire.go, wire/wire_test.go, .isekai/slime/wire/README.md
                   turn-end gate: right author ✓  verify (pre-turn lines) ✓
                                  duties @S ✓     doc moved ✓   → log.md "gate pass"
                   ▲
                   │ @S DONE  @F wire/wire.go:… @U territory … @E
                 ORC-ENGINE reads the report, @V on the landing
                   ▲
                   │ @S DONE  @V … @E          (courier envelope if > 600 bytes)
                 ELF-ENGINE ticks its deck
                   ▲
                   │ @S DONE  … @E
                 RIMURU answers the human in plain language
```

1. **Rimuru** reads the ask; `onto` names `wire/` as `slime-wire`'s, in `elf-engine`'s lane. It
   may not write `wire/` (throne never authors) and may reach only its direct children, so it
   dispatches `elf-engine`.
2. **elf-engine** writes its deck — one item, `who: orc-engine` — and dispatches `orc-engine`
   (an elf may not reach `slime-wire` under strict flow).
3. **orc-engine** checks `slime-wire`'s verify is coded — it is, so the order passes the TDD
   precondition — and dispatches `slime-wire`.
4. **slime-wire**, at the chain's floor, holds no dispatch. It edits `wire/wire.go` and its test
   (any path outside `wire/` would be refused at the step), moves its README's traits in the same
   wave, runs `go test ./wire/...`, and answers on the wire. Its gate runs the verify line that
   stood before the turn, then appends the verdict to `log.md`. Its `@U territory` line lands as
   an ontology fact and a shared note when the orc's record beat reads the report.
5. **orc-engine**'s dispatch step returns `@S DONE` and the paths written; they are recorded as
   court work, so the orc's own gate has nothing of its own to judge. It answers `@S DONE` with
   its verdict.
6. **elf-engine** ticks its deck and reports; **rimuru** tells the human what landed and where.

Had the slime's gate failed twice, it would have reported `@S FAIL — …`, the orc's dispatch step
would have come back failed, and the slime's strain would have grown by its re-asks plus one;
once the strain reaches 5, the report the orc receives carries the council's warrant, and the
orc answers as manager — `@INSTRUCT`, `@SPLIT`, or `@ESCALATE` up to the elf.

## Where the creed reads differently

- **"No change lands without its Orc's pass. The Orc checks…"** — the four checks are code, run
  by the harness at the end of the authoring body's own turn; the orc never runs them as a model.
  Its part is coding the verify lines beforehand (enforced at dispatch) and judging the report.
- **"Most Elves, Orcs and Slimes run as Court … A Kijin's Body is always a Keeper"** — there is
  no persistent body: every body, a kijin's included, is dispatched as a Court with a fresh
  context, and the rank table carries no `Body` column (config's `ranks.<r>.body` is a hole).
- **Ascension "distilled through Memory … with the traits that earned it"** — ascension is a
  metrics condition over Court reports, minted by the harness; `log.md` records the metrics that
  earned it, not distilled thoughts.
- **Rimuru "works through the Elf by default"** — with `law.territory.throneNeverAuthors` on
  (the default) the throne's writes are refused outright, the law file and the dimension manifest
  excepted; working through the elf is the only lawful path, not a preference.
- **Stress "read from an instrument … a failing test, a growing error rate, a timeout"** — the
  council's strain is the run's own governance count (failed acts, gate and wire re-asks, the
  landing); relief at `relieveAt` is automatic, splitting or retiring the creature's doc without
  a second naming (Nature 4's "named twice" does not gate it).
