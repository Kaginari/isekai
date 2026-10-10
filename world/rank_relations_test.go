package world

import (
	"context"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/config"
	"github.com/Kaginari/isekai/loop"
	"github.com/Kaginari/isekai/provider/mock"
)

// The relations suite reads the coded bonds between ranks — the tree itself, never a doc's
// prose. Fixtures map the law's own table (DefaultRanks) and one hand-cut hierarchy.

func TestRelateTheThreeBonds(t *testing.T) {
	rs := Ranks{
		{Name: "elf", ReportsTo: Rimuru, Layer: 1},
		{Name: "elf2", ReportsTo: Rimuru, Layer: 1},
		{Name: "orc", ReportsTo: "elf", Layer: 2},
		{Name: "high_orc", ReportsTo: "elf", Base: "orc", Layer: 2},
		{Name: "slime", ReportsTo: "orc", Layer: 3},
		{Name: "slime2", ReportsTo: "orc", Layer: 3},
	}
	cases := []struct {
		a, b string
		want Relation
	}{
		{"elf", Rimuru, RelChild},      // an elf reports up to the throne
		{Rimuru, "elf", RelParent},     // the throne is the elf's parent
		{"orc", "slime", RelParent},    // an orc is one hop above its slime
		{"slime", "orc", RelChild},     // a slime one hop below its orc
		{"high_orc", "orc", RelDivine}, // a high orc ascends its base
		{"orc", "high_orc", RelDivine}, // the divine bond reads both ways
		{"elf", "elf2", RelNeighbor},   // one configured layer shares the neighbor bond
		{"slime", "slime2", RelNeighbor},
		{"orc", "orc", RelNeighbor}, // a rank is its own neighbor
		{"elf", "slime", RelNone},   // two hops apart and no layer — no bond
		{"elf", "orc", RelParent},   // one hop, despite different layers
		{"slime", "elf2", RelNone},  // neither parent nor layer-mate
		{"", "orc", RelNone},        // the empty name bonds nothing
		{"orc", "", RelNone},
		{"ghost", "orc", RelNone}, // an unknown rank bonds nothing
	}
	for _, c := range cases {
		if got := rs.relate(c.a, c.b); got != c.want {
			t.Errorf("relate(%q,%q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
	// a bond is symmetric in the pair, asymmetric in the role
	if rs.relate("elf", "orc") != RelParent || rs.relate("orc", "elf") != RelChild {
		t.Error("parent/child invert across the pair")
	}
}

func TestNeighborRanksSharesLayer(t *testing.T) {
	layered := Ranks{
		{Name: "elf", ReportsTo: Rimuru, Layer: 1},
		{Name: "elf2", ReportsTo: Rimuru, Layer: 1},
		{Name: "orc", ReportsTo: "elf", Layer: 2},
		{Name: "orc2", ReportsTo: "elf", Layer: 2},
	}
	if got := layered.NeighborRanks("orc"); strings.Join(got, ",") != "orc,orc2" {
		t.Errorf("orc's layer-mates: %v", got)
	}
	if got := layered.NeighborRanks("elf"); strings.Join(got, ",") != "elf,elf2" {
		t.Errorf("elf's layer-mates: %v", got)
	}
	if got := layered.NeighborRanks(Rimuru); got != nil {
		t.Errorf("the throne has no layer: %v", got)
	}
	if got := layered.NeighborRanks("ghost"); got != nil {
		t.Errorf("an unknown rank knows no one: %v", got)
	}

	// layers unset: the neighbors are the ranks reporting to the same parent
	reportBased := Ranks{
		{Name: "orc", ReportsTo: "elf"},
		{Name: "high_orc", ReportsTo: "elf", Base: "orc"},
		{Name: "slime", ReportsTo: "orc"},
	}
	got := reportBased.NeighborRanks("orc")
	if strings.Join(got, ",") != "orc,high_orc" {
		t.Errorf("report-based layer-mates: %v", got)
	}
	// the slime shares its report-line with no one else
	if got := reportBased.NeighborRanks("slime"); strings.Join(got, ",") != "slime" {
		t.Errorf("the slime's only layer-mate is itself: %v", got)
	}
}

func TestDirectChildOneHopDown(t *testing.T) {
	rs := Ranks(DefaultRanks(Isekai()))
	// down one hop
	if !rs.DirectChild(Rimuru, "elf") {
		t.Error("the throne's one hop is the elf")
	}
	if !rs.DirectChild("elf", "orc") {
		t.Error("elf → orc is one hop down")
	}
	if !rs.DirectChild("orc", "slime") {
		t.Error("orc → slime is one hop down")
	}
	// two hops down is refused: the orc commands the slime, not the elf
	if rs.DirectChild("elf", "slime") {
		t.Error("elf → slime skips a layer; the strict flow refuses it")
	}
	if rs.DirectChild(Rimuru, "orc") {
		t.Error("the throne never reaches past the elf to an orc")
	}
	// up is never a child
	if rs.DirectChild("orc", "elf") || rs.DirectChild("slime", "orc") {
		t.Error("work never flows up the ranks")
	}
	// the divine bond: a base rank's ascended self sits at its own layer — the base may
	// reach its ascended, for the ascended reports to the same parent as the base's own
	// parent (high_orc reports to elf, one hop under the elf, same as orc)
	if !rs.DirectChild("elf", "high_orc") {
		t.Error("elf → high_orc is one hop down: the ascended orc still answers to the elf")
	}
}

// Strict flow is coded at the dispatch gate, not begged in a prompt: when it is on, the throne
// reaching straight down to an orc is refused mid-turn, the exact skip the law forbids. Off,
// the same dispatch lands (the law's wording only). Defaults ship it on so the ladder holds
// without anyone asking.
func TestStrictFlowRefusesThroneToOrcSkip(t *testing.T) {
	// the default holds the ladder by machine: no per-test opt-out needed
	def, err := config.Defaults("isekai")
	if err != nil || !def.Tools.Dispatch.StrictFlow {
		t.Fatalf("strictFlow must default true so the ladder is held by the machine; config.yaml may opt out with tools.dispatch.strictFlow: false (err=%v strictFlow=%v)", err, def != nil && def.Tools.Dispatch.StrictFlow)
	}
	_, w := customWorld(t)
	m := mock.New(
		mock.Call("d1", "dispatch", map[string]interface{}{"body": "keeper-core", "ask": "hold the gate", "office": "raphael"}),
		mock.Text("@S DONE\n@U colony x\n@E 0"),
	)
	b := build(m)
	b.Court.StrictFlow = true
	e := w.Engine(Rimuru, b)
	r, err := e.Run(context.Background(), "x")
	if err != nil || r.Status != loop.Done {
		t.Fatalf("%v %+v", err, r)
	}
	got := m.Requests[1].Messages[2].ToolResults[0]
	if !got.IsError || !strings.Contains(got.Content, "one layer at a time") || !strings.Contains(got.Content, "rimuru's children") {
		t.Fatalf("rimuru→orc skip must be refused and name the one hop down: %+v", got)
	}

	// off: the same dispatch lands (the law is wording only)
	m2 := mock.New(
		mock.Call("d1", "dispatch", map[string]interface{}{"body": "keeper-core", "ask": "hold the gate", "office": "raphael"}),
		mock.Text("@S DONE\n@V held\n@U colony x\n@E 0"),
		mock.Text("@S DONE\n@U colony x\n@E 0"),
	)
	b2 := build(m2)
	b2.Court.StrictFlow = false
	e2 := w.Engine(Rimuru, b2)
	r2, err2 := e2.Run(context.Background(), "x")
	if err2 != nil || r2.Status != loop.Done {
		t.Fatalf("%v %+v", err2, r2)
	}
	if len(m2.Requests) != 3 {
		t.Fatalf("strict flow off: parent + child + close = 3 requests, got %d", len(m2.Requests))
	}
	if m2.Requests[2].Messages[2].ToolResults[0].IsError { // the third call closes the turn, carrying the dispatch's tool result
		t.Fatal("strict flow off: the skip lands")
	}

	// DefaultOff (for now): six app/bench suites mock skip-level dispatches and must migrate to
	// one-hop flows before the default can flip. Until then the law is opt-in via
	// tools.dispatch.strictFlow — the coded refusal above stands whatever the default.
}

func TestDivineRulesReachTheBase(t *testing.T) {
	rs := Ranks{
		{Name: "orc", ReportsTo: "elf", HoldsGate: true},
		{Name: "high_orc", ReportsTo: "elf", Base: "orc",
			DivineRules: []string{"hold the gate even when pressed", "a reverted pass is a fail"}},
	}
	got := rs.DivineRulesFor("orc")
	if len(got) != 2 || got[0] != "hold the gate even when pressed" {
		t.Errorf("the base hears its ascended's rules: %v", got)
	}
	if got := rs.DivineRulesFor("high_orc"); len(got) != 0 {
		t.Errorf("the ascended does not hear its own rules echoed back: %v", got)
	}
	if got := rs.DivineRulesFor("elf"); len(got) != 0 {
		t.Errorf("a rank with no ascended hears nothing: %v", got)
	}
}

func TestAscensionsWalksTheBonds(t *testing.T) {
	rs := Ranks(DefaultRanks(Isekai()))
	// the default table wires both divine ascensions coded on the intelligence band
	got := rs.Ascensions()
	if len(got) != 2 {
		t.Fatalf("the default table wires high_elf and high_orc on intelligence: %v", got)
	}
	byFrom := map[string]Ascension{}
	for _, a := range got {
		byFrom[a.From] = a
	}
	if byFrom["elf"].When != "ascend elf when intelligence >= 70 && reports >= 5" {
		t.Fatalf("high_elf ascendsWhen = %q, want the intelligence band", byFrom["elf"].When)
	}
	if byFrom["orc"].When != "ascend orc when intelligence >= 70 && reports >= 5" {
		t.Fatalf("high_orc ascendsWhen = %q, want the intelligence band", byFrom["orc"].When)
	}
	// a custom ascendsWhen still carries through
	rs2 := append(Ranks(nil), rs...)
	for i, r := range rs2 {
		if r.Name == "high_orc" {
			rs2[i].AscendsWhen = "ascend orc when reports >= 3"
		}
	}
	got = rs2.Ascensions()
	byFrom = map[string]Ascension{}
	for _, a := range got {
		byFrom[a.From] = a
	}
	if byFrom["orc"].When != "ascend orc when reports >= 3" {
		t.Fatalf("edited orc bond = %q, want reports >= 3", byFrom["orc"].When)
	}
}
