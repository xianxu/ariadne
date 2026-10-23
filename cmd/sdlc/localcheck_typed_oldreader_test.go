//go:build unix

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Opt-in accepts the frozen pre-#556 production binary, never a downloaded
// executable. The provider and Git transport are private fixture processes.
func TestTypedAttemptFrozenReaderScopesRefusal(t *testing.T) {
	old := os.Getenv("SDLC_PRE_TYPED_BIN")
	if old == "" {
		t.Skip("set SDLC_PRE_TYPED_BIN to the verified frozen pre-change binary")
	}
	for _, key := range []string{"HOME", "TMPDIR", "GOTMPDIR", "WF_TEST_SCRATCH", "WF_SDLC_QUEUE_ROOT", "XDG_STATE_HOME", "XDG_CACHE_HOME"} {
		t.Setenv(key, t.TempDir())
	}
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("COUCH_TREE", "")
	d, _, _, _ := localRunRepo(t, "unused")
	wireLocalCheckSignalCLI(t, d.Root)
	decl := localCheckDecl{Version: 1, Checks: []localCheck{{Name: "lint", Command: []string{"true"}, TimeoutMinutes: 1, Weight: 1}, {Name: "suite", Command: []string{"true"}, TimeoutMinutes: 1, Weight: 1}}}
	raw, _ := json.Marshal(decl)
	writeRepoFile(t, d.Root, localChecksPath, string(raw))
	git(t, d.Root, "add", "-A")
	git(t, d.Root, "commit", "-qm", "private legacy declarations")
	git(t, d.Root, "push", "-qu", "origin", "main")
	sha := strings.TrimSpace(git(t, d.Root, "rev-parse", "HEAD"))
	statuses := filepath.Join(t.TempDir(), "statuses")
	t.Setenv("SDLC_PRIVATE_GH_STATE", statuses)
	fake := `#!/usr/bin/env python3
import os,sys
a=sys.argv[1:]
if a[:2]==['pr','list']: sys.exit(0)
if a[:3]==['api','-X','POST'] and a[3].startswith('repos/fixture/queue/statuses/'):
 with open(os.environ['SDLC_PRIVATE_GH_STATE'],'a') as f: f.write(repr(a)+'\n')
 print('{}');sys.exit(0)
sys.exit('unexpected private gh request: '+repr(a))
`
	if err := os.WriteFile(filepath.Join(os.Getenv("SDLC_QUEUE_BIN"), "gh"), []byte(fake), 0700); err != nil {
		t.Fatal(err)
	}
	store := filepath.Join(d.Root, ".git", "sdlc", "test-stats")
	if err := os.MkdirAll(filepath.Join(store, "attempts"), 0700); err != nil {
		t.Fatal(err)
	}
	log := []byte("ok example.org/pkg 0.1s\n")
	for _, check := range []string{"lint", "suite"} {
		w, err := beginSuiteAttempt(store)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(w.LogPath, log, 0600); err != nil {
			t.Fatal(err)
		}
		a := storedAttemptFixture(w.ID)
		a.Repo = "fixture/queue"
		a.SHA = sha
		a.Check = check
		a.Exit = 0
		a.Failures = nil
		a.LogSHA256 = resumeRecordDigest(log)
		wall := 1.0
		a.WallSeconds = &wall
		if check == "suite" {
			a.Version = 4
			a.RunnerContractSHA256 = strings.Repeat("f", 64)
		}
		if _, err := publishSuiteAttempt(store, w, a); err != nil {
			t.Fatal(err)
		}
	}
	before := map[string][]byte{}
	entries, err := os.ReadDir(filepath.Join(store, "attempts"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		p := filepath.Join(store, "attempts", e.Name())
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		before[p] = b
	}
	run := func(check string, wantOK bool) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, old, "local-check", "republish", "--name", check, "--rev", sha)
		cmd.Dir = d.Root
		cmd.Env = append(os.Environ(), "PATH="+os.Getenv("SDLC_QUEUE_BIN")+string(os.PathListSeparator)+os.Getenv("PATH"))
		cmd.WaitDelay = 5 * time.Second
		out, err := cmd.CombinedOutput()
		if (err == nil) != wantOK {
			t.Fatalf("old %s success=%v, want %v: %v\n%s", check, err == nil, wantOK, err, out)
		}
		return string(out)
	}
	run("lint", true)
	posted, err := os.ReadFile(statuses)
	if err != nil {
		t.Fatal(err)
	}
	out := run("suite", false)
	if !strings.Contains(out, "needs a newer sdlc") {
		t.Fatalf("missing scoped upgrade diagnostic: %s", out)
	}
	after, err := os.ReadFile(statuses)
	if err != nil || !bytes.Equal(posted, after) {
		t.Fatalf("unsafe old-reader publication: %v\n%s", err, after)
	}
	for p, b := range before {
		got, err := os.ReadFile(p)
		if err != nil || !bytes.Equal(b, got) {
			t.Fatalf("old reader rewrote %s: %v", p, err)
		}
	}
}
