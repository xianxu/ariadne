package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/scratchdisk"
)

func resumeEnvironmentReason(childID string) string { return "sdlc resume child " + childID }

// The existing Git worktree lock is the lifetime authority for both retained
// checkout inputs and scratch. No additional lock or ownership format is used.
func protectResumeEnvironment(d localRunDeps, wt, childID, inflight string) error {
	return protectResumeEnvironmentWithSync(d, wt, childID, inflight, syncResumeDirectory)
}
func protectResumeEnvironmentWithSync(d localRunDeps, wt, childID, inflight string, sync func(string) error) error {
	if !resumeValidID(childID) {
		return errors.New("invalid resume environment child")
	}
	// The execution fence already exists. Make its existing owner reference
	// durable before a Git lock can retain any checkout or scratch indefinitely.
	if err := attachRunnerArtifactOwner(inflight, wt); err != nil {
		return err
	}
	if err := sync(inflight); err != nil {
		return err
	}
	if err := sync(filepath.Dir(inflight)); err != nil {
		return err
	}
	_, err := d.Git("worktree", "lock", "--reason", resumeEnvironmentReason(childID), wt)
	return err
}

func resumeEnvironmentRead(path string, max int64) ([]byte, error) {
	f, err := regularResumeFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, errors.New("resume environment metadata exceeds bound")
	}
	return b, nil
}
func resumeEnvironmentPath(raw []byte) (string, error) {
	s := strings.TrimSuffix(string(raw), "\n")
	if s == "" || strings.ContainsAny(s, "\x00\r\n") || !filepath.IsAbs(s) {
		return "", errors.New("invalid resume environment path")
	}
	return filepath.Clean(s), nil
}
func resumeEnvironmentCanonicalDir(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	path = absolute
	i, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !i.IsDir() || i.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("resume environment directory is not regular")
	}
	return filepath.EvalSymlinks(path)
}
func resumeEnvironmentGit(d localRunDeps, args ...string) (string, error) {
	if d.CleanupGit != nil {
		return d.CleanupGit(args...)
	}
	return d.Git(args...)
}

type resumeEnvironmentBinding struct{ gitdir, wt, common, locked, reverse, commonRaw, head string }

