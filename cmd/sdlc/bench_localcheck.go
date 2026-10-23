package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// localCheckPublicationDeps is shared by run and republish. A bench context
// selects private identity before any GitHub repository or PR discovery.
func localCheckPublicationDeps(ctx context.Context, root string, git func(...string) (string, error), stderr io.Writer) (localRunDeps, error) {
	d := localRunDeps{Root: root, Git: git, Stderr: stderr, FlakeRegistry: gateFlakeRegistry(root)}
	b, err := loadBenchContext(ctx, root)
	if err != nil {
		return d, err
	}
	if b != nil {
		if root != b.WorkRoot {
			return d, errors.New("bench local-check publication requires the worker checkout")
		}
		d.Repo = b.OriginIdentity
		d.RecordStatus = func(ctx context.Context, rec localRunRecord) error {
			return writeBenchLocalCheckReceipt(ctx, root, rec)
		}
		return d, nil
	}
	d.Repo, err = detectRepo()
	if err != nil {
		return d, err
	}
	d.GH = ghClient
	d.PR = currentBranchPR(git, d.Repo)
	return d, nil
}

// This is status evidence, not alternate proof authority: the normal same-head
// execution/flake guard must pass before the writer is called. It stays public
// and contains no hidden acceptance or scorer inputs.
type benchLocalCheckReceipt struct {
	Version        int            `json:"version"`
	RunID          string         `json:"run_id"`
	WorkRoot       string         `json:"work_root"`
	OriginIdentity string         `json:"origin_identity"`
	Head           string         `json:"head"`
	Check          string         `json:"check"`
	Record         localRunRecord `json:"record"`
}

func benchLocalCheckReceiptPath(b *benchPublicContext, head, check string) (string, error) {
	if b == nil || (len(head) != 40 && len(head) != 64) || !suiteHex.MatchString(head) || !localCheckName.MatchString(check) {
		return "", errors.New("invalid bench status subject")
	}
	return filepath.Join(b.WorkRoot, ".git", "sdlc", "bench-statuses", head, check+".json"), nil
}

func decodeBenchLocalCheckReceipt(raw []byte, b *benchPublicContext, head, check string) (benchLocalCheckReceipt, error) {
	var r benchLocalCheckReceipt
	if len(raw) > benchRecordLimit {
		return r, errors.New("bench status receipt exceeds 1 MiB")
	}
	if _, err := benchLocalCheckReceiptPath(b, head, check); err != nil {
		return r, err
	}
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return r, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return r, err
	}
	for _, key := range []string{"version", "run_id", "work_root", "origin_identity", "head", "check", "record"} {
		if value, ok := fields[key]; !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return r, fmt.Errorf("bench status receipt missing %s", key)
		}
	}
	// localRunRecord emits all its fields. Missing/null scalar outcomes must
	// never decode as exit-zero or untimed-out defaults after a damaged write.
	var recordFields, required map[string]json.RawMessage
	if err := json.Unmarshal(fields["record"], &recordFields); err != nil {
		return r, err
	}
	zero, _ := json.Marshal(localRunRecord{})
	if err := json.Unmarshal(zero, &required); err != nil {
		return r, err
	}
	for key := range required {
		value, ok := recordFields[key]
		if !ok || (key != "Tail" && key != "KnownFlakes" && bytes.Equal(bytes.TrimSpace(value), []byte("null"))) {
			return r, fmt.Errorf("bench status receipt record missing %s", key)
		}
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return r, err
	}
	if r.Version != 1 || r.RunID != b.RunID || r.WorkRoot != b.WorkRoot || r.OriginIdentity != b.OriginIdentity || r.Head != head || r.Check != check || r.Record.SHA != head || r.Record.Name != check {
		return r, errors.New("bench status receipt identity/version mismatch")
	}
	if len(r.Record.LogSHA256) != 64 || !suiteHex.MatchString(r.Record.LogSHA256) || r.Record.Duration < 0 {
		return r, errors.New("bench status receipt lacks complete log evidence")
	}
	return r, nil
}

