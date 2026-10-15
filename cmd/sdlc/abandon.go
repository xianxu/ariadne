// abandon.go — `sdlc abandon` (#286): end an issue as wontfix or punt without
// losing its work. Started work is kept under refs/ariadne/abandoned/NNNNNN on
// the publication remote, the card goes terminal with that ref recorded (the
// owner stays as attribution), the details are archived on main, and the
// issue branch is deleted locally and remotely. Every step is recognised when
// already done, so a rerun finishes an interrupted abandon.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

type abandonFlags struct {
	Issue      int
	As, Reason string
	IssuesDir  string
}

// The archive's plans and history roots come from the environment, as the
// reopen that restores from them reads them (one source, BR-10).
func plansDir() string   { return envOr("WF_PLANS_DIR", "workshop/plans") }
func historyDir() string { return envOr("WF_HISTORY_DIR", "workshop/history") }

func NewAbandonCmd() *cobra.Command {
	f := abandonFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:           "abandon --issue N --as wontfix|punt --reason TEXT",
		Short:         "End an issue as wontfix or punt, keeping its work under an archive ref",
		Long:          "Placeholder — replaced by helptext.MustGet(\"abandon\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			guardSpineRepo(cmd.ErrOrStderr())
			return runAbandon(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "issue ID to abandon")
	cmd.Flags().StringVar(&f.As, "as", "", "the terminal status: wontfix (rejected) or punt (deferred)")
	cmd.Flags().StringVar(&f.Reason, "reason", "", "one line for the issue's ## Log: why it ends here")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue details")
	return cmd
}

// abandonEvent is the lifecycle model's event for each terminal --as.
var abandonEvent = map[string]string{"wontfix": "abandon", "punt": "defer"}

// abandonDecision is the card after abandon: the model's edge from its status
// (by abandon's event), the record, and today. The owner is kept.
func abandonDecision(card []byte, as, today string, rec issue.Abandoned) ([]byte, error) {
	fm, _, err := issue.Parse(string(card))
	if err != nil {
		return nil, err
	}
	prev, _ := issue.GetField(fm, "status")
	t := vocab.Issue().TransitionForEvent(prev, abandonEvent[as])
	if t == nil || t.To != as {
		return nil, fmt.Errorf("a %s issue cannot be abandoned as %s (legal: %s)", prev, as, strings.Join(vocab.Issue().LegalTransitions(prev), ", "))
	}
	out, err := issue.SetCardField(card, "status", as)
	if err == nil {
		out, err = issue.SetCardField(out, "updated", today)
	}
	if err == nil {
		out, err = issue.SetCardAbandoned(out, &rec)
	}
	return out, err
}

// abandonNote is the ## Log line abandon files.
func abandonNote(as, reason, today string) string {
	return fmt.Sprintf("- %s: abandoned (%s): %s", today, as, reason)
}

