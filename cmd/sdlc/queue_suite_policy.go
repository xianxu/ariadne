package main

import (
	"encoding/json"
	"fmt"
	"iter"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

const suiteOrderingPolicy = "fifo-grant-aging-v1"

type suiteQueueCompatibility struct {
	Incompatible []string
	Granted      bool
	Revoke       []string
}

// One pass shared by grant preflight and text status. The caller establishes
// inspection availability; this cannot make an old waiter change its policy.
func assessSuiteQueueCompatibility(tickets iter.Seq[suiteQueueTicket]) suiteQueueCompatibility {
	var result suiteQueueCompatibility
	seen := map[string]bool{}
	for ticket := range tickets {
		// wire: read-only projection; additive ticket metadata is not rewritten.
		var record struct {
			Repo, Worktree string
			Policy         string               `json:"ordering_policy"`
			Subject        suitePrioritySubject `json:"subject"`
		}
		err := json.Unmarshal(ticket.Record, &record)
		if err != nil || record.Policy != suiteOrderingPolicy {
			result.Incompatible = append(result.Incompatible, fmt.Sprintf("ticket %d (repo %q, worktree %q)", ticket.Number, record.Repo, record.Worktree))
		}
		if !ticket.Granted {
			continue
		}
		result.Granted = true
		common := record.Subject.CommonDir
		if err != nil || !filepath.IsAbs(common) || filepath.Clean(common) != common || len(common) > 4096 || strings.ContainsAny(common, "\x00\r\n") {
			continue
		}
		for _, id := range append([]int{record.Subject.Issue}, record.Subject.Issues...) {
			if id <= 0 {
				continue
			}
			hint := "sdlc landing revoke --issue " + issue.CLIRef(strconv.Itoa(id)) + " --repo " + quoteHelperShell(common) + " --reason 'mixed suite ordering (ariadne#570)'"
			if !seen[hint] {
				result.Revoke = append(result.Revoke, hint)
				seen[hint] = true
			}
		}
	}
	return result
}

func (compatibility suiteQueueCompatibility) detail() string {
	return "legacy/unknown ordering: " + strings.Join(compatibility.Incompatible, "; ")
}
