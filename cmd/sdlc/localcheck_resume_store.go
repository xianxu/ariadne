package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
)

// The immutable parent's inode is the execution resource. Its lock spans all
// resume effects, and pruning takes this same lock before retiring the parent.
type resumeParentLock struct {
	f       *os.File
	path    string
	Raw     []byte
	Attempt SuiteAttempt
}

func (l *resumeParentLock) Close() error { return l.f.Close() }

func regularResumeFile(path string) (*os.File, error) {
	i, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !i.Mode().IsRegular() {
		return nil, errors.New("resume evidence is not a regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	opened, err := f.Stat()
	if err != nil || !os.SameFile(i, opened) {
		f.Close()
		return nil, errors.New("resume evidence changed while opening")
	}
	return f, nil
}

func lockResumeParent(ctx context.Context, root, id string) (*resumeParentLock, error) {
	if len(id) < 2 || len(id) > 64 || !suiteHex.MatchString(id) {
		return nil, errors.New("resume requires a safe parent attempt ID")
	}
	f, err := regularResumeFile(filepath.Join(root, "attempts", id+".json"))
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			f.Close()
		}
	}()
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		err = flock.Exclusive(f, true)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			return nil, err
		}
		timer := time.NewTimer(20 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	raw, err := io.ReadAll(io.LimitReader(f, suiteAttemptMaxJSON+1))
	if err != nil {
		return nil, err
	}
	a, err := decodeSuiteAttempt(raw)
	if err != nil || a.ID != id {
		return nil, fmt.Errorf("invalid resume parent %s: %v", id, err)
	}
	l := &resumeParentLock{f: f, path: filepath.Join(root, "attempts", id+".json"), Raw: raw, Attempt: a}
	if err := l.verify(); err != nil {
		return nil, err
	}
	ok = true
	return l, nil
}

func (l *resumeParentLock) verify() error {
	i, err := os.Lstat(l.path)
	if err != nil {
		return err
	}
	opened, err := l.f.Stat()
	if err != nil || !i.Mode().IsRegular() || !os.SameFile(i, opened) {
		return errors.New("resume parent path/inode changed")
	}
	if _, err := l.f.Seek(0, 0); err != nil {
		return err
	}
	raw, err := io.ReadAll(io.LimitReader(l.f, suiteAttemptMaxJSON+1))
	if err != nil {
		return err
	}
	// Compare exact bytes, including unknown members; a digest is not needed
	// here because both bounded byte snapshots are already in memory.
	if string(raw) != string(l.Raw) {
		return errors.New("resume parent bytes changed")
	}
	return nil
}

// Publish only while holding the matching parent's live lock. Hard-linking the
// synced temporary file is an exclusive commit marker, unlike replacing rename.
func publishResumeEvidence(root string, lock *resumeParentLock, raw []byte) error {
	if err := lock.verify(); err != nil {
		return err
	}
	if len(raw) > suiteAttemptMaxJSON {
		return errors.New("resume evidence exceeds byte limit")
	}
	// Identity is used only for strict decoding, never directly as a path.
	var envelope struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return err
	}
	link, receipt, err := decodeResumeEvidence(raw, envelope.ID+".json")
	if err != nil {
		return err
	}
	id, parent, repo, sha := "", "", "", ""
	if link != nil {
		id, parent, repo, sha = link.ID, link.RerunOf, link.Repo, link.SHA
	}
	if receipt != nil {
		id, parent, repo, sha = receipt.ID, receipt.RerunOf, receipt.Repo, receipt.SHA
	}
	if parent != lock.Attempt.ID || repo != lock.Attempt.Repo || sha != lock.Attempt.SHA {
		return errors.New("resume publication does not match locked parent")
	}
	if filepath.Clean(lock.path) != filepath.Join(root, "attempts", parent+".json") {
		return errors.New("resume lock is for another store")
	}
	s, err := readSuiteEvidence(root, testStatsRecordLimit)
	if err != nil || len(s.Warnings) > 0 {
		return fmt.Errorf("resume publication needs a complete evidence snapshot: %v; %v", err, s.Warnings)
	}
	if err := s.proofError(proofSubject{Repo: repo, SHA: sha, Check: suiteCheckName, ID: parent}); err != nil {
		return err
	}
	if len(s.Attempts)+len(s.Resumes)+len(s.Reconciliations)+len(s.Unsupported) > testStatsRecordLimit-3 {
		return errors.New("resume evidence capacity exhausted; retain or reconcile existing bundles")
	}
	if s.recordBytes+int64(len(raw))+2*suiteAttemptMaxJSON > suiteAttemptReaderBudget {
		return errors.New("resume evidence byte capacity exhausted; retain or reconcile existing bundles")
	}
	var bound *ResumeLink
	for i := range s.Resumes {
		l := &s.Resumes[i]
		if l.RerunOf == parent {
			if bound != nil {
				return errors.New("conflicting resume links")
			}
			bound = l
		}
	}
	if link != nil {
		if bound != nil {
			return errors.New("parent already has an immutable resume link")
		}
		if link.ParentRecordSHA256 != resumeRecordDigest(lock.Raw) || link.ParentLogSHA256 != lock.Attempt.LogSHA256 {
			return errors.New("resume parent digest mismatch")
		}
	} else {
		if bound == nil || receipt.LinkID != bound.ID || receipt.ChildID != bound.ChildID || receipt.LinkSHA256 != s.recordDigests[bound.ID] {
			return errors.New("reconciliation does not bind exact resume link")
		}
		for _, a := range s.Attempts {
			if a.ID == bound.ChildID {
				return errors.New("completed child cannot be reconciled as interrupted")
			}
		}
		for _, r := range s.Reconciliations {
			if r.RerunOf == parent {
				return errors.New("parent already has immutable reconciliation")
			}
		}
	}
	dir := filepath.Join(root, "attempts")
	tmp, err := os.CreateTemp(dir, ".resume-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(raw); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = lock.verify(); err != nil {
		return err
	}
	if err = os.Link(tmp.Name(), filepath.Join(dir, id+".json")); err != nil {
		return err
	}
	return syncResumeDirectory(dir)
}

func syncResumeDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	return errors.Join(f.Sync(), f.Close())
}
