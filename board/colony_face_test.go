package board

import (
	"encoding/json"
	"testing"
)

// TestTreeUpBreaksATwoNodeCycle: two creatures that each name the other as the one above them
// never reach the throne — the loop is broken where it is found, the creature hangs under the
// throne by an implied reports edge, a note says so, and no parent chain loops.
func TestTreeUpBreaksATwoNodeCycle(t *testing.T) {
	c := &Colony{Nodes: []Node{{ID: "orc-a", Kind: "creature", Rank: "orc"}, {ID: "orc-b", Kind: "creature", Rank: "orc"}},
		Edges: []Edge{{From: "orc-a", To: "orc-b", Bond: "above"}, {From: "orc-b", To: "orc-a", Bond: "above"}}}
	treeUp(c)
	parent := map[string]string{}
	for _, n := range c.Nodes {
		parent[n.ID] = n.Parent
	}
	for _, id := range []string{"orc-a", "orc-b"} {
		seen := map[string]bool{}
		for at := id; at != c.Throne; at = parent[at] {
			if seen[at] || at == "" {
				t.Fatalf("%s never reaches the throne: parents %v", id, parent)
			}
			seen[at] = true
		}
	}
	implied := 0
	for _, e := range c.Edges {
		if e.Implied && e.To == c.Throne && e.Bond == "reports" {
			implied++
		}
	}
	if implied != 1 || len(c.Notes) != 1 {
		t.Errorf("one of the pair hangs under the throne by one implied edge and one note: %d edges, notes %v, parents %v", implied, c.Notes, parent)
	}
}

// TestAPIColonyOnAnEmptyWorld: /api/colony's HTTP face on a world with no creatures is still one
// tree — the throne alone, named, with empty (never null) node, edge and label lists.
func TestAPIColonyOnAnEmptyWorld(t *testing.T) {
	b := newBoard(t, "empty", Options{})
	code, body := get(t, b, "/api/colony")
	if code != 200 {
		t.Fatalf("/api/colony on the empty world: %d %s", code, body)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatalf("not json: %v\n%s", err, body)
	}
	for _, k := range []string{"nodes", "edges", "labels"} {
		if v, ok := raw[k]; !ok || string(v) == "null" {
			t.Errorf("%q must be a list on an empty world, got %s", k, v)
		}
	}
	var c struct {
		Throne string `json:"throne"`
		Nodes  []Node `json:"nodes"`
	}
	if err := json.Unmarshal([]byte(body), &c); err != nil {
		t.Fatal(err)
	}
	creatures := 0
	for _, n := range c.Nodes {
		if n.Kind == "creature" {
			creatures++
			if n.ID != c.Throne || !n.Throne {
				t.Errorf("the empty world holds only the throne, found %+v", n)
			}
		}
	}
	if c.Throne == "" || creatures != 1 {
		t.Errorf("the throne must stand alone: throne %q, %d creatures\n%s", c.Throne, creatures, body)
	}
}
