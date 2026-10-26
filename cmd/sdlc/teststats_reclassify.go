package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

const suiteReclassificationRevision = 1
const suiteReclassificationMaxJSON = 2 * suiteAttemptMaxJSON

var errReclassificationSubject = errors.New("corrected subject differs from original audited identity")

type suiteReclassificationAudit struct {
	Version        int                   `json:"version"`
	Original       []byte                `json:"original"`
	OriginalSHA256 string                `json:"original_sha256"`
	Actor          string                `json:"actor"`
	Reason         string                `json:"reason"`
	At             time.Time             `json:"at"`
	Diagnosis      suiteStartupDiagnosis `json:"diagnosis"`
	Classifier     int                   `json:"classifier"`
}

func suiteBytesSHA256(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func encodeReclassifiedAttempt(original []byte, auditSHA string) ([]byte, error) {
	a, err := decodeSuiteAttempt(original)
	if err != nil || a.Version == 3 || a.Version == suiteAttemptIncompleteVersion {
		return nil, errors.New("requires original v1/v2 attempt bytes")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(original, &fields); err != nil {
		return nil, err
	}
	fields["version"] = json.RawMessage("3")
	fields["outcome"] = json.RawMessage(`"could-not-run"`)
	fields["reclassification_sha256"], err = json.Marshal(auditSHA)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	if _, err := decodeSuiteAttempt(encoded); err != nil {
		return nil, err
	}
	return encoded, nil
}

// validateReclassificationProof binds all current facts to the original bytes,
// audit and complete immutable process output. No filesystem path is trusted.
func validateReclassificationProof(current, audit, log []byte) error {
	a, err := decodeSuiteAttempt(current)
	if err != nil || a.Version != 3 {
		return errors.New("invalid corrected attempt")
	}
	if len(audit) > suiteReclassificationMaxJSON || suiteBytesSHA256(audit) != a.ReclassificationSHA256 {
		return errors.New("audit SHA256 mismatch or oversized audit")
	}
	if err := jsonstrict.RejectDuplicateKeys(audit); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(audit, &fields); err != nil {
		return err
	}
	known := map[string]bool{"version": true, "original": true, "original_sha256": true, "actor": true, "reason": true, "at": true, "diagnosis": true, "classifier": true}
	for k := range fields {
		if !canonicalSuiteKey(k, known) {
			return errors.New("noncanonical audit field")
		}
	}
	for k := range known {
		if v, ok := fields[k]; !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return fmt.Errorf("missing/null audit %s", k)
		}
	}
	var proof suiteReclassificationAudit
	if err := json.Unmarshal(audit, &proof); err != nil {
		return err
	}
	if proof.Version != 1 || proof.Classifier != suiteReclassificationRevision || proof.At.IsZero() || proof.At.Location() != time.UTC || suiteBytesSHA256(proof.Original) != proof.OriginalSHA256 {
		return errors.New("invalid audit identity, classifier or time")
	}
	original, err := decodeSuiteAttempt(proof.Original)
	if err != nil {
		return err
	}
	if original.ID != a.ID || original.Repo != a.Repo || original.SHA != a.SHA || original.Check != a.Check {
		return errReclassificationSubject
	}
	diagnosis, err := planSuiteReclassification(original, a.Repo, proof.Actor, proof.Reason, log)
	if err != nil {
		return err
	}
	if diagnosis != proof.Diagnosis {
		return errors.New("audit diagnosis differs from original startup evidence")
	}
	expected, err := encodeReclassifiedAttempt(proof.Original, a.ReclassificationSHA256)
	if err != nil {
		return err
	}
	// Canonicalize only the top-level map, preserving opaque numeric values.
	var actual map[string]json.RawMessage
	if err := json.Unmarshal(current, &actual); err != nil {
		return err
	}
	normalized, err := json.Marshal(actual)
	if err != nil {
		return err
	}
	if !bytes.Equal(normalized, expected) {
		return errors.New("corrected record changes original attempt facts")
	}
	return nil
}

// boundedSuiteProofFile counts proof IO against the snapshot's existing budget.
func boundedSuiteProofFile(path string, limit int64, remaining *int64, open func(string) (*os.File, error)) ([]byte, error) {
	if *remaining <= 0 {
		return nil, errors.New("evidence byte budget reached")
	}
	f, err := open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("proof is not a regular file")
	}
	b, err := io.ReadAll(io.LimitReader(f, min(limit, *remaining)+1))
	*remaining -= int64(len(b))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit || *remaining < 0 {
		return nil, errors.New("evidence byte budget reached")
	}
	return b, nil
}

