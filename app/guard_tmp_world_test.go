package app

import (
	"strings"
	"testing"
)

// A path inside the world is never herded into scratch, wherever the world lives — /app in a
// benchmark's container, a dir under /tmp or /var; what leaves the world still is.
func TestTmpRedirectLeavesTheWorldItsOwnPaths(t *testing.T) {
	scratch := "/app/.isekai/tmp/scratch/r1"
	for _, tc := range []struct{ cmd, world string }{
		{"printf '%s\\n' 'hello bench' > /app/hello.txt && cat /app/hello.txt", "/app"},
		{"cat /tmp/w/notes.md", "/tmp/w"},
		{"ls /var/lib/proj/src", "/var/lib/proj"},
	} {
		if out, msg := rewriteTmpCommand(tc.cmd, scratch, tc.world, false); out != tc.cmd || msg != "" {
			t.Errorf("%q in world %s was rewritten: %q %q", tc.cmd, tc.world, out, msg)
		}
	}
	out, _ := rewriteTmpCommand("echo x > /tmp/out.txt && cat /app/hello.txt", scratch, "/app", false)
	if !strings.Contains(out, scratch+"/out.txt") || !strings.Contains(out, "/app/hello.txt") || strings.Contains(out, "root/app") {
		t.Fatalf("off-world /tmp is herded, the world's own path is kept: %q", out)
	}
}
