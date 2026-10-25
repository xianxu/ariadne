package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/xianxu/ariadne/cmd/internal/jsonstrict"
)

// CouchRecovery supplies the one producer-owned recovery snapshot. It replaces
// the separate SDLC inventory read for resume, not for other fleet commands.
type CouchRecovery interface {
	Recovery(context.Context, string) (RecoverySnapshot, error)
}

type RecoverySnapshot struct{ Slots []RecoverySlot }

// RecoverySlot is a read-only projection. Asleep and Pending are positive
// evidence; Unknown names evidence that prevents automatic recovery.
type RecoverySlot struct {
	FleetSlot
	Asleep, Pending bool
	Unknown         string
}

type recoveryWireRow struct {
	Address, Path string
	Git           *recoveryWireGit
	Agent         *recoveryWireAgent
	Claims        *recoveryWireClaims
}

type recoveryWireGit struct {
	Source, Branch, State, Dirty, Unlanded, Operation, Error string
}

type recoveryWireAgent struct {
	State   string
	Threads *int
	Rows    []struct {
		RepoScope  string `json:"repo_scope"`
		Tag, State string
		AsleepAt   json.RawMessage `json:"asleep_at"`
	}
}

type recoveryWireClaims struct {
	Quality, Error, Active string
	Inactive               []string
	Dependency             []struct{ Ref, Checkout, State string }
}

// decodeRecovery accepts additive unknown fields without rewriting the report.
// The shared strict authority rejects duplicate keys at every depth first.
func decodeRecovery(raw []byte) (RecoverySnapshot, error) {
	var snapshot RecoverySnapshot
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return snapshot, fmt.Errorf("Couch recovery report: %w", err)
	}
	var wire struct {
		SchemaVersion *int `json:"schema_version"`
		Rows          *[]recoveryWireRow
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return snapshot, fmt.Errorf("Couch recovery report: %w", err)
	}
	if wire.SchemaVersion == nil || *wire.SchemaVersion != 1 || wire.Rows == nil {
		return snapshot, fmt.Errorf("Couch recovery report needs schema_version 1 and a rows array; inspect couch --recover-plan-from-sdlc")
	}
	seen := make(map[string]bool, len(*wire.Rows))
	for _, row := range *wire.Rows {
		claim, err := parseHubClaim(row.Address)
		if err != nil || claim.Slot < 0 || seen[row.Address] {
			return RecoverySnapshot{}, fmt.Errorf("Couch recovery report has invalid or duplicate address %q", row.Address)
		}
		seen[row.Address] = true
		sl := RecoverySlot{FleetSlot: FleetSlot{Address: row.Address, Repo: claim.Repo, Slot: claim.Slot, HostPath: row.Path}}
		projectRecovery(&sl, row)
		snapshot.Slots = append(snapshot.Slots, sl)
	}
	sort.Slice(snapshot.Slots, func(i, j int) bool { return snapshot.Slots[i].Address < snapshot.Slots[j].Address })
	return snapshot, nil
}

func projectRecovery(sl *RecoverySlot, row recoveryWireRow) {
	if row.Agent == nil {
		sl.Unknown = "agent evidence missing"
		return
	}
	a := row.Agent
	marked := false
	for _, thread := range a.Rows {
		marked = marked || len(thread.AsleepAt) != 0
	}
	if !marked {
		return
	} // old producer and not asleep are indistinguishable
	if a.State != "parked" || a.Threads == nil || *a.Threads != 1 || len(a.Rows) != 1 {
		sl.Unknown = "asleep mark has ambiguous or non-parked threads"
		return
	}
	thread := a.Rows[0]
	var text string
	if thread.State != "parked" || thread.Tag == "" || thread.RepoScope == "" || json.Unmarshal(thread.AsleepAt, &text) != nil {
		sl.Unknown = "asleep thread evidence invalid"
		return
	}
	when, err := time.Parse(time.RFC3339Nano, text)
	if err != nil || when.IsZero() {
		sl.Unknown = "asleep_at is not a nonzero RFC3339 timestamp"
		return
	}
	sl.Asleep = true
	sl.Pending, sl.Unknown = recoveryPending(row)
}

func recoveryPending(row recoveryWireRow) (bool, string) {
	g, c := row.Git, row.Claims
	if row.Path == "" || g == nil || c == nil || c.Quality != "present" || c.Error != "" || g.Error != "" {
		return false, "pending-work evidence unavailable"
	}
	if (g.Source != "sdlc" && g.Source != "local-probe") || g.Branch == "" || g.Operation != "no" ||
		(g.Dirty != "yes" && g.Dirty != "no") || (g.Unlanded != "yes" && g.Unlanded != "no") {
		return false, "pending-work Git evidence unknown or operation in progress"
	}
	switch g.State {
	case "resting", "terminal-issue", "open-issue", "staged":
	default:
		return false, "pending-work branch evidence unknown"
	}
	pending := c.Active != "" || g.State == "open-issue" || g.State == "staged" || g.Dirty == "yes" || g.Unlanded == "yes"
	for _, dep := range c.Dependency {
		switch dep.State {
		case "active":
			pending = true
		case "resting": // an inactive held claim alone is not pending work
		default:
			return false, "dependency claim evidence unknown or conflicting"
		}
	}
	return pending, ""
}

type execRecovery struct{ bin string }

func (c execRecovery) Recovery(ctx context.Context, cwd string) (RecoverySnapshot, error) {
	// Reuse owned group teardown, two-minute acquisition budget, 2-MiB stdout
	// cap and bounded stderr rather than introducing another process runner.
	raw, err := docRepositoryOutput(ctx, cwd, c.bin, "--recover-plan-from-sdlc")
	if err != nil {
		return RecoverySnapshot{}, fmt.Errorf("Couch recovery unavailable: %w; inspect couch --recover-plan-from-sdlc before retrying", err)
	}
	return decodeRecovery(raw)
}
