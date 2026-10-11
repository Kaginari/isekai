package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A look switched anywhere is kept in the world's config.local.yaml, the config reloaded, the
// dashboard's stream told; an unknown palette is refused with the list and writes nothing; the
// dashboard's ask can switch it with /theme too.
func TestThemeKeptAndAnnounced(t *testing.T) {
	w := newTestWorld(t, "isekai", isekaiCreatures())
	w.write(".isekai/config.yaml", "models: {default: mock/m}\nproviders: {mock: {enabled: true}}\nui: {board: {autostart: false}}\n")
	a := w.open()
	if n, m := a.look(); n != "cyberpunk" || m != "dark" {
		t.Fatalf("the default look: %s %s", n, m)
	}
	bus := a.webBus()
	replay, ch, cancel := bus.Subscribe()
	defer cancel()
	_ = replay
	if err := a.setTheme("synthwave", "light"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(w.root, ".isekai", "config.local.yaml"))
	if err != nil || !strings.Contains(string(b), "theme: synthwave") || !strings.Contains(string(b), "mode: light") {
		t.Fatalf("config.local.yaml: %v\n%s", err, b)
	}
	if n, m := a.look(); n != "synthwave" || m != "light" || !strings.Contains(a.liveConfig().Where("ui.theme"), "config.local.yaml") {
		t.Fatalf("reloaded: %s %s", n, m)
	}
	if ev := string(<-ch); !strings.Contains(ev, `"type":"theme"`) || !strings.Contains(ev, `"name":"synthwave"`) || !strings.Contains(ev, `"mode":"light"`) {
		t.Fatalf("the stream: %s", ev)
	}
	if err := a.setTheme("vaporwave", ""); err == nil || !strings.Contains(err.Error(), "chromecore") {
		t.Fatalf("an unknown palette names the list: %v", err)
	}
	var out, errw strings.Builder
	a.themeSlash("glitch", &out, &errw)
	if n, m := a.look(); n != "glitch" || m != "light" || !strings.Contains(out.String(), "theme glitch · light") {
		t.Fatalf("/theme glitch: %s %s %q %q", n, m, out.String(), errw.String())
	}
	out.Reset()
	a.themeSlash("", &out, &errw)
	if !strings.Contains(out.String(), "▸ glitch") {
		t.Fatalf("/theme alone lists: %q", out.String())
	}
}
