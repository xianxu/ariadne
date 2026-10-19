package landingline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogAppendReadAndSkipMalformed(t *testing.T) {
	l := openLine(t, t.TempDir(), Options{Now: func() time.Time { return t0 }})
	if err := l.Record(Event{Kind: EventEnqueue, Seq: 1, Issue: 7}); err != nil {
		t.Fatal(err)
	}
	f, _ := os.OpenFile(filepath.Join(l.dir, EventsFile), os.O_APPEND|os.O_WRONLY, 0)
	f.WriteString("{truncated\n")
	f.Close()
	if err := l.Record(Event{Kind: EventLanded, Seq: 1, Issue: 7}); err != nil {
		t.Fatal(err)
	}
	events, skipped, err := l.Events()
	if err != nil || skipped != 1 || len(events) != 2 || events[1].Kind != EventLanded || !events[0].At.Equal(t0) {
		t.Fatalf("events %v skipped %d err %v", events, skipped, err)
	}
}

func TestEventSweepPreservesUnknownMembersInOriginalLine(t *testing.T) {
	now := t0
	l := openLine(t, t.TempDir(), Options{Now: func() time.Time { return now }})
	if err := l.Record(Event{Kind: EventEnqueue, Seq: 1}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(l.dir, EventsFile)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	raw := `{"kind":"admit","at":"` + t0.Format(time.RFC3339Nano) + `","seq":2,"future":{"nested":[1,2]}}`
	if _, err := f.WriteString(raw + "\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	now = t0.Add(25 * time.Hour)
	if err := l.Record(Event{Kind: EventLanded, Seq: 3}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), raw+"\n") {
		t.Fatalf("sweep changed a newer event's original JSON:\n%s", content)
	}
}

func TestLogSweepDropsOldEntriesOnceADay(t *testing.T) {
	now := t0
	l := openLine(t, t.TempDir(), Options{Now: func() time.Time { return now }})
	if err := l.Record(Event{Kind: EventEnqueue, Seq: 1}); err != nil {
		t.Fatal(err)
	}
	now = t0.Add(31 * 24 * time.Hour)
	if err := l.Record(Event{Kind: EventEnqueue, Seq: 2}); err != nil {
		t.Fatal(err)
	}
	events, _, _ := l.Events()
	if len(events) != 1 || events[0].Seq != 2 {
		t.Fatalf("after sweep: %v", events)
	}
	// A second sweep the same day doesn't run: an entry made old meanwhile stays.
	now = now.Add(time.Hour)
	old := Event{Kind: EventEnqueue, Seq: 3, At: t0}
	if err := l.Record(old); err != nil {
		t.Fatal(err)
	}
	if events, _, _ = l.Events(); len(events) != 2 {
		t.Fatalf("swept twice in a day: %v", events)
	}
}

func TestGrantsLogRoundTrip(t *testing.T) {
	l := openLine(t, t.TempDir(), Options{Now: func() time.Time { return t0 }})
	g := grant(391, t0, 4)
	if err := l.AddGrant(g); err != nil {
		t.Fatal(err)
	}
	entries, _, err := l.GrantEntries()
	if err != nil || len(ValidGrants(entries, t0.Add(time.Hour))) != 1 {
		t.Fatalf("grants %v err %v", entries, err)
	}
}

func TestDelays(t *testing.T) {
	events := []Event{
		{Kind: EventEnqueue, Seq: 1, At: t0},
		{Kind: EventEnqueue, Seq: 2, At: t0.Add(time.Minute)},
		{Kind: EventAdmit, Seq: 1, At: t0},
		{Kind: EventLanded, Seq: 1, At: t0.Add(10 * time.Minute)},
		{Kind: EventEnqueue, Seq: 3, At: t0},
		{Kind: EventRefused, Seq: 3, At: t0.Add(time.Minute)},
		{Kind: EventLanded, Seq: 2, At: t0.Add(31 * time.Minute)},
		{Kind: EventLanded, Seq: 9, At: t0}, // its enqueue was swept: no sample
	}
	got := Delays(events)
	if len(got) != 2 || got[0] != 600 || got[1] != 1800 {
		t.Fatalf("delays %v", got)
	}
}

// The admit hook runs outside the line mutex, so it can log the admission.
func TestAdmitHookMayRecord(t *testing.T) {
	var l *Line
	o := Options{Poll: time.Millisecond, OnAdmit: func(head Ticket, displaced []Ticket) {
		if err := l.Record(Event{Kind: EventAdmit, Seq: head.Seq}); err != nil {
			t.Error(err)
		}
	}}
	l = openLine(t, t.TempDir(), o)
	wantPos(t, l, enqueue(t, l, 1), 0)
	if events, _, _ := l.Events(); len(events) != 1 || events[0].Kind != EventAdmit {
		t.Fatalf("events %v", events)
	}
}

func TestLogSweepDropsFutureDatedAndCountsMalformed(t *testing.T) {
	now := t0
	l := openLine(t, t.TempDir(), Options{Now: func() time.Time { return now }})
	if err := l.Record(Event{Kind: EventEnqueue, Seq: 1, At: t0.Add(100 * 24 * time.Hour)}); err != nil {
		t.Fatal(err)
	}
	f, _ := os.OpenFile(filepath.Join(l.dir, EventsFile), os.O_APPEND|os.O_WRONLY, 0)
	f.WriteString("{torn\n")
	f.Close()
	now = t0.Add(25 * time.Hour) // the next day's sweep
	if err := l.Record(Event{Kind: EventEnqueue, Seq: 2}); err != nil {
		t.Fatal(err)
	}
	events, skipped, _ := l.Events()
	if len(events) != 1 || events[0].Seq != 2 || skipped != 0 {
		t.Fatalf("after sweep: %v skipped %d", events, skipped)
	}
	stamp, _ := os.ReadFile(filepath.Join(l.dir, EventsFile+".swept"))
	if string(stamp) != "malformed=1\n" {
		t.Fatalf("stamp %q", stamp)
	}
}

// A revoke must outlive the grant it revokes: AddGrant stamps it, and the
// sweep keys a grant's age on the later of Granted and Expires.
func TestSweptRevokeKeepsTheGrantRevoked(t *testing.T) {
	now := t0
	l := openLine(t, t.TempDir(), Options{Now: func() time.Time { return now }})
	g := grant(391, t0, 24)
	if err := l.AddGrant(g); err != nil {
		t.Fatal(err)
	}
	now = t0.Add(time.Hour)
	if err := l.AddGrant(Grant{Issue: 391, Grantor: g.Grantor, Reason: "landed", Revoked: true}); err != nil {
		t.Fatal(err)
	}
	now = t0.Add(26 * time.Hour) // a daily sweep runs on this write
	if err := l.AddGrant(grant(7, now, 1)); err != nil {
		t.Fatal(err)
	}
	entries, _, _ := l.GrantEntries()
	if _, ok := ValidGrants(entries, t0.Add(2*time.Hour))[391]; ok {
		t.Fatalf("revoked grant valid again after a sweep: %v", entries)
	}
}

// A sweep run while grants are in force keeps them (the M1 round-2 regression).
func TestSweepKeepsGrantsInForce(t *testing.T) {
	now := t0
	l := openLine(t, t.TempDir(), Options{Now: func() time.Time { return now }})
	if err := l.AddGrant(grant(391, now, 24)); err != nil {
		t.Fatal(err)
	}
	now = t0.Add(23 * time.Hour)
	if err := l.AddGrant(grant(392, now, 4)); err != nil { // no sweep yet: same day
		t.Fatal(err)
	}
	now = t0.Add(23*time.Hour + 30*time.Minute)
	stamp := filepath.Join(l.dir, GrantsFile+".swept")
	if err := os.Chtimes(stamp, t0.Add(-time.Hour), t0.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := l.AddGrant(grant(393, now, 4)); err != nil { // this write sweeps
		t.Fatal(err)
	}
	entries, _, _ := l.GrantEntries()
	if got := ValidGrants(entries, now); len(got) != 3 {
		t.Fatalf("valid after sweep %v", got)
	}
}

func TestLogReadsAnyLineLength(t *testing.T) {
	l := openLine(t, t.TempDir(), Options{Now: func() time.Time { return t0 }})
	long := make([]byte, 2<<20)
	for i := range long {
		long[i] = 'x'
	}
	if err := l.Record(Event{Kind: EventEnqueue, Seq: 1, Detail: string(long)}); err != nil {
		t.Fatal(err)
	}
	if err := l.Record(Event{Kind: EventLanded, Seq: 1}); err != nil {
		t.Fatal(err)
	}
	events, skipped, err := l.Events()
	if err != nil || skipped != 0 || len(events) != 2 {
		t.Fatalf("events %d skipped %d err %v", len(events), skipped, err)
	}
}

func TestReadGrantEntriesDoesNotCreateLine(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "absent", "landing")
	entries, skipped, err := ReadGrantEntries(dir)
	if err != nil || skipped != 0 || len(entries) != 0 {
		t.Fatalf("%v %d %v", entries, skipped, err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("read created directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, GrantsFile), 0755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadGrantEntries(dir); err == nil {
		t.Fatal("directory treated as valid log")
	}
}
