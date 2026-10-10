---
name: kijin-release-pipeline
description: The isekai release pipeline end to end — what CI runs, how a version is cut, what GoReleaser publishes (archives, checksums, GHCR image, provenance), what install.sh expects, and which steps are irreversible. Wear when touching .github/, .goreleaser.yaml, Dockerfile.release, install.sh, RELEASING.md or CHANGELOG.md, or when a release is asked for.
---

- CI (ci.yml): test → gofmt clean, `go vet ./...`, bubblewrap + `sysctl -w kernel.apparmor_restrict_unprivileged_userns=0`, `go test -race ./...`, `go run ./cmd/isekai selftest`, clean tree; build matrix linux/darwin × amd64/arm64 (CGO_ENABLED=0 -trimpath); goreleaser check; bench-smoke (bin/isekai + bin/fakevllm, Harbor image, bench/smoke.sh right → 1.0 / wrong → 0.0). Reproduce locally with the same commands; goreleaser is CI-only here.
- Cutting a version (RELEASING.md): main green → CHANGELOG.md: move `## [Unreleased]` notes under `## [X.Y.Z] - YYYY-MM-DD`, Highlights paragraph first, then Added / Changed / Fixed / Removed, and keep an empty `## [Unreleased]` above → commit `chore: release vX.Y.Z` → `git tag -a vX.Y.Z -m "vX.Y.Z" && git push origin main vX.Y.Z`. SemVer; before 1.0 a minor may break config/CLI; pre-releases `v0.3.0-rc.1`.
- release.yml on tag v*: `go test ./...`; awk extracts the tag's CHANGELOG section as the release header and FAILS if `## [X.Y.Z]` is absent — write the section before tagging; buildx + ghcr login (GITHUB_TOKEN); `goreleaser release --clean --release-header`; attest-build-provenance (skipped for private repos). Verify later with `gh attestation verify <archive> --repo Kaginari/isekai`.
- GoReleaser publishes: isekai_<os>_<arch>.tar.gz (no version in the name — install.sh relies on releases/latest/download), checksums.txt, LICENSE/README/CHANGELOG inside, a GitHub changelog grouped feat/fix (docs:, test:, chore(deps): excluded — use Conventional Commits), image ghcr.io/kaginari/isekai:<version> and :latest (only for non-pre-releases), linux/amd64 + linux/arm64 from Dockerfile.release (debian:bookworm-slim, COPY only, no RUN).
- install.sh: `curl -fsSL …/install.sh | sh`, VERSION=vX.Y.Z pins, BIN_DIR default ~/.local/bin, verifies sha256 against checksums.txt, then `isekai version`. Changing the archive name or checksum file breaks every installer already published.
- Irreversible acts (Veldora's word each time): pushing a tag, deleting or moving a tag or release, moving :latest, changing the GHCR image name, bumping a pinned action major. A dry run is `goreleaser release --snapshot --clean` in CI or on a machine that has goreleaser.
- Ldflags stamp main.version/commit/date; `isekai version` prints them — the smoke of a build is that line.

## Thoughts
