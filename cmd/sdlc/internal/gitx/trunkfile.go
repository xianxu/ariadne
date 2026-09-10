// trunkfile.go — read and compare-and-swap-write ONE path on a remote branch,
// with no working tree (ariadne#209).
//
// "No working tree" is precise: no git CHECKOUT is read or written. Two private
// temp files are used — one for the index, one to hand `hash-object` content,
// since it reads a file rather than stdin — and neither is a working tree nor
// lives inside the caller's repo.
//
// Why this exists: publishing a small file to the trunk from a feature branch
// used to route through whatever checkout had main out (`syncViaMainWorktree`,
// deleted in ariadne#207), which could be dirty, mid-rebase, owned by another
// actor, or simply absent. Every guard on that route exists to make a SHARED WORKING DIRECTORY
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
// stdout and stderr separately. A package-level var so tests can drive failure paths.
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

// A git query that can legitimately answer "absent" has THREE outcomes, and this
// file classifies every one of them by exit code rather than collapsing the
// nonzero cases into a benign default.
//
// The rule exists because the same defect was shipped three times here in
// different disguises: matching "does not exist" in git's prose, then
// `cat-file -e`'s bare nonzero, then `rev-parse`'s. Each turned "I could not
// tell" into "it is not there", and downstream that becomes an empty file, a
// skipped signature, or a silently truncated trunk. A bool cannot hold the
// difference, so nothing in this file returns one for a question git can fail to
// answer (ARCH-ORDER).
//
// git's convention on the queries used here: exit 1 means the thing is absent;
// any other nonzero is a failure that must propagate. (`config --get` also uses 1
// for a malformed key, which is safe to fold in only because every key this file
// passes is a literal, so "malformed" would be our own bug, not a runtime state.)
const gitAbsentExit = 1

// gitExitCode returns the child's exit status, or -1 when the failure was not an
// exit status at all — git missing from PATH, permission denied, a killed
// process. Those must never read as "absent", and without this they would: an
// ExitError type assertion that fails leaves a zero value.
func gitExitCode(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
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

// localRef is the local BRANCH, fully qualified.
//
// Qualification is not cosmetic. An unqualified "main" resolves through git's
// ref precedence, and with both refs/tags/main and refs/heads/main present
// `rev-parse --verify main` and `cat-file blob main:path` both return the TAG —
// so the offline fallback could serve a tag's bytes while announcing it had used
// the local branch. The same reason trackingRef is spelled out in full.
func (t *TrunkFile) localRef() string { return "refs/heads/" + t.branch }

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
	return t.readFrom(t.trackingRef(), path)
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
			"could not fetch from %s — %q read from the stale %s, which may be behind "+
				"anything published since the last fetch (git says: %s)",
			t.remote, path, t.trackingRef(), FirstLine(string(out)))
	}
	tracking, terr := t.refPresent(t.trackingRef())
	if terr != nil {
		return nil, warn, terr
	}
	if warn != "" && !tracking {
		// Never fetched (or no remote at all): fall back to the local branch so a
		// repo that has never talked to an origin still reads something, and say so.
		b, err := t.readFrom(t.localRef(), path)
		return b, warn + "; no " + t.trackingRef() + ", used local " + t.branch, err
	}
	b, err := t.readFrom(t.trackingRef(), path)
	return b, warn, err
}

// refPresent reports whether a ref resolves, distinguishing absent from
// could-not-tell per the rule above.
func (t *TrunkFile) refPresent(ref string) (bool, error) {
	sha, err := t.resolveOpt(ref)
	return sha != "", err
}

// resolveOpt is the single rev-parse question: the SHA, or "" when the ref is
// absent, or an error when git could not tell. refPresent and resolve were two
// functions asking it separately, and Update ran both per attempt.
func (t *TrunkFile) resolveOpt(ref string) (string, error) {
	out, errOut, err := runGitIn(t.dir, nil, "rev-parse", "--verify", "--quiet", ref)
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	if gitExitCode(err) == gitAbsentExit {
		return "", nil
	}
	return "", fmt.Errorf("rev-parse %s: %v\n%s", ref, err, errOut)
}

