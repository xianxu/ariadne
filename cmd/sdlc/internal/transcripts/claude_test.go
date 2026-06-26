package transcripts

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// cwd → Claude Code transcript-folder encoding ('/' and '.' → '-'). Moved from
// cmd/sdlc/actual_test.go (#134), renamed cwdToTranscriptDir → cwdToClaudeDir.
func TestCwdToClaudeDir(t *testing.T) {
	cases := map[string]string{
		"/Users/x/workspace/nous":    "-Users-x-workspace-nous",
		"/Users/x/workspace/brain":   "-Users-x-workspace-brain",
		"/Users/x/.claude/projects":  "-Users-x--claude-projects", // leading '/.' → '--'
		"/w/worktree/ariadne-000040": "-w-worktree-ariadne-000040",
	}
	for in, want := range cases {
		if got := cwdToClaudeDir(in); got != want {
			t.Errorf("cwdToClaudeDir(%q) = %q, want %q", in, got, want)
		}
	}
}

// The Claude harness contributes one dir per cwd, existing folders only, in cwd
// order — never an unrelated folder that merely exists on disk. Moved from
// TestSelectActualDirs (#134).
func TestClaudeHarnessSources(t *testing.T) {
	root := t.TempDir()
	repo := "/w/nous"
	brain := "/w/brain"
	mk := func(p string) {
		if err := os.Mkdir(filepath.Join(root, cwdToClaudeDir(p)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mk(repo)
	mk(brain)
	mk("/w/pair") // unrelated — present on disk but NOT in cwds → must be excluded

	got := ClaudeHarness(root).Sources([]string{brain, repo})
	want := Sources{Dirs: []string{
		filepath.Join(root, cwdToClaudeDir(brain)),
		filepath.Join(root, cwdToClaudeDir(repo)),
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Sources = %+v, want %+v (brain+repo only, no unrelated)", got, want)
	}

	// A cwd whose folder doesn't exist is silently skipped (not invented).
	got2 := ClaudeHarness(root).Sources([]string{"/w/does-not-exist", brain})
	want2 := Sources{Dirs: []string{filepath.Join(root, cwdToClaudeDir(brain))}}
	if !reflect.DeepEqual(got2, want2) {
		t.Errorf("missing folder should be skipped; got %+v want %+v", got2, want2)
	}

	// Empty root → no dirs, no panic.
	if got3 := ClaudeHarness("").Sources([]string{repo}); len(got3.Dirs) != 0 {
		t.Errorf("empty root should yield no dirs, got %+v", got3)
	}
}
