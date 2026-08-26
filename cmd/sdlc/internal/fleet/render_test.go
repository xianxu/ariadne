package fleet

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestRenderInventoryExactSemanticSnapshots(t *testing.T) {
	one, two, version := 1, 2, 1
	bounded := PolicyCapability{OK: true, Value: &PolicyCapabilityValue{
		PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{},
		Capacity: Capacity{Kind: "bounded", Limit: &two}, OnCapacity: "reject",
	}}
	unbounded := PolicyCapability{OK: true, Value: &PolicyCapabilityValue{
		PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{},
		Capacity: Capacity{Kind: "unbounded"},
	}}
	diagnostic := PolicyCapability{Diagnostic: &PolicyDiagnostic{
		Code: "invalid-policy", Message: "bad declaration", Path: "/repo/.sdlc/fleet.json", PolicyVersion: &version,
	}}
	available := MeasuredFacts{
		Available: true, Head: "head", CommitTimestamp: "2026-01-02T03:04:05Z", DirtyCount: &one,
		BaseAvailable: true, BaseRef: "main", Ahead: &two, Behind: &one,
	}

	branch := validTreeRow()
	branch.Policy = bounded
	branch.Facts = available
	branch.Locked = stringRef("maintenance")
	branch.Prunable = stringRef("gone")
	branch.Issues = []IssueAssociation{{Ref: "ariadne#000200", DeclaredStatus: "working", Provenance: IssueProvenanceBranchPrefix}}

	baseMissing := branch
	baseMissing.Locked, baseMissing.Prunable = nil, nil
	baseMissing.Issues = []IssueAssociation{}
	baseMissing.Policy = unbounded
	baseMissing.Facts.BaseAvailable = false
	baseMissing.Facts.BaseRef = ""
	baseMissing.Facts.Ahead, baseMissing.Facts.Behind = nil, nil
	baseMissing.Facts.BaseError = "no base reference available"

	detached := baseMissing
	detached.Branch, detached.Detached = "", true
	detached.Facts.BaseRef = "origin/main"
	detached.Facts.BaseError = "rev-list failed"

	bare := branch
	bare.Branch, bare.Bare = "", true
	bare.Locked, bare.Prunable = nil, nil
	bare.Issues = []IssueAssociation{}
	bare.Facts = MeasuredFacts{Error: "show failed", Head: "head"}
	bare.Policy = diagnostic

	staged := branch
	staged.Locked, staged.Prunable = nil, nil
	staged.Issues = []IssueAssociation{}
	staged.Facts = MeasuredFacts{Error: "status failed", Head: "head", CommitTimestamp: "2026-01-02T03:04:05Z", DirtyCount: &one}

	cases := []struct {
		name  string
		value Inventory
		want  string
	}{
		{
			name:  "branch bounded issues attributes diagnostics",
			value: Inventory{Rows: []TreeRow{branch}, Diagnostics: []RepoDiagnostic{{RepoIdentity: "/repo/.git", RepoPath: "/repo", TreePath: "/repo", Stage: "facts", Message: "later probe failed"}}},
			want: "tree=\"/repo\"\trepo_identity=\"/repo/.git\"\trepo_root=\"/repo\"\tbranch=\"main\"\n" +
				"  facts head=\"head\" commit_timestamp=\"2026-01-02T03:04:05Z\" dirty_count=1 base_ref=\"main\" ahead=2 behind=1\n" +
				"  locked=\"maintenance\"\n  prunable=\"gone\"\n" +
				"  issue=\"ariadne#000200\" status=\"working\" provenance=\"branch-prefix\"\n" +
				"  policy=capability version=1 digest=\"" + testPolicyDigest + "\" key_kind=\"repo\" roots= capacity=\"bounded\" limit=2 on_capacity=\"reject\"\n" +
				"diagnostic repo_path=\"/repo\" stage=\"facts\" message=\"later probe failed\" repo_identity=\"/repo/.git\" tree_path=\"/repo\"\n",
		},
		{
			name:  "branch base unavailable unbounded",
			value: Inventory{Rows: []TreeRow{baseMissing}, Diagnostics: []RepoDiagnostic{}},
			want: "tree=\"/repo\"\trepo_identity=\"/repo/.git\"\trepo_root=\"/repo\"\tbranch=\"main\"\n" +
				"  facts head=\"head\" commit_timestamp=\"2026-01-02T03:04:05Z\" dirty_count=1 base_unavailable error=\"no base reference available\"\n" +
				"  policy=capability version=1 digest=\"" + testPolicyDigest + "\" key_kind=\"repo\" roots= capacity=\"unbounded\"\n",
		},
		{
			name:  "detached selected base unavailable",
			value: Inventory{Rows: []TreeRow{detached}},
			want: "tree=\"/repo\"\trepo_identity=\"/repo/.git\"\trepo_root=\"/repo\"\tdetached\n" +
				"  facts head=\"head\" commit_timestamp=\"2026-01-02T03:04:05Z\" dirty_count=1 base_unavailable error=\"rev-list failed\" base_ref=\"origin/main\"\n" +
				"  policy=capability version=1 digest=\"" + testPolicyDigest + "\" key_kind=\"repo\" roots= capacity=\"unbounded\"\n",
		},
		{
			name:  "bare staged head capability diagnostic",
			value: Inventory{Rows: []TreeRow{bare}},
			want: "tree=\"/repo\"\trepo_identity=\"/repo/.git\"\trepo_root=\"/repo\"\tbare\n" +
				"  facts unavailable head=\"head\" error=\"show failed\"\n" +
				"  policy diagnostic code=\"invalid-policy\" message=\"bad declaration\" path=\"/repo/.sdlc/fleet.json\" policy_version=1\n",
		},
		{
			name:  "staged dirty prefix",
			value: Inventory{Rows: []TreeRow{staged}},
			want: "tree=\"/repo\"\trepo_identity=\"/repo/.git\"\trepo_root=\"/repo\"\tbranch=\"main\"\n" +
				"  facts unavailable head=\"head\" commit_timestamp=\"2026-01-02T03:04:05Z\" dirty_count=1 error=\"status failed\"\n" +
				"  policy=capability version=1 digest=\"" + testPolicyDigest + "\" key_kind=\"repo\" roots= capacity=\"bounded\" limit=2 on_capacity=\"reject\"\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before, err := json.Marshal(tc.value)
			if err != nil {
				t.Fatal(err)
			}
			var first, second bytes.Buffer
			if err := RenderInventory(&first, tc.value); err != nil {
				t.Fatal(err)
			}
			if err := RenderInventory(&second, tc.value); err != nil {
				t.Fatal(err)
			}
			if got := first.String(); got != tc.want {
				t.Fatalf("render=%q\nwant=%q", got, tc.want)
			}
			if second.String() != first.String() {
				t.Fatal("repeated render is not stable")
			}
			after, err := json.Marshal(tc.value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatal("render mutated input")
			}
			assertNoDerivedRenderLabels(t, first.String())
		})
	}
}

