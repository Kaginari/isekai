package board

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Kaginari/isekai/value"
)

// pngish is a stand-in portrait: the resolver reads bytes, it never decodes them.
var pngish = []byte("\x89PNG\r\n\x1a\nworld")

// The one portrait order: the world's own file (any spelling) before config's before the
// embedded race portrait; nothing at all reads nil.
func TestPortraitResolutionOrder(t *testing.T) {
	root := t.TempDir()
	emb, _ := Portrait("dark-elf")
	if b, src := ResolvePortrait(root, ".isekai", "dark_elf", ""); !bytes.Equal(b, emb) || src != "embedded" {
		t.Fatalf("the embedded portrait answers any spelling: %q", src)
	}
	_ = os.MkdirAll(filepath.Join(root, "art"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "art", "watch.png"), []byte("configured"), 0o644)
	if b, src := ResolvePortrait(root, ".isekai", "dark-elf", "art/watch.png"); string(b) != "configured" || src != "art/watch.png" {
		t.Fatalf("ranks.<r>.portrait comes before the embedded one: %q", src)
	}
	_ = os.MkdirAll(filepath.Join(root, ".isekai", "portraits"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".isekai", "portraits", "darkelf.png"), pngish, 0o644)
	if b, src := ResolvePortrait(root, ".isekai", "DarkElf", "art/watch.png"); !bytes.Equal(b, pngish) || src != ".isekai/portraits/darkelf.png" {
		t.Fatalf("the world's own portrait (an old spelling) comes first: %q", src)
	}
	if b, _ := ResolvePortrait(root, ".isekai", "night-watch", ""); b != nil {
		t.Fatal("a custom race with no portrait has none")
	}
	if names := strings.Join(PortraitNames(), ","); names != "dark-elf,elf,high-elf,high-orc,kijin,orc,rimuru,slime" {
		t.Fatalf("the embedded portraits are spelled the one way: %s", names)
	}
}

// The value law reaches the board: /api/value and /api/meters answer from the integrator, a
// custom race's node carries its family's glyph and its value, /portraits/<rank> resolves.
func TestValueAPIsAndIcons(t *testing.T) {
	b := newBoard(t, "empty", Options{})
	code, body := get(t, b, "/api/value")
	if code != 200 || !strings.Contains(body, `"lit":false`) {
		t.Fatalf("no integrator, a silent value API: %d %s", code, body)
	}
	var m struct {
		Session map[string]value.MeterDef `json:"session"`
	}
	if code, body := get(t, b, "/api/meters"); code != 200 || json.Unmarshal([]byte(body), &m) != nil || m.Session["ctx"].Direction != "down" || m.Session["score"].Bands.OK == 0 {
		t.Fatalf("the shipped meters: %d %s", code, body)
	}
	rep := map[string]any{"bodies": []any{}}
	view := &ValueView{Lit: true, Window: 7 * 24 * time.Hour, Def: value.DefaultDef(), Report: rep,
		Meters: map[string]value.MeterDef{"ctx": {Signal: "ctx", Direction: "down", Bands: value.Bands{OK: 50, Warn: 75}}},
		Bodies: map[string]NodeValue{"rimuru": {Score: 88, Confidence: "ok", Band: "ok", Family: "root"}}}
	b2 := newBoard(t, "empty", Options{Sources: Sources{
		Value: func() *ValueView { return view },
		Ranks: func() []RankMeta {
			return []RankMeta{{Rank: "rimuru", Family: "root"}, {Rank: "night_watch", Family: "auditor"}}
		},
	}})
	if code, body := get(t, b2, "/api/value"); code != 200 || !strings.Contains(body, `"lit":true`) || !strings.Contains(body, `"night-watch"`) {
		t.Fatalf("the integrator's value: %d %s", code, body)
	}
	if code, body := get(t, b2, "/api/meters"); code != 200 || !strings.Contains(body, `"warn":75`) {
		t.Fatalf("the meters are config's: %d %s", code, body)
	}
	var c struct {
		Nodes []Node `json:"nodes"`
	}
	_, body = get(t, b2, "/api/colony")
	if err := json.Unmarshal([]byte(body), &c); err != nil {
		t.Fatal(err)
	}
	if len(c.Nodes) != 1 || c.Nodes[0].Icon == nil || c.Nodes[0].Icon.Kind != "portrait" || c.Nodes[0].Icon.URL != "portraits/rimuru" || c.Nodes[0].Value == nil || c.Nodes[0].Value.Score != 88 || c.Nodes[0].Family != "root" {
		t.Fatalf("the throne's node carries its icon, family and value: %+v", c.Nodes)
	}
	if ic := b2.iconFor("NightWatch"); ic.Kind != "family" || ic.URL != "web/icons/icons.svg#family-auditor" {
		t.Fatalf("a custom race with no portrait wears its family's glyph: %+v", ic)
	}
	if ic := b2.iconFor("ghost"); ic.Kind != "generic" {
		t.Fatalf("a rank nobody knows wears the generic glyph: %+v", ic)
	}
	if code, body := get(t, b2, "/portraits/dark_elf"); code != 200 || !strings.HasPrefix(body, "\x89PNG") {
		t.Fatalf("/portraits/<rank> resolves any spelling: %d", code)
	}
	if code, body := get(t, b2, "/portraits/night-watch"); code != 200 || !strings.HasPrefix(body, "<svg") || !strings.Contains(body, "#8A8F98") {
		t.Fatalf("a custom race's portrait is its family's glyph: %d %s", code, body)
	}
	// the activity window and the failure word come from law.value
	if b2.activityDays() != 7 {
		t.Fatalf("the activity window is law.value.window: %d", b2.activityDays())
	}
	// the icon sprite ships with its licences and its source line
	for _, f := range []string{"web/icons/icons.svg", "web/icons/LICENSE-lucide", "web/icons/LICENSE-tabler-icons", "web/icons/SOURCE"} {
		if _, err := content.ReadFile(f); err != nil {
			t.Errorf("%s is embedded: %v", f, err)
		}
	}
	sprite, _ := content.ReadFile("web/icons/icons.svg")
	for _, id := range []string{"mind", "family-root", "family-planner", "family-gate", "family-author", "family-steward", "family-auditor", "family-generic"} {
		if !strings.Contains(string(sprite), `id="`+id+`"`) {
			t.Errorf("the sprite lacks #%s", id)
		}
	}
}

