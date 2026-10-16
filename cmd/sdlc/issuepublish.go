package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/repoenum"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

func newIssuePublishCmd() *cobra.Command {
	var source, issuesDir string
	var issues []int
	var dryRun bool
	cmd := markMutatingCommand(&cobra.Command{
		Use:   "publish --issue N[,N…] | --commit SHA",
		Short: "Publish later edits to issue details on main (#284, #274); finishes an incomplete creation",
		Args:  cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			guardSpineRepo(cmd.ErrOrStderr())
			root, err := gitx.RepoTopLevel()
			if err != nil {
				return err
			}
			cut, err := tracker.CutOver(root)
			if err != nil {
				return err
			}
			if !cut {
				if len(issues) > 0 {
					return errors.New("--issue publishes in an issue tracker repository; here, publish one documentation commit with --commit SHA")
				}
				if source == "" {
					return fmt.Errorf("--commit SHA is required; select one coherent documentation commit")
				}
				return publishIssueCommit(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), root, source,
					envOr("WF_ISSUES_DIR", "workshop/issues"), envOr("WF_HISTORY_DIR", "workshop/history"))
			}
			if source != "" {
				return errors.New("this repository uses the issue tracker: `--commit` (copying commits to main) is retired here; publish details with `sdlc issue publish --issue N`")
			}
			return runTrackedPublish(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), issues, issuesDir, dryRun)
		},
	})
	cmd.Flags().IntSliceVar(&issues, "issue", nil, "issue ID(s) whose details to publish (issue tracker repositories): --issue 284 or --issue 284,285")
	cmd.Flags().StringVar(&source, "commit", "", "one non-merge documentation commit to publish in full (legacy repositories)")
	cmd.Flags().StringVar(&issuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue details")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "check and describe; change nothing")
	return cmd
}

// runTrackedPublish publishes each issue's details to main (#284). Details
// already on main are edits (by the owner, or anyone while unclaimed),
// published together by republishOwned. Details not yet on main are an
// incomplete creation — filed before #308, when issue new wrote them locally —
// finished by the transfer, one issue at a time, needing no owner.
func runTrackedPublish(ctx context.Context, stdout, stderr io.Writer, nums []int, issuesDir string, dryRun bool) error {
	nums = claimIssues(&claimFlags{Issues: nums})
	if len(nums) == 0 {
		return errors.New("--issue N (or a list: --issue 284,285) is required")
	}
	dirs, err := resolveIDDirs(issuesDir, envOr("WF_HISTORY_DIR", "workshop/history"))
	if err != nil {
		return err
	}
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	local, _, err := issue.LocalPathsByID(dirs.Abs)
	if err != nil {
		return err
	}
	locations := map[string]string{}
	var first []int
	var republish []string
	for _, n := range nums {
		id := fmt.Sprintf("%06d", n)
		card, err := snap.Require(id)
		if err != nil {
			return err
		}
		var localDetails []string
		candidates := slices.DeleteFunc(local[n], func(p string) bool { return !issueFilename(filepath.Base(p)) })
		for _, f := range classifyFamily(n, candidates) {
			if f.Kind == kindIssue {
				localDetails = append(localDetails, f.Path)
			}
		}
		if len(localDetails) > 1 {
			return fmt.Errorf("#%s has duplicate details: %s", issue.CLIRef(id), strings.Join(localDetails, ", "))
		}
		rel, err := publishedDetailPath(view, dirs, id)
		if err != nil {
			return err
		}
		if rel != "" {
			locations[id] = rel
			republish = append(republish, id)
		} else {
			if len(localDetails) == 1 && filepath.Dir(localDetails[0]) != dirs.Abs[0] {
				return fmt.Errorf("#%s has local archived details at %s but none on main; refusing active creation", issue.CLIRef(id), localDetails[0])
			}
			if err := requireInitialDetails(card); err != nil {
				return err
			}
			first = append(first, n)
		}
	}
	for _, n := range first {
		if err := runMoveDetail(ctx, stdout, stderr, &moveDetailFlags{Issue: n, IssuesDir: issuesDir, DryRun: dryRun}); err != nil {
			return err
		}
	}
	if len(republish) == 0 {
		return nil
	}
	// #348: the details' classification (the card's, mirrored) is checked
	// against main's declaration before it is published.
	read := lazyDecl(func() (repoenum.Declaration, error) { return readTrackerRepoEnums(env) })
	for _, id := range republish {
		card, _ := snap.Require(id)
		file := filepath.Join(env.root, filepath.FromSlash(locations[id]))
		data, err := os.ReadFile(file)
		if err != nil {
			continue // republishOwned reports a missing details file
		}
		if err := detailsClassification(read, card.Path, data); err != nil {
			return err
		}
	}
	if len(first) > 0 {
		if env, err = openTracker(ctx); err != nil { // finishing a creation may have moved this checkout
			return err
		}
	}
	return republishOwned(env, stdout, stderr, republish, dirs.Rel[0], dryRun)
}

