package jsonstrict

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRejectDuplicateKeys(t *testing.T) {
	for in, want := range map[string]string{
		`{"future":{"opaque":1e9999}}`:        "",
		`{"a":1,"b":{"c":[{"d":1},{"d":2}]}}`: "",
		`{"a":1,"a":2}`:                       `duplicate object key "a"`,
		`{"g":[{"n":"x","n":"y"}]}`:           `duplicate object key "n"`,
		`{"a":{"b":1,"b":1}}`:                 `duplicate object key "b"`,
		`{"command":["x"],"COMMAND":["y"]}`:   `duplicate object key "COMMAND"`,
		`{"groups":[],"groupſ":[]}`:           `duplicate object key "groupſ"`, // long s folds to s
		`{"kind":1,"Kind":2}`:                 `duplicate object key`,          // Kelvin sign folds to k
		`{"straße":1,"STRASSE":2}`:            "",                              // not a simple fold: distinct keys
		`{"a":1} {"b":2}`:                     "trailing JSON value",
		`{"a":`:                               "EOF",
	} {
		err := RejectDuplicateKeys([]byte(in))
		if want == "" && err != nil || want != "" && (err == nil || !strings.Contains(err.Error(), want)) {
			t.Errorf("%s: err %v, want %q", in, err, want)
		}
	}
}

// foldKey agrees with strings.EqualFold, encoding/json's matching rule, on
// every pair FuzzFoldKey finds.
func FuzzFoldKey(f *testing.F) {
	for _, s := range [][2]string{{"groups", "groupſ"}, {"kind", "\u212aind"}, {"Command", "COMMAND"}, {"a", "b"}, {"ß", "SS"}} {
		f.Add(s[0], s[1])
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		if (foldKey(a) == foldKey(b)) != strings.EqualFold(a, b) {
			t.Fatalf("foldKey(%q)==foldKey(%q) is %v, EqualFold %v", a, b, foldKey(a) == foldKey(b), strings.EqualFold(a, b))
		}
	})
}

// #489: Scan bounds nesting when asked (receipts allow 12 levels), reads one
// value and leaves the rest of the stream to its caller; numbers stay opaque.
func TestScan(t *testing.T) {
	// The top-level value is depth 0, as receipts count it: 13 nested arrays
	// reach depth 12 and pass a bound of 12; 14 do not.
	deep := strings.Repeat("[", 14) + strings.Repeat("]", 14)
	if err := Scan(json.NewDecoder(strings.NewReader(deep)), 12); err == nil || !strings.Contains(err.Error(), "exceeds 12") {
		t.Fatalf("depth 13 under a 12 bound: %v", err)
	}
	if err := Scan(json.NewDecoder(strings.NewReader(deep[1:len(deep)-1])), 12); err != nil {
		t.Fatalf("depth 12: %v", err)
	}
	if err := Scan(json.NewDecoder(strings.NewReader(deep)), 0); err != nil {
		t.Fatalf("unbounded: %v", err)
	}
	dec := json.NewDecoder(strings.NewReader(`{"a":1} {"b":2}`))
	if err := Scan(dec, 0); err != nil || !dec.More() {
		t.Fatalf("one value, stream left: %v", err)
	}
	if err := RejectDuplicateKeys([]byte(`{"n":1e400}`)); err != nil {
		t.Fatalf("an out-of-range number is opaque: %v", err)
	}
}
