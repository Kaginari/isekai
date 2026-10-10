# Brief — file-changes Go skeleton (world-core slice)

**From:** elf-isekai (relaying rimuru commission, file-brief pattern — canonical copy in commissioner land; a working copy stands at `.isekai/slime/world-core/thoughts/2026-10-02-files-changes-skeleton-brief.md`)
**Date:** 2026-10-02 · **World root:** `/Users/j00000/Workspace/agentic/operation-room-litellm/isekai` — sandbox forbids `cd`; run everything from here.
**Your land:** `board/tree.go` + `board/sources.go` (probe-verified via onto: slime-world-core owns both; edit them DIRECTLY).

## Pairing law (Vitality — gate-enforced)
Every `board/` write batch must move `.isekai/slime/world-core/README.md` in the SAME turn (your doc). Add a dated line noting the new `FileChange` type / `Sources.Files` feed / `TreeEntry.Kind` field.

## EDIT 1 — `board/tree.go`, TreeEntry struct at :12-17 (fields Path/Name/Dir/Depth). Append one field:
```go
	Kind string `json:"kind,omitempty"` // "write" | "delete"; empty = unchanged
```

## EDIT 2 — `board/sources.go`
(a) Insert the type ABOVE the `// Sources is every feed...` comment block (:295), i.e. between :293 `}` and :295:
```go
// FileChange is one path the session touched, newest-last; Kind ∈ {"write","delete"}.
// Adds/Dels are diff tallies when known (else 0); At is when the touch happened.
type FileChange struct {
	Path string    `json:"path"`
	Kind string    `json:"kind"`
	Adds int       `json:"adds,omitempty"`
	Dels int       `json:"dels,omitempty"`
	At   time.Time `json:"at"`
}

```
(`time` is already imported and used in this file.)

(b) In `type Sources struct` (:297-314), immediately after the `Tree` field line (:313), add:
```go
	// Files is the activity feed: one FileChange per path the session touched (write/delete),
	// for the dashboard's writes-only panel. nil means unwired; consumers must nil-guard.
	Files func() []FileChange
```

(c) In `FileSources`' returned `Sources{...}` literal, immediately after `Tree: func() ([]TreeEntry, bool) { return WalkTree(f.root, treeCap) },` (:358), add:
```go
		Files:   func() []FileChange { return nil }, // no activity recorder yet — panel hydrates empty
```

(d) In `withDefaults` (:362-...), mirroring the `Tree` block at :402-404, add:
```go
	if s.Files == nil {
		s.Files = d.Files
	}
```

(e) `gofmt -w board/tree.go board/sources.go`.

## Verify (raw outputs to elf-isekai):
`GOCACHE=$PWD/.isekai/tmp/gocache go build ./board/ && GOCACHE=$PWD/.isekai/tmp/gocache go vet ./board/` — must be clean.

## Context hole (settled ahead — do not re-guess)
The commission premise said board.go/events.go "already use" these symbols; they do NOT on disk (build is green today). These skeleton edits are forward-schema: they must keep the build green, change no existing behaviour, and the `kind,omitempty` keeps the old `/api/tree` wire byte-identical. Slice A (tui-board, the TS panel) is a separate in-flight job — do not touch `board/web/`.

## Report back to elf-isekai on the wire: applied hunks as `@F <file:line>`, build+vet raw output, `@?` holes.

