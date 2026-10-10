# Census wave 2/2 — re-run relay brief (elf-isekai, 2026-10-11)

Carries the throne order 2026-10-08 ("census wave 2 of 2, take 2 — first dispatch landed no journal; re-dispatch per spawn-fail protocol"). One commission file, three rungs; each orc reads only its own rung plus the common armor. The throne's order text is embedded in rung context, not re-quoted verbatim — this file is the verbatim-deliverable down the chain.

## Why the re-run looks different from the 2026-10-08 attempt

Fresh disk-read this session contradicts the order's zone billing, recorded here as rung context, not silently dropped (all anchors byte-true at write time):

| Order claims | Disk truth | Ruling |
|---|---|---|
| slime-app-court DONE wave-1 (README.md:13 logger) | false — line 13 today is the app/hooks.go trait; the RESOLVED-DRIFT logger ledger sits at :17. Trait text deepened post-census at SOME commit ≤ c93055a (11:29, un-fingerprinted — waves 1-4 churn all across 10-08, briefs orders/20261008-* were NOT part of any wave commit). Commission pins must carry PRE-DISPATCH content-verifie
| slime-app-engine "no census traits landed" | false — .isekai/slime/app-engine/README.md:40-44 carries 15 verified census anchors + 4 DRIFT-FLAGS, landed 2026-10-07 (orc-host chronicle, slime-app-engine-202610075442.report.md) | demoted to verify-only, no writes (rung 1) |
| slime-app-commands "pairing notch landed README:108" | false — notch is the ONTO PREFIX-COLLISION section at :108-116; the pairing FALLBACK trait sits at :37 (mis-billing armor paired to slime-cmd zone). Tail anchor = session.go census line :117-119. Stale-plans anchors RE-verified this session after orc-host's stop: real code is app/worldtools.go — planRead :81, planArchiveDir rchivedIntoLiveView; `world/worldtools.go` DOES NOT EXIST (world.go + worldtools_test.go only) — the order's `world/` prefix on every anchor was wave-2 drift inherited from that night's sealed briefs. Tests green this session (go test -run TestPlanArchiveSweep ./app PASS 0.417s). | trait line RE-PINNED, evidence embedded, verify-envelope worded (rung 2) |
| slime-loop / slime-memory "no census traits" | false — loop README:73-81 first wave recorded trait blocks landed 2026-10-05 (memory zone timed out first take, closed with app-docs trait correction); both READMEs carry full Traits sections today | folded into rung 1 verify-only arm |
| slime-bench / slime-release / slime-found / slime-app-dash | NO such zone on the roster — `onto roster`: .isekai/** bills to slime-app-commands wholesale (elf @U colony law 2026-10-11); no slime zone owns bench/ (orc-bench = bench-root, slime-fakevllm owns bench/harbor + fakevllm); CHANGELOG.md lives in elf territory | @? holes to the throne, no guessing slimes (rung 3) |

## Common armor (binds every leaf, quote then comply)

1. **Anchored edits only** — wave 2 of this exact shape was the full-file-write head truncation that took app-commands :1-38 and app-engine :1-30 (orc-host census ledger). Repair sequencing: confirm-once head guards (line 1 exact line) before AND after; anchored append only; byte-stability = diff hunks touch ONLY the intended lines; line count delta = added lines only. Law 17 app
2. **Known onto mis-billing armor** — `.isekai/slime/<zone>/README.md` writes may gate-block billed to slime-app-commands (ZoneOwner prefix-trie, `.isekai/** → slime/app/...` lines 1-6 + 9-21 seed it). Fallback: report the write standing per fail "writes stand as they are", the orc seals the pair, name the strike on the wire. Never counter-write a refused file from the wrong hand
3. **Survey mode** — Verify build-acts suspended by this brief's order except: (a) the two greps/`go run ./cmd/isekai onto owner` read-lanes named per rung; (b) rung 2's confirm-once red tests as instructed. The suspension is reportable on the wire (@S line), never silent.
4. **One leaf at a time, one dispatch per leaf.** No full-file rewrites. Leave existing dirt standing (porcelain law) unless the rung names it.
5. **Report per leaf**: lines landed (with `grep -Fn` confirm-once outputs), byte-stability check, mtime/size before+after, suspension declared.

## Rung 1 — orc-world: census wave-2 verify-only arm (no writes unless drift)

Verify, do not write, six zones. For each: grep -Fn the pinned phrase; report line number + byte-true verdict. Any drift → @? to elf, do NOT repair.

- slime-app-court README:13 = "- app/hooks.go holds the beat hooks; app/goal.go the goal scoring; app/logger.go: newLogger only; app/usage.go the token spend ledger." — line 13 moved post-wave-1 (RESOLVED-DRIFT ledger took the logger correction to :17; trait text deepened at an un-fingerprinted commit ≤ c93055a).
- slime-app-court README:17 = marker "- RESOLVED-DRIFT" — ledger correcting the per-body-log-multiplexer error; NOT_FOUND (multiplexer site unsettled is lawful).
- slime-app-engine README line 1 = `# slime-app-engine` + census trait blocks stand — includes the pair-6a verify block anchoring app/worldtools.go:23 worldTools seam among them (app/ naming, wave-2's corrected geography helps here); four 2026-10-07 DRIFT-FLAG lines at tail (grep `DRIFT-FLAG 2026-10-07` → expect 4); the 2026-10-08 wave-2 truncation lesson record stands.
- slime-app-commands README:110 section header "## ONTO PREFIX-COLLISION (zone-billing, 2026-10-11)" + "zone protected from sliming" at :113 and :115 + :116 "Next census wave ... can un-slime ..." fingerprint trio stands; the pairing FALLBACK trait stands at :37 (grep -Fn 'must therefore carry a self-pairing notch' → expect :37).
- slime-loop README census traits stand — grep "captured during census" → expect 22 total anchored zone lines; grep "captured during liveliness check" → expect 4+ (byte-true test = dated census tags present; the 7-vs-4 count gap between brief header and loop surface is a recorded Known, not a fail gate).
- slime-memory README — confirm 6 census traits stand; specifically zip.go:35 evidence-*- prefix + zip.go:37 gob-encode regressed-since-2026-09-17 trait.
- If all six byte-true: wire DONE with the six confirm-once outputs; zero writes, zero slime dispatches. This rung is orc-world eyes only.

## Rung 2 — orc-host: census wave-2 write arm (one leaf: slime-app-commands)

One leaf, one dispatch, one edit site.

1. Confirm-once BEFORE dispatch: `grep -Fn 'zone protected from sliming' .isekai/slime/app-commands/README.md` → expect :113 and :115. `sed -n '1p'` → `# slime-app-commands`. `sed -n '117,119p'` → the SessionStore census line ending "session APIs ride app/session.go, not census-derive." If any fails → STOP, @?.
2. Dispatch slime-app-commands (fresh slime, one session, no inherited pacing/strain memory). If the slime lane dies silent (precedent LEC-2026-10-04-STALE-PLANS take-3, kimi-3-conduit silent on this shape twice), do NOT re-book same-shape — wire the failure evidence @?. The slime's instruction, verbatim-deliver:
   - First compile your own diagnostic: `go build ./...` and `go test -count=1 -run 'TestPlanArchiveSweep|TestPlanMergeArchivedIntoLiveView' ./app` in YOUR OWN drain. If red → STOP, wire output as findings, zero writes. (The anchors below ride this brief as elf-verified facts this session — verify the envelope, never re-derive; try-go compile is honest-build under probe-lie per bench
   - If green: append ONE single anchored trait line immediately after line 119 (tail phrase "not census-derive.") via the `edit` tool. Line text (single line):
     `- 2026-10-11 census wave-2: stale-plans archive surface — app/worldtools.go:81 planRead, :159 planArchiveDir (live+archive under plans dir), app/tuiboard.go:29 Plans() merge via :38 railPlans; guards TestPlanArchiveSweep app/app_test.go:63 + TestPlanMergeArchivedIntoLiveView app/app_test.go:102 green 2026-10-11.`
   - After edit: build + the two tests again green; `wc -l` delta = +1 exactly; `git diff --stat` names ONLY this README. Report all outputs with grep -Fn confirmations.
3. Byte-stability: diff hunks touch ONLY the new line. Everything else byte-stable. No full-file write.
4. Slime compiles its own diagnostic per step 2 (try-go, honest-build); survey-mode suspension covers the old raft only. Elf-side independent anchors already banked this session (all grep-verified): app/worldtools.go:81 planRead, :159 planArchiveDir, app/tuiboard.go:29 Plans() via :38 railPlans, TestPlanArchiveSweep ./app green 0.417s, HEAD c93055a; the stale-plans commit 534bbbb si ch is exactly why the first rung-2 brief's anchor map described a tree that did not yet exist.

## Rung 3 — orc-bench + orc-world: census wave-2 hole arm (no dispatches, holes only)

The four remaining billed zones are not on the onto roster. No slime to one-hop to. This rung is eyes-only; each orc appends ONE short chronicle line admitting the hole and wires @? up the chain. Do NOT invent slimes, do NOT dispatch, do NOT write leaf READMEs.

- orc-bench: already sealed (report orc-bench-202610122236.report.md — chronicle line landed .isekai/orc/bench/README.md:26, byte-surgeon discipline held, owned-file-only). fakevllm README byte-true verify outstanding in that report: brief's original pin 'officials run on :18001' is drifted-off (the README preamble today reads "Vendor-faithful OpenAI+Anthropic+llm-d fake; modes: str p -Fn 'default :18001' .isekai/slime/fakevllm/README.md and grep -n 'DRIFT-FLAG 2026-10-02' → expect 5 lines (DRIFT armory digest verified 2026-10-08). orc-world folds this verify into its re-dispatch if claimed.
- orc-world: @? three names — (a) slime-release: no such zone; CHANGELOG.md is elf territory (elf census-billed), release verify surfaces (.goreleaser Tag=LatestReleaseTag guard) sit inside slime-cmd's README 2026-10-08 trait. Throne to rule: crown a slime-release zone, or accept cmd-zone coverage as the census state. (b) slime-found: no zone, no `find`/`cmd-find` on disk (find = co unzoned by onto (`onto owner .isekai` bills slime-app-commands wholesale; dashboards = board/ soil per slime-tui-board traits, http-server = app/board.go per slime-app-court :14). Throne to rule the owner before any trait lands.

### THRONE RULINGS (2026-10-08, rimuru) — rung-3 holes CLOSED

1. **slime-release**: no such zone on disk (onto roster + census both found nothing) — cmd-zone coverage accepted as-is; CHANGELOG.md remains elf census territory. No new slime crowned. **Closes rung-3 hole (a).**
2. **slime-found**: `found` is not a disk truth — struck from census billing. It was desk-note phrasing in the original 10-read grouping, never a zone. **Closes rung-3 hole (b).**
3. **slime-app-dash**: split mapping — dashboards surface = board/ soil (slime-tui-board), http-server = app/board.go (app-court coverage, slime-app-court :14). Coverage accepted as-is; no new zone crowned. **Closes rung-3 hole (c).**

Record: these three rulings close all rung-3 @? holes (.isekai/orc/bench/README.md:26 chronicle + this section). The census needs no re-runs from these: zones verified, no gaps. Also on the record: the four wave-2 rung-1/4 briefs quoted world/worldtools.go anchors for code that lived in app/worldtools.go at brief-time — the gates caught it fail-loud; this amended re-run brief re-pi s — throne-verified).

## Close-out

Each orc seals its own chronicle (Vitality) and reports on the wire with @$ echo. Elf consolidates to the throne: six verifies, one write, three holes. The census wave-2 re-run is then DONE-with-holes, not silently partial.

