# slime-memory

- **Rank:** Slime
- **Territory:** `memory/`
- **Reports to:** orc-world
- **Minds:** slime-go-proving, slime-wire-and-journals
- **Purpose:** the ground truth of the memory instrument: the three tiers (short, long, shared), the index, recall ranked by meaning then kinship, remember, forget, status, and the review and clean arms — the Go port of memory.js that must stay interchangeable with it.

## Traits
- IndexV 2, DeskLimit 5, PipeBuf 4096. Files under the world dir: memory/long/index.json (derived, rebuildable), memory/short/<creature>.jsonl (recall cache, disposable), memory/shared/notes.jsonl (world-shared, append-only, tracked) and the machine tier ~/.isekai/shared/notes.jsonl. WriteAtomic and AppendLine are the only writers.
- Sources indexed (sources()): log.md (episodic); Minds from MindBases (crafted-skills and base-skills under the world dir, then .opencode/skills, .opencode/skill, .claude/skills), commands from CommandBases and the world's tools dir (procedural); each creature's `thoughts/*.md` (procedural); the law, creature docs of Races elf, orc, slime, kijin, canon and the root README (semantic).
- Ranking is local and model-free: BM25 with the stop list and Tokens tokenizer, cosine over IDF-weighted vectors, recency, then RelationBoost — kin first: the asker's own creature +0.25 (else same race +0.10), its name +0.15, parent and a child +0.10, a worn mind or its zone +0.08, its own shared note +0.05. A body's own doc and thoughts must outrank the chronicle (TestRecallKinFirst).
- Status reports silent tiers as holes (no index, HEAD moved since the index was built, no transcript); a stale index is a `@?`, and the world's Recall hook rebuilds it unless memory.long.rebuildOnStale is off.
- js.go reproduces JavaScript semantics on purpose (UTF16Len/UTF16Slice, JSTrim, LocaleCompare, ToFixed via big.Rat, JSNum) so status and recall output equal memory.js's byte for byte; a "simplification" here is a divergence. The interop tests skip when node or the JS tool is absent.
- review.go: `memory review` reads four tiers through their own lens — short (stems asked across bodies), shared (lessons restated), long (staleness, procedural thinness, recurring friction), value (today's journals, gate fails) — and routes each finding to [genesis], [skill], [law] or [harness]. Review never writes.
- scrub.go: `memory clean` runs the cure deck (ScrubRules) over log.md, both notes files and the short caches. Dry-run touches nothing; `--apply` backs each file up byte-exact under the world's tmp/scrub, rewrites atomically, and reseeds the long index. Records (log, notes) normalise spans in place and never lose lines; only a cache line whose query is left with no content tokens drops.
- remember --kind law|colony|territory are the unsaid kinds (Unsaid, IsUnsaid); Forget clears the short tier for a creature. CLI verbs: status, index, recall, remember, forget, clean, review, selftest; ParseArgs is the JS argv parser.

## Verify
- `go test -count=1 -run 'TestRecallKinFirst|TestReviewTiers|TestScrubDryVsApply' ./memory/`
- `go test -count=1 -run 'TestSelftest|TestJSHelpers|TestSectionsAndDesks|TestAPI|TestInteropWithMemoryJS' ./memory/`

## Thoughts
- (none yet)
