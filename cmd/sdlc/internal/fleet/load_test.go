package fleet

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestPolicyDeclarationPathUsesVocabularyAuthority(t *testing.T) {
	repoRoot := filepath.Join(t.TempDir(), "repo")
	want := filepath.Join(repoRoot, filepath.FromSlash(vocab.FleetPolicy().DeclarationPath))
	if got := PolicyDeclarationPath(repoRoot); got != want {
		t.Fatalf("PolicyDeclarationPath(%q) = %q, want %q", repoRoot, got, want)
	}
}

func TestLoadPolicySharedCorpus(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "..", "..", "construct", "vocabulary", "testdata", "fleet_policy_*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("shared fleet-policy corpus is empty")
	}

	for _, path := range paths {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			got := LoadPolicyFile(path)
			valid := strings.Contains(filepath.Base(path), "_valid_")
			if got.OK != valid {
				t.Fatalf("LoadPolicyFile(%q).OK = %v, want %v; diagnostic=%+v", path, got.OK, valid, got.Diagnostic)
			}
			if valid {
				if got.Value == nil {
					t.Fatal("successful capability has nil value")
				}
				if got.Value.PolicyVersion != 1 || got.Value.PolicyDigest == "" {
					t.Fatalf("successful capability lacks policy identity: %+v", got.Value)
				}
				if got.Value.Roots == nil {
					t.Fatal("successful capability has nil roots; JSON contract requires []")
				}
				return
			}
			if got.Diagnostic == nil || got.Diagnostic.Code != DiagnosticInvalidPolicy || got.Diagnostic.Message == "" {
				t.Fatalf("invalid fixture returned non-actionable diagnostic: %+v", got.Diagnostic)
			}
		})
	}
}

func TestLoadPolicyFileMissing(t *testing.T) {
	got := LoadPolicyFile(filepath.Join(t.TempDir(), "missing.json"))
	if got.OK || got.Diagnostic == nil || got.Diagnostic.Code != DiagnosticMissingPolicy {
		t.Fatalf("missing declaration = %+v, want missing-policy diagnostic", got)
	}
}

func TestDecodePolicyDigestIsSemantic(t *testing.T) {
	one := []byte(`{"version":1,"admission":{"key":{"kind":"repo","roots":[]},"capacity":{"kind":"bounded","limit":1},"onCapacity":"reject"}}`)
	reordered := []byte(`{
  "admission": {
    "onCapacity": "reject",
    "capacity": {"limit": 1, "kind": "bounded"},
    "key": {"roots": [], "kind": "repo"}
  },
  "version": 1
}`)
	changed := []byte(`{"version":1,"admission":{"key":{"kind":"repo","roots":[]},"capacity":{"kind":"bounded","limit":2},"onCapacity":"reject"}}`)
	rootsOne := []byte(`{"version":1,"admission":{"key":{"kind":"declared-root","roots":["competitions/*","benchmark-suites/*"]},"capacity":{"kind":"unbounded"}}}`)
	rootsReordered := []byte(`{"version":1,"admission":{"key":{"kind":"declared-root","roots":["benchmark-suites/*","competitions/*"]},"capacity":{"kind":"unbounded"}}}`)

	a := DecodePolicy("a.json", one)
	b := DecodePolicy("b.json", reordered)
	c := DecodePolicy("c.json", changed)
	d := DecodePolicy("d.json", rootsOne)
	e := DecodePolicy("e.json", rootsReordered)
	if !a.OK || !b.OK || !c.OK || !d.OK || !e.OK {
		t.Fatalf("valid policies failed: a=%+v b=%+v c=%+v d=%+v e=%+v", a, b, c, d, e)
	}
	if a.Value.PolicyDigest != b.Value.PolicyDigest {
		t.Fatalf("cosmetic JSON rewrite changed digest: %q != %q", a.Value.PolicyDigest, b.Value.PolicyDigest)
	}
	if a.Value.PolicyDigest == c.Value.PolicyDigest {
		t.Fatalf("semantic policy change preserved digest %q", a.Value.PolicyDigest)
	}
	if d.Value.PolicyDigest != e.Value.PolicyDigest {
		t.Fatalf("root-order permutation changed digest: %q != %q", d.Value.PolicyDigest, e.Value.PolicyDigest)
	}
}

