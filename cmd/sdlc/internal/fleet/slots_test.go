package fleet

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// #289: the per-checkout verdict over probe outcome × working-tree state ×
// branch state × claim. Needs-recovery is reported whenever its facts were
// read; otherwise a failed probe is unknown; otherwise unlanded commits, an
// open issue or a claim hold work; otherwise ready.
func TestJudgeCheckout(t *testing.T) {
	n := func(v int) *int { return &v }
	ok := MeasuredFacts{Available: true, Head: "h", BaseAvailable: true, BaseRef: "origin/main", Ahead: n(0), Behind: n(0), DirtyCount: n(0)}
	with := func(f func(*TreeRow)) TreeRow {
		r := TreeRow{Branch: "main-slot1", Facts: ok, Issues: []IssueAssociation{}, Claims: []ClaimAssociation{}, ClaimsState: ClaimsPresent}
		f(&r)
		return r
	}
	openIssue := []IssueAssociation{{Ref: "r#000007", DeclaredStatus: "working", Provenance: IssueProvenanceBranchPrefix}}
	doneIssue := []IssueAssociation{{Ref: "r#000007", DeclaredStatus: "done", Provenance: IssueProvenanceBranchPrefix}}
	for _, tc := range []struct {
		name    string
		row     TreeRow
		verdict Verdict
		reasons string // comma-joined reason codes, in order
	}{
		{"clean on resting", with(func(*TreeRow) {}), VerdictReady, ""},
		{"merged feature branch", with(func(r *TreeRow) { r.Branch = "topic" }), VerdictReady, ""},
		{"closed-issue branch", with(func(r *TreeRow) { r.Branch = "000007-x"; r.Issues = doneIssue }), VerdictReady, ""},
		{"unlanded commits", with(func(r *TreeRow) { r.Branch = "topic"; r.Facts.Ahead = n(2) }), VerdictHoldsWork, "unlanded-commits"},
		{"unlanded on resting", with(func(r *TreeRow) { r.Facts.Ahead = n(1) }), VerdictHoldsWork, "unlanded-commits"},
		{"zero-commit open-issue branch", with(func(r *TreeRow) { r.Branch = "000007-x"; r.Issues = openIssue }), VerdictHoldsWork, "open-issue:r#000007"},
		{"claimed on resting", with(func(r *TreeRow) {
			r.Claims = []ClaimAssociation{{Ref: "r#000009", Status: "working"}}
		}), VerdictHoldsWork, "claimed:r#000009"},
		{"dirty", with(func(r *TreeRow) { r.Facts.DirtyCount = n(3) }), VerdictNeedsRecovery, "dirty"},
		{"operation", with(func(r *TreeRow) { r.Facts.Operation = "rebase-merge" }), VerdictNeedsRecovery, "operation:rebase-merge"},
		{"detached", with(func(r *TreeRow) { r.Branch, r.Detached = "", true }), VerdictNeedsRecovery, "detached"},
		{"dirty and unlanded", with(func(r *TreeRow) { r.Facts.DirtyCount = n(1); r.Branch = "topic"; r.Facts.Ahead = n(1) }), VerdictNeedsRecovery, "dirty"},
		{"facts unavailable", with(func(r *TreeRow) { r.Facts = MeasuredFacts{Error: "status failed"} }), VerdictUnknown, "probe:facts"},
		{"base unavailable", with(func(r *TreeRow) {
			r.Facts.BaseAvailable, r.Facts.BaseError, r.Facts.Ahead, r.Facts.Behind = false, "no base", nil, nil
		}), VerdictUnknown, "probe:base"},
		{"base unavailable but dirty", with(func(r *TreeRow) {
			r.Facts.BaseAvailable, r.Facts.BaseError, r.Facts.Ahead, r.Facts.Behind = false, "no base", nil, nil
			r.Facts.DirtyCount = n(1)
		}), VerdictNeedsRecovery, "dirty,probe:base"},
		{"operation probe failed", with(func(r *TreeRow) { r.Facts.OperationError = "denied" }), VerdictUnknown, "probe:operation"},
		{"issue branch, lookup failed", with(func(r *TreeRow) { r.Branch, r.IssuesError = "000007-x", "lookup issue 000007: tracker unreadable" }), VerdictUnknown, "probe:issue"},
		{"issue branch, lookup ok, no match", with(func(r *TreeRow) { r.Branch = "000007-x" }), VerdictReady, ""},
		{"claims unread, otherwise ready", with(func(r *TreeRow) { r.ClaimsState, r.ClaimsError = ClaimsUnknown, "x" }), VerdictUnknown, "probe:claims"},
		{"claims unread, already holding work", with(func(r *TreeRow) {
			r.ClaimsState, r.ClaimsError = ClaimsUnknown, "x"
			r.Branch, r.Facts.Ahead = "topic", n(1)
		}), VerdictHoldsWork, "unlanded-commits"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := JudgeCheckout(tc.row, "main-slot1")
			if got.Verdict != tc.verdict || strings.Join(got.Reasons, ",") != tc.reasons {
				t.Fatalf("got %s %v, want %s %s", got.Verdict, got.Reasons, tc.verdict, tc.reasons)
			}
			if strings.Contains(tc.reasons, "probe:") && len(got.Errors) == 0 {
				t.Fatalf("a failed probe must carry its error: %+v", got)
			}
		})
	}
}

