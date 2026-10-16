package main

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// landingDeps are the command's seams: the clock and who is asking.
type landingDeps struct {
	now func() time.Time
	// caller is the grantor's workspace (address and slot); ok=false outside
	// any workspace, which is the operator's plain shell.
	caller func() (address string, slot *int, ok bool)
	// couchTree is the slot of the Couch tree this process runs in ($COUCH_TREE).
	couchTree func() (slot *int, ok bool)
	name      func() string
}

func defaultLandingDeps() landingDeps {
	resolve := func(dir string) (workspace.Identity, bool) {
		id, err := workspace.Resolve(execGitRunner{}, dir, "")
		return id, err == nil
	}
	return landingDeps{
		now: func() time.Time { return landingNow() },
		caller: func() (string, *int, bool) {
			id, ok := resolve(".")
			if !ok {
				return "", nil, false
			}
			return workspaceText(id.Address, id.Repo), id.Slot, true
		},
		couchTree: func() (*int, bool) {
			tree := os.Getenv("COUCH_TREE")
			if tree == "" {
				return nil, false
			}
			id, ok := resolve(tree)
			return id.Slot, ok
		},
		name: func() string {
			out, _ := execGitRunner{}.Git("config", "user.name")
			return strings.TrimSpace(string(out))
		},
	}
}

func NewLandingCmd() *cobra.Command {
	return newLandingCmd(defaultLandingDeps())
}

func newLandingCmd(d landingDeps) *cobra.Command {
	var repo string
	cmd := &cobra.Command{
		Use:   "landing",
		Short: "Show a repository's landing line, and grant or revoke fleet-blocking priority (#393)",
		Long:  "Placeholder — replaced by helptext.MustGet(\"landing\") in main.go.",
		Args:  cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLandingStatus(cmd.OutOrStdout(), repo, d)
		},
	}
	cmd.PersistentFlags().StringVar(&repo, "repo", ".", "a checkout of the repository whose line to use")
	status := &cobra.Command{
		Use: "status", Short: "Show the landing now, the waiting merges in order, live grants and the delay p50/p90",
		Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error { return runLandingStatus(cmd.OutOrStdout(), repo, d) },
	}
	var issueN int
	var reason string
	var hours float64
	grant := &cobra.Command{
		Use: "grant", Short: "Give an issue fleet-blocking priority (operator or :0 only)",
		Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLandingGrant(cmd.OutOrStdout(), repo, issueN, reason, hours, false, d)
		},
	}
	grant.Flags().IntVar(&issueN, "issue", 0, "the fleet-blocking issue's number")
	grant.Flags().StringVar(&reason, "reason", "", "which shared fleet workflow it blocks")
	grant.Flags().Float64Var(&hours, "hours", 4, "how long the grant lasts, at most 24")
	// revoke has its own flag variables: a shared one would take revoke's
	// default into grant.
	var revokeIssue int
	var revokeReason string
	revoke := &cobra.Command{
		Use: "revoke", Short: "End an issue's fleet-blocking priority (operator or :0 only)",
		Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLandingGrant(cmd.OutOrStdout(), repo, revokeIssue, revokeReason, 0, true, d)
		},
	}
	revoke.Flags().IntVar(&revokeIssue, "issue", 0, "the issue whose priority ends")
	revoke.Flags().StringVar(&revokeReason, "reason", "revoked", "why it ends")
	cmd.AddCommand(status, grant, revoke)
	return cmd
}

func openRepoLine(repo string) (*landingline.Line, error) {
	dir, err := landingLineDir(execGitRunner{}, repo)
	if err != nil {
		return nil, fmt.Errorf("--repo %s: not a git checkout: %w", repo, err)
	}
	return openLandingLine(dir, io.Discard)
}

func runLandingGrant(out io.Writer, repo string, issueN int, reason string, hours float64, revoke bool, d landingDeps) error {
	if issueN <= 0 {
		return errors.New("--issue N is required")
	}
	if !revoke && strings.TrimSpace(reason) == "" {
		return errors.New("--reason is required: name the shared fleet workflow the issue blocks")
	}
	if !revoke && (hours <= 0 || hours > landingline.MaxGrant.Hours()) {
		return fmt.Errorf("--hours must be in (0, %g]; renew a longer need", landingline.MaxGrant.Hours())
	}
	address, slot, ok := d.caller()
	if !ok {
		address = "operator shell"
	}
	tree, _ := d.couchTree()
	if err := landingline.Authorize(slot, tree); err != nil {
		return err
	}
	line, err := openRepoLine(repo)
	if err != nil {
		return err
	}
	now := d.now()
	g := landingline.Grant{Issue: issueN, Grantor: landingline.Grantor{Address: address, Slot: slot, Name: d.name()}, Reason: reason, Granted: now, Revoked: revoke}
	if !revoke {
		g.Expires = now.Add(time.Duration(hours * float64(time.Hour)))
	}
	if err := line.AddGrant(g); err != nil {
		return err
	}
	if revoke {
		_, err = fmt.Fprintf(out, "#%d: fleet-blocking priority revoked by %s\n", issueN, address)
	} else {
		_, err = fmt.Fprintf(out, "#%d: fleet-blocking until %s, granted by %s: %s\n", issueN, g.Expires.Format(time.RFC3339), address, reason)
	}
	return err
}

func runLandingStatus(out io.Writer, repo string, d landingDeps) error {
	line, err := openRepoLine(repo)
	if err != nil {
		return err
	}
	head, waiting, err := line.Status()
	if err != nil {
		return err
	}
	entries, skippedGrants, err := line.GrantEntries()
	if err != nil {
		return err
	}
	grants := landingline.ValidGrants(entries, d.now())
	events, skippedEvents, err := line.Events()
	if err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "landing now: %s\n", describeTicket(head))
	for i, t := range waiting {
		class := "normal"
		if _, ok := grants[t.Issue]; ok {
			class = "fleet-blocking"
		}
		fmt.Fprintf(&b, "  %d. %s [%s] waiting since %s\n", i+1, describeTicket(&t), class, t.Enqueued.Format(time.RFC3339))
	}
	for _, issueN := range slices.Sorted(maps.Keys(grants)) {
		g := grants[issueN]
		fmt.Fprintf(&b, "grant: #%d fleet-blocking until %s, by %s: %s\n", issueN, g.Expires.Format(time.RFC3339), g.Grantor.Address, g.Reason)
	}
	if q := suiteQuantiles(landingline.Delays(events)); q != nil {
		fmt.Fprintf(&b, "ready-to-landed delay: p50 %.0fs, p90 %.0fs (n=%d, last 30 days)\n", q.P50, q.P90, q.Samples)
	} else {
		b.WriteString("ready-to-landed delay: no landings recorded\n")
	}
	if skippedGrants+skippedEvents > 0 {
		fmt.Fprintf(&b, "warning: skipped %d malformed log line(s)\n", skippedGrants+skippedEvents)
	}
	_, err = io.WriteString(out, b.String())
	return err
}