func TestDecodePolicyDigestChangesWithEveryRepresentableSemanticField(t *testing.T) {
	policies := map[string]string{
		"repo bounded reject one":        `{"version":1,"admission":{"key":{"kind":"repo","roots":[]},"capacity":{"kind":"bounded","limit":1},"onCapacity":"reject"}}`,
		"worktree bounded reject one":    `{"version":1,"admission":{"key":{"kind":"worktree","roots":[]},"capacity":{"kind":"bounded","limit":1},"onCapacity":"reject"}}`,
		"repo bounded reject two":        `{"version":1,"admission":{"key":{"kind":"repo","roots":[]},"capacity":{"kind":"bounded","limit":2},"onCapacity":"reject"}}`,
		"repo bounded provision one":     `{"version":1,"admission":{"key":{"kind":"repo","roots":[]},"capacity":{"kind":"bounded","limit":1},"onCapacity":"provision-worktree"}}`,
		"repo unbounded":                 `{"version":1,"admission":{"key":{"kind":"repo","roots":[]},"capacity":{"kind":"unbounded"}}}`,
		"declared root competitions":     `{"version":1,"admission":{"key":{"kind":"declared-root","roots":["competitions/*"]},"capacity":{"kind":"unbounded"}}}`,
		"declared root benchmark suites": `{"version":1,"admission":{"key":{"kind":"declared-root","roots":["benchmark-suites/*"]},"capacity":{"kind":"unbounded"}}}`,
	}
	digests := make(map[string]string, len(policies))
	for name, raw := range policies {
		got := DecodePolicy(name+".json", []byte(raw))
		if !got.OK {
			t.Fatalf("valid semantic variant %q failed: %+v", name, got.Diagnostic)
		}
		if prior, exists := digests[got.Value.PolicyDigest]; exists {
			t.Fatalf("semantic variants %q and %q share digest %q", prior, name, got.Value.PolicyDigest)
		}
		digests[got.Value.PolicyDigest] = name
	}
}

func TestDecodePolicyRejectsNonCanonicalJSONLanguage(t *testing.T) {
	tests := map[string]string{
		"duplicate-key":  `{"version":1,"version":1,"admission":{"key":{"kind":"repo","roots":[]},"capacity":{"kind":"bounded","limit":1},"onCapacity":"reject"}}`,
		"unknown-field":  `{"version":1,"extra":true,"admission":{"key":{"kind":"repo","roots":[]},"capacity":{"kind":"bounded","limit":1},"onCapacity":"reject"}}`,
		"trailing-value": `{"version":1,"admission":{"key":{"kind":"repo","roots":[]},"capacity":{"kind":"bounded","limit":1},"onCapacity":"reject"}} {}`,
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			got := DecodePolicy(name+".json", []byte(raw))
			if got.OK || got.Diagnostic == nil || got.Diagnostic.Code != DiagnosticInvalidPolicy {
				t.Fatalf("DecodePolicy accepted %s: %+v", name, got)
			}
		})
	}
}

func TestDecodePolicyDiagnosticCarriesOnlyDecodedVersion(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want *int
	}{
		{name: "unknown field after version", raw: `{"version":1,"unknown":true}`, want: intPtr(1)},
		{name: "nested duplicate after version", raw: `{"version":1,"admission":{"key":{},"key":{}}}`, want: intPtr(1)},
		{name: "version after nested duplicate", raw: `{"admission":{"key":{},"key":{}},"version":1}`, want: intPtr(1)},
		{name: "truncated after version", raw: `{"version":1,"admission":`, want: intPtr(1)},
		{name: "missing", raw: `{"admission":{}}`},
		{name: "null", raw: `{"version":null}`},
		{name: "string", raw: `{"version":"1"}`},
		{name: "fractional", raw: `{"version":1.5}`},
		{name: "duplicate", raw: `{"version":1,"version":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecodePolicy(tt.name+".json", []byte(tt.raw))
			if got.Diagnostic == nil {
				t.Fatalf("invalid input returned no diagnostic: %+v", got)
			}
			if tt.want == nil {
				if got.Diagnostic.PolicyVersion != nil {
					t.Fatalf("fabricated version %d", *got.Diagnostic.PolicyVersion)
				}
				return
			}
			if got.Diagnostic.PolicyVersion == nil || *got.Diagnostic.PolicyVersion != *tt.want {
				t.Fatalf("diagnostic version = %v, want %d", got.Diagnostic.PolicyVersion, *tt.want)
			}
		})
	}
}

func intPtr(value int) *int { return &value }

func FuzzDecodePolicy(f *testing.F) {
	f.Add([]byte(`{"version":1}`))
	f.Add([]byte{0xff, 0x00, '{'})
	f.Fuzz(func(t *testing.T, raw []byte) {
		got := DecodePolicy("fuzz.json", raw)
		if got.OK && (got.Value == nil || got.Diagnostic != nil) {
			t.Fatalf("successful result is not a total success variant: %+v", got)
		}
		if !got.OK && (got.Diagnostic == nil || got.Value != nil) {
			t.Fatalf("failed result is not a total diagnostic variant: %+v", got)
		}
	})
}
