package config

import (
	"strings"
	"testing"
)

// Trace export is outward (Nature 7): off by default, on only from the machine's own config with
// an endpoint named, and never a credential in a file.
func TestOTLPExportIsTheMachinesWord(t *testing.T) {
	c := newWorld(t, nil).load()
	if o := c.Instruments.OTLP; o.Enabled || o.Endpoint != "" || len(o.HeadersEnv) != 0 {
		t.Fatalf("the shipped export is off with nothing named: %+v", o)
	}
	cases := []struct{ name, file, yaml, env, want string }{
		{"enabled from the project", ".isekai/config.yaml", "instruments: {otlp: {enabled: true, endpoint: \"https://collector.example/v1/traces\"}}", "", "never the project config"},
		{"endpoint from the project, enabled locally", ".isekai/config.yaml", "instruments: {otlp: {endpoint: \"https://collector.example/v1/traces\"}}", "instruments: {otlp: {enabled: true}}", "never the project config"},
		{"no endpoint", ".isekai/config.local.yaml", "instruments: {otlp: {enabled: true}}", "", "needs the collector's http(s) traces URL"},
		{"not a URL", ".isekai/config.local.yaml", "instruments: {otlp: {enabled: true, endpoint: \"collector:4318\"}}", "", "needs the collector's http(s) traces URL"},
		{"a header value", ".isekai/config.local.yaml", "instruments: {otlp: {headersEnv: {api_key: \"sk-live-123\"}}}", "", "must name an environment variable"},
		{"a literal key", ".isekai/config.local.yaml", "instruments: {otlp: {apiKey: abc}}", "", "a credential never lives in a config file"},
		{"literal headers", ".isekai/config.local.yaml", "instruments: {otlp: {headers: {Authorization: \"Bearer x\"}}}", "", "a credential never lives in a config file"},
	}
	for _, cs := range cases {
		files := map[string]string{cs.file: cs.yaml + "\n"}
		if cs.env != "" {
			files[".isekai/config.local.yaml"] = cs.env + "\n"
		}
		err := newWorld(t, files).loadErr()
		if err == nil || !strings.Contains(err.Error(), cs.want) {
			t.Errorf("%s: got %v, want %q", cs.name, err, cs.want)
		}
	}
	// env is not the human's written word either
	w := newWorld(t, map[string]string{".isekai/config.local.yaml": "instruments: {otlp: {endpoint: \"https://collector.example/v1/traces\"}}\n"})
	w.env["ISEKAI_CONFIG_CONTENT"] = "instruments: {otlp: {enabled: true}}"
	if err := w.loadErr(); err == nil || !strings.Contains(err.Error(), "never the project config, env or a flag") {
		t.Errorf("enabled from env: %v", err)
	}
	// the machine's own layer: accepted, header values by name
	c = newWorld(t, map[string]string{".isekai/config.local.yaml": "instruments: {otlp: {enabled: true, endpoint: \"https://collector.example/v1/traces\", headersEnv: {api_key: TRACE_KEY}}}\n"}).load()
	if o := c.Instruments.OTLP; !o.Enabled || o.HeadersEnv["api_key"] != "TRACE_KEY" {
		t.Errorf("local export: %+v", o)
	}
}
