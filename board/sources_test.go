package board

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Kaginari/isekai/onto"
)

var fixedNow = time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)

func fixtureRoot(t *testing.T, name string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func fixtureSources(t *testing.T, name string) Sources {
	return FileSources(fixtureRoot(t, name), ".isekai", onto.Layout{}, func() time.Time { return fixedNow })
}

func TestUsageReadAndRollup(t *testing.T) {
	s := fixtureSources(t, "world")
	recs, rd := s.Usage()
	if !rd.Lit {
		t.Fatalf("usage should be lit: %s", rd)
	}
	if len(recs) != 5 {
		t.Fatalf("want 5 records (one bad line skipped), got %d", len(recs))
	}
	if !strings.Contains(rd.Why, "skipped: 1") {
		t.Errorf("the skipped line must be named: %q", rd.Why)
	}
	since, rng := RangeSince("7d", fixedNow)
	r := Rollup(recs, since, rng)
	if r.Records != 5 || r.Total.Calls != 5 {
		t.Errorf("total calls %d", r.Total.Calls)
	}
	if r.Total.Tokens() != 1200+300+20000+5000+800+900+500+100+2000+400+30000+1000+700+1500 {
		t.Errorf("total tokens %d", r.Total.Tokens())
	}
	if r.Total.Unpriced != 1 {
		t.Errorf("one call is unpriced (usd null), got %d", r.Total.Unpriced)
	}
	if len(r.ByDay) != 2 || r.ByDay[0].Key != "2026-09-25" {
		t.Errorf("by day: %+v", r.ByDay)
	}
	if r.ByBody[0].Key != "rimuru" {
		t.Errorf("biggest body should be rimuru: %+v", r.ByBody)
	}
	if len(r.Sessions) != 2 {
		t.Errorf("sessions: %v", r.Sessions)
	}
	day, _ := RangeSince("24h", fixedNow)
	if today := Rollup(recs, day, "24h"); today.Records != 2 {
		t.Errorf("24h should hold today's 2 calls, got %d", today.Records)
	}
	if all, _ := RangeSince("all", fixedNow); !all.IsZero() {
		t.Error("all = zero time")
	}
}

func TestUsageSilentOnEmptyWorld(t *testing.T) {
	s := fixtureSources(t, "empty")
	recs, rd := s.Usage()
	if len(recs) != 0 || rd.Lit {
		t.Fatalf("empty world must be silent, not zero: %+v", rd)
	}
	if !strings.HasPrefix(rd.String(), "never lit") {
		t.Errorf("reading: %s", rd)
	}
	if _, rd := s.Loop(); rd.Lit {
		t.Error("loop should be silent")
	}
	if _, rd := s.Log(); rd.Lit {
		t.Error("log should be silent")
	}
	if c := s.Colony(); c.Reading.Lit {
		t.Error("colony should be silent with no creatures")
	}
	if m := s.Memory(); m.Reading.Lit {
		t.Error("memory should be silent")
	}
	if tb := s.Toolbox(); tb.Reading.Lit {
		t.Error("toolbox should be silent")
	}
}

func TestLogParse(t *testing.T) {
	s := fixtureSources(t, "world")
	entries, rd := s.Log()
	if !rd.Lit || len(entries) != 2 {
		t.Fatalf("log: %d entries, %s", len(entries), rd)
	}
	if entries[0].At != "2026-09-25T18:30:00+02:00" || !strings.HasPrefix(entries[0].Who, "court-body") {
		t.Errorf("newest first: %+v", entries[0])
	}
	if !strings.Contains(entries[0].Body, "token TTL is 15m") {
		t.Errorf("body kept: %q", entries[0].Body)
	}
}

func TestLoopRuns(t *testing.T) {
	s := fixtureSources(t, "world")
	runs, rd := s.Loop()
	if !rd.Lit || len(runs) != 1 {
		t.Fatalf("runs: %+v %s", runs, rd)
	}
	if runs[0].Status != "PASS" || runs[0].As != "slime-auth" || runs[0].Steps != 1 {
		t.Errorf("run: %+v", runs[0])
	}
}

func TestColonyGraph(t *testing.T) {
	s := fixtureSources(t, "world")
	c := s.Colony()
	if !c.Reading.Lit {
		t.Fatalf("colony: %s", c.Reading)
	}
	byID := map[string]Node{}
	for _, n := range c.Nodes {
		byID[n.ID] = n
	}
	if byID["slime-auth"].Rank != "slime" || byID["orc-core"].Rank != "orc" || byID["elf-voice"].Rank != "elf" {
		t.Errorf("ranks: %+v", byID)
	}
	if byID["mind-slime-auth-mind"].Lane != "zone" {
		t.Errorf("slime-auth-mind lane: %q", byID["mind-slime-auth-mind"].Lane)
	}
	if byID["mind-tdd"].Lane != "shared" || !byID["mind-tdd"].Shared {
		t.Errorf("tdd (host skill, no race prefix) is shared: %+v", byID["mind-tdd"])
	}
	if byID["slime-auth"].Doc != ".isekai/slime/auth/README.md" {
		t.Errorf("doc: %q", byID["slime-auth"].Doc)
	}
	want := map[string]bool{"slime-auth truth orc-core": false, "orc-core verdict elf-voice": false, "elf-voice reports rimuru": false, "slime-auth wears mind-tdd": false}
	for _, e := range c.Edges {
		k := e.From + " " + e.Bond + " " + e.To
		if _, ok := want[k]; ok {
			want[k] = true
		}
		if e.Bond == "above" {
			t.Errorf("transitive closure leaked into the drawing: %+v", e)
		}
	}
	for k, ok := range want {
		if !ok {
			t.Errorf("missing edge %s", k)
		}
	}
	if c.Lanes[0] != "mind:zone" || c.Lanes[len(c.Lanes)-1] != "mind:shared" {
		t.Errorf("lanes: %v", c.Lanes)
	}
}

func TestMemoryAndToolboxViews(t *testing.T) {
	s := fixtureSources(t, "world")
	m := s.Memory()
	if !m.Reading.Lit {
		t.Fatalf("memory: %s", m.Reading)
	}
	if len(m.Notes) != 3 || m.ByKind["colony"] != 1 || m.ByKind["territory"] != 1 || m.ByKind["untyped"] != 1 {
		t.Errorf("notes by kind: %v (%d notes)", m.ByKind, len(m.Notes))
	}
	if m.Notes[0].Text != "an untyped note" {
		t.Errorf("newest first: %+v", m.Notes[0])
	}
	if len(m.Desks) == 0 {
		t.Error("the fixture mind has a desk")
	}
	tb := s.Toolbox()
	if !tb.Reading.Lit {
		t.Fatalf("toolbox: %s", tb.Reading)
	}
	if !tb.Offered["tdd"] || !tb.Loaded["tdd"] || tb.Loaded["slime-auth-mind"] {
		t.Errorf("offered %v loaded %v", tb.Offered, tb.Loaded)
	}
	if len(tb.Entries) < 2 {
		t.Errorf("registry entries: %d", len(tb.Entries))
	}
}

func TestDocIsFenced(t *testing.T) {
	s := fixtureSources(t, "world")
	if _, err := s.Doc(".isekai/slime/auth/README.md"); err != nil {
		t.Errorf("world doc readable: %v", err)
	}
	if _, err := s.Doc(".claude/skills/tdd/SKILL.md"); err != nil {
		t.Errorf("mind doc readable: %v", err)
	}
	for _, bad := range []string{"../../go.mod", ".isekai/../.isekai/../sources.go", "/etc/passwd", ".isekai/name"} {
		if _, err := s.Doc(bad); err == nil {
			t.Errorf("%q must be refused", bad)
		}
	}
}

func TestColonyIsOneTreeUnderTheThrone(t *testing.T) {
	c := fixtureSources(t, "world").Colony()
	byID := map[string]Node{}
	for _, n := range c.Nodes {
		byID[n.ID] = n
	}
	throne, ok := byID[c.Throne]
	if c.Throne == "" || !ok || !throne.Throne || throne.Kind != "creature" {
		t.Fatalf("the throne must be a node: %q %+v", c.Throne, throne)
	}
	for _, n := range c.Nodes {
		if n.Kind != "creature" || n.Throne {
			if n.Parent != "" {
				t.Errorf("%s (a %s) has a parent %q", n.ID, n.Kind, n.Parent)
			}
			continue
		}
		seen := map[string]bool{}
		at := n.ID
		for at != c.Throne {
			if seen[at] {
				t.Fatalf("%s loops at %s", n.ID, at)
			}
			seen[at] = true
			p, ok := byID[byID[at].Parent]
			if !ok {
				t.Fatalf("%s: %s's parent %q is not a node", n.ID, at, byID[at].Parent)
			}
			at = p.ID
		}
	}
	if byID["slime-auth"].Parent != "orc-core" || byID["orc-core"].Parent != "elf-voice" || byID["elf-voice"].Parent != c.Throne {
		t.Errorf("the asserted bonds make the tree: %q %q %q", byID["slime-auth"].Parent, byID["orc-core"].Parent, byID["elf-voice"].Parent)
	}
	// an orphan hangs under the throne by an implied edge, and a note says so
	implied := 0
	for _, e := range c.Edges {
		if e.Implied {
			implied++
			if e.Bond != "reports" || e.To != c.Throne || byID[e.From].Parent != c.Throne {
				t.Errorf("implied edge: %+v", e)
			}
		}
	}
	if implied == 0 || !strings.Contains(strings.Join(c.Notes, "\n"), "drawn under "+c.Throne) {
		t.Errorf("the fixture's orphans must hang under the throne, said: %d %v", implied, c.Notes)
	}
	// a mind in the world's own base-skills is a node, and its wearer is bonded to it
	if byID["mind-slime-auth-craft"].Kind != "mind" {
		t.Fatalf("base-skills mind missing: %+v", byID["mind-slime-auth-craft"])
	}
	wears := false
	for _, e := range c.Edges {
		wears = wears || (e.From == "slime-auth" && e.Bond == "wears" && e.To == "mind-slime-auth-craft")
	}
	if !wears {
		t.Error("slime-auth wears slime-auth-craft (its Minds field) — no wears edge")
	}
}

func TestTreeUpAddsAMissingThrone(t *testing.T) {
	c := &Colony{Nodes: []Node{{ID: "elf-a", Kind: "creature", Rank: "elf"}, {ID: "orc-b", Kind: "creature", Rank: "orc"}, {ID: "orc-c", Kind: "creature", Rank: "orc"}, {ID: "m", Kind: "mind"}},
		Edges: []Edge{{From: "orc-b", To: "elf-a", Bond: "verdict"}, {From: "elf-a", To: "rimuru", Bond: "reports"}, {From: "orc-c", To: "orc-c", Bond: "above"}, {From: "orc-b", To: "m", Bond: "wears"}}}
	treeUp(c)
	var throne *Node
	for i := range c.Nodes {
		if c.Nodes[i].ID == c.Throne {
			throne = &c.Nodes[i]
		}
	}
	if throne == nil || !throne.Throne || c.Throne != "rimuru" {
		t.Fatalf("no throne node added: %+v", c.Nodes)
	}
	parents := map[string]string{}
	for _, n := range c.Nodes {
		parents[n.ID] = n.Parent
	}
	if parents["orc-b"] != "elf-a" || parents["elf-a"] != "rimuru" || parents["orc-c"] != "rimuru" || parents["m"] != "" {
		t.Errorf("parents: %v", parents)
	}
}

func TestActivityByDay(t *testing.T) {
	s := fixtureSources(t, "world")
	runs, _ := s.Loop()
	recs, _ := s.Usage()
	a := BuildActivity(runs, recs, fixedNow, ActivityDays)
	if len(a.Days) != ActivityDays || a.Days[ActivityDays-1] != fixedNow.UTC().Format("2006-01-02") {
		t.Fatalf("days: %v", a.Days)
	}
	sa := a.Bodies["slime-auth"]
	if sa == nil || sa.Total != 1 || sa.Last == "" {
		t.Fatalf("slime-auth's run: %+v", sa)
	}
	i := -1
	for k, d := range a.Days {
		if d == "2026-09-25" {
			i = k
		}
	}
	if i < 0 || sa.Runs[i] != 1 || sa.Tokens[i] != 800+900+500+100 {
		t.Errorf("slime-auth on 2026-09-25: runs %v tokens %v", sa.Runs, sa.Tokens)
	}
	if r := a.Bodies["rimuru"]; r == nil || r.Total != 0 || r.Tokens[i] == 0 {
		t.Errorf("rimuru spends tokens without a journaled run here: %+v", r)
	}
	if _, ok := a.Bodies["orc-core"]; ok {
		t.Error("a body the instruments never saw must be absent, not zero")
	}
}
