package gitx

import (
	"strings"
	"testing"
)

func TestParseIndexStageZ(t *testing.T) {
	oid := strings.Repeat("a", 40)
	entries, err := ParseIndexStageZ([]byte("H 100644 " + oid + " 0\tworkshop/issues/000001-a\nquoted.md\x00s 100644 " + oid + " 0\tworkshop/history/000002-b.md\x00M 100644 " + oid + " 2\tconflict.md\x00"))
	if err != nil || len(entries) != 3 || entries[0].Path != "workshop/issues/000001-a\nquoted.md" || entries[1].Tag != "s" || entries[2].Stage != 2 {
		t.Fatalf("entries %#v: %v", entries, err)
	}
	for _, raw := range []string{"H 100644 " + oid + " 0\tx", "H 100644 invalid 0\tx\x00", "H 100644 " + oid + " 4\tx\x00", "H 100644 " + oid + " 0\t../escape\x00", "Q 100644 " + oid + " 0\tx\x00"} {
		if _, err := ParseIndexStageZ([]byte(raw)); err == nil {
			t.Errorf("accepted malformed %q", raw)
		}
	}
}

func FuzzParseIndexStageZ(f *testing.F) {
	f.Add([]byte("H 100644 " + strings.Repeat("a", 40) + " 0\tfile.md\x00"))
	f.Add([]byte("truncated"))
	f.Fuzz(func(t *testing.T, raw []byte) {
		entries, err := ParseIndexStageZ(raw)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.Path == "" || e.Stage < 0 || e.Stage > 3 {
				t.Fatalf("invalid %#v", e)
			}
		}
	})
}
