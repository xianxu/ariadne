package gitx

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// runEnv is run with extra environment entries appended to the process
// environment. It returns stdout even on a non-zero exit (exec's Output keeps
// the bytes beside the *exec.ExitError), so callers can read both the output
// and the exit status via gitExitCode. Overridable like run.
var runEnv = func(env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	return cmd.Output()
}

// rebasedPatchIdentity fixes author and committer so S is a pure function of
// (B_r, B_now, reviewed): the same inputs always yield the same commit SHA.
var rebasedPatchIdentity = []string{
	"GIT_AUTHOR_NAME=sdlc", "GIT_AUTHOR_EMAIL=sdlc@invalid", "GIT_AUTHOR_DATE=@0 +0000",
	"GIT_COMMITTER_NAME=sdlc", "GIT_COMMITTER_EMAIL=sdlc@invalid", "GIT_COMMITTER_DATE=@0 +0000",
}

const rebasedPatchMessage = "sdlc: reviewed patch rebased (#304)"

// RebasedReviewedBase returns S: a deterministic, unreferenced commit whose tree
// is today's merge-base(mainRef, HEAD) plus the patch `reviewed` carried against
// its own merge-base with mainRef (#304, D2/D3). diff(S, HEAD) is therefore the
// interdiff since that review, however the branch integrated main meanwhile
// (merge or rebase). When the bases are equal, tree(S) == tree(reviewed).
//
// conflicted (with a nil error) names the paths where main rewrote lines the
// reviewed patch touched; S then carries git's conflicted tree (with markers),
// so diff(S, HEAD) shows the resolution hunks plus any new work.
//
// S is never referenced; git gc collects it after gc.pruneExpire. Requires git
// >= 2.40 (merge-tree --write-tree --merge-base). A criss-cross history (more
// than one merge base on either side) is an error naming the bases.
func RebasedReviewedBase(mainRef, reviewed string) (s string, conflicted []string, err error) {
	if err := requireGit240(); err != nil {
		return "", nil, err
	}
	if _, err := run("git", "rev-parse", "--verify", "-q", reviewed+"^{commit}"); err != nil {
		return "", nil, fmt.Errorf("reviewed head %q does not resolve to a commit", reviewed)
	}
	bR, err := SoleMergeBase(mainRef, reviewed)
	if err != nil {
		return "", nil, fmt.Errorf("reviewed head %s: %w", reviewed, err)
	}
	bNow, err := SoleMergeBase(mainRef, "HEAD")
	if err != nil {
		return "", nil, fmt.Errorf("HEAD: %w", err)
	}

	out, err := run("git", "merge-tree", "--write-tree", "--name-only", "-z",
		"--merge-base="+bR, bNow, reviewed)
	status := 0
	if err != nil {
		status = gitExitCode(err)
		if status != 1 {
			return "", nil, fmt.Errorf("git merge-tree --merge-base=%s %s %s: %v", bR, bNow, reviewed, err)
		}
	}
	// -z output: <tree>NUL then, on conflict, <path>NUL... and an empty entry
	// before the informational messages.
	fields := strings.Split(string(out), "\x00")
	tree, err := parseObjectID([]byte(fields[0]))
	if err != nil {
		return "", nil, fmt.Errorf("git merge-tree: %w", err)
	}
	if status == 1 {
		seen := map[string]bool{}
		for _, p := range fields[1:] {
			if p == "" {
				break
			}
			if !seen[p] {
				seen[p] = true
				conflicted = append(conflicted, p)
			}
		}
		sort.Strings(conflicted)
	}

	commit, err := runEnv(rebasedPatchIdentity, "git", "commit-tree", "--no-gpg-sign",
		tree, "-p", bNow, "-m", rebasedPatchMessage)
	if err != nil {
		return "", nil, fmt.Errorf("git commit-tree %s -p %s: %v", tree, bNow, err)
	}
	s, err = parseObjectID(commit)
	if err != nil {
		return "", nil, fmt.Errorf("git commit-tree: %w", err)
	}
	return s, conflicted, nil
}

// SoleMergeBase returns the single merge base of a and b, refusing an empty
// answer (no shared history) and a criss-cross one (several bases).
func SoleMergeBase(a, b string) (string, error) {
	out, err := run("git", "merge-base", "--all", a, b)
	if err != nil && gitExitCode(err) != 1 {
		return "", fmt.Errorf("git merge-base --all %s %s: %v", a, b, err)
	}
	bases := strings.Fields(string(out))
	switch len(bases) {
	case 0:
		return "", fmt.Errorf("no merge base between %s and %s", a, b)
	case 1:
		return bases[0], nil
	default:
		return "", fmt.Errorf("criss-cross history: merge bases %s", strings.Join(bases, ", "))
	}
}

var gitVersionRE = regexp.MustCompile(`git version (\d+)\.(\d+)`)

// requireGit240 refuses a git that predates merge-tree --merge-base (2.40).
func requireGit240() error {
	out, err := run("git", "version")
	if err != nil {
		return fmt.Errorf("git version: %v", err)
	}
	v := strings.TrimSpace(string(out))
	m := gitVersionRE.FindStringSubmatch(v)
	if m == nil {
		return fmt.Errorf("git >= 2.40 required for review windows (found %q)", v)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	if major < 2 || (major == 2 && minor < 40) {
		return fmt.Errorf("git >= 2.40 required for review windows (found %s.%s)", m[1], m[2])
	}
	return nil
}

// CommitsWithCloseToken returns the commits in `git -C dir log <rangeArgs...>` (newest
// first) carrying a `Close-Token:` trailer whose value equals token exactly.
// Trailers survive a rebase and a merge of main but not a squash that rewrites
// the message. Exact match, never a body grep: a substring search is the #194
// self-reference trap.
func CommitsWithCloseToken(dir string, rangeArgs []string, token string) ([]string, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("empty Close-Token")
	}
	var args []string
	if dir != "" {
		args = append(args, "-C", dir) // the checkout a caller names, never the process cwd
	}
	args = append(args, "log")
	args = append(args, rangeArgs...)
	args = append(args, "--format=%H%x00%(trailers:key=Close-Token,valueonly,separator=%x2C)")
	out, err := run("git", args...)
	if err != nil {
		return nil, fmt.Errorf("git log %s: %v", strings.Join(rangeArgs, " "), err)
	}
	var shas []string
	for _, line := range strings.Split(string(out), "\n") {
		sha, values, ok := strings.Cut(line, "\x00")
		if !ok {
			continue
		}
		for _, v := range strings.Split(values, ",") {
			if strings.TrimSpace(v) == token {
				shas = append(shas, sha)
				break
			}
		}
	}
	return shas, nil
}
