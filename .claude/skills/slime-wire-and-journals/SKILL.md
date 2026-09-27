---
name: slime-wire-and-journals
description: The machine formats the isekai binary reads and writes — the wire envelope (@S @F @V @? @U @E, commission tags), the loop journal events, the usage journal, log.md entries, and the memory/toolbox files shared with the JS instruments. Wear when a change touches any of those formats or a package that parses or renders them.
---

- Report (a Court's answer, a hook's stdout, `run --format wire`): first line `@S <status word>`; then `@F` findings (file:line fact), `@V` verdicts with evidence, `@?` holes (named, never guessed), `@U <law|colony|territory> text` the unsaid, `@E <n>` the size. Parsed by wire.ParseReport (ok only with @S); rendered by Report.Emit(cap, dump) — over cap it is cut and the cut becomes a hole. DefaultCap 2048.
- Commission (going down to a Court): `@ROOT`, `@SCOPE`, `@ASK`, `@CAP`, `@DUMP`, `@SIZE`; tool.Commission.String() renders it, world.Gate's Duties-done check parses it back and fails a commissioned turn whose answer carries no @S.
- On an openai-compatible provider a Court's report can be guided-decoded to openai.WireSchema (S, F, V, Q, U) and rendered back by RenderWire — keep the schema and the tags in step.
- Loop journal: .isekai/instruments/loop/<run-id>.jsonl, one JSON object per beat with "t" ∈ perceive, gate, act, verify, record, hook, drain-report, drain-abort and "id" s1, s2…; loop.ReadJournal / State / Runs rebuild a run; the board's Run view and `resume` read it. Add a field, never rename one.
- Usage journal: .isekai/instruments/usage/<session>.jsonl, app.UsageRecord per provider call (session, body, office, rank, model, input/output/cacheRead/cacheWrite, cost); `isekai usage` and the board's /usage roll it up (Ranges 24h 7d 30d all).
- log.md entry (world.Entry.Render): `### [RFC3339] <author> — <title>` then `- **Task:**`, `- **Files:**`, `- **Gate:**`, `- **Result:**`, `- **Learned:**` (one line or sub-bullets). Appended only through world.Log, which refuses when the remembered prefix changed (Law 4).
- Memory files (shared with .isekai/tools/memory.js): memory/long/index.json (IndexV 2), memory/short/<creature>.jsonl, memory/shared/notes.jsonl (append-only JSONL). Toolbox files (shared with toolbox.js): toolbox/registry.json (RegV 1), toolbox/extra.jsonl, loads under instruments/toolbox. Output of `memory status|recall` and `toolbox brief|status` must equal the JS tools' byte for byte — memory/js.go exists for that.
- Desk: .isekai/instruments/desk/<session>.md, `# desk — <body> — <date>` then `- thought` lines, at most 5 before it is a stress reading.

## Thoughts