func readSuiteReclassificationProof(root string, a SuiteAttempt, raw []byte, remaining *int64, open func(string) (*os.File, error)) (time.Time, error) {
	audit, err := boundedSuiteProofFile(filepath.Join(root, "reclassified", a.ID+".json"), suiteReclassificationMaxJSON, remaining, open)
	if err != nil {
		return time.Time{}, err
	}
	log, err := boundedSuiteProofFile(filepath.Join(root, "failed", a.ID+".log"), 64<<20, remaining, open)
	if err != nil {
		return time.Time{}, err
	}
	if err := validateReclassificationProof(raw, audit, log); err != nil {
		return time.Time{}, err
	}
	var proof suiteReclassificationAudit
	if err := json.Unmarshal(audit, &proof); err != nil {
		return time.Time{}, err
	}
	return proof.At, nil
}

// The caller owns the repository transaction lock. The injected publisher is
// the file-effect seam: audit first (exclusive), then the authority replacement.
func publishSuiteReclassification(root, id, repo, actor, reason string, at time.Time, publish func(string, []byte, bool) error) error {
	if len(id) < 2 || len(id) > 64 || !suiteHex.MatchString(id) {
		return errors.New("give an exact published --attempt ID")
	}
	// Share the existing publication/pruning boundary with resume evidence.
	// Otherwise pruning can unlink the original after the last exact reread,
	// and our replacement would resurrect authority without its proof.
	directory, err := tryLockResumeDirectory(root)
	if err != nil {
		return fmt.Errorf("lock attempt evidence for reclassification; retry the same attempt: %w", err)
	}
	defer directory.Close()
	remaining := int64(suiteAttemptReaderBudget)
	path := filepath.Join(root, "attempts", id+".json")
	raw, err := boundedSuiteProofFile(path, suiteAttemptMaxJSON, &remaining, os.Open)
	if err != nil {
		return err
	}
	a, err := decodeSuiteAttempt(raw)
	if err != nil || a.ID != id || a.Repo != repo {
		return errors.New("unsupported, invalid or foreign attempt")
	}
	if a.Version == 3 {
		if _, err := readSuiteReclassificationProof(root, a, raw, &remaining, os.Open); err != nil {
			return err
		}
		// Recheck caller authority/reason against the immutable original proof.
		proofRaw, err := boundedSuiteProofFile(filepath.Join(root, "reclassified", id+".json"), suiteReclassificationMaxJSON, &remaining, os.Open)
		if err != nil {
			return err
		}
		var proof suiteReclassificationAudit
		if err := json.Unmarshal(proofRaw, &proof); err != nil {
			return err
		}
		original, err := decodeSuiteAttempt(proof.Original)
		if err != nil {
			return err
		}
		log, err := boundedSuiteProofFile(filepath.Join(root, "failed", id+".log"), 64<<20, &remaining, os.Open)
		if err != nil {
			return err
		}
		_, err = planSuiteReclassification(original, repo, actor, reason, log)
		return err
	}
	log, err := boundedSuiteProofFile(filepath.Join(root, "failed", id+".log"), 64<<20, &remaining, os.Open)
	if err != nil {
		return err
	}
	diagnosis, err := planSuiteReclassification(a, repo, actor, reason, log)
	if err != nil {
		return err
	}
	if at.IsZero() {
		return errors.New("reclassification time unavailable")
	}
	auditPath := filepath.Join(root, "reclassified", id+".json")
	audit, readErr := boundedSuiteProofFile(auditPath, suiteReclassificationMaxJSON, &remaining, os.Open)
	if readErr != nil && !os.IsNotExist(readErr) {
		return readErr
	}
	if os.IsNotExist(readErr) {
		audit, err = json.Marshal(suiteReclassificationAudit{Version: 1, Original: raw, OriginalSHA256: suiteBytesSHA256(raw), Actor: actor, Reason: reason, At: at.UTC(), Diagnosis: diagnosis, Classifier: suiteReclassificationRevision})
		if err != nil {
			return err
		}
	} else {
		var previous suiteReclassificationAudit
		if err := json.Unmarshal(audit, &previous); err != nil {
			return err
		}
		if previous.OriginalSHA256 != suiteBytesSHA256(raw) || !bytes.Equal(previous.Original, raw) {
			return errors.New("original raw digest differs from retained audit; inspect before retrying")
		}
	}
	encoded, err := encodeReclassifiedAttempt(raw, suiteBytesSHA256(audit))
	if err != nil {
		return err
	}
	if err := validateReclassificationProof(encoded, audit, log); err != nil {
		return err
	}
	w, err := beginSuiteAttempt(root)
	if err != nil {
		return err
	}
	defer os.RemoveAll(filepath.Dir(w.LogPath))
	if publish == nil {
		publish = func(path string, b []byte, exclusive bool) error {
			return writeSuiteReclassificationFile(filepath.Dir(w.LogPath), path, b, exclusive)
		}
	}
	if readErr != nil {
		if err := publish(auditPath, audit, true); err != nil {
			return fmt.Errorf("retain recovery audit: %w", err)
		}
	}
	// A changed original cannot be silently replaced by a stale recovery plan.
	if err := directory.verify(); err != nil {
		return err
	}
	latest, err := boundedSuiteProofFile(path, suiteAttemptMaxJSON, &remaining, os.Open)
	if err != nil || !bytes.Equal(latest, raw) {
		return errors.New("attempt moved during recovery; inspect before retrying")
	}
	if err := publish(path, encoded, false); err != nil {
		return fmt.Errorf("audit retained; repeat the same attempt to finish: %w", err)
	}
	return nil
}

