package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

//go:embed helptext/bench.md
var benchCLIHelp string

func NewBenchCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "bench", Short: "Prepare, observe and score an isolated model replay", Long: benchCLIHelp, Args: cobra.NoArgs}
	var taskPath, runDir string
	prepare := &cobra.Command{Use: "prepare", Short: "Create a sanitized replay and fresh private tracker", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if taskPath == "" || runDir == "" {
			return errors.New("bench prepare requires --task and --run-dir")
		}
		if err := requireBenchObserver(cmd.Context()); err != nil {
			return err
		}
		raw, err := readBenchInput(taskPath, bench.TaskRecordLimit)
		if err != nil {
			return err
		}
		task, err := bench.DecodeTask(raw)
		if err != nil {
			return err
		}
		if !filepath.IsAbs(runDir) || filepath.Clean(runDir) != runDir {
			return errors.New("bench run directory must be an absolute canonical path")
		}
		if _, err := os.Lstat(runDir); !errors.Is(err, os.ErrNotExist) {
			return errors.New("bench prepare requires a fresh absent run directory")
		}
		if err := sweepBenchRuns(cmd.Context(), filepath.Dir(runDir), benchCouchDeps{}); err != nil {
			return err
		}
		s, err := createBenchStore(runDir, task)
		if err != nil {
			return err
		}
		if err := prepareBenchRun(cmd.Context(), s, benchPrepareDeps{}); err != nil {
			return fmt.Errorf("bench preparation retained at %s: %w", s.root, err)
		}
		return printBenchStatus(cmd.OutOrStdout(), s, false)
	}}
	prepare.Flags().StringVar(&taskPath, "task", "", "trusted pinned task JSON")
	prepare.Flags().StringVar(&runDir, "run-dir", "", "new absolute run directory; existing paths refuse")
	var statusDir string
	var asJSON bool
	status := &cobra.Command{Use: "status", Short: "Read the retained run phase and next action", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if statusDir == "" {
			return errors.New("bench status requires --run-dir")
		}
		s, err := openBenchStore(statusDir)
		if err != nil {
			return err
		}
		return printBenchStatus(cmd.OutOrStdout(), s, asJSON)
	}}
	status.Flags().StringVar(&statusDir, "run-dir", "", "absolute retained run directory")
	status.Flags().BoolVar(&asJSON, "json", false, "structured phase, runtime and recovery projection")
	cmd.AddCommand(prepare, status)
	addBenchExecutionCommands(cmd)
	return cmd
}

func addBenchExecutionCommands(parent *cobra.Command) {
	var flags benchRunFlags
	run := benchObserverCommand("run", "Dispatch once and observe exact disposable actors", func(ctx context.Context, s *benchStore) error {
		return runBenchObserver(ctx, s, flags, benchObserverDeps{})
	})
	run.Flags().StringVar(&flags.Hub, "hub", "", "exact provisioned TL slot")
	run.Flags().StringVar(&flags.Worker, "worker", "", "exact provisioned worker slot")
	run.Flags().StringVar(&flags.Provision, "provision", "", "trusted exact launch-profile and hub-claim JSON")
	run.Flags().StringVar(&flags.Ledger, "ledger", "", "external append-only result JSONL")
	run.Flags().DurationVar(&flags.Budget, "budget", time.Hour, "run wall budget, at most 24h; retained on resume")
	score := benchObserverCommand("score", "Capture measurements and evaluate hidden acceptance after retirement", func(ctx context.Context, s *benchStore) error {
		return scoreBenchRun(ctx, s, benchCouchDeps{})
	})
	var ledger string
	record := benchObserverCommand("record", "Append the exact immutable terminal result to the ledger", func(ctx context.Context, s *benchStore) error {
		return recordBenchRun(ctx, s, ledger, benchCouchDeps{})
	})
	record.Flags().StringVar(&ledger, "ledger", "", "external ledger for an undispatched run; otherwise must match retained path")
	var retirement string
	cancel := benchObserverCommand("cancel", "Request stop or confirm explicit supported actor retirement", func(ctx context.Context, s *benchStore) error {
		return cancelBenchRun(ctx, s, retirement, benchCouchDeps{})
	})
	cancel.Flags().StringVar(&retirement, "retirement-evidence", "", "trusted park/archive and no-surviving-process JSON")
	clean := benchObserverCommand("clean", "Remove recorded disposable trees after a fresh retirement check", func(ctx context.Context, s *benchStore) error {
		return cleanBenchRun(ctx, s, benchCouchDeps{})
	})
	var runID, kind, reason, candidate string
	decision := &cobra.Command{Use: "decision", Short: "Record a decision from the exact bound TL conversation", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		id, err := recordBenchProposal(cmd.Context(), "decision", runID, kind, reason, "", benchCouchDeps{})
		if err == nil {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "bench decision retained:", id)
		}
		return err
	}}
	decision.Flags().StringVar(&runID, "run-id", "", "exact public run identity")
	decision.Flags().StringVar(&kind, "kind", "", "approve, replan, ruling or stop")
	decision.Flags().StringVar(&reason, "reason", "", "decision rationale")
	var finishID string
	finish := &cobra.Command{Use: "finish", Short: "Propose reviewed code for independent would-land verification", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		id, err := recordBenchProposal(cmd.Context(), "finish", finishID, "", "", candidate, benchCouchDeps{})
		if err == nil {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "bench candidate proposed:", id)
		}
		return err
	}}
	finish.Flags().StringVar(&finishID, "run-id", "", "exact public run identity")
	finish.Flags().StringVar(&candidate, "candidate", "", "full reviewed code commit; metadata head is not the candidate")
	parent.AddCommand(run, score, record, cancel, clean, decision, finish)
}

