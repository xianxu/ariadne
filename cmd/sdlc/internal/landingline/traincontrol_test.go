package landingline

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

func pureTrainControl() TrainControl {
	return TrainControl{Version: 1, Train: 1, Phase: PhaseIntegrating, Members: []TrainMember{{Seq: 1, Issue: 1, Branch: "one", Head: strings.Repeat("a", 40), PR: 1, OwnerRoot: "/one"}, {Seq: 2, Issue: 2, Branch: "two", Head: strings.Repeat("b", 40), PR: 2, OwnerRoot: "/two"}}, Publication: TrainPublication{Repository: "test/repo", Remote: "origin", Branch: "train-1"}}
}

func controlFixture(t *testing.T) (*Line, *Handle, *Handle, TrainControl) {
	t.Helper()
	l := openLine(t, t.TempDir(), Options{Poll: time.Millisecond})
	h := enqueue(t, l, 1)
	if err := l.Wait(context.Background(), h); err != nil {
		t.Fatal(err)
	}
	w := enqueue(t, l, 2)
	members := []TrainMember{{Seq: h.Ticket().Seq, Issue: 1, Branch: "b", Head: strings.Repeat("a", 40), PR: 1, OwnerRoot: "/one"}, {Seq: w.Ticket().Seq, Issue: 2, Branch: "b", Head: strings.Repeat("b", 40), PR: 2, OwnerRoot: "/two"}}
	c := TrainControl{Version: 1, Train: h.Ticket().Seq, Phase: PhaseTesting, Members: members, Publication: TrainPublication{Repository: "test/repo", Remote: "origin", Branch: "train-1"}}
	if err := l.BeginTrain(h, c); err != nil {
		t.Fatal(err)
	}
	return l, h, w, c
}

func TestTrainControlDropWinsMergeFence(t *testing.T) {
	l, h, w, c := controlFixture(t)
	receipt, err := l.RequestTrainDrop(c.Train, c.Members[1], "repair")
	if err != nil || receipt.NotificationError != nil {
		t.Fatalf("drop: %+v %v", receipt, err)
	}
	tr := NewTrain(c.Train)
	tr.Phase = PhaseMerging
	tr.MergeStarted = true
	if _, err := l.StartTrainMerge(h, tr); !errors.Is(err, ErrTrainDropPending) {
		t.Fatalf("merge after accepted drop: %v", err)
	}
	if o := l.outcomeOf(w.Ticket().Seq); o == nil || o.Kind != OutcomeHeld {
		t.Fatalf("held: %+v", o)
	}
	again, err := l.RequestTrainDrop(c.Train, c.Members[1], "other")
	if err != nil || !again.Repeated || again.Drop.Reason != "repair" {
		t.Fatalf("retry: %+v %v", again, err)
	}
}

func TestTrainControlSelfDropSurvivesWaiterPoll(t *testing.T) {
	l, h, w, c := controlFixture(t)
	if _, err := l.RequestTrainDrop(c.Train, c.Members[0], "self repair"); err != nil {
		t.Fatal(err)
	}
	step(t, l, w)
	got, err := l.ReadTrainControl()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.CurrentDrops()) != 1 {
		t.Fatalf("live head's drop retired by another waiter's poll: %+v", got)
	}
	tr := NewTrain(c.Train)
	tr.Phase = PhaseMerging
	tr.MergeStarted = true
	if _, err := l.StartTrainMerge(h, tr); !errors.Is(err, ErrTrainDropPending) {
		t.Fatalf("self drop permitted merging: %v", err)
	}
}

func TestTrainControlConsumedHeldStillPreventsActiveTrainMerge(t *testing.T) {
	l, h, w, c := controlFixture(t)
	if _, err := l.RequestTrainDrop(c.Train, c.Members[1], "repair"); err != nil {
		t.Fatal(err)
	}
	w.Release() // the dropped runner consumed its outcome before watcher ran
	other := enqueue(t, l, 3)
	step(t, l, other)
	got, err := l.ReadTrainControl()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.CurrentDrops()) != 1 {
		t.Fatal("accepted request retired before active conductor applied it")
	}
	tr := NewTrain(c.Train)
	tr.Phase = PhaseMerging
	tr.MergeStarted = true
	if _, err := l.StartTrainMerge(h, tr); !errors.Is(err, ErrTrainDropPending) {
		t.Fatalf("consumed held permitted original train merge: %v", err)
	}
}

