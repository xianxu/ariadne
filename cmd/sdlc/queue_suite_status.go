package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/pkg/workspace"
)

const suiteQueueOutputLimit = 1 << 20

// These are transient observations from the supplying scheduler, not a second
// admission policy or part of queue status's frozen JSON snapshot.
// wire: the supplying helper may add observation fields before this binary upgrades.
type suiteQueueObservation struct {
	Suppliers map[int64]suiteQueueSupplier `json:"-"`

	Version            int                `json:"version"`
	Available          bool               `json:"available"`
	PolicySeconds      float64            `json:"policy_seconds"`
	ProtectedThrough   int64              `json:"protected_through"`
	PriorityOverridden bool               `json:"priority_overridden"`
	Reason             string             `json:"reason"`
	Tickets            []suiteQueueTicket `json:"tickets"`
}

// wire: ticket observations retain additive helper metadata across binary versions.
type suiteQueueTicket struct {
	Number             int64           `json:"number"`
	Position           int             `json:"position"`
	AgeSeconds         *float64        `json:"age_seconds"`
	Protected          bool            `json:"protected"`
	Granted            bool            `json:"granted"`
	PriorityOverridden bool            `json:"priority_overridden"`
	Reason             string          `json:"reason"`
	Record             json.RawMessage `json:"record"`
}

func decodeSuiteQueue(raw []byte) (suiteQueueObservation, error) {
	var observed suiteQueueObservation
	if len(raw) > suiteQueueOutputLimit {
		return observed, fmt.Errorf("suite inspection exceeds output bound")
	}
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return observed, err
	}
	if err := json.Unmarshal(raw, &observed); err != nil {
		return observed, err
	}
	if observed.Version != 1 || observed.PolicySeconds <= 0 || math.IsInf(observed.PolicySeconds, 0) || math.IsNaN(observed.PolicySeconds) || observed.Tickets == nil || len(observed.Tickets) > 4096 || observed.ProtectedThrough < 0 {
		return observed, fmt.Errorf("invalid suite inspection version, policy or ticket bound")
	}
	if !observed.Available {
		if observed.Reason == "" || len(observed.Tickets) != 0 {
			return observed, fmt.Errorf("invalid unavailable suite inspection")
		}
		return observed, nil
	}
	seen := map[int64]bool{}
	for i, ticket := range observed.Tickets {
		if ticket.Number <= 0 || seen[ticket.Number] || ticket.Position != i+1 || ticket.Reason == "" {
			return observed, fmt.Errorf("invalid suite inspection ticket identity or position")
		}
		seen[ticket.Number] = true
		if age := ticket.AgeSeconds; age != nil && (*age < 0 || math.IsInf(*age, 0) || math.IsNaN(*age)) {
			return observed, fmt.Errorf("invalid suite ticket age")
		}
		var record map[string]json.RawMessage
		if err := json.Unmarshal(ticket.Record, &record); err != nil || record == nil {
			return observed, fmt.Errorf("invalid suite ticket record")
		}
	}
	return observed, nil
}

// observeSuiteQueue resolves the same workflow supplier as make/local-check.
// The Python owner handles clocks, ticket liveness, grants and ordering. Its
// bounded inspect operation creates no queue resources and never takes slots.
// Supervised provider children inherit this group, so one owner collects the
// complete tree even when cancellation cannot run Python finally blocks.
func observeSuiteQueue(parent context.Context, root string) suiteQueueObservation {
	unavailable := func(err error) suiteQueueObservation { return suiteQueueObservation{Reason: err.Error()} }
	if err := parent.Err(); err != nil {
		return unavailable(err)
	}
	helper, err := workspace.SuiteHelper(root, "")
	if err != nil {
		return unavailable(fmt.Errorf("suite helper unavailable: %w", err))
	}
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	capture := &helperHelpCapture{cancel: cancel, limit: suiteQueueOutputLimit}
	stdout, stderr := &helperHelpWriter{capture: capture}, &helperHelpWriter{capture: capture}
	cmd := exec.CommandContext(ctx, "python3", helper, "inspect", "--supervised")
	cmd.Dir = root
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = 100 * time.Millisecond
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err = cmd.Run()
	_ = processgroup.Terminate(cmd, true)
	if parent.Err() != nil {
		return unavailable(parent.Err())
	}
	if capture.exceeded {
		return unavailable(fmt.Errorf("suite inspection exceeds output bound"))
	}
	if err != nil {
		return unavailable(suiteQueueInspectFailure(helper, err, stderr.buffer.Bytes()))
	}
	observed, err := decodeSuiteQueue(stdout.buffer.Bytes())
	if err != nil {
		return unavailable(fmt.Errorf("suite inspection invalid: %w", err))
	}
	return observed
}

