package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// The landing line (#393): one durable merge per repository integrates and
// lands at a time, so each landing runs one integration suite. The ordering
// and the files live in internal/landingline; this is the merge's glue.

// landingLinePoll is how often a waiting merge looks again; tests shorten it.
var landingLinePoll = 2 * time.Second

// landingLineDir is the repository's line: in its git common dir, which every
// slot of the repository shares, beside the repo transaction lock (#370).
func landingLineDir(r gitRunner, root string) (string, error) {
	common, err := landingGit(r, root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return filepath.Join(common, "sdlc", "landing"), nil
}

// openLandingLine opens dir with production options; stderr receives the
// waiting position whenever it changes.
func openLandingLine(dir string, stderr io.Writer) (*landingline.Line, error) {
	var line *landingline.Line
	o := landingline.Options{
		Poll: landingLinePoll,
		Grants: func() map[int]landingline.Grant {
			entries, _, err := line.GrantEntries()
			if err != nil {
				return nil // an unreadable log confers no priority
			}
			return landingline.ValidGrants(entries, time.Now())
		},
		OnWait: func(pos int, head *landingline.Ticket) {
			fmt.Fprintf(stderr, "landing line: position %d; landing now: %s (sdlc landing status)\n", pos, describeTicket(head))
		},
		OnAdmit: func(head landingline.Ticket, displaced []landingline.Ticket) {
			e := landingline.Event{Kind: landingline.EventAdmit, Seq: head.Seq, Issue: head.Issue, Branch: head.Branch}
			for _, d := range displaced {
				e.Displaced = append(e.Displaced, d.Seq)
			}
			_ = line.Record(e)
		},
	}
	line, err := landingline.Open(dir, o)
	return line, err
}

func describeTicket(t *landingline.Ticket) string {
	switch {
	case t == nil:
		return "none"
	case t.Unreadable:
		return "a merge whose ticket is unreadable"
	default:
		return fmt.Sprintf("#%d (%s, pid %d)", t.Issue, t.Branch, t.PID)
	}
}

// landingTurn is a merge's place in the line, from its ticket to its outcome.
type landingTurn struct {
	line   *landingline.Line
	handle *landingline.Handle
	PR     int // set once the PR is selected, after admission
}

// enterLandingLine takes a ticket and waits for this merge's turn. Nothing the
// landing acts on (branch head, PRs, selection) may be observed before it
// returns: a wait can be long, and an observation taken before it is stale.
func enterLandingLine(ctx context.Context, r gitRunner, t landingTarget, branch string, stderr io.Writer) (*landingTurn, error) {
	dir, err := landingLineDir(r, t.Root)
	if err != nil {
		return nil, err
	}
	line, err := openLandingLine(dir, stderr)
	if err != nil {
		return nil, err
	}
	h, err := line.Enqueue(landingline.Ticket{Issue: issue.IDFromFilename(branch), Branch: branch, PID: os.Getpid()})
	if err != nil {
		return nil, err
	}
	turn := &landingTurn{line: line, handle: h}
	tk := h.Ticket()
	_ = line.Record(landingline.Event{Kind: landingline.EventEnqueue, Seq: tk.Seq, Issue: tk.Issue, Branch: tk.Branch})
	if err := line.Wait(ctx, h); err != nil {
		turn.leave(err)
		return nil, err
	}
	return turn, nil
}

// leave records the outcome and frees the line. A landed issue's priority
// grant ends with it.
func (lt *landingTurn) leave(err error) {
	tk := lt.handle.Ticket()
	e := landingline.Event{Kind: landingline.EventLanded, Seq: tk.Seq, Issue: tk.Issue, Branch: tk.Branch, PR: lt.PR}
	if err != nil {
		e.Kind, e.Detail = landingline.EventRefused, err.Error()
	}
	_ = lt.line.Record(e)
	if err == nil && tk.Issue != 0 {
		if entries, _, rerr := lt.line.GrantEntries(); rerr == nil {
			if _, granted := landingline.ValidGrants(entries, time.Now())[tk.Issue]; granted {
				_ = lt.line.AddGrant(landingline.Grant{Issue: tk.Issue, Grantor: landingline.Grantor{Address: "sdlc merge"}, Reason: "landed", Revoked: true})
			}
		}
	}
	lt.handle.Release()
}

func anyMergedPR(prs []landingPR) bool {
	for _, p := range prs {
		if p.State == "MERGED" {
			return true
		}
	}
	return false
}
