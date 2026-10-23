package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

type benchProvision struct {
	Version           int                `json:"version"`
	RunID             string             `json:"run_id"`
	ProvidedBy        string             `json:"provided_by"`
	Evidence          string             `json:"evidence"`
	HubClaimConfirmed bool               `json:"hub_claim_confirmed"`
	Hub               bench.AgentProfile `json:"hub"`
	Worker            bench.AgentProfile `json:"worker"`
}
type benchRunFlags struct {
	Hub, Worker, Provision, Ledger string
	Budget                         time.Duration
}
type benchObserverDeps struct {
	Couch  benchCouchDeps
	Now    func() time.Time
	Wait   func(context.Context, time.Duration) error
	Verify func(context.Context, *benchStore, string) (bench.WouldLandFacts, error)
}
type benchRuntimeState struct {
	Version   int               `json:"version"`
	RunID     string            `json:"run_id"`
	Hub       benchRuntimeActor `json:"hub"`
	Worker    benchRuntimeActor `json:"worker"`
	Since     string            `json:"since"`
	Deadline  string            `json:"deadline"`
	Ledger    string            `json:"ledger"`
	Provision json.RawMessage   `json:"provision"`
}
type benchPublicRuntime struct {
	Version int                   `json:"version"`
	RunID   string                `json:"run_id"`
	Hub     bench.RuntimeIdentity `json:"hub"`
	Worker  bench.RuntimeIdentity `json:"worker"`
}
type benchRetirement struct {
	Version         int                       `json:"version"`
	RunID           string                    `json:"run_id"`
	Hub             bench.RuntimeIdentity     `json:"hub"`
	Worker          bench.RuntimeIdentity     `json:"worker"`
	ConfirmedBy     string                    `json:"confirmed_by"`
	ConfirmedAt     string                    `json:"confirmed_at"`
	Evidence        string                    `json:"evidence"`
	ProcessesGone   bool                      `json:"processes_gone"`
	Survivors       []string                  `json:"survivors"`
	ReviewerSources []benchSourceInput        `json:"reviewer_sources"`
	Autonomy        *benchAutonomyAttestation `json:"autonomy,omitempty"`
}

type benchAutonomyAttestation struct {
	Since           string `json:"since"`
	Until           string `json:"until"`
	NoOperatorInput bool   `json:"no_operator_input"`
	Evidence        string `json:"evidence"`
}

func decodeBenchRequired(raw []byte, target any, keys ...string) error {
	if len(raw) > benchRecordLimit || !utf8.Valid(raw) {
		return errors.New("invalid or oversized bench evidence")
	}
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return errors.New("bench evidence must be an object")
	}
	for _, key := range keys {
		value, present := fields[key]
		if !present || strings.TrimSpace(string(value)) == "null" {
			return fmt.Errorf("bench evidence missing %s", key)
		}
	}
	return json.Unmarshal(raw, target)
}
func benchProfileEqual(a, b bench.AgentProfile) bool {
	return a.Agent == b.Agent && a.Model == b.Model && a.Effort == b.Effort
}
func (d benchObserverDeps) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now().UTC()
}

func loadBenchRuntime(r benchRecord, state bench.Run) (benchRuntimeState, error) {
	var runtime benchRuntimeState
	if err := decodeBenchRequired(r.Data["runtime"], &runtime, "version", "run_id", "hub", "worker", "since", "deadline", "ledger", "provision"); err != nil {
		return runtime, err
	}
	h, w := state.Bindings()
	if runtime.Version != 1 || runtime.RunID != r.ID || runtime.Hub.Identity != h || runtime.Worker.Identity != w || h.Root != filepath.Join(r.Root, "public", "hub") || w.Root != filepath.Join(r.Root, "public", "work") {
		return runtime, errors.New("retained runtime differs from exact run bindings")
	}
	return runtime, nil
}

