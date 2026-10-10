package memory

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// memory review — the three-tier instrument (Nature 5 + the Memory-tiers law section).
//
// The world keeps three memories, and each answers a different question; reviewing one with
// another's lens is how stains survive. This command reads each tier with its own lens and
// routes every finding to the evolution act that tier's intelligence legitimately calls for:
//
// SHORT  — ATTENTION. The semantic caches (.isekai/memory/short/<body>.jsonl) say what bodies
//          keep ASKING. A question stem repeated across bodies is a live cross-orc current:
//          Nature 4 — if no creature owns the zone, it is a Genesis trigger (birth one);
//          if one does, the answer belongs to procedure (craft a skill under
//          .isekai/crafted-skills/, or slim the worn mind) so nobody researches it twice.
// SHARED — LESSONS. The notes buses (.isekai + machine) say what the world resolved to
//          REMEMBER. A lesson restated means the first cure died in practice: Nature 3 —
//          consolidate it into law or a base skill so it stops needing to be relearned;
//          a restated need with no owner is the second naming.
// LONG   — GENOMIC. The long index says what the world DID and what is TRUE (episodic over
//          log.md, procedural over minds/commands/tools, semantic over law/creature/canon).
//          Its signals are structural — staleness (the nag a body can only suffer through),
//          procedural thinness (the world knows what happened, not how), and recurring
//          friction families in the episode stream. Each names its act: reseed (harness
//          hygiene), craft a set of skills, or — a family recurring across time — the
//          second naming, decided aloud (genesis or a crafted skill), never silently.
//
// Review reads, never writes (`clean` is the write arm); decisions stay with the human
// (Containment, Law 6).

// ReviewFinding is one signal one tier measured, with the route its recurrence should take.
type ReviewFinding struct {
	Signal   string   `json:"signal"`
	N        int      `json:"n"`
	Evidence []string `json:"evidence"`
	Route    string   `json:"route"` // [genesis] birth a creature | [skill] craft a skill | [law] consolidate the lesson | [harness] fix the instrument
}

// ReviewTier is one memory's report block.
type ReviewTier struct {
	Name     string          `json:"name"` // short — attention | shared — lessons | long — genomic
	Reads    string          `json:"reads"`
	Facts    []string        `json:"facts"`
	Findings []ReviewFinding `json:"findings"`
}

// ReviewResult is `isekai memory review`'s full report.
type ReviewResult struct {
	Tiers []ReviewTier `json:"tiers"`
}

var reviewMu sync.Mutex // review reads caches while live runs write them; one tier at a time

// cacheStem is a question's comparable head: content tokens, ordered — two asks share their
// stem even when their stop-words and phrasing differ.
func cacheStem(q string) string {
	return strings.Join(Tokens(q), " ")
}

