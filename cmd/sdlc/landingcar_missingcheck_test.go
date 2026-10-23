package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/pkg/testgate"
)

func TestCarMissingCheckStopsRetainedWait(t *testing.T) {
	for _, agent := range []string{"claude", "codex"} {
		t.Run(agent, func(t *testing.T) {
			t.Setenv("PAIR_AGENT", agent)
			target, gh, _ := carFixture(t)
			line, held := heldLine(t, target.Root)
			oldPoll, oldEvery := landingLinePoll, carVouchEvery
			landingLinePoll, carVouchEvery = time.Millisecond, time.Millisecond
			t.Cleanup(func() { landingLinePoll, carVouchEvery = oldPoll, oldEvery })
			ctx, cancel := context.WithTimeout(context.Background(), testgate.HangCeiling)
			defer cancel()
			t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
			ctx, q, err := startQueueOperation(ctx, target.Root, gh.pr.HeadOID, []string{"merge", "landing"})
			if err != nil {
				t.Fatal(err)
			}
			defer q.finish(context.Background(), "")
			q.op.Workspace = "test:4"
			target.RepoIdentity = "github.com/test/repo"
			calls := 0
			gates := passGates()
			gates.ci = func(context.Context, landingPR) error {
				return fmt.Errorf("wrapped: %w", (ciVerdict{Reason: ciLocalNeeded, LocalNeeded: []string{affectedCheckName}}).refusal(535))
			}
			readiness := func(ctx context.Context) (landingline.Readiness, bool, error) {
				calls++
				if calls > 1 {
					cancel() // bounded negative control: old code repeats instead of stopping
				}
				return carReadiness(ctx, execGitRunner{}, gh, target, landingTestBranch, gates, landingNow())
			}
			sched := func() landingline.Schedule { return landingline.Schedule{Interval: time.Hour, MaxCars: 8} }
			turn, err := enterLandingLine(ctx, execGitRunner{}, target, landingTestBranch, sched, readiness, io.Discard)
			if turn != nil {
				turn.leave(errors.New("unexpected admission"))
			}
			if err == nil || !strings.Contains(err.Error(), "merge stopped: run sdlc local-check run --name affected --boarding, then re-enter sdlc merge") || calls != 1 {
				t.Fatalf("calls=%d turn=%v err=%v", calls, turn, err)
			}
			if head, waiting, err := line.Status(); err != nil || head == nil || head.Seq != held.Ticket().Seq || len(waiting) != 0 {
				t.Fatalf("refusal retained a ticket: head=%v waiting=%v err=%v", head, waiting, err)
			}
			events := lineEvents(t, line)
			if len(events[landingline.EventWaitStart]) != 1 || len(events[landingline.EventWaitEnd]) != 1 {
				t.Fatalf("missing evidence endpoints: %v", events)
			}
			// A passing record on retry is represented by the same gate now
			// passing. No check runner was invoked or separate holder released.
			held.Release()
			gates.ci = passGates().ci
			retry := func(ctx context.Context) (landingline.Readiness, bool, error) {
				return carReadiness(ctx, execGitRunner{}, gh, target, landingTestBranch, gates, landingNow())
			}
			output := &watchWriter{want: "unused", seen: make(chan struct{})}
			turn, err = enterLandingLine(ctx, execGitRunner{}, target, landingTestBranch, sched, retry, output)
			if err != nil || turn == nil {
				t.Fatalf("retry did not admit: turn=%v err=%v\n%s", turn, err, output.String())
			}
			turn.leave(nil)
		})
	}
}

