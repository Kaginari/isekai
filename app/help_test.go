package app

import (
	"bytes"
	"strings"
	"testing"
)

func mainIn(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	dir := t.TempDir()
	var out, errb bytes.Buffer
	env := func(k string) string {
		if k == "HOME" {
			return dir
		}
		return ""
	}
	code := Main("isekai", append([]string{"--root", dir}, args...), IO{In: strings.NewReader(""), Out: &out, Err: &errb, Env: env}, Version{Version: "test", Commit: "abc", Date: "today"})
	return code, out.String(), errb.String()
}

// --help, -h and --version answer at once, anywhere — never a session, never a world lookup.
func TestHelpAndVersionFlags(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"-h"}, {"help"}} {
		code, out, errs := mainIn(t, args...)
		if code != 0 || !strings.Contains(out, "isekai [command] [flags]") || errs != "" {
			t.Errorf("%v: %d %q %q", args, code, out, errs)
		}
	}
	for _, args := range [][]string{{"--version"}, {"version"}} {
		if code, out, _ := mainIn(t, args...); code != 0 || !strings.HasPrefix(out, "isekai test") {
			t.Errorf("%v: %d %q", args, code, out)
		}
	}
	// a command's own help: its line and the flags, nothing run
	code, out, errs := mainIn(t, "run", "--help")
	if code != 0 || !strings.Contains(out, "isekai run") || !strings.Contains(out, "run one ask to its end") || errs != "" {
		t.Errorf("run --help: %d %q %q", code, out, errs)
	}
}

// An unknown command is named as unknown, outside a world too.
func TestUnknownCommandOutsideAWorld(t *testing.T) {
	code, _, errs := mainIn(t, "frobnicate")
	if code != 2 || !strings.Contains(errs, `unknown command "frobnicate"`) {
		t.Fatalf("%d %q", code, errs)
	}
}

// init --level complex with no reachable model founds the light world, names the missing key and
// the rerun, and never claims the founding session started.
func TestInitComplexNeedsAModelFirst(t *testing.T) {
	code, out, errs := mainIn(t, "init", "--level", "complex")
	if code != 2 || strings.Contains(out, "founding session starts") {
		t.Fatalf("%d %q %q", code, out, errs)
	}
	if !strings.Contains(errs, "ANTHROPIC_API_KEY") || !strings.Contains(errs, "init --level complex") || !strings.Contains(errs, "founded with the law only") {
		t.Fatalf("the next step is not named: %q %q", out, errs)
	}
}
