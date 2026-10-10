package onto

import (
	"os"
	"strings"
)

// ownerLine mirrors world.Owner() against the onto graph and returns the
// winning creature, or "" with no owner. Kept as a helper so CLI can wire it
// with its `case` + stdout conventions (cli.go is the dispatch surface).
func ownerLine(root, p string) (string, error) {
	if root == "" {
		root = "."
		if v := os.Getenv("ISEKAI_ROOT"); v != "" {
			root = v
		}
	}
	w, err := Load(root)
	if err != nil {
		return "", err
	}
	return ownerOf(w, normPath(p)), nil
}

// ownerOf is the census-side mirror of world.Owner(). Longest matching
// territory wins (same under() semantics as the write gate); on a tie, an
// authoring rank (rimuru/elf/slime) beats a gate-holding one, mirroring
// world/rank.go's Authors table. Lives in onto/ because world imports onto —
// a reverse import would cycle.
func ownerOf(w *World, p string) string {
	bestLen := -1
	var candidates []string
	for _, tr := range w.Graph.triples {
		if tr.P != pOwns {
			continue
		}
		owner, ter := tr.S.Value, tr.O.Value
		if !(p == ter || strings.HasPrefix(p, ter+"/")) {
			continue
		}
		if len(ter) > bestLen {
			bestLen, candidates = len(ter), []string{strings.TrimPrefix(owner, NS)}
		} else if len(ter) == bestLen {
			candidates = append(candidates, strings.TrimPrefix(owner, NS))
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	for _, name := range candidates {
		if isAuthoringRank(name) {
			return name
		}
	}
	return candidates[0]
}

// isAuthoringRank maps a creature name to whether its rank authors territory
// (rimuru/elf/slime true; orc/kijin/etc. false) — a static mirror of
// world/rank.go rank.<name>.Authors. Name prefix before the first '-' is the
// rank word (slime-auth → slime).
func isAuthoringRank(name string) bool {
	rank := name
	if i := strings.IndexByte(name, '-'); i > 0 {
		rank = name[:i]
	}
	switch rank {
	case "rimuru", "elf", "slime":
		return true
	}
	return false
}
