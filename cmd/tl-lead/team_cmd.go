package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// scaffold is what `team new` writes for a team: .hub (claim lines only,
// the one ownership list) and team.md (prose that points at .hub, never a
// second list).
func scaffold(name string, claims []hubClaim) map[string][]byte {
	var hubFile strings.Builder
	for _, c := range claims {
		hubFile.WriteString(c.String() + "\n")
	}
	teamMD := fmt.Sprintf(`---
type: team
---

# %[1]s

The team's TL is %[2]s, the :0 slot of this ops repo. What it supervises is
the claims in [.hub](.hub), the one ownership list (Couch reads it; a bare
repo claims every slot of that repo, repo:N one slot). The steward is the
first hub on Couch's command line; see `+"`tl-lead --help`"+`, TEAMS.

## Asks other teams for

- (what this team needs from other teams, and from the steward)
`, name, name+":0")
	return map[string][]byte{".hub": []byte(hubFile.String()), "team.md": []byte(teamMD)}
}

// writeScaffold puts files into root. A file already holding the same bytes
// counts as written (a rerun after an interruption); one holding other bytes
// refuses before anything is written. Each write is exclusive.
func writeScaffold(root string, files map[string][]byte) (written []string, err error) {
	names := sortedKeys(files)
	var todo []string
	for _, name := range names {
		have, err := os.ReadFile(filepath.Join(root, name))
		switch {
		case err == nil && bytes.Equal(have, files[name]):
			continue
		case err == nil:
			return nil, fmt.Errorf("%s already exists with other content; edit it by hand or remove it, then rerun", filepath.Join(root, name))
		case !errors.Is(err, os.ErrNotExist):
			return nil, err
		}
		todo = append(todo, name)
	}
	for _, name := range todo {
		f, err := os.OpenFile(filepath.Join(root, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return written, err
		}
		_, werr := f.Write(files[name])
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil {
			return written, werr
		}
		written = append(written, name)
	}
	return written, nil
}

// teamNew creates or adopts the ops repo at path. A missing folder is made
// and git-initialised; an existing one must be a git checkout's top level.
// It never commits, pushes or creates a remote: those are the operator's and
// the hub's steps.
func teamNew(out io.Writer, path string, raw []string) error {
	if len(raw) == 0 {
		return errors.New("give at least one --claim (a repo, or repo:N): what this team supervises")
	}
	var claims []hubClaim
	for _, r := range raw {
		c, err := parseHubClaim(strings.TrimSpace(r))
		if err != nil {
			return fmt.Errorf("--claim %q: %w", r, err)
		}
		claims = append(claims, c)
	}
	root, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	switch info, err := os.Stat(root); {
	case errors.Is(err, os.ErrNotExist):
		if err := os.MkdirAll(root, 0o755); err != nil {
			return err
		}
		if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
			return fmt.Errorf("git init %s: %v: %s", root, err, strings.TrimSpace(string(out)))
		}
		fmt.Fprintf(out, "created %s (git init)\n", root)
	case err != nil:
		return err
	case !info.IsDir():
		return fmt.Errorf("%s is not a folder", root)
	default:
		top, err := git(root, "rev-parse", "--show-toplevel")
		if err != nil {
			return fmt.Errorf("%s exists but is not a git checkout: a team is an ops repo", root)
		}
		if real, _ := filepath.EvalSymlinks(root); top != root && top != real {
			return fmt.Errorf("%s is inside the checkout %s: give the repo's top level", root, top)
		}
	}
	name := filepath.Base(root)
	written, err := writeScaffold(root, scaffold(name, claims))
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "team %s (TL %s): wrote %s\n", name, name+":0", orNone(strings.Join(written, ", ")))
	fmt.Fprintf(out, "next (yours, not tl-lead's): fill in team.md's asks. For a new non-brain work repo, run weave link github.com/xianxu/ariadne then weave compile, commit the setup, configure the remote and push main, then finish setup with sdlc issue init. Existing legacy repos use reviewed sdlc issue migrate; brains and declined issue scaffolds skip init. Start Couch with this repo among its hubs (the first hub is the steward)\n")
	return nil
}

