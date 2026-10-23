package onto

import (
	"encoding/binary"
	"hash/fnv"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Kaginari/isekai/discover"
)

// Stamp fingerprints every input LoadLayout reads — the ontology's turtle, the rank dirs'
// creature docs, the skill bases' minds — by path, size and modification time, directories
// included (an entry added or removed moves its dir's time). Equal stamps load equal worlds; a
// walk of stats costs a small fraction of a parse, derive and reason.
func Stamp(root string, l Layout) uint64 {
	l = l.norm()
	roots := []string{OntologyDirIn(root, l)}
	for _, r := range l.Ranks {
		if r.Dir != "" {
			roots = append(roots, filepath.Join(root, l.WorldDir, r.Dir))
		}
	}
	for _, b := range discover.SkillBases(l.WorldDir) {
		roots = append(roots, filepath.Join(root, b))
	}
	h := fnv.New64a()
	var buf [16]byte
	for _, r := range roots {
		_ = filepath.WalkDir(r, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // an absent root is a stamp like any other: absent
			}
			if !d.IsDir() && !strings.HasSuffix(p, ".md") && !strings.HasSuffix(p, ".ttl") {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			h.Write([]byte(p))
			binary.LittleEndian.PutUint64(buf[:8], uint64(info.Size()))
			binary.LittleEndian.PutUint64(buf[8:], uint64(info.ModTime().UnixNano()))
			h.Write(buf[:])
			return nil
		})
	}
	return h.Sum64()
}

var cache struct {
	mu sync.Mutex
	at map[string]cached
}

type cached struct {
	stamp uint64
	w     *World
}

// Cached is LoadLayout memoized on Stamp: the boards poll it every second, and the world it
// answers moves only when a doc or a turtle file does. The World is shared — read it, never add
// to it (a fact is recorded on a world of its own, from LoadLayout).
func Cached(root string, l Layout) (*World, error) {
	key := root + "\x00" + strings.Join(append([]string{l.WorldDir, l.Law}, rankKeys(l.Ranks)...), "\x00")
	stamp := Stamp(root, l)
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if c, ok := cache.at[key]; ok && c.stamp == stamp {
		return c.w, nil
	}
	w, err := LoadLayout(root, l)
	if err != nil {
		return nil, err
	}
	if cache.at == nil {
		cache.at = map[string]cached{}
	}
	cache.at[key] = cached{stamp, w}
	return w, nil
}

func rankKeys(rs []RankDir) []string {
	var out []string
	for _, r := range rs {
		out = append(out, r.Name+"/"+r.Dir+"/"+r.Parent)
	}
	return out
}
