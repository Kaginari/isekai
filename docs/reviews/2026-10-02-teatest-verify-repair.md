# Repair — teatest module-cache entry dirtied (gate: elf-isekai verify)

Gate: `grep -q '^module ' go.mod && go mod verify` → exit 1
`github.com/charmbracelet/x/exp/teatest/v2 v2.0.0-20260927004216-9c77d672503d: dir has
been modified (~/go/pkg/mod/github.com/charmbracelet/x/exp/teatest/v2@v2.0.0-…)`

## What was found (all in the module cache; I cannot write there from inside this world)

The extracted module dir carries **one untracked file tree**, a nested paste that does not
belong to teatest:

```
…/teatest/v2@v2.0.0-20260927004216-9c77d672503d/powernap/pkg/lsp/protocol/LICENSE
```

(`powernap/pkg/lsp/protocol/` — an lsp `LICENSE` pasted under teatest, nothing to do with
the module.) `go mod verify` hashes the extracted tree against the recorded sum; the extra
`powernap/` subtree makes it "dir has been modified". Other files match their zip sums —

```json
Sum "h1:9aGB2Xfhn7jz/G2P2g0FQAd5FdtzPACWZUVue6IR3oY="
GoModSum "h1:aRoQwQWmN9LBG2xi3sVByMFt2fdkPCagd0GAJ1qwOfw="
Zip (pristine, in cache): ~/go/pkg/mod/cache/download/github.com/charmbracelet/x/exp/teatest/v2/@v/v2.0.0-20260927004216-9c77d672503d.zip
```

so the rest of the module is clean. Likely cause: an editor accidentally duplicating
sibling-package files into the read-only cache dir, then something giving them write
permission; not corruption.

## The fix (one go-toolchain pair — run outside the world)

The sandbox blocks my absolute-path writes into `~/go/pkg/mod/…` (the gate's
own containment — bwrap rewrites those into `.isekai/tmp/root/…`, so any `rm` of mine
would land inside world tmp, nowhere real; and my destructive class is gated to you).
Run this yourself:

```sh
VER=v2.0.0-20260927004216-9c77d672503d
M="~/go/pkg/mod/github.com/charmbracelet/x/exp/teatest/v2@${VER}"
# the extracted module dirs are read-only; grant write, drop it, re-extract from the cached zip
chmod -R u+w "$M" && rm -rf "$M"
go mod download github.com/charmbracelet/x/exp/teatest/v2   # re-extracts from the zip still in cache
cd <workspace>/isekai
go mod verify                                               # should print: all modules verified
```

Safer variant (leaves the cache's record of versions alone, just moves the bad dir aside
instead of deleting it):

```sh
VER=v2.0.0-20260927004216-9c77d672503d
M="~/go/pkg/mod/github.com/charmbracelet/x/exp/teatest/v2@${VER}"
chmod -R u+w "$M" && mv "$M" "$M.bak-$(date +%Y%m%d)"
go mod download github.com/charmbracelet/x/exp/teatest/v2
```

The zip at
`~/go/pkg/mod/cache/download/github.com/charmbracelet/x/exp/teatest/v2/@v/v2.0.0-20260927004216-9c77d672503d.zip`
is intact and hashed to the recorded sum — `go mod download` will re-extract from it; no
network needed, and no change to any `go.mod`/`go.sum`.

## What I did from inside the world

- Captured the exact offending file (path above) and the pristine-zip coordinates — the
  evidence needed to make this a *file-level* cleanup, not a `go clean -modcache`.
- Verified `go.mod` declares the module (line 16) and the recorded zip sum matches —
  nothing to change in the repo: **zero repo-file changes are needed**.
- Documented the repair here for the verdict trail (docs/reviews/2026-10-02-dash-redesign-verdict.md
  already landed; this note stands beside it for reviewers).

## Hole

I could not execute the rm + re-extract myself: the sandbox refuses absolute writes into
`~/go/pkg/mod/…`; my destructive class is human-gated in this world; and the
deleted dir is machine-global state, not territory. What is needed is the human's go
toolchain. The pre-repair state (dirtied dir) is preserved by the "move aside" variant if
you want to diff it.