// Initial authority must be fully present; this is not a recovery reader.
func inspectResumeEnvironment(d localRunDeps, owner string, l ResumeLink) (resumeEnvironmentBinding, error) {
	var b resumeEnvironmentBinding
	out, err := resumeEnvironmentGit(d, "rev-parse", "--git-common-dir")
	if err != nil {
		return b, err
	}
	common := strings.TrimSpace(out)
	if !filepath.IsAbs(common) {
		common = filepath.Join(d.callerRoot(), common)
	}
	common, err = filepath.EvalSymlinks(common)
	if err != nil {
		return b, err
	}
	dir, err := parseRunnerOwner([]byte(owner), "", false)
	if err != nil {
		return b, err
	}
	// Check lexical containment before resolving: a foreign path is never an
	// invitation to inspect or collect artifacts in that repository.
	registered := filepath.Join(common, "worktrees")
	canonicalOwner, err := filepath.EvalSymlinks(filepath.Dir(dir))
	if err != nil {
		return b, err
	}
	if canonicalOwner != registered || filepath.Base(dir) == "." || filepath.Base(dir) == ".." {
		return b, errors.New("resume environment belongs to another Git common directory")
	}
	gitdir, err := resumeEnvironmentCanonicalDir(dir)
	if err != nil {
		return b, err
	}
	if filepath.Dir(gitdir) != registered {
		return b, errors.New("resume environment Git directory escapes registration")
	}
	// Resolving worktrees itself outside the common directory is also forbidden.
	realRegistered, err := filepath.EvalSymlinks(registered)
	if err != nil || realRegistered != registered {
		return b, errors.New("resume environment registration is redirected")
	}
	raw, err := resumeEnvironmentRead(filepath.Join(gitdir, "locked"), 4096)
	if err != nil {
		return b, err
	}
	locked := string(raw)
	if strings.TrimSuffix(string(raw), "\n") != resumeEnvironmentReason(l.ChildID) {
		return b, errors.New("resume environment worktree lock is for another child")
	}
	raw, err = resumeEnvironmentRead(filepath.Join(gitdir, "gitdir"), 4096)
	if err != nil {
		return b, err
	}
	reverse := string(raw)
	pointer, err := resumeEnvironmentPath(raw)
	if err != nil || filepath.Base(pointer) != ".git" {
		return b, errors.New("invalid resume environment worktree reverse pointer")
	}
	wt, err := resumeEnvironmentCanonicalDir(filepath.Dir(pointer))
	if err != nil {
		return b, err
	}
	if !strings.HasPrefix(filepath.Base(wt), localCheckTmpPrefix) || len(filepath.Base(wt)) == len(localCheckTmpPrefix) {
		return b, errors.New("resume environment is not a private local-check worktree")
	}
	raw, err = resumeEnvironmentRead(filepath.Join(wt, ".git"), 4096)
	if err != nil {
		return b, err
	}
	back, err := parseRunnerOwner(raw, wt, true)
	if err != nil {
		return b, err
	}
	back, err = filepath.EvalSymlinks(back)
	if err != nil || back != gitdir {
		return b, errors.New("resume environment worktree does not point back to its registration")
	}

	raw, err = resumeEnvironmentRead(filepath.Join(gitdir, "commondir"), 4096)
	if err != nil {
		return b, err
	}
	commonRaw := string(raw)
	commonPointer := strings.TrimSuffix(string(raw), "\n")
	if commonPointer == "" || strings.ContainsAny(commonPointer, "\x00\r\n") {
		return b, errors.New("invalid resume environment common pointer")
	}
	if !filepath.IsAbs(commonPointer) {
		commonPointer = filepath.Join(gitdir, commonPointer)
	}
	realCommon, err := filepath.EvalSymlinks(commonPointer)
	if err != nil || realCommon != common {
		return b, errors.New("resume environment common pointer changed")
	}
	raw, err = resumeEnvironmentRead(filepath.Join(gitdir, "HEAD"), 256)
	if err != nil {
		return b, err
	}
	if strings.TrimSuffix(string(raw), "\n") != l.SHA {
		return b, errors.New("resume environment detached HEAD changed")
	}
	return resumeEnvironmentBinding{gitdir: gitdir, wt: wt, common: common, locked: locked, reverse: reverse, commonRaw: commonRaw, head: string(raw)}, nil
}

