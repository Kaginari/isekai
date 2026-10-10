# slime-providers

- **Rank:** Slime
- **Territory:** `provider/`
- **Reports to:** orc-engine
- **Minds:** slime-go-proving, slime-wire-and-journals
- **Purpose:** the ground truth of the one standing outward act: the Provider interface and its three clients — anthropic (Messages API), openai (chat-completions for OpenAI, OpenRouter, Ollama, vLLM, SGLang, Gemini-compatible endpoints) and mock.

## Traits
- anthropic: POST /v1/messages, anthropic-version 2023-06-01, DefaultModel claude-opus-5, defaultMax 16000 output tokens, FallbackBeta server-side-fallback-2026-07-01 sent when Fallbacks is on. It declares Claude's trained tools schema-less when a ToolDef names them (bash_20250124, text_editor_20250728), puts cache_control on the last stable system block, sends adaptive thinking and output_config.effort, and replays thinking blocks opaque on the next call as tool use requires. It has NO retry: a 429 or 5xx is returned as `anthropic: HTTP n: type: message` at once.
- openai: ToolCalls "text" makes tools tagged text — textToolIntro is appended to the system prompt, tagRe parses <tool_call> tags, a bad tag becomes a call named `_malformed`; Guided decoding constrains a Court's report to WireSchema and RenderWire turns the JSON back into @-lines; ContextWindow reads GET /models once (max_model_len, context_length or context_window) and errors if the model is not listed; doRetry retries 429/500/502/503/504 up to 3 times, honouring Retry-After ≤ 60 s, else 1 s, 2 s, 4 s; errorText reads OpenAI's and Gemini's error shapes. An empty APIKey is allowed (Ollama).
- Keys never cross the wire: provider.Getenv is a package var so tests inject an environment; the app passes the key read from the env var config.apiKeyEnv names. Config-declared Headers are set verbatim on every request — config refuses authorization, x-api-key, api-key and sk-/bearer values there, so this package trusts what it is given.
- Provider.Name() is anthropic/<model> or openai/<model>; the usage journal and the board key spend by that name.
- Usage.Context() = Input + CacheRead + CacheWrite is the number the instrument reads; Usage.Add sums a Court's spend into its parent.
- mock.Provider plays scripted Responses in order for tests and selftests; the app's bench uses its own scriptProvider (app/mockscript.go), not this one.
- Request.SystemText joins the system blocks; provider.Validate refuses a request with no messages or a tool result without its call.

## Verify
- `go test ./provider/...`

## Thoughts
