package world

import (
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Kaginari/isekai/onto"
)

// Self-heal: the committed docs are the store, the roster is their cache. A doc another hand
// moved — a second session's relief, a human's edit, a checkout — left the cache standing until
// the next gate re-read it, so a dispatch could still mint a court for a retired seat or refuse a
// body whose claim had grown. The roster keeps a stamp of what it was derived from (every rank
// dir's seats and their .md files: name, size, mtime); SelfHeal compares the stamp and re-derives
// when the docs moved. The decisions that act on the roster ask it first: the dispatch preflight,
// an engine's build, a write's territory policy, and a roster miss.
//
// Healing never moves trust. The verify lines the gate runs as "before the turn" (verifyBase)
// change only at an explicit Reload (Open, the gate, the harness's own mints); a heal adds the
// lines of a body it had never seen and leaves every known body's lines as they were — a body
// cannot pass its gate by rewriting its check and waiting for a heal.

// docStamp fingerprints the creature docs on disk.
func (w *World) docStamp() uint64 {
	h := fnv.New64a()
	for _, d := range rankDirs(w.Ranks) {
		base := filepath.Join(w.Dir(), d)
		seats, err := os.ReadDir(base)
		if err != nil {
			fmt.Fprintf(h, "%s:-\n", d)
			continue
		}
		for _, s := range seats {
			if !s.IsDir() {
				continue
			}
			files, _ := os.ReadDir(filepath.Join(base, s.Name()))
			var mds []string
			for _, f := range files {
				if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
					continue
				}
				if fi, err := f.Info(); err == nil {
					mds = append(mds, fmt.Sprintf("%s:%d:%d", f.Name(), fi.Size(), fi.ModTime().UnixNano()))
				}
			}
			sort.Strings(mds)
			fmt.Fprintf(h, "%s/%s:%s\n", d, s.Name(), strings.Join(mds, ","))
		}
	}
	return h.Sum64()
}

// SelfHeal re-derives the roster when the docs moved since it was read; true when it did.
func (w *World) SelfHeal() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.selfHealLocked()
}

// selfHealLocked is SelfHeal with w.mu held.
func (w *World) selfHealLocked() bool {
	st := w.docStamp()
	if st == w.stamp {
		return false
	}
	ow, err := onto.LoadLayout(w.Root, w.Ranks.Layout(w.Lex))
	if err != nil {
		return false // the cache stands until the docs parse again
	}
	w.mergeFreshLocked(ow)
	w.stamp = st
	return true
}

// mergeFreshLocked puts a fresh derivation on the roster: the docs' creatures, then the foreign
// bodies the docs do not name. A body the roster never held brings its verify lines into the
// gate's base; a known body's lines stay as the last Reload read them.
func (w *World) mergeFreshLocked(ow *onto.World) {
	w.Onto = ow
	w.Creatures = creaturesOf(w.Root, ow, w.Ranks)
	for _, f := range w.Foreign {
		if w.creatureLocked(f.Name) == nil {
			w.Creatures = append(w.Creatures, f)
		}
	}
	if w.verifyBase == nil {
		w.verifyBase = map[string][]string{}
	}
	for _, c := range w.Creatures {
		if _, known := w.verifyBase[c.Name]; !known {
			w.verifyBase[c.Name] = append([]string(nil), c.Verify...)
		}
	}
}
