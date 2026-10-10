package world

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// The ledger shelf (`<worldDir>/instruments/ledgers.jsonl`): the council's strain, its
// consult counts, its stress-death memos and the court-write attribution notebook used to
// live process-local and die with it — a world that forgot its scars on every restart
// relearned nothing (freeze review 2026-10-07). Every mutation appends one JSON row;
// Open replays the shelf (last row wins per key) so rests survive restarts. An unreadable
// shelf never blocks a boot: the world simply opens ledgerless, exactly like before.

const ledgerShelf = "instruments/ledgers.jsonl"

type ledgerRow struct {
	Kind   string    `json:"kind"` // stress | consults | stressClear | death | courtWrite
	Body   string    `json:"body"`
	Value  int       `json:"value,omitempty"`
	Lane   string    `json:"lane,omitempty"`
	Path   string    `json:"path,omitempty"`
	Writer string    `json:"writer,omitempty"`
	At     time.Time `json:"at"`
}

// ledgerNote appends one row. Call sites write under the world lock, so the file's order is
// the memory's order — the replay stands on that.
func (w *World) ledgerNote(r ledgerRow) {
	if r.At.IsZero() {
		r.At = time.Now().UTC()
	}
	b, err := json.Marshal(r)
	if err != nil {
		return
	}
	p := filepath.Join(w.Dir(), filepath.FromSlash(ledgerShelf))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(b, '\n'))
}

// loadLedgers replays the shelf into the in-memory ledgers. Court-write rows replay through
// the same per-creature bound as runtime writes; the clear row empties both strain ledgers,
// exactly like the run-time clean-landing law.
func (w *World) loadLedgers() {
	f, err := os.Open(filepath.Join(w.Dir(), filepath.FromSlash(ledgerShelf)))
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 16<<20)
	for sc.Scan() {
		var r ledgerRow
		if json.Unmarshal(sc.Bytes(), &r) != nil || r.Body == "" {
			continue
		}
		switch r.Kind {
		case "stress":
			w.stress[r.Body] = r.Value
		case "consults":
			w.consults[r.Body] = r.Value
		case "stressClear":
			delete(w.stress, r.Body)
			delete(w.consults, r.Body)
		case "death":
			w.stressDeaths[r.Body] = stressEnd{Lane: r.Lane, At: r.At}
		case "courtWrite":
			if r.Path == "" {
				continue
			}
			bm := w.courtWrites[r.Body]
			if bm == nil {
				bm = map[string]courtWrite{}
				w.courtWrites[r.Body] = bm
			}
			bm[normRel(r.Path)] = courtWrite{Writer: or(r.Writer, r.Body), At: r.At}
			for len(bm) > courtWritesBound {
				oldest := ""
				for k, cw := range bm {
					if oldest == "" || cw.At.Before(bm[oldest].At) {
						oldest = k
					}
				}
				delete(bm, oldest)
			}
		}
	}
}