func benchRetirementSubject(r benchRecord, state bench.Run) (benchRuntimeState, error) {
	h, w := state.Bindings()
	if h == (bench.RuntimeIdentity{}) && w == (bench.RuntimeIdentity{}) && state.DispatchOperation() == "" {
		// These are roots to retire, not invented launched actor identities.
		return benchRuntimeState{Version: 1, RunID: r.ID, Since: r.CreatedAt, Hub: benchRuntimeActor{Identity: bench.RuntimeIdentity{Root: filepath.Join(r.Root, "public", "hub")}}, Worker: benchRuntimeActor{Identity: bench.RuntimeIdentity{Root: filepath.Join(r.Root, "public", "work")}}}, nil
	}
	return loadBenchRuntime(r, state)
}

func runBenchObserver(ctx context.Context, s *benchStore, f benchRunFlags, d benchObserverDeps) error {
	release, err := s.acquireRun()
	if err != nil {
		return err
	}
	defer release()
	r, state, err := s.read()
	if err != nil {
		return err
	}
	if state.Phase() == bench.PhaseDispatching || state.Phase() == bench.PhaseDispatchUnknown || state.Phase() == bench.PhaseRunning || state.Phase() == bench.PhaseProving {
		runtime, err := loadBenchRuntime(r, state)
		if err != nil {
			return err
		}
		var receiptID, message string
		if err := json.Unmarshal(r.Data["dispatch_receipt_id"], &receiptID); err != nil || receiptID == "" {
			return errors.New("bench dispatch has no confirmed receipt identity; inspect retained reply and original actors, never resend")
		}
		if err := json.Unmarshal(r.Data["dispatch_message"], &message); err != nil {
			return err
		}
		return watchBenchRun(ctx, s, runtime, d, receiptID, message)
	}
	if state.Phase() != bench.PhasePrepared {
		return fmt.Errorf("bench run retained at %s; inspect status and exact receipts, never redispatch", state.Phase())
	}
	if _, bound := r.Data["runtime"]; bound {
		return fmt.Errorf("bench binding retained before dispatch authorization; no send is repeated: retire the prepared roots with sdlc bench cancel --run-dir %s --retirement-evidence FILE, then prepare a fresh run", s.root)
	}
	if f.Budget <= 0 || f.Budget > 24*time.Hour || f.Hub == "" || f.Worker == "" || f.Hub == f.Worker || f.Provision == "" || f.Ledger == "" {
		return errors.New("bench run requires distinct exact --hub/--worker, --provision, external --ledger and --budget within 24h")
	}
	if !filepath.IsAbs(f.Ledger) || filepath.Clean(f.Ledger) != f.Ledger || benchPathWithin(r.Root, f.Ledger) || benchPathWithin(r.Task.SourceRepo, f.Ledger) {
		return errors.New("bench ledger must be an absolute physical path outside the run and source repository")
	}
	if err := preflightBenchLedger(ctx, f.Ledger); err != nil {
		return err
	}
	raw, err := readBenchInput(f.Provision, benchRecordLimit)
	if err != nil {
		return err
	}
	var provision benchProvision
	if err := decodeBenchRequired(raw, &provision, "version", "run_id", "provided_by", "evidence", "hub_claim_confirmed", "hub", "worker"); err != nil {
		return err
	}
	if provision.Version != 1 || provision.RunID != r.ID || strings.TrimSpace(provision.ProvidedBy) == "" || len(provision.ProvidedBy) > 256 || strings.TrimSpace(provision.Evidence) == "" || len(provision.Evidence) > 4096 || !provision.HubClaimConfirmed || !benchProfileEqual(provision.Hub, r.Task.TL) || !benchProfileEqual(provision.Worker, r.Task.Worker) {
		return errors.New("bench provisioning must confirm this run's exact launch profiles and hub claim")
	}
	work, hub := filepath.Join(r.Root, "public", "work"), filepath.Join(r.Root, "public", "hub")
	public, err := loadBenchContext(ctx, work)
	if err != nil || public == nil || public.RunID != r.ID {
		return errors.New("prepared bench context unavailable or changed")
	}
	claims, err := readBenchInput(filepath.Join(hub, ".hub"), 64<<10)
	if err != nil {
		return err
	}
	claimed := false
	for _, line := range strings.Split(string(claims), "\n") {
		if strings.TrimSpace(line) == f.Worker {
			claimed = true
		}
	}
	if !claimed {
		return errors.New("provisioned hub .hub must claim the exact worker slot")
	}
	h, err := probeBenchActor(ctx, d.Couch, f.Hub, hub, provision.Hub)
	if err != nil {
		return err
	}
	w, err := probeBenchActor(ctx, d.Couch, f.Worker, work, provision.Worker)
	if err != nil {
		return err
	}
	now := d.now()
	runtime := benchRuntimeState{Version: 1, RunID: r.ID, Hub: h, Worker: w, Since: now.Format(time.RFC3339Nano), Deadline: now.Add(f.Budget).Format(time.RFC3339Nano), Ledger: f.Ledger, Provision: raw}
	if err := s.putData("runtime", runtime); err != nil {
		return err
	}
	projection := benchPublicRuntime{Version: 1, RunID: r.ID, Hub: h.Identity, Worker: w.Identity}
	if err := writeBenchJSONExclusive(filepath.Join(r.Root, "public", "runtime.json"), projection); err != nil {
		return err
	}
	op := "dispatch-" + r.ID
	_, effects, err := s.apply(bench.Event{ID: op + "-begin", RunID: r.ID, Kind: bench.BeginDispatch, OperationID: op, Hub: h.Identity, Worker: w.Identity})
	if err != nil {
		return err
	}
	if len(effects) != 1 || effects[0].Kind != bench.Dispatch {
		return errors.New("bench dispatch effect not authorized")
	}
	message := fmt.Sprintf("Bench replay %s. Read tl-lead --help and tl-lead bench --help. Public task: %s. Exact prepared worker: %s at %s. Follow real claim, start-plan, change-code, full suite and SHIP close; never merge or code as TL. Record decisions with --run-id %s and propose the reviewed code candidate with tl-lead bench finish. Fixed provisioned models; no fallback. Operator input stops autonomous work. Budget ends %s. Hidden acceptance and scoring belong to observer.", r.ID, filepath.Join(r.Root, "public", "task.json"), f.Worker, work, r.ID, runtime.Deadline)
	if err := s.putData("dispatch_message", message); err != nil {
		return err
	}
	reply, sendErr := d.Couch.command(ctx, "--send-to", f.Hub, "--message", message, "--json")
	if len(reply) > 0 {
		if err := s.putData("dispatch_reply", json.RawMessage(reply)); err != nil {
			sendErr = errors.Join(sendErr, err)
		}
	}
	receipt, receiptErr := readBenchDelivery(reply, h, "", message)
	if receiptErr == nil {
		if err := s.putData("dispatch_receipt_id", receipt.Message.ID); err != nil {
			sendErr = errors.Join(sendErr, err)
		}
	}
	sendErr = errors.Join(sendErr, receiptErr)
	// A successful CLI reply is not proof of delivery. The receipt watcher is
	// the only path to DispatchConfirmed; a lost reply is never sent again.
	_, _, stateErr := s.apply(bench.Event{ID: op + "-unconfirmed", RunID: r.ID, Kind: bench.DispatchUnconfirmed, OperationID: op, Hub: h.Identity, Worker: w.Identity, Reason: "inspect retained dispatch receipt; no automatic resend"})
	if sendErr != nil {
		return errors.Join(fmt.Errorf("bench dispatch unconfirmed; inspect %s, never resend blindly: %w", s.root, sendErr), stateErr)
	}
	if stateErr != nil {
		return stateErr
	}
	return watchBenchRun(ctx, s, runtime, d, receipt.Message.ID, message)
}

