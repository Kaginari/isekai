# Changelog

All notable changes to Isekai are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow
[Semantic Versioning](https://semver.org).

## [Unreleased]

### Added

- **Light a whole race on the dashboard's net, and take the net full screen.** Each race's
  heading in the 2D net, and each race in the 3D net's key, is now a button: press it to light
  every creature of that race and dim the rest; press again to let go. A Full screen button gives
  the net the whole screen in either view; Escape brings it back.
- **The dashboard's overview shows the plan, the shared board and the files written.** Under the
  neural net, a switch picks one of four slices: the timeline, the plan (your session's deck, step
  by step, or the one in hand), the board — the work waiting on the shared queue, every creature's
  plan deck with its progress, and the last verdict each creature returned — and the files written
  this session. Each tab shows its count, and the switch remembers your pick. On a phone they
  stack. A new `/api/board` serves the board.
- **A court that escalates, fails or is denied says so loudly.** In the terminal, a court landing
  ESCALATE, FAIL or DENIED now prints one bright line under its card naming the verdict, who sent
  it and to whom (the creature that dispatched it), and why — the first hole in its report. The
  sidebar's court list shows that verdict instead of a green DONE.
- **Traces for your observability stack.** Set `instruments.otlp` in your global config or
  `.isekai/config.local.yaml` and every turn, court and tool call is sent as OpenTelemetry spans
  to the collector you name (OTLP over HTTP, so Jaeger, Grafana Tempo and any other OpenTelemetry
  backend read it). A turn is named after the creature's rank and name, a court's work nests
  under the dispatch that started it, and each tool call that actually ran gets its own span,
  carrying its input and output (cut at 4 KB) — the collector sees what the model sees.
  It is off by default; it never switches on from a repository's own config file, environment
  variables or flags; API keys are named by environment variable, never written in the file. If
  the collector is down or refuses the key, you get one notice in the terminal or on the board —
  not a stream of errors.
- **Two knobs for the shipped ascension rule.** `law.ascend.greenFloor` adds a minimum share of
  runs that landed cleanly (`landRate`, in percent), and `law.ascend.tenure` a minimum number of
  days since the creature was born, to the built-in rule for rising to high-elf or high-orc. Both
  are 0 by default, which leaves the rule exactly as it was; a world that wrote its own rule keeps
  it untouched.
- **Every creature has a birth record.** Each time a creature is born — written by its parent,
  split off by the council, risen by ascension or remade by evolution — one line is added to
  `.isekai/attic/births/<creature>.jsonl` saying when, how and from whom. A new `tenure` value
  reads how many days a creature has existed, so an ascension rule can ask for experience over
  time (`tenure >= 14`). Creatures born before this record existed count from the commit that
  added their README.
- **Retiring a creature looks after the ones below it.** When the council splits or retires a
  creature that others report to, each of them is moved to the right new manager — the half of
  the split that holds its files, or the retired creature's own manager — and the retired
  creature's README names who took its place. A creature that still names the retired one is
  routed to that successor, and `onto check` reports any creature left reporting to a retired
  one with no successor. The files the council rewrote no longer count against the body whose
  turn it happened in.
- **A parent can write its new child's README.** An orc founding a slime for part of its zone, or
  the throne founding an elf, used to fail the end-of-turn gate as if it had written another
  creature's file. The gate now lets it through while the README is new (not yet committed), the
  child reports to the writer, and everything the child claims exists and belongs to the
  writer's line or to nobody. Once the README is committed it is the child's own.
- **Stall guards.** A run that reads the same thing and gets the same answer, with nothing written
  since, gets a note the second time and checkpoints the third (`law.guts.repin`, 3); a file
  changed by another hand reads as new. A turn that runs 40 model calls
  with no write or dispatch checkpoints too, warned at three quarters (`law.guts.actBudget`).
  Both stops are resumable, reset at every new ask and every landing, and a negative value turns
  them off.
- **A stream that stops is cut.** A streamed answer that falls silent for two minutes after it has
  begun ends with "stream stalled" instead of waiting out the client's ten-minute timeout (a slow
  first token on a queued or prefilling server is left alone), and a cancelled call
  returns the cancellation at once, whatever the transport.
- **A gateway restart no longer ends the run.** When a model call fails for a passing reason (the
  gateway refusing connections or answering 5xx for longer than the per-request retries last, or
  a streamed answer cut off or stalled halfway), the whole call is asked again after 5 seconds,
  then 15 — fallback model included. A gateway's 504, or a 503 whose text speaks of a timeout, is
  asked again like any 5xx; a refusal, a missing model, the client's own timeout or a cancelled
  run is never asked again.
- **A shared work queue.** Work ordered down the ranks now has one place to wait that every
  session, every body and the human can read: a plain file of snapshots, one line per move, never
  rewritten. An item is held, then claimed by the body it was ordered to, then closed as done or
  dropped. Two hands moving the same item at once cannot both win: the second is told it lost
  and to read again; a crash that leaves half a line at the end of the file costs no later move.
  `isekai q` reads and moves it — `list`, `show`, `add`, `claim`, `touch`, `requeue`, `close` —
  with `--json` for scripts; it lives on the world's shared shelf
  (`.isekai/shared/queue.jsonl`), which no body may write by hand. Bodies get a `queue` tool to
  read it and to note progress on, hand back or close the work they hold; reading is free, and
  only the holder (or a rank above it) may move an item (`tools.queue.enabled`, on by default).
