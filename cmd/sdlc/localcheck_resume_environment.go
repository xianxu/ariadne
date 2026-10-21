package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/scratchdisk"
)

func resumeEnvironmentReason(childID string) string { return "sdlc resume child " + childID }

// The existing Git worktree lock is the lifetime authority for both retained
// checkout inputs and scratch. No additional lock or ownership format is used.
func protectResumeEnvironment(d localRunDeps, wt, childID, inflight string) error {
	if !resumeValidID(childID) {
		return errors.New("invalid resume environment child")
	}
	if _, err := d.Git("worktree", "lock", "--reason", resumeEnvironmentReason(childID), wt); err != nil {
		return err
	}
	// On any uncertain marker write, retain the worktree lock. An unlocked
	// checkout could otherwise be swept while its child is still using it.
	return attachRunnerArtifactOwner(inflight, wt)
}

func retireResumeEnvironment(ctx context.Context, d localRunDeps, store string, l ResumeLink, r ResumeReconciliation) error {
	_, roots, _ := scratchdisk.Roots("tmp", os.TempDir())
	return retireResumeEnvironmentWithRoots(ctx, d, store, l, r, roots)
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

type resumeEnvironmentBinding struct{ gitdir, wt, common string }

// Missing registration after an earlier confirmed removal is a convergent
// no-op. Missing or inconsistent metadata inside an extant registration refuses.
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
	if os.IsNotExist(err) && filepath.Clean(filepath.Dir(dir)) == registered {
		return b, nil
	}
	if err != nil {
		return b, err
	}
	if canonicalOwner != registered || filepath.Base(dir) == "." || filepath.Base(dir) == ".." {
		return b, errors.New("resume environment belongs to another Git common directory")
	}
	gitdir, err := resumeEnvironmentCanonicalDir(dir)
	if os.IsNotExist(err) {
		return b, nil
	}
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
	if strings.TrimSuffix(string(raw), "\n") != resumeEnvironmentReason(l.ChildID) {
		return b, errors.New("resume environment worktree lock is for another child")
	}
	raw, err = resumeEnvironmentRead(filepath.Join(gitdir, "gitdir"), 4096)
	if err != nil {
		return b, err
	}
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
	return resumeEnvironmentBinding{gitdir: gitdir, wt: wt, common: common}, nil
}

func resumeEnvironmentReceipt(store string, l ResumeLink, r ResumeReconciliation) error {
	realStore, err := resumeEnvironmentCanonicalDir(store)
	if err != nil {
		return err
	}
	attempts, err := resumeEnvironmentCanonicalDir(filepath.Join(realStore, "attempts"))
	if err != nil {
		return err
	}
	if attempts != filepath.Join(realStore, "attempts") {
		return errors.New("resume environment receipt directory is redirected")
	}
	if validateResumeLink(l) != nil || validateResumeReconciliation(r) != nil || !retentionReceiptBound(l, r, r.LinkSHA256) {
		return errors.New("resume environment requires a bound terminal interrupted receipt")
	}
	raw, err := resumeEnvironmentRead(filepath.Join(store, "attempts", l.ID+".json"), suiteAttemptMaxJSON)
	if err != nil {
		return err
	}
	link, _, err := decodeResumeEvidence(raw, l.ID+".json")
	if err != nil || link == nil || *link != l || resumeRecordDigest(raw) != r.LinkSHA256 {
		return errors.New("resume environment link changed")
	}
	raw, err = resumeEnvironmentRead(filepath.Join(store, "attempts", r.ID+".json"), suiteAttemptMaxJSON)
	if err != nil {
		return err
	}
	_, receipt, err := decodeResumeEvidence(raw, r.ID+".json")
	if err != nil || receipt == nil || *receipt != r {
		return errors.New("resume environment terminal receipt is not published")
	}
	return nil
}

func retireResumeEnvironmentWithRoots(ctx context.Context, d localRunDeps, store string, l ResumeLink, r ResumeReconciliation, roots []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := resumeEnvironmentReceipt(store, l, r); err != nil {
		return err
	}
	realStore, err := resumeEnvironmentCanonicalDir(store)
	if err != nil {
		return err
	}
	inflight := filepath.Join(realStore, "inflight", l.ChildID)
	realInflight, err := resumeEnvironmentCanonicalDir(inflight)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if realInflight != inflight {
		return errors.New("resume environment inflight directory is redirected")
	}
	raw, err := resumeEnvironmentRead(filepath.Join(inflight, runnerArtifactOwnerFile), 4096)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	owner := string(raw)
	binding, err := inspectResumeEnvironment(d, owner, l)
	if err != nil {
		return err
	}
	if binding.wt == "" {
		return nil
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
	// Collect all candidates before mutation. Revalidate reverse pointers, exact
	// child reason and raw receipt before each irreversible removal.
	for _, path := range scratch {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := resumeEnvironmentReceipt(store, l, r); err != nil {
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
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := resumeEnvironmentReceipt(store, l, r); err != nil {
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
	if out, err := resumeEnvironmentGit(d, "worktree", "remove", "--force", "--force", binding.wt); err != nil {
		return fmt.Errorf("resume environment worktree removal unconfirmed: %w: %s", err, strings.TrimSpace(out))
	}
	return nil
}
