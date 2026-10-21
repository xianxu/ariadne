package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
	"github.com/xianxu/ariadne/pkg/workspace"
)

const (
	suiteAttemptKeep         = 30 * 24 * time.Hour
	suiteInflightStale       = 24 * time.Hour
	suiteAttemptMaxJSON      = 1 << 20
	suiteAttemptReaderBudget = 64 << 20
)

var suiteHex = regexp.MustCompile(`^[0-9a-fA-F]+$`)
var issueIDRE = regexp.MustCompile(`^[0-9]{6}$`)
var suiteRepo = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// encoding/json matches struct fields case-insensitively even with
// DisallowUnknownFields. Persisted evidence needs exact schema names so an
// alias cannot override a required outcome field after duplicate detection.
var suiteAttemptJSONFields = map[string]bool{
	"version": true, "id": true, "repo": true, "sha": true, "check": true,
	"slot": true, "agent": true, "model": true, "effort": true,
	"started_at": true, "finished_at": true, "queued_at": true, "acquired_at": true, "main_baseline": true,
	"queue_seconds": true, "wall_seconds": true, "exit": true, "timed_out": true,
	"log_sha256": true, "log_path": true, "failures": true, "units": true,
	"outcome": true, "known_flakes": true, "rerun_of": true, // version 2 only (#394)
	"reclassification_sha256": true,
}

// suiteAttemptVersion is the version the writer records. Version 2 (#394)
// adds the gate's outcome, its known flakes and a rerun's link; a v1 record
// cannot carry them, so a v1 history stays readable and unambiguous.
const suiteAttemptVersion = 2

// Ordinary runs remain v2 for older fleet readers. Only an audited nonpassing
// correction requires v3 semantics and therefore the newer reader.
const suiteAttemptReadVersion = 3

var suiteAttemptV2Fields = []string{"outcome", "known_flakes", "rerun_of"}

// suiteOutcomes are the outcome values a v2 attempt may record.
var suiteOutcomes = map[flakes.Kind]bool{flakes.Passed: true, flakes.KnownFlakePassed: true, flakes.KnownFlakeAccept: true,
	flakes.KnownFlakeTwice: true, flakes.Unknown: true, flakes.Unattributable: true, flakes.PersistentMain: true,
	flakes.RegistryUnusable: true, flakes.RuleOff: true, flakes.TooManyKnownFlaky: true, flakes.RerunInconclusive: true}

type suiteAttemptWriter struct{ ID, LogPath string }

// beginSuiteAttempt reserves a private log path. Only its owner writes there;
// published JSON files, not private directories, are reader commit markers.
func beginSuiteAttempt(root string) (*suiteAttemptWriter, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return nil, err
	}
	id := hex.EncodeToString(token[:])
	dir := filepath.Join(root, "inflight", id)
	if err := os.MkdirAll(filepath.Dir(dir), 0o700); err != nil {
		return nil, err
	}
	if err := os.Mkdir(dir, 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "owner"), []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
		return nil, err
	}
	return &suiteAttemptWriter{ID: id, LogPath: filepath.Join(dir, "run.log")}, nil
}

