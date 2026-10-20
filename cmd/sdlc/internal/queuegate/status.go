package queuegate

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

type Activity struct {
	ID        string `json:"id"`
	Workspace string `json:"workspace"`
	Kind      string `json:"kind"`
	State     string `json:"state"`
	Head      string `json:"head,omitempty"`
}
type Snapshot struct {
	SchemaVersion int          `json:"schema_version"`
	State         string       `json:"state"`
	Complete      bool         `json:"complete"`
	PauseOwners   []PauseOwner `json:"pause_owners"`
	Active        []Activity   `json:"active"`
	Detail        string       `json:"detail"`
}

func unknownSnapshot(detail string) Snapshot {
	detail = safeDiagnostic(detail)
	return Snapshot{SchemaVersion: 1, State: "unknown", PauseOwners: []PauseOwner{}, Active: []Activity{}, Detail: detail}
}

func safeDiagnostic(detail string) string {
	if !utf8.ValidString(detail) {
		return "observation detail contains invalid UTF-8"
	}
	if len(detail) > MaxText {
		detail = detail[:MaxText]
		for !utf8.ValidString(detail) {
			detail = detail[:len(detail)-1]
		}
	}
	return detail
}

// BuildSnapshot never turns corrupt or oversized observations into a complete
// empty registry. Waiting and retired leases are not active work.
func BuildSnapshot(c Control, ops []Operation, diagnostics []string) Snapshot {
	if err := validateControl(c); err != nil {
		return unknownSnapshot(err.Error())
	}
	if len(ops) > MaxOperations {
		return unknownSnapshot("operation capacity exceeded")
	}
	s := Snapshot{SchemaVersion: 1, State: "running", Complete: true, PauseOwners: cloneOwners(c.PauseOwners), Active: []Activity{}}
	if len(s.PauseOwners) > 0 {
		s.State = "paused"
	}
	details := append([]string(nil), diagnostics...)
	seen := map[string]bool{}
	for _, op := range ops {
		if err := validateOperation(op); err != nil {
			details = append(details, "invalid operation: "+err.Error())
			continue
		}
		if seen[op.ID] {
			details = append(details, "duplicate operation identity "+op.ID)
			continue
		}
		seen[op.ID] = true
		if op.Phase == "waiting" || op.Phase == "finished" || op.Phase == "interrupted" {
			continue
		}
		state := op.Phase
		if state == "unknown" || op.Workspace == "" {
			state = "unknown"
			detail := op.Detail
			if detail == "" {
				detail = "uncertain operation " + op.ID
			}
			details = append(details, detail)
		}
		for _, kind := range op.Kinds {
			s.Active = append(s.Active, Activity{ID: op.ID, Workspace: op.Workspace, Kind: kind, State: state, Head: op.Head})
		}
		if len(s.Active) > MaxOperations {
			return unknownSnapshot("active entry capacity exceeded")
		}
	}
	if len(details) > 0 {
		s.Complete = false
		s.Detail = strings.Join(details, "; ")
		if strings.TrimSpace(s.Detail) == "" {
			s.Detail = "observation is incomplete"
		}
		s.Detail = safeDiagnostic(s.Detail)
	}
	if data, err := json.Marshal(s); err != nil || len(data) > MaxJSON {
		return unknownSnapshot("status exceeds 1 MiB output bound")
	}
	return s
}

func ValidateSnapshot(s Snapshot) error {
	if s.SchemaVersion != 1 {
		return errors.New("unsupported snapshot schema_version")
	}
	if err := validateOwners(s.PauseOwners); err != nil {
		return err
	}
	if s.Active == nil || len(s.Active) > MaxOperations || !boundedText(s.Detail, MaxText) {
		return errors.New("invalid active array or detail bound")
	}
	switch s.State {
	case "running":
		if len(s.PauseOwners) != 0 {
			return errors.New("running snapshot has pause owners")
		}
	case "paused":
		if len(s.PauseOwners) == 0 {
			return errors.New("paused snapshot has no owner")
		}
	case "unknown":
		if s.Complete {
			return errors.New("unknown observation cannot be complete")
		}
	default:
		return errors.New("invalid snapshot state")
	}
	if (!s.Complete || s.State == "unknown") && strings.TrimSpace(s.Detail) == "" {
		return errors.New("incomplete snapshot requires detail")
	}
	seen := map[string]bool{}
	for _, a := range s.Active {
		if !validID(a.ID) || !boundedText(a.Workspace, MaxText) || !boundedText(a.Head, MaxToken) || !validKind(a.Kind) {
			return errors.New("invalid activity identity or kind")
		}
		switch a.State {
		case "running", "interrupting":
			if a.Workspace == "" {
				return errors.New("unproven workspace must be unknown")
			}
		case "unknown":
			if s.Complete {
				return errors.New("unknown activity cannot be complete")
			}
		default:
			return errors.New("invalid activity state")
		}
		key := a.ID + "\x00" + a.Kind
		if seen[key] {
			return fmt.Errorf("duplicate activity %s/%s", a.ID, a.Kind)
		}
		seen[key] = true
	}
	if data, err := json.Marshal(s); err != nil || len(data) > MaxJSON {
		return errors.New("status exceeds 1 MiB output bound")
	}
	return nil
}
