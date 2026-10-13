package main

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// #344: the target states every stage verb meets, crossed with stage and
// --clear. Each row is the acceptance or the refusal; the real-git tests prove
// a refused run changes nothing.
func TestCheckStageAndUnstage(t *testing.T) {
	change := []gitx.StatusEntry{{XY: " M", Path: "README"}}
	resting := moveSide{Root: "/s0", Address: ":0", Branch: "main", Resting: "main", Head: "b", Untracked: []string{"scratch"}}
	for _, tc := range []struct {
		name         string
		to           func(moveSide) moveSide
		stage, clear string // "" means allowed
		staged       bool   // --clear has something to clear
	}{
		{"resting", func(s moveSide) moveSide { return s }, "", "", false},
		{"dirty", func(s moveSide) moveSide { s.Changes = change; return s }, "uncommitted changes", "", false},
		{"mid-operation", func(s moveSide) moveSide { s.Operation = "MERGE_HEAD"; return s }, "operation in progress", "", false},
		{"other branch", func(s moveSide) moveSide { s.Branch = "000002-y"; return s }, "not its resting branch", "nothing to clear", false},
		{"detached, leavable (staged or in a ref)", func(s moveSide) moveSide { s.Branch, s.Leavable = "", true; return s }, "", "", true},
		{"detached, leavable but dirty", func(s moveSide) moveSide { s.Branch, s.Leavable, s.Changes = "", true, change; return s }, "uncommitted changes", "uncommitted changes", true},
		{"detached, unreferenced", func(s moveSide) moveSide { s.Branch = ""; return s }, "in no branch and was not staged", "in no branch and was not staged", false},
		{"collision", func(s moveSide) moveSide { s.Untracked = []string{"feature"}; return s }, "would overwrite: feature", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			to := tc.to(resting)
			f := stageFacts{From: moveSide{Root: "/s1", Address: ":1", Branch: "000001-x", Resting: "main-slot1", Head: "a1b2c3d4e5"}, To: to, SameRepo: true, Incoming: []string{"feature", "README"}}
			expect(t, "stage", checkStage(f), tc.stage)
			staged, err := checkUnstage(to)
			expect(t, "clear", err, tc.clear)
			if staged != tc.staged {
				t.Errorf("clear: staged = %v, want %v", staged, tc.staged)
			}
		})
	}
	base := stageFacts{From: moveSide{Root: "/s1", Address: ":1", Head: "a"}, To: resting, SameRepo: true}
	for _, tc := range []struct {
		name   string
		mutate func(*stageFacts)
		want   string
	}{
		{"different repo", func(f *stageFacts) { f.SameRepo = false }, "different repositories"},
		{"same slot", func(f *stageFacts) { f.To.Root = f.From.Root }, "is the current slot"},
		{"no source commit", func(f *stageFacts) { f.From.Head = "" }, "no commit to stage"},
		{"dirty source stages its HEAD", func(f *stageFacts) { f.From.Changes = change }, ""},
	} {
		f := base
		tc.mutate(&f)
		expect(t, tc.name, checkStage(f), tc.want)
	}
}

func expect(t *testing.T, what string, err error, want string) {
	t.Helper()
	if want == "" {
		if err != nil {
			t.Errorf("%s refused: %v", what, err)
		}
		return
	}
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("%s: err = %v, want it to contain %q", what, err, want)
	}
}