func watchBenchRun(ctx context.Context, s *benchStore, runtime benchRuntimeState, d benchObserverDeps, receiptID, message string) error {
	deadline, err := time.Parse(time.RFC3339Nano, runtime.Deadline)
	if err != nil {
		return err
	}
	r, state, err := s.read()
	if err != nil {
		return err
	}
	id, err := newBenchEventID()
	if err != nil {
		return err
	}
	_, effects, err := s.apply(bench.Event{ID: "watch-" + id, RunID: r.ID, Kind: bench.Observe, OperationID: state.DispatchOperation(), Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity})
	if err != nil {
		return err
	}
	if len(effects) != 1 || effects[0].Kind != bench.Watch {
		return errors.New("bench observation effect not authorized")
	}
	for {
		r, state, err = s.read()
		if err != nil {
			return err
		}
		if ctx.Err() != nil || !d.now().Before(deadline) {
			return requestBenchRetirement(ctx, s, runtime, d, bench.Timeout, "observer budget ended or invocation cancelled")
		}
		if _, requested := r.Data["cancel_request"]; requested {
			return requestBenchRetirement(ctx, s, runtime, d, bench.Stop, "operator requested cancellation")
		}
		if err := verifyBenchActorsPresent(ctx, d.Couch, runtime); err != nil {
			return requestBenchRetirement(ctx, s, runtime, d, bench.RuntimeLost, "native actor changed or unavailable: "+err.Error())
		}
		if state.Phase() == bench.PhaseDispatching || state.Phase() == bench.PhaseDispatchUnknown {
			raw, err := d.Couch.command(ctx, "--message-status", receiptID, "--json")
			if err != nil {
				return fmt.Errorf("bench receipt unconfirmed; resume observation without resending: %w", err)
			}
			receipt, err := readBenchDelivery(raw, runtime.Hub, receiptID, message)
			if err != nil {
				return err
			}
			switch receipt.Status {
			case "submitted":
				if _, _, err := s.apply(bench.Event{ID: state.DispatchOperation() + "-confirmed", RunID: r.ID, Kind: bench.DispatchConfirmed, OperationID: state.DispatchOperation(), Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity}); err != nil {
					return err
				}
			case "expired", "not-dispatched":
				return fmt.Errorf("bench delivery %s: %s; retire this run, never resend automatically", receiptID, receipt.Status)
			}
		}
		_, state, err = s.read()
		if err != nil {
			return err
		}
		if state.Phase() == bench.PhaseRunning || state.Phase() == bench.PhaseProving {
			terminal, err := observeBenchProposals(ctx, s, runtime, d)
			if err != nil || terminal {
				return err
			}
		}
		if d.Wait != nil {
			err = d.Wait(ctx, 5*time.Second)
		} else {
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case <-timer.C:
			}
			timer.Stop()
		}
		if err != nil {
			return requestBenchRetirement(ctx, s, runtime, d, bench.Stop, "observer wait cancelled: "+err.Error())
		}
	}
}

