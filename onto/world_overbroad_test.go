package onto

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mkOverbroadFixture builds a minimal single-or-two-creature world so each
// test controls exactly one README and measures its owns-emission in isolation.
func mkOverbroadFixture(t *testing.T, docs map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".isekai/law"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".isekai/law/isekai.md"), []byte("# stub\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for path, body := range docs {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// ownsSet returns the set of pOwns objects emitted for subject in world.
func ownsSet(t *testing.T, w *World, subject string) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	for _, tr := range w.Graph.All() {
		if tr.S == Is(subject) && tr.P == pOwns {
			out[tr.O.Value] = true
		}
	}
	return out
}

// TestDeriveOwns_ScratchSpansDoNotDerive — census class (a): a body journaling
// its walk of `.isekai/tmp/gosdk` or writing under `.isekai/tmp/slime-loop`
// is describing *shared per-run scratch soil*, never owned territory. The
// prose supplement must refuse `.isekai/tmp/<x>` spans regardless of self-naming.
func TestDeriveOwns_ScratchSpansDoNotDerive(t *testing.T) {
	dir := mkOverbroadFixture(t, map[string]string{
		".isekai/slime/world-core/README.md": "name: slime-world-core\n" +
			"Territory: `world/`\n" +
			"\n" +
			"- The Go SDK surface under `.isekai/tmp/gosdk` was walked by slime-world-core on commission; it is shared scratch.\n" +
			"- slime-world-core exercised the world fs and left a probe under `.isekai/tmp/netprobe`.\n",
	})
	w, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	owns := ownsSet(t, w, "slime-world-core")
	for _, bad := range []string{".isekai/tmp/gosdk", ".isekai/tmp/netprobe", ".isekai/tmp", ".isekai"} {
		if owns[bad] {
			t.Errorf("scratch span %q derived as owns of slime-world-core", bad)
		}
	}
	if !owns["world/"] && !owns["world"] {
		t.Errorf("structured Territory world/ was not emitted (owns=%v)", owns)
	}
	// positive: slitscan prose forbidden; plain `world` (normalised from `world/`) is the structured own.
}

// TestDeriveOwns_BareIsekaiDoesNotDerive — census class (b): lines like
// "the census-derived owns flagged `.isekai/` billed to THIS slime's zone" name
// `.isekai` as the *site of a mis-attribution*, not a territory claim. The
// prose supplement must refuse bare `.isekai` (no subdir) regardless of self-naming.
func TestDeriveOwns_BareIsekaiDoesNotDerive(t *testing.T) {
	dir := mkOverbroadFixture(t, map[string]string{
		".isekai/slime/app-commands/README.md": "name: slime-app-commands\n" +
			"Territory: `app/commands.go`\n" +
			"\n" +
			"- 2026-10-04 — the record tool's memo embeds ticked `.isekai/` inside slime-app-commands's spine.\n" +
			"- slime-app-commands itself notes the census attributes `.isekai/` writes to this slime's zone, none matching `app/*`.\n",
	})
	w, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	owns := ownsSet(t, w, "slime-app-commands")
	if owns[".isekai"] || owns[".isekai/"] {
		t.Errorf("bare .isekai derived as owns of slime-app-commands (owns=%v)", owns)
	}
	if !owns["app/commands.go"] {
		t.Errorf("structured Territory app/commands.go was not emitted (owns=%v)", owns)
	}
	// positive: ancestor `app` derived from the structured own's parent chain is pre-existing behaviour.
}

// TestDeriveOwns_QueryPlaceholdersDoNotDerive — census class (c): live tree
// leaks `api/plan?body=<id>` into owns. A path carrying a query template is
// a URI, not a filesystem path; the prose supplement must refuse it.
func TestDeriveOwns_QueryPlaceholdersDoNotDerive(t *testing.T) {
	dir := mkOverbroadFixture(t, map[string]string{
		".isekai/slime/tui-board/README.md": "name: slime-tui-board\n" +
			"Territory: `tui/`\n" +
			"\n" +
			"- live board at `api/plan?body=<id>` — slime-tui-board serves it.\n",
	})
	w, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	owns := ownsSet(t, w, "slime-tui-board")
	for o := range owns {
		if strings.ContainsAny(o, "?<>") {
			t.Errorf("query/placeholder owns leaked: %q", o)
		}
	}
	if !owns["tui"] && !owns["tui/"] {
		t.Errorf("structured Territory tui/ was not emitted (owns=%v)", owns)
	}
}

