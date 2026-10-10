---
name: review-core
description: How a body here reviews, reads and acts as an instrument — the verdict/draft vocabulary, the test-golden conventions, the naming law (dirs → creature names, dotted → paths), and the day's heartbeat: build + go test -short. Worn by orcs at their gates and by the elf when speaking.
---

# review-core

- Verdict vocabulary lives in app/review.go: verdict kinds (pass / fail-bug / fail-feature / fail-gate) arrive from a Court to the dispatcher; revise docs in the seat that owns the fact, not the verdict text.
- Golden tests live as files under testdata/golden/ — UI shots in ui/shots.go, markdown fences in tui/markdown_test.go, init pages in app/init_render_test.go. A golden test with _no_ testdata file fails loudly, this is the mechanism to trust.
- Conventions sticky across the code: package path as a sentence ("the wire keeps a report's tail"); file names describe the one thing the file does (cli.go for command surface, never helpers.go); test funcs are single sentences in CamelCase.
- Build: `go build -o .isekai/tmp/isekai ./cmd/isekai` (Makefile puts the binary in `.isekai/tmp/`; bin/ is in .gitignore). Package tests: `go test -short ./...`. The bench suite is not `-short` — see bench-run mind for bench-legality.
- gate verdicts flow: tool act → guard pass → gate/gate.go Needs(class) → ask? → verdict stamped in wire + logged in .isekai/log.md.
- When reading a finding, look first for the package's own README (pointers land there), then HANDOFF.md, then the log.
- Naming law reminder: `.isekai/<race>/<short>/README.md` — name stems from folder (`orc/gatekeeper/` → `orc-gatekeeper`); onto owns claims extract backticked paths from the doc.