// Old supplying schedulers may still support acquisition without the optional
// inspection protocol. Diagnose only explicit argparse incompatibility; other
// failures remain unavailable with their own bounded stderr, never "old".
func suiteQueueInspectFailure(helper string, err error, stderr []byte) error {
	const diagnosticLimit = 1024 // one status-line excerpt; capture itself stays capped at 1 MiB
	detail := strings.TrimSpace(string(stderr))
	if len(detail) > diagnosticLimit {
		detail = detail[:diagnosticLimit] + " [truncated]"
	}
	var exit *exec.ExitError
	unsupported := errors.As(err, &exit) && exit.ExitCode() == 2 &&
		(strings.Contains(detail, "error: argument cmd: invalid choice: 'inspect'") ||
			strings.Contains(detail, "error: unrecognized arguments: --supervised"))
	if unsupported {
		return fmt.Errorf("suite helper %s does not support queue inspection: %w; update the supplying ariadne checkout %s through its pin/refresh workflow before retrying; acquisition uses the separate acquire command; stderr: %s", helper, err, filepath.Dir(filepath.Dir(helper)), detail)
	}
	return fmt.Errorf("suite helper inspect unavailable at %s: %w; stderr: %s", helper, err, detail)
}

func renderSuiteQueue(out io.Writer, observed suiteQueueObservation) {
	if !observed.Available {
		fmt.Fprintf(out, "suite queue order unavailable: %q; retry: sdlc queue status\n", observed.Reason)
		return
	}
	compatibility := assessSuiteQueueCompatibility(slices.Values(observed.Tickets))
	if len(compatibility.Incompatible) != 0 {
		index := 0
		for _, ticket := range observed.Tickets {
			legacy, _ := suiteTicketLegacy(ticket.Record)
			if legacy {
				if supplier, ok := observed.Suppliers[ticket.Number]; ok {
					compatibility.Incompatible[index] += "; " + supplier.detail()
				}
				index++
			}
		}
		fmt.Fprintln(out, "suite queue mixed-policy / "+compatibility.detail())
		if compatibility.Granted {
			fmt.Fprintln(out, "  deadlock risk: a live granted waiter coexists with legacy/unknown ordering; TL must revoke the conflicting grant, then let old waiters drain")
			for _, hint := range compatibility.Revoke {
				fmt.Fprintln(out, "  recovery: "+hint)
			}
		}
	}
	fmt.Fprintf(out, "suite queue: GRANT_MAX_WAIT_SECONDS=%.0f (%.0f min grant overtaking limit); protected through ticket %d\n", observed.PolicySeconds, observed.PolicySeconds/60, observed.ProtectedThrough)
	if observed.PriorityOverridden {
		fmt.Fprintln(out, "  aging overrode grant priority; protected FIFO prefix runs first")
	}
	for _, ticket := range observed.Tickets {
		age := "unknown"
		if ticket.AgeSeconds != nil {
			age = fmt.Sprintf("%.0fs", *ticket.AgeSeconds)
		}
		var record struct {
			Repo     string `json:"repo"`
			Worktree string `json:"worktree"`
		}
		_ = json.Unmarshal(ticket.Record, &record)
		override := ""
		if ticket.PriorityOverridden {
			override = "; aging overrode grant priority"
		}
		fmt.Fprintf(out, "  ticket %d: position %d; age %s; %q; protected=%t; granted=%t%s; repo %q, worktree %q\n", ticket.Number, ticket.Position, age, ticket.Reason, ticket.Protected, ticket.Granted, override, record.Repo, record.Worktree)
	}
}

func defaultSuiteQueueObservation(ctx context.Context) suiteQueueObservation {
	root, err := os.Getwd()
	if err != nil {
		return suiteQueueObservation{Reason: err.Error()}
	}
	return annotateSuiteSuppliers(ctx, observeSuiteQueue(ctx, root), suiteSupplierIO{resolve: resolveQueueSupplier, read: readQueueSupplier})
}
