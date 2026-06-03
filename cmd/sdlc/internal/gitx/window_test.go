package gitx

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"time"
)

// TestCommitWindow_ExtendedCapIncludes45Days pins the #68 cap bump (31→61): a
// commit ~45 days old must still anchor the window. Under the old 31-day cap it
// would be excluded (empty window) and the issue's actuals would come up short.
// CommitWindow reads the cwd via direct git, so we build a throwaway repo and
// chdir into it.
func TestCommitWindow_ExtendedCapIncludes45Days(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "t@t")
	run("config", "user.name", "t")
	run("config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "f")
	// Author/committer date 45 days ago: inside the 61-day cap, outside the old 31.
	dated := time.Now().AddDate(0, 0, -45).Format("2006-01-02T15:04:05")
	cmd := exec.Command("git", "commit", "-q", "-m", "#99 M1: the work")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+dated, "GIT_COMMITTER_DATE="+dated)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}

	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	sha, _, _, err := CommitWindow("99")
	if err != nil {
		t.Fatalf("CommitWindow: %v", err)
	}
	if sha == "" {
		t.Errorf("45-day-old #99 commit should be within the %d-day cap, but the window was empty", WindowCapDays)
	}
}

// TestIssueRefRE_DiscoveryParsing exercises the regex used by
// DiscoverWindowIssues, in isolation from git.
func TestIssueRefRE_DiscoveryParsing(t *testing.T) {
	tests := []struct {
		subject string
		want    []string
	}{
		{"#15: subject text", []string{"15"}},
		{"close #31 M4: foo", []string{"31"}},
		{"chore: bump (refs #1, #2, #3)", []string{"1", "2", "3"}},
		{"no issue ref here", nil},
		{"#42abc not a real ref", nil}, // word boundary blocks #42 from "42abc"
		{"prefix#42 suffix", []string{"42"}},
		{"#1 and #11 distinct", []string{"1", "11"}},
	}
	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			matches := issueRefRE.FindAllStringSubmatch(tt.subject, -1)
			var got []string
			for _, m := range matches {
				got = append(got, m[1])
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
}

// TestSubjectAnchorRE verifies the regex used inside CommitWindow to
// filter --grep candidates down to true subject-anchored matches.
// We rebuild the same pattern here (CommitWindow compiles it inline from
// the issue number); equivalent to close-issue.py's
//
//	^(close\s+)?#NN(?!\d)
//
// Go's RE2 doesn't support negative lookahead, so we render (?!\d) as
// ($|[^0-9]). Equivalent behavior for the close-issue use case.
func TestSubjectAnchorRE(t *testing.T) {
	subjectRE := regexp.MustCompile(`^(close\s+)?#15($|[^0-9])`)
	tests := []struct {
		subject string
		match   bool
	}{
		{"#15: subject", true},
		{"close #15 done", true},
		{"close   #15: tabby", true},
		{"#15", true},
		{"#150: different issue", false},
		{"chore: see #15 in body", false},
		{"prefix #15 not anchored", false},
	}
	for _, tt := range tests {
		got := subjectRE.MatchString(tt.subject)
		if got != tt.match {
			t.Errorf("MatchString(%q) = %v, want %v", tt.subject, got, tt.match)
		}
	}
}
