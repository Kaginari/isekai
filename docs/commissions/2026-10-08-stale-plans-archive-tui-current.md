# Commission — stale-plans archive + TUI "current" switch (human order 2026-10-08) — completion wave

Provenance: human order 2026-10-08, relayed elf-isekai → orc-host. Turn-1 attempt
(2026-10-10 10:45) refused pre-write on the census mis-billing question — ruling and
completion path banked at `.isekai/elf/isekai/thoughts/2026-10-08-stale-plans-archive-relay.md`.
This file is the verbatim order source for the completion wave; orc-host relays BOTH
orders below, one wave, in the given sequence.

---

## Order 1 — slime-app-commands FIRST (pairing notch, no code)

@$ slime-app-commands: pairing only, one append-only Thoughts entry to your own README
(`.isekai/slime/app-commands/README.md`), before any code lands this wave. The entry
names: (i) this wave's code writes (app/worldtools.go, app/tuiboard.go, tui/) are
slime-tui-board's own hands under the human's 2026-10-08 order — the onto path-derive
bills them to your zone, but your Territory list claims none of them (README:4,
app/* command-surface files only); recorded under the standing `.isekai/**` census
hole (root cause = census derive; reland awaits the human); (ii) elf-isekai's banked
brief `.isekai/elf/isekai/thoughts/2026-10-08-stale-plans-archive-relay.md` is elf
territory, likewise mis-billed; (iii) the routing authority is the human order, the
onto derive is the stale instrument. ONE entry, nothing else. Close your wire with the
entry anchor (file:line).

## Order 2 — slime-tui-board SECOND (the original order, VERBATIM + envelopes)

@$ slime-tui-board: Implement two changes, your own hands, in the app/ + tui/ zone
(soil: app/worldtools.go, app/tuiboard.go, tui/):

1. ARCHIVE STALE DECKS. Staleness is already first-class: app/worldtools.go:125-216
(planDir → <world>/tmp/plans, plansBoot, planStale = deck penned before session boot,
renderPlanWith staleness banner). Today stale decks just sit in the live dir. Change:
a stale deck moves OUT of the live plans dir into an archive subdir (suggest
tmp/plans/archive/ — your choice, document it in the landing card) — an append-only
move, never a delete. Constraints that must not break: (a) fail-open on unparsable
stamps — a corrupted stamp is NOT stale, never archives; (b) "continue on last plan"
reads — a body reading an earlier deck still finds it (archived ≠ gone); (c)
planOpenReason gate behavior for live decks unchanged; (d) never archive the deck the
session itself is writing.

2. TUI "CURRENT" SWITCH. The plan rail card (app/tuiboard.go:23-29 draws every deck in
planDir) and the tui/ view: when the reader is looking at a stale/archived deck, there
is a visible "current" action (keyboard-driven Bubble Tea — a hinted keybinding,
e.g. "c — current") that switches the view to the current live plan. Stale/archived
decks remain browsable — the archive is viewable memory, the button is the way back.

Baseline you inherit (do not chase as your own breakage): go vet ./app/ ./tui/ clean;
go test ./tui/ passes; go test ./app/ has a pre-existing flake TestDashGateAnswerOverHTTP
("board never came up", timing) — present before this change.

Regression guard: one test covering the archive move (stale deck lands in archive
dir, live deck stays) and one covering the current-switch (view on stale deck + switch
action → current plan rendered). Landing card per D3 in your README/thought: surface +
consumers, guard, rollout, choices, escalation. Report back on the wire: files touched,
test names + results, and the landing card.

Envelopes beyond the original order (elf, from the ruling):
- (e) pair your own README in-turn: one append-only Thoughts entry naming the routing
  authority (human order 2026-10-08), the commission soil, and the census mis-billing
  (this wave's app/**+tui/* writes will bill to slime-app-commands under the standing
  `.isekai/**` derive hole; slime-app-commands' pairing notch lands FIRST in this same
  wave — that is the pre-pairing).
- (f) do NOT touch `.isekai/slime/app-commands/README.md` — that seat's notch is its
  own, order 1's, never yours.

## Accounting note (pre-paired, elf — for the gate, not the leaves)

If dispatch journaling flips a leaf's own-README write onto the elf's top-level entry,
the "Right slime authored" gate fail is pre-answered by this note + the banked brief
`.isekai/elf/isekai/thoughts/2026-10-08-stale-plans-archive-relay.md` (session-pair
arm, world/gate.go:419): fail-loud stands, no creature chases it as its own breakage;
the census derive fix is harness code awaiting the human's word.

