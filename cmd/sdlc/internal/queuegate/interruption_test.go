package queuegate

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func interruptionCodecFixture() InterruptionRecord {
	op := testOperation()
	op.Phase = "unknown"
	return InterruptionRecord{
		SchemaVersion: 1,
		ID:            op.ID,
		Outcome:       "interrupted",
		Operation:     op,
		Reason:        "fast pause; descendant lifetime remains unproved",
		RecordedAt:    time.Date(2026, 10, 20, 9, 5, 0, 0, time.UTC),
	}
}

func TestInterruptionRecordCodecPreservesEnvelopeAndOperationAdditions(t *testing.T) {
	r := interruptionCodecFixture()
	r.Extra = map[string]json.RawMessage{"future_envelope": json.RawMessage(`{ "large" : 1e99999 }`)}
	r.Operation.Extra = map[string]json.RawMessage{"future_operation": json.RawMessage(`[ 1e9999, { "value" : null } ]`)}
	raw, err := EncodeInterruptionRecord(r)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeInterruptionRecord(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != r.ID || got.Outcome != "interrupted" || got.Operation.Phase != "unknown" || got.Reason != r.Reason || !got.RecordedAt.Equal(r.RecordedAt) {
		t.Fatalf("round trip changed evidence: %+v", got)
	}
	if !bytes.Equal(got.Extra["future_envelope"], r.Extra["future_envelope"]) || !bytes.Equal(got.Operation.Extra["future_operation"], r.Operation.Extra["future_operation"]) {
		t.Fatalf("lost opaque fields: envelope %s operation %s", got.Extra["future_envelope"], got.Operation.Extra["future_operation"])
	}
	rewritten, err := EncodeInterruptionRecord(got)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(rewritten, r.Extra["future_envelope"]) || !bytes.Contains(rewritten, r.Operation.Extra["future_operation"]) {
		t.Fatalf("normalized opaque JSON on rewrite: %s", rewritten)
	}
	if _, err := DecodeInterruptionRecord(rewritten); err != nil {
		t.Fatal(err)
	}
}

func TestInterruptionRecordCodecRejectsMissingNullAndWrongEnvelopeFields(t *testing.T) {
	raw, err := EncodeInterruptionRecord(interruptionCodecFixture())
	if err != nil {
		t.Fatal(err)
	}
	wrong := map[string]json.RawMessage{
		"schema_version": json.RawMessage(`"1"`),
		"id":             json.RawMessage(`7`),
		"outcome":        json.RawMessage(`false`),
		"operation":      json.RawMessage(`[]`),
		"reason":         json.RawMessage(`{}`),
		"recorded_at":    json.RawMessage(`7`),
	}
	for field, bad := range wrong {
		for _, mode := range []string{"missing", "null", "wrong"} {
			t.Run(field+"/"+mode, func(t *testing.T) {
				var fields map[string]json.RawMessage
				if err := json.Unmarshal(raw, &fields); err != nil {
					t.Fatal(err)
				}
				switch mode {
				case "missing":
					delete(fields, field)
				case "null":
					fields[field] = json.RawMessage(`null`)
				case "wrong":
					fields[field] = bad
				}
				candidate, err := json.Marshal(fields)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := DecodeInterruptionRecord(candidate); err == nil {
					t.Fatalf("accepted %s", candidate)
				}
			})
		}
	}
}

func TestInterruptionRecordCodecRejectsMissingNullAndWrongNestedFields(t *testing.T) {
	raw, err := EncodeInterruptionRecord(interruptionCodecFixture())
	if err != nil {
		t.Fatal(err)
	}
	wrong := map[string]json.RawMessage{
		"schema_version": json.RawMessage(`"1"`), "id": json.RawMessage(`7`),
		"workspace": json.RawMessage(`[]`), "kinds": json.RawMessage(`"suite"`),
		"phase": json.RawMessage(`false`), "pid": json.RawMessage(`"123"`),
		"pid_started": json.RawMessage(`1`), "managed": json.RawMessage(`1`),
		"started_at": json.RawMessage(`1`), "boot_id": json.RawMessage(`false`),
		"root": json.RawMessage(`{}`), "detail": json.RawMessage(`1`),
	}
	for field, bad := range wrong {
		for _, mode := range []string{"missing", "null", "wrong"} {
			t.Run(field+"/"+mode, func(t *testing.T) {
				var envelope, operation map[string]json.RawMessage
				if err := json.Unmarshal(raw, &envelope); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(envelope["operation"], &operation); err != nil {
					t.Fatal(err)
				}
				switch mode {
				case "missing":
					delete(operation, field)
				case "null":
					operation[field] = json.RawMessage(`null`)
				case "wrong":
					operation[field] = bad
				}
				envelope["operation"], err = json.Marshal(operation)
				if err != nil {
					t.Fatal(err)
				}
				candidate, err := json.Marshal(envelope)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := DecodeInterruptionRecord(candidate); err == nil {
					t.Fatalf("accepted %s", candidate)
				}
			})
		}
	}
}

func TestInterruptionRecordCodecRejectsMismatchedIdentityAndOutcome(t *testing.T) {
	r := interruptionCodecFixture()
	r.ID = "different-operation"
	if _, err := EncodeInterruptionRecord(r); err == nil {
		t.Fatal("encoded mismatched operation identity")
	}
	raw, err := EncodeInterruptionRecord(interruptionCodecFixture())
	if err != nil {
		t.Fatal(err)
	}
	for field, bad := range map[string]json.RawMessage{
		"id": json.RawMessage(`"different-operation"`), "outcome": json.RawMessage(`"failed"`),
		"schema_version": json.RawMessage(`2`), "reason": json.RawMessage(`""`),
		"recorded_at": json.RawMessage(`"0001-01-01T00:00:00Z"`),
	} {
		t.Run(field, func(t *testing.T) {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			fields[field] = bad
			candidate, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeInterruptionRecord(candidate); err == nil {
				t.Fatalf("accepted inconsistent evidence %s", candidate)
			}
		})
	}
}
