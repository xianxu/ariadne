package fleet

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestJSONContract_InventoryUsesNonNullCollectionsAndSnakeCase(t *testing.T) {
	raw, err := json.Marshal(Inventory{})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"rows":[],"diagnostics":[]}` {
		t.Fatalf("empty Inventory JSON = %s, want non-null collections", raw)
	}

	row := validTreeRow()
	raw, err = json.Marshal(Inventory{Rows: []TreeRow{row}})
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{`"repo_identity"`, `"tree_path"`, `"commit_timestamp"`, `"dirty_count"`, `"policy_version"`, `"policy_digest"`, `"issues":[]`, `"roots":[]`} {
		if !strings.Contains(string(raw), required) {
			t.Errorf("inventory JSON %s lacks %s", raw, required)
		}
	}
	for _, forbidden := range []string{`"admission_key"`, `"cold"`, `"drift"`, `"liveness"`, `"repoIdentity"`} {
		if strings.Contains(string(raw), forbidden) {
			t.Errorf("inventory JSON %s contains forbidden %s", raw, forbidden)
		}
	}
}

func TestJSONContract_RejectsImpossiblePolicyVariants(t *testing.T) {
	limit := 1
	tests := []struct {
		name  string
		value any
	}{
		{"capability bounded missing action", PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{}, Capacity: Capacity{Kind: "bounded", Limit: &limit}}}},
		{"capability unbounded leaks action", PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{}, Capacity: Capacity{Kind: "unbounded"}, OnCapacity: "reject"}}},
		{"result missing identity", PolicyResult{OK: true, Value: &PolicyResultValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, Capacity: Capacity{Kind: "unbounded"}}}},
		{"diagnostic missing message", PolicyCapability{Diagnostic: &PolicyDiagnostic{Code: "invalid-policy"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := json.Marshal(tt.value); err == nil {
				t.Fatal("Marshal succeeded, want invariant error")
			}
		})
	}
}

func TestJSONContract_StrictlyRejectsNullUnknownAndLeakedFields(t *testing.T) {
	for _, raw := range []string{
		`{"ok":true,"value":{"policy_version":1,"policy_digest":"` + testPolicyDigest + `","key_kind":"repo","roots":null,"capacity":{"kind":"unbounded"}},"diagnostic":null}`,
		`{"ok":true,"value":{"policy_version":1,"policy_digest":"` + testPolicyDigest + `","key_kind":"repo","roots":[],"capacity":{"kind":"unbounded","limit":1}},"diagnostic":null}`,
		`{"ok":true,"value":{"policy_version":1,"policy_digest":"` + testPolicyDigest + `","key_kind":"repo","roots":[],"capacity":{"kind":"unbounded"},"head":"duplicate"},"diagnostic":null}`,
		`{"ok":true,"value":{"policy_version":1,"policy_digest":"` + testPolicyDigest + `","repo_identity":"/repo/.git","admission_key":"/repo","capacity":{"kind":"unbounded"}},"diagnostic":null,"unknown":true}`,
	} {
		var capability PolicyCapability
		var result PolicyResult
		if err := json.Unmarshal([]byte(raw), &capability); err == nil {
			t.Errorf("PolicyCapability accepted %s", raw)
		}
		if err := json.Unmarshal([]byte(raw), &result); err == nil {
			t.Errorf("PolicyResult accepted %s", raw)
		}
	}
}

func TestJSONContract_MeasuredAvailabilityDistinguishesAbsentValues(t *testing.T) {
	row := validTreeRow()
	one := 1
	row.Facts = MeasuredFacts{Available: false, Error: "git status failed"}
	if _, err := json.Marshal(row); err != nil {
		t.Fatalf("unavailable facts with explicit error rejected: %v", err)
	}
	row.Facts = MeasuredFacts{Available: false}
	if _, err := json.Marshal(row); err == nil {
		t.Fatal("unavailable facts without error marshaled")
	}
	row.Facts = MeasuredFacts{Available: false, Error: "git failed", Head: "head", CommitTimestamp: "2026-01-02T03:04:05Z", DirtyCount: &one}
	if _, err := json.Marshal(row); err != nil {
		t.Fatalf("staged partial facts rejected: %v", err)
	}
	row.Facts = MeasuredFacts{Available: false, Error: "git failed", Head: "head", DirtyCount: &one}
	if _, err := json.Marshal(row); err == nil {
		t.Fatal("out-of-order partial facts marshaled")
	}
	row.Facts = MeasuredFacts{Available: false, Error: "git failed", BaseError: "must not exist"}
	if _, err := json.Marshal(row); err == nil {
		t.Fatal("unavailable facts with base state marshaled")
	}
}

func TestJSONContract_RejectsUnknownBoundedAction(t *testing.T) {
	limit := 1
	value := PolicyCapability{OK: true, Value: &PolicyCapabilityValue{
		PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{},
		Capacity: Capacity{Kind: "bounded", Limit: &limit}, OnCapacity: "invented-action",
	}}
	if _, err := json.Marshal(value); err == nil {
		t.Fatal("unknown bounded on_capacity marshaled")
	}
}

func TestJSONContract_PolicySuccessUsesVocabularyAndDigestContract(t *testing.T) {
	limit := 1
	valid := PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{}, Capacity: Capacity{Kind: "bounded", Limit: &limit}, OnCapacity: "reject"}}
	cases := []struct {
		name  string
		value PolicyCapability
		valid bool
	}{
		{"valid repo", valid, true},
		{"unsupported version", PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: 2, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{}, Capacity: Capacity{Kind: "unbounded"}}}, false},
		{"invalid digest", PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: "DIGEST", KeyKind: "repo", Roots: []string{}, Capacity: Capacity{Kind: "unbounded"}}}, false},
		{"unknown key kind", PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "unknown", Roots: []string{}, Capacity: Capacity{Kind: "unbounded"}}}, false},
		{"repo roots", PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{"apps/*"}, Capacity: Capacity{Kind: "unbounded"}}}, false},
		{"declared roots", PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "declared-root", Roots: []string{"apps/*"}, Capacity: Capacity{Kind: "unbounded"}}}, true},
		{"bad declared roots", PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "declared-root", Roots: []string{"apps"}, Capacity: Capacity{Kind: "unbounded"}}}, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := json.Marshal(tt.value)
			if (err == nil) != tt.valid {
				t.Fatalf("MarshalJSON error = %v, valid=%v", err, tt.valid)
			}
		})
	}
	result := PolicyResult{OK: true, Value: &PolicyResultValue{PolicyVersion: 2, PolicyDigest: testPolicyDigest, RepoIdentity: "/repo/.git", AdmissionKey: "/repo", Capacity: Capacity{Kind: "unbounded"}}}
	if _, err := json.Marshal(result); err == nil {
		t.Fatal("result with unsupported version marshaled")
	}
}

func TestJSONContract_RejectsDuplicateKeysAtEveryDepth(t *testing.T) {
	validCapability := `{"ok":true,"value":{"policy_version":1,"policy_digest":"` + testPolicyDigest + `","key_kind":"repo","roots":[],"capacity":{"kind":"unbounded"}}}`
	validInventory := `{"rows":[{"repo_identity":"/repo/.git","repo_root":"/repo","tree_path":"/repo","branch":"main","detached":false,"bare":false,"facts":{"available":false,"error":"git failed","base_available":false},"issues":[],"policy":` + validCapability + `}],"diagnostics":[]}`
	for _, raw := range []string{
		`{"ok":true,"ok":true,"value":null,"diagnostic":{"code":"bad","message":"bad"}}`,
		`{"ok":true,"value":{"policy_version":1,"policy_digest":"` + testPolicyDigest + `","policy_digest":"` + testPolicyDigest + `","key_kind":"repo","roots":[],"capacity":{"kind":"unbounded"}}}`,
		`{"rows":[],"rows":[],"diagnostics":[]}`,
		strings.Replace(validInventory, `"available":false`, `"available":false,"available":false`, 1),
	} {
		var capability PolicyCapability
		var inventory Inventory
		if err := json.Unmarshal([]byte(raw), &capability); err == nil {
			t.Errorf("PolicyCapability accepted duplicate JSON %s", raw)
		}
		if err := json.Unmarshal([]byte(raw), &inventory); err == nil {
			t.Errorf("Inventory accepted duplicate JSON %s", raw)
		}
	}
}

func TestJSONContract_DeeplyValidatesRowsAndDiagnostics(t *testing.T) {
	valid := validTreeRow()
	valid.Issues = []IssueAssociation{{Ref: "ariadne#000200", DeclaredStatus: "working", Provenance: IssueProvenanceBranchPrefix}}
	valid.Facts.BaseAvailable = false
	valid.Facts.BaseRef = "main"
	valid.Facts.BaseError = "rev-list failed"
	valid.Facts.Ahead, valid.Facts.Behind = nil, nil
	if _, err := json.Marshal(Inventory{Rows: []TreeRow{valid}, Diagnostics: []RepoDiagnostic{{RepoIdentity: valid.RepoIdentity, RepoPath: valid.RepoRoot, TreePath: valid.TreePath, Stage: "facts", Message: "failed"}}}); err != nil {
		t.Fatalf("valid deep inventory rejected: %v", err)
	}
	invalidRows := []TreeRow{
		func() TreeRow { row := validTreeRow(); row.Detached = true; return row }(),
		func() TreeRow { row := validTreeRow(); row.Branch = ""; return row }(),
		func() TreeRow {
			row := validTreeRow()
			row.Issues = []IssueAssociation{{Ref: "ariadne#200", DeclaredStatus: "working", Provenance: IssueProvenanceBranchPrefix}}
			return row
		}(),
		func() TreeRow {
			row := validTreeRow()
			row.Issues = []IssueAssociation{{Ref: "ariadne#000200", DeclaredStatus: "nope", Provenance: IssueProvenanceBranchPrefix}}
			return row
		}(),
		func() TreeRow { row := validTreeRow(); row.Facts.DirtyCount = intRef(-1); return row }(),
	}
	for _, row := range invalidRows {
		if _, err := json.Marshal(row); err == nil {
			t.Fatalf("invalid row marshaled: %#v", row)
		}
	}
	if _, err := json.Marshal(Inventory{Diagnostics: []RepoDiagnostic{{RepoPath: "/repo", Stage: "facts", Message: "bad", TreePath: "/repo"}}}); err == nil {
		t.Fatal("tree-scoped diagnostic without repo identity marshaled")
	}
}

func intRef(value int) *int { return &value }

func TestJSONContract_CanonicalGoldens(t *testing.T) {
	limit := 1
	unboundedCapability := PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{}, Capacity: Capacity{Kind: "unbounded"}}}
	boundedCapability := PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{}, Capacity: Capacity{Kind: "bounded", Limit: &limit}, OnCapacity: "reject"}}
	diagnosticCapability := PolicyCapability{Diagnostic: &PolicyDiagnostic{Code: "invalid-policy", Message: "bad"}}
	unboundedResult := PolicyResult{OK: true, Value: &PolicyResultValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, RepoIdentity: "/repo/.git", AdmissionKey: "/repo", Capacity: Capacity{Kind: "unbounded"}}}
	boundedResult := PolicyResult{OK: true, Value: &PolicyResultValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, RepoIdentity: "/repo/.git", AdmissionKey: "/repo", Capacity: Capacity{Kind: "bounded", Limit: &limit}, OnCapacity: "reject"}}
	diagnosticResult := PolicyResult{Diagnostic: &PolicyDiagnostic{Code: "invalid-policy", Message: "bad"}}
	available := validTreeRow()
	unavailable := available
	unavailable.Facts = MeasuredFacts{Error: "git failed"}
	unavailable.Policy = diagnosticCapability

	for _, tt := range []struct {
		name  string
		value any
		want  string
	}{
		{"capability bounded", boundedCapability, `{"ok":true,"value":{"policy_version":1,"policy_digest":"` + testPolicyDigest + `","key_kind":"repo","roots":[],"capacity":{"kind":"bounded","limit":1},"on_capacity":"reject"}}`},
		{"capability unbounded", unboundedCapability, `{"ok":true,"value":{"policy_version":1,"policy_digest":"` + testPolicyDigest + `","key_kind":"repo","roots":[],"capacity":{"kind":"unbounded"}}}`},
		{"capability diagnostic", diagnosticCapability, `{"ok":false,"diagnostic":{"code":"invalid-policy","message":"bad"}}`},
		{"result bounded", boundedResult, `{"ok":true,"value":{"policy_version":1,"policy_digest":"` + testPolicyDigest + `","repo_identity":"/repo/.git","admission_key":"/repo","capacity":{"kind":"bounded","limit":1},"on_capacity":"reject"}}`},
		{"result unbounded", unboundedResult, `{"ok":true,"value":{"policy_version":1,"policy_digest":"` + testPolicyDigest + `","repo_identity":"/repo/.git","admission_key":"/repo","capacity":{"kind":"unbounded"}}}`},
		{"result diagnostic", diagnosticResult, `{"ok":false,"diagnostic":{"code":"invalid-policy","message":"bad"}}`},
		{"inventory available", Inventory{Rows: []TreeRow{available}, Diagnostics: []RepoDiagnostic{}}, `{"rows":[{"repo_identity":"/repo/.git","repo_root":"/repo","tree_path":"/repo","branch":"main","detached":false,"bare":false,"facts":{"available":true,"head":"head","commit_timestamp":"2026-01-02T03:04:05Z","base_available":true,"base_ref":"main","ahead":1,"behind":1,"dirty_count":1},"issues":[],"policy":{"ok":true,"value":{"policy_version":1,"policy_digest":"` + testPolicyDigest + `","key_kind":"repo","roots":[],"capacity":{"kind":"bounded","limit":1},"on_capacity":"reject"}}}],"diagnostics":[]}`},
		{"inventory unavailable", Inventory{Rows: []TreeRow{unavailable}, Diagnostics: []RepoDiagnostic{}}, `{"rows":[{"repo_identity":"/repo/.git","repo_root":"/repo","tree_path":"/repo","branch":"main","detached":false,"bare":false,"facts":{"available":false,"error":"git failed","base_available":false},"issues":[],"policy":{"ok":false,"diagnostic":{"code":"invalid-policy","message":"bad"}}}],"diagnostics":[]}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := json.Marshal(tt.value)
			if err != nil || string(raw) != tt.want {
				t.Fatalf("json.Marshal() = (%s, %v), want %s", raw, err, tt.want)
			}
		})
	}
}

