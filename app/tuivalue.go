package app

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Kaginari/isekai/tui"
	"github.com/Kaginari/isekai/world"
)

// footerValue fills the terminal's value readings from the value law (law.value), folded at most
// once a second (valueView): the session's meters as config resolves them, each race's standing,
// each body's meter for its court row, the throne's own score, and the nearest ascension.
func (a *App) footerValue(fv *tui.FooterView) {
	cfg := a.liveConfig()
	fv.Meters, _ = cfg.Meters()
	if a.World == nil {
		return
	}
	v := a.valueView()
	if v == nil || !v.Lit {
		return
	}
	rep, ok := v.Report.(world.ValueReport)
	if !ok {
		return
	}
	for _, r := range rep.Ranks {
		fv.Races = append(fv.Races, tui.RailRace{Rank: r.Rank, Family: r.Family, Meter: r.Reading, Bodies: r.Bodies, Confident: r.Confident, Mean: r.Mean})
	}
	fv.Values = map[string]tui.RailValue{}
	for _, b := range rep.Bodies {
		rv := tui.RailValue{Meter: b.Meter, Low: b.Confidence != "ok"}
		fv.Values[b.Body] = rv
		if b.Body == world.Rimuru {
			s := rv
			fv.Score = &s
		}
	}
	fv.Ascend = nearestAscension(rep.Bodies)
}

// nearestAscension is the body closest to its ascension, said in a line: one condition left
// ("orc-engine · reports 17/20 to ascend"), or how many are left and the nearest of them
// ("orc-engine · 2 to ascend · passRate 84/90"); "" when no body is on its way.
func nearestAscension(bodies []world.BodyValue) string {
	best, bestN, bestD := "", 0, 0.0
	var bestLeaf string
	for _, b := range bodies {
		if b.Ascension == nil || b.Ascension.Met {
			continue
		}
		n, near, d := 0, "", -1.0
		for _, l := range b.Ascension.Leaves {
			if l.Met {
				continue
			}
			n++
			if d < 0 || l.Distance < d {
				d = l.Distance
				near = leafName(l.Text) + " " + leafNum(l.Got) + "/" + leafNum(l.Want)
			}
		}
		if n == 0 {
			continue
		}
		if best == "" || n < bestN || n == bestN && d < bestD {
			best, bestN, bestD, bestLeaf = b.Body, n, d, near
		}
	}
	switch {
	case best == "":
		return ""
	case bestN == 1:
		return best + " · " + bestLeaf + " to ascend"
	}
	return fmt.Sprintf("%s · %d to ascend · %s", best, bestN, bestLeaf)
}

// leafName is a condition's left side ("reports" of "reports >= 20").
func leafName(text string) string {
	f := strings.Fields(text)
	if len(f) == 0 {
		return text
	}
	return f[0]
}

func leafNum(f float64) string { return strconv.FormatFloat(math.Round(f*10)/10, 'f', -1, 64) }
