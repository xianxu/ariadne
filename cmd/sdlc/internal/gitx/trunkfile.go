// trunkfile.go — read and compare-and-swap-write ONE path on a remote branch,
// with no working tree (ariadne#209).
//
// "No working tree" is precise: no git CHECKOUT is read or written. Two private
// temp files are used — one for the index, one to hand `hash-object` content,
// since it reads a file rather than stdin — and neither is a working tree nor
// lives inside the caller's repo.
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
	"bytes"
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
// stdout and stderr are returned SEPARATELY, not combined. Combining them looked
// fine until a repo with `*.md text eol=lf` made git print "CRLF will be replaced
// by LF" on hash-object — which landed inside the parsed blob hash, and would
// have landed inside file CONTENT on every read. A shim that returns a value and
// a diagnostic in one buffer cannot be used safely by a caller that parses.
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
var runGitIn = func(dir string, env []string, args ...string) (stdout, stderr []byte, err error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	var out, errBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errBuf
	err = cmd.Run()
	return out.Bytes(), errBuf.Bytes(), err
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
	_, errOut, err := runGitIn(t.dir, nil, "fetch", "--quiet", t.remote,
		"+refs/heads/"+t.branch+":"+t.trackingRef())
	return errOut, err
}

// Read returns the file's bytes on the trunk, requiring a reachable remote.
// Callers that can tolerate staleness should use ReadDegraded.
func (t *TrunkFile) Read(path string) ([]byte, error) {
	if out, err := t.fetch(); err != nil {
		return nil, offlineError(t.remote, err, out)
	}
	return t.readRef(path)
}

// ReadDegraded returns the file's bytes and a non-empty warning when the fetch
// failed and the answer came from a stale tracking ref.
//
// The read/write asymmetry is deliberate and is the policy issueids.go:40-49
// already settled for id allocation: creating or inspecting is not something to
// refuse when the network is down, so degrade LOUDLY; but a CAS push has no base
// to compare against, so a write must refuse. One offline policy in this binary,
// not two (ARCH-DRY).
func (t *TrunkFile) ReadDegraded(path string) ([]byte, string, error) {
	var warn string
	if out, err := t.fetch(); err != nil {
		warn = fmt.Sprintf(
			"%s unreachable — %q read from the stale %s, which may be behind "+
				"anything published since the last fetch (%s)",
			t.remote, path, t.trackingRef(), FirstLine(string(out)))
	}
	if warn != "" && !t.refExists(t.trackingRef()) {
		// Never fetched (or no remote at all): fall back to the local branch so a
		// repo that has never talked to an origin still reads something, and say so.
		b, err := t.readLocal(path)
		return b, warn + "; no " + t.trackingRef() + ", used local " + t.branch, err
	}
	b, err := t.readRef(path)
	return b, warn, err
}

// refExists reports whether a ref resolves, by exit code.
func (t *TrunkFile) refExists(ref string) bool {
	_, _, err := runGitIn(t.dir, nil, "rev-parse", "--verify", "--quiet", ref)
	return err == nil
}

// offlineError names the cause rather than surfacing raw git output, so the
// refusal reads as a next-action spec.
func offlineError(remote string, err error, out []byte) error {
	return fmt.Errorf("%s unreachable (offline?): %v\n%s", remote, err, FirstLine(string(out)))
}

// FirstLine keeps a git failure to one line: fetch errors are several lines of
// remote diagnostics, and a warning that scrolls reads as noise, not a warning.
//
// Exported and homed here because cmd/sdlc had its own copy (issueids.go) and
// gitx cannot import package main. gitx is where the git-invocation vocabulary
// already lives, so the shared definition belongs on this side of the seam
// (ARCH-DRY).
func FirstLine(s string) string {
	for _, ln := range strings.Split(s, "\n") {
		if ln = strings.TrimSpace(ln); ln != "" {
			return ln
		}
	}
	return "no output"
}

// readLocal reads the path from the local branch, for a repo with no remote.
func (t *TrunkFile) readLocal(path string) ([]byte, error) {
	if !t.exists(t.branch, path) {
		return nil, nil
	}
	out, errOut, err := runGitIn(t.dir, nil, "cat-file", "blob", t.branch+":"+path)
	if err != nil {
		return nil, fmt.Errorf("read %s from %s: %v\n%s", path, t.branch, err, errOut)
	}
	return out, nil
}

// readRef reads the path from the tracking ref. A path absent from the trunk is
// NOT an error — it reads as empty, so a first-ever write needs no special case.
func (t *TrunkFile) readRef(path string) ([]byte, error) {
	if !t.exists(t.trackingRef(), path) {
		return nil, nil // absent on the trunk reads as empty: a first write needs no special case
	}
	out, errOut, err := runGitIn(t.dir, nil, "cat-file", "blob", t.trackingRef()+":"+path)
	if err != nil {
		return nil, fmt.Errorf("read %s from %s: %v\n%s", path, t.trackingRef(), err, errOut)
	}
	return out, nil
}

// exists reports whether <ref>:<path> is in the tree, via `cat-file -e`'s EXIT
// CODE rather than by matching git's prose.
//
// The first version of this matched phrases like "does not exist" — which works
// until a git release rewords them, and worse, silently reclassifies a real
// failure as "absent" and hands the caller an empty file. Existence has a machine
// interface; use it (the same reason signs() asks for --type=bool below).
func (t *TrunkFile) exists(ref, path string) bool {
	_, _, err := runGitIn(t.dir, nil, "cat-file", "-e", ref+":"+path)
	return err == nil
}