func TestTrainControlAcceptedNotificationFailureCannotReboard(t *testing.T) {
	l, h, w, c := controlFixture(t)
	outcomes := filepath.Join(l.dir, "outcomes")
	if err := os.Rename(outcomes, outcomes+"-saved"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outcomes, []byte("blocked"), 0600); err != nil {
		t.Fatal(err)
	}
	receipt, err := l.RequestTrainDrop(c.Train, c.Members[1], "repair")
	if err != nil || receipt.NotificationError == nil || !receipt.Control.Cleanup.Pending {
		t.Fatalf("acceptance lost on notification failure: %+v %v", receipt, err)
	}
	h.Release()
	if _, _, _, err := l.tryAdmit(w); err == nil {
		t.Fatal("admission ignored notification failure")
	}
	if err := os.Remove(outcomes); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(outcomes+"-saved", outcomes); err != nil {
		t.Fatal(err)
	}
	if got := step(t, l, w); got == 0 {
		t.Fatal("accepted drop readmitted")
	}
	if o := l.outcomeOf(w.Ticket().Seq); o == nil || o.Kind != OutcomeHeld {
		t.Fatalf("notification not reconciled: %+v", o)
	}
}

func TestTrainControlCodecRequiresSafetyMarkers(t *testing.T) {
	_, _, _, c := controlFixture(t)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	for _, object := range []string{"publication", "cleanup"} {
		var nested map[string]json.RawMessage
		if err := json.Unmarshal(fields[object], &nested); err != nil {
			t.Fatal(err)
		}
		if object == "publication" {
			delete(nested, "started")
		} else {
			delete(nested, "pending")
		}
		modified := map[string]json.RawMessage{}
		for k, v := range fields {
			modified[k] = v
		}
		modified[object], err = json.Marshal(nested)
		if err != nil {
			t.Fatal(err)
		}
		bad, err := json.Marshal(modified)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := DecodeTrainControl(bad); err == nil {
			t.Fatalf("missing irreversible %s marker defaulted: %s", object, bad)
		}
	}
}

