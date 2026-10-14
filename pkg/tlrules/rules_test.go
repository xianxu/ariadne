package tlrules

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	rules, err := Parse(strings.NewReader("# comment\n\na\town\tphrase one\nb\tcited\t`cmd`\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []Rule{{"a", Own, "phrase one"}, {"b", Cited, "`cmd`"}}
	if len(rules) != len(want) {
		t.Fatalf("got %d rules, want %d", len(rules), len(want))
	}
	for i := range want {
		if rules[i] != want[i] {
			t.Errorf("row %d: got %+v, want %+v", i, rules[i], want[i])
		}
	}
}

func TestParseRefusesMalformed(t *testing.T) {
	for _, tc := range []struct{ in, wantErr string }{
		{"a\tphrase\n", "want name<TAB>scope<TAB>phrase"},
		{"a\tmaybe\tphrase\n", "unknown scope"},
		{"a\town\t\n", "want name<TAB>scope<TAB>phrase"},
	} {
		_, err := Parse(strings.NewReader(tc.in))
		if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
			t.Errorf("Parse(%q) = %v, want an error containing %q", tc.in, err, tc.wantErr)
		}
	}
}
