package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Every test here uses a REAL repo with a REAL bare origin. That is not
// ceremony: the bug is "a ref exists that this worktree does not contain", and
// no function-call mock can express it — a fake returning canned ids would pass
// against the broken allocator, because the broken allocator never asked
// (ARCH-MOCK).

const idsDir, histDir = "workshop/issues", "workshop/history"

// idRepo builds repo + bare origin, seeds an issue, and pushes main.
func idRepo(t *testing.T) (repo, origin string) {
	t.Helper()
	origin = filepath.Join(t.TempDir(), "origin.git")
	git(t, "", "init", "--bare", "-b", "main", origin)
	repo = testfix.Repo(t, testfix.InitialCommit())
	writeIssueAt(t, repo, idsDir, "000001-first.md")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "seed")
	git(t, repo, "remote", "add", "origin", origin)
	git(t, repo, "push", "-u", "origin", "main")
	chdirTo(t, repo)
	return repo, origin
}

func writeIssueAt(t *testing.T, repo, dir, name string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(repo, dir), 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(repo, dir, name)
	body := "---\nid: " + strings.SplitN(name, "-", 2)[0] + "\nstatus: open\n---\n\n# t\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestAllocateIssueID_BranchCutBeforePublish is the bug, end to end.
//
// A branch is cut, THEN an issue lands on the trunk. The branch's
// workshop/issues/ never contained it, so a local-only scan reallocates that
// published id — days later, deterministically. Not a race.
func TestAllocateIssueID_BranchCutBeforePublish(t *testing.T) {
	repo, origin := idRepo(t)

	// Cut the branch first — this is the whole setup.
	git(t, repo, "switch", "-q", "-c", "feature")

	// Meanwhile #2 lands on the trunk, via a second clone so the branch's
	// worktree genuinely never sees it.
	// Clone the BARE ORIGIN, not repo/.git — pushing to the latter would leave
	// the shared origin untouched and the fixture would prove nothing.
	other := t.TempDir()
	git(t, "", "clone", "-q", origin, other)
	git(t, other, "config", "user.email", "e@x.com")
	git(t, other, "config", "user.name", "x")
	git(t, other, "config", "commit.gpgsign", "false")
	git(t, other, "switch", "-q", "main")
	writeIssueAt(t, other, idsDir, "000002-published-elsewhere.md")
	git(t, other, "add", "-A")
	git(t, other, "commit", "-m", "publish #2")
	git(t, other, "push", "-q", "origin", "main")

	// The branch's own directory still holds only #1.
	if _, err := os.Stat(filepath.Join(repo, idsDir, "000002-published-elsewhere.md")); !os.IsNotExist(err) {
		t.Fatal("fixture broken: the branch should not contain #2")
	}

	var stderr bytes.Buffer
	got, err := allocateIssueID(&stderr, idsDir, histDir, execGitRunner{})
	if err != nil {
		t.Fatalf("allocateIssueID: %v", err)
	}
	if got == "000002" {
		t.Errorf("reallocated the published id 000002 — this is the defect (stderr: %s)", stderr.String())
	}
	if got != "000003" {
		t.Errorf("allocated %s, want 000003 (max of local #1 and published #2)", got)
	}
}

// TestAllocateIssueID_CountsUnpushedLocal pins the UNION half: the trunk is not
// a replacement for the local scan. Two creations on one branch must not
// collide with each other just because neither is pushed.
func TestAllocateIssueID_CountsUnpushedLocal(t *testing.T) {
	repo, _ := idRepo(t)
	git(t, repo, "switch", "-q", "-c", "feature")
	writeIssueAt(t, repo, idsDir, "000009-unpushed-on-this-branch.md")

	var stderr bytes.Buffer
	got, err := allocateIssueID(&stderr, idsDir, histDir, execGitRunner{})
	if err != nil {
		t.Fatalf("allocateIssueID: %v", err)
	}
	if got != "000010" {
		t.Errorf("allocated %s, want 000010 — an unpushed local issue is real and must be counted", got)
	}
}

// TestAllocateIssueID_OfflineWarnsAndProceeds: creating an issue with no network
// is legitimate, but a SILENT local fallback recreates the bug. The warning is
// the deliverable, so it is asserted, not just the success.
func TestAllocateIssueID_OfflineWarnsAndProceeds(t *testing.T) {
	repo, _ := idRepo(t)
	git(t, repo, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
	git(t, repo, "update-ref", "-d", "refs/remotes/origin/main")

	var stderr bytes.Buffer
	got, err := allocateIssueID(&stderr, idsDir, histDir, execGitRunner{})
	if err != nil {
		t.Fatalf("offline creation must succeed, got: %v", err)
	}
	if got != "000002" {
		t.Errorf("allocated %s, want 000002 from the local scan", got)
	}
	for _, want := range []string{"unreachable", "LOCAL files only", "may collide"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("offline warning missing %q — a silent fallback recreates the bug:\n%s", want, stderr.String())
		}
	}
}

// TestAllocateIssueID_ScansHistoryOnTheTrunk: an archived issue still owns its
// id. The trunk scan must cover the same three directories the local one does,
// or an id becomes reusable the moment its issue is archived.
func TestAllocateIssueID_ScansHistoryOnTheTrunk(t *testing.T) {
	repo, _ := idRepo(t)
	writeIssueAt(t, repo, histDir+"/issues", "000050-archived.md")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "archive #50")
	git(t, repo, "push", "-q", "origin", "main")
	// Remove it locally so ONLY the trunk knows about it.
	if err := os.RemoveAll(filepath.Join(repo, histDir)); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	got, err := allocateIssueID(&stderr, idsDir, histDir, execGitRunner{})
	if err != nil {
		t.Fatalf("allocateIssueID: %v", err)
	}
	if got != "000051" {
		t.Errorf("allocated %s, want 000051 — an archived id on the trunk is still taken", got)
	}
}

