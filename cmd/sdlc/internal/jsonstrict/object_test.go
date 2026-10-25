package jsonstrict

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

type record struct {
	Version int    `json:"version"`
	Name    string `json:"name,omitempty"`
}

func recordFields(r *record) map[string]any {
	return map[string]any{"version": &r.Version, "name": &r.Name}
}

func TestDecodeObjectStrictness(t *testing.T) {
	cases := []struct {
		in, err string
	}{
		{`{"version":1,"name":"a"}`, ""},
		{`{"version":1,"later":{"x":[1]}}`, ""},
		{`[1]`, "not one JSON object"},
		{`{"version":1,"version":2}`, `duplicate member "version"`},
		{`{"version":null}`, `member "version" is null`},
		{`{"version":"1"}`, `member "version"`},
		{`{"Version":1}`, ""}, // a different name: unknown, preserved, never case-folded onto version
		{`{"version":1} {}`, "trailing data"},
		{`{"version":1`, ""}, // truncated: some error
	}
	for _, c := range cases {
		var r record
		_, err := DecodeObject([]byte(c.in), recordFields(&r))
		switch {
		case c.in == `{"version":1`:
			if err == nil {
				t.Errorf("%s: truncated input decoded", c.in)
			}
		case c.err == "" && err != nil:
			t.Errorf("%s: %v", c.in, err)
		case c.err != "" && (err == nil || !strings.Contains(err.Error(), c.err)):
			t.Errorf("%s: err = %v, want %q", c.in, err, c.err)
		}
		if c.in == `{"Version":1}` && r.Version != 0 {
			t.Errorf("case-folded onto a known member: %+v", r)
		}
	}
}

// An older-shaped reader rewriting a newer record keeps every member it doesn't know, and a member it does know is
// never overridden by extra.
func TestEncodeWithExtraRoundTrip(t *testing.T) {
	in := []byte(`{"version":1,"name":"a","z":{"nested":[1,{"k":true}]},"b":"x"}`)
	var r record
	extra, err := DecodeObject(in, recordFields(&r))
	if err != nil {
		t.Fatal(err)
	}
	extra["name"] = json.RawMessage(`"spoof"`)
	base, _ := json.Marshal(r)
	known := func(k string) bool { _, ok := recordFields(&r)[k]; return ok }
	out := EncodeWithExtra(base, extra, known)
	want := `{"version":1,"name":"a","b":"x","z":{"nested":[1,{"k":true}]}}`
	if string(out) != want {
		t.Fatalf("got %s\nwant %s", out, want)
	}
	if got := EncodeWithExtra(base, nil, known); !bytes.Equal(got, base) {
		t.Fatalf("no extra changed the encoding: %s", got)
	}
}

func FuzzDecodeObject(f *testing.F) {
	for _, s := range []string{`{"version":1}`, `{"version":1,"x":[]}`, `{"version":1,"version":1}`, `{}`, `null`, `{"a":1} x`} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		var r record
		extra, err := DecodeObject(data, recordFields(&r))
		if err != nil {
			return
		}
		// Whatever decodes re-encodes to an object that decodes to the same known values and the same extra.
		base, _ := json.Marshal(r)
		out := EncodeWithExtra(base, extra, func(k string) bool { _, ok := recordFields(&r)[k]; return ok })
		var r2 record
		extra2, err := DecodeObject(out, recordFields(&r2))
		if err != nil || r2 != r || len(extra2) != len(extra) {
			t.Fatalf("lossy round trip: %s -> %s (%v)", data, out, err)
		}
		for k, v := range extra {
			var a, b any
			if json.Unmarshal(v, &a) != nil || json.Unmarshal(extra2[k], &b) != nil {
				continue
			}
			ja, _ := json.Marshal(a)
			jb, _ := json.Marshal(b)
			if !bytes.Equal(ja, jb) {
				t.Fatalf("member %q changed: %s -> %s", k, v, extra2[k])
			}
		}
	})
}
