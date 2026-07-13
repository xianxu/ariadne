package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func TestIssueFileRefFilters(t *testing.T) {
	refs := []issueFileRef{
		{Path: "working.md", Status: "working"},
		{Path: "done.md", Status: "done"},
		{Path: "codecomplete.md", Status: "codecomplete"},
		{Path: "missing.md"},
		{Path: "wontfix.md", Status: "wontfix"},
		{Path: "open.md", Status: "open"},
		{Path: "punt.md", Status: "punt"},
	}

	tests := []struct {
		name string
		got  []issueFileRef
		want []issueFileRef
	}{
		{
			name: "codecomplete",
			got:  codecompleteIssueFiles(refs),
			want: refs[2:3],
		},
		{
			name: "not done",
			got:  notDoneIssueFiles(refs),
			want: []issueFileRef{refs[0], refs[3], refs[5]},
		},
		{
			name: "terminal",
			got:  terminalIssueFiles(refs),
			want: []issueFileRef{refs[1], refs[4], refs[6]},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Fatalf("got %#v, want %#v", tt.got, tt.want)
			}
		})
	}
}

func TestScanIssueFilesWindowPreservesOrderAndParsedSnapshot(t *testing.T) {
	dir := t.TempDir()
	first := writeScanIssueFile(t, dir, "000001-first.md", "working", "# First\n")
	second := writeScanIssueFile(t, dir, "custom.md", "codecomplete", "# Second\n")

	var gotArgs []string
	runGit := func(args ...string) ([]byte, error) {
		gotArgs = append([]string(nil), args...)
		return []byte(second + "\n" + first + "\n"), nil
	}
	refs, err := scanIssueFiles("base", dir, runGit)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"diff", "--name-only", "base..HEAD", "--", dir + "/*.md"}; !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("git args = %#v, want %#v", gotArgs, want)
	}
	if got, want := issueFilePaths(refs), []string{second, first}; !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %#v, want %#v", got, want)
	}
	if refs[0].Status != "codecomplete" || refs[0].Body != "# Second\n" {
		t.Fatalf("parsed ref = %#v", refs[0])
	}
	updated := issue.SetField(refs[0].Frontmatter, "status", "done")
	if got := issue.Compose(updated, refs[0].Body); !strings.Contains(got, "status: done\n---\n# Second\n") {
		t.Fatalf("composed parsed snapshot = %q", got)
	}
}

func TestScanIssueFilesWindowUsesRealGitDiff(t *testing.T) {
	repo := hermeticRepo(t)
	issuesDir := filepath.Join("workshop", "issues")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeScanIssueFile(t, issuesDir, "000001-first.md", "working", "# First\n")
	writeScanIssueFile(t, issuesDir, "custom.md", "working", "# Custom\n")
	runGitCommand(t, repo, "add", ".")
	runGitCommand(t, repo, "commit", "-qm", "base")
	base := strings.TrimSpace(runGitCommand(t, repo, "rev-parse", "HEAD"))
	writeScanIssueFile(t, issuesDir, "000001-first.md", "codecomplete", "# First changed\n")
	writeScanIssueFile(t, issuesDir, "custom.md", "done", "# Custom changed\n")
	runGitCommand(t, repo, "add", ".")
	runGitCommand(t, repo, "commit", "-qm", "changed")

	runner := execGitRunner{}
	refs, err := scanIssueFiles(base, issuesDir, runner.Git)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := issueFilePaths(refs), []string{
		filepath.Join(issuesDir, "000001-first.md"),
		filepath.Join(issuesDir, "custom.md"),
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %#v, want %#v", got, want)
	}
}

func TestScanIssueFilesDirectoryUsesSharedGrammarAndSorts(t *testing.T) {
	dir := t.TempDir()
	second := writeScanIssueFile(t, dir, "000002-second.md", "done", "# Second\n")
	first := writeScanIssueFile(t, dir, "000001-first.md", "working", "# First\n")
	writeScanIssueFile(t, dir, "custom.md", "working", "# Custom\n")

	refs, err := scanIssueFiles("", dir, func(...string) ([]byte, error) {
		t.Fatal("directory scan invoked git")
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := issueFilePaths(refs), []string{first, second}; !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %#v, want %#v", got, want)
	}

	fixtures := map[string]bool{
		"000001-slug.md":  true,
		"000001-.md":      true,
		"00001-short.md":  false,
		"000001-slug.txt": false,
		"custom.md":       false,
	}
	for name, want := range fixtures {
		if got := issueFilename(name); got != want {
			t.Errorf("issueFilename(%q) = %v, want %v", name, got, want)
		}
	}

	id, slug, ok := issueFilenameParts("000001-slug.md")
	if !ok || id != "000001" || slug != "slug" {
		t.Fatalf("parts = %q, %q, %v", id, slug, ok)
	}
	if got := issueIDPrefix("/tmp/000001-.md"); got != "000001" {
		t.Fatalf("empty-slug prefix = %q, want 000001", got)
	}
	for _, name := range []string{"00001-short.md", "abcdef-slug.md", "000001-slug.txt"} {
		if got := issueIDPrefix(name); got != "" {
			t.Errorf("issueIDPrefix(%q) = %q, want empty", name, got)
		}
	}
}

func TestScanIssueFilesSkipsDeletedUnreadableAndMalformed(t *testing.T) {
	dir := t.TempDir()
	missingStatus := filepath.Join(dir, "000001-missing-status.md")
	if err := os.WriteFile(missingStatus, []byte("---\ntitle: Missing\n---\n# Body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	malformed := filepath.Join(dir, "000002-malformed.md")
	if err := os.WriteFile(malformed, []byte("no frontmatter"), 0o644); err != nil {
		t.Fatal(err)
	}
	unreadable := filepath.Join(dir, "000003-directory.md")
	if err := os.Mkdir(unreadable, 0o755); err != nil {
		t.Fatal(err)
	}
	deleted := filepath.Join(dir, "000004-deleted.md")

	runGit := func(...string) ([]byte, error) {
		return []byte(strings.Join([]string{deleted, malformed, unreadable, missingStatus}, "\n")), nil
	}
	refs, err := scanIssueFiles("base", dir, runGit)
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0].Path != missingStatus || refs[0].Status != "" {
		t.Fatalf("refs = %#v", refs)
	}
}

func TestScanIssueFilesRetainsGitFailureFacts(t *testing.T) {
	cause := errors.New("diff failed")
	runGit := func(...string) ([]byte, error) {
		return []byte("fatal detail"), cause
	}
	_, err := scanIssueFiles("base", "workshop/issues", runGit)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(%v, cause) = false", err)
	}
	var scanErr *issueFileScanError
	if !errors.As(err, &scanErr) {
		t.Fatalf("errors.As(%T, *issueFileScanError) = false", err)
	}
	if got := string(scanErr.Output); got != "fatal detail" {
		t.Fatalf("output = %q", got)
	}
}

func writeScanIssueFile(t *testing.T, dir, name, status, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	contents := fmt.Sprintf("---\ntitle: Test\nstatus: %s\n---\n%s", status, body)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func issueFilePaths(refs []issueFileRef) []string {
	paths := make([]string, 0, len(refs))
	for _, ref := range refs {
		paths = append(paths, ref.Path)
	}
	return paths
}

func runGitCommand(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}
