package world

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kaginari/isekai/board"
	"github.com/Kaginari/isekai/loop"
	"github.com/Kaginari/isekai/value"
	"github.com/Kaginari/isekai/wire"
)

// wireResult is a Court's closed run: the loop's status, the body's @S, its own @? lines and the
// holes the harness raised.
func wireResult(loopStatus, word string, self int, holes ...string) *loop.Result {
	r := &loop.Result{Status: loopStatus, IsWire: true, Report: wire.Report{Status: word}, Holes: holes}
	for i := 0; i < self; i++ {
		r.Report.Holes = append(r.Report.Holes, "an honest unknown")
	}
	return r
}

// The value ledger is always kept: a world whose table ascends nobody still measures its bodies,
// and the record replays at Open.
func TestValueLedgerKeptWithoutAscension(t *testing.T) {
	root := ascendWorld(t)
	w := openWorldRanks(t, root, Ranks(DefaultRanks(Isekai())))
	if w.ascend != nil {
		t.Fatal("the world's own table ascends nobody")
	}
	w.FoldMetrics("slime-login", "DONE", 0)
	w.FoldMetrics("slime-login", "FAIL", 0)
	if _, err := os.Stat(filepath.Join(root, ".isekai", filepath.FromSlash(value.Shelf))); err != nil {
		t.Fatalf("the ledger is written with no ascension: %v", err)
	}
	w2 := openWorldRanks(t, root, Ranks(DefaultRanks(Isekai())))
	if m := w2.Metrics("slime-login"); m["reports"] != 2 || m["passes"] != 1 {
		t.Fatalf("the ledger replays at Open: %v", m)
	}
}

// One definition of success: the ASL fold, the council, the board's activity and the score read
// the same word the same way — a body's own @S FAIL on a clean loop fails everywhere; the human's
// DENIED is neutral everywhere.
func TestFourDefinitionsOfSuccessAgree(t *testing.T) {
	root := ascendWorld(t)
	w := openWorldRanks(t, root, Ranks(DefaultRanks(Isekai())))
	court := CourtOptions{StressEscalateAt: 5}

	failed := wireResult(loop.Done, "FAIL — the tests are red", 0)
	w.councilLedger("slime-login", verdictOf(failed), failed, court)
	w.foldResult("slime-login", "orc-auth", failed, w.stress["slime-login"], false, false)
	denied := wireResult(loop.Denied, "", 0)
	w.councilLedger("slime-login", verdictOf(denied), denied, court)
	w.foldResult("slime-login", "orc-auth", denied, w.stress["slime-login"], false, false)

	m := w.Metrics("slime-login")
	if m["fails"] != 1 || m["reports"] != 1 || m["neutral"] != 1 {
		t.Fatalf("the fold: one fail, the DENIED neutral: %v", m)
	}
	if w.stress["slime-login"] != 1 {
		t.Fatalf("the council: the FAIL strains once, the DENIED never: %d", w.stress["slime-login"])
	}
	act := board.BuildActivityDef([]board.Run{
		// the journal's end line: the loop said DONE, the run's word is the body's FAIL
		{As: "slime-login", Status: failed.Status, Word: value.RunWord(failed.Status, failed.Report.Status), At: time.Now().UTC().Format(time.RFC3339Nano)},
		{As: "slime-login", Status: denied.Status, Word: value.RunWord(denied.Status, ""), At: time.Now().UTC().Format(time.RFC3339Nano)},
	}, nil, time.Now(), 28, w.ValueOpts().Def)
	if a := act.Bodies["slime-login"]; a == nil || a.Total != 2 || a.Failed != 1 {
		t.Fatalf("the board's activity: one failed of two: %+v", a)
	}
	if v := w.Value("slime-login"); v.Score != 0 || v.Confidence != "low" {
		t.Fatalf("the score reads the same record: %+v", v)
	}
	// a clean landing clears the council and lands in the fold
	ok := wireResult(loop.Done, "DONE", 0)
	w.councilLedger("slime-login", verdictOf(ok), ok, court)
	w.foldResult("slime-login", "orc-auth", ok, w.stress["slime-login"], false, false)
	if w.stress["slime-login"] != 0 || w.Metrics("slime-login")["passes"] != 1 {
		t.Fatal("a pass clears the strain and counts as a pass")
	}
}

// Holes are split by whose they are: the world's faults and the body's honest @? never cost it.
func TestInfraHolesDoNotCost(t *testing.T) {
	root := ascendWorld(t)
	w := openWorldRanks(t, root, Ranks(DefaultRanks(Isekai())))
	r := wireResult(loop.Done, "DONE", 2, "the provider refused the request", "recall: memory: index stale", "law.gate is off — the turn's writes were not checked", "answer off the wire — sent back (1 of 1)")
	w.foldResult("slime-login", "orc-auth", r, 0, false, false)
	rows := w.Ledger()
	got := rows[len(rows)-1].Holes
	if got.Infra != 3 || got.Harness != 1 || got.Self != 2 {
		t.Fatalf("holes by kind: %+v", got)
	}
	if m := w.Metrics("slime-login"); m["holes"] != 1 || m["holeRate"] != 100 {
		t.Fatalf("only the harness hole costs: %v", m)
	}
}