- **A `fold` tool brings folded output back.** Given the name a folded line quotes, it returns
  the original tool output exactly as it was, checked against its name so altered bytes are never
  passed off as the original. It only reads the fold folder. Every rank has it
  (`tools.fold.enabled`, on by default).
- **Reading a folder lists it.** `read` on a directory now answers with what is in it (folders
  marked with a slash, and a pointer to `skim` when it holds Go code) instead of failing with
  "is a directory" and costing a second call.
- **`read` takes a range in the path, in every form.** `f.go:40:20` (from line 40, 20 lines) and
  `f.go:-20` (the last 20 lines) now work for a single file as they did in a batch, and `:-20`
  in a batch is the tail instead of the whole file. A file whose real name has a colon in it is
  still read as named. A mistyped folder now gets a "did you mean" too.
- **Shell reads are pointed at `read`.** When a body prints a file with `cat`, `head`, `tail` or
  `sed -n '10,40p'`, the command still runs, and its output ends with one line naming the `read`
  call that gives the same lines numbered and is not re-sent when unchanged
  (`tools.preferFileTools`: `hint`, the default, or `off`). Nothing is ever refused for it.
- **Orders wait in the open.** Work a background court orders down the ranks (`@$`) is held on the
  shared queue until the session's next wake claims it in the ordered body's name and opens the
  next turn with it; the terminal's status bar shows how many are held. Two sessions on one world
  never both take the same order, and the terminal now runs these orders as the line session
  always did. An order the session may not dispatch itself (strict layer flow, a body outside its
  line) is not claimed for a body nobody will start: the session takes it, with the reason noted
  on the item, and is asked to route it down the chain and then close or requeue it. When the
  body reports DONE its item closes; any other answer leaves it claimed with the outcome noted.

### Changed

- **The law says whose a creature's seat is.** `ISEKAI.md`'s gate now states what the harness
  already does: a creature's own directory stays its own under a broader territory claim, an edit
  there owes no other creature's doc, and the parent that founds a creature may write its first
  README.

- **The law gains its negative half.** `ISEKAI.md` now says, one line per race, what each never
  does — each line a restatement of a rule already there, in the form a creature checks before it
  acts.

- **`make build` and `make install` stamp the binary.** `isekai version` on a local build now
  names the tag, commit and build date it came from, the way a release does, instead of
  `dev (none, unknown)`.

- **The dashboard's ID card opens when you ask for it.** On the overview a click on a neuron now
  only selects it and lights its path; double-click it (in 2D or 3D) or press Card to open its ID
  card. The card sits above everything on the page — the console and the files drawer included —
  stays open and follows whatever you select next, and closes with ✕, Escape or a click outside
  it, leaving the neuron selected. On a phone it opens as a sheet at the bottom of the screen.
- **An idle terminal stays idle.** With nothing running, the terminal now wakes only for the
  sidebar slime's blink instead of five times a second, and each redraw reuses every row that
  did not change, so a session left open costs a small fraction of the CPU it did. Anything that
  starts to move — a turn, a court, a toast, a question waiting on you — brings the full frame
  rate back at once.
- **Recall spreads its seats.** The memories handed to a model at each turn (and by the `recall`
  tool) now take at most two from any one file, so a long chronicle whose every entry matches no
  longer fills all five places and hides the one doc that answers. The `memory recall` command
  keeps its plain ranking, identical to `memory.js`.
- **Folded tool output is kept, not lost.** The bytes behind a folded line are saved once under
  `.isekai/tmp/fold/`, named by their content, and the line says where — so the exact output can
  be read back instead of running the tool again (which may cost a slow command, or give a
  different answer). The same output folded twice is kept once; nothing there is ever deleted.
- **A folded tool output says what it was.** Old tool output that the model no longer sees in full
  now leaves a line naming the tool, what it ran on (the file, the pattern, the command) and its
  size — "read pkg/x.go: 412 lines, 9.8kB" — instead of only the tool's name, so a body can tell
  which of its forty reads it would have to repeat.
- **One shared shelf for the world's bookkeeping.** The council's ledgers and the value ledger now
  live beside the work queue in `.isekai/shared/` instead of `.isekai/instruments/`. An older world's
  files are moved over, whole, the first time it opens, and if an older binary writes the old file
  again, both are still read, their rows taken in the order they happened; the value ledger an
  older binary seeds a second time from the ascension tally is not counted twice.
- **A court's timeout names how it died:** a provider that never answered, a body that only read
  (a recon loop), or one that worked and ran long, with its turns, acts and writes.
- **The boards poll the ontology from a cache** keyed on the files it is read from: a poll with
  nothing changed costs about half a millisecond instead of a full load, and the terminal board
  keeps its Graph page.

### Fixed

- **The plan meter in the dashboard's header lights up.** It read a deck file the plan tool never
  writes, so it stayed hidden; it now shows the same deck as the overview's plan. The plan reader
  also no longer follows a name that climbs out of the plans folder.
