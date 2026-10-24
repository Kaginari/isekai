package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/Kaginari/isekai/config"
	"github.com/Kaginari/isekai/loop"
)

// otlpCollector is an OTLP/HTTP endpoint on loopback: it keeps every span and header it is sent.
type otlpCollector struct {
	*httptest.Server
	mu      sync.Mutex
	spans   []otlpSpan
	headers []http.Header
	posts   int
}

func newCollector(t *testing.T, status int) *otlpCollector {
	c := &otlpCollector{}
	c.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ResourceSpans []struct {
				ScopeSpans []struct{ Spans []otlpSpan }
			}
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &req)
		c.mu.Lock()
		c.posts++
		c.headers = append(c.headers, r.Header.Clone())
		for _, rs := range req.ResourceSpans {
			for _, ss := range rs.ScopeSpans {
				c.spans = append(c.spans, ss.Spans...)
			}
		}
		c.mu.Unlock()
		w.WriteHeader(status)
	}))
	t.Cleanup(c.Close)
	return c
}

func (s otlpSpan) attr(k string) string {
	for _, a := range s.Attributes {
		if a.Key == k {
			return a.Value["stringValue"] + a.Value["intValue"]
		}
	}
	return ""
}

// openTraced opens the world with the given env, the board off, no TTY.
func openTraced(w *testWorld, env map[string]string) (*App, *strings.Builder) {
	w.t.Helper()
	errb := &strings.Builder{}
	a, err := New(Options{Dist: w.dist, Root: w.root, Cwd: w.root, Home: w.home, Env: func(k string) string {
		if k == "HOME" {
			return w.home
		}
		return env[k]
	}, In: strings.NewReader(""), Out: &strings.Builder{}, Err: errb, Quiet: true, IsTTY: func() bool { return false }, NoBoard: true})
	if err != nil {
		w.t.Fatal(err)
	}
	return a, errb
}

// TestTracerShipsTurnsCourtsAndCalls: one CHAIN per turn named rank · body, the dispatch is the
// court's AGENT span and the court's turn its child in the same trace, one TOOL span per executed
// call under its turn — a denied call never ran and leaves none; headers come from env by name.
func TestTracerShipsTurnsCourtsAndCalls(t *testing.T) {
	col := newCollector(t, http.StatusOK)
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.script(".isekai/tmp/session.json",
		when("map the zone", call("dispatch", map[string]string{"body": "elf-core", "ask": "draft on the readme: read it"})),
		call("bash", map[string]string{"command": "echo hi"}),
		text("@S DONE\n@F the court read it\n@E 40"),
	)
	w.script(".isekai/tmp/elf.json",
		when("draft on the readme", call("read", map[string]string{"path": "README.md"})),
		text("@S DONE read\n@U colony nothing more\n@E 40"),
	)
	w.write(".isekai/config.yaml", "models: {default: m1/mount, offices: {ciel: m3/elf}}\nproviders:\n  m1: {type: mock, script: .isekai/tmp/session.json}\n  m3: {type: mock, script: .isekai/tmp/elf.json}\ntools: {bash: {sandbox: none}}\nui: {board: {autostart: false}}\n")
	w.write(".isekai/config.local.yaml", "permissions: {rules: [{match: \"bash:echo*\", action: deny}]}\ninstruments: {otlp: {enabled: true, endpoint: \""+col.URL+"/v1/traces\", headersEnv: {api_key: TRACE_KEY}}}\n")
	a, _ := openTraced(w, map[string]string{"TRACE_KEY": "k1"})
	e, err := a.Engine()
	if err != nil {
		t.Fatal(err)
	}
	if r, err := e.Run(context.Background(), "map the zone"); err != nil || r.Status != loop.Done {
		t.Fatalf("turn: %v %+v", err, r)
	}
	a.Close() // flushes the exporter

	col.mu.Lock()
	defer col.mu.Unlock()
	if len(col.headers) == 0 || col.headers[0].Get("api_key") != "k1" || col.headers[0].Get("Content-Type") != "application/json" {
		t.Fatalf("headers: %v", col.headers)
	}
	by := map[string]otlpSpan{}
	for _, s := range col.spans {
		by[s.attr("openinference.span.kind")+" "+s.Name] = s
		if s.attr("tool.name") == "bash" {
			t.Errorf("a denied call left a span: %+v", s)
		}
	}
	throne, court, elf, read := by["CHAIN "+chainName(a, "rimuru")], by["AGENT court · elf-core"], by["CHAIN elf · elf-core"], by["TOOL read"]
	if throne.SpanID == "" || court.SpanID == "" || elf.SpanID == "" || read.SpanID == "" || len(col.spans) != 4 {
		t.Fatalf("spans: %+v", col.spans)
	}
	if court.ParentSpanID != throne.SpanID || elf.ParentSpanID != court.SpanID || read.ParentSpanID != elf.SpanID {
		t.Errorf("tree: throne %s court %s←%s elf %s←%s read ←%s", throne.SpanID, court.SpanID, court.ParentSpanID, elf.SpanID, elf.ParentSpanID, read.ParentSpanID)
	}
	if court.TraceID != throne.TraceID || elf.TraceID != throne.TraceID || read.TraceID != throne.TraceID {
		t.Errorf("one trace per throne turn: %s %s %s %s", throne.TraceID, court.TraceID, elf.TraceID, read.TraceID)
	}
	if elf.attr("isekai.rank") != "elf" || elf.attr("isekai.body") != "elf-core" || throne.attr("isekai.status") != loop.Done || !strings.Contains(read.attr("input.value"), "README.md") {
		t.Errorf("attributes: elf %+v throne %+v read %+v", elf.Attributes, throne.Attributes, read.Attributes)
	}
}

