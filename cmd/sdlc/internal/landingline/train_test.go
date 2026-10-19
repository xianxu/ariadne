package landingline

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

var tt0 = time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)

func hourly() Schedule { return Schedule{Interval: time.Hour, MaxCars: 3} }

func TestDepartureDue(t *testing.T) {
	trainAt := func(next time.Time) []Event {
		return []Event{{Kind: EventEnqueue, At: tt0}, {Kind: EventTrain, At: tt0, NextDue: next}}
	}
	cases := []struct {
		name   string
		s      Schedule
		events []Event
		now    time.Time
		want   bool
	}{
		{"off is the serial line", Schedule{}, trainAt(tt0.Add(time.Hour)), tt0, true},
		{"no train yet", hourly(), nil, tt0, true},
		{"before the next departure", hourly(), trainAt(tt0.Add(time.Hour)), tt0.Add(59 * time.Minute), false},
		{"at the next departure", hourly(), trainAt(tt0.Add(time.Hour)), tt0.Add(time.Hour), true},
		{"the latest train wins", hourly(), append(trainAt(tt0), Event{Kind: EventTrain, NextDue: tt0.Add(2 * time.Hour)}), tt0.Add(time.Hour), false},
		{"a train event without NextDue is not a schedule", hourly(), []Event{{Kind: EventTrain}}, tt0, true},
	}
	for _, c := range cases {
		if got, why := DepartureDue(c.now, c.s, c.events); got != c.want || why == "" {
			t.Errorf("%s: due %v (%q), want %v", c.name, got, why, c.want)
		}
	}
}

func TestNextDue(t *testing.T) {
	if got := NextDue(tt0, hourly(), false); !got.Equal(tt0.Add(time.Hour)) {
		t.Errorf("normal: %v", got)
	}
	if got := NextDue(tt0, hourly(), true); !got.Equal(tt0) {
		t.Errorf("left ready cars: next train should be at once, got %v", got)
	}
}

func oid(c byte) string { return strings.Repeat(string(c), 40) }

func vouch(seq int64, ready bool, at time.Time) Readiness {
	return Readiness{Version: ReadinessVersion, Seq: seq, Branch: "b", Head: oid('a'), PR: int(seq), Ready: ready, At: at}
}

