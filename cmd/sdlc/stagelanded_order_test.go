package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func TestLandedStageCaller(t *testing.T) {
	lease := fleet.StageLease{Holder: "#1", Branch: "000001-procedure", Slot: "sample:1", Commit: strings.Repeat("a", 40)}
	caller := moveSide{Address: "sample:1", Branch: "main-slot1", Resting: "main-slot1"}
	for _, tc := range []struct {
		name   string
		change func(*fleet.StageLease, *moveSide)
		want   bool
	}{
		{"original resting slot", func(*fleet.StageLease, *moveSide) {}, true},
		{"another slot", func(_ *fleet.StageLease, c *moveSide) { c.Address = "sample:2" }, false},
		{"no slot", func(l *fleet.StageLease, _ *moveSide) { l.Slot = "" }, false},
		{"relative slot", func(l *fleet.StageLease, c *moveSide) { l.Slot, c.Address = ":1", ":1" }, false},
		{"noncanonical slot", func(l *fleet.StageLease, c *moveSide) { l.Slot, c.Address = "sample:01", "sample:01" }, false},
		{"detached", func(_ *fleet.StageLease, c *moveSide) { c.Branch = "" }, false},
		{"another issue", func(_ *fleet.StageLease, c *moveSide) { c.Branch = "000002-other" }, false},
		{"no rest", func(_ *fleet.StageLease, c *moveSide) { c.Resting, c.Branch = "", "" }, false},
		{"holder mismatch", func(l *fleet.StageLease, _ *moveSide) { l.Holder = "#2" }, false},
		{"branch holder", func(l *fleet.StageLease, _ *moveSide) { l.Holder, l.Branch = "branch:feature", "feature" }, false},
		{"legacy commit only", func(l *fleet.StageLease, _ *moveSide) { l.Holder = "" }, false},
		{"missing commit", func(l *fleet.StageLease, _ *moveSide) { l.Commit = "" }, false},
		{"zero commit", func(l *fleet.StageLease, _ *moveSide) { l.Commit = strings.Repeat("0", 40) }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l, c := lease, caller
			tc.change(&l, &c)
			if got := landedStageCaller(l, c); got != tc.want {
				t.Fatalf("candidate=%v, want %v", got, tc.want)
			}
		})
	}
}

// Mutate private resources at the existing GitHub evidence boundary; no sleeps
// or test-only production seam. The stage lock must remain held during proof.
type stageProofHookGH struct {
	*landingFakeGH
	after func(context.Context)
}

func (g *stageProofHookGH) LandingPRs(ctx context.Context, repo, branch string) ([]landingPR, error) {
	prs, err := g.landingFakeGH.LandingPRs(ctx, repo, branch)
	if g.after != nil {
		f := g.after
		g.after = nil
		f(ctx)
	}
	return prs, err
}

func TestStageClearLandedProofOrdering(t *testing.T) {
	for _, mutation := range []string{"caller", "target", "lease", "completion", "completion after archive", "cancel", "lock"} {
		t.Run(mutation, func(t *testing.T) {
			f := newLandedStageFixture(t, "merge")
			stagedLandingBuild(t, f.roots[0])
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			beforeHead := procedureHead(t, f.roots[0])
			beforeLease, err := fleet.ReadStageLease(execGitRunner{}, f.roots[0])
			if err != nil {
				t.Fatal(err)
			}
			wantLease := beforeLease
			wantBranch := ""
			archiveHooked := false
			if mutation == "completion after archive" {
				previous := landingArchiveRunner
				t.Cleanup(func() { landingArchiveRunner = previous })
				landingArchiveRunner = landingHookRunner{gitRunner: previous, after: func(_ string, args []string) error {
					if !archiveHooked && len(args) > 0 && args[0] == "diff-tree" {
						archiveHooked = true
						f.changeCard(t, func(raw []byte) ([]byte, error) {
							b, _, err := issue.CardCompletion(raw)
							if err != nil {
								return nil, err
							}
							b.Token += "-after-proof"
							return issue.SetCardCompletion(raw, b)
						})
					}
					return nil
				}}
			}
			hooked := false
			ghClient = &stageProofHookGH{landingFakeGH: f.gh, after: func(proofCtx context.Context) {
				hooked = true
				if _, bounded := proofCtx.Deadline(); !bounded {
					t.Fatal("unbounded landing proof")
				}
				switch mutation {
				case "caller":
					git(t, f.roots[1], "switch", "-qc", "000002-other")
				case "target":
					git(t, f.roots[0], "switch", "-qc", "occupied")
					wantBranch = "occupied"
				case "lease":
					wantLease.Holder, wantLease.Branch = "#2", "000002-other"
					if err := fleet.WriteStageLease(execGitRunner{}, f.roots[0], wantLease); err != nil {
						t.Fatal(err)
					}
				case "completion":
					f.changeCard(t, func(raw []byte) ([]byte, error) {
						b, _, err := issue.CardCompletion(raw)
						if err != nil {
							return nil, err
						}
						b.Token += "-replacement"
						return issue.SetCardCompletion(raw, b)
					})
				case "cancel":
					cancel()
				case "lock":
					unlock, err := fleet.LockStage(execGitRunner{}, f.roots[0])
					if err == nil {
						unlock()
						t.Fatal("proof released the stage lock")
					}
				}
			}}
			err = runUnstage(ctx, f.roots[1], ":0", "", false, io.Discard, io.Discard)
			if mutation == "completion after archive" && (!archiveHooked || err == nil || !strings.Contains(err.Error(), "changed during")) {
				t.Fatalf("final fresh completion check was not exercised: %v", err)
			}
			if !hooked {
				t.Fatal("proof boundary was not exercised")
			}
			if mutation == "lock" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "sdlc stage --clear") {
				t.Fatalf("changed authority must refuse with exact retry: %v", err)
			}
			afterLease, readErr := fleet.ReadStageLease(execGitRunner{}, f.roots[0])
			if readErr != nil || afterLease != wantLease || procedureHead(t, f.roots[0]) != beforeHead || git(t, f.roots[0], "branch", "--show-current") != wantBranch {
				t.Fatalf("refusal changed target or lease: %+v %v", afterLease, readErr)
			}
			if _, err := os.Stat(filepath.Join(f.roots[0], "bin")); !os.IsNotExist(err) {
				t.Fatalf("refusal ran build: %v", err)
			}
		})
	}
}
