package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// The operation seam is per invocation: fault tests cannot perturb another
// pruner or a concurrently publishing runner.
type resumeRetentionIO struct {
	remove, removeAll, sync func(string) error
}

type resumeRetentionIndex struct {
	attempts  map[string]SuiteAttempt
	links     map[string][]ResumeLink
	receipts  map[string][]ResumeReconciliation
	children  map[string][]SuiteAttempt
	protected map[string]bool
	blocked   map[string]bool
	visits    int
}

func indexResumeRetention(s SuiteEvidenceSnapshot) resumeRetentionIndex {
	x := resumeRetentionIndex{attempts: make(map[string]SuiteAttempt), links: make(map[string][]ResumeLink), receipts: make(map[string][]ResumeReconciliation), children: make(map[string][]SuiteAttempt), protected: make(map[string]bool), blocked: make(map[string]bool)}
	owners := make(map[string]string)
	member := func(id, parent string) {
		if previous, ok := owners[id]; ok && previous != parent {
			x.blocked[previous] = true
			x.blocked[parent] = true
		}
		owners[id] = parent
	}
	for _, a := range s.Attempts {
		x.visits++
		x.attempts[a.ID] = a
		if a.RerunOf != "" {
			x.children[a.RerunOf] = append(x.children[a.RerunOf], a)
		}
	}
	for _, l := range s.Resumes {
		member(l.ID, l.RerunOf)
		member(l.RerunOf, l.RerunOf)
		member(l.ChildID, l.RerunOf)
		x.visits++
		x.links[l.RerunOf] = append(x.links[l.RerunOf], l)
		x.protected[l.ID] = true
		x.protected[l.RerunOf] = true
		x.protected[l.ChildID] = true
	}
	for _, r := range s.Reconciliations {
		member(r.ID, r.RerunOf)
		member(r.RerunOf, r.RerunOf)
		member(r.LinkID, r.RerunOf)
		member(r.ChildID, r.RerunOf)
		x.visits++
		x.receipts[r.RerunOf] = append(x.receipts[r.RerunOf], r)
		x.protected[r.ID] = true
		x.protected[r.RerunOf] = true
		x.protected[r.LinkID] = true
		x.protected[r.ChildID] = true
	}
	for _, u := range s.Unsupported {
		x.visits++
		x.protected[u.ID] = true
		if u.RerunOf != "" {
			x.protected[u.RerunOf] = true
		}
	}
	// Protect conflicting children as well, even if a link reserved another ID.
	for parent, children := range x.children {
		if x.protected[parent] {
			for _, c := range children {
				x.visits++
				x.protected[c.ID] = true
			}
		}
	}
	return x
}

func pruneResumeBundles(root string, now time.Time) (map[string]bool, error) {
	return pruneResumeBundlesWithIO(root, now, resumeRetentionIO{remove: os.Remove, removeAll: os.RemoveAll, sync: syncResumeDirectory})
}

func pruneResumeBundlesWithIO(root string, now time.Time, ops resumeRetentionIO) (map[string]bool, error) {
	s, err := readSuiteEvidence(root, testStatsRecordLimit)
	if err != nil {
		return nil, err
	}
	x := indexResumeRetention(s)
	// Interrupted graph retention follows the operator-selected policy.
	// With unavailable authority, even an apparently unrelated legacy row may be
	// a bundle member. Prevent the caller's individual-file pruning altogether.
	if len(s.Warnings) > 0 || len(s.Unsupported) > 0 {
		return x.protected, errors.New("resume retention requires a complete supported evidence snapshot")
	}
	logRefs := make(map[string]int)
	for _, a := range s.Attempts {
		if path := suiteAttemptLogFile(root, a); path != "" {
			logRefs[path]++
		}
	}
	for parent, links := range x.links {
		if x.blocked[parent] {
			continue
		}
		if len(links) != 1 || len(x.receipts[parent]) != 0 || len(x.children[parent]) > 1 {
			continue
		}
		l := links[0]
		memberIDs := []string{l.RerunOf, l.ID, l.ChildID}
		if resumeRetentionEnvironmentProtected(root, memberIDs) {
			continue
		}
		child, hasChild := x.attempts[l.ChildID]
		// Pending, interrupted and receipt-bearing graphs are permanently pinned.
		if !hasChild {
			continue
		}
		if len(x.children[parent]) > 0 && x.children[parent][0].ID != l.ChildID {
			continue
		}
		if !retentionChildBound(l, child) {
			continue
		}
		raw, e := readRetentionLog(suiteAttemptLogFile(root, child))
		if e != nil || resumeRecordDigest(raw) != child.LogSHA256 {
			continue
		}
		latest := retentionLatest(l.CreatedAt, child.FinishedAt)
		p, hasParent := x.attempts[parent]
		if hasParent {
			if !resumeMeasuredDuration(p.WallSeconds) || p.RerunOf != "" || p.Repo != l.Repo || p.SHA != l.SHA || p.Check != l.Check || p.LogSHA256 != l.ParentLogSHA256 || s.recordDigests[p.ID] != l.ParentRecordSHA256 || l.CreatedAt.Before(p.FinishedAt) {
				continue
			}
			raw, e := readRetentionLog(suiteAttemptLogFile(root, p))
			if e != nil || resumeRecordDigest(raw) != l.ParentLogSHA256 {
				continue
			}
			latest = retentionLatest(latest, p.FinishedAt)
		}
		if now.Sub(latest) <= suiteAttemptKeep {
			continue
		}
		// Directory → parent is safe ONLY as a strict try. A publisher may
		// already own parent while waiting for our directory mutation mutex.
		var lock *resumeParentLock
		if hasParent {
			lock, err = tryLockResumeParent(root, parent)
			if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
				continue
			}
			if err != nil {
				return x.protected, err
			}
			if resumeRecordDigest(lock.Raw) != l.ParentRecordSHA256 {
				lock.Close()
				continue
			}
		}
		err = retireResumeBundle(root, l, p, hasParent, child, s.recordDigests, lock, logRefs, ops)
		if lock != nil {
			err = errors.Join(err, lock.Close())
		}
		if err != nil {
			return x.protected, err
		}
	}
	return x.protected, nil
}

