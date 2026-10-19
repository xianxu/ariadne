package landingline

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestOutcomeEndsAWaitOrReleasesIt(t *testing.T) {
	var seen []OutcomeKind
	o := testOptions(nil)
	o.OnOutcome = func(out Outcome) { seen = append(seen, out.Kind) }
	l := openLine(t, t.TempDir(), o)
	head, car := enqueue(t, l, 1), enqueue(t, l, 2)
	wantPos(t, l, head, 0)
	seq := car.Ticket().Seq
	polls := 0
	l.o.Sleep = func(context.Context, time.Duration) error {
		polls++
		switch polls {
		case 1:
			return l.Settle(seq, Outcome{Train: 1, Kind: OutcomeBoarded})
		case 2:
			return l.Settle(seq, Outcome{Train: 1, Kind: OutcomeReleased})
		case 3:
			if _, err := os.Stat(outcomePath(l.dir, seq)); !os.IsNotExist(err) {
				t.Errorf("a released outcome was not cleared: %v", err)
			}
			return l.Settle(seq, Outcome{Train: 2, Kind: OutcomeLanded, PR: 77})
		}
		return errors.New("waited past the landing")
	}
	err := l.Wait(context.Background(), car)
	var settled *Settled
	if !errors.As(err, &settled) || settled.Outcome.Kind != OutcomeLanded || settled.Outcome.PR != 77 {
		t.Fatalf("wait: %v", err)
	}
	if want := []OutcomeKind{OutcomeBoarded, OutcomeReleased, OutcomeLanded}; len(seen) != 3 || seen[0] != want[0] || seen[1] != want[1] || seen[2] != want[2] {
		t.Fatalf("outcomes reported %v", seen)
	}
	car.Release()
	if _, err := os.Stat(outcomePath(l.dir, seq)); !os.IsNotExist(err) {
		t.Errorf("release left the outcome: %v", err)
	}
}

func TestSettleRefusesAGoneCar(t *testing.T) {
	l := openLine(t, t.TempDir(), testOptions(nil))
	h := enqueue(t, l, 1)
	seq := h.Ticket().Seq
	h.Release()
	if err := l.Settle(seq, Outcome{Train: 1, Kind: OutcomeLanded}); !errors.Is(err, ErrGone) {
		t.Fatalf("settle on a released ticket: %v", err)
	}
	if err := l.Settle(seq, Outcome{Train: 1, Kind: "teleported"}); err == nil {
		t.Fatal("an unknown outcome was written")
	}
}

func TestOutcomeOfADeadCarIsSwept(t *testing.T) {
	dir := t.TempDir()
	helper := startHelper(t, dir, "ticket")
	l := openLine(t, dir, testOptions(nil))
	_, waiting, err := l.Status()
	if err != nil || len(waiting) != 1 {
		t.Fatal(waiting, err)
	}
	if err := l.Settle(waiting[0].Seq, Outcome{Train: 1, Kind: OutcomeBoarded}); err != nil {
		t.Fatal(err)
	}
	kill(t, helper)
	if _, _, err := l.Status(); err != nil {
		t.Fatal(err)
	}
	if left, _ := filepath.Glob(filepath.Join(dir, "outcomes", "*")); len(left) != 0 {
		t.Fatalf("a dead car's outcome survived: %v", left)
	}
}

func TestOutcomeCodecIsStrictAndPreserving(t *testing.T) {
	dir := t.TempDir()
	l := openLine(t, dir, testOptions(nil))
	h := enqueue(t, l, 1)
	seq := h.Ticket().Seq
	s := strconv.FormatInt(seq, 10)
	later := `{"version":1,"seq":` + s + `,"train":3,"kind":"boarded","at":"2026-10-19T09:00:00Z","car":{"index":2}}`
	os.WriteFile(outcomePath(dir, seq), []byte(later), 0o644)
	if err := l.Settle(seq, Outcome{Train: 3, Kind: OutcomeLanded}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(outcomePath(dir, seq))
	o, err := DecodeOutcome(data)
	if err != nil || o.Kind != OutcomeLanded || string(o.Extra["car"]) != `{"index":2}` {
		t.Fatalf("%s: %+v %v", data, o, err)
	}
	for name, body := range map[string]string{
		"duplicate":    `{"version":1,"seq":` + s + `,"seq":` + s + `}`,
		"wrong seq":    `{"version":1,"seq":999,"train":3,"kind":"landed","at":"2026-10-19T09:00:00Z"}`,
		"unknown kind": `{"version":1,"seq":` + s + `,"train":3,"kind":"vanished","at":"2026-10-19T09:00:00Z"}`,
		"no train":     `{"version":1,"seq":` + s + `,"kind":"landed","at":"2026-10-19T09:00:00Z"}`,
	} {
		os.WriteFile(outcomePath(dir, seq), []byte(body), 0o644)
		if got := l.outcomeOf(seq); got != nil {
			t.Errorf("%s: read as %+v", name, got)
		}
	}
}
