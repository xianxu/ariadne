package bench

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestScanLeakSyntheticNormalizedCrossHarnessFixtures(t *testing.T) {
	// SYNTHETIC fixtures model the normalized values of Claude Read/Bash and
	// Codex exec_command calls. They are not recordings of actual agent runs.
	const fixture = `[
 {"id":"synthetic-claude-read","kind":"read","cwd":"/run/public/work","paths":["/host/source/fix.go"]},
 {"id":"synthetic-claude-bash","kind":"exec","cwd":"/run/public/work","command":"git fetch origin"},
 {"id":"synthetic-codex-exec-command","kind":"exec","cwd":"/run/public/work","command":"cat ../../observer/task.json"},
 {"id":"synthetic-unsupported-mcp","kind":"opaque","unsupported":"unknown external tool schema"}
 ]`
	var events []ToolEvent
	if err := json.Unmarshal([]byte(fixture), &events); err != nil {
		t.Fatal(err)
	}
	got := ScanLeakEvents(leakTestScope(), events)
	if got.Status != "possibly-leaked" || got.Coverage.Events != 4 || got.Coverage.Unknown != 3 {
		t.Fatalf("got %+v", got)
	}
	if got.Findings[0].EventID != "synthetic-claude-read" {
		t.Fatalf("lost provenance: %+v", got)
	}
}

func TestScanLeakBoundsOversizedEvent(t *testing.T) {
	got := ScanLeakEvents(leakTestScope(), []ToolEvent{{Kind: "exec", Command: strings.Repeat("x", 1<<20+1)}})
	if got.Status != "unknown" || got.Coverage.Unknown != 1 || got.Coverage.Complete || len(got.Issues) != 1 {
		t.Fatalf("got %+v", got)
	}
}

func leakTestScope() LeakScope {
	return LeakScope{PreparedRoots: []string{"/run/public/work", "/run/public/hub"}, ProtectedRoots: []string{"/host/source", "/run/observer"}, PrivateOrigin: "/run/public/origin.git", Complete: true}
}

func TestScanLeakEvents(t *testing.T) {
	tests := []struct {
		name    string
		events  []ToolEvent
		status  string
		finding string
		unknown int
		private int
	}{
		{"in repo read", []ToolEvent{{Kind: "read", CWD: "/run/public/work", Paths: []string{"src/main.go"}}}, "no-observed-leak", "", 0, 0},
		{"relative escape", []ToolEvent{{ID: "r1", Kind: "read", CWD: "/run/public/work", Paths: []string{"../../../host/source/fix.go"}}}, "possibly-leaked", "protected-read", 0, 0},
		{"observer", []ToolEvent{{Kind: "read", CWD: "/run/public/work", Paths: []string{"../../observer/acceptance.patch"}}}, "possibly-leaked", "protected-read", 0, 0},
		{"prefix sibling", []ToolEvent{{Kind: "read", Paths: []string{"/run/public/work-secret/fix.go"}}}, "possibly-leaked", "outside-read", 0, 0},
		{"clone", []ToolEvent{{Kind: "exec", Argv: []string{"git", "clone", "/host/source", "copy"}}}, "possibly-leaked", "git-transfer", 0, 0},
		{"fetch alias not proven", []ToolEvent{{Kind: "exec", Argv: []string{"git", "fetch", "origin"}}}, "possibly-leaked", "git-transfer", 0, 0},
		{"private fetch", []ToolEvent{{Kind: "exec", CWD: "/run/public/work", Argv: []string{"git", "fetch", "../origin.git", "main"}}}, "no-observed-leak", "", 0, 1},
		{"fetch CWD", []ToolEvent{{Kind: "exec", CWD: "/run/public/work", Argv: []string{"git", "-C", "/run/public/hub", "fetch", "../origin.git", "main"}}}, "no-observed-leak", "", 0, 1},
		{"clone private still flagged", []ToolEvent{{Kind: "exec", Argv: []string{"git", "clone", "/run/public/origin.git"}}}, "possibly-leaked", "git-transfer", 0, 0},
		{"github", []ToolEvent{{Kind: "exec", Argv: []string{"curl", "https://api.github.com/repos/x/y"}}}, "possibly-leaked", "github-access", 1, 0},
		{"gh", []ToolEvent{{Kind: "exec", Argv: []string{"gh", "api", "repos/x/y"}}}, "possibly-leaked", "github-access", 0, 0},
		{"shell wrapper", []ToolEvent{{Kind: "exec", CWD: "/run/public/work", Argv: []string{"sh", "-c", "cat /host/source/fix.go"}}}, "possibly-leaked", "protected-reference", 1, 0},
		{"shell cannot prove fetch", []ToolEvent{{Kind: "exec", Command: "git fetch /run/public/origin.git"}}, "possibly-leaked", "git-transfer", 1, 0},
		{"read command", []ToolEvent{{Kind: "exec", CWD: "/run/public/work", Argv: []string{"cat", "/host/source/fix.go"}}}, "possibly-leaked", "protected-read", 0, 0},
		{"unknown program", []ToolEvent{{Kind: "exec", Argv: []string{"python3", "-c", "print('hello')"}}}, "unknown", "", 1, 0},
		{"opaque", []ToolEvent{{Kind: "opaque", Unsupported: "unsupported tool"}}, "unknown", "", 1, 0},
		{"malformed read", []ToolEvent{{Kind: "read", Paths: []string{"relative"}}}, "unknown", "", 1, 0},
		{"empty", nil, "unknown", "", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScanLeakEvents(leakTestScope(), tt.events)
			if got.Status != tt.status || got.Coverage.Unknown != tt.unknown || got.Coverage.PrivateFetches != tt.private {
				t.Fatalf("got %+v", got)
			}
			found := false
			for _, f := range got.Findings {
				if f.Kind == tt.finding {
					found = true
				}
			}
			if tt.finding != "" && !found {
				t.Fatalf("missing %s: %+v", tt.finding, got)
			}
		})
	}
}