// hereHub is the team the command runs in: the git top level of cwd, which
// must hold .hub, and the workspace its siblings live in.
func hereHub(cwd string) (self string, workspace string, err error) {
	top, err := git(cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", "", fmt.Errorf("not in a git checkout: run from a team's ops repo")
	}
	if _, err := os.Stat(filepath.Join(top, ".hub")); err != nil {
		return "", "", fmt.Errorf("%s has no .hub at its root: run from a team's ops repo (a hub)", top)
	}
	return filepath.Base(top), filepath.Dir(top), nil
}

// renderTeams lists every team, or with repo the slot owners of that repo
// as Couch would resolve them from .hub.
func renderTeams(out io.Writer, hubs []hub, repo string) {
	if repo == "" {
		if len(hubs) == 0 {
			fmt.Fprintln(out, "no teams: no repo directly under the workspace has .hub at its root")
			return
		}
		claimed := map[string][]string{}
		for _, h := range hubs {
			if h.Err != nil {
				fmt.Fprintf(out, "%s  ?  .hub unreadable: %v\n", h.address(), h.Err)
				continue
			}
			var cs []string
			for _, c := range h.Claims {
				cs = append(cs, c.String())
				if c.Slot < 0 {
					claimed[c.Repo] = append(claimed[c.Repo], h.Name)
				}
			}
			fmt.Fprintf(out, "%s  %s\n", h.address(), orNone(strings.Join(cs, " ")))
		}
		for _, r := range sortedKeys(claimed) {
			if len(claimed[r]) > 1 {
				fmt.Fprintf(out, "warning: %s is claimed by %s; Couch gives it to whichever is first on its command line\n", r, strings.Join(claimed[r], " and "))
			}
		}
		return
	}
	o := ownerOf(repo, -2, hubs) // -2: no exact slot, the repo-level answer
	switch {
	case o.Unclaimed:
		fmt.Fprintf(out, "%s: unclaimed; Couch gives it to the first hub on its command line (the steward)\n", repo)
	case len(o.Owners) == 1:
		fmt.Fprintf(out, "%s: %s:0\n", repo, o.Owners[0])
	default:
		var as []string
		for _, n := range o.Owners {
			as = append(as, n+":0")
		}
		fmt.Fprintf(out, "%s: tied between %s; Couch gives it to whichever is first on its command line\n", repo, strings.Join(as, ", "))
	}
	for _, h := range hubs {
		for _, c := range h.Claims {
			if c.Repo == repo && c.Slot >= 0 {
				fmt.Fprintf(out, "  %s: %s:0 (an exact slot claim)\n", c, h.Name)
			}
		}
	}
}

func newTeamCmd(d deps) *cobra.Command {
	team := &cobra.Command{
		Use:   "team",
		Short: "Teams: an ops repo whose :0 is the TL, its .hub the ownership list",
	}
	var claims []string
	newCmd := &cobra.Command{
		Use:   "new PATH --claim C [--claim C]...",
		Short: "Create or adopt an ops repo as a team: write .hub (claims) and team.md; never commits or pushes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return teamNew(cmd.OutOrStdout(), args[0], claims)
		},
	}
	newCmd.Flags().StringArrayVar(&claims, "claim", nil, "a repo (every slot) or repo:N (one slot) this team supervises; repeatable")
	list := &cobra.Command{
		Use:   "list [REPO]",
		Short: "Every team in this workspace and its claims; with REPO, who owns it",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			top, err := git(cwd, "rev-parse", "--show-toplevel")
			if err != nil {
				return fmt.Errorf("not in a git checkout: run from a checkout in the workspace")
			}
			workspace := filepath.Dir(top)
			hubs, err := readHubs(workspace)
			if err != nil {
				return err
			}
			if len(hubs) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "workspace %s (the parent of this checkout; run from a :0 checkout, since slot worktrees sit elsewhere)\n", workspace)
			}
			repo := ""
			if len(args) == 1 {
				repo = args[0]
			}
			renderTeams(cmd.OutOrStdout(), hubs, repo)
			return nil
		},
	}
	team.AddCommand(newCmd, list, newTeamIssuesCmd(d))
	return team
}