// offlineError reports a failed fetch WITHOUT asserting why it failed.
//
// It used to say "unreachable (offline?)" for every failure, which is the same
// rule BR-37 turned on commit subjects: a message may only name state the code
// observed. A rejected credential, a missing remote, a bad refspec and a dead
// network all land here, and telling an operator with a working connection that
// their remote is unreachable sends them to debug the wrong thing. Git's own
// first line is the observation; the offline case is offered as a possibility,
// not a diagnosis.
func offlineError(remote string, err error, out []byte) error {
	return fmt.Errorf("could not fetch from %s (offline, auth, or a bad ref — git says): %v\n%s",
		remote, err, FirstLine(string(out)))
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

// pathPresent answers whether <ref>:<path> is in the tree, in THREE states:
// present, absent, or "could not tell" — see the rule at the top of this file.
//
// `ls-tree` is the interface that separates them: exit 0 with empty output means
// absent, exit 0 with output means present, and a non-zero exit is a real failure
// that must PROPAGATE rather than being flattened into "absent". Flattening it is
// how the trunk got silently truncated: the caller received an empty file, the
// transform appended to nothing, and Update published only the new line.
func (t *TrunkFile) pathPresent(ref, path string) (bool, error) {
	_, present, err := t.entryOf(ref, path)
	return present, err
}

// entryOf is the single ls-tree question — mode AND presence in one call.
//
// pathPresent and modeOf were two functions asking git the same thing with
// different flags, and Update ran both per attempt. Two functions differing only
// in which field of one answer they keep are one function (ARCH-DRY), and the
// split also doubled the git calls on the hot path (ARCH-CONSTRAINTS).
func (t *TrunkFile) entryOf(ref, path string) (mode string, present bool, err error) {
	out, errOut, err := runGitIn(t.dir, nil, "ls-tree", "--end-of-options", ref, "--", path)
	if err != nil {
		return "", false, fmt.Errorf("ls-tree %s -- %s: %v\n%s", ref, path, err, errOut)
	}
	f := strings.Fields(strings.TrimSpace(string(out)))
	if len(f) == 0 {
		return "", false, nil
	}
	return f[0], true, nil
}

// readFrom reads <ref>:<path>, answering empty for a ref or path that is simply
// not there and propagating anything it could not determine.
//
// This was two functions — readRef and readLocal — differing only in which ref
// they passed, and the ref-presence guard existed on just one of them. Two
// functions differing only in a parameter's value are one function; when they
// diverge, the divergence is a bug in the one that lacks it, not a feature of the
// one that has it (ARCH-DRY). Here the missing guard was exactly that bug: the
// trunk path had no protection against an unresolvable ref.
func (t *TrunkFile) readFrom(ref, path string) ([]byte, error) {
	refThere, err := t.refPresent(ref)
	if err != nil {
		return nil, err
	}
	if !refThere {
		return nil, nil // no such ref yet — nothing published there
	}
	present, err := t.pathPresent(ref, path)
	if err != nil {
		return nil, err // could not tell — never fall through to "empty"
	}
	if !present {
		return nil, nil // absent reads as empty: a first write needs no special case
	}
	out, errOut, err := runGitIn(t.dir, nil, "cat-file", "blob", ref+":"+path)
	if err != nil {
		return nil, fmt.Errorf("read %s from %s: %v\n%s", path, ref, err, errOut)
	}
	return out, nil
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
	// Hoisted: the repo's signing config cannot change mid-loop, and re-shelling
	// `git config` per attempt is repeated expensive work (ARCH-CONSTRAINTS).
	sign, err := t.signs()
	if err != nil {
		return err
	}

	var lastRejection []byte
	for attempt := 1; attempt <= maxUpdateAttempts; attempt++ {
		// Attempt 1 fetches here; every later attempt reuses the fetch the
		// previous iteration already did to decide retryability, so a contended
		// update costs one fetch per attempt rather than two.
		if attempt == 1 {
			if out, err := t.fetch(); err != nil {
				return offlineError(t.remote, err, out)
			}
		}
		base, err := t.resolve(t.trackingRef())
		if err != nil {
			return err
		}
		old, err := t.readFrom(t.trackingRef(), path)
		if err != nil {
			return err
		}
		next, err := transform(old)
		if err != nil {
			return err // caller's error, surfaced unwrapped so errors.Is works
		}
		// A transform that changed nothing must not produce a commit. The tree
		// would be identical, so the push writes an EMPTY commit whose subject is
		// a permanent claim about an edit that never happened — and a commit
		// subject is the most durable message this system emits.
		if bytes.Equal(old, next) {
			return nil
		}
		out, err := t.commitAndPush(path, msg, next, base, sign)
		if err == nil {
			return nil
		}
		// Retry only when the TRUNK ACTUALLY MOVED — observed by re-resolving the
		// ref, not by matching git's rejection prose. A bare "rejected" match also
		// catches refusals retrying cannot fix (a declined pre-receive hook, no
		// push permission), which burned the whole budget and then reported a
		// contention story that never happened.
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
	sha, err := t.resolveOpt(ref)
	if err != nil {
		return "", err
	}
	if sha == "" {
		return "", fmt.Errorf("resolve %s: no such ref", ref)
	}
	return sha, nil
}

// commitAndPush builds the tree in a temp index and pushes the new commit at the
// branch. Returns git's STDERR so the caller can classify the failure.
func (t *TrunkFile) commitAndPush(path, msg string, content []byte, base string, sign bool) ([]byte, error) {
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
	//
	// BOUNDED CLAIM: git resolves those attributes from the WORKING TREE's
	// .gitattributes, not from the ref being written. That is correct whenever the
	// checkout and the trunk agree about the path — the overwhelmingly common
	// case, and always true when the branch descends from the trunk. It is wrong
	// only if the trunk's .gitattributes has diverged from this checkout's for
	// this path, and git offers no "use attributes from ref X" for hash-object, so
	// this is a stated limit rather than something the code enforces.
	out, errOut, err := runGitIn(t.dir, env, "hash-object", "-w", "--path", path, blobFile)
	if err != nil {
		return errOut, err
	}
	blob := strings.TrimSpace(string(out))

	// Preserve the path's existing mode. Rebuilding the entry as a hardcoded
	// 100644 silently drops the executable bit from any file that had it — which
	// a general primitive cannot do — ariadne#207 is specced to point this at
	// arbitrary repo paths.
	mode, err := t.modeOf(base, path)
	if err != nil {
		return nil, err
	}
	if _, errOut, err := runGitIn(t.dir, env, "update-index", "--add",
		"--cacheinfo", mode+","+blob+","+path); err != nil {
		return errOut, err
	}
	out, errOut, err = runGitIn(t.dir, env, "write-tree")
	if err != nil {
		return errOut, err
	}
	tree := strings.TrimSpace(string(out))

	args := []string{"commit-tree", tree, "-p", base, "-m", msg}
	if sign {
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

// modeOf returns the path's mode in the base tree, defaulting to a regular file
// when the path is new.
func (t *TrunkFile) modeOf(ref, path string) (string, error) {
	mode, present, err := t.entryOf(ref, path)
	if err != nil {
		return "", err
	}
	if !present || len(mode) != 6 {
		return "100644", nil // new path
	}
	return mode, nil
}

// signs reports whether this repo signs commits. A signing repo that silently
// produced unsigned commits through this path would be a regression, and
// commit-tree does not read commit.gpgsign on its own.
//
// --type=bool matters: git accepts 1/yes/on/true and stores the string VERBATIM,
// so `--get` on a repo configured with `yes` returns "yes" and a naive == "true"
// yields exactly the silent unsigned commit this function exists to prevent.
func (t *TrunkFile) signs() (bool, error) {
	out, errOut, err := runGitIn(t.dir, nil, "config", "--type=bool", "--get", "commit.gpgsign")
	if err == nil {
		return strings.TrimSpace(string(out)) == "true", nil
	}
	if gitExitCode(err) == gitAbsentExit {
		return false, nil // key simply not set — the common case
	}
	// Anything else means we could not determine the policy. Returning false here
	// would publish an UNSIGNED commit in a repo that requires signing, which is
	// the exact regression this function exists to prevent.
	return false, fmt.Errorf("read commit.gpgsign: %v\n%s", err, errOut)
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