// Mass is the code a creature owns: its own lines (the files whose deepest owner it is), its
// subtree's summed up the tree, the throne the whole world; tests and generated code left out.
func TestMasses(t *testing.T) {
	root := t.TempDir()
	w := func(p, s string) {
		_ = os.MkdirAll(filepath.Dir(filepath.Join(root, p)), 0o755)
		_ = os.WriteFile(filepath.Join(root, p), []byte(s), 0o644)
	}
	w("go.mod", "module x\n")
	w("auth/login/login.go", "package login\n\nfunc A() {}\n")
	w("auth/session.go", "package auth\n")
	w("auth/login/login_test.go", "package login\n\n\n\n\n")
	w("auth/gen.go", "// Code generated by x. DO NOT EDIT.\npackage auth\n\n\n")
	c := &Colony{Nodes: []Node{
		{ID: "rimuru", Kind: "creature", Throne: true},
		{ID: "orc-auth", Kind: "creature", Parent: "rimuru", Owns: []string{"auth"}},
		{ID: "slime-login", Kind: "creature", Parent: "orc-auth", Owns: []string{"auth/login"}},
		{ID: "mind-x", Kind: "mind"},
	}, Edges: []Edge{{From: "orc-auth", To: "mind-x", Bond: "wears"}, {From: "slime-login", To: "mind-x", Bond: "wears"}}}
	sourceMemo.root = "" // a fresh walk
	Masses(root, c)
	by := map[string]Node{}
	for _, n := range c.Nodes {
		by[n.ID] = n
	}
	if by["slime-login"].Mass != 3 || by["orc-auth"].Own != 1 || by["orc-auth"].Mass != 4 || by["rimuru"].Mass != 4 || by["mind-x"].Mass != 2 {
		t.Fatalf("masses: %+v", c.Nodes)
	}
}

// Every creature wears its face at every depth (a slime too), the 3D overlay draws it as a badge
// through the one resolver, and the focus card shows the race's image large. The page's code is
// read as text: the rule is that no depth gate stands between a node and its icon.
func TestFacesAtEveryDepth(t *testing.T) {
	net, _ := content.ReadFile("web/components/net/net.js")
	if strings.Contains(string(net), "p.d <= 2") || !strings.Contains(string(net), "if (icon) {") {
		t.Fatal("the 2D net draws every creature's icon, whatever its depth")
	}
	n3, _ := content.ReadFile("web/components/net/net3d.js")
	if !strings.Contains(string(n3), "'portraits/' + encodeURIComponent(rank)") {
		t.Fatal("the 3D net's badges come from /portraits/<rank>")
	}
	dash, _ := content.ReadFile("web/dash.js")
	if !strings.Contains(string(dash), "kcard-face") || !strings.Contains(string(dash), "kcard-portrait") || !strings.Contains(string(dash), "kcard-glyph") {
		t.Fatal("the focus card shows the race's portrait, else its family's glyph")
	}
	for _, old := range []string{"web/portraits/' + n.rank", "PORTRAITS", "BRAIN"} {
		if strings.Contains(string(net), old) || strings.Contains(string(n3), old) {
			t.Errorf("%q is gone: the resolver and the sprite are the one source", old)
		}
	}
}
