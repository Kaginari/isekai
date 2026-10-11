package world

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseASLShape(t *testing.T) {
	good := []string{
		"ascend orc when reports >= 25 && passRate >= 90%",
		"ascend orc when reports>=3",
		"ascend slime when (reports > 10 && failRate <= 5%) || holes == 0",
		"ascend orc when !(fails > 0)",
	}
	for _, g := range good {
		a, err := ParseASL(g)
		if err != nil {
			t.Fatalf("%q: %v", g, err)
		}
		if a.Target == "" || a.Cond == nil {
			t.Fatalf("%q: target %q cond nil", g, a.Target)
		}
		// round-trip: the parsed form re-renders as spec words
		if !strings.HasPrefix(a.String(), "ascend "+a.Target+" when ") {
			t.Fatalf("%q rendered %q", g, a.String())
		}
	}
	bad := []string{
		"",                            // empty
		"orc when reports >= 1",       // no ascend
		"ascend orc reports >= 1",     // no when
		"ascend orc when",             // no condition
		"ascend orc when foo >= 1",    // unknown metric
		"ascend orc when reports ~ 1", // bad op
	}
	for _, b := range bad {
		if _, err := ParseASL(b); err == nil {
			t.Fatalf("%q must not parse", b)
		}
	}
}

func TestASLEval(t *testing.T) {
	as, err := ParseASL("ascend orc when reports >= 4 && passRate >= 75%")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		m    Metrics
		want bool
	}{
		{Tally(4, 4, 0, 0), true},  // 4 reports, all pass → 100%
		{Tally(4, 3, 0, 0), true},  // 75% exactly
		{Tally(4, 2, 0, 0), false}, // 50%
		{Tally(3, 3, 0, 0), false}, // too few reports
		{Tally(0, 0, 0, 0), false}, // no record → rates read zero
	}
	for i, c := range cases {
		if got := as.Cond.Eval(c.m); got != c.want {
			t.Fatalf("case %d: %v → %v, want %v", i, c.m, got, c.want)
		}
	}
}

func TestASLBooleanPrecedence(t *testing.T) {
	// && binds tighter than ||; parens group; ! negates
	as, err := ParseASL("ascend orc when reports >= 10 && failRate <= 5% || reports >= 50")
	if err != nil {
		t.Fatal(err)
	}
	// reports 50 with bad failRate crosses on the high-report lane
	if !as.Cond.Eval(Tally(50, 20, 30, 0)) {
		t.Fatal("the || lane must fire on a high report count")
	}
	// reports 10 clean crosses the && lane
	if !as.Cond.Eval(Tally(10, 10, 0, 0)) {
		t.Fatal("the && lane must fire on a clean record")
	}
	// reports 10 with a failure does not
	if as.Cond.Eval(Tally(10, 9, 1, 0)) {
		t.Fatal("failRate <= 5% with 1/10 a failure must not cross")
	}
}

func TestFoldMetricsFoldsStatus(t *testing.T) {
	root := ascendWorld(t)
	w := openWorldRanks(t, root, Ranks(DefaultRanks(Isekai())))
	w.FoldMetrics("orc-a", "PASS", 0)
	w.FoldMetrics("orc-a", "DONE", 1)
	w.FoldMetrics("orc-a", "FAIL", 2)
	w.FoldMetrics("slime-x", "PASS", 0)
	m := w.Metrics("orc-a")
	if m["reports"] != 3 || m["passes"] != 2 || m["fails"] != 1 || m["holes"] != 3 {
		t.Fatalf("tally: %+v", m)
	}
	if got := w.Metrics("nobody"); got["reports"] != 0 {
		t.Fatalf("an unknown body reads zero: %+v", got)
	}
}

func TestRankValidateAscendsWhen(t *testing.T) {
	lex := Isekai()
	mk := func(rs []Rank) Ranks { return rs }
	good := mk([]Rank{
		{Name: "elf", ReportsTo: Rimuru},
		{Name: "orc", ReportsTo: "elf", HoldsGate: true},
		{Name: "high-orc", ReportsTo: "elf", Base: "orc", AscendsWhen: "ascend orc when reports >= 5"},
	})
	if err := good.Validate(); err != nil {
		t.Fatalf("a lawful ascendsWhen: %v", err)
	}
	badBase := mk([]Rank{
		{Name: "elf", ReportsTo: Rimuru},
		{Name: "orc", ReportsTo: "elf", HoldsGate: true},
		{Name: "high-orc", ReportsTo: "elf", Base: "orc", AscendsWhen: "ascend slime when reports >= 5"},
	})
	if err := badBase.Validate(); err == nil {
		t.Fatal("ascendsWhen of a different rank must be refused")
	}
	badGrammar := mk([]Rank{
		{Name: "elf", ReportsTo: Rimuru},
		{Name: "orc", ReportsTo: "elf", HoldsGate: true},
		{Name: "high-orc", ReportsTo: "elf", Base: "orc", AscendsWhen: "quickly now"},
	})
	if err := badGrammar.Validate(); err == nil {
		t.Fatal("a non-ASL ascendsWhen must be refused")
	}
	noBase := mk([]Rank{
		{Name: "elf", ReportsTo: Rimuru},
		{Name: "elf", ReportsTo: Rimuru, AscendsWhen: "ascend elf when reports >= 5"},
	})
	noBase[1].Name = "watcher"
	if err := noBase.Validate(); err == nil {
		t.Fatal("ascendsWhen on a non-ascended rank must be refused")
	}
	_ = lex // lex read by implication; the table is the test's own
}

