package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// --- misinvoke guards (exit 2) — pure, no git/files ---

func TestRunActiveTimeMisinvoke(t *testing.T) {
	cases := []struct {
		name string
		opts activetime.Options
		want string
	}{
		{"no dir", activetime.Options{Issues: []string{"8"}, GitRepo: "/r"}, "no --dir given"},
		{"no issue", activetime.Options{Dirs: []string{"/d"}, GitRepo: "/r"}, "--issue required"},
		{"no git-repo", activetime.Options{Dirs: []string{"/d"}, Issues: []string{"8"}}, "--git-repo required"},
	}
	for _, c := range cases {
		var out, errOut bytes.Buffer
		code := runActiveTime(c.opts, &out, &errOut)
		if code != 2 {
			t.Errorf("%s: exit code = %d, want 2", c.name, code)
		}
		if !strings.Contains(errOut.String(), c.want) {
			t.Errorf("%s: stderr %q missing %q", c.name, errOut.String(), c.want)
		}
	}
}

// --- real-git end-to-end: telemetry gap (exit 3) + measured table (exit 0) ---

func atGitRepo(t *testing.T, commits map[string]string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	repo := testfix.Repo(t)
	gitcmd := func(env []string, args ...string) {
		c := exec.Command("git", append([]string{"-C", repo}, args...)...)
		c.Env = append(os.Environ(), env...)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	for iso, msg := range commits {
		if err := os.WriteFile(filepath.Join(repo, "f"), []byte(iso), 0o644); err != nil {
			t.Fatal(err)
		}
		gitcmd([]string{"GIT_AUTHOR_DATE=" + iso, "GIT_COMMITTER_DATE=" + iso}, "add", "f")
		gitcmd([]string{"GIT_AUTHOR_DATE=" + iso, "GIT_COMMITTER_DATE=" + iso}, "commit", "-q", "-m", msg)
	}
	return repo
}

func TestRunActiveTimeTelemetryGap(t *testing.T) {
	repo := atGitRepo(t, map[string]string{"2026-03-01T12:00:00+00:00": "#8 did work"})
	var out, errOut bytes.Buffer
	code := runActiveTime(activetime.Options{
		Dirs: []string{t.TempDir()}, GitRepo: repo, // empty transcript dir
		SinceISO: "2026-01-01T00:00:00Z", UntilISO: "2026-06-01T00:00:00Z",
		Issues: []string{"8"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	}, &out, &errOut)
	if code != 3 {
		t.Fatalf("exit code = %d, want 3", code)
	}
	if !strings.Contains(errOut.String(), "TELEMETRY UNAVAILABLE") {
		t.Errorf("stderr missing TELEMETRY UNAVAILABLE: %q", errOut.String())
	}
}

func TestRunActiveTimeMeasuredTable(t *testing.T) {
	repo := atGitRepo(t, map[string]string{"2026-03-01T10:45:00+00:00": "#8 work"})
	dir := t.TempDir()
	writeJSONLmain(t, filepath.Join(dir, "s.jsonl"),
		`{"timestamp":"2026-03-01T10:00:00Z","type":"user","message":{"content":"#8 a"}}`,
		`{"timestamp":"2026-03-01T10:30:00Z","type":"user","message":{"content":"#8 b"}}`,
		`{"timestamp":"2026-03-01T11:00:00Z","type":"user","message":{"content":"#8 c"}}`,
	)
	var out, errOut bytes.Buffer
	code := runActiveTime(activetime.Options{
		Dirs: []string{dir}, GitRepo: repo,
		SinceISO: "2026-03-01T00:00:00Z", UntilISO: "2026-03-02T00:00:00Z",
		Issues: []string{"8"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	}, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, errOut.String())
	}
	s := out.String()
	for _, want := range []string{
		"# v3 global-boundary attribution",
		"# per-issue totals",
		"#8:",    // per-issue total line
		"commit", // table header
	} {
		if !strings.Contains(s, want) {
			t.Errorf("stdout missing %q:\n%s", want, s)
		}
	}
}

func TestRunActiveTimeRendersWarnings(t *testing.T) {
	repo := atGitRepo(t, map[string]string{"2026-03-01T10:20:00+00:00": "chore: no refs"})
	dir := t.TempDir()
	writeJSONLmain(t, filepath.Join(dir, "s.jsonl"),
		`{"timestamp":"2026-03-01T10:00:00Z","type":"user","message":{"content":"#8 a"}}`,
		`{"timestamp":"2026-03-01T10:10:00Z","type":"user","message":{"content":"#8 b"}}`,
	)
	var out, errOut bytes.Buffer
	code := runActiveTime(activetime.Options{
		Dirs: []string{dir}, GitRepo: repo,
		SinceISO: "2026-03-01T00:00:00Z", UntilISO: "2026-03-02T00:00:00Z",
		Issues: []string{"8"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	}, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "# attribution warnings") || !strings.Contains(out.String(), "fallback") {
		t.Fatalf("stdout missing attribution warning:\n%s", out.String())
	}
}

func TestActiveTimeCmd_Registered(t *testing.T) {
	cmd := NewActiveTimeCmd()
	for _, flag := range []string{"dir", "git-repo", "since", "until", "issue", "commit-weight", "prefix-commit-weight", "threshold-min", "include-assistant"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("active-time command missing flag: --%s", flag)
		}
	}
}

func writeJSONLmain(t *testing.T, path string, lines ...string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A foreign-ref warning is window-scoped: no minutes, no time range. Rendering its zero time
// as "0000-12-31 16:07" would be noise dressed as data (#190).
func TestFormatAttributionWarningOmitsZeroTimeRange(t *testing.T) {
	got := formatAttributionWarning(activetime.AttributionWarning{
		Issue:  "pair#127",
		Reason: "foreign ref ignored — another repo's issue, not attributable here (×2)",
	})
	if strings.Contains(got, "0000") {
		t.Errorf("zero time rendered: %q", got)
	}
	if !strings.HasPrefix(got, "pair#127 foreign ref ignored") {
		t.Errorf("got %q — want the qualified key unprefixed, then the reason", got)
	}
	// A real segment warning keeps its range.
	seg := formatAttributionWarning(activetime.AttributionWarning{
		Issue: "187", Active: 46.1, Share: 0.77, Reason: "mention fallback",
		Start: time.Date(2026, 7, 29, 10, 25, 0, 0, time.Local),
		End:   time.Date(2026, 7, 29, 12, 36, 0, 0, time.Local),
	})
	if !strings.Contains(seg, "2026-07-29 10:25 → 2026-07-29 12:36") {
		t.Errorf("a segment warning must keep its range: %q", seg)
	}
	if !strings.HasPrefix(seg, "#187 ") {
		t.Errorf("a bare key must still be #-prefixed: %q", seg)
	}
}
