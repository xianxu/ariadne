package activetime

import (
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

// #270: integrating main into an issue branch must not change its measured
// time. The fixture is pair#247's incident in miniature: a session works on #9
// from 00:00 to 01:00 and commits `#9 a` (00:20) and `#9 b` (00:50) on its
// branch, while another session lands `#5 foreign` (00:30) on main. Before the
// fix, every integration made #5 a boundary that took this session's minutes,
// and a plain rebase also moved the window (committer dates are restamped).

func gitAt(t *testing.T, repo, iso string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	cmd.Env = os.Environ()
	if iso != "" {
		cmd.Env = append(cmd.Env, "GIT_AUTHOR_DATE="+iso, "GIT_COMMITTER_DATE="+iso)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func branchCommit(t *testing.T, repo, iso, file, msg string) {
	t.Helper()
	if err := os.WriteFile(repo+"/"+file, []byte(iso), 0o644); err != nil {
		t.Fatal(err)
	}
	gitAt(t, repo, iso, "add", file)
	gitAt(t, repo, iso, "commit", "-q", "-m", msg)
}

// integrationFixture returns a repo whose `main` carries the foreign commit and
// whose `issue` branch carries this session's commits, not yet integrated.
// When mergeMid is set, main is merged into the branch between `#9 a` and
// `#9 b` (the merge-main-mid-work practice), with mergeMid as its subject.
func integrationFixture(t *testing.T, mergeMid string) string {
	repo := gitInit(t)
	branchCommit(t, repo, "2025-12-31T23:00:00+00:00", "base", "base")
	gitAt(t, repo, "", "branch", "-M", "main")
	gitAt(t, repo, "", "switch", "-q", "-c", "issue")
	branchCommit(t, repo, "2026-01-01T00:20:00+00:00", "own", "#9 a")
	gitAt(t, repo, "", "switch", "-q", "main")
	branchCommit(t, repo, "2026-01-01T00:30:00+00:00", "theirs", "#5 foreign")
	gitAt(t, repo, "", "switch", "-q", "issue")
	if mergeMid != "" {
		gitAt(t, repo, "2026-01-01T00:35:00+00:00", "merge", "-q", "-m", mergeMid, "main")
	}
	branchCommit(t, repo, "2026-01-01T00:50:00+00:00", "own", "#9 b")
	return repo
}

func scopeOf(t *testing.T, repo string) Scope {
	return Scope{BranchPoint: gitAt(t, repo, "", "merge-base", "main", "HEAD"), Issue: "9"}
}

func boundaryShape(cs []Commit) []string {
	var out []string
	for _, c := range cs {
		out = append(out, c.Time.UTC().Format("15:04")+" "+c.Subject)
	}
	return out
}

func TestBoundariesSurviveIntegratingMain(t *testing.T) {
	const since, until = "2026-01-01T00:00:00Z", "2026-01-01T00:50:00Z"
	want := []string{"00:20 #9 a", "00:50 #9 b"}

	states := map[string]func(t *testing.T) string{
		"not integrated": func(t *testing.T) string { return integrationFixture(t, "") },
		"rebased": func(t *testing.T) string {
			repo := integrationFixture(t, "")
			gitAt(t, repo, "", "rebase", "-q", "main") // restamps committer dates to now
			return repo
		},
		"rebased, committer date kept": func(t *testing.T) string {
			repo := integrationFixture(t, "")
			gitAt(t, repo, "", "rebase", "-q", "--committer-date-is-author-date", "main")
			return repo
		},
		"merged at the end": func(t *testing.T) string {
			repo := integrationFixture(t, "")
			gitAt(t, repo, "2026-01-01T00:55:00+00:00", "merge", "-q", "--no-edit", "main")
			return repo
		},
		"merged mid-work": func(t *testing.T) string { return integrationFixture(t, "Merge branch 'main' into issue") },
		// #317: ariadne#304's merge subject. Naming the issue kept it a
		// boundary, and its citations split the run with #5.
		"merged mid-work, subject names the issue": func(t *testing.T) string {
			return integrationFixture(t, "#9: merge main (#5 landed)")
		},
	}
	for name, build := range states {
		t.Run(name, func(t *testing.T) {
			repo := build(t)
			scope := scopeOf(t, repo)
			commits, err := loadWindowCommits(repo, since, until, scope)
			if err != nil {
				t.Fatal(err)
			}
			if got := boundaryShape(commits); !reflect.DeepEqual(got, want) {
				t.Errorf("boundaries = %v, want %v", got, want)
			}
			peers, err := WindowIssues(repo, since, until, scope)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(peers, []string{"9"}) {
				t.Errorf("peers = %v, want [9]: #5 reached the branch only through main", peers)
			}

			dir := eventsDir(t, sessionEvents()...)
			res, err := Compute(Options{
				Dirs: []string{dir}, GitRepo: repo, Scope: scope,
				SinceISO: since, UntilISO: until, Issues: peers,
				CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
			})
			if err != nil {
				t.Fatal(err)
			}
			if !approx(res.PerIssue["9"], 50) || res.PerIssue["5"] != 0 {
				t.Errorf("per issue = %v, want #9 = 50 min and nothing for #5", res.PerIssue)
			}
		})
	}
}

// Without a branch point (working directly on main) every commit in the window
// stays a boundary, as before #270.
func TestBoundariesUnscopedOnMain(t *testing.T) {
	repo := integrationFixture(t, "")
	gitAt(t, repo, "", "switch", "-q", "main")
	commits, err := loadWindowCommits(repo, "2026-01-01T00:00:00Z", "2026-01-01T00:50:00Z", Scope{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := boundaryShape(commits), []string{"00:30 #5 foreign"}; !reflect.DeepEqual(got, want) {
		t.Errorf("boundaries = %v, want %v", got, want)
	}
}

// A main commit naming the measured issue (its filing, say) is still its
// activity even though it is not on the branch.
func TestScopedBoundaryKeepsTrunkCommitsNamingTheIssue(t *testing.T) {
	repo := gitInit(t)
	branchCommit(t, repo, "2026-01-01T00:05:00+00:00", "base", "#9: issue-sync: new issue")
	gitAt(t, repo, "", "branch", "-M", "main")
	branchCommit(t, repo, "2026-01-01T00:10:00+00:00", "theirs", "#5 foreign")
	gitAt(t, repo, "", "switch", "-q", "-c", "issue")
	branchCommit(t, repo, "2026-01-01T00:20:00+00:00", "own", "#9 a")
	commits, err := loadWindowCommits(repo, "2026-01-01T00:00:00Z", "2026-01-01T00:50:00Z", scopeOf(t, repo))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"00:05 #9: issue-sync: new issue", "00:20 #9 a"}
	if got := boundaryShape(commits); !reflect.DeepEqual(got, want) {
		t.Errorf("boundaries = %v, want %v", got, want)
	}
}

// #317: a citation names an issue without claiming for it. A main commit that
// cites #9 is not #9's boundary, and the branch's own commit citing #5 claims
// for #9 only, while #5 stays in the mention scope.
func TestCitationsDoNotClaim(t *testing.T) {
	repo := gitInit(t)
	branchCommit(t, repo, "2025-12-31T23:00:00+00:00", "base", "base")
	gitAt(t, repo, "", "branch", "-M", "main")
	gitAt(t, repo, "", "switch", "-q", "-c", "issue")
	branchCommit(t, repo, "2026-01-01T00:20:00+00:00", "own", "#9 a")
	gitAt(t, repo, "", "switch", "-q", "main")
	branchCommit(t, repo, "2026-01-01T00:30:00+00:00", "theirs", "#5: fix the thing #9 found (#9)")
	gitAt(t, repo, "", "switch", "-q", "issue")
	gitAt(t, repo, "", "rebase", "-q", "main")
	branchCommit(t, repo, "2026-01-01T00:50:00+00:00", "own", "#9: log: rationale (F1, #5)")
	const since, until = "2026-01-01T00:00:00Z", "2026-01-01T00:50:00Z"
	scope := scopeOf(t, repo)
	commits, err := loadWindowCommits(repo, since, until, scope)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"00:20 #9 a", "00:50 #9: log: rationale (F1, #5)"}
	if got := boundaryShape(commits); !reflect.DeepEqual(got, want) {
		t.Fatalf("boundaries = %v, want %v", got, want)
	}
	if got := commits[1].Issues; !reflect.DeepEqual(got, []string{"9"}) {
		t.Errorf("claimants = %v, want [9]: (F1, #5) is a citation", got)
	}
	peers, err := WindowIssues(repo, since, until, scope)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(peers, []string{"5", "9"}) {
		t.Errorf("peers = %v, want [5 9]: a citation stays in the mention scope", peers)
	}
	res, err := Compute(Options{
		Dirs: []string{eventsDir(t, sessionEvents()...)}, GitRepo: repo, Scope: scope,
		SinceISO: since, UntilISO: until, Issues: peers,
		CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !approx(res.PerIssue["9"], 50) || res.PerIssue["5"] != 0 {
		t.Errorf("per issue = %v, want #9 = 50 min and nothing for #5", res.PerIssue)
	}
}

// One user event every 10 minutes from 00:00 through 00:50: 50 active minutes.
func sessionEvents() []string {
	var lines []string
	for _, m := range []string{"00", "10", "20", "30", "40", "50"} {
		lines = append(lines, `{"timestamp":"2026-01-01T00:`+m+`:00Z","type":"user","message":{"content":"x"}}`)
	}
	return lines
}
