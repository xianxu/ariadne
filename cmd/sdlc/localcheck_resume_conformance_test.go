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
 if os.environ.get('PRIVATE_RESUME_BARRIER')=='statuspost' and fields['state']=='success':
  with open(os.path.join(os.environ['PRIVATE_RESUME_MARKS'],'barrier.json'),'w') as f: json.dump({'pid':os.getpid(),'wt':''},f)
  with open(os.path.join(os.environ['PRIVATE_RESUME_MARKS'],'barrier.fifo'),'r') as f: f.read()
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

// The wrappers stop at confirmed external effects; readiness is an explicit
// file/FIFO handshake, never a delay chosen to race a native publication.
func (f *privateResumeFixture) installBarrier(phase string) {
	f.t.Helper()
	bin := os.Getenv("SDLC_QUEUE_BIN")
	transport := filepath.Join(bin, "git-private-transport")
	if err := os.Rename(filepath.Join(bin, "git"), transport); err != nil {
		f.t.Fatal(err)
	}
	script := `#!/usr/bin/env python3
import json, os, subprocess, sys
a=sys.argv[1:]; phase=os.environ.get('PRIVATE_RESUME_BARRIER','')
def stop(wt):
 with open(os.path.join(os.environ['PRIVATE_RESUME_MARKS'],'barrier.json'),'w') as f: json.dump({'pid':os.getpid(),'wt':wt},f)
 with open(os.path.join(os.environ['PRIVATE_RESUME_MARKS'],'barrier.fifo'),'r') as f: f.read()
op=''
if 'worktree' in a:
 i=a.index('worktree'); op=a[i+1]
if phase=='markerbeforelock' and op=='lock': stop(a[-1])
if phase=='receiptcleanup' and op=='remove': stop(a[-1])
r=subprocess.run([os.path.join(os.path.dirname(__file__),'git-private-transport')]+a)
if r.returncode==0 and phase=='prelink' and op=='add': stop(a[-2])
if r.returncode==0 and phase=='afterlock' and op=='lock': stop(a[-1])
if r.returncode==0 and phase=='cleanupbeforecommit' and op=='remove': stop(a[-1])
sys.exit(r.returncode)
`
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0700); err != nil {
		f.t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(f.marks, "barrier.fifo"), 0600); err != nil {
		f.t.Fatal(err)
	}
	f.t.Setenv("PRIVATE_RESUME_BARRIER", phase)
}

type privateResumeBarrier struct {
	PID int    `json:"pid"`
	WT  string `json:"wt"`
}

func (f *privateResumeFixture) killAtBarrier(args ...string) privateResumeBarrier {
	f.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := f.command(ctx, args...)
	output, err := os.Create(filepath.Join(f.t.TempDir(), "barrier-cli.log"))
	if err != nil {
		f.t.Fatal(err)
	}
	defer output.Close()
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Start(); err != nil {
		f.t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var barrier privateResumeBarrier
	finished := false
	f.t.Cleanup(func() {
		if barrier.PID > 1 {
			_ = syscall.Kill(barrier.PID, syscall.SIGKILL)
		}
		if !finished {
			_ = cmd.Process.Kill()
			<-done
		}
	})
	deadline := time.After(45 * time.Second)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
waiting:
	for {
		select {
		case err := <-done:
			finished = true
			f.t.Fatalf("CLI exited before barrier: %v\n%s", err, f.read(output.Name()))
		case <-deadline:
			f.t.Fatal("CLI did not reach deterministic barrier")
		case <-tick.C:
			raw, err := os.ReadFile(filepath.Join(f.marks, "barrier.json"))
			if err == nil && json.Unmarshal(raw, &barrier) == nil && barrier.PID > 1 {
				break waiting
			}
		}
	}
	if barrier.PID == os.Getpid() || barrier.PID == cmd.Process.Pid {
		f.t.Fatal("barrier did not identify its private child")
	}
	if err := cmd.Process.Kill(); err != nil {
		f.t.Fatal(err)
	}
	// Only the private wrapper/GH subprocess remains; real Git has either not
	// started or has returned. Kill it explicitly before operator attestation.
	if err := syscall.Kill(barrier.PID, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
		f.t.Fatal(err)
	}
	if err := <-done; err == nil {
		f.t.Fatal("killed owner returned success")
	}
	finished = true
	if err := testfix.AwaitExit(barrier.PID, 5*time.Second); err != nil {
		f.t.Fatal(err)
	}
	f.t.Setenv("PRIVATE_RESUME_BARRIER", "")
	return barrier
}

func privateResumeConformanceEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("SDLC_PRIVATE_RESUME_CONFORMANCE") != "1" {
		t.Skip("opt in with SDLC_PRIVATE_RESUME_CONFORMANCE=1")
	}
}

