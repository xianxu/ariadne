package tracker

import (
	"strings"
	"testing"
	"time"
)

func record(when, op string, paths ...string) string {
	return "\x01" + when + "\x00#1: tracker: update card\n\nTracker-Operation: " + op + "\n\x00\n" + strings.Join(paths, "\n") + "\n"
}

// #284: the newest claim-kind commit per wanted card; other operations, other
// cards and malformed records are skipped; reading stops once all are found.
func TestClaimTimes(t *testing.T) {
	a, b := "workshop/issue-cards/000001-a.md", "workshop/issue-cards/000002-b.md"
	stream := record("2026-10-07T12:00:00-07:00", "set-aaa", a) + // not a claim
		record("2026-10-07T11:00:00-07:00", "claim-bbb", a, b) + // a set claim
		"\x01garbage-without-fields" +
		record("2026-10-06T09:00:00-07:00", "reclaim-ccc", a) + // older: ignored
		record("not-a-time", "claim-ddd", b)
	got, err := ClaimTimes(strings.NewReader(stream), map[string]bool{a: true, b: true})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := time.Parse(time.RFC3339, "2026-10-07T11:00:00-07:00")
	if !got[a].Equal(want) || !got[b].Equal(want) {
		t.Fatalf("got %v", got)
	}
	if got, _ := ClaimTimes(strings.NewReader(record("2026-10-07T11:00:00-07:00", "unclaim-x", a)), map[string]bool{a: true}); len(got) != 0 {
		t.Fatalf("an unclaim is not a claim: %v", got)
	}
}

// ClaimTimes stops reading once every wanted card is resolved.
func TestClaimTimesStopsEarly(t *testing.T) {
	a := "workshop/issue-cards/000001-a.md"
	r := &countingReader{r: strings.NewReader(record("2026-10-07T11:00:00-07:00", "claim-x", a) + strings.Repeat(record("2026-01-01T00:00:00Z", "claim-y", a), 20000))}
	if _, err := ClaimTimes(r, map[string]bool{a: true}); err != nil {
		t.Fatal(err)
	}
	if r.n > 1<<20 {
		t.Fatalf("read %d bytes after the answer was known", r.n)
	}
}

type countingReader struct {
	r interface{ Read([]byte) (int, error) }
	n int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += n
	return n, err
}

// FuzzClaimTimes: any stream, malformed or truncated, never panics.
func FuzzClaimTimes(f *testing.F) {
	f.Add(record("2026-10-07T11:00:00-07:00", "claim-x", "p"))
	f.Add("\x01\x00\x00")
	f.Add("\x01x\x00Tracker-Operation: claim-\x00")
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = ClaimTimes(strings.NewReader(s), map[string]bool{"p": true})
	})
}
