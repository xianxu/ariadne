package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// weeklyHeadline counts distinct observed landing identities, then places the
// ratio beside its coverage. A source gap makes the covered count zero: we
// cannot certify the numerator for even an otherwise validated landing.
func weeklyHeadline(raw weeklyRepoEvidence, window WeeklyWindow, attribution weeklyAttribution) WeeklyHeadline {
	observed := map[int]bool{}
	for _, e := range raw.Events {
		if e.Kind == landingline.EventLanded && e.Issue > 0 && window.Contains(e.At) {
			observed[e.Issue] = true
		}
	}
	for id, card := range raw.Cards {
		if card.Completion == nil || card.Completion.LandedCommit == "" {
			continue
		}
		if at, ok := raw.RecentMainMerges[card.Completion.LandedCommit]; ok && window.Contains(at) {
			observed[id] = true
		}
	}
	covered := attribution.LandedIssues
	if attribution.AttemptsPerLandedIssue == nil {
		covered = 0
	}
	if covered > len(observed) {
		covered = len(observed)
	}
	counts := map[string]int{}
	for _, reason := range attribution.CoverageUnknown {
		if strings.HasPrefix(reason, "issue ") {
			if i := strings.Index(reason, ": "); i >= 0 {
				reason = reason[i+2:]
			}
		}
		counts[reason]++
	}
	result := WeeklyHeadline{
		AttemptsPerLandedIssue: attribution.AttemptsPerLandedIssue,
		ObservedLandedIssues:   len(observed),
		CoveredLandedIssues:    covered,
		UnknownLandedIssues:    len(observed) - covered,
		TopUnknownReasons:      []WeeklyUnknownReason{},
	}
	for reason, count := range counts {
		result.TopUnknownReasons = append(result.TopUnknownReasons, WeeklyUnknownReason{Reason: reason, Count: count})
	}
	sort.Slice(result.TopUnknownReasons, func(i, j int) bool {
		a, b := result.TopUnknownReasons[i], result.TopUnknownReasons[j]
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		return a.Reason < b.Reason
	})
	if len(result.TopUnknownReasons) > 3 {
		result.TopUnknownReasons = result.TopUnknownReasons[:3]
	}
	value := "n/a"
	if result.AttemptsPerLandedIssue != nil {
		value = fmt.Sprintf("%.2f", *result.AttemptsPerLandedIssue)
	}
	result.Text = fmt.Sprintf("full suites per landed issue = %s over %d of %d observed landed issues (%d unknown)", value, covered, len(observed), result.UnknownLandedIssues)
	if len(result.TopUnknownReasons) > 0 {
		result.Text += "; top reasons: "
		for i, reason := range result.TopUnknownReasons {
			if i > 0 {
				result.Text += ", "
			}
			result.Text += fmt.Sprintf("%s (%d)", reason.Reason, reason.Count)
		}
	}
	return result
}

// weeklyLandingEvidence is a validated landing window and its issue-only Git
// ancestry. Construction, rather than attribution, owns cross-source trust.
type weeklyLandingEvidence struct {
	Issue      int
	PR         int
	Branch     string
	EnqueuedAt time.Time
	LandedAt   time.Time
	BranchSHAs map[string]bool
}

type weeklyUnknownAttempt struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

type weeklyAttribution struct {
	LandedIssues           int                    `json:"landed_issues"`
	AttributedAttempts     int                    `json:"attributed_attempts"`
	AttributedPassed       int                    `json:"attributed_passed"`
	AttributedFailed       int                    `json:"attributed_failed"`
	AttributedRepeats      int                    `json:"attributed_repeats"`
	OutsideWeekAttributed  int                    `json:"outside_week_attributed"`
	InWeekOutsideCohort    int                    `json:"in_week_outside_cohort"`
	PreEnqueue             int                    `json:"pre_enqueue"`
	UnassignableAttempts   int                    `json:"unassignable_attempts"`
	AttemptsPerLandedIssue *float64               `json:"attempts_per_landed_issue"`
	UnknownAttempts        []weeklyUnknownAttempt `json:"unknown_attempts"`
	CoverageUnknown        []string               `json:"coverage_unknown"`
}

