package refresh

import (
	"strings"
	"testing"
)

func TestEligibility(t *testing.T) {
	for _, tc := range []struct{ equal, ancestor, rebase, want bool }{{true, false, false, true}, {false, true, false, true}, {false, false, false, false}, {false, false, true, true}} {
		if eligibility(tc.equal, tc.ancestor, tc.rebase) != tc.want {
			t.Fatalf("%+v", tc)
		}
	}
}
func TestAdvance(t *testing.T) {
	valid := map[[2]int]Phase{{int(inspecting), int(checked)}: ready, {int(ready), int(validated)}: applying, {int(applying), int(confirmed)}: applying, {int(applying), int(finished)}: compiling, {int(compiling), int(compiled)}: complete}
	for p := inspecting; p <= stopped; p++ {
		for e := checked; e <= failed; e++ {
			got, err := advance(p, e)
			want, ok := valid[[2]int{int(p), int(e)}]
			if e == failed && p != complete && p != stopped {
				want, ok = stopped, true
			}
			if (err == nil) != ok || ok && got != want {
				t.Fatalf("%v %v: %v %v", p, e, got, err)
			}
		}
	}
}
func TestParsers(t *testing.T) {
	for _, s := range []string{"", strings.Repeat("a", 39), strings.Repeat("g", 40), strings.Repeat("a", 40) + "\nextra\n"} {
		if _, err := parseOID(s); err == nil {
			t.Fatalf("accepted %q", s)
		}
	}
	if _, err := parseOID(strings.Repeat("a", 40) + "\n"); err != nil {
		t.Fatal(err)
	}
	if got, err := parseRecord(" /path with space \n"); err != nil || got != " /path with space " {
		t.Fatalf("%q %v", got, err)
	}
	for _, s := range []string{"HEAD\n", "refs/heads/a..b\n", "refs/heads/x\nextra\n"} {
		if _, err := parseBranch(s); err == nil {
			t.Fatalf("accepted %q", s)
		}
	}
}
func FuzzParseRecord(f *testing.F) {
	f.Add(" /tmp/path \n")
	f.Fuzz(func(t *testing.T, s string) {
		v, e := parseRecord(s)
		if e == nil && v+"\n" != s {
			t.Fatal("framing lost")
		}
	})
}
func TestNULRecords(t *testing.T) {
	for _, s := range []string{"unterminated", "\x00", "a\x00\x00"} {
		if _, e := parseNULRecords(s); e == nil {
			t.Fatalf("accepted %q", s)
		}
	}
	v, e := parseNULRecords(" path \n\x00next\x00")
	if e != nil || len(v) != 2 || v[0] != " path \n" {
		t.Fatalf("%q %v", v, e)
	}
}
func TestInvalidGitEvidence(t *testing.T) {
	if _, e := parseOID(strings.Repeat("0", 40) + "\n"); e == nil {
		t.Fatal("zero ID accepted")
	}
	for _, s := range []string{"refs/heads/a\x01b\n", "refs/heads/a\x7fb\n"} {
		if _, e := parseBranch(s); e == nil {
			t.Fatal("control ref accepted")
		}
	}
}
