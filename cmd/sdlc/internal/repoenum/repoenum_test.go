package repoenum

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

var model = vocab.Issue().RepoEnums

func TestParse(t *testing.T) {
	d, err := Parse([]byte(`{"version":1,"enums":{"area":["couch/broadcast/web","sdlc/merge"]}}`), model)
	if err != nil || len(d.Enums["area"]) != 2 {
		t.Fatalf("%+v %v", d, err)
	}
	for _, empty := range []string{"", " \n"} {
		if d, err := Parse([]byte(empty), model); err != nil || d.Allowed("area", "x") {
			t.Fatalf("empty %q: %+v %v", empty, d, err)
		}
	}
	for _, tc := range []struct{ name, in, want string }{
		{"unknown top-level", `{"version":1,"enums":{},"areas":[]}`, "areas"},
		{"unknown field", `{"version":1,"enums":{"component":["x"]}}`, "component"},
		{"version", `{"version":2,"enums":{}}`, "version"},
		{"uppercase", `{"version":1,"enums":{"area":["Couch/x"]}}`, "Couch/x"},
		{"too deep", `{"version":1,"enums":{"area":["a/b/c/d/e"]}}`, "a/b/c/d/e"},
		{"empty segment", `{"version":1,"enums":{"area":["a//b"]}}`, "a//b"},
		{"trailing slash", `{"version":1,"enums":{"area":["a/"]}}`, "a/"},
		{"duplicate", `{"version":1,"enums":{"area":["a","a"]}}`, "twice"},
		{"trailing data", `{"version":1,"enums":{}} {}`, "trailing"},
		{"not json", `nope`, ".sdlc/vocabulary.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.in), model)
			if err == nil || !strings.Contains(err.Error(), tc.want) || !strings.Contains(err.Error(), model.DeclarationPath) {
				t.Fatalf("err %v, want %q and the declaration path", err, tc.want)
			}
		})
	}
}

func TestAllowedAndUndeclared(t *testing.T) {
	d := Declaration{Version: 1, Enums: map[string][]string{"area": {"couch/broadcast/web", "sdlc"}}}
	for v, want := range map[string]bool{
		"couch/broadcast/web":   true,
		"couch/broadcast":       true, // ancestor
		"couch":                 true,
		"sdlc":                  true,
		"couch/session":         false, // sibling
		"couch/broad":           false, // not a segment prefix
		"couch/broadcast/web/x": false, // descendant of a declared value
		"sdlcx":                 false,
	} {
		if got := d.Allowed("area", v); got != want {
			t.Errorf("Allowed(%q) = %v, want %v", v, got, want)
		}
	}
	if got := d.Undeclared("area", []string{"couch", "x/y", "sdlc", "z"}); strings.Join(got, ",") != "x/y,z" {
		t.Fatalf("undeclared %v", got)
	}
	err := Refusal(model, "area", []string{"x/y"})
	if !strings.Contains(err.Error(), `"x/y"`) || !strings.Contains(err.Error(), model.DeclarationPath) || !strings.Contains(err.Error(), "local-gates-verified") {
		t.Fatal(err)
	}
}

func TestMatchPrefix(t *testing.T) {
	for _, tc := range []struct {
		p, v string
		want bool
	}{
		{"couch/broadcast", "couch/broadcast", true},
		{"couch/broadcast", "couch/broadcast/web", true},
		{"couch/broadcast", "couch/broadcaster", false},
		{"couch", "couchx/y", false},
		{"couch/broadcast/web", "couch/broadcast", false},
	} {
		if got := MatchPrefix(tc.p, tc.v); got != tc.want {
			t.Errorf("MatchPrefix(%q, %q) = %v", tc.p, tc.v, got)
		}
	}
}

// Properties: every ancestor of a declared value is allowed, and a prefix match
// means the prefix's segments lead the value's.
func TestAncestorProperty(t *testing.T) {
	values := []string{"a", "a/b", "a/b/c", "x-1/y/z/w", "q9/r"}
	d := Declaration{Version: 1, Enums: map[string][]string{"area": values}}
	for _, v := range values {
		segs := strings.Split(v, "/")
		for i := 1; i <= len(segs); i++ {
			anc := strings.Join(segs[:i], "/")
			if !d.Allowed("area", anc) {
				t.Errorf("ancestor %q of %q not allowed", anc, v)
			}
			if !MatchPrefix(anc, v) {
				t.Errorf("MatchPrefix(%q, %q) false", anc, v)
			}
		}
	}
}

// FuzzParse: any input errors or yields values that all have the field's
// syntax, and a parsed declaration survives a JSON round-trip unchanged.
func FuzzParse(f *testing.F) {
	for _, s := range []string{`{"version":1,"enums":{"area":["a/b"]}}`, `{"version":1,"enums":{}}`, ``, `{"version":1}`, `[]`, `{"version":1,"enums":{"area":["a","b/c/d/e"]}}`} {
		f.Add([]byte(s))
	}
	spec := model.Fields["area"]
	f.Fuzz(func(t *testing.T, data []byte) {
		d, err := Parse(data, model)
		if err != nil {
			return
		}
		for _, v := range d.Enums["area"] {
			if !ValidValue(spec, v) {
				t.Fatalf("accepted invalid %q", v)
			}
		}
		out, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		again, err := Parse(out, model)
		if err != nil || len(again.Enums["area"]) != len(d.Enums["area"]) {
			t.Fatalf("round-trip: %v %+v vs %+v", err, again, d)
		}
	})
}