func readBenchLocalCheckReceipt(ctx context.Context, root, head, check string) (benchLocalCheckReceipt, error) {
	b, err := loadBenchContext(ctx, root)
	if err != nil {
		return benchLocalCheckReceipt{}, err
	}
	if b == nil || root != b.WorkRoot {
		return benchLocalCheckReceipt{}, errors.New("no worker bench context")
	}
	path, err := benchLocalCheckReceiptPath(b, head, check)
	if err != nil {
		return benchLocalCheckReceipt{}, err
	}
	raw, err := readBenchStatusFile(path)
	if err != nil {
		return benchLocalCheckReceipt{}, err
	}
	r, err := decodeBenchLocalCheckReceipt(raw, b, head, check)
	if err != nil || !r.Record.passed() {
		return r, err
	}
	query := func(args ...string) (string, error) {
		out, diag, err := gitx.RunGitInContext(ctx, root, nil, args...)
		if err != nil {
			return "", fmt.Errorf("bench receipt Git proof: %w: %s", err, diag)
		}
		return string(out), nil
	}
	d := localRunDeps{Root: root, Git: query, Repo: b.OriginIdentity, Stderr: io.Discard, FlakeRegistry: gateFlakeRegistry(root)}
	logDir, err := localCheckLogDir(d)
	if err != nil {
		return r, err
	}
	saved, newest, err := loadSavedAttemptCurrent(ctx, d, suiteStoreRoot(logDir), head, check, "", r.Record.Host)
	if err != nil {
		return r, err
	}
	// Saved wall seconds are a float64; decoding can round one nanosecond.
	delta := saved.Duration - r.Record.Duration
	if !newest || !saved.passed() || saved.Exit != r.Record.Exit || saved.TimedOut != r.Record.TimedOut || saved.Outcome != r.Record.Outcome || saved.LogSHA256 != r.Record.LogSHA256 || delta > time.Nanosecond || delta < -time.Nanosecond || saved.ResumeParent != r.Record.ResumeParent || saved.ResumeChild != r.Record.ResumeChild || !slices.Equal(saved.KnownFlakes, r.Record.KnownFlakes) {
		return r, errors.New("bench status receipt no longer matches the newest verified attempt")
	}
	if isTestCheck(check) {
		if err := guardSuitePassRecord(ctx, d, check, head, &r.Record); err != nil {
			return r, err
		}
	}
	if saved.ResumeParent == "" && (saved.Outcome == flakes.KnownFlakePassed || saved.Outcome == flakes.KnownFlakeAccept) {
		c, err := declaredCheck(ctx, query, check, head)
		if err != nil {
			return r, err
		}
		if next := revalidateKnownFlake(ctx, d, c, saved.knownFlakeOutcome()); next.Kind != saved.Outcome {
			return r, errors.New("bench status known-flake allowance no longer holds: " + next.Reason)
		}
	}
	return r, nil
}

func readBenchStatusFile(path string) ([]byte, error) {
	// Directory ancestry cannot redirect an otherwise no-follow leaf open.
	canonical, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	if canonical != filepath.Dir(path) {
		return nil, errors.New("bench status directory has symlink ancestry")
	}
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("bench status receipt is not a regular file")
	}
	raw, err := io.ReadAll(io.LimitReader(f, benchRecordLimit+1))
	if len(raw) > benchRecordLimit {
		return nil, errors.New("bench status receipt exceeds 1 MiB")
	}
	return raw, err
}

