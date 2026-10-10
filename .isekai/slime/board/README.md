# slime-board

- **Rank:** Slime
- **Territory:** `board/`
- **Reports to:** orc-interfaces
- **Minds:** slime-go-proving, slime-wire-and-journals
- **Purpose:** the ground truth of the board — the world seen over HTTP on 127.0.0.1 from the same instruments the session writes — and the dashboard's live routes.

## Traits
- An http.Handler with everything embedded (templates, static, assets, web): a page works offline. Default port 7411 (config board.addr); README.md in this directory documents the pages.
- Routes: / overview, /court, /usage, /colony, /memory, /toolbox, /log, /config, /events (SSE hub; watch() polls directory and file signatures and broadcasts), /api/doc, /api/court, /api/usage, /api/colony, /assets/ and /static/, plus liveRoutes: /dash and its stream, and POST acts (ask, answer, interrupt) that pass act() — POST only, local() Host and Origin, X-Board-Token equal to this run's token, and a Live session attached (409 otherwise).
- Names maps every word the pages show through the lexicon (DefaultNames for isekai); a new page word goes in Names, never in a template literally.
- sources.go reads the instruments: Reading is lit or silent (a silent instrument is drawn as a finding, per Nature 9), Body for a live Court, UsageRecord/UsageSum/Rollup with Ranges 24h 7d 30d all, Run from loop journals, LogEntry from log.md, Off from config's honesty findings, Node/Edge/Colony from the ontology, MemoryView and ToolboxView.
- Live is the interface the app's session implements (Ask, Answer, Interrupt, Status); the board never holds a session of its own. `isekai board` serves read-only without one.

## Verify
- `go test ./board/...`

## Thoughts
