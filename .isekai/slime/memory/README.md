# slime-memory

- **Rank:** Slime
- **Territory:** `memory/`
- **Reports to:** orc-world
- **Minds:** slime-go-proving, slime-wire-and-journals
- **Purpose:** the ground truth of the memory instrument: the three tiers (short, long, shared), the index, recall ranked by meaning then relation, remember, forget, status — the Go port of memory.js that must stay interchangeable with it.

## Traits
- IndexV 2, DeskLimit 5, PipeBuf 4096. Files: memory/long/index.json (derived, rebuildable, gitignored), memory/short/<creature>.jsonl (recall cache, disposable), memory/shared/notes.jsonl (world-shared, append-only, tracked) and the machine tier ~/.isekai/shared/notes.jsonl. WriteAtomic and AppendLine are the only writers.
- Sources indexed: log.md entries (episodic), Minds and commands (procedural — MindBases .opencode/skills, .opencode/skill, .claude/skills; CommandBases .claude/commands, .opencode/commands, .opencode/command), the law, creature docs under .isekai/{elf,orc,slime,kijin} (Creatures), canon, README (semantic).
- Ranking is local and model-free: BM25 with the stop list and Tokens tokenizer, cosine over IDF-weighted vectors, Recency, then a relation boost when a memory names the asker, its orc or a worn mind. NormAs canonicalises the asker's name.
- Status reports silent tiers as holes (no index, HEAD moved since the index was built — GitHead — no transcript); a stale index is a `@?`, and the world's Recall hook rebuilds it.
- js.go reproduces JavaScript semantics on purpose: UTF16Len and UTF16Slice, JSTrim's whitespace set, LocaleCompare with ICU punctuation order, ToFixed via big.Rat, JSNum formatting — so `memory status|recall` output equals memory.js's byte for byte. A "simplification" here is a divergence.
- remember --kind law|colony|territory are the unsaid kinds (Unsaid, IsUnsaid); Forget clears the short tier for a creature. CLI ParseArgs is the JS argv parser; FormatResult renders JSON or the text the JS tool prints.

## Verify
- `go test ./memory/...`

## Thoughts