func TestJSONContract_InventoryRejectsNullAndUnknownNestedFields(t *testing.T) {
	for _, raw := range []string{
		`{"rows":null,"diagnostics":[]}`,
		`{"rows":[],"diagnostics":null}`,
		`{"rows":[{"repo_identity":"/repo/.git","repo_root":"/repo","tree_path":"/repo","detached":false,"bare":false,"facts":{"available":false},"issues":[],"policy":{"ok":false,"diagnostic":{"code":"invalid","message":"bad"}}}],"diagnostics":[]}`,
	} {
		var inventory Inventory
		if err := json.Unmarshal([]byte(raw), &inventory); err == nil {
			t.Errorf("Inventory accepted %s", raw)
		}
	}
}

func FuzzJSONContractRoundTrip(f *testing.F) {
	f.Add(byte(0))
	f.Add(byte(255))
	f.Fuzz(func(t *testing.T, mode byte) {
		keyKind, roots := "repo", []string{}
		switch mode & 3 {
		case 1:
			keyKind = "worktree"
		case 2:
			keyKind, roots = "declared-root", []string{"apps/*"}
		case 3:
			keyKind = "unknown"
		}
		if mode&4 != 0 {
			roots = []string{"apps"}
		}
		digest := testPolicyDigest
		if mode&8 != 0 {
			digest = "not-a-digest"
		}
		version := 1
		if mode&16 != 0 {
			version = 2
		}
		valid := version == 1 && digest == testPolicyDigest && keyKind != "unknown" &&
			((keyKind == "declared-root" && len(roots) == 1 && roots[0] == "apps/*") ||
				((keyKind == "repo" || keyKind == "worktree") && len(roots) == 0))
		value := PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: version, PolicyDigest: digest, KeyKind: keyKind, Roots: roots, Capacity: Capacity{Kind: "unbounded"}}}
		raw, err := json.Marshal(value)
		if !valid {
			if err == nil {
				t.Fatalf("invalid policy marshaled: %#v", value)
			}
			return
		}
		if err != nil {
			t.Fatalf("valid policy rejected: %v", err)
		}
		var roundTrip PolicyCapability
		if err := json.Unmarshal(raw, &roundTrip); err != nil {
			t.Fatalf("round trip %s: %v", raw, err)
		}
		if !reflect.DeepEqual(roundTrip, value) {
			t.Fatalf("round trip = %#v", roundTrip)
		}

		result := PolicyResult{OK: true, Value: &PolicyResultValue{PolicyVersion: version, PolicyDigest: digest, RepoIdentity: "/repo/.git", AdmissionKey: "/repo", Capacity: Capacity{Kind: "unbounded"}}}
		raw, err = json.Marshal(result)
		resultValid := version == 1 && digest == testPolicyDigest
		if !resultValid {
			if err == nil {
				t.Fatalf("invalid policy result marshaled: %#v", result)
			}
			return
		}
		if err != nil {
			t.Fatalf("valid policy result rejected: %v", err)
		}
		var resultRoundTrip PolicyResult
		if err := json.Unmarshal(raw, &resultRoundTrip); err != nil {
			t.Fatalf("policy result round trip %s: %v", raw, err)
		}
		if !reflect.DeepEqual(resultRoundTrip, result) {
			t.Fatalf("policy result round trip = %#v", resultRoundTrip)
		}
	})
}

