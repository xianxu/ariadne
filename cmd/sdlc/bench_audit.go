package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// Agent-visible proposals are evidence, never authority to pass a gate. The
// observer binds these records to its independently verified runtime identities.
type benchAuditEvent struct {
	Version      int      `json:"version"`
	ID           string   `json:"id"`
	RunID        string   `json:"run_id"`
	Kind         string   `json:"kind"`
	At           string   `json:"at"`
	Root         string   `json:"root"`
	Slot         string   `json:"slot"`
	Session      string   `json:"session"`
	Generation   string   `json:"generation"`
	Command      []string `json:"command"`
	Route        []string `json:"route,omitempty"`
	ParentID     string   `json:"parent_id"`
	Success      *bool    `json:"success"`
	Error        string   `json:"error"`
	DecisionKind string   `json:"decision_kind"`
	Reason       string   `json:"reason"`
	Candidate    string   `json:"candidate"`
}

func benchAuditPath(c *benchPublicContext) string {
	return filepath.Join(c.WorkRoot, ".git", "sdlc", "bench-audit.jsonl")
}

func newBenchEventID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func withBenchCommand(ctx context.Context, c *benchPublicContext, root string, args []string, invoke func() error) error {
	if c == nil {
		return invoke()
	}
	id, err := newBenchEventID()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		args = []string{"help"}
	}
	start := benchAuditEvent{Version: 1, ID: id, RunID: c.RunID, Kind: "command-start", At: time.Now().UTC().Format(time.RFC3339Nano), Root: root, Command: append([]string(nil), args...)}
	if route, ok := ctx.Value(benchCommandRouteKey{}).([]string); ok {
		start.Route = append([]string(nil), route...)
	}
	if err := appendBenchAudit(ctx, c, start); err != nil {
		return fmt.Errorf("bench audit unavailable before command: %w", err)
	}
	finished := false
	finish := func(outcome error) error {
		if finished {
			return nil
		}
		finished = true
		end := start
		end.ID = id + "-end"
		end.Kind = "command-end"
		end.ParentID = id
		end.At = time.Now().UTC().Format(time.RFC3339Nano)
		success := outcome == nil
		end.Success = &success
		if outcome != nil {
			end.Error = outcome.Error()
			if len(end.Error) > 4096 {
				end.Error = end.Error[:4096]
				for !utf8.ValidString(end.Error) {
					end.Error = end.Error[:len(end.Error)-1]
				}
			}
		}
		return appendBenchAudit(ctx, c, end)
	}
	unregister := registerDieCleanup(func() {
		if auditErr := finish(errors.New("legacy command exited before returning; see gate output")); auditErr != nil {
			fmt.Fprintln(os.Stderr, "bench refusal audit unavailable:", auditErr)
		}
	})
	defer unregister()
	err = guardBenchCommand(c, args)
	if err == nil {
		err = invoke()
	}
	// Preserve a failed gate's error even if recording its outcome also fails.
	return errors.Join(err, finish(err))
}

func validateBenchAuditEvent(c *benchPublicContext, e benchAuditEvent) error {
	if e.Version != 1 || e.ID == "" || len(e.ID) > 128 || e.RunID != c.RunID || (e.Root != c.WorkRoot && e.Root != c.HubRoot) {
		return errors.New("bench audit identity mismatch")
	}
	if _, err := time.Parse(time.RFC3339Nano, e.At); err != nil {
		return fmt.Errorf("bench audit time: %w", err)
	}
	if len(e.Error) > 4096 || len(e.Reason) > 4096 || len(e.DecisionKind) > 128 || len(e.Command) > 128 {
		return errors.New("bench audit field exceeds bound")
	}
	for _, value := range []string{e.ID, e.RunID, e.Root, e.Slot, e.Session, e.Generation, e.DecisionKind, e.Candidate} {
		if strings.ContainsAny(value, "\x00\r\n") {
			return errors.New("invalid bench audit identity")
		}
	}
	for _, arg := range e.Command {
		if len(arg) > 16<<10 || strings.ContainsRune(arg, 0) {
			return errors.New("bench audit command exceeds bound")
		}
	}
	switch e.Kind {
	case "command-start":
		if len(e.Command) == 0 || e.Success != nil || e.ParentID != "" {
			return errors.New("invalid command start")
		}
	case "command-end":
		if len(e.Command) == 0 || e.Success == nil || e.ParentID == "" {
			return errors.New("invalid command end")
		}
	case "decision", "finish":
		if e.Slot == "" || e.Session == "" || e.Generation == "" || e.Root != c.HubRoot {
			return errors.New("bench TL proposal needs bound hub identity")
		}
		if e.Kind == "decision" && (strings.TrimSpace(e.DecisionKind) == "" || strings.TrimSpace(e.Reason) == "") {
			return errors.New("bench decision needs kind and reason")
		}
		if e.Kind == "finish" && !benchCandidatePattern.MatchString(e.Candidate) {
			return errors.New("bench finish needs full candidate OID")
		}
	default:
		return errors.New("unsupported bench audit kind")
	}
	return nil
}

