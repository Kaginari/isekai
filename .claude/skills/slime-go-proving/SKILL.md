---
name: slime-go-proving
description: How this Go module proves itself — the exact build, vet, format and test commands per package, what CI adds (race, bwrap, selftest, clean tree), and the pieces the root ./... does not cover. Wear before running or writing any verify line in the isekai repository.
---

- Toolchain: go 1.26.8 in go.mod; on this machine `export PATH=$HOME/.local/go-current/bin:$PATH` (go1.27.1). Build the binary with `go build -o bin/isekai ./cmd/isekai`.
- Prove a territory with its package: `go test ./<pkg>/...` (e.g. `go test ./loop/... ./instrument/...`). Whole module: `go build ./... && go vet ./...` then `go test ./...`. All of it passes today and is what the creature docs' Verify lines run.
- CI (.github/workflows/ci.yml) is stricter than local: `gofmt -l .` must print nothing (`test -z "$(gofmt -l .)"` is the exit-code form), `go vet ./...`, `go test -race ./...`, `go run ./cmd/isekai selftest`, and `git status --porcelain` must be empty after the tests. Run gofmt and vet before calling work done; run `-race` on shell/, loop/, board/ when touching goroutines.
- The sandbox tests need bubblewrap (/usr/bin/bwrap here; on Ubuntu 24.04 also `sysctl -w kernel.apparmor_restrict_unprivileged_userns=0`). Without it sandbox falls to Mode None and TestBwrapConfinement / TestSandboxedNetworkOneShot skip or fail — read the reason, do not route around.
- Not covered by the root `./...`: bench/fakevllm (its own go.mod — `cd bench/fakevllm && go build -o /dev/null .`), bench/smoke.sh (needs docker and the Harbor image), `goreleaser check` (not installed locally, CI only), screenshots (`isekai ui check`, needs Chromium — UI_CHROMIUM), TUI shots (TUI_SHOTS=<dir> go test ./tui/ -run TestShots).
- Goldens: `go test ./tui/ -goldens` rewrites tui/testdata/golden; look at the new files before committing them.
- `bin/isekai selftest` prints `@S PASS n checks` (195 today, < 1 s); `bin/isekai onto check` must report 0 findings after any change to creature docs or skills.
- A verify line goes in a creature doc only after it ran and passed in this session; a command that fails today is written as a trait and a log finding instead.

## Thoughts
