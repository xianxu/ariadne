package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

// runResumeBatch is the thin effect shell. The coordinator alone writes output
// and applies transitions; workers only return immutable outcomes. A stop ends
// new dispatch, not an already-admitted remote operation.
func runResumeBatch(ctx context.Context, out io.Writer, d deps, hub string, rows []resumeRow, limit int) error {
	b, err := newResumeBatch(rows, limit)
	if err != nil {
		return err
	}
	todo := b.untouched()
	fmt.Fprintf(out, "resuming %s; concurrency %d per invocation\n", orNone(strings.Join(todo, ", ")), limit)
	results := make(chan resumeBatchResult, min(limit, len(todo)))
	var calls sync.WaitGroup
	cancelCh := ctx.Done()
	observeCancel := func() {
		if cancelCh != nil && ctx.Err() != nil {
			advanceResumeBatch(b, resumeBatchEvent{kind: batchCancel, err: ctx.Err()})
			fmt.Fprintln(out, "caller canceled; joining admitted calls before returning")
			cancelCh = nil
		}
	}
	accept := func(result resumeBatchResult) {
		wasStopped := b.stopped()
		advanceResumeBatch(b, resumeBatchEvent{kind: batchResult, addr: result.Addr, err: result.Err})
		if b.stopped() && !wasStopped {
			fmt.Fprintf(out, "dispatch stopped; not attempted: %s\n", orNone(strings.Join(b.untouched(), ", ")))
		}
		reportResumeResult(out, d, result)
	}
	for !b.finished() {
		observeCancel()
		// Do not admit from a partial view of the already-ready completions.
		draining := true
		for draining {
			select {
			case result := <-results:
				accept(result)
			default:
				draining = false
			}
		}
		observeCancel()
		if b.finished() {
			break
		}
		if effects := advanceResumeBatch(b, resumeBatchEvent{kind: batchAdmit}); len(effects) != 0 {
			launch := effects[0]
			fmt.Fprintf(out, "resuming  %s; wave %d; waiting for Couch's terminal receipt\n", launch.Addr, launch.Wave)
			calls.Add(1)
			go func(launch resumeLaunch) {
				defer calls.Done()
				results <- resumeBatchResult{Addr: launch.Addr, Err: d.resume.Resume(launch.Addr)}
			}(launch)
			continue
		}
		if b.finished() {
			break
		}
		select {
		case result := <-results:
			accept(result)
		case <-cancelCh:
			observeCancel()
		}
	}
	calls.Wait() // a returned result is not yet proof its worker goroutine exited
	err = b.resultError()
	var blocked []string
	for _, row := range rows {
		if row.Class == classUnknown || row.Class == classAmbiguous {
			blocked = append(blocked, row.Addr)
		}
	}
	if len(blocked) > 0 {
		err = errors.Join(err, fmt.Errorf("unknown or ambiguous worker evidence: %s", strings.Join(blocked, ", ")))
	}
	if err != nil {
		fmt.Fprintf(out, "worker waves incomplete for %s; not attempted: %s\n", hub, orNone(strings.Join(b.untouched(), ", ")))
		fmt.Fprintf(out, "inspect unresolved evidence: %s; keep scheduled tasks held and report this incomplete handoff to the machine steward\n", orNone(strings.Join(blocked, ", ")))
		return err
	}
	fmt.Fprintf(out, "worker waves 1 and 2 complete for %s (explicit eligible set); report this handoff to the machine steward\n", hub)
	fmt.Fprintln(out, "keep scheduled tasks held: only the machine steward releases wave 3 after every hub's handoff or explicit operator go (pair#874); per-hub automation is a separate veto; this command did not inspect or change their state")
	return nil
}

func reportResumeResult(out io.Writer, d deps, result resumeBatchResult) {
	addr, err := result.Addr, result.Err
	if err == nil {
		fmt.Fprintf(out, "resumed   %s\n", addr)
		return
	}
	if !errors.Is(err, errUncertain) {
		fmt.Fprintf(out, "failed    %s: %v\n", addr, err)
		return
	}
	fmt.Fprintf(out, "uncertain %s: %v\n", addr, err)
	fmt.Fprintf(out, "inspect the request with couch --operation-status ID (from the diagnostics) and couch --peek %s before a manual resume or switcher action\n", addr)
	actors, aerr := readActors(d.actors)
	switch {
	case aerr != nil:
		fmt.Fprintf(out, "observation unavailable: %v\n", aerr)
	case actors[addr].Agent != "":
		fmt.Fprintf(out, "observed %s running; operation outcome remains uncertain\n", addr)
	default:
		fmt.Fprintf(out, "observed no actor at %s; operation outcome remains uncertain\n", addr)
	}
}
