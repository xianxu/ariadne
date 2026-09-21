package staging

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInheritedLeaseOutlivesParentCopy(t *testing.T) {
	stage, err := New(filepath.Join(t.TempDir(), "generation"))
	if err != nil {
		t.Fatal(err)
	}
	lease, err := Lease(stage)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", "-c", "echo ready; read release")
	cmd.ExtraFiles = []*os.File{lease}
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	if err := lease.Close(); err != nil {
		t.Fatal(err)
	}
	scan := bufio.NewScanner(output)
	if !scan.Scan() || scan.Text() != "ready" {
		t.Fatal("child did not start")
	}
	if err := Remove(stage); !errors.Is(err, ErrInUse) {
		t.Fatalf("stage removed while descendant owns lease: %v", err)
	}
	if _, err := os.Stat(filepath.Join(stage, "owner.json")); err != nil {
		t.Fatal("ownership metadata lost", err)
	}
	_, _ = input.Write([]byte("release\n"))
	_ = input.Close()
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	if err := Remove(stage); err != nil {
		t.Fatal(err)
	}
}
func TestMissingLeaseNeverAuthorizesCleanup(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "generation")
	stage, err := New(dest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(stage, "lease")); err != nil {
		t.Fatal(err)
	}
	if err := Remove(stage); err == nil {
		t.Fatal("missing producer evidence permitted cleanup")
	}
	if _, err := os.Stat(filepath.Join(stage, "owner.json")); err != nil {
		t.Fatal(err)
	}
}
