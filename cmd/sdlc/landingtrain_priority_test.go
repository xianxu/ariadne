package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// Observe the subject at the subprocess boundary, after production has built
// the train and selected the exact revision to test.
func trainPriorityCapture(t *testing.T, root string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "subjects.jsonl")
	t.Setenv("TRAIN_PRIORITY_SUBJECTS", out)
	procedureWrite(t, root, "scripts/test-lock.py", fakeSuiteLockHelper("import os\nwith open(os.environ['TRAIN_PRIORITY_SUBJECTS'], 'a') as f:\n    f.write(os.environ['WF_TEST_PRIORITY_SUBJECT'] + '\\n')\nprint('acquired')\n"))
	return out
}

func readTrainPrioritySubjects(t *testing.T, path string) []suitePrioritySubject {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var subjects []suitePrioritySubject
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		var subject suitePrioritySubject
		if err := json.Unmarshal([]byte(line), &subject); err != nil {
			t.Fatal(err)
		}
		subjects = append(subjects, subject)
	}
	return subjects
}

func TestLandingTrainPriorityBindsOnlyKeptCars(t *testing.T) {
	s := newTrainSetup(t, "train")
	out := trainPriorityCapture(t, s.server)
	git(t, s.server, "add", "scripts/test-lock.py")
	git(t, s.server, "commit", "-qm", "capture train queue subjects")
	git(t, s.server, "push", "-q", "origin", "main")
	second := s.pushCar(t, "000002-second", "second", 2, 2)
	procedureWrite(t, second, "shipped.txt", "conflicts with the first car\n")
	git(t, second, "add", "shipped.txt")
	git(t, second, "commit", "-qm", "conflicting car")
	git(t, second, "push", "-q", "upstream", "000002-second")
	s.addCar(t, "000002-second", 2)
	s.pushCar(t, "000003-third", "third", 3, 3)
	one := s.merge(s.first)
	waitQueued(t, s.line, 1, []chan error{one})
	two := s.waitingCar(t, "000002-second", 2, 2)
	three := s.waitingCar(t, "000003-third", 3, 3)
	waitVouched(t, s.line, 3, []chan error{one, two, three})
	s.blocker.Release()
	if err := finished(t, one); err != nil {
		t.Fatal(err)
	}
	wantSettled(t, two, landingline.OutcomeHeld)
	wantSettled(t, three, landingline.OutcomeLanded)
	common := git(t, s.first, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if resolved, err := filepath.EvalSymlinks(common); err == nil {
		common = resolved
	}
	found := false
	for _, subject := range readTrainPrioritySubjects(t, out) {
		if len(subject.Issues) == 0 { // a car may run its own preparation check
			continue
		}
		found = true
		if subject.CommonDir != common || subject.Issue != 0 || !slices.Equal(subject.Issues, []int{1, 3}) {
			t.Fatalf("train subject %+v; want only admitted cars [1 3] in %s", subject, common)
		}
	}
	if !found {
		t.Fatal("real conductor's integration suite never carried its admitted car IDs")
	}
}

func TestLandingTrainPriorityExactPrefixAfterReintegration(t *testing.T) {
	s := newTrainSetup(t, "train")
	s.pushCar(t, "000002-second", "second", 2, 2)
	c := &conductor{ctx: context.Background(), r: execGitRunner{}, t: landingTarget{Root: s.first, Remote: "upstream", Repo: "test/repo"}, f: landingFlags(), stderr: io.Discard, id: 535}
	t.Cleanup(func() {
		if c.wt != "" {
			git(t, s.first, "worktree", "remove", "--force", c.wt)
		}
	})
	cars := []landingline.Car{
		{Seq: 1, Issue: 1, Branch: landingTestBranch, Head: s.heads[landingTestBranch], PR: 1},
		{Seq: 2, Issue: 2, Branch: "000002-second", Head: s.heads["000002-second"], PR: 2},
	}
	if _, err := c.build(cars); err != nil {
		t.Fatal(err)
	}
	if len(c.heads) != 2 {
		t.Fatalf("built heads %v", c.heads)
	}
	old := slices.Clone(c.heads)
	for i, sha := range old {
		if got := c.priorityIssues(sha); !slices.Equal(got, []int{1, 2}[:i+1]) {
			t.Fatalf("split prefix %d: got %v", i+1, got)
		}
	}
	if got := c.priorityIssues(c.base); len(got) != 0 {
		t.Fatalf("main baseline borrowed car IDs %v", got)
	}
	integrationAdvance(t, s.server, "new-main.go", "package main\n")
	if _, err := c.reintegrate(2); err != nil {
		t.Fatal(err)
	}
	if c.heads[1] == old[1] || c.heads[0] != "" {
		t.Fatalf("code-changing reintegration retained stale prefix: %v", c.heads)
	}
	if got := c.priorityIssues(c.heads[1]); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("reintegrated prefix lost its original cars: %v", got)
	}
	for _, sha := range append(old, "", c.base, strings.Repeat("f", 40)) {
		if got := c.priorityIssues(sha); len(got) != 0 {
			t.Errorf("unmatched or obsolete revision %q borrowed %v", sha, got)
		}
	}
}

func TestLandingTrainPriorityRecordedTestChecksOnly(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "acquired")
	out := trainPriorityCapture(t, d.Root)
	git(t, d.Root, "add", "scripts/test-lock.py")
	git(t, d.Root, "commit", "-qm", "capture detached train subjects")
	head := git(t, d.Root, "rev-parse", "HEAD")
	gh.setHead(9, head)
	c := &conductor{heads: []string{head}, carIssues: []int{535}}
	d.PriorityIssues = c.priorityIssues
	for _, tc := range []struct {
		name       string
		sha        string
		diagnostic bool
		want       []int
	}{
		{suiteCheckName, head, false, []int{535}},
		{affectedCheckName, head, false, []int{535}},
		{"smoke", head, false, nil},
		{suiteCheckName, head, true, nil},
		{suiteCheckName, sha, false, nil},
	} {
		d.SkipRecord = tc.diagnostic
		rec, err := runLocalCheck(context.Background(), d, localCheck{Name: tc.name, Command: []string{"true"}, TimeoutMinutes: 5}, tc.sha)
		if err != nil || !rec.passed() {
			t.Fatalf("%s diagnostic=%t: %+v %v", tc.name, tc.diagnostic, rec, err)
		}
		subjects := readTrainPrioritySubjects(t, out)
		got := subjects[len(subjects)-1]
		if got.Issue != 0 || !slices.Equal(got.Issues, tc.want) {
			t.Fatalf("%s diagnostic=%t sha=%s: subject %+v want %v", tc.name, tc.diagnostic, tc.sha, got, tc.want)
		}
	}
}
