package main

import (
	"reflect"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// #283 D3: shaping under a claim may leave this issue's details modified on the
// resting branch; start-plan carries exactly that. Any entry touching the
// details path on either side of a rename or copy, or any other path, blocks.
func TestPlanningDirtyBlocking(t *testing.T) {
	const d = "workshop/issues/000031-a b \"q\".md"
	for _, c := range []struct {
		name    string
		entries []gitx.StatusEntry
		want    []string
	}{
		{"clean", nil, nil},
		{"details modified (worktree)", []gitx.StatusEntry{{XY: " M", Path: d}}, nil},
		{"details modified (staged + worktree)", []gitx.StatusEntry{{XY: "MM", Path: d}}, nil},
		{"details plus code", []gitx.StatusEntry{{XY: " M", Path: d}, {XY: " M", Path: "cmd/x.go"}}, []string{"cmd/x.go"}},
		{"another issue's details", []gitx.StatusEntry{{XY: " M", Path: "workshop/issues/000032-b.md"}}, []string{"workshop/issues/000032-b.md"}},
		{"details deleted", []gitx.StatusEntry{{XY: " D", Path: d}}, []string{d}},
		{"details added", []gitx.StatusEntry{{XY: "A ", Path: d}}, []string{d}},
		{"renamed away from details", []gitx.StatusEntry{{XY: "R ", Path: "x.md", Orig: d}}, []string{"x.md"}},
		{"renamed onto details", []gitx.StatusEntry{{XY: "R ", Path: d, Orig: "x.md"}}, []string{d}},
		{"copied from details", []gitx.StatusEntry{{XY: "C ", Path: "y.md", Orig: d}}, []string{"y.md"}},
	} {
		if got := planningDirtyBlocking(c.entries, d); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
