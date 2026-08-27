package judge

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	fullObjectID = regexp.MustCompile(`^[0-9a-f]{40}$`)
	shellSafeArg = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./-]+$`)
)

// ReviewWindowManifest is the resolved, immutable input to one boundary review.
// WorkingTree selects base-vs-working-tree scope; otherwise HeadSHA closes a
// committed range. Git resolution belongs to the caller's IO shell.
type ReviewWindowManifest struct {
	RepoRoot       string
	BaseSHA        string
	HeadSHA        string
	AmbientHeadSHA string
	WorkingTree    bool
	IssuesDir      string
	HistoryDir     string
	IssueFile      string
	PlanFile       string
}

// ReviewCommand is one read-only repository-inspection recipe. Args stay
// structured; Shell renders them for a human but is never executed by sdlc.
type ReviewCommand struct {
	Label string
	Args  []string
}

func (c ReviewCommand) Shell() string {
	quoted := make([]string, len(c.Args))
	for i, arg := range c.Args {
		quoted[i] = shellQuoteArg(arg)
	}
	return strings.Join(quoted, " ")
}

// RenderReviewWindow validates already-resolved data and renders the compact
// boundary-review manifest. It performs no filesystem or Git IO.
func RenderReviewWindow(m ReviewWindowManifest) (string, error) {
	if err := validateReviewWindow(m); err != nil {
		return "", err
	}

	mode := "committed range"
	headLine := "head: " + m.HeadSHA
	if m.WorkingTree {
		mode = "working tree"
		headLine = "ambient HEAD: " + m.AmbientHeadSHA
	}

	lines := []string{
		"mode: " + mode,
		"repository root: " + m.RepoRoot,
		"base: " + m.BaseSHA,
		headLine,
	}
	if m.WorkingTree {
		lines = append(lines, "scope: includes committed-after-base plus staged and unstaged tracked changes; excludes untracked files")
	}
	if m.IssueFile != "" {
		lines = append(lines, "issue file: "+m.IssueFile)
	}
	if m.PlanFile != "" {
		lines = append(lines, "plan file: "+m.PlanFile)
	}
	lines = append(lines, "commands:")
	for _, cmd := range reviewCommands(m) {
		lines = append(lines, "  "+cmd.Label+": "+cmd.Shell())
	}
	return strings.Join(lines, "\n"), nil
}

func validateReviewWindow(m ReviewWindowManifest) error {
	if !filepath.IsAbs(m.RepoRoot) {
		return fmt.Errorf("review repository root must be absolute: %q", m.RepoRoot)
	}
	if !fullObjectID.MatchString(m.BaseSHA) {
		return fmt.Errorf("review base must be a full commit object id: %q", m.BaseSHA)
	}
	if m.IssuesDir == "" || m.HistoryDir == "" {
		return fmt.Errorf("review issue/history exclusions must be non-empty")
	}
	if m.WorkingTree {
		if m.HeadSHA != "" {
			return fmt.Errorf("working-tree review must not carry a committed head")
		}
		if !fullObjectID.MatchString(m.AmbientHeadSHA) {
			return fmt.Errorf("working-tree review ambient HEAD must be a full commit object id: %q", m.AmbientHeadSHA)
		}
		return nil
	}
	if !fullObjectID.MatchString(m.HeadSHA) {
		return fmt.Errorf("committed review head must be a full commit object id: %q", m.HeadSHA)
	}
	if m.AmbientHeadSHA != "" {
		return fmt.Errorf("committed review must not carry an ambient HEAD")
	}
	return nil
}

func reviewCommands(m ReviewWindowManifest) []ReviewCommand {
	prefix := []string{"git", "-C", m.RepoRoot, "diff"}
	rangeArgs := []string{m.BaseSHA}
	if !m.WorkingTree {
		rangeArgs = append(rangeArgs, m.HeadSHA)
	}
	excludes := []string{"--", ":!" + strings.TrimSuffix(m.IssuesDir, "/") + "/", ":!" + strings.TrimSuffix(m.HistoryDir, "/") + "/"}
	command := func(label string, option, target string) ReviewCommand {
		args := append([]string{}, prefix...)
		if option != "" {
			args = append(args, option)
		}
		args = append(args, rangeArgs...)
		args = append(args, "--")
		if target != "" {
			args = append(args, target)
		}
		args = append(args, excludes[1:]...)
		return ReviewCommand{Label: label, Args: args}
	}
	return []ReviewCommand{
		command("stat", "--stat", ""),
		command("names", "--name-status", ""),
		command("full", "", ""),
		command("targeted", "", "<path-from-name-status>"),
	}
}

func shellQuoteArg(arg string) string {
	if arg != "" && shellSafeArg.MatchString(arg) {
		return arg
	}
	return "'" + strings.ReplaceAll(arg, "'", `'"'"'`) + "'"
}
