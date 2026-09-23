package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
	"github.com/xianxu/ariadne/pkg/workspace"
)

type setupGitReader struct{ context context.Context }

func (g setupGitReader) GitInDir(dir string, args ...string) ([]byte, error) {
	// GitReader owns path framing: trailing whitespace can belong to a path.
	cmd := exec.CommandContext(g.context, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("git %v in %q: %w", args, dir, err)
	}
	return out, nil
}

// prepareSetup discovers policy once before effects. The caller holds the lease
// through final composition and cleanup; subprocesses inherit the same open file
// description so a killed caller cannot admit another setup over live writers.
func prepareSetup(ctx context.Context, root string, dryRun bool, out, errOut io.Writer) (acquire.Client, weavefs.ExecRunner, func() error, error) {
	client := acquire.Client{}
	runner := weavefs.ExecRunner{Context: ctx, Stdout: out, Stderr: errOut}
	closeSetup := func() error { return nil }
	env, err := workspace.DiscoverEnvironment(setupGitReader{ctx}, root)
	if err != nil {
		return client, runner, closeSetup, err
	}
	if env == nil {
		return client, runner, closeSetup, nil
	}
	client.Policy = &acquire.Policy{EnvironmentRoot: env.Root, HostRoot: env.Host.WorktreeRoot, HostCommonDir: env.Host.RepoIdentity}
	if !dryRun {
		lease, err := staging.AcquireSetup(env.Root)
		if err != nil {
			return client, runner, closeSetup, err
		}
		runner.ExtraFiles = []*os.File{lease}
		client.Git = acquire.ExecGit{ExtraFiles: runner.ExtraFiles}
		closeSetup = lease.Close
	}
	return client, runner, closeSetup, nil
}
