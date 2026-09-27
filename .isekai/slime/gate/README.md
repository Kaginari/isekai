# slime-gate

- **Rank:** Slime
- **Territory:** `gate/`
- **Reports to:** orc-safety
- **Minds:** slime-go-proving, slime-class-contract
- **Purpose:** the ground truth of the human gate: who answers, in what order, and the exact shape of an Answer with its provenance.

## Traits
- Decisions: not-needed, approved, denied, would-ask. Gate.Needs(c) is `c >= Outward || (Strict && c == Write)`; Request.Force (a permission rule said ask) makes the gate ask whatever the class.
- Ask order: not needed → pre-approved by class (Approve map, skipped under Force; By "pre-approved") → DryRun (would-ask, By "dry-run") → Gate.Answer (the TUI's choice block; anything but Approved becomes Denied; By "tui") → the TTY prompt (yes regex `^(?i)y(es)?$`; By "tty") → no TTY: Denied with the hint `pass --approve <class>` (By "no-tty"). Nothing in this list approves on its own.
- Flag is the name shown in a denial ("--approve" by default) so agent-one can name its own; Prompt replaces DefaultPrompt; In/Out default to stdin/stderr; IsTTY is injectable for tests.
- ParseApprove reads "outward,destructive" into the class set and rejects unknown names — a typo is an error, never a silent no-op.
- Standing pre-approvals are the human's written word: config refuses them from env (checkGate), and `config check` lists every one as a loosening.

## Verify
- `go test ./gate/...`

## Thoughts
