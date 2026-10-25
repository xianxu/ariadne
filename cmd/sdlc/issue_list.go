// issue_list.go — `sdlc issue list --json` (#633): every card of the
// repository, terminal ones included, for readers outside sdlc (tl-lead's
// triage stats) that count opened and closed issues per week.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
)

// IssueListJSON is the --json document. A field is added, never renamed or
// re-meant, without a schema_version bump.
type IssueListJSON struct {
	SchemaVersion int `json:"schema_version"`
	// Tracker says the repository has an issue tracker; without one the list
	// holds only details files, so closes are not visible.
	Tracker bool         `json:"tracker"`
	Issues  []IssueState `json:"issues"`
}

// issueListSchema is IssueListJSON's current schema_version.
const issueListSchema = 1

func runIssueListJSON(ctx context.Context, stdout, stderr io.Writer, issuesDir string, filter issueListFilter) error {
	abs, err := filepath.Abs(issuesDir)
	if err != nil {
		return err
	}
	tracked, err := repositoryTracked(ctx, abs)
	if err != nil {
		return fmt.Errorf("list issues: %w", err)
	}
	issues, stale, err := issueStates(ctx, issuesDir, true)
	if err != nil {
		return fmt.Errorf("list issues: %w", err)
	}
	if stale {
		cwarn(stderr, staleTrackerNote)
	}
	doc := IssueListJSON{SchemaVersion: issueListSchema, Tracker: tracked, Issues: []IssueState{}}
	for _, is := range issues {
		// An unreadable row has no status or classification to filter on;
		// it is kept so a reader can report it.
		if is.Status == "unreadable" || filter.matches(is) {
			doc.Issues = append(doc.Issues, is)
		}
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}
