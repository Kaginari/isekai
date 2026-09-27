# slime-config

- **Rank:** Slime
- **Territory:** `config/`, `examples/`
- **Reports to:** orc-config
- **Minds:** slime-go-proving
- **Purpose:** the ground truth of the switchboard: layers and precedence, the strict YAML/JSON readers, the schema with an `enabled` on every feature, origins, refusals, permission rules, models and providers, ranks as data, rules and checks, patching — and the example configs.

## Traits
- Seven layers, lowest first: built-in defaults (defaults.go is a JSON text with `[dist]` / `[dist-dir]` / `~/` substitution), global ~/.config/<dist>/config.{yaml,yml,json}, project <dist-dir>/config.*, project local config.local.* (gitignored), env whole file (<PREFIX>CONFIG, <PREFIX>CONFIG_CONTENT), env keys (<PREFIX>MODEL, SMALL_MODEL, MODE, LOG_LEVEL, PROVIDER_<NAME>_API_KEY, DISABLE_PROJECT_CONFIG), CLI flags (--model, --approve, --dry-run, --strict, --set k=v, --no-<feature>). Beside config.yaml a file named after a top-level section (models.yaml, providers.yaml, rules.yaml, tools.yaml, mcp.yaml, guard.yaml) loads in the same layer and wins there; two files for one section is a load error.
- config/yaml is a strict YAML subset and JSON with comments into one line-numbered Node tree; anything outside the subset is an error with file:line; unknown keys are errors; the decoder is reflection over struct tags. Origins records where every value came from (`config explain`); Patch/ApplyPatch edit a YAML file preserving layout and return a diff; Reload yields Changes.
- Permissions: rules `<tool>:<glob>` with allow/ask/deny; Decide picks the most specific rule (Specificity, actionRank deny > ask > allow), ClassDefault is the floor; checkPermissions refuses `allow` with a wildcard-only pattern on OutwardCapable tools or an outward/destructive floor; Loosenings lists every silenced question; `question` is an alias of `ask`.
- checkGate: humanGate.enabled false is honoured only from a file layer; an approve entry from env is refused. checkCompaction: trigger.tokens must be below law.budget.contextTokens; stressTokens ≤ contextTokens.
- Ranks as data: rankSet extend|replace over builtinRanks; `rimuru` may not be configured; an ascended rank (ascendsFrom) inherits every field it did not set and may not drop authors, holdsGate or sideways; every rank has reportsTo; no cycles; an authoring rank needs a gate above it. Vocabulary folding: analyst/judge/drafter → great-sage/raphael/ciel; coord/domain/zone/service/auditor → elf/orc/slime/kijin/dark-elf.
- Models: Offices great-sage, raphael, ciel; Tasks drain, gate, log, bench; ResolveModel(creature, race, office, task) with per-slot fallbacks and origins, Mount is the session's model; Price.Cost prices tokens; providers default table has anthropic and openai on, openrouter and ollama off; a non-mock provider needs baseURL; keys come from apiKeyEnv only.
- rules.go: injected rules with scopes all | rank:<race> | creature:<name>, PromptRules for the prompt, LawFacts for the ontology, Checks for the gate (RunCheck via the shell). tools.go: builtin switches and classes, custom tools with {{param}} argv/shell templates, MCP server definitions (checkMCP) and MCPClass.
- Selftest (config.SelftestIn) is the instrument that says whether docs/canon/config.md and the code still agree. examples/gateway is the vLLM gateway example the README cites.

## Verify
- `go test ./config/...`

## Thoughts
