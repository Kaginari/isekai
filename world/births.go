package world

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// The births ledger (`<worldDir>/attic/births/<creature>.jsonl`): every time the harness births a
// body — a zone-parent's genesis passing its gate, a relief's young, an ascension, an evolution's
// split or rebirth — one line lands in that body's own file, never rewritten. The doc says who a
// body is; the ledger says when it came to be, which is what tenure reads. The attic is harness
// soil: no body writes it, and the tree watch never bills it.

// birthRecord is one birth fact.
type birthRecord struct {
	At   time.Time `json:"at"`
	Body string    `json:"body"`
	Kind string    `json:"kind"`           // genesis | split | ascend | evolve-split | rebirth
	From string    `json:"from,omitempty"` // the body it came from (the retiree, the base, the parent)
	By   string    `json:"by,omitempty"`   // who birthed it: the writer of a genesis, else the harness's act
	Doc  string    `json:"doc,omitempty"`
}

// BirthLedgerPath is one body's births file.
func (w *World) BirthLedgerPath(body string) string {
	return filepath.Join(w.Dir(), "attic", "births", strings.ToLower(body)+".jsonl")
}

// logBirth appends one birth; a failure is a note (the birth itself stands on its doc).
func (w *World) logBirth(r birthRecord) {
	if r.At.IsZero() {
		r.At = time.Now().UTC()
	}
	b, err := json.Marshal(r)
	if err == nil {
		p := w.BirthLedgerPath(r.Body)
		if err = os.MkdirAll(filepath.Dir(p), 0o755); err == nil {
			var f *os.File
			if f, err = os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
				_, err = f.Write(append(b, '\n'))
				f.Close()
			}
		}
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if err != nil {
		w.Notes = append(w.Notes, "births ledger: "+r.Body+": "+err.Error())
		return
	}
	if t, ok := w.born[r.Body]; !ok || r.At.Before(t) || t.IsZero() {
		w.born[r.Body] = r.At
	}
}

// Births is a body's birth facts, oldest first (a torn line is skipped).
func (w *World) Births(body string) []birthRecord {
	f, err := os.Open(w.BirthLedgerPath(body))
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []birthRecord
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var r birthRecord
		if json.Unmarshal(sc.Bytes(), &r) == nil && !r.At.IsZero() {
			out = append(out, r)
		}
	}
	return out
}

// bornAt is when a body was first born: its first ledger line, else — a body born before the
// ledger is grandparented — the first commit that added its doc. ok is false when neither speaks.
func (w *World) bornAt(body string) (time.Time, bool) {
	w.mu.Lock()
	t, known := w.born[body]
	w.mu.Unlock()
	if known {
		return t, !t.IsZero()
	}
	t = time.Time{}
	if bs := w.Births(body); len(bs) > 0 {
		t = bs[0].At
		for _, b := range bs[1:] {
			if b.At.Before(t) {
				t = b.At
			}
		}
	} else if doc := w.docOf(body); doc != "" {
		if out, err := w.gitOut("log --diff-filter=A --format=%at -- " + doc); err == nil {
			if ls := strings.Fields(out); len(ls) > 0 {
				if sec, err := strconv.ParseInt(ls[len(ls)-1], 10, 64); err == nil {
					t = time.Unix(sec, 0).UTC()
				}
			}
		}
	}
	w.mu.Lock()
	w.born[body] = t
	w.mu.Unlock()
	return t, !t.IsZero()
}
