package world

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/config"
)

// ascendRanks is the law's table with a short ascension band, so a fixture can earn it.
func ascendRanks(when string) Ranks {
	rs := Ranks(DefaultRanks(Isekai()))
	for i := range rs {
		if rs[i].Name == "high_orc" {
			rs[i].AscendsWhen = when
		}
	}
	return rs
}

// ascendWorld is one lane: elf-core → orc-auth (verify, territory auth/) → slime-login.
func ascendWorld(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeDoc(t, root, ".isekai/isekai.md", lawText)
	writeDoc(t, root, ".isekai/log.md", "# Chronicle\n\nAppend-only.\n\n---\n")
	writeDoc(t, root, ".isekai/elf/core/README.md", "# elf-core\n\n- **Rank:** Elf\n- **Reports to:** rimuru\n")
	writeDoc(t, root, ".isekai/orc/auth/README.md",
		"# orc-auth\n\n- **Rank:** Orc\n- **Territory:** `auth/`\n- **Reports to:** elf-core\n- **Verify:** `test -d auth`\n\n## Traits\n\n1. tokens expire after one hour\n")
	writeDoc(t, root, ".isekai/slime/login/README.md",
		"# slime-login\n\n- **Rank:** Slime\n- **Territory:** `auth/login/`\n- **Reports to:** orc-auth\n- **Verify:** `test -d auth/login`\n")
	writeDoc(t, root, "auth/login/login.go", "package login\n")
	return root
}

