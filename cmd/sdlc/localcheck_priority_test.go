package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalCheckPrioritySubjectBinding(t *testing.T) {
	d, baseline, gh, _ := localRunRepo(t, "acquired")
	output := filepath.Join(t.TempDir(), "subject.json")
	t.Setenv("PRIORITY_SUBJECT_OUT", output)
	helper := fakeSuiteLockHelper("import os\nopen(os.environ['PRIORITY_SUBJECT_OUT'], 'w').write(os.environ['WF_TEST_PRIORITY_SUBJECT'])\nprint('acquired')\n")
	if err := os.WriteFile(filepath.Join(d.Root, "scripts/test-lock.py"), []byte(helper), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Git("switch", "-qc", "000452-priority"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Git("add", "scripts/test-lock.py"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Git("commit", "-qm", "priority fixture"); err != nil {
		t.Fatal(err)
	}
	head, err := d.Git("rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	head = strings.TrimSpace(head)
	gh.setHead(9, head)
	common, err := filepath.EvalSymlinks(filepath.Join(d.Root, ".git"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, sha  string
		diagnostic bool
		want       int
	}{
		{suiteCheckName, head, false, 452}, {"smoke", head, false, 0}, {suiteCheckName, head, true, 0}, {suiteCheckName, baseline, false, 0},
	} {
		t.Run(fmt.Sprintf("%s-%t-%d", tc.name, tc.diagnostic, tc.want), func(t *testing.T) {
			d.SkipRecord = tc.diagnostic
			// Both commits contain the script marker; helper comes from original checkout.
			rec, err := runLocalCheck(context.Background(), d, localCheck{Name: tc.name, Command: []string{"sh", "-c", "echo passed"}, TimeoutMinutes: 5}, tc.sha)
			if err != nil || !rec.passed() {
				t.Fatalf("run %+v %v", rec, err)
			}
			data, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			var got suitePrioritySubject
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if got.Issue != tc.want || got.CommonDir != common {
				t.Fatalf("subject %+v want issue %d original common dir", got, tc.want)
			}
		})
	}
}
