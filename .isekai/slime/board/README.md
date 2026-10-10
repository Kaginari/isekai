# slime-board

- **Rank:** Slime
- **Territory:** `board/`
- **Reports to:** orc-interfaces
- **Minds:** slime-go-proving, slime-wire-and-journals
- **Purpose:** the ground truth of the board and the dash — the world seen over HTTP on loopback from the same instruments the session writes, the live session driven from the browser through token-guarded acts.

## Traits
- `board.New(Options)` is an http.Handler with everything embedded (templates/, static/, assets/ with pinned Bootstrap, web/): a page works offline. The app binds it to 127.0.0.1 on ui.board.port (default 7411); `isekai board` serves it read-only without a session. `board/README.md` documents the pages, feeds and the usage schema (a test holds the schema doc true).
- Routes: the pages / court usage colony memory toolbox log config (one html/template each over layout.html); /events (SSE; watch polls file signatures — name, size, mtime — and the court's JSON, stdlib only); /api/doc court usage colony tree tree/children courts plan; liveRoutes add /web/, /dash, /dash/stream, /api/world, /api/live, and the acts /api/ask, /api/answer, /api/interrupt. /api/courts sends counts only, never body fields; /api/tree/children serves one lazy level and answers 404 for a path escaping the root; /api/plan answers `{body, empty: true}` with 200 when no deck exists.
- Acts pass `act()` or nothing happens: POST only (405), local() — Host is localhost, 127.0.0.1 or ::1 and Origin, when present, is this host (DNS rebinding) — and X-Board-Token equal to this board's random per-run token (403), and a Live attached (409 otherwise). /dash/stream also refuses a non-local Host.
- The answer path: Live (board/live.go) is the interface the app's webBus implements — Subscribe (replay, then the stream), Ask, Answer(id, index, text), Interrupt, Status. The board never holds a session itself; /dash/stream sends `reset`, the replay, then events with a 15 s ping, or `nosession` when Live is nil.
- /dash serves web/dash.html with {{TOKEN}}, {{WORDS}}, {{DIST}} substituted, Cache-Control no-store and a strict CSP (script-src 'self', no inline script, frame-ancestors 'none'). dash.js has no dependencies, builds DOM with textContent or escaped markup, sends the token on every act, and draws the live net (bodies and minds on one grid, rank portraits from web/portraits) as SVG.
- The dash is built on the ui system: web/tokens.css, palettes.css and layout.css are byte copies of the ui foundation, and web/components/<name>/ must lint clean under ui.Lint — TestDashboardHoldsTheUISystem fails on drift ("copy it again"). Colours and lengths come from tokens only.
- Names maps every word a page shows through the lexicon (DefaultNames for isekai, merged with the distribution's); a new page word goes into Names, never literally into a template or dash.js.
- `board/sources.go` reads the instruments: every figure carries a Reading (lit or silent, why); a silent or missing instrument renders as "silent since …" or "never lit", never as zero. Usage rollups treat a null usd as unpriced.

## Verify
- `go test -count=1 -run 'TestDashActsNeedTokenLoopbackAndOrigin|TestDashPageCarriesTokenAndCSP|TestDashStreamReplaysThenStreams|TestDashboardHoldsTheUISystem|TestAPICourtsNeverLeaksBodies|TestEmptyWorldIsSilentNotZero|TestLexiconSwapChangesLabels' ./board/`
- `go test -count=1 ./board/...`

## Thoughts
- (none yet)