// #289: the verdict order the slot fold uses.
func TestVerdictOrder(t *testing.T) {
	order := []Verdict{VerdictReady, VerdictHoldsWork, VerdictUnknown, VerdictMissing, VerdictNeedsRecovery}
	for i := 1; i < len(order); i++ {
		if order[i-1].rank() >= order[i].rank() {
			t.Fatalf("%s must rank below %s", order[i-1], order[i])
		}
	}
	if Worst(VerdictHoldsWork, VerdictMissing, VerdictReady) != VerdictMissing {
		t.Fatal("worst-of")
	}
}

func slotRow(repoRoot, tree, branch string) TreeRow {
	row := validTreeRow()
	zero := 0
	row.RepoIdentity, row.RepoRoot, row.TreePath, row.Branch = repoRoot+"/.git", repoRoot, tree, branch
	row.Facts.Ahead, row.Facts.Behind, row.Facts.DirtyCount = &zero, &zero, &zero
	return row
}

// #289: slots come from row paths alone — a fleet primary is repo:0, the
// canonical slot path of the row's own repository is repo:N; look-alikes,
// feature worktrees and clones are not slots.
func TestDiscoverSlots(t *testing.T) {
	const f = "/f"
	rows := []TreeRow{
		slotRow(f+"/pair", f+"/pair", "main"),
		slotRow(f+"/pair", f+"/worktree/pair-slot2/pair", "main-slot2"),
		slotRow(f+"/pair", f+"/worktree/pair-slot1/pair", "main-slot1"),
		slotRow(f+"/pair", f+"/worktree/pair-slot01/pair", "x"),                             // non-canonical number
		slotRow(f+"/pair", f+"/worktree/ariadne-slot1/pair", "x"),                           // another repository's slot path
		slotRow(f+"/pair", "/tmp/feature", "000001-x"),                                      // feature worktree
		slotRow(f+"/worktree/pair-slot1/ariadne", f+"/worktree/pair-slot1/ariadne", "main"), // dependency clone
		slotRow(f+"/ariadne", f+"/ariadne", "main"),
	}
	var got []string
	for _, h := range discoverSlots(rows, f) {
		got = append(got, h.Address+"@"+h.HostPath+"#"+h.Resting+"|"+h.EnvRoot)
	}
	want := []string{
		"ariadne:0@/f/ariadne#main|",
		"pair:0@/f/pair#main|",
		"pair:1@/f/worktree/pair-slot1/pair#main-slot1|/f/worktree/pair-slot1",
		"pair:2@/f/worktree/pair-slot2/pair#main-slot2|/f/worktree/pair-slot2",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// #289: a slot folds its members to the worst verdict; members keep their own.
// A declared clone that is missing, outside the environment, or not a
// checkout is judged without a row; a declaration error lands on its declarer.
func TestAssembleSlots(t *testing.T) {
	const env = "/f/worktree/pair-slot1"
	host := SlotHost{Repo: "pair", Slot: 1, Address: "pair:1", Resting: "main-slot1", HostPath: env + "/pair", EnvRoot: env}
	zero := SlotHost{Repo: "pair", Address: "pair:0", Resting: "main", HostPath: "/f/pair"}
	dirty := slotRow(env+"/ariadne", env+"/ariadne", "main")
	three := 3
	dirty.Facts.DirtyCount = &three
	rows := []TreeRow{slotRow("/f/pair", env+"/pair", "main-slot1"), dirty, slotRow("/f/pair", "/f/pair", "main")}
	for _, tc := range []struct {
		name    string
		decls   []MemberDecl
		declErr map[string]string
		slot    Verdict
		members string
	}{
		{"clean dependency", []MemberDecl{{env + "/ariadne", MemberPresent}}, nil, VerdictNeedsRecovery, "host:ready,dependency:needs-recovery"},
		{"missing dependency", []MemberDecl{{env + "/gone", MemberMissing}}, nil, VerdictMissing, "host:ready,dependency:missing"},
		{"outside", []MemberDecl{{"/f/elsewhere", MemberOutside}}, nil, VerdictUnknown, "host:ready,dependency:unknown"},
		{"not a checkout", []MemberDecl{{env + "/plain-dir", MemberPresent}}, nil, VerdictUnknown, "host:ready,dependency:unknown"},
		{"unreadable declaration", nil, map[string]string{env + "/pair": "permission denied"}, VerdictUnknown, "host:unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			slots := AssembleSlots([]SlotHost{host, zero}, map[string]SlotDeclaration{host.HostPath: {Members: tc.decls, Errors: tc.declErr}}, rows)
			s := slots[0]
			var got []string
			for _, m := range s.Members {
				got = append(got, m.Role+":"+string(m.Verdict))
				if m.Verdict != VerdictReady && len(m.Reasons) == 0 {
					t.Fatalf("%s %s has no reasons", m.Role, m.Verdict)
				}
			}
			if s.Verdict != tc.slot || strings.Join(got, ",") != tc.members {
				t.Fatalf("slot %s %v, want %s %s", s.Verdict, got, tc.slot, tc.members)
			}
			if z := slots[1]; z.Address != "pair:0" || len(z.Members) != 1 || z.Verdict != VerdictReady {
				t.Fatalf(":0 is its host alone: %+v", z)
			}
		})
	}
}

func sampleSlotsInventory(t *testing.T) Inventory {
	t.Helper()
	const env = "/f/worktree/pair-slot1"
	dirty := slotRow(env+"/ariadne", env+"/ariadne", "main")
	two := 2
	dirty.Facts.DirtyCount = &two
	rows := []TreeRow{slotRow("/f/pair", env+"/pair", "main-slot1"), dirty}
	host := SlotHost{Repo: "pair", Slot: 1, Address: "pair:1", Resting: "main-slot1", HostPath: env + "/pair", EnvRoot: env}
	slots := AssembleSlots([]SlotHost{host}, map[string]SlotDeclaration{host.HostPath: {Members: []MemberDecl{{env + "/ariadne", MemberPresent}}}}, rows)
	return Inventory{Rows: rows, Machine: me, Slots: slots}
}

// #289: the slots contract round-trips and each invariant has a rejection.
func TestSlotsContract(t *testing.T) {
	raw, err := json.Marshal(sampleSlotsInventory(t))
	if err != nil {
		t.Fatal(err)
	}
	var back Inventory
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("%v\n%s", err, raw)
	}
	if len(back.Slots) != 1 || back.Slots[0].Verdict != VerdictNeedsRecovery || back.Slots[0].Members[1].Reasons[0] != "dirty" {
		t.Fatalf("round trip: %s", raw)
	}
	for name, mutate := range map[string]func(string) string{
		"other version":   func(s string) string { return strings.Replace(s, `"schema_version":1`, `"schema_version":2`, 1) },
		"no version":      func(s string) string { return strings.Replace(s, `"schema_version":1,`, ``, 1) },
		"null slots":      func(s string) string { return strings.Replace(s, `"slots":[{`, `"slots":null,"z":[{`, 1) },
		"unknown verdict": func(s string) string { return strings.Replace(s, `"verdict":"ready"`, `"verdict":"fine"`, 1) },
		"slot not the worst": func(s string) string {
			return strings.Replace(s, `"verdict":"needs-recovery","members"`, `"verdict":"ready","members"`, 1)
		},
		"non-ready, no reasons": func(s string) string { return strings.Replace(s, `"reasons":["dirty"]`, `"reasons":[]`, 1) },
		"host not first":        func(s string) string { return strings.Replace(s, `"role":"host"`, `"role":"dependency"`, 1) },
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

// #289: the human view names each slot, its verdict and each member's reasons.
func TestRenderSlots(t *testing.T) {
	var b bytes.Buffer
	if err := RenderInventory(&b, sampleSlotsInventory(t)); err != nil {
		t.Fatal(err)
	}
	want := "slot=\"pair:1\" verdict=needs-recovery resting_branch=\"main-slot1\" environment=\"/f/worktree/pair-slot1\"\n" +
		"  member=host verdict=ready path=\"/f/worktree/pair-slot1/pair\" branch=\"main-slot1\"\n" +
		"  member=dependency verdict=needs-recovery path=\"/f/worktree/pair-slot1/ariadne\" branch=\"main\" reasons=\"dirty\"\n"
	if !strings.Contains(b.String(), want) {
		t.Fatalf("render lacks the slot block:\n%s", b.String())
	}
}
