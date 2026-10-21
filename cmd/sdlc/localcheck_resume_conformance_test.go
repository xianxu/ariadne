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
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

type privateResumeFixture struct {
	t                                 *testing.T
	root, sha, store, marks, statuses string
	registry                          *tracker.Repository
	parent                            SuiteAttempt
	parentRaw, parentLog              []byte
}

// This reuses the production-main subprocess entry from phase-1 conformance.
// All stores, transports, registrations, barriers and process IDs are private.
func newPrivateResumeFixture(t *testing.T) *privateResumeFixture {
	t.Helper()
	for _, key := range []string{"HOME", "TMPDIR", "GOTMPDIR", "WF_TEST_SCRATCH", "WF_SDLC_QUEUE_ROOT", "XDG_STATE_HOME", "XDG_CACHE_HOME"} {
		t.Setenv(key, t.TempDir())
	}
	for key, val := range map[string]string{
		"GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": os.DevNull, "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off", "GOWORK": "off", "GOFLAGS": "",
		"WF_TEST_LOCK_FILE": filepath.Join(t.TempDir(), "suite.lock"), "WF_TEST_SUITES": "1", "WF_TEST_LOCK": "", "WF_TEST_LOCK_HELD_BY": "", "WF_TEST_LOCK_TIMEOUT": "30", "WF_ISSUES_DIR": "workshop/issues", "COUCH_TREE": "",
	} {
		t.Setenv(key, val)
	}
	d, _, _, _ := localRunRepo(t, "unused")
	f := &privateResumeFixture{t: t, root: d.Root, marks: t.TempDir(), statuses: filepath.Join(t.TempDir(), "statuses.jsonl")}
	f.store = filepath.Join(f.root, ".git", "sdlc", "test-stats")
	t.Setenv("PRIVATE_RESUME_MARKS", f.marks)
	write := func(rel, body string) { writeRepoFile(t, f.root, rel, body) }
	helper, err := os.ReadFile("../../scripts/test-lock.py")
	if err != nil {
		t.Fatal(err)
	}
	write("scripts/test-lock.py", string(helper))
	write("go.mod", "module example.com/private-resume\n\ngo 1.22\n")
	for _, pkg := range []string{"a", "b"} {
		selected, decoy := "TestA", "TestB"
		if pkg == "b" {
			selected, decoy = decoy, selected
		}
		write(pkg+"/resume_test.go", fmt.Sprintf(`package %s
import ("fmt"; "os"; "path/filepath"; "syscall"; "testing"; "time")
func %s(t *testing.T) {
 root := os.Getenv("PRIVATE_RESUME_MARKS")
 file, err := os.OpenFile(filepath.Join(root, "%s-count"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); if err != nil { t.Fatal(err) }; fmt.Fprintln(file, "executed"); file.Close()
 once := filepath.Join(root, "%s-once")
 if _, err := os.Stat(once); os.IsNotExist(err) { if err := os.WriteFile(once, nil, 0600); err != nil { t.Fatal(err) }; t.Fatal("private failure before registration") }
 if _, err := os.Stat(filepath.Join(root, "block")); err == nil {
  pgid, err := syscall.Getpgid(0); if err != nil { t.Fatal(err) }
  if err := os.WriteFile(filepath.Join(root, "started"), []byte(fmt.Sprintf("%%d %%d", os.Getpid(), pgid)), 0600); err != nil { t.Fatal(err) }
  for { time.Sleep(time.Second) }
 }
}
func %s(t *testing.T) { if err := os.WriteFile(filepath.Join(os.Getenv("PRIVATE_RESUME_MARKS"), "%s-decoy"), nil, 0600); err != nil { t.Fatal(err) } }
`, pkg, selected, pkg, pkg, decoy, pkg))
	}
	git(t, f.root, "add", "-A")
	git(t, f.root, "commit", "-qm", "private resume input")
	wireLocalCheckSignalCLI(t, f.root)
	decl := localCheckDecl{Version: 1, Checks: []localCheck{{Name: suiteCheckName, Command: []string{"sh", "-c", `printf 'suite\n' >> "$PRIVATE_RESUME_MARKS/suite-count"; go test -count=1 -json -run '^TestA$' ./a; a=$?; go test -count=1 -json -run '^TestB$' ./b; b=$?; test "$a" = 0 && test "$b" = 0`}, TimeoutMinutes: 5, Weight: 1}}}
	raw, err := json.Marshal(decl)
	if err != nil {
		t.Fatal(err)
	}
	write(localChecksPath, string(raw))
	git(t, f.root, "add", "-A")
	git(t, f.root, "commit", "-qm", "private resume declaration")
	git(t, f.root, "push", "-qu", "origin", "main")
	tf, err := gitx.NewTrunkFileContext(context.Background(), f.root, "origin", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	_, err = tf.Bootstrap(map[string][]byte{tracker.ManifestPath: tracker.ManifestBytes(), "workshop/issue-cards/000001-flakes.md": []byte("---\nid: 000001\nstatus: open\n---\n\n# Private flakes\n\n## Problem\nPrivately registered failures.\n")}, "private empty registry", func(gitx.BootstrapResult) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	trackerRoot := strings.TrimSpace(git(t, f.root, "rev-parse", "refs/remotes/origin/issue-tracker"))
	write(tracker.CutoverMarkerPath, string(tracker.CutoverMarkerBytes(trackerRoot)))
	git(t, f.root, "add", "-A")
	git(t, f.root, "commit", "-qm", "private tracker cutover")
	git(t, f.root, "push", "-q", "origin", "main")
	f.sha = strings.TrimSpace(git(t, f.root, "rev-parse", "HEAD"))
	f.registry, err = tracker.NewRepository(context.Background(), f.root, "origin")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("SDLC_PRIVATE_GH_STATE", f.statuses)
	fake := `#!/usr/bin/env python3
import json, os, sys
a=sys.argv[1:]
if a[:2]==['pr','list']: sys.exit(0)
if a[:3]==['api','-X','POST'] and a[3].startswith('repos/fixture/queue/statuses/'):
 fields=dict(x.split('=',1) for x in a[4:] if x!='-f'); fields['endpoint']=a[3]
 with open(os.environ['SDLC_PRIVATE_GH_STATE'],'a') as f: f.write(json.dumps(fields)+'\n')
 print('{}'); sys.exit(0)
sys.exit('unexpected private gh command: '+repr(a))
`
	if err := os.WriteFile(filepath.Join(os.Getenv("SDLC_QUEUE_BIN"), "gh"), []byte(fake), 0700); err != nil {
		t.Fatal(err)
	}
	f.cli(false, "local-check", "run", "--name", "suite")
	s := f.snapshot()
	if len(s.Attempts) != 1 || s.Attempts[0].Exit == 0 || len(s.Resumes) != 0 {
		t.Fatalf("initial real failure: %+v", s)
	}
	f.parent = s.Attempts[0]
	f.parentRaw = f.read(filepath.Join(f.store, "attempts", f.parent.ID+".json"))
	f.parentLog = f.read(suiteAttemptLogFile(f.store, f.parent))
	// Registration is a separate stateful private tracker update, after failure.
	err = f.registry.ChangeCard("000001", "workshop/issue-cards/000001-flakes.md", "private registration", "register-private", func(raw []byte) ([]byte, error) {
		return bytes.Replace(raw, []byte("status: open\n"), []byte("status: open\nflake_tests: [example.com/private-resume/a.TestA, example.com/private-resume/b.TestB]\n"), 1), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *privateResumeFixture) read(path string) []byte {
	f.t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		f.t.Fatal(e)
	}
	return b
}
func (f *privateResumeFixture) snapshot() SuiteEvidenceSnapshot {
	f.t.Helper()
	s, e := readSuiteEvidence(f.store, 100)
	if e != nil || len(s.Warnings) != 0 {
		f.t.Fatalf("snapshot %+v %v", s, e)
	}
	return s
}
func (f *privateResumeFixture) command(ctx context.Context, args ...string) *exec.Cmd {
	raw, _ := json.Marshal(args)
	c := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLocalCheckPhase1ConformanceProcess$")
	processgroup.Configure(c)
	c.Cancel = func() error { return processgroup.Terminate(c, true) }
	c.Dir = f.root
	c.WaitDelay = 2 * time.Second
	c.Env = append(os.Environ(), "SDLC_PRIVATE_CONFORMANCE_CHILD=1", "SDLC_PRIVATE_CONFORMANCE_ARGS="+string(raw))
	return c
}
func (f *privateResumeFixture) cli(ok bool, args ...string) string {
	f.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	out, err := f.command(ctx, args...).CombinedOutput()
	f.t.Logf("sdlc %s\n%s", strings.Join(args, " "), out)
	if (err == nil) != ok {
		f.t.Fatalf("success=%v want %v: %v\n%s", err == nil, ok, err, out)
	}
	return string(out)
}
func (f *privateResumeFixture) resumeArgs() []string {
	return []string{"local-check", "resume", "--name", "suite", "--attempt", f.parent.ID, "--rev", f.sha}
}
func (f *privateResumeFixture) unchanged() {
	f.t.Helper()
	if !bytes.Equal(f.parentRaw, f.read(filepath.Join(f.store, "attempts", f.parent.ID+".json"))) || !bytes.Equal(f.parentLog, f.read(suiteAttemptLogFile(f.store, f.parent))) {
		f.t.Fatal("original record or raw full-suite log changed")
	}
}
func (f *privateResumeFixture) noDecoys() {
	f.t.Helper()
	for _, p := range []string{"a", "b"} {
		if _, e := os.Stat(filepath.Join(f.marks, p+"-decoy")); !os.IsNotExist(e) {
			f.t.Fatalf("same-name decoy %s ran: %v", p, e)
		}
	}
}
func (f *privateResumeFixture) counts() string {
	f.t.Helper()
	return string(f.read(filepath.Join(f.marks, "suite-count"))) + "/" + string(f.read(filepath.Join(f.marks, "a-count"))) + "/" + string(f.read(filepath.Join(f.marks, "b-count")))
}

// Opt in: SDLC_PRIVATE_RESUME_CONFORMANCE=1 go test ./cmd/sdlc
// -run '^TestLocalCheckResumeConformance$' -count=1 -v. No live evidence is used.
func TestLocalCheckResumeConformance(t *testing.T) {
	if os.Getenv("SDLC_PRIVATE_RESUME_CONFORMANCE") != "1" {
		t.Skip("opt in with SDLC_PRIVATE_RESUME_CONFORMANCE=1")
	}
	t.Run("completed", func(t *testing.T) {
		f := newPrivateResumeFixture(t)
		f.cli(true, f.resumeArgs()...)
		s := f.snapshot()
		if len(s.Resumes) != 1 || len(s.Reconciliations) != 0 || len(s.Attempts) != 2 {
			t.Fatalf("completed bundle: %+v", s)
		}
		var child *SuiteAttempt
		for i := range s.Attempts {
			if s.Attempts[i].ID == s.Resumes[0].ChildID {
				child = &s.Attempts[i]
			}
		}
		if child == nil || child.Exit != 0 || child.RerunOf != f.parent.ID || child.SHA != f.sha {
			t.Fatalf("exact reserved child: %+v", s)
		}
		counts := f.counts()
		if counts != "suite\n/executed\nexecuted\n/executed\nexecuted\n" {
			t.Fatalf("expected one full suite and one exact rerun each: %q", counts)
		}
		f.cli(true, f.resumeArgs()...)
		f.cli(true, "local-check", "republish", "--name", "suite", "--attempt", f.parent.ID, "--rev", f.sha)
		if f.counts() != counts {
			t.Fatal("repeat resume/republish executed tests")
		}
		f.unchanged()
		f.noDecoys()
		statuses := f.read(f.statuses)
		if !bytes.Contains(statuses, []byte(`"state": "success"`)) {
			t.Fatalf("no successful production publication: %s", statuses)
		}
		t.Log("PRIVATE resume: real pre-registration failure -> exact JSON rerun -> success/pass guard; original bytes retained; repeated resume executes nothing")
	})
	t.Run("owner crash and terminal reconcile", func(t *testing.T) {
		f := newPrivateResumeFixture(t)
		if e := os.WriteFile(filepath.Join(f.marks, "block"), nil, 0600); e != nil {
			t.Fatal(e)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		cmd := f.command(ctx, f.resumeArgs()...)
		output, e := os.Create(filepath.Join(t.TempDir(), "crashed-cli.log"))
		if e != nil {
			t.Fatal(e)
		}
		defer output.Close()
		cmd.Stdout = output
		cmd.Stderr = output
		if e := cmd.Start(); e != nil {
			t.Fatal(e)
		}
		finished := false
		pgid, pid := 0, 0
		t.Cleanup(func() {
			if pgid > 1 {
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
			}
			if !finished {
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
			}
		})
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			raw, e := os.ReadFile(filepath.Join(f.marks, "started"))
			if e == nil {
				parts := strings.Fields(string(raw))
				if len(parts) == 2 {
					pid, _ = strconv.Atoi(parts[0])
					pgid, _ = strconv.Atoi(parts[1])
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
		if pid <= 1 || pgid <= 1 || pgid == syscall.Getpgrp() || pgid == cmd.Process.Pid {
			t.Fatal("fixture rerun did not report safe process IDs")
		}
		if e := cmd.Process.Kill(); e != nil {
			t.Fatal(e)
		}
		if e := cmd.Wait(); e == nil {
			t.Fatal("killed CLI returned success")
		}
		finished = true
		// Kill only recorded fixture descendants after the CLI owner has died.
		if e := syscall.Kill(-pgid, syscall.SIGKILL); e != nil && e != syscall.ESRCH {
			t.Fatal(e)
		}
		if e := testfix.AwaitExit(pid, 5*time.Second); e != nil {
			t.Fatal(e)
		}
		s := f.snapshot()
		if len(s.Resumes) != 1 || len(s.Attempts) != 1 || len(s.Reconciliations) != 0 {
			t.Fatalf("crash fence: %+v", s)
		}
		link := s.Resumes[0]
		linkRaw := f.read(filepath.Join(f.store, "attempts", link.ID+".json"))
		digest := resumeRecordDigest(linkRaw)
		counts := f.counts()
		before := f.read(f.statuses)
		args := []string{"local-check", "reconcile", "--attempt", f.parent.ID, "--link-digest", digest, "--note", "private operator verified recorded fixture descendants stopped"}
		f.cli(true, args...)
		s = f.snapshot()
		if len(s.Reconciliations) != 1 || len(s.Attempts) != 1 || s.Reconciliations[0].Outcome != "interrupted" {
			t.Fatalf("terminal interruption: %+v", s)
		}
		r := s.Reconciliations[0]
		receiptPath := filepath.Join(f.store, "attempts", r.ID+".json")
		receipt := f.read(receiptPath)
		f.cli(true, args...)
		f.cli(false, f.resumeArgs()...)
		if f.counts() != counts || !bytes.Equal(before, f.read(f.statuses)) || !bytes.Equal(receipt, f.read(receiptPath)) || !bytes.Equal(linkRaw, f.read(filepath.Join(f.store, "attempts", link.ID+".json"))) {
			t.Fatal("reconciliation/repeat changed execution, immutable proof or published status")
		}
		if e := os.Remove(filepath.Join(f.marks, "block")); e != nil {
			t.Fatal(e)
		}
		refusal := f.cli(false, "local-check", "run", "--name", "suite")
		if !strings.Contains(refusal, "interrupted") {
			t.Fatalf("later green did not refuse terminal interruption: %s", refusal)
		}
		afterGreen := f.snapshot()
		green := 0
		for _, attempt := range afterGreen.Attempts {
			if attempt.Check == suiteCheckName && attempt.Exit == 0 {
				green++
			}
		}
		if green != 1 {
			t.Fatalf("expected one real green suite behind the refusing guard: %+v", afterGreen.Attempts)
		}
		if bytes.Contains(f.read(f.statuses), []byte(`"state": "success"`)) {
			t.Fatal("terminal interruption allowed later green status")
		}
		f.unchanged()
		f.noDecoys()
		t.Log("PRIVATE crash: CLI owner killed after real rerun barrier; recorded descendants stopped; operator reconcile immutable interruption only; no retry/pass; later green blocked")
	})
}
