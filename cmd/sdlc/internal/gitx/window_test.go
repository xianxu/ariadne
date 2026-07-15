package gitx

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestCommitWindow_ExtendedCapIncludes45Days pins the #68 cap bump (31→61): a
// commit ~45 days old must still anchor the window. Under the old 31-day cap it
// would be excluded (empty window) and the issue's actuals would come up short.
// CommitWindow reads the cwd via direct git, so we build a throwaway repo and
// chdir into it.
func TestCommitWindow_ExtendedCapIncludes45Days(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "t@t")
	run("config", "user.name", "t")
	run("config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "f")
	// Author/committer date 45 days ago: inside the 61-day cap, outside the old 31.
	dated := time.Now().AddDate(0, 0, -45).Format("2006-01-02T15:04:05")
	cmd := exec.Command("git", "commit", "-q", "-m", "#99 M1: the work")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+dated, "GIT_COMMITTER_DATE="+dated)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}

	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	sha, _, _, err := CommitWindow("99")
	if err != nil {
		t.Fatalf("CommitWindow: %v", err)
	}
	if sha == "" {
		t.Errorf("45-day-old #99 commit should be within the %d-day cap, but the window was empty", WindowCapDays)
	}
}

// TestIssueRefRE_DiscoveryParsing exercises the regex used by
// DiscoverWindowIssues, in isolation from git.
func TestIssueRefRE_DiscoveryParsing(t *testing.T) {
	tests := []struct {
		subject string
		want    []string
	}{
		{"#15: subject text", []string{"15"}},
		{"close #31 M4: foo", []string{"31"}},
		{"chore: bump (refs #1, #2, #3)", []string{"1", "2", "3"}},
		{"no issue ref here", nil},
		{"#42abc not a real ref", nil}, // word boundary blocks #42 from "42abc"
		{"prefix#42 suffix", []string{"42"}},
		{"#1 and #11 distinct", []string{"1", "11"}},
	}
	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			matches := issueRefRE.FindAllStringSubmatch(tt.subject, -1)
			var got []string
			for _, m := range matches {
				got = append(got, m[1])
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
}

// TestWorkingTransitionISO pins the #113 anchor: WorkingTransitionISO returns
// the EARLIEST `status: working` flip (the claim) and ignores the status:open
// creation, so a DESIGN commit made between the claim and the first `#N` code
// commit falls inside the window [claim, firstWork] — the attention the old
// parent-of-first-#N anchor cut off. Builds a real throwaway repo (the
// established gitx pattern) since the risk lives in the `git log -G` behavior.
func TestWorkingTransitionISO(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	git("config", "user.email", "t@t")
	git("config", "user.name", "t")
	git("config", "commit.gpgsign", "false")

	issueFile := "000113-foo.md"
	writeStatus := func(status string) {
		if err := os.WriteFile(filepath.Join(dir, issueFile),
			[]byte("---\nid: 000113\nstatus: "+status+"\n---\n# Foo\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	commit := func(daysAgo int, addPath, msg string) {
		git("add", addPath)
		d := time.Now().AddDate(0, 0, -daysAgo).Format("2006-01-02T15:04:05")
		cmd := exec.Command("git", "commit", "-q", "-m", msg)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+d, "GIT_COMMITTER_DATE="+d)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit %q: %v\n%s", msg, err, out)
		}
	}
	lastISO := func(pathspec ...string) string {
		return git(append([]string{"log", "-1", "--format=%aI", "--"}, pathspec...)...)
	}

	// open (create) → working (claim) → design (non-#N) → #N work. The two
	// bookkeeping commits carry non-#N "issue-sync" subjects, exactly like a
	// real `sdlc claim`.
	writeStatus("open")
	commit(4, issueFile, "issue-sync: create #113")
	writeStatus("working")
	commit(3, issueFile, "issue-sync: update issues")
	claimISO := lastISO(issueFile) // the working-flip commit

	if err := os.WriteFile(filepath.Join(dir, "plan.md"), []byte("design\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	commit(2, "plan.md", "design: spec + plan (no #N yet)")
	designISO := lastISO()

	if err := os.WriteFile(filepath.Join(dir, "code.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	commit(1, "code.go", "#113: implement")
	workISO := lastISO()

	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	got, ok := WorkingTransitionISO(issueFile)
	if !ok {
		t.Fatal("expected a working-transition commit, got ok=false")
	}
	if got != claimISO {
		t.Errorf("WorkingTransitionISO = %q, want the claim (working-flip) commit %q", got, claimISO)
	}
	// The design commit sits strictly between the claim and the first #N commit,
	// so it lands in the window [claim, firstWork] — the #113 win.
	if !(got <= designISO && designISO <= workISO) {
		t.Errorf("design commit %q should fall in window [%q, %q]", designISO, got, workISO)
	}

	// An issue file that never flipped to working has no anchor.
	writeStatus("open")
	if err := os.WriteFile(filepath.Join(dir, "999-never.md"),
		[]byte("---\nid: 999\nstatus: open\n---\n# N\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	commit(1, "999-never.md", "issue-sync: create #999 (stays open)")
	if iso, ok := WorkingTransitionISO("999-never.md"); ok {
		t.Errorf("never-working file should have no anchor, got %q", iso)
	}
}

// TestIsShippedWorkSubject exercises the #76 pure classifier: subject-anchored
// on #N, minus the bookkeeping denylist. This is the discriminator that keeps a
// bare `--grep #N` count from flagging a filing/claim/close commit as shipped
// implementation work.
func TestIsShippedWorkSubject(t *testing.T) {
	tests := []struct {
		issue   string
		subject string
		want    bool
	}{
		{"76", "#76: surface close-off candidates", true},
		{"76", "tools: #76 surface close-off candidates", true},
		{"76", "#76: file issue — sdlc state surfaces drift", false}, // bookkeeping
		{"76", "sdlc: #76 close issue", false},                       // bookkeeping with area prefix
		{"51", "#51: ticket — in-place branch workflow", false},      // bookkeeping
		{"51", "#51: close (done, actual 4h) + archive", false},      // bookkeeping
		{"51", "#51 M1-M3: in-place branch becomes default", true},   // milestone work
		{"80", "#80: archive stages only moved issue paths", true},   // 'archive' is work, not denylisted
		{"82", "#82 M1: issue new auto-syncs the new file", true},    // 'issue new' ≠ 'issue-sync'
		{"76", "#76: close-off candidate surfacing", true},           // whole-token: 'close-off' ≠ 'close'
		{"51", "#54: push.md (dogfood of #51)", false},               // not subject-anchored to 51
		{"51", "docs: mention #51 in a note", false},                 // loose mention, not the subject's issue
		{"51", "#510: a different issue entirely", false},            // #510 must not match #51
		{"51", "issue-sync: update issues", false},                   // never anchors #N
	}
	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			if got := IsShippedWorkSubject(tt.issue, tt.subject); got != tt.want {
				t.Errorf("IsShippedWorkSubject(%q, %q) = %v, want %v", tt.issue, tt.subject, got, tt.want)
			}
		})
	}
}

// TestIssueSubjectDescriptor_WindowOwnership pins the matcher CommitWindow uses
// to decide which commits belong to an issue's active-time window (#134). It is
// the allowClosePrefix=true entry point — distinct from IsShippedWorkSubject's
// probe: the window COUNTS bookkeeping/close commits (they carry real
// design/close minutes), so there is no denylist here. The contract under test:
// accept the canonical `#N …` and the documented `<area>: #N …`; reject a loose
// reference anywhere but the anchor (`docs: mention #N`, `see #N`) and a longer
// number (`#1340` ≠ `#134`).
func TestIssueSubjectDescriptor_WindowOwnership(t *testing.T) {
	tests := []struct {
		issue   string
		subject string
		owned   bool
	}{
		{"134", "#134: make actuals robust", true},            // canonical
		{"134", "#134 M2: estimator source", true},            // canonical + milestone
		{"134", "sdlc: #134 measure codex transcripts", true}, // <area>: #N
		{"134", "side-quest: #134 robustness pass", true},     // hyphenated area
		{"134", "close #134: archive", true},                  // window counts close (allowClosePrefix)
		{"134", "#134: close (done, actual 3.9h)", true},      // close commit IS in-window (no denylist)
		{"134", "docs: mention #134 in a note", false},        // loose ref after colon, not anchored
		{"134", "see #134 for context", false},                // not subject-anchored
		{"134", "#1340: a different issue entirely", false},   // #1340 must not match #134
		{"134", "issue-sync: update issues", false},           // never anchors #N
	}
	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			_, ok := issueSubjectDescriptor(tt.issue, tt.subject, true)
			if ok != tt.owned {
				t.Errorf("issueSubjectDescriptor(%q, %q, true) owned = %v, want %v", tt.issue, tt.subject, ok, tt.owned)
			}
		})
	}
}

// TestShippedWorkOnMain wires the classifier to the git-log probe via the
// package `run` shim — pinning that the pure helper is actually invoked on real
// log output (so it can't be silently un-wired, lessons.md #72), that the first
// *work* commit wins over a leading bookkeeping commit, and that a missing main
// ref / filing-only history degrade to not-shipped.
func TestShippedWorkOnMain(t *testing.T) {
	orig := run
	defer func() { run = orig }()

	// fakeRun dispatches on argv: MainRef's rev-parse + the log scan.
	fakeRun := func(originMainOK bool, logOut string) func(string, ...string) ([]byte, error) {
		return func(name string, args ...string) ([]byte, error) {
			if len(args) >= 2 && args[0] == "rev-parse" && args[1] == "--verify" {
				ref := args[len(args)-1]
				if (ref == "origin/main" && originMainOK) || ref == "main" {
					return []byte("deadbeef\n"), nil
				}
				return nil, exec.ErrNotFound
			}
			if args[0] == "log" {
				return []byte(logOut), nil
			}
			return nil, exec.ErrNotFound
		}
	}

	t.Run("first work commit wins over leading bookkeeping", func(t *testing.T) {
		run = fakeRun(true, "sha1\x00#76: file issue — drift\nsha2\x00#76: surface close-off\n")
		sha, subj, ok := ShippedWorkOnMain("76")
		if !ok || sha != "sha2" || subj != "#76: surface close-off" {
			t.Errorf("got (%q,%q,%v), want (sha2, #76: surface close-off, true)", sha, subj, ok)
		}
	})

	t.Run("filing-only history → not shipped", func(t *testing.T) {
		run = fakeRun(true, "sha1\x00#76: file issue — drift\n")
		if _, _, ok := ShippedWorkOnMain("76"); ok {
			t.Error("filing-only history should not count as shipped")
		}
	})

	t.Run("no main ref → not shipped (degrades)", func(t *testing.T) {
		run = func(name string, args ...string) ([]byte, error) { return nil, exec.ErrNotFound }
		if _, _, ok := ShippedWorkOnMain("76"); ok {
			t.Error("absent main ref should degrade to not-shipped, never hard-fail")
		}
	})
}

// TestSubjectAnchorRE verifies the regex used inside CommitWindow to
// filter --grep candidates down to true subject-anchored matches.
// We rebuild the same pattern here (CommitWindow compiles it inline from
// the issue number); equivalent to close-issue.py's
//
//	^(close\s+)?#NN(?!\d)
//
// Go's RE2 doesn't support negative lookahead, so we render (?!\d) as
// ($|[^0-9]). Equivalent behavior for the close-issue use case.
func TestSubjectAnchorRE(t *testing.T) {
	subjectRE := regexp.MustCompile(`^(close\s+)?#15($|[^0-9])`)
	tests := []struct {
		subject string
		match   bool
	}{
		{"#15: subject", true},
		{"close #15 done", true},
		{"close   #15: tabby", true},
		{"#15", true},
		{"#150: different issue", false},
		{"chore: see #15 in body", false},
		{"prefix #15 not anchored", false},
	}
	for _, tt := range tests {
		got := subjectRE.MatchString(tt.subject)
		if got != tt.match {
			t.Errorf("MatchString(%q) = %v, want %v", tt.subject, got, tt.match)
		}
	}
}

// TestDiffNameStatus pins the A/M/R/D parsing the #124 grandfather design rests on:
// the status truncation (`R100`→`R`, via fields[0][:1]) and the rename-destination
// (last tab field) path. Previously these were only ever reached through the gate's
// stubbed seam — the real git-output parser had zero direct coverage (#124 M2 review).
func TestDiffNameStatus(t *testing.T) {
	orig := run
	defer func() { run = orig }()
	run = func(name string, args ...string) ([]byte, error) {
		return []byte("A\tx.md\nM\ty.md\nD\tz.md\nR100\told.md\tnew.md\n"), nil
	}
	got, err := DiffNameStatus("base", "head")
	if err != nil {
		t.Fatal(err)
	}
	want := []FileChange{
		{Status: "A", Path: "x.md"},
		{Status: "M", Path: "y.md"},
		{Status: "D", Path: "z.md"},
		{Status: "R", Path: "new.md"}, // R100 → "R"; rename destination is the last field
	}
	if len(got) != len(want) {
		t.Fatalf("DiffNameStatus = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("change[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// TestDiffNameStatus_Empty: no changes → empty slice + nil (not an error).
func TestDiffNameStatus_Empty(t *testing.T) {
	orig := run
	defer func() { run = orig }()
	run = func(name string, args ...string) ([]byte, error) { return []byte("\n"), nil }
	got, err := DiffNameStatus("base", "")
	if err != nil || len(got) != 0 {
		t.Fatalf("DiffNameStatus(empty) = (%+v, %v), want (nil, nil)", got, err)
	}
}
