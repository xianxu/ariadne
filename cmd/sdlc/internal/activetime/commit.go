package activetime

import (
	"fmt"
	"os/exec"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueref"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// Commit is a window commit with its subject issue refs (deduped,
// order-preserving). Time is the author date (%aI). Commits define global time
// boundaries; commits with issue refs can claim nearby activity runs, while
// no-ref commits remain neutral boundaries.
type Commit struct {
	Time    time.Time
	SHA     string // short (7)
	Subject string
	Issues  []string
}

// gitRun is the package-level git runner (mirrors gitx's run shim) so
// commit_test.go can drive fixtures without spawning git. repo is passed via
// `-C` so the standalone `--git-repo <path>` flag can target arbitrary repos.
var gitRun = func(repo string, args ...string) ([]byte, error) {
	full := append([]string{"-C", repo}, args...)
	return exec.Command("git", full...).Output()
}

// Scope limits a measurement's boundaries to one issue branch (#270). Main's
// history reaches a branch whenever main is integrated (rebase or merge), and
// its commits belong to whatever sessions produced them; letting them bound
// this branch's segments hands this session's time to those issues. With a
// BranchPoint, a commit is a boundary only if it is the branch's own non-merge
// commit (in BranchPoint..HEAD) or its subject names Issue (the filing commit
// on main, the issue's tracker claim/close). The zero Scope keeps every commit
// in the window: work done directly on main has no branch to scope to.
type Scope struct {
	BranchPoint string
	Issue       string
}

// loadWindowCommits reads `git -C <repo> log` and returns the commits whose
// AUTHOR date lies in [sinceISO, untilISO], oldest-first, with their
// tracked-issue subject refs. The window is filtered here rather than with
// `--since/--until`, which compare committer dates: a plain rebase restamps
// those, and would move commits in or out of the window (#270). Reading the
// whole history costs ~50 ms on ariadne's. Mirrors active-time-v3.py
// load_commits; %s never contains a newline, so one tab-delimited line per
// commit is unambiguous.
func loadWindowCommits(repo, sinceISO, untilISO string, scope Scope, extraRefs ...string) ([]Commit, error) {
	since, until, err := windowBounds(sinceISO, untilISO)
	if err != nil {
		return nil, err
	}
	args := []string{"log", "--pretty=format:%H%x09%aI%x09%s", "--reverse"}
	if len(extraRefs) > 0 {
		// Histories beside HEAD (#252: the issue tracker's claim/close commits);
		// git log lists a commit reachable from several refs once.
		args = append(append(args, "HEAD"), extraRefs...)
	}
	out, err := gitRun(expandUser(repo), args...)
	if err != nil {
		return nil, err
	}
	text := strings.TrimRight(string(out), "\n")
	if text == "" {
		return nil, nil
	}
	own, err := branchOwn(repo, scope)
	if err != nil {
		return nil, err
	}
	// Resolved once from the repo the commits came from, not per line.
	self, err := selfQualifier(repo)
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		ts, err := parseISO(parts[1])
		if err != nil {
			continue
		}
		if (!since.IsZero() && ts.Before(since)) || (!until.IsZero() && ts.After(until)) {
			continue
		}
		issues := issueref.LocalNums(parts[2], self)
		if own != nil && !own[parts[0]] && !slices.Contains(issues, scope.Issue) {
			continue
		}
		commits = append(commits, Commit{
			Time:    ts,
			SHA:     short7(parts[0]),
			Subject: parts[2],
			Issues:  issues,
		})
	}
	return commits, nil
}

// branchOwn returns the full SHAs of the branch's own non-merge commits, or nil
// for the zero Scope. A merge commit is integration, not work: merging main in
// mid-issue must not add a boundary the unintegrated branch lacked.
func branchOwn(repo string, scope Scope) (map[string]bool, error) {
	if scope.BranchPoint == "" {
		return nil, nil
	}
	out, err := gitRun(expandUser(repo), "rev-list", "--no-merges", "HEAD", "^"+scope.BranchPoint)
	if err != nil {
		return nil, fmt.Errorf("branch commits since %s: %w", scope.BranchPoint, err)
	}
	own := map[string]bool{}
	for _, sha := range strings.Fields(string(out)) {
		own[sha] = true
	}
	return own, nil
}

// windowBounds parses the window edges; "" leaves that side open.
func windowBounds(sinceISO, untilISO string) (since, until time.Time, err error) {
	if sinceISO != "" {
		if since, err = parseISO(sinceISO); err != nil {
			return since, until, fmt.Errorf("window start %q: %w", sinceISO, err)
		}
	}
	if untilISO != "" {
		if until, err = parseISO(untilISO); err != nil {
			return since, until, fmt.Errorf("window end %q: %w", untilISO, err)
		}
	}
	return since, until, nil
}

// WindowIssues returns the distinct local issues the window's boundary commits
// reference plus scope.Issue (measured even before its first commit),
// numerically sorted. It reads the same commits Compute segments on,
// so the tracked set and the boundaries cannot disagree: an issue that reached
// the branch only through main is in neither (#270).
func WindowIssues(repo, sinceISO, untilISO string, scope Scope, extraRefs ...string) ([]string, error) {
	commits, err := loadWindowCommits(repo, sinceISO, untilISO, scope, extraRefs...)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []string
	if scope.Issue != "" {
		seen[scope.Issue] = true
		out = append(out, scope.Issue)
	}
	for _, c := range commits {
		for _, iss := range c.Issues {
			if !seen[iss] {
				seen[iss] = true
				out = append(out, iss)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		a, _ := strconv.Atoi(out[i])
		b, _ := strconv.Atoi(out[j])
		return a < b
	})
	return out, nil
}

// selfQualifier derives local issue qualifiers from the explicitly targeted Git
// repository, preserving identity across linked checkouts and nested paths.
// Unverified paths are errors; a filesystem basename is never evidence.
func selfQualifier(repo string) (string, error) {
	identity, err := resolveCommitWorkspace(expandUser(repo))
	if err != nil {
		return "", err
	}
	return identity.Repo, nil
}

type commitGitReader struct{}

func (commitGitReader) GitInDir(dir string, args ...string) ([]byte, error) {
	return gitRun(dir, args...)
}

var resolveCommitWorkspace = func(dir string) (workspace.Identity, error) {
	return workspace.Resolve(commitGitReader{}, dir, "")
}

// short7 truncates a SHA to 7 chars (Python's sha[:7]).
func short7(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
