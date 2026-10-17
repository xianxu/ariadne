package fleet

import (
	"testing"
	"time"
)

// #421: the lease round-trips, a legacy one-line record is a commit with no
// holder, and damaged content is an error, never a free lease.
func TestStageLeaseParse(t *testing.T) {
	l := StageLease{Commit: "abc", Holder: "#421", Branch: "000421-x", Slot: "ariadne:2", At: time.Date(2026, 10, 16, 22, 0, 0, 0, time.UTC)}
	got, err := ParseStageLease(l.Format())
	if err != nil || got != l {
		t.Fatalf("round trip = %+v, %v", got, err)
	}
	if got, err := ParseStageLease("abc\n"); err != nil || got != (StageLease{Commit: "abc"}) {
		t.Fatalf("legacy = %+v, %v", got, err)
	}
	for _, bad := range []string{"", "commit abc\nholder #1\n", "commit a\nholder b\nbranch c\nslot d\nwhen e\n", "commit a\nholder b\nbranch c\nslot d\nat yesterday\n"} {
		if _, err := ParseStageLease(bad); err == nil {
			t.Fatalf("damaged %q parsed", bad)
		}
	}
}
