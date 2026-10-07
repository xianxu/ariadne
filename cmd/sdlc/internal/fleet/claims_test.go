package fleet

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
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
		{Ref: "a#000005", Status: "open", Revision: "r5", Claimant: claimantOn(meFP, "/a/slot")},       // a shaping claim: holds the lock (#283)
		{Ref: "a#000006", Status: "done", Revision: "r6", Claimant: claimantOn(meFP, "/a/slot")},       // terminal: attribution, no lock
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
		{"present", RepoClaims{State: ClaimsPresent, Cards: classes}, me, ClaimsPresent, []string{"a#000001", "a#000005"}, []string{"a#000002"}, ""},
		{"stale", RepoClaims{State: ClaimsStale, Error: "offline", Cards: classes}, me, ClaimsStale, []string{"a#000001", "a#000005"}, []string{"a#000002"}, "offline"},
		{"partial", RepoClaims{State: ClaimsPartial, Error: unreadableError([]string{"a#000009"}), Cards: classes, Unreadable: []string{"a#000009"}}, me, ClaimsPartial, []string{"a#000001", "a#000005"}, []string{"a#000002"}, "a#000009"},
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
		"null claims": func(s string) string {
			return editJSON(t, s, func(m map[string]any) { m["rows"].([]any)[0].(map[string]any)["claims"] = nil })
		},
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
		"raw machine id":       func(s string) string { return strings.Replace(s, `"fingerprint":"`+meFP, `"fingerprint":"RAW`, 1) },
		"missing machine":      func(s string) string { return editJSON(t, s, func(m map[string]any) { delete(m, "machine") }) },
		"null dangling claims": func(s string) string { return editJSON(t, s, func(m map[string]any) { m["dangling_claims"] = nil }) },
		"lockless claim":       func(s string) string { return strings.Replace(s, `"status":"working"`, `"status":"done"`, 1) }, // #283: terminal holds no lock
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

// #288 BR-9: two clones of one repository read the same tracker. A claim on
// clone B's tree is placed on B's row and is not dangling from A's read; a
// claim on no tree is dangling once, not once per clone.
func TestPlaceClaimsAcrossClones(t *testing.T) {
	cards := []ClaimCard{
		{Ref: "x#000001", Status: "working", Revision: "r1", Claimant: claimantOn(meFP, "/b")},
		{Ref: "x#000002", Status: "working", Revision: "r2", Claimant: claimantOn(meFP, "/gone")},
	}
	rowsIn := []TreeRow{claimsRow("/a", "/a"), claimsRow("/b", "/b")}
	byRepo := map[string]RepoClaims{"/a/.git": {State: ClaimsPresent, Cards: cards}, "/b/.git": {State: ClaimsPresent, Cards: cards}}
	rows, dangling := PlaceClaims(rowsIn, byRepo, me)
	if len(rows[0].Claims) != 0 || len(rows[1].Claims) != 1 || rows[1].Claims[0].Ref != "x#000001" {
		t.Fatalf("placement: %+v / %+v", rows[0].Claims, rows[1].Claims)
	}
	if len(dangling) != 1 || dangling[0].Ref != "x#000002" {
		t.Fatalf("dangling must be the one claim on no tree, once: %+v", dangling)
	}
}