// An ascension carries the base whole into the ascended rank's own dir, retires the base's seat,
// re-points the base's slimes to the risen body, keeps the ontology valid, and is logged.
func TestAscensionCarriesTheBodyAndRetiresTheBase(t *testing.T) {
	root := ascendWorld(t)
	w := openWorldRanks(t, root, ascendRanks("ascend orc when reports >= 3 && passRate >= 90%"))
	for i := 0; i < 3; i++ {
		w.FoldMetrics("orc-auth", "DONE", 0)
	}
	if got := w.Ascensions(); len(got) != 1 || got[0].New != "high-orc-auth" {
		t.Fatalf("one ascension to high-orc-auth: %+v holes %v", got, w.AscendHoles())
	}
	doc, err := os.ReadFile(filepath.Join(root, ".isekai/high-orc/auth/README.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, frag := range []string{"# high-orc-auth", "- **Rank:** high_orc", "**Ascendant of:** `orc-auth`", "- **Territory:** `auth/`", "- **Reports to:** elf-core", "`test -d auth`", "tokens expire after one hour", "## Ascension", "reports 3, passes 3"} {
		if !strings.Contains(string(doc), frag) {
			t.Errorf("the ascended doc must carry %q:\n%s", frag, doc)
		}
	}
	base, _ := os.ReadFile(filepath.Join(root, ".isekai/orc/auth/README.md"))
	if !strings.Contains(string(base), retiredMark) || !strings.Contains(string(base), "rose as high-orc-auth") {
		t.Fatalf("the base's seat retires, naming its ascendant:\n%s", base)
	}
	// the roster: one body on the soil, the slimes' line runs through it
	for _, n := range namesOf(w) {
		if n == "orc-auth" {
			t.Error("the retired base leaves the roster")
		}
	}
	hi := w.Creature("high-orc-auth")
	if hi == nil || hi.Rank != "high_orc" || strings.Join(hi.Territory, ",") != "auth" || hi.Parent != "elf-core" || !hi.VerifyCoded() {
		t.Fatalf("the risen body holds the base's territory, line and verify: %+v", hi)
	}
	if sl := w.Creature("slime-login"); sl == nil || sl.Parent != "high-orc-auth" {
		t.Fatalf("the base's slime now reports to the ascendant: %+v", sl)
	}
	if o := w.Owner("auth/session.go"); o == nil || o.Name != "high-orc-auth" {
		t.Errorf("the gate-holder fallback owner is the risen body: %v", o)
	}
	// it commands the base's slimes
	if err := w.mayDispatch("high-orc-auth", w.Creature("slime-login"), CourtOptions{StrictFlow: true}); err != nil {
		t.Errorf("the high orc orders its slimes: %v", err)
	}
	// the ontology stays valid
	if fs := w.Onto.Validate(); len(fs) != 0 {
		t.Errorf("the ascension must leave the ontology valid: %v", fs)
	}
	// log.md carries the ascension with the reading that earned it (Law 4)
	logText, _ := os.ReadFile(w.LogPath())
	if !strings.Contains(string(logText), "high-orc-auth — ascension — orc-auth rose as high-orc-auth") || !strings.Contains(string(logText), "earned: reports 3, passes 3") {
		t.Errorf("log.md must record the ascension:\n%s", logText)
	}
	// the dispatcher hears it once
	lines := w.ascendLines()
	if len(lines) != 1 || !strings.Contains(lines[0], "[ascension] orc-auth rose as high-orc-auth") {
		t.Errorf("the dispatcher reads the ascension: %v", lines)
	}
	if again := w.ascendLines(); len(again) != 0 {
		t.Errorf("announced once: %v", again)
	}
}

// The tally is a ledger replayed at Open: a record earned across sessions counts, and a body
// that ascended in an earlier session never mints again.
func TestAscensionTallyOutlivesTheSession(t *testing.T) {
	root := ascendWorld(t)
	rs := ascendRanks("ascend orc when reports >= 3 && passRate >= 90%")
	w1 := openWorldRanks(t, root, rs)
	w1.FoldMetrics("orc-auth", "DONE", 0)
	w1.FoldMetrics("orc-auth", "DONE", 0)
	if len(w1.Ascensions()) != 0 {
		t.Fatal("two reports do not cross")
	}
	w2 := openWorldRanks(t, root, rs)
	if m := w2.Metrics("orc-auth"); m.Reports != 2 || m.Passes != 2 {
		t.Fatalf("the tally replays at Open: %+v", m)
	}
	w2.FoldMetrics("orc-auth", "DONE", 0)
	if got := w2.Ascensions(); len(got) != 1 {
		t.Fatalf("the third report, a session later, crosses: %+v holes %v", got, w2.AscendHoles())
	}
	shelf, err := os.ReadFile(filepath.Join(root, ".isekai", filepath.FromSlash(ascensionShelf)))
	if err != nil || strings.Count(string(shelf), `"kind":"report"`) != 3 || strings.Count(string(shelf), `"kind":"ascend"`) != 1 {
		t.Fatalf("the shelf is append-only, one row per report and per ascension: %v\n%s", err, shelf)
	}
	w3 := openWorldRanks(t, root, rs)
	if m := w3.Metrics("orc-auth"); m.Reports != 3 {
		t.Fatalf("replayed again: %+v", m)
	}
	w3.FoldMetrics("orc-auth", "DONE", 0)
	if got := w3.Ascensions(); len(got) != 0 {
		t.Fatalf("an ascended body never mints twice: %+v", got)
	}
}

// A mint that cannot land is a hole: logged, and surfaced to the dispatcher as @?.
func TestAscensionHoleSurfaces(t *testing.T) {
	root := ascendWorld(t)
	rs := ascendRanks("ascend orc when reports >= 1")
	for i := range rs {
		if rs[i].Name == "high_orc" { // the old shape: the ascended rank on its base's dir
			rs[i].Dir, rs[i].Prefix = "orc", "orc-"
		}
	}
	w := openWorldRanks(t, root, rs)
	w.FoldMetrics("orc-auth", "DONE", 0)
	if len(w.Ascensions()) != 0 {
		t.Fatal("no mint on the base's own dir")
	}
	lines := w.ascendLines()
	if len(lines) != 1 || !strings.HasPrefix(lines[0], "@? ascension hole — ascend orc-auth") || !strings.Contains(lines[0], "a dir and prefix of its own") {
		t.Fatalf("the hole reaches the dispatcher as @?: %v", lines)
	}
	logText, _ := os.ReadFile(w.LogPath())
	if !strings.Contains(string(logText), "ascension hole — orc-auth") {
		t.Errorf("the hole is logged:\n%s", logText)
	}
	if w.Creature("orc-auth") == nil {
		t.Error("a failed mint leaves the base standing")
	}
}

// A body's own @S FAIL is no pass, whatever the loop's status said.
func TestAscensionCountsTheBodysVerdict(t *testing.T) {
	root := ascendWorld(t)
	w := openWorldRanks(t, root, ascendRanks("ascend orc when reports >= 2 && passRate >= 100%"))
	w.FoldMetrics("orc-auth", "DONE", 0)
	w.FoldMetrics("orc-auth", "FAIL", 0)
	if m := w.Metrics("orc-auth"); m.Passes != 1 || m.Fails != 1 || len(w.Ascensions()) != 0 {
		t.Fatalf("a fail is a fail: %+v", m)
	}
}

// The binary's table (config) and the world's default table ascend on the same conditions, in
// the ascended ranks' own dirs.
func TestDefaultAscensionAgrees(t *testing.T) {
	if config.DefaultAscendElf != DefaultAscendElf || config.DefaultAscendOrc != DefaultAscendOrc {
		t.Fatal("config and world disagree on the default ascension conditions")
	}
	for _, s := range []string{DefaultAscendElf, DefaultAscendOrc} {
		if _, err := ParseASL(s); err != nil {
			t.Fatalf("%q: %v", s, err)
		}
	}
	rs := Ranks(DefaultRanks(Isekai()))
	for _, n := range []string{"high_elf", "high_orc"} {
		r, _ := rs.Get(n)
		b, _ := rs.Get(r.Base)
		if r.Dir == b.Dir || r.Prefix == b.Prefix || r.Prefix != r.Dir+"-" {
			t.Errorf("%s needs its own dir and prefix: %+v", n, r)
		}
	}
	if r, _ := Ranks(DefaultRanks(AgentOne())).Get("high_orc"); r.Prefix != "principal-domain-" {
		t.Errorf("agent-one's words: %+v", r)
	}
	// the conservative band: 19 clean reports do not cross, 20 do; one hole in five blocks it
	a, _ := ParseASL(DefaultAscendOrc)
	if a.Cond.Eval(Metrics{Reports: 19, Passes: 19}) || !a.Cond.Eval(Metrics{Reports: 20, Passes: 20}) || a.Cond.Eval(Metrics{Reports: 20, Passes: 20, Holes: 4}) || a.Cond.Eval(Metrics{Reports: 20, Passes: 17}) {
		t.Error("the default band")
	}
}
