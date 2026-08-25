package fleet

import (
	"encoding/json"
	"testing"
)

func TestPolicyEnvelopesRoundTripOnlyTotalVariants(t *testing.T) {
	capabilityValue := &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: "digest", KeyKind: "repo", Roots: []string{}, Capacity: Capacity{Kind: "unbounded"}}
	diagnostic := &PolicyDiagnostic{Code: DiagnosticInvalidPolicy, Message: "bad declaration"}
	resultValue := &PolicyResultValue{PolicyVersion: 1, PolicyDigest: "digest", RepoIdentity: "/repo/.git", AdmissionKey: "/repo/.git", Capacity: Capacity{Kind: "unbounded"}}

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
	value := &PolicyCapabilityValue{PolicyVersion: 1, PolicyDigest: "digest", KeyKind: "repo", Roots: []string{}, Capacity: Capacity{Kind: "unbounded"}}
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

	resultValue := &PolicyResultValue{PolicyVersion: 1, PolicyDigest: "digest", RepoIdentity: "/repo/.git", AdmissionKey: "/repo/.git", Capacity: Capacity{Kind: "unbounded"}}
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
