package tool

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/queue"
)

// TestQueueToolReadsFreeMovesWrite: list and show settle as reads; touch, requeue and close as
// writes that name no path — the shared shelf is the harness's, never billed to the body.
func TestQueueToolReadsFreeMovesWrite(t *testing.T) {
	q := QueueTool(queue.Open(filepath.Join(t.TempDir(), "queue.jsonl"), nil), "slime-auth")
	for in, want := range map[string]Class{
		`{"op":"list"}`: Read, `{"op":"show","id":"q1"}`: Read,
		`{"op":"touch","id":"q1"}`: Write, `{"op":"requeue","id":"q1"}`: Write, `{"op":"close","id":"q1","drop":true}`: Write,
	} {
		c, _ := q.Settle(Env{}, json.RawMessage(in))
		if c.Class != want || len(c.Paths) != 0 {
			t.Errorf("%s: %v %v", in, c.Class, c.Paths)
		}
	}
}

// TestQueueToolMovesOnlyHeldWork: a body touches and closes what it holds; another body's work
// is refused, and a closed item stays closed.
func TestQueueToolMovesOnlyHeldWork(t *testing.T) {
	s := queue.Open(filepath.Join(t.TempDir(), "queue.jsonl"), nil)
	s.Add(queue.Item{For: "slime-auth", Ask: "tighten the ttl"}, "orc-security")
	s.Add(queue.Item{For: "slime-api", Ask: "page the routes"}, "orc-security")
	s.Claim("q1", "slime-auth")
	s.Claim("q2", "slime-api")
	run := func(as, in string) Result {
		return QueueTool(s, as).Run(context.Background(), Env{}, json.RawMessage(in))
	}
	if r := run("slime-auth", `{"op":"list"}`); r.Err || !strings.Contains(r.Output, "q1") || !strings.Contains(r.Output, "← yours") || strings.Count(r.Output, "← yours") != 1 {
		t.Fatalf("list: %+v", r)
	}
	if r := run("slime-auth", `{"op":"touch","id":"q2","note":"mine"}`); !r.Err || !strings.Contains(r.Output, "held by slime-api") {
		t.Fatalf("a body touched work it does not hold: %+v", r)
	}
	if r := run("slime-auth", `{"op":"touch","id":"q1","note":"halfway"}`); r.Err || len(r.Wrote) != 0 {
		t.Fatalf("touch: %+v", r)
	}
	if r := run("slime-auth", `{"op":"close","id":"q1","note":"landed"}`); r.Err || !strings.Contains(r.Output, "done") {
		t.Fatalf("close: %+v", r)
	}
	if r := run("slime-auth", `{"op":"requeue","id":"q1"}`); !r.Err {
		t.Fatalf("closed work moved: %+v", r)
	}
	if r := run("slime-auth", `{"op":"show","id":"q1"}`); r.Err || !strings.Contains(r.Output, "r4 close by slime-auth — landed") {
		t.Fatalf("show: %+v", r)
	}
	if r := run("slime-auth", `{"op":"touch"}`); !r.Err {
		t.Fatal("a move with no id ran")
	}
}
