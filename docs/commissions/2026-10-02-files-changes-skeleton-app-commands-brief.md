# Brief — file-changes Go skeleton (app-commands slice)

**From:** elf-isekai (relaying rimuru commission, file-brief pattern — canonical copy in commissioner land; a working copy stands at `.isekai/slime/app-commands/thoughts/2026-10-02-files-wiring-brief.md`)
**Date:** 2026-10-02 · **World root:** `/Users/j00000/Workspace/agentic/operation-room-litellm/isekai` — sandbox forbids `cd`; run everything from here.
**Your land:** `app/board.go` ONLY (probe-verified via onto: slime-app-commands owns it).

## Hard precondition (do NOT edit until true)
`board.FileChange` must already exist and `./board` must build: world-core adds `type FileChange`, `Sources.Files func() []FileChange`, and a `Files:` entry in `FileSources` in the same chain (brief `docs/commissions/2026-10-02-files-changes-skeleton-world-core-brief.md`). Check: `grep -n "FileChange" board/sources.go` — if absent, report `@S blocked` immediately, edit nothing.

## Pairing law (Vitality — gate-enforced)
Your `app/` write must move `.isekai/slime/app-commands/README.md` in the SAME turn. Add a dated line noting the `Files:` wiring (honest-nil recorder placeholder).

## EDIT — `app/board.go`, in the `src := board.Sources{` literal (:42-63)
Immediately after the `Off: a.offList,` line (:45), insert ONE line:
```go
		Files:  func() []board.FileChange { return nil },
```
(`Files` then aligns with `Court:`, `Config:`, `Off:` above it; the panel self-feeds from session SSE and the `/api/files` hydrate legitimately returns empty until a real recorder lands. `Off:`/`Files:` share the 4-rune width — gofmt keeps the existing wider alignment untouched.)

Then `gofmt -w app/board.go`.

## Verify (raw outputs to elf-isekai):
`GOCACHE=$PWD/.isekai/tmp/gocache go build ./app/ && GOCACHE=$PWD/.isekai/tmp/gocache go vet ./app/` — must be clean.

## Touch nothing else: no `board/**`, no `app/web.go`, no Makefile/package.json. New symbols only — if you believe a consumer edit is needed, report it as `@?`, do NOT do it.

## Report back to elf-isekai on the wire: applied hunk as `@F app/board.go:N`, build+vet raw output, `@?` holes.

