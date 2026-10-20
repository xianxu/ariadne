package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// RepoIssue is the part of one `sdlc state --json` issue row tl-lead reads.
type RepoIssue struct {
	ID       string
	Status   string
	Title    string
	Owner    string // the owner's workspace (repo:N) or worktree; "" when nobody owns it
	OwnerErr string // the owner could not be read: unknown, never unowned
}

func (s execSDLC) RepoIssues(repoRoot string) ([]RepoIssue, error) {
	raw, err := s.run(repoRoot, "state", "--json")
	if err != nil {
		return nil, err
	}
	return decodeRepoIssues(raw)
}

// decodeRepoIssues reads `sdlc state --json`'s issues, ignoring the fields
// and sections tl-lead does not use.
func decodeRepoIssues(raw []byte) ([]RepoIssue, error) {
	var in struct {
		Issues []struct {
			ID, Status, Title string
			Owner             *struct{ Workspace, Worktree string }
			OwnerError        string `json:"owner_error"`
		}
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, fmt.Errorf("sdlc state --json: unreadable reply: %v", err)
	}
	out := make([]RepoIssue, 0, len(in.Issues))
	for _, is := range in.Issues {
		r := RepoIssue{ID: is.ID, Status: is.Status, Title: is.Title, OwnerErr: is.OwnerError}
		if is.Owner != nil {
			r.Owner = is.Owner.Workspace
			if r.Owner == "" {
				r.Owner = is.Owner.Worktree
			}
		}
		out = append(out, r)
	}
	return out, nil
}

// unclaimedIssues keeps a repo's open issues nobody owns, newest (highest id)
// first, at most limit; owner-unreadable ones come back apart, since a failed
// read is never absence.
func unclaimedIssues(issues []RepoIssue, limit int) (open, unknown []RepoIssue) {
	for _, is := range issues {
		if is.Status != "open" {
			continue
		}
		switch {
		case is.OwnerErr != "":
			unknown = append(unknown, is)
		case is.Owner == "":
			open = append(open, is)
		}
	}
	sort.Slice(open, func(i, j int) bool { return open[i].ID > open[j].ID })
	if limit > 0 && len(open) > limit {
		open = open[:limit]
	}
	return open, unknown
}

// claimedRepos are the whole repos a hub claims, in .hub order.
func claimedRepos(h hub) []string {
	var out []string
	seen := map[string]bool{}
	for _, c := range h.Claims {
		if c.Slot < 0 && !seen[c.Repo] {
			seen[c.Repo] = true
			out = append(out, c.Repo)
		}
	}
	return out
}

// teamIssues prints, per repo this hub claims, the open issues nobody owns.
// A repo whose read fails prints `?` and the others still print.
func teamIssues(out io.Writer, s SDLC, vantage string, h hub, limit int) {
	for _, repo := range claimedRepos(h) {
		root, err := s.RepoRoot(vantage, repo)
		var issues []RepoIssue
		if err == nil {
			issues, err = s.RepoIssues(root)
		}
		if err != nil {
			fmt.Fprintf(out, "%s: ? (%v)\n", repo, err)
			continue
		}
		open, unknown := unclaimedIssues(issues, limit)
		fmt.Fprintf(out, "%s: %d unclaimed\n", repo, len(open))
		for _, is := range open {
			fmt.Fprintf(out, "  %s#%s  %s\n", repo, strings.TrimLeft(is.ID, "0"), is.Title)
		}
		for _, is := range unknown {
			fmt.Fprintf(out, "  %s#%s  owner unreadable (%s)  %s\n", repo, strings.TrimLeft(is.ID, "0"), is.OwnerErr, is.Title)
		}
	}
}

func newTeamIssuesCmd(d deps) *cobra.Command {
	limit := 10
	c := &cobra.Command{
		Use:   "issues [--limit N]",
		Short: "New issues nobody has claimed, in each repo this team's .hub claims, newest first",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			self, workspace, err := hereHub(cwd)
			if err != nil {
				return err
			}
			hubs, err := readHubs(workspace)
			if err != nil {
				return err
			}
			for _, h := range hubs {
				if h.Name == self {
					if h.Err != nil {
						return fmt.Errorf("%s/.hub: %w", self, h.Err)
					}
					teamIssues(cmd.OutOrStdout(), d.sdlc, cwd, h, limit)
					return nil
				}
			}
			return fmt.Errorf("%s: no .hub found in %s", self, workspace)
		},
	}
	c.Flags().IntVar(&limit, "limit", limit, "newest unclaimed issues shown per repo (0: all)")
	return c
}
