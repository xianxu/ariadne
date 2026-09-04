package main

import (
	"bytes"
	"fmt"
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

	clashes, err := introducedIDClashes(base, "HEAD", base, idsDir, histDir, execGitRunner{})
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
	clean, err := introducedIDClashes(base, "HEAD", base, idsDir, histDir, execGitRunner{})
	if err != nil {
		t.Fatalf("introducedIDClashes (clean): %v", err)
	}
	if len(clean) != 0 {
		t.Errorf("clean range reported clashes: %v", clean)
	}
}

// TestMergeCheckScript_RefusesGivenMergeBase is the CI adapter's refusal path
// (#213 close review BR-5 — the previous test asserted only the SKIP path, so
// the check could have been inert and still "covered").
//
// It also pins BR-1, the reason the check did not work at all: the runner
// contract hands merge-base(base, head), and comparing against THAT cannot see
// this collision — the branch was cut before the colliding id was published, so
// the merge-base predates that file and the id looks new on both sides. The
// script must resolve the trunk TIP itself. The fixture is built in that exact
// order: cut, publish, collide.
func TestMergeCheckScript_RefusesGivenMergeBase(t *testing.T) {
	script, err := filepath.Abs(filepath.Join("..", "..", "scripts", "merge-checks.d", "40-duplicate-issue-id.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if _, serr := os.Stat(script); serr != nil {
		t.Skipf("check script not present: %v", serr)
	}
	ariadne, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	origin := filepath.Join(t.TempDir(), "origin.git")
	git(t, "", "init", "--bare", "-b", "main", origin)
	repo := testfix.Repo(t, testfix.InitialCommit())
	writeIssueAt(t, repo, idsDir, "000001-seed.md")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "seed")
	git(t, repo, "remote", "add", "origin", origin)
	git(t, repo, "push", "-u", "origin", "main")

	git(t, repo, "switch", "-q", "-c", "feature") // cut BEFORE the publish
	git(t, repo, "switch", "-q", "main")
	writeIssueAt(t, repo, idsDir, "000500-theirs.md")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "publish 500")
	git(t, repo, "push", "-q", "origin", "main")

	git(t, repo, "switch", "-q", "feature")
	writeIssueAt(t, repo, idsDir, "000500-mine.md")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "reuse 500")

	// A derivative resolves sdlc through construct/dev-aliases.sh rather than
	// owning cmd/sdlc — the shape BR-6 was about.
	if mkErr := os.MkdirAll(filepath.Join(repo, "construct"), 0o755); mkErr != nil {
		t.Fatal(mkErr)
	}
	resolver := "#!/usr/bin/env bash\n[ \"${1:-}\" = \"--list\" ] && printf 'sdlc\\t" + ariadne + "\\n'\n"
	if wErr := os.WriteFile(filepath.Join(repo, "construct", "dev-aliases.sh"), []byte(resolver), 0o755); wErr != nil {
		t.Fatal(wErr)
	}

	mergeBase := strings.TrimSpace(git(t, repo, "merge-base", "main", "HEAD"))
	cmd := exec.Command("bash", script, mergeBase, "HEAD")
	cmd.Dir = repo
	// The script mktemp's a build dir; point it somewhere writable so a
	// restricted default TMPDIR makes it SKIP rather than exercising the check.
	cmd.Env = envWithTMPDIR(t)
	out, runErr := cmd.CombinedOutput()

	if runErr == nil {
		t.Fatalf("check PASSED a reused id — comparing against merge-base cannot see this collision:\n%s", out)
	}
	for _, want := range []string{"000500", "000500-mine.md", "000500-theirs.md"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("refusal must name %q so the operator can repair it:\n%s", want, out)
		}
	}
}

// TestMergeCheckScript_SkipsWhenSdlcIsUnresolvable pins the other half: a repo
// where sdlc cannot be located exits 0 with an ANNOUNCED skip, so a tracker-less
// or unbootstrapped derivative is not failed for lacking the binary.
func TestMergeCheckScript_SkipsWhenSdlcIsUnresolvable(t *testing.T) {
	script, err := filepath.Abs(filepath.Join("..", "..", "scripts", "merge-checks.d", "40-duplicate-issue-id.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if _, serr := os.Stat(script); serr != nil {
		t.Skipf("check script not present: %v", serr)
	}
	repo := testfix.Repo(t, testfix.InitialCommit())
	writeIssueAt(t, repo, idsDir, "000001-seed.md")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "seed")

	cmd := exec.Command("bash", script, "", "HEAD")
	cmd.Dir = repo
	cmd.Env = envWithTMPDIR(t)
	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		t.Errorf("must exit 0 when sdlc is unresolvable, got %v:\n%s", runErr, out)
	}
	if !strings.Contains(string(out), "SKIPPING") && !strings.Contains(string(out), "skipping") {
		t.Errorf("the skip must be announced, not silent:\n%s", out)
	}
}

