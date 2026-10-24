// engagement.go derives bounded ownership intervals from pinned tracker evidence.
package tracker

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// OwnershipEvent is one chronological card snapshot. A nil Owner releases it.
type OwnershipEvent struct {
	At        time.Time
	Owner     *issue.Claimant
	Operation string
	Started   time.Time
	// Status distinguishes a live reservation from a retained terminal owner.
	// DeriveEngagements deliberately keeps both as attribution evidence.
	Status string
}

// EngagementWindow is a half-open interval with all source workspaces that
// participated in this engagement. A relocation keeps its original Began.
type EngagementWindow struct {
	Start, End time.Time
	Began      time.Time
	Owners     []issue.Claimant
}

// DeriveEngagements rejects inconsistent evidence rather than guessing. Only
// the first evidenced claim may use Started; a later claim always begins anew.
func DeriveEngagements(events []OwnershipEvent, until time.Time) ([]EngagementWindow, error) {
	if until.IsZero() {
		return nil, fmt.Errorf("engagement cutoff is missing")
	}
	var previous time.Time
	for i, e := range events {
		if e.At.IsZero() || (!previous.IsZero() && e.At.Before(previous)) {
			return nil, fmt.Errorf("ownership event %d has missing or out-of-order time", i)
		}
		if !e.Started.IsZero() && e.Started.After(e.At) {
			return nil, fmt.Errorf("ownership event %d started after its commit", i)
		}
		if e.Owner != nil {
			if err := validateEngagementOwner(*e.Owner); err != nil {
				return nil, fmt.Errorf("ownership event %d: %w", i, err)
			}
		}
		previous = e.At
	}
	var windows []EngagementWindow
	var active *EngagementWindow
	var owner *issue.Claimant
	firstClaim := true
	closeAt := func(at time.Time) {
		if active == nil {
			return
		}
		if at.After(until) {
			at = until
		}
		active.End = at
		if active.Start.Before(at) {
			copyWindow := *active
			copyWindow.Owners = append([]issue.Claimant(nil), active.Owners...)
			windows = append(windows, copyWindow)
		}
		active = nil
	}
	for i, e := range events {
		if e.At.After(until) {
			break
		}
		if e.Owner == nil {
			closeAt(e.At)
			owner = nil
			continue
		}
		unchanged := owner != nil && owner.Machine == e.Owner.Machine && owner.Repository == e.Owner.Repository && owner.Worktree == e.Owner.Worktree && owner.Workspace == e.Owner.Workspace
		if unchanged && e.Operation != OpReclaim {
			continue
		}
		start, began := e.At, e.At
		roots := []issue.Claimant{*e.Owner}
		if active != nil && e.Operation == OpRelocate {
			if owner.Machine != e.Owner.Machine || owner.Repository != e.Owner.Repository {
				return nil, fmt.Errorf("relocation changes machine or repository")
			}
			began = active.Began
			roots = append(append([]issue.Claimant(nil), active.Owners...), *e.Owner)
		}
		if firstClaim && e.Operation == OpClaim && !e.Started.IsZero() && (i == 0 || !e.Started.Before(events[0].At)) {
			start = e.Started
			began = start
		}
		closeAt(e.At)
		active = &EngagementWindow{Start: start, Began: began, Owners: roots}
		copied := *e.Owner
		owner = &copied
		firstClaim = false
	}
	closeAt(until)
	return windows, nil
}

func validateEngagementOwner(c issue.Claimant) error {
	if !issue.ValidFingerprint(c.Machine) || !filepath.IsAbs(c.Worktree) {
		return fmt.Errorf("invalid owner machine or worktree")
	}
	for _, s := range []string{c.Operator, c.MachineName, c.Repository, c.Worktree} {
		if strings.TrimSpace(s) == "" || strings.ContainsAny(s, "\r\n") {
			return fmt.Errorf("invalid owner identity")
		}
	}
	if strings.ContainsAny(c.Workspace, "\r\n") {
		return fmt.Errorf("invalid owner workspace")
	}
	return nil
}

