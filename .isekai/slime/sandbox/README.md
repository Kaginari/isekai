# slime-sandbox

- **Rank:** Slime
- **Territory:** `sandbox/`, `shell/`
- **Reports to:** orc-safety
- **Minds:** slime-go-proving, gate-safety
- **Purpose:** the ground truth of containment at the process level: bubblewrap around every command, env scrubbing, process-group containment, and the one persistent bash per body with its background jobs.

## Traits
- baseArgs: `--ro-bind / /`, `--tmpfs /tmp`, `--dev /dev`, `--proc /proc`, `--die-with-parent`, `--new-session`, `--bind <root> <root>` read-write, each RWPaths entry that exists, `--unshare-net` unless Call.Network (the gate let it out), `--chdir <cwd>`. The probe runs once, lazily; ProbeBwrap(bin, timeout) runs /bin/true under those flags and keeps the first line of the failure as Why.
- Mode None with Why when bwrap is missing ("bwrap not on PATH"), cannot run (user namespaces restricted), Options.Enabled is false ("disabled by config") or the Sandbox is nil. Status() is the line `status` prints; Wrap returns argv untouched under None.
- ScrubEnv drops any name matching `(?i)(^|_)(API_KEY|TOKEN|SECRET|PASSWORD|PASSWD|PRIVATE_KEY|ACCESS_KEY|CREDENTIALS?)(_|$)` — so AWS_SECRET_ACCESS_KEY and GITHUB_TOKEN_X go — any value that is a URL carrying user:password (DATABASE_URL), plus EnvDrop names or path.Match globs. EnvAllow wins over every drop; Sandbox.Scrub applies EnvSet after scrubbing and it wins over the process env (the package registries).
- Contain(cmd) gives a command its own process group (Setpgid), kills the whole group with SIGKILL when its context ends, and bounds Wait by WaitDelay (2 s) so a backgrounded child holding the pipes cannot outlive the deadline. Sandbox.Command always ends in Contain; every command a body or the config can name (the bash tool, custom tools, hooks, rule checks) runs through Contain or Command — the binary's own fixed probes (git rev-parse, docker version) do not need it.
- shell: one bash per body started through Sandbox.Command; every command is framed by a unique sentinel carrying the exit code and the resulting cwd (exitTrap on EXIT catches a command that exits the shell); default per-command timeout 10 min; a timeout kills the whole tree via /proc descendants; `restart` throws the session away; Background(name, cmd, network) jobs are named `^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`, logged to a file, listed and killed on Close; a command the gate let reach the network runs in a one-shot net-enabled shell seeded with the session's exports and cwd (oneShot).
- TestBwrapConfinement and TestSandboxedNetworkOneShot need a working bwrap; TestTimeoutKillsTree and TestCommandDeadlineKillsTheTree read /proc and are timing-sensitive under -race.

## Verify
- `go test -count=1 -run 'TestScrubEnv|TestScrubCommonCredentialNames|TestEnvSetWinsAndTokenAllowed|TestNoneMode|TestCommandDeadlineKillsTheTree' ./sandbox/`
- `go test -count=1 ./sandbox/... ./shell/...`

## Thoughts
- (none yet)