func FuzzJSONContractAlgebra(f *testing.F) {
	f.Add([]byte{0})
	f.Add([]byte{7, 255, 3})
	f.Fuzz(func(t *testing.T, seed []byte) {
		var mode byte
		if len(seed) > 0 {
			mode = seed[0]
		}
		limit := int(mode%9) + 1
		capacity := Capacity{Kind: "unbounded"}
		action := ""
		if mode&1 != 0 {
			capacity = Capacity{Kind: "bounded", Limit: &limit}
			action = "reject"
		}

		capability := PolicyCapability{OK: true, Value: &PolicyCapabilityValue{
			PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{}, Capacity: capacity, OnCapacity: action,
		}}
		result := PolicyResult{OK: true, Value: &PolicyResultValue{
			PolicyVersion: 1, PolicyDigest: testPolicyDigest, RepoIdentity: "/repo/.git", AdmissionKey: "/repo", Capacity: capacity, OnCapacity: action,
		}}
		if mode&2 != 0 {
			diagnostic := &PolicyDiagnostic{Code: "invalid-policy", Message: "bad"}
			capability = PolicyCapability{Diagnostic: diagnostic}
			result = PolicyResult{Diagnostic: diagnostic}
		}
		assertJSONRoundTrip(t, capability, new(PolicyCapability))
		assertJSONRoundTrip(t, result, new(PolicyResult))

		row := validTreeRow()
		row.Policy = capability
		if mode&4 != 0 {
			row.Facts = MeasuredFacts{Error: "git failed"}
		}
		inventory := Inventory{Rows: []TreeRow{row}, Diagnostics: []RepoDiagnostic{}}
		assertJSONRoundTrip(t, inventory, new(Inventory))
	})
}

