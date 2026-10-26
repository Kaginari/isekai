# Benchmark results

Generated from `runs.jsonl` by `record.py` — do not edit.

## Head to head

### Run `tb2-gpt-oss-20b`

tb2 slice (10 tasks), isekai v1.1.0, cheap OpenRouter model  

Each cell: pass rate (trials) · errors · input/output tokens · cost $. Tokens and cost show only when every launch in the cell reported them.

| model | isekai |
|---|---|
| `openai/gpt-oss-20b` | **0.00** (10) · 6 err · 231.3k/50.6k · 0.0000 |

### Run `tb2-free-nemotron`

tb2 slice (10 tasks), isekai v1.1.0, free OpenRouter model  

Each cell: pass rate (trials) · errors · input/output tokens · cost $. Tokens and cost show only when every launch in the cell reported them.

| model | isekai |
|---|---|
| `nvidia/nemotron-3-super-120b-a12b:free` | **0.00** (10) · 10 err · 0/0 · 0.0000 |

## Every launch

| when | run | agent | model | dataset | trials | mean reward | errors | input tok | output tok | cost $ | commit | purpose |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 2026-10-26 00:38 | tb2-gpt-oss-20b | isekai | gateway/openai/gpt-oss-20b | terminal-bench | 10 | 0.0000 | 6 | 231304 | 50607 | 0.0000 | e6f3578+dirty | tb2 slice (10 tasks), isekai v1.1.0, cheap OpenRouter model |
| 2026-10-26 00:20 | tb2-free-nemotron | isekai | gateway/nvidia/nemotron-3-super-120b-a12b:free | terminal-bench | 10 | 0.0000 | 10 | 0 | 0 | 0.0000 | e6f3578+dirty | tb2 slice (10 tasks), isekai v1.1.0, free OpenRouter model |
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
