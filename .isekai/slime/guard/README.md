# slime-guard

- **Rank:** Slime
- **Territory:** `guard/`
- **Reports to:** orc-safety
- **Minds:** slime-go-proving, gate-safety, slime-class-contract
- **Purpose:** the ground truth of the global dangerous-command guard: the denylist refused before the classifier and the human gate, its command-word reading, its corpus, and its installation into other agents' hooks.

## Traits
- patterns.txt is embedded and always on: one POSIX ERE (RE2) pattern per line, `#` comments; it blocks only the catastrophic and irreversible — rm -rf of /, a top-level dir or home, any sudo rm, dd/mkfs/diskutil onto a device, raw disk redirects, the fork bomb, curl|sh, git push --force / --delete / :ref / +ref, reflog expire and gc --prune=now, chmod 777 /, chown -R … /, GitHub repo and secret deletions, secret-store reads. Recoverable local commands stay allowed on purpose.
- Sources add, never remove: Default() is the built-in list; Load adds files (a missing file is no finding, an unreadable one or a bad pattern is a Note, never a crash); ForHome adds the machine-wide ~/.agents/hooks/dangerous-patterns.txt (SharedPath) and a rule for the home by its own path; AddPatterns takes config's guard.patterns. Match returns the first Rule with its Source; Refusal renders the message the model reads.
- Match tests each rule against the raw command AND against CommandWords(cmd): command words at a command position (start, after ; & | ( ` $( or a quote) lose their path or alias escape — `/usr/bin/git push --force`, `\rm -rf ~` — and wrappers are seen through (xargs, exec, command, nohup, nice, env, time, sudo, timeout, with their flags and assignments). Arguments stay: `cat /usr/bin/rm` is allowed. tool.ClassifyCommand reads commands through the same CommandWords.
- corpus.txt is the test: `block <cmd>` / `allow <cmd>` lines (155 cases), `\n` unescaped by Cases(); Guard.Test returns passed and wrong. The same file is read by tool's TestGuardCorpusNeedsTheHuman, so a new block line must also classify ≥ outward without the guard.
- Hook(in, out, errw, cursor) answers a pre-tool hook from stdin (.tool_input.command, .toolInput.command or .command): blocked is exit 2 with the reason on stderr, or a deny JSON on stdout under cursor. ClaudeSettings patches a settings JSON to add it; OpenCodePlugin(binary) emits the JS plugin. `isekai guard install --yes` writes both; without --yes it only shows the change.
- It guards against accidents, not a determined agent: `python -c "shutil.rmtree(...)"` slips past any regex. The sandbox and the gate remain the containment.

## Verify
- `go test -count=1 ./guard/...`

## Thoughts
- (none yet)
