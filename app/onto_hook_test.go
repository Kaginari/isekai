package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/loop"
	"github.com/Kaginari/isekai/tool"
	"github.com/Kaginari/isekai/world"
)

func intoResult() tool.Result { return tool.Result{} }

// the onto hooks: ledger rows always on write-class acts, one unowned-soil advisory per run;
// claimed soil never nags.
func TestOntoNote(t *testing.T) {
	root := t.TempDir()
	write := func(p, s string) {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(root+"/.isekai/isekai.md", "# law\n")
	write(root+"/.isekai/orc/engine/README.md",
		"# orc-engine\n\n- **Rank:** Orc\n- **Territory:** `onto/`\n- **Reports to:** elf-isekai\n")
	a := testTmpApp(t, root)
	a.Cfg.Law.Territory.Enabled = true
	w, err := world.Open(root, world.Isekai(), world.Options{Ontology: true})
	if err != nil {
		t.Fatalf("world: %v", err)
	}
	a.World = w
	note := a.ontoNote()
	if note == nil {
		t.Fatal("territory law on + a world → the onto hooks ride")
	}
	jpath := filepath.Join(t.TempDir(), "run.jsonl")
	newSess := func() *loop.Session {
		return &loop.Session{RunID: "r-1", Engine: &loop.Engine{As: "elf-isekai"}, Journal: &loop.Journal{Path: jpath}}
	}
	call := func(s *loop.Session, tool_, path string) string {
		in, _ := json.Marshal(map[string]string{"path": path})
		st := &loop.StepRecord{ID: "s1", Tool: tool_, Input: in}
		st.Result = intoResult() // success — status stamp timing should not matter
		return note(context.Background(), s, st)
	}
	s := newSess()
	// an owned path: ledger row, no advisory
	if out := call(s, "edit", "onto/world.go"); out != "" {
		t.Fatalf("claimed soil never nags: %q", out)
	}
	// unowned: advisory fires once, then silence (one per run by law)
	out := call(s, "write", "docs/x.md")
	if !strings.Contains(out, "unowned soil") || !strings.Contains(out, "docs/x.md") {
		t.Fatalf("advisory must name the unowned path: %q", out)
	}
	if out := call(s, "write", "docs/y.md"); out != "" {
		t.Fatalf("one nudge per run: %q", out)
	}
	// reads don't count
	if out := call(s, "read", "onto/world.go"); out != "" {
		t.Fatalf("only write-class acts: %q", out)
	}
	// the journal grew two onto rows + the mapped bodies
	data, _ := os.ReadFile(jpath)
	rows := strings.Count(string(data), `"t":"onto"`)
	if rows != 3 {
		t.Fatalf("ledger rows must be logged per write-class act (3 writes, the read-class absent): %d rows in\n%s", rows, string(data))
	}
	if !strings.Contains(string(data), `"owner":"orc-engine"`) || !strings.Contains(string(data), `"owner":"none"`) {
		t.Errorf("ledger must carry owner names: %s", string(data))
	}
	// territory law off → hook away (court-class safety)
	a.Cfg.Law.Territory.Enabled = false
	if a.ontoNote() != nil {
		t.Error("law.territory.enabled=false → no onto hooks")
	}
}
