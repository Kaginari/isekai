Found the colony of this {{WORLD}} — you are hiring a team of experts for `{{PROJECT}}` ({{LANG}}).
The law is `{{WORLD}}/isekai.md`; its sections "Minds & Bodies" and "The world" are the rules for
what you create. Nothing outside `{{WORLD}}/`, `.claude/skills/` and `.claude/agents/` is written;
the project's code is read, never changed.

## 1. Read everything first

List every source file (the find/glob tool), then read them — all of them, not a sample: entry
points, packages, models, services, tests, build and CI files, the README. Note as you go:
- the domains: areas that change for their own reasons and need their own judgement (a gate);
- the vertical slices inside each domain: the smallest units someone owns the truth of;
- the conventions every slice repeats (how a new one is added), the contracts they share;
- the hazards: what can break production, leak a secret, open access, or cost money;
- how the project proves itself: the commands that build and test it, and whether they pass now.

## 2. Decide the team — by what the code needs, not by a quota

- **One elf** (`{{VOICE}}<name>`) — the shared mind and voice: owns the cross-domain rules and the
  project-wide files (README, the module/package manifest). A second elf only when the project is
  really two products with separate audiences.
- **Orcs** (`{{GATE}}<domain>`) — one per domain that deserves its own gate. Two to five is usual; one
  for a small tool. Merge domains too thin to judge on their own; never make an orc per folder
  when the folders are layers of one feature.
- **Slimes** (`{{TRUTH}}<zone>`) — one per vertical slice: the files that change together (e.g. a
  command, its service, its model and its tests). Each slime reports to exactly one orc (its truth
  link); its territory is those files, precisely.
- **A kijin** (`{{LEAD}}<subsystem>`) — only for a standing subsystem someone must own across
  sessions (CI and releases, deploy, infrastructure). Reports to rimuru.
- **Minds** (skills) — know-how worth wearing: a contract or convention two or more creatures share
  (zone minds, worn by slimes), a review checklist an orc applies (verdict minds), a pipeline or
  cross-domain practice (global minds). Three to six is usual. A mind is instructions, concrete and
  specific to this code — paths, names, commands — never generic advice.
- **Bodies** (agents) — mint the elf, each orc (court) and each kijin (keeper, `mode: all`). Slimes
  run as court when dispatched; mint one only if it will run often on its own.

Before writing, show the plan as a table: creature · rank · reports to · territory · minds.

## 3. Write it, in exactly these shapes

A creature doc, at `{{WORLD}}/<rank dir>/<short>/README.md` (dirs: elf `{{VOICE_DIR}}`, orc
`{{GATE_DIR}}`, slime `{{TRUTH_DIR}}`, kijin `{{LEAD_DIR}}`); its name is the prefix plus `<short>`:

```
# {{TRUTH}}<short>

- **Rank:** {{TRUTH_RANK}}
- **Territory:** `path/one/`, `path/file.go`
- **Reports to:** {{GATE}}<domain>
- **Minds:** <mind>, <mind>
- **Purpose:** the ground truth of … (one sentence, what it owns and why it matters)

## Traits
- durable facts about this territory a newcomer would get wrong (from the code, with names)

## Verify
- `<a command that proves this territory — run it first; keep it only if it passes now>`

## Thoughts
```

A mind, at `.claude/skills/<name>/SKILL.md`: front matter `name` and `description` (one line: what
it knows and when to wear it), then the know-how as terse bullets, then an empty `## Thoughts`.

A body, at `.claude/agents/<creature>.md`: front matter `name` (the creature's name), `description`,
and `mode: all` for a keeper; the body text says whom it is, which doc to read first, which minds to
wear, and what waits for the human.

Rules that make the colony true:
- Territories do not overlap between slimes; an orc's territory is its slimes' union (or omitted).
- Every verify line is run before it is written. A command that fails today is not a verify line:
  write the failure as a trait and as a finding in the log instead.
- Hazards found in reading go into the traits of the creature that owns them.

## 4. Prove it

- Run `{{BIN}} onto check` and fix every finding it reports, until it passes.
- Append one entry to `{{WORLD}}/log.md`: what you founded, and the findings from the read.
- Answer with the table of the team and the findings, briefly.

A first sketch from the file tree alone (no reading, no judgement — a hint to improve on, not a plan):

{{SKETCH}}
