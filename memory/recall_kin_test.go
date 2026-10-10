package memory

import (
	"strings"
	"testing"
	"time"
)

// kin-first recall: a body's own creature doc and its dated thoughts outrank the chronicle's
// mass — and another race's doc still beats log.md. Census drift 2026-10-03: bodies re-derived
// context from the chronicle because kin sources were never on top.
func TestRecallKinFirst(t *testing.T) {
	root := t.TempDir()
	isekaiDir := root + "/.isekai"
	w := &World{Root: root, Home: root, Now: time.Now}
	// an elf with a doc and dated thoughts carrying the task's terms
	writeFile(isekaiDir+"/elf/isekai/README.md",
		"# elf-isekai\n\n- **Rank:** Elf\n- **Territory:** the whole court\n\n## Thoughts\n\n- census territory onto mapping lands through isowns\n")
	writeFile(isekaiDir+"/elf/isekai/thoughts/2026-10-03-census-brief.md",
		"# census brief\n\nThe onto census derives isowns territory claims from the structured Territory field; the write-gate's Owner() reads the same structure. Land the derivation in onto/world.go.\n")
	// the chronicle drags matching terms by mass alone
	writeFile(isekaiDir+"/log.md",
		"# Chronicle\n\n### [2026-10-02] rimuru — census territory chatter\n- census territory onto isowns mentioned at length census territory onto isowns census territory\n")
	writeFile(isekaiDir+"/isekai.md", "# law\n")
	if _, err := w.Index(); err != nil {
		t.Fatalf("index: %v", err)
	}
	res, err := w.Recall("how does the onto census land isowns territory", RecallOpts{As: "elf-isekai", K: 3, NoCache: true})
	if err != nil {
		t.Fatalf("recall: %v", err)
	}
	if len(res.Results) == 0 {
		t.Fatal("no results")
	}
	top := res.Results[0]
	if !strings.Contains(top.Src, "elf/isekai") {
		t.Fatalf("kin-first rule: the elf's own territory must top the recall, got %+v", res.Results)
	}
	// the chronicle only surfaces when kin has nothing to say
	res2, _ := w.Recall("how does the onto census land isowns territory", RecallOpts{As: "slime-app-commands", K: 5, NoCache: true})
	logRank := -1
	for i, r := range res2.Results {
		if strings.HasSuffix(r.Src, "log.md") {
			logRank = i
		}
	}
	if logRank == 0 {
		t.Errorf("the chronicle must never be the highest-ranked anchor while kin memory exists, got %+v", res2.Results)
	}
}
