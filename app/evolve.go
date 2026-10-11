package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Kaginari/isekai/config"
	"github.com/Kaginari/isekai/tui"
	"github.com/Kaginari/isekai/value"
	"github.com/Kaginari/isekai/world"
)

// law.evolution as the app arms it (world/evolve.go does the work): the founding sketch reads a
// territory from the code, the gate's sandbox contains the verify lines a plan runs, and every
// evolution reaches the surfaces — the session stream and the dashboard (an `evolution` event),
// the terminal (a toast and a sidebar line), the plain console (one line).

// evolutionOptions is law.evolution for the world: config's numbers (armEvolution adds the
// app's hands).
func evolutionOptions(cfg *config.Config) world.EvolutionOptions {
	e := cfg.Law.Evolution
	return world.EvolutionOptions{Enabled: e.Enabled, Mode: e.Mode, ThoughtsMax: e.ThoughtsMax, StressAt: cfg.EvolutionStressAt(), FailStreak: e.FailStreak,
		CooldownRuns: e.CooldownRuns, Arms: append([]string(nil), e.Arms...), SplitLines: e.SplitLines, SplitFiles: e.SplitFiles, SkillStatus: cfg.Law.Forge.EvolvedStatus}
}

// armEvolution hands the world its law.evolution with the app's hands: the survey, the sandbox,
// the verify clock, the surfaces.
func (a *App) armEvolution() {
	if a.World == nil {
		return
	}
	o := evolutionOptions(a.Cfg)
	o.Survey = a.surveyTerritory
	o.Sandbox = a.Sandbox
	o.VerifyTimeout = a.Cfg.Law.Gate.VerifyTimeout.D()
	o.On = a.evolved
	a.World.SetEvolution(o)
}

// evolved carries one evolution to every surface that is up.
func (a *App) evolved(ev world.EvolveEvent) {
	msg := tui.EvEvolution{ID: ev.ID, Body: ev.Body, Arm: ev.Arm, Into: ev.Into, Reason: ev.Reason, Status: ev.Status, Mode: ev.Mode, Hole: ev.Hole, Skill: ev.Skill, Line: ev.Line()}
	a.mu.Lock()
	sink, web := a.evoSink, a.web
	a.mu.Unlock()
	switch {
	case sink != nil:
		sink(msg) // the terminal and the dashboard (the host fans out)
	case web != nil:
		web.Send(msg)
	}
	if sink == nil && !a.Opt.Quiet {
		mark := "✦"
		if ev.Status == "hole" {
			mark = "@?"
		}
		fmt.Fprintf(a.Opt.Err, "%s %s\n", mark, ev.Line())
	}
}

// evolveSweep is the run-start pass (law.evolution): every stressed body whose desk overflowed
// is weighed before the first turn.
func (a *App) evolveSweep() {
	if a.World == nil || !a.Cfg.EvolutionOn() {
		return
	}
	a.World.EvolveSweep()
}

// --- the survey: the founding sketch's eyes on one territory ---

var surveyCache struct {
	mu   sync.Mutex
	root string
	at   time.Time
	s    *survey
}

// surveyTerritory reads a territory the way the medium founding reads a slime's unit: its files,
// lines, packages and exports, its tests and the runner that proves them.
func (a *App) surveyTerritory(prefixes []string) world.TerritorySurvey {
	surveyCache.mu.Lock()
	if surveyCache.s == nil || surveyCache.root != a.Root || time.Since(surveyCache.at) > 30*time.Second {
		surveyCache.s, surveyCache.root, surveyCache.at = readSurvey(a.Root, a.Cfg.Dist.WorldDir), a.Root, time.Now()
	}
	s := surveyCache.s
	surveyCache.mu.Unlock()
	u := &unit{langs: map[string]int{}}
	var clean []string
	for _, p := range prefixes {
		p = strings.Trim(strings.TrimSpace(p), "`")
		p = strings.TrimSuffix(p, "/")
		if p == "" {
			continue
		}
		clean = append(clean, p)
		if fi, err := os.Stat(filepath.Join(a.Root, filepath.FromSlash(p))); err == nil && fi.IsDir() {
			u.dirs = append(u.dirs, p)
		} else {
			u.files = append(u.files, p)
		}
	}
	if len(clean) == 0 {
		return world.TerritorySurvey{}
	}
	u.name = slug(filepath.Base(clean[len(clean)-1]))
	files := 0
	for i := range s.files {
		f := &s.files[i]
		for _, p := range clean {
			if f.rel == p || strings.HasPrefix(f.rel, p+"/") {
				s.absorb(u, f)
				files++
				break
			}
		}
	}
	if files == 0 {
		return world.TerritorySurvey{}
	}
	s.place(u)
	pl := s.slimePlan(u, []*unit{u}, [][]float64{{0}}, a.Lex.Prefixes)
	return world.TerritorySurvey{Files: files, Lines: u.lines, Purpose: pl.Purpose, Traits: pl.Traits, Verify: pl.Verify}
}