func assertJSONRoundTrip[T any](t *testing.T, value T, decoded *T) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %T: %v", value, err)
	}
	if err := json.Unmarshal(raw, decoded); err != nil {
		t.Fatalf("unmarshal %T %s: %v", value, raw, err)
	}
	if !reflect.DeepEqual(value, *decoded) {
		t.Fatalf("round trip %T = %#v, want %#v", value, *decoded, value)
	}
	unknown := append(append([]byte{}, raw[:len(raw)-1]...), []byte(`,"unknown":true}`)...)
	if err := json.Unmarshal(unknown, decoded); err == nil {
		t.Fatalf("strict decoder accepted unknown field in %s", unknown)
	}
}

func validTreeRow() TreeRow {
	limit := 1
	return TreeRow{
		RepoIdentity: "/repo/.git", RepoRoot: "/repo", TreePath: "/repo", Branch: "main",
		Facts:  MeasuredFacts{Available: true, Head: "head", CommitTimestamp: "2026-01-02T03:04:05Z", BaseAvailable: true, BaseRef: "main", Ahead: &limit, Behind: &limit, DirtyCount: &limit},
		Issues: []IssueAssociation{},
		Policy: PolicyCapability{OK: true, Value: &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{}, Capacity: Capacity{Kind: "bounded", Limit: &limit}, OnCapacity: "reject"}},
	}
}
