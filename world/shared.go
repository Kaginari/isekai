package world

import (
	"path/filepath"

	"github.com/Kaginari/isekai/queue"
)

// The shared shelf (`<worldDir>/shared/`): the world's own bookkeeping that every session,
// every body and the human read — the work queue first. It is harness soil: the harness
// writes it (the queue's moves go through the queue tool or the q command, never a body's
// write), so the territory gate refuses a body's write there and the tree watch never bills
// a change there to the body whose turn it fell in (loop/snapshot.go).
const sharedShelf = "shared"

// SharedDir is the shared shelf's directory.
func (w *World) SharedDir() string { return filepath.Join(w.Dir(), sharedShelf) }

// QueuePath is the shared work queue's file.
func (w *World) QueuePath() string { return filepath.Join(w.SharedDir(), "queue.jsonl") }

// Queue opens the shared work queue under the rank law: a body moves the work it holds, and
// any creature up the holder's Reports-to chain (rimuru tops every chain) may move it too.
func (w *World) Queue() *queue.Store { return queue.Open(w.QueuePath(), w.Outranks) }

// Outranks reports whether `by` stands above `owner` in the Reports-to chain. Work held by no
// one is the throne's to move.
func (w *World) Outranks(by, owner string) bool {
	if by == Rimuru {
		return true
	}
	if owner == "" || by == "" {
		return false
	}
	for _, up := range w.lineOf(owner) {
		if up == by {
			return true
		}
	}
	return false
}
