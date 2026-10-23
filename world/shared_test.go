package world

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/loop"
	"github.com/Kaginari/isekai/provider/mock"
	"github.com/Kaginari/isekai/queue"
)

// TestQueueOutranks: the world's queue moves a holder's work for the holder and for anyone up
// its Reports-to chain (rimuru tops it); a sibling or a child is refused.
func TestQueueOutranks(t *testing.T) {
	w := open(t, fixture(t))
	for _, c := range []struct {
		by, owner string
		want      bool
	}{
		{"orc-security", "slime-auth", true}, {"elf-core", "slime-auth", true}, {Rimuru, "slime-auth", true},
		{Rimuru, "", true}, {"slime-api", "slime-auth", false}, {"slime-auth", "orc-security", false}, {"elf-core", "", false},
	} {
		if got := w.Outranks(c.by, c.owner); got != c.want {
			t.Errorf("Outranks(%s, %s) = %v", c.by, c.owner, got)
		}
	}
	q := w.Queue()
	if q.Path() != filepath.Join(w.Dir(), "shared", "queue.jsonl") {
		t.Fatalf("the queue lives on the shared shelf: %s", q.Path())
	}
	q.Add(queue.Item{For: "slime-auth", Ask: "fix"}, "orc-security")
	q.Claim("q1", "slime-auth")
	if _, err := q.Requeue("q1", "slime-api", ""); !queue.IsRefusal(err) {
		t.Fatal("a sibling requeued work it does not hold")
	}
	if _, err := q.Requeue("q1", "orc-security", "reroute"); err != nil {
		t.Fatalf("the orc over the holder: %v", err)
	}
}

// TestTerritorySharedShelfIsHarnessSoil: the shared shelf is the harness's — a body's write
// there is refused like engine soil, naming the shelf.
func TestTerritorySharedShelfIsHarnessSoil(t *testing.T) {
	dir := fixture(t)
	w := open(t, dir)
	m := mock.New(
		writeCall("1", ".isekai/shared/queue.jsonl", "{}\n"),
		mock.Text("@S DONE\n@U colony c\n@E 0"),
	)
	e := w.Engine("slime-auth", build(m))
	r, err := e.Run(context.Background(), "edit the queue by hand")
	if err != nil || r.Status != loop.Done {
		t.Fatalf("%v %+v", err, r)
	}
	if len(r.Steps) != 1 || r.Steps[0].Status != "refused" {
		t.Fatalf("the write must refuse: %+v", r.Steps)
	}
	if res := m.Requests[1].Messages[2].ToolResults[0]; !strings.Contains(res.Content, "shared shelf") {
		t.Fatalf("the refusal names the shelf: %q", res.Content)
	}
	if _, err := os.Stat(filepath.Join(dir, ".isekai", "shared", "queue.jsonl")); err == nil {
		t.Fatal("the write landed on the shelf")
	}
}

// TestSharedShelfMigrates: an old world's ledgers and value ledger move onto the shared shelf
// at Open — renamed whole, every line kept — and read back as before; new rows land there.
func TestSharedShelfMigrates(t *testing.T) {
	dir := fixture(t)
	inst := filepath.Join(dir, ".isekai", "instruments")
	os.MkdirAll(inst, 0o755)
	ledger := `{"kind":"stress","body":"slime-auth","value":2,"at":"2026-10-09T09:00:00Z"}` + "\n"
	val := `{"kind":"seed","at":"2026-10-01T00:00:00Z","from":"instruments/ascension.jsonl"}` + "\n" +
		`{"kind":"run","at":"2026-10-09T09:00:00Z","body":"slime-auth","loop":"DONE","word":"DONE","status":"DONE"}` + "\n"
	os.WriteFile(filepath.Join(inst, "ledgers.jsonl"), []byte(ledger), 0o644)
	os.WriteFile(filepath.Join(inst, "value.jsonl"), []byte(val), 0o644)
	// an ascension tally too: a moved ledger already carries its seed, it is never seeded twice
	os.WriteFile(filepath.Join(inst, "ascension.jsonl"), []byte(`{"kind":"report","body":"slime-api","status":"DONE","at":"2026-09-01T00:00:00Z"}`+"\n"), 0o644)
	w := open(t, dir)
	for name, want := range map[string]string{"ledgers.jsonl": ledger, "value.jsonl": val} {
		b, err := os.ReadFile(filepath.Join(w.SharedDir(), name))
		if err != nil || string(b) != want {
			t.Fatalf("%s moved whole: %v %q", name, err, b)
		}
		if _, err := os.Stat(filepath.Join(inst, name)); err == nil {
			t.Fatalf("%s left behind", name)
		}
	}
	if w.stress["slime-auth"] != 2 || len(w.val.rows) != 2 {
		t.Fatalf("the moved shelf read back: stress %v, %d value rows", w.stress, len(w.val.rows))
	}
	w.ledgerNote(ledgerRow{Kind: "consults", Body: "slime-auth", Value: 1})
	if b, _ := os.ReadFile(filepath.Join(w.SharedDir(), "ledgers.jsonl")); strings.Count(string(b), "\n") != 2 {
		t.Fatalf("a new row lands on the shared shelf: %q", b)
	}
}

