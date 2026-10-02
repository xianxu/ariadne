package fleet

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

var (
	meFP    = issue.MachineFingerprint("this-machine")
	otherFP = issue.MachineFingerprint("other-machine")
	me      = MachineFrom(MachineIdentity{Fingerprint: meFP, Name: "here"}, nil)
)

func claimantOn(machine, worktree string) *Claimant {
	return &Claimant{Operator: "op", Machine: machine, MachineName: "m", Worktree: worktree, Repository: "r"}
}

func claimsRow(repo, tree string) TreeRow {
	row := validTreeRow()
	row.RepoIdentity, row.RepoRoot, row.TreePath = repo+"/.git", repo, tree
	return row
}

// #288: placement over the cross product of a repository's read quality, the
// machine identity's availability, and each claimant class. Only this
// machine's active claims are local state; one on a tree is placed there, one
// on no tree is dangling; nothing is placed or dangling unless the read
// carries value and the identity is known.
func TestPlaceClaims(t *testing.T) {
	classes := []ClaimCard{
		{Ref: "a#000001", Status: "working", Revision: "r1", Claimant: claimantOn(meFP, "/a/slot")},    // placed
		{Ref: "a#000002", Status: "blocked", Revision: "r2", Claimant: claimantOn(meFP, "/gone")},      // dangling
		{Ref: "a#000003", Status: "working", Revision: "r3", Claimant: claimantOn(otherFP, "/a/slot")}, // other machine
		{Ref: "a#000004", Status: "working", Revision: "r4"},                                           // unattributed
		{Ref: "a#000005", Status: "open", Revision: "r5", Claimant: claimantOn(meFP, "/a/slot")},       // inactive
		{Ref: "a#000006", Status: "done", Revision: "r6", Claimant: claimantOn(meFP, "/a/slot")},       // inactive
	}
	unknownMe := MachineFrom(MachineIdentity{}, errors.New("ioreg failed"))
	for _, tc := range []struct {
		name         string
		read         RepoClaims
		machine      Machine
		state        string
		placed, dang []string
		errHas       string
	}{
		{"present", RepoClaims{State: ClaimsPresent, Cards: classes}, me, ClaimsPresent, []string{"a#000001"}, []string{"a#000002"}, ""},
		{"stale", RepoClaims{State: ClaimsStale, Error: "offline", Cards: classes}, me, ClaimsStale, []string{"a#000001"}, []string{"a#000002"}, "offline"},
		{"partial", RepoClaims{State: ClaimsPartial, Error: unreadableError([]string{"a#000009"}), Cards: classes, Unreadable: []string{"a#000009"}}, me, ClaimsPartial, []string{"a#000001"}, []string{"a#000002"}, "a#000009"},
		{"unknown read", RepoClaims{State: ClaimsUnknown, Error: "no tracker read"}, me, ClaimsUnknown, nil, nil, "no tracker read"},
		{"no tracker", RepoClaims{State: ClaimsAbsent}, me, ClaimsAbsent, nil, nil, ""},
		{"present, identity unknown", RepoClaims{State: ClaimsPresent, Cards: classes}, unknownMe, ClaimsUnknown, nil, nil, "ioreg failed"},
		{"no tracker, identity unknown", RepoClaims{State: ClaimsAbsent}, unknownMe, ClaimsAbsent, nil, nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows := []TreeRow{claimsRow("/a", "/a"), claimsRow("/a", "/a/slot")}
			got, dangling := PlaceClaims(rows, map[string]RepoClaims{"/a/.git": tc.read}, tc.machine)
			var placed []string
			for _, row := range got {
				if row.ClaimsState != tc.state || !strings.Contains(row.ClaimsError, tc.errHas) || row.Claims == nil {
					t.Fatalf("row %s: state %q error %q", row.TreePath, row.ClaimsState, row.ClaimsError)
				}
				if err := validateClaims(row.ClaimsState, row.ClaimsError, row.Claims); err != nil {
					t.Fatalf("row %s violates the contract: %v", row.TreePath, err)
				}
				for _, c := range row.Claims {
					if row.TreePath != "/a/slot" {
						t.Fatalf("claim %s placed on %s", c.Ref, row.TreePath)
					}
					placed = append(placed, c.Ref)
				}
			}
			var dang []string
			for _, d := range dangling {
				if d.RepoIdentity != "/a/.git" || d.RepoRoot != "/a" {
					t.Fatalf("dangling repo: %+v", d)
				}
				dang = append(dang, d.Ref)
			}
			if strings.Join(placed, ",") != strings.Join(tc.placed, ",") || strings.Join(dang, ",") != strings.Join(tc.dang, ",") {
				t.Fatalf("placed %v dangling %v, want %v %v", placed, dang, tc.placed, tc.dang)
			}
		})
	}
	t.Run("a repository never read", func(t *testing.T) {
		got, _ := PlaceClaims([]TreeRow{claimsRow("/b", "/b")}, map[string]RepoClaims{}, me)
		if got[0].ClaimsState != ClaimsUnknown || got[0].ClaimsError == "" {
			t.Fatalf("%+v", got[0])
		}
	})
}