func TestTrainControlRecoveryBranchAndNestedDropsRoundTrip(t *testing.T) {
	_, _, _, c := controlFixture(t)
	c.Publication.RecoveryBranch = "000001-recovery"
	c.Publication.Extra = map[string]json.RawMessage{"future_publication": json.RawMessage(`1`)}
	c.Drops = []TrainDrop{{Train: c.Train, Member: c.Members[0], Reason: "repair", At: time.Now(), Extra: map[string]json.RawMessage{"future_drop": json.RawMessage(`true`)}}}
	c.Cleanup = TrainCleanup{Pending: true, Publication: c.Publication, Extra: map[string]json.RawMessage{"future_cleanup": json.RawMessage(`{"keep":true}`)}}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeTrainControl(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Cleanup.Publication.RecoveryBranch != "000001-recovery" || len(got.Cleanup.Extra) != 1 || len(got.Drops[0].Extra) != 1 || len(got.Drops[0].Member.Extra) != len(c.Members[0].Extra) {
		t.Fatalf("nested metadata lost: %+v", got)
	}
}

func TestTrainControlConcurrentDropAndMergeHaveOneWinner(t *testing.T) {
	for i := 0; i < 30; i++ {
		l, h, _, c := controlFixture(t)
		start := make(chan struct{})
		var wg sync.WaitGroup
		var dropErr, mergeErr error
		wg.Add(2)
		go func() { defer wg.Done(); <-start; _, dropErr = l.RequestTrainDrop(c.Train, c.Members[1], "repair") }()
		go func() {
			defer wg.Done()
			<-start
			tr := NewTrain(c.Train)
			tr.Phase = PhaseMerging
			tr.MergeStarted = true
			_, mergeErr = l.StartTrainMerge(h, tr)
		}()
		close(start)
		wg.Wait()
		if (dropErr == nil) == (mergeErr == nil) {
			t.Fatalf("drop/merge must have exactly one winner: %v / %v", dropErr, mergeErr)
		}
		got, err := l.ReadTrainControl()
		if err != nil {
			t.Fatal(err)
		}
		if got.MergeStarted && len(got.CurrentDrops()) > 0 {
			t.Fatalf("double safety winner: %+v", got)
		}
	}
}

func TestTrainControlSecondDropEitherJoinsSealedBatchOrRefuses(t *testing.T) {
	for i := 0; i < 30; i++ {
		l, h, _, c := controlFixture(t)
		if _, err := l.RequestTrainDrop(c.Train, c.Members[1], "first"); err != nil {
			t.Fatal(err)
		}
		tr := NewTrain(c.Train)
		tr.Phase = PhaseTesting
		for _, m := range c.Members {
			tr.Cars = append(tr.Cars, m.Car())
		}
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		var requestErr, applyErr error
		var effects []Effect
		var sealed *TrainControl
		go func() { defer wg.Done(); <-start; _, requestErr = l.RequestTrainDrop(c.Train, c.Members[0], "second") }()
		go func() { defer wg.Done(); <-start; _, effects, sealed, applyErr = l.ApplyTrainDrops(h, tr) }()
		close(start)
		wg.Wait()
		if applyErr != nil {
			t.Fatal(applyErr)
		}
		held := map[int64]bool{}
		for _, effect := range effects {
			if e, ok := effect.(Hold); ok {
				held[e.Seq] = true
			}
		}
		if !held[c.Members[1].Seq] || held[c.Members[0].Seq] != (requestErr == nil) {
			t.Fatalf("accepted request missing from effects: %v %v", effects, requestErr)
		}
		if !sealed.Phase.Terminal() || len(sealed.CurrentDrops()) != len(held) {
			t.Fatalf("batch not sealed: %+v", sealed)
		}
	}
}

func TestTrainControlCapacityAndCleanupIdentityRefuseWithoutLosingAcceptance(t *testing.T) {
	l, h, _, c := controlFixture(t)
	receipt, err := l.RequestTrainDrop(c.Train, c.Members[1], "repair")
	if err != nil {
		t.Fatal(err)
	}
	expected := receipt.Control.Cleanup
	expected.Publication.Branch = "other"
	if err := l.FinishTrainControl(h, c.Train, expected); err == nil {
		t.Fatal("different cleanup identity accepted")
	}
	before, err := os.ReadFile(filepath.Join(l.dir, trainControlFilename))
	if err != nil {
		t.Fatal(err)
	}
	tooLarge := receipt.Control
	tooLarge.Extra = map[string]json.RawMessage{"huge": json.RawMessage(`"` + strings.Repeat("x", maxTicket) + `"`)}
	if err := l.withMutex(func() error { return l.writeTrainControlLocked(tooLarge) }); err == nil {
		t.Fatal("oversized control accepted")
	}
	after, err := os.ReadFile(filepath.Join(l.dir, trainControlFilename))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("capacity refusal lost accepted request")
	}
}

func TestTrainControlMergeWinsDropAndStaysStarted(t *testing.T) {
	l, h, _, c := controlFixture(t)
	tr := NewTrain(c.Train)
	tr.Phase = PhaseMerging
	tr.MergeStarted = true
	if _, err := l.StartTrainMerge(h, tr); err != nil {
		t.Fatal(err)
	}
	tr.Phase = PhaseReintegrating
	if _, err := l.UpdateTrain(h, tr); err != nil {
		t.Fatal(err)
	}
	if _, err := l.RequestTrainDrop(c.Train, c.Members[1], "late"); !errors.Is(err, ErrTrainMergeStarted) {
		t.Fatalf("late drop: %v", err)
	}
}

