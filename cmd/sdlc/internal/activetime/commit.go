package activetime

import (
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// Commit is a window commit with its tracked-issue subject refs (deduped,
// order-preserving). Time is the author date (%aI). Commits define segment
// boundaries; each non-suffix segment is anchored by the commit at its end.
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

// loadWindowCommits runs `git -C <repo> log` over [sinceISO, untilISO] and
// returns the commits oldest-first with their tracked-issue subject refs.
// Mirrors active-time-v3.py load_commits. The `%x00%n` record terminator the
// Python used is unnecessary here: %s (subject) never contains a newline, so one
// tab-delimited line per commit is unambiguous.
func loadWindowCommits(repo, sinceISO, untilISO string, pat *regexp.Regexp) ([]Commit, error) {
	args := []string{"log", "--pretty=format:%H%x09%aI%x09%s", "--reverse"}
	if sinceISO != "" {
		args = append(args, "--since="+sinceISO)
	}
	if untilISO != "" {
		args = append(args, "--until="+untilISO)
	}
	out, err := gitRun(expandUser(repo), args...)
	if err != nil {
		return nil, err
	}
	text := strings.TrimRight(string(out), "\n")
	if text == "" {
		return nil, nil
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
		commits = append(commits, Commit{
			Time:    ts,
			SHA:     short7(parts[0]),
			Subject: parts[2],
			Issues:  uniqueRefs(pat, parts[2]),
		})
	}
	return commits, nil
}

// uniqueRefs returns the tracked-issue refs in s, deduped preserving order.
// Mirrors load_commits's seen/uniq pass. nil pattern → no refs.
func uniqueRefs(pat *regexp.Regexp, s string) []string {
	if pat == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, m := range pat.FindAllStringSubmatch(s, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

// short7 truncates a SHA to 7 chars (Python's sha[:7]).
func short7(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