func TestScanLeakIncompleteCoverage(t *testing.T) {
	scope := leakTestScope()
	scope.Complete = false
	got := ScanLeakEvents(scope, []ToolEvent{{Kind: "read", Paths: []string{"/run/public/work/a"}}})
	if got.Status != "unknown" {
		t.Fatalf("incomplete coverage: %+v", got)
	}
	scope.PreparedRoots = nil
	if got := ScanLeakEvents(scope, []ToolEvent{{Kind: "read", Paths: []string{"/host/source/a"}}}); got.Status != "unknown" {
		t.Fatalf("invalid scope: %+v", got)
	}
}

func TestScanLeakPrivateFetchRejectsExecutionOverrides(t *testing.T) {
	for _, argv := range [][]string{
		{"git", "fetch", "/run/public/origin.git", "--upload-pack=cat /host/source/fix.go"},
		{"git", "-c", "remote.origin.url=/run/public/origin.git", "fetch", "origin"},
	} {
		got := ScanLeakEvents(leakTestScope(), []ToolEvent{{Kind: "exec", Argv: argv}})
		if got.Status == "no-observed-leak" || got.Coverage.PrivateFetches != 0 {
			t.Fatalf("unsafe exception: %+v", got)
		}
	}
}

func TestScanLeakEventsLinearVisits(t *testing.T) {
	for _, n := range []int{100, 200} {
		events := make([]ToolEvent, n)
		for i := range events {
			events[i] = ToolEvent{ID: fmt.Sprint(i), Kind: "read", CWD: "/run/public/work", Paths: []string{"a", "b"}}
		}
		got := ScanLeakEvents(leakTestScope(), events)
		if got.Coverage.Events != n || got.Coverage.Operands != 2*n {
			t.Fatalf("n=%d: %+v", n, got.Coverage)
		}
	}
}

func FuzzScanLeakEvents(f *testing.F) {
	for _, seed := range []string{"", "\x00", "../../observer/a", "sh -c 'git fetch origin'", strings.Repeat("a", 4096)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got := ScanLeakEvents(leakTestScope(), []ToolEvent{{Kind: "opaque", Command: s}, {Kind: "read", CWD: s, Paths: []string{s}}})
		if got.Status == "no-observed-leak" {
			t.Fatalf("opaque input became clean: %+v", got)
		}
	})
}
