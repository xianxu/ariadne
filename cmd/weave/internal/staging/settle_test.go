package staging

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// deadOwnerStage is a stage whose recorded owner has exited and been reaped,
// with a producer lease held through the returned file, as an orphaned
// producer of a killed owner holds it (#343).
func deadOwnerStage(t *testing.T) (destination, stage string, producer *os.File) {
	t.Helper()
	destination = filepath.Join(t.TempDir(), "generation")
	stage, err := New(destination)
	if err != nil {
		t.Fatal(err)
	}
	owner := exec.Command("true")
	if err := owner.Run(); err != nil {
		t.Fatal(err)
	}
	host, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	metadata, _ := json.Marshal(Owner{Version: 1, Destination: destination, Host: host, PID: owner.Process.Pid})
	if err := os.WriteFile(filepath.Join(stage, "owner.json"), metadata, 0600); err != nil {
		t.Fatal(err)
	}
	if producer, err = Lease(stage); err != nil {
		t.Fatal(err)
	}
	return destination, stage, producer
}

// An orphaned producer still exiting when the retry reclaims: its lease is
// released within the settle, so the stage is reclaimed rather than kept.
func TestReclaimWaitsForAnExitingProducer(t *testing.T) {
	destination, stage, producer := deadOwnerStage(t)
	released := time.AfterFunc(Settle/10, func() { producer.Close() })
	defer released.Stop()
	if err := Reclaim(destination); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(stage); !os.IsNotExist(err) {
		t.Fatalf("stage of an exiting producer survived reclaim: %v", err)
	}
}

// A producer still alive past the settle keeps its stage: reclaim waits the
// bound, then preserves it.
func TestReclaimPreservesALiveOrphanAfterTheSettle(t *testing.T) {
	destination, stage, producer := deadOwnerStage(t)
	defer producer.Close()
	start := time.Now()
	if err := Reclaim(destination); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed < Settle {
		t.Fatalf("reclaim gave up after %v, before the settle %v", elapsed, Settle)
	}
	if _, err := os.Stat(filepath.Join(stage, "owner.json")); err != nil {
		t.Fatalf("live producer's stage removed: %v", err)
	}
}
