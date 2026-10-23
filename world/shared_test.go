package world

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/loop"
	"github.com/Kaginari/isekai/provider/mock"
	"github.com/Kaginari/isekai/queue"
)

// TestQueueOutranks: the world's queue moves a holder's work for the holder and for anyone up
// its Reports-to chain (rimuru tops it); a sibling or a child is refused.
func TestQueueOutranks(t *testing.T) {
	w := open(t, fixture(t))
	for _, c := range []struct {
		by, owner string
		want      bool
	}{
		{"orc-security", "slime-auth", true}, {"elf-core", "slime-auth", true}, {Rimuru, "slime-auth", true},
		{Rimuru, "", true}, {"slime-api", "slime-auth", false}, {"slime-auth", "orc-security", false}, {"elf-core", "", false},
	} {
		if got := w.Outranks(c.by, c.owner); got != c.want {
			t.Errorf("Outranks(%s, %s) = %v", c.by, c.owner, got)
		}
	}
	q := w.Queue()
	if q.Path() != filepath.Join(w.Dir(), "shared", "queue.jsonl") {
		t.Fatalf("the queue lives on the shared shelf: %s", q.Path())
	}
	q.Add(queue.Item{For: "slime-auth", Ask: "fix"}, "orc-security")
	q.Claim("q1", "slime-auth")
	if _, err := q.Requeue("q1", "slime-api", ""); !queue.IsRefusal(err) {
		t.Fatal("a sibling requeued work it does not hold")
	}
	if _, err := q.Requeue("q1", "orc-security", "reroute"); err != nil {
		t.Fatalf("the orc over the holder: %v", err)
	}
}

// TestTerritorySharedShelfIsHarnessSoil: the shared shelf is the harness's — a body's write
// there is refused like engine soil, naming the shelf.
func TestTerritorySharedShelfIsHarnessSoil(t *testing.T) {
	dir := fixture(t)
	w := open(t, dir)
	m := mock.New(
		writeCall("1", ".isekai/shared/queue.jsonl", "{}\n"),
		mock.Text("@S DONE\n@U colony c\n@E 0"),
	)
	e := w.Engine("slime-auth", build(m))
	r, err := e.Run(context.Background(), "edit the queue by hand")
	if err != nil || r.Status != loop.Done {
		t.Fatalf("%v %+v", err, r)
	}
	if len(r.Steps) != 1 || r.Steps[0].Status != "refused" {
		t.Fatalf("the write must refuse: %+v", r.Steps)
	}
	if res := m.Requests[1].Messages[2].ToolResults[0]; !strings.Contains(res.Content, "shared shelf") {
		t.Fatalf("the refusal names the shelf: %q", res.Content)
	}
	if _, err := os.Stat(filepath.Join(dir, ".isekai", "shared", "queue.jsonl")); err == nil {
		t.Fatal("the write landed on the shelf")
	}
}
