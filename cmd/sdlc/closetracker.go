// closetracker.go — close for tracker-era issues (#252). The close decision
// becomes durable in two places: an evidence commit on the issue branch (the
// details' Log line, ledgers, sidecars and project records, with the verdict
// trailers) and, bound to that commit, codecomplete on the card. A merge — from
// any clone — later selects the issue by that binding.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// gitEvidence commits a close's files on the source branch. It builds the
// commit in a temporary index from HEAD, so staged unrelated work is neither
// committed nor disturbed, and moves the branch by compare-and-swap.
type gitEvidence struct{ env *trackerEnv }

func (g gitEvidence) Prepare(spec tracker.ReceiptSpec) (string, error) {
	env := g.env
	if contains, err := env.gitTest("merge-base", "--is-ancestor", spec.ReviewedHEAD, "HEAD"); err != nil {
		return "", err
	} else if !contains {
		return "", fmt.Errorf("%s no longer contains the reviewed commit %s; re-run `sdlc close`", env.branch, shortOID(spec.ReviewedHEAD))
	}
	head, err := env.git("rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return "", err
	}
	idx, cleanup, err := tempIndexFile()
	defer cleanup()
	if err != nil {
		return "", err
	}
	withIndex := []string{"GIT_INDEX_FILE=" + idx}
	if _, err := env.gitEnv(withIndex, "read-tree", head); err != nil {
		return "", err
	}
	for _, p := range spec.EvidencePathList() {
		abs := filepath.Join(env.root, filepath.FromSlash(p))
		info, err := os.Lstat(abs)
		switch {
		case errors.Is(err, os.ErrNotExist):
			if _, err := env.gitEnv(withIndex, "update-index", "--force-remove", "--", p); err != nil {
				return "", err
			}
			continue
		case err != nil:
			return "", err
		case !info.Mode().IsRegular():
			return "", fmt.Errorf("evidence %s is not an ordinary file", p)
		}
		blob, err := env.git("hash-object", "-w", "--path", p, "--", abs)
		if err != nil {
			return "", err
		}
		mode := "100644"
		if info.Mode().Perm()&0o111 != 0 {
			mode = "100755"
		}
		if _, err := env.gitEnv(withIndex, "update-index", "--add", "--cacheinfo", mode+","+blob+","+p); err != nil {
			return "", err
		}
	}
	tree, err := env.gitEnv(withIndex, "write-tree")
	if err != nil {
		return "", err
	}
	if headTree, err := env.git("rev-parse", head+"^{tree}"); err != nil {
		return "", err
	} else if headTree == tree {
		return "", errors.New("the close changed none of its evidence files; nothing to record")
	}
	args := []string{"commit-tree", tree, "-p", head, "-m", spec.EvidenceMessage}
	// Honour the repository's signing policy (unset exits 1: no signing).
	if value, err := env.git("config", "--bool", "--get", "commit.gpgsign"); err == nil && value == "true" {
		args = append(args, "-S")
	}
	return env.git(args...)
}

func (g gitEvidence) Apply(spec tracker.ReceiptSpec, commit string) error {
	env := g.env
	parent, err := env.git("rev-parse", "--verify", commit+"^")
	if err != nil {
		return err
	}
	if _, err := env.git("update-ref", "-m", "sdlc close", spec.SourceBranch, commit, parent); err != nil {
		return err
	}
	// The worktree already holds these bytes; only their index entries follow HEAD.
	_, err = env.git(append([]string{"reset", "-q", "--"}, spec.EvidencePathList()...)...)
	return err
}

func (g gitEvidence) Applied(spec tracker.ReceiptSpec, commit string) (bool, error) {
	return g.env.gitTest("merge-base", "--is-ancestor", commit, spec.SourceBranch)
}

func tempIndexFile() (string, func(), error) {
	dir, err := os.MkdirTemp("", "sdlc-evidence-*")
	if err != nil {
		return "", func() {}, err
	}
	return filepath.Join(dir, "index"), func() { os.RemoveAll(dir) }, nil
}

