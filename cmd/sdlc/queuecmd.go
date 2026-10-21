package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
	"github.com/xianxu/ariadne/pkg/workspace"
)

type queueCommandDeps struct {
	store     func() (*queuegate.Store, error)
	authority landingDeps
	legacy    func(context.Context, []queuegate.Operation) ([]queuegate.Activity, []string)
}

// queueStateRoot is also the Python helper's resolution contract. An explicit
// private machine-lock fixture must not accidentally consult production owners.
func queueStateRoot() (string, error) {
	root := os.Getenv("WF_SDLC_QUEUE_ROOT")
	if root == "" {
		if lock := os.Getenv("WF_TEST_LOCK_FILE"); lock != "" {
			root = lock + ".queue-state"
		}
	}
	if root == "" {
		state := os.Getenv("XDG_STATE_HOME")
		if state == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			state = filepath.Join(home, ".local", "state")
		}
		root = filepath.Join(state, "ariadne", "queue")
	}
	return filepath.Abs(root)
}

func defaultQueueCommandDeps() queueCommandDeps {
	return queueCommandDeps{
		store: func() (*queuegate.Store, error) {
			root, err := queueStateRoot()
			if err != nil {
				return nil, err
			}
			return &queuegate.Store{Root: root, BootID: queueBootID}, nil
		},
		authority: defaultLandingDeps(),
		legacy: func(ctx context.Context, known []queuegate.Operation) ([]queuegate.Activity, []string) {
			fleetRoot := os.Getenv("WF_SDLC_QUEUE_FLEET_ROOT")
			lock := os.Getenv("WF_TEST_LOCK_FILE")
			private := fleetRoot != "" && lock != "" && os.Getenv("WF_SDLC_QUEUE_ROOT") != ""
			if fleetRoot == "" {
				v, err := workspace.NormalizeVantage(queueGitReader{ctx}, ".")
				if err != nil {
					return nil, []string{"legacy fleet topology unavailable: " + err.Error()}
				}
				fleetRoot = v.FleetRoot
			}
			if lock == "" {
				state := os.Getenv("XDG_STATE_HOME")
				if state == "" {
					home, err := os.UserHomeDir()
					if err != nil {
						return nil, []string{err.Error()}
					}
					state = filepath.Join(home, ".local", "state")
				}
				lock = filepath.Join(state, "ariadne", "test-suite.lock")
			}
			slots := 2
			if raw := os.Getenv("WF_TEST_SUITES"); raw != "" {
				n, err := strconv.Atoi(raw)
				if err != nil || n < 1 || n > 4096 {
					return nil, []string{"invalid machine suite slot count"}
				}
				slots = n
			}
			return observeLegacyQueue(ctx, queueLegacyOptions{FleetRoot: fleetRoot, LockBase: lock, Slots: slots, Private: private}, known)
		},
	}
}

type queueGitReader struct{ ctx context.Context }

func (g queueGitReader) GitInDir(dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(g.ctx, "git", args...)
	cmd.Dir = dir
	cmd.WaitDelay = 100 * time.Millisecond
	return cmd.Output()
}

func NewQueueCmd() *cobra.Command { return newQueueCmd(defaultQueueCommandDeps()) }

func newQueueCmd(d queueCommandDeps) *cobra.Command {
	cmd := &cobra.Command{Use: "queue", Short: "Pause, resume and inspect machine-wide test and landing admission", Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true}
	var owner, reason string
	var fast bool
	pause := &cobra.Command{Use: "pause", Short: "Add an owner-scoped hold (operator or :0 only)", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		return runQueueMutation(cmd.Context(), cmd.OutOrStdout(), d, owner, reason, fast, false)
	}}
	pause.Flags().StringVar(&owner, "owner", "", "required token identifying only this caller's pause")
	pause.Flags().StringVar(&reason, "reason", "", "opaque explanation for the pause")
	pause.Flags().BoolVar(&fast, "cancel-running", false, "request interruption of running suites, never record a test failure")
	var resumeOwner string
	resume := &cobra.Command{Use: "resume", Short: "Remove only the named hold; interrupted work does not restart", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		return runQueueMutation(cmd.Context(), cmd.OutOrStdout(), d, resumeOwner, "", false, true)
	}}
	resume.Flags().StringVar(&resumeOwner, "owner", "", "required token to remove; other owners remain")
	var asJSON bool
	status := &cobra.Command{Use: "status", Short: "Observe owners and admitted work; incomplete never means quiescent", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		return runQueueStatus(cmd.Context(), cmd.OutOrStdout(), d, asJSON)
	}}
	status.Flags().BoolVar(&asJSON, "json", false, "print only the frozen schema_version 1 JSON snapshot")
	cmd.AddCommand(pause, resume, status, newQueueReconcileCmd(d))
	return cmd
}

