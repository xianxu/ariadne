package gatestate

import (
	"reflect"
	"strings"
	"testing"
)

func TestRenderParseRoundTrip(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Critical/seam in wrong layer", "Minor/naming")),
		round(2, dispose("PQ-1", "addressed"), findings("Important/lock contract unstated")),
	)
	l.Rounds[0].Blocked = true
	l.Rounds[1].Blocked = true
	l.Rounds[1].New[0].Detail = "who owns the lock across the retry?"
	l.ContentHash = ContentHash("issue", "plan")

	got, err := ParseSidecar(Render(l, "ariadne"))
	if err != nil {
		t.Fatalf("ParseSidecar: %v", err)
	}
	if !reflect.DeepEqual(got, l) {
		t.Errorf("round-trip lost data:\n got %+v\nwant %+v", got, l)
	}
}

// The human prose must actually carry the findings — a reader opening the sidecar should
// not have to read YAML to see what the gate said.
func TestRenderProseCarriesFindings(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Critical/seam in wrong layer")))
	l.Rounds[0].Blocked = true
	out := Render(l, "ariadne")
	for _, want := range []string{
		"# Gate ledger — ariadne#187 (plan-quality)",
		"## Round 1", "BLOCKED", "PQ-1", "Critical", "seam in wrong layer",
		"## Open findings",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered sidecar missing %q", want)
		}
	}
}

// A protocol-error round and a forced round must both be visible to a human reader — they
// are the audit trail for "the gate ran but produced nothing" and "we went anyway".
func TestRenderProseShowsProtocolErrorAndForce(t *testing.T) {
	l := ledgerWith(round(1, nil, nil))
	l.Rounds[0].ProtocolError = "no valid findings block"
	l.Rounds[0].Forced = "shipping the hotfix"
	l.Rounds[0].Blocked = true
	out := Render(l, "ariadne")
	if !strings.Contains(out, "Protocol error") || !strings.Contains(out, "no valid findings block") {
		t.Error("prose must disclose a protocol-error round")
	}
	if !strings.Contains(out, "Forced past") || !strings.Contains(out, "shipping the hotfix") {
		t.Error("prose must disclose a forced round with its rationale")
	}
}

func TestRenderDisposedSection(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Critical/seam")),
		round(2, dispose("PQ-1", "addressed"), nil),
	)
	l.Rounds[1].Dispositions[0].Note = "moved to the filter"
	out := Render(l, "ariadne")
	if !strings.Contains(out, "### Disposed") || !strings.Contains(out, "PQ-1 — addressed — moved to the filter") {
		t.Errorf("disposed section missing or malformed:\n%s", out)
	}
	if !strings.Contains(out, "(none — every finding has been disposed)") {
		t.Error("open-findings section should say so when empty")
	}
}

// A sidecar that EXISTS but doesn't parse must error, never yield an empty ledger —
// silently resetting would re-open every disposed finding.
func TestParseSidecarRejectsMissingFrontmatter(t *testing.T) {
	if _, err := ParseSidecar("# Gate ledger\n\nno frontmatter here\n"); err == nil {
		t.Error("a sidecar without frontmatter must error, not yield an empty ledger")
	}
}

func TestParseSidecarRejectsBadYAML(t *testing.T) {
	if _, err := ParseSidecar("---\n:::not: [valid: yaml\n---\n\nbody\n"); err == nil {
		t.Error("unparseable frontmatter must error")
	}
}

// FuzzRenderParseRoundTrip is the property that makes the sidecar safe to hold
// agent-authored text. Two claims, both load-bearing:
//
//  1. **Render NEVER emits a document ParseSidecar cannot read.** An unreadable ledger
//     destroys the gate's memory for that issue permanently — every disposition lost,
//     every finding re-opened — which is the exact failure this package exists to prevent.
//  2. **Canonical form is a fixed point.** Whatever normalization Render applies, applying
//     it again changes nothing, so a ledger is stable across arbitrarily many rounds.
//
// The hazards are real, not hypothetical. Render writes into a `---`-fenced frontmatter
// that frontmatter.Split (pkg/frontmatter/frontmatter.go:26) terminates at the first line
// that is exactly `---`. And go.yaml.in/yaml/v3 mis-emits leading-newline strings with a
// block-scalar indent indicator that contradicts its own output — this fuzz target found
// that within one second of its first run, on input "\n0".
func FuzzRenderParseRoundTrip(f *testing.F) {
	f.Add("seam in wrong layer", "moves the filter boundary")
	f.Add("x", "has\n---\na fence line")
	f.Add("x", "```findings nested inside")
	f.Add("---", "---")
	f.Add("", "")
	f.Add("multi\nline\ntitle", "tab\there")
	f.Add("x", "\n0") // the yaml/v3 emitter bug this target caught

	f.Fuzz(func(t *testing.T, title, detail string) {
		l := Ledger{
			Gate: "plan-quality", IssueNum: 187, IDPrefix: "PQ",
			ContentHash: "deadbeef",
			Rounds: []Round{{
				N: 1, Timestamp: testTimestamp, Agent: testAgent, Blocked: true,
				New: []Finding{{ID: "PQ-1", Severity: "Critical", Title: title, Detail: detail, Round: 1}},
			}},
		}
		got, err := ParseSidecar(Render(l, "ariadne"))
		if err != nil {
			t.Fatalf("Render emitted an unreadable ledger for title=%q detail=%q: %v", title, detail, err)
		}
		again, err := ParseSidecar(Render(got, "ariadne"))
		if err != nil {
			t.Fatalf("re-render unreadable for title=%q detail=%q: %v", title, detail, err)
		}
		if !reflect.DeepEqual(got, again) {
			t.Fatalf("canonical form is not a fixed point for title=%q detail=%q:\n got %+v\nthen %+v",
				title, detail, got, again)
		}
	})
}

// TestRenderParseRoundTripPreservesCanonicalText pins that normalization only strips
// SURROUNDING whitespace — interior structure in a multi-line detail must survive intact,
// or the gate would quietly mangle what the judge actually said.
func TestRenderParseRoundTripPreservesCanonicalText(t *testing.T) {
	detail := "first line\n  indented second\n\nafter a blank"
	l := ledgerWith(round(1, nil, findings("Critical/seam")))
	l.Rounds[0].New[0].Detail = detail

	got, err := ParseSidecar(Render(l, "ariadne"))
	if err != nil {
		t.Fatalf("ParseSidecar: %v", err)
	}
	if got.Rounds[0].New[0].Detail != detail {
		t.Errorf("detail mangled:\n got %q\nwant %q", got.Rounds[0].New[0].Detail, detail)
	}
}