// #288 BR-10: the claim read's quality, derived from the composed records,
// over its state space: tracker presence × freshness × unreadable cards ×
// claimant record, with duplicates never counted twice.
func TestRepoClaimsFrom(t *testing.T) {
	base := "---\nid: 000001\nstatus: working\ncreated: 2026-10-01\nupdated: 2026-10-02\n---\n\n# C\n\n## Problem\n\nx\n"
	mk := func(id, status string, owner *issue.Claimant) tracker.Record {
		raw := strings.NewReplacer("000001", id, "status: working", "status: "+status).Replace(base)
		b := []byte(raw)
		if owner != nil {
			var err error
			if b, err = issue.SetCardClaimant(b, *owner); err != nil {
				t.Fatal(err)
			}
		}
		card, err := issue.ParseCard(b)
		if err != nil {
			t.Fatal(err)
		}
		return tracker.Record{ID: id, Path: "workshop/issue-cards/" + id + "-c.md", BlobOID: "blob" + id, Card: card, Raw: b}
	}
	owner := issue.Claimant{Operator: "op", Machine: meFP, MachineName: "m", Worktree: "/w/raw", Repository: "r"}
	cards := []tracker.Record{mk("000001", "working", &owner), mk("000002", "open", nil)}
	bad := []tracker.UnreadableCard{{ID: "000003", Path: "workshop/issue-cards/000003-c.md", Err: errors.New("invalid claimant")}}
	dup := []tracker.DetailFile{{Path: "/d/000001-c.md", Raw: []byte(base)}, {Path: "/d/000001-copy.md", Raw: []byte(base)}}
	canon := func(p string) string { return "/canon" + p }
	offline := errors.New("dial: no route")
	for _, tc := range []struct {
		name       string
		rs         tracker.Records
		unreadable []tracker.UnreadableCard
		state      string
		errHas     string
		cards      int
	}{
		{"no tracker", tracker.Records{}, nil, ClaimsAbsent, "", 0},
		{"unconfirmed tracker", tracker.Records{Stale: true}, nil, ClaimsUnknown, "could not confirm", 0},
		{"never fetched", tracker.Records{Stale: true, FetchErr: offline}, nil, ClaimsUnknown, "no route", 0},
		{"present", tracker.Records{Tracker: true}, nil, ClaimsPresent, "", 2},
		{"stale", tracker.Records{Tracker: true, Stale: true, FetchErr: offline}, nil, ClaimsStale, "no route", 2},
		{"partial", tracker.Records{Tracker: true}, bad, ClaimsPartial, "r#000003", 2},
		{"stale and partial", tracker.Records{Tracker: true, Stale: true, FetchErr: offline}, bad, ClaimsPartial, "no route", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var in []tracker.Record
			if tc.rs.Tracker {
				in = cards
			}
			got := repoClaimsFrom(tracker.ComposeRecords(tc.rs, in, tc.unreadable, dup), "r", canon)
			if got.State != tc.state || !strings.Contains(got.Error, tc.errHas) || len(got.Cards) != tc.cards {
				t.Fatalf("got %+v", got)
			}
			if (got.Error == "") != (tc.state == ClaimsPresent || tc.state == ClaimsAbsent) {
				t.Fatalf("error %q for state %s", got.Error, tc.state)
			}
			if tc.cards > 0 {
				if c := got.Cards[0]; c.Ref != "r#000001" || c.Revision != "blob000001" || c.Claimant == nil || c.Claimant.Worktree != "/canon/w/raw" {
					t.Fatalf("claimed card: %+v", c)
				}
				if got.Cards[1].Claimant != nil {
					t.Fatalf("unclaimed card has a claimant: %+v", got.Cards[1])
				}
			}
		})
	}
}

// #288: marshalling never writes into the caller's rows.
func TestInventoryMarshalDoesNotMutate(t *testing.T) {
	row := validTreeRow()
	row.Claims, row.ClaimsState = nil, ""
	inv := Inventory{Rows: []TreeRow{row}}
	if _, err := json.Marshal(inv); err != nil {
		t.Fatal(err)
	}
	if inv.Rows[0].ClaimsState != "" || inv.Rows[0].Claims != nil {
		t.Fatalf("marshal mutated the caller's row: %+v", inv.Rows[0])
	}
}

// editJSON changes exactly one thing in a JSON document — so a contract
// rejection test isolates the invariant it names, with no unknown key that
// strict decoding would reject on its own.
func editJSON(t *testing.T, raw string, edit func(map[string]any)) string {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	var m map[string]any
	if err := dec.Decode(&m); err != nil {
		t.Fatal(err)
	}
	edit(m)
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}