func TestTrainDropTerminalRefusalAfterHeadRelease(t *testing.T) {
	for _, phase := range []Phase{PhaseLanded, PhaseUnconfirmed} {
		t.Run(string(phase), func(t *testing.T) {
			l, h, _, c := controlFixture(t)
			tr := NewTrain(c.Train)
			tr.Phase = PhaseMerging
			tr.MergeStarted = true
			if _, err := l.StartTrainMerge(h, tr); err != nil {
				t.Fatal(err)
			}
			tr.Phase = phase
			if _, err := l.UpdateTrain(h, tr); err != nil {
				t.Fatal(err)
			}
			h.Release()
			_, err := l.RequestTrainDrop(c.Train, c.Members[1], "late")
			if !errors.Is(err, ErrTrainMergeStarted) || !strings.Contains(err.Error(), string(phase)) {
				t.Fatalf("terminal state collapsed into generic stale refusal: %v", err)
			}
		})
	}
}

func TestTrainControlCrashAfterAcceptanceReconcilesBeforeAdmission(t *testing.T) {
	l, h, w, c := controlFixture(t)
	// Simulate interruption after the durable safety write but before notification.
	c.Drops = []TrainDrop{{Train: c.Train, Member: c.Members[1], Reason: "repair", At: time.Now()}}
	c.Cleanup = TrainCleanup{Pending: true, Publication: c.Publication}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(l.dir, "train-control.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	h.Release()
	next := enqueue(t, l, 3)
	if got := step(t, l, next); got != 0 {
		t.Fatalf("successor admission: %d", got)
	}
	if o := l.outcomeOf(w.Ticket().Seq); o == nil || o.Kind != OutcomeHeld {
		t.Fatalf("lost accepted exclusion: %+v", o)
	}
	replacement := TrainControl{Version: 1, Train: next.Ticket().Seq, Phase: PhaseForming}
	if err := l.BeginTrain(next, replacement); !errors.Is(err, ErrTrainCleanupPending) {
		t.Fatalf("replaced pending cleanup: %v", err)
	}
	if err := l.FinishTrainControl(next, c.Train, c.Cleanup); err != nil {
		t.Fatal(err)
	}
	if err := l.BeginTrain(next, replacement); err != nil {
		t.Fatal(err)
	}
	got, err := l.ReadTrainControl()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Drops) != 1 || len(got.CurrentDrops()) != 0 {
		t.Fatalf("retained requests: %+v", got)
	}
}

func TestTrainControlRetainedDropIsNotABoardingCandidate(t *testing.T) {
	l, h, w, c := controlFixture(t)
	receipt, err := l.RequestTrainDrop(c.Train, c.Members[1], "repair")
	if err != nil {
		t.Fatal(err)
	}
	if err := l.FinishTrainControl(h, c.Train, receipt.Control.Cleanup); err != nil {
		t.Fatal(err)
	}
	h.Release()
	next := enqueue(t, l, 3)
	if got := step(t, l, next); got != 0 {
		t.Fatalf("next head position %d", got)
	}
	_, waiting, err := l.Status()
	if err != nil {
		t.Fatal(err)
	}
	if slices.ContainsFunc(waiting, func(t Ticket) bool { return t.Seq == w.Ticket().Seq }) {
		t.Fatal("retained dropped ticket exposed to successor Board")
	}
	retry := enqueue(t, l, 2)
	_, waiting, err = l.Status()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(waiting, func(t Ticket) bool { return t.Seq == retry.Ticket().Seq }) {
		t.Fatal("old drop excluded a new explicit retry identity")
	}
}