func TestReadinessValidAndCurrent(t *testing.T) {
	good := vouch(4, true, tt0)
	if err := good.Valid(4); err != nil {
		t.Fatal(err)
	}
	bad := map[string]Readiness{}
	r := good
	r.Version = 2
	bad["future version"] = r
	r = good
	r.Head = "abc"
	bad["short head"] = r
	r = good
	r.Head = "$(rm -rf /)" + oid('a')
	bad["shell text"] = r
	r = good
	r.PR = 0
	bad["no PR"] = r
	r = good
	r.At = time.Time{}
	bad["no time"] = r
	for name, r := range bad {
		if r.Valid(4) == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if good.Valid(5) == nil {
		t.Error("a vouch filed under another ticket was accepted")
	}
	notReady := Readiness{Version: ReadinessVersion, Seq: 4, Branch: "b", Reason: "CI pending", At: tt0}
	if err := notReady.Valid(4); err != nil {
		t.Errorf("a not-ready vouch needs no head: %v", err)
	}
	if !good.Current(tt0.Add(MaxReadinessAge)) || good.Current(tt0.Add(MaxReadinessAge+time.Second)) {
		t.Error("currency window")
	}
	if good.Current(tt0.Add(-time.Second)) {
		t.Error("a future-dated vouch is current")
	}
	if notReady.Current(tt0) {
		t.Error("a not-ready vouch is current")
	}
}

func tickets(n int) []Ticket {
	var out []Ticket
	for i := 2; i <= n+1; i++ {
		out = append(out, Ticket{Seq: int64(i), Issue: 100 + i, Branch: "b"})
	}
	return out
}

func boardedSeqs(b Boarding) []int64 {
	var out []int64
	for _, c := range b.Cars {
		out = append(out, c.Seq)
	}
	return out
}

func TestBoard(t *testing.T) {
	head := Car{Seq: 1, Issue: 101, Branch: "h", Head: oid('h'), PR: 1}
	ready := func(seqs ...int64) map[int64]Readiness {
		m := map[int64]Readiness{}
		for _, s := range seqs {
			m[s] = vouch(s, true, tt0)
		}
		return m
	}
	t.Run("boards every ready waiter in order", func(t *testing.T) {
		b := Board(head, false, tickets(2), ready(2, 3), hourly(), tt0)
		if got := boardedSeqs(b); !slices.Equal(got, []int64{1, 2, 3}) || b.LeftReady != 0 {
			t.Fatalf("cars %v left %d", got, b.LeftReady)
		}
	})
	t.Run("a not-ready waiter is passed over, not a stop", func(t *testing.T) {
		r := ready(3)
		r[2] = Readiness{Version: ReadinessVersion, Seq: 2, Branch: "b", Reason: "CI pending", At: tt0}
		b := Board(head, false, tickets(2), r, hourly(), tt0)
		if got := boardedSeqs(b); !slices.Equal(got, []int64{1, 3}) || !strings.Contains(b.Why[2], "CI pending") {
			t.Fatalf("cars %v why %v", got, b.Why)
		}
	})
	t.Run("a vouch naming another branch does not board", func(t *testing.T) {
		r := ready(2, 3)
		v := r[2]
		v.Branch = "000999-someone-else"
		r[2] = v
		waiters := tickets(2)
		for i := range waiters {
			waiters[i].Branch = "b"
		}
		b := Board(head, false, waiters, r, hourly(), tt0)
		if got := boardedSeqs(b); !slices.Equal(got, []int64{1, 3}) || !strings.Contains(b.Why[2], "000999-someone-else") {
			t.Fatalf("cars %v why %v", got, b.Why)
		}
		if b.Cars[1].Branch != "b" {
			t.Fatalf("a car's branch comes from its ticket, got %q", b.Cars[1].Branch)
		}
	})
	t.Run("an unvouched or stale waiter does not board", func(t *testing.T) {
		r := map[int64]Readiness{3: vouch(3, true, tt0.Add(-time.Hour))}
		b := Board(head, false, tickets(2), r, hourly(), tt0)
		if got := boardedSeqs(b); !slices.Equal(got, []int64{1}) || b.Why[2] == "" || b.Why[3] == "" {
			t.Fatalf("cars %v why %v", got, b.Why)
		}
	})
	t.Run("the cap leaves ready cars for an immediate train", func(t *testing.T) {
		b := Board(head, false, tickets(4), ready(2, 3, 4, 5), hourly(), tt0)
		if got := boardedSeqs(b); !slices.Equal(got, []int64{1, 2, 3}) || b.LeftReady != 2 {
			t.Fatalf("cars %v left %d", got, b.LeftReady)
		}
	})
	t.Run("an Alone waiter ends boarding and rides next, by itself", func(t *testing.T) {
		r := ready(2, 4)
		r[3] = Readiness{Version: ReadinessVersion, Seq: 3, Branch: "b", Head: oid('c'), PR: 3, Ready: true, Alone: true, Reason: "touches local-gates paths", At: tt0}
		b := Board(head, false, tickets(3), r, hourly(), tt0)
		if got := boardedSeqs(b); !slices.Equal(got, []int64{1, 2}) || b.LeftReady != 2 || !strings.Contains(b.Why[3], "alone") {
			t.Fatalf("cars %v left %d why %v", got, b.LeftReady, b.Why)
		}
	})
	t.Run("an Alone head boards nobody", func(t *testing.T) {
		b := Board(head, true, tickets(2), ready(2, 3), hourly(), tt0)
		if got := boardedSeqs(b); !slices.Equal(got, []int64{1}) || b.LeftReady != 2 {
			t.Fatalf("cars %v left %d", got, b.LeftReady)
		}
	})
	t.Run("trains off: the head rides alone", func(t *testing.T) {
		b := Board(head, false, tickets(2), ready(2, 3), Schedule{}, tt0)
		if got := boardedSeqs(b); !slices.Equal(got, []int64{1}) {
			t.Fatalf("cars %v", got)
		}
	})
}

// bisectCost runs Bisect against a train whose first failing prefix is
// culprit, counting suites after the first (full-train) one.
func bisectCost(t *testing.T, n, culprit int) int {
	t.Helper()
	results := map[int]bool{n: false}
	suites := 0
	for {
		next, c, lo, done := Bisect(n, results)
		if done {
			if c != culprit || lo != culprit-1 {
				t.Fatalf("n=%d culprit %d: found %d (lo %d)", n, culprit, c, lo)
			}
			return suites
		}
		suites++
		results[next] = next < culprit
	}
}

// The hot-path counted test (tl-worker): a split costs ceil(log2 n) suites at
// most, at N and 2N, for every culprit position; a linear scan would cost n-1.
func TestBisectCostIsLogarithmic(t *testing.T) {
	for _, n := range []int{8, 16} {
		bound := int(math.Ceil(math.Log2(float64(n))))
		for culprit := 1; culprit <= n; culprit++ {
			if got := bisectCost(t, n, culprit); got > bound {
				t.Errorf("n=%d culprit %d: %d suites, bound %d", n, culprit, got, bound)
			}
		}
	}
}

func TestBisectEdges(t *testing.T) {
	if _, c, lo, done := Bisect(1, map[int]bool{1: false}); !done || c != 1 || lo != 0 {
		t.Errorf("one car: culprit %d lo %d done %v", c, lo, done)
	}
	if _, _, _, done := Bisect(3, map[int]bool{3: true}); done {
		t.Error("nothing failed, yet bisect finished")
	}
	// A pass above the first failure (a flaky prefix) never lifts lo past it.
	if _, c, lo, done := Bisect(4, map[int]bool{4: false, 2: false, 3: true, 1: true}); !done || c != 2 || lo != 1 {
		t.Errorf("contradicting results: culprit %d lo %d done %v", c, lo, done)
	}
}

func TestReadinessSidecarSurvivesScansWhileItsTicketIsHeld(t *testing.T) {
	dir := t.TempDir()
	helper := startHelper(t, dir, "vouch")
	l := openLine(t, dir, testOptions(nil))
	for range 3 { // every scan; a sidecar in tickets/ would vanish on the first
		if _, _, err := l.Status(); err != nil {
			t.Fatal(err)
		}
	}
	ready, err := l.Readiness()
	if err != nil || len(ready) != 1 {
		t.Fatalf("readiness %v %v", ready, err)
	}
	for _, r := range ready {
		if !r.Ready || r.Head != oid('a') || r.PR != 9 {
			t.Fatalf("vouch %+v", r)
		}
	}
	kill(t, helper)
	if ready, err := l.Readiness(); err != nil || len(ready) != 0 {
		t.Fatalf("a dead waiter's vouch survived: %v %v", ready, err)
	}
	if left, _ := filepath.Glob(filepath.Join(dir, "ready", "*")); len(left) != 0 {
		t.Fatalf("sidecars left: %v", left)
	}
}

func TestReadinessSidecarIsUntrustedInput(t *testing.T) {
	dir := t.TempDir()
	l := openLine(t, dir, testOptions(nil))
	h := enqueue(t, l, 1)
	if err := l.Vouch(h, Readiness{Branch: "b", Head: oid('b'), PR: 1, Ready: true}); err != nil {
		t.Fatal(err)
	}
	path := readyPath(dir, h.Ticket().Seq)
	for name, body := range map[string]string{
		"torn":          `{"version":1,"seq":`,
		"wrong seq":     `{"version":1,"seq":99,"branch":"b","head":"` + oid('b') + `","pr":1,"ready":true,"at":"2026-10-19T09:00:00Z"}`,
		"newer version": `{"version":2,"seq":1,"branch":"b","head":"` + oid('b') + `","pr":1,"ready":true,"at":"2026-10-19T09:00:00Z"}`,
	} {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if ready, err := l.Readiness(); err != nil || len(ready) != 0 {
			t.Errorf("%s: read as a vouch: %v %v", name, ready, err)
		}
	}
	if err := l.Vouch(h, Readiness{Branch: "b", Head: "nope", PR: 1, Ready: true}); err == nil {
		t.Error("an invalid vouch was written")
	}
	h.Release()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("release left the vouch: %v", err)
	}
	if err := l.Vouch(h, Readiness{Branch: "b", Head: oid('b'), PR: 1, Ready: true}); err == nil {
		t.Error("a released ticket vouched")
	}
}

