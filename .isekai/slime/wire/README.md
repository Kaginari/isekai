# slime-wire

- **Rank:** Slime
- **Territory:** `wire/`
- **Reports to:** orc-engine
- **Minds:** slime-go-proving, slime-wire-and-journals
- **Purpose:** the ground truth of Absolute Rule II in code: parsing and emitting the commission (@ROOT @SCOPE @ASK @CAP @DUMP @SIZE) and the report (@S @F @V @? @U @E), with @CAP enforced.

## Traits
- A report is recognised by its @S line (ParseReport returns ok only then); tags match tagLine `^@([A-Z]+|\?)(\s+(.*))?$`, so an @-word mid-line is prose, not a tag.
- DefaultCap is 2048 characters; Report.Emit(cap, dump) cuts a report over its ceiling and names the cut as a hole — the cut is visible, never silent. Commission.CapOrDefault applies the default when the commission names none.
- UnsaidKinds are exactly law, colony, territory; an @U line with another kind is a hole. memory.Unsaid and compact's landing use the same three words.
- Consumers outside this package: the openai provider's guided schema (WireSchema) and RenderWire, the world gate's Duties-done check (a commission on the wire needs an @S back), app's shell hooks (a hook's stdout beginning with @S is parsed as a report and journaled), the dispatch tool and the TUI's report view. A tag added here is a tag added in each of them.

## Verify
- `go test ./wire/...`

## Thoughts
