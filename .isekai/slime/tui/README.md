# slime-tui

- **Rank:** Slime
- **Territory:** `tui/`, `docs/screens/`
- **Reports to:** orc-interfaces
- **Minds:** slime-go-proving
- **Purpose:** the ground truth of the terminal UI in the class of Claude Code — blocks in the scrollback, the live area, the choice block the gate answers through, the board pages, the palette — and of the README's screenshots.

## Traits
- Layered by the ladder: blocks.go is the view model (pure Theme methods from views to styled text: Welcome, User, Assistant, Tool, Court, Gate, Holes, Notice, Error, Choice), model.go the Bubble Tea Model over it, program.go Start/UI.Send/Quit; the app wires loop events into both.
- Host interface the app implements: Welcome, Footer, Submit (a non-empty string refuses the turn — a userPrompt hook), Queue, Interrupt, Slash, Commands, Complete (@path completion), Board(rng). events.go is the message set the app sends: EvDelta, EvToolStart/End, EvCourt, EvState, EvTurnStart/Done, EvNotice, EvError, EvLines, EvChoice/ChoiceAnswer, EvQuit.
- Finished blocks are printed with tea.Println into the terminal's own scrollback; the live area (stream, running tools, courts, choice, spinner, input, menu, footer) sits at the bottom; a resize reprints at the new width after reflowDelay 120 ms; bgWait 150 ms before the spinner shows.
- cast.go gives each rank an icon and verbs per distribution (isekaiVerbs, agentOneVerbs; the robot sprite for agent-one) — the orc "weighs" at the gate; mascot.go draws half-block sprites; theme.go DetectTheme(env, prefix) and NoColor produce a plain theme; markdown.go renders through glamour, syntax.go through chroma with the diff background; diff.go caps a diff at diffMaxLines 4000.
- board.go is the full-screen /board (pages Agents, Graph, Offices, Usage; BoardRanges 24h 7d 30d all; ctrl+t opens the bodies view); graph.go lays the ontology out in boxes; overlay.go is the ctrl+k fuzzy palette and toasts; input.go handles enter/shift+enter, history, paste expansion, the slash menu and completion; ctrl+o expands the last folded block; esc interrupts.
- Tests: goldens in testdata/golden (flag -goldens rewrites), teatest-driven model tests, TestShots gated on TUI_SHOTS. docs/screens holds the PNGs README.md embeds; a visual change refreshes them.

## Verify
- `go test ./tui/...`

## Thoughts