// retainedHistoryFloor is the last date on which an issue might have existed
// before a retained source begins. Both the attempt recorder and landing log
// sweep at 30 days; the first retained event and structured attempt also bound
// observable admissions and suites.
func retainedHistoryFloor(raw weeklyRepoEvidence) (time.Time, bool) {
	if raw.CollectedAt.IsZero() {
		return time.Time{}, false
	}
	floor := raw.CollectedAt.Add(-suiteAttemptKeep)
	if eventCutoff := raw.CollectedAt.Add(-landingline.Retention); eventCutoff.After(floor) {
		floor = eventCutoff
	}
	var firstEvent time.Time
	for _, event := range raw.Events {
		if event.At.IsZero() || event.At.After(raw.CollectedAt) {
			continue
		}
		if firstEvent.IsZero() || event.At.Before(firstEvent) {
			firstEvent = event.At
		}
	}
	if firstEvent.IsZero() {
		return time.Time{}, false
	}
	if firstEvent.After(floor) {
		floor = firstEvent
	}
	var firstAttempt time.Time
	for _, attempt := range raw.Attempts {
		if attempt.Check != suiteCheckName || attempt.Legacy || attempt.ID == "" || attempt.FinishedAt.IsZero() || attempt.FinishedAt.After(raw.CollectedAt) {
			continue
		}
		if firstAttempt.IsZero() || attempt.FinishedAt.Before(firstAttempt) {
			firstAttempt = attempt.FinishedAt
		}
	}
	if firstAttempt.IsZero() {
		return time.Time{}, false
	}
	if firstAttempt.After(floor) {
		floor = firstAttempt
	}
	return floor.UTC().Truncate(24 * time.Hour), true
}