func verifyBenchActorsPresent(ctx context.Context, d benchCouchDeps, runtime benchRuntimeState) error {
	actors, err := benchActors(ctx, d)
	if err != nil {
		return err
	}
	for _, expected := range []benchRuntimeActor{runtime.Hub, runtime.Worker} {
		found := 0
		for _, actual := range actors {
			if actual.Binding.Slot == expected.Identity.Slot {
				found++
				if !actual.Known || !actual.Supported || actual.Binding != expected.binding() {
					return errors.New("exact runtime incarnation no longer matches")
				}
			}
		}
		if found != 1 {
			return errors.New("exact runtime actor absent or ambiguous")
		}
	}
	return nil
}

func verifyBenchConversations(ctx context.Context, d benchCouchDeps, runtime benchRuntimeState) error {
	for _, expected := range []benchRuntimeActor{runtime.Hub, runtime.Worker} {
		p := bench.AgentProfile{Agent: expected.Identity.Agent, Model: expected.Identity.Model, Effort: expected.Identity.Effort}
		actual, err := probeBenchActor(ctx, d, expected.Identity.Slot, expected.Identity.Root, p)
		if err != nil {
			return err
		}
		if actual.Identity != expected.Identity || actual.binding() != expected.binding() || actual.NativeSession != expected.NativeSession {
			return errors.New("exact runtime conversation changed; prior candidate authority cannot be reused")
		}
	}
	return nil
}

