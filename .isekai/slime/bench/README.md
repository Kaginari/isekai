# slime-bench

- **Rank:** Slime
- **Territory:** `bench/`
- **Reports to:** orc-app
- **Minds:** slime-go-proving, bench-run
- **Purpose:** the ground truth of the benchmarks: the Harbor runner and its adapters, the fake vLLM, the local tasks and the Terminal-Bench slice's image recipe, the €0 smoke, the head-to-head launcher, and the append-only record of every launch.

## Traits
- `bench/fakevllm` is its own Go module (`module fakevllm`, go 1.23): the root's `go test ./...` and `go vet ./...` never see it; CI builds it with `cd bench/fakevllm && go build -o ../../bin/fakevllm .`. It serves GET /v1/models (max_model_len 32768) and POST /v1/chat/completions with tool_calls and usage, playing a fixed script; FAKE_MODE right|wrong (anything else is fatal) picks the answer, FAKE_FILE the file written (default /app/hello.txt), -addr the listen address.
- `bench/harbor/harbor_agent.py` (HarborAgent) reads BENCH_BIN, BENCH_DIST, BENCH_CONFIG (written into the world as its local config) and BENCH_FORWARD_ENV (key NAMES; values only from the environment, never on disk); it installs the static binary at /usr/local/bin/<dist>, seeds a world at /app, runs `<dist> run --json`, copies the usage and loop journals to /logs/agent, and reports tokens and cost from the usage journal — cost is None when any call is unpriced, never a guess.
- `bench/harbor/harbor_opencode.py` (HarborOpenCode) is Harbor's OpenCode minus the internet install: task images prebake node and npm, the npm token rides by name (NPM_AUTH_TOKEN). The task images reach packages through the Artifactory remotes only.
- `bench/smoke.sh` is the €0 smoke CI runs: it needs docker, bin/<dist>, bin/fakevllm and the `<dist>-harbor` runner image; DIST defaults to `ls cmd`; the fake listens on the bridge gateway, port 18000; each mode is recorded by record.py with `--expect 1.0` (right) / `0.0` (wrong) — a wrong score fails the script. The runner mounts the repository at its own absolute path so Harbor's paths resolve on the host's Docker daemon.
- `bench/harbor.sh` is the real run: isekai and OpenCode on the same tasks, gateway models and attempts, one Harbor job per (agent, model, task set) under jobs/<run>/; flags win over BENCH_* env; the key travels by variable name (`--key-env`, default GATEWAY_API_KEY); `--dry-run` prints every command and runs none; `--run <name>` resumes unfinished jobs.
- `bench/runs.jsonl` is append-only (one line per launch); `bench/RESULTS.md` is generated from it by record.py (agent × model tables for the most recent runs) — never hand-edited. jobs/ is Harbor's raw output and untracked. configs/fake-vllm.yaml is the exact smoke config; configs/gateway.yaml the template a real run renders.
- `bench/tasks/hello-file` is the plumbing task. `bench/tb2-base/build.sh` bakes the Terminal-Bench 2 slice images (prebuilts wrapped with node, opencode and each task's verifier pins, secrets via temp file / BuildKit secret, never layered); the `tasks-tb2/` task tree it tags images for is not in this repository.
- A run on a real provider costs money: only the fake vLLM runs in CI, and a real benchmark waits for Veldora's word.

## Verify
- `bash -n bench/smoke.sh && bash -n bench/harbor.sh && bash -n bench/tb2-base/build.sh && python3 -m py_compile bench/record.py bench/harbor/harbor_agent.py bench/harbor/harbor_opencode.py && (cd bench/fakevllm && go build -o /dev/null .)`

## Thoughts
- (none yet)