// reviewShort reads every body's semantic cache and asks what they keep asking.
func reviewShort(shortDir string, facts *[]string) []ReviewFinding {
	files, err := filepath.Glob(filepath.Join(shortDir, "*.jsonl"))
	if err != nil || len(files) == 0 {
		*facts = append(*facts, "no semantic cache files yet")
		return nil
	}
	type stemStat struct {
		bodies map[string]int
		total  int
	}
	stems := map[string]*stemStat{}
	junk, withHits := 0, 0
	entries := 0
	for _, p := range files {
		body := strings.TrimSuffix(filepath.Base(p), ".jsonl")
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(raw), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var e struct {
				Q       string `json:"q"`
				Results []struct {
					Src string `json:"src"`
				} `json:"results"`
			}
			if json.Unmarshal([]byte(line), &e) != nil || e.Q == "" {
				continue
			}
			entries++
			stem := cacheStem(e.Q)
			if stem != "" {
				st := stems[stem]
				if st == nil {
					st = &stemStat{bodies: map[string]int{}}
					stems[stem] = st
				}
				st.bodies[body]++
				st.total++
			}
			if len(e.Results) > 0 {
				withHits++
				allJunk := true
				for _, r := range e.Results {
					if !strings.Contains(r.Src, "log.md") && !strings.Contains(r.Src, "notes.jsonl") {
						allJunk = false
						break
					}
				}
				if allJunk {
					junk++
				}
			}
		}
	}
	*facts = append(*facts, fmt.Sprintf("%d cache entr%s across %d bodies — %d with results", entries, plural(entries), len(files), withHits))
	var out []ReviewFinding
	for stem, st := range stems {
		if len(st.bodies) >= 2 {
			out = append(out, ReviewFinding{
				Signal:   "cross-body question stem — bodies of different ranks keep asking the same thing",
				N:        st.total,
				Evidence: []string{stem, strings.Join(stemBodies(st.bodies), ", ")},
				Route:    "[genesis] if no creature owns this zone the recurring current is a birth trigger; otherwise [skill] make the answer procedural so the asking stops",
			})
		} else if st.total >= 3 {
			out = append(out, ReviewFinding{
				Signal:   "repeated question stem — one body researches the same thing every run",
				N:        st.total,
				Evidence: []string{stem, strings.Join(stemBodies(st.bodies), ", ")},
				Route:    "[skill] the answer belongs to .isekai/crafted-skills/ — a discovered answer that stays memorable costs nothing to recall",
			})
		}
	}
	if junk > 0 {
		out = append(out, ReviewFinding{
			Signal:   "junk-anchor recalls — cached answers quote only log.md/notes.jsonl, never creature docs or canon",
			N:        junk,
			Evidence: []string{filepath.Base(shortDir)},
			Route:    "[harness] the index was stale or thin — `isekai memory index` reseeds; if junks persist after, the tier is starved of procedural truth: [skill] craft what the bodies keep needing",
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].N > out[j].N })
	return out
}

func stemBodies(m map[string]int) []string {
	var bs []string
	for b, n := range m {
		bs = append(bs, fmt.Sprintf("%s×%d", b, n))
	}
	sort.Strings(bs)
	return bs
}

// reviewShared reads the notes buses — the lessons the world resolved to remember — and
// asks which lessons had to be restated.
func reviewShared(paths []string, facts *[]string) []ReviewFinding {
	type lesson struct {
		by   string
		text string
	}
	var lessons []lesson
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(raw), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var e struct {
				By   string `json:"by"`
				Text string `json:"text"`
			}
			if json.Unmarshal([]byte(line), &e) != nil || e.Text == "" {
				continue
			}
			lessons = append(lessons, lesson{by: e.By, text: e.Text})
		}
	}
	if len(lessons) == 0 {
		*facts = append(*facts, "no shared lessons on the buses")
		return nil
	}
	*facts = append(*facts, fmt.Sprintf("%d lesson%s on the shared/machine buses", len(lessons), plural(len(lessons))))
	// a restated lesson = two notes whose heads match after normalising — the first cure
	// died in practice and the world paid the same tuition twice.
	seen := map[string]lesson{}
	var out []ReviewFinding
	done := map[string]bool{}
	for _, l := range lessons {
		head := strings.ToLower(strings.Join(Tokens(l.text), " "))
		if len(head) > 60 {
			head = head[:60]
		}
		if prev, ok := seen[head]; ok && !done[head] {
			done[head] = true
			out = append(out, ReviewFinding{
				Signal:   "lesson restated — the world paid the same tuition twice",
				N:        2,
				Evidence: []string{fmt.Sprintf("%s: %.80s", prev.by, prev.text), fmt.Sprintf("%s: %.80s", l.by, l.text)},
				Route:    "[law] fold the lesson into the genome (creature trait or base-skill) so it stops needing memory to survive; if the restatement names an unowned need, it is the second naming → [genesis]",
			})
			continue
		}
		seen[head] = l
	}
	return out
}

