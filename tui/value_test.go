package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/Kaginari/isekai/value"
)

func meterOf(def value.MeterDef, v float64, basis bool) value.Meter {
	return value.Render(def, v, basis)
}

// The sidebar wears the value law: each race's meter (label, reading, a bar in its band)
// confident races first by their mean, a race with a short record dimmed and unranked ("low"),
// a custom race in its family's hue; a court row carries its body's reading; the vitals carry
// the throne's score and the nearest ascension.
func TestRailValueMeters(t *testing.T) {
	th := depthTheme(true)
	landed := value.MeterDef{Signal: "landRate", Label: "Landed", Scale: "percent", Bands: value.Bands{OK: 80, Warn: 50}}
	gate := value.MeterDef{Signal: "gateRight", Label: "Gate right", Scale: "percent", Bands: value.Bands{OK: 90, Warn: 70}}
	asks := value.MeterDef{Signal: "askDoneRate", Label: "Asks done", Scale: "percent", Bands: value.Bands{OK: 80, Warn: 50}}
	v := RailView{World: "isekai", CtxKnown: true, CtxPct: 40,
		Layers: []RailLayer{{Rank: "rimuru", Tokens: 9000}, {Rank: "slime", Tokens: 3000}},
		Races: []RailRace{
			{Rank: "slime", Family: "author", Meter: meterOf(landed, 62, true), Bodies: 3, Confident: 2, Mean: 61},
			{Rank: "orc", Family: "gate", Meter: meterOf(gate, 95, true), Bodies: 1, Confident: 1, Mean: 93},
			{Rank: "scribe", Family: "gate", Meter: meterOf(gate, 0, false), Bodies: 1, Confident: 0},
			{Rank: "kijin", Family: "steward", Bodies: 0},
		},
		Score:  &RailValue{Meter: meterOf(asks, 88, true)},
		Ascend: "orc-engine · reports 17/20 to ascend",
		Courts: []RailCourt{{Name: "slime-auth", State: "thinking", Value: &RailValue{Meter: meterOf(landed, 45, true)}}},
	}
	rows := railTexts(th.railRows(v, 48))
	plain := ansi.Strip(strings.Join(rows, "\n"))
	for _, want := range []string{"asks done", "88%", "↑ orc-engine · reports 17/20 to ascend", "gate right", "95%", "landed", "62%", "low", "45%"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("the sidebar misses %q:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "kijin") {
		t.Fatal("a race with no body and no spend is listed")
	}
	// order: confident by mean (orc 93, slime 61), then the low and the spend-only
	at := func(s string) int { return strings.Index(plain, "● "+s) }
	if !(at("orc") < at("slime") && at("slime") < at("scribe") && at("slime") < at("rimuru")) {
		t.Fatalf("the races' order:\n%s", plain)
	}
	// the bands' colours: orc's reading ok, slime's warn, the court's 45% bad
	joined := strings.Join(rows, "\n")
	for name, hue := range map[string]string{"95%": th.pal.OK, "62%": th.pal.Warn, "45%": th.pal.Danger} {
		if !strings.Contains(joined, sgrFg(hexRGB(hue).col())) {
			t.Errorf("%s is not in its band's colour %s", name, hue)
		}
	}
	// a custom race wears its family's hue: scribe (gate) the orc's lane
	if got, want := th.raceHue("scribe", "gate"), th.lane("orc"); toRGB(got) != toRGB(want) {
		t.Errorf("scribe wears %s, want the orc's %s", toRGB(got).hex(), toRGB(want).hex())
	}
}

// The context gauge bands by config's meter, not by numbers of its own: with ui.meters.ctx
// banding 30% as bad, a 40% gauge is drawn in the danger colour.
func TestCtxGaugeBandsFromConfig(t *testing.T) {
	th := depthTheme(true)
	strict := map[string]value.MeterDef{"ctx": {Signal: "ctx", Label: "Context", Scale: "percent", Direction: "down", Bands: value.Bands{OK: 10, Warn: 20}}}
	rows := ansi.Strip(strings.Join(railTexts(th.railRows(RailView{World: "w", CtxKnown: true, CtxPct: 40, Meters: strict}, 40)), "\n"))
	if !strings.Contains(rows, "40%") {
		t.Fatal(rows)
	}
	raw := strings.Join(railTexts(th.railRows(RailView{World: "w", CtxKnown: true, CtxPct: 40, Meters: strict}, 40)), "\n")
	if !strings.Contains(raw, sgrFg(hexRGB(th.pal.Danger).col())) {
		t.Fatal("a 40% context under a 10/20 law is not drawn as danger")
	}
	calm := strings.Join(railTexts(th.railRows(RailView{World: "w", CtxKnown: true, CtxPct: 40}, 40)), "\n")
	if strings.Contains(calm, sgrFg(hexRGB(th.pal.Danger).col())) {
		t.Fatal("a 40% context under the default law is danger")
	}
}