func runQueueMutation(parent context.Context, out io.Writer, d queueCommandDeps, owner, reason string, fast, resume bool) error {
	_, slot, _ := d.authority.caller()
	tree, _ := d.authority.couchTree()
	if err := landingline.Authorize(slot, tree); err != nil {
		return fmt.Errorf("queue mutation: %w", err)
	}
	if strings.TrimSpace(owner) == "" {
		return fmt.Errorf("--owner TOKEN is required")
	}
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()
	s, err := d.store()
	if err != nil {
		return err
	}
	if resume {
		err = s.Resume(ctx, owner)
	} else {
		err = s.Pause(ctx, owner, reason, fast)
	}
	if err != nil {
		return err
	}
	verb := "paused by"
	if resume {
		verb = "removed pause owner"
	}
	_, err = fmt.Fprintf(out, "queue: %s %q\n", verb, owner)
	return err
}

func queueUnknown(detail string) queuegate.Snapshot {
	return queuegate.Snapshot{SchemaVersion: 1, State: "unknown", Complete: false, PauseOwners: []queuegate.PauseOwner{}, Active: []queuegate.Activity{}, Detail: queueDetail(detail)}
}

func queueDetail(detail string) string {
	if len(detail) > queuegate.MaxText {
		detail = detail[:queuegate.MaxText]
		for !utf8.ValidString(detail) && len(detail) > 0 {
			detail = detail[:len(detail)-1]
		}
	}
	return detail
}

func runQueueStatus(parent context.Context, out io.Writer, d queueCommandDeps, asJSON bool) error {
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()
	var snapshot queuegate.Snapshot
	s, err := d.store()
	var known []queuegate.Operation
	if err == nil {
		snapshot, known, err = s.Observe(ctx)
	}
	if err != nil {
		snapshot = queueUnknown("queue observation unavailable: " + err.Error())
	}
	if err == nil && d.legacy != nil {
		active, details := d.legacy(ctx, known)
		snapshot.Active = append(snapshot.Active, active...)
		if len(details) > 0 {
			snapshot.Complete = false
			if snapshot.Detail != "" {
				details = append([]string{snapshot.Detail}, details...)
			}
			snapshot.Detail = queueDetail(strings.Join(details, "; "))
		}
	}
	if err := queuegate.ValidateSnapshot(snapshot); err != nil {
		snapshot = queueUnknown("queue observation invalid/incomplete: " + err.Error())
	}
	data, err := json.Marshal(snapshot)
	if err != nil || len(data) > queuegate.MaxJSON {
		data, _ = json.Marshal(queueUnknown("queue status exceeds output bound"))
	}
	if asJSON {
		_, err = fmt.Fprintf(out, "%s\n", data)
		return err
	}
	fmt.Fprintf(out, "queue: %s; complete=%t\n", snapshot.State, snapshot.Complete)
	for _, p := range snapshot.PauseOwners {
		fmt.Fprintf(out, "  owner %q: %q (cancel_running=%t)\n", p.Owner, p.Reason, p.CancelRunning)
	}
	for _, a := range snapshot.Active {
		fmt.Fprintf(out, "  %s %s: %s (%s)\n", a.Kind, a.ID, a.State, a.Workspace)
	}
	if snapshot.Detail != "" {
		fmt.Fprintf(out, "  %s\n", snapshot.Detail)
	}
	return nil
}
