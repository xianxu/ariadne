package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssetChangeCodeBranchVerifiesMaterialization(t *testing.T) {
	for _, mode := range []string{"inplace-existing", "inplace-new", "worktree-existing", "worktree-new"} {
		for _, kind := range []string{"attributes", "hook"} {
			t.Run(mode+"/"+kind, func(t *testing.T) {
				r, _, target, asset := planningAssetFixture(t, true)
				if strings.HasSuffix(mode, "-new") {
					r.git("switch", "-q", target)
					target = "000010-new"
				}
				poisonAssetCheckout(t, r, asset, kind)
				before := r.git("rev-parse", "HEAD")
				run := func() error {
					var out, diag bytes.Buffer
					if strings.HasPrefix(mode, "worktree") {
						_, err := createWorktreeBranch(context.Background(), &out, &diag, target, execGitRunner{}, "workshop/issues")
						return err
					}
					_, err := createInPlaceBranch(context.Background(), &out, &diag, target, execGitRunner{}, "workshop/issues")
					return err
				}
				err := run()
				if err == nil {
					t.Fatal("change-code branch creation accepted converted or hook-corrupted bytes")
				}
				if kind == "attributes" {
					if r.git("rev-parse", "HEAD") != before {
						t.Fatal("attribute refusal changed HEAD")
					}
				} else {
					if !strings.Contains(err.Error(), "restore") {
						t.Fatalf("hook failure lacks recovery: %v", err)
					}
					if retryErr := run(); retryErr == nil {
						t.Fatal("retry shortcut accepted unrepaired evidence")
					}
				}
				if _, e := os.Stat(filepath.Join(r.root, ".goto")); !os.IsNotExist(e) {
					t.Fatal("failure announced a worktree through .goto")
				}
			})
		}
	}
}

type failingAssetMetadataRunner struct{ gitRunner }

func (r failingAssetMetadataRunner) Git(args ...string) ([]byte, error) {
	if len(args) > 0 && args[0] == "ls-tree" {
		return nil, errors.New("injected asset metadata failure")
	}
	return r.gitRunner.Git(args...)
}
func TestAssetChangeCodeBranchMetadataFailureRefusesBeforeEffects(t *testing.T) {
	hermeticRepo(t)
	for _, worktree := range []bool{false, true} {
		base := &captureRunner{currentBranch: "main", branchExists: false}
		r := failingAssetMetadataRunner{base}
		var out bytes.Buffer
		var err error
		if worktree {
			_, err = createWorktreeBranch(context.Background(), &out, &out, "new-branch", r, "workshop/issues")
		} else {
			_, err = createInPlaceBranch(context.Background(), &out, &out, "new-branch", r, "workshop/issues")
		}
		if err == nil || !strings.Contains(err.Error(), "restore a readable") {
			t.Fatalf("metadata error did not fail closed: %v", err)
		}
		if anyGitCallStartsWith(base.gitCalls, "checkout") || anyGitCallStartsWith(base.gitCalls, "worktree", "add") || len(base.mkdirs) > 0 || len(base.writes) > 0 {
			t.Fatal("metadata failure caused filesystem or branch effects")
		}
	}
}
