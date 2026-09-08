// trunkfile.go — read and compare-and-swap-write ONE path on a remote branch,
// with no working tree (ariadne#209).
//
// Why this exists: publishing a small file to the trunk from a feature branch
// currently routes through whatever checkout has main out (`syncViaMainWorktree`,
// claim.go), which can be dirty, mid-rebase, owned by another actor, or simply
// absent. Every guard on that route exists to make a SHARED WORKING DIRECTORY
// safe. Here there is no working directory: fetch, build the tree in a temp
// index, commit-tree, and push the commit at the ref. `push <commit>:main` IS the
// concurrency primitive — a compare-and-swap that a local cleanliness check
// cannot approximate, because it sees the remote and the check only sees here.
//
// The retry loop is generic and the TRANSFORM decides mergeability (ARCH-ORDER):
// Update re-reads and re-calls the transform on a moved base, so a caller whose
// transform replays an intent ("add this line") preserves a peer's concurrent
// edit, while a caller whose transform sets content keeps last-writer-wins. One
// primitive, per-caller semantics — which is what lets ariadne#207 consume this
// rather than growing a second retry loop (ARCH-DRY).
package gitx

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// runGitIn runs git in `dir` with `env` appended to the environment, returning
// combined output. A package-level var so tests can drive failure paths.
//
// It is a SIBLING of `run` (window.go), not a widening of it. The capability
// audit for #209 enumerated every exec.Cmd parameter the plumbing needs against
// `run`'s signature: `run` is exec.Command(...).Output(), which carries argv and
// nothing else. Dir is needed (otherwise git runs against the process cwd — in a
// test, the real repo and its real origin); Env is needed (GIT_INDEX_FILE, since
// only read-tree has an --index-output escape); stderr is needed (three call
// sites render git's own text). Stdin is not (blobs come from a temp file via
// `hash-object --path`), and no git call in this binary takes a context today, so
// introducing one here would be a second convention.
//
// `run` is left alone deliberately: its existing callers were written against
// .Output() semantics, and folding stderr into strings they parse would break
// them silently.
var runGitIn = func(dir string, env []string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	return cmd.CombinedOutput()
}

// TrunkFile reads and CAS-writes one path on <remote>/<branch>.
type TrunkFile struct {
	dir    string // repo to run git in — never empty, see NewTrunkFile
	remote string
	branch string
}

// NewTrunkFile refuses an empty dir.
//
// This guard is the reason the type is safe to test: gitx's other shim carries
// no Dir, so a TrunkFile that forgot to pass one would silently operate on the
// process cwd — fetching from and pushing to the real origin during `go test`.
// Making that unrepresentable beats remembering not to do it.
func NewTrunkFile(dir, remote, branch string) (*TrunkFile, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("trunkfile: dir must not be empty (git would run against the process cwd)")
	}
	if remote == "" || branch == "" {
		return nil, errors.New("trunkfile: remote and branch are required")
	}
	return &TrunkFile{dir: dir, remote: remote, branch: branch}, nil
}

// trackingRef is the local remote-tracking ref this type treats as the base.
func (t *TrunkFile) trackingRef() string {
	return "refs/remotes/" + t.remote + "/" + t.branch
}

// fetch updates the tracking ref. Fetching INTO the ref explicitly (rather than
// `git fetch origin main`, which only guarantees FETCH_HEAD) is the same form
// issueids.go uses.
func (t *TrunkFile) fetch() ([]byte, error) {
	return runGitIn(t.dir, nil, "fetch", "--quiet", t.remote,
		"+refs/heads/"+t.branch+":"+t.trackingRef())
}

// Read returns the file's bytes on the trunk. A path absent from the trunk is
// NOT an error — it reads as empty, so a first-ever write needs no special case.
func (t *TrunkFile) Read(path string) ([]byte, error) {
	if _, err := t.fetch(); err != nil {
		// Offline is handled by the caller-facing ReadDegraded; a bare Read
		// still tries the stale ref rather than failing outright.
		_ = err
	}
	out, err := runGitIn(t.dir, nil, "cat-file", "blob", t.trackingRef()+":"+path)
	if err != nil {
		if isMissingPath(out) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s from %s: %v\n%s", path, t.trackingRef(), err, out)
	}
	return out, nil
}

// isMissingPath distinguishes "the tree has no such path" from a real failure.
// git phrases this several ways across versions, so match on the stable parts.
func isMissingPath(out []byte) bool {
	s := string(out)
	return strings.Contains(s, "does not exist") ||
		strings.Contains(s, "exists on disk, but not in") ||
		strings.Contains(s, "Not a valid object name")
}

// tempIndexPath returns an ABSOLUTE path for GIT_INDEX_FILE.
//
// Absolute matters: a relative value resolves against the child's working
// directory, which is cmd.Dir here — so a relative path would silently land
// inside the caller's repo. And it must never be $GIT_DIR/index, which would
// corrupt whatever checkout shares the git dir.
func tempIndexPath(dir string) (string, func(), error) {
	f, err := os.CreateTemp("", "sdlc-trunk-index-*")
	if err != nil {
		return "", func() {}, err
	}
	p := f.Name()
	f.Close()
	os.Remove(p) // git wants to create it itself
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", func() { os.Remove(p) }, err
	}
	return abs, func() { os.Remove(abs) }, nil
}

