package landingline

import (
	"os"
	"path/filepath"
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
	admitted(t, admitAsync(l, enqueue(t, l, 1)))
	if events, _, _ := l.Events(); len(events) != 1 || events[0].Kind != EventAdmit {
		t.Fatalf("events %v", events)
	}
}
