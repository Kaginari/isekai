package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/world"
)

// TestQueueCommand: `q` reads and moves the world's shared queue as the human — add, claim
// for a body, close; --json prints the items; a refusal exits 2.
func TestQueueCommand(t *testing.T) {
	w, a := forgeWorld(t)
	a.Close()
	run := func(args ...string) (int, string, string) {
		var out, errb bytes.Buffer
		code := Main("isekai", append([]string{"--root", w.root, "--quiet", "--no-board"}, args...), IO{In: strings.NewReader(""), Out: &out, Err: &errb, Env: func(k string) string {
			if k == "HOME" {
				return w.home
			}
			return ""
		}}, Version{Version: "test"})
		return code, out.String(), errb.String()
	}
	if code, out, errs := run("q", "add", "--for", "slime-auth", "--rank", "0", "tighten", "the", "ttl"); code != 0 || !strings.Contains(out, "queued q1") {
		t.Fatalf("q add: %d %q %q", code, out, errs)
	}
	if code, out, _ := run("q", "claim", "q1", "--as", "slime-auth"); code != 0 || !strings.Contains(out, "held by slime-auth") {
		t.Fatalf("q claim --as: %d %q", code, out)
	}
	if code, out, _ := run("--json", "q", "list"); code != 0 || !strings.HasPrefix(out, `[{"id":"q1"`) {
		t.Fatalf("q list --json: %d %q", code, out)
	}
	if code, out, _ := run("q", "close", "q1", "landed"); code != 0 || !strings.Contains(out, "done q1") {
		t.Fatalf("the human closes any work: %d %q", code, out)
	}
	if code, _, errs := run("q", "requeue", "q1"); code != 2 || !strings.Contains(errs, "only claimed work") {
		t.Fatalf("a refusal exits 2: %d %q", code, errs)
	}
	if b, err := os.ReadFile(filepath.Join(w.root, ".isekai", "shared", "queue.jsonl")); err != nil || strings.Count(string(b), "\n") != 3 {
		t.Fatalf("three moves, three lines: %v %q", err, b)
	}
	if _, out, _ := run("--help"); !strings.Contains(out, "q list") {
		t.Fatalf("help lists q: %s", out)
	}
}

// TestQueueToolOnEveryShelf: the queue tool reaches a slime's cut shelf and the throne's, and
// tools.queue.enabled: false takes it off with the origin named.
func TestQueueToolOnEveryShelf(t *testing.T) {
	_, a := forgeWorld(t)
	defer a.Close()
	for _, body := range []string{"slime-auth", world.Rimuru} {
		e := a.World.Engine(body, a.Build())
		if !strings.Contains(","+strings.Join(e.Tools.Names(), ",")+",", ",queue,") {
			t.Fatalf("%s's shelf lacks queue: %v", body, e.Tools.Names())
		}
		a.World.Close(e)
	}
	a.Cfg.Tools.Queue.Enabled = false
	if why := a.Shelf.Disabled()["queue"]; !strings.Contains(why, "tools.queue.enabled") {
		t.Fatalf("the switch names its origin: %q", why)
	}
}
