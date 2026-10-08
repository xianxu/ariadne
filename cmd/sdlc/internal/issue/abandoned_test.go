package issue

import (
	"strings"
	"testing"
)

// #286: the abandon record round-trips beside other records, nil clears it,
// and a malformed one fails closed both when written and when read back (a
// reopen fetches the ref it names).
func TestAbandonedRecord(t *testing.T) {
	card, err := SetCardRelease(releaseCard(t), &Release{By: ReleasedBy(releaser)})
	if err != nil {
		t.Fatal(err)
	}
	want := Abandoned{Ref: AbandonedRef("000286"), Head: strings.Repeat("c", 40)}
	out, err := SetCardAbandoned(card, &want)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok, err := CardAbandoned(out); err != nil || !ok || got != want {
		t.Fatalf("round trip: %+v %v %v", got, ok, err)
	}
	if _, ok, _ := CardRelease(out); !ok {
		t.Fatal("writing the abandon record dropped the release")
	}
	cleared, err := SetCardAbandoned(out, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := CardAbandoned(cleared); ok || err != nil {
		t.Fatalf("clear: %v %v", ok, err)
	}
	for _, bad := range []Abandoned{
		{Ref: "refs/heads/000286-x", Head: want.Head},
		{Ref: "refs/ariadne/abandoned/286", Head: want.Head},
		{Ref: want.Ref, Head: "HEAD"},
		{Ref: want.Ref},
	} {
		if _, err := SetCardAbandoned(card, &bad); err == nil || !strings.Contains(err.Error(), "tracker.abandoned") {
			t.Errorf("%+v accepted: %v", bad, err)
		}
	}
	tampered := []byte(strings.Replace(string(out), want.Ref, "refs/heads/main", 1))
	if _, _, err := CardAbandoned(tampered); err == nil || !strings.Contains(err.Error(), "tracker.abandoned.ref") {
		t.Fatalf("a tampered record read back: %v", err)
	}
}
