package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SessionSync is the seam over `couch --session-sync` (pair). Couch owns the
// file's guards: the operator's lock (exit 3), compare-and-replace by section
// hash (exit 6 on stale), and the owed-event ledger. tl-lead only drives it.
type SessionSync interface {
	ReadSection(file, heading string) (body, hash string, err error)
	// WriteSection replaces the section with the body in bodyPath if the
	// section still has hash, folding the given owed event IDs.
	WriteSection(file, heading, bodyPath, hash string, fold []string) error
	Status(file string) (SyncStatus, error)
	Diff(file string) (string, error)
	// MarkAllSeen reads the whole file (`read --all`), which tells Couch the
	// agent has now seen every edit.
	MarkAllSeen(file string) error
	// ReadFrontmatter reads the closed leading `---` block and its hash
	// (`read --frontmatter --json`, pair#643), for the discussions: index
	// (#424, #426). It marks only the frontmatter seen: the operator's
	// unread section edits stay in the diff.
	ReadFrontmatter(file string) (block, hash string, err error)
	// WriteFrontmatter replaces only the block with bodyPath's if it still
	// has hash; every body byte survives.
	WriteFrontmatter(file, bodyPath, hash string) error
	// Peek reads the whole file without moving couch's diff baseline, for
	// the body's links and for a decision that must not read through couch.
	Peek(file string) (string, error)
}

// SyncStatus is the part of `status --json` tl-lead reads.
type SyncStatus struct {
	Holder string
	Owed   []string
}

var (
	// ErrLocked: the operator holds the file's lock (couch exit 3).
	ErrLocked = errors.New("the operator holds the lock")
	// ErrStale: the section changed since it was read (couch exit 6).
	ErrStale = errors.New("stale: the section changed since it was read")
	// ErrDeliveryGap: couch failed to record a delivery to this slot, so the
	// owed events are unknown, not zero. Folding is impossible until the TL has
	// checked its recent messages; a write without --fold-owed still works.
	ErrDeliveryGap = errors.New("delivery gap: couch failed to record a delivery, so the owed events are unknown; " +
		"check recent messages, then write without --fold-owed")
)

const (
	couchExitLocked = 3
	couchExitStale  = 6
)

// couchSessionSync runs the real couch binary with an argv array, never a
// shell, so a quoted value can't merge with a flag.
type couchSessionSync struct{ bin string }

func (c couchSessionSync) run(args ...string) ([]byte, error) {
	cmd := exec.Command(c.bin, append([]string{"--session-sync"}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	switch {
	case err == nil:
		return stdout.Bytes(), nil
	case errors.As(err, &exit) && exit.ExitCode() == couchExitLocked:
		return nil, ErrLocked
	case errors.As(err, &exit) && exit.ExitCode() == couchExitStale:
		return nil, ErrStale
	}
	msg := strings.TrimSpace(stderr.String())
	if msg == "" {
		msg = strings.TrimSpace(stdout.String())
	}
	return nil, fmt.Errorf("couch --session-sync %s: %v: %s", args[0], err, msg)
}

// couchReply is the JSON envelope every --json verb prints.
type couchReply struct {
	Code   string
	Error  string
	Status *struct {
		Holder string
		Owed   []struct{ ID string }
	}
	Read *struct {
		Hash string
		Body string
	}
}

func decodeReply(verb string, raw []byte) (couchReply, error) {
	var r couchReply
	if err := json.Unmarshal(raw, &r); err != nil {
		return r, fmt.Errorf("couch --session-sync %s: unreadable reply: %v", verb, err)
	}
	if r.Code == "delivery-gap" {
		return r, fmt.Errorf("couch --session-sync %s: %w (%s)", verb, ErrDeliveryGap, r.Error)
	}
	if r.Code != "ok" {
		return r, fmt.Errorf("couch --session-sync %s refused (%s): %s", verb, r.Code, r.Error)
	}
	return r, nil
}

func (c couchSessionSync) ReadSection(file, heading string) (string, string, error) {
	raw, err := c.run("read", "--section", heading, file, "--json")
	if err != nil {
		return "", "", err
	}
	r, err := decodeReply("read", raw)
	if err != nil {
		return "", "", err
	}
	if r.Read == nil || r.Read.Hash == "" {
		return "", "", fmt.Errorf("couch --session-sync read: no hash in the reply for ## %s", heading)
	}
	return r.Read.Body, r.Read.Hash, nil
}

func (c couchSessionSync) WriteSection(file, heading, bodyPath, hash string, fold []string) error {
	args := []string{"write", "--section", heading, bodyPath, file, "--if-hash", hash}
	if len(fold) > 0 {
		args = append(args, "--fold", strings.Join(fold, ","))
	}
	_, err := c.run(args...)
	return err
}

func (c couchSessionSync) Status(file string) (SyncStatus, error) {
	raw, err := c.run("status", "--json", file)
	if err != nil {
		return SyncStatus{}, err
	}
	r, err := decodeReply("status", raw)
	if err != nil {
		return SyncStatus{}, err
	}
	if r.Status == nil {
		return SyncStatus{}, fmt.Errorf("couch --session-sync status: no status in the reply")
	}
	st := SyncStatus{Holder: r.Status.Holder}
	for _, o := range r.Status.Owed {
		st.Owed = append(st.Owed, o.ID)
	}
	return st, nil
}

func (c couchSessionSync) Diff(file string) (string, error) {
	raw, err := c.run("diff", file)
	return string(raw), err
}

func (c couchSessionSync) MarkAllSeen(file string) error {
	_, err := c.run("read", "--all", file)
	return err
}

func (c couchSessionSync) ReadFrontmatter(file string) (string, string, error) {
	raw, err := c.run("read", "--frontmatter", file, "--json")
	if err != nil {
		return "", "", err
	}
	r, err := decodeReply("read", raw)
	if err != nil {
		return "", "", err
	}
	if r.Read == nil || r.Read.Hash == "" {
		return "", "", fmt.Errorf("couch --session-sync read --frontmatter: no hash in the reply")
	}
	return r.Read.Body, r.Read.Hash, nil
}

func (c couchSessionSync) WriteFrontmatter(file, bodyPath, hash string) error {
	_, err := c.run("write", "--frontmatter", bodyPath, file, "--if-hash", hash)
	return err
}

// Peek reads the file on disk, as submit does for its markers: couch keeps
// the file current and only its own reads move the baseline.
func (c couchSessionSync) Peek(file string) (string, error) {
	raw, err := os.ReadFile(file)
	return string(raw), err
}