func retentionLatest(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}
func retentionChildBound(l ResumeLink, c SuiteAttempt) bool {
	return (c.Version == 1 || c.Version == 2) && c.ID == l.ChildID && c.Repo == l.Repo && c.SHA == l.SHA && c.Check == suiteRerunCheck && c.RerunOf == l.RerunOf && !c.Legacy && !c.MainBaseline && resumeMeasuredDuration(c.WallSeconds) && resumeUTC(c.StartedAt) && resumeUTC(c.FinishedAt) && !c.StartedAt.Before(l.CreatedAt) && !c.FinishedAt.Before(c.StartedAt) && resumeValidDigest(c.LogSHA256)
}
func readRetentionLog(path string) ([]byte, error) {
	if path == "" {
		return nil, errors.New("missing retained log")
	}
	f, err := regularResumeFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, (64<<20)+1))
	if len(raw) > 64<<20 {
		return nil, errors.New("retained log exceeds budget")
	}
	return raw, err
}

// Retirement revalidates each immutable record under the parent's execution
// lock. A missing parent is allowed only after terminal binding has been proved.
func retireResumeBundle(root string, l ResumeLink, p SuiteAttempt, hasParent bool, c SuiteAttempt, digests map[string]string, lock *resumeParentLock, logRefs map[string]int, ops resumeRetentionIO) error {
	dir := filepath.Join(root, "attempts")
	record := func(id string) string { return filepath.Join(dir, id+".json") }
	ids := []string{l.ID, c.ID}
	for _, id := range ids {
		raw, err := readRetentionLog(record(id))
		if err != nil {
			return err
		}
		if resumeRecordDigest(raw) != digests[id] {
			return fmt.Errorf("resume retention record %s changed", id)
		}
	}
	if hasParent {
		if err := lock.verify(); err != nil {
			return err
		}
	}
	// Always resync even on a parent-missing retry: an earlier unlink may not
	// have become durable before the interrupted process exited.
	if err := retentionRemove(ops.remove, record(l.RerunOf)); err != nil {
		return err
	}
	if err := ops.sync(dir); err != nil {
		return err
	}
	// The rest of this sequence is finalized with the durable-retirement order.
	return retireResumeMembers(root, l, p, hasParent, c, logRefs, ops)
}
func retentionRemove(remove func(string) error, path string) error {
	err := remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// The terminal witness survives link removal until that removal is durable.
// An interrupted parent-missing retry can therefore prove terminality again.
func retireResumeMembers(root string, l ResumeLink, p SuiteAttempt, hasParent bool, c SuiteAttempt, logRefs map[string]int, ops resumeRetentionIO) error {
	dir := filepath.Join(root, "attempts")
	removeLog := func(a SuiteAttempt) error {
		path := suiteAttemptLogFile(root, a)
		if path == "" || logRefs[path] > 1 {
			return nil
		}
		return retentionRemove(ops.remove, path)
	}
	if hasParent {
		if err := removeLog(p); err != nil {
			return err
		}
	} else {
		// The failed raw log remains addressable after a crash lost the parent row.
		path := filepath.Join(root, "failed", l.RerunOf+".log")
		if logRefs[path] == 0 {
			if err := retentionRemove(ops.remove, path); err != nil {
				return err
			}
		}
	}
	for _, id := range []string{l.RerunOf, l.ID, l.ChildID} {
		if err := retentionRemove(ops.removeAll, filepath.Join(root, "inflight", id)); err != nil {
			return err
		}
	}
	if err := retentionRemove(ops.remove, filepath.Join(dir, l.ID+".json")); err != nil {
		return err
	}
	if err := ops.sync(dir); err != nil {
		return err
	}
	if err := removeLog(c); err != nil {
		return err
	}
	if err := retentionRemove(ops.remove, filepath.Join(dir, c.ID+".json")); err != nil {
		return err
	}
	return ops.sync(dir)
}

func resumeRetentionEnvironmentProtected(root string, ids []string) bool {
	for _, id := range ids {
		inflight := filepath.Join(root, "inflight", id)
		info, err := os.Lstat(filepath.Join(inflight, runnerArtifactOwnerFile))
		if err != nil && !os.IsNotExist(err) {
			return true
		}
		if err == nil && !info.Mode().IsRegular() {
			return true
		}
		if runnerArtifactProtected(inflight) {
			return true
		}
	}
	return false
}
