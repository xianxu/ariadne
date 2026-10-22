package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestLandingWaitUnsupportedIdentityPreservesAdmission(t *testing.T) {
	root, _, _ := integrationFixture(t)
	var stderr bytes.Buffer
	turn, err := enterLandingLine(context.Background(), mergeRunner, landingTarget{Root: root}, "000547-test", nil, nil, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	defer turn.handle.Release()
	if !strings.Contains(stderr.String(), "canonical invocation identity unavailable") {
		t.Fatal(stderr.String())
	}
	events, _, err := turn.line.Events()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range events {
		if strings.HasPrefix(e.Kind, "wait_") {
			t.Fatalf("fabricated unsupported identity: %+v", e)
		}
	}
}
func TestLandingWaitSupportedIdentityRecordsOwnerReturn(t *testing.T) {
	root, _, _ := integrationFixture(t)
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	head := git(t, root, "rev-parse", "HEAD")
	ctx, q, err := startQueueOperation(context.Background(), root, head, []string{"merge", "landing"})
	if err != nil {
		t.Fatal(err)
	}
	defer q.finish(context.Background(), "")
	q.op.Workspace = "test:4"
	var stderr bytes.Buffer
	turn, err := enterLandingLine(ctx, mergeRunner, landingTarget{Root: root, RepoIdentity: "github.com/test/repo"}, "000547-test", nil, nil, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	defer turn.handle.Release()
	events, _, err := turn.line.Events()
	if err != nil {
		t.Fatal(err)
	}
	var starts, ends int
	for _, e := range events {
		switch e.Kind {
		case landingline.EventWaitStart:
			starts++
		case landingline.EventWaitEnd:
			ends++
		}
	}
	if starts != 1 || ends != 1 {
		t.Fatalf("endpoints=%d,%d stderr=%s", starts, ends, stderr.String())
	}
}

func TestLandingWaitBeginFailureRefusesAndCleansTicket(t *testing.T) {
	root, _, _ := integrationFixture(t)
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	ctx, q, err := startQueueOperation(context.Background(), root, git(t, root, "rev-parse", "HEAD"), []string{"merge", "landing"})
	if err != nil {
		t.Fatal(err)
	}
	defer q.finish(context.Background(), "")
	q.op.Workspace = "test:4"
	dir, err := landingLineDir(mergeRunner, root)
	if err != nil {
		t.Fatal(err)
	}
	line, err := landingline.Open(dir, landingline.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(filepath.Join(dir, landingline.EventsFile), 0755); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	turn, err := enterLandingLine(ctx, mergeRunner, landingTarget{Root: root, RepoIdentity: "github.com/test/repo"}, "000547-test", nil, nil, &stderr)
	if err == nil || turn != nil || !strings.Contains(err.Error(), "cannot record retained wait start") {
		t.Fatalf("turn=%+v err=%v", turn, err)
	}
	head, waiting, err := line.Status()
	if err != nil || head != nil || len(waiting) != 0 {
		t.Fatalf("leaked ticket: head=%+v waiting=%+v err=%v", head, waiting, err)
	}
}
func TestLandingWaitReadinessJoinAndLateFailure(t *testing.T) {
	for _, mode := range []string{"merged", "settled", "end-write-failure"} {
		t.Run(mode, func(t *testing.T) {
			root, _, _ := integrationFixture(t)
			line, held := heldLine(t, root)
			defer held.Release()
			t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
			ctx, q, err := startQueueOperation(context.Background(), root, git(t, root, "rev-parse", "HEAD"), []string{"merge", "landing"})
			if err != nil {
				t.Fatal(err)
			}
			defer q.finish(context.Background(), "")
			q.op.Workspace = "test:4"
			var clock atomic.Int64
			clock.Store(time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC).UnixNano())
			oldNow := landingNow
			landingNow = func() time.Time { return time.Unix(0, clock.Load()).UTC() }
			defer func() { landingNow = oldNow }()
			dir, err := landingLineDir(mergeRunner, root)
			if err != nil {
				t.Fatal(err)
			}
			var callbackErr error
			readiness := func(vctx context.Context) (landingline.Readiness, bool, error) {
				if mode == "settled" {
					_, waiting, e := line.Status()
					if e != nil || len(waiting) != 1 {
						callbackErr = fmt.Errorf("waiting=%+v err=%v", waiting, e)
						return landingline.Readiness{}, true, nil
					}
					callbackErr = line.Settle(waiting[0].Seq, landingline.Outcome{Train: held.Ticket().Seq, Kind: landingline.OutcomeRefused, Reason: "fixture refusal"})
					if callbackErr != nil {
						return landingline.Readiness{}, true, nil
					}
					<-vctx.Done()
				}
				if mode == "end-write-failure" {
					path := filepath.Join(dir, landingline.EventsFile)
					callbackErr = os.Rename(path, path+".saved")
					if callbackErr == nil {
						callbackErr = os.Mkdir(path, 0755)
					}
				}
				clock.Add(int64(time.Minute))
				return landingline.Readiness{}, mode != "settled", nil
			}
			stderr := watchWriter{want: "no-match-for-this-fixture", seen: make(chan struct{})}
			turn, err := enterLandingLine(ctx, mergeRunner, landingTarget{Root: root, RepoIdentity: "github.com/test/repo"}, "000547-test", func() landingline.Schedule { return landingline.Schedule{Interval: time.Hour, MaxCars: 8} }, readiness, &stderr)
			if callbackErr != nil {
				t.Fatal(callbackErr)
			}
			if turn != nil || err == nil {
				t.Fatalf("turn=%+v err=%v", turn, err)
			}
			if mode != "settled" && !errors.Is(err, errMergedWhileWaiting) {
				t.Fatal(err)
			}
			_, waiting, e := line.Status()
			if e != nil || len(waiting) != 0 {
				t.Fatalf("ticket leaked: %+v %v", waiting, e)
			}
			if mode == "end-write-failure" {
				if !strings.Contains(stderr.String(), "end could not be recorded") {
					t.Fatal(stderr.String())
				}
				return
			}
			events, _, e := line.Events()
			if e != nil {
				t.Fatal(e)
			}
			var starts, ends int
			for _, event := range events {
				switch event.Kind {
				case landingline.EventWaitStart:
					starts++
				case landingline.EventWaitEnd:
					ends++
					if event.At.UnixNano() != clock.Load() {
						t.Fatalf("end before readiness join: %v", event.At)
					}
				}
			}
			if starts != 1 || ends != 1 {
				t.Fatalf("start/end=%d/%d events=%+v", starts, ends, events)
			}
		})
	}
}
