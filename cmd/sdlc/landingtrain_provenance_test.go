package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Exercise the real conductor, runner, records and CLI consumer. GitHub and
// capacity are the existing stateful/private fixture seams; no fleet work runs.
func TestLandingTrainRecordsConductorProvenance(t *testing.T) {
	s := newTrainSetup(t, "train")
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	common := git(t, s.first, "rev-parse", "--git-common-dir")
	observations := filepath.Join(t.TempDir(), "owners.jsonl")
	helper := fmt.Sprintf(`import os, json, glob
op = json.load(open(os.path.join(os.environ['WF_SDLC_QUEUE_ROOT'], 'operations', os.environ['WF_SDLC_QUEUE_ID'] + '.json')))
leases = [json.load(open(p))['root'] for p in glob.glob(%q)]
with open(%q, 'a') as out:
    out.write(json.dumps(dict(owner=os.environ['WF_TEST_LOCK_OWNER'], root=op['root'], workspace=op['workspace'], leases=leases)) + '\n')
print('skipped')
`, filepath.Join(common, "sdlc", "local-check-runs", "*.json"), observations)
	procedureWrite(t, s.server, "scripts/test-lock.py", fakeSuiteLockHelper(helper))
	git(t, s.server, "add", ".")
	git(t, s.server, "commit", "-qm", "observe train suite ownership")
	git(t, s.server, "push", "-q", "origin", "main")
	trainCarWithIssue(t, s, "---\nid: 000002\nstatus: codecomplete\n---\n## Log\nclosed\n")
	one := s.merge(s.first)
	waitQueued(t, s.line, 1, one)
	s.queuedCar(t, "000002-second", 2, 2)
	waitVouched(t, s.line, 2)
	s.blocker.Release()
	if err := finished(t, one); err != nil {
		t.Fatalf("conductor: %v", err)
	}
	if len(s.gh.created) != 1 {
		t.Fatalf("train PRs: %v", s.gh.created)
	}
	attempts, warnings, err := readSuiteAttempts(filepath.Join(common, "sdlc", "test-stats"), 100)
	if err != nil || len(warnings) != 0 || len(attempts) != 1 {
		t.Fatalf("train attempts: %+v warnings=%v err=%v", attempts, warnings, err)
	}
	a := attempts[0]
	if a.SHA != s.gh.pr(s.gh.created[0]).HeadOID || a.Slot != "sample:1" || a.MainBaseline {
		t.Errorf("train attempt lost conductor identity: %+v", a)
	}
	raw, err := os.ReadFile(observations)
	if err != nil {
		t.Fatal(err)
	}
	var observed struct {
		Owner, Root, Workspace string
		Leases                 []string
	}
	if err := json.Unmarshal(bytes.TrimSpace(raw), &observed); err != nil {
		t.Fatal(err)
	}
	if observed.Owner != s.first || observed.Root != s.first || observed.Workspace != "sample:1" || len(observed.Leases) != 1 || observed.Leases[0] != s.first {
		t.Errorf("live train ownership uses disposable content root: %+v; want %s", observed, s.first)
	}
	trainAttemptActualConsumer(t, a)
}

func trainAttemptActualConsumer(t *testing.T, produced SuiteAttempt) {
	t.Helper()
	r, _, _, owner, source := ownedResolverFixture(t, func(r *trackerRepo) { trainActualGitHubTransport(t, r) })
	f, err := os.OpenFile(source, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	for _, minute := range []int{2, 8} {
		_, err = fmt.Fprintf(f, "{\"timestamp\":\"2026-10-14T00:%02d:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":[{\"type\":\"input_text\",\"text\":\"#335\"}]}}\n", minute)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	common := git(t, r.root, "rev-parse", "--git-common-dir")
	if !filepath.IsAbs(common) {
		common = filepath.Join(r.root, common)
	}
	dir := filepath.Join(common, "sdlc", "test-stats", "attempts")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	// Transplant the real writer's record into the existing fixed-clock actuals
	// fixture. Rebind only repository/head/time; keep its produced slot and exact
	// queue duration, so the foreign conductor overlaps issue 335's ownership.
	a := produced
	a.Repo, a.SHA = strings.TrimPrefix(owner.Repository, "github.com/"), procedureHead(t, r.root)
	a.LogPath = filepath.Join("local-checks", a.SHA, a.Check+".log")
	queued := time.Date(2026, 10, 14, 0, 4, 0, 0, time.UTC)
	acquired := queued.Add(produced.AcquiredAt.Sub(*produced.QueuedAt))
	a.QueuedAt, a.AcquiredAt = &queued, &acquired
	a.StartedAt, a.FinishedAt = acquired, acquired.Add(produced.FinishedAt.Sub(produced.StartedAt))
	brain := t.TempDir()
	t.Chdir(r.root)
	for _, missing := range []bool{false, true} {
		candidate := a
		if missing {
			candidate.Slot = ""
		}
		raw, err := json.Marshal(candidate)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := decodeSuiteAttempt(raw); err != nil {
			t.Fatalf("retimed producer record is invalid: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, candidate.ID+".json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
		var output bytes.Buffer
		cmd := NewActualCmd()
		cmd.SetOut(&output)
		cmd.SetErr(&output)
		cmd.SetArgs([]string{"--issue", "335", "--brain-dir", brain})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		if missing {
			if !strings.Contains(output.String(), "unknown slot provenance") {
				t.Errorf("missing-slot control did not refuse: %s", &output)
			}
		} else if !strings.Contains(output.String(), "0.10h") || strings.Contains(output.String(), "queue evidence incomplete") {
			t.Errorf("foreign train blocked overlapping actuals: %s", &output)
		}
	}
}

// Keep GitHub identity reads real while the existing bare repository supplies
// only the network transport. Every other Git operation delegates unchanged.
func trainActualGitHubTransport(t *testing.T, r *trackerRepo) {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	const url = "https://github.com/xianxu/ariadne.git"
	git(t, r.root, "remote", "set-url", "origin", url)
	dir := t.TempDir()
	body := fmt.Sprintf(`#!/usr/bin/env python3
import os, sys
args = sys.argv[1:]
for i, arg in enumerate(args):
    if arg in ('fetch', 'push', 'ls-remote'):
        for j in range(i + 1, len(args)):
            if args[j] in ('origin', %q):
                args[j] = %q
                break
        break
os.execv(%q, [%q] + args)
`, url, r.origin, realGit, realGit)
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(body), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestLocalCheckRelativeCommonDirUsesCaller(t *testing.T) {
	d, _, _, _ := localRunRepo(t, "skipped")
	caller := d.Root
	d.Root, d.WorkflowAnchor = t.TempDir(), caller
	got, err := localCheckLogDir(d)
	want := filepath.Join(caller, ".git", "sdlc", "local-checks")
	if err != nil || got != want {
		t.Fatalf("caller-relative common dir: %s, %v; want %s", got, err, want)
	}
}