// closeEvidencePaths lists the files a close records: the details, every
// changed issue-family artifact in the plans directory (plan, gate ledgers,
// review sidecars) and the project records edited in this repository.
func closeEvidencePaths(env *trackerEnv, r closeResult, plansDir string) ([]string, error) {
	rel := func(p string) (string, error) {
		abs, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		out, err := filepath.Rel(env.root, abs)
		if err != nil || strings.HasPrefix(out, "..") {
			return "", fmt.Errorf("%s is outside the checkout", p)
		}
		return filepath.ToSlash(out), nil
	}
	details, err := rel(r.issuePath)
	if err != nil {
		return nil, err
	}
	paths := []string{details}
	stem := strings.TrimSuffix(filepath.Base(r.issuePath), ".md")
	if plans, err := rel(plansDir); err == nil {
		changed, err := env.git("status", "--porcelain", "-z", "--untracked-files=all", "--", path.Join(plans, stem+"-*"))
		if err != nil {
			return nil, err
		}
		for _, entry := range strings.Split(changed, "\x00") {
			if len(entry) > 3 {
				paths = append(paths, entry[3:])
			}
		}
	}
	for _, e := range r.projectEdits {
		if e.repoDir != "" && e.repoDir != r.repoTop {
			continue // a peer repository's record rides its own commit
		}
		p, err := rel(e.path)
		if err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	seen := map[string]bool{}
	var out []string
	for _, p := range paths {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out, nil
}

// publishTrackerClose records a finalized close: the evidence commit, then
// codecomplete bound to it on the card. On FIX-THEN-SHIP the receipt is stored
// unstarted, so the agent's fixes land first and the evidence commit — the
// anchor merge checks the reviewed state against — comes last.
func publishTrackerClose(stdout, stderr io.Writer, f *closeFlags, r closeResult, review reviewResult) error {
	env, err := openTracker(commandContext(f.Context))
	if err != nil {
		return err
	}
	if env.branch == "" {
		return errors.New("a tracker-era close needs a checked-out branch for its evidence commit")
	}
	id, _, ok := issue.ParseFilename(filepath.Base(r.issuePath))
	if !ok {
		return fmt.Errorf("%s is not an issue filename", r.issuePath)
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	card, ok := snap.Card(id)
	if !ok {
		return fmt.Errorf("no card #%s on the tracker", id)
	}
	mainView, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	reviewed, err := env.git("rev-parse", "--verify", review.Head+"^{commit}")
	if err != nil {
		return fmt.Errorf("resolve the reviewed commit: %w", err)
	}
	paths, err := closeEvidencePaths(env, r, f.plansDir())
	if err != nil {
		return err
	}
	details, err := os.ReadFile(r.issuePath)
	if err != nil {
		return err
	}
	blob, err := gitx.WriteBlob(env.ctx, env.root, details)
	if err != nil {
		return err
	}
	actual := f.Actual
	if actual == "" {
		actual = issue.ActualNotApplicableSentinel
	}
	message := fmt.Sprintf("#%s: close\n\n%s\nClose-Actual: %s", issue.CLIRef(id), strings.Join(reviewTrailers(review), "\n"), actual)
	spec := tracker.ReceiptSpec{Token: operationToken("close"), Repository: env.target.Repository, IssueID: id,
		CardPath: card.Path, SourcePath: paths[0], DestinationPath: paths[0], SourceBranch: env.branchRef(),
		SourceBase: reviewed, SourceHEAD: reviewed, ReviewedHEAD: reviewed, SourceBlob: blob, CardOID: card.BlobOID,
		TrackerBase: snap.Ref(), MainBase: mainView.Ref(), Source: tracker.LocalSource,
		EvidenceMessage: message, EvidencePaths: strings.Join(paths, "\n")}
	c, err := tracker.NewCompletion(spec)
	if err != nil {
		return err
	}
	receipts, err := env.receipts()
	if err != nil {
		return err
	}
	if review.Verdict == judge.VerdictFixThenShip {
		if err := receipts.Save(c.Receipt()); err != nil {
			return err
		}
		cwarn(stderr, fmt.Sprintf("FIX-THEN-SHIP: commit the fixes on %s, then `sdlc issue recovery reconcile --issue %s` records this close\n"+
			"      (the evidence commit lands after your fixes; codecomplete is published bound to it)", env.branch, issue.CLIRef(id)))
		return nil
	}
	op := tracker.NewCompletionOp(env.ctx, env.repo, env.branchRef(), gitEvidence{env}, time.Now().Format("2006-01-02"))
	final, err := tracker.Drive(c.Receipt(), tracker.CompletionStepper, op, receipts)
	if err != nil {
		return fmt.Errorf("%w\n      the close is recorded; finish it with `sdlc issue recovery reconcile --issue %s`", err, issue.CLIRef(id))
	}
	cok(stderr, fmt.Sprintf("#%s codecomplete on %s, bound to evidence commit %s", id, env.branch, shortOID(tracker.EvidenceCommit(final))))
	fmt.Fprintln(stdout, tracker.EvidenceCommit(final))
	return nil
}

