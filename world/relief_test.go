package world

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/loop"
)

// endSite mirrors the court.go call-site law (warrant, then lift the act arm): this is the
// "court is done, what happens" seam the tests can hold onto honestly.
func endSite(w *World, body string, verdict string, r *loop.Result, court CourtOptions) (stamp string, relieved string) {
	stamp, _ = w.councilLedger(body, verdict, r, court)
	if rel := court.StressRelieveAt; rel > 0 {
		w.mu.Lock()
		strainNow := w.stress[body]
		w.mu.Unlock()
		if strainNow >= rel {
			if act, ok := w.relieve(body); ok {
				relieved = "council relief — " + body + " crosses relieveAt: " + act
			}
		}
	}
	return stamp, relieved
}

// TestCouncilReliefSplit: strain crossing the act rung splits the body's zone into two
// same-race young, retires the parent, and the roster re-reads the truth — the graph keeps
// only memory, never a hidden claim.
func TestCouncilReliefSplit(t *testing.T) {
	dir := t.TempDir()
	for p, s := range map[string]string{
		".isekai/isekai.md":              lawText,
		".isekai/log.md":                 "# Chronicle\n",
		".isekai/orc/security/README.md": "# orc-security\n\n- **Rank:** Orc\n- **Territory:** `src/`\n",
		".isekai/slime/two/README.md":    "# slime-two\n\n- **Rank:** Slime\n- **Territory:** `src/a/`, `src/b/`\n- **Reports to:** orc-security\n- **Verify:** `true`\n",
	} {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w, err := Open(dir, Isekai(), Options{Ontology: true})
	if err != nil {
		t.Fatal(err)
	}
	c1 := w.Creature("slime-two")
	if c1 == nil || len(c1.Territory) != 2 {
		t.Fatalf("split body must carry exactly 2 prefixes: %+v", c1)
	}
	parentBreadth := len(c1.Territory)
	court := CourtOptions{StressEscalateAt: 1, StressMaxConsults: 3, StressRelieveAt: 2}
	r := &loop.Result{Status: loop.Fail}
	if _, act := endSite(w, "slime-two", "FAIL", r, court); act != "" {
		t.Fatalf("strain below the rung must never act: %q", act)
	}
	_, act := endSite(w, "slime-two", "FAIL", r, court) // second miss: the strain hits the rung
	if !strings.Contains(act, "split into") || !strings.Contains(act, "slime-two-a") || !strings.Contains(act, "slime-two-b") {
		t.Fatalf("the act must be the split naming both young: %q", act)
	}
	if c := w.Creature("slime-two"); c != nil {
		t.Fatalf("the retiree must drop off the roster: %+v", c)
	}
	ca, cb := w.Creature("slime-two-a"), w.Creature("slime-two-b")
	if ca == nil || cb == nil {
		t.Fatalf("the young never reached the roster: %v / %v", ca, cb)
	}
	if len(ca.Territory)+len(cb.Territory) != parentBreadth {
		t.Fatalf("halves must total the parent's breadth: %d + %d ≠ %d", len(ca.Territory), len(cb.Territory), parentBreadth)
	}
	if own := w.Owner("src/a/x.go"); own == nil || own.Name != "slime-two-a" {
		t.Fatalf("src/a rides with slime-two-a: %+v", own)
	}
	if own := w.Owner("src/b/x.go"); own == nil || own.Name != "slime-two-b" {
		t.Fatalf("src/b rides with slime-two-b: %+v", own)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".isekai/instruments/ledgers.jsonl"))
	if !strings.Contains(string(b), `"kind":"relief"`) || !strings.Contains(string(b), "split") || !strings.Contains(string(b), "slime-two-a") || !strings.Contains(string(b), "slime-two-b") {
		t.Fatalf("the act must write the birth + the cut to the shelf:\n%s", b)
	}
	if w.Creature("orc-security") == nil {
		t.Fatal("relief must not slip the dispatching rank up")
	}
}

