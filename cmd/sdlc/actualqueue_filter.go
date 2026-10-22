package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// potentialIssueQueueAttempts uses the existing selector only as a conservative
// relevance probe. Its errors and warnings are not adopted or returned here:
// immutable membership must precede deciding whether they block this issue.
func potentialIssueQueueAttempts(attempts []SuiteAttempt, engagements []tracker.EngagementWindow, repository string, since, until time.Time) ([]SuiteAttempt, error) {
	seen := map[string]bool{}
	var relevant []SuiteAttempt
	for _, a := range attempts {
		if a.Repo != strings.TrimPrefix(repository, "github.com/") || a.Check != "suite" || a.MainBaseline {
			continue
		}
		if seen[a.ID] {
			return nil, fmt.Errorf("duplicate suite attempt %s", a.ID)
		}
		seen[a.ID] = true
		waits, notes, err := selectOwnedQueueAttempts([]SuiteAttempt{a}, engagements, repository, since, until)
		if err != nil || len(waits) > 0 || len(notes) > 0 {
			relevant = append(relevant, a)
		}
	}
	return relevant, nil
}

func filterIssueQueueAttempts(attempts []SuiteAttempt, memberships map[string]queueHeadMembership, issueID int) ([]SuiteAttempt, error) {
	seen := map[string]bool{}
	var own []SuiteAttempt
	var blockers []string
	for _, a := range attempts {
		if seen[a.ID] {
			return nil, fmt.Errorf("duplicate suite attempt %s", a.ID)
		}
		seen[a.ID] = true
		membership := memberships[a.SHA]
		if !membership.Known {
			blockers = append(blockers, fmt.Sprintf("suite attempt %s SHA %s has unknown issue-head attribution: %s; restore immutable commit evidence and retry sdlc actual --issue %d", a.ID, a.SHA, membership.Reason, issueID))
			continue
		}
		for _, id := range membership.Issues {
			if id == issueID {
				own = append(own, a)
				break
			}
		}
	}
	if len(blockers) > 0 {
		return nil, &queueEvidenceError{Blockers: blockers}
	}
	return own, nil
}
