package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

const (
	flakeUnitSHAEnv      = "SDLC_FLAKE_SHA"
	flakeUnitContractEnv = "SDLC_FLAKE_CONTRACT"
	flakeUnitResultsEnv  = "SDLC_FLAKE_RESULTS"
	flakeUnitNameEnv     = "SDLC_FLAKE_UNIT"
	flakeUnitMaxSummary  = 16 << 20
)

func newFlakeRunCmd() *cobra.Command {
	var unit string
	cmd := &cobra.Command{
		Use: "run --unit script:PATH|lua:PATH", Short: "Run one declared typed unit inside a managed local check",
		Long: `The suite's typed-unit adapter. Reads the exact tested commit's version-3
runner declaration and verifies the enclosing managed check's contract binding.
Script exit is observed by this wrapper; Plenary results use a private channel.
Child stdout is diagnostic only. Outside a managed check use sdlc local-check run;
ordinary make recipes should keep their direct-runner fallback.`,
		Args: cobra.NoArgs, SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runFlakeUnit(cmd.Context(), cmd.OutOrStdout(), unit, execGitRunner{})
		},
	}
	cmd.Flags().StringVar(&unit, "unit", "", "exact script or Lua file execution unit")
	return cmd
}

// runFlakeUnit is an inner adapter, never a process-group owner. The enclosing
// managed helper alone owns cancellation, group join, capacity and retirement.
func runFlakeUnit(ctx context.Context, out io.Writer, selector string, runner gitRunner) error {
	sha, contract := os.Getenv(flakeUnitSHAEnv), os.Getenv(flakeUnitContractEnv)
	if !resumeValidDigest(contract) || !suiteHex.MatchString(sha) || (len(sha) != 40 && len(sha) != 64) {
		return fmt.Errorf("flake run requires a managed local-check invocation; use sdlc local-check run")
	}
	unit, err := flakes.ParseUnit(selector)
	if err != nil {
		return err
	}
	gitRunner := newBoundedLandingGit(ctx, runner, 10*time.Second)
	git := func(args ...string) (string, error) {
		b, err := gitRunner.Git(args...)
		return strings.TrimSpace(string(b)), err
	}
	root, err := git("rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("unit checkout: %w", err)
	}
	head, err := git("rev-parse", "HEAD")
	if err != nil || head != sha {
		return fmt.Errorf("unit checkout does not match the managed check head")
	}
	decl, err := readLocalChecks(git, sha)
	if err != nil {
		return err
	}
	check, exists := decl.check(suiteCheckName)
	if !exists || decl.Version != 3 || check.FlakeRunners.ContractDigest() != contract {
		return fmt.Errorf("unit runner contract differs from the managed check; repair and rerun the exact check")
	}
	declared, exists := check.FlakeRunners[unit.Kind()]
	if !exists {
		return fmt.Errorf("suite declares no %s runner", unit.Kind())
	}
	argv, err := declared.Project(unit)
	if err != nil {
		return err
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return err
	}
	invocation := hex.EncodeToString(nonce[:])
	env := withoutFlakeUnitChannel(os.Environ())
	results := ""
	if unit.Kind() == flakes.KindLua {
		// TMPDIR belongs to the enclosing run's existing scratch lifetime.
		// This channel is output, never a standalone resume authority.
		dir, err := os.MkdirTemp("", "sdlc-unit-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)
		results = filepath.Join(dir, "results.json")
		f, err := os.OpenFile(results, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		env = append(env, flakeUnitNameEnv+"="+selector, flakeUnitResultsEnv+"="+results)
	}
	output := &unitDiagnosticWriter{out: out, lineStart: true}
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	child := exec.CommandContext(childCtx, argv[0], argv[1:]...)
	// Deliberately inherit the outer managed group. Configuring a new group
	// here would let the child escape cleanup when this wrapper dies (#556).
	child.Dir, child.Env, child.WaitDelay = root, env, 5*time.Second
	child.Stdout, child.Stderr = output, output
	if err := child.Start(); err != nil {
		return fmt.Errorf("unit runner did not start: %w", err)
	}
	event := flakes.UnitEvent{Version: 1, Invocation: invocation, Unit: selector, Event: "start"}
	if err := output.emit(event); err != nil {
		cancel()
		_ = child.Wait()
		return err
	}
	waitErr := child.Wait()
	if childCtx.Err() != nil {
		return childCtx.Err() // interruption cannot emit a completed PASS
	}
	exit := 0
	if waitErr != nil {
		var ended *exec.ExitError
		if !errors.As(waitErr, &ended) {
			return fmt.Errorf("unit output/completion unavailable: %w", waitErr)
		}
		exit = ended.ExitCode()
		if exit < 0 {
			exit = 255 // observed signal termination, never a passing unit
		}
	}
	complete, nativeErrors := true, 0
	var proofErr error
	if unit.Kind() == flakes.KindLua {
		data, err := readFlakeUnitSummary(results)
		var cases []flakes.CaseResult
		if err == nil {
			cases, nativeErrors, err = flakes.DecodeUnitSummary(data, selector)
		}
		if err != nil {
			complete, proofErr = false, err
		} else {
			for _, c := range cases {
				e := flakes.UnitEvent{Version: 1, Invocation: invocation, Unit: selector, Event: "case", Case: c.Name, Result: c.Result}
				if err := output.emit(e); err != nil {
					return err
				}
			}
		}
	}
	event.Event, event.Exit, event.Complete, event.Errors = "finish", &exit, complete, nativeErrors
	if err := output.emit(event); err != nil {
		return err
	}
	if proofErr != nil {
		return fmt.Errorf("Lua unit has no complete private result summary: %w", proofErr)
	}
	if nativeErrors != 0 {
		return fmt.Errorf("Lua unit reports %d native setup/hook errors", nativeErrors)
	}
	return waitErr
}

func withoutFlakeUnitChannel(env []string) []string {
	filtered := make([]string, 0, len(env))
	for _, value := range env {
		if !strings.HasPrefix(value, flakeUnitNameEnv+"=") && !strings.HasPrefix(value, flakeUnitResultsEnv+"=") {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func readFlakeUnitSummary(path string) ([]byte, error) {
	// Reuse the existing regular-file identity check; a replaced symlink or
	// special file is not the private result channel we prepared.
	f, err := regularResumeFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > flakeUnitMaxSummary {
		return nil, fmt.Errorf("invalid or oversized private unit channel")
	}
	data, err := io.ReadAll(io.LimitReader(f, flakeUnitMaxSummary+1))
	if len(data) > flakeUnitMaxSummary {
		return nil, fmt.Errorf("private unit channel exceeds 16 MiB")
	}
	return data, err
}

// Both child streams share one writer. Every diagnostic line, including a
// partial line split across writes, stays indented outside control-frame scope.
type unitDiagnosticWriter struct {
	mu        sync.Mutex
	out       io.Writer
	lineStart bool
}

func (w *unitDiagnosticWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	consumed := 0
	for len(data) != 0 {
		if w.lineStart {
			if _, err := io.WriteString(w.out, "    "); err != nil {
				return consumed, err
			}
			w.lineStart = false
		}
		end := len(data)
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			end = i + 1
		}
		n, err := w.out.Write(data[:end])
		consumed += n
		if err != nil {
			return consumed, err
		}
		if n != end {
			return consumed, io.ErrShortWrite
		}
		w.lineStart = data[end-1] == '\n'
		data = data[end:]
	}
	return consumed, nil
}

func (w *unitDiagnosticWriter) emit(event flakes.UnitEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.lineStart {
		if _, err := io.WriteString(w.out, "\n"); err != nil {
			return err
		}
	}
	data = append(data, '\n')
	n, err := w.out.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	w.lineStart = true
	return err
}
