package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

type netFetchFailureRunner struct {
	gitRunner
	shim string
}

func (r netFetchFailureRunner) GitInDir(dir string, args ...string) ([]byte, error) {
	if len(args) > 0 && args[0] == "fetch" {
		cmd := exec.Command(r.shim, args...)
		cmd.Dir = dir
		return runGitCmd(cmd)
	}
	return r.gitRunner.GitInDir(dir, args...)
}

func TestNetCheckPreparationFailureIsNotSuiteEvidence(t *testing.T) {
	dir := tempRepoWith(t, map[string]string{"Makefile": "test:\n\tfalse\n"})
	ci := newCIFakeGH()
	swapMergeDeps(t, &e2eGH{ciSeamGH: ciSeamGH{ci: ci}}, nil)
	shim := filepath.Join(t.TempDir(), "git")
	if err := os.WriteFile(shim, []byte("#!/bin/sh\nprintf '%s\\n' \"fatal: unable to access 'https://git.example/repo': Could not resolve host: git.example\" >&2\nexit 128\n"), 0700); err != nil {
		t.Fatal(err)
	}
	previous := mergeRunner
	mergeRunner = netFetchFailureRunner{gitRunner: previous, shim: shim}
	t.Cleanup(func() { mergeRunner = previous })
	err := runLocalCheckCmd(context.Background(), io.Discard, localCheckFlags{Names: []string{"suite"}, Rev: "HEAD"})
	var network *gitx.GitNetworkFailure
	if !errors.As(err, &network) {
		t.Fatalf("preparation lost infrastructure cause: %v", err)
	}
	if len(ci.statuses) != 0 || len(ci.comments) != 0 {
		t.Fatalf("preparation failure published test evidence: statuses=%v comments=%v", ci.statuses, ci.comments)
	}
	for _, name := range []string{"local-checks", "test-stats"} {
		p := filepath.Join(dir, ".git", "sdlc", name)
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("preparation minted test record %s: %v", p, err)
		}
	}
}