func TestVouchTerminalRace(t *testing.T) {
	for _, mode := range []string{"admission", "cancel", "terminal"} {
		t.Run(mode, func(t *testing.T) {
			line, _, err := openLandingLine(t.TempDir(), nil, io.Discard)
			if err != nil {
				t.Fatal(err)
			}
			h, err := line.Enqueue(landingline.Ticket{Issue: 1, Branch: "b"})
			if err != nil {
				t.Fatal(err)
			}
			defer h.Release()
			ctx, cancel := context.WithTimeout(context.Background(), testgate.HangCeiling)
			defer cancel()
			entered, release := make(chan struct{}), make(chan struct{})
			terminal := errors.New("missing fixture record")
			var calls, stops atomic.Int32
			readiness := func(context.Context) (landingline.Readiness, bool, error) {
				calls.Add(1)
				close(entered)
				select {
				case <-release:
				case <-ctx.Done():
				}
				return landingline.Readiness{Branch: "b", Reason: "missing", At: landingNow()}, false, terminal
			}
			done := vouchWhileWaiting(ctx, line, h, func() landingline.Schedule { return landingline.Schedule{Interval: time.Hour} }, readiness, func(cause error) {
				if cause != terminal {
					t.Errorf("wrong cause: %v", cause)
				}
				stops.Add(1)
			}, io.Discard)
			select {
			case <-entered:
			case <-ctx.Done():
				t.Fatal("no readiness callback")
			}
			if mode == "admission" {
				if err := line.Wait(ctx, h); err != nil {
					t.Fatal(err)
				}
			} else if mode == "cancel" {
				cancel()
			}
			close(release)
			select {
			case <-done:
			case <-ctx.Done():
				if mode != "cancel" {
					t.Fatal("vouch did not join")
				}
				<-done
			}
			want := int32(0)
			if mode == "terminal" {
				want = 1
			}
			if stops.Load() != want || calls.Load() != 1 {
				t.Fatalf("stops=%d calls=%d", stops.Load(), calls.Load())
			}
		})
	}
}

func TestEnterLandingLineTerminalHandoff(t *testing.T) {
	target, _, _ := carFixture(t)
	called := false
	ctx := context.WithValue(context.Background(), trainReformReleaseKey{}, func() { called = true })
	terminal := errors.New("missing record before handoff")
	readiness := func(context.Context) (landingline.Readiness, bool, error) {
		return landingline.Readiness{}, false, terminal
	}
	turn, err := enterLandingLine(ctx, execGitRunner{}, target, landingTestBranch, nil, readiness, io.Discard)
	if turn != nil || err != terminal || called {
		t.Fatalf("turn=%v err=%v released=%v", turn, err, called)
	}
}

func TestEnterLandingLineSettledWinsLateTerminal(t *testing.T) {
	target, _, _ := carFixture(t)
	line, held := heldLine(t, target.Root)
	ctx, cancel := context.WithTimeout(context.Background(), testgate.HangCeiling)
	defer cancel()
	terminal := errors.New("late missing record")
	readiness := func(vctx context.Context) (landingline.Readiness, bool, error) {
		_, waiting, err := line.Status()
		if err != nil || len(waiting) != 1 {
			t.Errorf("waiting=%v err=%v", waiting, err)
			return landingline.Readiness{}, false, terminal
		}
		if err := line.Settle(waiting[0].Seq, landingline.Outcome{Train: held.Ticket().Seq, Kind: landingline.OutcomeLanded, PR: 5}); err != nil {
			t.Error(err)
		}
		<-vctx.Done() // Wait observed the outcome before readiness completes.
		return landingline.Readiness{}, false, terminal
	}
	turn, err := enterLandingLine(ctx, execGitRunner{}, target, landingTestBranch, func() landingline.Schedule { return landingline.Schedule{Interval: time.Hour} }, readiness, io.Discard)
	var settled *landingline.Settled
	if turn != nil || !errors.As(err, &settled) || settled.Outcome.Kind != landingline.OutcomeLanded {
		t.Fatalf("turn=%v err=%v", turn, err)
	}
}

func TestCarMissingCheckNamesStayOnePass(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			target, gh, _ := carFixture(t)
			gates := passGates()
			names := make([]string, n)
			for i := range names {
				names[i] = fmt.Sprintf("check-%02d", i)
			}
			probes := 0
			gates.ci = func(context.Context, landingPR) error {
				probes++
				return (ciVerdict{Reason: ciPending, LocalNeeded: names}).refusal(1)
			}
			_, _, err := carReadiness(context.Background(), execGitRunner{}, gh, target, landingTestBranch, gates, landingNow())
			if err == nil || probes != 1 {
				t.Fatalf("probes=%d err=%v", probes, err)
			}
			var gate *ciGateError
			if !errors.As(err, &gate) || len(gate.LocalNeeded) != n {
				t.Fatalf("lost needed names: %v", err)
			}
			for _, name := range names {
				if !strings.Contains(err.Error(), localCheckCommand(name)) {
					t.Fatalf("missing command for %s", name)
				}
			}
		})
	}
}
