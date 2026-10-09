// archivemain.go — archiving one issue's details and plans on main in one
// narrow commit, outside a landing: `sdlc abandon` (#286) and the bookkeeping
// for a merge done outside sdlc (#287). It follows the landing archive's
// rules (archiveDestination, and archivedDetails mirroring the terminal card).
package main

import (
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// mainArchive says what one archive commit carries and who may make it.
type mainArchive struct {
	Message string
	Final   []byte            // the details to archive; nil archives main's copy
	Plans   map[string][]byte // plan contents by basename, replacing main's copies
	Note    string            // a ## Log line added to main's copy (Final nil only)
	Allow   func(tracker.Record) error
}

// archiveIssueOnMain moves the issue's live details (mirrored to its card) and
// its plan artifacts on main into the history archive. Main without a live
// copy is already archived: nothing to do.
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
	prepare := func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
		live, err := view.Exists(detailRel)
		if err != nil || !live {
			return gitx.TrunkWrite{}, errors.Join(err, tracker.ErrNoChange)
		}
		content := a.Final
		if content == nil {
			if content, err = view.Read(detailRel); err != nil {
				return gitx.TrunkWrite{}, err
			}
			if a.Note != "" {
				fm, body, err := issue.Parse(string(content))
				if err != nil {
					return gitx.TrunkWrite{}, err
				}
				content = []byte(issue.Compose(fm, insertLogLine(body, a.Note)))
			}
		}
		w := gitx.TrunkWrite{Write: map[string][]byte{
			archiveDestination(historyDir(), vocab.ArchiveIssues, base): mirrorTerminal(env, content, card.Raw),
		}, Delete: []string{detailRel}, ExactBytes: true}
		plans, err := view.Files(plansDir())
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		for _, p := range plans {
			if planArtifactBelongsToIssue(base, path.Base(p.Path)) {
				w.Write[archiveDestination(historyDir(), vocab.ArchivePlans, path.Base(p.Path))] = p.Content
				w.Delete = append(w.Delete, p.Path)
			}
		}
		for name, content := range a.Plans {
			w.Write[archiveDestination(historyDir(), vocab.ArchivePlans, name)] = content
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