// maxUpdateAttempts bounds the CAS retry. Three is the same bound ariadne#207
// specs; past it the contention is not transient and a caller looping forever is
// worse than a refusal that names why.
const maxUpdateAttempts = 3

// Update applies `transform` to the file's trunk content and pushes the result,
// retrying on a moved base.
//
// The retry RE-READS and RE-CALLS the transform rather than re-pushing the bytes
// it built the first time. That is the entire concurrency contract: a transform
// that replays an intent ("append this line") preserves whatever a peer landed in
// the meantime, while one that sets content keeps last-writer-wins. The loop does
// not know or care which — mergeability is the caller's property, expressed in
// the transform (ARCH-ORDER: the interleaving policy is written down at the seam
// where a reader can see it, not spread across call sites).
//
// Nondeterminism enters at exactly one place — the order peers' pushes reach the
// remote — and it is reproduced in tests by a real bare origin, not by timing.
func (t *TrunkFile) Update(path, msg string, transform func([]byte) ([]byte, error)) error {
	var lastRejection []byte
	for attempt := 1; attempt <= maxUpdateAttempts; attempt++ {
		old, err := t.Read(path)
		if err != nil {
			return err
		}
		next, err := transform(old)
		if err != nil {
			return err // caller's error, surfaced unwrapped so errors.Is works
		}
		out, err := t.commitAndPush(path, msg, next)
		if err == nil {
			return nil
		}
		if !isNonFastForward(out) {
			return fmt.Errorf("publish %s: %v\n%s", path, err, out)
		}
		lastRejection = out
	}
	return fmt.Errorf("publish %s: the trunk moved under %d attempts; last rejection:\n%s",
		path, maxUpdateAttempts, lastRejection)
}

// isNonFastForward reports whether git refused the push because the ref moved —
// the CAS failure this type retries — as opposed to any other push failure,
// which is not retryable and must surface immediately.
func isNonFastForward(out []byte) bool {
	s := string(out)
	return strings.Contains(s, "non-fast-forward") ||
		strings.Contains(s, "fetch first") ||
		strings.Contains(s, "rejected")
}

// commitAndPush builds the tree in a temp index and pushes the new commit at the
// branch. Returns git's combined output so the caller can classify the failure.
func (t *TrunkFile) commitAndPush(path, msg string, content []byte) ([]byte, error) {
	idx, cleanupIdx, err := tempIndexPath(t.dir)
	defer cleanupIdx()
	if err != nil {
		return nil, err
	}
	env := []string{"GIT_INDEX_FILE=" + idx}

	if out, err := runGitIn(t.dir, env, "read-tree", t.trackingRef()); err != nil {
		return out, err
	}

	blobFile, cleanupBlob, err := writeTemp(content)
	defer cleanupBlob()
	if err != nil {
		return nil, err
	}
	// --path (not bare hash-object) so .gitattributes filters and EOL
	// normalization for THIS path apply. Without it the stored blob can differ
	// from what a checkout of the resulting commit produces.
	out, err := runGitIn(t.dir, env, "hash-object", "-w", "--path", path, blobFile)
	if err != nil {
		return out, err
	}
	blob := strings.TrimSpace(string(out))

	if out, err := runGitIn(t.dir, env, "update-index", "--add",
		"--cacheinfo", "100644,"+blob+","+path); err != nil {
		return out, err
	}
	out, err = runGitIn(t.dir, env, "write-tree")
	if err != nil {
		return out, err
	}
	tree := strings.TrimSpace(string(out))

	args := []string{"commit-tree", tree, "-p", t.trackingRef(), "-m", msg}
	if t.signs() {
		args = append([]string{"commit-tree", "-S"}, args[1:]...)
	}
	out, err = runGitIn(t.dir, nil, args...)
	if err != nil {
		return out, err
	}
	commit := strings.TrimSpace(string(out))

	return runGitIn(t.dir, nil, "push", t.remote, commit+":refs/heads/"+t.branch)
}

// signs reports whether this repo signs commits. A signing repo that silently
// produced unsigned commits through this path would be a regression, and
// commit-tree does not read commit.gpgsign on its own.
func (t *TrunkFile) signs() bool {
	out, err := runGitIn(t.dir, nil, "config", "--get", "commit.gpgsign")
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// writeTemp stages content for hash-object, which reads a file rather than stdin.
func writeTemp(content []byte) (string, func(), error) {
	f, err := os.CreateTemp("", "sdlc-trunk-blob-*")
	if err != nil {
		return "", func() {}, err
	}
	p := f.Name()
	cleanup := func() { os.Remove(p) }
	if _, err := f.Write(content); err != nil {
		f.Close()
		return "", cleanup, err
	}
	return p, cleanup, f.Close()
}
