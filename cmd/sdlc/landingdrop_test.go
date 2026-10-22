package main

import (
	"github.com/xianxu/ariadne/pkg/workspace"
	"strings"
	"testing"
)

func TestLandingDropRequiresIssue(t *testing.T) {
	_, err := runLanding(t, testLandingDeps(slotPtr(0), slotPtr(0)), "drop")
	if err == nil || !strings.Contains(err.Error(), "--issue N") {
		t.Fatalf("drop must name the required issue, got %v", err)
	}
}

func TestAuthorizeTrainDrop(t *testing.T) {
	own := workspace.Identity{RepoIdentity: "repo", WorktreeRoot: "/fleet/worktree/sample-slot2/sample", Address: dropAddress("sample:2"), Slot: slotPtr(2)}
	target := workspace.Identity{RepoIdentity: "repo", WorktreeRoot: "/fleet/sample", Address: dropAddress("sample:0"), Slot: slotPtr(0)}
	other := workspace.Identity{RepoIdentity: "repo", WorktreeRoot: "/fleet/worktree/sample-slot3/sample", Address: dropAddress("sample:3"), Slot: slotPtr(3)}
	foreign := target
	foreign.RepoIdentity = "foreign"
	for _, tc := range []struct {
		name                  string
		caller, target, owner workspace.Identity
		tree                  *int
		allow                 bool
	}{
		{"owner", own, target, own, slotPtr(2), true},
		{"steward", target, target, own, slotPtr(0), true},
		{"foreign worker", other, target, own, slotPtr(3), false},
		{"repo flag cannot impersonate", foreign, target, own, nil, false},
		{"worker entered zero", target, target, own, slotPtr(3), false},
		{"unknown owner", own, target, workspace.Identity{}, nil, false},
		{"unknown caller", workspace.Identity{}, target, own, nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := authorizeTrainDrop(tc.caller, tc.target, tc.owner, tc.tree)
			if (err == nil) != tc.allow {
				t.Fatalf("allow%v error%v", tc.allow, err)
			}
		})
	}
}

func dropAddress(s string) *string { return &s }