- **The dashboard answers only your own machine, on every address.** The dashboard page itself
  (which carries the key the console's actions need), its stylesheets and scripts, the theme list,
  and the live-session and world summaries now refuse a request that does not come addressed to
  `localhost`, as the world's records already did. A web page elsewhere that points a borrowed
  domain name at your machine can no longer read any of them.
- **A creature's notes no longer claim whole folders.** Mentioning `src/` or `docs/` (or a pattern
  like `**/`) in the free text of a creature's README, near its own name, used to make it the
  owner of that entire folder. Only the Territory line can claim a whole top-level folder now.
  When two claims overlap, `onto check` says whether each came from the Territory line or from
  the text, and points at the line that made it.
- **A refused order says how to get there.** When a body may not order a creature outside its
  own line, the refusal now names a route that is checked step by step: the nearest creature in
  that line the body may order itself, or up through its own managers to where the two lines
  meet and back down. If no step-by-step route holds, it says so and hands the question to the
  human instead of guessing. A creature whose manager no longer exists is reported as such,
  never named as the one to ask, and the list of creatures a body may order leaves out anything
  the order would then be refused for, and names each one once.
- **The world notices docs that changed under it.** When another session, a checkout or a person
  edited a creature's README, the running session kept its old picture until the next gate: it
  could still send work to a creature that had been retired, or refuse a write the creature had
  just been granted. Sending work, building a body and checking a write now re-read the docs
  when they changed. What a body's own verify lines are trusted to be still changes only at the
  gate, so a body cannot pass by rewriting its own check.
- **One answer to "whose file is this?"** The gate, the write rules and `isekai onto owner` now
  read ownership the same way. A creature's own folder is always its own, even when another
  creature's claim covers the whole `.isekai/` tree, so editing your own README no longer owes
  someone else's doc. The shared shelf and the instruments shelf belong to no creature and are
  never billed to one. The scratch folder (`.isekai/tmp/`) is every body's own: writing there is
  lawful and owes no doc, except another body's plan deck, which is refused.
- **The test suite no longer leaves folders in `/tmp`.** Two dashboard tests let a turn keep
  writing after their scratch folder was swept, so it came back; the end-to-end binaries were
  built into a folder nothing removed. Each run now leaves nothing behind.
- **The "memory is out of date" notice says how to fix it.** A world that turned off automatic
  memory rebuilds is told once per turn that recall is running on an old index; the notice now
  names the command that rebuilds it (`isekai memory index`) instead of "rebuild by hand".
- **A lesson written twice is recalled once.** A note saved with `remember` and the log entry it
  repeats (or the same note kept for this folder and for the whole machine) used to take two of
  the five places a turn's recall offers; they now take one.
- **A file read again after its output was folded away is shown again.** Reading a file a second
  time used to answer "unchanged since the first read" even when that first read had already been
  folded out of what the model sees — leaving it with neither the text nor a way to get it short of
  asking for a different range. The second read now ships the text again when the first one has
  been folded; while the first is still in view, it is still withheld.
- **The record of who wrote what no longer grows for the life of a session.** The book that keeps
  a background court's writes off its dispatcher's account kept every file ever written; it now
  forgets a write as soon as no running turn could still need it.
- **No script can forge a line on the work queue.** Refusing shell writes onto `.isekai/shared/`
  only works for commands the harness can read; a script file, a path built in a variable or a
  Python one-liner still got through. The harness now seals the queue and both ledgers as it
  writes them, checks them after every act, and moves anything it did not write into
  `.isekai/shared/attic/` (kept, never deleted) before any reader sees it; the act that wrote
  it fails and says so. A file rewritten by hand is set aside whole and replaced by an empty
  one — to put it back, move it into place and delete its `.seal` file.
- **The ledgers survive a torn line too.** A crash that left half a line at the end of the
  council's ledgers or the value ledger glued the next row onto it, and that row was lost on the
  next start. Both now start a new row on a line of its own, as the work queue already did; the
  three files share one way of appending.