// A gate verdict credits the body that held the gate; a pass a later fail on the same paths undoes
// is an overturn against the holder.
func TestGateVerdictFoldsAndOverturns(t *testing.T) {
	root := ascendWorld(t)
	w := openWorldRanks(t, root, Ranks(DefaultRanks(Isekai())))
	w.noteGate("slime-login", Verdict{Word: "pass", VerifyRun: 2}, []string{"auth/login/login.go"})
	w.foldResult("slime-login", "orc-auth", wireResult(loop.Done, "DONE", 0), 0, false, false)
	w.noteGate("slime-login", Verdict{Word: "pass", VerifyRun: 1}, []string{"auth/login/other.go"})
	w.foldResult("slime-login", "orc-auth", wireResult(loop.Done, "DONE", 0), 0, false, false)
	w.noteGate("slime-login", Verdict{Word: "fail", VerifyRun: 2, VerifyRed: 1}, []string{"auth/login/login.go"})
	w.foldResult("slime-login", "orc-auth", wireResult(loop.Fail, "", 0), 0, false, false)
	o := w.Metrics("orc-auth")
	if o["verdicts"] != 3 || o["overturned"] != 1 || int(o["gateRight"]) != 66 {
		t.Fatalf("the holder's record: %v", o)
	}
	s := w.Metrics("slime-login")
	if int(s["gateRate"]) != 66 || int(s["verifyGreenRate"]) != 80 {
		t.Fatalf("the author's gate record: gateRate %v verifyGreenRate %v", s["gateRate"], s["verifyGreenRate"])
	}
	// the overturn replays: a reopened world knows the standing passes and the undone one
	w2 := openWorldRanks(t, root, Ranks(DefaultRanks(Isekai())))
	if w2.Metrics("orc-auth")["overturned"] != 1 {
		t.Fatal("the overturn is on the ledger")
	}
	w2.noteGate("slime-login", Verdict{Word: "fail"}, []string{"auth/login/other.go"})
	w2.foldResult("slime-login", "orc-auth", wireResult(loop.Fail, "", 0), 0, false, false)
	if w2.Metrics("orc-auth")["overturned"] != 2 {
		t.Fatal("a pass standing from before the reopen is overturned too")
	}
}

// The throne's turns are the root's record: the Ended hook folds every human turn.
func TestThroneTurnsFold(t *testing.T) {
	root := ascendWorld(t)
	w := openWorldRanks(t, root, Ranks(DefaultRanks(Isekai())))
	end := w.turnEnded(Rimuru)
	end(nil, &loop.Result{Status: loop.Done, Wrote: []string{"a.go"}})
	end(nil, &loop.Result{Status: loop.Done, Interrupted: true})
	m := w.Metrics(Rimuru)
	if m["runs"] != 2 || m["reports"] != 1 || m["neutral"] != 1 || m["askDoneRate"] != 100 {
		t.Fatalf("the root's record (the interrupted turn is neutral): %v", m)
	}
	if v := w.Value(Rimuru); v.Family != "root" || v.Meter.Signal != "askDoneRate" {
		t.Fatalf("the throne is measured as the root: %+v", v)
	}
}

// A custom race in config gets its family's value law and meter from its flags; one that names
// its own score keeps it.
func TestCustomRaceGetsItsFamilyValue(t *testing.T) {
	rs := append(Ranks(DefaultRanks(Isekai())),
		Rank{Name: "night-watch", ReportsTo: Rimuru, Tools: []string{"read", "grep"}, Dir: "night-watch", Prefix: "night-watch-"},
		Rank{Name: "scribe", ReportsTo: "orc", Authors: true, Dir: "scribe", Prefix: "scribe-", Score: "passRate", Meter: value.MeterDef{Signal: "passRate", Label: "Pass", Bands: value.Bands{OK: 90, Warn: 60}}})
	if err := rs.Validate(); err != nil {
		t.Fatal(err)
	}
	if rs.FamilyOf("night-watch") != "auditor" || rs.FamilyOf("night_watch") != "auditor" || rs.FamilyOf("scribe") != "author" || rs.FamilyOf("orc") != "gate" || rs.FamilyOf("elf") != "planner" || rs.FamilyOf("kijin") != "steward" || rs.FamilyOf("dark-elf") != "auditor" {
		t.Fatal("families from flags")
	}
	root := ascendWorld(t)
	w := openWorldRanks(t, root, rs)
	if sp := w.Spec("night-watch"); sp.Score != value.FamilyDefault("auditor").Score || sp.Meter.Signal != "upheldRate" {
		t.Fatalf("the auditor family's law: %+v", sp)
	}
	if sp := w.Spec("scribe"); sp.Score != "passRate" || sp.Meter.Bands.OK != 90 {
		t.Fatalf("a race's own law: %+v", sp)
	}
}
