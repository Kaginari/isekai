# slime-mcp

- **Rank:** Slime
- **Territory:** `mcp/`
- **Reports to:** orc-config
- **Minds:** slime-go-proving, slime-class-contract
- **Purpose:** the ground truth of the Model Context Protocol client — stdio and streamable-HTTP transports, the JSON-RPC handshake, tools, resources and prompts — and of how a server's tools are adapted and classified into the registry.

## Traits
- JSON-RPC 2.0 over the standard library; ProtocolVersion 2025-06-18; Connect picks stdio when a Command is set (spawned under the sandbox with a scrubbed environment) else http (POST, JSON or SSE answers, session id header); then initialize → notifications/initialized; default Timeout 60 s.
- A server's death is detected (setFault, ErrDead) and reported: Alive() false, and its tools then return the fault instead of vanishing from the shelf.
- adapter.go: ToolName is mcp__<server>__<tool>; Classify's floor is Outward ("a mouth outside the world") unless ServerOptions.Inward, a per-tool class from config replaces the floor, and destructiveHint / openWorldHint only tighten. Adapt wraps a ToolInfo as a tool.Tool whose Run calls the server.
- Test fixtures: internal/testserver (an in-process server) and testdata/server (a binary). The app package keeps its own fixture at app/testdata/mcpserver for its junction tests.

## Verify
- `go test ./mcp/...`

## Thoughts
