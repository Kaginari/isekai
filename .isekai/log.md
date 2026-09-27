# Chronicle — change log

Append-only. Newest entries at the bottom. One entry per change.

---

### [2026-09-27T21:19:56Z] rimuru — founding: the colony of the isekai engine (init --level complex)
- **Task:** read every Go package, the build, CI and release files; decide the team the code needs; write creatures, minds and bodies; verify lines run first; onto check clean
- **Files:** .isekai/elf/isekai/README.md, .isekai/orc/{engine,safety,world,interfaces,config,app}/README.md, .isekai/slime/{loop,providers,wire,compact,guard,gate,sandbox,tools,world,onto,memory,toolbox,tui,board,ui,config,discover,mcp,app,bench}/README.md, .isekai/kijin/release/README.md, .claude/skills/{slime-go-proving,slime-class-contract,slime-wire-and-journals,slime-creature-docs,orc-gate-verdict,kijin-release-pipeline}/SKILL.md, .claude/agents/{elf-isekai,orc-engine,orc-safety,orc-world,orc-interfaces,orc-config,orc-app,kijin-release}.md
- **Gate:** n/a (founding; law.gate.rightAuthor off for this session) — bin/isekai onto check: @S PASS 0 findings · 29 creatures · 6 minds · 635 triples
- **Result:** founded elf-isekai; orcs engine (loop, instrument, provider, wire, compact), safety (guard, gate, sandbox, shell, tool), world (world, onto, memory, toolbox), interfaces (tui, board, ui, docs/screens), config (config, discover, mcp, examples), app (app, cmd, bench); 20 slimes with disjoint territories; kijin-release (.github, .goreleaser.yaml, Dockerfile.release, install.sh, RELEASING.md, CHANGELOG.md) as the keeper; every verify line ran and passed (go build/vet, go test per package, gofmt clean, CHANGELOG has [Unreleased], install.sh and smoke.sh parse, fakevllm builds); a previous founder crashed after init --level light, so init was not rerun
- **Learned:**
  - The anthropic provider has no retry; the openai provider retries 429/500/502/503/504 three times with Retry-After — asymmetric on purpose or by omission, worth a second naming
  - sandbox.ScrubEnv only drops names ending in API_KEY, TOKEN, SECRET, PASSWORD, PASSWD: AWS_SECRET_ACCESS_KEY passes through unless bash.envDrop names it (trait on slime-sandbox and orc-safety)
  - app's TestEmbeddedLawsMatchTheWorld compares app/law/isekai.md with this repository's own .isekai/isekai.md — now that a world exists here, the two must stay byte-equal (Law 6 for both)
  - onto's territory derivation takes every back-ticked token with a slash anywhere in a doc: foreign paths in traits must be plain text or a slime overlaps its neighbour
  - docs/canon/README.md points at docs/canon/agents/ which does not exist; docs/canon/binary.md says Go ≥ 1.23 and names cmd/agent-one — canon lags code (elf trait)
  - bench/fakevllm is a separate Go module: the root go test/vet skip it; goreleaser is not installed locally, so goreleaser check is CI-only
  - Config shell hooks run with sh -c outside the sandbox with the full process environment (trait on slime-app)
  - Two budget defaults coexist: loop's constants (50 steps, 30 min, 2 retries) and config's law.budget (40 steps, 60 min); config wins when loaded
