package world

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Kaginari/isekai/config/yaml"
)

// DimWorld is one world the manifest names.
type DimWorld struct {
	Name string
	Path string
	Role string
}

// DimRelation is one typed bond between two manifest worlds ("from -> to : type", ask how).
type DimRelation struct {
	From string
	To   string
	Type string
	Ask  string
}

// Dimension is the parsed `<root>/.isekai/dimension.yaml`: the worlds that work together and
// their typed relations. The manifest is world soft-code, of one class with the law.
type Dimension struct {
	Name      string
	Above     string // "" when the manifest says null (a top-level dimension)
	Worlds    []DimWorld
	Relations []DimRelation
}

// LoadDimension parses the manifest at path (the caller joins `<root>/<worldDir>/dimension.yaml`).
// A missing file, malformed yaml or a wrong shape is an error for the caller to note — the
// world still opens (the LoadLaw missing-file precedent: a hole surfaces, never a failed open).
// Anything outside name/above/worlds/relations is left unread.
func LoadDimension(path string) (*Dimension, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("dimension manifest: %v", err)
	}
	n, err := yaml.Parse(b, path)
	if err != nil {
		return nil, err
	}
	if n.Kind != yaml.Map {
		return nil, fmt.Errorf("dimension manifest: expected a map at the top, got %s (%s)", n.Kind, n.Where())
	}
	d := &Dimension{}
	if s, err := scalarAt(n, "name"); err != nil {
		return nil, err
	} else {
		d.Name = s
	}
	if s, err := scalarAt(n, "above"); err != nil {
		return nil, err
	} else {
		d.Above = s
	}
	if l := n.Get("worlds"); l != nil {
		if l.Kind != yaml.List {
			return nil, fmt.Errorf("dimension manifest: `worlds` must be a list (%s)", l.Where())
		}
		for _, it := range l.Items {
			if it.Kind != yaml.Map {
				return nil, fmt.Errorf("dimension manifest: `worlds` items must be maps (%s)", it.Where())
			}
			var dw DimWorld
			if dw.Name, err = scalarAt(it, "name"); err != nil {
				return nil, err
			}
			if dw.Path, err = scalarAt(it, "path"); err != nil {
				return nil, err
			}
			if dw.Role, err = scalarAt(it, "role"); err != nil {
				return nil, err
			}
			d.Worlds = append(d.Worlds, dw)
		}
	}
	if l := n.Get("relations"); l != nil {
		if l.Kind != yaml.List {
			return nil, fmt.Errorf("dimension manifest: `relations` must be a list (%s)", l.Where())
		}
		for _, it := range l.Items {
			if it.Kind != yaml.Map {
				return nil, fmt.Errorf("dimension manifest: `relations` items must be maps (%s)", it.Where())
			}
			var dr DimRelation
			if dr.From, err = scalarAt(it, "from"); err != nil {
				return nil, err
			}
			if dr.To, err = scalarAt(it, "to"); err != nil {
				return nil, err
			}
			if dr.Type, err = scalarAt(it, "type"); err != nil {
				return nil, err
			}
			if dr.Ask, err = scalarAt(it, "ask"); err != nil {
				return nil, err
			}
			d.Relations = append(d.Relations, dr)
		}
	}
	return d, nil
}

// scalarAt reads key's plain scalar text ("" for absent or null), refusing a container.
func scalarAt(m *yaml.Node, key string) (string, error) {
	v := m.Get(key)
	if v == nil || v.Kind == yaml.Null {
		return "", nil
	}
	if !v.Scalar() {
		return "", fmt.Errorf("dimension manifest: `%s` must be a plain scalar (%s)", key, v.Where())
	}
	return v.Text(), nil
}

// GitSelfInit versions root when nothing has: `git rev-parse --is-inside-work-tree` with
// cmd.Dir = root (the memory.GitHead precedent, memory/files.go) — an error or "false" means
// no work-tree, so it runs a bare `git init -b main`. Returns whether the init ran. A git
// failure is an error for the caller to note, never a failed open. Bare init only: no
// `.gitignore` is seeded (a guessed ignore risks committing secrets; the human authors ignores
// before the first commit), and no commit is made — the first commit is the human's act.
// Nested dirs with their own `.git` (sovereign clones) stay sovereign: init without add never
// recurses, absorbs or submodules them.
func GitSelfInit(root string) (bool, error) {
	probe := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	probe.Dir = root
	out, err := probe.Output()
	if err == nil && strings.TrimSpace(string(out)) == "true" {
		return false, nil
	}
	init := exec.Command("git", "init", "-b", "main")
	init.Dir = root
	if out, err := init.CombinedOutput(); err != nil {
		return false, fmt.Errorf("git init: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return true, nil
}
