package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/internal/jsonstrict"
)

// resumeResult describes only observed process facts; classification performs no IO.
type resumeResult struct {
	stdout, stderr, processError string
	started, normal, timedOut    bool
	exitCode                     int
	outputExceeded               bool
}

type resumeReceipt struct{ ID, Op, Target, Status, Code, Diagnostic string }

type resumeFailure struct {
	receipt resumeReceipt
	details string
}

func (e *resumeFailure) Error() string { return e.receipt.Status + ": " + e.details }
func (e *resumeFailure) failureKey() (resumeFailureKey, bool) {
	k := resumeFailureKey{Status: e.receipt.Status, Code: e.receipt.Code, Diagnostic: e.receipt.Diagnostic}
	return k, (k.Status == "refused" || k.Status == "failed") && (k.Code != "" || k.Diagnostic != "")
}

// classifyResumeResult requires a matching terminal receipt and consistent normal
// exit. A process error after possible admission cannot prove the request failed.
func classifyResumeResult(addr string, r resumeResult) error {
	details := fmt.Sprintf("stdout: %s; stderr: %s", orNone(strings.TrimSpace(r.stdout)), orNone(strings.TrimSpace(r.stderr)))
	if r.outputExceeded {
		details = "output exceeded capture limit; " + details
	}
	if !r.started {
		return fmt.Errorf("did not run: %s; %s", r.processError, details)
	}
	var receipt resumeReceipt
	valid := !r.outputExceeded && jsonstrict.RejectDuplicateKeys([]byte(r.stdout)) == nil && json.Unmarshal([]byte(r.stdout), &receipt) == nil && receipt.ID != "" && receipt.Op == "resume" && receipt.Target == addr
	if valid && r.normal && !r.timedOut {
		if r.exitCode == 0 && receipt.Status == "succeeded" && r.processError == "" {
			return nil
		}
		if r.exitCode > 0 && (receipt.Status == "refused" || receipt.Status == "failed") {
			return &resumeFailure{receipt: receipt, details: details}
		}
	}
	return fmt.Errorf("%w: no confirmed terminal outcome (%s); %s", errUncertain, orNone(r.processError), details)
}

// Each pipe is consumed fully while retaining bounded diagnostics. Overflow
// makes the receipt unusable; a valid truncated prefix cannot prove success.
type resumeCapture struct {
	bytes.Buffer
	limit    int
	overflow bool
}

func (b *resumeCapture) Write(p []byte) (int, error) {
	n := len(p)
	remaining := b.limit - b.Len()
	if n > remaining {
		b.overflow = true
		p = p[:remaining]
	}
	_, _ = b.Buffer.Write(p)
	return n, nil
}

func (c execResume) Resume(addr string) error {
	limit := orDuration(c.resume, resumeTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.bin, "--resume", addr, "--json")
	cmd.WaitDelay = time.Second
	stop := operatingCommandGroup(cmd)
	defer stop()
	stdout, stderr := resumeCapture{limit: 1 << 20}, resumeCapture{limit: 64 << 10}
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	r := resumeResult{stdout: stdout.String(), stderr: stderr.String(), started: cmd.Process != nil, timedOut: ctx.Err() != nil, exitCode: -1, outputExceeded: stdout.overflow || stderr.overflow}
	if err != nil {
		r.processError = err.Error()
	}
	if cmd.ProcessState != nil {
		r.normal = cmd.ProcessState.Exited()
		r.exitCode = cmd.ProcessState.ExitCode()
	}
	return classifyResumeResult(addr, r)
}
