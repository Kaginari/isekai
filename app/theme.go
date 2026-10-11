package app

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Kaginari/isekai/config"
	"github.com/Kaginari/isekai/theme"
	"github.com/Kaginari/isekai/tui"
)

// The look (ui.theme, ui.mode) is one choice for both faces. Whichever surface switches it —
// the terminal's /theme or /settings, the dashboard's menu, /theme typed into the dashboard's
// ask — it lands here: written to the world's config.local.yaml (machine-local, like every
// live switch), the config reloaded, and every surface told (EvTheme: the terminal follows,
// the dashboard swaps its stylesheet).

// look is the look in force.
func (a *App) look() (name, mode string) {
	cfg := a.liveConfig()
	return cfg.UI.Theme, cfg.UI.Mode
}

// setTheme keeps a look: written, reloaded, announced. Unknown names are refused with the list.
func (a *App) setTheme(name, mode string) error {
	name, mode = strings.ToLower(strings.TrimSpace(name)), strings.ToLower(strings.TrimSpace(mode))
	cur, curMode := a.look()
	if name == "" {
		name = cur
	}
	if mode == "" {
		mode = curMode
	}
	if err := theme.Check(name, mode); err != nil {
		return err
	}
	a.themeMu.Lock()
	defer a.themeMu.Unlock()
	if name != cur || mode != curMode {
		cfg := a.liveConfig()
		file := filepath.Join(a.Root, cfg.Dist.WorldDir, "config.local.yaml")
		for _, l := range cfg.Layers {
			if l.Name == "local" && l.Present && l.Path != "" {
				file = l.Path
			}
		}
		if err := cfg.ApplyPatch(file, []config.Op{{Path: "ui.theme", Value: name}, {Path: "ui.mode", Value: mode}}); err != nil {
			return err
		}
		next, _, err := cfg.Reload()
		if err != nil {
			return fmt.Errorf("written, but the reload failed: %w", err)
		}
		// the switch is its own notice on every surface: the reload's change lines stay quiet
		a.reload(next, nil)
	}
	a.announceTheme(name, mode)
	return nil
}

// announceTheme tells every surface: through the host when one is attached (the terminal and
// the dashboard), else straight onto the dashboard's stream.
func (a *App) announceTheme(name, mode string) {
	ev := tui.EvTheme{Name: name, Mode: mode}
	a.mu.Lock()
	w := a.web
	a.mu.Unlock()
	if w == nil {
		return
	}
	w.mu.Lock()
	h := w.host
	w.mu.Unlock()
	if h != nil {
		go h.send(ev) // the terminal may be the caller: never block its update
		return
	}
	w.Send(ev)
}

// themeSlash is /theme [name] [dark|light|system] where the terminal's own handler is not in the
// way (the REPL, the dashboard's ask): alone it lists the palettes.
func (a *App) themeSlash(rest string, out, errw io.Writer) {
	args := strings.Fields(strings.ToLower(rest))
	if len(args) == 0 || args[0] == "list" {
		fmt.Fprint(out, a.liveConfig().ExplainTheme())
		return
	}
	name, mode := "", ""
	for _, x := range args {
		if x == "dark" || x == "light" || x == "system" {
			mode = x
		} else {
			name = x
		}
	}
	if err := a.setTheme(name, mode); err != nil {
		fmt.Fprintln(errw, "theme: "+err.Error())
		return
	}
	n, m := a.look()
	fmt.Fprintf(out, "theme %s · %s — the terminal and the dashboard wear it (config.local.yaml)\n", n, m)
}

// SetTheme is the terminal's switch (tui.ThemeHost).
func (h *tuiHost) SetTheme(name, mode string) error { return h.a.setTheme(name, mode) }
