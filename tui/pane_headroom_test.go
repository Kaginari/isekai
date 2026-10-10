package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// The first screen (2026-10-10, VS Code): the panes fill the window, but the window's top row
// is not ours to keep — once the launching command line scrolls above the viewport, VS Code's
// sticky scroll pins it over that row, and it covered the slime's head and the welcome's top
// border ("everything is one line down"). The frame keeps the window's height exactly (the
// inline renderer's rectangle) and leaves its first row as headroom: the pane edge only, no
// content. Every row is exactly the window's width, the divider sits at one column all the way
// down, and the mascot's rows sit whole right under the headroom.
func TestPaneFirstScreenMatrix(t *testing.T) {
	head, _ := mascotArt("isekai", false)
	if len(head) == 0 {
		t.Fatal("the isekai mascot has no art")
	}
	for _, w := range []int{100, 110, 120, 130, 140, 160} {
		for _, h := range []int{26, 30, 34, 40, 50} {
			t.Run(fmt.Sprintf("%dx%d", w, h), func(t *testing.T) {
				m := New(&fakeHost{}, NewTheme(true), DefaultWords())
				m.Update(tea.WindowSizeMsg{Width: w, Height: h})
				m.Update(evBgWait{}) // the welcome lands
				rail := m.railVisible()
				if rail == 0 {
					t.Fatalf("%dx%d must carry the sidebar", w, h)
				}
				cw := w - rail
				f := m.Render()
				rows := strings.Split(f, "\n")
				if len(rows) != h {
					t.Fatalf("frame is %d rows, want exactly the window's %d", len(rows), h)
				}
				for i, l := range rows {
					if got := ansi.StringWidth(l); got != w {
						t.Fatalf("row %d is %d cells, want %d: %q", i, got, w, ansi.Strip(l))
					}
				}
				strip := strings.Split(ansi.Strip(f), "\n")
				// the headroom: nothing but the pane edge
				r0 := []rune(strip[0])
				if strings.TrimSpace(string(r0[:cw])) != "" || strings.TrimSpace(string(r0[cw+3:])) != "" {
					t.Fatalf("the headroom row carries content: %q", strip[0])
				}
				// the divider at one column on every pane row, headroom included
				for i := 0; i < h-4; i++ {
					r := []rune(strip[i])
					if string(r[cw:cw+3]) != "░▒│" {
						t.Fatalf("row %d: pane edge at col %d is %q, want ░▒│: %q", i, cw, string(r[cw:cw+3]), strip[i])
					}
				}
				// the welcome's top border opens the chat pane right under the headroom
				if !strings.HasPrefix(strip[paneTop], "╭") {
					t.Fatalf("row %d must open the welcome card: %q", paneTop, strip[paneTop])
				}
				// the mascot, every row whole, from the first pane row
				if h >= 30 && ansi.StringWidth(head[0])+6 <= rail-3 {
					for i, a := range head {
						want := strings.TrimSpace(ansi.Strip(a))
						if want == "" {
							continue
						}
						got := strings.TrimSpace(string([]rune(strip[paneTop+i])[cw+3:]))
						if got != want {
							t.Fatalf("mascot row %d: got %q, want %q", i, got, want)
						}
					}
				}
			})
		}
	}
}

// A click lands where the human sees it: the window's row is the pane's row plus the headroom.
func TestPaneMouseCountsFromBelowTheHeadroom(t *testing.T) {
	stubCopy(t)
	m := selModel(t)
	y := rowOf(t, m, "bravo line") // a window row (the frame includes the headroom)
	m.Update(tea.MouseClickMsg{X: 0, Y: y, Button: tea.MouseLeft})
	m.Update(tea.MouseClickMsg{X: 0, Y: y, Button: tea.MouseLeft}) // a double click: the word
	if m.sel == nil || !m.sel.shown {
		t.Fatal("the double click selected nothing")
	}
	if got := m.selText(); got != "bravo" {
		t.Fatalf("a double click on row %d selected %q, want %q", y, got, "bravo")
	}
}
