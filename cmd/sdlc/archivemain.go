// archivemain.go — archiving one issue's details and plans on main in one
// narrow commit, outside a landing: `sdlc abandon` (#286) and the bookkeeping
// for a merge done outside sdlc (#287). It follows the landing archive's
// rules (archiveDestination, and archivedDetails mirroring the terminal card).
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// archiveDirs are the issue, plans and history roots one archive reads and
// writes; each caller passes its own (a verb's flags, or the environment).
type archiveDirs struct{ Issues, Plans, History string }

// envArchiveDirs is the roots from the environment, for verbs without flags.
func envArchiveDirs() archiveDirs {
	return archiveDirs{Issues: envOr("WF_ISSUES_DIR", "workshop/issues"), Plans: plansDir(), History: historyDir()}
}

// mainArchive says what one archive commit carries and who may make it.
type mainArchive struct {
	Message string
	Dirs    archiveDirs
	Assets  *issueassets.Bundle // validated kept-branch assets; nil reads main
	Final   []byte              // the details to archive; nil archives main's copy
	Plans   map[string][]byte   // plan contents by basename, replacing main's copies
	Sides   []gitx.TreeFile     // kept-branch side-folder files at their live paths (#424); nil reads main
	Note    string              // a ## Log line added to main's copy (Final nil only)
	Allow   func(tracker.Record) error
}