func decodeBenchAudit(c *benchPublicContext, raw []byte) ([]benchAuditEvent, error) {
	if len(raw) > benchRecordLimit || !utf8.Valid(raw) {
		return nil, errors.New("bench audit exceeds bound or is invalid UTF8")
	}
	if len(raw) > 0 && raw[len(raw)-1] != '\n' {
		return nil, errors.New("bench audit has a partial tail; retain and reconcile")
	}
	var events []benchAuditEvent
	ids := map[string]bool{}
	for _, line := range bytes.Split(raw, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		if len(events) >= benchEventLimit {
			return nil, errors.New("bench audit record budget exhausted")
		}
		if err := jsonstrict.RejectDuplicateKeys(line); err != nil {
			return nil, err
		}
		var e benchAuditEvent
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, err
		}
		if err := validateBenchAuditEvent(c, e); err != nil {
			return nil, err
		}
		// Known-field spellings are authority-bearing; unknown members remain in
		// the append-only original bytes and are never rewritten by this reader.
		canonical, _ := json.Marshal(e)
		var known, object map[string]json.RawMessage
		if err := json.Unmarshal(canonical, &known); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(line, &object); err != nil {
			return nil, err
		}
		for key := range object {
			for name := range known {
				if strings.EqualFold(key, name) && key != name {
					return nil, errors.New("noncanonical bench audit field")
				}
			}
		}
		if ids[e.ID] {
			return nil, errors.New("duplicate bench audit event identity")
		}
		ids[e.ID] = true
		events = append(events, e)
	}
	return events, nil
}

func readBenchAudit(ctx context.Context, c *benchPublicContext) ([]benchAuditEvent, error) {
	current, err := loadBenchContext(ctx, c.WorkRoot)
	if err != nil {
		return nil, err
	}
	if current == nil || *current != *c {
		return nil, errors.New("bench audit context changed")
	}
	f, err := os.OpenFile(benchAuditPath(c), os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, benchRecordLimit+1))
	if err != nil {
		return nil, err
	}
	return decodeBenchAudit(c, raw)
}

func appendBenchAudit(ctx context.Context, c *benchPublicContext, e benchAuditEvent) error {
	if err := validateBenchAuditEvent(c, e); err != nil {
		return err
	}
	current, err := loadBenchContext(ctx, c.WorkRoot)
	if err != nil {
		return err
	}
	if current == nil || *current != *c {
		return errors.New("bench audit context changed")
	}
	dir := filepath.Dir(benchAuditPath(c))
	if err := os.Mkdir(dir, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("bench audit directory substituted")
	}
	lock, err := os.OpenFile(filepath.Join(dir, "bench-audit.lock"), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err := flock.Exclusive(lock, true); err != nil {
		return fmt.Errorf("bench audit busy; retry command: %w", err)
	}
	f, err := os.OpenFile(benchAuditPath(c), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err = f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return errors.New("bench audit is not a regular file")
	}
	raw, err := io.ReadAll(io.LimitReader(f, benchRecordLimit+1))
	if err != nil {
		return err
	}
	events, err := decodeBenchAudit(c, raw)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(e)
	if err != nil {
		return err
	}
	for _, previous := range events {
		if previous.ID == e.ID {
			old, err := json.Marshal(previous)
			if err != nil {
				return err
			}
			if !bytes.Equal(old, encoded) {
				return errors.New("conflicting bench audit event")
			}
			return nil
		}
	}
	if len(events) >= benchEventLimit || len(raw)+len(encoded)+1 > benchRecordLimit {
		return errors.New("bench audit budget exhausted; retain run")
	}
	if _, err := f.Write(append(encoded, '\n')); err != nil {
		return err
	}
	return f.Sync()
}
