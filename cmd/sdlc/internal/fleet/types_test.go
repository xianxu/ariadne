package fleet

import (
	"encoding/json"
	"testing"
)

const testPolicyDigest = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestPolicyEnvelopesRoundTripOnlyTotalVariants(t *testing.T) {
	capabilityValue := &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{}, Capacity: Capacity{Kind: "unbounded"}}
	diagnostic := &PolicyDiagnostic{Code: DiagnosticInvalidPolicy, Message: "bad declaration"}
	resultValue := &PolicyResultValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, RepoIdentity: "/repo/.git", AdmissionKey: "/repo/.git", Capacity: Capacity{Kind: "unbounded"}}

	tests := []struct {
		name  string
		value any
		out   any
	}{
		{name: "capability success", value: PolicyCapability{OK: true, Value: capabilityValue}, out: &PolicyCapability{}},
		{name: "capability diagnostic", value: PolicyCapability{Diagnostic: diagnostic}, out: &PolicyCapability{}},
		{name: "result success", value: PolicyResult{OK: true, Value: resultValue}, out: &PolicyResult{}},
		{name: "result diagnostic", value: PolicyResult{Diagnostic: diagnostic}, out: &PolicyResult{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, tt.out); err != nil {
				t.Fatalf("round trip failed for %s: %v; JSON=%s", tt.name, err, raw)
			}
		})
	}
}

func TestPolicyEnvelopesRejectImpossibleVariants(t *testing.T) {
	value := &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, KeyKind: "repo", Roots: []string{}, Capacity: Capacity{Kind: "unbounded"}}
	diagnostic := &PolicyDiagnostic{Code: DiagnosticInvalidPolicy, Message: "bad declaration"}

	for name, envelope := range map[string]PolicyCapability{
		"success without value":      {OK: true},
		"success with diagnostic":    {OK: true, Value: value, Diagnostic: diagnostic},
		"failure without diagnostic": {},
		"failure with value":         {Value: value, Diagnostic: diagnostic},
	} {
		t.Run("marshal "+name, func(t *testing.T) {
			if _, err := json.Marshal(envelope); err == nil {
				t.Fatalf("json.Marshal accepted impossible variant: %+v", envelope)
			}
		})
	}

	invalidJSON := map[string]string{
		"success without value":      `{"ok":true}`,
		"success with diagnostic":    `{"ok":true,"value":{"policy_version":1,"policy_digest":"d","key_kind":"repo","roots":[],"capacity":{"kind":"unbounded"}},"diagnostic":{"code":"bad","message":"bad"}}`,
		"failure without diagnostic": `{"ok":false}`,
		"failure with value":         `{"ok":false,"value":{"policy_version":1,"policy_digest":"d","key_kind":"repo","roots":[],"capacity":{"kind":"unbounded"}},"diagnostic":{"code":"bad","message":"bad"}}`,
		"unknown field":              `{"ok":false,"diagnostic":{"code":"bad","message":"bad"},"extra":true}`,
	}
	for name, raw := range invalidJSON {
		t.Run("unmarshal "+name, func(t *testing.T) {
			var got PolicyCapability
			if err := json.Unmarshal([]byte(raw), &got); err == nil {
				t.Fatalf("json.Unmarshal accepted impossible variant: %s", raw)
			}
		})
	}

	resultValue := &PolicyResultValue{PolicyVersion: 1, PolicyDigest: testPolicyDigest, RepoIdentity: "/repo/.git", AdmissionKey: "/repo/.git", Capacity: Capacity{Kind: "unbounded"}}
	for name, envelope := range map[string]PolicyResult{
		"success without value":      {OK: true},
		"success with diagnostic":    {OK: true, Value: resultValue, Diagnostic: diagnostic},
		"failure without diagnostic": {},
		"failure with value":         {Value: resultValue, Diagnostic: diagnostic},
	} {
		t.Run("marshal result "+name, func(t *testing.T) {
			if _, err := json.Marshal(envelope); err == nil {
				t.Fatalf("json.Marshal accepted impossible result variant: %+v", envelope)
			}
		})
	}
	for name, raw := range map[string]string{
		"success without value":      `{"ok":true}`,
		"failure without diagnostic": `{"ok":false}`,
		"unknown field":              `{"ok":false,"diagnostic":{"code":"bad","message":"bad"},"extra":true}`,
	} {
		t.Run("unmarshal result "+name, func(t *testing.T) {
			var got PolicyResult
			if err := json.Unmarshal([]byte(raw), &got); err == nil {
				t.Fatalf("json.Unmarshal accepted impossible result variant: %s", raw)
			}
		})
	}
}