// requireInitialDetails prevents first-publication recovery resurrecting closed work.
func requireInitialDetails(card tracker.Record) error {
	status, _ := issue.GetField(card.Card.Frontmatter, "status")
	if vocab.Issue().IsTerminal(status) {
		return fmt.Errorf("#%s is %s but has no published details; refusing to create an active copy", issue.CLIRef(card.ID), status)
	}
	return nil
}

// publishedDetailPath resolves one ID on a pinned main, including legacy archives.
// Plans/reviews are classified by the same family reader as resolve.
func publishedDetailPath(view *gitx.TrunkView, dirs idDirs, id string) (string, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return "", err
	}
	seen := map[string]bool{}
	var paths []string
	for _, dir := range dirs.Rel {
		files, err := view.FilesSelected(dir, func(f gitx.TreeFile) bool {
			got, _, ok := issue.ParseFilename(path.Base(f.Path))
			return ok && got == id && path.Dir(f.Path) == dir
		})
		if err != nil {
			return "", err
		}
		for _, f := range files {
			if !seen[f.Path] {
				seen[f.Path] = true
				paths = append(paths, f.Path)
			}
		}
	}
	var details []string
	for _, f := range classifyFamily(n, paths) {
		if f.Kind == kindIssue {
			details = append(details, f.Path)
		}
	}
	if len(details) > 1 {
		return "", fmt.Errorf("#%s has duplicate details on main: %s", issue.CLIRef(id), strings.Join(details, ", "))
	}
	if len(details) == 1 {
		return details[0], nil
	}
	return "", nil
}

func publishIssueCommit(ctx context.Context, stdout, stderr io.Writer, root, source, issuesDir, historyDir string) error {
	ctx = commandContext(ctx)
	tf, err := gitx.NewTrunkFileContext(ctx, root, "origin", "main")
	if err != nil {
		return err
	}
	selected, err := tf.SelectCommit(source)
	if err != nil {
		return err
	}
	roots := []string{issuesDir, envOr("WF_PLANS_DIR", "workshop/plans"), historyDir, "workshop/projects"}
	for i, path := range roots {
		roots[i], err = gitx.InsideRoot(root, path)
		if err != nil {
			return err
		}
	}
	if err := validateCommitSelection(selected, roots); err != nil {
		return err
	}
	if err := checkPublishedClassification(tf, selected, roots[0]); err != nil {
		return err
	}
	var changed []string
	for _, c := range selected.Changes {
		changed = append(changed, c.Path)
	}
	result, err := tf.PublishCommitChecked(selected, func(base, next *gitx.TrunkView) error {
		return validateAssetPublication(ctx, root, base, next, changed, issuesDir, historyDir)
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(stderr, "Source commit %s: %s", result.Source, result.Outcome)
	if result.Candidate != "" {
		fmt.Fprintf(stderr, " (%s)", result.Candidate)
	}
	fmt.Fprintln(stderr)
	fmt.Fprintln(stdout, result.Outcome)
	return nil
}

// checkPublishedClassification checks the area and severity of every issue
// file a legacy publication adds or changes (#348), against main's declaration
// (read only when some file carries a classification).
func checkPublishedClassification(tf *gitx.TrunkFile, sel gitx.SelectedCommit, issuesRoot string) error {
	read := lazyDecl(func() (repoenum.Declaration, error) {
		raw, err := tf.Read(vocab.Issue().RepoEnums.DeclarationPath)
		if err != nil {
			return repoenum.Declaration{}, err
		}
		return parseRepoEnums(raw)
	})
	for _, ch := range sel.Changes {
		if strings.Trim(ch.NewMode, "0") == "" || !strings.HasPrefix(ch.Path, issuesRoot+"/") || !strings.HasSuffix(ch.Path, ".md") {
			continue
		}
		data, err := tf.ReadAt(sel.Source, ch.Path)
		if err != nil {
			return err
		}
		if err := detailsClassification(read, ch.Path, data); err != nil {
			return err
		}
	}
	return nil
}

// detailsClassification checks one issue file's area and severity; a file
// with neither (or no frontmatter) passes.
func detailsClassification(read declReader, name string, data []byte) error {
	fm, _, err := issue.Parse(string(data))
	if err != nil {
		return nil // structure is the instance gate's concern, not this check's
	}
	areas, severity, err := frontmatterClassification(fm)
	if err == nil {
		err = checkClassification(read, areas, severity)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
