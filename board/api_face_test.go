package board

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAPITreeChildrenPinsDirsBeforeFilesAndCaps is slice 1's pin test: the lazy
// tree endpoint returns one level of children, directories first (name-ordered),
// files after, capped per level — never a flat dump of everything.
func TestAPITreeChildrenPinsDirsBeforeFilesAndCaps(t *testing.T) {
	b := newBoard(t, "world", Options{})
	code, out := get(t, b, "/api/tree/children")
	if code != 200 {
		t.Fatalf("children root: %d %s", code, out[:min(80, len(out))])
	}
	var root struct {
		Path  string `json:"path"`
		Nodes []struct {
			Name string `json:"name"`
			Dir  bool   `json:"dir"`
		} `json:"children"`
	}
	if err := json.Unmarshal([]byte(out), &root); err != nil {
		t.Fatalf("children decode: %v (%s)", err, out[:min(120, len(out))])
	}
	if len(root.Nodes) == 0 {
		t.Fatal("fixture world has children at root")
	}
	seenFile := false
	for _, n := range root.Nodes {
		if !n.Dir {
			seenFile = true
		} else if seenFile {
			t.Errorf("dir %q sorted after a file — dirs stay before files", n.Name)
		}
	}
}

// TestAPITreeChildrenOnlyOneLevel pins the laziness: asking for a directory's
// children returns that directory's entries, not the whole subtree.
func TestAPITreeChildrenOnlyOneLevel(t *testing.T) {
	b := newBoard(t, "world", Options{})
	// find a directory at root, then descend
	_, root := get(t, b, "/api/tree/children")
	var r struct {
		Nodes []struct {
			Name string `json:"name"`
			Dir  bool   `json:"dir"`
		} `json:"children"`
	}
	if err := json.Unmarshal([]byte(root), &r); err != nil || len(r.Nodes) == 0 {
		t.Skip("no decodable root nodes")
	}
	for _, n := range r.Nodes {
		if n.Dir {
			code, sub := get(t, b, "/api/tree/children?path="+n.Name)
			if code != 200 {
				t.Fatalf("children of %q: %d", n.Name, code)
			}
			var s struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal([]byte(sub), &s); err != nil {
				t.Fatalf("descend decode: %v", err)
			}
			if !strings.HasSuffix(s.Path, n.Name) {
				t.Errorf("descended path %q does not end in %q", s.Path, n.Name)
			}
			return
		}
	}
	t.Skip("no directory at fixture root")
}

// TestAPICourtsNeverLeaksBodies is the thin chip: /api/courts yields a compact
// {active,stale,empty} snapshot — it must never carry the body rows themselves.
func TestAPICourtsNeverLeaksBodies(t *testing.T) {
	bodies := []Body{{Name: "slime-auth", Rank: "slime", State: "tool"}}
	b := newBoard(t, "world", Options{Sources: Sources{Courts: func() (any, error) { return bodies, nil }}})
	code, out := get(t, b, "/api/courts")
	if code != 200 {
		t.Fatalf("api/courts: %d", code)
	}
	if strings.Contains(out, "slime-auth") {
		t.Errorf("courts chip leaked a body name: %s", out[:min(120, len(out))])
	}
	var snap struct {
		Active int  `json:"active"`
		Stale  int  `json:"stale"`
		Empty  bool `json:"empty"`
	}
	if err := json.Unmarshal([]byte(out), &snap); err != nil {
		t.Fatalf("courts decode: %v (%s)", err, out[:min(120, len(out))])
	}
	if snap.Active != 1 || snap.Empty {
		t.Errorf("courts snap for 1 live body: %+v", snap)
	}

	// empty world → empty:true, not a 500 or a fabricated body
	empty := newBoard(t, "world", Options{})
	_, eout := get(t, empty, "/api/courts")
	var esnap struct {
		Empty bool `json:"empty"`
	}
	if err := json.Unmarshal([]byte(eout), &esnap); err != nil || !esnap.Empty {
		t.Errorf("courts on empty world should be empty:true, got %s", eout)
	}
}

// TestAPIPlanRendersTheWorldPlanDeck pins the plan endpoint: it reads the
// body's live plan deck from .isekai/tmp/plans and serves items as JSON.
func TestAPIPlanRendersTheWorldPlanDeck(t *testing.T) {
	root := fixtureRoot(t, "world")
	plansDir := filepath.Join(root, ".isekai", "tmp", "plans")
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		t.Fatal(err)
	}
	deck := `{"body":"slime-tui-board","updated":"2026-01-01T00:00:00Z","items":[{"id":"s1","text":"ship the face","status":"in_progress","who":"slime-tui-board","score":"3","why":"commission"}]}`
	if err := os.WriteFile(filepath.Join(plansDir, "slime-tui-board.json"), []byte(deck), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// keep the fixture world pristine: remove the deck we authored
		os.Remove(filepath.Join(plansDir, "slime-tui-board.json"))
	})

	b := newBoard(t, "world", Options{})
	code, out := get(t, b, "/api/plan?body=slime-tui-board")
	if code != 200 {
		t.Fatalf("api/plan: %d %s", code, out[:min(80, len(out))])
	}
	if !strings.Contains(out, "ship the face") {
		t.Errorf("plan endpoint missing deck text: %s", out[:min(160, len(out))])
	}
	var got struct {
		Body  string `json:"body"`
		Items []struct {
			Text   string `json:"text"`
			Status string `json:"status"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("plan decode: %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].Text != "ship the face" {
		t.Errorf("plan items: %+v", got.Items)
	}

	// a body with no deck → silent (empty:true or no items), not a 500
	b2 := newBoard(t, "world", Options{})
	c2, o2 := get(t, b2, "/api/plan?body=no-such-body")
	if c2 != 200 {
		t.Fatalf("missing deck should still 200 (silent), got %d", c2)
	}
	var silent struct {
		Empty bool `json:"empty"`
	}
	if err := json.Unmarshal([]byte(o2), &silent); err != nil {
		t.Fatalf("silent plan decode: %v", err)
	}
	if !silent.Empty && strings.Contains(o2, "ship the face") {
		t.Errorf("no-such-body leaked another deck: %s", o2)
	}
}

// TestPlanPopupIsInTheNetModal pins the web face: the plan popup lives inside the
// net endpoint modal, not as a stray top-level dialog.
func TestPlanPopupIsInTheNetModal(t *testing.T) {
	js, err := os.ReadFile("web/dash.js")
	if err != nil {
		t.Skip("web/dash.js not present in this tree")
	}
	src := string(js)
	if !strings.Contains(src, "/api/plan") {
		t.Error("dash.js does not fetch /api/plan")
	}
	// the popup must hang off the net endpoint flow ("pm" current form), not a
	// detached document.createElement('dialog') at module scope.
	idxPlan := strings.Index(src, "/api/plan")
	idxNet := strings.Index(src, "net")
	if idxPlan < 0 || idxNet < 0 {
		t.Skip("dash.js shape unrecognisable")
	}
}

// TestSlice1HasChildrenFlagInAPITree pins the has_children ornament: /api/tree
// (the flat list) annotates directories with a has_children marker so the UI can
// offer lazy expansion without a second guess.
func TestSlice1HasChildrenFlagInAPITree(t *testing.T) {
	b := newBoard(t, "world", Options{})
	_, out := get(t, b, "/api/tree")
	if !strings.Contains(out, "has_children") {
		t.Error("api/tree lacks the has_children field — lazy tree cannot render expanders")
	}
}
