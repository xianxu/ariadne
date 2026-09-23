package main

import (
	"fmt"
	"strconv"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
	"go.yaml.in/yaml/v3"
)

// claimDecision changes only reservation metadata on the observed remote record.
func claimDecision(raw []byte, id int, today, started string) ([]byte, error) {
	fm, body, err := issue.Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("remote issue #%d is malformed: %w", id, err)
	}
	var fields map[string]interface{}
	if err := yaml.Unmarshal([]byte(fm), &fields); err != nil {
		return nil, fmt.Errorf("remote issue #%d is malformed: %w", id, err)
	}
	rawID, _ := issue.GetField(fm, "id")
	parsedID, idErr := strconv.Atoi(rawID)
	if idErr != nil || parsedID != id {
		return nil, fmt.Errorf("remote issue #%d has missing or mismatched id", id)
	}
	status, ok := fields["status"].(string)
	if !ok || !vocab.Issue().IsOpen(status) {
		return nil, fmt.Errorf("remote issue #%d is not open (status %q); already-working issues are taken; continue existing work without claiming again", id, status)
	}
	fm = issue.SetField(fm, "status", "working")
	fm = issue.SetField(fm, "updated", today)
	if value, _ := issue.GetField(fm, "started"); value == "" {
		fm = issue.SetField(fm, "started", started)
	}
	return []byte(issue.Compose(fm, body)), nil
}