func TestAdmissibleGatesTheHead(t *testing.T) {
	due := false
	held := []string{}
	o := testOptions(nil)
	attempts := 0
	o.Admission = func() func(Ticket, *Readiness, bool) (bool, string) {
		attempts++
		return func(tk Ticket, r *Readiness, granted bool) (bool, string) {
			if !due {
				return false, "next train at 10:00"
			}
			if r == nil || !r.Ready {
				return false, "not ready"
			}
			return true, ""
		}
	}
	var audit []int64
	o.OnAdmit = func(head Ticket, displaced []Ticket) {
		for _, d := range displaced {
			audit = append(audit, d.Seq)
		}
	}
	o.OnHeld = func(why string) { held = append(held, why) }
	l := openLine(t, t.TempDir(), o)
	a, b := enqueue(t, l, 1), enqueue(t, l, 2)
	// Wait reports why it is held, then the test's Sleep ends the wait.
	l.o.Sleep = func(ctx context.Context, d time.Duration) error { return context.Canceled }
	if err := l.Wait(context.Background(), a); err != context.Canceled {
		t.Fatalf("wait: %v", err)
	}
	wantPos(t, l, a, 1) // first in line, but the train is not due
	due = true
	wantPos(t, l, a, 1) // due, but not vouched
	if err := l.Vouch(b, Readiness{Branch: "b", Head: oid('b'), PR: 2, Ready: true}); err != nil {
		t.Fatal(err)
	}
	// The ready ticket behind a not-ready one is the first admissible: it is
	// admitted, so a not-ready merge never holds the line.
	before := attempts
	wantPos(t, l, b, 0)
	if attempts != before+1 {
		t.Errorf("the rule was built %d times for one attempt over two tickets", attempts-before)
	}
	if !slices.Equal(audit, []int64{a.Ticket().Seq}) {
		t.Errorf("the not-ready ticket the head went ahead of is missing from the audit: %v", audit)
	}
	b.Release()
	if _, _, why, err := l.tryAdmit(a); err != nil || why != "not ready" {
		t.Errorf("why %q %v", why, err)
	}
	if !slices.Contains(held, "next train at 10:00") {
		t.Errorf("OnHeld never saw the schedule: %v", held)
	}
}

