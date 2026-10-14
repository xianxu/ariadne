// Package reviewusage retains nonsemantic reviewer attribution independently of
// gate acceptance. Requested/Started receipts survive process death as unknown
// progress; they never authorize a provider retry or a gate round.
package reviewusage

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const Version = 1
const MaxBytes = 2 << 10

type Phase string

const (
	Requested Phase = "requested"
	Started   Phase = "started"
	Finished  Phase = "finished"
)

type Intent struct {
	ID           string `json:"id" yaml:"id"`
	Repository   string `json:"repository" yaml:"repository"`
	IssueID      string `json:"issue_id" yaml:"issue_id"`
	Recipe       string `json:"recipe" yaml:"recipe"`
	Round        int    `json:"round" yaml:"round"`
	Attempt      int    `json:"attempt" yaml:"attempt"`
	Agent        string `json:"agent" yaml:"agent"`
	Root         string `json:"root" yaml:"root"`
	ReviewedHEAD string `json:"reviewed_head" yaml:"reviewed_head"`
	Slot         string `json:"slot" yaml:"slot"`
	RequestedAt  string `json:"requested_at" yaml:"requested_at"`
}
type Record struct {
	Version    int    `json:"version" yaml:"version"`
	Intent     Intent `json:"intent" yaml:"intent"`
	Phase      Phase  `json:"phase" yaml:"phase"`
	PID        int    `json:"pid,omitempty" yaml:"pid,omitempty"`
	StartedAt  string `json:"started_at,omitempty" yaml:"started_at,omitempty"`
	FinishedAt string `json:"finished_at,omitempty" yaml:"finished_at,omitempty"`
	SessionID  string `json:"session_id,omitempty" yaml:"session_id,omitempty"`
	Outcome    string `json:"outcome,omitempty" yaml:"outcome,omitempty"`
	Diagnostic string `json:"diagnostic,omitempty" yaml:"diagnostic,omitempty"`
}

type EventKind string

const (
	Start  EventKind = "start"
	Finish EventKind = "finish"
)

type Event struct {
	Kind                                          EventKind
	Intent                                        Intent
	PID                                           int
	At, StartedAt, SessionID, Outcome, Diagnostic string
}

var token = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
var oid = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)
var issueID = regexp.MustCompile(`^[0-9]{6}$`)
var sessionUUID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func singleLine(v string) bool {
	return strings.TrimSpace(v) != "" && utf8.ValidString(v) && strings.IndexFunc(v, unicode.IsControl) < 0
}
func timestamp(v string) (time.Time, error) { return time.Parse(time.RFC3339Nano, v) }
func Validate(r Record) error {
	i := r.Intent
	if r.Version != Version || !token.MatchString(i.ID) || !issueID.MatchString(i.IssueID) || i.IssueID == "000000" || !oid.MatchString(i.ReviewedHEAD) || i.Round < 1 || i.Attempt < 1 || i.Attempt > 2 {
		return errors.New("invalid reviewer intent identity/version")
	}
	for _, v := range []string{i.Repository, i.Recipe, i.Root, i.Slot} {
		if !singleLine(v) {
			return errors.New("invalid reviewer intent field")
		}
	}
	if !filepath.IsAbs(i.Root) {
		return errors.New("reviewer root must be absolute")
	}
	if i.Agent != "claude" && i.Agent != "codex" && i.Agent != "gemini" {
		return errors.New("unsupported reviewer agent")
	}
	requested, err := timestamp(i.RequestedAt)
	if err != nil {
		return errors.New("invalid reviewer requested timestamp")
	}
	if len(r.Diagnostic) > 256 || !utf8.ValidString(r.Diagnostic) || strings.IndexFunc(r.Diagnostic, unicode.IsControl) >= 0 {
		return errors.New("invalid reviewer diagnostic")
	}
	if r.SessionID != "" && !sessionUUID.MatchString(r.SessionID) {
		return errors.New("invalid exact reviewer session ID")
	}
	var started time.Time
	if r.StartedAt != "" {
		started, err = timestamp(r.StartedAt)
		if err != nil || started.Before(requested) || r.PID <= 0 {
			return errors.New("invalid reviewer start confirmation")
		}
	}
	switch r.Phase {
	case Requested:
		if r.PID != 0 || r.StartedAt != "" || r.FinishedAt != "" || r.SessionID != "" || r.Outcome != "" || r.Diagnostic != "" {
			return errors.New("requested reviewer contains unconfirmed progress")
		}
	case Started:
		if r.PID <= 0 || r.StartedAt == "" || r.FinishedAt != "" || r.SessionID != "" || r.Outcome != "" || r.Diagnostic != "" {
			return errors.New("invalid started reviewer shape")
		}
	case Finished:
		finish, err := timestamp(r.FinishedAt)
		if err != nil || finish.Before(requested) || (!started.IsZero() && finish.Before(started)) {
			return errors.New("invalid reviewer finish confirmation")
		}
		switch r.Outcome {
		case "returned", "exit-error", "agent-error", "canceled", "deadline":
			if r.PID <= 0 || r.StartedAt == "" {
				return errors.New("launched reviewer finish requires confirmed start")
			}
		case "launch-failed":
			if r.PID != 0 || r.StartedAt != "" || r.SessionID != "" {
				return errors.New("launch failure cannot carry launched identity")
			}
		default:
			return errors.New("unknown reviewer process outcome")
		}
	default:
		return errors.New("unknown reviewer accounting phase")
	}
	raw, err := json.Marshal(r)
	if err != nil || len(raw) > MaxBytes {
		return errors.New("reviewer record exceeds 2 KiB")
	}
	return nil
}
func NewRequested(i Intent) (Record, error) {
	r := Record{Version: Version, Intent: i, Phase: Requested}
	return r, Validate(r)
}
func (r Record) Uncertain() bool { return r.Phase != Finished }

