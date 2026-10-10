# slime-wire

- **Rank:** Slime
- **Territory:** `wire/`
- **Reports to:** orc-engine
- **Minds:** slime-go-proving, slime-wire-and-journals
- **Purpose:** the ground truth of Absolute Rule II in code: parsing and emitting the commission (@ROOT @SCOPE @ASK @CAP @DUMP @SIZE) and the report (@S @F @V @? @U @T @$ @FORGE @E), with @CAP enforced.

## Traits
- A report is recognised by its @S line (ParseReport returns ok only then); tags match tagLine `^@([A-Z]+|[?$])(?:\s+(.*))?$`, so an @-word mid-line is prose. The raw pipe form (`S|PASS`, `F|path:ln|claim`) is read for S F V ? U E T $ only. Unrecognised lines are kept in Report.Other, in order.
- UnsaidKinds are exactly law, colony, territory; an @U whose first word is not one of them keeps Kind "" with the whole text. memory.Unsaid and compact's landing use the same three words.
- @$ `<body>: <ask>` is an Enqueue (body folded lower-case) — work ordered one layer down that the Court drains after the turn; EnqueuesOf reads the same lines from any text, so a commission can state the relay it expects. A @$ without `<body>:` stays prose.
- @FORGE `<kind>:<name> | <text>` is a Forge draft, kind only tool or skill (folded lower), riding up for curation — the wire never mints it. A malformed @FORGE stays prose.
- DefaultCap is 2048 bytes; Commission.CapOrDefault applies it. Report.Emit(cap, dump) closes with an exact @E (it counts its own line) and, over the cap, drops lines in this order — Other, @T, @FORGE, @F, @V — never @S, @? or @U, and adds a `@CAP n: k lines cut` hole naming the dump path when set.
- Consumers outside this package: the openai provider's WireSchema/RenderWire, the loop's wire and +unsaid bounces, the world gate's Duties-done check, world/hooks.go (Forges), app/court.go (Enqueues), app.go (EnqueuesOf on an ask), app's shell hooks, the dispatch tool and the TUI's report view. A tag added here is a tag added in each of them.

## Verify
- `go test -count=1 ./wire/...`

## Thoughts
- (none yet)
