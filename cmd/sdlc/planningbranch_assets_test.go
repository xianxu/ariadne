package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
)

func planningAssetFixture(t *testing.T, existing bool) (*trackerRepo, string, string, string) {
	t.Helper()
	cp, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cp: card}, map[string]string{detailPath: detail})
	data := []byte("evidence\nexact LF bytes\n")
	name, err := issueassets.Name(data, "bin")
	if err != nil {
		t.Fatal(err)
	}
	rel := "assets/000009/" + name
	asset := "workshop/issues/" + rel
	target := "000009-nine"
	if existing {
		r.git("switch", "-q", "-c", target)
		writeRepoFile(t, r.root, asset, string(data))
		appendDetail(t, r, detailPath, "[evidence]("+rel+")")
		r.git("add", "workshop/issues")
		r.git("commit", "-qm", "evidence")
		r.git("switch", "-q", "main")
	} else {
		peerAdd(t, r, asset, string(data))
		peerAdd(t, r, detailPath, detail+"\n[evidence]("+rel+")\n")
	}
	return r, detailPath, target, asset
}

func poisonAssetCheckout(t *testing.T, r *trackerRepo, asset, kind string) {
	t.Helper()
	if kind == "attributes" {
		writeRepoFile(t, r.root, ".git/info/attributes", asset+" text eol=crlf\n")
		return
	}
	hookDir := r.git("config", "core.hooksPath")
	// All fixture path components are fixed portable characters.
	script := "#!/bin/sh\nif test -f '" + asset + "'; then printf 'changed by checkout hook' > '" + asset + "'; fi\n"
	if err := os.WriteFile(filepath.Join(hookDir, "post-checkout"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
}

func TestAssetPlanningBranchVerifiesMaterialization(t *testing.T) {
	for _, existing := range []bool{false, true} {
		for _, kind := range []string{"attributes", "hook"} {
			label := "create/" + kind
			if existing {
				label = "existing/" + kind
			}
			t.Run(label, func(t *testing.T) {
				r, detail, target, asset := planningAssetFixture(t, existing)
				poisonAssetCheckout(t, r, asset, kind)
				before := r.git("rev-parse", "HEAD")
				env, err := openTracker(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				_, err = preparePlanningBranch(env, "000009", detail)
				if err == nil {
					t.Fatal("planning reported success after unsafe asset materialization")
				}
				if kind == "attributes" {
					if r.git("rev-parse", "HEAD") != before || r.git("branch", "--show-current") != "main" {
						t.Fatal("attribute refusal switched branch")
					}
					if _, e := os.Stat(filepath.Join(r.root, asset)); !os.IsNotExist(e) {
						t.Fatal("attribute refusal installed evidence")
					}
				} else {
					if !strings.Contains(err.Error(), "differs") || !strings.Contains(err.Error(), "restore") {
						t.Fatalf("hook failure lacks recovery: %v", err)
					}
					if r.git("branch", "--show-current") != target {
						t.Fatal("fixture hook did not run after branch switch")
					}
					if _, retryErr := preparePlanningBranch(env, "000009", detail); retryErr == nil {
						t.Fatal("retry accepted the unrepaired hook-corrupted asset")
					}
				}
			})
		}
	}
}

func TestAssetHandoffSwitchVerifiesMaterialization(t *testing.T) {
	for _, kind := range []string{"attributes", "hook"} {
		t.Run(kind, func(t *testing.T) {
			r, _, target, asset := planningAssetFixture(t, true)
			poisonAssetCheckout(t, r, asset, kind)
			before := r.git("rev-parse", "HEAD")
			env, err := openTracker(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			var diagnostics bytes.Buffer
			if switchClean(env, &diagnostics, target) {
				t.Fatal("handoff reported success after unsafe materialization")
			}
			if kind == "attributes" {
				if r.git("rev-parse", "HEAD") != before {
					t.Fatal("unsafe handoff changed HEAD")
				}
			} else if !strings.Contains(diagnostics.String(), "restore") {
				t.Fatalf("hook error lacks recovery: %s", diagnostics.String())
			}
			if kind == "hook" && switchClean(env, &diagnostics, target) {
				t.Fatal("handoff retry accepted the unrepaired hook-corrupted asset")
			}
		})
	}
}
