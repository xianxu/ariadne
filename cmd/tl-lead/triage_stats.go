// triage_stats.go — `tl-lead triage stats` (#633): per repository, bugs and
// features opened and closed per week, the open-bug count at each week's end,
// and the weekly bug quota for one bug slot, from every card that
// `sdlc issue list --json` reads. Read-only; the weekly triage duty (help.md)
// starts from it.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// RepoCard is one card as the triage reads it. Created and Updated are civil
// dates (YYYY-MM-DD); a terminal card's Updated is its close date, since sdlc's
// setters leave it alone.
type RepoCard struct {
	ID, Status, Kind, Created, Updated string
	Workspace                          string // the claimant's slot (repo:N), "" when none
	Unreadable                         string // why sdlc could not read the card
}

// RepoCards is one repository's listing. Without a tracker, closed issues are
// archived details sdlc does not list, so only open counts are real.
type RepoCards struct {
	Tracker bool
	Stale   bool // the tracker fetch failed: the cards are a checkout's last copy
	Cards   []RepoCard
}

// issueListSchema is the `sdlc issue list --json` schema_version this reader
// understands; another is refused rather than misread.
const issueListSchema = 1

func (s execSDLC) RepoCards(repoRoot string) (RepoCards, error) {
	raw, err := s.run(repoRoot, "issue", "list", "--json")
	if err != nil {
		return RepoCards{}, err
	}
	return decodeRepoCards(raw)
}

// decodeRepoCards reads `sdlc issue list --json`. It is tolerant of fields it
// does not use, which a v1 writer may add, and strict on schema_version:
// another version is refused, naming it, rather than misread.
func decodeRepoCards(raw []byte) (RepoCards, error) {
	var in struct {
		SchemaVersion int  `json:"schema_version"`
		Tracker       bool `json:"tracker"`
		Stale         bool `json:"stale"`
		Cards         []struct {
			ID, Status, Kind, Created, Updated, Unreadable string
			ClaimantWorkspace                              string `json:"claimant_workspace"`
		}
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return RepoCards{}, fmt.Errorf("sdlc issue list --json: unreadable reply: %v", err)
	}
	if in.SchemaVersion != issueListSchema {
		return RepoCards{}, fmt.Errorf("sdlc issue list --json: schema_version %d, this tl-lead reads %d (rebuild one of them)", in.SchemaVersion, issueListSchema)
	}
	out := RepoCards{Tracker: in.Tracker, Stale: in.Stale, Cards: make([]RepoCard, 0, len(in.Cards))}
	for _, c := range in.Cards {
		out.Cards = append(out.Cards, RepoCard{ID: c.ID, Status: c.Status, Kind: c.Kind, Created: c.Created, Updated: c.Updated, Workspace: c.ClaimantWorkspace, Unreadable: c.Unreadable})
	}
	return out, nil
}

// day is a civil date counted in days, so week arithmetic never meets a DST
// change or a zone.
type day int

func parseDay(s string) (day, bool) {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return 0, false
	}
	return day(t.Unix() / 86400), true
}

func dayOf(t time.Time) day {
	y, m, d := t.Date()
	return day(time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix() / 86400)
}

func (d day) String() string { return time.Unix(int64(d)*86400, 0).UTC().Format(time.DateOnly) }

// monday is the Monday starting d's week (1970-01-01, day 0, was a Thursday).
func (d day) monday() day { return d - (d+3)%7 }

// weekRow is one week: [Start, Start+7).
type weekRow struct {
	Start                          day
	BugsOpened, BugsClosed         int
	OpenBugs                       int // open at the week's end
	FeaturesOpened, FeaturesClosed int
}

// cardDays is a readable card's dates: closed is -1 while it is not terminal.
type cardDays struct {
	kind           string
	created, close day
	done           bool
	workspace      string
}

// readable keeps the cards whose dates parse; the rest are counted apart.
func readable(cards []RepoCard) (out []cardDays, unreadable int) {
	model := vocab.Issue()
	for _, c := range cards {
		created, ok := parseDay(c.Created)
		if c.Unreadable != "" || !ok {
			unreadable++
			continue
		}
		cd := cardDays{kind: c.Kind, created: created, close: -1, done: c.Status == "done", workspace: c.Workspace}
		if model.IsTerminal(c.Status) {
			closed, ok := parseDay(c.Updated)
			if !ok {
				unreadable++
				continue
			}
			cd.close = closed
		}
		out = append(out, cd)
	}
	return out, unreadable
}

// weekStats buckets the cards into the weeks ending with today's: opened by
// created, closed by a terminal card's close date, and open at a week's end
// when created before it and not closed before it. A reopened card counts as
// open from its creation, so the open line is approximate across a reopen.
func weekStats(cards []cardDays, today day, weeks int) []weekRow {
	first := today.monday() - day(7*(weeks-1))
	rows := make([]weekRow, weeks)
	for i := range rows {
		rows[i].Start = first + day(7*i)
	}
	week := func(d day) int {
		if d < first {
			return -1
		}
		return int(d-first) / 7
	}
	for _, c := range cards {
		bug := c.kind == "bug"
		if i := week(c.created); i >= 0 && i < weeks {
			if bug {
				rows[i].BugsOpened++
			} else if c.kind == "feature" {
				rows[i].FeaturesOpened++
			}
		}
		if c.close >= 0 {
			if i := week(c.close); i >= 0 && i < weeks {
				if bug {
					rows[i].BugsClosed++
				} else if c.kind == "feature" {
					rows[i].FeaturesClosed++
				}
			}
		}
		if !bug {
			continue
		}
		for i := range rows {
			end := rows[i].Start + 7
			if c.created < end && (c.close < 0 || c.close >= end) {
				rows[i].OpenBugs++
			}
		}
	}
	return rows
}

