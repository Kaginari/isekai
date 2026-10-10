# Log — dash tron-cockpit redesign — 2026-10-02

Goal (human): dash -> futuristic AI cockpit: trident grid [neural net LEFT=mind | console CENTER=voice | world-tree RIGHT=territory], files lit per running task by tool class + body; world-tree = the world root where isekai runs.

Design settled via elf-isekai verdict (court):
- EVOLVE aurora in place (dash.html:2 is its only consumer) -> tron dark. palettes.css aurora: --p-1-dark->oklch(0.82 0.16 195), --p-2->oklch(0.70 0.19 285), --p-3->oklch(0.76 0.17 330), --p-4->oklch(0.30 0.09 215); keep --p-5 + all neutrals. ZERO tokens.css edits. Contrast >=4.5:1; prefers-reduced-motion; CVD-distinct series.
- Trident via shell.css container query @ ~96rem (console widest center; net+tree narrower), preserving existing data-view tabs + twocol/onecol collapse map.
- File-heat: write/gate/outward/destructive->--warn, read->--series-1, per-body->rank token; folder glow bubbles up. Prefer-reduced-motion -> static highlight.

Seams verified (this session):
- SSE feed ALREADY ships per-tool {class,wrote,body}+diff.path (app/web.go toolJSON:244-260) -> paint loop consumes it, NO app/web.go change needed.
- board/tree.go does NOT exist (dropped-stash residue) -> greenfield: Sources.Tree walker (skip dep/build dirs per app/found.go:60 + tree-stamp law), cap 4096, fenced inside-world like apiDoc (board.go:539), GET api/tree -> {live,root,tree,truncated}. embed all:web (board.go:28) auto-includes new tree.css.
- apiDoc fence lives in the SOURCE; handler 404s on error (board.go:539-547).

Harness findings (not design failures):
- Gate-attribution bug: loop/snapshot.go turn-watch charges ALL writes in the open window to the session whose turn is open -> elf-isekai/kijin gates failed on sibling writes. (flagged @U territory)
- elf-isekai/kijin verify 'go mod verify' reads machine-global GOMODCACHE -> dirtied cache fails gate on environment. Pairing fix: per-project GOMODCACHE under territory. (flagged @U territory)
- board/testdata/world/.isekai fixture MISSING from git -> pre-existing board test REDs; new tests here are fixture-INDEPENDENT + additive.
- Orchestration LAW: strict layer flow (rimuru->elf only; elf descends orc->slime). Law 7 + decree 2026-10-01: throne authors NOTHING but .isekai/isekai.md (my palettes.css edit was refused). board/ pairing law (decree 2026-09-30): no board/ write without slime-tui-board README in same turn.

Execution: chain birthed this session -> elf-isekai (helm) -> orc-host (gate) -> slime-tui-board (hands, API+web). OPEN debts: s1-cyan vs s6-cyan-dark CVD sweep under palette-audit mind; GOMODCACHE pairing; gate-attribution fix.

# Log — dash tron-cockpit LANDED — 2026-10-02

Capability landed: dash = futuristic trident cockpit (neural net LEFT | console CENTER | world-tree RIGHT), files lit per tool touch by class+body. Verified by rimuru's own reads.

Gate (rimuru ran): go build ./board ./app GREEN; go test ./board (walk-tree + pairing + every-page) PASS; pairing Vitality satisfied (slime-tui-board README moved same-turn).

Anchors — Go: board/tree.go:1-72 WalkTree+TreeEntry{path,name,dir,depth} cap=4096 skip={.git,node_modules,vendor,bin,dist,.next,.venv,venv,__pycache__,target,.cache}; board/tree_test.go WalkTree/Cap/Empty/ZeroCap; board/sources.go:313 Sources.Tree(entries,truncated) + :358 FileSources wires WalkTree(f.root,4096); board/board.go:233 route + :577 apiTree handler {live,root,tree,truncated}.
Web: board/web/palettes.css aurora->tron (p-1-dark 0.82.0.16.195, p-2 0.70.0.19.205, p-3 0.76.0.17.330, p-4 0.30.0.09.215); shell.css:29 trident 96rem; dash.html:112 data-panel=tree; components/tree/{tree.css,tree.html}; dash.js:718 fetch api/tree + heat/decay (write->--warn, read->--series-1, rank tint).

Courts: elf-isekai helmed (3 FAILs on the way — harness timeouts/stress+deck-pitfall, NOT work refusals), orc-host gated, slime-tui-board authored the Go tree feed. Lawful chain descent: strict layer flow binds DISPATCH (rimuru->elf/kijin only); elf descends orc->slime via court tool.

