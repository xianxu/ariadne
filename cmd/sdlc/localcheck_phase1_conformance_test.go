//go:build unix

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// The subprocess enters production main, preserving the CLI's argument parsing,
// Git transport, disk evidence, registry reads, slot helper, and status publisher.
func TestLocalCheckPhase1ConformanceProcess(t *testing.T) {
	if os.Getenv("SDLC_PRIVATE_CONFORMANCE_CHILD") != "1" {
		return
	}
	if bin := os.Getenv("SDLC_QUEUE_BIN"); bin != "" {
		_ = os.Setenv("PATH", testfix.PrependPath(bin, os.Getenv("PATH")))
	}
	var args []string
	if err := json.Unmarshal([]byte(os.Getenv("SDLC_PRIVATE_CONFORMANCE_ARGS")), &args); err != nil {
		t.Fatal(err)
	}
	os.Args = append([]string{"sdlc"}, args...)
	main()
	os.Exit(0)
}

// Run explicitly with SDLC_PRIVATE_CONFORMANCE=1 go test ./cmd/sdlc
// -run '^TestLocalCheckPhase1Conformance$' -count=1 -v. No installation or network
// access is needed. t.TempDir owns all fixture artifacts and removes them on exit.
func TestLocalCheckPhase1Conformance(t *testing.T) {
	if os.Getenv("SDLC_PRIVATE_CONFORMANCE") != "1" {
		t.Skip("opt in with SDLC_PRIVATE_CONFORMANCE=1 (real private Git and Go subprocesses)")
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("TMPDIR", t.TempDir())
	t.Setenv("GOTMPDIR", t.TempDir())
	t.Setenv("WF_TEST_SCRATCH", t.TempDir())
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	t.Setenv("WF_TEST_LOCK_FILE", filepath.Join(t.TempDir(), "suite.lock"))
	t.Setenv("WF_TEST_SUITES", "1")
	t.Setenv("WF_TEST_LOCK", "")
	t.Setenv("WF_TEST_LOCK_HELD_BY", "")
	t.Setenv("WF_TEST_LOCK_TIMEOUT", "30")
	t.Setenv("WF_ISSUES_DIR", "workshop/issues")
	t.Setenv("GOTOOLCHAIN", "local")
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOSUMDB", "off")
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	d, _, _, _ := localRunRepo(t, "unused")
	write := func(rel, body string) {
		t.Helper()
		writeRepoFile(t, d.Root, rel, body)
	}
	helper, err := os.ReadFile("../../scripts/test-lock.py")
	if err != nil {
		t.Fatal(err)
	}
	write("scripts/test-lock.py", string(helper))
	write("go.mod", "module example.com/exact\n\ngo 1.22\n")
	marks := t.TempDir()
	t.Setenv("EXACT_RERUN_MARKS", marks)
	for _, pkg := range []string{"a", "b"} {
		selected, decoy := "TestA", "TestB"
		if pkg == "b" {
			selected, decoy = decoy, selected
		}
		write(pkg+"/exact_test.go", fmt.Sprintf(`package %s
import ("os"; "path/filepath"; "testing")
func %s(t *testing.T) {
 p := filepath.Join(os.Getenv("EXACT_RERUN_MARKS"), "%s-once")
 if _, err := os.Stat(p); err != nil { if err := os.WriteFile(p, nil, 0600); err != nil { t.Fatal(err) }; t.Fatal("registered once") }
 for _, name := range []string{"case(value)", "path/file", "spaces here", "λ", "", "trailing/", "double//slash"} { t.Run(name, func(t *testing.T) {}) }
 if err := os.WriteFile(filepath.Join(os.Getenv("EXACT_RERUN_MARKS"), "%s-rerun"), nil, 0600); err != nil { t.Fatal(err) }
}
func %s(t *testing.T) { if err := os.WriteFile(filepath.Join(os.Getenv("EXACT_RERUN_MARKS"), "%s-decoy"), nil, 0600); err != nil { t.Fatal(err) } }
`, pkg, selected, pkg, pkg, decoy, pkg))
	}
	git(t, d.Root, "add", "-A")
	git(t, d.Root, "commit", "-qm", "private conformance inputs")
	wireLocalCheckSignalCLI(t, d.Root)
	decl := localCheckDecl{Version: 1, Checks: []localCheck{
		{Name: "lint", Command: []string{"true"}, TimeoutMinutes: 1, Weight: 1},
		{Name: "smoke", Command: []string{"true"}, TimeoutMinutes: 1, Weight: 1},
		{Name: suiteCheckName, Command: []string{"sh", "-c", "go test -count=1 -json -run '^TestA$' ./a; go test -count=1 -json -run '^TestB$' ./b; exit 1"}, TimeoutMinutes: 5, Weight: 1},
	}}
	raw, err := json.Marshal(decl)
	if err != nil {
		t.Fatal(err)
	}
	write(localChecksPath, string(raw))
	git(t, d.Root, "add", "-A")
	git(t, d.Root, "commit", "-qm", "private check declarations")
	git(t, d.Root, "push", "-q", "origin", "main")
	git(t, d.Root, "branch", "--set-upstream-to=origin/main", "main")
	trunk, err := gitx.NewTrunkFileContext(context.Background(), d.Root, "origin", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	_, err = trunk.Bootstrap(map[string][]byte{
		tracker.ManifestPath:                    tracker.ManifestBytes(),
		"workshop/issue-cards/000001-flakes.md": []byte("---\nid: 000001\nstatus: open\nflake_tests: [example.com/exact/a.TestA, example.com/exact/b.TestB]\n---\n\n# Private flakes\n\n## Problem\nPrivate registered once failures.\n"),
	}, "private registry", func(gitx.BootstrapResult) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	trackerRoot := strings.TrimSpace(git(t, d.Root, "rev-parse", "refs/remotes/origin/issue-tracker"))
	write(tracker.CutoverMarkerPath, string(tracker.CutoverMarkerBytes(trackerRoot)))
	git(t, d.Root, "add", "-A")
	git(t, d.Root, "commit", "-qm", "private tracker cutover")
	git(t, d.Root, "push", "-q", "origin", "main")
	sha := strings.TrimSpace(git(t, d.Root, "rev-parse", "HEAD"))

	// This fake is stateful: it records every publication, and refuses any
	// unexpected gh command rather than falling through to the user's gh.
	ghState := filepath.Join(t.TempDir(), "statuses.jsonl")
	t.Setenv("SDLC_PRIVATE_GH_STATE", ghState)
	fakeGH := `#!/usr/bin/env python3
import json, os, sys
a = sys.argv[1:]
if a[:2] == ['pr', 'list']:
    sys.exit(0)
if a[:3] == ['api', '-X', 'POST'] and a[3].startswith('repos/fixture/queue/statuses/'):
    fields = dict(x.split('=', 1) for x in a[4:] if x != '-f')
    fields['endpoint'] = a[3]
    with open(os.environ['SDLC_PRIVATE_GH_STATE'], 'a') as f:
        f.write(json.dumps(fields) + '\n')
    print('{}')
    sys.exit(0)
sys.exit('unexpected private gh command: ' + repr(a))
`
	if err := os.WriteFile(filepath.Join(os.Getenv("SDLC_QUEUE_BIN"), "gh"), []byte(fakeGH), 0700); err != nil {
		t.Fatal(err)
	}
	cli := func(wantOK bool, args ...string) string {
		t.Helper()
		raw, _ := json.Marshal(args)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLocalCheckPhase1ConformanceProcess$")
		processgroup.Configure(cmd)
		cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
		cmd.Dir = d.Root
		cmd.WaitDelay = 5 * time.Second
		cmd.Env = append(os.Environ(), "SDLC_PRIVATE_CONFORMANCE_CHILD=1", "SDLC_PRIVATE_CONFORMANCE_ARGS="+string(raw))
		out, err := cmd.CombinedOutput()
		t.Logf("sdlc %s\n%s", strings.Join(args, " "), out)
		if (err == nil) != wantOK {
			t.Fatalf("CLI success=%v, want %v: %v\n%s", err == nil, wantOK, err, out)
		}
		return string(out)
	}
	cli(true, "local-check", "run", "--name", "lint,smoke")
	store := filepath.Join(d.Root, ".git", "sdlc", "test-stats")
	future := []byte(fmt.Sprintf(`{"version":4,"id":"d3","repo":"fixture/queue","sha":%q,"check":"smoke","future":{"opaque":true}}`, sha))
	futurePath := filepath.Join(store, "attempts", "d3.json")
	if err := os.WriteFile(futurePath, future, 0600); err != nil {
		t.Fatal(err)
	}
	cli(true, "local-check", "republish", "--name", "lint")
	before, err := os.ReadFile(ghState)
	if err != nil {
		t.Fatal(err)
	}
	out := cli(false, "local-check", "republish", "--name", "smoke")
	if !strings.Contains(out, "needs a newer sdlc") {
		t.Fatalf("related future evidence lacked upgrade refusal: %s", out)
	}
	after, err := os.ReadFile(ghState)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("refused proof published a status: %s %v", after, err)
	}
	cli(true, "local-check", "run", "--name", "suite")
	for _, pkg := range []string{"a", "b"} {
		for _, suffix := range []string{"once", "rerun"} {
			if _, err := os.Stat(filepath.Join(marks, pkg+"-"+suffix)); err != nil {
				t.Fatalf("selected %s %s did not execute: %v", pkg, suffix, err)
			}
		}
		if _, err := os.Stat(filepath.Join(marks, pkg+"-decoy")); !os.IsNotExist(err) {
			t.Fatalf("unrelated %s decoy executed: %v", pkg, err)
		}
	}
	snapshot, err := readSuiteEvidence(store, 100)
	if err != nil || len(snapshot.Warnings) != 0 || len(snapshot.Unsupported) != 1 {
		t.Fatalf("private evidence: %+v %v", snapshot, err)
	}
	var parent, child *SuiteAttempt
	count := 0
	for i := range snapshot.Attempts {
		a := &snapshot.Attempts[i]
		if a.Check == suiteCheckName {
			parent = a
			count++
		} else if a.Check == suiteRerunCheck {
			child = a
			count++
		}
	}
	if count != 2 || parent == nil || child == nil || child.RerunOf != parent.ID || parent.Exit != 1 || child.Exit != 0 || parent.Outcome != string(flakes.KnownFlakePassed) {
		t.Fatalf("want one retained original and one linked exact rerun: %+v", snapshot.Attempts)
	}
	for _, a := range []*SuiteAttempt{parent, child} {
		if _, err := os.Stat(suiteAttemptLogFile(store, *a)); err != nil {
			t.Fatalf("retained %s log missing: %v", a.Check, err)
		}
	}
	preserved, err := os.ReadFile(futurePath)
	if err != nil || !bytes.Equal(preserved, future) {
		t.Fatalf("opaque future record changed: %s %v", preserved, err)
	}
	statuses, err := os.ReadFile(ghState)
	if err != nil {
		t.Fatal(err)
	}
	var published []map[string]string
	for _, line := range bytes.Split(bytes.TrimSpace(statuses), []byte("\n")) {
		var status map[string]string
		if err := json.Unmarshal(line, &status); err != nil {
			t.Fatal(err)
		}
		published = append(published, status)
	}
	if len(published) != 4 || published[2]["context"] != "sdlc-local/lint" || published[3]["context"] != "sdlc-local/suite" || published[3]["state"] != "success" || !strings.Contains(published[3]["description"], "known flake") {
		t.Fatalf("unexpected status publications: %+v", published)
	}
	t.Log("PRIVATE conformance: unrelated saved proof usable; related future proof refused without publication; exact package reruns exclude decoys; one linked child and both logs retained")
}
