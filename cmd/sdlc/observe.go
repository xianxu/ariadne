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
		in.Evidence = collectEvidence(root, remote, stem, status, in.Branch)
	} else if stem != "" {
		in.MainArchiveErr, in.Branch.Err, in.Evidence.Err = envErr, envErr, envErr
		in.Evidence.Source = "(publication remote unresolved)"
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
		if id, err := workspace.Resolve(execGitRunner{}, w.Path, ""); err == nil && id.Address != nil && id.UsesSlotLayout() {
			h.Address = *id.Address
		}
		var err error
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
// merge or a deleted branch never strands it.
func collectEvidence(root, remote, stem, status string, branch observe.BranchFacts) observe.Evidence {
	history := envOr("WF_HISTORY_DIR", "workshop/history")
	plans := envOr("WF_PLANS_DIR", "workshop/plans")
	issues := envOr("WF_ISSUES_DIR", "workshop/issues")
	type place struct{ ref, plans, details string }
	var places []place
	main := "refs/remotes/" + remote + "/main"
	if status == "done" {
		places = append(places,
			place{main, vocab.ArchiveSubdir(history, vocab.ArchivePlans), path.Join(vocab.ArchiveSubdir(history, vocab.ArchiveIssues), stem+".md")},
			place{main, plans, path.Join(issues, stem+".md")})
	}
	if branch.Ref != "" {
		places = append(places, place{branch.Ref, plans, path.Join(issues, stem+".md")})
	}
	for _, p := range places {
		ev := observe.Evidence{Source: p.ref + ":" + p.plans, Artifacts: map[string]observe.ArtifactFacts{}}
		listing, err := observeGit(root, "ls-tree", "--name-only", p.ref, "--", p.plans+"/")
		if err != nil {
			ev.Err = fmt.Errorf("list %s: %w", ev.Source, err)
			return ev
		}
		files := map[string]bool{}
		for _, f := range strings.Split(strings.TrimSpace(string(listing)), "\n") {
			if strings.HasPrefix(path.Base(f), stem+"-") {
				files[path.Base(f)] = true
			}
		}
		details, derr := observeGit(root, "show", p.ref+":"+p.details)
		if len(files) == 0 && derr != nil {
			continue // nothing of this issue here; try the next place
		}
		if derr == nil {
			ev.Details = details
		}
		show := func(name string) (string, error) {
			out, err := observeGit(root, "show", p.ref+":"+path.Join(p.plans, name))
			return string(out), err
		}
		// The plan-quality ledger: the plan boundary's open blocking findings.
		if files[stem+"-plan-gate.md"] {
			art := observe.ArtifactFacts{Found: true}
			if text, err := show(stem + "-plan-gate.md"); err != nil {
				art.LedgerErr = err
			} else if l, err := gatestate.ParseSidecar(text); err != nil {
				art.LedgerErr = err
			} else {
				d := gatestate.Decide(l, gatestate.DefaultRoundCap)
				art.OpenBlocking = len(d.OpenBlocking) + len(d.Demoted) // cap-independent
			}
			ev.PlanGate = &art
		}
		// The issue-wide boundary ledger, scoped per boundary as its gate scopes it.
		var ledger *gatestate.Ledger
		var ledgerErr error
		if files[stem+"-close-gate.md"] {
			if text, err := show(stem + "-close-gate.md"); err != nil {
				ledgerErr = err
			} else if l, err := gatestate.ParseSidecar(text); err != nil {
				ledgerErr = err
			} else {
				ledger = &l
			}
		}
		for name := range files {
			boundary := ""
			switch {
			case name == stem+"-close-review.md":
				boundary = "close"
			case strings.HasSuffix(name, "-review.md") && strings.HasPrefix(strings.TrimPrefix(name, stem+"-"), "m"):
				boundary = "M" + strings.TrimSuffix(strings.TrimPrefix(name, stem+"-m"), "-review.md")
			default:
				continue
			}
			art := observe.ArtifactFacts{Found: true, LedgerErr: ledgerErr}
			if text, err := show(name); err != nil {
				art.SidecarErr = err
			} else {
				art.Sidecar = text
			}
			if ledger != nil {
				milestone := boundary
				if boundary == "close" {
					milestone = ""
				}
				d := gatestate.DecideScoped(gatestate.FilterBoundary(*ledger, milestone), openScopeFor(*ledger, milestone), gatestate.DefaultRoundCap)
				art.OpenBlocking = len(d.OpenBlocking) + len(d.Demoted) // cap-independent
			}
			ev.Artifacts[boundary] = art
		}
		return ev
	}
	return observe.Evidence{}
}
