// closetracker.go — close for tracker-era issues (#252). The close decision
// becomes durable in two places: an evidence commit on the issue branch (the
// details' Log line, ledgers, sidecars and project records, with the verdict
// trailers) and, bound to that commit, codecomplete on the card. A merge — from
// any clone — later selects the issue by that binding.
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

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// gitEvidence commits a close's files on the source branch. It builds the
// commit in a temporary index from HEAD, so staged unrelated work is neither
// committed nor disturbed, and moves the branch by compare-and-swap. A pinned
// file a later commit superseded is kept, warned on stderr and named in a
// Close-Kept trailer, so no close evidence is dropped silently.
type gitEvidence struct {
	env    *trackerEnv
	stderr io.Writer
}

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
	message := spec.EvidenceMessage
	for _, e := range spec.EvidenceEntries() {
		base, _, err := g.treeEntry(spec.ReviewedHEAD, e.Path)
		if err != nil {
			return "", err
		}
		current, mode, err := g.treeEntry(head, e.Path)
		if err != nil {
			return "", err
		}
		if e.Superseded(base, current) {
			message += "\n" + tracker.EvidenceKeptTrailer + ": " + e.Path
			cwarn(g.stderr, fmt.Sprintf("kept %s as committed after the close; it differs from the close's pinned version (check it still carries the close record)", e.Path))
		}
		if !e.Replays(base, current) {
			continue // already pinned, or edited in a later commit that must survive
		}
		if e.Blob == "" {
			if _, err := env.gitEnv(withIndex, "update-index", "--force-remove", "--", e.Path); err != nil {
				return "", err
			}
			continue
		}
		if mode != "100755" {
			mode = "100644"
		}
		if _, err := env.gitEnv(withIndex, "update-index", "--add", "--cacheinfo", mode+","+e.Blob+","+e.Path); err != nil {
			return "", err
		}
	}
	tree, err := env.gitEnv(withIndex, "write-tree")
	if err != nil {
		return "", err
	}
	// An unchanged tree is legitimate: a fix commit may already carry the pinned
	// evidence. The commit still records the verdict trailers and the anchor.
	args := []string{"commit-tree", tree, "-p", head, "-m", message}
	// Honour the repository's signing policy (unset exits 1: no signing).
	if value, err := env.git("config", "--bool", "--get", "commit.gpgsign"); err == nil && value == "true" {
		args = append(args, "-S")
	}
	return env.git(args...)
}