// validateLandingEvidence refuses to turn syntactically valid but contradictory
// landing records into a metric. A single bad record can make the cohort partial.
func validateLandingEvidence(raw weeklyRepoEvidence, window WeeklyWindow) ([]weeklyLandingEvidence, []string) {
	bySeq := map[int64][]landingline.Event{}
	byIssue := map[int][]landingline.Event{}
	for _, e := range raw.Events {
		if e.Seq > 0 {
			bySeq[e.Seq] = append(bySeq[e.Seq], e)
		}
		if e.Issue > 0 {
			byIssue[e.Issue] = append(byIssue[e.Issue], e)
		}
	}
	var landings []weeklyLandingEvidence
	var unknown []string
	for id, card := range raw.Cards {
		if card.Completion == nil || card.Completion.LandedCommit == "" {
			continue
		}
		at, ok := raw.RecentMainMerges[card.Completion.LandedCommit]
		if !ok || !window.Contains(at) {
			continue
		}
		found := false
		for _, event := range byIssue[id] {
			if event.Kind == landingline.EventLanded {
				found = true
				break
			}
		}
		if !found {
			unknown = append(unknown, fmt.Sprintf("issue %d: main landing has no landed event", id))
		}
	}
	issues := map[int]int{}
	tickets := map[int64]int{}
	branches := map[string]int{}
	prs := map[int]int{}
	merges := map[string]int{}
	for _, e := range raw.Events {
		if e.Kind != landingline.EventLanded || !window.Contains(e.At) {
			continue
		}
		issues[e.Issue]++
		tickets[e.Seq]++
		branches[e.Branch]++
		prs[e.PR]++
		if card, ok := raw.Cards[e.Issue]; ok && card.Completion != nil {
			merges[card.Completion.LandedCommit]++
		}
	}
	mergedStatus := vocab.Issue().FirstTransitionForEvent("merge")
	for _, e := range raw.Events {
		if e.Kind != landingline.EventLanded || !window.Contains(e.At) {
			continue
		}
		fail := func(why string) {
			unknown = append(unknown, fmt.Sprintf("issue %d ticket %d: %s", e.Issue, e.Seq, why))
		}
		if e.Issue <= 0 || e.Seq <= 0 || e.Branch == "" || e.PR <= 0 || e.At.IsZero() {
			fail("incomplete landed event")
			continue
		}
		merge := ""
		if card, ok := raw.Cards[e.Issue]; ok && card.Completion != nil {
			merge = card.Completion.LandedCommit
		}
		if issues[e.Issue] > 1 || tickets[e.Seq] > 1 || branches[e.Branch] > 1 || prs[e.PR] > 1 || (merge != "" && merges[merge] > 1) {
			fail("conflicting cohort landing identity")
			continue
		}
		var paired int
		var enqueued time.Time
		var terminal int
		conflict := false
		for _, sibling := range bySeq[e.Seq] {
			if sibling.Issue != e.Issue || sibling.Branch != e.Branch || sibling.At.IsZero() {
				conflict = true
			}
			switch sibling.Kind {
			case landingline.EventEnqueue:
				paired++
				enqueued = sibling.At
			case landingline.EventLanded, landingline.EventRefused, landingline.EventResumed:
				terminal++
				if sibling.Kind == landingline.EventLanded && sibling.PR != e.PR {
					conflict = true
				}
			}
		}
		if conflict || paired != 1 || terminal != 1 || enqueued.After(e.At) {
			fail("conflicting or missing ticket lifecycle")
			continue
		}
		// Earlier refused admissions belong to the same integration-tax window.
		priorInvalid := false
		successfulEnqueue := enqueued
		for _, prior := range byIssue[e.Issue] {
			if prior.Kind == landingline.EventEnqueue && prior.Branch == e.Branch && !prior.At.IsZero() && prior.At.Before(successfulEnqueue) {
				if prior.Seq <= 0 {
					priorInvalid = true
					continue
				}
				priorEnqueue, priorRefused, priorTerminal := 0, 0, 0
				for _, sibling := range bySeq[prior.Seq] {
					if sibling.Issue != e.Issue || sibling.Branch != e.Branch || sibling.At.IsZero() || sibling.At.Before(prior.At) {
						priorInvalid = true
					}
					if sibling.Kind == landingline.EventEnqueue {
						priorEnqueue++
					}
					if sibling.Kind == landingline.EventRefused {
						priorRefused++
					}
					if sibling.Kind == landingline.EventRefused || sibling.Kind == landingline.EventLanded || sibling.Kind == landingline.EventResumed {
						priorTerminal++
					}
				}
				if priorEnqueue != 1 || priorRefused != 1 || priorTerminal != 1 {
					priorInvalid = true
				}
				if prior.At.Before(enqueued) {
					enqueued = prior.At
				}
			}
		}
		if priorInvalid {
			fail("conflicting or incomplete prior enqueue")
			continue
		}
		card, ok := raw.Cards[e.Issue]
		if !ok || mergedStatus == nil || card.Status != mergedStatus.To || card.Completion == nil || card.Completion.LandedCommit == "" {
			fail("no done completion binding")
			continue
		}
		if !raw.CollectedAt.IsZero() {
			floor, known := retainedHistoryFloor(raw)
			if !known || card.CreatedAt.IsZero() || !card.CreatedAt.UTC().Truncate(24*time.Hour).After(floor) {
				fail("history does not cover issue creation and all possible admissions")
				continue
			}
		}
		graph, ok := raw.Graphs[card.Completion.LandedCommit]
		if !ok || graph.MergeSHA != card.Completion.LandedCommit || graph.MainParent == "" || graph.IssueParent == "" || graph.PR != e.PR || graph.Branch != e.Branch || !graph.ReachableFromMain || !graph.EvidenceReachable || len(graph.BranchSHAs) == 0 {
			fail("unproved landing merge or branch ancestry")
			continue
		}
		landings = append(landings, weeklyLandingEvidence{Issue: e.Issue, PR: e.PR, Branch: e.Branch, EnqueuedAt: enqueued, LandedAt: e.At, BranchSHAs: graph.BranchSHAs})
	}
	sort.Slice(landings, func(i, j int) bool { return landings[i].Issue < landings[j].Issue })
	sort.Strings(unknown)
	return landings, unknown
}

