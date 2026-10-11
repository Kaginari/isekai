package tui

import "github.com/Kaginari/isekai/theme"

// The terminal's palettes are the binary's (package theme): named schemes as data, role → hex
// for a dark and a light variant, shared with the dashboard. This file only names them in the
// terminal's words; the one constructor that reads a palette's roles is NewThemeFrom.

// Palette, Roles and Swatch are the shared types.
type (
	Palette = theme.Palette
	Roles   = theme.Roles
	Swatch  = theme.Swatch
)

// Cyberpunk is the default palette (omoro.io's Cyberpunk).
var Cyberpunk = theme.Cyberpunk

// ThemeNamed builds the theme from the palette config names (the default when unknown).
func ThemeNamed(name string, dark bool) Theme { return NewThemeFrom(theme.Must(name), dark) }

// Flip is the same palette for the other background (or the same one: dark says which).
func (t Theme) Flip(dark bool) Theme {
	n := ThemeNamed(t.Name, dark)
	n.Forced = true
	return n
}