// TestSharedShelfReadsBothHomes: when an older binary wrote the old file again after the move,
// Open reads the old file's lines first, then the shared shelf's — nothing is lost, nothing
// moved over a file that exists.
func TestSharedShelfReadsBothHomes(t *testing.T) {
	dir := fixture(t)
	for rel, row := range map[string]string{
		"instruments/ledgers.jsonl": `{"kind":"stress","body":"slime-auth","value":1,"at":"2026-10-09T09:00:00Z"}`,
		"shared/ledgers.jsonl":      `{"kind":"stress","body":"slime-auth","value":3,"at":"2026-10-10T09:00:00Z"}` + "\n" + `{"kind":"stress","body":"slime-api","value":2,"at":"2026-10-10T09:00:00Z"}`,
	} {
		p := filepath.Join(dir, ".isekai", filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(row+"\n"), 0o644)
	}
	w := open(t, dir)
	if w.stress["slime-auth"] != 3 || w.stress["slime-api"] != 2 {
		t.Fatalf("both homes, old first: %v", w.stress)
	}
	if _, err := os.Stat(filepath.Join(dir, ".isekai", "instruments", "ledgers.jsonl")); err != nil {
		t.Fatal("a move over an existing shared file")
	}
}

// TestSharedShelfRefusesTheShell: a body's shell write onto the shared shelf is refused like a
// write tool's — a heredoc appended to the queue cannot forge a claim on another body's work.
func TestSharedShelfRefusesTheShell(t *testing.T) {
	dir := fixture(t)
	w := open(t, dir)
	q := w.Queue()
	if _, err := q.Add(queue.Item{For: "slime-api", Ask: "fix the api"}, "orc-security"); err != nil {
		t.Fatal(err)
	}
	forged := `{"id":"q1","rev":2,"op":"claim","state":"claimed","for":"slime-api","ask":"fix the api","owner":"slime-auth","by":"slime-auth"}`
	for i, cmd := range []string{
		"cat >> .isekai/shared/queue.jsonl <<'EOF'\n" + forged + "\nEOF",
		"echo '" + forged + "' | tee -a ./.isekai/shared/queue.jsonl",
		"cd .isekai/shared && printf '%s\\n' '" + forged + "' >>queue.jsonl",
	} {
		m := mock.New(
			mock.Call("1", "bash", map[string]string{"command": cmd}),
			mock.Text("@S DONE\n@U colony c\n@E 0"),
		)
		e := w.Engine("slime-auth", build(m))
		r, err := e.Run(context.Background(), "take the api's work")
		w.Close(e)
		if err != nil || len(r.Steps) != 1 || r.Steps[0].Status != "refused" {
			t.Fatalf("%d: the shell write must refuse: %v %+v %+v", i, err, r.Steps, r)
		}
		if res := m.Requests[1].Messages[2].ToolResults[0]; !strings.Contains(res.Content, "shared shelf") || !strings.Contains(res.Content, "queue tool") {
			t.Fatalf("%d: the refusal names the shelf and the body's door to it: %q", i, res.Content)
		}
		if it, _, _ := q.Get("q1"); it.Owner != "" || it.State != queue.Held {
			t.Fatalf("%d: the forged line landed: %+v", i, it)
		}
	}
}

// TestSharedShelfMergesHomesByTime: rows an older binary appended to the old file after the move
// interleave with the shared file's by their time — a newer clear in the old file is not undone
// by an older strain on the shared shelf — and a second seed an older binary wrote there (its
// ledger looked missing) carries nothing over twice; the rows it appended after its seed stay.
func TestSharedShelfMergesHomesByTime(t *testing.T) {
	dir := fixture(t)
	for rel, rows := range map[string]string{
		"instruments/ledgers.jsonl":   `{"kind":"stressClear","body":"slime-auth","at":"2026-10-11T09:00:00Z"}`,
		"shared/ledgers.jsonl":        `{"kind":"stress","body":"slime-auth","value":3,"at":"2026-10-10T09:00:00Z"}`,
		"instruments/ascension.jsonl": `{"kind":"report","body":"slime-api","status":"DONE","at":"2026-09-01T00:00:00Z"}`,
		"shared/value.jsonl": `{"kind":"run","at":"2026-09-01T00:00:00Z","body":"slime-api","loop":"DONE","word":"DONE","status":"DONE"}` + "\n" +
			`{"kind":"seed","at":"2026-10-01T00:00:00Z","from":"instruments/ascension.jsonl"}` + "\n" +
			`{"kind":"run","at":"2026-10-09T09:00:00Z","body":"slime-auth","loop":"DONE","word":"DONE","status":"DONE"}`,
		"instruments/value.jsonl": `{"kind":"run","at":"2026-09-01T00:00:00Z","body":"slime-api","loop":"DONE","word":"DONE","status":"DONE"}` + "\n" +
			`{"kind":"seed","at":"2026-10-12T00:00:00Z","from":"instruments/ascension.jsonl"}` + "\n" +
			`{"kind":"run","at":"2026-10-12T09:00:00Z","body":"slime-auth","loop":"FAIL","word":"FAIL","status":"FAIL"}`,
	} {
		p := filepath.Join(dir, ".isekai", filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(rows+"\n"), 0o644)
	}
	w := open(t, dir)
	if v, ok := w.stress["slime-auth"]; ok {
		t.Fatalf("the newer clear in the old file stands: stress %d", v)
	}
	var kinds []string
	for _, r := range w.val.rows {
		kinds = append(kinds, r.Kind+":"+r.Body+":"+r.Word)
	}
	if got := strings.Join(kinds, ","); got != "run:slime-api:DONE,seed::,run:slime-auth:DONE,run:slime-auth:FAIL" {
		t.Fatalf("one seed, both homes' own rows by time: %s", got)
	}
}
