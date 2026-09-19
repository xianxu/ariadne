//go:build manual

// smalldiffreplay_test.go — #231 M3's fixture: run the small-diff review recipe
// ONCE, with the real judge, against a window another repo actually shipped, and
// save what it raised. Manual-tagged: it spends real agent latency.
//
//	REPLAY_REPO=../parley.nvim REPLAY_BASE=bbe05eef REPLAY_HEAD=ac60a055 \
//	REPLAY_ISSUE_NUM=263 REPLAY_OUT=$TMPDIR/sd-263-r1.md \
//	  go test ./cmd/sdlc -tags manual -run TestReplaySmallDiff -v -timeout 30m
//
// It builds the dispatch through boundaryReviewDispatchOptions — the path close
// takes — with Category set to the small-diff recipe, so the prompt, the pinned
// review window and the tool allowlist are exactly what a quick-flow close sends.
// It does not touch any ledger: this is an aim check on the recipe, not a gate.
package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

func TestReplaySmallDiff(t *testing.T) {
	repo, base, head := os.Getenv("REPLAY_REPO"), os.Getenv("REPLAY_BASE"), os.Getenv("REPLAY_HEAD")
	num, _ := strconv.Atoi(os.Getenv("REPLAY_ISSUE_NUM"))
	out := os.Getenv("REPLAY_OUT")
	if repo == "" || base == "" || head == "" || num <= 0 || out == "" {
		t.Skip("set REPLAY_REPO, REPLAY_BASE, REPLAY_HEAD, REPLAY_ISSUE_NUM and REPLAY_OUT")
	}
	chdirTo(t, repo)
	full := func(ref string) string {
		b, err := exec.Command("git", "rev-parse", "--verify", ref+"^{commit}").Output()
		if err != nil {
			t.Fatalf("rev-parse %s: %v", ref, err)
		}
		return strings.TrimSpace(string(b))
	}
	baseLong, headLong := full(base), full(head)
	opts, ok, why := boundaryReviewDispatchOptions(io.Discard, os.Stderr, boundaryReviewParams{
		Label: "#" + strconv.Itoa(num), Base: base, BaseLong: baseLong, Head: headLong,
		IssuesDir: "workshop/issues", IssueNum: num, PlansDir: "workshop/plans",
		Category: judge.SmallDiffReview,
	})
	if !ok {
		t.Fatalf("dispatch options: %s", why)
	}
	if !strings.Contains(opts.Prompt, "## Small-diff focus") {
		t.Fatal("the dispatch did not build the small-diff recipe")
	}
	reply, err := judge.Dispatch(context.Background(), opts)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if err := os.WriteFile(out, []byte(reply), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("reply (%d bytes) written to %s", len(reply), out)
}
