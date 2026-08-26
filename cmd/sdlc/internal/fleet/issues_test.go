package fleet

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

func TestAssociateBranchIssue(t *testing.T) {
	one := func(id string) ([]IssueRecord, error) {
		return []IssueRecord{{Ref: "#" + id, DeclaredStatus: "working"}}, nil
	}
	lookupErr := errors.New("read same-repo issues")
	want := []IssueAssociation{{Ref: "#000149", DeclaredStatus: "working", Provenance: "branch-prefix"}}

	tests := []struct {
		name    string
		branch  string
		lookup  IssueLookup
		want    []IssueAssociation
		wantErr error
	}{
		{name: "valid", branch: "000149-opaque-tags", lookup: one, want: want},
		{name: "empty slug follows issue filename grammar", branch: "000149-", lookup: one, want: want},
		{name: "slash-prefixed lookalike", branch: "feature/000149-opaque-tags", lookup: one, want: []IssueAssociation{}},
		{name: "slash after anchored prefix", branch: "000149-opaque/tags", lookup: one, want: want},
		{name: "main", branch: "main", lookup: one, want: []IssueAssociation{}},
		{name: "detached", branch: "(detached)", lookup: one, want: []IssueAssociation{}},
		{name: "empty", branch: "", lookup: one, want: []IssueAssociation{}},
		{name: "short prefix", branch: "00149-opaque-tags", lookup: one, want: []IssueAssociation{}},
		{name: "overlong prefix", branch: "0000149-opaque-tags", lookup: one, want: []IssueAssociation{}},
		{name: "numeric bleed", branch: "0001490-opaque-tags", lookup: one, want: []IssueAssociation{}},
		{name: "malformed prefix", branch: "00014x-opaque-tags", lookup: one, want: []IssueAssociation{}},
		{name: "missing same-repo issue", branch: "000149-opaque-tags", lookup: func(string) ([]IssueRecord, error) { return []IssueRecord{}, nil }, want: []IssueAssociation{}},
		{name: "ambiguous same-repo issue", branch: "000149-opaque-tags", lookup: func(string) ([]IssueRecord, error) {
			return []IssueRecord{{Ref: "#149", DeclaredStatus: "open"}, {Ref: "#149-copy", DeclaredStatus: "working"}}, nil
		}, want: []IssueAssociation{}},
		{name: "lookup failure preserves identity and drops partial record", branch: "000149-opaque-tags", lookup: func(string) ([]IssueRecord, error) {
			return []IssueRecord{{Ref: "#149", DeclaredStatus: "working"}}, lookupErr
		}, want: []IssueAssociation{}, wantErr: lookupErr},
		{name: "nil lookup", branch: "000149-opaque-tags", want: []IssueAssociation{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AssociateBranchIssue(tt.branch, tt.lookup)
			if got == nil {
				t.Fatal("AssociateBranchIssue returned nil; JSON collection must remain []")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("AssociateBranchIssue(%q) error = %v, want errors.Is(_, %v)", tt.branch, err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("AssociateBranchIssue(%q) = %#v, want %#v", tt.branch, got, tt.want)
			}
		})
	}
}

func TestAssociateBranchIssueProperty(t *testing.T) {
	property := func(rawID [6]byte, rawSuffix string) bool {
		id := decimalIssueID(rawID)
		validBranch := id + "-" + safeBranchSuffix(rawSuffix)

		// Exercise the full lookup-cardinality partition for every generated valid
		// branch. Exactly one same-repo record is the only associating case.
		for cardinality := 0; cardinality <= 3; cardinality++ {
			got, err, calls, lookupID := associateWithCardinality(validBranch, cardinality)
			if err != nil || calls != 1 || lookupID != id || got == nil {
				return false
			}
			if cardinality == 1 {
				want := []IssueAssociation{{
					Ref:            "same-repo#" + id,
					DeclaredStatus: "open",
					Provenance:     "branch-prefix",
				}}
				if !reflect.DeepEqual(got, want) {
					return false
				}
			} else if len(got) != 0 {
				return false
			}
		}

		// Mutate each generated success shape across the grammar boundaries that
		// previously admitted basename lookalikes or numeric-prefix bleed. Invalid
		// branches must not consult the same-repo lookup at any cardinality.
		mutations := []string{
			"topic/" + validBranch,
			id + "0-" + safeBranchSuffix(rawSuffix),
			id[:5] + "-" + safeBranchSuffix(rawSuffix),
			id[:3] + "x" + id[4:] + "-" + safeBranchSuffix(rawSuffix),
		}
		for _, branch := range mutations {
			for cardinality := 0; cardinality <= 3; cardinality++ {
				got, err, calls, _ := associateWithCardinality(branch, cardinality)
				if err != nil || got == nil || len(got) != 0 || calls != 0 {
					return false
				}
			}
		}
		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 2_000}); err != nil {
		t.Fatal(err)
	}
}

func associateWithCardinality(branch string, cardinality int) (got []IssueAssociation, err error, calls int, lookupID string) {
	got, err = AssociateBranchIssue(branch, func(id string) ([]IssueRecord, error) {
		calls++
		lookupID = id
		records := make([]IssueRecord, cardinality)
		for i := range records {
			records[i] = IssueRecord{Ref: "same-repo#" + id, DeclaredStatus: "open"}
		}
		return records, nil
	})
	return got, err, calls, lookupID
}

func decimalIssueID(raw [6]byte) string {
	id := make([]byte, len(raw))
	for i, b := range raw {
		id[i] = '0' + b%10
	}
	return string(id)
}

func safeBranchSuffix(raw string) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789-_"
	var suffix strings.Builder
	for _, b := range []byte(raw) {
		suffix.WriteByte(alphabet[int(b)%len(alphabet)])
	}
	return suffix.String()
}
