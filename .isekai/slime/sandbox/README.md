# slime-sandbox

- **Rank:** Slime
- **Territory:** `sandbox/`, `shell/`
- **Reports to:** orc-safety
- **Minds:** slime-go-proving, slime-class-contract
- **Purpose:** the ground truth of containment at the process level: bubblewrap around every command, env scrubbing, and the one persistent bash per body with its background jobs.

## Traits
- baseArgs: `--ro-bind / /`, `--tmpfs /tmp`, `--dev /dev`, `--proc /proc`, `--die-with-parent`, `--new-session`, `--bind <root> <root>` read-write, each RWPaths entry that exists, `--unshare-net` unless Call.Network (the gate let it out), `--chdir`. The probe runs once, lazily; ProbeBwrap(bin, timeout) tries a real bwrap and keeps the reason on failure.
- Mode None with Why when bwrap is missing or user namespaces are restricted (Ubuntu 24.04's kernel.apparmor_restrict_unprivileged_userns — CI sets it to 0 and installs bubblewrap; this machine has /usr/bin/bwrap). Options.Enabled false is Mode None "disabled by config". Status() is the line `status` prints.
- ScrubEnv drops names matching `(^|_)(API_KEY|TOKEN|SECRET|PASSWORD|PASSWD)$` (case-insensitive) plus EnvDrop globs (path.Match); EnvAllow wins over every drop; EnvSet is applied after scrubbing and wins over the process env (the package registries). Names like AWS_SECRET_ACCESS_KEY are NOT dropped by default.
- shell: one bash per body; every command is framed by a unique sentinel carrying the exit code and the resulting cwd (exitTrap on EXIT catches a command that exits the shell); default per-command timeout 10 min; a timeout kills the whole tree via /proc descendants; `restart` throws the session away; Background(name, cmd, network) jobs are named `^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`, logged to a file, listed and killed on Close; a command the gate let reach the network runs in a one-shot net-enabled shell seeded with the session's exports and cwd (oneShot).
- TestBwrapConfinement and TestSandboxedNetworkOneShot need bwrap; TestTimeoutKillsTree was flaky under -race on CI (commit 80ba99d gives the tree a moment before reading /proc).

## Verify
- `go test ./sandbox/... ./shell/...`

## Thoughts
