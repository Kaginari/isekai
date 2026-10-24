package board

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kaginari/isekai/queue"
)

// The board panel reads the shared shelf whole and writes nothing: the open briefs (claimed
// first, the wire's envelope dropped, closed work only counted), every deck with its step in
// hand (a finished deck after the ones in hand, an empty one left off), and each body's last
// verdict.
func TestAPIBoardReadsTheSharedShelf(t *testing.T) {
	root := t.TempDir()
	copyTree(t, fixtureRoot(t, "world"), root)
	dir := filepath.Join(root, ".isekai")
	q := queue.Open(filepath.Join(dir, "shared", "queue.jsonl"), nil)
	for _, it := range []queue.Item{
		{Ask: "@ASK draft\nshorten the token TTL to 15 minutes", For: "slime-auth", From: "orc-security"},
		{Ask: "weigh the import", For: "orc-security", Rank: 1},
		{Ask: "landed already", For: "slime-auth"},
	} {
		if _, err := q.Add(it, "rimuru"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := q.Claim("q2", "orc-security"); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Close("q3", "slime-auth", queue.Done, "ok"); err != nil {
		t.Fatal(err)
	}
	plans := filepath.Join(dir, "tmp", "plans")
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(plans, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("elf-engine.json", `{"body":"elf-engine","updated":"2026-09-25T10:00:00Z","items":[{"text":"survey","status":"done"},{"text":"slice the loop","status":"in_progress"},{"text":"land it","status":"pending"}]}`)
	write("orc-security.json", `{"body":"orc-security","updated":"2026-09-26T10:00:00Z","items":[{"text":"gate","status":"done"}]}`)
	write("slime-empty.json", `{"body":"slime-empty","items":[]}`)
	before := snapshotTree(t, dir)

	b := New(Options{WorldRoot: root, Now: func() time.Time { return fixedNow }})
	t.Cleanup(b.Close)
	var v SharedView
	getJSONLocal(t, b, "/api/board", &v)
	if !v.Queue.Reading.Lit || v.Queue.Held != 1 || v.Queue.Claimed != 1 || v.Queue.Closed != 1 || len(v.Queue.Items) != 2 {
		t.Fatalf("queue: %+v", v.Queue)
	}
	if v.Queue.Items[0].ID != "q2" || v.Queue.Items[0].Owner != "orc-security" || v.Queue.Items[1].Ask != "shorten the token TTL to 15 minutes" {
		t.Fatalf("the open briefs, claimed first, envelope dropped: %+v", v.Queue.Items)
	}
	if len(v.Decks) != 2 || v.Decks[0].Body != "elf-engine" || v.Decks[0].Done != 1 || v.Decks[0].Total != 3 || v.Decks[0].Current != "slice the loop" || v.Decks[1].Body != "orc-security" || v.Decks[1].Current != "" {
		t.Fatalf("decks: %+v", v.Decks)
	}
	if len(v.Reports) == 0 || v.Reports[0].Status == "" {
		t.Fatalf("reports: %+v", v.Reports)
	}
	seen := map[string]bool{}
	for _, r := range v.Reports {
		if seen[r.As] {
			t.Fatalf("a body's report twice: %+v", v.Reports)
		}
		seen[r.As] = true
	}
	if after := snapshotTree(t, dir); after != before {
		t.Fatal("reading the board wrote to the world")
	}
}

// snapshotTree is every file under dir with its size and mtime: equal snapshots, nothing written.
func snapshotTree(t *testing.T, dir string) string {
	t.Helper()
	var s string
	_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			s += p + " " + info.ModTime().String() + " " + fmt.Sprint(info.Size()) + "\n"
		}
		return nil
	})
	return s
}

// A deck name is a file name as the plan tool pens it: a separator in ?body= never climbs out
// of the decks' directory.
func TestAPIPlanStaysInTheDecks(t *testing.T) {
	root := t.TempDir()
	copyTree(t, fixtureRoot(t, "world"), root)
	if err := os.WriteFile(filepath.Join(root, ".isekai", "secret.json"), []byte(`{"items":[{"text":"not a deck","status":"pending"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	b := New(Options{WorldRoot: root, Now: func() time.Time { return fixedNow }})
	t.Cleanup(b.Close)
	var v map[string]any
	getJSONLocal(t, b, "/api/plan?body=..%2F..%2Fsecret", &v)
	if v["empty"] != true || v["items"] != nil {
		t.Fatalf("a climbing deck name read outside the decks: %v", v)
	}
}
