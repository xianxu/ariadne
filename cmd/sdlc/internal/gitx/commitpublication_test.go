package gitx

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func sourceFile(t *testing.T, repo, path, body string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filepath.Join(repo, path)), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, path), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "--", path)
	testfix.Git(t, repo, "commit", "-qm", "selected document")
	return strings.TrimSpace(testfix.Capture(t, repo, "rev-parse", "HEAD"))
}

func TestCommitPublication_ThreeWayPreservesCallerAndRemote(t *testing.T) {
	repo, origin := trunkFixture(t, "first\n\nsecond\n\nthird\n")
	testfix.Git(t, repo, "checkout", "-qb", "feature")
	sourceFile(t, repo, "unpublished.go", "package unpublished\n")
	source := sourceFile(t, repo, "note.md", "FIRST\n\nsecond\n\nthird\n")
	pushPeerFile(t, origin, "note.md", "first\n\nsecond\n\nTHIRD\n")
	sourceFile(t, repo, "after.md", "later source work\n")
	if err := os.WriteFile(filepath.Join(repo, "note.md"), []byte("dirty staged\n"), 0644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "note.md")
	if err := os.WriteFile(filepath.Join(repo, "note.md"), []byte("dirty working\n"), 0644); err != nil {
		t.Fatal(err)
	}
	head := testfix.Capture(t, repo, "rev-parse", "HEAD")
	index := testfix.Capture(t, repo, "ls-files", "--stage")
	tf, _ := NewTrunkFile(repo, "origin", "main")
	selected, err := tf.SelectCommit(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected.Changes) != 1 || selected.Changes[0].Path != "note.md" {
		t.Fatalf("selection=%+v", selected)
	}
	result, err := tf.PublishCommit(selected)
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != CommitPublished {
		t.Fatalf("result=%+v", result)
	}
	if got := showTrunk(t, origin, "note.md"); got != "FIRST\n\nsecond\n\nTHIRD\n" {
		t.Fatalf("merged=%q", got)
	}
	if got := testfix.Capture(t, origin, "ls-tree", "--name-only", "main"); strings.Contains(got, "unpublished.go") || strings.Contains(got, "after.md") {
		t.Fatalf("leaked ancestor/descendant: %s", got)
	}
	if testfix.Capture(t, repo, "rev-parse", "HEAD") != head || testfix.Capture(t, repo, "ls-files", "--stage") != index {
		t.Fatal("caller branch/index changed")
	}
	if b, _ := os.ReadFile(filepath.Join(repo, "note.md")); string(b) != "dirty working\n" {
		t.Fatal("worktree changed")
	}
	t.Logf("fixture objects: %s", testfix.Capture(t, repo, "count-objects", "-vH"))
	started := time.Now()
	result, err = tf.PublishCommit(selected)
	t.Logf("provenance duration: %s", time.Since(started))
	if err != nil || result.Outcome != CommitAlreadyApplied {
		t.Fatalf("retry=%+v %v", result, err)
	}
	// A deliberate revert is later intent; retry must not replay the old selection.
	pushPeerFile(t, origin, "note.md", "reverted deliberately\n")
	result, err = tf.PublishCommit(selected)
	if err != nil || result.Outcome != CommitAlreadyApplied {
		t.Fatalf("retry after revert=%+v %v", result, err)
	}
	if got := showTrunk(t, origin, "note.md"); got != "reverted deliberately\n" {
		t.Fatalf("revert overwritten: %q", got)
	}
}

func TestCommitPublication_ConflictIsAtomic(t *testing.T) {
	repo, origin := trunkFixture(t, "base\n")
	source := sourceFile(t, repo, "note.md", "local\n")
	pushPeerFile(t, origin, "note.md", "remote\n")
	before := testfix.Capture(t, origin, "rev-parse", "main")
	tf, _ := NewTrunkFile(repo, "origin", "main")
	selected, err := tf.SelectCommit(source)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tf.PublishCommit(selected)
	if err == nil || !strings.Contains(err.Error(), "note.md") {
		t.Fatalf("want named conflict: %v", err)
	}
	if testfix.Capture(t, origin, "rev-parse", "main") != before {
		t.Fatal("conflict changed remote")
	}
}