func TestTrainControlBeginCannotRegisterRetainedDroppedMember(t *testing.T) {
	l, h, _, c := controlFixture(t)
	receipt, err := l.RequestTrainDrop(c.Train, c.Members[1], "repair")
	if err != nil {
		t.Fatal(err)
	}
	if err := l.FinishTrainControl(h, c.Train, receipt.Control.Cleanup); err != nil {
		t.Fatal(err)
	}
	h.Release()
	next := enqueue(t, l, 3)
	if got := step(t, l, next); got != 0 {
		t.Fatalf("next head position %d", got)
	}
	replacement := TrainControl{Version: 1, Train: next.Ticket().Seq, Phase: PhaseForming, Members: []TrainMember{c.Members[1]}}
	if err := l.BeginTrain(next, replacement); err == nil {
		t.Fatal("registered a retained dropped ticket as successor member")
	}
}

func TestCarsDroppedAbortsOnlySurvivorsAndRejectsStarted(t *testing.T) {
	c := pureTrainControl()
	tr := NewTrain(c.Train)
	tr.Phase = PhaseTesting
	for _, m := range c.Members {
		tr.Cars = append(tr.Cars, m.Car())
	}
	drop := TrainDrop{Train: c.Train, Member: c.Members[0], Reason: "repair", At: time.Now()}
	next, effects, err := Transition(tr, CarsDropped{Drops: []TrainDrop{drop}})
	if err != nil || next.Phase != PhaseAbandoned {
		t.Fatalf("drop transition: %+v %v", next, err)
	}
	var held, released []int64
	for _, effect := range effects {
		switch e := effect.(type) {
		case Hold:
			held = append(held, e.Seq)
		case Release:
			released = append(released, e.Seqs...)
		}
	}
	if len(held) != 1 || held[0] != c.Members[0].Seq || len(released) != 1 || released[0] != c.Members[1].Seq {
		t.Fatalf("effects: %+v", effects)
	}
	tr.MergeStarted = true
	tr.Phase = PhaseReintegrating
	if _, _, err := Transition(tr, CarsDropped{Drops: []TrainDrop{drop}}); err == nil {
		t.Fatal("accepted drop after merge start")
	}
}

func TestCarsDroppedAllPreMergePhasesPreserveSurvivorOrder(t *testing.T) {
	c := pureTrainControl()
	c.Members = append(c.Members, TrainMember{Seq: 3, Issue: 3, Branch: "three", Head: strings.Repeat("c", 40), PR: 3, OwnerRoot: "/three"})
	for _, phase := range []Phase{PhaseIntegrating, PhaseTesting, PhaseBisecting, PhaseAttributing, PhaseChecking, PhaseLanding, PhaseReintegrating} {
		for mask := 1; mask < 8; mask++ {
			tr := NewTrain(c.Train)
			tr.Phase = phase
			var drops []TrainDrop
			var survivors []int64
			for i, m := range c.Members {
				tr.Cars = append(tr.Cars, m.Car())
				if mask&(1<<i) != 0 {
					drops = append(drops, TrainDrop{Train: c.Train, Member: m, Reason: "repair", At: time.Now()})
				} else {
					survivors = append(survivors, m.Seq)
				}
			}
			before := tr
			before.Cars = slices.Clone(tr.Cars)
			before.Results = clone(tr.Results)
			next, effects, err := Transition(tr, CarsDropped{Drops: drops})
			if err != nil {
				t.Fatalf("%s mask%d: %v", phase, mask, err)
			}
			if !reflect.DeepEqual(tr, before) || next.Phase != PhaseAbandoned {
				t.Fatalf("mutated source or nonterminal drop in %s", phase)
			}
			var released []int64
			holds, closes := 0, 0
			for _, effect := range effects {
				switch e := effect.(type) {
				case Hold:
					holds++
				case Release:
					released = append(released, e.Seqs...)
				case ClosePR:
					closes++
				case Merge:
					t.Fatal("drop emitted merge")
				}
			}
			if !slices.Equal(released, survivors) || holds != len(drops) || closes != 1 {
				t.Fatalf("%s mask%d effects %+v", phase, mask, effects)
			}
			if _, _, err := Transition(next, CIDone{Green: true}); err == nil {
				t.Fatal("obsolete completion escaped terminal drop")
			}
		}
	}
}