// TestCouncilReliefRetire: a one-prefix zone with nothing under it to divide is relieved by
// retirement — the seat keeps its graph memory while the soil opens again for dispatches.
// (A one-prefix zone WITH tree content splits instead — TestCouncilReliefTreeSplit below.)
func TestCouncilReliefRetire(t *testing.T) {
	dir := t.TempDir()
	for p, s := range map[string]string{
		".isekai/isekai.md":              lawText,
		".isekai/log.md":                 "# Chronicle\n",
		".isekai/orc/host/README.md":     "# orc-host\n\n- **Rank:** Orc\n- **Territory:** `zone/`\n",
		".isekai/slime/narrow/README.md": "# slime-narrow\n\n- **Rank:** Slime\n- **Territory:** `zone/narrow/`\n- **Reports to:** orc-host\n",
	} {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w, err := Open(dir, Isekai(), Options{Ontology: true})
	if err != nil {
		t.Fatal(err)
	}
	court := CourtOptions{StressEscalateAt: 1, StressMaxConsults: 3, StressRelieveAt: 1}
	r := &loop.Result{Status: loop.Fail}
	_, act := endSite(w, "slime-narrow", "FAIL", r, court)
	if !strings.Contains(act, "retired") || strings.Contains(act, "split") {
		t.Fatalf("a narrow zone relieves by retirement, never birth: %q", act)
	}
	if w.Creature("slime-narrow") != nil {
		t.Fatal("the retiree keeps graph memory, never claims soil again")
	}
	if own := w.Owner("zone/narrow/file.go"); own != nil && own.Name == "slime-narrow" {
		t.Fatalf("retired soil must release the slime's claim (the orc's umbrella stays lawful): %+v", own)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".isekai/slime/narrow/README.md"))
	if !strings.Contains(string(b), "**Retired:** true") {
		t.Fatalf("the retiree's doc owns its marker:\n%s", b)
	}
	// under rel=1 the second miss must not re-fire idempotently — one act per streak
	_, act = endSite(w, "slime-narrow", "FAIL", r, court)
	if act != "" {
		t.Fatalf("the already-retired must not re-fire: %q", act)
	}
}

// TestCouncilReliefTreeSplit: a single-prefix zone holding a real tree narrows by it —
// the 2026-10-09 zone-sizing law (one wide dir, every new file on one keeper, and no
// count-split possible with a single prefix): sub-dirs and direct files halve into two
// same-race young; nothing under the prefix is orphaned.
//
// The fixture roots the zone at `src/` — the roster's derive admits src-rooted structured
// claims under every genome state landed so far; an `app`-rooted single-prefix zone rides
// the census class-(f) commission in flight (its seat-slug admission law) and re-verifies
// when that slice lands.
func TestCouncilReliefTreeSplit(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		".isekai/isekai.md":            lawText,
		".isekai/log.md":               "# Chronicle\n",
		".isekai/orc/app/README.md":    "# orc-app\n\n- **Rank:** Orc\n- **Territory:** `src/`\n",
		".isekai/slime/wide/README.md": "# slime-wide\n\n- **Rank:** Slime\n- **Territory:** `src/`\n- **Reports to:** orc-app\n- **Verify:** `true`\n",
		"src/a/one.go":                 "package a\n",
		"src/b/two.go":                 "package b\n",
		"src/c/three.go":               "package c\n",
		"src/d/four.go":                "package d\n",
		"src/root.go":                  "package src\n",
	}
	for p, s := range files {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w, err := Open(dir, Isekai(), Options{Ontology: true})
	if err != nil {
		t.Fatal(err)
	}
	court := CourtOptions{StressEscalateAt: 1, StressMaxConsults: 3, StressRelieveAt: 1}
	r := &loop.Result{Status: loop.Fail}
	_, act := endSite(w, "slime-wide", "FAIL", r, court)
	if !strings.Contains(act, "split into") || !strings.Contains(act, "slime-wide-a") || !strings.Contains(act, "slime-wide-b") || !strings.Contains(act, "by the tree") {
		t.Fatalf("a one-prefix zone with a tree must split by it: %q", act)
	}
	ca, cb := w.Creature("slime-wide-a"), w.Creature("slime-wide-b")
	if ca == nil || cb == nil {
		t.Fatalf("the young never reached the roster: %v / %v", ca, cb)
	}
	// the union of the halves is the tree under src/ — no entry orphaned
	claims := append(append([]string{}, ca.Territory...), cb.Territory...)
	sort.Strings(claims)
	want := []string{"src/a", "src/b", "src/c", "src/d", "src/root.go"}
	sort.Strings(want)
	if strings.Join(claims, ",") != strings.Join(want, ",") {
		t.Fatalf("the halves must carry the whole tree: %v", claims)
	}
	for _, p := range []string{"src/a/one.go", "src/b/two.go", "src/c/three.go", "src/d/four.go", "src/root.go"} {
		own := w.Owner(p)
		if own == nil || (own.Name != "slime-wide-a" && own.Name != "slime-wide-b") {
			t.Fatalf("%s must ride with a young one: %+v", p, own)
		}
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".isekai/instruments/ledgers.jsonl"))
	if !strings.Contains(string(b), `"lane":"split-tree"`) {
		t.Fatalf("the act must write the tree-split to the shelf:\n%s", b)
	}
}
