package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func TestOwnedBranchRequestReadsOncePerCapturedIdentity(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			ctx := ownedBranchRequestContext(t.Context())
			request := ctx.Value(ownedBranchRequestKey{}).(*ownedBranchRequest)
			histories, captures := 0, 0
			request.readHistory = func(*tracker.Repository, string, string) ([]tracker.OwnershipEvent, error) {
				histories++
				return []tracker.OwnershipEvent{{Status: "working"}}, nil
			}
			request.capture = func(root string, since, until time.Time) ownedBranchCapture {
				captures++
				return ownedBranchCapture{Root: root, Unavailable: "private counted unknown"}
			}
			for repeat := 0; repeat < 100; repeat++ {
				for i := 0; i < n; i++ {
					if _, err := loadOwnedBranchHistory(ctx, nil, "pin", fmt.Sprint(i)); err != nil {
						t.Fatal(err)
					}
					captureOwnedBranchCached(ctx, fmt.Sprint(i), time.Now(), time.Now().Add(time.Second))
				}
			}
			if histories != n || captures != n {
				t.Fatalf("per-event IO: histories=%d captures=%d roots=%d", histories, captures, n)
			}
			// A different immutable tracker generation is a distinct authority.
			if _, err := loadOwnedBranchHistory(ctx, nil, "next-pin", "0"); err != nil {
				t.Fatal(err)
			}
			if histories != n+1 {
				t.Fatal("reused history across different pins")
			}
		})
	}
}