func TestTrainPublicationFencePinsCleanupAndRefusesNextExternalEffect(t *testing.T) {
	l, h, _, c := controlFixture(t)
	p := c.Publication
	p.Started = true
	p.ExpectedTip = strings.Repeat("c", 40)
	p.PreviousTip = strings.Repeat("d", 40)
	p.PR = 100
	p.RecoveryBranch = "b"
	if _, err := l.UpdateTrainPublication(h, p); err != nil {
		t.Fatal(err)
	}
	receipt, err := l.RequestTrainDrop(c.Train, c.Members[1], "repair")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(receipt.Control.Cleanup.Publication, p) {
		t.Fatalf("cleanup not pinned: %+v", receipt)
	}
	p.ExpectedTip = strings.Repeat("e", 40)
	if _, err := l.UpdateTrainPublication(h, p); !errors.Is(err, ErrTrainDropPending) {
		t.Fatalf("publication after accepted drop: %v", err)
	}
	if err := l.FinishTrainControl(h, c.Train, receipt.Control.Cleanup); err != nil {
		t.Fatal(err)
	}
	got, err := l.ReadTrainControl()
	if err != nil {
		t.Fatal(err)
	}
	if got.Cleanup.Pending || len(got.CurrentDrops()) != 1 {
		t.Fatalf("completion lost pinned drop: %+v", got)
	}
}

func TestTrainDropRequestsPreserveCleanupPublicationUnknownFields(t *testing.T) {
	l, h, _, c := controlFixture(t)
	if _, err := l.RequestTrainDrop(c.Train, c.Members[1], "repair waiter"); err != nil {
		t.Fatal(err)
	}
	got, err := l.ReadTrainControl()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]json.RawMessage{"future_cleanup_proof": json.RawMessage(`{"confirmed":"future"}`)}
	got.Cleanup.Publication.Extra = want
	if err := l.withMutex(func() error { return l.writeTrainControlLocked(*got) }); err != nil {
		t.Fatal(err)
	}
	receipt, err := l.RequestTrainDrop(c.Train, c.Members[0], "repair conductor")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(receipt.Control.Cleanup.Publication.Extra, want) {
		t.Fatalf("second acceptance lost cleanup publication metadata: %+v", receipt.Control.Cleanup.Publication.Extra)
	}
	tr := NewTrain(c.Train)
	tr.Phase = c.Phase
	for _, m := range c.Members {
		tr.Cars = append(tr.Cars, m.Car())
	}
	_, _, sealed, err := l.ApplyTrainDrops(h, tr)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.FinishTrainControl(h, c.Train, sealed.Cleanup); err != nil {
		t.Fatal(err)
	}
	got, err = l.ReadTrainControl()
	if err != nil {
		t.Fatal(err)
	}
	if got.Cleanup.Pending || !reflect.DeepEqual(got.Cleanup.Publication.Extra, want) {
		t.Fatalf("sealing/finishing lost cleanup publication metadata: %+v", got.Cleanup)
	}
}

func TestTrainControlCodecPreservesNestedUnknownAndRejectsMalformed(t *testing.T) {
	_, _, _, c := controlFixture(t)
	c.Extra = map[string]json.RawMessage{"future": json.RawMessage(`{"keep":1}`)}
	c.Members[0].Extra = map[string]json.RawMessage{"future_member": json.RawMessage(`true`)}
	c.Publication.Extra = map[string]json.RawMessage{"future_publication": json.RawMessage(`"keep"`)}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeTrainControl(data)
	if err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"future", "future_member", "future_publication"} {
		if !strings.Contains(string(out), key) {
			t.Fatalf("lost %s: %s", key, out)
		}
	}
	for _, bad := range []string{`{}`, `{"version":1,"version":1}`, `{"version":2,"train":1,"phase":"forming"}`, `{"version":1,"train":1,"phase":null}`, `{"version":1,"train":1,"phase":"forming","members":[{"seq":1,"seq":2}]}`} {
		if _, err := DecodeTrainControl([]byte(bad)); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
}
