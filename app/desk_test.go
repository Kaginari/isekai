package app

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/tool"
	"github.com/Kaginari/isekai/world"
)

// A court's desk is its own run's file — the one the drainer and a resume read — and the throne
// keeps the session's: no body writes over another's working memory.
func TestDeskIsTheRunsOwn(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.write(".isekai/config.yaml", "ui: {board: {autostart: false}}\n")
	a := w.open()
	add := func(body, run, thought string) {
		t.Helper()
		in, _ := json.Marshal(map[string]string{"add": thought})
		if r := a.deskTool(body).Run(context.Background(), tool.Env{Root: a.Root, RunID: run}, in); r.Err {
			t.Fatal(r.Output)
		}
	}
	add(world.Rimuru, "rimuru-run", "the throne's plan")
	add("slime-auth", "slime-auth-r1", "the court's finding")
	add("slime-api", "slime-api-r2", "another court's finding")
	for path, want := range map[string]string{
		a.deskPath(a.SessionID):     "the throne's plan",
		a.deskPath("slime-auth-r1"): "the court's finding",
		a.deskPath("slime-api-r2"):  "another court's finding",
	} {
		b, err := os.ReadFile(path)
		if err != nil || strings.Count(string(b), "\n- ") != 1 || !strings.Contains(string(b), want) {
			t.Errorf("%s: %q %v, want only %q", path, b, err, want)
		}
	}
}
