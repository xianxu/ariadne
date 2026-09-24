package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

const laIssue = "workshop/issues/000246-landing.md"
const laHistory = "workshop/history/issues/000246-landing.md"

func laBody(status, body string) string {
	return "---\nid: 000246\nstatus: " + status + "\nactual_hours: 1\nupdated: 2026-09-20\n---\n# Landing\n\n" + body + "\n"
}
func laGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	return strings.TrimSpace(testfix.Capture(t, root, args...))
}
func laWrite(t *testing.T, root, path, body string) {
	t.Helper()
	mkArtifact(t, filepath.Join(root, path), body)
}
func laCommit(t *testing.T, root, msg string) string {
	t.Helper()
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", msg)
	return laGit(t, root, "rev-parse", "HEAD")
}
func laFixture(t *testing.T) (root, origin string, pr landingPR) {
	t.Helper()
	root = testfix.Repo(t, testfix.InitialCommit())
	origin = filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, root, "remote", "add", "origin", origin)
	laWrite(t, root, laIssue, laBody("working", "base"))
	laWrite(t, root, "workshop/issues/000247-other.md", strings.Replace(laBody("codecomplete", "unrelated"), "000246", "000247", 1))
	base := laCommit(t, root, "base")
	testfix.Git(t, root, "push", "-q", "origin", "HEAD:main")
	testfix.Git(t, root, "checkout", "-qb", "000246-landing")
	laWrite(t, root, laIssue, laBody("codecomplete", "reviewed"))
	laWrite(t, root, "workshop/plans/000246-landing-plan.md", "the plan\n")
	laWrite(t, root, "workshop/plans/000246-landing-close-review.md", "the review\n")
	head := laCommit(t, root, "#246: close")
	testfix.Git(t, root, "push", "-q", "origin", "HEAD:main")
	return root, origin, landingPR{Number: 246, State: "MERGED", Repo: "test/repo", HeadRef: "000246-landing", HeadOID: head, BaseRef: "main", BaseOID: base, MergeOID: head}
}
func laArchive(root string, pr landingPR) error {
	return archiveLandingPR(root, "origin", pr.Repo, pr, "workshop/issues", "workshop/plans", "workshop/history")
}
func laProof(root, tip string, pr landingPR) (bool, error) {
	return landingArchiveComplete(root, tip, pr.Repo, pr, "workshop/issues", "workshop/plans", "workshop/history")
}

