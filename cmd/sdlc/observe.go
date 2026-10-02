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
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/observe"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
	"github.com/xianxu/ariadne/pkg/workspace"
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
			} else {
				in.CardErr = rec.CardErr
			}
			if rec.Card == nil && rec.DetailPath != "" {
				stem = strings.TrimSuffix(path.Base(rec.DetailPath), ".md")
			}
		}
		if rs.Ref != "" {
			if oid, rerr := observeGit(root, "rev-parse", "--verify", rs.Ref+"^{commit}"); rerr == nil {
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
	if out, err := observeGit(root, "worktree", "list", "--porcelain", "-z"); err != nil {
		in.WorktreesErr = fmt.Errorf("git worktree list: %w", err)
	} else if trees, perr := gitx.ParseWorktrees(out); perr != nil {
		in.WorktreesErr = perr
	} else {
		for _, w := range trees {
			in.Worktrees = append(in.Worktrees, observe.LocalWorktree{Path: canonRoot(w.Path), Branch: strings.TrimPrefix(w.Branch, "refs/heads/")})
		}
	}
	if stem != "" && envErr == nil {
		remote := env.target.Remote
		in.MainArchive, in.MainArchiveErr = archivedOnMain(root, remote, stem)
		in.Branch = collectBranch(root, remote, stem)
		if in.WorktreesErr == nil {
			in.Holding = collectHolding(in.Worktrees, remote, stem)
		}
		status := ""
		if rec, ok := rs.Get(id); ok {
			status = rec.Status()
		}
		if rel, rerr := filepath.Rel(canonRoot(root), canonRoot(issuesDir)); rerr != nil || strings.HasPrefix(rel, "..") {
			in.Evidence = observe.Evidence{Source: issuesDir, Err: fmt.Errorf("the issues directory %s is outside %s", issuesDir, root)}
		} else {
			in.Evidence = collectEvidence(root, filepath.ToSlash(rel), remote, stem, status, in.Branch)
		}
	} else if stem != "" {
		in.MainArchiveErr, in.Branch.Err, in.Evidence.Err = envErr, envErr, envErr
		in.Evidence.Source = "(publication remote unresolved)"
	}
	return observe.Assemble(in)
}

// archivedOnMain is the issue's archived details path on the publication
// remote's main as last fetched ("" when not archived there).
func archivedOnMain(root, remote, stem string) (string, error) {
	want := path.Join(vocab.ArchiveSubdir(vocab.Issue().Discovery().Archive, vocab.ArchiveIssues), stem+".md")
	out, err := observeGit(root, "ls-tree", "--name-only", "refs/remotes/"+remote+"/main", "--", want)
	if err != nil {
		return "", fmt.Errorf("read %s/main: %w", remote, err)
	}
	if strings.TrimSpace(string(out)) == want {
		return want, nil
	}
	return "", nil
}

// observeGitReader routes workspace.Resolve's git reads through observeGit,
// so the operating-envelope count sees them too.
type observeGitReader struct{}

func (observeGitReader) GitInDir(dir string, args ...string) ([]byte, error) {
	return observeGit(dir, args...)
}

// observeGit runs one read-only git command in dir (counted by tests for the
// operating-envelope bound).
var observeGit = func(dir string, args ...string) ([]byte, error) {
	return gitx.RunGit(append([]string{"-C", dir}, args...)...)
}

func gitLine(dir string, args ...string) (string, error) {
	out, err := observeGit(dir, args...)
	return strings.TrimSpace(string(out)), err
}

// refExists is a definite yes/no for a ref; any other failure is an error.
func refExists(root, ref string) (bool, error) {
	out, err := observeGit(root, "for-each-ref", "--format=%(refname)", ref)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == ref, nil
}

// collectBranch finds the issue branch here: local, else the remote's copy.
func collectBranch(root, remote, stem string) observe.BranchFacts {
	var b observe.BranchFacts
	for _, ref := range []string{"refs/heads/" + stem, "refs/remotes/" + remote + "/" + stem} {
		ok, err := refExists(root, ref)
		if err != nil {
			b.Err = err
			return b
		}
		if ok {
			b.Ref = ref
			break
		}
	}
	if b.Ref == "" {
		return b
	}
	var err error
	if b.Head, err = gitLine(root, "rev-parse", "--verify", b.Ref+"^{commit}"); err != nil {
		b.Err = err
		return b
	}
	if b.LastCommitAt, err = gitLine(root, "log", "-1", "--format=%cI", b.Ref); err != nil {
		b.Err = err
		return b
	}
	count, err := gitLine(root, "rev-list", "--count", "refs/remotes/"+remote+"/main.."+b.Ref)
	if err == nil {
		_, err = fmt.Sscan(count, &b.AheadOfMain)
	}
	b.Err = err
	return b
}

// collectHolding reads activity in each local worktree that holds the branch.
func collectHolding(trees []observe.LocalWorktree, remote, stem string) []observe.HoldingFacts {
	var out []observe.HoldingFacts
	for _, w := range trees {
		if w.Branch != stem {
			continue
		}
		h := observe.HoldingFacts{Path: w.Path, Branch: w.Branch}
		var err error
		if id, rerr := workspace.Resolve(observeGitReader{}, w.Path, ""); rerr != nil {
			err = fmt.Errorf("resolve the workspace: %w", rerr)
		} else if id.Address != nil && id.UsesSlotLayout() {
			h.Address = *id.Address
		}
		if err != nil {
			h.Err = err
			out = append(out, h)
			continue
		}
		if h.Head, err = gitLine(w.Path, "rev-parse", "HEAD"); err == nil {
			var status []byte
			if status, err = observeGit(w.Path, "status", "--porcelain=v1", "-z", "--untracked-files=all"); err == nil {
				entries, perr := gitx.ParseStatusZ(status)
				h.DirtyCount, err = len(entries), perr
			}
		}
		if err == nil {
			var lr string
			if lr, err = gitLine(w.Path, "rev-list", "--left-right", "--count", "HEAD...refs/remotes/"+remote+"/main"); err == nil {
				_, err = fmt.Sscan(lr, &h.Ahead, &h.Behind)
			}
		}
		h.Err = err
		out = append(out, h)
	}
	return out
}

// collectEvidence reads the committed review evidence: on the issue branch
// before landing, on main after (archived with the issue, #143) — so a squash
// merge or a deleted branch never strands it. Every path derives from the given
// issues dir and the vocabulary's discovery; artifact names from the writers'
// own path helpers (sidecarPath, sidecarPathFor, the gate suffixes).
func collectEvidence(root, issuesRel, remote, stem, status string, branch observe.BranchFacts) observe.Evidence {
	d := vocab.Issue().Discovery()
	issueFile := stem + ".md"
	type place struct{ ref, plans, details string }
	var places []place
	main := "refs/remotes/" + remote + "/main"
	if status == "done" {
		places = append(places,
			place{main, vocab.ArchiveSubdir(d.Archive, vocab.ArchivePlans), path.Join(vocab.ArchiveSubdir(d.Archive, vocab.ArchiveIssues), issueFile)},
			place{main, d.Plans, path.Join(issuesRel, issueFile)})
	}
	if branch.Ref != "" {
		places = append(places, place{branch.Ref, d.Plans, path.Join(issuesRel, issueFile)})
	}
	for _, p := range places {
		ev := observe.Evidence{Source: p.ref + ":" + p.plans, Artifacts: map[string]observe.ArtifactFacts{}}
		listing, err := observeGit(root, "ls-tree", "--name-only", p.ref, "--", p.plans+"/", p.details)
		if err != nil {
			ev.Err = fmt.Errorf("list %s: %w", ev.Source, err)
			return ev
		}
		present := map[string]bool{}
		for _, f := range strings.Split(strings.TrimSpace(string(listing)), "\n") {
			present[f] = true
		}
		slash := func(p string) string { return filepath.ToSlash(p) }
		planGate := slash(sidecarPathFor(p.plans, issueFile, planGateSuffix))
		closeGate := slash(sidecarPathFor(p.plans, issueFile, boundaryGateSuffix))
		reviews := map[string]string{} // boundary → artifact path
		if close := slash(sidecarPath(p.plans, issueFile, "")); present[close] {
			reviews["close"] = close
		}
		for f := range present {
			if m := reviewMilestoneRe.FindStringSubmatch(f); m != nil {
				if want := slash(sidecarPath(p.plans, issueFile, "M"+m[1])); want == f {
					reviews["M"+m[1]] = f
				}
			}
		}
		if !present[p.details] && !present[planGate] && !present[closeGate] && len(reviews) == 0 {
			continue // nothing of this issue here; try the next place
		}
		show := func(file string) (string, error) {
			out, err := observeGit(root, "show", p.ref+":"+file)
			return string(out), err
		}
		if present[p.details] {
			text, err := show(p.details)
			ev.Details, ev.DetailsErr = []byte(text), err
			if err != nil {
				ev.Details = nil
			}
		}
		readLedger := func(file string) (*gatestate.Ledger, error) {
			text, err := show(file)
			if err != nil {
				return nil, err
			}
			l, err := gatestate.ParseSidecar(text)
			return &l, err
		}
		// Open blocking-severity findings, independent of the round cap
		// (DecideScoped demotes past it).
		openBlocking := func(d gatestate.Decision) int { return len(d.OpenBlocking) + len(d.Demoted) }
		if present[planGate] {
			art := observe.ArtifactFacts{Found: true}
			if l, err := readLedger(planGate); err != nil {
				art.LedgerErr = err
			} else {
				art.OpenBlocking = openBlocking(gatestate.Decide(*l, gatestate.DefaultRoundCap))
			}
			ev.PlanGate = &art
		}
		var ledger *gatestate.Ledger
		var ledgerErr error
		if present[closeGate] {
			ledger, ledgerErr = readLedger(closeGate)
		}
		for boundary, file := range reviews {
			art := observe.ArtifactFacts{Found: true, LedgerErr: ledgerErr}
			art.Sidecar, art.SidecarErr = show(file)
			if ledger != nil && ledgerErr == nil {
				milestone := boundary
				if boundary == "close" {
					milestone = ""
				}
				art.OpenBlocking = openBlocking(gatestate.DecideScoped(gatestate.FilterBoundary(*ledger, milestone), openScopeFor(*ledger, milestone), gatestate.DefaultRoundCap))
			}
			ev.Artifacts[boundary] = art
		}
		return ev
	}
	return observe.Evidence{}
}
