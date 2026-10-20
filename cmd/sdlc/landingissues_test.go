package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestLandingIssueOnlyMainKeepsReadyHead(t *testing.T) {
	root, server, gh := integrationFixture(t)
	gh.readinessStatus = "CLEAN"
	var tested string
	gh.afterStatus = func(head string) {
		tested = head
		for n := 0; n < 3; n++ {
			integrationAdvance(t, server, fmt.Sprintf("workshop/issues/%06d-filed.md", n+700), fmt.Sprintf("---\nid: %06d\nstatus: open\nupdated: 2026-10-19\n---\n# Filed record\n\n## Log\n", n+700))
		}
	}
	var stderr bytes.Buffer
	if err := runMerge(io.Discard, &stderr, landingFlags()); err != nil {
		t.Fatalf("%v\n%s", err, &stderr)
	}
	if tested == "" || gh.pr.HeadOID != tested || gh.merges != 1 {
		t.Fatalf("ready head changed: tested=%s merged=%s merges=%d\n%s", tested, gh.pr.HeadOID, gh.merges, &stderr)
	}
	if strings.Count(stderr.String(), "attempt 1 of 3") != 1 || strings.Contains(stderr.String(), "attempt 2 of 3") {
		t.Fatalf("issue-only movement consumed integration budget:\n%s", &stderr)
	}
	for n := 0; n < 3; n++ {
		if !gitSucceeds(root, "cat-file", "-e", fmt.Sprintf("upstream/main:workshop/issues/%06d-filed.md", n+700)) {
			t.Fatal("server merge lost a main issue record")
		}
	}
}

func TestLandingIssueOnlyFallbackDoesNotExhaustAttempts(t *testing.T) {
	for _, policy := range []string{"BEHIND", "UNKNOWN"} {
		t.Run(policy, func(t *testing.T) {
			_, server, gh := integrationFixture(t)
			gh.readinessStatus = policy
			published := 0
			gh.afterReadiness = func() {
				if published < 4 {
					published++
					integrationAdvance(t, server, fmt.Sprintf("workshop/plans/%06d-note.md", published+700), "plan\n")
				}
			}
			suites := 0
			gh.afterStatus = func(string) { suites++ }
			var stderr bytes.Buffer
			if err := runMerge(io.Discard, &stderr, landingFlags()); err != nil {
				t.Fatalf("%v\n%s", err, &stderr)
			}
			if published != 4 || gh.merges != 1 || suites != 1 {
				t.Fatalf("published=%d merges=%d suites=%d\n%s", published, gh.merges, suites, &stderr)
			}
			if strings.Contains(stderr.String(), "attempt 2 of 3") {
				t.Fatalf("metadata fallback charged:\n%s", &stderr)
			}
		})
	}
}
