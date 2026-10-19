package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The vectors are pair's own (cmd/internal/couchcore/hub_claim_test.go,
// TestHubClaimsParse): the contract tl-lead's copy of the grammar keeps.
func TestHubClaimVectors(t *testing.T) {
	claims, err := parseHub([]byte("# team\npair\nparley.nvim:7\n\n"))
	if err != nil || len(claims) != 2 || claims[0] != (hubClaim{"pair", -1}) || claims[1] != (hubClaim{"parley.nvim", 7}) {
		t.Fatalf("claims %+v, err %v", claims, err)
	}
	for _, raw := range []string{"../pair", "/work/pair", "pair:-1", "pair:01", "pair:1:2", "!pair:7", "pair\x00", strings.Repeat("x", 65537)} {
		if _, err := parseHub([]byte(raw)); err == nil {
			t.Errorf("accepted %q", raw[:min(len(raw), 40)])
		}
	}
	if _, err := parseHub([]byte("pair\nnot a claim\n")); err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Errorf("a malformed line must refuse naming it: %v", err)
	}
}

func TestOwnerOf(t *testing.T) {
	ops := hub{Name: "ops", Claims: []hubClaim{{"pair", -1}, {"ariadne", -1}}}
	echo := hub{Name: "ops-echo", Claims: []hubClaim{{"echo", -1}, {"pair", 7}}}
	other := hub{Name: "ops-x", Claims: []hubClaim{{"ariadne", -1}}}
	broken := hub{Name: "ops-y", Err: os.ErrPermission}
	hubs := []hub{ops, echo, other, broken}
	for _, tc := range []struct {
		repo  string
		slot  int
		want  []string
		exact bool
		none  bool
	}{
		{"pair", 3, []string{"ops"}, false, false},
		{"pair", 7, []string{"ops-echo"}, true, false},         // exact beats repo
		{"ariadne", 2, []string{"ops", "ops-x"}, false, false}, // a tie: every claimant
		{"ops-echo", 0, []string{"ops-echo"}, true, false},     // a hub owns its own :0
		{"brain", 1, nil, false, true},                         // unclaimed: the first hub's, unseen
		{"ops-y", 0, []string{"ops-y"}, true, false},           // a hub with a broken .hub still owns its :0
	} {
		got := ownerOf(tc.repo, tc.slot, hubs)
		if strings.Join(got.Owners, ",") != strings.Join(tc.want, ",") || got.Exact != tc.exact || got.Unclaimed != tc.none {
			t.Errorf("%s:%d = %+v", tc.repo, tc.slot, got)
		}
		if _, ok := got.sole(); ok != (len(tc.want) == 1) {
			t.Errorf("%s:%d sole = %v", tc.repo, tc.slot, ok)
		}
	}
}

func TestReadHubs(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("ops/.hub", "ariadne\npair\n")
	write("ops-echo/.hub", "echo\n")
	write("ops-bad/.hub", "../x\n")
	write("pair/README.md", "a product repo, no .hub\n")
	write("ops/teams/devinfra/.hub", "ignored: only a repo root's .hub counts\n")
	hubs, err := readHubs(ws)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, h := range hubs {
		names = append(names, h.Name)
	}
	if strings.Join(names, ",") != "ops,ops-bad,ops-echo" {
		t.Fatalf("hubs %v", names)
	}
	if hubs[1].Err == nil || len(hubs[0].Claims) != 2 || hubs[2].address() != "ops-echo:0" {
		t.Fatalf("hubs %+v", hubs)
	}
}