func (f *privateResumeFixture) noPreparedWorktree() {
	f.t.Helper()
	out := git(f.t, f.root, "worktree", "list", "--porcelain")
	if strings.Count(out, "worktree ") != 1 || strings.Contains(out, "locked") {
		f.t.Fatalf("prepared checkout/registration remains: %s", out)
	}
}

func (f *privateResumeFixture) discoverableProtectedInputs() {
	f.t.Helper()
	s := f.snapshot()
	if len(s.Resumes) != 1 {
		f.t.Fatal("protected environment lacks unique durable link")
	}
	l := s.Resumes[0]
	marker := f.read(filepath.Join(f.store, "inflight", l.ChildID, runnerArtifactOwnerFile))
	gitdir := strings.TrimSpace(strings.TrimPrefix(string(marker), "gitdir: "))
	if strings.TrimSpace(string(f.read(filepath.Join(gitdir, "locked")))) != resumeEnvironmentReason(l.ChildID) {
		f.t.Fatal("protected worktree lacks exact discoverable child owner binding")
	}
}

func (f *privateResumeFixture) reconcileArgs() []string {
	s := f.snapshot()
	if len(s.Resumes) != 1 {
		f.t.Fatalf("expected one pending fence: %+v", s)
	}
	l := s.Resumes[0]
	return []string{"local-check", "reconcile", "--attempt", f.parent.ID, "--link-digest", resumeRecordDigest(f.read(filepath.Join(f.store, "attempts", l.ID+".json"))), "--note", "private operator stopped all recorded fixture descendants"}
}

func (f *privateResumeFixture) reconcileWithoutExecution() {
	f.t.Helper()
	counts := f.counts()
	statuses := f.read(f.statuses)
	args := f.reconcileArgs()
	f.cli(true, args...)
	s := f.snapshot()
	if len(s.Reconciliations) != 1 || len(s.Attempts) != 1 {
		f.t.Fatalf("expected interruption without committed child: %+v", s)
	}
	r := s.Reconciliations[0]
	path := filepath.Join(f.store, "attempts", r.ID+".json")
	raw := f.read(path)
	f.cli(true, args...)
	f.cli(false, f.resumeArgs()...)
	if f.counts() != counts || !bytes.Equal(statuses, f.read(f.statuses)) || !bytes.Equal(raw, f.read(path)) {
		f.t.Fatal("recovery changed execution, publication or immutable receipt")
	}
	f.unchanged()
}

func TestResumeCrashPreLink(t *testing.T) {
	privateResumeConformanceEnabled(t)
	f := newPrivateResumeFixture(t)
	counts := f.counts()
	f.installBarrier("prelink")
	b := f.killAtBarrier(f.resumeArgs()...)
	if len(f.snapshot().Resumes) != 0 || f.counts() != counts {
		t.Fatal("pre-link preparation executed or published a fence")
	}
	if runnerArtifactProtected(b.WT) {
		t.Fatal("pre-link preparation is protected")
	}
	old := time.Now().Add(-2 * localCheckStaleAge)
	if err := os.Chtimes(b.WT, old, old); err != nil {
		t.Fatal(err)
	}
	sweepStaleLocalWorktrees(localRunDeps{Root: f.root, Git: func(args ...string) (string, error) {
		out, e := exec.Command("git", append([]string{"-C", f.root}, args...)...).CombinedOutput()
		return string(out), e
	}})
	f.noPreparedWorktree()
	f.unchanged()
}

