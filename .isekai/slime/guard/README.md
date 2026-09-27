# slime-guard

- **Rank:** Slime
- **Territory:** `guard/`
- **Reports to:** orc-safety
- **Minds:** slime-go-proving, slime-class-contract
- **Purpose:** the ground truth of the global dangerous-command guard: the denylist refused before the classifier and the human gate, its corpus, and its installation into Claude Code and OpenCode.

## Traits
- patterns.txt is embedded and always on: one POSIX ERE (RE2) pattern per line, `#` comments; it blocks only the catastrophic and irreversible — rm -rf of /, a top-level dir or home (by ~, $HOME or the path), any sudo rm, dd/mkfs/diskutil onto a device, raw disk redirects, the fork bomb, curl|sh, git push --force / --delete / +ref, reflog expire and gc --prune=now, chmod 777 /, chown -R /, GitHub repo and secret deletions, secret-store reads. Recoverable local commands stay allowed on purpose.
- Two more sources add, never remove: the machine-wide ~/.agents/hooks/dangerous-patterns.txt (SharedPath, read by other agents' hooks too) and the config's guard.files / guard.patterns. Load and ForHome build the Guard; Match returns the Rule with its source; Refusal renders the message the model reads.
- corpus.txt is the test: `block <cmd>` / `allow <cmd>` lines, `\n` unescaped; Guard.Test returns passed and wrong. The same file is read by the tool package's classifier test, so a new block line must also classify ≥ outward or that test fails.
- Hook(in, out, errw, cursor) is the stdin hook for Claude Code's PreToolUse; ClaudeSettings patches a settings JSON to add it; OpenCodePlugin(binary) emits the JS plugin. `isekai guard install --yes` writes both; without --yes it only shows the change.
- It guards against accidents, not a determined agent: `python -c "shutil.rmtree(...)"` slips past any regex. The sandbox and the gate remain the containment.

## Verify
- `go test ./guard/...`

## Thoughts