- **The shell is no side door onto the harness's files.** A body could append a line to the work
  queue (or to the engine's journals) with a shell redirect or `tee` — forging a claim on work
  ordered to someone else — because only the write tools were held to that soil. Now a command
  that writes into `.isekai/shared/` or the engine's journal folders is refused the same way,
  whether the path is written plainly, after a `cd`, through `$ISEKAI_ROOT`, a symlink or inside
  `sh -c '…'`, and the refusal points the body at its `queue` tool. A command that is only a
  redirection (`>file`) no longer crashes the classifier.
- **A dispatch to a body that does not exist fails at once.** In the background it used to be
  acknowledged as started, the refusal arriving only later as a failed report (and a second
  "spawn failed" after it). Now the dispatch itself is refused before anything starts, naming
  the closest real body when the name is a typo (`slime-auht` → slime-auth) and every body
  that may be dispatched from there; the same holds for an order the dispatch law forbids.
- **A court's writes are its own.** A court working in the background changes the same files its
  dispatcher's turn is watching; the dispatcher's gate used to see those changes on disk and
  judge them as the dispatcher's own (a false "belongs to …, written by rimuru"). A file a
  court's tool reported writing is now left to that court's gate, as long as the file still
  stands as the court left it; anything the dispatcher changes itself afterwards, even in the
  same turn, is still its own and is judged.
- **A cancelled gate is not a failed one.** When a run is stopped while its end-of-turn gate is
  still reading, the gate stops spawning checks and says "cancelled" instead of recording every
  check it skipped as a red failure against the body, and the turn is no longer sent back to the
  model to fix a failure that never happened. The writes it never judged are not forgotten: the
  next turn's gate judges them along with its own.
- **Each court keeps its own desk.** Every body's `desk` wrote the throne's file; a court now
  writes its run's, the one compaction and `isekai resume` read.
- **One body, one live court.** A second dispatch to a body whose court is still running is
  refused, naming the run, instead of two courts sharing one desk and scratch.
- **Court reports no longer overwrite each other.** They were stamped on a 12-hour clock (3 a.m.
  and 3 p.m. shared a name); each is now named for its run, beside its journal.
- **Recall sends each anchor once** when two sources return the same one.
- **Patterns are not paths.** The scratch redirect leaves `sed -n /^func/p`, `awk /x/` and
  `grep -E '/a|/b'` alone (globs are still redirected), and the classifier no longer reads
  `sed -n 1,/^x/p` or `git show HEAD:path` as leaving the world; a brace expansion is expanded as
  the shell does (`/etc/{hosts,passwd}`), and `ls /etc/` and `cat </etc/hosts` now ask, as reads
  outside the world should.
- **`gofmt -l` and `goimports -d` are reads; `go fmt` and `-w` are writes.** `go fmt` was read
  as read-only.

### Removed

- **The isekai binary no longer carries a copy of agent-one's law.** Agent-one's law ships with
  the agent-one binary, from its own repository; the second copy here only drifted from it.
  `isekai --dist agent-one init` now refuses and names the agent-one binary instead of founding
  a workspace with a stale law.

## [1.0.0] - 2026-10-11

The first stable release. From here the configuration keys, the `isekai` commands and flags, and
the memory server's MCP tools (`recall`, `remember`, `memory_status`) change only with a major
version; a key or flag that goes away is deprecated for a minor release first.

### Changed

- **The 3D net's neurons are points of light** — scaled down to about a third of their old size,
  growing with the mass under them, so a large world's net reads as a constellation rather than a
  cluster of spheres.

## [0.3.1] - 2026-10-11

### Added

- **The world's memory over MCP.** `isekai mcp serve` is an MCP server on stdio — `recall`,
  `remember`, `memory_status` — so Claude Code, OpenCode or any MCP client recalls what the world
  remembers and files notes every later session finds. `claude mcp add isekai-memory -- isekai mcp serve`.
- **`/board` in the terminal wears the session's frame** and theme: the rounded border, the raised
  header with the page tabs as chips, the page on the plane, the keys in the status bar.
- **The pinned neuron's card is an ID block** on the dashboard's net, with ✕ to put it away while
  the neuron stays selected.

### Fixed

- **`--docker-agent` builds a working runtime.** The image is built on the host network (a bridge
  that cannot resolve failed every apt step) and is the full one — Go, Node.js, Python, build tools,
  Chromium — the same as `docker/runtime/`.
- **The scratch redirect leaves the world's own paths alone** wherever the world lives: a world at
  `/app` (a benchmark container) or under `/tmp` had its own files herded into scratch.
- **Two data races:** a background Court's spend landing on its parent's session, and the terminal
  reading its session while a resumed run swapped it — both under a lock now.
- **The bench images default to Docker Hub**, a private mirror by `REGISTRY`.

## [0.3.0] - 2026-10-11

### Highlights

**The world measures, answers and heals itself — and looks the part.** Every race is valued by its
own job (`law.value`, families, meters); a creature whose runs never land draws the council, then
relief, then evolution (`law.evolution`, auto by default). The forge mints the skills a run proved.
The terminal is one framed, themed cockpit with a theme picker, the dashboard one page with the neural
net, the 3D brain and a config tab that says what every key does — eight palettes shared by both. The
sandbox hides the host's processes and your credentials (and runs under Seatbelt on macOS),
`--docker-agent` replaces `--containered`, the board guards every read, and private package registries
log in — or, with `policy: private-only`, are the only place packages may come from.

**The world versions and defends its own seat.** A world root bearing `.isekai/dimension.yaml`
outside any work-tree is versioned `git init -b main` once at open — a world is born under git, by
the harness itself. And the gate now attributes the throne's own seat files (the dimension manifest
and the law doc) to Rimuru: before, a gate-holder whose territory reached `.isekai/` silently owned
them through Owner's fallback, so Rimuru's lawful relocation of the manifest failed right-author and
Vitality — demanding an orc's README for the throne's own file. The exemption mirrors the creature
doc rule and polices non-throne bodies exactly as before.

### Added

- **A full runtime image for `--docker-agent`.** `docker/runtime/` builds the machine the agent works
  in: Go, Node.js LTS with npm, pnpm and yarn, Python with pip and uv, build essentials, gopls,
  staticcheck, ruff, pytest and a headless Chromium, every cache and global install under `/tmp` for
  the session's uid; published as `montabouagina/kaginari:v0.3.0` (`--image`, or
  `registry.containers.image`). The README's terminal screens are the current TUI's, written by
  `README_SHOTS`.

- **Private package registries log in, and can be the only source.** `registry.packages.auth` (and
  `npmAuth`, `pipAuth`, `goAuth` over it) is the login each manager presents, written at every start
  under the world's `tmp/registry/`, owner-only: npm an npmrc whose bearer token stays a `${VAR}`
  reference, pip and uv the login inside their index URL, go a netrc for its proxy host — checked
  against a live server for all three. Env-first; a literal token is refused from the tracked config.
  `registry.packages.policy: private-only` keeps every manager on those URLs: GOPROXY loses `direct`,
  a manager with no private URL is switched off, and the guard (and the Claude Code hook) refuses a
  command that points a manager elsewhere, installs from a URL or a repository, or fetches a public
  registry host.
