package sysvshm

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNativeObservationUsesBoundedPlatformProjection(t *testing.T) {
	for _, platform := range []string{"darwin", "linux", "unsupported"} {
		t.Run(platform, func(t *testing.T) {
			calls := 0
			n := Native{Platform: platform, UID: 501, Owner: "xianxu", Read: func(ctx context.Context, platform string) ([]byte, error) {
				calls++
				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("unbounded capture")
				}
				if platform == "darwin" {
					return []byte(darwinHeader + darwinRow), nil
				}
				return []byte(linuxHeader + linuxRow), nil
			}}
			got := n.Snapshot(context.Background())
			if platform == "unsupported" {
				if calls != 0 || got.Unavailable == "" {
					t.Fatalf("%d %+v", calls, got)
				}
			} else if calls != 1 || got.Unavailable != "" || len(got.Segments) != 1 {
				t.Fatalf("%d %+v", calls, got)
			}
		})
	}
	n := Native{Platform: "linux", Read: func(context.Context, string) ([]byte, error) { return nil, errors.New("denied") }}
	if got := n.Snapshot(context.Background()); !strings.Contains(got.Unavailable, "denied") {
		t.Fatalf("%+v", got)
	}
}

func TestGroupWitnessRequiresKernelMembershipAndIncarnation(t *testing.T) {
	rows := []byte("42 10\n43 10\n44 20\n")
	got := groupWitnesses(rows, 10, func(pid, group int) (Witness, bool) {
		if pid == 42 {
			return Witness{Name: "postgres", Start: "123"}, true
		}
		return Witness{}, false // exited or moved/reused between ps and kernel read
	})
	if len(got) != 1 || got[42].Name != "postgres" {
		t.Fatalf("%+v", got)
	}
	if got := groupWitnesses([]byte("42 invalid\n"), 10, func(int, int) (Witness, bool) { t.Fatal("invalid row probed"); return Witness{}, false }); len(got) != 0 {
		t.Fatalf("%+v", got)
	}
}
