package gitx

import (
	"reflect"
	"testing"
)

// The first entry's leading status space is part of the record: a modified
// file listed first keeps its whole path (#259), as do spaces and renames.
func TestParseStatusZ(t *testing.T) {
	raw := []byte(" M workshop/plans/000259-a-close-gate.md\x00?? workshop/plans/000259-a-close-review.md\x00R  new name.md\x00old name.md\x00M  a b/c.go\x00")
	got, err := ParseStatusZ(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []StatusEntry{
		{XY: " M", Path: "workshop/plans/000259-a-close-gate.md"},
		{XY: "??", Path: "workshop/plans/000259-a-close-review.md"},
		{XY: "R ", Path: "new name.md", Orig: "old name.md"},
		{XY: "M ", Path: "a b/c.go"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseStatusZ = %#v\nwant %#v", got, want)
	}
	if got, err := ParseStatusZ(nil); err != nil || got != nil {
		t.Fatalf("empty status = %v, %v", got, err)
	}
	for _, bad := range []string{
		"M workshop/x.md\x00", // a trimmed first entry: the lost space is refused, not guessed
		" M x.md",             // no NUL terminator
		"R  new.md\x00",       // a rename without its source
		"ZZ x.md\x00",         // unknown status
	} {
		if _, err := ParseStatusZ([]byte(bad)); err == nil {
			t.Errorf("ParseStatusZ(%q) accepted malformed status", bad)
		}
	}
}
