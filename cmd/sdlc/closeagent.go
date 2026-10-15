package main

import (
	"context"
	"fmt"
	"io"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/agentid"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// reportMissingAgent warns about the window's own commits (first parent, no
// merges, so main's commits merged in don't count) that carry no Agent
// trailer (#367). A report, not a gate: agents adopt the convention first.
func reportMissingAgent(ctx context.Context, stderr io.Writer, base, head string) {
	out, err := gitx.GitOutputContext(ctx, "log", "--first-parent", "--no-merges", "--format="+agentid.LogFormat, base+".."+head)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("Agent trailers: could not read %s..%s: %v", shortSHA(base), shortSHA(head), err))
		return
	}
	missing := agentid.MissingAgent(agentid.ParseLog(string(out)))
	if len(missing) == 0 {
		return
	}
	cwarn(stderr, fmt.Sprintf("%d commit(s) in the window lack an Agent: trailer (#367; `sdlc agent-id` prints yours):", len(missing)))
	for i, c := range missing {
		if i == 10 {
			fmt.Fprintf(stderr, "      … and %d more\n", len(missing)-i)
			break
		}
		fmt.Fprintf(stderr, "      %s %s\n", shortSHA(c.SHA), c.Subject)
	}
}