// No external command separates native link publication from marker creation.
// This deterministic process harness exits at that seam using the production
// publisher and actual parent lock, rather than racing a filesystem watcher.
func TestResumeCrashLinkBeforeMarkerProcess(t *testing.T) {
	if os.Getenv("PRIVATE_RESUME_LINK_PROCESS") != "1" {
		return
	}
	root, parent := os.Getenv("PRIVATE_RESUME_ROOT"), os.Getenv("PRIVATE_RESUME_PARENT")
	store := filepath.Join(root, ".git", "sdlc", "test-stats")
	lock, err := lockResumeParent(context.Background(), store, parent)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	gitCall := func(args ...string) (string, error) {
		out, e := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
		return string(out), e
	}
	wt, _, err := detachedWorktree(gitCall, lock.Attempt.SHA)
	if err != nil {
		t.Fatal(err)
	}
	w, err := beginSuiteAttempt(store)
	if err != nil {
		t.Fatal(err)
	}
	started, err := processStartTime(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	boot, err := queueBootID()
	if err != nil {
		t.Fatal(err)
	}
	id, err := resumeEvidenceID()
	if err != nil {
		t.Fatal(err)
	}
	l := ResumeLink{Version: 3, Kind: "resume-link", ID: id, Repo: lock.Attempt.Repo, SHA: lock.Attempt.SHA, Check: suiteCheckName, RerunOf: parent, ParentRecordSHA256: resumeRecordDigest(lock.Raw), ParentLogSHA256: lock.Attempt.LogSHA256, ChildID: w.ID, Owner: ResumeOwner{PID: os.Getpid(), PIDStarted: started, BootID: boot, Host: hostName()}, CreatedAt: time.Now().UTC()}
	raw, err := json.Marshal(l)
	if err != nil {
		t.Fatal(err)
	}
	if err := publishResumeEvidence(store, lock, raw); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(os.Getenv("PRIVATE_RESUME_MARKS"), "native-wt"), []byte(wt), 0600); err != nil {
		t.Fatal(err)
	}
	os.Exit(42)
}

func TestResumeCrashLinkBeforeMarker(t *testing.T) {
	privateResumeConformanceEnabled(t)
	f := newPrivateResumeFixture(t)
	counts := f.counts()
	cmd := exec.Command(os.Args[0], "-test.run=^TestResumeCrashLinkBeforeMarkerProcess$")
	cmd.Dir = f.root
	cmd.Env = append(os.Environ(), "PRIVATE_RESUME_LINK_PROCESS=1", "PRIVATE_RESUME_ROOT="+f.root, "PRIVATE_RESUME_PARENT="+f.parent.ID)
	out, err := cmd.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 42 {
		t.Fatalf("native seam did not exit immediately after publication: %v\n%s", err, out)
	}
	s := f.snapshot()
	if len(s.Resumes) != 1 || len(s.Attempts) != 1 || f.counts() != counts {
		t.Fatalf("link seam launched or lost fence: %+v", s)
	}
	wt := string(f.read(filepath.Join(f.marks, "native-wt")))
	if runnerArtifactProtected(wt) {
		t.Fatal("link-before-marker checkout was protected")
	}
	if _, err := os.Stat(filepath.Join(f.store, "inflight", s.Resumes[0].ChildID, runnerArtifactOwnerFile)); !os.IsNotExist(err) {
		t.Fatalf("marker exists before marker creation: %v", err)
	}
	f.reconcileWithoutExecution()
	old := time.Now().Add(-2 * localCheckStaleAge)
	if err := os.Chtimes(wt, old, old); err != nil {
		t.Fatal(err)
	}
	sweepStaleLocalWorktrees(localRunDeps{Root: f.root, Git: func(args ...string) (string, error) {
		out, e := exec.Command("git", append([]string{"-C", f.root}, args...)...).CombinedOutput()
		return string(out), e
	}})
	f.noPreparedWorktree()
	f.unchanged()
}

func TestResumeCrashMarkerBeforeLock(t *testing.T) {
	privateResumeConformanceEnabled(t)
	f := newPrivateResumeFixture(t)
	counts := f.counts()
	f.installBarrier("markerbeforelock")
	b := f.killAtBarrier(f.resumeArgs()...)
	s := f.snapshot()
	if len(s.Resumes) != 1 || f.counts() != counts {
		t.Fatal("marker-before-lock launched tests or lacks durable link")
	}
	marker := f.read(filepath.Join(f.store, "inflight", s.Resumes[0].ChildID, runnerArtifactOwnerFile))
	if !bytes.HasPrefix(marker, []byte("gitdir: ")) {
		t.Fatal("owner binding not discoverable before Git lock")
	}
	if runnerArtifactProtected(b.WT) {
		t.Fatal("Git lock acquired before lock-call barrier")
	}
	f.reconcileWithoutExecution()
	// The exact child Git lock was never acquired, so destructive receipt
	// cleanup has no lock authority. Ordinary unlocked preparation GC owns it.
	old := time.Now().Add(-2 * localCheckStaleAge)
	if err := os.Chtimes(b.WT, old, old); err != nil {
		t.Fatal(err)
	}
	sweepStaleLocalWorktrees(localRunDeps{Root: f.root, Git: func(args ...string) (string, error) {
		out, e := exec.Command("git", append([]string{"-C", f.root}, args...)...).CombinedOutput()
		return string(out), e
	}})
	f.noPreparedWorktree()
}

