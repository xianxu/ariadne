package weavefs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
	"github.com/xianxu/ariadne/pkg/procgroup/procgrouptest"
)

// #361: when the producer group can no longer be signalled because its members
// are zombies or exiting (macOS killpg: EPERM), a holder may still be closing
// its descriptors. finishProducers must settle on the lease, not stop on the
// signal's answer. The group here is a real zombie-only group (the kernel's
// EPERM, measured); the lease holder is a separate process that exits shortly
// after, as an exiting producer releases fd 3 late. Before #361 finishProducers
// returned EPERM at once and the stage could not be removed.
func TestFinishProducersSettlesWhenTheGroupIsAlreadyDone(t *testing.T) {
	stage, err := staging.New(filepath.Join(t.TempDir(), "generation"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = staging.Remove(stage) })

	// The producer group: its only member has exited and is not reaped.
	pgid := procgrouptest.ZombieGroup(t).Process.Pid

	// A holder still closing its inherited lease: released 300ms from now.
	lease, err := staging.Lease(stage)
	if err != nil {
		t.Fatal(err)
	}
	holder := exec.Command("sh", "-c", "sleep 0.3")
	holder.ExtraFiles = []*os.File{lease}
	if err := holder.Start(); err != nil {
		t.Fatal(err)
	}
	lease.Close() // only the holder keeps the lease now
	holderDone := make(chan struct{})
	go func() { _ = holder.Wait(); close(holderDone) }()
	t.Cleanup(func() { _ = holder.Process.Kill(); <-holderDone })

	if err := finishProducers(stage, pgid); err != nil {
		t.Fatalf("finishProducers stopped instead of settling on the lease: %v", err)
	}
	if err := staging.Remove(stage); err != nil {
		t.Fatalf("stage not removable after a settled finish: %v", err)
	}
}