func writeBenchLocalCheckReceipt(ctx context.Context, root string, rec localRunRecord) error {
	b, err := loadBenchContext(ctx, root)
	if err != nil {
		return err
	}
	if b == nil || root != b.WorkRoot {
		return errors.New("no worker bench context")
	}
	r := benchLocalCheckReceipt{Version: 1, RunID: b.RunID, WorkRoot: b.WorkRoot, OriginIdentity: b.OriginIdentity, Head: rec.SHA, Check: rec.Name, Record: rec}
	raw, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if _, err := decodeBenchLocalCheckReceipt(raw, b, rec.SHA, rec.Name); err != nil {
		return err
	}
	path, err := benchLocalCheckReceiptPath(b, rec.SHA, rec.Name)
	if err != nil {
		return err
	}
	// Make one directory component at a time and reject links before descending.
	for _, dir := range []string{filepath.Join(root, ".git", "sdlc"), filepath.Join(root, ".git", "sdlc", "bench-statuses"), filepath.Dir(path)} {
		if err := os.Mkdir(dir, 0700); err != nil && !os.IsExist(err) {
			return err
		}
		canonical, err := filepath.EvalSymlinks(dir)
		if err != nil || canonical != dir {
			return fmt.Errorf("bench status directory changed: %s", dir)
		}
		info, err := os.Lstat(dir)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New("bench status directory is not a directory")
		}
	}
	if old, err := readBenchStatusFile(path); err == nil {
		if _, err := decodeBenchLocalCheckReceipt(old, b, rec.SHA, rec.Name); err != nil {
			return err
		}
		raw, err = mergeBenchStatusExtensions(old, raw)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if len(raw) > benchRecordLimit {
		return errors.New("bench status receipt exceeds 1 MiB")
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".status-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, writeErr := f.Write(raw)
	if writeErr == nil {
		writeErr = f.Sync()
	}
	if err := errors.Join(writeErr, f.Close()); err != nil {
		return err
	}
	current, err := loadBenchContext(ctx, root)
	if err != nil {
		return err
	}
	if current == nil || *current != *b {
		return errors.New("bench context changed before status publication")
	}
	return os.Rename(f.Name(), path)
}

// Preserve object extensions when replacing a latest receipt with newly
// verified evidence. Known fields take the new value; no private fields enter.
func mergeBenchStatusExtensions(old, current []byte) ([]byte, error) {
	var previous, next map[string]json.RawMessage
	if err := json.Unmarshal(old, &previous); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(current, &next); err != nil {
		return nil, err
	}
	for key, value := range previous {
		replacement, ok := next[key]
		if !ok {
			next[key] = value
			continue
		}
		if bytes.HasPrefix(bytes.TrimSpace(value), []byte("{")) && bytes.HasPrefix(bytes.TrimSpace(replacement), []byte("{")) {
			merged, err := mergeBenchStatusExtensions(value, replacement)
			if err != nil {
				return nil, err
			}
			next[key] = merged
		}
		if key == "KnownFlakes" {
			merged, err := mergeBenchFlakeExtensions(value, replacement)
			if err != nil {
				return nil, err
			}
			next[key] = merged
		}
	}
	return json.Marshal(next)
}

// Flake entries have stable test+issue identity; preserve future members only
// on that same entry, including when the registry's ordering has changed.
func mergeBenchFlakeExtensions(old, current []byte) ([]byte, error) {
	var previous, next []json.RawMessage
	if err := json.Unmarshal(old, &previous); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(current, &next); err != nil {
		return nil, err
	}
	identity := func(raw json.RawMessage) (flakes.Entry, error) {
		var entry flakes.Entry
		err := json.Unmarshal(raw, &entry)
		return entry, err
	}
	byIdentity := make(map[flakes.Entry]json.RawMessage, len(previous))
	for _, raw := range previous {
		entry, err := identity(raw)
		if err != nil {
			return nil, err
		}
		byIdentity[entry] = raw
	}
	for i, raw := range next {
		entry, err := identity(raw)
		if err != nil {
			return nil, err
		}
		if former, ok := byIdentity[entry]; ok {
			next[i], err = mergeBenchStatusExtensions(former, raw)
			if err != nil {
				return nil, err
			}
		}
	}
	return json.Marshal(next)
}
