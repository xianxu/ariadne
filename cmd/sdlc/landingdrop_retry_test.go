package main

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func TestLandingDropRetryReleasesRetainedHead(t *testing.T) {
	for _, mode := range []string{"early error", "cancellation", "resume before enqueue", "vouched handoff"} {
		t.Run(mode, func(t *testing.T) {
			line, err := landingline.Open(t.TempDir(), landingline.Options{})
			if err != nil {
				t.Fatal(err)
			}
			h, err := line.Enqueue(landingline.Ticket{Issue: 1, Branch: landingTestBranch})
			if err != nil {
				t.Fatal(err)
			}
			defer h.Release()
			if err := line.Wait(context.Background(), h); err != nil {
				t.Fatal(err)
			}
			turn := &landingTurn{line: line, handle: h, log: &lineLog{line: line, stderr: io.Discard}, stderr: io.Discard}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			failure := errors.New("next attempt failed before enqueue")
			attempts := 0
			err = retryTrainReformation(ctx, io.Discard, func(ctx context.Context) error {
				attempts++
				if attempts == 1 {
					if mode == "cancellation" {
						cancel()
					}
					return &trainReformError{train: h.Ticket().Seq, turn: turn}
				}
				if attempts > 2 {
					t.Fatal("unexpected repeated attempt")
				}
				switch mode {
				case "early error":
					return failure
				case "resume before enqueue":
					return nil
				case "vouched handoff":
					fresh, err := line.Enqueue(landingline.Ticket{Issue: 1, Branch: landingTestBranch})
					if err != nil {
						return err
					}
					defer fresh.Release()
					if err := line.Vouch(fresh, landingline.Readiness{Ready: true, Branch: landingTestBranch, Head: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", PR: 1}); err != nil {
						return err
					}
					release, ok := ctx.Value(trainReformReleaseKey{}).(func())
					if !ok {
						t.Fatal("missing scoped handoff")
					}
					release()
					head, _, err := line.Status()
					if err != nil {
						return err
					}
					if head != nil {
						t.Fatal("vouched handoff retained old head")
					}
					return nil
				}
				return errors.New("unexpected next attempt")
			})
			if mode == "early error" && !errors.Is(err, failure) {
				t.Errorf("error%v", err)
			}
			if mode == "cancellation" && !errors.Is(err, context.Canceled) {
				t.Errorf("error%v", err)
			}
			if (mode == "resume before enqueue" || mode == "vouched handoff") && err != nil {
				t.Errorf("error%v", err)
			}
			head, waiting, statusErr := line.Status()
			if statusErr != nil || head != nil || len(waiting) != 0 {
				t.Fatalf("retained resources after return: head%v waiting%v err%v", head, waiting, statusErr)
			}
		})
	}
}
