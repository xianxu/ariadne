package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// fixedFlakeIssue is a terminal card joined to its published detail body.
// Unknown detail evidence must be retained: it could claim any test.
type fixedFlakeIssue struct {
	ID, Merge, Details string
	Tests              []string
	DetailKnown        bool
}

type fixedFlakeState uint8

const (
	flakeUnknown fixedFlakeState = iota
	flakeFixed
	flakeRecurred
)

var errFixedGitUnknown = errors.New("Git ancestry is unknown")

type flakeListView struct {
	Registry flakes.Registry
	Cards    []flakes.Card
	Fixed    []fixedFlakeIssue
	Main     string
	FixedErr error
}

// loadFlakeListView reads the tracker only once. Fixed-issue evidence is
// advisory to this read-only view, so its failure never masks a failure row.
func loadFlakeListView(ctx context.Context, issuesDir string) (flakeListView, error) {
	rs, err := loadIssueRecords(ctx, issuesDir, tracker.Fresh)
	if err != nil {
		return flakeListView{}, err
	}
	cards, err := flakeCards(rs)
	if err != nil {
		return flakeListView{}, err
	}
	registry, err := flakes.BuildRegistry(cards, flakeIssueOpen)
	if err != nil {
		return flakeListView{}, err
	}
	view := flakeListView{Registry: registry, Cards: cards}
	view.Fixed, view.Main, view.FixedErr = publishedFixedIssues(ctx, issuesDir, rs)
	return view, nil
}

