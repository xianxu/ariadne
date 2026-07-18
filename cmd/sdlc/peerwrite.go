// peerwrite.go — safe peer-write commit mechanics (#171 M3).
//
// The close gate updates every project file across the fleet that references
// the closing issue (M2). A peer repo's edit must not be left as a surprise in
// its working tree, but blindly committing into someone else's repo is worse.
// The split: `planPeerWrites` is the pure decision core (git state in,
// decisions out — no IO); `readRepoGitState` + `applyPeerWrites` are the thin
// shell around the shared `gitRunner` seam. A peer auto-commits only when the
// commit is unambiguous — on main, clean index, not a brain capture repo
// (#176). Everything else is report-only: the file stays written and the
// operator gets the reason plus the exact command to finish. A report-only
// outcome never fails the close.
package main

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// RepoGitState is the snapshot of a peer repo's git state that the commit
// decision keys on. Read by readRepoGitState; consumed by planPeerWrites.
type RepoGitState struct {
	Branch           string
	HasStagedChanges bool
	IsBrain          bool // gitx.IsBrainRepo — a brain is never auto-committed into (#176)
}

// PeerWriteDecision is one peer repo's verdict: commit the project-file edits
// (Commit true, Message set) or leave them uncommitted with the reason and the
// exact manual command that finishes the job.
type PeerWriteDecision struct {
	RepoDir    string
	Files      []string // repo-relative paths to stage
	Commit     bool
	Message    string // scoped commit message (set when Commit)
	Reason     string // why not committed (set when !Commit)
	NextAction string // exact manual command to finish (set when !Commit)
}

// planPeerWrites decides, per peer repo, whether git state authorizes an
// automatic scoped commit of the project-file edits. The current repo is
// omitted — its project edit rides the normal close commit. A peer commits
// only on main with a clean index and only if it is not a brain; otherwise
// the file is left updated but uncommitted and the operator gets the reason +
// exact next action. closingRef ("<repo>#<id>") labels the commit so the peer's
// history says which close produced it. Pure: no IO, deterministic output
// (sorted by RepoDir).
func planPeerWrites(edits map[string][]string, states map[string]RepoGitState, curRepoDir, closingRef string) []PeerWriteDecision {
	var out []PeerWriteDecision
	for repoDir, files := range edits {
		if repoDir == curRepoDir {
			continue
		}
		msg := fmt.Sprintf("project: close-time update (%s)", closingRef)
		manual := fmt.Sprintf("cd %s && git add %s && git commit -m %q", repoDir, strings.Join(files, " "), msg)
		d := PeerWriteDecision{RepoDir: repoDir, Files: files}
		st, known := states[repoDir]
		base := filepath.Base(repoDir)
		switch {
		case !known:
			d.Reason = fmt.Sprintf("%s git state is unknown — refusing to commit blind", base)
			d.NextAction = manual
		case st.IsBrain:
			d.Reason = fmt.Sprintf("%s is a brain capture repo — never auto-committed into (#176)", base)
			d.NextAction = manual
		case st.Branch != "main":
			d.Reason = fmt.Sprintf("%s is on branch %q, not main", base, st.Branch)
			d.NextAction = manual
		case st.HasStagedChanges:
			d.Reason = fmt.Sprintf("%s has pre-existing staged changes — refusing to absorb another session's index", base)
			d.NextAction = manual + "  # after handling your staged changes"
		default:
			d.Commit = true
			d.Message = msg
		}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RepoDir < out[j].RepoDir })
	return out
}

// readRepoGitState snapshots the git state the peer-write decision needs.
// `diff --cached --quiet` exits non-zero iff there are staged changes;
// execGitRunner surfaces that as err != nil (CombinedOutput).
func readRepoGitState(r gitRunner, repoDir string) RepoGitState {
	branch, _ := r.GitInDir(repoDir, "rev-parse", "--abbrev-ref", "HEAD")
	_, err := r.GitInDir(repoDir, "diff", "--cached", "--quiet")
	return RepoGitState{
		Branch:           strings.TrimSpace(string(branch)),
		HasStagedChanges: err != nil,
		IsBrain:          gitx.IsBrainRepo(repoDir),
	}
}

// applyPeerWrites executes the decisions: a committing decision stages exactly
// its Files and commits with the scoped message; a report-only decision prints
// the reason + next action. File writes already happened in applyClose — this
// only handles the commit. Failures warn and continue; the close never fails
// on a peer-write outcome.
func applyPeerWrites(r gitRunner, decisions []PeerWriteDecision, stdout, stderr io.Writer) {
	for _, d := range decisions {
		base := filepath.Base(d.RepoDir)
		if !d.Commit {
			cwarn(stderr, fmt.Sprintf("project updated but NOT committed in %s: %s\n      finish: %s", base, d.Reason, d.NextAction))
			continue
		}
		args := append([]string{"add", "--"}, d.Files...)
		if out, err := r.GitInDir(d.RepoDir, args...); err != nil {
			cwarn(stderr, fmt.Sprintf("git add in %s failed: %v — %s", base, err, strings.TrimSpace(string(out))))
			continue
		}
		commitArgs := []string{"commit", "-m", d.Message, "--"}
		commitArgs = append(commitArgs, d.Files...)
		if out, err := r.GitInDir(d.RepoDir, commitArgs...); err != nil {
			cwarn(stderr, fmt.Sprintf("git commit in %s failed: %v — %s", base, err, strings.TrimSpace(string(out))))
			continue
		}
		cinfo(stdout, fmt.Sprintf("committed %s in %s", strings.Join(d.Files, ", "), base))
	}
}
