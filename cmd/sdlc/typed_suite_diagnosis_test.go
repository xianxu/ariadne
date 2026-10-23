package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func TestTypedSuiteSplitRemainsUnverified(t *testing.T) {
	dir := t.TempDir()
	record := func(name, log string, exit int) localRunRecord {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(log), 0600); err != nil {
			t.Fatal(err)
		}
		return localRunRecord{SHA: name, Exit: exit, LogPath: path, LogSHA256: resumeRecordDigest([]byte(log))}
	}
	c := conductor{ctx: context.Background(), base: "base", heads: []string{"head"}, evidence: map[int]localRunRecord{0: record("base", "ok example.org/pkg 0.1s\n", 0), 1: record("head", typedLogFrames(t, "script:tests/route.sh", "", 124), 1)}}
	got := c.assessSplit(landingline.AssessSplit{Lo: 0, Hi: 1})
	if got.Verdict.Kind != landingline.AttributionUnverified || !strings.Contains(got.Verdict.Reason, "typed failure has no Go package impact") {
		t.Fatalf("typed split attribution: %+v", got)
	}
}

func TestTypedSuiteBaselineUsesSharedFacts(t *testing.T) {
	frames := typedLogFrames(t, "script:tests/route.sh", "", 124)
	var recipe strings.Builder
	for _, line := range strings.Split(strings.TrimSpace(frames), "\n") {
		fmt.Fprintf(&recipe, "\t@printf '%%s\\n' '%s'\n", line)
	}
	recipe.WriteString("\t@exit 1\n")
	d, sha, _ := realLockLocalRunRepo(t, recipe.String())
	path := filepath.Join(t.TempDir(), "merged.log")
	if err := os.WriteFile(path, []byte(frames), 0600); err != nil {
		t.Fatal(err)
	}
	failure := &localCheckFailure{Check: localCheck{Name: suiteCheckName, Command: []string{"make", "test"}, TimeoutMinutes: 1, Weight: 2}, Run: localRunRecord{Name: suiteCheckName, SHA: sha, Exit: 1, LogPath: path}, Head: sha}
	err := diagnoseSuiteFailure(context.Background(), failure, sha, d)
	if !strings.Contains(err.Error(), "pre-existing: script:tests/route.sh") || strings.Contains(err.Error(), "(unparsed)") {
		t.Fatalf("typed baseline: %v", err)
	}
}