func assertNoDerivedRenderLabels(t *testing.T, rendered string) {
	t.Helper()
	for _, label := range renderFieldLabels(rendered) {
		switch strings.ToLower(label) {
		case "cold", "drift", "liveness", "staleness":
			t.Fatalf("render includes derived label %q: %s", label, rendered)
		}
	}
}

func renderFieldLabels(rendered string) []string {
	var labels []string
	for _, line := range strings.Split(rendered, "\n") {
		for start := 0; start < len(line); {
			for start < len(line) && (line[start] == ' ' || line[start] == '\t') {
				start++
			}
			end, quoted, escaped := start, false, false
			for end < len(line) {
				c := line[end]
				if escaped {
					escaped = false
				} else if quoted && c == '\\' {
					escaped = true
				} else if c == '"' {
					quoted = !quoted
				} else if !quoted && (c == ' ' || c == '\t') {
					break
				}
				end++
			}
			if eq := strings.IndexByte(line[start:end], '='); eq >= 0 {
				labels = append(labels, line[start:start+eq])
			}
			start = end + 1
		}
	}
	return labels
}

func TestRenderPolicyExactVariants(t *testing.T) {
	limit := 2
	version := 1
	cases := []struct {
		value PolicyResult
		want  string
	}{
		{PolicyResult{OK: true, Value: &PolicyResultValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, RepoIdentity: "/repo/.git", AdmissionKey: "/repo", Capacity: Capacity{Kind: "bounded", Limit: &limit}, OnCapacity: "reject"}}, "policy version=1 digest=\"" + testPolicyDigest + "\" repo_identity=\"/repo/.git\" admission_key=\"/repo\" capacity=\"bounded\" limit=2 on_capacity=\"reject\"\n"},
		{PolicyResult{OK: true, Value: &PolicyResultValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, RepoIdentity: "/repo/.git", AdmissionKey: "/repo", Capacity: Capacity{Kind: "unbounded"}}}, "policy version=1 digest=\"" + testPolicyDigest + "\" repo_identity=\"/repo/.git\" admission_key=\"/repo\" capacity=\"unbounded\"\n"},
		{PolicyResult{Diagnostic: &PolicyDiagnostic{Code: "invalid-policy", Message: "bad cold=storage", Path: "/p/drift", PolicyVersion: &version}}, "policy diagnostic code=\"invalid-policy\" message=\"bad cold=storage\" path=\"/p/drift\" policy_version=1\n"},
	}
	for _, tc := range cases {
		var out bytes.Buffer
		if err := RenderPolicy(&out, tc.value); err != nil {
			t.Fatal(err)
		}
		if out.String() != tc.want {
			t.Fatalf("got=%q want=%q", out.String(), tc.want)
		}
		assertNoDerivedRenderLabels(t, out.String())
	}
}