// --- the desk: isekai evolve list | show <id> | apply <id> | reject <id> <why> ---

func (a *App) cmdEvolve(args []string, io IO) int {
	if a.World == nil {
		fmt.Fprintf(io.Err, "@S FAIL\n@? no %s here\n", worldWord(a.Cfg.Dist.Name))
		return 2
	}
	sub := "list"
	if len(args) > 0 {
		sub = args[0]
	}
	switch sub {
	case "list", "ls":
		ps := a.World.Proposals()
		open := 0
		for _, p := range ps {
			if p.Status == "open" {
				open++
			}
		}
		mode := a.Cfg.Law.Evolution.Mode
		fmt.Fprintf(io.Out, "evolution · mode %s · %d open proposal%s of %d\n", mode, open, plural(open), len(ps))
		for i := len(ps) - 1; i >= 0; i-- {
			p := ps[i]
			fmt.Fprintf(io.Out, "  %s  %-9s %-8s %-18s %s — %s\n", p.ID, p.Status, p.Arm, p.Body, p.At.Local().Format("2006-01-02 15:04"), p.Reason)
		}
		for _, r := range a.recentEvolutions(5) {
			fmt.Fprintln(io.Out, "  "+r)
		}
		return 0
	case "show":
		if len(args) < 2 {
			fmt.Fprintf(io.Err, "@S FAIL\n@? evolve show <id>\n")
			return 2
		}
		p, err := a.World.Proposal(args[1])
		if err != nil {
			fmt.Fprintf(io.Err, "@S FAIL\n@? %v\n", err)
			return 2
		}
		fmt.Fprintf(io.Out, "%s · %s · %s · %s\nwhy: %s\ntrigger: %s\n", p.ID, p.Body, p.Arm, p.Status, p.Reason, strings.Join(p.Evidence, "; "))
		if len(p.Into) > 0 {
			fmt.Fprintf(io.Out, "into: %s\n", strings.Join(p.Into, ", "))
		}
		if p.Why != "" {
			fmt.Fprintf(io.Out, "note: %s\n", p.Why)
		}
		fmt.Fprint(io.Out, p.Diff())
		return 0
	case "apply":
		if len(args) < 2 {
			fmt.Fprintf(io.Err, "@S FAIL\n@? evolve apply <id>\n")
			return 2
		}
		ev, err := a.World.ApplyProposal(context.Background(), args[1])
		if err != nil {
			fmt.Fprintf(io.Err, "@S FAIL\n@? %v\n", err)
			return 1
		}
		fmt.Fprintf(io.Out, "@S OK %s (log.md)\n", ev.Line())
		return 0
	case "reject":
		if len(args) < 3 {
			fmt.Fprintf(io.Err, "@S FAIL\n@? evolve reject <id> <why>\n")
			return 2
		}
		p, err := a.World.RejectProposal(args[1], strings.Join(args[2:], " "))
		if err != nil {
			fmt.Fprintf(io.Err, "@S FAIL\n@? %v\n", err)
			return 2
		}
		fmt.Fprintf(io.Out, "@S OK rejected %s (%s %s): %s\n", p.ID, p.Arm, p.Body, p.Why)
		return 0
	}
	fmt.Fprintf(io.Err, "@S FAIL\n@? evolve %s — list | show <id> | apply <id> | reject <id> <why>\n", sub)
	return 2
}

// recentEvolutions are the ledger's latest evolution rows, one line each, newest first.
func (a *App) recentEvolutions(n int) []string {
	var out []string
	for _, r := range a.World.EvolutionView(n).Recent {
		out = append(out, evolutionRow(r))
	}
	return out
}

func evolutionRow(r value.Row) string {
	what := r.Arm
	if len(r.Into) > 0 && r.Arm == "split" {
		what += " → " + strings.Join(r.Into, ", ")
	}
	return fmt.Sprintf("%s %s %s: %s (%s) — %s", r.At.Local().Format("2006-01-02 15:04"), r.Body, or(r.Status, "applied"), what, r.Rank, r.Reason)
}

// evolutionStatus is status's evolution block: the law as armed, the queue, the latest rows.
func (a *App) evolutionStatus() []string {
	if a.World == nil {
		return nil
	}
	v := a.World.EvolutionView(5)
	state := "on"
	if !v.Enabled {
		state = "off"
	}
	lines := []string{fmt.Sprintf("evolution: %s · mode %s · desk > %d · stressAt %d · cooldown %d runs · %d open proposal%s", state, v.Mode, v.ThoughtsMax, v.StressAt, v.CooldownRuns, v.Open, plural(v.Open))}
	for _, r := range v.Recent {
		lines = append(lines, "  "+evolutionRow(r))
	}
	return lines
}
