//go:build (darwin && !ios) || linux

package sysvshm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// Explicit operator-cleared qualification only. Default suites never allocate
// kernel IPC. Cleanup is limited to this child's returned ID and creator PID.
func TestNativeOwnedSysVQualification(t *testing.T) {
	if os.Getenv("WF_SYSV_QUALIFY") != "1" {
		t.Skip("requires operator-cleared IPC capacity and WF_SYSV_QUALIFY=1")
	}
	if os.Getenv("WF_SYSV_QUALIFY_CHILD") == "1" {
		id, err := unix.SysvShmGet(unix.IPC_PRIVATE, 56, unix.IPC_CREAT|unix.IPC_EXCL|0600)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Fprintln(os.Stdout, id)
		os.Exit(0) // deliberately abandon only this allocation
	}
	native := NewNative()
	before := native.Snapshot(context.Background())
	if before.Unavailable != "" {
		t.Fatal(before.Unavailable)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeOwnedSysVQualification$")
	cmd.Env = append(os.Environ(), "WF_SYSV_QUALIFY_CHILD=1")
	raw, err := cmd.Output()
	if err != nil {
		t.Fatalf("owned allocation child: %v", err)
	}
	id, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil || id < 0 {
		t.Fatalf("invalid owned ID %q: %v", raw, err)
	}
	// Refuse removal unless the kernel independently confirms this exact child's
	// identity and current owner; never trust a broad before/after difference.
	var desc unix.SysvShmDesc
	if _, err := unix.SysvShmCtl(id, unix.IPC_STAT, &desc); err != nil {
		t.Fatal(err)
	}
	if int(desc.Cpid) != cmd.Process.Pid || int(desc.Perm.Uid) != os.Geteuid() {
		t.Fatalf("ID %d not owned by child %d; operator reconciliation required", id, cmd.Process.Pid)
	}
	t.Cleanup(func() {
		if _, err := unix.SysvShmCtl(id, unix.IPC_RMID, nil); err != nil {
			t.Errorf("remove test-owned segment %d: %v", id, err)
		}
	})
	r := Delta(before, native.Snapshot(context.Background()), os.Geteuid())
	for _, candidate := range r.Candidates {
		if candidate.Segment.ID == id && candidate.Segment.Creator == cmd.Process.Pid {
			t.Logf("observed owned abandoned segment id=%d creator=%d; exact-ID cleanup registered", id, cmd.Process.Pid)
			return
		}
	}
	t.Fatalf("owned abandoned segment %d missing: %+v", id, r)
}