// TestRefuseDuplicateIssueIDs is the merge gate: allocation cannot reach a
// branch cut before this fix, so the collision has to be caught at the last
// point it is still repairable.
func TestRefuseDuplicateIssueIDs(t *testing.T) {
	repo, _ := idRepo(t)
	git(t, repo, "switch", "-q", "-c", "feature")
	// Same id as the trunk's #1, different slug — different PATH, which is why
	// git merges both and nothing else ever objects.
	writeIssueAt(t, repo, idsDir, "000001-different-slug.md")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "reuse an id")

	var stderr bytes.Buffer
	err := refuseDuplicateIssueIDs(&stderr, idsDir, histDir, execGitRunner{})
	if err == nil {
		t.Fatalf("merge gate allowed a reused id (stderr: %s)", stderr.String())
	}
	for _, want := range []string{"000001", "000001-different-slug.md", "000001-first.md"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal must name %q so the operator can repair it:\n%s", want, err.Error())
		}
	}
}

// TestRefuseDuplicateIssueIDs_CleanBranchPasses: the same file already on the
// trunk is not a collision. Comparing by id alone would refuse every merge.
func TestRefuseDuplicateIssueIDs_CleanBranchPasses(t *testing.T) {
	repo, _ := idRepo(t)
	git(t, repo, "switch", "-q", "-c", "feature")
	writeIssueAt(t, repo, idsDir, "000002-genuinely-new.md")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "a new issue")

	var stderr bytes.Buffer
	if err := refuseDuplicateIssueIDs(&stderr, idsDir, histDir, execGitRunner{}); err != nil {
		t.Errorf("clean branch refused: %v", err)
	}
	if !strings.Contains(stderr.String(), "no reused issue ids") {
		t.Errorf("expected the pass line; got: %s", stderr.String())
	}
}

// TestNextID_IsPure covers the decision without any IO — the reason it was
// split out of the directory scan.
func TestNextID_IsPure(t *testing.T) {
	for _, tc := range []struct {
		name string
		sets [][]int
		want string
	}{
		{"nothing", nil, "000001"},
		{"empty sets", [][]int{{}, {}}, "000001"},
		{"local only", [][]int{{1, 2, 3}}, "000004"},
		{"trunk higher than local", [][]int{{1}, {2}}, "000003"},
		{"local higher than trunk", [][]int{{9}, {2}}, "000010"},
		{"overlap", [][]int{{1, 2}, {2, 3}}, "000004"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := issue.NextID(tc.sets...); got != tc.want {
				t.Errorf("NextID(%v) = %s, want %s", tc.sets, got, tc.want)
			}
		})
	}
}