func retireResumeEnvironmentInputs(ctx context.Context, d localRunDeps, store string, l ResumeLink, roots []string, ops resumeEnvironmentIO, authorize func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := authorize(); err != nil {
		return err
	}
	realStore, err := resumeEnvironmentCanonicalDir(store)
	if err != nil {
		return err
	}
	inflight := filepath.Join(realStore, "inflight", l.ChildID)
	realInflight, err := resumeEnvironmentCanonicalDir(inflight)
	if os.IsNotExist(err) {
		return errors.New("completed resume environment ownership marker is missing")
	}
	if err != nil {
		return err
	}
	if realInflight != inflight {
		return errors.New("resume environment inflight directory is redirected")
	}
	raw, err := resumeEnvironmentRead(filepath.Join(inflight, runnerArtifactOwnerFile), 4096)
	if os.IsNotExist(err) {
		return errors.New("completed resume environment ownership marker is missing")
	}
	if err != nil {
		return err
	}
	owner := string(raw)
	binding, err := inspectResumeEnvironment(d, owner, l)
	if err != nil {
		return err
	}
	var scratch []string
	for _, root := range roots {
		realRoot, err := filepath.EvalSymlinks(root)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		entries, err := os.ReadDir(realRoot)
		if err != nil {
			return err
		}
		for _, e := range entries {
			match := runDirRE.FindStringSubmatch(e.Name())
			if !e.IsDir() || match == nil {
				continue
			}
			path := filepath.Join(realRoot, e.Name())
			marker, err := resumeEnvironmentRead(filepath.Join(path, runnerArtifactOwnerFile), 4096)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return err
			}
			if string(marker) != owner {
				continue
			}
			dir, err := parseRunnerOwner(marker, path, false)
			if err != nil {
				return err
			}
			dir, err = filepath.EvalSymlinks(dir)
			if err != nil {
				return err
			}
			if dir != binding.gitdir {
				continue
			}
			pid, err := strconv.Atoi(match[1])
			if err != nil || pid != l.Owner.PID {
				return errors.New("resume environment scratch has another owner PID")
			}
			scratch = append(scratch, path)
		}
	}
	confirmOwner := func() error {
		current, err := resumeEnvironmentCanonicalDir(inflight)
		if err != nil {
			return err
		}
		if current != inflight {
			return errors.New("resume environment inflight directory changed")
		}
		marker, err := resumeEnvironmentRead(filepath.Join(inflight, runnerArtifactOwnerFile), 4096)
		if err != nil {
			return err
		}
		if string(marker) != owner {
			return errors.New("resume environment inflight owner changed")
		}
		return nil
	}
	// Cancellation revokes authority even when the preceding effect succeeded.
	// Reauthorize before every removal and directory sync; never retry this pass.
	// Collect all candidates before mutation. Revalidate reverse pointers, exact
	// child reason and original-owner authority before each irreversible removal.
	for _, path := range scratch {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := authorize(); err != nil {
			return err
		}
		current, err := inspectResumeEnvironment(d, owner, l)
		if err != nil {
			return err
		}
		if current != binding {
			return errors.New("resume environment binding changed")
		}
		if err := confirmOwner(); err != nil {
			return err
		}
		if _, err := resumeEnvironmentCanonicalDir(path); err != nil {
			return err
		}
		marker, err := resumeEnvironmentRead(filepath.Join(path, runnerArtifactOwnerFile), 4096)
		if err != nil {
			return err
		}
		dir, err := parseRunnerOwner(marker, path, false)
		if err != nil {
			return err
		}
		dir, err = filepath.EvalSymlinks(dir)
		if err != nil || dir != binding.gitdir {
			return errors.New("resume environment scratch binding changed")
		}
		if err := resumeCleanupEffect(ctx, authorize, func() error { return ops.removeAll(path) }); err != nil {
			return err
		}
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			return errors.New("resume scratch cleanup did not confirm absence")
		}
		if err := resumeCleanupEffect(ctx, authorize, func() error { return ops.sync(filepath.Dir(path)) }); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := authorize(); err != nil {
		return err
	}
	current, err := inspectResumeEnvironment(d, owner, l)
	if err != nil {
		return err
	}
	if current != binding {
		return errors.New("resume environment binding changed")
	}
	if err := confirmOwner(); err != nil {
		return err
	}
	// This invocation advances using its original validated binding. A partial
	// state from any previous invocation is never reconstructed or repaired.
	if err := resumeCleanupEffect(ctx, authorize, func() error { return ops.removeAll(binding.wt) }); err != nil {
		return err
	}
	if _, err := os.Lstat(binding.wt); !os.IsNotExist(err) {
		return errors.New("resume checkout removal unconfirmed")
	}
	if err := resumeCleanupEffect(ctx, authorize, func() error { return ops.sync(filepath.Dir(binding.wt)) }); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := authorize(); err != nil {
		return err
	}
	if err := confirmResumeRegistration(binding); err != nil {
		return err
	}
	if err := confirmOwner(); err != nil {
		return err
	}
	if err := resumeCleanupEffect(ctx, authorize, func() error {
		out, err := resumeEnvironmentGit(d, "worktree", "remove", "--force", "--force", binding.wt)
		if err != nil {
			return fmt.Errorf("resume environment worktree removal unconfirmed: %w: %s", err, strings.TrimSpace(out))
		}
		return nil
	}); err != nil {
		return err
	}
	for _, path := range []string{binding.wt, binding.gitdir} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			return errors.New("resume environment removal did not confirm both paths absent")
		}
	}
	parent := filepath.Dir(binding.gitdir)
	if _, err := os.Lstat(parent); os.IsNotExist(err) {
		parent = binding.common
	} else if err != nil {
		return err
	}
	if err := resumeCleanupEffect(ctx, authorize, func() error { return ops.sync(parent) }); err != nil {
		return err
	}
	return releaseResumeEnvironmentOwner(ctx, inflight, owner, ops, authorize)
}

