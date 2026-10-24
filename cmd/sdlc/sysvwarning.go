package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/sysvshm"
)

func nativeSysVObserver() sysvshm.Observer { return sysvshm.NewNative() }

// Every configured host run appends a final summary, including clean runs. An
// earlier child-output lookalike therefore cannot create its warning marker.
func writeSysVReport(out io.Writer, check, sha string, r sysvshm.Report) {
	counts := fmt.Sprintf("owner-before=%d owner-after=%d", r.Before, r.After)
	if r.Unavailable != "" {
		counts = "owner-before=unknown owner-after=unknown"
	}
	fmt.Fprintf(out, "\n[sdlc SysV observation] check=%q head=%q %s\n", check, sha, counts)
	state := "observed"
	if r.Unavailable != "" {
		state = "unavailable"
		fmt.Fprintf(out, "[sdlc SysV observation] observation unavailable: %q\n", boundedSysVText(r.Unavailable))
	} else if len(r.Candidates) > 0 {
		state = "warning"
		fmt.Fprintln(out, "[sdlc SysV observation] new segments appeared during this run, not attributed unless explicitly witnessed below")
		for _, c := range r.Candidates {
			relationship, name := "not attributed", "unknown"
			if c.Witness.Name != "" && c.Witness.Start != "" {
				relationship, name = "witnessed in run process group", filepath.Base(c.Witness.Name)
			}
			fmt.Fprintf(out, "[sdlc SysV observation] id=%d creator=%d attachments=%d bytes=%d command=%s relationship=%q\n", c.Segment.ID, c.Segment.Creator, c.Segment.Attachments, c.Segment.Size, strconv.Quote(boundedSysVText(name)), relationship)
		}
	}
	fmt.Fprintf(out, "sdlc-sysv-v1 check=%q head=%q state=%s\n", check, sha, state)
}

func boundedSysVText(s string) string {
	if len(s) > 240 {
		return s[:240]
	}
	return s
}

func hasSysVWarning(tail []string, check, sha string) bool {
	if len(tail) == 0 {
		return false
	}
	return tail[len(tail)-1] == fmt.Sprintf("sdlc-sysv-v1 check=%q head=%q state=warning", check, sha)
}

// A canceled run still gets a separately bounded observation. There is no
// monitor goroutine or cleanup authority, and no probe before admission.
func captureSysV(caller context.Context, observer sysvshm.Observer) sysvshm.Snapshot {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(commandContext(caller)), 2*time.Second)
	defer cancel()
	return observer.Snapshot(ctx)
}

func sysVRunReport(caller context.Context, observer sysvshm.Observer, before sysvshm.Snapshot, witnesses map[int]sysvshm.Witness, out io.Writer, check, sha string) {
	after := captureSysV(caller, observer)
	writeSysVReport(out, check, sha, sysvshm.Delta(before, after, os.Geteuid(), witnesses))
}