func TestLandingArchiveAtomicAndRetryProof(t *testing.T) {
	root, origin, pr := laFixture(t)
	head := laGit(t, root, "rev-parse", "HEAD")
	index := testfix.Capture(t, root, "ls-files", "--stage")
	laWrite(t, root, laIssue, "local dirty preserved\n")
	before := laGit(t, origin, "rev-list", "--count", "main")
	if complete, err := laProof(root, pr.MergeOID, pr); err != nil || complete {
		t.Fatalf("premature proof %v %v", complete, err)
	}
	if err := laArchive(root, pr); err != nil {
		t.Fatal(err)
	}
	tip := laGit(t, origin, "rev-parse", "main")
	if laGit(t, origin, "rev-parse", "main^") != pr.MergeOID {
		t.Fatal("archive not single atomic commit")
	}
	if laGit(t, origin, "rev-list", "--count", "main") == before {
		t.Fatal("archive no commit")
	}
	files := testfix.Capture(t, origin, "ls-tree", "-r", "--name-only", "main")
	for _, want := range []string{laHistory, "workshop/history/plans/000246-landing-plan.md", "workshop/history/plans/000246-landing-close-review.md", "workshop/issues/000247-other.md"} {
		if !strings.Contains(files, want) {
			t.Fatalf("missing %s: %s", want, files)
		}
	}
	if strings.Contains(files, laIssue) || strings.Contains(files, "workshop/plans/000246-") {
		t.Fatal("active artifacts remain")
	}
	if b := testfix.Capture(t, origin, "show", "main:"+laHistory); !strings.Contains(b, "status: done") || !strings.Contains(b, "reviewed") {
		t.Fatal(b)
	}
	if complete, err := laProof(root, tip, pr); err != nil || !complete {
		t.Fatalf("proof %v %v", complete, err)
	}
	if err := laArchive(root, pr); err != nil {
		t.Fatal(err)
	}
	if laGit(t, origin, "rev-parse", "main") != tip {
		t.Fatal("retry duplicated archive")
	}
	if laGit(t, root, "rev-parse", "HEAD") != head || testfix.Capture(t, root, "ls-files", "--stage") != index {
		t.Fatal("local refs/index changed")
	}
	if b, _ := os.ReadFile(filepath.Join(root, laIssue)); string(b) != "local dirty preserved\n" {
		t.Fatal("local dirt lost")
	}
}
func TestLandingArchiveIndependentlyPublishedBody(t *testing.T) {
	root, origin, pr := laFixture(t)
	// A different commit with the same document body becomes the PR base. A tree
	// diff would miss the owned close; ancestry retains it.
	testfix.Git(t, root, "checkout", "-qb", "published-copy", pr.BaseOID)
	laWrite(t, root, laIssue, laBody("codecomplete", "reviewed"))
	pr.BaseOID = laCommit(t, root, "independent documentation publication")
	testfix.Git(t, root, "merge", "--no-ff", "-m", "integration", pr.HeadOID)
	pr.MergeOID = laGit(t, root, "rev-parse", "HEAD")
	testfix.Git(t, root, "push", "--force", "-q", "origin", "HEAD:main")
	if err := laArchive(root, pr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(testfix.Capture(t, origin, "ls-tree", "-r", "--name-only", "main"), laHistory) {
		t.Fatal("independently published body hid owning close")
	}
}
func TestLandingArchiveCollisionAndReopen(t *testing.T) {
	for _, kind := range []string{"occupied", "reopened", "wrong-id", "symlink", "reclosed-generation", "missing-plan", "rewound-main"} {
		t.Run(kind, func(t *testing.T) {
			root, origin, pr := laFixture(t)
			switch kind {
			case "reclosed-generation":
				laWrite(t, root, laIssue, strings.Replace(laBody("codecomplete", "reviewed"), "actual_hours: 1", "actual_hours: 2\nstarted: 2026-09-22", 1))
			case "missing-plan":
				testfix.Git(t, root, "rm", "workshop/plans/000246-landing-plan.md")
			case "rewound-main":
				testfix.Git(t, root, "checkout", "-qb", "independent-copy", pr.BaseOID)
				testfix.Git(t, root, "checkout", pr.HeadOID, "--", "workshop")
			case "occupied":
				laWrite(t, root, laHistory, laBody("done", "old archive"))
			case "reopened":
				laWrite(t, root, laIssue, laBody("working", "new work"))
			case "wrong-id":
				laWrite(t, root, laIssue, strings.Replace(laBody("codecomplete", "reviewed"), "000246", "000248", 1))
			case "symlink":
				if err := os.Remove(filepath.Join(root, laIssue)); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("outside", filepath.Join(root, laIssue)); err != nil {
					t.Fatal(err)
				}
			}
			laCommit(t, root, "concurrent change")
			testfix.Git(t, root, "push", "-q", "--force", "origin", "HEAD:main")
			before := laGit(t, origin, "rev-parse", "main")
			if err := laArchive(root, pr); err == nil {
				t.Fatal("unsafe archive accepted")
			}
			if laGit(t, origin, "rev-parse", "main") != before {
				t.Fatal("ref changed on refusal")
			}
		})
	}
}
func TestLandingArchiveEmptyAndUnsafeRoots(t *testing.T) {
	root, origin, pr := laFixture(t)
	pr.BaseOID = pr.HeadOID
	tip := laGit(t, origin, "rev-parse", "main")
	if done, err := laProof(root, tip, pr); err != nil || !done {
		t.Fatalf("empty proof %v %v", done, err)
	}
	if err := laArchive(root, pr); err != nil {
		t.Fatal(err)
	}
	if laGit(t, origin, "rev-parse", "main") != tip {
		t.Fatal("empty archive committed")
	}
	if _, err := landingArchiveComplete(root, tip, pr.Repo, pr, "../issues", "workshop/plans", "workshop/history"); err == nil {
		t.Fatal("outside root accepted")
	}
}

// Keep these imports available for the controlled publisher tests below.

func laPeer(t *testing.T, origin string) string {
	t.Helper()
	peer := testfix.Repo(t)
	testfix.Git(t, peer, "remote", "add", "origin", origin)
	testfix.Git(t, peer, "fetch", "-q", "origin", "main")
	testfix.Git(t, peer, "checkout", "-qB", "main", "origin/main")
	return peer
}
func TestLandingArchiveProofRejectsWrongOrChangedGeneration(t *testing.T) {
	for _, kind := range []string{"wrong-provenance", "partial", "changed-history", "reopened", "renamed-reopen", "new-sidecar", "unrelated"} {
		t.Run(kind, func(t *testing.T) {
			root, origin, pr := laFixture(t)
			if err := laArchive(root, pr); err != nil {
				t.Fatal(err)
			}
			peer := laPeer(t, origin)
			switch kind {
			case "wrong-provenance":
				testfix.Git(t, peer, "commit", "--amend", "-qm", "same filenames without provenance")
			case "partial":
				testfix.Git(t, peer, "rm", "workshop/history/plans/000246-landing-plan.md")
				testfix.Git(t, peer, "commit", "--amend", "--no-edit", "-q")
			case "changed-history":
				laWrite(t, peer, laHistory, laBody("done", "later changed archive"))
				laCommit(t, peer, "changed history")
			case "reopened":
				laWrite(t, peer, laIssue, laBody("working", "reopened"))
				laCommit(t, peer, "reopened")
			case "renamed-reopen":
				laWrite(t, peer, "workshop/issues/000246-renamed.md", laBody("working", "reopened"))
				laCommit(t, peer, "renamed reopen")
			case "new-sidecar":
				laWrite(t, peer, "workshop/plans/000246-extra.md", "later generation")
				laCommit(t, peer, "new sidecar")
			case "unrelated":
				laWrite(t, peer, "workshop/issues/000249-other.md", strings.Replace(laBody("working", "other"), "000246", "000249", 1))
				laCommit(t, peer, "other work")
			}
			testfix.Git(t, peer, "push", "--force", "-q", "origin", "main")
			testfix.Git(t, root, "fetch", "-q", "origin", "main")
			tip := laGit(t, origin, "rev-parse", "main")
			done, err := laProof(root, tip, pr)
			if kind == "unrelated" {
				if err != nil || !done {
					t.Fatalf("unrelated work invalidated proof: %v %v", done, err)
				}
			} else if err == nil || done {
				t.Fatalf("%s accepted: %v %v", kind, done, err)
			}
			if laGit(t, origin, "rev-parse", "main") != tip {
				t.Fatal("read-only proof published")
			}
		})
	}
}

type laPublisher struct {
	tf           *gitx.TrunkFile
	afterPrepare func()
	unknown      bool
	calls        int
}

func (p *laPublisher) UpdateMany(msg string, prepare func(*gitx.TrunkView) (gitx.TrunkWrite, error)) error {
	err := p.tf.UpdateMany(msg, func(v *gitx.TrunkView) (gitx.TrunkWrite, error) {
		p.calls++
		set, err := prepare(v)
		if err == nil && p.calls == 1 && p.afterPrepare != nil {
			p.afterPrepare()
		}
		return set, err
	})
	if err == nil && p.unknown {
		return gitx.ErrPublicationUncertain
	}
	return err
}
func TestLandingArchiveFreshRetryAndUnknown(t *testing.T) {
	for _, unknown := range []bool{false, true} {
		t.Run(map[bool]string{false: "fresh-edit", true: "unknown"}[unknown], func(t *testing.T) {
			root, origin, pr := laFixture(t)
			tf, err := gitx.NewTrunkFile(root, "origin", "main")
			if err != nil {
				t.Fatal(err)
			}
			pub := &laPublisher{tf: tf, unknown: unknown}
			if !unknown {
				pub.afterPrepare = func() {
					peer := laPeer(t, origin)
					laWrite(t, peer, laIssue, laBody("codecomplete", "compatible concurrent documentation"))
					laCommit(t, peer, "document update")
					testfix.Git(t, peer, "push", "-q", "origin", "main")
				}
			}
			orig := newLandingArchivePublisher
			newLandingArchivePublisher = func(string, string) (trunkPublisher, error) { return pub, nil }
			defer func() { newLandingArchivePublisher = orig }()
			err = laArchive(root, pr)
			if unknown {
				if !errors.Is(err, gitx.ErrPublicationUncertain) {
					t.Fatalf("unknown lost: %v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if pub.calls != 2 {
					t.Fatalf("prepare count %d want 2", pub.calls)
				}
				if !strings.Contains(testfix.Capture(t, origin, "show", "main:"+laHistory), "compatible concurrent documentation") {
					t.Fatal("remote documentation overwritten")
				}
			}
			tip := laGit(t, origin, "rev-parse", "main")
			if done, err := laProof(root, tip, pr); err != nil || !done {
				t.Fatalf("recovery proof %v %v", done, err)
			}
			newLandingArchivePublisher = orig
			if err := laArchive(root, pr); err != nil {
				t.Fatal(err)
			}
			if laGit(t, origin, "rev-parse", "main") != tip {
				t.Fatal("recovery duplicated archive")
			}
		})
	}
}

type laReadFailure struct {
	gitRunner
	fail  string
	limit bool
}

func (r laReadFailure) GitInDir(dir string, args ...string) ([]byte, error) {
	if strings.Contains(strings.Join(args, " "), r.fail) {
		if r.limit {
			return []byte(strings.Repeat(strings.Repeat("a", 40)+"\n", 10001)), nil
		}
		return nil, errors.New("injected read failure")
	}
	return r.gitRunner.GitInDir(dir, args...)
}
func TestLandingArchiveReadFailuresAndLimits(t *testing.T) {
	for _, tc := range []struct {
		fail  string
		limit bool
	}{{"rev-list --max-count", false}, {"rev-list --max-count", true}, {"trailers:key=Landing-PR", false}} {
		t.Run(tc.fail+map[bool]string{false: "failure", true: "limit"}[tc.limit], func(t *testing.T) {
			root, origin, pr := laFixture(t)
			original := landingArchiveRunner
			landingArchiveRunner = laReadFailure{original, tc.fail, tc.limit}
			defer func() { landingArchiveRunner = original }()
			before := laGit(t, origin, "rev-parse", "main")
			if _, err := laProof(root, before, pr); err == nil {
				t.Fatal("read failure/limit treated as absence")
			}
			if err := laArchive(root, pr); err == nil {
				t.Fatal("read failure/limit allowed archive")
			}
			if laGit(t, origin, "rev-parse", "main") != before {
				t.Fatal("failed read mutated remote")
			}
		})
	}
}
