package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

type resumeFlags struct{ Name, Attempt, Rev string }

func newLocalCheckResumeCmd() *cobra.Command {
	var f resumeFlags
	c := &cobra.Command{Use: "resume", Short: "Rerun the registered flakes of one saved suite failure on its original pushed commit", Long: renderLong("local-check"), Args: cobra.NoArgs, SilenceUsage: true, RunE: func(c *cobra.Command, _ []string) error {
		if f.Name == "" {
			return errors.New("local-check resume: give --name suite")
		}
		if f.Name != suiteCheckName {
			return errors.New("local-check resume: only suite is supported")
		}
		if f.Attempt == "" {
			return errors.New("local-check resume: give --attempt")
		}
		if !resumeValidID(f.Attempt) {
			return errors.New("local-check resume requires a safe parent attempt ID")
		}
		d, err := resumeCommandDeps(commandContext(c.Context()), c.ErrOrStderr())
		if err != nil {
			return err
		}
		if out, err := d.Git("fetch", "--quiet", "--prune", "origin"); err != nil {
			return fmt.Errorf("resume refresh origin: %w: %s", err, out)
		}
		rev, err := d.Git("rev-parse", "--verify", f.Rev+"^{commit}")
		if err != nil {
			return err
		}
		sha := strings.TrimSpace(rev)
		refs, err := d.Git("for-each-ref", "--contains", sha, "--format=%(refname)", "refs/remotes/origin/")
		if err != nil {
			return err
		}
		if strings.TrimSpace(refs) == "" {
			return errors.New("local-check resume: original commit must remain pushed on origin")
		}
		check, err := declaredCheck(c.Context(), d.Git, suiteCheckName, sha)
		if err != nil {
			return err
		}
		rec, err := resumeSavedSuite(commandContext(c.Context()), d, check, sha, f.Attempt)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.ErrOrStderr(), "==> local check suite: %s recorded on %s\n", rec.statusDescription(), abbrevSHA(sha))
		return nil
	}}
	c.Flags().StringVar(&f.Name, "name", "", "suite (required)")
	c.Flags().StringVar(&f.Attempt, "attempt", "", "original failed suite attempt ID (required)")
	c.Flags().StringVar(&f.Rev, "rev", "HEAD", "original pushed commit")
	return c
}
func resumeCommandDeps(ctx context.Context, stderr io.Writer) (localRunDeps, error) {
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return localRunDeps{}, err
	}
	git := func(args ...string) (string, error) {
		out, err := mergeRunner.GitInDir(root, args...)
		return string(out), err
	}
	repo, err := detectRepo()
	if err != nil {
		return localRunDeps{}, err
	}
	return localRunDeps{GH: ghClient, Repo: repo, Root: root, Git: git, CleanupGit: localCheckCleanupGit(ctx, mergeRunner, root), Stderr: stderr, FlakeRegistry: gateFlakeRegistry(root)}, nil
}
func newLocalCheckReconcileCmd(authority landingDeps) *cobra.Command {
	var parent, digest, note string
	c := &cobra.Command{Use: "reconcile", Short: "Record terminal interruption after verifying the resume owner and descendants stopped", Long: renderLong("local-check"), Args: cobra.NoArgs, SilenceUsage: true, RunE: func(c *cobra.Command, _ []string) error {
		if parent == "" {
			return errors.New("local-check reconcile: give --attempt")
		}
		if !resumeValidID(parent) {
			return errors.New("local-check reconcile requires a safe parent attempt ID")
		}
		if !resumeValidDigest(digest) {
			return errors.New("local-check reconcile: --link-digest requires the exact raw link SHA256")
		}
		if !resumeText(note, 4096) {
			return errors.New("local-check reconcile: --note must attest that descendants stopped")
		}
		address, slot, _ := authority.caller()
		tree, _ := authority.couchTree()
		if err := landingline.Authorize(slot, tree); err != nil {
			return err
		}
		d, err := resumeCommandDeps(commandContext(c.Context()), c.ErrOrStderr())
		if err != nil {
			return err
		}
		actor := fmt.Sprintf("uid=%d workspace=%s", os.Getuid(), address)
		if err := reconcileSavedResume(commandContext(c.Context()), d, parent, digest, note, actor); err != nil {
			return err
		}
		fmt.Fprintf(c.ErrOrStderr(), "==> resume %s terminally interrupted; surviving inputs retained, no cleanup attempted; fix/advance code and run a suite on a new head\n", parent)
		return nil
	}}
	c.Flags().StringVar(&parent, "attempt", "", "parent attempt ID")
	c.Flags().StringVar(&digest, "link-digest", "", "SHA256 of immutable raw link bytes")
	c.Flags().StringVar(&note, "note", "", "operator attestation that no descendants remain (not automatically discovered)")
	return c
}