func TestPolicyEnvelopesRejectDiagnosticsOutsideTheirClosedVariants(t *testing.T) {
	capabilityCodes := []string{DiagnosticMissingPolicy, DiagnosticInvalidPolicy}
	for _, code := range capabilityCodes {
		if _, err := json.Marshal(PolicyCapability{Diagnostic: &PolicyDiagnostic{Code: code, Message: "refused"}}); err != nil {
			t.Fatalf("PolicyCapability rejected modeled code %q: %v", code, err)
		}
	}

	resultCodes := []string{DiagnosticMissingPolicy, DiagnosticInvalidPolicy, DiagnosticOutsideDeclaredScope}
	for _, code := range resultCodes {
		if _, err := json.Marshal(PolicyResult{Diagnostic: &PolicyDiagnostic{Code: code, Message: "refused"}}); err != nil {
			t.Fatalf("PolicyResult rejected modeled code %q: %v", code, err)
		}
	}

	for _, tt := range []struct {
		name  string
		value any
		raw   string
	}{
		{
			name:  "capability rejects result-only code",
			value: PolicyCapability{Diagnostic: &PolicyDiagnostic{Code: DiagnosticOutsideDeclaredScope, Message: "refused"}},
			raw:   `{"ok":false,"diagnostic":{"code":"outside-declared-scope","message":"refused"}}`,
		},
		{
			name:  "capability rejects unknown code",
			value: PolicyCapability{Diagnostic: &PolicyDiagnostic{Code: "invalid-polciy", Message: "refused"}},
			raw:   `{"ok":false,"diagnostic":{"code":"invalid-polciy","message":"refused"}}`,
		},
		{
			name:  "result rejects unknown code",
			value: PolicyResult{Diagnostic: &PolicyDiagnostic{Code: "invalid-polciy", Message: "refused"}},
			raw:   `{"ok":false,"diagnostic":{"code":"invalid-polciy","message":"refused"}}`,
		},
		{
			name:  "result rejects removed unreachable code",
			value: PolicyResult{Diagnostic: &PolicyDiagnostic{Code: "path-outside-repo", Message: "refused"}},
			raw:   `{"ok":false,"diagnostic":{"code":"path-outside-repo","message":"refused"}}`,
		},
	} {
		t.Run(tt.name+" on marshal", func(t *testing.T) {
			if _, err := json.Marshal(tt.value); err == nil {
				t.Fatalf("json.Marshal accepted diagnostic outside closed variants: %+v", tt.value)
			}
		})
		t.Run(tt.name+" on unmarshal", func(t *testing.T) {
			switch tt.value.(type) {
			case PolicyCapability:
				var got PolicyCapability
				if err := json.Unmarshal([]byte(tt.raw), &got); err == nil {
					t.Fatalf("json.Unmarshal accepted diagnostic outside capability variants: %s", tt.raw)
				}
			case PolicyResult:
				var got PolicyResult
				if err := json.Unmarshal([]byte(tt.raw), &got); err == nil {
					t.Fatalf("json.Unmarshal accepted diagnostic outside result variants: %s", tt.raw)
				}
			}
		})
	}
}
