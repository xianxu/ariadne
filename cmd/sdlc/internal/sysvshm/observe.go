package sysvshm

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"time"
)

// Observer is injected at the run boundary; observations never authorize cleanup.
type Observer interface {
	Snapshot(context.Context) Snapshot
}

type Native struct {
	Platform string
	UID      int
	Owner    string
	Read     func(context.Context, string) ([]byte, error)
}

func NewNative() Native {
	uid := os.Geteuid()
	owner := strconv.Itoa(uid)
	if u, err := user.LookupId(owner); err == nil {
		owner = u.Username
	}
	return Native{Platform: runtime.GOOS, UID: uid, Owner: owner}
}

func (n Native) Snapshot(ctx context.Context) Snapshot {
	if n.Platform != "darwin" && n.Platform != "linux" {
		return Snapshot{Unavailable: "host SysV observation unsupported"}
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	read := n.Read
	if read == nil {
		read = n.read
	}
	data, err := read(ctx, n.Platform)
	if err != nil {
		return Snapshot{Unavailable: fmt.Sprintf("host SysV observation: %v", err)}
	}
	if n.Platform == "darwin" {
		return ParseDarwin(data, n.UID, n.Owner)
	}
	return ParseLinux(data, n.UID)
}

func (n Native) read(ctx context.Context, platform string) ([]byte, error) {
	if platform == "linux" {
		f, err := os.Open("/proc/sysvipc/shm")
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return io.ReadAll(io.LimitReader(f, MaxBytes+1))
	}
	cmd := exec.CommandContext(ctx, "ipcs", "-m", "-a", "-u", strconv.Itoa(n.UID))
	cmd.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	cmd.WaitDelay = 100 * time.Millisecond
	output := &boundedOutput{}
	cmd.Stdout, cmd.Stderr = output, output
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ipcs: %w", err)
	}
	return output.data, nil
}

type boundedOutput struct{ data []byte }

func (b *boundedOutput) Write(p []byte) (int, error) {
	if len(b.data)+len(p) > MaxBytes {
		return 0, fmt.Errorf("SysV output exceeds byte limit")
	}
	b.data = append(b.data, p...)
	return len(p), nil
}
