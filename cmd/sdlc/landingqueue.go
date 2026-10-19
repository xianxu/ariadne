package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// The landing line (#393): one durable merge per repository integrates and
// lands at a time, so each landing runs one integration suite. The ordering
// and the files live in internal/landingline; this is the merge's glue.

// landingLinePoll is how often a waiting merge looks again; tests shorten it.
var landingLinePoll = 2 * time.Second

// landingNow is the line's clock in merge and the landing commands.
var landingNow = time.Now

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
// waiting position whenever it changes. sched (#449) is the train setting,
// read again on every admission attempt so a change on main counts while a
// merge waits; nil, or a schedule with trains off, is the serial line.
func openLandingLine(dir string, sched func() landingline.Schedule, stderr io.Writer) (*landingline.Line, *lineLog, error) {
	var line *landingline.Line
	log := &lineLog{stderr: stderr}
	o := landingline.Options{
		Now:  landingNow,
		Poll: landingLinePoll,
		Grants: func() map[int]landingline.Grant {
			entries, _, err := line.GrantEntries()
			if err != nil {
				return nil // an unreadable log confers no priority
			}
			return landingline.ValidGrants(entries, landingNow())
		},
		OnWait: func(pos int, head *landingline.Ticket) {
			fmt.Fprintf(stderr, "landing line: position %d; landing now: %s (sdlc landing status)\n", pos, describeTicket(head))
		},
		OnAdmit: func(head landingline.Ticket, displaced []landingline.Ticket) {
			e := landingline.Event{Kind: landingline.EventAdmit, Seq: head.Seq, Issue: head.Issue, Branch: head.Branch}
			for _, d := range displaced {
				e.Displaced = append(e.Displaced, d.Seq)
			}
			log.record(e)
		},
	}
	if sched != nil {
		o.Admission = trainAdmission(sched, func() []landingline.Event {
			events, _, err := line.Events()
			if err != nil {
				return nil // an unreadable log: the first train, never a stall
			}
			return events
		}, landingNow)
		o.OnHeld = func(why string) { fmt.Fprintf(stderr, "landing train: held: %s\n", why) }
	}
	line, err := landingline.Open(dir, o)
	log.line = line
	return line, log, err
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
	line    *landingline.Line
	handle  *landingline.Handle
	log     *lineLog
	stderr  io.Writer
	PR      int  // set once the PR is selected, after admission
	Resumed bool // the selected PR had already merged: a resume, not a landing
}

// lineLog records a merge's events. A write failure warns on stderr once, so
// lost audit and pilot data is visible without failing the landing.
type lineLog struct {
	line   *landingline.Line
	stderr io.Writer
	warned bool
}

func (l *lineLog) record(e landingline.Event) {
	if err := l.line.Record(e); err != nil && !l.warned {
		l.warned = true
		fmt.Fprintf(l.stderr, "landing line: could not log %s (%v); landing continues, unaudited\n", e.Kind, err)
	}
}

