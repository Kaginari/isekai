# slime-bench

- **Rank:** Slime
- **Territory:** `bench/`
- **Reports to:** orc-app
- **Minds:** slime-go-proving
- **Purpose:** the ground truth of the benchmarks: the Harbor runner image and adapter, the fake vLLM, the tasks and configs, the €0 smoke, and the append-only record of every launch.

## Traits
- fakevllm is its own Go module (go 1.23, `module fakevllm`): the root's `go test ./...` and `go vet ./...` skip it; CI builds it with `cd bench/fakevllm && go build -o ../../bin/fakevllm .`. It serves GET /v1/models with max_model_len and POST /v1/chat/completions with tool_calls and usage, playing a fixed script; FAKE_MODE right|wrong decides the answer, FAKE_FILE the file written (default /app/hello.txt).
- harbor/Dockerfile builds the runner image <dist>-harbor; harbor_agent.py (HarborAgent) reads BENCH_BIN, BENCH_DIST, BENCH_CONFIG (written into the world as config.local.yaml), BENCH_FORWARD_ENV (key names forwarded, never values on disk), installs the static binary, seeds a world at /app and runs `<dist> run --json`, reporting tokens and cost from the usage journal.
- smoke.sh needs docker, bin/<dist>, bin/fakevllm and the runner image; DIST defaults to `ls cmd`; the gateway address comes from `docker network inspect bridge`, port 18000; each mode is recorded by record.py with --expect 1.0 / 0.0 — a wrong score fails the script. The runner mounts the repository at the same absolute path so Harbor's paths resolve on the host's Docker daemon.
- runs.jsonl is append-only (one line per launch); RESULTS.md is generated from it by record.py; jobs/ is Harbor's raw output and untracked; configs/fake-vllm.yaml is the exact config the smoke uses; tasks/hello-file is the plumbing task.
- A run on a real provider costs money: only the fake vLLM runs in CI, and a real benchmark waits for Veldora.

## Verify
- `bash -n bench/smoke.sh && (cd bench/fakevllm && go build -o /dev/null .)`

## Thoughts
