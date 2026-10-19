package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// SDLC is the seam over the sdlc JSON contracts `tl-lead trains` reads:
// `sdlc fleet inventory --json` (every slot and its branch),
// `sdlc workspace <repo> --json` (a repository's primary checkout), and
// `sdlc issue show N --json` (a car's status, claimant and holding checkouts),
// and `sdlc state --json` (a repository's issues and their owners). The
// inventory's rows also carry each checkout's claims on this machine.
type SDLC interface {
	FleetInventory(vantage string) (Fleet, error)
	RepoRoot(vantage, repo string) (string, error)
	IssueShow(repoPath string, n int) (Observation, error)
	// RepoIssues reads `sdlc state --json` in a repository's checkout: its
	// issues with status and owner (ariadne#456, team issues).
	RepoIssues(repoRoot string) ([]RepoIssue, error)
}

// Fleet is the part of the inventory tl-lead reads.
type Fleet struct {
	Slots []FleetSlot
}

// FleetSlot is one slot; Host is its own checkout.
type FleetSlot struct {
	Address       string // repo:N
	Repo          string
	Slot          int
	RestingBranch string
	HostPath      string
	HostBranch    string
	// Claims are the non-terminal issues claimed by this slot's checkout on
	// this machine, as the inventory places them (sdlc's own claim-to-tree
	// join, bound to the machine identity). ClaimsKnown is false when the
	// inventory could not read them: unknown, never "none".
	Claims      []string
	ClaimsKnown bool
	ClaimsWhy   string
}

// Observation is the part of an issue observation tl-lead reads.
type Observation struct {
	Status   string // card status
	Claimant string // the claimant's workspace (repo:N), "" when unclaimed
	// ClaimantPath is the claimant's checkout holding the issue branch, if any.
	ClaimantPath string
}

const sdlcCallTimeout = 60 * time.Second

type execSDLC struct{ bin string }

func (s execSDLC) run(dir string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sdlcCallTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.bin, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("sdlc %s: timed out after %s", strings.Join(args, " "), sdlcCallTimeout)
		}
		return nil, fmt.Errorf("sdlc %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

func (s execSDLC) FleetInventory(vantage string) (Fleet, error) {
	raw, err := s.run(vantage, "fleet", "inventory", "--json")
	if err != nil {
		return Fleet{}, err
	}
	return decodeFleet(raw)
}

func (s execSDLC) RepoRoot(vantage, repo string) (string, error) {
	raw, err := s.run(vantage, "workspace", repo, "--json")
	if err != nil {
		return "", err
	}
	var in struct {
		SchemaVersion int    `json:"schema_version"`
		PrimaryRoot   string `json:"primary_root"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", fmt.Errorf("sdlc workspace %s: unreadable JSON: %v", repo, err)
	}
	if in.SchemaVersion != 2 || in.PrimaryRoot == "" {
		return "", fmt.Errorf("sdlc workspace %s: schema_version %d with primary_root %q, tl-lead reads 2", repo, in.SchemaVersion, in.PrimaryRoot)
	}
	return in.PrimaryRoot, nil
}

func (s execSDLC) IssueShow(repoPath string, n int) (Observation, error) {
	raw, err := s.run(repoPath, "issue", "show", strconv.Itoa(n), "--json")
	if err != nil {
		return Observation{}, err
	}
	return decodeObservation(raw)
}

// decodeFleet reads only the fields tl-lead needs; sdlc owns the contract's
// strictness, and a version tl-lead wasn't written against is refused.
func decodeFleet(raw []byte) (Fleet, error) {
	var in struct {
		SchemaVersion int `json:"schema_version"`
		Rows          []struct {
			TreePath    string `json:"tree_path"`
			ClaimsState string `json:"claims_state"`
			Claims      []struct {
				Ref    string `json:"ref"`
				Status string `json:"status"`
			} `json:"claims"`
		} `json:"rows"`
		Slots []struct {
			Address       string `json:"address"`
			Repo          string `json:"repo"`
			Slot          int    `json:"slot"`
			RestingBranch string `json:"resting_branch"`
			Members       []struct {
				Role   string `json:"role"`
				Path   string `json:"path"`
				Branch string `json:"branch"`
			} `json:"members"`
		} `json:"slots"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return Fleet{}, fmt.Errorf("sdlc fleet inventory: unreadable JSON: %v", err)
	}
	if in.SchemaVersion != 1 {
		return Fleet{}, fmt.Errorf("sdlc fleet inventory: schema_version %d, tl-lead reads 1", in.SchemaVersion)
	}
	type rowClaims struct {
		refs  []string
		known bool
		state string
	}
	byTree := map[string]rowClaims{}
	for _, r := range in.Rows {
		rc := rowClaims{known: r.ClaimsState == "present" || r.ClaimsState == "absent", state: r.ClaimsState}
		for _, c := range r.Claims {
			switch c.Status {
			case "done", "wontfix", "punt":
				continue
			}
			rc.refs = append(rc.refs, c.Ref)
		}
		byTree[r.TreePath] = rc
	}
	var f Fleet
	for _, s := range in.Slots {
		fs := FleetSlot{Address: s.Address, Repo: s.Repo, Slot: s.Slot, RestingBranch: s.RestingBranch}
		for _, m := range s.Members {
			if m.Role == "host" {
				fs.HostPath, fs.HostBranch = m.Path, m.Branch
			}
		}
		if rc, ok := byTree[fs.HostPath]; ok {
			fs.Claims, fs.ClaimsKnown = rc.refs, rc.known
			if !rc.known {
				fs.ClaimsWhy = "claims " + orNone(rc.state)
			}
		} else {
			fs.ClaimsWhy = "no inventory row for " + orNone(fs.HostPath)
		}
		f.Slots = append(f.Slots, fs)
	}
	return f, nil
}

func decodeObservation(raw []byte) (Observation, error) {
	var in struct {
		SchemaVersion int `json:"schema_version"`
		Card          struct {
			State  string `json:"state"`
			Error  string `json:"error"`
			Status string `json:"status"`
		} `json:"card"`
		Assignment struct {
			Claimant *struct {
				Workspace string `json:"workspace"`
			} `json:"claimant"`
		} `json:"assignment"`
		Workspaces struct {
			Holding []struct {
				Path       string `json:"path"`
				IsClaimant bool   `json:"is_claimant"`
			} `json:"holding"`
		} `json:"workspaces"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return Observation{}, fmt.Errorf("sdlc issue show: unreadable JSON: %v", err)
	}
	if in.SchemaVersion != 1 {
		return Observation{}, fmt.Errorf("sdlc issue show: schema_version %d, tl-lead reads 1", in.SchemaVersion)
	}
	// A card that wasn't read is not an answer (stale is: the last fetch).
	if in.Card.State != "present" && in.Card.State != "stale" {
		return Observation{}, fmt.Errorf("sdlc issue show: card %s: %s", in.Card.State, in.Card.Error)
	}
	o := Observation{Status: in.Card.Status}
	if c := in.Assignment.Claimant; c != nil {
		o.Claimant = c.Workspace
	}
	for _, h := range in.Workspaces.Holding {
		if h.IsClaimant {
			o.ClaimantPath = h.Path
		}
	}
	return o, nil
}