// archiveIssueOnMain moves the issue's live details (mirrored to its card) and
// its plan artifacts and assets on main into the history archive. A retry
// without live details must still validate the archived bundle.
func archiveIssueOnMain(env *trackerEnv, stderr io.Writer, id, detailRel string, a mainArchive) error {
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	card, err := snap.Require(id)
	if err != nil {
		return err
	}
	base := path.Base(detailRel)
	archived := false
	replacingEpisode := false
	prepare := func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
		replacingEpisode = false
		live, err := view.Exists(detailRel)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		archivedDetail := archiveDestination(a.Dirs.History, vocab.ArchiveIssues, base)
		if !live {
			if exists, err := view.Exists(archivedDetail); err != nil {
				return gitx.TrunkWrite{}, err
			} else if !exists {
				return gitx.TrunkWrite{}, fmt.Errorf("archived issue %s is missing: restore the archived details before retrying", archivedDetail)
			}
			if _, err := assetFilesFromView(view, detailRel, id, nil); err != nil {
				return gitx.TrunkWrite{}, fmt.Errorf("active asset residue for %s: %w", detailRel, err)
			}
			// #424: likewise a side folder left behind by the details' move.
			if residue, err := trunkSideResidue(view, detailRel, a.Dirs.Plans); err != nil {
				return gitx.TrunkWrite{}, err
			} else if residue != "" {
				return gitx.TrunkWrite{}, fmt.Errorf("active side-doc residue for %s: %s; move it beside the archived details, or remove it, before retrying", detailRel, residue)
			}
			raw, err := view.Read(archivedDetail)
			if err != nil {
				return gitx.TrunkWrite{}, fmt.Errorf("cannot confirm archived issue %s: %w; restore the archived details before retrying", archivedDetail, err)
			}
			bundle, err := assetFilesFromView(view, archivedDetail, id, raw)
			if err != nil {
				return gitx.TrunkWrite{}, err
			}
			if a.Final != nil && a.Assets != nil {
				successor, err := archiveAbandonSuccessor(env, detailRel, raw, bundle, card.Raw)
				if err != nil {
					return gitx.TrunkWrite{}, err
				}
				if successor {
					replacingEpisode = true
					if err := requireNewArchiveAssets(env, view, bundle, *a.Assets); err != nil {
						return gitx.TrunkWrite{}, err
					}
					w := gitx.TrunkWrite{Write: assetBundleWrite(*a.Assets, archivedDetail), ExactBytes: true}
					w.Write[archivedDetail] = mirrorTerminal(env, a.Final, card.Raw)
					// The archived side folders become this episode's, as its
					// assets do (#424).
					if err := trunkSideReplace(view, archivedDetail, a, &w); err != nil {
						return gitx.TrunkWrite{}, err
					}
					for p := range bundle.Files {
						if _, keep := a.Assets.Files[p]; !keep {
							w.Delete = append(w.Delete, path.Join(path.Dir(archivedDetail), p))
						}
					}
					for name, body := range a.Plans {
						w.Write[archiveDestination(a.Dirs.History, vocab.ArchivePlans, name)] = body
					}
					archived = true
					return w, nil
				}
			}
			if a.Final != nil && !bytes.Equal(raw, mirrorTerminal(env, a.Final, card.Raw)) {
				return gitx.TrunkWrite{}, fmt.Errorf("archived details %s differ from the kept generation: restore the kept issue details before retrying", archivedDetail)
			}
			if a.Assets != nil && !assetBundlesEqual(bundle, *a.Assets) {
				return gitx.TrunkWrite{}, fmt.Errorf("archived assets for %s differ from the kept generation: restore the kept asset bundle before retrying", archivedDetail)
			}
			return gitx.TrunkWrite{}, tracker.ErrNoChange
		}
		if exists, err := view.Exists(archivedDetail); err != nil {
			return gitx.TrunkWrite{}, err
		} else if exists {
			return gitx.TrunkWrite{}, fmt.Errorf("archive destination occupied: %s; reconcile or remove that destination before retrying", archivedDetail)
		}
		mainRaw, err := view.Read(detailRel)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		mainAssets, err := assetFilesFromView(view, detailRel, id, mainRaw)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		content := a.Final
		if content == nil {
			content = mainRaw
			if a.Note != "" {
				fm, body, err := issue.Parse(string(content))
				if err != nil {
					return gitx.TrunkWrite{}, err
				}
				content = []byte(issue.Compose(fm, insertLogLine(body, a.Note)))
			}
		}
		w := gitx.TrunkWrite{Write: map[string][]byte{
			archiveDestination(a.Dirs.History, vocab.ArchiveIssues, base): mirrorTerminal(env, content, card.Raw),
		}, Delete: []string{detailRel}, ExactBytes: true}
		residue := ""
		_, err = view.FilesSelected(path.Join(path.Dir(archivedDetail), "assets", id), func(f gitx.TreeFile) bool {
			if residue == "" {
				residue = f.Path
			}
			return false
		})
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		if residue != "" {
			return gitx.TrunkWrite{}, fmt.Errorf("archive asset already exists: %s; reconcile or remove that archived asset before retrying", residue)
		}
		bundle := mainAssets
		if a.Assets != nil {
			bundle = *a.Assets
		}
		if err := requireNewArchiveAssets(env, view, mainAssets, bundle); err != nil {
			return gitx.TrunkWrite{}, err
		}
		for dest, data := range assetBundleWrite(bundle, archivedDetail) {
			if exists, err := view.Exists(dest); err != nil {
				return gitx.TrunkWrite{}, err
			} else if exists {
				return gitx.TrunkWrite{}, fmt.Errorf("archive destination occupied: %s; reconcile or remove that destination before retrying", dest)
			}
			w.Write[dest] = data
		}
		for src := range assetBundleWrite(mainAssets, detailRel) {
			w.Delete = append(w.Delete, src)
		}
		plans, err := view.Files(a.Dirs.Plans)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		// Side ownership comes from the tree inventory, including orphan
		// plan-side folders, independently of the plan document's presence.
		paths := make([]string, 0, len(plans)+len(a.Plans)+len(a.Sides))
		for _, p := range plans {
			paths = append(paths, p.Path)
			if path.Dir(p.Path) == a.Dirs.Plans && planArtifactBelongsToIssue(base, path.Base(p.Path)) {
				dest := archiveDestination(a.Dirs.History, vocab.ArchivePlans, path.Base(p.Path))
				w.Write[dest] = p.Content
				w.Delete = append(w.Delete, p.Path)
			}
		}
		for name := range a.Plans {
			paths = append(paths, path.Join(a.Dirs.Plans, name))
		}
		for _, f := range a.Sides {
			paths = append(paths, f.Path)
		}
		sideDocs := sideArchiveDocsForIssue(detailRel, archivedDetail, paths, a.Dirs.Plans, a.Dirs.History)
		// #424/#623: every owned side folder moves, bytes unchanged.
		if err := trunkSideArchive(view, sideDocs, &w, a.Sides); err != nil {
			return gitx.TrunkWrite{}, err
		}
		for name, content := range a.Plans {
			w.Write[archiveDestination(a.Dirs.History, vocab.ArchivePlans, name)] = content
		}
		archived = true
		return w, nil
	}
	err = mainPublish(env, a.Message, prepare, func(string, string) error {
		fresh, err := env.repo.Snapshot()
		if err != nil {
			return err
		}
		c, err := fresh.Require(id)
		if err != nil {
			return err
		}
		if replacingEpisode && !bytes.Equal(c.Raw, card.Raw) {
			return fmt.Errorf("card #%s changed while preparing the successor archive; rerun sdlc abandon to recompute the current generation", issue.CLIRef(id))
		}
		return a.Allow(c)
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

// Relocation of already-published bytes survives policy tightening. Evidence
// present only on a kept branch is a new admission to main and rechecks policy.
func requireNewArchiveAssets(env *trackerEnv, destination *gitx.TrunkView, published, next issueassets.Bundle) error {
	if len(issueassets.BundleDelta(published, next).Writes) == 0 {
		return nil
	}
	return requireAssetPublicationPolicy(env.ctx, env.root, destination)
}

// A reopen restores the kept branch, while main retains the previous episode's
// archive. Only a proven successor may replace it; retrying the same episode
// must keep the exact archived generation. The prior mirror pins its immutable
// terminal card, and that card names the source commit whose body/assets we
// compare before allowing replacement.
func archiveAbandonSuccessor(env *trackerEnv, detailRel string, archived []byte, bundle issueassets.Bundle, currentCard []byte) (bool, error) {
	oid, err := issue.MirrorBaselineOID(archived)
	if err != nil {
		return false, nil
	}
	oldCard, err := env.repo.ReadCardBlob(oid)
	if err != nil {
		return false, fmt.Errorf("cannot read prior archived card %s: %w; fetch the issue tracker before retrying", oid, err)
	}
	old, oldOK, err := issue.CardAbandoned(oldCard)
	if err != nil {
		return false, err
	}
	next, nextOK, err := issue.CardAbandoned(currentCard)
	if err != nil {
		return false, err
	}
	if !oldOK || !nextOK || !old.Started() || !next.Started() || old.Head == next.Head || old.Ref != next.Ref || old.Branch != next.Branch {
		return false, nil
	}
	ancestor, err := env.ancestorOf(old.Head, next.Head)
	if err != nil {
		return false, err
	}
	if !ancestor {
		return false, fmt.Errorf("prior archived issue is not an ancestor of kept commit %s: reconcile the abandoned branch before retrying", next.Head)
	}
	priorView := env.main.ViewOf(old.Head)
	source, err := priorView.Read(detailRel)
	if err != nil {
		return false, err
	}
	if !bytes.Equal(archived, mirrorTerminal(env, source, oldCard)) {
		return false, fmt.Errorf("prior archived details %s changed: restore their recorded generation before retrying", detailRel)
	}
	original, err := assetFilesFromView(priorView, detailRel, issueIDPrefix(path.Base(detailRel)), source)
	if err != nil {
		return false, err
	}
	if !assetBundlesEqual(original, bundle) {
		return false, fmt.Errorf("prior archived assets for %s changed: restore their recorded generation before retrying", detailRel)
	}
	return true, nil
}

func publishedAbandonAssets(view *gitx.TrunkView, detailRel, id string) (issueassets.Bundle, error) {
	source := detailRel
	if exists, err := view.Exists(source); err != nil {
		return issueassets.Bundle{}, err
	} else if !exists {
		source = archiveDestination(historyDir(), vocab.ArchiveIssues, path.Base(detailRel))
	}
	raw, err := view.Read(source)
	if err != nil {
		return issueassets.Bundle{}, err
	}
	return assetFilesFromView(view, source, id, raw)
}
