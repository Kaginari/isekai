---
name: orc-gate-verdict
description: The review an orc (or the kijin) applies before a change lands in the isekai repository — the four gate checks as world.Gate implements them, plus this codebase's own rules (honesty switches, lexicon words, law copies, records, corpus, README and CHANGELOG). Wear when judging a slime's work or writing a verdict in log.md.
---

- Right author: every written path is inside the authoring slime's Territory (world.Gate check 1; onto owns edges). A path nobody owns is a routing gap — a finding, then a birth when named twice (Nature 4), never a silent widening of a territory.
- Traits hold: run the owner's Verify lines as they stood before the turn (`go test ./<pkg>/...`, plus `test -z "$(gofmt -l .)"` and `go vet ./...` for anything Go). A verify line removed or changed this turn is a hole the orc confirms. Tests intact: no test deleted, skipped or narrowed to pass (world.testsIntact; loop.IsTestFile decides what is a test).
- Duties done: the commission's ask answered; a wire commission answered with @S; @U when +unsaid was asked. Doc truthful (Vitality): the owner's README changed in the same turn when behaviour, shape or an invariant changed — a stale trait is a fail, not a nit.
- Honesty rule (config): a new feature has an `enabled` switch with an origin, its default is in config/defaults.go and docs/canon/config.md; a switched-off law feature shows in `status` as `@? off`. A behaviour added to world/, loop/ or app/ without a switch fails here.
- Words: user-facing text in a shared package goes through world.Lexicon / config.Dist.Word (both distributions); no hard-coded "isekai", "slime", ".isekai" outside the isekai lexicon entries and the documented fallbacks (tool.DefaultWorldDir, memory.DefaultWorldDir).
- Law copies: app/law/isekai.md must equal .isekai/isekai.md byte for byte (app's TestEmbeddedLawsMatchTheWorld); neither changes without Veldora's order (Law 6). Records (log.md, notes.jsonl, runs.jsonl) are appended, never rewritten.
- Safety loosenings need their evidence in the same change: a wider readOnly allowlist, a new git read subcommand, a softened guard pattern, a permission `allow` — the corpus updated, `go test ./guard/... ./tool/...` green, and the reason in the log entry.
- Surface changes: a new command, flag or config key updates README.md (CLI table, Configuration) and the `## [Unreleased]` section of CHANGELOG.md; a visual change refreshes docs/screens; a format change (wire, journal, memory, toolbox files) keeps the JS instruments' compatibility.
- Verdict: `pass` or `fail — reasons`, appended to .isekai/log.md as an entry (Gate line), never edited afterwards. When in doubt the answer is `@?` up to the elf, not a guess.

## Thoughts