// #288: the contract round-trips, and each invariant has a rejection.
func TestInventoryClaimsContract(t *testing.T) {
	rows := []TreeRow{claimsRow("/a", "/a/slot")}
	rows, dangling := PlaceClaims(rows, map[string]RepoClaims{"/a/.git": {State: ClaimsPresent, Cards: []ClaimCard{
		{Ref: "a#000001", Status: "working", Revision: "r1", Claimant: claimantOn(meFP, "/a/slot")},
		{Ref: "a#000002", Status: "working", Revision: "r2", Claimant: claimantOn(meFP, "/gone")},
	}}}, me)
	inv := Inventory{Rows: rows, Diagnostics: []RepoDiagnostic{}, Machine: me, DanglingClaims: dangling}
	raw, err := json.Marshal(inv)
	if err != nil {
		t.Fatal(err)
	}
	var back Inventory
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("%v\n%s", err, raw)
	}
	if back.Machine.Fingerprint != meFP || len(back.Rows[0].Claims) != 1 || len(back.DanglingClaims) != 1 || back.DanglingClaims[0].Claimant.Worktree != "/gone" {
		t.Fatalf("round trip: %s", raw)
	}
	for name, mutate := range map[string]func(string) string{
		"null claims": func(s string) string { return strings.Replace(s, `"claims":[{`, `"claims":null,"x":[{`, 1) },
		"unknown state": func(s string) string {
			return strings.Replace(s, `"claims_state":"present"`, `"claims_state":"maybe"`, 1)
		},
		"present with error": func(s string) string {
			return strings.Replace(s, `"claims_state":"present"`, `"claims_state":"present","claims_error":"x"`, 1)
		},
		"stale without error": func(s string) string {
			return strings.Replace(s, `"claims_state":"present"`, `"claims_state":"stale"`, 1)
		},
		"unknown with claims": func(s string) string {
			return strings.Replace(s, `"claims_state":"present"`, `"claims_state":"unknown","claims_error":"x"`, 1)
		},
		"raw machine id":  func(s string) string { return strings.Replace(s, `"fingerprint":"`+meFP, `"fingerprint":"RAW`, 1) },
		"missing machine": func(s string) string { return strings.Replace(s, `"machine":{`, `"machinx":{`, 1) },
		"null dangling claims": func(s string) string {
			return strings.Replace(s, `"dangling_claims":[`, `"dangling_claims":null,"y":[`, 1)
		},
		"inactive claim": func(s string) string { return strings.Replace(s, `"status":"working"`, `"status":"open"`, 1) },
	} {
		t.Run(name, func(t *testing.T) {
			bad := mutate(string(raw))
			if bad == string(raw) {
				t.Fatal("mutation did not apply")
			}
			var inv Inventory
			if err := json.Unmarshal([]byte(bad), &inv); err == nil {
				t.Fatalf("accepted %s", bad)
			}
		})
	}
}