// Step consumes complete callback identity. Exact confirmations repeat safely;
// out-of-order or rewritten callbacks return the prior state without mutation.
func Step(r Record, e Event) (Record, error) {
	refuse := func(err error) (Record, error) { return r, err }
	if err := Validate(r); err != nil {
		return refuse(err)
	}
	if e.Intent != r.Intent {
		return refuse(errors.New("reviewer intent binding mismatch"))
	}
	next := r
	switch e.Kind {
	case Start:
		if r.Phase == Finished {
			return refuse(errors.New("reviewer start out of order"))
		}
		if r.Phase == Started {
			if e.PID == r.PID && e.At == r.StartedAt {
				return r, nil
			}
			return refuse(errors.New("reviewer start confirmation rewritten"))
		}
		next.Phase = Started
		next.PID = e.PID
		next.StartedAt = e.At
	case Finish:
		if r.Phase == Requested && e.Outcome != "launch-failed" {
			return refuse(errors.New("reviewer finish without start confirmation"))
		}
		if r.Phase != Requested && (e.PID != r.PID || e.StartedAt != r.StartedAt) {
			return refuse(errors.New("reviewer start binding mismatch"))
		}
		next.Phase = Finished
		next.FinishedAt = e.At
		next.SessionID = e.SessionID
		next.Outcome = e.Outcome
		next.Diagnostic = e.Diagnostic
		if r.Phase == Finished && next != r {
			return refuse(errors.New("reviewer finish confirmation rewritten"))
		}
	default:
		return refuse(fmt.Errorf("unknown reviewer event %q", e.Kind))
	}
	if err := Validate(next); err != nil {
		return refuse(err)
	}
	return next, nil
}

// Advance validates the complete confirmed prefix for CAS persistence. A finish
// may persist a known launch and return together after its first write failed.
func Advance(prior, next Record) error {
	if err := Validate(next); err != nil {
		return err
	}
	if prior == next {
		return nil
	}
	current := prior
	var err error
	if current.Phase == Requested && next.StartedAt != "" {
		current, err = Step(current, Event{Kind: Start, Intent: next.Intent, PID: next.PID, At: next.StartedAt})
		if err != nil {
			return err
		}
	}
	if next.Phase == Finished {
		current, err = Step(current, Event{Kind: Finish, Intent: next.Intent, PID: next.PID, StartedAt: next.StartedAt, At: next.FinishedAt, SessionID: next.SessionID, Outcome: next.Outcome, Diagnostic: next.Diagnostic})
		if err != nil {
			return err
		}
	}
	if current != next {
		return errors.New("reviewer persistence would rewrite or regress confirmed progress")
	}
	return nil
}
