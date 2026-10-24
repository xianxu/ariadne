package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/pkg/workspace"
)

var weaveHelperRefresh = runInstalledHelperRefresh

// prepareSuiteHelper is the single new-suite boundary for ordinary, integrated,
// train, baseline and saved-resume launches. Proof-only reposts never reach it.
func (d localRunDeps) prepareSuiteHelper(ctx context.Context) (string, error) {
	refresh := d.refreshSuiteHelper
	if refresh == nil {
		refresh = weaveHelperRefresh
	}
	if err := refresh(ctx, d.callerRoot(), d.Stderr); err != nil {
		return "", err
	}
	return prepareSuiteHelper(ctx, d.Root, d.WorkflowAnchor)
}

// Optional refresh never substitutes a scheduler or waives its capability gate.
// Capture child output so an old weave emits one actionable line, not usage noise.
func runInstalledHelperRefresh(parent context.Context, root string, out io.Writer) error {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	helper, err := workspace.SuiteHelper(root, "")
	if err != nil {
		return parent.Err()
	} // existing preflight owns this refusal
	supplier := filepath.Dir(filepath.Dir(helper))
	env, envErr := workspace.DiscoverEnvironment(queueGitReader{ctx}, root)
	if envErr == nil && (env == nil || filepath.Clean(root) == filepath.Join(env.Root, "ariadne")) {
		return parent.Err()
	}
	if out == nil {
		out = io.Discard
	}
	warning := func(reason string) {
		fmt.Fprintf(out, "suite helper refresh skipped %q: %s; repair: inspect the private sibling and run weave refresh --suite-helper, then retry the original command\n", supplier, reason)
	}
	if envErr != nil {
		warning("numbered environment unavailable")
		return parent.Err()
	}
	cmd := exec.CommandContext(ctx, "weave", "refresh", "--suite-helper")
	cmd.Dir = root
	cmd.Env = weaveDependencyCheckEnv(cmd)
	capture := &helperHelpCapture{cancel: cancel}
	stdout, stderr := &helperHelpWriter{capture: capture}, &helperHelpWriter{capture: capture}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = time.Second
	err = cmd.Run()
	_ = processgroup.Terminate(cmd, true)
	if parent.Err() != nil {
		return parent.Err()
	}
	if err != nil || capture.exceeded || ctx.Err() != nil {
		warning("optional weave mode unavailable or its ending is uncertain")
		return nil
	}
	if line := strings.TrimSpace(stdout.buffer.String()); line != "" {
		if strings.ContainsAny(line, "\r\n") {
			warning("unexpected refresh response; inspect the supplier")
		} else {
			fmt.Fprintln(out, line)
		}
	}
	return nil
}
