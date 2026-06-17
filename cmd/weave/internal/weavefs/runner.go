package weavefs

import (
	"fmt"
	"os"
	"os/exec"
)

// runner.go is weave's process-exec seam — DELIBERATELY SEPARATE from FS, which
// stays filesystem-only by documented stance (fs.go). The dynamic-skill generate
// stage (#111) execs a package's tracked `.dynamic-skill` to regenerate its
// committed SKILL.md at compile time; that one bounded "run a package's marker"
// capability is the only exec weave does (#95 M5 retired the open-ended go.mod
// editor — cmd/weave otherwise carries zero os/exec). Injecting it as an
// interface lets the generate stage be unit-tested with a fake (no real binary).
type Runner interface {
	// Run executes argv with the working directory set to dir, streaming the
	// child's stdout/stderr to the parent's. A non-zero exit (or a spawn failure)
	// returns a non-nil error so the generate stage can fail the compile loudly —
	// a dynamic skill never fails silently.
	Run(dir string, argv []string) error
}

// ExecRunner is the production Runner: it wraps os/exec, sets cmd.Dir = dir (so a
// `.dynamic-skill` resolves its relative paths from the package dir), and inherits
// the parent's stdout/stderr (so the marker's diagnostics stream through). Its
// zero value is ready to use.
type ExecRunner struct{}

// Run spawns argv[0] with argv[1:] as arguments, cwd = dir. An empty argv is a
// programmer error (the caller always supplies the marker path). A non-zero exit
// is wrapped so the failing dir is visible in the compile error.
func (ExecRunner) Run(dir string, argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("run in %s: empty argv", dir)
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %v in %s: %w", argv, dir, err)
	}
	return nil
}

// ensure ExecRunner satisfies Runner at compile time.
var _ Runner = ExecRunner{}
