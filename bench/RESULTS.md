# Benchmark results

Generated from `runs.jsonl` by `record.py` — do not edit.

## Head to head

### Run `tb2-deepseek-v4.1-flash`

tb2 slice (10 tasks), isekai v1.1.0+fixes, deepseek-v4.1-flash  

Each cell: pass rate (trials) · errors · input/output tokens · cost $. Tokens and cost show only when every launch in the cell reported them.

| model | isekai |
|---|---|
| `deepseek/deepseek-v4.1-flash` | **0.20** (10) · 8 err · 11558.9k/665.3k · — |

## Every launch

| when | run | agent | model | dataset | trials | mean reward | errors | input tok | output tok | cost $ | commit | purpose |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 2026-10-26 02:43 | tb2-deepseek-v4.1-flash | isekai | gateway/deepseek/deepseek-v4.1-flash | terminal-bench | 10 | 0.2000 | 8 | 11558862 | 665336 | — | 565d34c | tb2 slice (10 tasks), isekai v1.1.0+fixes, deepseek-v4.1-flash |
| 2026-10-25 22:22 | — | isekai | vllm/fake/vllm-coder | adhoc | 1 | 0.0000 | 0 | 2700 | 80 | — | fb48136+dirty | smoke: fake vLLM (wrong) |
| 2026-10-25 22:21 | — | isekai | vllm/fake/vllm-coder | adhoc | 1 | 1.0000 | 0 | 2700 | 80 | — | fb48136+dirty | smoke: fake vLLM (right) |
| 2026-09-26 12:05 | — | isekai | vllm/fake/vllm-coder | adhoc | 1 | 0.0000 | 0 | 2700 | 80 | — | 443ff26+dirty | smoke: fake vLLM (wrong) |
| 2026-09-26 12:04 | — | isekai | vllm/fake/vllm-coder | adhoc | 1 | 1.0000 | 0 | 2700 | 80 | — | 443ff26+dirty | smoke: fake vLLM (right) |
| 2026-09-26 12:02 | — | isekai | vllm/fake/vllm-coder | adhoc | 1 | 0.0000 | 0 | 2700 | 80 | — | 443ff26+dirty | smoke: fake vLLM (wrong) |
| 2026-09-26 12:02 | — | isekai | vllm/fake/vllm-coder | adhoc | 1 | 1.0000 | 0 | 2700 | 80 | — | 443ff26+dirty | smoke: fake vLLM (right) |
| 2026-09-26 11:59 | — | isekai | vllm/fake/vllm-coder | adhoc | 1 | 0.0000 | 0 | 2700 | 80 | — | 443ff26+dirty | smoke: fake vLLM (wrong) |
| 2026-09-26 11:58 | — | isekai | vllm/fake/vllm-coder | adhoc | 1 | 1.0000 | 0 | 2700 | 80 | — | 443ff26+dirty | smoke: fake vLLM (right) |
| 2026-09-26 11:30 | — | isekai | vllm/fake/vllm-coder | fake/vllm-coder__adhoc | 1 | 0.0000 | 0 | 2700 | 80 | — | a815eb6+dirty | smoke: fake vLLM (wrong) |
| 2026-09-26 11:30 | — | isekai | vllm/fake/vllm-coder | fake/vllm-coder__adhoc | 1 | 1.0000 | 0 | 2700 | 80 | — | a815eb6+dirty | smoke: fake vLLM (right) |
| 2026-09-26 10:43 | — | nop | — | local:tasks/hello-file | 1 | 0.0000 | 0 | — | — | — | e6ddce8 | task validation (reference solution must pass, nop must fail) |
| 2026-09-26 10:43 | — | oracle | — | local:tasks/hello-file | 1 | 1.0000 | 0 | — | — | — | e6ddce8 | task validation (reference solution must pass, nop must fail) |
