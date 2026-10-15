package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

type usageFlags struct {
	Since string
	Burn  time.Duration
	Issue int
	JSON  bool
	Brain string
}
type agentGlance struct {
	Agent             string           `json:"agent"`
	LastHourTokens    int64            `json:"last_hour_tokens"`
	BurnTokensPerHour float64          `json:"burn_tokens_per_hour"`
	BurnWindow        usage.Window     `json:"burn_window"`
	Scope             string           `json:"scope"`
	Projection        usage.Projection `json:"projection"`
	Coverage          usage.Coverage   `json:"coverage"`
}
type usageIssuePart struct {
	Role     string `json:"role"`
	Tokens   int64  `json:"tokens"`
	Rounds   int    `json:"rounds"`
	Attempts int    `json:"attempts"`
}
type usageOutput struct {
	SchemaVersion int              `json:"schema_version"`
	CapturedAt    time.Time        `json:"captured_at"`
	AgentGlance   []agentGlance    `json:"agent_glance"`
	Window        usage.Window     `json:"window"`
	Report        usage.Report     `json:"report"`
	Coverage      usage.Coverage   `json:"coverage"`
	Issue         int              `json:"issue,omitempty"`
	IssueParts    []usageIssuePart `json:"issue_parts,omitempty"`
	Actual        *actualUsage     `json:"actual,omitempty"`
	Provenance    string           `json:"provenance"`
	Unbound       *usage.Report    `json:"unbound_fleet_evidence,omitempty"`
	UnboundWindow *usage.Window    `json:"unbound_fleet_window,omitempty"`
}
type actualUsage struct {
	Hours    *float64 `json:"hours"`
	Window   string   `json:"window"`
	Status   string   `json:"status"`
	Warnings []string `json:"warnings"`
}