func TestRenderWriteFailure(t *testing.T) {
	errSentinel := errors.New("write failed")
	row := validTreeRow()
	if err := RenderInventory(failingRenderWriter{err: errSentinel}, Inventory{Rows: []TreeRow{row}}); !errors.Is(err, errSentinel) {
		t.Fatalf("err=%v", err)
	}
}

type failingRenderWriter struct{ err error }

func (w failingRenderWriter) Write([]byte) (int, error) { return 0, w.err }

func TestRenderRejectsInvalidAndRendersPolicy(t *testing.T) {
	var out bytes.Buffer
	if err := RenderInventory(&out, Inventory{Rows: []TreeRow{{}}}); err == nil {
		t.Fatal("invalid inventory rendered")
	}
	result := PolicyResult{Diagnostic: &PolicyDiagnostic{Code: "invalid-policy", Message: "bad", Path: "/repo/.sdlc/fleet.json"}}
	if err := RenderPolicy(&out, result); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); !strings.Contains(got, `policy diagnostic code="invalid-policy"`) || !strings.Contains(got, `path="/repo/.sdlc/fleet.json"`) {
		t.Fatalf("policy render = %q", got)
	}
}

func TestRenderQuotesFreeFormValuesAndKeepsLabelsStructural(t *testing.T) {
	row := validTreeRow()
	row.TreePath = "/repo/with space\tand\nnewline"
	row.RepoRoot = row.TreePath
	row.Branch = "feature/\"quoted\"\\branch"
	row.Locked = stringRef("cold=storage\x01\nlocked=false")
	row.Facts = MeasuredFacts{Error: "status\tfailed\n tree=forged", Head: "abc\\def"}
	row.Policy = PolicyCapability{Diagnostic: &PolicyDiagnostic{Code: "invalid-policy", Message: "bad\npolicy=forged", Path: "/p\tq"}}
	var out bytes.Buffer
	if err := RenderInventory(&out, Inventory{Rows: []TreeRow{row}}); err != nil {
		t.Fatal(err)
	}
	wantFragments := []string{
		`tree="/repo/with space\tand\nnewline"`,
		`branch="feature/\"quoted\"\\branch"`,
		`locked="cold=storage\x01\nlocked=false"`,
		`error="status\tfailed\n tree=forged"`,
		`message="bad\npolicy=forged"`,
		`path="/p\tq"`,
	}
	for _, want := range wantFragments {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("render missing %q: %q", want, out.String())
		}
	}
	assertNoDerivedRenderLabels(t, out.String())
}

func TestRenderRejectsDuplicateTreeIdentity(t *testing.T) {
	row := validTreeRow()
	if err := RenderInventory(&bytes.Buffer{}, Inventory{Rows: []TreeRow{row, row}}); err == nil || !strings.Contains(err.Error(), "duplicate tree identity") {
		t.Fatalf("err=%v", err)
	}
}

func TestRenderDiagnosticOrderingUsesFieldTuple(t *testing.T) {
	row := validTreeRow()
	diagnostics := []RepoDiagnostic{
		{RepoIdentity: "a", RepoPath: "b", Stage: "c\x00d", Message: "z"},
		{RepoIdentity: "a\x00b", RepoPath: "c", Stage: "d", Message: "a"},
	}
	first := Inventory{Rows: []TreeRow{row}, Diagnostics: diagnostics}
	second := Inventory{Rows: []TreeRow{row}, Diagnostics: []RepoDiagnostic{diagnostics[1], diagnostics[0]}}
	var a, b bytes.Buffer
	if err := RenderInventory(&a, first); err != nil {
		t.Fatal(err)
	}
	if err := RenderInventory(&b, second); err != nil {
		t.Fatal(err)
	}
	if a.String() != b.String() {
		t.Fatalf("diagnostic order depends on input:\n%s\n%s", a.String(), b.String())
	}
}

func stringRef(value string) *string { return &value }
