# slime-tools

- **Rank:** Slime
- **Territory:** `tool/`
- **Reports to:** orc-safety
- **Minds:** slime-go-proving, slime-class-contract
- **Purpose:** the ground truth of the tools a body may run — their shapes, the registry and shelf, the built-ins, the editor, patch, git, web, dispatch, ask, custom tools — and of the per-command classifier that decides what reaches the human.

## Traits
- Tool{Name, Description, Schema, Class, Classify, Run}; Settle(env, input) returns the effective Classification and holes; Def.DeclareFor(provider) hands anthropic the schema-less trained-tool JSON (AnthropicBash, AnthropicEditor) and everyone else the custom schema. Registry keeps order; Only(names…) cuts a rank's shelf; Builtins() is read, write, edit, bash, glob, grep.
- classify.go: per shell segment, destructiveRules (rm, shred, mv, truncate/dd/mkfs, git reset/clean/rebase/gc/prune/rm/mv, branch -d/-D/-M, push --force/-f/+, checkout -- / restore ., stash drop/pop, reflog expire, fork bomb, raw disk, tag -d, find -delete, and any redirect/sed -i/tee/cp onto a record) → outwardRules (git remote verbs, curl/wget/ssh/scp/rsync/nc…, npm/pnpm/yarn network verbs, npx, pip install, docker push/pull/login/build/run, gh/aws/gcloud/kubectl/terraform…, mail, any http(s):// URL, xdg-open, go get/install/mod download|tidy, secret stores pass/op/bw…, security keychain, gpg --export-secret, brew) → writeRules (any redirect, tee/cp/mkdir/touch/chmod/…, sed -i, git local writes, memory.js remember) → a command not on the readOnly allowlist is `write: unknown command`; a path outside the world (~, $HOME, absolute, ../ climbs) or a cd outside is outward. The strongest class of any segment wins.
- Confine(p) refuses `..` climbs, absolute paths elsewhere and symlinks of any existing prefix pointing out of the root; every path tool goes through it. IsRecord marks isekai.md, log.md, canon/*.md, notes.jsonl.
- editor.go is str_replace_based_edit_tool (view / create / str_replace / insert), declared as text_editor_20250728 under anthropic; writeAtomic writes then renames. patch.go parses unified diffs (ParseUnified, FilePatch.Apply with findHunk drift). git.go classifies by subcommand tables (gitReadSubs, gitWriteSubs, gitOutwardSubs, gitDestructiveSubs; a `+ref` push is destructive). web.go WebFetch is GET only, textual bodies only; WebSearch needs a Backend or is unavailable.
- dispatch.go: Commission{Body, Ask, Office, Cap, Unsaid…}, OfficeOf(ask) picks great-sage/raphael/ciel from the ask's verb; DispatchTool runs a Dispatcher the world supplies. ask.go: Question with options, TTYAsker on a terminal. missing.go: aliases (e.g. other harnesses' tool names) → Closest suggestions; DoomLoop stops a model that keeps calling the same missing tool. custom.go: exactly one of run (argv, {{param}} placeholders, elements never a shell) or shell (a template classified as a command); floor class write; name `^[a-z][a-z0-9_-]{0,63}$`.
- guard_corpus_test.go reads the guard slime's corpus — a `block` line there that this classifier rates below outward fails this package's tests.
- Output clipping: builtins clip a result (2000 lines / 50 KB) and point at the dump under the world's tmp/tool-out.

## Verify
- `go test ./tool/...`

## Thoughts
