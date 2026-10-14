// repoenums.go — the IO shell around internal/repoenum (#348): reading a
// repository's declaration from main, and the one classification check every
// writer and gate applies.
package main

import (
	"fmt"
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

// declReader reads a repository's declaration. Callers pass one that is only
// invoked when there is an area to check (#348 BR-1: input carrying no area
// never depends on the declaration being readable), memoized by lazyDecl.
type declReader func() (repoenum.Declaration, error)

// declUnreadable marks an error reading or parsing the declaration itself (as
// opposed to a value it refuses), so a gate can fail closed on it.
type declUnreadable struct{ err error }

func (e declUnreadable) Error() string { return e.err.Error() }
func (e declUnreadable) Unwrap() error { return e.err }

// lazyDecl memoizes read: at most one read per command or gate run. Its
// errors are declUnreadable.
func lazyDecl(read declReader) declReader {
	var (
		done bool
		decl repoenum.Declaration
		err  error
	)
	return func() (repoenum.Declaration, error) {
		if !done {
			decl, err = read()
			if err != nil {
				err = declUnreadable{err}
			}
			done = true
		}
		return decl, err
	}
}

// checkClassification checks areas (syntax, no duplicates, membership in the
// declaration) and a severity (the shared enum). No areas and no severity
// pass without reading anything; the declaration is read only for areas.
// Every writer and gate applies this one check.
func checkClassification(read declReader, areas []string, severity string) error {
	if err := issue.CheckSeverity(severity); err != nil {
		return err
	}
	if len(areas) == 0 {
		return nil
	}
	spec := vocab.Issue().RepoEnums.Fields["area"]
	seen := map[string]bool{}
	for _, a := range areas {
		if err := issue.CheckAreaSyntax(a); err != nil {
			return err
		}
		if seen[a] {
			return fmt.Errorf("area %q listed twice", a)
		}
		seen[a] = true
	}
	if spec.MaxItems > 0 && len(areas) > spec.MaxItems {
		return fmt.Errorf("at most %d areas, got %d", spec.MaxItems, len(areas))
	}
	decl, err := read()
	if err != nil {
		return err
	}
	if undeclared := decl.Undeclared("area", areas); len(undeclared) > 0 {
		return repoenum.Refusal(vocab.Issue().RepoEnums, "area", undeclared)
	}
	return nil
}

// uniqueAreas drops repeated areas, keeping first occurrences in order (a
// writer's convenience; the check refuses duplicates in hand-edited input).
func uniqueAreas(areas []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, a := range areas {
		if !seen[a] {
			seen[a] = true
			out = append(out, a)
		}
	}
	return out
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
