package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// A team is an ops repo whose :0 slot is the TL (ariadne#456, one team per
// repo): its checkout has .hub at the root, the ownership list Couch reads
// (pair couchcore/hub_claim.go). tl-lead reads .hub; it never rewrites one.

// hubClaim is one .hub claim: a whole repo (Slot -1) or one exact slot.
type hubClaim struct {
	Repo string
	Slot int
}

func (c hubClaim) String() string {
	if c.Slot < 0 {
		return c.Repo
	}
	return c.Repo + ":" + strconv.Itoa(c.Slot)
}

// maxHubBytes and maxHubLines are Couch's limits; a larger .hub refuses there.
const (
	maxHubBytes = 64 << 10
	maxHubLines = 1024
)

// parseHubClaim mirrors pair's parseHubClaim: a workspace-relative repo, or
// repo:N with N a canonical non-negative integer. The vectors in
// team_test.go are pair's own (hub_claim_test.go), the contract until Couch
// offers a parse verb to conform against.
func parseHubClaim(text string) (hubClaim, error) {
	c := hubClaim{Repo: text, Slot: -1}
	if at := strings.LastIndexByte(text, ':'); at >= 0 {
		c.Repo = text[:at]
		n, err := strconv.Atoi(text[at+1:])
		if err != nil || n < 0 || strconv.Itoa(n) != text[at+1:] {
			return c, fmt.Errorf("invalid exact slot %q", text)
		}
		c.Slot = n
	}
	p := c.Repo
	if p == "" || filepath.IsAbs(p) || filepath.Clean(p) != p || p == "." || p == ".." || strings.HasPrefix(p, "../") ||
		strings.ContainsAny(p, "\\:!#") || strings.IndexFunc(p, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
		return c, fmt.Errorf("invalid workspace-relative repository %q", p)
	}
	return c, nil
}

// parseHub reads a .hub document: one claim per line; blank lines and #
// comments carry none. A malformed line refuses, naming it, as Couch does.
func parseHub(raw []byte) ([]hubClaim, error) {
	if len(raw) > maxHubBytes {
		return nil, fmt.Errorf(".hub exceeds 64 KiB")
	}
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(lines) > maxHubLines {
		return nil, fmt.Errorf(".hub exceeds %d lines", maxHubLines)
	}
	var out []hubClaim
	for i, line := range lines {
		text := strings.TrimSpace(line)
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		c, err := parseHubClaim(text)
		if err != nil {
			return nil, fmt.Errorf(".hub line %d: %w", i+1, err)
		}
		out = append(out, c)
	}
	return out, nil
}

// hub is one team as tl-lead sees it: an ops checkout with .hub at its root.
type hub struct {
	Name   string // the repo's folder name; its TL is Name:0
	Root   string
	Claims []hubClaim
	Err    error // .hub unreadable or malformed: the team is listed, never resolved
}

func (h hub) address() string { return h.Name + ":0" }

// ownership is what ownerOf can say about one slot.
type ownership struct {
	Owners    []string // hub names at the best rank; one means a sole owner
	Exact     bool     // the best rank is an exact slot claim
	Unclaimed bool     // no hub claims it: Couch gives it to the first hub, which tl-lead cannot see
}

// sole is the slot's owner when exactly one hub holds it.
func (o ownership) sole() (string, bool) {
	if len(o.Owners) == 1 {
		return o.Owners[0], true
	}
	return "", false
}

// ownerOf mirrors Couch's Resolve (pair couchcore/hub_ownership.go): a hub
// owns its own :0; an exact repo:N claim outranks a repo claim; hubs tied at
// the best rank are all returned, since the tie goes to the first hub on
// Couch's command line, an order tl-lead cannot see. A hub whose .hub could
// not be read claims nothing here. Runtime moves (couch --move-slot) are not
// on disk; a caller acting on a slot re-checks with Couch.
func ownerOf(repo string, slot int, hubs []hub) ownership {
	for _, h := range hubs {
		if h.Name == repo && slot == 0 {
			return ownership{Owners: []string{h.Name}, Exact: true}
		}
	}
	best := 0
	var owners []string
	for _, h := range hubs {
		rank := 0
		for _, c := range h.Claims {
			if c.Repo != repo {
				continue
			}
			if c.Slot == -1 && rank < 1 {
				rank = 1
			}
			if c.Slot == slot {
				rank = 2
			}
		}
		switch {
		case rank > best:
			best, owners = rank, []string{h.Name}
		case rank > 0 && rank == best:
			owners = append(owners, h.Name)
		}
	}
	if best == 0 {
		return ownership{Unclaimed: true}
	}
	return ownership{Owners: owners, Exact: best == 2}
}

// readHubs finds the teams in a workspace: each directory directly under it
// with .hub at its root. An unreadable .hub is kept with its error.
func readHubs(workspace string) ([]hub, error) {
	entries, err := os.ReadDir(workspace)
	if err != nil {
		return nil, err
	}
	var out []hub
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		root := filepath.Join(workspace, e.Name())
		raw, err := os.ReadFile(filepath.Join(root, ".hub"))
		if os.IsNotExist(err) {
			continue
		}
		h := hub{Name: e.Name(), Root: root, Err: err}
		if err == nil {
			h.Claims, h.Err = parseHub(raw)
		}
		out = append(out, h)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
