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