func TestAscensionMintsABody(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, ".isekai/isekai.md", lawText)
	writeDoc(t, root, ".isekai/orc/auth/README.md",
		"# orc-auth\n\n- **Rank:** Orc\n- **Territory:** `auth/`\n- **Reports to:** rimuru\n")
	writeDoc(t, root, "auth/login.go", "package auth\n")
	rs := Ranks{
		{Name: "elf", ReportsTo: Rimuru},
		{Name: "orc", ReportsTo: "elf", HoldsGate: true, Dir: "orc", Prefix: "orc-"},
		{Name: "high-orc", ReportsTo: "elf", Base: "orc", Dir: "high-orc", Prefix: "high-orc-",
			AscendsWhen: "ascend orc when reports >= 3 && passRate >= 100%"},
	}
	w := openWorldRanks(t, root, rs)
	if w.Creature("orc-auth") == nil {
		t.Fatalf("the orc must be on the roster: %v", namesOf(w))
	}
	// two reports do not yet cross
	w.FoldMetrics("orc-auth", "PASS", 0)
	w.FoldMetrics("orc-auth", "PASS", 0)
	if got := w.Ascensions(); len(got) != 0 {
		t.Fatalf("two reports below the condition must not ascend: %+v", got)
	}
	if w.Creature("orc-auth") == nil {
		t.Fatal("the base stays put")
	}
	// the third clean report crosses: the harness mints the high body itself
	w.FoldMetrics("orc-auth", "PASS", 0)
	events := w.Ascensions()
	if len(events) != 1 {
		t.Fatalf("a crossing must mint one ascension, got %+v holes %v", events, w.AscendHoles())
	}
	ev := events[0]
	if ev.Base != "orc-auth" || ev.Rank != "high-orc" {
		t.Fatalf("event: %+v", ev)
	}
	if ev.New != "high-orc-auth" {
		t.Fatalf("the ascended id rides its own prefix: %q", ev.New)
	}
	want := ".isekai/high-orc/auth/README.md"
	if filepath.ToSlash(ev.Doc) != want {
		t.Fatalf("the minted doc: %q, want %q", ev.Doc, want)
	}
	full := filepath.Join(root, filepath.FromSlash(want))
	text, err := os.ReadFile(full)
	if err != nil {
		t.Fatal(err)
	}
	for _, frag := range []string{"high-orc-auth", "**Rank:** high-orc", "Ascendant of:** `orc-auth`", "ascend orc when"} {
		if !strings.Contains(string(text), frag) {
			t.Fatalf("the minted doc must carry %q:\n%s", frag, text)
		}
	}
	// the reloaded roster holds the minted body, ranked as its ascended rank
	if c := w.Creature("high-orc-auth"); c == nil || c.Rank != "high-orc" {
		var rank string
		if c != nil {
			rank = c.Rank
		}
		t.Fatalf("the minted body must be on the roster as high_orc, got %q — roster %v", rank, namesOf(w))
	}
	// the base's doc was not overwritten
	base, err := os.ReadFile(filepath.Join(root, ".isekai/orc/auth/README.md"))
	if err != nil || !strings.Contains(string(base), "# orc-auth") {
		t.Fatalf("the base doc must stand: %v\n%s", err, base)
	}
	// the minted snapshot marks the crossing
	if m := w.Metrics("orc-auth"); m["reports"] != 3 || m["passes"] != 3 {
		t.Fatalf("snapshot: %+v", m)
	}
	// ascending again from the same body does not double-mint
	w.FoldMetrics("orc-auth", "PASS", 0)
	if got := w.Ascensions(); len(got) != 1 {
		t.Fatalf("a second report from the ascended body must not remint: %+v", got)
	}
}

