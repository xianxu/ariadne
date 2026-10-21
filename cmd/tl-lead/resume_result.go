package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// resumeResult describes only observed process facts; classification performs no IO.
type resumeResult struct {
	stdout, stderr, processError string
	started, normal, timedOut    bool
	exitCode                     int
}

type resumeReceipt struct{ ID, Op, Target, Status string }

// classifyResumeResult requires a matching terminal receipt and consistent normal
// exit. A process error after possible admission cannot prove the request failed.
func classifyResumeResult(addr string, r resumeResult) error {
	details := fmt.Sprintf("stdout: %s; stderr: %s", orNone(strings.TrimSpace(r.stdout)), orNone(strings.TrimSpace(r.stderr)))
	if !r.started {
		return fmt.Errorf("did not run: %s; %s", r.processError, details)
	}
	var receipt resumeReceipt
	valid := json.Unmarshal([]byte(r.stdout), &receipt) == nil && receipt.ID != "" && receipt.Op == "resume" && receipt.Target == addr
	if valid && r.normal && !r.timedOut {
		if r.exitCode == 0 && receipt.Status == "succeeded" {
			return nil
		}
		if r.exitCode > 0 && (receipt.Status == "refused" || receipt.Status == "failed") {
			return fmt.Errorf("%s: %s", receipt.Status, details)
		}
	}
	return fmt.Errorf("%w: no confirmed terminal outcome (%s); %s", errUncertain, orNone(r.processError), details)
}

func (c execResume) Resume(addr string) error {
	limit := orDuration(c.resume, resumeTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.bin, "--resume", addr, "--json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	r := resumeResult{stdout: stdout.String(), stderr: stderr.String(), started: cmd.Process != nil, timedOut: ctx.Err() != nil, exitCode: -1}
	if err != nil {
		r.processError = err.Error()
	}
	if cmd.ProcessState != nil {
		r.normal = cmd.ProcessState.Exited()
		r.exitCode = cmd.ProcessState.ExitCode()
	}
	return classifyResumeResult(addr, r)
}
