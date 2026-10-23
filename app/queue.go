package app

import (
	"fmt"
	"strings"

	"github.com/Kaginari/isekai/queue"
)

// cmdQueue is `q`: the shared work queue, read and moved by the human. The human stands over
// every body, so any held or claimed item is theirs to move; --as names a body to act for.
func (a *App) cmdQueue(args []string, io IO) int {
	if a.World == nil {
		fmt.Fprintf(io.Err, "@S FAIL\n@? no %s here\n", worldWord(a.Cfg.Dist.Name))
		return 2
	}
	s := queue.Open(a.World.QueuePath(), func(string, string) bool { return true })
	by := strings.ToLower(a.World.Lex.Human)
	if by == "" {
		by = "human"
	}
	return queue.CLI(s, by, args, a.Cfg.Output.Format == "json", io.Out, io.Err)
}
