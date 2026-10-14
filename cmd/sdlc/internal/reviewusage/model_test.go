package reviewusage

import (
	"reflect"
	"strings"
	"testing"
)

func testIntent() Intent {
	return Intent{ID: "review-one", Repository: "github.com/example/repo", IssueID: "000365", Recipe: "plan-quality", Round: 1, Attempt: 1, Agent: "claude", Root: "/repo", ReviewedHEAD: strings.Repeat("a", 40), Slot: "repo:3", RequestedAt: "2026-10-14T10:00:00Z"}
}
func TestAttemptConfirmedProgress(t *testing.T) {
	intent := testIntent()
	requested, err := NewRequested(intent)
	if err != nil {
		t.Fatal(err)
	}
	if requested.Phase != Requested || !requested.Uncertain() {
		t.Fatal("unconfirmed launch must remain uncertain")
	}
	start := Event{Kind: Start, Intent: intent, PID: 42, At: "2026-10-14T10:00:01Z"}
	started, err := Step(requested, start)
	if err != nil {
		t.Fatal(err)
	}
	if again, err := Step(started, start); err != nil || again != started {
		t.Fatal("exact start repeat changed progress")
	}
	finish := Event{Kind: Finish, Intent: intent, PID: 42, StartedAt: start.At, At: "2026-10-14T10:00:02Z", SessionID: "11111111-1111-4111-8111-111111111111", Outcome: "returned"}
	complete, err := Step(started, finish)
	if err != nil {
		t.Fatal(err)
	}
	if complete.Phase != Finished || complete.Uncertain() {
		t.Fatal("finished explicit binding not final")
	}
	if again, err := Step(complete, finish); err != nil || again != complete {
		t.Fatal("exact finish repeat changed progress")
	}
	if _, err := Step(requested, finish); err == nil {
		t.Fatal("finished launched process without start confirmation")
	}
	for _, change := range []func(*Event){func(e *Event) { e.Intent.Repository = "foreign" }, func(e *Event) { e.Intent.ID = "foreign" }, func(e *Event) { e.PID = 43 }, func(e *Event) { e.StartedAt = "2026-10-14T10:00:00Z" }, func(e *Event) { e.SessionID = "22222222-2222-4222-8222-222222222222" }, func(e *Event) { e.Outcome = "canceled" }} {
		event := finish
		change(&event)
		got, err := Step(complete, event)
		if err == nil || got != complete {
			t.Fatal("rewritten confirmation changed state")
		}
	}
	// Every immutable intent member is consumed by callbacks.
	value := reflect.ValueOf(&finish.Intent).Elem()
	for i := 0; i < value.NumField(); i++ {
		bad := finish
		field := reflect.ValueOf(&bad.Intent).Elem().Field(i)
		switch field.Kind() {
		case reflect.String:
			field.SetString(field.String() + "foreign")
		case reflect.Int:
			field.SetInt(field.Int() + 1)
		}
		if got, err := Step(started, bad); err == nil || got != started {
			t.Fatalf("ignored intent member %s", value.Type().Field(i).Name)
		}
	}
}

func TestAttemptLaunchFailureAndDeathEquivalent(t *testing.T) {
	r, err := NewRequested(testIntent())
	if err != nil {
		t.Fatal(err)
	}
	dead := r
	if err := Validate(dead); err != nil || !dead.Uncertain() {
		t.Fatal("requested death equivalent not retained")
	}
	s, err := Step(r, Event{Kind: Start, Intent: r.Intent, PID: 42, At: "2026-10-14T10:00:01Z"})
	if err != nil {
		t.Fatal(err)
	}
	if !s.Uncertain() {
		t.Fatal("started death equivalent not incomplete")
	}
	failed, err := Step(r, Event{Kind: Finish, Intent: r.Intent, At: "2026-10-14T10:00:01Z", Outcome: "launch-failed"})
	if err != nil || failed.Phase != Finished {
		t.Fatalf("known launch refusal not recorded: %v", err)
	}
}

func TestAttemptBoundsAndOrdering(t *testing.T) {
	r, _ := NewRequested(testIntent())
	for _, mutate := range []func(*Record){func(r *Record) { r.Version = 2 }, func(r *Record) { r.Intent.ID = "bad/id" }, func(r *Record) { r.Intent.Attempt = 3 }, func(r *Record) { r.Intent.Round = 0 }, func(r *Record) { r.Intent.Root = "relative" }, func(r *Record) { r.Intent.Repository = "" }, func(r *Record) { r.Intent.RequestedAt = "bad" }, func(r *Record) { r.Phase = "unknown" }, func(r *Record) { r.PID = 4 }, func(r *Record) { r.Intent.Root = "/" + strings.Repeat("x", MaxBytes) }} {
		bad := r
		mutate(&bad)
		if Validate(bad) == nil {
			t.Fatal("malformed/oversize record accepted")
		}
	}
	for _, at := range []string{"bad", "2026-10-14T09:59:59Z"} {
		if got, err := Step(r, Event{Kind: Start, Intent: r.Intent, PID: 42, At: at}); err == nil || got != r {
			t.Fatal("invalid start timestamp admitted")
		}
	}
}
