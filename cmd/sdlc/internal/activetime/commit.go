package activetime

import (
	"os/exec"
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

// loadWindowCommits runs `git -C <repo> log` over [sinceISO, untilISO] and
// returns the commits oldest-first with their tracked-issue subject refs.
// Mirrors active-time-v3.py load_commits. The `%x00%n` record terminator the
// Python used is unnecessary here: %s (subject) never contains a newline, so one
// tab-delimited line per commit is unambiguous.
func loadWindowCommits(repo, sinceISO, untilISO string) ([]Commit, error) {
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
		commits = append(commits, Commit{
			Time:    ts,
			SHA:     short7(parts[0]),
			Subject: parts[2],
			Issues:  issueref.LocalNums(parts[2], self),
		})
	}
	return commits, nil
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