// attributeWeeklyAttempts counts the landed-in-week cohort, not the separate
// physical attempts finished in the week. Unknown coverage makes the ratio nil.
func attributeWeeklyAttempts(attempts []SuiteAttempt, landings []weeklyLandingEvidence, window WeeklyWindow, coverage []string) weeklyAttribution {
	out := weeklyAttribution{LandedIssues: len(landings), UnknownAttempts: []weeklyUnknownAttempt{}, CoverageUnknown: append([]string{}, coverage...)}
	rows := append([]SuiteAttempt(nil), attempts...)
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].FinishedAt.Equal(rows[j].FinishedAt) {
			return rows[i].ID < rows[j].ID
		}
		return rows[i].FinishedAt.Before(rows[j].FinishedAt)
	})
	ids := map[string]int{}
	for _, a := range rows {
		if a.Check == suiteCheckName && a.ID != "" {
			ids[a.ID]++
		}
	}
	shaCount := map[string]int{}
	for _, a := range rows {
		if a.Check != suiteCheckName {
			continue
		}
		inWeek := window.Contains(suiteAttemptCompletedAt(a))
		if a.Legacy || a.ID == "" || a.FinishedAt.IsZero() || ids[a.ID] != 1 {
			for _, landing := range landings {
				if landing.BranchSHAs[a.SHA] {
					out.CoverageUnknown = append(out.CoverageUnknown, "partial attempt on cohort branch: "+a.SHA)
					break
				}
			}
			if inWeek {
				out.UnassignableAttempts++
				out.UnknownAttempts = append(out.UnknownAttempts, weeklyUnknownAttempt{a.ID, "partial or duplicate attempt"})
			}
			if ids[a.ID] > 1 {
				out.CoverageUnknown = append(out.CoverageUnknown, "duplicate attempt id: "+a.ID)
			}
			continue
		}
		matches := 0
		preEnqueue := false
		for _, landing := range landings {
			if !landing.BranchSHAs[a.SHA] {
				continue
			}
			if a.FinishedAt.Before(landing.EnqueuedAt) {
				preEnqueue = true
				continue
			}
			if a.FinishedAt.After(landing.LandedAt) {
				continue
			}
			matches++
		}
		if matches != 1 {
			if matches > 1 {
				out.CoverageUnknown = append(out.CoverageUnknown, "attempt "+a.ID+": ambiguous branch ancestry")
			}
			if inWeek {
				out.UnassignableAttempts++
				reason := "outside confirmed landing cohort"
				if preEnqueue {
					reason = "pre-enqueue"
					out.PreEnqueue++
				} else {
					out.InWeekOutsideCohort++
				}
				if matches > 1 {
					reason = "ambiguous branch ancestry"
				}
				out.UnknownAttempts = append(out.UnknownAttempts, weeklyUnknownAttempt{a.ID, reason})
			}
			continue
		}
		out.AttributedAttempts++
		if a.Exit == 0 && !a.TimedOut {
			out.AttributedPassed++
		} else {
			out.AttributedFailed++
		}
		if shaCount[a.SHA] > 0 {
			out.AttributedRepeats++
		}
		shaCount[a.SHA]++
		if !inWeek {
			out.OutsideWeekAttributed++
		}
	}
	if out.LandedIssues > 0 && len(out.CoverageUnknown) == 0 {
		ratio := float64(out.AttributedAttempts) / float64(out.LandedIssues)
		out.AttemptsPerLandedIssue = &ratio
	}
	sort.Strings(out.CoverageUnknown)
	return out
}
