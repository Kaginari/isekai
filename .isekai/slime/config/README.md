# slime-config

- **Rank:** Slime
- **Territory:** `config/`, `examples/`
- **Reports to:** orc-config
- **Minds:** slime-go-proving, gate-safety
- **Purpose:** the ground truth of the switchboard: layers and precedence, the strict YAML/JSON readers, the schema with an `enabled` on every feature, origins, refusals, permission rules, models and providers, ranks as data, rules and checks, patching, the local scaffold — and the example configs.

## Traits
- Seven layers, lowest first: built-in defaults (defaults.go, a JSON text with [dist], [dist-dir] and ~/ substitution), global ~/.config/<dist>/config.*, project <dist-dir>/config.* (tracked), local config.local.* (machine-only), env whole file (<PREFIX>CONFIG, <PREFIX>CONFIG_CONTENT), env keys (MODEL, SMALL_MODEL, MODE, LOG_LEVEL, PROVIDER_<NAME>_API_KEY, DISABLE_PROJECT_CONFIG), CLI flags (--model, --approve, --dry-run, --strict, --set k=v, --no-<feature>).
- Section files (load.go): beside config.yaml, a file named after any top-level default section (models.yaml, guards.yaml for guard, …) loads in the same layer and wins there; `<section>.local.yaml|yml|json` does the same in the local layer (localSections), after config.local.*; a name that is no section is ignored; two files for one section in one layer is a load error.
- `ui.theme` and `ui.mode` (theme.go): checkTheme lower-cases them, defaults cyberpunk and dark, refuses an unknown palette naming theme.Names() and a mode outside dark | light | system; ExplainTheme is `config explain theme`.
- `config scaffold` (scaffold.go) works on paths, not a loaded config: it refuses a non-world (exit 2), mints config.local.yaml plus one `<section>.local.yaml` per curated section this binary knows, never overwrites an existing file, and adds `*.local.yaml` to the world dir's .gitignore. `--check` writes nothing: unknown local sections are `@?`, sections with no local file `@F`; exit 1 on drift.
- Refusals from the tracked project layer (validate.go, permissions.go): law.humanGate.enabled false is honoured only from the global or local file; an approve entry from env is refused, and from the project layer for outward or destructive; in Decide a project-layer `allow` never pre-approves a tool whose class default is ask. `allow` with a wildcard-only pattern on an outward-capable tool is refused from any layer.
- Credentials: providers.<p>.apiKey is the one literal key field (ProviderKey prefers the apiKeyEnv variable, then the literal); a credential anywhere else — other key/token/secret/password keys under providers. or mcp., authorization/x-api-key/api-key headers, sk-/bearer values — is a load error naming apiKeyEnv or headersEnv. checkCompaction: trigger.tokens below law.budget.contextTokens, stressTokens ≤ contextTokens.
- The pin law (models.go): ResolveModel goes creature → office → task → default and stops at the first pin; a creature pin forces over everything. Ranks are not on the ladder — models.ranks.* and ranks.<r>.model parse but only raise an "ignored, move it to models.offices" hole; ModelTable prints a rank as its Office's resolution. Offices great-sage, raphael, ciel; Tasks drain, gate, log, bench; Mount is the session's model.
- Ranks as data (ranks.go): rankSet extend|replace over builtinRanks; `rimuru` may not be configured; an ascended rank (ascendsFrom) inherits every field it did not set (layer included) except dir and prefix, which default to its own name, and may not drop authors, holdsGate or sideways; ascendsWhen must read `ascend <base> when …` naming its own base (high-elf and high-orc carry DefaultAscendElf/Orc); divineRules and ascendsWhen belong to ascended ranks only and are never inherited. `body` is no field of the resolved rank: a set ranks.<r>.body is a hole. tools.dispatch.lineage (default true) keeps dispatch within a creature's line. Vocabulary folding: analyst/judge/drafter → great-sage/raphael/ciel; coord/domain/zone/service/auditor → elf/orc/slime/kijin/dark-elf.
- Off() names law.forge.demandAtSteps when a layer sets it to 0 (the forge question switched off); tools.custom.<name>.crafted {by, run, at, draft} is a minted tool's provenance (the forge writes it).
- rules.go: injected rules with scopes all | rank:<race> | creature:<name>; RunCheck runs a check under sandbox.Contain. examples/gateway is the vLLM gateway world README's Configuration section cites — it must pass `config check`; examples/world-rules.yaml is the annotated config.local.yaml example. config.SelftestIn is the instrument that says whether docs/canon/config.md and the code still agree.
- law.guts.repin (3) and law.guts.actBudget (40) arm the loop's stall guards; 0 takes the default, negative = off; repin 1 is refused (the first read repeats nothing).

## Verify
- `go test -count=1 -run 'TestSectionFiles|TestLocalSectionFiles|TestPrecedence|TestModelsResolution|TestRanks|TestPermissionRefusals|TestProjectAllowNeverPreApprovesOutward|TestProviderKeyInlineFallback|TestSelftest' ./config/`
- `go run ./cmd/isekai config check --root examples/gateway`

## Thoughts
- (none yet)