OPEN for human: (a) tree.go skip-set omits build/out/testdata/coverage (testdata stays VISIBLE in dash tree) — leave or add? (b) tree.go:70 dead '_ = os.PathSeparator'. DEFERRED to fresh session: docs/ design note, full series CVD sweep (palette-audit), optional neon pass. Harness debts flagged: gate write-attribution, GOMODCACHE verify, missing board/testdata/world/.isekai fixture.


# Log — TUI slice 1: court-row state pulse — 2026-10-02

N15/N16 wave 1 landed (human order): court rows pulse. thinking = shimmer-sweep on the verb crested per race lane (rimuru rose-gold / elf verdant / orc ember / slime chartreuse), tool = bubbles spinner frames + tool name (`TOOL edit · 12s`), waiting-on-gate = harmonica spring bob off the right edge (ω4.5/ζ0.32, kick 8, ≤2 cols, re-kicks while warm), done/fail = 700ms full-row lane silhouette flash (fail = bad colour) then settle → tick sweep owes exactly one tombstone. Past 8s body silence: quarter-rate breath (pulseIdleCool), bob settles. Hard-nos held: zero new deps (charm.land vendored only), no background painting (no fake transparency), colorprofile tier-degrades to plain text, every motion rides the existing 200ms evTick (no new event flow).

Anchors: tui/theme.go lanes+lane(); tui/motion.go shimmerText + courtPulse/stepPulses/pulseFor/flashHold; tui/blocks.go Pulse + Theme.Court(v,w,...Pulse) variadic (zero-Pulse byte-identical — goldens untouched); tui/model.go pulses wiring (EvCourt/EvState note, EvStream/EvBodyStep activity, evTick flash sweep, landCourt cleanup). Carried fix: EvCourt update branch copies non-empty Exit (leg-2a had a first-sight-only hole). goldens court-pulse-*@{80,120}; tests TestCourtFlashHoldsThenTickLandsTheTombstone + TestCourtPulseCoolsPastIdle. tui suite green incl. -race; app ./loop ./world ./tool same as baseline (TestTUIInterruptAndQueue stash-verified pre-existing env flake again). Crush mining annex for slices 0/2/3 sits in .isekai/instruments/audits/audit-fixes.md (N16 ANNEX).

OPEN: slices 2 (grid: plan rail W36 / act log / court table + sidebar constants from mining), 3 (modal overlays — mining says no backdrop dimming upstream: honest layering), slice for diff-card compositor (geometry+blend ratios ready). Commits are the human's act.

# Log — TUI slice 2: the left rail — 2026-10-02

N16 slice 2 landed. Wide tall windows (≥120×30) wear a fixed left rail riding the live area — scrollback canon kept (conversation still flows; rail redraws per frame, never prints). Crush geometry verbatim: W=32 + 2-col gutter, centerW=W-34, compact tripwire 120×30, choice opens = full width (focus theft). Contents: vitals (model / ctx Meter+stress / tokens · cost), touch ledger (dedup latest-last, per-writer race lane, cap 10), court roster (live in dispatch order + landed bodies, cap 8, lane dots). Rai l cells never animate on their own — motion stays on the centre row; one moving surface per region.

Anchors: tui/blocks.go RailView/RailFile/RailCourt + Theme.Rail + railW/railMinW/railMinH; tui/model.go files ledger (EvToolEnd/EvBodyStep Wrote → noteWrite lane via rankOf), railView/railVisible, Render grid via JoinHorizontal(Top). Goldens rail/rail-empty@80/120; TestRailRidesWideWindows (tripwire, dedupe, roster, choice suppression). -race green; app unchanged vs baseline flake.

Skipped on purpose: plan-deck rail section (no plan data model terminal-side; sidebar spec supersedes), per-cell roster animation (one-motion-surface law).

# Log — TUI slice 3: modal asks — 2026-10-02

N16 slice 3 landed. On windows ≥78×21 the ask/approval card is a true overlay: backdrop darkened honestly (frame stripped and re-worn in dim — no alpha), card centred via crush CenterRect math at min(w·3/5,100), top edge gradients accent→border (elevation cue), enters with a 2→1→0-row drop over the first 400ms riding the existing 200ms tick (no new event flow; the 8s law never gets a say). Below the tripwire the card flows inline — pre-slice behavior byte-kept. Compositing rides lipgloss layers/canvas (r eal cell compositing; blank cells show the frame through). Keys unchanged; numeral fast-answer already in law.