// decodeSuiteAttempt is the untrusted-disk boundary. It rejects a whole bad
// record, so malformed data cannot become a passing run or flake evidence.
func decodeSuiteAttempt(data []byte) (SuiteAttempt, error) {
	var a SuiteAttempt
	if len(data) > suiteAttemptMaxJSON {
		return a, fmt.Errorf("record exceeds %d bytes", suiteAttemptMaxJSON)
	}
	// Decode into tokens first: encoding/json otherwise accepts omitted scalar
	// fields as zero values and silently keeps the last duplicate key. Either
	// behavior could fabricate a successful run from damaged evidence.
	keys := map[string]bool{}
	probe := json.NewDecoder(bytes.NewReader(data))
	start, err := probe.Token()
	if err != nil || start != json.Delim('{') {
		return a, errors.New("attempt must be one JSON object")
	}
	for probe.More() {
		keyToken, err := probe.Token()
		if err != nil {
			return a, err
		}
		key, ok := keyToken.(string)
		if !ok || !canonicalSuiteKey(key, suiteAttemptJSONFields) || keys[key] {
			return a, errors.New("duplicate or noncanonical attempt field")
		}
		keys[key] = true
		var value json.RawMessage
		if err := probe.Decode(&value); err != nil {
			return a, err
		}
		if err := jsonstrict.RejectDuplicateKeys(value); err != nil {
			return a, err
		}
		if key == "units" {
			if err := validateSuiteUnitsJSON(value); err != nil {
				return a, err
			}
		}
		if key == "known_flakes" {
			if err := validateSuiteKnownFlakesJSON(value); err != nil {
				return a, err
			}
		}
		if suiteAttemptJSONFields[key] && key != "queue_seconds" && key != "wall_seconds" && key != "agent" && key != "model" && key != "effort" && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return a, fmt.Errorf("null required attempt field %s", key)
		}
	}
	end, err := probe.Token()
	if err != nil || end != json.Delim('}') {
		return a, errors.New("invalid attempt object")
	}
	if _, err := probe.Token(); !errors.Is(err, io.EOF) {
		return a, errors.New("trailing JSON")
	}
	for _, key := range []string{"version", "id", "repo", "sha", "check", "started_at", "finished_at", "main_baseline", "queue_seconds", "wall_seconds", "exit", "timed_out", "log_sha256", "log_path"} {
		if !keys[key] {
			return a, fmt.Errorf("missing required attempt field %s", key)
		}
	}
	d := json.NewDecoder(bytes.NewReader(data))
	if err := d.Decode(&a); err != nil {
		return a, err
	}
	if _, err := d.Token(); !errors.Is(err, io.EOF) {
		return a, errors.New("trailing JSON")
	}
	if a.Version != 1 && a.Version != 2 && a.Version != 3 {
		return a, fmt.Errorf("unsupported version %d", a.Version)
	}
	if a.Version == 1 {
		for _, key := range suiteAttemptV2Fields {
			if keys[key] {
				return a, fmt.Errorf("version 1 attempt carries the version 2 field %s", key)
			}
		}
	}
	if a.Version == 3 {
		if len(a.ReclassificationSHA256) != 64 || !suiteHex.MatchString(a.ReclassificationSHA256) ||
			a.Outcome != "could-not-run" || !isTestCheck(a.Check) || a.Exit <= 0 || a.TimedOut || len(a.Failures) != 0 || len(a.Units) != 0 {
			return a, errors.New("invalid nonpassing reclassification envelope")
		}
	} else if keys["reclassification_sha256"] {
		return a, errors.New("reclassification requires version 3")
	}
	if a.Outcome != "" && !(a.Version == 3 && a.Outcome == "could-not-run") && !suiteOutcomes[flakes.Kind(a.Outcome)] {
		return a, fmt.Errorf("invalid outcome %q", a.Outcome)
	}
	if a.RerunOf != "" && (!suiteHex.MatchString(a.RerunOf) || len(a.RerunOf) < 2 || len(a.RerunOf) > 64 || a.RerunOf == a.ID) {
		return a, errors.New("invalid rerun_of")
	}
	if (a.RerunOf != "") != (a.Check == suiteRerunCheck) {
		return a, fmt.Errorf("a %s attempt names its rerun_of, and only it does", suiteRerunCheck)
	}
	if len(a.KnownFlakes) > 10000 {
		return a, errors.New("too many facts")
	}
	for _, e := range a.KnownFlakes {
		if flakes.ValidTest(e.Test) != nil || !issueIDRE.MatchString(e.Issue) {
			return a, errors.New("invalid known flake")
		}
	}
	if !suiteHex.MatchString(a.ID) || len(a.ID) < 2 || len(a.ID) > 64 {
		return a, errors.New("invalid attempt ID")
	}
	if !suiteRepo.MatchString(a.Repo) {
		return a, errors.New("invalid repository")
	}
	if !suiteHex.MatchString(a.SHA) || len(a.SHA) != 40 && len(a.SHA) != 64 {
		return a, errors.New("invalid head SHA")
	}
	if !localCheckName.MatchString(a.Check) {
		return a, errors.New("invalid check name")
	}
	if a.Slot != "" {
		address, err := workspace.ParseAddress(a.Slot)
		if err != nil || address.Repo == "" || address.String() != a.Slot {
			return a, errors.New("invalid attempt slot")
		}
	}
	if !suiteHex.MatchString(a.LogSHA256) || len(a.LogSHA256) != 64 {
		return a, errors.New("invalid log hash")
	}
	if a.StartedAt.IsZero() || a.FinishedAt.IsZero() || a.FinishedAt.Before(a.StartedAt) || a.StartedAt.Location() != time.UTC || a.FinishedAt.Location() != time.UTC {
		return a, errors.New("invalid UTC time window")
	}
	if (a.QueuedAt == nil) != (a.AcquiredAt == nil) {
		return a, errors.New("queue endpoints must be an all-or-none pair")
	}
	if a.QueuedAt != nil {
		if a.QueuedAt.IsZero() || a.AcquiredAt.IsZero() || a.QueuedAt.Location() != time.UTC || a.AcquiredAt.Location() != time.UTC || a.AcquiredAt.Before(*a.QueuedAt) || a.AcquiredAt.After(a.StartedAt) || a.QueueSeconds == nil || math.Abs(a.AcquiredAt.Sub(*a.QueuedAt).Seconds()-*a.QueueSeconds) > 1e-6 {
			return a, errors.New("invalid queue endpoints or duration")
		}
	}
	if a.Exit < -1 || a.TimedOut && a.Exit == 0 {
		return a, errors.New("invalid exit/timeout")
	}
	for _, p := range []*float64{a.QueueSeconds, a.WallSeconds} {
		if p != nil && (*p < 0 || math.IsNaN(*p) || math.IsInf(*p, 0)) {
			return a, errors.New("invalid duration")
		}
	}
	if len(a.Failures) > 10000 || len(a.Units) > 10000 {
		return a, errors.New("too many facts")
	}
	for _, f := range a.Failures {
		if f == "" || len(f) > 512 {
			return a, errors.New("invalid failure name")
		}
	}
	for _, u := range a.Units {
		if u.Kind != "package" && u.Kind != "package-group" && u.Kind != "shard" || u.Name == "" || len(u.Name) > 512 || u.Seconds < 0 || math.IsNaN(u.Seconds) || math.IsInf(u.Seconds, 0) {
			return a, errors.New("invalid duration unit")
		}
	}
	if a.LogPath != "" && a.LogPath != filepath.Join("failed", a.ID+".log") && a.LogPath != filepath.Join("local-checks", a.SHA, a.Check+".log") {
		return a, errors.New("log path escapes generated locations")
	}
	return a, nil
}