func observeBenchProposals(ctx context.Context, s *benchStore, runtime benchRuntimeState, d benchObserverDeps) (bool, error) {
	public, err := loadBenchContext(ctx, runtime.Worker.Identity.Root)
	if err != nil || public == nil {
		return false, errors.New("bench public context changed while observing")
	}
	events, err := readBenchAudit(ctx, public)
	if err != nil {
		return false, err
	}
	r, state, err := s.read()
	if err != nil {
		return false, err
	}
	for _, proposal := range events {
		if proposal.Kind != "decision" && proposal.Kind != "finish" {
			continue
		}
		if proposal.RunID != r.ID || proposal.Root != runtime.Hub.Identity.Root || proposal.Slot != runtime.Hub.Identity.Slot || proposal.Session != runtime.Hub.Identity.Session || proposal.Generation != runtime.Hub.Identity.Generation {
			return false, errors.New("bench TL proposal differs from original runtime binding")
		}
		if proposal.Kind == "decision" {
			if proposal.DecisionKind == "stop" {
				return true, requestBenchRetirement(ctx, s, runtime, d, bench.Stop, "TL stopped: "+proposal.Reason)
			}
			continue
		}
		op := "proof-" + proposal.ID
		if err := verifyBenchConversations(ctx, d.Couch, runtime); err != nil {
			return true, requestBenchRetirement(ctx, s, runtime, d, bench.RuntimeLost, err.Error())
		}
		if state.Phase() == bench.PhaseRunning {
			_, effects, err := s.apply(bench.Event{ID: op + "-begin", RunID: r.ID, Kind: bench.ProposeWouldLand, OperationID: op, Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity, Facts: bench.WouldLandFacts{Candidate: proposal.Candidate}})
			if err != nil {
				return false, err
			}
			if len(effects) == 0 {
				continue
			}
			if len(effects) != 1 || effects[0].Kind != bench.CheckProof {
				return false, errors.New("bench proof effect not authorized")
			}
		} else {
			if state.ProofOperation() != op {
				continue
			}
		}
		verify := d.Verify
		if verify == nil {
			verify = verifyBenchWouldLand
		}
		deadline, err := time.Parse(time.RFC3339Nano, runtime.Deadline)
		if err != nil {
			return false, err
		}
		remaining := deadline.Sub(d.now())
		if remaining <= 0 {
			return true, requestBenchRetirement(ctx, s, runtime, d, bench.Timeout, "run budget expired before candidate verification")
		}
		if remaining > 2*time.Minute {
			remaining = 2 * time.Minute
		}
		proofCtx, cancel := context.WithTimeout(ctx, remaining)
		facts, proofErr := verify(proofCtx, s, proposal.Candidate)
		cancel()
		kind := bench.ProofAccepted
		reason := ""
		if proofErr != nil {
			kind = bench.ProofRejected
			reason = proofErr.Error()
			if len(reason) > 4096 {
				reason = truncate(reason, 1000)
			}
		}
		_, _, err = s.apply(bench.Event{ID: op + "-result", RunID: r.ID, Kind: kind, OperationID: op, Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity, Facts: facts, Reason: reason})
		if err != nil {
			return false, err
		}
		if proofErr == nil {
			return true, requestBenchRetirement(ctx, s, runtime, d, bench.BeginCancel, "reviewed candidate admitted; retire actors before scoring")
		}
		return false, fmt.Errorf("bench would-land proof refused (retained as a bench finding): %w; correct evidence and resume observation", proofErr)
	}
	return false, nil
}