// envWithTMPDIR REPLACES TMPDIR rather than appending a second one — the script
// mktemp's a build dir, and appending left the restricted default in force, so
// the check skipped and the test read as a pass against broken code.
func envWithTMPDIR(t *testing.T) []string {
	t.Helper()
	dir := t.TempDir()
	out := []string{"TMPDIR=" + dir, "TMP=" + dir, "TEMP=" + dir}
	for _, kv := range os.Environ() {
		switch {
		case strings.HasPrefix(kv, "TMPDIR="), strings.HasPrefix(kv, "TMP="), strings.HasPrefix(kv, "TEMP="):
			continue
		}
		out = append(out, kv)
	}
	return out
}

// TestIntroducedIDClashes_IndependentOfSlugSortOrder is BR-2: issueFilesByID
// kept only the FIRST path per id, so when the head tree carried BOTH files —
// a rebased PR, or any branch that pulled main after the trunk file landed —
// head[id] could equal base[id] and nothing was reported.
//
// Detection depended on which slug `ls-tree` sorted first. Measured on the real
// repo at the time: `000213-aaa-collision.md` was refused, while an otherwise
// identical `000213-planted-collision.md` passed. Both orders are asserted here.
func TestIntroducedIDClashes_IndependentOfSlugSortOrder(t *testing.T) {
	for _, slug := range []string{"000001-aaa-sorts-first.md", "000001-zzz-sorts-last.md"} {
		t.Run(slug, func(t *testing.T) {
			repo, _ := idRepo(t) // trunk carries 000001-first.md
			base := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD"))
			git(t, repo, "switch", "-q", "-c", "feature")
			// The head tree keeps the trunk's file AND adds a colliding one —
			// the rebased-PR shape where the first-path-only map hid the clash.
			writeIssueAt(t, repo, idsDir, slug)
			git(t, repo, "add", "-A")
			git(t, repo, "commit", "-m", "reuse an id")

			clashes, err := introducedIDClashes(base, "HEAD", base, idsDir, histDir, execGitRunner{})
			if err != nil {
				t.Fatalf("introducedIDClashes: %v", err)
			}
			if len(clashes) != 1 {
				t.Fatalf("found %d clashes, want 1 — detection must not depend on slug order: %v", len(clashes), clashes)
			}
			if !strings.Contains(clashes[0], slug) {
				t.Errorf("clash must name the introduced file %q:\n%s", slug, clashes[0])
			}
		})
	}
}

// TestPublishedIssueIDs_PartialReadIsAnError is BR-7: a per-directory ls-tree
// failure used to be swallowed, so a partial trunk read was indistinguishable
// from a complete one — and allocation would hand out a colliding id while
// reporting success. Failing loudly routes it to the offline warning instead.
func TestPublishedIssueIDs_PartialReadIsAnError(t *testing.T) {
	repo, _ := idRepo(t)
	_ = repo
	// A ref that exists but whose ls-tree calls fail: point at a bogus ref via a
	// runner that answers rev-parse but errors on ls-tree.
	r := &lsTreeFailRunner{}
	if _, err := publishedIssueIDs(idsDir, histDir, r); err == nil {
		t.Error("a failed ls-tree must be an error — a partial trunk read allocates colliding ids silently")
	}
}

type lsTreeFailRunner struct{ execGitRunner }

func (l *lsTreeFailRunner) Git(args ...string) ([]byte, error) {
	if len(args) > 0 && args[0] == "ls-tree" {
		return []byte("boom"), fmt.Errorf("simulated ls-tree failure")
	}
	return l.execGitRunner.Git(args...)
}

