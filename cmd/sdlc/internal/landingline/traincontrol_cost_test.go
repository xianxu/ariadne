package landingline

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/pkg/testgate"
)

// Run the actual count guard under the ordinary Go suite. Its instrumented
// child binary selects only TestTrainDropCountedWork's fixtures, so it cannot
// recursively invoke this driver. At most parent+one child test binary run.
func TestTrainDropCostIsLinear(t *testing.T) {
	// The inherited test deadline owns the budget. Reserve the existing
	// process-group drain allowance so cancellation finishes before it.
	drain := 5 * time.Second
	parent, hasDeadline := t.Deadline()
	deadline, err := trainDropCostDeadline(time.Now(), parent, hasDeadline, drain)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	script := filepath.Join("..", "..", "..", "..", "scripts", "test-landing-drop-cost.py")
	cmd := exec.CommandContext(ctx, "python3", script, "--evidence", t.TempDir(), "--deadline-unix-nanos", strconv.FormatInt(deadline.UnixNano(), 10))
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = drain
	// Python never detaches supervised children. Also collect the group on
	// ordinary script failure, including its own deadline error.
	defer func() { _ = processgroup.Terminate(cmd, true) }()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("counted train drop work: %v\n%s", err, out)
	}
	t.Logf("%s", out)
}

func trainDropCostDeadline(now, parent time.Time, hasDeadline bool, drain time.Duration) (time.Time, error) {
	if !hasDeadline {
		parent = now.Add(testgate.HangCeiling)
	}
	deadline := parent.Add(-drain)
	if !deadline.After(now) {
		return time.Time{}, fmt.Errorf("counted train drop work has no remaining caller budget after process-group drain")
	}
	return deadline, nil
}

func TestTrainDropCostDeadlineUsesCallerBudget(t *testing.T) {
	now := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)
	drain := 5 * time.Second
	parent := now.Add(30 * time.Minute)
	got, err := trainDropCostDeadline(now, parent, true, drain)
	if err != nil || !got.Equal(parent.Add(-drain)) {
		t.Fatalf("parent deadline was replaced: %v %v", got, err)
	}
	got, err = trainDropCostDeadline(now, time.Time{}, false, drain)
	if err != nil || !got.Equal(now.Add(testgate.HangCeiling-drain)) {
		t.Fatalf("deadline-free fallback: %v %v", got, err)
	}
	if _, err := trainDropCostDeadline(now, now.Add(drain), true, drain); err == nil {
		t.Fatal("launched work without shutdown budget")
	}
}

// TestTrainDropCountedWork provides identical production operations at three
// sizes for scripts/test-landing-drop-cost.py. Profiles count executed
// production basic blocks, including callbacks, rather than wall time. The
// accepted batch is seeded once: N independent requests are N separate events
// and would intentionally do more work than this one batch transition.
func TestTrainDropCountedWork(t *testing.T) {
	for _, n := range []int{4, 8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			l := openLine(t, t.TempDir(), Options{Poll: time.Millisecond})
			c := TrainControl{Version: TrainControlVersion, Phase: PhaseTesting}
			tr := NewTrain(1)
			tr.Phase = PhaseTesting
			var head *Handle
			for i := 1; i <= n; i++ {
				branch, root := fmt.Sprintf("branch-%d", i), fmt.Sprintf("/owner/%d", i)
				h, err := l.Enqueue(Ticket{Issue: i, Branch: branch, QueueRoot: root})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(h.Release)
				if i == 1 {
					head = h
					if err := l.Wait(context.Background(), h); err != nil {
						t.Fatal(err)
					}
					c.Train = h.Ticket().Seq
					tr.ID = c.Train
				}
				member := TrainMember{Seq: h.Ticket().Seq, Issue: i, Branch: branch, Head: strings.Repeat("a", 40), PR: i, OwnerRoot: root}
				c.Members = append(c.Members, member)
				tr.Cars = append(tr.Cars, member.Car())
			}
			if err := l.BeginTrain(head, c); err != nil {
				t.Fatal(err)
			}
			stamp := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)
			for _, member := range c.Members {
				c.Drops = append(c.Drops, TrainDrop{Train: c.Train, Member: member, Reason: "counted batch", At: stamp})
			}
			c.Cleanup = TrainCleanup{Pending: true, Publication: c.Publication}
			if err := l.withMutex(func() error { return l.writeTrainControlLocked(c) }); err != nil {
				t.Fatal(err)
			}
			next, effects, sealed, err := l.ApplyTrainDrops(head, tr)
			if err != nil {
				t.Fatal(err)
			}
			held, closed := 0, 0
			for _, effect := range effects {
				switch effect.(type) {
				case Hold:
					held++
				case ClosePR:
					closed++
				default:
					t.Fatalf("unexpected dropped-batch effect %T", effect)
				}
			}
			if next.Phase != PhaseAbandoned || held != n || closed != 1 || len(sealed.CurrentDrops()) != n {
				t.Fatalf("batch did not account for all %d pinned cars", n)
			}
		})
	}
}

func TestTrainDropCandidateIsNotComposed(t *testing.T) {
	c := pureTrainControl()
	_, _, _, err := decideTrainDrop(c, c.Train, c.Members[1], "candidate", time.Now())
	if err == nil {
		t.Fatal("initial integrating candidate accepted as composed member")
	}
	if len(c.Drops) != 0 || c.Cleanup.Pending {
		t.Fatal("candidate refusal changed control")
	}
}

func TestCarsDroppedRejectsAmbiguousRiders(t *testing.T) {
	c := pureTrainControl()
	tr := NewTrain(c.Train)
	tr.Phase = PhaseTesting
	tr.Cars = []Car{c.Members[0].Car(), c.Members[0].Car()}
	drop := TrainDrop{Train: c.Train, Member: c.Members[0], Reason: "repair", At: time.Now()}
	if _, _, err := Transition(tr, CarsDropped{Drops: []TrainDrop{drop}}); err == nil {
		t.Fatal("accepted ambiguous duplicated riding identity")
	}
}
