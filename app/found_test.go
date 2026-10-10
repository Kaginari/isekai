package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaginari/isekai/onto"
)

// project writes a fixture tree: path → content ("" writes a placeholder line).
func project(t *testing.T, files map[string]string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "proj")
	for f, body := range files {
		p := filepath.Join(root, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if body == "" {
			body = "x\n"
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func tree(t *testing.T, files ...string) string {
	t.Helper()
	m := map[string]string{}
	for _, f := range files {
		m[f] = ""
	}
	return project(t, m)
}

// goFile is a Go source file of about n lines: a doc, imports, exported funcs.
func goFile(pkg, doc string, n int, imports ...string) string {
	var b strings.Builder
	if doc != "" {
		fmt.Fprintf(&b, "// Package %s %s\n", pkg, doc)
	}
	fmt.Fprintf(&b, "package %s\n\n", pkg)
	for _, im := range imports {
		fmt.Fprintf(&b, "import %q\n", im)
	}
	for i := 0; b.Len() == 0 || strings.Count(b.String(), "\n") < n; i++ {
		fmt.Fprintf(&b, "\n// F%d does a thing.\nfunc F%d() int {\n\treturn %d\n}\n", i, i, i)
	}
	if pkg == "main" {
		b.WriteString("\nfunc main() {}\n")
	}
	return b.String()
}

func goTest(pkg string) string {
	return "package " + pkg + "\n\nimport \"testing\"\n\nfunc TestIt(t *testing.T) {}\n"
}

// pad is n lines of filler for a non-Go source file.
func pad(prefix string, n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "%s%d\n", prefix, i)
	}
	return b.String()
}

// sketchAt founds a medium colony at root and returns the sketch and the write result.
func sketchAt(t *testing.T, root string) (sketch, sketchWrite) {
	t.Helper()
	if _, err := Init("isekai", root); err != nil {
		t.Fatal(err)
	}
	d, _ := parseDist("isekai")
	lex := lexiconFor("isekai")
	s := sketchColony(root, d)
	res, err := writeSketch(root, d, s, lex.Ranks, lex.Prefixes, lex.Display)
	if err != nil {
		t.Fatal(err)
	}
	return s, res
}

// sound asserts the colony the ontology reads is whole: no finding, nothing housekeep would
// rewire, every claim derives under its creature's name, the ranks' fan-out bounded.
func sound(t *testing.T, root string, s sketch) {
	t.Helper()
	w, err := onto.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if f := w.Validate(); len(f) != 0 {
		t.Fatalf("onto check: %v\n%s", f, s.table(lexiconFor("isekai").Prefixes))
	}
	rep, left, err := onto.Housekeep(root, false, false)
	if err != nil || len(rep.DeadBonds) != 0 || len(left) != 0 {
		t.Fatalf("housekeep would touch the sketch: dead bonds %v, left %v, %v", rep.DeadBonds, left, err)
	}
	for _, p := range s.all() {
		for _, c := range p.Territory {
			if !zoneOK(p.Short, c) {
				t.Errorf("%s-%s claims %s, which the ontology does not derive under that name", p.Rank, p.Short, c)
			}
		}
		if p.Rank == "slime" {
			for _, c := range p.Territory {
				if strings.Contains(c, "/") && !strings.HasSuffix(c, "/") {
					t.Errorf("slime-%s claims the nested file %s: territories are directories", p.Short, c)
				}
			}
		}
	}
	for _, o := range s.Orcs {
		if len(o.Slimes) == 0 {
			t.Errorf("orc-%s carries no slime", o.Short)
		}
	}
	if len(s.Elves) > 1 {
		for _, e := range s.Elves[1:] {
			if n := len(e.Orcs); n == 0 || n > orcsPerElf {
				t.Errorf("lane elf-%s carries %d orcs", e.Short, n)
			}
		}
	}
}

func byName(s sketch) map[string]plan {
	m := map[string]plan{}
	for _, p := range s.all() {
		m[p.Rank+"-"+p.Short] = p
	}
	return m
}

func readDoc(t *testing.T, root, rank, short string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".isekai", rank, short, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// A small Go module: one elf, one orc, a slime per package dir plus the root's own files;
// verify lines only where tests are.
func TestMediumSmallGoModule(t *testing.T) {
	root := project(t, map[string]string{
		"go.mod":                         "module example.com/tiny\n\ngo 1.22\n",
		"README.md":                      "# tiny\n",
		"main.go":                        goFile("main", "", 30, "example.com/tiny/internal/greet"),
		"internal/greet/greet.go":        goFile("greet", "says hello to whoever asks.", 140),
		"internal/greet/greet_test.go":   goTest("greet"),
		"internal/store/store.go":        goFile("store", "keeps the greetings on disk.", 160, "example.com/tiny/internal/greet"),
		"vendor/x/y.go":                  goFile("y", "", 400),
		"internal/store/testdata/fix.go": goFile("fix", "", 400),
	})
	s, res := sketchAt(t, root)
	sound(t, root, s)
	if len(s.Elves) != 1 || len(s.Orcs) != 1 || s.Kijin != nil {
		t.Fatalf("a small module is one elf and one orc, no kijin:\n%s", s.table(lexiconFor("isekai").Prefixes))
	}
	m := byName(s)
	greet, store, entry := m["slime-internal-greet"], m["slime-internal-store"], m["slime-proj"]
	if len(greet.Territory) != 1 || greet.Territory[0] != "internal/greet/" || len(store.Territory) != 1 || store.Territory[0] != "internal/store/" {
		t.Fatalf("one slime per package dir: %+v / %+v", greet.Territory, store.Territory)
	}
	if strings.Join(entry.Territory, ",") != "main.go" {
		t.Fatalf("the root's own files are the entry slime's: %+v", entry.Territory)
	}
	if len(greet.Verify) != 1 || greet.Verify[0] != "go test ./internal/greet/..." {
		t.Errorf("a tested package verifies with go test: %v", greet.Verify)
	}
	if len(store.Verify) != 0 || len(entry.Verify) != 0 {
		t.Errorf("no tests, no verify line: store %v entry %v", store.Verify, entry.Verify)
	}
	if e := s.Elves[0]; strings.Join(e.Verify, ",") != "go build ./...,go vet ./..." || !contains(e.Territory, "go.mod") || !contains(e.Territory, "README.md") {
		t.Errorf("the elf holds the manifests and builds the module: %+v", e)
	}
	for _, p := range s.all() {
		for _, c := range p.Territory {
			if strings.HasPrefix(c, "vendor") || strings.Contains(c, "testdata") {
				t.Errorf("vendored and fixture dirs are no one's: %s", c)
			}
		}
	}
	doc := readDoc(t, root, "slime", "internal-greet")
	for _, want := range []string{"- **Rank:** Slime", "- **Territory:** `internal/greet/`", "- **Reports to:** orc-", "- **Minds:** none",
		"says hello to whoever asks", "- sketch: Go package greet: 1 source file, 1 test file", "- sketch: Exports F0, F1", "Imported by slime-",
		"## Verify\n- `go test ./internal/greet/...`"} {
		if !strings.Contains(doc, want) {
			t.Errorf("slime doc lacks %q:\n%s", want, doc)
		}
	}
	if doc := readDoc(t, root, "slime", "internal-store"); !strings.Contains(doc, "No tests found in this territory") || strings.Contains(doc, "## Verify") {
		t.Errorf("an untested slime says so and carries no verify:\n%s", doc)
	}
	if doc := readDoc(t, root, "slime", "proj"); !strings.Contains(doc, "Entry point: main.go") {
		t.Errorf("the entry point is observed:\n%s", doc)
	}
	if len(res.wrote) != len(s.all()) || len(res.kept) != 0 {
		t.Errorf("a fresh world writes every creature: %+v", res)
	}
}

// A larger multi-package Go repository: imports cluster the packages into orcs, lane elves and
// a canon elf above them, a kijin for the CI; a tiny cmd/ folds into what it imports.
func TestMediumMultiPackageGo(t *testing.T) {
	const mod = "example.com/big/"
	files := map[string]string{
		"go.mod":                   "module example.com/big\n",
		"go.sum":                   "",
		"README.md":                "# big\n",
		"LICENSE":                  "",
		"docs/guide.md":            "",
		".github/workflows/ci.yml": "",
		"Dockerfile":               "",
		"cmd/big/main.go":          goFile("main", "", 10, mod+"app"),
	}
	add := func(pkg string, tested bool, imports ...string) {
		var ims []string
		for _, im := range imports {
			ims = append(ims, mod+im)
		}
		files[pkg+"/"+filepath.Base(pkg)+".go"] = goFile(filepath.Base(pkg), "is the "+pkg+" part.", 300, ims...)
		files[pkg+"/more.go"] = goFile(filepath.Base(pkg), "", 200, ims...)
		if tested {
			files[pkg+"/"+filepath.Base(pkg)+"_test.go"] = goTest(filepath.Base(pkg))
		}
	}
	add("loop", true, "provider", "wire", "compact")
	add("provider", true, "wire")
	add("wire", true)
	add("compact", false, "wire", "provider")
	add("guard", true)
	add("gate", true, "guard")
	add("sandbox", true, "guard")
	add("shell", false, "sandbox", "guard")
	add("tool", true, "shell", "sandbox", "gate", "guard")
	add("tui", true)
	add("web", true, "tui")
	add("board", false, "web", "tui")
	add("config", true)
	add("discover", true, "config")
	add("app", true, "loop", "tool", "tui", "web", "board", "config", "discover", "provider")
	root := project(t, files)
	s, _ := sketchAt(t, root)
	sound(t, root, s)
	m := byName(s)
	if len(s.Elves) < 3 || s.Elves[0].Short != "canon" {
		t.Fatalf("a large project gets a canon elf and lane elves:\n%s", s.table(lexiconFor("isekai").Prefixes))
	}
	if c := s.Elves[0]; !contains(c.Territory, "docs/") || !contains(c.Territory, "go.mod") || len(c.Orcs) != len(s.Elves)-1 {
		t.Errorf("the canon elf holds docs and manifests and speaks for the lanes: %+v", c)
	}
	if app := m["slime-app"]; strings.Join(app.Territory, ",") != "app/,cmd/" {
		t.Errorf("tiny cmd/ folds into app/, which it imports: %+v", app.Territory)
	}
	orcOf := map[string]string{}
	for _, o := range s.Orcs {
		if len(o.Slimes) > maxPerOrc {
			t.Errorf("orc-%s carries %d slimes", o.Short, len(o.Slimes))
		}
		for _, sl := range o.Slimes {
			orcOf[sl.Short] = o.Short
		}
	}
	for _, pair := range [][2]string{{"loop", "provider"}, {"provider", "wire"}, {"gate", "guard"}, {"sandbox", "shell"}, {"tui", "web"}, {"config", "discover"}} {
		if orcOf[pair[0]] != orcOf[pair[1]] {
			t.Errorf("%s and %s are bound by imports but split: %v", pair[0], pair[1], orcOf)
		}
	}
	if orcOf["loop"] == orcOf["tui"] || orcOf["guard"] == orcOf["tui"] {
		t.Errorf("unrelated packages share an orc: %v", orcOf)
	}
	if k := s.Kijin; k == nil || !contains(k.Territory, ".github/") || !contains(k.Territory, "Dockerfile") || k.Short != "ci" {
		t.Errorf("CI and a Dockerfile make a kijin-ci: %+v", s.Kijin)
	}
	if v := m["slime-shell"].Verify; len(v) != 0 {
		t.Errorf("shell has no tests: %v", v)
	}
	if v := m["slime-wire"].Verify; strings.Join(v, "") != "go test ./wire/..." {
		t.Errorf("wire is tested: %v", v)
	}
	for _, o := range s.Orcs {
		for _, v := range o.Verify {
			if strings.Count(v, "go test") != 1 {
				t.Errorf("an orc folds its slimes' go tests into one line: %q", v)
			}
		}
	}
}

// A JS workspace: one slime per workspace package, named so the ontology derives each claim,
// verify lines only for packages with a real test script.
func TestMediumJSWorkspace(t *testing.T) {
	root := project(t, map[string]string{
		"package.json":                     `{"name":"acme","private":true,"workspaces":["packages/*","apps/*"]}`,
		"README.md":                        "",
		"packages/ui/package.json":         `{"name":"@acme/ui","description":"The design system's components.","scripts":{"test":"vitest run"}}`,
		"packages/ui/src/index.ts":         "export function Button() {}\n" + pad("// ", 200),
		"packages/ui/src/button.test.ts":   "test('x', () => {})\n",
		"packages/core/package.json":       `{"name":"@acme/core","scripts":{"test":"vitest run"},"dependencies":{"@acme/ui":"*"}}`,
		"packages/core/src/index.ts":       "import { Button } from '@acme/ui'\nexport const core = 1\n" + pad("// ", 200),
		"packages/core/src/core.test.ts":   "test('x', () => {})\n",
		"packages/config/package.json":     `{"name":"@acme/config","scripts":{"test":"echo \"Error: no test specified\" && exit 1"}}`,
		"packages/config/index.js":         "module.exports = {}\n",
		"apps/web/package.json":            `{"name":"web","scripts":{"test":"vitest run"},"dependencies":{"@acme/ui":"*","@acme/core":"*"}}`,
		"apps/web/src/main.tsx":            "import { core } from '@acme/core'\n" + pad("// ", 300),
		"apps/web/src/main.test.tsx":       "test('x', () => {})\n",
		"packages/ui/node_modules/x/i.js":  pad("// ", 500),
		"packages/ui/dist/index.js":        pad("// ", 500),
		"apps/web/src/generated/schema.ts": "// Code generated by x. DO NOT EDIT.\n" + pad("// ", 500),
	})
	s, _ := sketchAt(t, root)
	sound(t, root, s)
	m := byName(s)
	for name, want := range map[string]string{"packages-ui": "npm test --workspace packages/ui", "packages-core": "npm test --workspace packages/core", "apps-web": "npm test --workspace apps/web"} {
		p, ok := m["slime-"+name]
		if !ok || strings.Join(p.Verify, "|") != want {
			t.Errorf("slime-%s verify %v, want %q\n%s", name, p.Verify, want, s.table(lexiconFor("isekai").Prefixes))
		}
	}
	if p, ok := m["slime-packages-config"]; !ok || len(p.Verify) != 0 {
		t.Errorf("npm's placeholder test script is no verify line: %+v", p)
	}
	if !strings.Contains(m["slime-packages-ui"].Purpose, "design system's components") {
		t.Errorf("the purpose comes from the package's description: %q", m["slime-packages-ui"].Purpose)
	}
	if o := m["orc-packages"]; len(o.Slimes) != 3 {
		t.Errorf("the packages share an orc named from their root: %+v", o)
	}
	if !strings.Contains(strings.Join(m["slime-apps-web"].Traits, " "), "Imports from slime-packages-core, slime-packages-ui") {
		t.Errorf("workspace dependencies are bonds: %v", m["slime-apps-web"].Traits)
	}
	if !strings.Contains(strings.Join(m["slime-apps-web"].Traits, " "), "1 source file, 1 test file") {
		t.Errorf("generated code is not counted: %v", m["slime-apps-web"].Traits)
	}
}

// A Python package: the package and its tests are slimes; pytest only where test files are.
func TestMediumPythonPackage(t *testing.T) {
	root := project(t, map[string]string{
		"pyproject.toml":            "[project]\nname = \"acme\"\n",
		"README.md":                 "",
		"src/acme/__init__.py":      "\"\"\"Acme turns invoices into ledgers.\"\"\"\n",
		"src/acme/core/__init__.py": "",
		"src/acme/core/engine.py":   "from acme.io import reader\n\ndef run():\n    pass\n" + pad("# ", 200),
		"src/acme/io/__init__.py":   "",
		"src/acme/io/reader.py":     "class Reader:\n    pass\n" + pad("# ", 200),
		"tests/conftest.py":         "",
		"tests/test_engine.py":      "from acme.core import engine\n\ndef test_run():\n    engine.run()\n" + pad("# ", 150),
		".venv/lib/x.py":            pad("# ", 900),
	})
	s, _ := sketchAt(t, root)
	sound(t, root, s)
	m := byName(s)
	acme, tests := m["slime-acme"], m["slime-tests"]
	if strings.Join(acme.Territory, ",") != "src/acme/" || strings.Join(tests.Territory, ",") != "tests/" {
		t.Fatalf("the package and its tests:\n%s", s.table(lexiconFor("isekai").Prefixes))
	}
	if !strings.Contains(acme.Purpose, "Acme turns invoices into ledgers") {
		t.Errorf("the package docstring is the purpose: %q", acme.Purpose)
	}
	if len(acme.Verify) != 0 || strings.Join(tests.Verify, "") != "python -m pytest -q tests" {
		t.Errorf("pytest only where tests are: acme %v tests %v", acme.Verify, tests.Verify)
	}
	if !strings.Contains(strings.Join(tests.Traits, " "), "Imports from slime-acme") {
		t.Errorf("the tests' imports bind them to the package: %v", tests.Traits)
	}
	for _, p := range s.all() {
		if strings.Contains(strings.Join(p.Territory, ","), ".venv") {
			t.Errorf("a virtualenv is no one's: %+v", p)
		}
	}
}

// A mixed repo with CI, release tooling and infra: the kijin holds them, never a slime; docs
// stay with the elf; each language gets its own test runner.
func TestMediumMixedRepoWithCI(t *testing.T) {
	root := project(t, map[string]string{
		"go.mod":                        "module example.com/mixed\n",
		"package.json":                  `{"scripts":{"test":"jest"}}`,
		"README.md":                     "",
		"CHANGELOG.md":                  "",
		".goreleaser.yaml":              "",
		".github/workflows/release.yml": "",
		"Dockerfile":                    "",
		"deploy/k8s.yaml":               "",
		"deploy/hook.py":                pad("# ", 200),
		"docs/index.md":                 "",
		"docs/site.js":                  pad("// ", 300),
		"server/server.go":              goFile("server", "serves the API.", 300),
		"server/server_test.go":         goTest("server"),
		"web/app.js":                    "export function App() {}\n" + pad("// ", 300),
		"web/app.test.js":               "test('x', () => {})\n",
		"scripts/migrate.py":            "def migrate():\n    pass\n" + pad("# ", 200),
	})
	s, _ := sketchAt(t, root)
	sound(t, root, s)
	m := byName(s)
	k := s.Kijin
	if k == nil || k.Short != "release" {
		t.Fatalf("CI with release tooling is kijin-release: %+v", k)
	}
	for _, want := range []string{".github/", ".goreleaser.yaml", "CHANGELOG.md", "Dockerfile", "deploy/"} {
		if !contains(k.Territory, want) {
			t.Errorf("kijin lacks %s: %v", want, k.Territory)
		}
	}
	for _, p := range s.all() {
		if p.Rank == "slime" && (contains(p.Territory, "deploy/") || contains(p.Territory, "docs/")) {
			t.Errorf("infra and docs are never a slime's: %+v", p)
		}
	}
	if !contains(s.Elves[0].Territory, "docs/") {
		t.Errorf("docs are the elf's: %v", s.Elves[0].Territory)
	}
	if v := strings.Join(m["slime-server"].Verify, ""); v != "go test ./server/..." {
		t.Errorf("server: %q", v)
	}
	if v := strings.Join(m["slime-web"].Verify, ""); v != "npx jest web" {
		t.Errorf("web is scoped through the root's jest: %q", v)
	}
	if v := m["slime-scripts"].Verify; len(v) != 0 {
		t.Errorf("scripts has no tests: %v", v)
	}
	if !strings.Contains(strings.Join(k.Traits, " "), "release.yml") {
		t.Errorf("the kijin names the workflows it saw: %v", k.Traits)
	}
}

// Medium is idempotent and safe: a second run writes nothing, a refined doc is never touched,
// and over a colony someone already cut it adds only what nobody holds.
func TestMediumIdempotentAndSafe(t *testing.T) {
	root := project(t, map[string]string{
		"go.mod":            "module example.com/m\n",
		"api/api.go":        goFile("api", "serves.", 300),
		"store/store.go":    goFile("store", "stores.", 300),
		"jobs/jobs.go":      goFile("jobs", "runs jobs.", 300),
		"jobs/jobs_test.go": goTest("jobs"),
	})
	s, first := sketchAt(t, root)
	sound(t, root, s)
	refined := filepath.Join(root, ".isekai", "slime", "store", "README.md")
	mine := "# slime-store\n\n- **Rank:** Slime\n- **Territory:** `store/`\n- **Reports to:** " + s.Orcs[0].Slimes[0].Parent + "\n- **Purpose:** mine\n\n## Traits\n- refined by hand\n"
	if err := os.WriteFile(refined, []byte(mine), 0o644); err != nil {
		t.Fatal(err)
	}
	_, again := sketchAt(t, root)
	if len(again.wrote) != 0 || len(again.kept) != len(first.wrote) {
		t.Fatalf("a re-run writes nothing and keeps all: %+v (first %d)", again, len(first.wrote))
	}
	if b, _ := os.ReadFile(refined); string(b) != mine {
		t.Fatalf("a refined doc was touched:\n%s", b)
	}
	// a world whose own colony already holds api/ under another name: the sketch's slime-api
	// is left out (covered), the rest is added, and the whole still checks clean
	root2 := project(t, map[string]string{
		"go.mod":         "module example.com/m\n",
		"api/api.go":     goFile("api", "serves.", 300),
		"store/store.go": goFile("store", "stores.", 300),
	})
	if _, err := Init("isekai", root2); err != nil {
		t.Fatal(err)
	}
	for p, body := range map[string]string{
		".isekai/elf/team/README.md":     "# elf-team\n\n- **Rank:** Elf\n- **Reports to:** rimuru\n- **Purpose:** ours\n",
		".isekai/orc/http/README.md":     "# orc-http\n\n- **Rank:** Orc\n- **Territory:** `api/`\n- **Reports to:** elf-team\n- **Purpose:** ours\n",
		".isekai/slime/server/README.md": "# slime-server\n\n- **Rank:** Slime\n- **Territory:** `api/`\n- **Reports to:** orc-http\n- **Purpose:** ours\n",
	} {
		_ = os.MkdirAll(filepath.Dir(filepath.Join(root2, p)), 0o755)
		if err := os.WriteFile(filepath.Join(root2, p), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s2, res := sketchAt(t, root2)
	if !contains(res.covered, "slime-api") {
		t.Fatalf("api/ is already held: %+v", res)
	}
	for _, w := range res.wrote {
		if strings.Contains(w, "slime/api/") {
			t.Fatalf("the sketch wrote a second keeper for api/: %v", res.wrote)
		}
	}
	if w, err := onto.Load(root2); err != nil || len(w.Validate()) != 0 {
		t.Fatalf("the merged colony checks clean: %v %v\n%s", err, w.Validate(), s2.table(lexiconFor("isekai").Prefixes))
	}
}

// The minds a world already has are worn where they fit; none is invented.
func TestMediumWearsFittingMinds(t *testing.T) {
	root := project(t, map[string]string{
		"go.mod":          "module example.com/m\n",
		"api/api.go":      goFile("api", "serves.", 300),
		"api/api_test.go": goTest("api"),
		".isekai/base-skills/slime-go-proving/SKILL.md": "---\nname: slime-go-proving\ndescription: how this Go module proves itself\n---\n",
		".isekai/base-skills/slime-py-proving/SKILL.md": "---\nname: slime-py-proving\ndescription: how the python half runs pytest\n---\n",
		".isekai/base-skills/api-contract/SKILL.md":     "---\nname: api-contract\ndescription: the api's wire contract\n---\n",
	})
	s, _ := sketchAt(t, root)
	sl := byName(s)["slime-api"]
	if strings.Join(sl.Minds, ",") != "api-contract,slime-go-proving" {
		t.Errorf("minds worn: %v", sl.Minds)
	}
	if doc := readDoc(t, root, "slime", "api"); !strings.Contains(doc, "- **Minds:** api-contract, slime-go-proving") {
		t.Errorf("the doc names its minds:\n%s", doc)
	}
}

// `init --level medium` on the wire: the census, the lanes, what was written and kept.
func TestInitMediumWire(t *testing.T) {
	root := project(t, map[string]string{
		"go.mod":                   "module example.com/w\n",
		"api/api.go":               goFile("api", "serves.", 300),
		"api/api_test.go":          goTest("api"),
		"store/store.go":           goFile("store", "stores.", 300, "example.com/w/api"),
		".github/workflows/ci.yml": "",
	})
	run := func() string {
		var out, errb bytes.Buffer
		env := func(k string) string {
			if k == "HOME" {
				return root
			}
			return ""
		}
		if code := Main("isekai", []string{"--root", root, "init", "--level", "medium", "--plain"}, IO{In: strings.NewReader(""), Out: &out, Err: &errb, Env: env}, Version{}); code != 0 {
			t.Fatalf("init: %d %s %s", code, out.String(), errb.String())
		}
		return out.String()
	}
	out := run()
	for _, want := range []string{"@F .isekai/slime/api/README.md", "@F lane elf-proj: orc-proj (2)", "(1 elf · 1 orc · 2 slimes · 1 kijin): 5 creatures written (0 kept as they were)"} {
		if !strings.Contains(out, want) {
			t.Errorf("wire lacks %q:\n%s", want, out)
		}
	}
	if out := run(); !strings.Contains(out, "0 creatures written (5 kept as they were)") {
		t.Errorf("a second run keeps everything:\n%s", out)
	}
}

func TestNormLevel(t *testing.T) {
	for in, want := range map[string]string{"": "light", "soft": "light", "Light": "light", "medium": "medium", "complex": "complex", "huge": ""} {
		if got := normLevel(in); got != want {
			t.Errorf("normLevel(%q) = %q, want %q", in, got, want)
		}
	}
}

// The ontology's zone rule, mirrored: what a name may claim.
func TestZoneOK(t *testing.T) {
	for _, c := range []struct {
		short, entry string
		ok           bool
	}{
		{"app", "cmd/", true}, {"app", "main.go", true}, {"anything", "src/a/b/", true},
		{"isekai-loop", "isekai/loop/", true}, {"isekai", "isekai/loop/", true}, {"loop", "isekai/loop/", false},
		{"isekai-loop", "isekai/wire/", false}, {"packages-ui", "packages/ui/src/", true}, {"agent-one", "agent-one/engine/", false},
	} {
		if got := zoneOK(c.short, c.entry); got != c.ok {
			t.Errorf("zoneOK(%q, %q) = %v", c.short, c.entry, got)
		}
	}
}

// The complex level's brief carries the sketch as its first hint, in the distribution's words.
func TestFoundingBrief(t *testing.T) {
	root := tree(t, "go.mod", "svc/a/x.go", "svc/a/y.go", "svc/b/z.go")
	d, _ := parseDist("isekai")
	lex := lexiconFor("isekai")
	s := sketchColony(root, d)
	ask := foundingAsk(root, d, s, lex.Ranks, lex.Prefixes, lex.Display)
	if strings.Contains(ask, "{{") || !strings.Contains(ask, "slime-svc → orc-") || !strings.Contains(ask, "isekai onto check") {
		t.Fatalf("the founding brief is not filled in:\n%s", ask)
	}
	a1 := lexiconFor("agent-one")
	d1, _ := parseDist("agent-one")
	ask1 := foundingAsk(root, d1, sketchColony(root, d1), a1.Ranks, a1.Prefixes, a1.Display)
	if !strings.Contains(ask1, a1.Prefixes["slime"]+"svc → "+a1.Prefixes["orc"]) || !strings.Contains(ask1, "agent-one onto check") || !strings.Contains(ask1, "**Rank:** Zone worker") {
		t.Fatalf("the brief speaks the distribution's words:\n%s", ask1)
	}
}