// TestIntroducedCollisions_ArchiveAndRenameAreNotCollisions is BR-13, the
// Critical the close review caught: the first definition of "collision" was
// "head has a path base lacks", which is what EVERY archive looks like —
// `sdlc merge` moves workshop/issues/NNN-x.md to history/issues/NNN-x.md on
// every close. The gate refused a routine archive, which would have broken
// nearly every merge in the fleet. Renames and renumbers have the same shape.
//
// A collision is one TREE contradicting itself: two live paths claiming one id.
func TestIntroducedCollisions_ArchiveAndRenameAreNotCollisions(t *testing.T) {
	for _, tc := range []struct {
		name string
		base map[int][]string
		head map[int][]string
		want []int
	}{
		{"archive: issues/ -> history/issues/",
			map[int][]string{1: {"workshop/issues/000001-a.md"}},
			map[int][]string{1: {"workshop/history/issues/000001-a.md"}},
			nil},
		{"rename: slug changed",
			map[int][]string{1: {"workshop/issues/000001-old-slug.md"}},
			map[int][]string{1: {"workshop/issues/000001-new-slug.md"}},
			nil},
		{"renumber: id changed",
			map[int][]string{1: {"workshop/issues/000001-a.md"}},
			map[int][]string{9: {"workshop/issues/000009-a.md"}},
			nil},
		{"a genuinely new issue",
			map[int][]string{1: {"workshop/issues/000001-a.md"}},
			map[int][]string{1: {"workshop/issues/000001-a.md"}, 2: {"workshop/issues/000002-b.md"}},
			nil},
		{"COLLISION: two live paths for one id",
			map[int][]string{1: {"workshop/issues/000001-a.md"}},
			map[int][]string{1: {"workshop/issues/000001-a.md", "workshop/issues/000001-b.md"}},
			[]int{1}},
		{"pre-existing: base already had two, so not introduced",
			map[int][]string{1: {"workshop/issues/000001-a.md", "workshop/issues/000001-b.md"}},
			map[int][]string{1: {"workshop/issues/000001-a.md", "workshop/issues/000001-b.md"}},
			nil},
		{"pre-existing that GREW is still introduced-at-this-id? no — base was already dirty",
			map[int][]string{1: {"a", "b"}},
			map[int][]string{1: {"a", "b", "c"}},
			nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := introducedCollisions(tc.head, tc.base, tc.base)
			if len(got) != len(tc.want) {
				t.Fatalf("introducedCollisions = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("introducedCollisions = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

// TestMergedPathsFor_ModelsTheMergeResult pins the three-tree model directly.
// Two earlier definitions of "collision" were wrong in opposite directions —
// one refused every archive, the next was blind to the collision it existed for
// — because both asked about a single tree. A merge gate has to ask what the
// MERGE RESULT looks like, which needs merge-base (to see deletions), head, and
// the trunk tip (where the colliding file actually lives).
func TestMergedPathsFor_ModelsTheMergeResult(t *testing.T) {
	for _, tc := range []struct {
		name              string
		base, head, trunk map[int][]string
		wantIDs           []int // ids the merge result would have >1 path for
	}{
		{
			name:  "archive: head moves the file, trunk still has the old path",
			base:  map[int][]string{1: {"workshop/issues/000001-a.md"}},
			head:  map[int][]string{1: {"workshop/history/issues/000001-a.md"}},
			trunk: map[int][]string{1: {"workshop/issues/000001-a.md"}},
			// the move deletes the old path — one survivor
		},
		{
			name:  "renumber: id 1 vacated, id 9 created",
			base:  map[int][]string{1: {"workshop/issues/000001-a.md"}},
			head:  map[int][]string{9: {"workshop/issues/000009-a.md"}},
			trunk: map[int][]string{1: {"workshop/issues/000001-a.md"}},
		},
		{
			name:  "THE BUG: branch cut before the trunk published the same id",
			base:  map[int][]string{},
			head:  map[int][]string{500: {"workshop/issues/000500-mine.md"}},
			trunk: map[int][]string{500: {"workshop/issues/000500-theirs.md"}},
			// the branch never contains theirs — only the MERGE reveals it
			wantIDs: []int{500},
		},
		{
			// BR-18, and this table asserted the WRONG answer first: I encoded
			// the false refusal as expected behaviour. Honouring only head's
			// deletions meant every PR open across a close on main was refused —
			// most PRs. Deletions count from EITHER side.
			name:  "trunk archived it while the PR was open",
			base:  map[int][]string{7: {"workshop/issues/000007-x.md"}},
			head:  map[int][]string{7: {"workshop/issues/000007-x.md"}},
			trunk: map[int][]string{7: {"workshop/history/issues/000007-x.md"}},
		},
		{
			name:    "both sides ADD a file for one id — neither deleted anything",
			base:    map[int][]string{},
			head:    map[int][]string{8: {"workshop/issues/000008-mine.md"}},
			trunk:   map[int][]string{8: {"workshop/issues/000008-theirs.md"}},
			wantIDs: []int{8},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := introducedCollisions(tc.head, tc.base, tc.trunk)
			if len(got) != len(tc.wantIDs) {
				t.Fatalf("introducedCollisions = %v, want %v\n  merged = %v",
					got, tc.wantIDs, mergedPathsFor(tc.head, tc.base, tc.trunk))
			}
			for i := range got {
				if got[i] != tc.wantIDs[i] {
					t.Fatalf("introducedCollisions = %v, want %v", got, tc.wantIDs)
				}
			}
		})
	}
}
