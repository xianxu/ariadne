// issuelist_json.go — `sdlc issue list --json` (#633): every card of the
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

// IssueListJSON is the --json document, schema_version 1. Its rows are their
// own type, not `state --json`'s IssueState, so a change to state never
// changes this format; TestIssueListJSONFieldSetIsPinned holds v1's keys.
// Within a version fields are only added, never renamed or re-meant.
type IssueListJSON struct {
	SchemaVersion int `json:"schema_version"`
	// Tracker says the repository has an issue tracker; without one the list
	// holds only details files, so closes are not visible.
	Tracker bool      `json:"tracker"`
	Cards   []CardRow `json:"cards"`
}

// CardRow is one issue. A terminal card's Updated is its close date: setters
// leave it alone. A row sdlc could not read has Status "unreadable" and says
// why; it is never dropped, and no filter removes it.
type CardRow struct {
	ID       string   `json:"id"`
	Title    string   `json:"title,omitempty"`
	Status   string   `json:"status"`
	Created  string   `json:"created,omitempty"`
	Updated  string   `json:"updated,omitempty"`
	Kind     string   `json:"kind,omitempty"`
	Severity string   `json:"severity,omitempty"`
	Area     []string `json:"area,omitempty"`
	// ClaimantWorkspace is the claimant's slot (repo:N), which a closed card
	// keeps: the slot that closed it.
	ClaimantWorkspace string `json:"claimant_workspace,omitempty"`
	Unreadable        string `json:"unreadable,omitempty"`
}

// issueListSchema is IssueListJSON's current schema_version.
const issueListSchema = 1

// cardRowOf projects one state row into the format. A details file that
// failed to parse leaves an empty status; it is reported as unreadable.
func cardRowOf(is IssueState) CardRow {
	row := CardRow{ID: is.ID, Title: is.Title, Status: is.Status, Created: is.Created, Updated: is.Updated,
		Kind: is.Kind, Severity: is.Severity, Area: is.Area, Unreadable: is.Unreadable}
	if is.Owner != nil {
		row.ClaimantWorkspace = is.Owner.Workspace
	}
	if row.Status == "" {
		row.Status = "unreadable"
	}
	if row.Status == "unreadable" && row.Unreadable == "" {
		row.Unreadable = "details file " + is.Path + " could not be read"
	}
	return row
}

func runIssueListJSON(ctx context.Context, stdout, stderr io.Writer, issuesDir string, filter issueListFilter) error {
	abs, err := filepath.Abs(issuesDir)
	if err != nil {
		die(stderr, fmt.Sprintf("list issues: %v", err))
	}
	tracked, err := repositoryTracked(ctx, abs)
	if err != nil {
		die(stderr, fmt.Sprintf("list issues: %v", err))
	}
	issues, stale, err := issueStates(ctx, issuesDir, true)
	if err != nil {
		die(stderr, fmt.Sprintf("list issues: %v", err))
	}
	if stale {
		cwarn(stderr, staleTrackerNote)
	}
	doc := IssueListJSON{SchemaVersion: issueListSchema, Tracker: tracked, Cards: []CardRow{}}
	for _, is := range issues {
		row := cardRowOf(is)
		// An unreadable row has no status or classification to filter on;
		// it is kept so a reader can report it.
		if row.Status == "unreadable" || filter.matches(is) {
			doc.Cards = append(doc.Cards, row)
		}
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}
