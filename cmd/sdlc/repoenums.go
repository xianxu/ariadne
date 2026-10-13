// repoenums.go — the IO shell around internal/repoenum (#348): reading a
// repository's declaration from main, and the one classification check every
// writer and gate applies.
package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/repoenum"
	"github.com/xianxu/ariadne/pkg/vocab"
)

func parseRepoEnums(data []byte) (repoenum.Declaration, error) {
	return repoenum.Parse(data, vocab.Issue().RepoEnums)
}

// readTrackerRepoEnums reads the declaration from the publication target's
// main, freshly fetched (a tracker writer's view). Absent is empty.
func readTrackerRepoEnums(env *trackerEnv) (repoenum.Declaration, error) {
	data, err := env.main.Read(vocab.Issue().RepoEnums.DeclarationPath)
	if err != nil {
		return repoenum.Declaration{}, fmt.Errorf("read %s on main: %w", vocab.Issue().RepoEnums.DeclarationPath, err)
	}
	return parseRepoEnums(data)
}

// repoEnumsAtRev reads the declaration from rev's tree; absent is empty.
func repoEnumsAtRev(git func(args ...string) (string, error), rev string) (repoenum.Declaration, error) {
	path := vocab.Issue().RepoEnums.DeclarationPath
	out, err := git("ls-tree", "--name-only", rev, "--", path)
	if err != nil {
		return repoenum.Declaration{}, fmt.Errorf("read %s's tree for %s: %w", rev, path, err)
	}
	if strings.TrimSpace(out) == "" {
		return parseRepoEnums(nil)
	}
	data, err := git("show", rev+":"+path)
	if err != nil {
		return repoenum.Declaration{}, fmt.Errorf("read %s on %s: %w", path, rev, err)
	}
	return parseRepoEnums([]byte(data))
}

// repoEnumsMainRev names the rev a legacy writer reads the declaration from:
// origin/main, else the local main, never HEAD (a branch must not declare its
// own values).
func repoEnumsMainRev(git func(args ...string) (string, error)) (string, error) {
	for _, ref := range []string{"refs/remotes/origin/main", "refs/heads/main"} {
		if _, err := git("rev-parse", "--verify", "--quiet", ref); err == nil {
			return ref, nil
		}
	}
	return "", fmt.Errorf("no main to read %s from", vocab.Issue().RepoEnums.DeclarationPath)
}

// checkClassification checks area values (syntax and membership in the
// declaration) and a severity (membership in the shared enum). An empty
// severity and no areas pass. Every writer and gate applies this one check.
func checkClassification(decl repoenum.Declaration, areas []string, severity string) error {
	m := vocab.Issue()
	spec := m.RepoEnums.Fields["area"]
	for _, a := range areas {
		if !repoenum.ValidValue(spec, a) {
			return fmt.Errorf("area %q is not a /-path of lowercase segments (at most %d)", a, spec.MaxDepth)
		}
	}
	if spec.MaxItems > 0 && len(areas) > spec.MaxItems {
		return fmt.Errorf("at most %d areas, got %d", spec.MaxItems, len(areas))
	}
	if undeclared := decl.Undeclared("area", areas); len(undeclared) > 0 {
		return repoenum.Refusal(m.RepoEnums, "area", undeclared)
	}
	if severity != "" && !slices.Contains(m.Enums["severity"], severity) {
		return fmt.Errorf("severity %q is not one of %s", severity, strings.Join(m.Enums["severity"], ", "))
	}
	return nil
}

// frontmatterClassification reads area and severity from frontmatter (a card's
// or a details file's), for the gates.
func frontmatterClassification(fm string) (areas []string, severity string, err error) {
	areas, _, err = issue.FieldList(fm, "area")
	if err != nil {
		return nil, "", err
	}
	severity, err = issue.FieldString(fm, "severity")
	return areas, severity, err
}

// flowList renders values as the card's one-line flow list.
func flowList(values []string) string { return "[" + strings.Join(values, ", ") + "]" }