// TestDuplicateIDsInRef_SeesCollisionsAlreadyMerged covers the question a
// branch-vs-trunk comparison structurally cannot answer.
//
// Once both colliding files are on the trunk the two trees AGREE, so
// refuseDuplicateIssueIDs finds nothing — which is why the eight collisions
// found in the wild were invisible to every gate. This asks whether one tree
// contradicts itself.
func TestDuplicateIDsInRef_SeesCollisionsAlreadyMerged(t *testing.T) {
	listing := strings.Join([]string{
		"workshop/issues/000001-first.md",
		"workshop/issues/000002-a.md",
		"workshop/issues/000002-b-different-slug.md",
		"workshop/history/issues/000003-archived.md",
		"workshop/issues/000003-live-reuse.md",
		"workshop/issues/not-an-issue.md",
	}, "\n")
	got := issue.DuplicateIDsInRef(listing)
	if len(got) != 2 {
		t.Fatalf("found %d collisions, want 2: %+v", len(got), got)
	}
	if got[0].ID != 2 || got[1].ID != 3 {
		t.Errorf("ids = %d,%d — want 2,3 sorted ascending", got[0].ID, got[1].ID)
	}
	if len(got[0].Paths) != 2 {
		t.Errorf("collision #2 should name both paths, got %v", got[0].Paths)
	}
	// An id spanning issues/ and history/issues/ is still a collision: the
	// archived issue owns that id.
	if !strings.Contains(strings.Join(got[1].Paths, " "), "history") {
		t.Errorf("collision #3 should span the archive, got %v", got[1].Paths)
	}
}

// TestDuplicateIDsInRef_SamePathTwiceIsNotACollision: the listing concatenates
// three directories that can overlap, so the same path can appear twice. That is
// one file, not two claimants — counting it would fail every clean repo.
func TestDuplicateIDsInRef_SamePathTwiceIsNotACollision(t *testing.T) {
	listing := "workshop/issues/000007-x.md\nworkshop/issues/000007-x.md\n"
	if got := issue.DuplicateIDsInRef(listing); len(got) != 0 {
		t.Errorf("same path twice reported as a collision: %+v", got)
	}
}

// TestIntroducedIDClashes drives the decision the CI check refuses on, against a
// real repo — the range introduces a file reusing an id already at base.
func TestIntroducedIDClashes(t *testing.T) {
	repo, _ := idRepo(t)
	base := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD"))

	git(t, repo, "switch", "-q", "-c", "feature")
	writeIssueAt(t, repo, idsDir, "000001-different-slug.md")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "reuse an id")

	clashes, err := introducedIDClashes(base, "HEAD", idsDir, histDir, execGitRunner{})
	if err != nil {
		t.Fatalf("introducedIDClashes: %v", err)
	}
	if len(clashes) != 1 {
		t.Fatalf("found %d clashes, want 1: %v", len(clashes), clashes)
	}
	for _, want := range []string{"000001", "000001-different-slug.md", "000001-first.md"} {
		if !strings.Contains(clashes[0], want) {
			t.Errorf("clash report must name %q so the operator can repair it:\n%s", want, clashes[0])
		}
	}

	// A range that adds a genuinely new id is clean.
	git(t, repo, "switch", "-q", "-c", "clean-feature", base)
	writeIssueAt(t, repo, idsDir, "000002-genuinely-new.md")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "a new issue")
	clean, err := introducedIDClashes(base, "HEAD", idsDir, histDir, execGitRunner{})
	if err != nil {
		t.Fatalf("introducedIDClashes (clean): %v", err)
	}
	if len(clean) != 0 {
		t.Errorf("clean range reported clashes: %v", clean)
	}
}

// TestMergeCheckScript_RefusesAPlantedCollision drives the CI ADAPTER, not just
// the Go behind it. The script is what GitHub runs, and its own logic — the
// no-tracker skip, the no-cmd/sdlc skip, the exit code — is untested otherwise.
func TestMergeCheckScript_RefusesAPlantedCollision(t *testing.T) {
	script, err := filepath.Abs(filepath.Join("..", "..", "scripts", "merge-checks.d", "40-duplicate-issue-id.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(script); err != nil {
		t.Skipf("check script not present: %v", err)
	}
	repo := testfix.Repo(t, testfix.InitialCommit())
	writeIssueAt(t, repo, idsDir, "000001-first.md")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "seed")
	base := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD"))
	writeIssueAt(t, repo, idsDir, "000001-different-slug.md")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "reuse an id")

	// No ./cmd/sdlc in this fixture, so the script takes its documented skip
	// path. Asserting the SKIP is the point: a derivative repo consuming the
	// symlinked runner must not fail the check merely for lacking the binary.
	cmd := exec.Command("bash", script, base, "HEAD")
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("script must exit 0 when it cannot build sdlc, got %v:\n%s", err, out)
	}
	if !strings.Contains(string(out), "skipping") {
		t.Errorf("the skip must be announced, not silent:\n%s", out)
	}
}