func writeSuiteReclassificationFile(private, path string, b []byte, exclusive bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	// A new audit directory's own entry must be durable too. Syncing only
	// the child after linking its file would not persist that parent entry.
	parent, err := os.Open(filepath.Dir(filepath.Dir(path)))
	if err != nil {
		return err
	}
	if err := parent.Sync(); err != nil {
		parent.Close()
		return err
	}
	if err := parent.Close(); err != nil {
		return err
	}
	f, err := os.CreateTemp(private, ".reclass-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if exclusive {
		err = os.Link(f.Name(), path)
	} else {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

// A crash after the first prune unlink, or an audit linked after its expired
// original was pruned, must not create a permanent side ledger. Completed audit
// files get the existing 30-day grace; private partial writes retain the existing
// 24-hour dead-owner lifecycle. Unknown presence never proves an orphan.
func pruneSuiteReclassificationAudits(root string, now time.Time, remove func(string) error) error {
	dir, err := os.Open(filepath.Join(root, "reclassified"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer dir.Close()
	for seen := 0; seen < testStatsRecordLimit; {
		entries, readErr := dir.ReadDir(min(128, testStatsRecordLimit-seen))
		if readErr != nil && readErr != io.EOF {
			return readErr
		}
		seen += len(entries)
		for _, entry := range entries {
			id := strings.TrimSuffix(entry.Name(), ".json")
			if entry.Name() != id+".json" || len(id) < 2 || len(id) > 64 || !suiteHex.MatchString(id) {
				continue
			}
			info, err := entry.Info()
			if err != nil || !info.Mode().IsRegular() || now.Sub(info.ModTime()) <= suiteAttemptKeep {
				continue
			}
			if _, err := os.Lstat(filepath.Join(root, "attempts", id+".json")); !os.IsNotExist(err) {
				continue
			}
			for _, path := range []string{filepath.Join(root, "failed", id+".log"), filepath.Join(root, "reclassified", entry.Name())} {
				if err := remove(path); err != nil && !os.IsNotExist(err) {
					return err
				}
			}
		}
		if readErr == io.EOF {
			break
		}
	}
	return nil
}
