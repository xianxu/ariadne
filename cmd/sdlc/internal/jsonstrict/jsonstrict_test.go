package jsonstrict

import (
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
