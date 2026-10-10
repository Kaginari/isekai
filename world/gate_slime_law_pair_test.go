package world

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kaginari/isekai/wire"
)

func TestSlimeLawPairingGatePassthrough(t *testing.T) {
	dir := fixture(t)
	w := open(t, dir)

	// replicate the live world: slime-law claims .isekai/config.local.yaml as territory
	// (per b413fdc) and the README is inside its own dir. Mutate the creature in place.
	docPath := ".isekai/slime/law/README.md"
	os.WriteFile(filepath.Join(dir, docPath), []byte("# slime-law\n- **Rank:** Slime\n- **Territory:** `.isekai/config.local.yaml`\n- **Reports to:** orc-core\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".isekai/slime/law"), 0o755)
	w.Reload()

	fmt.Printf("=== roster after law creature added ===\n")
	for _, c := range w.Creatures {
		fmt.Printf("  %s dir=%q terr=%v\n", c.Name, c.Dir, c.Territory)
	}

	// what rimuru's turn wrote: config.local.yaml AND the README. The gate fired on both.
	owner1 := w.Owner(".isekai/config.local.yaml")
	owner2 := w.Owner(docPath)
	fmt.Printf("owner(config.local.yaml)=%v  owner(README)=%v\n", nameOf(owner1), nameOf(owner2))

	opt := DefaultHooks().Gate
	opt.TraitsHold = false
	opt.DocTruthful = false
	v := w.Gate(context.Background(), nil, opt, "rimuru", []string{".isekai/config.local.yaml", docPath}, &wire.Report{Status: "DONE"}, false, "founding")
	fmt.Printf("verdict: %q reasons=%v\n", v.Word, v.Reasons)
}

func nameOf(c *Creature) string {
	if c == nil {
		return "<nil>"
	}
	return c.Name
}