// EngagementHistory reads complete per-card history at an immutable tracker
// commit. The extra record detects truncation, which must never look like an
// initial claim. No fetch or ambient checkout participates in these reads.
func (r *Repository) EngagementHistory(pin, cardPath string) ([]OwnershipEvent, error) {
	stream, finish, err := r.trunk.HistoryStreamAt(pin, "%x01%H%x00%cI%x00%B%x00", cardPath, ClaimHistoryLimit+1)
	if err != nil {
		return nil, err
	}
	records, err := readEngagementHistoryRecords(stream, ClaimHistoryLimit)
	if err != nil {
		_ = finish(true)
		return nil, err
	}
	if err := finish(false); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no engagement history for %s", cardPath)
	}
	slices.Reverse(records)
	events := make([]OwnershipEvent, 0, len(records))
	for _, rec := range records {
		raw, err := r.trunk.ViewOf(rec.commit).Read(cardPath)
		if err != nil {
			return nil, err
		}
		card, err := issue.ParseCard(raw)
		if err != nil {
			return nil, fmt.Errorf("engagement card %s at %s: %w", cardPath, rec.commit, err)
		}
		owner, has, err := issue.CardClaimant(raw)
		if err != nil {
			return nil, err
		}
		event := OwnershipEvent{At: rec.at, Operation: rec.operation}
		event.Status, err = issue.FieldString(card.Frontmatter, "status")
		if err != nil {
			return nil, err
		}
		if has {
			event.Owner = &owner
		}
		started, err := issue.FieldString(card.Frontmatter, "started")
		if err != nil {
			return nil, err
		}
		if started != "" {
			event.Started, err = time.Parse(time.RFC3339, started)
			if err != nil {
				return nil, fmt.Errorf("engagement started: %w", err)
			}
		}
		events = append(events, event)
	}
	if events[0].Owner != nil && events[0].Operation != OpClaim {
		return nil, fmt.Errorf("incomplete engagement history for %s: initial ownership lacks claim evidence", cardPath)
	}
	return events, nil
}

type engagementHistoryRecord struct {
	commit    string
	at        time.Time
	operation string
}

func readEngagementHistoryRecords(r io.Reader, limit int) ([]engagementHistoryRecord, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	sc.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		start := 0
		if len(data) > 0 && data[0] == 1 {
			start = 1
		}
		if i := strings.IndexByte(string(data[start:]), 1); i >= 0 {
			return start + i, data[start : start+i], nil
		}
		if atEOF && len(data) > start {
			return len(data), data[start:], nil
		}
		return 0, nil, nil
	})
	var records []engagementHistoryRecord
	for sc.Scan() {
		if len(records) >= limit {
			return nil, fmt.Errorf("incomplete engagement history: exceeds %d card commits", limit)
		}
		fields := strings.SplitN(sc.Text(), "\x00", 4)
		if len(fields) != 4 {
			return nil, fmt.Errorf("malformed engagement history record")
		}
		commit := strings.Clone(strings.TrimSpace(fields[0]))
		if len(commit) != 40 && len(commit) != 64 {
			return nil, fmt.Errorf("malformed engagement commit")
		}
		for _, c := range commit {
			if !strings.ContainsRune("0123456789abcdef", c) {
				return nil, fmt.Errorf("malformed engagement commit")
			}
		}
		at, err := time.Parse(time.RFC3339, strings.TrimSpace(fields[1]))
		if err != nil {
			return nil, fmt.Errorf("engagement commit time: %w", err)
		}
		operation := ""
		if m := operationLine.FindStringSubmatch(fields[2]); m != nil {
			operation, _, _ = strings.Cut(m[1], "-")
			operation = strings.Clone(operation)
		}
		records = append(records, engagementHistoryRecord{commit: commit, at: at, operation: operation})
	}
	return records, sc.Err()
}