func NewUsageCmd() *cobra.Command {
	f := usageFlags{}
	cmd := &cobra.Command{Use: "usage", Short: "Report local response tokens, attribution and account burn", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()
		out, e := collectUsage(ctx, f, time.Now().UTC())
		if e != nil {
			return e
		}
		if f.JSON {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(out)
		}
		return renderUsage(cmd.OutOrStdout(), out)
	}}
	cmd.Flags().StringVar(&f.Since, "since", "24h", "Detail window duration or RFC3339 timestamp; omitted with --issue means whole issue")
	cmd.Flags().DurationVar(&f.Burn, "burn-window", 3*time.Hour, "Wall-clock burn window, 1h through 3h")
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "Issue ownership and reviewer token cost")
	cmd.Flags().BoolVar(&f.JSON, "json", false, "Emit schema version 1 JSON")
	cmd.Flags().StringVar(&f.Brain, "brain-dir", "", "Explicit brain root for issue attribution")
	cmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		if cmd.Flags().Changed("since") && f.Since == "" {
			return fmt.Errorf("since must not be empty")
		}
		if f.Burn < time.Hour || f.Burn > 3*time.Hour {
			return fmt.Errorf("burn-window must be 1h through 3h")
		}
		if f.Issue < 0 || cmd.Flags().Changed("issue") && f.Issue == 0 {
			return fmt.Errorf("issue must be positive")
		}
		if f.Issue > 0 && !cmd.Flags().Changed("since") {
			f.Since = ""
		}
		return nil
	}
	var agent, unit string
	var remaining float64
	cal := &cobra.Command{Use: "calibrate", Short: "Record an account meter observation; two points are required", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		path, e := usage.MeterPath()
		if e != nil {
			return e
		}
		m, e := usage.RecordMeter(cmd.Context(), path, usage.Observation{Agent: agent, Unit: unit, Remaining: remaining, At: time.Now().UTC()})
		if e != nil {
			return e
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Recorded %s %g %s (%d observations); two observations and complete token evidence are required for a projection.\n", agent, remaining, unit, len(usage.MeterPoints(m, agent, unit)))
		return nil
	}}
	cal.Flags().StringVar(&agent, "agent", "", "codex or claude")
	cal.Flags().StringVar(&unit, "unit", "", "credits or percent")
	cal.Flags().Float64Var(&remaining, "remaining", 0, "Account meter remaining")
	for _, flag := range []string{"agent", "unit", "remaining"} {
		_ = cal.MarkFlagRequired(flag)
	}
	cmd.AddCommand(cal)
	return cmd
}
func usageSince(raw string, now time.Time) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	if d, e := time.ParseDuration(raw); e == nil {
		if d <= 0 {
			return time.Time{}, fmt.Errorf("since duration must be positive")
		}
		return now.Add(-d), nil
	}
	t, e := time.Parse(time.RFC3339, raw)
	if e != nil || !t.Before(now) {
		return time.Time{}, fmt.Errorf("since must be a positive duration or past RFC3339 timestamp")
	}
	return t, nil
}
func collectUsage(ctx context.Context, f usageFlags, now time.Time) (usageOutput, error) {
	out := usageOutput{SchemaVersion: 1, CapturedAt: now, Issue: f.Issue, Provenance: "reported local response usage; account quota coverage may differ"}
	since, e := usageSince(f.Since, now)
	if e != nil {
		return out, e
	}
	root, e := usageRoot()
	if e != nil {
		return out, e
	}
	roots, repos, e := usageFleetSlots(ctx, root)
	if e != nil {
		return out, e
	}
	var authority usageAuthority
	until := now
	if f.Issue > 0 {
		authority, e = resolveUsageAuthority(ctx, root, f.Issue, now)
		if e != nil {
			return out, e
		}
		if f.Brain != "" {
			brain, err := filepath.Abs(f.Brain)
			if err != nil {
				return out, err
			}
			f.Brain = brain

		}
		until = authority.Until
		if since.IsZero() {
			since = authority.Windows[0].Start
		}
		out.Provenance = authority.Label
	}
	scanSince := since
	if glanceSince := now.Add(-f.Burn); scanSince.IsZero() || glanceSince.Before(scanSince) {
		scanSince = glanceSince
	}
	meterPath, e := usage.MeterPath()
	if e != nil {
		return out, e
	}
	meter, e := usage.LoadMeter(meterPath)
	if e != nil {
		return out, e
	}
	for _, agent := range []string{"claude", "codex"} {
		for _, unit := range []string{"credits", "percent"} {
			points := usage.MeterPoints(meter, agent, unit)
			if len(points) == 2 && now.Sub(points[0].At) <= 7*24*time.Hour && points[0].At.Before(scanSince) {
				scanSince = points[0].At
			}
		}
	}
	discovery, e := transcripts.DiscoverUsageEvidence(ctx, transcripts.DefaultHarnesses(), scanSince)
	if e != nil {
		return out, e
	}
	coverage := usage.Coverage{Counts: map[string]int64{}, Examples: discovery.Examples}
	for k, v := range discovery.Diagnostics {
		if k == "sources" {
			k = "discovered_sources"
		}
		coverage.Counts[k] = int64(v)
	}
	var commonCoverage usage.Coverage
	commonCoverage.Merge(coverage)
	parserCoverage := map[string]*usage.Coverage{"codex": {}, "claude": {}}
	var responses []usage.Response
	for _, s := range discovery.Sources {
		if e = ctx.Err(); e != nil {
			return out, e
		}
		file, err := os.Open(s.File)
		if err != nil {
			coverage.Add("unavailable_sources", s.Harness)
			parserCoverage[s.Harness].Add("unavailable_sources", "")
			continue
		}
		before, beforeErr := file.Stat()
		rows, c, err := usage.Decode(ctx, s.Harness, s.ID, io.LimitReader(file, s.Length))
		stat, statErr := file.Stat()
		file.Close()
		coverage.Merge(c)
		parserCoverage[s.Harness].Merge(c)
		if beforeErr != nil || !usageSourceStable(s.Info, before, stat) {
			coverage.Add("source_mutations", s.Harness)
			parserCoverage[s.Harness].Add("source_mutations", "")
		}
		if statErr != nil || stat.Size() < s.Length {
			coverage.Add("truncated_tails", s.Harness)
			parserCoverage[s.Harness].Add("truncated_tails", "")
		}
		if err != nil {
			if ctx.Err() != nil {
				return out, ctx.Err()
			}
			coverage.Add("unavailable_sources", s.Harness)
			parserCoverage[s.Harness].Add("unavailable_sources", "")
			continue
		}
		if len(responses)+len(rows) > 250000 {
			coverage.Add("bound_exceeded", "response identities")
			commonCoverage.Add("bound_exceeded", "response identities")
			break
		}
		responses = append(responses, rows...)
	}
	responses, c, normalizedCoverage, e := normalizeUsageEvidence(responses)
	if e != nil {
		return out, e
	}
	coverage.Merge(c)
	slots := usageSessionSlots(discovery.Sessions, roots)
	reviewers, reviewCoverage, e := loadUsageReviewers(ctx, reviewRepositories(root, repos), usageReviewerScope{Repository: authority.Repository, Issue: f.Issue})
	if e != nil {
		return out, e
	}
	coverage.Merge(reviewCoverage)
	fileByID := map[string]string{}
	for _, s := range discovery.Sessions {
		fileByID[s.Harness+":"+s.ID] = s.File
	}
	workerResolve := func(r usage.Response) bool { return true }
	if f.Issue > 0 {
		windows, e := usageWorkerWindows(ctx, discovery.Sessions, authority)
		if e != nil {
			return out, e
		}
		workerResolve = func(r usage.Response) bool {
			return usageWindowContains(windows[fileByID[r.Agent+":"+r.SourceID]], r.At)
		}
	}
	reviewerMap := bindUsageReviewers(discovery.Sessions, reviewers, now)
	fleetResolve := func(r usage.Response) usage.Attribution {
		slot := slots[r.Agent+":"+r.SourceID]
		if slot == "" {
			slot = "unknown-slot"
		}
		return usage.Attribution{Slot: slot, Role: "worker", Include: true}
	}
	detailResolve := func(r usage.Response) usage.Attribution {
		return usageDetailAttribution(r, fleetResolve(r), reviewerMap, authority, f.Issue, workerResolve(r))
	}
	if since.After(until) {
		since = until
	}
	out.Window = usage.Window{Since: since, Until: until}
	burn := usage.Window{Since: now.Add(-f.Burn), Until: now}
	out.Report, e = usage.AggregateUsage(responses, out.Window, burn, detailResolve)
	if e != nil {
		return out, e
	}
	coverage.Merge(out.Report.Coverage)
	out.Coverage = coverage
	for _, agent := range []string{"claude", "codex"} {
		only := func(r usage.Response) usage.Attribution { a := fleetResolve(r); a.Include = r.Agent == agent; return a }
		hour, e := usage.AggregateUsage(responses, usage.Window{Since: now.Add(-time.Hour), Until: now}, burn, only)
		if e != nil {
			return out, e
		}
		total, _ := hour.Totals.Total()
		glance := agentGlance{Agent: agent, LastHourTokens: total, BurnTokensPerHour: hour.BurnTokensPerHour, BurnWindow: burn, Scope: "fleet"}
		unit := "credits"
		latest := time.Time{}
		for _, o := range meter.Observations {
			if o.Agent == agent && o.At.After(latest) {
				unit = o.Unit
				latest = o.At
			}
		}
		points := usage.MeterPoints(meter, agent, unit)
		glance.Coverage.Merge(commonCoverage)
		if agentCoverage := parserCoverage[agent]; agentCoverage != nil {
			glance.Coverage.Merge(*agentCoverage)
		}
		glance.Coverage.Merge(normalizedCoverage[agent])
		viewCoverage := hour.Coverage
		var calibrationTokens int64
		if len(points) == 2 && now.Sub(points[0].At) <= 7*24*time.Hour {
			report, e := usage.AggregateUsage(responses, usage.Window{Since: points[0].At, Until: points[1].At}, burn, only)
			if e != nil {
				return out, e
			}
			viewCoverage = report.Coverage
			calibrationTokens, _ = report.Totals.Total()
		}
		glance.Coverage.Merge(viewCoverage)
		glance.Projection = usage.ProjectMeter(meter, agent, unit, now, calibrationTokens, hour.BurnTokensPerHour, glance.Coverage.SpendComplete())
		out.AgentGlance = append(out.AgentGlance, glance)
	}
	if f.Issue > 0 {
		out.IssueParts = usageIssueParts(out.Report)
		unknown := func(r usage.Response) usage.Attribution {
			a := usageDetailAttribution(r, fleetResolve(r), reviewerMap, usageAuthority{}, 0, true)
			a.Include = a.Slot == "unknown-slot" || a.Conflict
			return a
		}
		unbound, e := usage.AggregateUsage(responses, burn, burn, unknown)
		if e != nil {
			return out, e
		}
		out.Unbound = &unbound
		out.UnboundWindow = &burn
		for _, g := range unbound.Groups {
			out.Coverage.Counts["excluded_unbound_responses"] += g.Responses
		}
		actual := computeActual(context.WithValue(ctx, localMeasurementContextKey{}, true), root, f.Brain, fmt.Sprint(f.Issue))
		out.Actual = &actualUsage{Window: actual.Window, Status: actualUsageStatus(actual), Warnings: actual.Warnings}
		if actual.Status == actualMeasured {
			out.Actual.Hours = &actual.Hours
		}
	}
	return out, nil
}
func usageIssueParts(report usage.Report) []usageIssuePart {
	parts := []usageIssuePart{{Role: "worker"}, {Role: "reviewer"}, {Role: "total"}}
	rounds, attempts := map[string]bool{}, map[string]bool{}
	for _, g := range report.Groups {
		if g.Responses == 0 {
			continue
		}
		n, _ := g.Tokens.Total()
		i := 0
		if g.Role == "reviewer" {
			i = 1
			rounds[g.Round] = true
			attempts[g.Attempt] = true
		}
		parts[i].Tokens += n
		parts[2].Tokens += n
	}
	parts[1].Rounds = len(rounds)
	parts[1].Attempts = len(attempts)
	return parts
}
func renderUsage(w io.Writer, out usageOutput) error {
	for _, g := range out.AgentGlance {
		projection := g.Projection.Status
		if g.Projection.HoursRemaining != nil {
			projection = fmt.Sprintf("%.1fh remaining (%s)", *g.Projection.HoursRemaining, g.Projection.Unit)
		}
		fmt.Fprintf(w, "%s: last hour %d tokens; burn %.0f tokens/h (%gh); %s\n", g.Agent, g.LastHourTokens, g.BurnTokensPerHour, g.BurnWindow.Until.Sub(g.BurnWindow.Since).Hours(), projection)
	}
	fmt.Fprintf(w, "\n%s → %s\n%s\n", out.Window.Since.Format(time.RFC3339), out.Window.Until.Format(time.RFC3339), out.Provenance)
	for _, p := range out.IssueParts {
		fmt.Fprintf(w, "%s: %d tokens", p.Role, p.Tokens)
		if p.Role == "reviewer" {
			fmt.Fprintf(w, " (%d rounds, %d attempts)", p.Rounds, p.Attempts)
		}
		fmt.Fprintln(w)
	}
	if out.Unbound != nil {
		n, _ := out.Unbound.Totals.Total()
		fmt.Fprintf(w, "Unbound fleet evidence: %d tokens (excluded from issue totals; %s → %s)\n", n, out.UnboundWindow.Since.Format(time.RFC3339), out.UnboundWindow.Until.Format(time.RFC3339))
	}
	for _, g := range out.Report.Groups {
		n, _ := g.Tokens.Total()
		if g.Responses > 0 {
			fmt.Fprintf(w, "%s %s %s %s: %d tokens (input %d, output %d; cache read %s, write %s; reasoning %s)\n", g.Slot, g.Agent, g.Model, g.Role, n, g.Tokens.Input, g.Tokens.Output, usageDimension(g.Tokens.CacheRead), usageDimension(g.Tokens.CacheWrite), usageDimension(g.Tokens.Reasoning))
		}
	}
	for _, h := range out.Report.Hourly {
		n, _ := h.Tokens.Total()
		fmt.Fprintf(w, "Hour %s (%gh): %d tokens\n", h.At.Format(time.RFC3339), h.DurationHours, n)
	}
	if out.Actual != nil {
		if out.Actual.Hours != nil {
			fmt.Fprintf(w, "Active time: %.2fh (%s)\n", *out.Actual.Hours, out.Actual.Window)
		} else {
			fmt.Fprintf(w, "Active time unavailable: %s\n", out.Actual.Status)
		}
	}
	keys := []string{}
	for k, n := range out.Coverage.Counts {
		if n > 0 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	fmt.Fprintf(w, "Coverage: spend_complete=%t attribution_complete=%t", out.Coverage.SpendComplete(), out.Coverage.Complete())
	for _, k := range keys {
		fmt.Fprintf(w, " %s=%d", k, out.Coverage.Counts[k])
	}
	fmt.Fprintln(w)
	return nil
}

func usageDetailAttribution(r usage.Response, a usage.Attribution, reviewers map[string]usageReviewBinding, authority usageAuthority, issue int, workerEligible bool) usage.Attribution {
	if b, ok := reviewers[r.Agent+":"+r.SourceID]; ok {
		a.Role = "reviewer"
		a.Round = b.Round
		a.Attempt = b.Attempt
		a.Conflict = b.Conflict
		if b.Slot != "" {
			a.Slot = b.Slot
		}
		if issue > 0 {
			a.Include = !b.Conflict && b.Repository == authority.Repository && b.Issue == fmt.Sprintf("%06d", issue) && !r.At.Before(b.Since) && r.At.Before(b.Until)
		}
		return a
	}
	if issue > 0 {
		a.Include = workerEligible
	}
	return a
}

func usageDimension(v *int64) string {
	if v == nil {
		return "unsupported"
	}
	return fmt.Sprint(*v)
}

func actualUsageStatus(a actualResult) string {
	switch a.Status {
	case actualMeasured:
		return "measured"
	case actualTelemetryGap:
		return "telemetry_gap: " + a.Detail
	case actualEmptyWindow:
		return "empty_window"
	case actualNoWindow:
		return "no_window"
	default:
		return "error: " + a.Detail
	}
}

// Reviewer dispatch can run in another issue's owned sessions, including bound
// descendants working across repositories. Keep fleet identities for exclusion
// before selecting an issue's totals, with the current repository first.
func reviewRepositories(root string, repos []string) []string {
	out := []string{root}
	seen := map[string]bool{filepath.Clean(root): true}
	for _, repo := range repos {
		key := filepath.Clean(repo)
		if !seen[key] {
			seen[key] = true
			out = append(out, repo)
		}
	}
	return out
}

func usageSourceStable(pin, before, after os.FileInfo) bool {
	if pin == nil || before == nil || after == nil {
		return false
	}
	if !os.SameFile(pin, before) || !os.SameFile(before, after) || before.Size() < pin.Size() || after.Size() < before.Size() {
		return false
	}
	if before.Size() == pin.Size() && !before.ModTime().Equal(pin.ModTime()) {
		return false
	}
	if after.Size() == before.Size() && !after.ModTime().Equal(before.ModTime()) {
		return false
	}
	return true
}

// normalizeUsageEvidence deduplicates globally within the agent-qualified
// identity namespace and retains each account's own completeness diagnostics.
func normalizeUsageEvidence(in []usage.Response) ([]usage.Response, usage.Coverage, map[string]usage.Coverage, error) {
	buckets := map[string][]usage.Response{}
	for _, r := range in {
		buckets[r.Agent] = append(buckets[r.Agent], r)
	}
	agents := []string{}
	for a := range buckets {
		agents = append(agents, a)
	}
	sort.Strings(agents)
	var out []usage.Response
	var global usage.Coverage
	byAgent := map[string]usage.Coverage{}
	for _, a := range agents {
		rows, c, e := usage.NormalizeResponses(buckets[a])
		if e != nil {
			return nil, global, byAgent, e
		}
		global.Merge(c)
		byAgent[a] = c
		out = append(out, rows...)
	}
	return out, global, byAgent, nil
}
