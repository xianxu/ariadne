// projectfind.go — fleet-wide project navigation (#171 M4).
//
// `sdlc project find --issue <ref>` and `sdlc resolve --kind project <ref>`
// both answer "which project records reference this issue, anywhere in the
// fleet?" via the shared M2 discovery (`project.DiscoverByIssueRef`) under
// ActiveAndArchive scope — navigation must find a record regardless of its
// lifecycle location (active workshop/projects, archived history/projects, or
// the deprecated brain legacy home). Read-only: never tagged
// markMutatingCommand, so parley can shell to it on a keypress like resolve.
package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
)

type projectFindFlags struct {
	Issue string
	root  string // test injection; "" ⇒ gitx.RepoTopLevel()
}

func newProjectFindCmd() *cobra.Command {
	f := projectFindFlags{}
	cmd := &cobra.Command{Use: "find", Short: "Find every project record referencing an issue, fleet-wide", Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProjectFind(cmd.OutOrStdout(), &f)
		}}
	cmd.Flags().StringVar(&f.Issue, "issue", "", "issue ref (e.g. metis#18 or #171)")
	_ = cmd.MarkFlagRequired("issue")
	return cmd
}

// discoverProjectsForRef parses an issue ref and returns every project record
// referencing it across the fleet, archive-inclusive. Shared by `project find`
// and `resolve --kind project` (ARCH-DRY). The ref's repo token resolves with
// the same sibling matching as `sdlc resolve` (exact, then unique prefix).
func discoverProjectsForRef(refStr, root string) ([]projectdoc.ProjectMatch, ArtifactRef, error) {
	ref, err := parseRef(refStr)
	if err != nil {
		return nil, ArtifactRef{}, err
	}
	if ref.GitHub {
		return nil, ref, fmt.Errorf("github refs have no project records (projects reference workshop issues)")
	}
	repoDir, err := resolveRepoDir(ref, root)
	if err != nil {
		return nil, ref, err
	}
	repoName := filepath.Base(repoDir)
	matches, err := projectdoc.DiscoverByIssueRef(filepath.Dir(root), repoName, strconv.Itoa(ref.ID), projectdoc.ActiveAndArchive)
	if err != nil {
		return nil, ref, err
	}
	if len(matches) == 0 {
		return nil, ref, fmt.Errorf("no project across the fleet references %s#%d", repoName, ref.ID)
	}
	return matches, ref, nil
}

func runProjectFind(stdout io.Writer, f *projectFindFlags) error {
	root, err := currentRoot(f.root)
	if err != nil {
		return err
	}
	matches, _, err := discoverProjectsForRef(f.Issue, root)
	if err != nil {
		return err
	}
	for _, m := range matches {
		line := m.Path
		if m.Legacy {
			line += " (legacy)"
		}
		fmt.Fprintln(stdout, line)
	}
	return nil
}