// TestDeriveOwns_StructuredForeignPathDoesNotDerive — census class (e):
// a structured Territory entry that resolves OUTSIDE the claimant's zone
// is a cross-reference, never ownership — mirror of the prose rule at
// onto/world_test.go:463. The fixture creature whose zone is
// `.isekai/slime/owna/core` claims `app/cli.go` (foreign-zone file): that
// structured claim must mint NO owns (neither `app` nor `app/cli.go`).
// Meanwhile a structured bare filename (`local.go` — the adopted
// world_test.go:387 semantics) and a full in-zone path MUST still derive,
// proving only the foreign claim is dropped.
func TestDeriveOwns_StructuredForeignPathDoesNotDerive(t *testing.T) {
	name := "slime-owna"
	docBody := "name: " + name + "\n" +
		"Territory: `.isekai/slime/owna/core`, `.isekai/slime/owna/inner/deep.go`, `local.go`, `app/cli.go`\n" +
		"\n" +
		name + " touches app/cli.go only as a caller.\n"
	dir := mkOverbroadFixture(t, map[string]string{
		".isekai/slime/owna/README.md":          docBody,
		".isekai/slime/owna/core/inner/deep.go": "package inner\n",
		".isekai/slime/owna/core/local.go":      "package owna\n",
		"app/cli.go":                            "package main\n",
	})
	w, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	owns := ownsSet(t, w, "slime-owna")
	if owns["app"] || owns["app/cli.go"] {
		t.Errorf("structured foreign claim app/cli.go derived owns — out-of-zone structured claims are cross-references, never soil (owns=%v)", owns)
	}
	if !owns[".isekai/slime/owna/inner/deep.go"] {
		t.Errorf("full in-zone structured claim was suppressed along with the foreign one (owns=%v)", owns)
	}
	if !owns["local.go"] {
		t.Errorf("bare zone-local filename claim local.go was suppressed — the adopted bare-filename semantics (world_test.go:387) MUST still derive (owns=%v)", owns)
	}
}

// TestDeriveOwns_APIRoutesDoNotDerive — census class (d): prose-narrated
// leading-slash spans that hold NO dot-extension (`/api/tree`,
// `/api/tree/children`) are REST routes, not filesystem soil; normPath
// normalised them merely because they hold a slash, and self-naming prose
// beside the creature's name minted owns for them (live: slime-world-core
// and slime-tui-board READMEs both narrate `/api/tree`). Meanwhile
// catalogued real soil with dots (`bench/board.go:255`) must survive.
func TestDeriveOwns_APIRoutesDoNotDerive(t *testing.T) {
	name := "slime-tui-board"
	docBody := "name: " + name + "\n" +
		"Territory: `tui/`\n" +
		"\n" +
		name + " answers the board handlers `/api/tree` and `/api/tree/children`.\n"
	// the spans above sit in backticks with the creature name adjoining —
	// exactly the self-naming prose that mints them (live: slime-world-core
	// and slime-tui-board READMEs both narrate `/api/tree`).
	dir := mkOverbroadFixture(t, map[string]string{
		".isekai/slime/tui-board/README.md": docBody,
	})
	w, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	owns := ownsSet(t, w, "slime-tui-board")
	for o := range owns {
		if strings.HasPrefix(o, "/") && !strings.Contains(o, ".") {
			t.Errorf("REST route %q minted as owns of slime-tui-board — leading-slash, dot-less prose spans are API routes, never soil (owns=%v)", o, owns)
		}
	}
	if !owns["tui"] && !owns["tui/"] {
		t.Errorf("structured Territory tui/ was not emitted (owns=%v)", owns)
	}
}