func TestAscensionMintRefusesCollision(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, ".isekai/isekai.md", lawText)
	writeDoc(t, root, ".isekai/orc/auth/README.md",
		"# orc-auth\n\n- **Rank:** Orc\n- **Territory:** `auth/`\n- **Reports to:** rimuru\n")
	rs := Ranks{
		{Name: "elf", ReportsTo: Rimuru},
		{Name: "orc", ReportsTo: "elf", HoldsGate: true, Dir: "orc", Prefix: "orc-"},
		{Name: "high-orc", ReportsTo: "elf", Base: "orc", Dir: "high-orc", Prefix: "high-orc-",
			AscendsWhen: "ascend orc when reports >= 1"},
	}
	w := openWorldRanks(t, root, rs)
	w.FoldMetrics("orc-auth", "PASS", 0)
	if got := w.Ascensions(); len(got) != 1 {
		t.Fatalf("first crossing must mint: %+v holes %v", got, w.AscendHoles())
	}
	// a fresh world over the same disk: the doc stands, so a re-crossing must skip the mint
	// and log a hole rather than overwrite
	w2 := openWorldRanks(t, root, rs)
	if w2.Creature("high-orc-auth") == nil {
		t.Fatalf("the on-disk ascended body must rejoin the roster: %v", namesOf(w2))
	}
	w2.FoldMetrics("orc-auth", "PASS", 0)
	if got := w2.Ascensions(); len(got) != 0 {
		t.Fatalf("a doc-backed body never remints: %+v", got)
	}
}

func TestAscensionNeverWithoutABaseRecord(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, ".isekai/isekai.md", lawText)
	writeDoc(t, root, ".isekai/orc/auth/README.md",
		"# orc-auth\n\n- **Rank:** Orc\n- **Territory:** `auth/`\n- **Reports to:** rimuru\n")
	rs := Ranks{
		{Name: "elf", ReportsTo: Rimuru},
		{Name: "orc", ReportsTo: "elf", HoldsGate: true, Dir: "orc", Prefix: "orc-"},
		{Name: "high-orc", ReportsTo: "elf", Base: "orc", Dir: "high-orc", Prefix: "high-orc-",
			AscendsWhen: "ascend orc when reports >= 1"},
	}
	w := openWorldRanks(t, root, rs)
	// a body that never filed a report reads zero across every metric — a zero record never crosses
	w.FoldMetrics("orc-ghost", "PASS", 0)
	if got := w.Ascensions(); len(got) != 0 {
		t.Fatalf("nobody on the roster named orc-ghost: %+v", got)
	}
}

// helpers shared across world tests live in world_test.go; these two are local fakes to keep
// the ascension suite self-contained.

func writeDoc(t *testing.T, root, rel, text string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func namesOf(w *World) []string {
	var out []string
	for _, c := range w.Creatures {
		out = append(out, c.Name)
	}
	return out
}

func openWorldRanks(t *testing.T, root string, rs Ranks) *World {
	t.Helper()
	w, err := Open(root, Isekai(), Options{Ranks: rs, Ontology: true})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

// TestIntelligence is the coded evaluator the High ascension reads: the deterministic fold of
// the body's passRate against its holeRate — zero at zero reports, floored at zero (never
// negative) when holes outnumber passes. It is a world metric — no prompt, no ask.
func TestIntelligence(t *testing.T) {
	// no record: nothing thinks, nothing ascends
	if got := Tally(0, 0, 0, 0)["intelligence"]; got != 0 {
		t.Fatalf("zero record must read zero: got %v", got)
	}
	// pass with no holes reads its full passRate
	if got := Tally(10, 9, 0, 0)["intelligence"]; got != 90 {
		t.Fatalf("9/10 pass w/ 0 holes: want 90, got %v", got)
	}
	// holes discount the passes: 90 - 30 = 60
	if got := Tally(10, 9, 0, 3)["intelligence"]; got != 60 {
		t.Fatalf("9/10 pass w/ 3 holes: want 60, got %v", got)
	}
	// holes outnumbering passes floors at zero, never negative
	if got := Tally(10, 3, 0, 9)["intelligence"]; got != 0 {
		t.Fatalf("3/10 pass w/ 9 holes must floor at 0, got %v", got)
	}
	// the ASL reads it as a named operand
	cond, err := ParseCondition("intelligence >= 70 && reports >= 5")
	if err != nil {
		t.Fatal(err)
	}
	if !cond.Eval(Tally(10, 9, 0, 1)) {
		t.Fatal("9/10 pass w/ 1 hole must cross intelligence >= 70 && reports >= 5")
	}
	if cond.Eval(Tally(10, 7, 0, 4)) {
		t.Fatal("7/10 pass w/ 4 holes (intelligence 30) must stay under the band")
	}
}
