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

	"github.com/xianxu/ariadne/pkg/workspace"
)

func TestLocalCheckReclassifyPrivateRepository(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	identity, err := workspace.Resolve(execGitRunner{}, d.Root, "")
	if err != nil || identity.Address == nil {
		t.Fatalf("private repo identity: %+v %v", identity, err)
	}
	logDir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	store := suiteStoreRoot(logDir)
	source, a, raw := reclassificationFixture(t)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	fields["repo"], _ = json.Marshal(d.Repo)
	fields["sha"], _ = json.Marshal(sha)
	fields["slot"], _ = json.Marshal(identity.Repo + ":5")
	raw, err = json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"attempts", "failed"} {
		if err := os.MkdirAll(filepath.Join(store, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(store, "attempts", a.ID+".json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	log, err := os.ReadFile(filepath.Join(source, "failed", a.ID+".log"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store, "failed", a.ID+".log"), log, 0600); err != nil {
		t.Fatal(err)
	}
	oldRepo, oldGH := detectRepo, ghClient
	detectRepo = func() (string, error) { return d.Repo, nil }
	ghClient = gh
	t.Cleanup(func() { detectRepo, ghClient = oldRepo, oldGH })
	gh.statuses[sha] = map[string]string{localStatusContext(suiteCheckName): "failure"}
	t.Chdir(d.Root)
	var output bytes.Buffer
	cmd := buildRoot()
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"local-check", "reclassify", "--attempt", a.ID, "--reason", "cache repaired"})
	run := cmd.Execute
	if cli := os.Getenv("SDLC_RECLASSIFY_CLI"); cli != "" {
		if !filepath.IsAbs(cli) {
			t.Fatal("SDLC_RECLASSIFY_CLI must be absolute")
		}
		if out, err := d.Git("remote", "add", "origin", "https://github.com/"+d.Repo+".git"); err != nil {
			t.Fatalf("private origin: %s %v", out, err)
		}
		bin := t.TempDir()
		probe := filepath.Join(bin, "gh-called")
		t.Setenv("SDLC_RECLASSIFY_GH_PROBE", probe)
		if err := os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\nprintf called > \"$SDLC_RECLASSIFY_GH_PROBE\"\nexit 91\n"), 0700); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
		child := exec.Command(cli, "local-check", "reclassify", "--attempt", a.ID, "--reason", "cache repaired")
		child.Dir, child.Stdout, child.Stderr = d.Root, &output, &output
		run = func() error {
			err := child.Run()
			if _, probeErr := os.Stat(probe); !os.IsNotExist(probeErr) {
				t.Fatalf("real CLI attempted GitHub effect: %v", probeErr)
			}
			return err
		}
	}
	if err := run(); err != nil {
		t.Fatalf("native :0 correction: %v\n%s", err, &output)
	}
	if !strings.Contains(output.String(), "could-not-run") || gh.statuses[sha][localStatusContext(suiteCheckName)] != "failure" || len(gh.comments[9]) != 0 {
		t.Fatalf("status rewritten or missing recovery: %s %+v", &output, gh.statuses)
	}
	if !strings.Contains(output.String(), "--name suite --rev "+sha) {
		t.Fatalf("recovery does not name the exact recorded head/check: %s", &output)
	}
	if _, _, err := loadSavedAttempt(store, d.Repo, sha, suiteCheckName, "", ""); err == nil {
		t.Fatal("CLI correction became a saved result")
	}
	pass := localCheck{Name: suiteCheckName, Command: []string{"true"}, TimeoutMinutes: 1}
	if rec, err := runLocalCheck(context.Background(), d, pass, sha); err != nil || !rec.passed() || gh.statuses[sha][localStatusContext(suiteCheckName)] != "success" {
		t.Fatalf("genuine repaired execution failed: %+v %v", rec, err)
	}
	if rec, err := runLocalCheck(context.Background(), d, suiteFailing("o/r/pkg.TestReal"), sha); err != nil || rec.passed() {
		t.Fatalf("real code failure missing: %+v %v", rec, err)
	}
	if rec, err := runLocalCheck(context.Background(), d, pass, sha); err == nil || gh.statuses[sha][localStatusContext(suiteCheckName)] != "failure" {
		t.Fatalf("neutral correction explained unrelated code failure: %+v %v %+v", rec, err, gh.statuses)
	}
}
