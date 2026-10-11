package app

import (
	"testing"

	"github.com/Kaginari/isekai/value"
	"github.com/Kaginari/isekai/world"
)

// The nearest ascension: the body with the fewest conditions left, then the shortest distance;
// one left reads as its reading, several as their count and the nearest.
func TestNearestAscension(t *testing.T) {
	leaf := func(text string, got, want float64, met bool) value.Leaf {
		d := 0.0
		if !met {
			d = want - got
		}
		return value.Leaf{Text: text, Got: got, Want: want, Met: met, Distance: d}
	}
	bodies := []world.BodyValue{
		{Body: "orc-a", Ascension: &world.AscendProgress{To: "high-orc", Leaves: []value.Leaf{leaf("reports >= 20", 12, 20, false), leaf("passRate >= 90%", 84.333, 90, false)}}},
		{Body: "orc-engine", Ascension: &world.AscendProgress{To: "high-orc", Leaves: []value.Leaf{leaf("reports >= 20", 17, 20, false), leaf("passRate >= 90%", 95, 90, true)}}},
		{Body: "orc-done", Ascension: &world.AscendProgress{To: "high-orc", Met: true}},
		{Body: "slime-x"},
	}
	if got := nearestAscension(bodies); got != "orc-engine · reports 17/20 to ascend" {
		t.Fatalf("got %q", got)
	}
	if got := nearestAscension(bodies[:1]); got != "orc-a · 2 to ascend · passRate 84.3/90" {
		t.Fatalf("got %q", got)
	}
	if got := nearestAscension(bodies[2:]); got != "" {
		t.Fatalf("no body on its way: %q", got)
	}
}