func requestBenchRetirement(ctx context.Context, s *benchStore, runtime benchRuntimeState, d benchObserverDeps, kind bench.EventKind, reason string) error {
	r, state, err := s.read()
	if err != nil {
		return err
	}
	op := state.CancelOperation()
	if op == "" {
		op = "cancel-" + r.ID
		_, effects, err := s.apply(bench.Event{ID: op + "-begin", RunID: r.ID, Kind: kind, OperationID: op, Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity, Reason: reason})
		if err != nil {
			return err
		}
		if len(effects) != 1 || effects[0].Kind != bench.Cancel {
			return errors.New("bench cancellation effect not authorized")
		}
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		message := fmt.Sprintf("Stop work for bench %s, original TL generation %s. %s. Operator must park/archive both bound actors and confirm no surviving processes; no scoring until explicit retirement evidence.", r.ID, runtime.Hub.Identity.Generation, reason)
		reply, notifyErr := d.Couch.command(cleanupCtx, "--send-to", runtime.Hub.Identity.Slot, "--message", message, "--json")
		cancel()
		if len(reply) > 0 {
			if err := s.putData("stop_reply", json.RawMessage(reply)); err != nil {
				notifyErr = errors.Join(notifyErr, err)
			}
		}
		if notifyErr != nil {
			reason += "; stop request unconfirmed: " + notifyErr.Error()
		}
	}
	if len(reason) > 4096 {
		reason = truncate(reason, 1000)
	}
	_, _, err = s.apply(bench.Event{ID: op + "-unconfirmed", RunID: r.ID, Kind: bench.CancelUnconfirmed, OperationID: op, Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity, Reason: reason})
	return errors.Join(fmt.Errorf("bench needs reconciliation: park/archive the exact actors and confirm their processes are gone, then run sdlc bench cancel --run-dir %s --retirement-evidence FILE", s.root), err)
}

func benchPathWithin(root, path string) bool {
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}

// This checks explicit trusted retirement evidence AND a fresh native census.
// Absence from actors alone never proves that detached descendants are gone.
func verifyBenchRetirement(ctx context.Context, s *benchStore, d benchCouchDeps) error {
	r, state, err := s.read()
	if err != nil {
		return err
	}
	if !state.RuntimeStopped() {
		return errors.New("bench runtime retirement is not confirmed")
	}
	runtime, err := benchRetirementSubject(r, state)
	if err != nil {
		return err
	}
	var proof benchRetirement
	if err := decodeBenchRequired(r.Data["retirement"], &proof, "version", "run_id", "hub", "worker", "confirmed_by", "confirmed_at", "evidence", "processes_gone", "survivors", "reviewer_sources"); err != nil {
		return err
	}
	if err := validateBenchRetirement(runtime, proof); err != nil {
		return err
	}
	return verifyBenchRootsVacant(ctx, d, runtime.Hub.Identity.Root, runtime.Worker.Identity.Root)
}

func validateBenchRetirement(runtime benchRuntimeState, proof benchRetirement) error {
	if proof.Version != 1 || proof.RunID != runtime.RunID || proof.Hub != runtime.Hub.Identity || proof.Worker != runtime.Worker.Identity || !proof.ProcessesGone || len(proof.Survivors) != 0 || strings.TrimSpace(proof.ConfirmedBy) == "" || len(proof.ConfirmedBy) > 256 || strings.TrimSpace(proof.Evidence) == "" || len(proof.Evidence) > 4096 || len(proof.ReviewerSources) > 128 {
		return errors.New("retirement needs exact original actors and explicit trusted confirmation of no surviving processes")
	}
	at, err := time.Parse(time.RFC3339Nano, proof.ConfirmedAt)
	if err != nil {
		return err
	}
	since, err := time.Parse(time.RFC3339Nano, runtime.Since)
	if err != nil || at.Before(since) {
		return errors.New("retirement evidence predates this run")
	}
	return nil
}

func verifyBenchRootsVacant(ctx context.Context, d benchCouchDeps, roots ...string) error {
	actors, err := benchActors(ctx, d)
	if err != nil {
		return err
	}
	for _, actor := range actors {
		for _, root := range roots {
			if actor.Binding.Repository == filepath.Join(root, ".git") {
				return fmt.Errorf("bench root still has a Couch actor %s; explicit retirement required", actor.Binding.Slot)
			}
		}
	}
	return nil
}
