package jsonstrict

import (
	"strings"
	"testing"
)

func TestRejectDuplicateKeys(t *testing.T) {
	for in, want := range map[string]string{
		`{"a":1,"b":{"c":[{"d":1},{"d":2}]}}`: "",
		`{"a":1,"a":2}`:                       `duplicate object key "a"`,
		`{"g":[{"n":"x","n":"y"}]}`:           `duplicate object key "n"`,
		`{"a":{"b":1,"b":1}}`:                 `duplicate object key "b"`,
		`{"a":1} {"b":2}`:                     "trailing JSON value",
		`{"a":`:                               "EOF",
	} {
		err := RejectDuplicateKeys([]byte(in))
		if want == "" && err != nil || want != "" && (err == nil || !strings.Contains(err.Error(), want)) {
			t.Errorf("%s: err %v, want %q", in, err, want)
		}
	}
}