func chainName(a *App, body string) string {
	if r := a.World.RankOf(body).Name; r != "" {
		return r + " · " + body
	}
	return body
}

// TestTracerOffShipsNothing: export is off by default and with no endpoint — not a byte leaves.
func TestTracerOffShipsNothing(t *testing.T) {
	if NewTracer(config.OTLP{Endpoint: "http://127.0.0.1:1/v1/traces"}, "isekai", nil, nil, nil) != nil ||
		NewTracer(config.OTLP{Enabled: true}, "isekai", nil, nil, nil) != nil {
		t.Fatal("a tracer without enabled and an endpoint")
	}
	var off *Tracer // every method is a no-op on nil
	off.State(nil, "thinking")
	off.Observe(nil, nil, "start")
	off.Ended(nil, nil)
	off.Close()
	col := newCollector(t, http.StatusOK)
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.script(".isekai/tmp/s.json", text("@S DONE\n@E 40"))
	w.write(".isekai/config.yaml", "models: {default: mock/m}\nproviders: {mock: {enabled: true, script: .isekai/tmp/s.json}}\nui: {board: {autostart: false}}\n")
	w.write(".isekai/config.local.yaml", "instruments: {otlp: {endpoint: \""+col.URL+"\"}}\n")
	a, _ := openTraced(w, nil)
	e, err := a.Engine()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.Run(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	a.Close()
	col.mu.Lock()
	defer col.mu.Unlock()
	if a.Tracer != nil || col.posts != 0 {
		t.Errorf("export off, yet %d posts", col.posts)
	}
}

// TestTracerWarnsOncePerClassOnTheNoticeLine: a refusing collector is one notice on the session's
// notice line (never stderr) however many batches it refuses; each other failure class — an
// unreachable collector, an unset header variable — speaks once too.
func TestTracerWarnsOncePerClassOnTheNoticeLine(t *testing.T) {
	col := newCollector(t, http.StatusUnauthorized)
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.script(".isekai/tmp/s.json", text("@S DONE one\n@E 40"), text("@S DONE two\n@E 40"))
	w.write(".isekai/config.yaml", "models: {default: mock/m}\nproviders: {mock: {enabled: true, script: .isekai/tmp/s.json}}\nui: {board: {autostart: false}}\n")
	w.write(".isekai/config.local.yaml", "instruments: {otlp: {enabled: true, endpoint: \""+col.URL+"/v1/traces?key=s3cret\"}}\n")
	a, errb := openTraced(w, nil)
	var mu sync.Mutex
	var notes []string
	a.Notify = func(s string) { mu.Lock(); notes = append(notes, s); mu.Unlock() }
	e, err := a.Engine()
	if err != nil {
		t.Fatal(err)
	}
	for _, ask := range []string{"one", "two"} {
		if _, err := e.Run(context.Background(), ask); err != nil {
			t.Fatal(err)
		}
		a.Tracer.post([]otlpSpan{{Name: ask}}) // a batch of its own per turn
	}
	a.Close()
	mu.Lock()
	if len(notes) != 1 || !strings.Contains(notes[0], "401") || strings.Contains(notes[0], "s3cret") {
		t.Errorf("notices: %q", notes)
	}
	mu.Unlock()
	if strings.Contains(errb.String(), "traces:") {
		t.Errorf("a tracer warning reached stderr: %q", errb.String())
	}

	// with no live surface the warning is a hole on the report, still never stderr
	dead := httptest.NewServer(http.NotFoundHandler())
	dead.Close()
	var got []string
	tr := NewTracer(config.OTLP{Enabled: true, Endpoint: dead.URL, HeadersEnv: map[string]string{"api_key": "TRACE_KEY"}}, "isekai", func(string) string { return "" }, nil, func(s string) { got = append(got, s) })
	for i := 0; i < 3; i++ {
		tr.post([]otlpSpan{{Name: "x"}})
	}
	if len(got) != 2 || !strings.Contains(got[0], "TRACE_KEY is not set") || !strings.Contains(got[1], "did not answer") {
		t.Errorf("unset + unreachable, once each: %q", got)
	}
	b := &App{}
	b.notice("traces: x")
	if h := b.Holes(); len(h) != 1 || h[0] != "traces: x" {
		t.Errorf("headless notice: %v", h)
	}
}
