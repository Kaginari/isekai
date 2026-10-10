package board

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRelatesCoupleImportsAndSharedMinds(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/w\n\ngo 1.22\n")
	write("store/db.go", "package store\n\nimport (\n\t\"fmt\"\n\t\"example.com/w/entity\"\n)\n\nvar _ = fmt.Sprint\nvar _ = entity.X\n")
	write("store/more.go", "package store\n\nimport \"example.com/w/entity\"\n\nvar _ = entity.X\n")
	write("store/db_test.go", "package store\n\nimport \"example.com/w/web\"\n\nvar _ = web.Y\n")
	write("entity/e.go", "package entity\n\nvar X = 1\n")
	write("web/ui/app.js", "import { y } from '../lib/util.js';\nconst z = require('./local');\n")
	write("web/lib/util.js", "export const y = 1;\n")
	write("cli/main.py", "from tools import helper\nimport os\n")
	write("tools/helper.py", "x = 1\n")
	nodes := []Node{
		{ID: "rimuru", Kind: "creature", Rank: "rimuru", Throne: true},
		{ID: "orc-data", Kind: "creature", Rank: "orc", Parent: "rimuru", Owns: []string{"store", "entity"}},
		{ID: "slime-db", Kind: "creature", Rank: "slime", Parent: "orc-data", Owns: []string{"store"}},
		{ID: "slime-entities", Kind: "creature", Rank: "slime", Parent: "orc-data", Owns: []string{"Entity"}},
		{ID: "slime-ui", Kind: "creature", Rank: "slime", Parent: "rimuru", Owns: []string{"web/ui"}},
		{ID: "slime-lib", Kind: "creature", Rank: "slime", Parent: "rimuru", Owns: []string{"web/lib/*.js"}},
		{ID: "slime-cli", Kind: "creature", Rank: "slime", Parent: "rimuru", Owns: []string{"cli"}},
		{ID: "slime-tools", Kind: "creature", Rank: "slime", Parent: "rimuru", Owns: []string{"tools"}},
		{ID: "mind-core", Kind: "mind"},
	}
	edges := []Edge{
		{From: "slime-ui", To: "mind-core", Bond: "wears"},
		{From: "slime-db", To: "mind-core", Bond: "wears"},
		{From: "orc-data", To: "mind-core", Bond: "wears"}, // the slime's ancestor: no relate
	}
	got := map[string]Relate{}
	for _, r := range Relates(root, nodes, edges) {
		got[r.From+"~"+r.To] = r
	}
	if r := got["slime-db~slime-entities"]; r.Imports != 2 || r.Weight != 1 {
		t.Errorf("db ↔ entities: want 2 imports at weight 1 (the strongest), got %+v", r)
	}
	if r := got["slime-lib~slime-ui"]; r.Imports != 1 {
		t.Errorf("a relative JS import through a glob-owned path: want 1 import, got %+v", r)
	}
	if r := got["slime-cli~slime-tools"]; r.Imports != 1 {
		t.Errorf("a Python import: want 1, got %+v", r)
	}
	if r := got["slime-db~slime-ui"]; r.Minds != 1 || r.Imports != 0 {
		t.Errorf("a shared mind (the test file's import left out): want 1 mind and no import, got %+v", r)
	}
	for k := range got {
		if k == "orc-data~slime-db" || k == "orc-data~slime-entities" {
			t.Errorf("an ancestor pair is the tree's business, not a relate: %s", k)
		}
	}
	if rs := Relates(t.TempDir(), nil, nil); rs == nil || len(rs) != 0 {
		t.Errorf("an empty world: want an empty, non-nil list, got %#v", rs)
	}
}
