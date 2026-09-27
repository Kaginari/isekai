---
name: slime-class-contract
description: The four act classes and the fixed order of checks a tool step passes in the isekai binary — declared floor, classifier, guard, territory policy, permission rule, human gate — plus the record files and the guard corpus every change must keep true. Wear when touching tool/, guard/, gate/, sandbox/, shell/, mcp/adapter.go or loop.Session.step.
---

- Classes are ordered ints in tool: read (0) < write (1) < outward (2) < destructive (3). A Tool declares a floor (Tool.Class); Tool.Classify may return a higher class, never a lower one (Settle takes tool.Max). "A declared class only tightens" is enforced in code, not prose.
- Step order in loop.Session.step — do not reorder: step/wall-clock budgets → missing tool (Hooks.Missing, `_malformed`) → Settle → Hooks.Guard (refused, by guard) → env.Policy territory (refused, by policy) → Hooks.Decide permission rule (deny refuses; allow silences the gate; ask forces it) → gate.Ask → dry-run (would-ask / would-run) → Hooks.PreTool → Run → Hooks.PostTool → verify → record.
- Gate: Needs(c) is c ≥ outward, or write under --strict. Answer order: not needed → pre-approved by class → dry-run → TUI choice → TTY yes/no → no TTY = denied. Nothing approves on its own; a permission rule `allow` on an outward-capable tool must be a pattern, never `*` (config.checkPermissions).
- Guard before everything: guard/patterns.txt (embedded, always on) + ~/.agents/hooks/dangerous-patterns.txt + guard.files; a match is refused with no gate at all. Corpus guard/corpus.txt: every `block` line must be refused by guard.Test AND classified ≥ outward by tool.Env.ClassifyCommand alone (tool/guard_corpus_test.go). Change a pattern or a rule → run `go test ./guard/... ./tool/...`.
- Classifier facts: per shell segment; unknown command = write; any http(s):// URL, network tools, package managers, cloud CLIs, secret stores, `go get`, a path outside the world, a `cd` outside = outward; rm/mv/dd/mkfs, git history rewrites, force pushes, `find -delete`, and any redirect/sed -i/tee/cp onto a record = destructive.
- Records: isekai.md, log.md, canon/*.md, notes.jsonl (tool.IsRecord and the classifier's `record` group). Overwriting one is destructive; log.md and notes.jsonl are appended only.
- Paths: every path tool passes tool.Env.Confine — no `..` climb, no absolute path outside the root, no symlink (of any existing prefix) that resolves outside.
- MCP tools are outward by default (mcp.Classify), write only when the server is marked inward in config; destructiveHint / openWorldHint tighten only. Custom tools floor at write; a shell template is classified like any command.
- Sandbox: bwrap read-only root, the world read-write, private /tmp, no network unless the gate approved the act as outward; env scrubbed by `(^|_)(API_KEY|TOKEN|SECRET|PASSWORD|PASSWD)$` plus bash.envDrop globs — names like AWS_SECRET_ACCESS_KEY need envDrop.

## Thoughts
