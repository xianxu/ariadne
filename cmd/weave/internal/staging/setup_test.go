package staging

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupLeaseSerializesEnvironmentAndKeepsStableFile(t *testing.T) {
	env := t.TempDir()
	lease, err := AcquireSetup(env)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	before, err := os.Stat(filepath.Join(env, ".weave-setup.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if other, err := AcquireSetup(env); !errors.Is(err, ErrSetupInUse) || !strings.Contains(err.Error(), env) || !strings.Contains(err.Error(), "retry") {
		if other != nil {
			other.Close()
		}
		t.Fatalf("contention diagnostic: %v", err)
	}
	other, err := AcquireSetup(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	other.Close()
	if err := lease.Close(); err != nil {
		t.Fatal(err)
	}
	lease, err = AcquireSetup(env)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	after, err := os.Stat(filepath.Join(env, ".weave-setup.lock"))
	if err != nil || !os.SameFile(before, after) {
		t.Fatalf("setup lock replaced: %v", err)
	}
}

func TestSetupLeaseRejectsNonRegularFile(t *testing.T) {
	env := t.TempDir()
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(env, ".weave-setup.lock")); err != nil {
		t.Fatal(err)
	}
	if lease, err := AcquireSetup(env); err == nil {
		lease.Close()
		t.Fatal("accepted symlink lock")
	}
}