// tempIndexPath returns an ABSOLUTE path for GIT_INDEX_FILE, inside a private
// temp DIRECTORY.
//
// Absolute matters: a relative value resolves against the child's working
// directory, which is cmd.Dir here — so a relative path would silently land
// inside the caller's repo. And it must never be $GIT_DIR/index, which would
// corrupt whatever checkout shares the git dir.
//
// The directory matters too. Creating a temp FILE and deleting it so git can
// make its own leaves the name unclaimed in a world-writable /tmp between the
// remove and git's create — anyone can plant a symlink there and redirect the
// index write (ARCH-SECURE: this process does not own /tmp). A 0700 directory
// created atomically by MkdirTemp has no such window, and the name inside it is
// ours alone.
func tempIndexPath() (string, func(), error) {
	dir, err := os.MkdirTemp("", "sdlc-trunk-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { os.RemoveAll(dir) }
	abs, err := filepath.Abs(filepath.Join(dir, "index"))
	if err != nil {
		return "", cleanup, err
	}
	return abs, cleanup, nil
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
			// Includes the offline refusal: a CAS push has no base to compare
			// against, so unlike a read this cannot degrade.
			return err
		}
		next, err := transform(old)
		if err != nil {
			return err // caller's error, surfaced unwrapped so errors.Is works
		}
		base, err := t.resolve(t.trackingRef())
		if err != nil {
			return err
		}
		out, err := t.commitAndPush(path, msg, next, base)
		if err == nil {
			return nil
		}
		// Retry only when the TRUNK ACTUALLY MOVED — observed by re-resolving the
		// ref, not by matching git's rejection prose. The first version matched a
		// bare "rejected", which git emits for refusals that retrying cannot fix
		// (a hook decline, no push permission, a stale-info tag clobber); those
		// burned the whole attempt budget and then reported a contention story
		// that never happened.
		if _, ferr := t.fetch(); ferr != nil {
			return fmt.Errorf("publish %s: %v\n%s", path, err, out)
		}
		now, rerr := t.resolve(t.trackingRef())
		if rerr != nil || now == base {
			return fmt.Errorf("publish %s: %v\n%s", path, err, out)
		}
		lastRejection = out
	}
	return fmt.Errorf("publish %s: the trunk moved under %d attempts; last rejection:\n%s",
		path, maxUpdateAttempts, lastRejection)
}

// resolve returns the SHA a ref points at.
func (t *TrunkFile) resolve(ref string) (string, error) {
	out, errOut, err := runGitIn(t.dir, nil, "rev-parse", "--verify", ref)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %v\n%s", ref, err, errOut)
	}
	return strings.TrimSpace(string(out)), nil
}

// commitAndPush builds the tree in a temp index and pushes the new commit at the
// branch. Returns git's combined output so the caller can classify the failure.
func (t *TrunkFile) commitAndPush(path, msg string, content []byte, base string) ([]byte, error) {
	idx, cleanupIdx, err := tempIndexPath()
	defer cleanupIdx()
	if err != nil {
		return nil, err
	}
	env := []string{"GIT_INDEX_FILE=" + idx}

	if _, errOut, err := runGitIn(t.dir, env, "read-tree", base); err != nil {
		return errOut, err
	}

	blobFile, cleanupBlob, err := writeTemp(content)
	defer cleanupBlob()
	if err != nil {
		return nil, err
	}
	// --path (not bare hash-object) so .gitattributes filters and EOL
	// normalization for THIS path apply. Without it the stored blob can differ
	// from what a checkout of the resulting commit produces.
	out, errOut, err := runGitIn(t.dir, env, "hash-object", "-w", "--path", path, blobFile)
	if err != nil {
		return errOut, err
	}
	blob := strings.TrimSpace(string(out))

	if _, errOut, err := runGitIn(t.dir, env, "update-index", "--add",
		"--cacheinfo", "100644,"+blob+","+path); err != nil {
		return errOut, err
	}
	out, errOut, err = runGitIn(t.dir, env, "write-tree")
	if err != nil {
		return errOut, err
	}
	tree := strings.TrimSpace(string(out))

	args := []string{"commit-tree", tree, "-p", base, "-m", msg}
	if t.signs() {
		args = append([]string{"commit-tree", "-S"}, args[1:]...)
	}
	out, errOut, err = runGitIn(t.dir, nil, args...)
	if err != nil {
		return errOut, err
	}
	commit := strings.TrimSpace(string(out))

	_, errOut, err = runGitIn(t.dir, nil, "push", t.remote, commit+":refs/heads/"+t.branch)
	return errOut, err
}

// signs reports whether this repo signs commits. A signing repo that silently
// produced unsigned commits through this path would be a regression, and
// commit-tree does not read commit.gpgsign on its own.
//
// --type=bool matters: git accepts 1/yes/on/true and stores the string VERBATIM,
// so `--get` on a repo configured with `yes` returns "yes" and a naive == "true"
// yields exactly the silent unsigned commit this function exists to prevent.
func (t *TrunkFile) signs() bool {
	out, _, err := runGitIn(t.dir, nil, "config", "--type=bool", "--get", "commit.gpgsign")
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// writeTemp stages content for hash-object, which reads a file rather than
// stdin. Unlike the index path this file is created and kept by us, so
// CreateTemp's atomic O_EXCL is sufficient — there is no delete-then-recreate
// window for anyone to occupy.
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
