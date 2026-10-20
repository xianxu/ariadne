package landingline

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
)

func TestQueuePauseRetainsLandingTicketAndAdmitsAfterResume(t *testing.T) {
	ctx := context.Background()
	store := &queuegate.Store{Root: t.TempDir()}
	op := queuegate.Operation{SchemaVersion: 1, ID: "merge-1", Workspace: "ariadne:3", Kinds: []string{"merge", "landing"}, Phase: "waiting", PID: os.Getpid(), PIDStarted: "fixture", StartedAt: time.Now(), Root: t.TempDir(), BootID: "fixture", Managed: true}
	lease, err := store.Enroll(ctx, op)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	if err := store.Pause(ctx, "drain", "hold", false); err != nil {
		t.Fatal(err)
	}
	o := testOptions(nil)
	o.Gate = func(resource func() (bool, error)) (bool, error) {
		admitted, err := lease.TryAdmit(ctx, resource)
		var paused *queuegate.PausedError
		if errors.As(err, &paused) {
			return false, &HeldError{Reason: err.Error()}
		}
		return admitted, err
	}
	l := openLine(t, t.TempDir(), o)
	h := enqueue(t, l, 481)
	pos, _, why, err := l.tryAdmit(h)
	if err != nil || pos == 0 || why == "" {
		t.Fatalf("paused admission pos=%d why=%q err=%v", pos, why, err)
	}
	head, waiting, err := l.Status()
	if err != nil || head != nil || len(waiting) != 1 {
		t.Fatalf("head=%v waiting=%v err=%v", head, waiting, err)
	}
	snapshot, err := store.Snapshot(ctx)
	if err != nil || len(snapshot.Active) != 0 {
		t.Fatalf("waiting operation active: %+v %v", snapshot, err)
	}
	if err := store.Resume(ctx, "drain"); err != nil {
		t.Fatal(err)
	}
	wantPos(t, l, h, 0)
	snapshot, err = store.Snapshot(ctx)
	if err != nil || len(snapshot.Active) != 2 {
		t.Fatalf("admitted operation missing: %+v %v", snapshot, err)
	}
}

func TestQueueGateEnclosesHeadPublicationAfterSlowInputs(t *testing.T) {
	inside := false
	order := []string{}
	o := testOptions(nil)
	o.Admission = func() AdmissionRule {
		if inside {
			t.Fatal("slow admission read inside gate")
		}
		order = append(order, "inputs")
		return nil
	}
	o.Grants = func() map[int]Grant {
		if inside {
			t.Fatal("grant read inside gate")
		}
		return nil
	}
	// Head publication is observed inside the gate, before it is released.
	dir := t.TempDir()
	o.Gate = func(resource func() (bool, error)) (bool, error) {
		order = append(order, "gate")
		inside = true
		defer func() { inside = false }()
		admitted, err := resource()
		if admitted {
			data, e := os.ReadFile(filepath.Join(dir, "head.lock"))
			if e != nil || len(data) == 0 {
				t.Fatalf("head not published inside gate: %q %v", data, e)
			}
		}
		return admitted, err
	}
	l := openLine(t, dir, o)
	h := enqueue(t, l, 481)
	wantPos(t, l, h, 0)
	if len(order) != 2 || order[0] != "inputs" || order[1] != "gate" {
		t.Fatal(order)
	}
}

// Pause cannot observe a published head without the operation becoming active:
// both publications are inside the same real machine gate transaction.
func TestQueuePauseSerializesWithLandingHeadPublication(t *testing.T) {
	ctx := context.Background()
	store := &queuegate.Store{Root: t.TempDir()}
	op := queuegate.Operation{SchemaVersion: 1, ID: "merge-barrier", Workspace: "ariadne:3", Kinds: []string{"merge", "landing"}, Phase: "waiting", PID: os.Getpid(), PIDStarted: "fixture", StartedAt: time.Now(), Root: t.TempDir(), BootID: "fixture", Managed: true}
	lease, err := store.Enroll(ctx, op)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	published, release := make(chan struct{}), make(chan struct{})
	o := testOptions(nil)
	o.Gate = func(resource func() (bool, error)) (bool, error) {
		return lease.TryAdmit(ctx, func() (bool, error) {
			admitted, err := resource()
			if admitted {
				close(published)
				<-release
			}
			return admitted, err
		})
	}
	dir := t.TempDir()
	line := openLine(t, dir, o)
	ticket := Ticket{Issue: 481, Branch: "branch", QueueOperationID: op.ID, QueuePID: op.PID, QueuePIDStarted: op.PIDStarted, QueueRoot: store.Root}
	h, err := line.Enqueue(ticket)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Release()
	done := make(chan error, 1)
	go func() { done <- line.Wait(ctx, h) }()
	select {
	case <-published:
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("head not published")
	}
	headFile, err := os.Open(filepath.Join(dir, "head.lock"))
	if err != nil {
		close(release)
		t.Fatal(err)
	}
	head := heldTicket(headFile)
	headFile.Close()
	if head.QueueOperationID != op.ID || head.QueuePID != op.PID || head.QueuePIDStarted != op.PIDStarted || head.QueueRoot != store.Root {
		close(release)
		t.Fatalf("lost binding: %+v", head)
	}
	pauseDone := make(chan error, 1)
	go func() { pauseDone <- store.Pause(ctx, "drain", "publication barrier", false) }()
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := <-pauseDone; err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.Snapshot(ctx)
	if err != nil || snapshot.State != "paused" || len(snapshot.Active) != 2 {
		t.Fatalf("pause lost admitted work: %+v %v", snapshot, err)
	}
}
