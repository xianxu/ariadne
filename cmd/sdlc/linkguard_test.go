package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/linkguard"
)

const sideDocQ = "workshop/issues/000009-s09.side/2026-10/2026-10-23-01-why.md"
const sideDocNested = "workshop/issues/000009-s09.side/2026-10/2026-10-23-02-how.md"

// #579 (D1): an issue shaped on the resting branch links a side doc, which
// links a further one; publish carries both with the details in its one main
// commit and leaves the rest equal to main and clean.
func TestIssuePublishCarriesLinkedSideDocs(t *testing.T) {
	r, _ := claimSetRepo(t)
	claimFor(t, 9)
	writeRepoFile(t, r.root, sideDocQ, "---\ntype: side-doc\nparent: ../../000009-s09.md\n---\n# Why?\n\n→ [discussion](2026-10-23-02-how.md)\n")
	writeRepoFile(t, r.root, sideDocNested, "# How?\n")
	writeRepoFile(t, r.root, "workshop/issues/000009-s09.side/2026-10/2026-10-23-03-unlinked.md", "# not linked\n")
	appendDetail(t, r, "workshop/issues/000009-s09.md", "A question → [discussion](000009-s09.side/2026-10/2026-10-23-01-why.md)")
	before := r.originMain()
	if out, err := publish(t, 9); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if parent := r.git("rev-parse", r.originMain()+"^"); parent != before {
		t.Fatal("the publication took more than one main commit")
	}
	for _, p := range []string{sideDocQ, sideDocNested} {
		if r.git("cat-file", "-t", r.originMain()+":"+p) != "blob" {
			t.Fatalf("main lacks %s", p)
		}
	}
	if out := r.git("ls-tree", "--name-only", r.originMain(), "--", "workshop/issues/000009-s09.side/2026-10/2026-10-23-03-unlinked.md"); out != "" {
		t.Fatal("publish carried an unlinked side doc")
	}
	if r.git("rev-parse", "HEAD") != r.originMain() {
		t.Fatal("rest not fast-forwarded to main")
	}
	if st := r.git("status", "--porcelain"); st != "?? workshop/issues/000009-s09.side/2026-10/2026-10-23-03-unlinked.md" {
		t.Fatalf("rest not clean apart from the unlinked doc:\n%s", st)
	}
	// A rerun with the side doc edited carries the edit.
	writeRepoFile(t, r.root, sideDocNested, "# How?\n\nAnswered.\n")
	appendDetail(t, r, "workshop/issues/000009-s09.md", "More.")
	if out, err := publish(t, 9); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(r.git("show", r.originMain()+":"+sideDocNested), "Answered.") {
		t.Fatal("main lacks the side doc's edit")
	}
}

// #579: publish refuses, before any effect, a link it cannot carry: a file
// outside the issue's side folder that main lacks, an ignored file, or a
// symlink in the side folder. Removing the link lets it publish.
func TestIssuePublishRefusesUncommittedLink(t *testing.T) {
	for name, setup := range map[string]func(t *testing.T, r *trackerRepo) string{
		"outside the side folder": func(t *testing.T, r *trackerRepo) string {
			writeRepoFile(t, r.root, "workshop/notes/draft.md", "# draft\n")
			return "../notes/draft.md"
		},
		"ignored": func(t *testing.T, r *trackerRepo) string {
			writeRepoFile(t, r.root, ".git/info/exclude", "session.side/\n")
			writeRepoFile(t, r.root, "workshop/issues/session.side/q.md", "# private\n")
			return "session.side/q.md"
		},
		"a symlink in the side folder": func(t *testing.T, r *trackerRepo) string {
			writeRepoFile(t, r.root, "outside.md", "# x\n")
			dir := filepath.Join(r.root, "workshop/issues/000009-s09.side")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(r.root, "outside.md"), filepath.Join(dir, "l.md")); err != nil {
				t.Fatal(err)
			}
			return "000009-s09.side/l.md"
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, _ := claimSetRepo(t)
			claimFor(t, 9)
			target := setup(t, r)
			original := r.git("show", "HEAD:workshop/issues/000009-s09.md")
			appendDetail(t, r, "workshop/issues/000009-s09.md", "See → [discussion]("+target+")")
			before := r.originMain()
			out, err := publish(t, 9)
			if err == nil || !strings.Contains(err.Error(), "links to files that will not reach the repository") ||
				!strings.Contains(err.Error(), "nothing was published") || r.originMain() != before {
				t.Fatalf("want a refusal publishing nothing: %v\n%s", err, out)
			}
			writeRepoFile(t, r.root, "workshop/issues/000009-s09.md", original+"\nNo link now.\n")
			if out, err := publish(t, 9); err != nil {
				t.Fatalf("without the link: %v\n%s", err, out)
			}
		})
	}
}