func runAbandon(ctx context.Context, stdout, stderr io.Writer, f *abandonFlags) error {
	ctx = commandContext(ctx)
	if f.Issue <= 0 {
		return errors.New("--issue N is required")
	}
	if _, ok := abandonEvent[f.As]; !ok {
		return fmt.Errorf("--as must be wontfix (rejected) or punt (deferred), not %q", f.As)
	}
	reason := strings.TrimSpace(f.Reason)
	if reason == "" || strings.ContainsAny(reason, "\r\n") {
		return errors.New("--reason is required: one line saying why the issue ends here")
	}
	if tracked, err := repositoryTracked(ctx, "."); err != nil {
		return err
	} else if !tracked {
		return errors.New("`sdlc abandon` runs in issue tracker repositories; here, use `sdlc issue set-status`")
	}
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%06d", f.Issue)
	issueStr := issue.CLIRef(id)
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	card, err := snap.Require(id)
	if err != nil {
		return err
	}
	if dirty, err := cleanTree(env); err != nil {
		return err
	} else if dirty != "" {
		return fmt.Errorf("#%s: abandon needs a clean tree; commit or discard:\n%s", issueStr, dirty)
	}
	detailRel := path.Join(f.IssuesDir, path.Base(card.Path))
	today := time.Now().Format("2006-01-02")
	status, _ := issue.GetField(card.Card.Frontmatter, "status")
	rec, recorded, err := issue.CardAbandoned(card.Raw)
	if err != nil {
		return fmt.Errorf("card #%s: %w", id, err)
	}
	resume := recorded && vocab.Issue().IsTerminal(status)

	if !resume {
		if vocab.Issue().IsTerminal(status) {
			return fmt.Errorf("#%s is already %s; nothing to abandon", issueStr, status)
		}
		if err := requireCardOwnership(env, card); err != nil {
			return err
		}
		if _, err := abandonDecision(card.Raw, f.As, today, issue.Abandoned{}); err != nil {
			return fmt.Errorf("#%s: %w", issueStr, err)
		}
		if !vocab.Issue().IsOpen(status) {
			// Started work: this checkout is on the branch being abandoned.
			if env.branch == "" || env.onRest() || env.branch == "main" {
				return fmt.Errorf("#%s has started work: run abandon from its issue branch (this checkout is on %s)", issueStr, valueOr(env.branch, "a detached HEAD"))
			}
			if rec, err = keepAbandonedWork(env, stderr, id, detailRel, f.As, reason, today); err != nil {
				return err
			}
		}
		next, err := abandonDecision(card.Raw, f.As, today, rec)
		if err != nil {
			return fmt.Errorf("#%s: %w", issueStr, err)
		}
		err = uncertainCardWrite(cardPublish(env, card, next, operationToken("abandon"), nil, nil), fmt.Sprintf("sdlc abandon --issue %s --as %s --reason …", issueStr, f.As))
		invalidateIssueRecords(env.ctx)
		if err != nil {
			return err
		}
		cok(stderr, fmt.Sprintf("#%s is %s; the owner stays as attribution", issueStr, f.As))
		status = f.As
	} else {
		// The owner stays as attribution: only that workspace finishes it.
		if own, recorded, _, err := ownership(env, card); err != nil {
			return err
		} else if own == issue.OwnershipForeign {
			return fmt.Errorf("#%s was abandoned by %s; only that workspace finishes it", issueStr, describeClaimant(recorded))
		}
		cinfo(stderr, fmt.Sprintf("#%s is already %s by abandon; finishing its remaining steps", issueStr, status))
	}

	if err := archiveAbandoned(env, stderr, id, detailRel, f, rec, abandonNote(status, reason, today)); err != nil {
		return fmt.Errorf("%w\n      the card is %s; rerun `sdlc abandon --issue %s --as %s --reason …` to finish", err, status, issueStr, status)
	}
	if rec.Started() {
		if err := dropAbandonedBranch(env, stderr, rec); err != nil {
			return fmt.Errorf("%w\n      the work is kept at %s; rerun `sdlc abandon --issue %s --as %s --reason …` to finish", err, rec.Ref, issueStr, status)
		}
	}
	if w := unpinReviewed(env.root, id); w != "" { // #304 D4: the issue ended
		cwarn(stderr, w)
	}
	fmt.Fprintf(stdout, "abandoned #%s as %s\n", issueStr, status)
	return nil
}

// keepAbandonedWork commits the abandon note on the issue branch and pushes
// the tip to the archive ref, returning the record the card will carry.
func keepAbandonedWork(env *trackerEnv, stderr io.Writer, id, detailRel, as, reason, today string) (issue.Abandoned, error) {
	// Check admission before even the abandon note or archive-ref push can expose
	// unpublished branch evidence. archiveIssueOnMain repeats it on each retry.
	tip, err := env.git("rev-parse", "HEAD")
	if err != nil {
		return issue.Abandoned{}, err
	}
	raw, err := env.main.ViewOf(tip).Read(detailRel)
	if err != nil {
		return issue.Abandoned{}, err
	}
	assets, err := assetFilesFromView(env.main.ViewOf(tip), detailRel, id, raw)
	if err != nil {
		return issue.Abandoned{}, err
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return issue.Abandoned{}, err
	}
	published, err := publishedAbandonAssets(view, detailRel, id)
	if err != nil {
		return issue.Abandoned{}, err
	}
	if err := requireNewArchiveAssets(env, view, published, assets); err != nil {
		return issue.Abandoned{}, err
	}
	issueStr := issue.CLIRef(id)
	msg := fmt.Sprintf("#%s: log: abandon (%s)", issueStr, as)
	if subject, err := env.git("log", "-1", "--format=%s"); err != nil {
		return issue.Abandoned{}, err
	} else if subject != msg { // a rerun finds its note already committed
		abs := filepath.Join(env.root, filepath.FromSlash(detailRel))
		raw, err := os.ReadFile(abs)
		if err != nil {
			return issue.Abandoned{}, fmt.Errorf("#%s: the issue branch has no details at %s: %w", issueStr, detailRel, err)
		}
		fm, body, err := issue.Parse(string(raw))
		if err != nil {
			return issue.Abandoned{}, err
		}
		if err := os.WriteFile(abs, []byte(issue.Compose(fm, insertLogLine(body, abandonNote(as, reason, today)))), 0o644); err != nil {
			return issue.Abandoned{}, err
		}
		if err := commitOnly(env, msg, detailRel); err != nil {
			return issue.Abandoned{}, err
		}
	}
	head, err := env.git("rev-parse", "HEAD")
	if err != nil {
		return issue.Abandoned{}, err
	}
	rec := issue.Abandoned{Ref: issue.AbandonedRef(id), Branch: env.branch, Head: head}
	if err := pushArchiveRef(env, rec.Ref, head); err != nil {
		return issue.Abandoned{}, fmt.Errorf("#%s: keeping the work at %s failed: %w — nothing else changed; rerun", issueStr, rec.Ref, err)
	}
	cok(stderr, fmt.Sprintf("#%s's work kept at %s (%s)", issueStr, rec.Ref, shortOID(head)))
	return rec, nil
}

// pushArchiveRef points ref on the publication remote at head. The lease
// accepts the ref absent, already at head (a rerun), or at an ancestor of
// head (an orphan an interrupted reopen left, whose work the branch holds).
func pushArchiveRef(env *trackerEnv, ref, head string) error {
	cur, err := remoteRefTip(env.git, env.target.Remote, ref)
	if err != nil {
		return err
	}
	switch {
	case cur == head:
		return nil
	case cur != "":
		if _, err := env.git("fetch", "-q", env.target.Remote, ref); err != nil {
			return err
		}
		if inBranch, err := env.ancestorOf(cur, head); err != nil {
			return err
		} else if !inBranch {
			return fmt.Errorf("%s already holds %s, which this branch does not contain; inspect it before abandoning", ref, shortOID(cur))
		}
	}
	_, err = env.git("push", "-q", "--force-with-lease="+ref+":"+cur, env.target.Remote, head+":"+ref)
	return err
}

// archiveAbandoned archives an abandoned issue on main. Started work's final
// details and plans are the kept tip's; an issue abandoned from open gets the
// note in main's copy. Another workspace's card is refused (requireOwnedToPublish).
func archiveAbandoned(env *trackerEnv, stderr io.Writer, id, detailRel string, f *abandonFlags, rec issue.Abandoned, note string) error {
	spec := mainArchive{
		Message: fmt.Sprintf("#%s: issue: abandon — archive details", issue.CLIRef(id)),
		Dirs:    archiveDirs{Issues: f.IssuesDir, Plans: plansDir(), History: historyDir()},
		Note:    note,
		Allow:   func(c tracker.Record) error { return requireOwnedToPublish(env, c) },
	}
	if rec.Started() {
		var err error
		if spec.Final, err = abandonedDetails(env, rec, detailRel); err != nil {
			return err
		}
		if spec.Plans, err = keptPlans(env, rec, path.Base(detailRel)); err != nil {
			return err
		}
		bundle, err := assetFilesFromView(env.main.ViewOf(rec.Head), detailRel, id, spec.Final)
		if err != nil {
			return err
		}
		spec.Assets = &bundle
		spec.Note = ""
	}
	return archiveIssueOnMain(env, stderr, id, detailRel, spec)
}

// abandonedDetails is the kept tip's details, fetching the archive ref when
// this clone no longer has the commit (a rerun elsewhere).
func abandonedDetails(env *trackerEnv, rec issue.Abandoned, detailRel string) ([]byte, error) {
	if have, err := env.gitTest("cat-file", "-e", rec.Head+"^{commit}"); err != nil {
		return nil, err
	} else if !have {
		if _, err := env.git("fetch", "-q", env.target.Remote, rec.Ref); err != nil {
			return nil, err
		}
	}
	raw, err := env.gitRaw(nil, "show", rec.Head+":"+detailRel)
	if err != nil {
		return nil, fmt.Errorf("the kept tip %s has no %s: %w", shortOID(rec.Head), detailRel, err)
	}
	return raw, nil
}

// keptPlans is the kept tip's plan artifacts for the issue, by basename.
func keptPlans(env *trackerEnv, rec issue.Abandoned, base string) (map[string][]byte, error) {
	listed, err := env.git("ls-tree", "--name-only", rec.Head, plansDir()+"/")
	if err != nil {
		return nil, err
	}
	out := map[string][]byte{}
	for _, p := range strings.Fields(listed) {
		if !planArtifactBelongsToIssue(base, path.Base(p)) {
			continue
		}
		content, err := env.gitRaw(nil, "show", rec.Head+":"+p)
		if err != nil {
			return nil, err
		}
		out[path.Base(p)] = content
	}
	return out, nil
}

// mirrorTerminal brings details to the terminal card through the archive's
// one projection; details it cannot refresh are archived as they are.
func mirrorTerminal(env *trackerEnv, content, card []byte) []byte {
	oid, err := issue.MirrorBaselineOID(content)
	if err != nil {
		return content
	}
	baseline, err := env.gitRaw(nil, "cat-file", "blob", oid)
	if err != nil {
		return content
	}
	return archivedDetails(content, baseline, card)
}

// dropAbandonedBranch deletes the abandoned branch from the remote (leased on
// the kept tip), returns this checkout to rest, and deletes the local branch.
// Each step is skipped when already done.
func dropAbandonedBranch(env *trackerEnv, stderr io.Writer, rec issue.Abandoned) error {
	var restingView *gitx.TrunkView
	if env.branch == rec.Branch {
		restingHead, err := env.git("rev-parse", env.resting)
		if err != nil {
			return err
		}
		restingView = env.main.ViewOf(restingHead)
		if err := verifyAssetMaterialization(env.ctx, env.root, restingView, envOr("WF_ISSUES_DIR", "workshop/issues")); err != nil {
			return err
		}
	}
	// The remote copy may be any earlier push of the branch; it is deleted
	// (leased on what it holds) only when the kept tip contains it.
	remoteTip, err := remoteRefTip(env.git, env.target.Remote, "refs/heads/"+rec.Branch)
	if err != nil {
		return err
	}
	if remoteTip != "" {
		if _, err := env.git("fetch", "-q", env.target.Remote, "refs/heads/"+rec.Branch); err != nil {
			return err
		}
		if kept, err := env.ancestorOf(remoteTip, rec.Head); err != nil {
			return err
		} else if !kept {
			return fmt.Errorf("%s on %s holds %s, which the kept tip %s does not contain; inspect it before deleting", rec.Branch, env.target.Remote, shortOID(remoteTip), shortOID(rec.Head))
		}
		if err := deleteRemoteBranch(env.git, env.target.Remote, rec.Branch, remoteTip); err != nil {
			return err
		}
	}
	if env.branch == rec.Branch {
		if _, err := env.git("switch", "-q", env.resting); err != nil {
			return err
		}
		env.branch = env.resting
		if err := verifyMaterializedAssets(env.root, restingView, envOr("WF_ISSUES_DIR", "workshop/issues")); err != nil {
			return fmt.Errorf("returned to %s, but asset verification failed: %w", env.resting, err)
		}
	}
	if tip, err := env.git("for-each-ref", "--format=%(objectname)", "refs/heads/"+rec.Branch); err != nil {
		return err
	} else if tip != "" {
		if tip != rec.Head {
			return fmt.Errorf("local %s moved to %s after it was kept at %s; inspect it before deleting", rec.Branch, shortOID(tip), shortOID(rec.Head))
		}
		if _, err := env.git("branch", "-q", "-D", rec.Branch); err != nil {
			return err
		}
	}
	cok(stderr, fmt.Sprintf("%s deleted here and on %s; this checkout is on %s", rec.Branch, env.target.Remote, env.resting))
	return nil
}

// restoreAbandoned is a reopen's work for an abandoned issue (#286 D9), run
// before set-status's card write: the branch is recreated at the kept tip,
// main is merged in, the archived details and plans move back out of the
// archive, and the branch is pushed. Each step is recognised on rerun; a
// merge conflict stops here with the card still terminal.
func restoreAbandoned(env *trackerEnv, stderr io.Writer, card tracker.Record, rec issue.Abandoned, issuesDir string) error {
	issueStr := issue.CLIRef(card.ID)
	own, recorded, _, err := ownership(env, card)
	if err != nil {
		return err
	}
	if own == issue.OwnershipForeign {
		return fmt.Errorf("#%s was abandoned by %s; only that workspace reopens it (reassignment is `sdlc reclaim`)", issueStr, describeClaimant(recorded))
	}
	if merging, err := env.gitTest("rev-parse", "-q", "--verify", "MERGE_HEAD"); err != nil {
		return err
	} else if merging {
		return fmt.Errorf("#%s: a merge of main is in progress on %s; resolve and commit it, then rerun", issueStr, rec.Branch)
	}
	if dirty, err := cleanTree(env); err != nil {
		return err
	} else if dirty != "" {
		return fmt.Errorf("#%s: reopening restores a branch and needs a clean tree; commit or discard:\n%s", issueStr, dirty)
	}
	if env.branch != rec.Branch && !env.onRest() {
		return fmt.Errorf("#%s: reopen from the resting branch (%s) or from %s, not %s", issueStr, env.resting, rec.Branch, valueOr(env.branch, "a detached HEAD"))
	}
	// 1–2: locate the kept tip without materializing its files.
	createBranch := false
	branchHead := ""
	if tip, err := env.git("for-each-ref", "--format=%(objectname)", "refs/heads/"+rec.Branch); err != nil {
		return err
	} else if tip == "" {
		if _, err := env.git("fetch", "-q", env.target.Remote, rec.Ref); err != nil {
			return fmt.Errorf("#%s: fetching the kept work %s failed: %w", issueStr, rec.Ref, err)
		}
		if fetched, err := env.git("rev-parse", "FETCH_HEAD"); err != nil {
			return err
		} else if fetched != rec.Head {
			return fmt.Errorf("#%s: %s is at %s, not the recorded %s; inspect it before reopening", issueStr, rec.Ref, shortOID(fetched), shortOID(rec.Head))
		}
		createBranch = true
		branchHead = rec.Head
	} else if has, err := env.ancestorOf(rec.Head, "refs/heads/"+rec.Branch); err != nil {
		return err
	} else if !has {
		return fmt.Errorf("#%s: local %s does not contain the kept tip %s; inspect it before reopening", issueStr, rec.Branch, shortOID(rec.Head))
	}
	if branchHead == "" {
		branchHead, err = env.git("rev-parse", "refs/heads/"+rec.Branch)
		if err != nil {
			return err
		}
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	base := path.Base(card.Path)
	archivedRel := archiveDestination(historyDir(), vocab.ArchiveIssues, base)
	liveRel := path.Join(issuesDir, base)
	// Switching to kept work and merging main each materialize source paths.
	// Check all participating attributes, including the later active destination,
	// before creating a branch, switching, or starting a merge.
	for _, source := range []*gitx.TrunkView{env.main.ViewOf(branchHead), view} {
		if err := verifyAssetMaterialization(env.ctx, env.root, source, issuesDir, branchHead, view.Ref()); err != nil {
			return err
		}
		paths, err := assetMaterializationPaths(env.root, source, issuesDir)
		if err != nil {
			return err
		}
		prefix := path.Join(path.Dir(archivedRel), "assets", card.ID) + "/"
		for _, p := range paths {
			if !strings.HasPrefix(p, prefix) {
				continue
			}
			dst := path.Join(path.Dir(liveRel), "assets", card.ID, strings.TrimPrefix(p, prefix))
			for _, ref := range []string{"", branchHead, view.Ref()} {
				if err := verifyAssetAttributesAt(env.ctx, env.root, dst, ref); err != nil {
					return err
				}
			}
		}
	}
	if createBranch {
		if _, err := env.git("branch", "-q", "--no-track", rec.Branch, rec.Head); err != nil {
			return err
		}
	}
	if env.branch != rec.Branch {
		if _, err := env.git("switch", "-q", rec.Branch); err != nil {
			return err
		}
		env.branch = rec.Branch
	}
	if err := verifyMaterializedAssets(env.root, env.main.ViewOf(branchHead), issuesDir); err != nil {
		return err
	}
	// 3: merge main, which carries the archive commit.
	archiveCommit, err := env.git("rev-list", "-1", view.Ref(), "--", archivedRel)
	if err != nil {
		return err
	}
	if archiveCommit != "" {
		if merged, err := env.ancestorOf(archiveCommit, "HEAD"); err != nil {
			return err
		} else if !merged {
			if _, err := env.git("merge", "-q", "--no-commit", "--no-edit", view.Ref()); err != nil {
				if rerr := resolveArchiveConflicts(env, base, issuesDir, plansDir()); rerr != nil {
					return fmt.Errorf("#%s: merging main into %s stopped on a conflict: %v\n      resolve and commit the merge (for the issue's details, keep main's archived copy), then rerun `sdlc issue set-status working --issue %s`; the card is unchanged", issueStr, rec.Branch, rerr, issueStr)
				}
			}
			pending, err := env.gitTest("rev-parse", "-q", "--verify", "MERGE_HEAD")
			if err != nil {
				return err
			}
			if pending {
				if err := removeDuplicateReopenAssets(env, base, issuesDir); err != nil {
					return err
				}
				if err := verifyIndexedAssetBundles(env, issuesDir); err != nil {
					return err
				}
				if _, err := env.git("commit", "-q", "--no-verify", "--no-edit"); err != nil {
					return err
				}
			} else if err := verifyMaterializedAssets(env.root, view, issuesDir); err != nil {
				return err
			}
		}
	}
	// 4: the details and plans come back out of the archive.
	if err := unarchiveIssue(env, base, issuesDir, fmt.Sprintf("#%s: reopen: restore from %s", issueStr, rec.Ref)); err != nil {
		return err
	}
	// 5: the branch on the remote, before the card says it is back.
	if err := pushIssueBranch(env, rec.Branch); err != nil {
		return fmt.Errorf("#%s: pushing the restored %s failed: %w — the card is unchanged; rerun", issueStr, rec.Branch, err)
	}
	cok(stderr, fmt.Sprintf("#%s restored on %s from %s and pushed", issueStr, rec.Branch, rec.Ref))
	return nil
}

// unarchiveIssue moves an issue's archived details and plans at HEAD back to
// their live paths in one commit (#286 D9, #305), carrying any working-tree
// edit to them; a no-op once they are out.
func unarchiveIssue(env *trackerEnv, base, issuesDir, message string) error {
	archivedRel := archiveDestination(historyDir(), vocab.ArchiveIssues, base)
	liveRel := path.Join(issuesDir, base)
	id := issueIDPrefix(base)
	if archived, err := env.has("HEAD", archivedRel); err != nil {
		return err
	} else if archived {
		moves := [][2]string{{archivedRel, liveRel}}
		head, err := env.git("rev-parse", "HEAD")
		if err != nil {
			return err
		}
		raw, err := env.main.ViewOf(head).Read(archivedRel)
		if err != nil {
			return err
		}
		bundle, err := assetFilesFromView(env.main.ViewOf(head), archivedRel, id, raw)
		if err != nil {
			return err
		}
		materialized, err := readAssetFiles(env.root, archivedRel, id)
		if err != nil {
			return err
		}
		actual, err := issueassets.BuildBundle(id, issueAssetBody(raw), materialized)
		if err != nil {
			return fmt.Errorf("restored archive assets: %w; restore the archived bundle from %s before retrying", err, head)
		}
		if !assetBundlesEqual(bundle, actual) {
			return fmt.Errorf("restored archive assets differ from %s; restore the archived bundle from that commit before retrying", head)
		}
		for rel := range bundle.Files {
			moves = append(moves, [2]string{path.Join(path.Dir(archivedRel), rel), path.Join(path.Dir(liveRel), rel)})
		}
		plans, err := env.git("ls-tree", "--name-only", "HEAD", vocab.ArchiveSubdir(historyDir(), vocab.ArchivePlans)+"/")
		if err != nil {
			return err
		}
		for _, p := range strings.Fields(plans) {
			if planArtifactBelongsToIssue(base, path.Base(p)) {
				moves = append(moves, [2]string{p, path.Join(plansDir(), path.Base(p))})
			}
		}
		// Validate every asset destination before any restore mutation.
		for _, m := range moves {
			if strings.HasPrefix(m[1], path.Join(issuesDir, "assets", id)+"/") {
				if err := verifyAssetAttributes(env.ctx, env.root, m[1]); err != nil {
					return err
				}
				if err := assetDirectories(env.root, path.Dir(m[1]), false); err != nil && !os.IsNotExist(err) {
					return err
				}
			}
			if live, err := env.has("HEAD", m[1]); err != nil {
				return err
			} else if live {
				if strings.Contains(m[1], "/assets/") {
					current, err := env.main.ViewOf(head).Read(m[1])
					if err != nil {
						return err
					}
					want, err := env.main.ViewOf(head).Read(m[0])
					if err != nil {
						return err
					}
					if !bytes.Equal(current, want) {
						return fmt.Errorf("asset destination %s differs from the archive: reconcile that file before reopening", m[1])
					}
				}
			}
		}
		var paths []string
		for _, m := range moves {
			if live, err := env.has("HEAD", m[1]); err != nil {
				return err
			} else if live {
				if _, err := env.git("rm", "-q", "--", m[1]); err != nil { // the archived copy is newer
					return err
				}
			}
			if strings.HasPrefix(m[1], path.Join(issuesDir, "assets", id)+"/") {
				if err := assetDirectories(env.root, path.Dir(m[1]), true); err != nil {
					return err
				}
			}
			if err := os.MkdirAll(filepath.Join(env.root, filepath.FromSlash(path.Dir(m[1]))), 0o755); err != nil {
				return err // the merge may have removed the last file in it
			}
			if _, err := env.git("mv", m[0], m[1]); err != nil {
				return err
			}
			paths = append(paths, m[0], m[1])
		}
		if err := verifyAssetBundleOnDisk(env.ctx, env.root, liveRel, id, raw, bundle); err != nil {
			return err
		}
		if _, err := env.git(append([]string{"commit", "-q", "--no-verify", "-m", message, "--"}, paths...)...); err != nil {
			return err
		}
	}
	return nil
}

// finishReopen commits the card mirror the reopen's card write refreshed (as
// close commits its own), pushes the branch, and drops any archive ref, so a
// reopen ends clean, on origin, with nothing left in the archive.
func finishReopen(r *reopenedIssue, stderr io.Writer, issuesDir string) {
	env := r.env
	detailRel := path.Join(issuesDir, path.Base(r.cardPath))
	if dirty, err := env.git("status", "--porcelain", "--", detailRel); err != nil {
		cwarn(stderr, fmt.Sprintf("reopen: the details mirror was not checked: %v", err))
	} else if dirty != "" {
		if err := commitOnly(env, mirrorSubject(r.id), detailRel); err != nil {
			cwarn(stderr, fmt.Sprintf("reopen: the refreshed details mirror was not committed: %v", err))
		}
	}
	boundaryPush(env, stderr, "reopen")
	if r.rec.Ref != "" {
		dropArchiveRef(env, stderr, r.rec)
	}
}

func mirrorSubject(id string) string {
	return fmt.Sprintf("#%s: mirror working card", issue.CLIRef(id))
}

// dropArchiveRef deletes a reopened issue's archive ref, leased on the kept
// tip. Best-effort: a leftover ref is harmless (its work is on the branch) and
// a later abandon overwrites it.
func dropArchiveRef(env *trackerEnv, stderr io.Writer, rec issue.Abandoned) {
	if _, err := env.git("push", "-q", "--force-with-lease="+rec.Ref+":"+rec.Head, env.target.Remote, ":"+rec.Ref); err != nil {
		cwarn(stderr, fmt.Sprintf("%s was not deleted: %v — harmless: its work is on the branch", rec.Ref, err))
	}
}

// resolveArchiveConflicts finishes a reopen's merge of main when the only
// conflicts are the issue's own live details and plans against main's
// archive of them: main's side (the archived copy, which holds the abandon
// note and is newer) is taken, and the next step moves it back out. Any other
// conflict is the agent's to resolve; the merge is left in progress.
func resolveArchiveConflicts(env *trackerEnv, base, issuesDir, plansDir string) error {
	out, err := env.git("diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return err
	}
	var own, other []string
	for _, p := range strings.Fields(out) {
		dir, name := path.Dir(p), path.Base(p)
		if (dir == issuesDir && name == base) || (dir == plansDir && planArtifactBelongsToIssue(base, name)) {
			own = append(own, p)
		} else {
			other = append(other, p)
		}
	}
	if len(other) > 0 {
		return fmt.Errorf("conflicts in %s", strings.Join(other, ", "))
	}
	if len(own) == 0 {
		return errors.New("the merge failed without a conflict")
	}
	if _, err := env.git(append([]string{"rm", "-q", "--"}, own...)...); err != nil {
		return err
	}
	return nil
}

// A kept branch may introduce active assets which main first saw at archive
// time. Once the merge removes the active body, remove only proven duplicate
// members before committing that intermediate tree. The archive remains the
// single index and source for the following unarchive step.
func removeDuplicateReopenAssets(env *trackerEnv, base, issuesDir string) error {
	tree, err := env.git("write-tree")
	if err != nil {
		return err
	}
	view := env.main.ViewOf(tree)
	if live, err := view.Exists(path.Join(issuesDir, base)); err != nil || live {
		return err
	}
	id := issueIDPrefix(base)
	activeDir := path.Join(issuesDir, "assets", id)
	var active []gitx.TreeFile
	_, err = view.FilesSelected(activeDir, func(f gitx.TreeFile) bool {
		active = append(active, f)
		return false
	})
	if err != nil {
		return err
	}
	if len(active) == 0 {
		return nil
	}
	archived := archiveDestination(historyDir(), vocab.ArchiveIssues, base)
	raw, err := view.Read(archived)
	if err != nil {
		return err
	}
	bundle, err := assetFilesFromView(view, archived, id, raw)
	if err != nil {
		return err
	}
	activeBundle, err := assetFilesFromView(view, path.Join(issuesDir, base), id, raw)
	if err != nil {
		return err
	}
	var paths []string
	for _, f := range active {
		member := path.Join("assets", id, path.Base(f.Path))
		want, ok := bundle.Files[member]
		if !ok || path.Dir(f.Path) != activeDir || f.Mode != "100644" || !bytes.Equal(activeBundle.Files[member], want) {
			return fmt.Errorf("active asset %s is not an exact archived duplicate; restore its archived bytes before retrying reopen", f.Path)
		}
		paths = append(paths, f.Path)
	}
	if err := verifyMaterializedAssetPaths(env.root, view, paths); err != nil {
		return err
	}
	_, err = env.git(append([]string{"rm", "-q", "--"}, paths...)...)
	return err
}
