package weavefs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
	"github.com/xianxu/ariadne/pkg/procgroup"
)

// Runner is the existing process seam shared by startup and generators.
type Runner interface {
	// Run executes argv with the working directory set to dir, streaming the
	// child's stdout/stderr to the parent's. A non-zero exit (or a spawn failure)
	// returns a non-nil error so the generate stage can fail the compile loudly —
	// a dynamic skill never fails silently.
	Run(dir string, argv []string) error
}

// ExecRunner runs a child with optional context, environment and streams.
// Its zero value inherits the parent environment and output.
type ExecRunner struct {
	Context context.Context
	Env     []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	// ExtraFiles are inherited by every payload and remain caller-owned. An
	// owned producer reserves fd 3 for its stage lease and appends these files.
	ExtraFiles []*os.File
}

// Run spawns argv[0] with argv[1:] as arguments, cwd = dir. An empty argv is a
// programmer error (the caller always supplies the marker path). A non-zero exit
// is wrapped so the failing dir is visible in the compile error.
func (r ExecRunner) Run(dir string, argv []string) error { return r.run(dir, argv, "") }

// OwnedRunner includes the producer lifetime boundary used by real processes and
// stateful synchronous fakes. A successful return proves no writer lease remains.
type OwnedRunner interface {
	RunOwned(dir string, argv []string, stage string) error
}

func (r ExecRunner) RunOwned(dir string, argv []string, stage string) error {
	if stage == "" {
		return fmt.Errorf("producer stage is required")
	}
	return r.run(dir, argv, stage)
}

func (r ExecRunner) run(dir string, argv []string, stage string) error {
	if len(argv) == 0 {
		return fmt.Errorf("run in %s: empty argv", dir)
	}
	ctx := r.Context
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return procgroup.Kill(cmd.Process.Pid, syscall.SIGKILL) }
	// Descendants retaining output pipes must not hold Wait open indefinitely.
	cmd.WaitDelay = time.Second
	var lease *os.File
	if stage != "" {
		var err error
		lease, err = staging.Lease(stage)
		if err != nil {
			return err
		}
		cmd.ExtraFiles = []*os.File{lease} // fd 3, inherited before payload starts
	}
	cmd.ExtraFiles = append(cmd.ExtraFiles, r.ExtraFiles...)
	cmd.Env = r.Env
	cmd.Stdin = r.Stdin
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if r.Stdout != nil {
		cmd.Stdout = r.Stdout
	}
	if r.Stderr != nil {
		cmd.Stderr = r.Stderr
	}
	err := cmd.Run()
	if lease != nil {
		err = errors.Join(err, lease.Close()) // never unlock an inherited lease
		if cmd.Process != nil {
			err = errors.Join(err, finishProducers(stage, cmd.Process.Pid))
		}
	}
	if err != nil {
		return fmt.Errorf("run %v in %s: %w", argv, dir, err)
	}
	return nil
}

// With the parent's descriptor closed, a busy lease proves a descendant remains.
// Under the marker contract it remains in this group, keeping the PGID reserved;
// only then is signalling after Wait safe from process-group ID reuse. Zombie
// processes hold no descriptors and therefore cannot block cleanup forever.
//
// A group that is already done (os.ErrProcessDone: empty, or on macOS only
// zombies and exiting members) still settles on the lease: an exiting holder
// keeps its descriptor for a moment after it can no longer be signalled (#361).
func finishProducers(stage string, pgid int) error {
	lease, err := staging.Exclusive(stage)
	if err == nil {
		return lease.Close()
	}
	if !errors.Is(err, staging.ErrInUse) {
		return err
	}
	if err := procgroup.Kill(pgid, syscall.SIGKILL); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	if lease, err = staging.AwaitExclusive(stage, staging.Settle); err != nil {
		return fmt.Errorf("producer lifetime remains uncertain; preserving stage %s: %w", stage, err)
	}
	return lease.Close()
}

// ensure ExecRunner satisfies Runner at compile time.
var _ Runner = ExecRunner{}

// InputRunner is the process boundary for commands that consume stdin.
type InputRunner interface {
	RunInput(dir string, argv []string, input string) error
}

func (r ExecRunner) RunInput(dir string, argv []string, input string) error {
	r.Stdin = strings.NewReader(input)
	return r.Run(dir, argv)
}