// suiteAttemptQueueWindow preserves the documented late-shifted convention for
// old records: StartedAt follows post-acquisition setup, not lock acquisition.
func suiteAttemptQueueWindow(a SuiteAttempt) (time.Time, time.Time, bool) {
	if a.QueuedAt != nil && a.AcquiredAt != nil {
		return *a.QueuedAt, *a.AcquiredAt, true
	}
	if a.QueueSeconds == nil {
		return time.Time{}, time.Time{}, false
	}
	return a.StartedAt.Add(-time.Duration(*a.QueueSeconds * float64(time.Second))), a.StartedAt, false
}

func ptrTime(t time.Time) *time.Time { return &t }

func validateSuiteUnitsJSON(data []byte) error {
	return validateSuiteFactObjects(data, "duration unit", map[string]bool{"kind": true, "name": true, "seconds": true}, []string{"kind", "name", "seconds"})
}

func validateSuiteKnownFlakesJSON(data []byte) error {
	return validateSuiteFactObjects(data, "known flake", map[string]bool{"test": true, "issue": true}, []string{"test", "issue"})
}

func canonicalSuiteKey(key string, known map[string]bool) bool {
	if known[key] {
		return true
	}
	for name := range known {
		if strings.EqualFold(key, name) {
			return false // encoding/json would otherwise bind this alias to a known field
		}
	}
	return true
}

