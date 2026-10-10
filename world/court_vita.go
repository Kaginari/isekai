package world

import (
	"strings"
	"time"
)

// Stress-burn ledger: a body that dies fighting the court clock leaves a memory of its
// lane here; the next dispatch of the same creature rides the next-fit lane and reads a
// one-line law note at ask-top. Law: the instrument remembers stress-deaths; the model
// does not re-fight stress every respawn — the harness carries the past forward.
//
// This is the legal answer to Veldora's 2026-10-02 observance: "the slime stays thinking
// long — that means the opposite of confidence", and "it's auto split or remove or born
// new — the genome changes to the lane best fit for the situation". The bodies themselves
// decided nothing; the harness swaps their DNA between spawns.

// stressEnd marks one body's court run as ending under clock or zone burn. Only
// stress-shaped deaths are kept; a FAIL from other shapes stays in the report's own story.
type stressEnd struct {
	Lane string    // the lane the run rode (provider/model key seen on the dying run)
	At   time.Time // when it burned out under the axe
}

// laneLadder is the fallback chain the law knows, weakest → strongest (provider-key form).
// A burn on the left of the chain upgrades; a burn already at the peak calls out the rank's
// allocation as the hole instead. Tuned by hand, kept short on purpose — the harness only
// nags within the ladder, never invents a lane.
// Lane names on the wire are provider-prefixed the way Provider.n stamps them
// ("openai/muse-glimmer-30b"): the ladder lives in the same form the journal sees it.
var laneLadder = []string{
	"openai/muse-glimmer-30b",
	"openai/kimi-3-conduit",
	"openai/glm-5-3-conduit",
}

// laneIndex is ladder position; -1 when the lane is unknown (the law then leaves the
// mutation to the world config as pinned).
func laneIndex(l string) int {
	for i, v := range laneLadder {
		if v == l {
			return i
		}
	}
	return -1
}

// laneUp gives the preferred next lane, or "" when the dying lane is out-of-ladder or
// already at its ceiling (then the advice is to split the ask, not to mutate).
// Both engine-cased forms ("openai/x") and machine-config forms ("litellm/x") ladder match:
// provider/mount differences don't deny the law.
func laneUp(l string) string {
	norm := strings.Replace(l, "litellm/", "openai/", 1)
	i := laneIndex(norm)
	if i < 0 || i+1 >= len(laneLadder) {
		return ""
	}
	return laneLadder[i+1]
}
