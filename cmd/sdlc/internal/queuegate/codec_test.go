package queuegate

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestStrictObjectSharedBoundaryPreservesOpaqueValuesAndRejectsAmbiguity(t *testing.T) {
	fields, err := StrictObject([]byte(`{"owner":"A","future": { "n" : 1e99999 }}`))
	if err != nil || string(fields["future"]) != `{ "n" : 1e99999 }` {
		t.Fatalf("opaque object %s %v", fields["future"], err)
	}
	for _, raw := range [][]byte{[]byte(`{"x":1,"x":2}`), []byte(`{"future":{"x":1,"x":2}}`), []byte(`{"future":[{"x":1,"x":2}]}`), []byte(`[]`), []byte(`null`), []byte(`{} {}`), []byte(`{"truncated":`), append([]byte(`{"x":"`), 0xff), bytes.Repeat([]byte(" "), MaxJSON+1)} {
		if _, err := StrictObject(raw); err == nil {
			t.Fatalf("accepted ambiguous object %q", raw)
		}
	}
}

func TestControlCodecPreservesOpaqueAdditions(t *testing.T) {
	raw := []byte(`{"schema_version":1,"pause_owners":[{"owner":"A","reason":"quoted ; $(opaque)","cancel_running":false,"new":{"n":1e9999}}],"future":[null,1e99999]}`)
	c, err := DecodeControl(raw)
	if err != nil {
		t.Fatal(err)
	}
	c, err = Transition(c, Action{Kind: "pause", Owner: "A", CancelRunning: true})
	if err != nil {
		t.Fatal(err)
	}
	got, err := EncodeControl(c)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(got, []byte(`"new":{"n":1e9999}`)) || !bytes.Contains(got, []byte(`"future":[null,1e99999]`)) {
		t.Fatalf("lost opaque fields: %s", got)
	}
	again, err := DecodeControl(got)
	if err != nil || !again.PauseOwners[0].CancelRunning {
		t.Fatalf("roundtrip %+v %v", again, err)
	}
}

func TestControlCodecRejectsMalformedKnownAndNestedJSON(t *testing.T) {
	for _, raw := range []string{`{}`, `null`, `[]`, `{"schema_version":1,"pause_owners":null}`, `{"schema_version":1,"pause_owners":[],"schema_version":1}`, `{"Schema_Version":1,"pause_owners":[]}`, `{"schema_version":2,"pause_owners":[]}`, `{"schema_version":1.0,"pause_owners":[]}`, `{"schema_version":1,"pause_owners":[{"owner":"A","reason":"r"}]}`, `{"schema_version":1,"pause_owners":[{"owner":"A","reason":"r","cancel_running":0}]}`, `{"schema_version":1,"pause_owners":[{"owner":"A","owner":"B","reason":"r","cancel_running":false}]}`, `{"schema_version":1,"pause_owners":[],"future":{"duplicate":1,"duplicate":2}}`, `{"schema_version":1,"pause_owners":[],"future":[{"x":1,"x":2}]}`, `{"schema_version":1,"pause_owners":[]} true`, `{"schema_version":1,"pause_owners":[{"owner":"A","reason":"r","cancel_running":false},{"owner":"A","reason":"r","cancel_running":false}]}`} {
		if _, err := DecodeControl([]byte(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	if _, err := DecodeControl(bytes.Repeat([]byte(" "), MaxJSON+1)); err == nil {
		t.Fatal("oversized input")
	}
	c := NewControl()
	c.Extra = map[string]json.RawMessage{"future": json.RawMessage(`{"x":1,"x":2}`)}
	if _, err := EncodeControl(c); err == nil {
		t.Fatal("encoded duplicate unknown fields")
	}
	c = NewControl()
	c.Extra = map[string]json.RawMessage{"schema_version": json.RawMessage(`2`)}
	if _, err := EncodeControl(c); err == nil {
		t.Fatal("extra overrides known field")
	}
}

func TestOperationCodecRoundTripAndStrictTypes(t *testing.T) {
	op := testOperation()
	op.Extra = map[string]json.RawMessage{"future": json.RawMessage(`1e9999`)}
	raw, err := EncodeOperation(op)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeOperation(raw)
	if err != nil || got.ID != op.ID || string(got.Extra["future"]) != "1e9999" {
		t.Fatalf("roundtrip %+v %v", got, err)
	}
	for _, pair := range [][2]string{{`"phase":"waiting"`, `"phase":"other"`}, {`"kinds":["merge","landing"]`, `"kinds":["suite","suite"]`}, {`"pid":123`, `"pid":"123"`}, {`"managed":false`, `"managed":null`}, {`"workspace":"pair:5"`, `"workspace":null`}, {`"id":"op-1"`, `"id":"../escape"`}, {`"pid_started":"incarnation"`, `"PID_STARTED":"incarnation"`}} {
		bad := strings.Replace(string(raw), pair[0], pair[1], 1)
		if bad == string(raw) {
			t.Fatalf("replacement not found %s in %s", pair[0], raw)
		}
		if _, err := DecodeOperation([]byte(bad)); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
}

func TestCodecPreservesUnknownValueBytesAndRejectsInvalidUTF8(t *testing.T) {
	raw := []byte(`{"schema_version":1,"pause_owners":[],"future": { "value" : 1e99999 } }`)
	c, err := DecodeControl(raw)
	if err != nil {
		t.Fatal(err)
	}
	out, err := EncodeControl(c)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte(`{ "value" : 1e99999 }`)) {
		t.Fatalf("normalized opaque value %s", out)
	}
	bad := append([]byte(`{"schema_version":1,"pause_owners":[],"future":"`), 0xff)
	bad = append(bad, []byte(`"}`)...)
	if _, err := DecodeControl(bad); err == nil {
		t.Fatal("invalid UTF8 accepted")
	}
	deep := `{"schema_version":1,"pause_owners":[],"future":` + strings.Repeat("[", 130) + "0" + strings.Repeat("]", 130) + "}"
	if _, err := DecodeControl([]byte(deep)); err == nil {
		t.Fatal("nesting limit accepted")
	}
}

func FuzzOperationDecode(f *testing.F) {
	seed, _ := EncodeOperation(testOperation())
	f.Add(seed)
	f.Fuzz(func(t *testing.T, b []byte) {
		op, err := DecodeOperation(b)
		if err != nil {
			return
		}
		out, err := EncodeOperation(op)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := DecodeOperation(out); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzControlDecode(f *testing.F) {
	f.Add([]byte(`{"schema_version":1,"pause_owners":[]}`))
	f.Add([]byte(`{"schema_version":1,"pause_owners":[],"x":1e99999}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		c, err := DecodeControl(b)
		if err != nil {
			return
		}
		out, err := EncodeControl(c)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := DecodeControl(out); err != nil {
			t.Fatal(err)
		}
	})
}
