package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewprepare"
)

func declarePreparationForTest(t *testing.T, command []string) {
	t.Helper()
	if err := os.MkdirAll(".sdlc", 0755); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(reviewprepare.Config{Version: 1, Command: command, TimeoutMinutes: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".sdlc/review-prepare.json", raw, 0644); err != nil {
		t.Fatal(err)
	}
	git(t, "", "add", ".sdlc/review-prepare.json")
	git(t, "", "commit", "-qm", "declare preparation")
}

func TestBoundaryPreparationDryRunHasNoProcesses(t *testing.T) {
	p := boundaryWindow(t)
	marker := filepath.Join(t.TempDir(), "executed")
	declarePreparationForTest(t, []string{"touch", marker})
	p.Head = git(t, "", "rev-parse", "HEAD")
	original := judge.Run
	t.Cleanup(func() { judge.Run = original })
	judge.Run = func(context.Context, func(int), string, ...string) (judge.ProcessOutput, error) {
		t.Fatal("dry run invoked reviewer")
		return judge.ProcessOutput{}, nil
	}
	var out, stderr bytes.Buffer
	if err := printBoundaryReviewDryRun(&out, &stderr, p); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("dry run executed preparation: %v", err)
	}
	if !strings.Contains(out.String(), "touch") || !strings.Contains(out.String(), marker) {
		t.Fatalf("dry run omitted preparation: %s %s", out.String(), stderr.String())
	}
}

func TestClosePreparationInvalidDiagnosticsRetention(t *testing.T) {
	issues := closeRepo(t, 69)
	plans := filepath.Join("workshop", "plans")
	if err := os.MkdirAll(plans, 0755); err != nil {
		t.Fatal(err)
	}
	declarePreparationForTest(t, []string{"sh", "-c", "printf preparation-evidence"})
	before := readIssue(t, issues)
	head := git(t, "", "rev-parse", "HEAD")
	original := judge.Run
	t.Cleanup(func() { judge.Run = original })
	calls := 0
	judge.Run = func(context.Context, func(int), string, ...string) (judge.ProcessOutput, error) {
		calls++
		return judge.ProcessOutput{Stdout: []byte(strings.Repeat("noise ", 14000) + "\nVERDICT: REWORK\n```findings\n{}\n```\n")}, nil
	}
	file := filepath.Base(mustIssuePath(t, issues, 69))
	path := sidecarPath(plans, file, "")
	countRE := regexp.MustCompile(`added diagnostic bytes: ([0-9]+)`)
	retainedRE := regexp.MustCompile(`retained-bytes: ([0-9]+)`)
	total := 0
	for n := 1; n <= 4; n++ {
		_, stderr, err := executeSDLCTestCommand("close", "--issue", "69", "--actual", "1", "--verified", "tests pass", "--no-atlas", "--issues-dir", issues, "--plans-dir", plans, "--brain-dir", "../nonexistent-brain")
		if err == nil {
			t.Fatal("invalid review finalized close")
		}
		if calls != 2*n {
			t.Fatalf("dispatch attempts=%d want=%d; stderr=%s", calls, 2*n, stderr)
		}
		if after := readIssue(t, issues); after != before {
			t.Fatal("invalid review changed issue")
		}
		if got := git(t, "", "rev-parse", "HEAD"); got != head {
			t.Fatalf("invalid review committed %s", got)
		}
		if _, err := os.Stat(boundaryGatePath(plans, file)); !os.IsNotExist(err) {
			t.Fatalf("invalid review persisted ledger: %v", err)
		}
		matches := countRE.FindStringSubmatch(stderr)
		if len(matches) != 2 {
			t.Fatalf("missing measured append: %s", stderr)
		}
		count, err := strconv.Atoi(matches[1])
		if err != nil || count <= 0 || count > reviewprepare.DiagnosticLimit {
			t.Fatalf("bad appended count %v %v", matches, err)
		}
		total += count
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		retained := retainedRE.FindAllStringSubmatch(string(data), -1)
		if len(retained) != n {
			t.Fatalf("retained invocations=%d want=%d", len(retained), n)
		}
		if retained[n-1][1] != fmt.Sprintf("%010d", count) {
			t.Fatalf("printed %d differs from retained %s", count, retained[n-1][1])
		}
		start := bytes.LastIndex(data, []byte("\n## Review preparation and protocol diagnostics\n"))
		if start < 0 || len(data)-start != count {
			t.Fatalf("reported %d diagnostic bytes, actual appended envelope=%d", count, len(data)-start)
		}
		if !strings.Contains(string(data), "preparation-evidence") {
			t.Fatal("preparation evidence lost")
		}
	}
	if total > 1<<20 {
		t.Fatalf("four calls retained %d diagnostic bytes", total)
	}
}