func benchObserverCommand(name, short string, effect func(context.Context, *benchStore) error) *cobra.Command {
	var runDir string
	cmd := &cobra.Command{Use: name, Short: short, Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if runDir == "" {
			return fmt.Errorf("bench %s requires --run-dir", name)
		}
		if err := requireBenchObserver(cmd.Context()); err != nil {
			return err
		}
		s, err := openBenchStore(runDir)
		if err != nil {
			return err
		}
		if err := effect(cmd.Context(), s); err != nil {
			return err
		}
		return printBenchStatus(cmd.OutOrStdout(), s, false)
	}}
	cmd.Flags().StringVar(&runDir, "run-dir", "", "absolute retained run directory")
	return cmd
}

func requireBenchObserver(ctx context.Context) error {
	root, err := gitx.RepoTopLevelContext(ctx)
	if err != nil {
		return nil
	}
	c, err := loadBenchContext(ctx, root)
	if err != nil {
		return err
	}
	if c != nil {
		return errors.New("bench observer effects must run outside the prepared agent checkouts")
	}
	return nil
}

func readBenchInput(path string, limit int) ([]byte, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	physical, err := filepath.EvalSymlinks(path)
	if err != nil || physical != path {
		return nil, errors.New("bench input must be an existing physical path without symlinks")
	}
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("bench input must be a regular file")
	}
	raw, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > limit {
		return nil, errors.New("bench input exceeds its bound")
	}
	return raw, nil
}

func printBenchStatus(out io.Writer, s *benchStore, asJSON bool) error {
	r, state, err := s.read()
	if err != nil {
		return err
	}
	next := "inspect retained evidence; no automatic redispatch or cleanup"
	switch state.Phase() {
	case bench.PhaseNew:
		next = "prepare a fresh run with a valid task"
	case bench.PhasePrepared:
		next = "provision exact disposable hub/worker actors, then run sdlc bench run"
		if r.Data["runtime"] != nil {
			next = "binding retained before dispatch; retire the prepared roots with sdlc bench cancel --retirement-evidence FILE, then prepare a fresh run"
		}
	case bench.PhaseDispatchUnknown:
		next = "inspect the exact dispatch receipt; never resend blindly"
	case bench.PhaseRunning, bench.PhaseProving:
		next = "retain the existing observer; wait for exact candidate proof or budget expiry"
	case bench.PhaseWouldLand, bench.PhaseReconciliation, bench.PhaseCancelling:
		next = "park/archive both exact actors and confirm their processes are gone; provide retirement evidence to sdlc bench cancel"
		if state.Phase() == bench.PhaseWouldLand && state.RuntimeStopped() {
			next = "sdlc bench score captures measurements and evaluates hidden acceptance"
		}
	case bench.PhaseStopped, bench.PhaseScored, bench.PhaseAcceptanceUnavailable:
		next = "collect retained evidence and record the terminal row with sdlc bench record"
		if state.Phase() == bench.PhaseStopped && r.Data["measurements"] == nil {
			next = "sdlc bench score captures terminal measurements without hidden execution"
		}
	case bench.PhaseScoring:
		next = "scoring outcome unconfirmed; inspect retained scorer logs and native worktree leases; no automatic retry or cleanup"
	case bench.PhaseRecording, bench.PhaseRecordUnknown:
		next = "sdlc bench record retries only the exact retained ledger row"
	case bench.PhaseRecorded:
		next = "sdlc bench clean removes only confirmed owned disposable trees"
	case bench.PhaseCleaned:
		next = "complete; ledger and compact evidence retained"
	}
	projection := struct {
		Version        int         `json:"version"`
		RunID          string      `json:"run_id"`
		Phase          bench.Phase `json:"phase"`
		Candidate      string      `json:"candidate"`
		RuntimeStopped bool        `json:"runtime_stopped"`
		Next           string      `json:"next"`
	}{1, r.ID, state.Phase(), state.Candidate(), state.RuntimeStopped(), next}
	if asJSON {
		return json.NewEncoder(out).Encode(projection)
	}
	_, err = fmt.Fprintf(out, "bench %s: %s (runtime stopped: %t)\nnext: %s\n", r.ID, state.Phase(), state.RuntimeStopped(), next)
	return err
}