// velocityWeeks is the window the bug quota is measured over: the four whole
// weeks before the current one.
const velocityWeeks = 4

// slotVelocity is the median of issues closed (done, any kind) per slot per
// week over the velocityWeeks before today's week, counting only slot-weeks
// that closed something, so an idle or newly added slot does not pull it down.
// samples is how many slot-weeks it rests on (0: nothing measured).
func slotVelocity(cards []cardDays, today day) (median float64, samples int) {
	end := today.monday()
	start := end - 7*velocityWeeks
	counts := map[string]int{}
	for _, c := range cards {
		if !c.done || c.workspace == "" || c.close < start || c.close >= end {
			continue
		}
		counts[fmt.Sprintf("%s\x00%d", c.workspace, (c.close-start)/7)]++
	}
	var ns []int
	for _, n := range counts {
		ns = append(ns, n)
	}
	if len(ns) == 0 {
		return 0, 0
	}
	sort.Ints(ns)
	mid := len(ns) / 2
	if len(ns)%2 == 1 {
		return float64(ns[mid]), len(ns)
	}
	return float64(ns[mid-1]+ns[mid]) / 2, len(ns)
}

// bugQuota is the bugs one slot takes for the week: the median slot velocity,
// rounded down, never below one.
func bugQuota(median float64) int {
	if n := int(median); n > 1 {
		return n
	}
	return 1
}

// unclassified counts the readable cards with no kind, open and closed.
func unclassified(cards []cardDays) (open, closed int) {
	for _, c := range cards {
		if c.kind != "" {
			continue
		}
		if c.close < 0 {
			open++
		} else {
			closed++
		}
	}
	return open, closed
}

// renderRepoStats prints one repository's report.
func renderRepoStats(out io.Writer, repo string, rc RepoCards, today day, weeks int) {
	cards, bad := readable(rc.Cards)
	fmt.Fprintf(out, "%s: %d weeks to %s (weeks start Monday)\n", repo, weeks, today)
	if rc.Stale {
		fmt.Fprintf(out, "  stale: the tracker could not be fetched; these are the checkout's last copy of the cards\n")
	}
	if !rc.Tracker {
		fmt.Fprintf(out, "  no tracker: open counts only (closed issues are archived details sdlc does not list)\n")
	}
	fmt.Fprintf(out, "  %-10s  %7s %7s %5s  %9s %9s\n", "week", "bugs +", "bugs -", "open", "feature +", "feature -")
	for _, r := range weekStats(cards, today, weeks) {
		fmt.Fprintf(out, "  %-10s  %7d %7d %5d  %9d %9d\n", r.Start, r.BugsOpened, r.BugsClosed, r.OpenBugs, r.FeaturesOpened, r.FeaturesClosed)
	}
	uo, uc := unclassified(cards)
	line := fmt.Sprintf("  unclassified (no kind): %d open, %d closed", uo, uc)
	if bad > 0 {
		line += fmt.Sprintf("; unreadable: %d", bad)
	}
	fmt.Fprintln(out, line)
	median, samples := slotVelocity(cards, today)
	if samples == 0 {
		fmt.Fprintf(out, "  bug quota: 1 per week (no slot closed an issue in the last %d weeks)\n", velocityWeeks)
		return
	}
	fmt.Fprintf(out, "  bug quota: %d per week (median %s issues closed per slot per week, %d slot-weeks, last %d whole weeks)\n",
		bugQuota(median), strings.TrimSuffix(fmt.Sprintf("%.1f", median), ".0"), samples, velocityWeeks)
}

// triageStats prints the report for each repository this hub claims (or the
// one named). A repository whose read fails prints `?` and the rest continue.
func triageStats(out io.Writer, s SDLC, vantage string, h hub, only string, today day, weeks int) {
	for _, repo := range claimedRepos(h) {
		if only != "" && repo != only {
			continue
		}
		root, err := s.RepoRoot(vantage, repo)
		var rc RepoCards
		if err == nil {
			rc, err = s.RepoCards(root)
		}
		if err != nil {
			fmt.Fprintf(out, "%s: ? (%v)\n", repo, err)
			continue
		}
		renderRepoStats(out, repo, rc, today, weeks)
	}
}

func newTriageCmd(d deps) *cobra.Command {
	c := &cobra.Command{
		Use:   "triage",
		Short: "The weekly backlog triage's read-only numbers (see WEEKLY BACKLOG TRIAGE in tl-lead --help)",
		Args:  cobra.NoArgs,
	}
	c.AddCommand(newTriageStatsCmd(d))
	return c
}

func newTriageStatsCmd(d deps) *cobra.Command {
	weeks := 12
	var only string
	c := &cobra.Command{
		Use:   "stats [--repo R] [--weeks N]",
		Short: "Per repo this team's .hub claims: bugs and features opened and closed per week, open bugs, and the bug quota",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if weeks < 1 || weeks > 52 {
				return fmt.Errorf("--weeks %d: give 1 to 52", weeks)
			}
			cwd, h, err := thisHub()
			if err != nil {
				return err
			}
			if only != "" && !slices.Contains(claimedRepos(h), only) {
				return fmt.Errorf("--repo %s: this team's .hub does not claim it (claims %s)", only, strings.Join(claimedRepos(h), ", "))
			}
			triageStats(cmd.OutOrStdout(), d.sdlc, cwd, h, only, dayOf(d.clock()), weeks)
			return nil
		},
	}
	c.Flags().StringVar(&only, "repo", "", "report only this repository")
	c.Flags().IntVar(&weeks, "weeks", weeks, "weeks shown, ending with the current one (1 to 52)")
	return c
}
