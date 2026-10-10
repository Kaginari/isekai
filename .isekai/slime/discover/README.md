# slime-discover

- **Rank:** Slime
- **Territory:** `discover/`
- **Reports to:** orc-config
- **Minds:** slime-go-proving
- **Purpose:** the ground truth of harness parity: reading what Claude Code and OpenCode wrote — instructions, skills, commands, agents, MCP imports — as plain structs, read-only, each source behind its own switch.

## Traits
- Options: Root, Home, and the five switches Instructions, Skills, Commands, Agents, MCP; Sources.Claude and Sources.OpenCode are *bool, nil meaning on. Nothing here writes.
- Instructions: CLAUDE.md / AGENTS.md walk-up from the root, then ~/.claude/CLAUDE.md and ~/.config/opencode/AGENTS.md; each carries its Scope and Path.
- Skills: level 1 only (name and description from front matter); Skill.Load() reads the body on demand; Builtin(text) wraps an embedded skill (the ui Mind). Skill dirs: .claude/skills, .agents/skills, .opencode/skill(s) and their home equivalents.
- Commands: .claude/commands and .opencode/command(s) markdown; Command.Expand substitutes $ARGUMENTS and $1..$9 (argN).
- Agents: .claude/agents and .opencode/agent(s), walked recursively; front matter name (else the file name), description, model, mode (default "subagent"), tools; the first file to claim a name wins, sorted by name.
- MCP: .mcp.json (Claude Code) and opencode.json(c)'s mcp block; a malformed file is returned as an error, never skipped silently.

## Verify
- `go test ./discover/...`

## Thoughts
