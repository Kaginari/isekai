package app

import (
	"encoding/json"
	"strings"
	"testing"
)

// Every tool a body is offered declares a schema that is JSON: one that is not breaks the request
// encoding of every model call that carries the shelf (a mock never encodes, so it passes unseen).
func TestEveryShelfSchemaIsJSON(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.write(".isekai/config.yaml", "tools: {bash: {sandbox: none}}\nui: {board: {autostart: false}}\n")
	a, err := New(Options{Dist: "isekai", Root: w.root, Cwd: w.root, Home: w.home, Env: func(k string) string { return map[string]string{"HOME": w.home}[k] },
		In: strings.NewReader(""), Out: &strings.Builder{}, Err: &strings.Builder{}, Quiet: true, IsTTY: func() bool { return false }, NoBoard: true})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	for _, body := range []string{"rimuru", "slime-auth"} {
		r, done := a.Shelf.Build(body)
		for _, d := range r.Defs() {
			if len(d.Schema) > 0 && !json.Valid(d.Schema) {
				t.Errorf("%s's %s schema is not JSON: %s", body, d.Name, d.Schema)
			}
		}
		if done != nil {
			done()
		}
	}
}