// enterLandingLine takes a ticket and waits for this merge's turn. Nothing the
// landing acts on (branch head, PRs, selection) may be observed before it
// returns: a wait can be long, and an observation taken before it is stale.
//
// With trains on (#449) the merge also vouches its readiness while it waits
// (readiness reads its own gates); a vouch is a claim, so it observes for
// the conductor, never for this merge's own landing, which reads again.
func enterLandingLine(ctx context.Context, r gitRunner, t landingTarget, branch string, sched func() landingline.Schedule, readiness func(context.Context) landingline.Readiness, stderr io.Writer) (*landingTurn, error) {
	dir, err := landingLineDir(r, t.Root)
	if err != nil {
		return nil, err
	}
	line, log, err := openLandingLine(dir, sched, stderr)
	if err != nil {
		return nil, err
	}
	h, err := line.Enqueue(landingline.Ticket{Issue: issue.IDFromFilename(branch), Branch: branch, PID: os.Getpid()})
	if err != nil {
		return nil, err
	}
	turn := &landingTurn{line: line, handle: h, log: log, stderr: stderr}
	tk := h.Ticket()
	log.record(landingline.Event{Kind: landingline.EventEnqueue, Seq: tk.Seq, Issue: tk.Issue, Branch: tk.Branch})
	var stopVouching func()
	if sched != nil && readiness != nil {
		vctx, cancel := context.WithCancel(ctx)
		done := vouchWhileWaiting(vctx, line, h, sched, readiness, stderr)
		stopVouching = func() { cancel(); <-done }
	}
	err = line.Wait(ctx, h)
	if stopVouching != nil {
		stopVouching()
	}
	if err != nil {
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
	switch {
	case err != nil:
		e.Kind, e.Detail = landingline.EventRefused, err.Error()
	case lt.Resumed:
		e.Kind = landingline.EventResumed // another merge landed it while this one waited
	}
	lt.log.record(e)
	if err == nil && tk.Issue != 0 {
		entries, _, rerr := lt.line.GrantEntries()
		if rerr != nil {
			fmt.Fprintf(lt.stderr, "landing line: could not read grants to end #%d's priority (%v); a grant expires on its own\n", tk.Issue, rerr)
		} else {
			if _, granted := landingline.ValidGrants(entries, landingNow())[tk.Issue]; granted {
				if gerr := lt.line.AddGrant(landingline.Grant{Issue: tk.Issue, Grantor: landingline.Grantor{Address: "sdlc merge"}, Reason: "landed", Revoked: true}); gerr != nil {
					fmt.Fprintf(lt.stderr, "landing line: could not end #%d's priority grant (%v); it expires on its own\n", tk.Issue, gerr)
				}
			}
		}
	}
	lt.handle.Release()
}

// resumeOnly reports whether the branch's PRs leave only a resume: a merged PR
// and no open one. An empty listing is not a resume (it may lag a new PR).
func resumeOnly(prs []landingPR) bool {
	merged := false
	for _, p := range prs {
		switch p.State {
		case "OPEN":
			return false
		case "MERGED":
			merged = true
		}
	}
	return merged
}

// landingScheduleTTL is how long a read train setting is reused. The setting
// changes rarely (an operator-verified commit to main), and admission polls
// every few seconds: without the cache each poll would run two git processes.
var landingScheduleTTL = 30 * time.Second

// landingScheduleSource reads the train setting from the local remote-tracking
// main, reusing a read for landingScheduleTTL. It is never fetched here:
// whatever last fetched main is what counts. An unreadable setting is the
// serial line, warned once: serial spends more suites but never lands
// anything a train would not. The source is shared by the vouching goroutine
// and the admission poll, so it is safe for concurrent use.
func landingScheduleSource(git func(args ...string) (string, error), remote string, stderr io.Writer) func() landingline.Schedule {
	var mu sync.Mutex
	var cached landingline.Schedule
	var readAt time.Time
	warned := false
	return func() landingline.Schedule {
		mu.Lock()
		defer mu.Unlock()
		now := landingNow()
		if !readAt.IsZero() && !now.Before(readAt) && now.Sub(readAt) < landingScheduleTTL {
			return cached
		}
		s, err := readLandingSchedule(git, landingMainRef(git, remote))
		if err != nil {
			if !warned {
				warned = true
				fmt.Fprintf(stderr, "landing train: could not read %s (%v); landing serially\n", landingConfigPath, err)
			}
			s = landingline.Schedule{}
		}
		cached, readAt = s, now
		return cached
	}
}

// landingMainRef is the remote-tracking main the repository lands into:
// remote's main when the remote is known (sdlc merge's target), else main's
// configured upstream remote, else origin. The one definition both merge and
// status use.
func landingMainRef(git func(args ...string) (string, error), remote string) string {
	if remote == "" {
		if out, err := git("config", "--get", "branch.main.remote"); err == nil && strings.TrimSpace(out) != "" {
			remote = strings.TrimSpace(out)
		} else {
			remote = "origin"
		}
	}
	return remote + "/main"
}
