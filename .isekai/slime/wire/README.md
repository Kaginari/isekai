# slime-wire

- **Rank:** Slime
- **Territory:** `wire/`, `queue/`
- **Reports to:** orc-engine
- **Minds:** slime-go-proving, slime-wire-and-journals
- **Purpose:** the ground truth of Absolute Rule II in code: parsing and emitting the commission (@ROOT @SCOPE @ASK @CAP @DUMP @SIZE) and the report (@S @F @V @? @U @T @$ @FORGE @E), with @CAP enforced; and the shared work queue the @$ orders are held on (queue/).

## Traits
- A report is recognised by its @S line (ParseReport returns ok only then); tags match tagLine `^@([A-Z]+|[?$])(?:\s+(.*))?$`, so an @-word mid-line is prose. The raw pipe form (`S|PASS`, `F|path:ln|claim`) is read for S F V ? U E T $ only. Unrecognised lines are kept in Report.Other, in order.
- UnsaidKinds are exactly law, colony, territory; an @U whose first word is not one of them keeps Kind "" with the whole text. memory.Unsaid and compact's landing use the same three words.
- @$ `<body>: <ask>` is an Enqueue (body folded lower-case) — work ordered one layer down, held on the shared queue until the wake after the turn claims it; EnqueuesOf reads the same lines from any text, so a commission can state the relay it expects. A @$ without `<body>:` stays prose.
- @FORGE `<kind>:<name> | <text>` is a Forge draft, kind only tool or skill (folded lower), riding up for curation — the wire never mints it. A malformed @FORGE stays prose. Forge.Hash is its content address (fnv64 over kind, name and whitespace-folded text, 10 hex) — the same at every relay hop; ForgesOf reads @FORGE lines out of any text, prose included.
- DefaultCap is 2048 bytes; Commission.CapOrDefault applies it. Report.Emit(cap, dump) closes with an exact @E (it counts its own line) and, over the cap, drops lines in this order — Other, @T, @FORGE, @F, @V — never @S, @? or @U, and adds a `@CAP n: k lines cut` hole naming the dump path when set.
- Consumers outside this package: the openai provider's WireSchema/RenderWire, the loop's wire and +unsaid bounces, the world gate's Duties-done check, world/court.go and world/forge.go (Forges), app/claude.go (ForgesOf on the Stop hook's final text), app/court.go (Enqueues), app.go (EnqueuesOf on an ask), app's shell hooks, the dispatch tool and the TUI's report view. A tag added here is a tag added in each of them.
- queue/: Item snapshots (id q<n>, rev, op, state, rank, for, from, ask, owner, by, note, added, at) appended one JSON line per move to one file, never rewritten (Law 4); Fold seats a line only when it is an add at rev 1 or follows its item's last folded rev on a non-terminal item, so a racing second writer reads back a Refusal and never a double claim — no lock, no database. States held → claimed → done | dropped (terminal); claim takes held work ordered to the claimant (or May over it), touch/requeue/close take what the holder holds (or May over it), held work closes by whoever queued it. Ranked: claimed, held, closed; rank 0 first, then age. CLI (`q list [--all] | show | add | claim | touch | requeue | close [--drop]`, --as, --json) exits 2 on a Refusal, 1 on an unreadable file.

## Verify
- `go test -count=1 ./wire/... ./queue/...`

## Thoughts
- (none yet)