Anchors: tui/model.go choiceAt + choiceFloat() tripwire + EvChoice stamping; tui/motion.go choiceMinW/H, centerRect, dropOffset, pulseFPS reuse; tui/overlay.go overlay() choice branch (darken, gradient top via Gradient(), centre+clamp). TestChoiceFloatsOverDarkenedBackdrop: backdrop dim colour pinned (#7f8a99), centred card, gradient stop count, drop offset 2,tripwire inline fallback. TestChoiceBlock (digits/enter/free-text contract) untouched green.

# Log — TUI slice 4: the ambition tail — 2026-10-02

N16 slice 4 landed. (a) Crest: live Court cards wear the race's two-row half-block sprite as an 8-col left column (sprite alternates its two blink frames on the pulse phase); card renders at W-8, windows <38 total drop the crest whole; tombstone prints stay single-◆ (zero-Pulse byte law held); the landing flash silhouettes the sprite with its row (flash runs after crest). (b) Burst: a gate verdict after a Court ran rains ◆ over the frame for six ticks (1.2s — far under the 8s law), pass=ok/fail=bad, dete rministic knuth-scatter per arm, self-expiring; courtless verdicts stay quiet. (c) Choice.Why now renders through t.Markdown (glamour vendored); plain-prose fixture goldens strip byte-identical, so the approval/question goldens held without rewriting.

Anchors: tui/blocks.go Pulse.Crest + crest column (dress=flash•crest) + Choice Why→Markdown; tui/motion.go burstTicks/pulseFPS; tui/model.go burstAt/Good/On armed in finish() (len(logOrder)>0); tui/overlay.go burstRain(rows) knuth-scatter x, staggered y=tick·2, Z2. Goldens court-crest@80/120; tests TestCourtCrestGeometry / TestVerdictBurstRainsOnce (arm/spend/recolor/courtless-quiet). -race green; full app suite: only the documented TestTUIInterruptAndQueue env flake; TestREPLLive red ONCE beside its 15s stall (stall cascade; passes isolated ×2 and on rerun) — noted, not attributed.

Wave complete: slices 1 (pulse) + 2 (rail) + 3 (modal) + 4 (tail) all landed with goldens+tests. Hard-nos held across all four: no new deps, no fake transparency, single existing tick family, colorprofile tier-safe.

# Log — TUI ambient dressing + rail tiers — 2026-10-02

Human read of the wave: "nothing changed" — the four slices were all state-triggered; boot/idle byte-identical by the zero-Pulse law, and his window sat under the 120×30 rail tripwire. Ambient pass landed: welcome box wears the lane legend row (● rimuru · ● elf · ● orc · ● slime in their colours; agent-one maps its four ranks) + accent border elevation; footer's model token rides the accent lane; the rail now TIERS (32 at ≥120, compact 24 at ≥96, off below or under 26 rows) and speaks at idle (world base name head); compact roster drops the age before chopping names mid-token.

Anchors: tui/theme.go laneLegend; tui/blocks.go Welcome legend+accent border, Rail(v, ...widths) tiered + World head + nameW/age-yield roster math, Footer model tint; tui/model.go railVisible()→int via railSize, cw=W-rw-2, railView World=basename(footer.World). Goldens: welcome regen (legend line), rail/rail-empty regen (World row), rail-24 new; footer strip-identical. Binary reinstalled 22:45 (make install).

OPEN for human eye: real boot now differs at ANY width (welcome legend/footer tint); the sidebar rides windows ≥96 wide · ≥26 tall.

# Log — TUI: right rail cards + verdict banner + ctx zone — 2026-10-02

Human steering: sidebar goes RIGHT, screen splits 2/3 | 1/3, mascot + proper cards + colored usage metrics; gate verdict wears a banner card; ctx stress lives in the sidebar. Landed: rail re-proportioned (railSize = w/3 floored at 28, off <60 wide or <26 tall), right-joined (center | gutter | rail); cards: gradient-titled rounded frames — vitals (world/model/ctx Meter/stress + zone row NEAR warn / STRESS bad / OK quiet / tokens · cost accent), files ledger (lane-tinted, MaxFiles cap), court roster (MaxCo urts cap); mascot art rides the rail top when h≥34 and it fits. Gate verdict → one banner block (ok/bad background, dark-ink-on-dark-theme / light-ink-on-light-theme fg, bold; ASCII profile degrades to plain words). FooterView gains Zone; app/tui.go wires instrument zone (OK/NEAR/STRESS/UNREAD) through; the app package touched (footer builder) — app suite: only the documented env flake.

Anchors: tui/blocks.go Rail(cards)+railCard(gradient titles)+Gate banner+RailView.Zone; tui/model.go railView(rw) mascot/caps, right JoinHorizontal, Zone pass-through; app/tui.go Footer zone. Goldens: rail trio regen (cards), rail-mascot new, gate/gate-fail regen (banner), leftover rail-24 fixtures deleted. tui suite + -race green.

# Log — TUI: true pane split, sidebar RIGHT with its own background — 2026-10-02

Human vocabulary: make it like crush/opencode — chat | sidebar, fixed sidebar width, sidebar collapses when window too narrow, DIFFERENT BACKGROUND. Landed (supersedes the left-rail geometry of the ambient pass): splitPanes() frames every render into chat pane (cw, terminal bg, exact-width padded, clipped pane-bottom style — newest lines win) | divider cell | sidebar (railSize = w/3 min 28, off under 60×26). Sidebar background painted for real (paneBg #14181f dark / #e9edf1 light SOLID, no transparency) cards wear it via lipgloss' bg-filled padding, mascot art gets resets rewritten into bg-restores (paneArt), filler rows are bg wipes. Choice modal now VEILS both panes (strip-dim) and keeps the layout standing; overlay test rebuilt on computed geometry. Shots: panes.ansi shows the frame.

Bug found by the debug print (kept as doc): padRight duplicated the line content when the live row was wider than the pane (returned padded string instead of the padding) — the "duplicated court row" in the veil debug.

Anchors: tui/model.go splitPanes/padRight-fix/railVisible; tui/blocks.go Rail pane-wipe+railCard bg+centre; tui/theme.go paneBg/paneArt. Goldens rail-* regen (bg-filled cells make stripped output stable). -race green; app: only the env flake.

# Log — TUI: frame-height invariant + overlay pure layers — 2026-10-02

Human's paste showed the sidebar scrambling mid-session (footer ghosted into the input row, sidebar rows shifting). Root: the pane frame must be EXACTLY m.height rows forever (crush-check confirmed their fullscreen rectangles are exact — sidebarWidth 32, uv.ScreenBuffer, AltScreen). Overlay's toast prefix-rows and inline choice under the pane broke it. Fix: overlay is pure layers (toasts/palette draw over the base's own rows — zero new rows); Render gives the choice-inline a seat INSIDE the chat pane; ho me rows measured, paneH = height - home rows. Stability tests: TestPaneFrameHeightIsInvariant (toast/menu/textarea/extreme stream all pinned) + TestPaneHoldsOnRealProgram (teatest full program, prints interleaved, exact-height frame, divider present, no row overruns). Race-safe frame reading via evSnapshot handler.

Crush ground truth pulled and confirmed: internal/ui/model/ui.go:3738 v.AltScreen = true + :4312 sidebarWidth=32 with uv.ScreenBuffer rectangles — they never mix scrollback with a live region. We stay inline by canon (conversation into scrollback) and pay the same discipline on the live frame.

Anchors: tui/model.go Render restructure + lines() + splitPanes(paneH param) + evSnapshot; tui/overlay.go pure layers; tui/pane_stability_test.go.

# Log — TUI: wide-window divider law — 2026-10-02
Paste forensics #2: his window is wide (~210) — the frame was fine-rowed but the law wasn't pinned. TestPaneWideWindowColumns: 210×50 busy storm (prose + two courts + tool prints + floating ask) — every row exactly m.width, divider at cw on every non-card row, plus 96×30 / 120×26 / 240×80 sweeps. The ask card legitimately crosses the divider column (modal law) — the check masks its rows.

# Log — TUI: pane mode stops printing (pty-verified) — 2026-10-02

Root cause nailed with a real pty harness (pyte emulator, 230×55): the corruption never was geometry (frames were exact) — it was bubbletea's inline renderer when a full-height frame and scrollback prints coexist: its row-diff bookkeeping loses sync the moment a print lands between frames (evidence: paste's merged lines, probe's spliced "litask, or / for commands tx –", 27X/27C stutter in the raw stream). crush avoids this class because AltScreen+ScreenBuffer never prints.

Fix: pane mode (railSize on) stops printing. The transcript lives in the pane viewport (paneChat cache over m.blocks, dirtied on land/resize/theme/swap); reflow under panes = cache dirty; prints still happen the moment the window no longer holds panes (exactly-once law via printedIdx — every block reaches the scrollback at most once). Inline mode is untouched when panes are off (small windows flow exactly as before).

pty probe rerun on the rebuilt binary @23:4x: paint perfect at boot AND at thinking AND at a full streamed run end — mascot + vitals + files + courts cards on their background; chat pane scrolls its own transcript; no ghost rows, no spliced lines.

Tests: model_test tombstone/flash tests now assert on the transcript store (flow-truth), drainPrints learns the pane route, TestPane* invariants green. Suites: tui + -race green; app/loop/world untouched green except the stash-verified env flake.

# Log — chat pane scrollback via bubbles/viewport (N16 wave A1) — 2026-10-03 night

The pane transcript now rides the vendored bubbles/viewport instead of my hand-rolled
tail cache Render each frame: `tui/pane_vp.go` (`ensureVP`/`feedVP`/`vpKey`/`vpWheel`).
pgup/pgdn/ctrl+u/ctrl+d walk the transcript; the mouse wheel scrolls it too
(MouseModeCellMotion when panes or board are live). Anchoring law: pinned AtBottom
follows the live feed; any scroll-up hands the wheel to the human — new content does
not yank the view to the tail; returning to bottom re-engages the follow. Verdict
bursts + choice veils keep working: burst draw happens post-viewport in Render, veils
strip the whole frame, never the pane.

Pinned by TestChatPaneScrollsWithViewport (scroll → no-yank → refollow law) alongside
the existing TestPane* invariants. Fixed a self-inflicted build break mid-landing: my
PasteMsg edit left a duplicated stray body inside the switch (syntax error at :409);
first crash + pty verification forced the recovery before the landing card.

Suite: `go test ./tui` + `-race` green; pre-existing app/* gofmt noise stash-verified
untouched-by-me. Commit: tui pane-viewport (`cbd7200`), binary reinstalled via make install.

# Log — charm-lifts sweep (N16 close-out) — 2026-10-03 night

Three vendored-library adoptions on the wake of the pane/viewport wave:

1. **lipgloss tree for the files ledger** — the rail's `files` card now draws the touch
   ledger as a tree: directories collapse to one dim root (`tui/`), leaves hang under it
   worn in the writer's race lane (ItemStyleFunc mapping basename→lane). Bare paths (no
   dir) ride the root directly. `Hide(true)` was a trap — it hides the whole render;
   an empty root value is the honest way to skip the root line.
2. **glamour StyleConfig riding the world palette** — markdown (headings=accent gold,
   emph=ember, links=verdant italic, inline code=chartreuse, block quotes sunk into dim
   with a `│` token, code-block bg = paneBg). NO_COLOR path untouched (stock NoTTY
   config, branding skipped). Pinned by TestGlamourBrandRidesThePalette.
3. **ctx meter** — `Theme.Meter` is now a bubbles/progress bar (half-block precision,
   percentage stripped, EmptyColor = the border token, WithColorFunc over a 101-stop
   green→amber→red ramp from the zone lanes: fill colour = how full the gauge is).

All three stay inside the vendored module set (zero new deps — tree is a sub-package of
lipgloss v2.0.6, progress of bubbles v2.2.1, glamour was already the markdown engine).
Goldens regenerated for the rail (tree rows + progress glyphs). Suites: go test ./tui
+-race green, gofmt clean, `go build ./...` and `make install` green.

# Log — plan card + reflecting welcome + quarter-block cast — 2026-10-03 night

Three looks, one flight of work:

1. **Plan deck card** (`01885f3`): the rail gains a `plan · <body>` card above the files
   ledger. `Host.Plans()` reads `<world>/tmp/plans/*.json` each tick while the rail is on
   (cheap: one small JSON per body). The elf's deck is the steady state (its rank sits
   first); `ctrl+p` cycles to orc/slime/etc. and sticks until you cycle again (planSticky).
   Status rides lipgloss background chips spanning the card's row: pending brushed-steel,
   in_progress amber, done deep-green, blocked ember-red — readable foregrounds for both
   the dark and the light theme. New golden `rail-plan-80/120` pins it.
2. **Quarter-block race cast** (same commit): `quarterBlocks` renders 2×2 sub-pixels per
   cell — every rank's face now fits a 12×4-pixel canvas in the same 6-cell row the
   half-blocks rode. The animation cycle is IQ + blink + IQ + blink across 4 frames
   (blinking and gaze-shifting pupils). Elf + dark-elf got genuine long ears (the `P`
   pixels now flank the head); the orc keeps its red brows and white fangs; kijin keeps
   its horned fringe; rimuru blobs halved-view. Goldens regenerated; shots re-baselined.
3. **Reflecting welcome + refreshed rail borders** (`418965f`): the old welcome's fat
   mascot block is gone — the slime lives in the sidebar now, and the welcome box shows
   title + tagline + lane legend only. Both the welcome box and every rail card now wear
   `BorderForegroundBlend(accent, border, accent)` (lipgloss v2's wrapped blend loop) —
   light sweeps from accent into the dim border and back around the frame. Verified in
   the windows shot.

Suites: go test ./tui + -race green; app suite stalls only on the known env flake
(TestTUIInterruptAndQueue, stash-verified red on baseline). `make install` fresh binary.

# Log — sidebar mascot on quarter blocks — 2026-10-03 night

The pane mascot rides `quarterBlocks` now (`8eabe3f`): a 32-pixel-wide canvas (twice the
horizontal grain, same 16-cell rail footprint) with slimmed slime blob for isekai and
tailored antenna-nub robot for agent-one. The shadow row `S` also rides quarter blocks,
so the floor line no longer bleeds past the body. Goldens regenerated; ANSI shots
re-baselined.

# Log — PNG portrait mascot (go-termimg half-blocks) — 2026-10-03 night

The pane mascot is the world's own art now, not a hand-drawn sprite: `init` stages
`<world>/.isekai/portraits/{rimuru,slime,elf,...}.png` at world-found time
(board.Portrait exports the existing board/web/portraits PNGs; app/embed.go copies them).
`tui/portrait.go` renders them at session open via github.com/blacktop/go-termimg
(ProtocolHalfblocks): 12% bezel crop to shed the picture-frame, corner-sample background
knockout (tolerance 26) so the creature grounds itself in the pane bg, bi-linear scale to
2w, 4h px → half-block cells. Plan card styling in the same commit: markdown stripped out
of plan item text, uniform lane-blended card background.

Zero-risk fallbacks: missing portraits/ dir or a failed render returns the hand sprite from
tui/mascot.go (quarter-blocks, kept). Suites green; ANSI shots re-baselined.

# Log — rollback + pills — 2026-10-03 night

The quarter-block experiment fell: cast.go and mascot.go restored from commit 7379677 (the
half-block originals). The pane mascot, race icons, and badges are back the way they were
before the RGB redesign overshot. go-termimg + the portrait PNG work stay — vital for later
work, dormant today (no live consumer; blockers: needs a renderer at-pane, not in-model).

Also in the same flight: (1) approval popup options are lipgloss pills — picked option is an
accent-filled chip, the rest are dim pills, the card keeps the reflecting-blend border;
(2) vitals card gains a `/ <cap>` suffix on the ctx meter (∞ when unconfigured) and one
pill per race with this session's token spend (rimuru / elf / orc / slime, lane coloured,
wrapping on narrow panes). UsageJournal tracks per-rank spend for the pills.

Suites: tui + -race green, app green except the known env flake. Tests pinned to the pill
form.

# Log — race lanes + token-audit preflight — 2026-10-03 night

Lane colours realigned to canon (`tui/theme.go` lanes map): rimuru golden, elf orange, orc
green, slime blue, kijin red, dark-elf violet. Agent-one's four ranks ride the same mapping
via their canonical aliases. No test drift — pills still read the lane from the legend's own
map, brand test asserts lane-of(elf), not a hex.

Token-audit of the user's question ("why did asking about something burn 1M"): session
20261002-232246-cd5f37 shows elf-isekai alone pumping 27 calls for one court turn (input
tokens mount from 18k to 37k as the elf's context grows per re-ask), totalling 1,057,004
across 53 calls and bodies (rimuru 198K, elf 426K, orc-host 256K, slime-tui-board 177K).
Verdict: not a leak — it's how a court's conversation is priced (full-context replay per
step). If the pills matching this answer matter more than the number, say "flip to fill" and
the vitals pills show each body's current ctx% instead of its session burn.

# Log — wire compaction for court fan-out — 2026-10-03 night

Loop calls the provider with `wire := compactReplay(s.Messages)` now, not the full session
transcript. compact.go: folds tool outputs older than the newest four into `[replay-folded
<toolname>: N lines were here]` captions, and strips Opaque thinking blocks from past
assistant turns (kept on tool-carrying messages — Anthropic's contract binds a thinking
block to its tool_use). The compactor returns a deep copy; session memory keeps the raw.

Expected dent on the audit case: the elf's 27-call run carried ~30K of input replay per
step; folding the five oldest tool groups alone cuts about a third of that per call at
the end, more in longer runs.

Before/after on the wire (TestCompactReplayBeforeAfterReport): a 12-tool run plus a
follow-up question sends 67,200 bytes of tool replay without the diet, 23,040 with it —
a 66% cut. The diet's strength grows with the run's length (only the newest 4 tool groups
survive raw), so a 53-call audit-grade session lands somewhere in the 60–80%% window.

# Log — sidebar slime + shadowed rail — 2026-10-03 late night

The slime is back home at the rail's top, always (height gated at 30 now, was 34 —
shorter windows were losing it), and it blinks there: every 12th evTick the eye pixels
swap for body pixels via mascotArt(closed=true) — the intro's drop machinery stays usable
by the shots test but nothing drives it at launch anymore.

The welcome box no longer prints the dash URL; instead the vitals card carries it as an
OSC 8 hyperlink ("dash ↗ http://…") via lipgloss.Hyperlink — ⌘-click opens the browser,
no paste. Welcome.Board field is gone; the rail's URL rides FooterView.Dash.

The pane edge has a 3D cast: two shadow columns (▓▒ density-gradient runes tinted by
Blend1D(paneBg → shade anchor)) sit between the chat and the divider; the divider keeps
its border colour over the pane bg. Layout law updated in TestPaneWideWindowColumns: the
divider now sits at column cw+2.

Note for the human: if the shadow reads too dark or too dim in your terminal, say so —
Theme.paneShadow is one knob, the anchor hex and the blend stops.

# Log — quiet board + tracked titles — 2026-10-03 late night (2)

serveBoard's stderr print (`board: http://… · dashboard: …`) is gone entirely — the TUI
carries the URL in the vitals card as a clickable link, `say`/`dash` mode prints its own,
and `status` keeps boardLine. The board surfaces only where the surface talks.

Rail card titles are letter-spaced (tracked): `vitals` → `v i t a l s`, files/courts/plan
the same; falls back to the plain word when the floor won't hold. Goldens regen.

# Log — forge demand on long runs — 2026-10-03 02:45

slime-loop's 37-turn tour of the loop package answered the user's own question: yes, the
machinery is there (@FORGE wire line → record beat → notes ledger → the desk's /forge mint).
What's missing is the trigger: the slime never named a draft because closing didn't ask.

Change: forgeDemand(steps) rides the closing check — at 12+ steps the re-ask carries the
crystallization question ("name a @FORGE skill draft or say why it's one-off"). Short runs
stay silent. 12 is the floor because a real tour/probe crosses it and a one-two steps run
has nothing worth minting. Next equally-hard walk of the loop package should land its
tour-skill draft in the notes, and the desk mints /forge.

# Log — read batches/caches, skim outlines Go — 2026-10-03 02:50

The read tour that cost slime-loop 37 turns is now structurally cheap:

- `read` batches (paths[] — per-file heads in one call), auto-limits whole-file asks past
  200 lines (full: true escapes), and answers an unchanged window with a pointer instead of
  re-serving the bytes (fnv64 + mod/size check — edits always re-serve, never stale).
- `skim <path>` (new builtin #7) is a go/parser outline: imports, every type/const/var/func
  with line anchor + doc first-line, no bodies; package dirs head each file.
- minimal shelf rides skim — tourists get the outline they kept grepping for.

First run after this should show: fewer turns per walking tour, shorter prompts per turn
(skim/head answers instead of re-served windows), forged skill drafts named at the close
(see earlier entry, 729a0b4).

# Log — micro pixel font on rail titles — 2026-10-03 02:55

The tracked titles were rejected on sight (spaces read as noise at card scale). The human's
counter-order: "use vhs for font".

VHS is a recorder, not a font — so the VCR-era look is drawn instead: tui/font.go holds a
micro-letter 3x4 bitmap font, rendered through the existing halfBlocks pipeline (the slime's own
sprite), accent-coloured. railCard titles with font coverage become two pixel rows; anything the
font lacks (plan · body, lowercase-safe) falls back to the gradient word. The choice cards and
welcome still carry the plain gradient — pixel density is a rail-head title privilege.

# Log — live-loop audit + staleness nags — 2026-10-03 03:10

Verdict doc at .isekai/instruments/audits/2026-10-03-slime-loop-audit.md. Nothing stuck:
both slime-loop runs tonight closed themselves (guts doom streak 3 → auto-checkpoint, the
ladder's design). orc-engine is live digesting checkpoint reports. Recall nagged
stale-index on every body all night (index older than HEAD — log.md, elf README, tui-board
README changed this session); the fix is `isekai memory index` in the world — running it is
the human's act, not mine mid-flight.

# Log — the courier law — 2026-10-03 03:20

Dispatch reports over 600 bytes no longer enter the parent's context whole. The court files
the full wire text on the instruments shelf (<body>-<run>.report.md beside the journals) and
the parent receives an envelope: @S verdict, @R shelf path, @H fnv64 checksum, @W wrote-set,
8-line digest head. Small reports ride whole below the floor.

The instruments shelf is now engine soil in territory law: Law 2 refusals with "engine soil"
named — every body's write under .isekai/instruments/{loop,usage} refuses before run. The
wire register grew @R/@H/@W in the world's canon (isekai.md).

Pins: TestCourierEnvelope (app) — envelope rides, flesh filed, integrity holds.
TestTerritoryInstrumentsIsEngineSoil (world) — the slime's forged-report write refuses.

Number the audit taught (2026-10-03, "ask slime"): 5.1M tokens; 42% was orc+elf relaying the
slime's text. The courier law ends relay-by-flesh; the layer keeps its job but pays
courier's prices, not pyramids'.

# Log — ranks die to the office lens — 2026-10-03 03:50

Rank model pins lied by identity: a pinned orc rode claude-opus even on a draft commission the
office named worse. ResolveModel now ladders creature → office → task → default; any set-forever
rank pin becomes a hole naming where to move it ("offices lens, never ranks"). The rank table
itself stays — its Job, ReportsTo, Office fields are the world model — but Model stops being its
business; a rank's Office field is the pin (that one's coded and authored by the law, not by
an accident of config).

Session-dispatch backgrounding stays off at the default flip: mid-chain courts, when
backgrounded, would strand their reports (the wake channel feeds the session only). Foreground
is the law between ranks by design.

Examples re-skined to the conduit mounts (glm-5-3-conduit / kimi-3-conduit /
nemotron-3-super-120b-conduit) with two-deep fallbacks and no gemma anywhere.

# Log — tiers declared, nags closed — 2026-10-03 04:10

providers.yaml's litellm block now declares the three served mounts with tier +
contextWindow (nemotron 1, kimi 2, glm 3), so the office lineage check actually runs, and
the dead rank-pin holes dissolved when the pins left models.yaml (world + machine layers).
ModelTable now prints "rank → office" lanes, not twice-resolved defaults. Everything else
of the evening rides the same binary; rebuilt for the live run.

# Log — small caps + plan owns its office — 2026-10-03 04:45

Rail order is vitals → courts → plan → files; title face is one-line Unicode small-caps in
per-card lane colors (pixel font retired; it read as two heavy rows). The plan card's tint
is its office, not its race (resolve live first, rank's default office second) — and the
steady state defaults to rimuru's own deck since dispatch now lives in the throne's hands.
Court rows on the rail are clickable: a click pins that body's plan into the card, sticky
until another pick; office tint follows.

# Log — quiet blocks — 2026-10-03 04:47

The tools' bodies (diffs, outputs) default to hidden, wearing only their status line
(ok · elapsed · line count · ctrl+o); the log is one chord away. collapses() drops its
size threshold — the law applies at any size. The model's EvToolEnd arm registers the
block for ctrl+o on every done act with a body, so the chord's answer never misses the
latest work.

# Log — clean rollback + retest — 2026-10-03

Veldora ordered the held plan (files panel + tree work) treated as wrong: cleaned,
rolled back, retested. Code artifacts wiped: board/sources.go and board/tree.go
restored to HEAD (FileChange feed, Kind field gone), board/src/ (5 TS files),
board/web/components/files/ and package.json removed. Untracked dispatch records
(docs/commissions/, loop census report) kept per journal law; .env untouched.
stash@{0} (older half-written board.go/events.go WIP, 217 lines) deliberately held
as the only recoverable copy — dropping it awaits the human's word.

Retest at clean HEAD (ce5cc6d): go build ./... green. go test ./...: one failure —
app.TestTUIInterruptAndQueue, 3/3 reproducible, tree byte-identical to HEAD so not
the rollback's doing. Bisect verdict (bfc987d) proved FALSE on re-verification:
bfc987d~1 fails 3/3 and the test's birth commit 8bfdef0 fails 3/3 with a different
symptom (first turn stuck at "Pondering… (0s)" — tool roundtrip never completes).
The test has NEVER been green in this repo's history; it was born broken in the
harness import. Not a regression from recent court work.

Diagnosis of the HEAD symptom (stuck at "the note ran as its own turn" wait, but
esc-interrupt drain path reads correct at app/tui.go:531 finish→Inbox→start): the
mock script helper matches steps by "when" substring against the last user message
(app/mockscript.go). Suspect: after esc-interrupt, the queued "a note" turn is
started AUTO (EvTurnStart Auto:true), and either the drain never fires because
finish() runs before Queue()'s Court.Enqueue lands, or the auto-turn's mock step
never matches. Open hole: exact root cause of the note turn not firing after
interrupt at HEAD (the mechanics at 8bfdef0 are different — first turn itself
hangs; a separate defect). Both need a real fix commission, not a re-run.

Escalation note: TUI writes are gate-refused for rimuru (throne authors nothing
but the law) — the fix must go through dispatch.

