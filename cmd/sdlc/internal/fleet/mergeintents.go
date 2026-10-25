package fleet

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// InventoryCapabilities is additive. Zero means the producer did not claim
// support; decoding an older inventory must never upgrade its evidence.
type InventoryCapabilities struct {
	MergeIntents int `json:"merge_intents,omitempty"`
}

func (c *InventoryCapabilities) UnmarshalJSON(raw []byte) error {
	var wire struct {
		MergeIntents json.RawMessage `json:"merge_intents"`
	}
	if string(raw) == "null" {
		return errors.New("capabilities must be an object")
	}
	if err := additiveUnmarshal(raw, &wire); err != nil {
		return err
	}
	value := InventoryCapabilities{}
	if wire.MergeIntents != nil {
		if string(wire.MergeIntents) == "null" {
			return errors.New("merge_intents capability must be an integer")
		}
		if err := json.Unmarshal(wire.MergeIntents, &value.MergeIntents); err != nil {
			return err
		}
		if value.MergeIntents != 1 {
			return errors.New("unsupported merge_intents capability")
		}
	}
	*c = value
	return nil
}

// MergeIntentReport is shared by inventory, state and landing status. Identity
// names the physical record, not a process incarnation or permission to act.
type MergeIntentReport struct {
	State     string    `json:"state"`
	Branch    string    `json:"branch,omitempty"`
	Issue     int       `json:"issue,omitempty"`
	Head      string    `json:"head,omitempty"`
	Root      string    `json:"root,omitempty"`
	Workspace string    `json:"workspace,omitempty"`
	PID       int       `json:"pid,omitempty"`
	StartedAt time.Time `json:"started_at,omitzero"`
	Reentry   string    `json:"reentry,omitempty"`
	Note      string    `json:"note,omitempty"`
	Path      string    `json:"path"`
	Identity  string    `json:"identity"`
}

func (r *MergeIntentReport) UnmarshalJSON(raw []byte) error {
	type wire MergeIntentReport
	var value wire
	if err := additiveUnmarshal(raw, &value); err != nil {
		return err
	}
	if err := MergeIntentReport(value).validate(); err != nil {
		return err
	}
	*r = MergeIntentReport(value)
	return nil
}

func (r MergeIntentReport) validate() error {
	if !filepath.IsAbs(r.Identity) || filepath.Clean(r.Identity) != r.Identity || r.Identity != r.Path {
		return errors.New("merge intent requires its canonical physical identity/path")
	}
	if r.State == "unreadable" {
		if r.Note == "" || r.Reentry != "" {
			return errors.New("unreadable intent requires a note and forbids reentry")
		}
		return nil
	}
	if r.State != "live" && r.State != "orphaned" && r.State != "stale" {
		return fmt.Errorf("unknown merge intent state %q", r.State)
	}
	in := landingline.Intent{Version: landingline.IntentVersion, Branch: r.Branch, Issue: r.Issue, Head: r.Head, Root: r.Root, Workspace: r.Workspace, PID: r.PID, StartedAt: r.StartedAt}
	if err := in.Valid(); err != nil {
		return err
	}
	if r.State == "orphaned" {
		if r.Reentry != "sdlc merge --branch "+r.Branch+" --yes" {
			return errors.New("orphaned intent requires its exact validated reentry command")
		}
	} else if r.Reentry != "" {
		return errors.New("only orphaned intents have a reentry command")
	}
	return nil
}

func (r TreeRow) validateMergeIntents(required bool) error {
	if r.MergeIntentsState == "" && !required {
		if len(r.MergeIntents) != 0 || r.MergeIntentsError != "" {
			return errors.New("merge evidence without read quality")
		}
		return nil
	}
	if r.MergeIntents == nil {
		return errors.New("merge_intents must be non-null")
	}
	switch r.MergeIntentsState {
	case "complete":
		if r.MergeIntentsError != "" {
			return errors.New("complete merge read cannot carry an error")
		}
	case "partial", "unreadable":
		if r.MergeIntentsError == "" {
			return errors.New("incomplete merge read requires an error")
		}
		if r.MergeIntentsState == "unreadable" && len(r.MergeIntents) != 0 {
			return errors.New("an unreadable store cannot carry observations; use partial")
		}
	default:
		return errors.New("missing or invalid merge_intents_state")
	}
	for _, report := range r.MergeIntents {
		if report.Root != r.TreePath {
			return errors.New("merge intent root must equal its inventory row")
		}
		if err := report.validate(); err != nil {
			return err
		}
	}
	return nil
}

// Inventory is a read-only observation, not a persisted rewrite authority.
// Ignore future fields while keeping duplicate keys and known types strict.
func additiveUnmarshal(raw []byte, target any) error {
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}
