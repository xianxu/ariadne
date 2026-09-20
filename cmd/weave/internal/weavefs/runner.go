package weavefs

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
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
}

// Run spawns argv[0] with argv[1:] as arguments, cwd = dir. An empty argv is a
// programmer error (the caller always supplies the marker path). A non-zero exit
// is wrapped so the failing dir is visible in the compile error.
func (r ExecRunner) Run(dir string, argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("run in %s: empty argv", dir)
	}
	ctx := r.Context
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
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
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %v in %s: %w", argv, dir, err)
	}
	return nil
}

// ensure ExecRunner satisfies Runner at compile time.
var _ Runner = ExecRunner{}