func TestCommitPublication_SelectionRejectsInvalidModesAndParents(t *testing.T) {
	repo, _ := trunkFixture(t, "base\n")
	tf, _ := NewTrunkFile(repo, "origin", "main")
	root := strings.TrimSpace(testfix.Capture(t, repo, "rev-list", "--max-parents=0", "HEAD"))
	if _, err := tf.SelectCommit(root); err == nil {
		t.Fatal("root selected")
	}
	if _, err := tf.SelectCommit("--help"); err == nil {
		t.Fatal("option accepted")
	}
	if err := os.Symlink("note.md", filepath.Join(repo, "link.md")); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "link.md")
	testfix.Git(t, repo, "commit", "-qm", "symlink")
	if _, err := tf.SelectCommit("HEAD"); err == nil {
		t.Fatal("symlink selected")
	}
	testfix.Git(t, repo, "checkout", "-qb", "other", "HEAD^")
	sourceFile(t, repo, "other.md", "other\n")
	testfix.Git(t, repo, "merge", "--no-ff", "-m", "merge", "main")
	if _, err := tf.SelectCommit("HEAD"); err == nil {
		t.Fatal("merge selected")
	}
}

func TestUpdateMany_LostAcknowledgmentIsUncertainWithoutReplay(t *testing.T) {
	repo, origin := trunkFixture(t, "base\n")
	tf, _ := NewTrunkFile(repo, "origin", "main")
	original := runGitIn
	defer func() { runGitIn = original }()
	runGitIn = func(dir string, env []string, args ...string) ([]byte, []byte, error) {
		out, diagnostic, err := original(dir, env, args...)
		if len(args) > 0 && args[0] == "push" && err == nil {
			return out, []byte("lost acknowledgment"), errors.New("connection lost")
		}
		return out, diagnostic, err
	}
	calls := 0
	err := tf.UpdateMany("claim", func(*TrunkView) (TrunkWrite, error) {
		calls++
		if calls > 1 {
			return TrunkWrite{}, errors.New("already working")
		}
		return TrunkWrite{Write: map[string][]byte{"note.md": []byte("working\n")}}, nil
	})
	if !errors.Is(err, ErrPublicationUncertain) || calls != 1 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
	if got := showTrunk(t, origin, "note.md"); got != "working\n" {
		t.Fatal(got)
	}
}

func TestCommitPublication_RaceSchedules(t *testing.T) {
	for _, schedule := range []string{"peer", "rewind", "lost-ack", "unknown", "confirm-offline", "exhausted"} {
		t.Run(schedule, func(t *testing.T) {
			repo, origin := trunkFixture(t, "base\n")
			source := sourceFile(t, repo, "selected.md", "selected\n")
			tf, _ := NewTrunkFile(repo, "origin", "main")
			selected, err := tf.SelectCommit(source)
			if err != nil {
				t.Fatal(err)
			}
			original := runGitIn
			defer func() { runGitIn = original }()
			pushes := 0
			lost := false
			runGitIn = func(dir string, env []string, args ...string) ([]byte, []byte, error) {
				if args[0] == "fetch" && lost && schedule == "confirm-offline" {
					return nil, nil, errors.New("offline confirmation")
				}
				if args[0] != "push" {
					return original(dir, env, args...)
				}
				pushes++
				if schedule == "unknown" {
					return nil, nil, errors.New("transport disconnected")
				}
				if pushes == 1 || schedule == "exhausted" {
					switch schedule {
					case "peer", "exhausted":
						pushPeerFile(t, origin, "peer.md", strings.Repeat("peer\n", pushes))
					case "rewind":
						testfix.Git(t, origin, "update-ref", "refs/heads/main", strings.TrimSpace(testfix.Capture(t, origin, "rev-parse", "main^")))
					}
				}
				out, diag, err := original(dir, env, args...)
				if err == nil && (schedule == "lost-ack" || schedule == "confirm-offline") {
					lost = true
					return out, diag, errors.New("lost acknowledgment")
				}
				return out, diag, err
			}
			result, err := tf.PublishCommit(selected)
			switch schedule {
			case "unknown", "confirm-offline":
				if !errors.Is(err, ErrPublicationUncertain) || result.Outcome != CommitUncertain || pushes != 1 {
					t.Fatalf("result=%+v pushes=%d error=%v", result, pushes, err)
				}
			case "exhausted":
				if !errors.Is(err, ErrTrunkMoved) || pushes != 3 {
					t.Fatalf("pushes=%d err=%v", pushes, err)
				}
			default:
				if err != nil || result.Outcome != CommitPublished {
					t.Fatalf("result=%+v err=%v", result, err)
				}
				if schedule == "lost-ack" && pushes != 1 {
					t.Fatal("lost ack replayed")
				}
				if schedule == "rewind" && strings.Contains(testfix.Capture(t, origin, "ls-tree", "--name-only", "main"), "note.md") {
					t.Fatal("rewind resurrected old remote files")
				}
			}
		})
	}
}

