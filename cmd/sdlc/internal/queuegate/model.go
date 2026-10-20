// Package queuegate contains the machine-wide admission gate's pure model.
package queuegate

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxOwners     = 128
	MaxOperations = 4096
	MaxJSON       = 1 << 20
	MaxToken      = 256
	MaxText       = 4096
)

func boundedText(s string, limit int) bool { return len(s) <= limit && utf8.ValidString(s) }

type PauseOwner struct {
	Owner         string                     `json:"owner"`
	Reason        string                     `json:"reason"`
	CancelRunning bool                       `json:"cancel_running"`
	Extra         map[string]json.RawMessage `json:"-"`
}
type Control struct {
	SchemaVersion int                        `json:"schema_version"`
	PauseOwners   []PauseOwner               `json:"pause_owners"`
	Extra         map[string]json.RawMessage `json:"-"`
}
type Action struct {
	Kind, Owner, Reason string
	CancelRunning       bool
}

func NewControl() Control { return Control{SchemaVersion: 1, PauseOwners: []PauseOwner{}} }

func validateOwners(owners []PauseOwner) error {
	if owners == nil || len(owners) > MaxOwners {
		return errors.New("pause_owners must be an array of at most 128 owners")
	}
	seen := map[string]bool{}
	for _, p := range owners {
		if strings.TrimSpace(p.Owner) == "" || !boundedText(p.Owner, MaxToken) || !boundedText(p.Reason, MaxText) {
			return errors.New("invalid pause owner or reason bound")
		}
		if seen[p.Owner] {
			return fmt.Errorf("duplicate pause owner %q", p.Owner)
		}
		seen[p.Owner] = true
	}
	return nil
}
func validateControl(c Control) error {
	if c.SchemaVersion != 1 {
		return errors.New("unsupported control schema_version")
	}
	return validateOwners(c.PauseOwners)
}
func cloneExtra(in map[string]json.RawMessage) map[string]json.RawMessage {
	if in == nil {
		return nil
	}
	out := make(map[string]json.RawMessage, len(in))
	for k, v := range in {
		out[k] = append(json.RawMessage(nil), v...)
	}
	return out
}
func cloneOwners(in []PauseOwner) []PauseOwner {
	out := make([]PauseOwner, len(in))
	copy(out, in)
	for i := range out {
		out[i].Extra = cloneExtra(out[i].Extra)
	}
	return out
}

// Transition changes exactly one owner's hold. Replays retain the original
// reason and can strengthen cancellation, but never undo it.
func Transition(c Control, a Action) (Control, error) {
	if err := validateControl(c); err != nil {
		return Control{}, err
	}
	if strings.TrimSpace(a.Owner) == "" || !boundedText(a.Owner, MaxToken) || !boundedText(a.Reason, MaxText) {
		return Control{}, errors.New("invalid owner or reason bound")
	}
	if a.Kind != "pause" && a.Kind != "resume" {
		return Control{}, errors.New("unknown queue action")
	}
	out := c
	out.PauseOwners = cloneOwners(c.PauseOwners)
	out.Extra = cloneExtra(c.Extra)
	for i, p := range out.PauseOwners {
		if p.Owner == a.Owner {
			if a.Kind == "resume" {
				out.PauseOwners = append(out.PauseOwners[:i], out.PauseOwners[i+1:]...)
			} else {
				out.PauseOwners[i].CancelRunning = p.CancelRunning || a.CancelRunning
			}
			return out, nil
		}
	}
	if a.Kind == "pause" {
		if len(out.PauseOwners) == MaxOwners {
			return Control{}, errors.New("pause owner capacity reached")
		}
		out.PauseOwners = append(out.PauseOwners, PauseOwner{Owner: a.Owner, Reason: a.Reason, CancelRunning: a.CancelRunning})
	}
	return out, nil
}

type Operation struct {
	SchemaVersion int                        `json:"schema_version"`
	ID            string                     `json:"id"`
	Workspace     string                     `json:"workspace"`
	Head          string                     `json:"head,omitempty"`
	Kinds         []string                   `json:"kinds"`
	Phase         string                     `json:"phase"`
	PID           int                        `json:"pid"`
	PIDStarted    string                     `json:"pid_started"`
	Managed       bool                       `json:"managed"`
	StartedAt     time.Time                  `json:"started_at"`
	BootID        string                     `json:"boot_id"`
	Root          string                     `json:"root"`
	Detail        string                     `json:"detail"`
	Extra         map[string]json.RawMessage `json:"-"`
}

func validID(s string) bool {
	if s == "" || len(s) > MaxToken || s == "." || s == ".." {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' || r == ':') {
			return false
		}
	}
	return true
}
func validKind(k string) bool { return k == "suite" || k == "merge" || k == "landing" }
func validateOperation(o Operation) error {
	if o.SchemaVersion != 1 {
		return errors.New("unsupported operation schema_version")
	}
	if !validID(o.ID) || !boundedText(o.Workspace, MaxText) || !boundedText(o.Head, MaxToken) || !boundedText(o.Detail, MaxText) || !boundedText(o.Root, MaxText) || !boundedText(o.BootID, MaxToken) || !boundedText(o.PIDStarted, MaxToken) || o.PID < 0 || o.StartedAt.IsZero() {
		return errors.New("invalid operation identity, timestamp or text bound")
	}
	if o.Managed && (o.PID <= 1 || o.PIDStarted == "") {
		return errors.New("managed operation lacks process incarnation")
	}
	if len(o.Kinds) == 0 || len(o.Kinds) > 3 {
		return errors.New("operation requires distinct kinds")
	}
	seen := map[string]bool{}
	for _, k := range o.Kinds {
		if !validKind(k) || seen[k] {
			return errors.New("invalid or duplicate operation kind")
		}
		seen[k] = true
	}
	switch o.Phase {
	case "waiting", "running", "interrupting", "unknown", "finished", "interrupted":
	default:
		return errors.New("invalid operation phase")
	}
	return nil
}

// TransitionOperation conservatively records a finish after cancellation as
// interrupted. Terminal operations cannot be restarted by replaying an event.
func TransitionOperation(o Operation, event string) (Operation, error) {
	if err := validateOperation(o); err != nil {
		return Operation{}, err
	}
	next := ""
	switch o.Phase {
	case "waiting":
		switch event {
		case "admit":
			next = "running"
		case "cancel":
			next = "interrupted"
		case "unknown":
			next = "unknown"
		}
	case "running":
		switch event {
		case "finish":
			next = "finished"
		case "cancel":
			next = "interrupting"
		case "unknown":
			next = "unknown"
		}
	case "interrupting":
		switch event {
		case "finish", "interrupted":
			next = "interrupted"
		case "cancel":
			next = "interrupting"
		case "unknown":
			next = "unknown"
		}
	case "unknown":
		if event == "unknown" {
			next = "unknown"
		}
	}
	if next == "" {
		return Operation{}, fmt.Errorf("invalid operation transition %s/%s", o.Phase, event)
	}
	out := o
	out.Phase = next
	out.Kinds = append([]string(nil), o.Kinds...)
	out.Extra = cloneExtra(o.Extra)
	return out, nil
}