func validateSuiteFactObjects(data []byte, kind string, known map[string]bool, required []string) error {
	var units []json.RawMessage
	if err := json.Unmarshal(data, &units); err != nil {
		return err
	}
	for _, raw := range units {
		dec := json.NewDecoder(bytes.NewReader(raw))
		start, err := dec.Token()
		if err != nil || start != json.Delim('{') {
			return fmt.Errorf("%s must be an object", kind)
		}
		keys := map[string]bool{}
		for dec.More() {
			keyToken, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok || !canonicalSuiteKey(key, known) || keys[key] {
				return fmt.Errorf("duplicate or noncanonical %s field", kind)
			}
			keys[key] = true
			var value json.RawMessage
			if err := dec.Decode(&value); err != nil {
				return err
			}
			if known[key] && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				return fmt.Errorf("null required %s field %s", kind, key)
			}
		}
		if _, err := dec.Token(); err != nil {
			return err
		}
		for _, key := range required {
			if !keys[key] {
				return fmt.Errorf("%s missing field %s", kind, key)
			}
		}
	}
	return nil
}

// suiteAttemptLogFile resolves a published attempt's relative log_path, the
// inverse of publishSuiteAttempt's layout: failed/<id>.log under root, and
// local-checks/<sha>/<check>.log beside it. Empty when the record names none.
func suiteAttemptLogFile(root string, a SuiteAttempt) string {
	switch {
	case a.LogPath == "":
		return ""
	case a.LogPath == filepath.Join("failed", a.ID+".log"):
		return filepath.Join(root, a.LogPath)
	default:
		return filepath.Join(filepath.Dir(root), a.LogPath)
	}
}

// publishSuiteAttempt closes the private lifecycle before posting status. It
// preserves a failed log under an immutable ID and atomically publishes JSON.
func publishSuiteAttempt(root string, w *suiteAttemptWriter, a SuiteAttempt) (string, error) {
	if w == nil || a.ID != w.ID {
		return "", errors.New("attempt writer mismatch")
	}
	if err := os.MkdirAll(filepath.Join(root, "attempts"), 0o700); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(root, "failed"), 0o700); err != nil {
		return "", err
	}
	final := filepath.Join(root, "attempts", a.ID+".json")
	if _, err := os.Stat(final); err == nil {
		return "", errors.New("attempt ID already published")
	} else if !os.IsNotExist(err) {
		return "", err
	}
	log, err := os.OpenFile(w.LogPath, os.O_RDWR, 0)
	if err != nil {
		return "", err
	}
	if err := log.Sync(); err != nil {
		_ = log.Close()
		return "", err
	}
	if err := log.Close(); err != nil {
		return "", err
	}
	// A failure's log, and a known-flake rerun's (#394) whatever its exit,
	// is kept immutably under its ID: the per-head legacy path below is
	// rewritten by every later attempt on the head, so a rerun's evidence
	// there would be overwritten by the suite attempt it reran.
	immutable := a.Exit != 0 || a.TimedOut || a.RerunOf != ""
	if immutable {
		a.LogPath = filepath.Join("failed", a.ID+".log")
		// Link is an exclusive, same-filesystem publication: even an ID
		// collision cannot replace a prior failure's evidence.
		failedPath := filepath.Join(root, a.LogPath)
		if err := os.Link(w.LogPath, failedPath); err != nil {
			if !os.IsExist(err) {
				return "", err
			}
			sourceInfo, sourceErr := os.Stat(w.LogPath)
			failedInfo, failedErr := os.Stat(failedPath)
			if sourceErr != nil || failedErr != nil || !os.SameFile(sourceInfo, failedInfo) {
				return "", err
			}
		}
	} else {
		a.LogPath = filepath.Join("local-checks", a.SHA, a.Check+".log")
	}
	// Compatibility copy happens before the JSON commit marker. A failed copy
	// leaves no published record that falsely claims a complete attempt.
	from := w.LogPath
	if immutable {
		from = filepath.Join(root, a.LogPath)
	}
	// The legacy per-head copy is the suite's log; a rerun of a few tests
	// never replaces it.
	if a.RerunOf == "" {
		legacy := filepath.Join(filepath.Dir(root), "local-checks", a.SHA, a.Check+".log")
		if err := copySuiteLogAtomically(from, legacy, filepath.Dir(w.LogPath)); err != nil {
			return "", fmt.Errorf("legacy log copy: %w", err)
		}
	}
	data, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	if _, err := decodeSuiteAttempt(data); err != nil {
		return "", fmt.Errorf("attempt validation: %w", err)
	}
	if _, err := os.Stat(final); err == nil {
		return "", errors.New("attempt ID already published")
	} else if !os.IsNotExist(err) {
		return "", err
	}
	// Name the temp for its private owner. A crash leaves its ID available to
	// the dead-owner pruner, while active writers are never swept.
	tmp, err := os.CreateTemp(filepath.Join(root, "attempts"), ".attempt-"+a.ID+"-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmp.Name(), final); err != nil {
		return "", err
	}
	_ = os.RemoveAll(filepath.Dir(w.LogPath))
	return final, nil
}