- **Every config key says what it does.** `config.Describe` holds a line per key (a switch says
  on/off first); `/api/config` sends them and the dashboard's config tree shows each under its key
  (a section's beside its name). A test fails a key of the defaults that has none.
- **The theme picker.** `/theme` alone (and the palette's `theme`) opens a picker: every palette with
  its swatches, the one in use marked, the mode under them; moving previews live, enter keeps it
  (the dashboard following), esc goes back. `/settings` lists every palette too.
- **The dashboard in the README**: the neural net with a creature's card, the 3D brain, the config tab.

- **Themes for both faces.** One package of named palettes (`theme/`) dresses the terminal and
  the dashboard, the same roles in a dark and a light variant: omoro.io's Cyberpunk (unchanged,
  the default) and six more of its cyberpunk family — Neon Dystopia, Corporate Rebellion, Rain
  City, Synthwave, Glitch, Chromecore — plus Realm, the dashboard's own neon HUD. Each role is
  derived from the palette's swatches, every retune documented, and every text role holds 4.5:1
  on every surface of every palette in both modes (a table test; the terminal's own over every
  gradient row). `ui.theme` and `ui.mode` (dark · light · system) pick the look; an unknown palette
  is refused at load with the list; `isekai config explain theme`. `/theme <name>
  [dark|light|system]`, `/settings` and the dashboard's theme and mode menus switch it live for
  every surface and write it to the world's `config.local.yaml`; the dashboard's colour roles are
  generated at `/theme.css` and swapped in place (a `theme` event on the session stream carries a
  terminal switch; `GET`/`POST /api/theme`), the 3D net repainting with them.
- **Value meters in the terminal.** The sidebar's races tile reads each race's value meter (its
  label, its reading, a bar in its band; confident races first by their mean, a short record
  dimmed and unranked, a custom race in its family's hue), each Court's row its body's reading,
  the vitals the throne's score and the body nearest its ascension; the context and plan gauges
  band by config's meters (`cfg.Meters()`, `ui.meters`) instead of numbers of their own.

- **Stress-driven evolution.** A creature that is stressed — council strain at
  `law.evolution.stressAt` (the council's `escalateAt` by default), its last `failStreak` gate
  verdicts failed, or a doom checkpoint — while its Thoughts desk (its doc's plus its worn
  Minds') holds more than `thoughtsMax` (7) entries evolves in the same wave, weighed after every
  Court's report and at run start. One arm is chosen from evidence and its reason logged:
  **split** (a large zone whose failures and Thoughts cluster in two sub-directories becomes two
  young named for their halves, each with its Thoughts and the survey's traits and verify; an
  orc's slimes are divided with it; the parent retires), **rebirth** (a genome its own verify or
  the tree contradicts is written again from the code, the old doc rests in
  `<world>/attic/<name>@<date>.md`, a `rebirth` row starts its value record again) or **mutate**
  (each repeated lesson becomes a trait, a verify line lands where tests cover it, a recurring how
  is minted as a crafted skill the body wears, at `law.forge.evolvedStatus`). Every plan is
  staged and proven before it lands — its verify green, the ontology over an overlay of the new
  docs no worse, every body on the soil meant — else the old docs stand and a hole is logged. An
  `evolution` row in the value ledger carries the cooldown (`cooldownRuns`, 10) across restarts.
  `mode: propose` queues the diff for `isekai evolve list | show | apply | reject`. The session
  stream carries an `evolution` event: the terminal toasts it and keeps it in the sidebar, the
  dashboard adds a timeline row, rings the neuron in the arm's colour and reloads the colony;
  `isekai status` lists recent evolutions and `/api/value` each body's desk, stress and
  cooldown. `config explain evolution` prints the law with its origins; switched off, it is an
  honesty finding.

- **Each race is valued by its own job.** One definition of success (`law.value`: pass, fail and
  neutral words; a run's word is the loop's when it did not close `DONE`, else the body's `@S`) is
  read by the ascension fold, the council, the dashboard's leaderboard and activity, and the loop
  journal's new `word`. Every run of every body is one row of an append-only value ledger
  (`instruments/value.jsonl`), kept whether or not anything ascends, replayed at open and seeded
  once from the old ascension tally. A gate verdict credits the body that held the gate, and a
  pass later failed on the same paths is an overturn. Holes split by kind — infra, harness, self —
  and only harness holes cost; `holeRate` is per report and never passes 100%; the human's
  `DENIED` and an interrupted run are neutral. A rank's flags give it a family (root, planner,
  gate, author, steward, auditor, generic) and the family a score — an arithmetic expression over
  the ledger's signals (`landRate`, `gateRate`, `gateRight`, `slimeHealth`, `planLandRate`,
  `askDoneRate`, …) — and a meter with config's bands; `ranks.<r>.value` overrides both,
  `ascendsWhen` reads `score` and every signal with arithmetic, and a signal a race can never
  produce is a load-time hole. `/api/value` and `/api/meters` serve it; `isekai config explain
  value` prints each race's law with its origins.
- **The dashboard ranks by value.** The Metrics leaderboard sorts by score (score, the race's own
  measure in its band, strain, work, tokens; a body under its race's record hatched and unranked)
  under a row of each race's meter; every neuron wears a band ring; the 3D health tint is the
  ranked bodies' score; every band the page draws comes from `/api/meters` (the context bands from
  `law.budget`, the one line the context instrument's NEAR zone reads too).
- **Faces and sizes.** One portrait resolver (`board/icons.go`): the world's own
  `portraits/<rank>.png`, `ranks.<r>.portrait`, the embedded race portrait, the family's glyph
  (Tabler outline icons, MIT), the generic glyph; `/portraits/<rank>` serves it and `/api/colony`
  nodes carry `icon`, `family`, `value` and `mass` (lines of code owned). Every creature wears its
  face at every depth, the focus card shows it large, the minds wear Lucide's `brain-circuit`
  (ISC); neurons are small and sized by the code they own, and the 3D tint switch carries its key.
- **The guts speak any language.** `law.guts.benches` (go, pytest, the JS runners, cargo, make —
  each with red and green lexemes), `lexemes`, `weights`, `clamp` and `ring` are config; a reading
  journals a stable `class` (fail, neg-lexeme, echo, thrash, repeat, red-pin) and `lessons` groups
  by it.

- **The terminal UI wears the Cyberpunk palette, with volume.** The colours are one named palette
  (`tui/palette.go`, role → hex for dark and light, read by one constructor): omoro.io's Cyberpunk —
  Void ground, surfaces stepping toward Smog Purple, Scanner Cyan the primary, a Danger Red rose the
  human's accent, Amber Haze the warning, the plan in Smog Purple's violet — retuned where a swatch
  fails 4.5:1. Titles wear colour, each region its own gradient; a Court's card is tinted in its
  race's lane with a state chip, the plan its own violet family in the chat and the sidebar. Slabs
  gain volume: a gradient face, a gloss under the top edge, a two-tone bevel, a soft two-column
  shadow with quarter-block corners, a rim glow on the running Court, a fresh toast and the ask's
  well while typing. The conversation and the sidebar wear a draggable scrollbar while they
  overflow. The sidebar is a cockpit column: a header tile (sprite, world, model, state), the
  vitals as a grid of meters, each race's share of the work, the courts live first, the plan's
  progress and current step, the files. The ask's well starts three rows tall with a
  placeholder, soft-wraps, grows to eight rows (or 35% of the window), then scrolls inside, its
  rule saying the draft's size and what enter does.

- **The 3D net is a brain.** The throne sits at the centre and each rank is a shell further out
  (elves inner, orcs middle, slimes on the outer cortex); each lane is a region of the brain, tinted
  by a faint glow and switchable to its four weeks of work (heat) or its runs' pass rate (health).
  A seeded force layout on two hemispheres keeps a creature near its parent and pulls related
  creatures together: the colony's new `relates` (code coupling between territories from their
  imports — Go, Python, JS/TS — and minds both wear, in `board/relate.go`), days worked together and
  runs commissioned together. The minds are a crown above the throne, their `wears` filaments
  falling in the mind's colour, glowing, with particles travelling down while the wearer works.
  Every rank is a lit solid, one instanced draw per shape — a star crystal in orbit rings for the
  throne, slender bipyramids for elves, hexagonal prisms for orcs, soft wobbling blobs for slimes, a
  tetrahedron for the kijin, a crown ring on an ascended rank, a torus for a mind — with a key in
  the lexicon's words. Synapses taper from the parent; a dispatch lights each synapse it crosses
  as a pulse with a trail and the report comes back, a firing neuron flashes and rings; dendrites,
  slow motes and a bloom pass give depth. Hovering a relates bond says why it holds, and "replay"
  scrubs through the last turn's signals. Still under reduced motion, and the idle frame budget
  holds.

- **The terminal UI has depth.** On a truecolor terminal the frame is layered planes drawn with
  colour and the block elements: the borders on the terminal's own ground; the conversation, the
  sidebar and the status bar each a plane of their own tint; the ask and each answer raised as
  slabs with a spine (magenta, the primary), a lit top face, a shaded edge and a soft drop shadow;
  tools and Courts as smaller chips in their class's or race's hue; the welcome as the hero slab on
  a primary → magenta gradient; the sidebar's sections as stacked tiles; the input sunk in a well;
  toasts, the ask card and the palette floating highest. One tonal scale stepped from the
  terminal's background, inverted on a light theme (white slabs on a grey ground, soft grey
  shadows); every text role holds 4.5:1 on every tint. A selection copies text, never the
  drawing. 256-colour and colourless terminals keep the flat frame.

- **The dashboard is the board's one page, at the root.** `http://127.0.0.1:<port>/` opens the
  cockpit; the tabs are Overview · Console · Neural net · Metrics · Court · Memory · Toolbox ·
  Relations · Tree · Config, and a deep link opens one (`/#memory`, `/#config`, `/#court/<run>`).
  Four tabs are new, in the realm HUD: **Court** — every journaled run, filtered by body, outcome
  or words, and one run whole (the commission, each step with the gate's class and reason, the
  holes); **Memory** — the three tiers as instruments, a read-only recall, the desks and notes, the
  memory's review on demand; **Toolbox** — the registry by kind and state with costs, offers and
  loads, the skills the session can wear, the minted bodies, the loads journal; **Config** — the
  resolved config as a tree with each value's layer, credentials shown as set but never their
  value, and the off-list. Ten tabs fit any width: what does not fit folds into a "more" menu,
  arrows walk the tabs. New read-only APIs: `/api/runs`, `/api/run`, `/api/memory`,
  `/api/memory/recall`, `/api/memory/review`, `/api/toolbox`, `/api/config` (loopback Host only).

- **A native sandbox on macOS: Seatbelt.** On a Mac every shell command, verify line, hook,
  custom tool and stdio MCP server now runs under `/usr/bin/sandbox-exec` with a profile that
  mirrors bwrap's fence: writes only to the world, its read-write paths and the temp dirs; no
  network but loopback and unix sockets unless the act passed the gate as outward; the home's
  credentials (plus the keychains, cookies and browser profiles) unreadable; other processes not
  inspectable or signalable; Apple Events, LaunchServices and the clipboard closed. Paths cross as
  `-D` parameters, never as profile text. `status` prints `sandbox: seatbelt (sandbox-exec)`, or
  `sandbox: none — sandbox-exec refused: …` when a macOS refuses the profile;
  `tools.bash.sandbox: none` opts out. README's Security section has the table and the limits (no
  pid namespace; allow-default with mach IPC open); `go test ./sandbox/ -run Seatbelt -v` on a Mac
  proves it there.

- **The terminal UI is one frame.** A window of 60×26 or more is a single box-drawn grid under the
  headroom row: a header (the brand on its gradient, the world, the model, the session and its
  state), the conversation titled in its edge, the input inside the main column only (never under
  the sidebar), a full-height sidebar, and a status bar (the context meter, tokens, cost, live
  bodies, key hints) — shared edges with real junctions, every row the window's width at every
  size. The sidebar is ruled sections (vitals beside the sprite, courts with their state as a
  chip, the plan deck in its office's lane, the files as a tree) that fold on a click or
  `alt+c`/`alt+p`/`alt+f`; `alt+s` or `/sidebar` hides it. The welcome is a compact card — the
  sprite beside the name, the lane legend as chips, a label/value grid with the offices folded to
  one line (`/offices` has them whole) and a row of where to start. The human's ask wears a
  magenta bar, the answer a `✦` in the gutter. The slash menu, completions and the `?` sheet rise
  over the conversation instead of reflowing it; `/settings` is a huh form for the theme, the
  mouse and the sidebar. The palette follows the dashboard's neon HUD — cyan primary, magenta for
  the human, borders quieter than text, deep-ink twins on light terminals, every text role at
  4.5:1. Clicks and drags read the frame's own geometry.

- **The dashboard's overview is a cockpit.** The neural net is the hero, the console docks beside
  it (drag the splitter to resize, fold it to a rail — both remembered), and a timeline of Courts
  and verdicts sits under the net (`/api/activity` now carries the newest runs, their wire headers
  dropped); a row focuses its neuron. The header gains the plan's progress. The folder tree leaves
  the grid for a drawer over every view — the files button, `t`, or Alt+T from the prompt; Esc or a
  click away closes it — and hovering a path lights the neuron that owns it. Every action pulses
  on its own neuron in 2D and 3D: reads ripple softly, writes pulse in the rank's colour, outward or
  destructive steps and a waiting gate ring amber, verdicts flare green or red; the throne breathes
  while the session runs.

- **The dashboard wears "realm", a neon HUD.** A near-black ground, neon cyan and magenta,
  hairline panels with corner brackets, Orbitron display type and tracked capital labels over
  Rajdhani body text and JetBrains Mono numbers (all SIL OFL, vendored and served by the board), a
  faint slow scanline and a one-time glitch on the brand — still under reduced motion. Dark is the
  default; light is a real light mode in the same language; the theme switch cycles dark, light
  and system. One palette of named roles (ground, panel, border, text, muted, primary, accent,
  ok/warn/danger, a hue per rank), every text colour at 4.5:1 or better in both themes. Metrics
  gains a leaderboard of the creatures by their work. `ui check` lets a page's own `theme.css`
  hold literal values beside the foundation's token files.

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

- **The forge (`@FORGE`, `/forge`, `isekai forge`).** A run that proves a repeating need carries
  its draft on the report's wire — `@FORGE skill:<name> | <when> | <how>` or
  `@FORGE tool:<name> | <summary> | <argv>` (holes `{{param}}`). The harness files each draft
  from the report it holds (the dispatcher's own result — a report past the 600-byte envelope or
  cut by `@CAP` keeps its draft; the salvage turn, the throne's final answer and the Claude Code
  Stop hook file theirs too) as a shared note tagged `forge` with the drafting body, its run and
  a content hash; a line relayed up the ranks is filed once, as its drafter's. The desk lists
  open drafts by that hash (`/forge`, `isekai forge list`: name, version, body, run, age, kind,
  a tool's class), shows what a mint would write (`show <id>`), and closes a draft with an
  append-only mark (`reject <id> <why>`). A skill mint writes
  `.isekai/crafted-skills/<name>/SKILL.md` in Claude Code's format (`name` in `[a-z0-9-]{1,64}`,
  `description` the when, the body the how) plus `version`, `status`, `by`, `born`, `minted`; it
  never overwrites — a newer draft of a crafted name shows as `v2` and lands only with
  `--version`, the old file kept under `versions/`, and a name another home wears is refused. A
  later body whose doc names the skill wears it (`@T mind <name>`); any body loads it with
  `skill{name}`. A tool mint writes `tools.custom.<name>` with its argv split by shell quoting (a
  pipeline refused), its class from the classifier (never a hard-coded `write`), an interpreter
  running a string (`bash -c`, `python -c`, `node -e`, `eval`) refused unless `--force`, an
  outward or destructive tool landing only on the human's yes (`--yes` headless), the YAML
  printed before it is written, and `crafted: {by, run, at, draft}` provenance. The `remember`
  tool refuses the forge's tags. `law.forge.demandAtSteps` (12; 0 = off, reported as switched
  off) asks any turn that long, once, for its draft — with or without `law.closingCheck`.

- **One skill registry.** `discover.SkillBases` is the one list of skill homes, in one
  precedence — the world's `crafted-skills`, its `base-skills`, the host's project dirs, its
  global dirs, then the binary's own — read alike by the session's discovery (`status`'s minds,
  the `skill` tool), the toolbox, memory and the ontology, under the distribution's world dir. A
  name with two homes is an `@?` hole naming both paths, never a silent shadow.

- **The container's trust (law.humanGate.trustContainer).** Inside the harness's own runtime
  container — `<PREFIX>DOCKER_AGENT=1`, set only by the harness's own container argv, spoofable by
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

### Security

- **The sandbox hides the host's processes and your credentials.** bwrap now gives every command its
  own pid, ipc and uts namespaces, and masks the home's credential paths (`~/.ssh`, `~/.gnupg`,
  `~/.aws`, `~/.kube`, `~/.docker`, `~/.netrc`, `~/.config/gcloud`, `~/.config/gh`, keyrings, browser
  profiles …) — they read as empty inside, while the world stays writable. README's new Security
  section says what the sandbox fences, what it does not, and how to fence a Mac.
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

