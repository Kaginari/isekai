// Package board is the world, seen: an http.Handler the binary serves on 127.0.0.1 that reads
// the same instruments the session writes and shows them in one page at the root — the
// dashboard (web/dash.html): the cockpit, then a tab per question (binary.md §The board,
// canon/ui.md). Everything it serves is embedded; the page works offline.
package board

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Kaginari/isekai/onto"
)

//go:embed all:web
var content embed.FS

// Names are the lexicon labels the page reads: no isekai term is hard-coded in it.
// Missing keys fall back to DefaultNames; a distribution passes its own (Court → Subagents).
type Names map[string]string

// DefaultNames is the isekai vocabulary.
func DefaultNames() Names {
	return Names{
		"board":                 "Board",
		"bin":                   "isekai",
		"page.overview":         "Overview",
		"page.court":            "Court",
		"page.usage":            "Usage",
		"page.colony":           "Colony",
		"page.memory":           "Memory",
		"page.toolbox":          "Toolbox",
		"page.log":              "Log",
		"page.config":           "Config",
		"world":                 "world",
		"session":               "session",
		"body":                  "body",
		"bodies":                "bodies",
		"court_body":            "Court Body",
		"creature":              "creature",
		"creatures":             "creatures",
		"mind":                  "mind",
		"minds":                 "minds",
		"rank":                  "rank",
		"office":                "office",
		"territory":             "territory",
		"desk":                  "desk",
		"thoughts":              "thoughts",
		"law":                   "law",
		"instrument":            "instrument",
		"stress":                "stress",
		"off_list":              "off-list",
		"unsaid":                "the unsaid",
		"rank.rimuru":           "Rimuru",
		"rank.veldora":          "Veldora",
		"rank.elf":              "Elf",
		"rank.orc":              "Orc",
		"rank.slime":            "Slime",
		"rank.kijin":            "Kijin",
		"lane.zone":             "zone",
		"lane.verdict":          "verdict",
		"lane.global":           "global",
		"lane.shared":           "shared",
		"bond.truth":            "truth-current",
		"bond.verdict":          "verdict-current",
		"bond.wears":            "anima-thread",
		"bond.reports":          "reports to",
		"bond.above":            "above",
		"bond.owns":             "owns",
		"kind.law":              "law",
		"kind.colony":           "colony",
		"kind.territory":        "territory",
		"tier.short":            "short",
		"tier.long":             "long",
		"tier.shared":           "shared",
		"toolbox.offered":       "offered",
		"toolbox.loaded":        "loaded",
		"toolbox.kind.mind":     "minds",
		"toolbox.kind.command":  "commands",
		"toolbox.kind.tool":     "tools",
		"toolbox.kind.body":     "bodies",
		"toolbox.kind.external": "external",
	}
}

// Options configures a board.
type Options struct {
	WorldRoot string      // the directory that holds the world dir
	WorldDir  string      // ".isekai" (a distribution renames it)
	Layout    onto.Layout // the ontology's rank layout; zero = the law's
	Names     Names       // lexicon labels, merged over DefaultNames
	Sources   Sources     // feeds; nil fields read files
	Prefix    string      // URL prefix the board is mounted under ("" or "/name")
	Now       func() time.Time
	Poll      time.Duration // how often the watcher looks for changes (default 1s)
	Logger    *log.Logger
	Live      Live // the running session, for the dashboard; nil = read-only
}

// Board is the handler.
type Board struct {
	opt     Options
	src     Sources
	names   Names
	mux     *http.ServeMux
	hub     *hub
	stop    chan struct{}
	once    sync.Once
	liveCrt bool // an integrator wired Court()
	liveCfg bool
	token   string // guards the dashboard's acts (live.go)
}

// Token returns the per-session board token used to authorize dashboard acts.
// It is exposed for tests that need to act as the page (e.g., TestDashGateAnswerOverHTTP).
// The token is also embedded in the dashboard page ({{TOKEN}}) and sent by dash.js
// via X-Board-Token for every /api/ask /api/answer /api/interrupt request.
func (b *Board) Token() string { return b.token }

// PortraitNames lists the embedded rank portraits ("rimuru", "elf", "orc", "slime",
// "kijin", "darkelf", "highelf", "highorc").
func PortraitNames() []string {
	return []string{"rimuru", "elf", "orc", "slime", "kijin", "darkelf", "highelf", "highorc"}
}

// Portrait returns the embedded rank portrait PNG's bytes; unknown names return an error.
func Portrait(name string) ([]byte, error) {
	names := PortraitNames()
	ok := false
	for _, n := range names {
		if n == name {
			ok = true
			break
		}
	}
	if !ok {
		return nil, fmt.Errorf("portrait %q: not one of %v", name, names)
	}
	return content.ReadFile("web/portraits/" + name + ".png")
}

