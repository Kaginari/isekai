# slime-world

- **Rank:** Slime
- **Territory:** `world/`
- **Reports to:** orc-world
- **Minds:** slime-go-proving, slime-creature-docs, slime-wire-and-journals
- **Purpose:** the ground truth of the world package: discovery of the root, the law loader, the lexicon, ranks and creatures, territory enforcement, the Orc gate with Vitality, the append-only log, the system prompt, the loop hooks and the Court dispatcher.

## Traits
- lexicon.json holds both vocabularies (isekai, agent-one): Ranks dirs (elf, orc, slime, kijin, dark-elf — note the hyphen for dark_elf), Prefixes, Display names, unsaid tokens, check names, the crest heading. Races is elf, orc, slime, kijin.
- DefaultRanks is the law's table: elf → rimuru (court, office ciel, all tools incl. dispatch); orc → elf (holds the gate, sideways, raphael); slime → orc (authors, great-sage, baseTools WITHOUT dispatch); kijin → rimuru (authors and holds the gate, keeper, raphael); high_elf/high_orc carry Base and share the base dir; dark_elf is a keeper with read, bash, glob, grep, law only. Ranks.Validate refuses a cycle, a missing parent, and an authoring rank with no gate above.
- Gate(): RightAuthor fails when a written file belongs to another creature and is not in the writer's territory, or lies outside a slime's declared territory; DocTruthful fails when a file under an owner's territory changed and the owner's doc did not change in the same turn; TraitsHold runs every verify line through `bash -c` in the root (default timeout 120 s, exit 124 on timeout, 160-char tail) — a verify line removed this turn is a hole and still runs; DutiesDone parses the commission; the word is pass, fail, or the lexicon's GateNA when no gate holder exists. `init --level complex` runs with law.gate.rightAuthor off so a founder can write every doc.
- Dispatcher: a Court is minted only for a name on the roster; work goes down the ranks (or orc ⇄ orc), never up; the child's step and minute budgets are the parent's remainder; the Court runs on the wire (e.Wire = true) with the commission's cap; its context dies in Close; what it wrote and gated is subtracted from the parent's own writes (courtGated / ownWrites).
- Policy(body) enforces territory only for ranks that author (Authors) and only on write-class acts with paths; the refusal carries the one-hop-up hint.
- LoadLaw splits the law by headings: the crest section is always in the prompt; the `law` tool loads any other section by title on demand. Prompt order: crest → identity → rules in force → creature line (doc, territory, one hop up, minds, verify) → @ONTO projection (800 tokens) → instructions → the memory line.
- log.Entry renders `### [RFC3339] author — title` then Task, Files, Gate, Result, Learned bullets; OpenLog + Append are the only writers.
- Recall rebuilds a stale memory index or toolbox registry instead of routing around it; Record lands @U lines from tool output; hooks.go MergeHooks composes the world's hooks with the app's.

## Verify
- `go test ./world/...`

## Thoughts
