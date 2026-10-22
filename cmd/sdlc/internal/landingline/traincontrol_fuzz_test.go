package landingline

import (
	"encoding/json"
	"testing"
)

// The codec is a cross-process safety boundary: successful decodes must remain
// valid after a canonical rewrite, including later writers' nested fields.
func FuzzTrainControlCodec(f *testing.F) {
	c := pureTrainControl()
	c.Extra = map[string]json.RawMessage{"future": json.RawMessage(`{"keep":true}`)}
	c.Members[0].Extra = map[string]json.RawMessage{"future_member": json.RawMessage(`17`)}
	seed, err := json.Marshal(c)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"version":1,"version":1}`))
	f.Add([]byte(`{"version":null}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		c, err := DecodeTrainControl(data)
		if err != nil {
			return
		}
		first, err := json.Marshal(c)
		if err != nil {
			t.Fatal(err)
		}
		round, err := DecodeTrainControl(first)
		if err != nil {
			t.Fatalf("accepted record cannot survive rewrite: %v", err)
		}
		second, err := json.Marshal(round)
		if err != nil {
			t.Fatal(err)
		}
		if string(first) != string(second) {
			t.Fatalf("rewrite not stable:\n%s\n%s", first, second)
		}
	})
}