func TestAdmissibleNilIsTheSerialLine(t *testing.T) {
	l := openLine(t, t.TempDir(), testOptions(nil))
	a := enqueue(t, l, 1)
	wantPos(t, l, a, 0) // no vouch, no schedule: admitted as before #449
}

func TestTrainEventFieldsSurviveTheSweep(t *testing.T) {
	dir := t.TempDir()
	now := tt0
	l := openLine(t, dir, Options{Now: func() time.Time { return now }})
	e := Event{Kind: EventTrain, Seq: 7, Train: 7, Cars: []int64{7, 8}, Result: "landed", Splits: 1, Culprit: 9, NextDue: tt0.Add(time.Hour)}
	if err := l.Record(e); err != nil {
		t.Fatal(err)
	}
	// A line with a member no reader knows, as a later sdlc might write.
	f, err := os.OpenFile(filepath.Join(dir, EventsFile), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"kind":"car","at":"2026-10-19T09:00:00Z","seq":8,"train":7,"future":{"x":1}}` + "\n")
	f.Close()
	now = tt0.Add(25 * time.Hour) // the next append sweeps
	if err := l.Record(Event{Kind: EventEnqueue, Seq: 10}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, EventsFile))
	if !strings.Contains(string(data), `"future":{"x":1}`) || !strings.Contains(string(data), `"next_due"`) {
		t.Fatalf("the sweep dropped fields:\n%s", data)
	}
	events, skipped, err := l.Events()
	if err != nil || skipped != 0 {
		t.Fatal(err, skipped)
	}
	if next, ok := NextDeparture(events); !ok || !next.Equal(tt0.Add(time.Hour)) {
		t.Errorf("next departure %v %v", next, ok)
	}
	if d := Delays(events); len(d) != 0 {
		t.Errorf("train events counted as landings: %v", d)
	}
	// An older reader's shape decodes the new line, ignoring what it doesn't know.
	var old struct {
		Kind string `json:"kind"`
		Seq  int64  `json:"seq"`
	}
	if err := json.Unmarshal([]byte(strings.SplitN(string(data), "\n", 2)[0]), &old); err != nil || old.Kind != EventTrain {
		t.Errorf("old reader: %+v %v", old, err)
	}
}

func TestDecodeReadinessStrictKnownTolerantUnknown(t *testing.T) {
	good := `{"version":1,"seq":4,"branch":"b","head":"` + oid('a') + `","pr":4,"ready":true,"at":"2026-10-19T09:00:00Z","later":{"nested":[1,{"x":true}]}}`
	r, err := DecodeReadiness([]byte(good))
	if err != nil || r.Valid(4) != nil || string(r.Extra["later"]) != `{"nested":[1,{"x":true}]}` {
		t.Fatalf("%+v %v", r, err)
	}
	for name, body := range map[string]string{
		"duplicate known":   `{"version":1,"seq":4,"seq":5}`,
		"duplicate unknown": `{"version":1,"x":1,"x":2}`,
		"wrong type":        `{"version":"1","seq":4}`,
		"null known":        `{"version":1,"head":null}`,
		"not an object":     `[1]`,
		"trailing data":     `{"version":1} {}`,
	} {
		if _, err := DecodeReadiness([]byte(body)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	// encoding/json matches names case-insensitively; an alias stays unknown
	// and never overrides the exact member.
	r, err = DecodeReadiness([]byte(`{"version":1,"seq":4,"head":"` + oid('a') + `","HEAD":"` + oid('b') + `"}`))
	if err != nil || r.Head != oid('a') || string(r.Extra["HEAD"]) != `"`+oid('b')+`"` {
		t.Fatalf("alias: %+v %v", r, err)
	}
	out, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	back, err := DecodeReadiness(out)
	if err != nil || back.Head != oid('a') {
		t.Fatalf("round trip through the alias: %s → %+v %v", out, back, err)
	}
}

// A rewrite of the same ticket's vouch by this sdlc keeps the members a
// later sdlc wrote, nested ones included (ARCH-EVOLVE).
func TestVouchRewriteKeepsUnknownMembers(t *testing.T) {
	dir := t.TempDir()
	l := openLine(t, dir, testOptions(nil))
	h := enqueue(t, l, 1)
	seq := h.Ticket().Seq
	later := `{"version":1,"seq":` + strconv.FormatInt(seq, 10) + `,"branch":"b","head":"` + oid('a') + `","pr":1,"ready":true,"at":"2026-10-19T09:00:00Z","train_hint":{"lane":"fast","weights":[1,2]}}`
	if err := os.WriteFile(readyPath(dir, seq), []byte(later), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := l.Vouch(h, Readiness{Branch: "b", Reason: "CI gate: pending", At: tt0}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(readyPath(dir, seq))
	r, err := DecodeReadiness(data)
	if err != nil {
		t.Fatalf("%s: %v", data, err)
	}
	if r.Ready || r.Reason != "CI gate: pending" || string(r.Extra["train_hint"]) != `{"lane":"fast","weights":[1,2]}` {
		t.Fatalf("rewrite lost or kept the wrong members: %s", data)
	}
	// A previous vouch that doesn't parse carries nothing forward.
	os.WriteFile(readyPath(dir, seq), []byte(`{"version":1,"x":1,"x":2}`), 0o644)
	if err := l.Vouch(h, Readiness{Branch: "b", Reason: "pending", At: tt0}); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(readyPath(dir, seq))
	if strings.Contains(string(data), `"x"`) {
		t.Fatalf("a malformed vouch's members survived: %s", data)
	}
}

// The vouching goroutine and the admission run concurrently in production
// (vouchWhileWaiting beside Wait); run with -race, this is their regression.
func TestVouchConcurrentWithAdmission(t *testing.T) {
	l := openLine(t, t.TempDir(), testOptions(nil))
	h := enqueue(t, l, 1)
	done := make(chan error, 1)
	go func() {
		for {
			err := l.Vouch(h, Readiness{Branch: "b", Head: oid('a'), PR: 1, Ready: true})
			if errors.Is(err, ErrNotWaiting) {
				done <- nil
				return
			}
			if err != nil {
				done <- err
				return
			}
		}
	}()
	if err := l.Wait(context.Background(), h); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatalf("vouch failed other than by admission: %v", err)
	}
}