func copySuiteLogAtomically(src, dst, privateDir string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	// Keep this publication temp in the same private in-flight directory as
	// the writer log, so dead-owner cleanup also reclaims an interrupted copy.
	out, err := os.CreateTemp(privateDir, ".suite-*")
	if err != nil {
		return err
	}
	defer os.Remove(out.Name())
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Rename(out.Name(), dst)
}

// readSuiteEvidence never opens record-derived log paths. It reports malformed
// input as coverage loss and admits at most limit published files.
func readSuiteEvidence(root string, limit int) (SuiteEvidenceSnapshot, error) {
	return readSuiteEvidenceBudget(root, limit, suiteAttemptReaderBudget)
}

func readSuiteEvidenceBudget(root string, limit int, maxBytes int64) (SuiteEvidenceSnapshot, error) {
	return readSuiteEvidenceWithOpen(root, limit, maxBytes, os.Open)
}

func readSuiteEvidenceWithOpen(root string, limit int, maxBytes int64, open func(string) (*os.File, error)) (snapshot SuiteEvidenceSnapshot, err error) {
	defer snapshot.indexLinkedParents()
	dir, err := open(filepath.Join(root, "attempts"))
	if os.IsNotExist(err) {
		return snapshot, nil
	}
	if err != nil {
		return snapshot, err
	}
	defer dir.Close()
	seen := 0
	var bytesRead int64
	for {
		entries, dirErr := dir.ReadDir(128)
		if dirErr != nil && dirErr != io.EOF {
			return snapshot, dirErr
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			if seen >= limit {
				snapshot.Warnings = append(snapshot.Warnings, "record limit reached; later attempts omitted")
				return snapshot, nil
			}
			seen++
			p := filepath.Join(root, "attempts", e.Name())
			f, openErr := open(p)
			if _, lerr := os.Lstat(p); errors.Is(openErr, fs.ErrNotExist) && errors.Is(lerr, fs.ErrNotExist) {
				continue // the entry itself is gone: pruned since the listing (a dangling link is still unreadable)
			}
			if openErr != nil {
				snapshot.Warnings = append(snapshot.Warnings, p+": "+openErr.Error())
				continue
			}
			remaining := maxBytes - bytesRead
			if remaining <= 0 {
				_ = f.Close()
				snapshot.Warnings = append(snapshot.Warnings, "record byte budget reached; later attempts omitted")
				return snapshot, nil
			}
			b, readErr := io.ReadAll(io.LimitReader(f, min(int64(suiteAttemptMaxJSON+1), remaining+1)))
			_ = f.Close()
			if readErr != nil {
				snapshot.Warnings = append(snapshot.Warnings, p+": "+readErr.Error())
				continue
			}
			if int64(len(b)) > remaining {
				snapshot.Warnings = append(snapshot.Warnings, "record byte budget reached; later attempts omitted")
				return snapshot, nil
			}
			bytesRead += int64(len(b))
			a, decodeErr := decodeSuiteAttempt(b)
			if decodeErr != nil || e.Name() != a.ID+".json" {
				if unsupported, identityErr := decodeUnsupportedProof(b, e.Name()); identityErr == nil {
					unsupported.Path = p
					snapshot.addUnsupported(unsupported)
				} else {
					snapshot.Warnings = append(snapshot.Warnings, p+": invalid attempt record")
				}
				continue
			}
			if a.Version == 3 {
				remaining := maxBytes - bytesRead
				var proofErr error
				a.ReclassifiedAt, proofErr = readSuiteReclassificationProof(root, a, b, &remaining, open)
				bytesRead = maxBytes - remaining
				if proofErr != nil {
					snapshot.addUnsupported(UnsupportedProof{Version: a.Version, ID: a.ID, Repo: a.Repo, SHA: a.SHA, Check: a.Check, Path: p, Raw: b, Problem: proofErr.Error()})
					continue
				}
			}
			snapshot.Attempts = append(snapshot.Attempts, a)
		}
		if dirErr == io.EOF {
			break
		}
	}
	return snapshot, nil
}

