package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	suiteAttemptKeep         = 30 * 24 * time.Hour
	suiteInflightStale       = 24 * time.Hour
	suiteAttemptMaxJSON      = 1 << 20
	suiteAttemptReaderBudget = 64 << 20
)

var suiteHex = regexp.MustCompile(`^[0-9a-fA-F]+$`)
var suiteRepo = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

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
		if !ok || keys[key] {
			return a, errors.New("duplicate or invalid attempt field")
		}
		keys[key] = true
		var value json.RawMessage
		if err := probe.Decode(&value); err != nil {
			return a, err
		}
		if key != "queue_seconds" && key != "wall_seconds" && key != "agent" && key != "model" && key != "effort" && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
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
	d.DisallowUnknownFields()
	if err := d.Decode(&a); err != nil {
		return a, err
	}
	if _, err := d.Token(); !errors.Is(err, io.EOF) {
		return a, errors.New("trailing JSON")
	}
	if a.Version != 1 {
		return a, fmt.Errorf("unsupported version %d", a.Version)
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
	if !suiteHex.MatchString(a.LogSHA256) || len(a.LogSHA256) != 64 {
		return a, errors.New("invalid log hash")
	}
	if a.StartedAt.IsZero() || a.FinishedAt.IsZero() || a.FinishedAt.Before(a.StartedAt) || a.StartedAt.Location() != time.UTC || a.FinishedAt.Location() != time.UTC {
		return a, errors.New("invalid UTC time window")
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
		if u.Name == "" || len(u.Name) > 512 || u.Seconds < 0 || math.IsNaN(u.Seconds) || math.IsInf(u.Seconds, 0) {
			return a, errors.New("invalid duration unit")
		}
	}
	if a.LogPath != "" && a.LogPath != filepath.Join("failed", a.ID+".log") && a.LogPath != filepath.Join("local-checks", a.SHA, a.Check+".log") {
		return a, errors.New("log path escapes generated locations")
	}
	return a, nil
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
	if a.Exit != 0 || a.TimedOut {
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
	if a.Exit != 0 || a.TimedOut {
		from = filepath.Join(root, a.LogPath)
	}
	legacy := filepath.Join(filepath.Dir(root), "local-checks", a.SHA, a.Check+".log")
	if err := copySuiteLogAtomically(from, legacy, filepath.Dir(w.LogPath)); err != nil {
		return "", fmt.Errorf("legacy log copy: %w", err)
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

// readSuiteAttempts never opens record-derived log paths. It reports malformed
// input as coverage loss and admits at most limit published files.
func readSuiteAttempts(root string, limit int) ([]SuiteAttempt, []string, error) {
	return readSuiteAttemptsBudget(root, limit, suiteAttemptReaderBudget)
}

func readSuiteAttemptsBudget(root string, limit int, maxBytes int64) ([]SuiteAttempt, []string, error) {
	dir, err := os.Open(filepath.Join(root, "attempts"))
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	defer dir.Close()
	var out []SuiteAttempt
	var warnings []string
	seen := 0
	var bytesRead int64
	for {
		entries, dirErr := dir.ReadDir(128)
		if dirErr != nil && dirErr != io.EOF {
			return out, warnings, dirErr
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			if seen >= limit {
				warnings = append(warnings, "record limit reached; later attempts omitted")
				return out, warnings, nil
			}
			seen++
			p := filepath.Join(root, "attempts", e.Name())
			f, openErr := os.Open(p)
			if openErr != nil {
				warnings = append(warnings, p+": "+openErr.Error())
				continue
			}
			remaining := maxBytes - bytesRead
			if remaining <= 0 {
				_ = f.Close()
				warnings = append(warnings, "record byte budget reached; later attempts omitted")
				return out, warnings, nil
			}
			b, readErr := io.ReadAll(io.LimitReader(f, min(int64(suiteAttemptMaxJSON+1), remaining+1)))
			_ = f.Close()
			if readErr != nil {
				warnings = append(warnings, p+": "+readErr.Error())
				continue
			}
			if int64(len(b)) > remaining {
				warnings = append(warnings, "record byte budget reached; later attempts omitted")
				return out, warnings, nil
			}
			bytesRead += int64(len(b))
			a, decodeErr := decodeSuiteAttempt(b)
			if decodeErr != nil || e.Name() != a.ID+".json" {
				warnings = append(warnings, p+": invalid attempt record")
				continue
			}
			out = append(out, a)
		}
		if dirErr == io.EOF {
			break
		}
	}
	return out, warnings, nil
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
			if err == nil && now.Sub(info.ModTime()) > suiteInflightStale && suiteOwnerDead(filepath.Join(root, "inflight", id)) {
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
		if !suiteOwnerDead(filepath.Join(root, "inflight", e.Name())) {
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