- **A retired seat claims no soil.** The ontology keeps a retired creature's name, bonds and Minds
  but derives no territory from its doc, so the young of a split (or an ascension, or a council
  relief) never overlap it and `onto check` stays clean. The council's relief no longer fires on a
  body that just evolved, and relief now runs after the report's fold.
- **A rank has one spelling.** `rank.Canon` (lower case, hyphenated) is used by config, the
  world, the ontology's colony, the board, the portraits' file names and the TUI's keys;
  `dark_elf`, `darkelf` and `DarkElf` are read as `dark-elf`.
- **Config is the one source of the default ascension band**; the world's table carries none.
- **`status` names numeric and string switches** (`law.stress.escalateAt`, `.relieveAt`,
  `law.doomLoop.threshold`, `law.guts.benchThrashVisits`, `.digestEvery`,
  `law.budget.salvageTokens`, an empty `ascendsWhen`) like any `enabled: false`; a guts floor of 0
  is honoured and a positive one refused.

- **`--containered` is now `--docker-agent`** — the flag that runs the whole binary inside a Docker
  container; the old name still works and says what it is called now. The marker inside the
  container is `<PREFIX>DOCKER_AGENT`.
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

- **A theme pick is kept.** A world whose `config.local.yaml` is the scaffold's comments over `{}` refused
  every write (a block key after a flow document); the YAML editor now re-renders a flow-map document
  as a block, comments kept. A save that fails stays in the conversation instead of a passing status line.