// reviewLong reads the index as an instrument: its freshness against sources, its kind
// balance, and the friction families the episode stream carries. The episode scan runs with
// or without an index — friction in log.md exists whether or not the index over it does.
func (w *World) reviewLong(facts *[]string) []ReviewFinding {
	var out []ReviewFinding
	idx, hole := w.LoadIndex()
	if idx == nil {
		*facts = append(*facts, "@? "+hole)
		out = append(out, ReviewFinding{
			Signal: "no long-term index at all",
			N:      0,
			Route:  "[harness] rerun `isekai memory index` — a world without its genomic layer forgets its own episodes",
		})
	} else {
		epi, proc, sem := 0, 0, 0
		for _, m := range idx.Mems {
			switch m.Kind {
			case "episodic":
				epi++
			case "procedural":
				proc++
			case "semantic":
				sem++
			}
		}
		*facts = append(*facts, fmt.Sprintf("index built %s — %d memories (episodic %d · procedural %d · semantic %d) → %s", idx.BuiltAt, idx.N, epi, proc, sem, w.Rel(w.IndexPath())))
		if holes := w.IndexHoles(idx, ""); len(holes) > 0 {
			out = append(out, ReviewFinding{
				Signal:   "index stale against its sources",
				N:        len(holes),
				Evidence: holes[:min(3, len(holes))],
				Route:    "[harness] rerun `isekai memory index`; bodies only suffer this nag, nothing cures it for them",
			})
		}
		if proc == 0 && epi+sem > 0 {
			out = append(out, ReviewFinding{
				Signal: "procedural memory is empty — the world knows what happened and what is true, never how",
				N:      proc,
				Route:  "[skill] teach the how: minds (base-skills) or crafted-skills under .isekai/, one per recurring how",
			})
		}
	}
	// friction ecology: same rule deck as `clean`, read over the episode stream only — the
	// families the world's own speech keeps re-birthing.
	logPath := filepath.Join(w.Isekai(), "log.md")
	if raw, err := os.ReadFile(logPath); err == nil {
		for _, r := range ScrubRules() {
			if n := len(r.re.FindAllString(string(raw), -1)); n > 0 {
				out = append(out, ReviewFinding{
					Signal:   fmt.Sprintf("friction family %s in the episode stream", r.name),
					N:        n,
					Evidence: []string{filepath.Base(logPath)},
					Route:    "[skill/law] " + r.evolve,
				})
			}
		}
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// reviewValue reads the decision evidence lane (scenario-3-2-1): the scorecard's three feeds.
// Today-scoped per body (the funnel judges today's fresh ground): end verdicts + gate fails
// from the loop journals; decks with scored cards (ritual adoption); the tap's outcome rows.
func reviewValue(worldDir string, facts *[]string) []ReviewFinding {
	today := time.Now().UTC().Format("20060102")
	var out []ReviewFinding

	// per-body today's endings: end verdicts and gate fails
	type bodyDay struct {
		ends  map[string]int
		gates int
	}
	byBody := map[string]*bodyDay{}
	journals := 0
	dir := filepath.Join(worldDir, "instruments", "loop")
	for _, p := range listFiles(dir) {
		name := filepath.Base(p)
		if !strings.Contains(name, today) || !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		journals++
		body := strings.SplitN(name, "-20", 2)[0]
		bd := byBody[body]
		if bd == nil {
			bd = &bodyDay{ends: map[string]int{}}
			byBody[body] = bd
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(raw), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var e struct {
				T      string `json:"t"`
				Status string `json:"status"`
				Kind   string `json:"kind"`
			}
			if json.Unmarshal([]byte(line), &e) != nil {
				continue
			}
			switch {
			case e.T == "end":
				bd.ends[e.Status]++
			case e.T == "gate" && e.Kind == "" && strings.Contains(line, "\"verdict\":\"fail"):
				bd.gates++
			}
		}
	}
	*facts = append(*facts, fmt.Sprintf("today: %d journals across %d bodies", journals, len(byBody)))
	for body, bd := range byBody {
		fails := bd.ends["FAIL"] + bd.ends["ESCALATE"] + bd.ends["DENIED"]
		if fails >= 2 {
			out = append(out, ReviewFinding{
				Signal:   "body grinding today — end verdicts tilted against it",
				N:        fails,
				Evidence: []string{fmt.Sprintf("%s: FAIL×%d ESCALATE×%d DENIED×%d, gate fails %d", body, bd.ends["FAIL"], bd.ends["ESCALATE"], bd.ends["DENIED"], bd.gates)},
				Route:    "[scorecard] cards crossing this body's zone carry today's friction on their Why rows — value stays honest only if the score reads last-24h ground",
			})
		}
	}

	// ritual adoption: decks and their scored cards
	plans := filepath.Join(worldDir, "tmp", "plans")
	for _, p := range listFiles(plans) {
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var d struct {
			Items []struct {
				Score string `json:"score"`
			} `json:"items"`
		}
		if json.Unmarshal(raw, &d) == nil {
			total := len(d.Items)
			n := 0
			for _, it := range d.Items {
				if it.Score != "" {
					n++
				}
			}
			if total > 0 {
				*facts = append(*facts, fmt.Sprintf("deck %s: %d of %d cards scored", filepath.Base(p), n, total))
			}
		}
	}
	streamToday := filepath.Join(worldDir, "instruments", "stream", time.Now().UTC().Format("2006-01-02")+".jsonl")
	outcomes := 0
	if raw, err := os.ReadFile(streamToday); err == nil {
		outcomes = strings.Count(string(raw), `"type":"plan-outcome"`)
		choices := strings.Count(string(raw), `"type":"choice-answered"`)
		*facts = append(*facts, fmt.Sprintf("stream today: %d plan-outcome rows · %d choice rows", outcomes, choices))
	}
	if outcomes == 0 {
		out = append(out, ReviewFinding{
			Signal: "no plan-outcome rows yet — the funnel's learning loop has no ground below it",
			N:      0,
			Route:  "[harness] first funnel run lands them; until then, cards score off friction/outcomes/journals alone",
		})
	}
	return out
}

func listFiles(dir string) []string {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range ents {
		if !e.IsDir() {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out
}

func (w *World) Review() *ReviewResult {
	reviewMu.Lock()
	defer reviewMu.Unlock()
	res := &ReviewResult{}

	facts := []string{}
	shortFindings := reviewShort(w.shortDir(), &facts)
	res.Tiers = append(res.Tiers, ReviewTier{
		Name: "short — attention", Reads: w.Rel(w.shortDir()) + "/*.jsonl (semantic caches, per body)",
		Facts: facts, Findings: shortFindings,
	})

	facts = []string{}
	sharedFindings := reviewShared([]string{w.sharedNotes(), w.machineNotes()}, &facts)
	res.Tiers = append(res.Tiers, ReviewTier{
		Name: "shared — lessons", Reads: "notes.jsonl (world-shared + machine-shared)",
		Facts: facts, Findings: sharedFindings,
	})

	facts = []string{}
	longFindings := w.reviewLong(&facts)
	res.Tiers = append(res.Tiers, ReviewTier{
		Name: "long — genomic", Reads: "long/index.json over log.md · minds · commands · law · creatures · canon",
		Facts: facts, Findings: longFindings,
	})

	facts = []string{}
	valueFindings := reviewValue(w.Isekai(), &facts)
	res.Tiers = append(res.Tiers, ReviewTier{
		Name: "value — decisions", Reads: "today's journals (end verdicts, gate fails) · decks' scored cards · the tap's plan-outcome rows",
		Facts: facts, Findings: valueFindings,
	})
	return res
}

// printReview renders the review on the wire register: per tier, facts then routed findings.
func printReview(out io.Writer, r *ReviewResult) {
	for _, t := range r.Tiers {
		fmt.Fprintf(out, "@F tier %s (reads %s)\n", t.Name, t.Reads)
		for _, f := range t.Facts {
			fmt.Fprintf(out, "@F   %s\n", f)
		}
		for _, fd := range t.Findings {
			fmt.Fprintf(out, "@?   %s ×%d %s\n", fd.Signal, fd.N, fd.Route)
			for _, ev := range fd.Evidence[:min(2, len(fd.Evidence))] {
				fmt.Fprintf(out, "@F     %s\n", truncateRunes(ev, 140))
			}
		}
		if len(t.Findings) == 0 {
			fmt.Fprintln(out, "@F   no findings")
		}
	}
}