// The checkout was removed earlier in this pass. Check the captured registration
// bytes without deriving or accepting a reconstructed checkout binding.
func confirmResumeRegistration(b resumeEnvironmentBinding) error {
	current, err := resumeEnvironmentCanonicalDir(b.gitdir)
	if err != nil {
		return err
	}
	if current != b.gitdir {
		return errors.New("resume registration directory changed")
	}
	for name, want := range map[string]string{"locked": b.locked, "gitdir": b.reverse, "commondir": b.commonRaw, "HEAD": b.head} {
		raw, err := resumeEnvironmentRead(filepath.Join(b.gitdir, name), 4096)
		if err != nil {
			return err
		}
		if string(raw) != want {
			return errors.New("resume registration bytes changed during cleanup")
		}
	}
	return nil
}

func releaseResumeEnvironmentOwner(ctx context.Context, inflight, owner string, ops resumeEnvironmentIO, authorize func() error) error {
	if err := authorize(); err != nil {
		return err
	}
	marker := filepath.Join(inflight, runnerArtifactOwnerFile)
	raw, err := resumeEnvironmentRead(marker, 4096)
	if err != nil {
		return err
	}
	if string(raw) != owner {
		return errors.New("resume environment owner changed before release")
	}
	if err := resumeCleanupEffect(ctx, authorize, func() error { return ops.remove(marker) }); err != nil {
		return err
	}
	if err := resumeCleanupEffect(ctx, authorize, func() error { return ops.sync(inflight) }); err != nil {
		return err
	}
	return ctx.Err()
}

func retireCompletedResumeEnvironment(ctx context.Context, d localRunDeps, store string, l ResumeLink, lock *resumeParentLock) error {
	_, roots, _ := scratchdisk.Roots("tmp", os.TempDir())
	return retireCompletedResumeEnvironmentWithRoots(ctx, d, store, l, lock, roots)
}
func retireCompletedResumeEnvironmentWithRoots(ctx context.Context, d localRunDeps, store string, l ResumeLink, lock *resumeParentLock, roots []string) error {
	return retireCompletedResumeEnvironmentWithIO(ctx, d, store, l, lock, roots, resumeEnvironmentIO{removeAll: os.RemoveAll, remove: os.Remove, sync: syncResumeDirectory})
}
func retireCompletedResumeEnvironmentWithIO(ctx context.Context, d localRunDeps, store string, l ResumeLink, lock *resumeParentLock, roots []string, ops resumeEnvironmentIO) error {
	var linkDigest string
	authorize := func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if lock == nil || d.runnerLifetime == nil || !d.runnerLifetime.mayRetire() {
			return errors.New("resume managed process cleanup is not confirmed")
		}
		if err := lock.verify(); err != nil {
			return err
		}
		if validateResumeLink(l) != nil || filepath.Clean(lock.path) != filepath.Join(store, "attempts", l.RerunOf+".json") || lock.Attempt.ID != l.RerunOf || lock.Attempt.Repo != l.Repo || lock.Attempt.SHA != l.SHA || lock.Attempt.Check != l.Check || resumeRecordDigest(lock.Raw) != l.ParentRecordSHA256 || lock.Attempt.LogSHA256 != l.ParentLogSHA256 {
			return errors.New("completed resume does not bind the locked parent")
		}
		started, err := processStartTime(os.Getpid())
		if err != nil {
			return err
		}
		boot, err := queueBootID()
		if err != nil {
			return err
		}
		if l.Owner.PID != os.Getpid() || l.Owner.PIDStarted != started || l.Owner.Host != hostName() || l.Owner.BootID != boot {
			return errors.New("completed resume is not owned by this process incarnation")
		}
		raw, err := resumeEnvironmentRead(filepath.Join(store, "attempts", l.ID+".json"), suiteAttemptMaxJSON)
		if err != nil {
			return err
		}
		decoded, _, err := decodeResumeEvidence(raw, l.ID+".json")
		if err != nil || decoded == nil || !reflect.DeepEqual(*decoded, l) {
			return errors.New("completed resume link changed")
		}
		digest := resumeRecordDigest(raw)
		if linkDigest != "" && linkDigest != digest {
			return errors.New("completed resume raw link changed")
		}
		linkDigest = digest
		return nil
	}
	return retireResumeEnvironmentInputs(ctx, d, store, l, roots, ops, authorize)
}

type resumeEnvironmentIO struct{ removeAll, remove, sync func(string) error }