// #579: change-code and close refuse a link to an uncommitted or ignored
// file in the issue or its plan, and pass once it is committed or removed.
func TestIssueSideLinkRefusals(t *testing.T) {
	r, _ := claimSetRepo(t)
	t.Chdir(r.root)
	identity, err := resolveWorkspace(r.root)
	if err != nil {
		t.Fatal(err)
	}
	issueRel := "workshop/issues/000009-s09.md"
	planRel := "workshop/plans/000009-s09-plan.md"
	writeRepoFile(t, r.root, ".git/info/exclude", "session.side/\n")
	writeRepoFile(t, r.root, "session.side/private.md", "# private\n")
	writeRepoFile(t, r.root, sideDocQ, "# q\n")
	writeRepoFile(t, r.root, planRel, "# plan\n\nSee [private](../../session.side/private.md).\n")
	appendDetail(t, r, issueRel, "→ [discussion](000009-s09.side/2026-10/2026-10-23-01-why.md)")
	refusals := func() []linkguard.Refusal {
		t.Helper()
		rs, err := issueSideLinkRefusals(execSideLinkGit, r.root, issueRel, "workshop/plans", identity)
		if err != nil {
			t.Fatal(err)
		}
		return rs
	}
	rs := refusals()
	if len(rs) != 2 || rs[0].Doc != issueRel || rs[0].State != linkguard.Uncommitted || rs[1].Doc != planRel || rs[1].State != linkguard.Ignored {
		t.Fatalf("want the issue's uncommitted side doc and the plan's ignored link: %+v", rs)
	}
	if err := changeCodeSideLinks(issueRel, "workshop/plans"); err == nil || !strings.Contains(err.Error(), "--no-side-links (or --force <reason>)") {
		t.Fatalf("change-code: %v", err)
	}
	r.git("add", sideDocQ)
	r.git("commit", "-q", "-m", "side doc")
	writeRepoFile(t, r.root, planRel, "# plan\n\nFolded in.\n")
	if rs := refusals(); len(rs) != 0 {
		t.Fatalf("committed and folded, still refused: %+v", rs)
	}
	if err := changeCodeSideLinks(issueRel, "workshop/plans"); err != nil {
		t.Fatal(err)
	}
}

// #579: close refuses an issue linking an ignored file, closes with
// --no-side-links (acknowledged), and closes once the link is folded in.
func TestCloseRefusesIgnoredSideLink(t *testing.T) {
	issuesDir := closeRepo(t, 187)
	t.Setenv("WF_CALIB_LEDGER", "")
	issuePath := mustIssuePath(t, issuesDir, 187)
	writeRepoFile(t, ".", ".git/info/exclude", "session.side/\n")
	writeRepoFile(t, ".", "session.side/q.md", "# private\n")
	body, err := os.ReadFile(issuePath)
	if err != nil {
		t.Fatal(err)
	}
	linked := strings.Replace(string(body), "Thing.", "Thing → [discussion](../../session.side/q.md).", 1)
	writeRepoFile(t, ".", issuePath, linked)
	flags := func() *closeFlags {
		return &closeFlags{Issue: 187, Actual: "1.0", Verified: "done", NoAtlas: true, NoJudge: true, IssuesDir: issuesDir, BrainDir: "../nonexistent-brain"}
	}
	msg, died := expectDie(t, func() { _ = runClose(io.Discard, io.Discard, flags()) })
	if !died || !strings.Contains(msg, `links "../../session.side/q.md", which is git-ignored`) || !strings.Contains(msg, "--no-side-links (or --force)") {
		t.Fatalf("want the side-link refusal, got died=%v %q", died, msg)
	}
	f := flags()
	f.NoSideLinks = true
	f.DryRun = true
	var errb bytes.Buffer
	if err := runClose(io.Discard, &errb, f); err != nil || !strings.Contains(errb.String(), "--no-side-links (or --force): closing with 1 link") {
		t.Fatalf("bypass: %v\n%s", err, errb.String())
	}
	writeRepoFile(t, ".", issuePath, string(body))
	f = flags()
	f.DryRun = true
	if err := runClose(io.Discard, io.Discard, f); err != nil {
		t.Fatalf("folded in: %v", err)
	}
}