// treeEntry is path's blob and mode in commit ("" when absent).
func (g gitEvidence) treeEntry(commit, p string) (blob, mode string, err error) {
	entry, err := g.env.git("ls-tree", "--full-tree", commit, "--", p)
	if err != nil || entry == "" {
		return "", "", err
	}
	meta, _, _ := strings.Cut(entry, "\t")
	fields := strings.Fields(meta)
	if len(fields) != 3 || fields[1] != "blob" {
		return "", "", fmt.Errorf("%s in %s is not a file (%s)", p, shortOID(commit), entry)
	}
	return fields[2], fields[0], nil
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

// closeEvidence pins the files a close records — the details, every changed
// issue-family artifact in the plans directory (plan, gate ledgers, review
// sidecars) and the project records edited in this repository — as blobs of
// their bytes now, so a deferred evidence commit replays exactly these.
func closeEvidence(env *trackerEnv, r closeResult, plansDir string) ([]tracker.EvidenceEntry, error) {
	root := canonRoot(env.root)
	rel := func(p string) (string, error) {
		// Both sides canonical: Git's toplevel resolves symlinks, a caller's
		// cwd-relative path may not (/tmp vs /private/tmp on macOS).
		out, err := filepath.Rel(root, canonRoot(p))
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
		changed, err := env.statusEntries(env.root, "--untracked-files=all", "--", path.Join(plans, stem+"-*"))
		if err != nil {
			return nil, err
		}
		for _, entry := range changed {
			paths = append(paths, entry.Path)
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
	var entries []tracker.EvidenceEntry
	for _, p := range paths {
		if seen[p] {
			continue
		}
		seen[p] = true
		raw, err := os.ReadFile(filepath.Join(env.root, filepath.FromSlash(p)))
		switch {
		case errors.Is(err, os.ErrNotExist):
			entries = append(entries, tracker.EvidenceEntry{Path: p})
			continue
		case err != nil:
			return nil, err
		}
		blob, err := gitx.WriteBlob(env.ctx, env.root, raw)
		if err != nil {
			return nil, err
		}
		entries = append(entries, tracker.EvidenceEntry{Blob: blob, Path: p})
	}
	return entries, nil
}

// trackerClosePrep carries the verdict-independent preconditions of a
// tracker-era close, checked in computeClose before any review or write.
type trackerClosePrep struct {
	env        *trackerEnv
	card       tracker.Record
	trackerRef string
	receipts   *tracker.RecoveryReceipts
	superseded []tracker.Receipt // unstarted earlier closes this one replaces
}

// prepareTrackerClose checks what a tracker-era close needs before its review:
// a branch to commit on, the card, and no half-finished close in progress (an
// unstarted one — a FIX-THEN-SHIP never resumed — is superseded by this close).
func prepareTrackerClose(ctx context.Context, id string) (*trackerClosePrep, error) {
	env, err := openTracker(ctx)
	if err != nil {
		return nil, err
	}
	if env.branch == "" {
		return nil, errors.New("a tracker-era close commits its evidence on a branch; check out the issue branch")
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		return nil, err
	}
	card, ok := snap.Card(id)
	if !ok {
		return nil, fmt.Errorf("no card #%s on the tracker", id)
	}
	receipts, err := env.receipts()
	if err != nil {
		return nil, err
	}
	pending, err := receiptsFor(receipts, id)
	if err != nil {
		return nil, err
	}
	prep := &trackerClosePrep{env: env, card: card, trackerRef: snap.Ref(), receipts: receipts}
	for _, r := range pending {
		if r.Operation() != "completion" {
			continue
		}
		if !r.Discardable() {
			return nil, fmt.Errorf("#%s has a close in progress (%s at %s); finish it with `sdlc issue recovery reconcile --issue %s` before closing again", id, r.Spec().Token, r.Stage(), issue.CLIRef(id))
		}
		prep.superseded = append(prep.superseded, r)
	}
	return prep, nil
}

// publishTrackerClose records a finalized close: the evidence commit, then
// codecomplete bound to it on the card. On FIX-THEN-SHIP the receipt is stored
// unstarted, so the agent's fixes land first and the evidence commit — the
// anchor merge checks the reviewed state against — comes last.
func publishTrackerClose(stdout, stderr io.Writer, f *closeFlags, r closeResult, review reviewResult) error {
	prep := r.trackerPrep
	if prep == nil {
		return errors.New("tracker close was not prepared before its review")
	}
	env, card := prep.env, prep.card
	id := card.ID
	mainView, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	reviewed, err := env.git("rev-parse", "--verify", review.Head+"^{commit}")
	if err != nil {
		return fmt.Errorf("resolve the reviewed commit: %w", err)
	}
	entries, err := closeEvidence(env, r, f.plansDir())
	if err != nil {
		return err
	}
	blob := entries[0].Blob // the details, pinned as they are now
	actual := f.Actual
	if actual == "" {
		actual = issue.ActualNotApplicableSentinel
	}
	message := fmt.Sprintf("#%s: close\n\n%s\nClose-Actual: %s", issue.CLIRef(id), strings.Join(reviewTrailers(review), "\n"), actual)
	spec := tracker.ReceiptSpec{Token: operationToken("close"), Repository: env.target.Repository, IssueID: id,
		CardPath: card.Path, SourcePath: entries[0].Path, DestinationPath: entries[0].Path, SourceBranch: env.branchRef(),
		SourceBase: reviewed, SourceHEAD: reviewed, ReviewedHEAD: reviewed, SourceBlob: blob, CardOID: card.BlobOID,
		TrackerBase: prep.trackerRef, MainBase: mainView.Ref(), Source: tracker.LocalSource,
		EvidenceMessage: message, EvidencePaths: tracker.FormatEvidenceEntries(entries)}
	c, err := tracker.NewCompletion(spec)
	if err != nil {
		return err
	}
	receipts := prep.receipts
	for _, old := range prep.superseded {
		if err := receipts.Discard(old); err != nil {
			return fmt.Errorf("release the superseded close %s: %w", old.Spec().Token, err)
		}
		cinfo(stderr, fmt.Sprintf("superseded the unfinished close %s", old.Spec().Token))
	}
	if review.Verdict == judge.VerdictFixThenShip {
		if err := receipts.Save(c.Receipt()); err != nil {
			return err
		}
		cwarn(stderr, fmt.Sprintf("FIX-THEN-SHIP: commit the fixes on %s, then `sdlc issue recovery reconcile --issue %s` records this close\n"+
			"      (the evidence commit lands after your fixes; codecomplete is published bound to it)", env.branch, issue.CLIRef(id)))
		return nil
	}
	op := tracker.NewCompletionOp(env.ctx, env.repo, env.branchRef(), gitEvidence{env, stderr}, time.Now().Format("2006-01-02"), env.ancestorOf)
	final, err := tracker.Drive(c.Receipt(), tracker.CompletionStepper, op, receipts)
	invalidateIssueRecords(env.ctx)
	if err != nil {
		return fmt.Errorf("%w\n      the close is recorded; finish it with `sdlc issue recovery reconcile --issue %s`", err, issue.CLIRef(id))
	}
	cok(stderr, fmt.Sprintf("#%s codecomplete on %s, bound to evidence commit %s", id, env.branch, shortOID(tracker.EvidenceCommit(final))))
	fmt.Fprintln(stdout, tracker.EvidenceCommit(final))
	return nil
}
