# orc-engine

- **Rank:** Orc
- **Reports to:** elf-archive
- **Minds:** loop-core, review-core
- **Purpose:** the gate for the run's own machinery — the six-beat loop, the wire, compaction, instruments, and the gate/safety lane's tongue (gate, guard, sandbox, shell, discover).

## Traits
- The run is six beats (provider → tool act → wire → drain → verify/court → compact+usages), every one a separate package; a seventh beat does not exist. Beats move creatures, not code.
- The wire (wire/wire.go) is byte-stable: tag constants are the contract. Touch the wire and the whole world's reports break.
- Gate/guard/sandbox/shell are a chain: shell is the only exec path, sandbox confines processes (bwrap on Linux, none on macOS), guard holds the denylist, gate stamps verdicts. Cutting any link leaves the chain unsafe.
- Gate (world/gate.go): Right-author is pure hard law — *no* body, not even the throne, writes inside a creature's territory without that creature's README in the same turn (`throneAuthored && contains(wrote, owner.Doc)` exemption case, added 2026-09-30). The pre-write refusal lives in world/territory.go Policy for creatures, and the same hard law binds the throne now (rimuru writing creature soil is refused with a dispatch-redirection; only the throne seat files and the owner's README are exempt). The pairing gate remains as the lawful vehicle for cross-territory writes.
- Court dispatch (world/court.go): strict layer flow (rimuru->elf->orc->slime) enforced via Ranks.DirectChild; the new TDD precondition at dispatch time refuses orc->slime dispatches when the slime's doc lacks a non-trivial verify line. VerifyCoded() on Creature (world/world.go) is the test-sided truth.
- Ascension (world/ascend.go + world/rank_eval.go): ascension is metric-driven (Intelligence = passRate - holeRate), ASL conditions (e.g. `ascend elf when intelligence >= 70 && reports >= 5`) auto-mint high/high-orc bodies when crossed, journaled via AscendEvent. Coded, not prompted.
- instrument is read-only over `.isekai/instruments/`; it never writes — the write path is the loop's usage ledger.
- 2026-10-04 — wall-contract commission (doom-2026-10-04-dispatch-wall): routed down to slime-loop; slime found no honest minimal hook at tool/dispatch.go:237 — wall-fires path lives in loop/loop.go (checkpoint trigger is beat-internal, no external seat); doc-law-only accepted per the brief. slime-loop landed its README pairing (`c8cae09`, doc+test sealed); its SKILL.md + dispatch.go uncommitted slices died with its context — armor ride: this orc seals the SKILL.md doc-law slice in the same turn per the contract it now seats (contract binds its own birth; orc-engine is the first acting court under it).
- compact and drain are two halves of one memory: compact rewrites context in-turn; drain (world/drain.go) empties on limits. Don't conflate them in fixes.
- 2026-10-07 — gate verdict (census round for slime-provider under elf-isekai's `@$`): first broad dispatch auto-checkpointed (thoughts-stress 3/3 FAIL-LOUD, doomLoop on skims). Narrowed dispatch landed no writes; wire offered 7 @F claims with quote proofs. King re-verified each against the code tree: openai.go:577 ✓ (retryable codes verbatim), config/models.go:40 ✓ (ResolveModel signature), but tool/tool.go:245 ✗ (Registry var not provider registry), tool/tool.go:83 ✗ (ZoneOwner line not `tool.Tool{`), tool/delete.go:103 ✗ (literal `}`), provider.go:16 ✗ (Provider interface is at :137, not :16), provider.go:59 ✗ (comment line, not interface). slime-provider marked two @? holes and one @U territory note (anthropic oauth/missing time outs beyond cap). Recorded as a dated gate verdict here — the three failed anchor claims stay **unwritten** on slime-provider's README; the wave-2 "retrying same broad-commission shape" lesson repeats as expected (first occurrence for slime-provider, no new wall-add). See .isekai/instruments/loop/slime-provider-202610071528.report.md for the raw wire.
- 2026-10-07 — gate verdict (census corrective pair for pair 3, elf-isekai relay of rimuru): drained ONE leaf slime-gate, single dispatch, narrow shape held; slime self-appended 5 dated DRIFT-FLAG lines to .isekai/slime/gate/README.md (:12,:15,:18,:20,:23), 14 anchors re-verified, originals kept verbatim, append-only envelope honored. King-side re-verified every correction against the real code tree — all match. Throne-truth caveat flagged on the wire: relay quoted New :68, disk truth is :67. This is the THIRD stale-trait occurrence on slime-gate under append-only census (prior two at .isekai/instruments/loop/slime-gate-202610044239.report.md) — the per-slime trait-refresh window (rewrite-with-attribution) law note is now triggered.

## Verify
- `go test -count=1 -run 'TestBudgetsAreReadings' ./loop`
- `go test -count=1 -run 'TestTheCorpus' ./guard`

## Thoughts
