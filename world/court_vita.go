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
// ("openai/meta/muse-glimmer-30b"): the ladder lives in the same form the journal sees it.
// The rungs are OpenRouter model ids; the model part keeps its own vendor slash.
var laneLadder = []string{
	"openai/meta/muse-glimmer-30b",
	"openai/moonshotai/kimi-k3",
	"openai/z-ai/glm-5.3",
}

// laneIndex is ladder position; -1 when the lane is unknown (the law then leaves the
// mutation to the world config as pinned). Lanes match on the model part, after the first
// slash: the engine-cased form ("openai/x") and the machine-config form ("openrouter/x")
// name the same rung.
func laneIndex(l string) int {
	_, m, _ := strings.Cut(l, "/")
	for i, v := range laneLadder {
		if _, vm, _ := strings.Cut(v, "/"); vm == m {
			return i
		}
	}
	return -1
}

// laneUp gives the preferred next lane, or "" when the dying lane is out-of-ladder or
// already at its ceiling (then the advice is to split the ask, not to mutate).
// Provider/mount differences don't deny the law (laneIndex matches the model part).
func laneUp(l string) string {
	i := laneIndex(l)
	if i < 0 || i+1 >= len(laneLadder) {
		return ""
	}
	return laneLadder[i+1]
}