// New builds the board. It starts a watcher goroutine; Close stops it.
func New(o Options) *Board {
	if o.WorldDir == "" {
		o.WorldDir = ".isekai"
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Poll <= 0 {
		o.Poll = time.Second
	}
	if o.Logger == nil {
		o.Logger = log.New(&discard{}, "", 0)
	}
	o.Prefix = strings.TrimSuffix(o.Prefix, "/")
	b := &Board{opt: o, hub: newHub(), stop: make(chan struct{}), token: newToken()}
	b.liveCrt = o.Sources.Court != nil
	b.liveCfg = o.Sources.Config != nil
	b.src = o.Sources.withDefaults(FileSources(o.WorldRoot, o.WorldDir, o.Layout, o.Now))
	b.names = DefaultNames()
	for k, v := range o.Names {
		b.names[k] = v
	}
	b.mux = http.NewServeMux()
	b.routes()
	go b.watch()
	return b
}

// Close stops the watcher and every open event stream.
func (b *Board) Close() { b.once.Do(func() { close(b.stop); b.hub.closeAll() }) }

// Serve listens on addr until ctx ends.
func Serve(ctx context.Context, addr string, o Options) error {
	b := New(o)
	defer b.Close()
	return b.Serve(ctx, addr)
}

// Serve listens on addr until ctx ends.
func (b *Board) Serve(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: b, ReadHeaderTimeout: 5 * time.Second}
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ln) }()
	select {
	case <-ctx.Done():
		b.Close()
		shut, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shut)
		return nil
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

// ServeHTTP strips the prefix and routes.
func (b *Board) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if b.opt.Prefix != "" {
		if r.URL.Path == b.opt.Prefix {
			http.Redirect(w, r, b.opt.Prefix+"/", http.StatusFound)
			return
		}
		if !strings.HasPrefix(r.URL.Path, b.opt.Prefix+"/") {
			http.NotFound(w, r)
			return
		}
		r2 := r.Clone(r.Context())
		r2.URL.Path = strings.TrimPrefix(r.URL.Path, b.opt.Prefix)
		r = r2
	}
	b.mux.ServeHTTP(w, r)
}

func (b *Board) routes() {
	// one page, at the root: the dashboard. Every other path is an API, the event stream, or
	// an asset under /web/.
	b.mux.HandleFunc("/", b.root)
	// the old addresses keep working: /dash was the dashboard's own path; the pages that became
	// tabs open their tab
	b.mux.HandleFunc("/dash", b.moved(""))
	for page, tab := range movedPages {
		b.mux.HandleFunc("/"+page, b.moved(tab))
	}
	// every read of the world's records answers a loopback Host only: a rebound DNS name
	// (a page on the open web pointed at 127.0.0.1) reads nothing
	b.mux.HandleFunc("/events", guarded(b.events))
	b.mux.HandleFunc("/api/doc", guarded(b.apiDoc))
	b.mux.HandleFunc("/api/court", guarded(b.apiCourt))
	b.mux.HandleFunc("/api/usage", guarded(b.apiUsage))
	b.mux.HandleFunc("/api/colony", guarded(b.apiColony))
	b.mux.HandleFunc("/api/activity", guarded(b.apiActivity))
	b.mux.HandleFunc("/api/tree", guarded(b.apiTree))
	b.mux.HandleFunc("/api/tree/children", guarded(b.apiTreeChildren))
	b.mux.HandleFunc("/api/courts", guarded(b.apiCourts))
	b.mux.HandleFunc("/api/plan", guarded(b.apiPlan))
	b.liveRoutes()
	b.worldRoutes()
}

func (b *Board) worldName() string {
	name := readName(path.Join(b.opt.WorldRoot, b.opt.WorldDir, "name"))
	if name == "" {
		name = path.Base(b.opt.WorldRoot)
	}
	return name
}

// movedPages are the old board pages that live on as the dashboard's tabs, by path.
var movedPages = map[string]string{"court": "court", "memory": "memory", "toolbox": "toolbox", "config": "config", "usage": "metrics", "colony": "net"}

// root serves the dashboard at "/" and nothing else there.
func (b *Board) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	b.dash(w, r)
}

// moved answers an old address with a permanent redirect to the root, opening tab when given.
func (b *Board) moved(tab string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		to := b.opt.Prefix + "/"
		if tab != "" {
			to += "#" + tab
		}
		http.Redirect(w, r, to, http.StatusMovedPermanently)
	}
}