- **`/settings`, `/sidebar` and the terminal's own commands are in the slash menu**, not only the palette.
- **The pinned creature's card floats over the net's corner** instead of the timeline's band; the
  timeline keeps more of the height.
- **The conversation follows its bottom across a resize.** A pane that shrank decided it had been
  scrolled up and stopped following (so `/status` and every later block landed out of sight); the
  welcome card heads the conversation even when a fast first turn lands before it.
- **The dashboard's stream journal and section-local config stay untracked** in a new world.

- The terminal's conversation pane kept following the bottom only while its height held: a
  smaller window or a draft growing the ask box read as the human scrolling up, and new blocks —
  a slash command's output among them — landed out of sight, the pane sometimes pinned to its top.
  The welcome now heads the conversation even when a first ask lands inside its background wait,
  and no block prints before the window's size is known.
- The dashboard's topbar no longer sets the page's width: its vitals row contains its inline
  size, so the shell stopped overflowing sideways (by 176px at 1024px wide).
- A background Court's report that arrived while the session's turn was closing was lost; it now
  waits in the inbox for the turn-done drain.

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

- **Throne-seat attribution at the gate.** The gate's right-author and Vitality checks exempt the
  throne's seat files (`.isekai/dimension.yaml`, `.isekai/isekai.md`) when the throne authors them —
  Owner's gate-holder fallback no longer lets an orc whose territory is `.isekai/` claim the throne's
  own files. A non-throne body writing them still fails right-author.

### Removed

- **The board pages and Bootstrap.** The server-rendered pages (Overview, Court, Usage, Colony,
  Memory, Toolbox, Log, Config), their templates, `static/` and the embedded Bootstrap are gone,
  and with them the dashboard's "board pages" link. `/dash` redirects to `/` (301); `/court`,
  `/memory`, `/toolbox`, `/config`, `/usage` and `/colony` redirect to their tab; `/log` is gone
  (the log stays in `log.md` and the memory's episodic tier).

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
