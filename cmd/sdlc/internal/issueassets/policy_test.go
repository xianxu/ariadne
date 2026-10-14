package issueassets

import (
	"strings"
	"testing"
)

func TestPolicyStrictAndFailClosed(t *testing.T) {
	for _, s := range []string{`{}`, `null`, `{"version":2,"policy":"allowed"}`, `{"version":1,"policy":"allowed","extra":1}`, `{"version":1,"policy":"allowed","policy":"text-only"}`, `{"version":1,"policy":"allowed"} {}`} {
		if _, e := ParsePolicy([]byte(s)); e == nil || !strings.Contains(e.Error(), "operator") {
			t.Errorf("%s: %v", s, e)
		}
	}
	p, e := ParsePolicy([]byte(`{"version":1,"policy":"allowed"}`))
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range []string{"public", "unknown", "", "private", "PRIVATE"} {
		d := ResolvePolicy(nil, v)
		if d.Allowed != (v == "private") {
			t.Fatal(v, d)
		}
		if !ResolvePolicy(&p, v).Allowed {
			t.Fatal("explicit policy ignored")
		}
	}
	p.Policy = ModeTextOnly
	if ResolvePolicy(&p, "private").Allowed {
		t.Fatal("explicit deny ignored")
	}
}
func FuzzParsePolicy(f *testing.F) {
	f.Add(`{"version":1,"policy":"allowed"}`)
	f.Add(`null`)
	f.Fuzz(func(t *testing.T, s string) {
		p, e := ParsePolicy([]byte(s))
		if e == nil && p.Version != 1 {
			t.Fatal(p)
		}
	})
}
