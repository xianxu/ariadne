// observe.go — the IO side of `sdlc issue show --json` (#279): read the
// existing authorities for one issue into observe.Inputs. It writes nothing; the
// one side effect is the tracker fetch's remote-tracking ref update, which every
// read-only view already performs (PreferFresh).
package main

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/observe"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// observeNow is the observation clock (tests pin it).
var observeNow = time.Now

// collectObservation reads issue id of the repository rooted at root. Every
// read failure is carried into Inputs, never collapsed into "absent".
func collectObservation(ctx context.Context, root, issuesDir, id string) observe.Observation {
	in := observe.Inputs{Issue: id, ObservedAt: observeNow()}
	if root == "" {
		// Details outside any repository are the whole record: no tracker, no
		// workspace identity, no worktrees — each said, none faked.
		in.MeErr = errors.New("the issues directory is not inside a repository")
		in.WorktreesErr = in.MeErr
		return observe.Assemble(in)
	}
	rs, err := loadIssueRecordsAt(ctx, root, issuesDir, tracker.PreferFresh)
	switch {
	case err != nil:
		in.Tracked, in.TrackerErr = true, err
	case rs.Stale && !rs.Tracker:
		in.Tracked, in.TrackerErr = true, errors.Join(errors.New("tracker unreachable and never fetched here"), rs.FetchErr)
	default:
		in.Tracked, in.TrackerStale, in.TrackerErr = rs.Tracker, rs.Stale, rs.FetchErr
	}
	stem := ""
	if err == nil {
		if rec, ok := rs.Get(id); ok {
			if rec.Card != nil {
				in.Card, in.CardPath, in.CardBlob = rec.Card.Raw, rec.Card.Path, rec.Card.BlobOID
				stem = strings.TrimSuffix(path.Base(rec.Card.Path), ".md")
			} else if rec.DetailPath != "" {
				stem = strings.TrimSuffix(path.Base(rec.DetailPath), ".md")
			}
		}
		if rs.Ref != "" {
			if oid, rerr := gitx.RunGit("-C", root, "rev-parse", "--verify", rs.Ref+"^{commit}"); rerr == nil {
				in.TrackerRef = strings.TrimSpace(string(oid))
			} else {
				in.TrackerRefErr = fmt.Errorf("resolve %s: %w", rs.Ref, rerr)
			}
		}
	}
	env, envErr := openTrackerAt(ctx, root)
	if envErr == nil {
		if me, err := claimantIdentity(env); err == nil {
			in.Me = &me
		} else {
			in.MeErr = err
		}
	} else {
		in.MeErr = envErr
	}
	if out, err := gitx.RunGit("-C", root, "worktree", "list", "--porcelain", "-z"); err != nil {
		in.WorktreesErr = fmt.Errorf("git worktree list: %w", err)
	} else if trees, perr := gitx.ParseWorktrees(out); perr != nil {
		in.WorktreesErr = perr
	} else {
		for _, w := range trees {
			in.Worktrees = append(in.Worktrees, observe.LocalWorktree{Path: canonRoot(w.Path), Branch: strings.TrimPrefix(w.Branch, "refs/heads/")})
		}
	}
	if stem != "" && envErr == nil {
		in.MainArchive, in.MainArchiveErr = archivedOnMain(root, env.target.Remote, stem)
	} else if stem != "" {
		in.MainArchiveErr = envErr
	}
	return observe.Assemble(in)
}

// archivedOnMain is the issue's archived details path on the publication
// remote's main as last fetched ("" when not archived there).
func archivedOnMain(root, remote, stem string) (string, error) {
	history := envOr("WF_HISTORY_DIR", "workshop/history")
	want := path.Join(vocab.ArchiveSubdir(history, vocab.ArchiveIssues), stem+".md")
	out, err := gitx.RunGit("-C", root, "ls-tree", "--name-only", "refs/remotes/"+remote+"/main", "--", want)
	if err != nil {
		return "", fmt.Errorf("read %s/main: %w", remote, err)
	}
	if strings.TrimSpace(string(out)) == want {
		return want, nil
	}
	return "", nil
}
