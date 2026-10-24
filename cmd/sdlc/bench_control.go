package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

func recordBenchProposal(ctx context.Context, kind, runID, decision, reason, candidate string, d benchCouchDeps) (string, error) {
	if kind != "decision" && kind != "finish" {
		return "", errors.New("unsupported bench proposal kind")
	}
	if kind == "decision" {
		if decision != "approve" && decision != "replan" && decision != "ruling" && decision != "stop" {
			return "", errors.New("bench decision kind must be approve, replan, ruling or stop")
		}
		if strings.TrimSpace(reason) == "" || len(reason) > 4096 || !utf8.ValidString(reason) || strings.ContainsRune(reason, 0) {
			return "", errors.New("bench decision needs bounded UTF8 --reason")
		}
	} else if !benchCandidatePattern.MatchString(candidate) {
		return "", errors.New("bench finish needs full lowercase --candidate commit ID")
	}
	root, err := gitx.RepoTopLevelContext(ctx)
	if err != nil {
		return "", err
	}
	c, err := loadBenchContext(ctx, root)
	if err != nil {
		return "", err
	}
	if c == nil || root != c.HubRoot || runID != c.RunID {
		return "", errors.New("bench TL proposal requires the exact root-bound hub and --run-id")
	}
	raw, err := readBenchInput(filepath.Join(filepath.Dir(c.WorkRoot), "runtime.json"), 16<<10)
	if err != nil {
		return "", err
	}
	var runtime benchPublicRuntime
	if err := decodeBenchRequired(raw, &runtime, "version", "run_id", "hub", "worker"); err != nil {
		return "", err
	}
	if runtime.Version != 1 || runtime.RunID != runID || runtime.Hub.Root != c.HubRoot || runtime.Worker.Root != c.WorkRoot || runtime.Hub.Validate() != nil || runtime.Worker.Validate() != nil || runtime.Hub.Slot == runtime.Worker.Slot {
		return "", errors.New("bench public runtime projection is invalid")
	}
	h := runtime.Hub
	actual, err := probeBenchActor(ctx, d, h.Slot, h.Root, bench.AgentProfile{Agent: h.Agent, Model: h.Model, Effort: h.Effort})
	if err != nil {
		return "", err
	}
	if actual.Identity != h || os.Getenv("COUCH_TREE") != h.Root || os.Getenv("COUCH_THREAD_SCOPE") != actual.Scope || os.Getenv("COUCH_THREAD_TAG") != actual.Tag || os.Getenv("PAIR_LAUNCH_NONCE") != h.Generation || os.Getenv("PAIR_SESSION_NAME") != h.Session {
		return "", errors.New("bench TL invoker differs from exact provisioned conversation")
	}
	id, err := newBenchEventID()
	if err != nil {
		return "", err
	}
	e := benchAuditEvent{Version: 1, ID: id, RunID: runID, Kind: kind, At: time.Now().UTC().Format(time.RFC3339Nano), Root: h.Root, Slot: h.Slot, Session: h.Session, Generation: h.Generation, DecisionKind: decision, Reason: reason, Candidate: candidate}
	if err := appendBenchAudit(ctx, c, e); err != nil {
		return "", err
	}
	return id, nil
}

// acceptBenchRetirement consumes explicit operator proof inside the live
// session. It never reopens or takes over a crashed observer invocation.
func acceptBenchRetirement(ctx context.Context, s *benchSession, proofPath string, d benchCouchDeps) error {
	r, state, err := s.read()
	if err != nil {
		return err
	}
	runtime, err := benchRetirementSubject(r, state)
	if err != nil {
		return err
	}
	if proofPath == "" {
		return errors.New("explicit retirement evidence is required")
	}
	raw, err := readBenchInput(proofPath, benchRecordLimit)
	if err != nil {
		return err
	}
	var proof benchRetirement
	if err := decodeBenchRequired(raw, &proof, "version", "run_id", "hub", "worker", "confirmed_by", "confirmed_at", "evidence", "processes_gone", "survivors", "reviewer_sources"); err != nil {
		return err
	}
	if err := validateBenchRetirement(runtime, proof); err != nil {
		return err
	}
	if err := verifyBenchRootsVacant(ctx, d, runtime.Hub.Identity.Root, runtime.Worker.Identity.Root); err != nil {
		return err
	}
	if err := s.putData("retirement", json.RawMessage(raw)); err != nil {
		return err
	}
	if state.RuntimeStopped() {
		return nil
	}
	if state.Phase() == bench.PhasePrepared {
		_, effects, err := s.apply(bench.Event{ID: "cancel-before-dispatch", RunID: r.ID, Kind: bench.Stop, OperationID: "cancel-" + r.ID, Reason: "no dispatch authorized; roots explicitly retired"})
		if err != nil {
			return err
		}
		if len(effects) != 0 {
			return errors.New("undispatched cancellation unexpectedly authorized an external effect")
		}
		return nil
	}
	op := state.CancelOperation()
	if op == "" {
		op = "cancel-" + r.ID
		_, effects, err := s.apply(bench.Event{ID: op + "-begin", RunID: r.ID, Kind: bench.BeginCancel, OperationID: op, Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity, Reason: "explicit retirement evidence received"})
		if err != nil {
			return err
		}
		if len(effects) != 1 || effects[0].Kind != bench.Cancel {
			return errors.New("retirement reconciliation effect not authorized")
		}
	}
	_, _, err = s.apply(bench.Event{ID: op + "-retired", RunID: r.ID, Kind: bench.CancelConfirmed, OperationID: op, Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity, Reason: fmt.Sprintf("explicit retirement by %s at %s", proof.ConfirmedBy, proof.ConfirmedAt)})
	return err
}
