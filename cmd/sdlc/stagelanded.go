package main

import (
	"bytes"
	"context"
	"fmt"
	"path"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// landedStageProof lasts only through this locked clear. A fresh card read at
// the effect boundary rejects a reopened/replaced generation during the proof.
type landedStageProof struct {
	env  *trackerEnv
	card tracker.Record
}

func (p *landedStageProof) revalidate() error {
	snapshot, err := p.env.repo.Snapshot()
	if err != nil {
		return err
	}
	current, err := snapshot.Require(p.card.ID)
	if err != nil {
		return err
	}
	if current.Path != p.card.Path || !bytes.Equal(current.Raw, p.card.Raw) {
		return fmt.Errorf("completion card #%s changed during landed-holder proof", issue.CLIRef(p.card.ID))
	}
	return p.env.ctx.Err()
}

// proveLandedStageHolder reuses landing's immutable PR, completion ownership
// and archive proof. Empty ownership is deliberately NOT successful here.
// main is the very same pin the caller will restore and build, not a new fetch.
func proveLandedStageHolder(ctx context.Context, root, main string, held fleet.StageLease) (*landedStageProof, error) {
	ctx = withIssueRecordsScope(ctx) // no earlier command-scoped card observation
	env, err := openTrackerAt(ctx, root)
	if err != nil {
		return nil, err
	}
	// Clear's existing main pin is origin/main. Its repository must agree with
	// the publication remote that owns this completion and tracker.
	origin, err := env.git("remote", "get-url", "--all", "origin")
	if err != nil {
		return nil, err
	}
	repository, err := gitx.PublicationRepository(root, origin)
	if err != nil || repository != env.target.Repository {
		return nil, fmt.Errorf("origin/main and the completion tracker name different repositories")
	}
	snapshot, err := env.repo.Snapshot()
	if err != nil {
		return nil, err
	}
	id, _, ok := issue.ParseFilename(held.Branch + ".md")
	if !ok {
		return nil, fmt.Errorf("stage does not name an issue branch")
	}
	card, err := snapshot.Require(id)
	if err != nil {
		return nil, err
	}
	status, _ := issue.GetField(card.Card.Frontmatter, "status")
	binding, present, err := issue.CardCompletion(card.Raw)
	if err != nil {
		return nil, err
	}
	if status != "done" || !present || binding.Repository != env.target.Repository || !landingOIDValid(binding.LandedCommit) {
		return nil, fmt.Errorf("%s requires its current done completion and landed commit (status %s)", held.Holder, status)
	}
	if onMain, err := evidenceInHead(env, binding.LandedCommit, main); err != nil {
		return nil, err
	} else if !onMain {
		return nil, fmt.Errorf("%s completion is not confirmed on fresh main", held.Holder)
	}
	// Use the existing landing transport seam (including effective URL
	// validation), rather than introduce a second GitHub protocol.
	url, err := landingGit(boundedLandingGit(ctx, mergeRunner), root, "remote", "get-url", "--all", env.target.Remote)
	if err != nil {
		return nil, err
	}
	repo, err := landingRepoURL(url)
	if err != nil {
		return nil, err
	}
	gh, ok := ghClient.(landingGH)
	if !ok {
		return nil, fmt.Errorf("GitHub landing evidence unavailable")
	}
	prs, err := gh.LandingPRs(ctx, repo, held.Branch)
	if err != nil {
		return nil, err
	}
	if len(prs) > landingPRLimit {
		return nil, fmt.Errorf("landing PR evidence exceeds %d records", landingPRLimit)
	}
	dirs, err := resolveArchiveRoots(root, envOr("WF_ISSUES_DIR", "workshop/issues"), envOr("WF_PLANS_DIR", "workshop/plans"), envOr("WF_HISTORY_DIR", "workshop/history"))
	if err != nil {
		return nil, err
	}
	matches := 0
	for _, pr := range prs {
		if pr.Repo != repo || pr.HeadRef != held.Branch || pr.State != "MERGED" {
			continue
		}
		var probeErr error
		pr.LandedOID, ok = landedCommit(pr, func(oid string) bool {
			yes, err := evidenceInHead(env, oid, main)
			if err != nil {
				probeErr = err
			}
			return yes
		})
		if probeErr != nil {
			return nil, probeErr
		}
		if !ok {
			continue
		}
		if err := validateArchivePR(repo, pr); err != nil {
			return nil, err
		}
		carried, err := env.ancestorOf(held.Commit, pr.HeadOID)
		if err != nil {
			return nil, err
		}
		if !carried {
			continue
		}
		selected, err := selectArchiveLandingIssues(ctx, root, pr, dirs)
		if err != nil {
			return nil, err
		}
		owns := false
		for _, owned := range selected {
			ownedID, _, valid := issue.ParseFilename(path.Base(owned.path))
			if valid && ownedID == id && owned.tracked && bytes.Equal(owned.card, card.Raw) {
				owns = true
			}
		}
		if !owns {
			continue // an issue-free archive proof grants no stage authority
		}
		// Archive's shared legacy policy can prove byte-preserving moves
		// without a card mirror. Stage release needs the stronger generation
		// claim: this archive must pin this current close's done binding.
		archivedPath := archiveDestination(dirs.history, vocab.ArchiveIssues, path.Base(card.Path))
		archived, err := env.main.ViewOf(main).Read(archivedPath)
		if err != nil {
			return nil, err
		}
		mirror, err := issue.MirrorBaselineOID(archived)
		if err != nil {
			return nil, fmt.Errorf("%s archive does not pin a completion generation: %w", held.Holder, err)
		}
		pinned, err := env.repo.ReadCardBlob(mirror)
		if err != nil {
			return nil, err
		}
		archivedCard, err := issue.ParseCard(pinned)
		if err != nil {
			return nil, err
		}
		archivedBinding, present, err := issue.CardCompletion(pinned)
		archivedStatus, _ := issue.GetField(archivedCard.Frontmatter, "status")
		if err != nil || !present || archivedCard.ID != card.ID || archivedStatus != "done" || archivedBinding != binding {
			return nil, fmt.Errorf("%s archive does not pin its current done completion", held.Holder)
		}
		complete, err := confirmLandingArchive(ctx, root, main, repo, pr, dirs, selected)
		if err != nil {
			return nil, err
		}
		if !complete {
			return nil, fmt.Errorf("%s archive is not confirmed on fresh main", held.Holder)
		}
		matches++
	}
	if matches != 1 {
		return nil, fmt.Errorf("%s needs one matching merged PR carrying the staged commit, current completion and archive; found %d", held.Holder, matches)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &landedStageProof{env: env, card: card}, nil
}

func landedStageRefusal(held fleet.StageLease, target string, err error) error {
	return fmt.Errorf("unverified landed holder %s on %s: %w; finish/reconcile landing or repair evidence, then run sdlc stage --clear %s from %s on its resting branch", held.Holder, target, err, target, held.Slot)
}