func pruneSuiteAttempts(root string, now time.Time) error {
	entries, err := os.ReadDir(filepath.Join(root, "attempts"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if id, ok := attemptTempOwner(e.Name()); ok {
			info, err := e.Info()
			if err == nil && now.Sub(info.ModTime()) > suiteInflightStale && suiteOwnerDead(filepath.Join(root, "inflight", id)) && !runnerArtifactProtected(filepath.Join(root, "inflight", id)) {
				_ = os.Remove(filepath.Join(root, "attempts", e.Name()))
			}
			continue
		}
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) <= suiteAttemptKeep {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		if !suiteHex.MatchString(id) {
			continue
		}
		_ = os.Remove(filepath.Join(root, "attempts", e.Name()))
		_ = os.Remove(filepath.Join(root, "failed", id+".log"))
		_ = os.Remove(filepath.Join(root, "reclassified", id+".json"))
	}
	// An active writer owns its in-flight directory even if its clock or run
	// duration exceeds the stale threshold. Uncertain liveness retains it.
	inflight, err := os.ReadDir(filepath.Join(root, "inflight"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, e := range inflight {
		if !e.IsDir() {
			continue
		}
		if !suiteHex.MatchString(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil || now.Sub(info.ModTime()) <= suiteInflightStale {
			continue
		}
		if runnerArtifactProtected(filepath.Join(root, "inflight", e.Name())) || !suiteOwnerDead(filepath.Join(root, "inflight", e.Name())) {
			continue
		}
		log := filepath.Join(root, "failed", e.Name()+".log")
		if _, err := os.Stat(filepath.Join(root, "attempts", e.Name()+".json")); os.IsNotExist(err) {
			if info, err := os.Stat(log); err == nil && now.Sub(info.ModTime()) > suiteInflightStale {
				_ = os.Remove(log)
			}
		}
		_ = os.RemoveAll(filepath.Join(root, "inflight", e.Name()))
	}
	return nil
}

func attemptTempOwner(name string) (string, bool) {
	const prefix = ".attempt-"
	if !strings.HasPrefix(name, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(name, prefix)
	if len(rest) < 34 || rest[32] != '-' || !suiteHex.MatchString(rest[:32]) {
		return "", false
	}
	return rest[:32], true
}

func suiteOwnerDead(dir string) bool {
	b, err := os.ReadFile(filepath.Join(dir, "owner"))
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	return err == nil && pid > 0 && syscall.Kill(pid, 0) == syscall.ESRCH
}
