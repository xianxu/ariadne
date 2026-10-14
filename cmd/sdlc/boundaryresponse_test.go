package main

import (
	"fmt"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"strings"
	"testing"
)

func TestValidateBoundaryResponse(t *testing.T) {
	ledger := gatestate.Ledger{Rounds: []gatestate.Round{{N: 1, New: []gatestate.Finding{{ID: "BR-1", Severity: "Important", Title: "old"}, {ID: "BR-2", Severity: "Minor", Title: "minor"}}}}}
	cases := []struct {
		name, verdict, body string
		good                bool
	}{
		{"ship", "SHIP", "{}", true},
		{"unknown verdict", "BOGUS", "{}", false},
		{"bare rework", "REWORK", "{}", false},
		{"new blocker", "REWORK", "findings:\n- id: new\n  severity: Important\n  title: broken\n  detail: concrete defect", true},
		{"implicit new", "REWORK", "findings:\n- severity: Critical\n  title: broken\n  detail: concrete defect", true},
		{"no detail", "REWORK", "findings:\n- id: new\n  severity: Important\n  title: broken", false},
		{"minor only", "REWORK", "findings:\n- id: new\n  severity: Minor\n  title: broken\n  detail: detail", false},
		{"forged new ID", "REWORK", "findings:\n- id: BR-88\n  severity: Important\n  title: broken\n  detail: detail", false},
		{"explicit old", "REWORK", "dispose:\n- id: BR-1\n  disposition: not-addressed", true},
		{"minor old", "REWORK", "dispose:\n- id: BR-2\n  disposition: not-addressed", false},
		{"unknown ID ship", "SHIP", "dispose:\n- id: BR-99\n  disposition: addressed", false},
		{"implementer disposition", "SHIP", "dispose:\n- id: BR-1\n  disposition: disputed", false},
		{"unknown kind", "SHIP", "findings:\n- severity: Minor\n  kind: invented\n  title: title", false},
		{"malformed", "SHIP", "[bad", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := "VERDICT: " + tc.verdict + "\n```findings\n" + tc.body + "\n```"
			if err := validateBoundaryResponse(out, ledger); (err == nil) != tc.good {
				t.Fatalf("error = %v, good = %v", err, tc.good)
			}
		})
	}
	if validateBoundaryResponse("VERDICT: SHIP", ledger) == nil {
		t.Fatal("missing findings accepted")
	}
}

func TestValidateBoundaryResponseDispute(t *testing.T) {
	for _, kind := range []string{"", "defect", "unverifiable"} {
		ledger := gatestate.Ledger{IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, Boundary: "M1", New: []gatestate.Finding{{ID: "BR-1", Kind: kind, Severity: "Important", Title: "measurement blocked", Detail: "sandbox"}}}}, Disputes: []gatestate.Dispute{{ID: "BR-1", After: 1, Evidence: "host command passed", Rationale: "host has access"}}}
		text := gatestate.Render(ledger, "test")
		parsed, err := gatestate.ParseSidecar(text)
		if err != nil {
			t.Fatal(err)
		}
		scoped := gatestate.FilterBoundary(parsed, "M1")
		for _, tc := range []struct {
			body     string
			answered bool
		}{
			{"{}", false},
			{"dispose:\n- id: BR-1\n  disposition: withdrawn", false},
			{"dispose:\n- id: BR-1\n  disposition: withdrawn\n  note: host command demonstrates the invariant", true},
			{"dispose:\n- id: BR-1\n  disposition: addressed\n  note: host evidence establishes the check passes", true},
		} {
			if err := validateBoundaryResponse("VERDICT: SHIP\n```findings\n"+tc.body+"\n```", scoped); (err == nil) != (kind != "unverifiable" || tc.answered) {
				t.Fatalf("kind %q body %q error %v", kind, tc.body, err)
			}
		}
		// Another boundary cannot dispose this ID, even though the unfiltered ledger knows it.
		if validateBoundaryResponse("VERDICT: SHIP\n```findings\ndispose:\n- id: BR-1\n  disposition: withdrawn\n  note: host passed\n```", gatestate.FilterBoundary(parsed, "M2")) == nil {
			t.Fatal("out of scope ID accepted")
		}
	}
}

func FuzzValidateBoundaryResponse(f *testing.F) {
	for _, s := range []string{"", "VERDICT: SHIP\n```findings\n{}\n```", "VERDICT: REWORK\n```findings\nfindings:\n- severity: Important\n  title: bug\n  detail: details\n```", "VERDICT: SHIP\n```findings\n[bad\n```"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, output string) {
		ledger := gatestate.Ledger{}
		err := validateBoundaryResponse(output, ledger)
		if (err == nil) != (validateBoundaryResponse(output, ledger) == nil) {
			t.Fatal("nondeterministic admission")
		}
		if err == nil {
			if _, ok := gatestate.ParseFindingsBlock(output); !ok {
				t.Fatal("invalid findings admitted")
			}
		}
	})
}

func TestValidateBoundaryResponseUnverifiableRequiresDetail(t *testing.T) {
	for _, verdict := range []string{"SHIP", "FIX-THEN-SHIP", "REWORK"} {
		for _, detail := range []string{"", "  ", "host check needed after sandbox refused process lookup"} {
			output := "VERDICT: " + verdict + "\n```findings\nfindings:\n- id: new\n  kind: unverifiable\n  severity: Minor\n  title: Cannot verify\n  detail: " + fmt.Sprintf("%q", detail) + "\n- id: new\n  severity: Important\n  title: Separate blocker\n  detail: Concrete defect\n```"
			err := validateBoundaryResponse(output, gatestate.Ledger{})
			if (err == nil) != (strings.TrimSpace(detail) != "") {
				t.Errorf("verdict=%s detail=%q error=%v", verdict, detail, err)
			}
		}
	}
}