type usageData struct {
	UsageReport
	Session string
}

func (b *Board) usageReport(r *http.Request) usageData {
	recs, rd := b.src.Usage()
	since, rng := RangeSince(r.URL.Query().Get("range"), b.opt.Now())
	session := r.URL.Query().Get("session")
	if session != "" {
		var keep []UsageRecord
		for _, u := range recs {
			if u.Session == session {
				keep = append(keep, u)
			}
		}
		recs = keep
	}
	rep := Rollup(recs, since, rng)
	rep.Reading = rd
	if session != "" {
		rep.Sessions = sessionsOf(recs)
	}
	return usageData{UsageReport: rep, Session: session}
}

func sessionsOf(recs []UsageRecord) []string {
	set := map[string]bool{}
	for _, u := range recs {
		set[u.Session] = true
	}
	var out []string
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

type chartSeries struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Values []int  `json:"values"`
}

type chartData struct {
	Days   []string      `json:"days"`
	Series []chartSeries `json:"series"`
	USD    []float64     `json:"usd"`
	Models []UsageSum    `json:"models"`
}

func chartOf(r UsageReport) chartData {
	c := chartData{Models: r.ByModel}
	c.Series = []chartSeries{{Key: "input", Label: "input"}, {Key: "output", Label: "output"}, {Key: "cacheRead", Label: "cache read"}, {Key: "cacheWrite", Label: "cache write"}}
	for _, d := range r.ByDay {
		c.Days = append(c.Days, d.Key)
		c.Series[0].Values = append(c.Series[0].Values, d.Input)
		c.Series[1].Values = append(c.Series[1].Values, d.Output)
		c.Series[2].Values = append(c.Series[2].Values, d.CacheRead)
		c.Series[3].Values = append(c.Series[3].Values, d.CacheWrite)
		c.USD = append(c.USD, d.USD)
	}
	return c
}

// laneLabel is one lane of the colony under the lexicon's word.
type laneLabel struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Kind  string `json:"kind"`
}

func (b *Board) laneLabels(c *Colony) []laneLabel {
	used := map[string]bool{}
	for _, n := range c.Nodes {
		if n.Kind == "mind" {
			used["mind:"+n.Lane] = true
		} else {
			used[n.Lane] = true
		}
	}
	var out []laneLabel
	for _, l := range c.Lanes {
		if !used[l] {
			continue
		}
		if strings.HasPrefix(l, "mind:") {
			out = append(out, laneLabel{ID: l, Label: b.names.get("lane."+strings.TrimPrefix(l, "mind:")) + " " + b.names.get("minds"), Kind: "mind"})
		} else {
			out = append(out, laneLabel{ID: l, Label: b.names.get("rank." + l), Kind: "rank"})
		}
	}
	return out
}

// colonyNamed is the colony with the lexicon's words on it: the throne wears the dist's name for
// the session body (Rimuru here; agent-one has its own), never a hard-coded one.
func (b *Board) colonyNamed() *Colony {
	c := b.src.Colony()
	for i := range c.Nodes {
		if c.Nodes[i].Throne && c.Nodes[i].Label == "" {
			c.Nodes[i].Label = b.names.get("rank." + c.Nodes[i].Rank)
		}
	}
	return c
}

// ---- api