func TestCommitPublication_DeleteRenameAndEmpty(t *testing.T) {
	repo, origin := trunkFixture(t, "base\n")
	testfix.Git(t, repo, "mv", "note.md", "renamed.md")
	if err := os.WriteFile(filepath.Join(repo, "empty.md"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "empty.md")
	testfix.Git(t, repo, "commit", "-qm", "rename and empty")
	tf, _ := NewTrunkFile(repo, "origin", "main")
	selected, err := tf.SelectCommit("HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if len(selected.Changes) != 3 {
		t.Fatalf("changes=%+v", selected.Changes)
	}
	result, err := tf.PublishCommit(selected)
	if err != nil || result.Outcome != CommitPublished {
		t.Fatalf("%+v %v", result, err)
	}
	if got := testfix.Capture(t, origin, "ls-tree", "--name-only", "main"); !strings.Contains(got, "renamed.md") || !strings.Contains(got, "empty.md") || strings.Contains(got, "note.md") {
		t.Fatal(got)
	}
	testfix.Git(t, repo, "rm", "renamed.md")
	testfix.Git(t, repo, "commit", "-qm", "delete")
	selected, err = tf.SelectCommit("HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tf.PublishCommit(selected); err != nil {
		t.Fatal(err)
	}
	if got := testfix.Capture(t, origin, "ls-tree", "--name-only", "main"); strings.Contains(got, "renamed.md") {
		t.Fatal(got)
	}
}

func TestCommitPublication_NoChangeAndDirectAncestry(t *testing.T) {
	for _, direct := range []bool{false, true} {
		t.Run(map[bool]string{false: "same-bytes", true: "direct"}[direct], func(t *testing.T) {
			repo, origin := trunkFixture(t, "base\n")
			source := sourceFile(t, repo, "note.md", "local\n")
			if direct {
				testfix.Git(t, repo, "push", "origin", "main")
			} else {
				pushPeerFile(t, origin, "note.md", "local\n")
			}
			before := testfix.Capture(t, origin, "rev-parse", "main")
			tf, _ := NewTrunkFile(repo, "origin", "main")
			selected, err := tf.SelectCommit(source)
			if err != nil {
				t.Fatal(err)
			}
			result, err := tf.PublishCommit(selected)
			if err != nil {
				t.Fatal(err)
			}
			want := CommitNoChange
			if direct {
				want = CommitAlreadyApplied
			}
			if result.Outcome != want {
				t.Fatalf("result=%+v want %s", result, want)
			}
			if testfix.Capture(t, origin, "rev-parse", "main") != before {
				t.Fatal("no-change committed")
			}
		})
	}
}

func TestCommitPublication_ProvenanceFailuresRefuse(t *testing.T) {
	for _, query := range []string{"merge-base", "log"} {
		t.Run(query, func(t *testing.T) {
			repo, origin := trunkFixture(t, "base\n")
			source := sourceFile(t, repo, "selected.md", "selected\n")
			tf, _ := NewTrunkFile(repo, "origin", "main")
			selected, err := tf.SelectCommit(source)
			if err != nil {
				t.Fatal(err)
			}
			original := runGitInContext
			defer func() { runGitInContext = original }()
			runGitInContext = func(ctx context.Context, dir string, env []string, args ...string) ([]byte, []byte, error) {
				if args[0] == query {
					if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > 30*time.Second {
						t.Fatal("history query unbounded")
					}
					return nil, nil, context.DeadlineExceeded
				}
				return original(ctx, dir, env, args...)
			}
			before := testfix.Capture(t, origin, "rev-parse", "main")
			_, err = tf.PublishCommit(selected)
			if err == nil {
				t.Fatal("failed provenance treated as absent")
			}
			if testfix.Capture(t, origin, "rev-parse", "main") != before {
				t.Fatal("published on failed provenance")
			}
		})
	}
}

func TestCommitPublication_ProvenanceRequiresExactTrailer(t *testing.T) {
	repo, origin := trunkFixture(t, "base\n")
	source := sourceFile(t, repo, "selected.md", "selected\n")
	peer := testfix.Repo(t)
	testfix.Git(t, peer, "remote", "add", "origin", origin)
	testfix.Git(t, peer, "fetch", "origin", "main")
	testfix.Git(t, peer, "checkout", "-qB", "main", "origin/main")
	testfix.Git(t, peer, "commit", "--allow-empty", "-qm", "Source-Commit: "+source+"\n\nThis is a subject, not provenance.")
	testfix.Git(t, peer, "push", "origin", "main")
	tf, _ := NewTrunkFile(repo, "origin", "main")
	selected, err := tf.SelectCommit(source)
	if err != nil {
		t.Fatal(err)
	}
	result, err := tf.PublishCommit(selected)
	if err != nil || result.Outcome != CommitPublished {
		t.Fatalf("subject incorrectly counted as provenance: %+v %v", result, err)
	}
}

func TestCommitPublication_PreservesSourceMessage(t *testing.T) {
	repo, origin := trunkFixture(t, "base\n")
	source := sourceFile(t, repo, "selected.md", "selected\n")
	tf, _ := NewTrunkFile(repo, "origin", "main")
	selected, err := tf.SelectCommit(source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tf.PublishCommit(selected); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(testfix.Capture(t, origin, "show", "-s", "--format=%s", "main")); got != "selected document" {
		t.Fatalf("source subject lost: %q", got)
	}
}

func TestCommitPublication_MalformedObjectResponseRefuses(t *testing.T) {
	repo, _ := trunkFixture(t, "base\n")
	tf, _ := NewTrunkFile(repo, "origin", "main")
	original := runGitIn
	defer func() { runGitIn = original }()
	runGitIn = func(dir string, env []string, args ...string) ([]byte, []byte, error) {
		if args[0] == "rev-parse" {
			return []byte("--malformed-object"), nil, nil
		}
		if args[0] == "rev-list" {
			return []byte("--malformed-object abc"), nil, nil
		}
		if args[0] == "diff-tree" || args[0] == "show" {
			return nil, nil, nil
		}
		return original(dir, env, args...)
	}
	if _, err := tf.SelectCommit("HEAD"); err == nil {
		t.Fatal("malformed Git object response accepted")
	}
}

func TestUpdateMany_IdenticalPeerCandidateDoesNotAcquireClaim(t *testing.T) {
	repo, _ := trunkFixture(t, "base\n")
	t.Setenv("GIT_AUTHOR_DATE", "2026-01-01T00:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-01-01T00:00:00Z")
	tf, _ := NewTrunkFile(repo, "origin", "main")
	set := TrunkWrite{Write: map[string][]byte{"note.md": []byte("working\n")}}
	taken := errors.New("remote issue already working")
	calls := 0
	err := tf.UpdateMany("identical claim", func(*TrunkView) (TrunkWrite, error) {
		calls++
		if calls > 1 {
			return TrunkWrite{}, taken
		}
		// Identical identity, tree, message and second-resolution time yield the same
		// Git object. The peer wins before this caller sends its stale expected ref.
		if err := tf.UpdateMany("identical claim", func(*TrunkView) (TrunkWrite, error) { return set, nil }); err != nil {
			t.Fatal(err)
		}
		return set, nil
	})
	if !errors.Is(err, taken) || calls != 2 {
		t.Fatalf("identical peer candidate mistaken for own claim: calls=%d err=%v", calls, err)
	}
}

func TestCommitPublication_UnpublishedPrerequisite(t *testing.T) {
	repo, origin := trunkFixture(t, "base\n")
	first := sourceFile(t, repo, "depends.md", "first\n")
	second := sourceFile(t, repo, "depends.md", "second\n")
	tf, _ := NewTrunkFile(repo, "origin", "main")
	selected, err := tf.SelectCommit(second)
	if err != nil {
		t.Fatal(err)
	}
	before := testfix.Capture(t, origin, "rev-parse", "main")
	if _, err = tf.PublishCommit(selected); err == nil || !strings.Contains(err.Error(), "depends.md") {
		t.Fatalf("missing prerequisite did not refuse with path: %v", err)
	}
	if testfix.Capture(t, origin, "rev-parse", "main") != before {
		t.Fatal("prerequisite widened selected commit")
	}
	prerequisite, err := tf.SelectCommit(first)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tf.PublishCommit(prerequisite); err != nil {
		t.Fatal(err)
	}
	if _, err = tf.PublishCommit(selected); err != nil {
		t.Fatal(err)
	}
	if got := showTrunk(t, origin, "depends.md"); got != "second\n" {
		t.Fatal(got)
	}
}
