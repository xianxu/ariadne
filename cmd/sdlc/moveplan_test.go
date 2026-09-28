package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

func TestUntrackedCollisions(t *testing.T) {
	incoming := []string{"a/b", "c", "README"}
	for _, tc := range []struct {
		untracked []string
		want      []string
	}{
		{[]string{"README"}, []string{"README"}},
		{[]string{"a"}, []string{"a"}},     // untracked file where the branch has a directory
		{[]string{"c/d"}, []string{"c/d"}}, // untracked under a path the branch has as a file
		{[]string{"ab", "a/bc", "cd"}, nil},
		{[]string{"scratch", "README"}, []string{"README"}},
	} {
		if got := untrackedCollisions(tc.untracked, incoming); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("untrackedCollisions(%v) = %v, want %v", tc.untracked, got, tc.want)
		}
	}
}

func TestCheckMove(t *testing.T) {
	base := func() moveFacts {
		return moveFacts{
			From:     moveSide{Root: "/s1", Address: ":1", Branch: "000001-x", Resting: "main-slot1", Head: "a"},
			To:       moveSide{Root: "/s0", Address: ":0", Branch: "main", Resting: "main", Head: "b", Untracked: []string{"scratch"}},
			SameRepo: true,
			Incoming: []string{"feature", "README"},
			Parked:   []string{"b local main work"},
		}
	}
	change := []gitx.StatusEntry{{XY: " M", Path: "README"}}
	for _, tc := range []struct {
		name   string
		mutate func(*moveFacts)
		want   string // "" means allowed
	}{
		{"valid", func(*moveFacts) {}, ""},
		{"different repo", func(f *moveFacts) { f.SameRepo = false }, "different repositories"},
		{"same slot", func(f *moveFacts) { f.To.Root = f.From.Root }, "already the current slot"},
		{"source detached", func(f *moveFacts) { f.From.Branch = "" }, "detached"},
		{"source resting", func(f *moveFacts) { f.From.Branch = "main-slot1" }, "no issue branch to move"},
		{"source changes", func(f *moveFacts) { f.From.Changes = change }, ":1 has uncommitted changes ( M README)"},
		{"source untracked", func(f *moveFacts) { f.From.Untracked = []string{"tmp"} }, ":1 has untracked files (tmp)"},
		{"source operation", func(f *moveFacts) { f.From.Operation = "MERGE_HEAD" }, ":1 has a Git operation in progress"},
		{"destination not resting", func(f *moveFacts) { f.To.Branch = "000002-y" }, "not its resting branch"},
		{"destination changes", func(f *moveFacts) { f.To.Changes = change }, ":0 has uncommitted changes"},
		{"destination operation", func(f *moveFacts) { f.To.Operation = "rebase-merge" }, ":0 has a Git operation in progress"},
		{"destination collision", func(f *moveFacts) { f.To.Untracked = []string{"scratch", "feature"} }, "would overwrite: feature"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := base()
			tc.mutate(&f)
			err := checkMove(f)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("refused: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}