func flakeGitContains(ctx context.Context, root, ancestor, head string) (bool, error) {
	_, err := gitx.GitOutputContext(ctx, "-C", root, "merge-base", "--is-ancestor", ancestor, head)
	if err == nil {
		return true, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

// A legacy one-word package label proves ownership only if the issue's merge
// tree has the expected repository-relative package and no other Go package
// with that basename. A later tree cannot establish that historical identity.
func flakeUniquePackageAtMerge(ctx context.Context, root, merge, pkg string) (bool, error) {
	rel := packageRelativePath(pkg)
	if rel == "" {
		return false, nil
	}
	raw, err := gitx.GitOutputContext(ctx, "-C", root, "ls-tree", "-r", "--name-only", merge)
	if err != nil {
		return false, err
	}
	found := false
	for _, file := range strings.Split(string(raw), "\n") {
		if path.Ext(file) != ".go" {
			continue
		}
		dir := path.Dir(file)
		if path.Base(dir) != path.Base(rel) {
			continue
		}
		if dir != rel {
			return false, nil
		}
		found = true
	}
	return found, nil
}

func publishedFixedIssues(ctx context.Context, issuesDir string, rs tracker.Records) ([]fixedFlakeIssue, string, error) {
	root := repoRootOfContext(ctx, issuesDir)
	if root == "" {
		return nil, "", errors.New("issue details are outside a repository")
	}
	trunk, err := peerTrunk(ctx, root)
	if err != nil {
		return nil, "", err
	}
	main, err := trunk.Snapshot()
	if err != nil {
		return nil, "", err
	}
	active, err := gitx.InsideRoot(root, issuesDir)
	if err != nil {
		return nil, "", err
	}
	archived, err := gitx.InsideRoot(root, vocab.ArchiveSubdir(historyDir(), vocab.ArchiveIssues))
	if err != nil {
		return nil, "", err
	}
	details := map[string]string{}
	duplicates := map[string]bool{}
	for _, dir := range []string{active, archived} {
		files, err := main.FilesSelected(filepath.ToSlash(dir), func(file gitx.TreeFile) bool {
			_, _, ok := issue.ParseFilename(path.Base(file.Path))
			return ok && path.Dir(file.Path) == filepath.ToSlash(dir)
		})
		if err != nil {
			return nil, "", err
		}
		for _, file := range files {
			id, _, _ := issue.ParseFilename(path.Base(file.Path))
			if _, exists := details[id]; exists {
				duplicates[id] = true
			}
			_, body, err := issue.Parse(string(file.Content))
			if err != nil {
				duplicates[id] = true
				continue
			}
			details[id] = body
		}
	}
	var fixed []fixedFlakeIssue
	for _, rec := range rs.All() {
		if rec.Duplicate || rec.Card == nil || !vocab.Issue().IsTerminal(rec.Status()) {
			continue
		}
		tests, _, err := issue.FieldList(rec.Card.Card.Frontmatter, flakeTestsField)
		if err != nil {
			return nil, "", fmt.Errorf("issue %s flake_tests: %w", rec.ID, err)
		}
		completion, _, err := issue.CardCompletion(rec.Card.Raw)
		if err != nil {
			return nil, "", fmt.Errorf("issue %s completion: %w", rec.ID, err)
		}
		_, present := details[rec.ID]
		fixed = append(fixed, fixedFlakeIssue{ID: rec.ID, Merge: completion.LandedCommit, Tests: tests, Details: details[rec.ID], DetailKnown: present && !duplicates[rec.ID]})
	}
	return fixed, main.Ref(), nil
}

func fixedIssueCovers(tests []string, details, name string, bareCollision bool) bool {
	strong, _ := fixedIssueCoverage(tests, details, name, bareCollision)
	return strong
}

// fixedIssueCoverage distinguishes an exact package path from a bare package
// label. A bare label needs repository evidence before it can claim a test.
func fixedIssueCoverage(tests []string, details, name string, bareCollision bool) (strong, bareLabel bool) {
	for _, test := range tests {
		if test == name {
			return true, false
		}
	}
	pkg, bare, ok := flakes.SplitTest(name)
	if !ok {
		return false, false
	}
	rel := packageRelativePath(pkg)
	lines := strings.Split(details, "\n")
	fenced := issue.FenceSpans(lines, issue.UnterminatedIsProse)
	section := ""
	for n, line := range lines {
		if fenced[n] {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			section = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			continue
		}
		if section == "Log" {
			trim := strings.TrimSpace(line)
			if !strings.HasPrefix(trim, "- Test:") {
				continue
			}
			line = strings.TrimSpace(strings.TrimPrefix(trim, "- Test:"))
		} else if section != "" && section != "Problem" && section != "Spec" && section != "Done when" && section != "Plan" {
			continue
		}
		if containsTestToken(line, name) || (strings.Contains(rel, "/") && containsTestToken(line, bare) && containsPackagePath(line, rel)) {
			return true, false
		}
		if !bareCollision && containsTestToken(line, bare) && containsStandalonePackage(line, path.Base(pkg)) {
			bareLabel = true
		}
	}
	return false, bareLabel
}

func packageRelativePath(pkg string) string {
	parts := strings.SplitN(pkg, "/", 4)
	if len(parts) != 4 {
		return ""
	}
	return parts[3]
}

func containsPackagePath(line, rel string) bool {
	for start := 0; start < len(line); {
		i := strings.Index(line[start:], rel)
		if i < 0 {
			return false
		}
		i += start
		end := i + len(rel)
		beforeOK := i == 0 || !packagePathByte(line[i-1])
		afterOK := end == len(line) || !packagePathByte(line[end])
		if end < len(line) && line[end] == '/' {
			// A source file directly inside this package is context. A nested
			// directory names another package even when the parent prefix matches.
			rest := line[end+1:]
			goEnd := strings.Index(rest, ".go")
			firstSlash := strings.IndexByte(rest, '/')
			afterOK = goEnd >= 0 && (firstSlash < 0 || goEnd < firstSlash) &&
				(goEnd+3 == len(rest) || !packagePathByte(rest[goEnd+3]))
		}
		if beforeOK && afterOK {
			return true
		}
		start = i + 1
	}
	return false
}

func containsStandalonePackage(line, name string) bool {
	for start := 0; start < len(line); {
		i := strings.Index(line[start:], name)
		if i < 0 {
			return false
		}
		i += start
		end := i + len(name)
		beforeOK := i == 0 || !packagePathByte(line[i-1])
		afterOK := end == len(line) || !packagePathByte(line[end])
		if beforeOK && afterOK {
			return true
		}
		start = i + 1
	}
	return false
}

func packagePathByte(b byte) bool {
	return testWordByte(b) || b == '/' || b == '-'
}

func containsTestToken(line, name string) bool {
	for start := 0; start < len(line); {
		i := strings.Index(line[start:], name)
		if i < 0 {
			return false
		}
		i += start
		end := i + len(name)
		beforeOK := i == 0 || !testWordByte(line[i-1])
		afterOK := end == len(line) || !testWordByte(line[end])
		if beforeOK && afterOK {
			return true
		}
		start = i + 1
	}
	return false
}

func testWordByte(b byte) bool {
	return b == '_' || b == '.' || unicode.IsLetter(rune(b)) || unicode.IsDigit(rune(b))
}

// classifyFixedFlake coordinates ownership with injected Git ancestry. It is
// integration logic: fixedIssueCovers is the pure document matcher.
func classifyFixedFlake(name string, issues []fixedFlakeIssue, bareCollision bool, contains func(string, string) (bool, error), packageUnique func(string, string) (bool, error), main, head string) (fixedFlakeState, fixedFlakeIssue, string) {
	var candidates []fixedFlakeIssue
	for _, candidate := range issues {
		if !candidate.DetailKnown {
			// A legacy terminal card with neither details nor a landing merge
			// cannot prove a fix. Its explicit flake_tests can still make this
			// test's ownership uncertain; an unrelated empty card cannot.
			if candidate.Merge != "" || fixedIssueCovers(candidate.Tests, "", name, bareCollision) {
				return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("issue #%s details unreadable", candidate.ID)
			}
			continue
		}
		strong, bareLabel := fixedIssueCoverage(candidate.Tests, candidate.Details, name, bareCollision)
		if strong {
			candidates = append(candidates, candidate)
			continue
		}
		if bareLabel {
			if candidate.Merge == "" {
				return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("issue #%s has no landing merge for its package label", candidate.ID)
			}
			pkg, _, _ := flakes.SplitTest(name)
			if packageUnique == nil {
				return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("package label for issue #%s cannot be verified", candidate.ID)
			}
			unique, err := packageUnique(candidate.Merge, pkg)
			if err != nil {
				return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("package label for issue #%s cannot be verified: %v", candidate.ID, err)
			}
			if !unique {
				return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("package label for issue #%s is ambiguous or absent at its merge", candidate.ID)
			}
			candidates = append(candidates, candidate)
		}
	}
	if len(candidates) > 1 {
		return flakeUnknown, fixedFlakeIssue{}, "multiple terminal issues cover this test"
	}
	if len(candidates) == 0 {
		return flakeUnknown, fixedFlakeIssue{}, ""
	}
	if candidates[0].Merge == "" {
		return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("issue #%s has no landing merge", candidates[0].ID)
	}
	onMain, err := contains(candidates[0].Merge, main)
	if err != nil {
		return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("Git ancestry unavailable for issue #%s: %v", candidates[0].ID, err)
	}
	if !onMain {
		return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("issue #%s landing merge is not on main", candidates[0].ID)
	}
	inHead, err := contains(candidates[0].Merge, head)
	if err != nil {
		return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("Git ancestry unavailable for issue #%s: %v", candidates[0].ID, err)
	}
	if inHead {
		return flakeRecurred, candidates[0], ""
	}
	return flakeFixed, candidates[0], ""
}
