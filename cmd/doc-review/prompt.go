package main

import (
	_ "embed"
	"fmt"
	"strings"
	"time"
)

// reviewPrompt is the baked review contract handed to the reviewer agent. It is
// the heart of "instructions baked into the binary": the skill carries none of
// it. {{FILE}} is replaced with the absolute document path.
//
//go:embed review-prompt.md
var reviewPrompt string

// buildPrompt instantiates the baked review prompt for one document.
func buildPrompt(absFile string) string {
	return strings.ReplaceAll(reviewPrompt, "{{FILE}}", absFile)
}

// wrapReport prepends provenance frontmatter + a header to the reviewer's raw
// report, so the sidecar is self-describing (which agent, read-only, what it
// reviews). The body below is the reviewer's own words, unedited.
func wrapReport(docName string, agent AgentCLI, body string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "---\n")
	fmt.Fprintf(&b, "type: memory.data\n")
	fmt.Fprintf(&b, "status: active\n")
	fmt.Fprintf(&b, "reviews: %s\n", docName)
	fmt.Fprintf(&b, "reviewer: %s\n", agent)
	fmt.Fprintf(&b, "review_kind: fresh-context-fact-and-reference (read-only)\n")
	fmt.Fprintf(&b, "generated: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "---\n\n")
	fmt.Fprintf(&b, "# Fresh-Context Review of `%s` — by `%s`\n\n", docName, agent)
	fmt.Fprintf(&b, "Independent read-only review: a second-vendor agent (`%s`) with no conversation "+
		"history checked each factual claim and whether its cited reference supports it. The source "+
		"document was NOT modified. Findings below are the reviewer's, verbatim — **not yet applied**; "+
		"the main agent triages and edits the doc where it agrees.\n\n---\n\n", agent)
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}