func TestResumeCrashCheckoutRetiredBeforeChildCommit(t *testing.T) {
	privateResumeConformanceEnabled(t)
	f := newPrivateResumeFixture(t)
	f.installBarrier("cleanupbeforecommit")
	b := f.killAtBarrier(f.resumeArgs()...)
	if _, err := os.Stat(b.WT); !os.IsNotExist(err) {
		t.Fatalf("real removal did not retire checkout: %v", err)
	}
	f.noPreparedWorktree()
	s := f.snapshot()
	if len(s.Attempts) != 1 || len(s.Resumes) != 1 {
		t.Fatalf("unconfirmed cleanup reply committed child: %+v", s)
	}
	_ = f.read(filepath.Join(f.store, "inflight", s.Resumes[0].ChildID, runnerArtifactOwnerFile))
	if f.counts() != "suite\n/executed\nexecuted\n/executed\nexecuted\n" {
		t.Fatal("expected real uncommitted targeted PASS")
	}
	f.reconcileWithoutExecution()
	f.noPreparedWorktree()
}

func TestResumeCrashCommittedChildStatusPost(t *testing.T) {
	privateResumeConformanceEnabled(t)
	f := newPrivateResumeFixture(t)
	f.installBarrier("statuspost")
	f.killAtBarrier(f.resumeArgs()...)
	s := f.snapshot()
	if len(s.Attempts) != 2 || len(s.Resumes) != 1 {
		t.Fatalf("status post preceded committed child: %+v", s)
	}
	f.noPreparedWorktree()
	counts := f.counts()
	raw := map[string][]byte{}
	for _, a := range s.Attempts {
		raw[a.ID] = f.read(filepath.Join(f.store, "attempts", a.ID+".json"))
	}
	for _, link := range s.Resumes {
		raw[link.ID] = f.read(filepath.Join(f.store, "attempts", link.ID+".json"))
	}
	f.cli(true, f.resumeArgs()...)
	f.cli(true, "local-check", "republish", "--name", "suite", "--attempt", f.parent.ID, "--rev", f.sha)
	if f.counts() != counts {
		t.Fatal("post-crash recovery executed tests")
	}
	for id, b := range raw {
		if !bytes.Equal(b, f.read(filepath.Join(f.store, "attempts", id+".json"))) {
			t.Fatal("post-crash recovery rewrote proof")
		}
	}
	f.noPreparedWorktree()
	f.unchanged()
}

func TestResumeCrashTerminalReceiptBeforeCleanup(t *testing.T) {
	privateResumeConformanceEnabled(t)
	f := newPrivateResumeFixture(t)
	f.installBarrier("afterlock")
	f.killAtBarrier(f.resumeArgs()...)
	// The first owner's marker is durable, and no execution was launched.
	if err := os.Remove(filepath.Join(f.marks, "barrier.json")); err != nil {
		t.Fatal(err)
	}
	f.t.Setenv("PRIVATE_RESUME_BARRIER", "receiptcleanup")
	args := f.reconcileArgs()
	f.killAtBarrier(args...)
	s := f.snapshot()
	if len(s.Reconciliations) != 1 || len(s.Attempts) != 1 {
		t.Fatalf("cleanup preceded immutable interrupted receipt: %+v", s)
	}
	f.discoverableProtectedInputs()
	r := s.Reconciliations[0]
	path := filepath.Join(f.store, "attempts", r.ID+".json")
	raw := f.read(path)
	counts := f.counts()
	statuses := f.read(f.statuses)
	f.cli(true, args...)
	f.cli(true, args...)
	if !bytes.Equal(raw, f.read(path)) || f.counts() != counts || !bytes.Equal(statuses, f.read(f.statuses)) {
		t.Fatal("lost cleanup reply changed receipt/execution/publication")
	}
	f.noPreparedWorktree()
	f.unchanged()
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
	t.Run("running", func(t *testing.T) {
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
		f.discoverableProtectedInputs()
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
		f.noPreparedWorktree()
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
