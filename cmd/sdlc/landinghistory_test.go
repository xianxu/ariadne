package main

import (
	"io"
	"strings"
	"testing"
)

// GitHub retains completed PRs when an issue branch is recreated. Exercise
// selection both before and after the integration changes that branch's head.
func TestLandingIgnoresHistoricalPRHeads(t *testing.T) {
	for _, state := range []string{"MERGED", "CLOSED"} {
		for _, phase := range []string{"fresh", "interrupted", "pushed-but-unregistered"} {
			t.Run(state+"/"+phase, func(t *testing.T) {
				root, server, gh := integrationFixture(t)
				older := gh.pr
				older.Number, older.State = 380, state
				older.HeadOID, older.MergeOID = gh.pr.BaseOID, gh.pr.BaseOID
				gh.older = []landingPR{older}
				gh.headLag = 3
				if phase != "fresh" {
					main := procedureHead(t, server)
					git(t, root, "fetch", "-q", "upstream")
					git(t, root, "merge", "--no-ff", "-m", "integration\n\nIntegrated-Main: "+main, main)
					if phase == "pushed-but-unregistered" {
						git(t, root, "push", "-q", "upstream", landingTestBranch)
					}
				}
				if err := runMerge(io.Discard, io.Discard, landingFlags()); err != nil {
					t.Fatalf("historical PR blocked current landing: %v", err)
				}
				if gh.merges != 1 || gh.pr.State != "MERGED" || gh.pr.Number != 1 {
					t.Fatalf("wrong PR landed: merges=%d pr=%+v", gh.merges, gh.pr)
				}
				if gh.older[0] != older {
					t.Fatal("historical PR changed")
				}
				if got := git(t, root, "branch", "--show-current"); got != "main-slot1" {
					t.Fatalf("landing did not finish cleanup: %s", got)
				}
			})
		}
	}
}

func TestLandingTwoOpenPRsRefuseWithIdentities(t *testing.T) {
	for _, sameHead := range []bool{false, true} {
		t.Run(map[bool]string{false: "different-heads", true: "same-head"}[sameHead], func(t *testing.T) {
			root, _, gh := integrationFixture(t)
			other := gh.pr
			other.Number = 380
			if !sameHead {
				other.HeadOID = gh.pr.BaseOID
			}
			gh.older = []landingPR{other}
			before := procedureHead(t, root)
			err := runMerge(io.Discard, io.Discard, landingFlags())
			if err == nil {
				t.Fatal("two OPEN PRs accepted")
			}
			for _, want := range []string{"#1", "#380"} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("ambiguity does not identify %s: %v", want, err)
				}
			}
			if gh.merges != 0 || procedureHead(t, root) != before || git(t, gh.remote, "rev-parse", "refs/heads/"+landingTestBranch) != before {
				t.Fatal("ambiguous selection changed the branch or merged")
			}
		})
	}
}

func TestSelectLandingCandidate(t *testing.T) {
	head := strings.Repeat("a", 40)
	open := landingPR{Number: 1, State: "OPEN", Repo: "test/repo", HeadRef: landingTestBranch, HeadOID: head, BaseRef: "main"}
	history := open
	history.Number, history.State, history.HeadOID = 380, "MERGED", strings.Repeat("b", 40)
	currentMerged := open
	currentMerged.State = "MERGED"
	wrongRepo := history
	wrongRepo.Repo = "other/repo"
	for _, tc := range []struct {
		name string
		prs  []landingPR
		head string
		want int
	}{
		{"historical-head", []landingPR{history, open}, head, 1},
		{"current-merged-recovery", []landingPR{currentMerged}, head, 1},
		{"history-alone-is-not-current", []landingPR{history}, head, 0},
		{"absent-local-ref-single-recovery", []landingPR{history}, "", 380},
		{"ambiguous-current-evidence", []landingPR{currentMerged, open}, head, 0},
		{"historical-identity-still-validated", []landingPR{wrongRepo, open}, head, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := selectLandingCandidate(tc.prs, landingTarget{Repo: "test/repo"}, landingTestBranch, tc.head)
			if tc.want == 0 {
				if err == nil {
					t.Fatalf("uncertain evidence accepted: %+v", got)
				}
				return
			}
			if err != nil || got.Number != tc.want {
				t.Fatalf("selection %+v %v; want #%d", got, err, tc.want)
			}
		})
	}
}