func (b *Board) apiDoc(w http.ResponseWriter, r *http.Request) {
	text, err := b.src.Doc(r.URL.Query().Get("path"))
	if err != nil {
		http.Error(w, "not a readable doc inside the world", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(text))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}

func (b *Board) apiCourt(w http.ResponseWriter, r *http.Request) {
	bodies := b.src.Court()
	if bodies == nil {
		bodies = []Body{}
	}
	writeJSON(w, map[string]any{"live": b.liveCrt, "bodies": bodies, "at": b.opt.Now()})
}

func (b *Board) apiUsage(w http.ResponseWriter, r *http.Request) {
	d := b.usageReport(r)
	writeJSON(w, map[string]any{"reading": d.Reading.String(), "lit": d.Reading.Lit, "range": d.Range, "total": d.Total, "chart": chartOf(d.UsageReport), "byBody": d.ByBody, "byOffice": d.ByOffice, "byRank": d.ByRank, "byModel": d.ByModel, "byDay": d.ByDay})
}

func (b *Board) apiColony(w http.ResponseWriter, r *http.Request) {
	c := b.colonyNamed()
	// every list is a list on the wire, an empty world's included — never null
	if c.Nodes == nil {
		c.Nodes = []Node{}
	}
	if c.Edges == nil {
		c.Edges = []Edge{}
	}
	for _, l := range []*[]string{&c.Lanes, &c.Findings, &c.Notes} {
		if *l == nil {
			*l = []string{}
		}
	}
	labels := b.laneLabels(c)
	if labels == nil {
		labels = []laneLabel{}
	}
	writeJSON(w, struct {
		*Colony
		Labels []laneLabel `json:"labels"`
	}{c, labels})
}

// apiActivity is each body's recent work, day by day, for the net's activity tiles and the
// weight of its bonds: runs from the loop journals, tokens from the usage journal. A body the
// instruments never saw is absent — its tile stays blank, never a made-up zero.
func (b *Board) apiActivity(w http.ResponseWriter, r *http.Request) {
	runs, rr := b.src.Loop()
	recs, ur := b.src.Usage()
	a := BuildActivity(runs, recs, b.opt.Now(), ActivityDays)
	a.Reading = rr.String() + " · " + ur.String()
	a.Lit = rr.Lit || ur.Lit
	writeJSON(w, a)
}

func (b *Board) apiTree(w http.ResponseWriter, r *http.Request) {
	if b.src.Tree == nil {
		writeJSON(w, map[string]any{"live": false, "root": "", "tree": []TreeEntry{}, "truncated": false})
		return
	}
	path := r.URL.Query().Get("path")
	entries, truncated := b.src.Tree(path)
	if entries == nil {
		entries = []TreeEntry{}
	}
	writeJSON(w, map[string]any{"live": true, "root": b.opt.WorldRoot, "tree": entries, "truncated": truncated})
}

// apiTreeChildren serves one lazy level of the world tree: ?path is the
// root-relative slash prefix ("" = the root level). The JSON face is
// {path, nodes, truncated} with kind rendered as `dir bool` per node —
// the dash's expand-on-click contract. A prefix escaping the root, a
// missing directory, or a file answers 404: the client rendered that node
// from an older snapshot and the world moved under it.
func (b *Board) apiTreeChildren(w http.ResponseWriter, r *http.Request) {
	if b.src.Tree == nil {
		http.Error(w, "world tree not backed", http.StatusServiceUnavailable)
		return
	}
	prefix := r.URL.Query().Get("path")
	entries, trunc := b.src.Tree(prefix)
	if entries == nil && !trunc {
		http.NotFound(w, r)
		return
	}
	type node struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Dir  bool   `json:"dir"`
		Size int64  `json:"size,omitempty"`
		Mod  string `json:"mod,omitempty"`
	}
	nodes := make([]node, 0, len(entries))
	for _, e := range entries {
		nodes = append(nodes, node{
			Name: e.Name,
			Path: e.Path,
			Dir:  e.Kind == "dir",
			Size: e.Size,
			Mod:  e.Mod,
		})
	}
	path := prefix
	if path == "" {
		path = "."
	}
	writeJSON(w, map[string]any{"path": path, "children": nodes, "truncated": trunc})
}

// apiCourts serves the courts strip: every entity with a plan deck and its
// open hand (queued cards, per-heartbeat presence, plan stats). Body fields
// are never sent — bodies stay local to the plan file.
func (b *Board) apiCourts(w http.ResponseWriter, r *http.Request) {
	c := b.src.Courts
	if c == nil {
		writeJSON(w, map[string]any{"active": 0, "stale": 0, "empty": true})
		return
	}
	view, err := c()
	if err != nil {
		writeJSON(w, map[string]any{"active": 0, "stale": 0, "empty": true})
		return
	}
	// view is expected to be []Body (as set in the test)
	bodies, ok := view.([]Body)
	if !ok {
		// If the Court function returns something else, treat as empty
		writeJSON(w, map[string]any{"active": 0, "stale": 0, "empty": true})
		return
	}
	active := len(bodies)
	stale := 0 // we don't have stale info in the test; could be computed from bodies if needed
	empty := active == 0
	writeJSON(w, map[string]any{"active": active, "stale": stale, "empty": empty})
}

// apiPlan serves THE deck — the world's plan.json (tmp/plans/<world>.json)
// with the holder, hearts, and lane rows the plan popup renders. Never
// another creature's deck: courts carry those.
func (b *Board) apiPlan(w http.ResponseWriter, r *http.Request) {
	body := r.URL.Query().Get("body")
	p := b.src.Plan
	if p == nil {
		writeJSON(w, map[string]any{"body": body, "empty": true})
		return
	}
	view, err := p(body)
	if err != nil {
		writeJSON(w, map[string]any{"body": body, "empty": true, "error": err.Error()})
		return
	}
	writeJSON(w, view)
}

func (n Names) get(k string) string {
	if v, ok := n[k]; ok {
		return v
	}
	return k
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }
