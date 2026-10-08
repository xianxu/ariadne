// abandon.go — `sdlc abandon` (#286): end an issue as wontfix or punt without
// losing its work. Started work is kept under refs/ariadne/abandoned/NNNNNN on
// the publication remote, the card goes terminal with that ref recorded (the
// owner stays as attribution), the details are archived on main, and the
// issue branch is deleted locally and remotely. Every step is recognised when
// already done, so a rerun finishes an interrupted abandon.
package main

import (
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
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

type abandonFlags struct {
	Issue                           int
	As, Reason                      string
	IssuesDir, PlansDir, HistoryDir string
}

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
	cmd.Flags().StringVar(&f.PlansDir, "plans-dir", envOr("WF_PLANS_DIR", "workshop/plans"), "directory holding plan artifacts")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", envOr("WF_HISTORY_DIR", "workshop/history"), "archive root")
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
	fmt.Fprintf(stdout, "abandoned #%s as %s\n", issueStr, status)
	return nil
}

// keepAbandonedWork commits the abandon note on the issue branch and pushes
// the tip to the archive ref, returning the record the card will carry.
func keepAbandonedWork(env *trackerEnv, stderr io.Writer, id, detailRel, as, reason, today string) (issue.Abandoned, error) {
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
	out, err := env.git("ls-remote", env.target.Remote, ref)
	if err != nil {
		return err
	}
	cur := ""
	if f := strings.Fields(out); len(f) > 0 {
		cur = f[0]
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

// archiveAbandoned archives the issue's details (and its plan artifacts on
// main) in one narrow main commit, through the landing archive's rules: the
// final details mirror the terminal card. Started work's final details are
// the kept tip's; an issue abandoned from open gets the note in main's copy.
func archiveAbandoned(env *trackerEnv, stderr io.Writer, id, detailRel string, f *abandonFlags, rec issue.Abandoned, note string) error {
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	card, err := snap.Require(id)
	if err != nil {
		return err
	}
	var final []byte
	if rec.Started() {
		if final, err = abandonedDetails(env, rec, detailRel); err != nil {
			return err
		}
	}
	base := path.Base(detailRel)
	archived := false
	prepare := func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
		live, err := view.Exists(detailRel)
		if err != nil || !live {
			return gitx.TrunkWrite{}, errors.Join(err, tracker.ErrNoChange)
		}
		content := final
		if content == nil {
			if content, err = view.Read(detailRel); err != nil {
				return gitx.TrunkWrite{}, err
			}
			fm, body, err := issue.Parse(string(content))
			if err != nil {
				return gitx.TrunkWrite{}, err
			}
			content = []byte(issue.Compose(fm, insertLogLine(body, note)))
		}
		w := gitx.TrunkWrite{Write: map[string][]byte{
			archiveDestination(f.HistoryDir, vocab.ArchiveIssues, base): mirrorTerminal(env, content, card.Raw),
		}, Delete: []string{detailRel}, ExactBytes: true}
		plans, err := view.Files(f.PlansDir)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		for _, p := range plans {
			if planArtifactBelongsToIssue(base, path.Base(p.Path)) {
				w.Write[archiveDestination(f.HistoryDir, vocab.ArchivePlans, path.Base(p.Path))] = p.Content
				w.Delete = append(w.Delete, p.Path)
			}
		}
		archived = true
		return w, nil
	}
	msg := fmt.Sprintf("#%s: issue: abandon — archive details", issue.CLIRef(id))
	err = mainPublish(env, msg, prepare, func(string, string) error {
		fresh, err := env.repo.Snapshot()
		if err != nil {
			return err
		}
		c, err := fresh.Require(id)
		if err != nil {
			return err
		}
		return requireOwnedToPublish(env, c)
	})
	if errors.Is(err, tracker.ErrNoChange) {
		return nil
	}
	if err != nil {
		return err
	}
	if archived {
		cok(stderr, fmt.Sprintf("#%s's details archived on main", issue.CLIRef(id)))
	}
	return nil
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
	// The remote copy may be any earlier push of the branch; it is deleted
	// (leased on what it holds) only when the kept tip contains it.
	out, err := env.git("ls-remote", "--heads", env.target.Remote, "refs/heads/"+rec.Branch)
	if err != nil {
		return err
	}
	if f := strings.Fields(out); len(f) > 0 {
		remoteTip := f[0]
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
