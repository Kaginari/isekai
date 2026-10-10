# slime-discover

- **Rank:** Slime
- **Territory:** `discover/`
- **Reports to:** orc-config
- **Minds:** slime-go-proving
- **Purpose:** the ground truth of harness parity: reading what Claude Code and OpenCode wrote — instructions, skills, commands, agents, MCP imports — as plain structs, read-only, each source behind its own switch.

## Traits
- Options: Root, Home, and the five switches Instructions, Skills, Commands, Agents, MCP; Sources.Claude and Sources.OpenCode are *bool, nil meaning on. Nothing here writes; the only consumer is app (app/discover.go and the MCP import in app/mcp.go), which turns the structs into prompt text, toolbox entries, slash commands and foreign bodies.
- Instructions: AGENTS.md and CLAUDE.md walked up from the root (nearest first, AGENTS.md before CLAUDE.md in one directory), then ~/.claude/CLAUDE.md and ~/.config/opencode/AGENTS.md; each carries its Scope and Path.
- Skills: level 1 only (name and description from front matter); Skill.Load() reads the body on demand; Builtin(text) wraps a skill the binary ships. Skill dirs: .claude/skills, .opencode/skill, .opencode/skills, .agents/skills and their home equivalents.
- Commands: .claude/commands and .opencode/command(s) markdown; Command.Expand substitutes $ARGUMENTS and $1..$9.
- Agents: .claude/agents and .opencode/agent(s), walked recursively; front matter name (else the file name), description, model, mode (default "subagent"), tools; the first file to claim a name wins, sorted by name.
- MCP: .mcp.json (Claude Code) and opencode.json(c)'s mcp block, each Server tagged with its Source; a malformed file is returned as an error, never skipped silently.

## Verify
- `go test -count=1 ./discover/`

## Thoughts
- (none yet)
