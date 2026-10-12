package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// gitIn runs git in dir, failing the test on error.
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return testfix.Git(t, dir, args...)
}

// initFleetRepo creates parent/name as a git repo on main with one initial commit.
func initFleetRepo(t *testing.T, parent, name string) string {
	t.Helper()
	return testfix.Repo(t, testfix.At(parent, name), testfix.InitialCommit())
}

// seedPeerProject commits a project file in the peer referencing ref, returning
// its repo-relative path.
func seedPeerProject(t *testing.T, peerDir, ref string) string {
	t.Helper()
	rel := filepath.Join("workshop", "projects", "p.md")
	abs := filepath.Join(peerDir, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nstatus: executing\n---\n\n# p\n\n## tasks\n\n- [ ] [" + ref + "] the thing\n"
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, peerDir, "add", rel)
	gitIn(t, peerDir, "commit", "-q", "-m", "project: seed")
	return rel
}

// fleetCloseFixture builds a temp fleet: parent/ariadne is the closing repo
// (chdir'd into, with issue #42 committed under a "#42" subject) and returns
// (parent, issuesDir). Peers are added by the caller before running the close.
func fleetCloseFixture(t *testing.T) (string, string) {
	t.Helper()
	parent := t.TempDir()
	repo := initFleetRepo(t, parent, "ariadne")
	issuesDir := "workshop/issues"
	if err := os.MkdirAll(filepath.Join(repo, issuesDir), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nid: 000042\nstatus: working\nestimate_hours: 1\n---\n\n" +
		"# x\n\n## Spec\n\nThing.\n\n## Plan\n\n- [x] do it\n\n## Log\n"
	if err := os.WriteFile(filepath.Join(repo, issuesDir, "000042-x.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "add", ".")
	gitIn(t, repo, "commit", "-q", "-m", "#42: implement the thing")
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	return parent, issuesDir
}

func closeFlagsFor42(issuesDir string) *closeFlags {
	return &closeFlags{
		Issue:     42,
		NoActual:  true,
		Verified:  "peer-write test close",
		NoAtlas:   true,
		IssuesDir: issuesDir,
		BrainDir:  "../nonexistent-brain",
	}
}

// withOrigin gives a fleet repo a bare origin with main pushed and tracked, and
// returns the origin's path.
func withOrigin(t *testing.T, repo string) string {
	t.Helper()
	origin := filepath.Join(t.TempDir(), filepath.Base(repo)+".git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	gitIn(t, repo, "remote", "add", "origin", origin)
	gitIn(t, repo, "push", "-q", "-u", "origin", "main")
	return origin
}

// originFile reads a file from origin's main ("" when absent).
func originFile(t *testing.T, origin, rel string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", origin, "show", "main:"+filepath.ToSlash(rel)).Output()
	if err != nil {
		return ""
	}
	return string(out)
}

// peerState is what "untouched" means for a peer checkout.
func peerState(t *testing.T, peer, rel string) string {
	t.Helper()
	raw, _ := os.ReadFile(filepath.Join(peer, rel))
	return gitIn(t, peer, "rev-parse", "HEAD") + "|" + gitIn(t, peer, "status", "--porcelain", "--untracked-files=all") + "|" + string(raw)
}

// #308 Done-when 4 (C1): a peer repo's project is ticked on the peer's origin
// main by one commit; the peer's checkout (a TL's live tree) is untouched.
func TestClose_PeerProjectLandsOnOriginAndLeavesCheckoutUntouched(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	peer := initFleetRepo(t, parent, "nous")
	rel := seedPeerProject(t, peer, "ariadne#42")
	origin := withOrigin(t, peer)
	before := peerState(t, peer, rel)

	var out bytes.Buffer
	if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
		t.Fatalf("runClose: %v\n%s", err, out.String())
	}
	if got := originFile(t, origin, rel); !strings.Contains(got, "- [x] [ariadne#42]") {
		t.Errorf("origin's project row not ticked:\n%s\n%s", got, out.String())
	}
	subject := strings.TrimSpace(testfix.Capture(t, origin, "log", "-1", "--format=%s", "main"))
	if subject != "project: close-time update (ariadne#42)" {
		t.Errorf("origin commit subject %q", subject)
	}
	if after := peerState(t, peer, rel); after != before {
		t.Errorf("the peer checkout was touched:\nbefore %q\nafter  %q", before, after)
	}
}

// The ops layout: workshop/projects/x.md is a symlink into teams/…; the commit
// changes the real file, not the link (kink 38 committed the link: empty).
func TestClose_PeerProjectTickResolvesSymlink(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	peer := initFleetRepo(t, parent, "ops")
	realRel := filepath.Join("teams", "devinfra", "projects", "p", "project.md")
	if err := os.MkdirAll(filepath.Join(peer, filepath.Dir(realRel)), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nstatus: executing\n---\n\n# p\n\n## tasks\n\n- [ ] [ariadne#42] the thing\n"
	if err := os.WriteFile(filepath.Join(peer, realRel), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(peer, "workshop", "projects"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join("..", "..", realRel), filepath.Join(peer, "workshop", "projects", "p.md")); err != nil {
		t.Fatal(err)
	}
	gitIn(t, peer, "add", ".")
	gitIn(t, peer, "commit", "-q", "-m", "project: seed via link")
	origin := withOrigin(t, peer)

	var out bytes.Buffer
	if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
		t.Fatalf("runClose: %v\n%s", err, out.String())
	}
	if got := originFile(t, origin, realRel); !strings.Contains(got, "- [x] [ariadne#42]") {
		t.Fatalf("the real project file was not ticked on origin:\n%s\n%s", got, out.String())
	}
	changed := strings.TrimSpace(testfix.Capture(t, origin, "show", "--name-only", "--format=", "main"))
	if changed != filepath.ToSlash(realRel) {
		t.Errorf("the commit changed %q, want only the real file", changed)
	}
}

// A moved origin is re-read: another writer's edit to the same project lands
// first, and the tick is re-derived on top of it, keeping both.
func TestClose_PeerProjectTickRecomputesOnMovedOrigin(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	peer := initFleetRepo(t, parent, "nous")
	rel := seedPeerProject(t, peer, "ariadne#42")
	origin := withOrigin(t, peer)
	other := filepath.Join(t.TempDir(), "other")
	testfix.Git(t, "", "clone", "-q", origin, other)
	gitIn(t, other, "config", "user.name", "tl")
	gitIn(t, other, "config", "user.email", "tl@x")
	raw, _ := os.ReadFile(filepath.Join(other, rel))
	if err := os.WriteFile(filepath.Join(other, rel), []byte(strings.Replace(string(raw), "# p\n", "# p\n\nThe TL's note.\n", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, other, "commit", "-qam", "tl: note")
	gitIn(t, other, "push", "-q", "origin", "main")

	var out bytes.Buffer
	if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
		t.Fatalf("runClose: %v\n%s", err, out.String())
	}
	got := originFile(t, origin, rel)
	if !strings.Contains(got, "The TL's note.") || !strings.Contains(got, "- [x] [ariadne#42]") {
		t.Errorf("want both the moved origin's edit and the tick:\n%s", got)
	}
}

// The local copy close previewed had no tl:, but origin's does now: refused,
// nothing published (the TL owns it, #326), and the close still succeeds.
func TestClose_PeerProjectTickRefusesWhenOriginGainedTL(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	peer := initFleetRepo(t, parent, "nous")
	rel := seedPeerProject(t, peer, "ariadne#42")
	origin := withOrigin(t, peer)
	other := filepath.Join(t.TempDir(), "other")
	testfix.Git(t, "", "clone", "-q", origin, other)
	gitIn(t, other, "config", "user.name", "tl")
	gitIn(t, other, "config", "user.email", "tl@x")
	raw, _ := os.ReadFile(filepath.Join(other, rel))
	if err := os.WriteFile(filepath.Join(other, rel), []byte(strings.Replace(string(raw), "status: executing\n", "status: executing\ntl: tl@ops:0\n", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, other, "commit", "-qam", "tl: take the project")
	gitIn(t, other, "push", "-q", "origin", "main")
	tip := strings.TrimSpace(testfix.Capture(t, origin, "rev-parse", "main"))

	var out bytes.Buffer
	if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
		t.Fatalf("runClose: %v\n%s", err, out.String())
	}
	if now := strings.TrimSpace(testfix.Capture(t, origin, "rev-parse", "main")); now != tip {
		t.Error("published over a TL-owned origin copy")
	}
	if !strings.Contains(out.String(), "NOT updated on its origin") || !strings.Contains(out.String(), "TL-owned on origin") {
		t.Errorf("missing refusal warning:\n%s", out.String())
	}
}

// No upstream to publish to, or a brain: a warning, nothing written anywhere.
func TestClose_PeerProjectTickRefusesBrainAndNoOrigin(t *testing.T) {
	for _, kind := range []string{"no-origin", "brain"} {
		t.Run(kind, func(t *testing.T) {
			parent, issuesDir := fleetCloseFixture(t)
			peer := initFleetRepo(t, parent, "nous")
			rel := seedPeerProject(t, peer, "ariadne#42")
			var origin string
			if kind == "brain" {
				origin = withOrigin(t, peer)
				if err := os.MkdirAll(filepath.Join(peer, ".brain"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(peer, ".brain", "config.md"), []byte("brain\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			before := peerState(t, peer, rel)
			var out bytes.Buffer
			if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
				t.Fatalf("runClose: %v\n%s", err, out.String())
			}
			// A brain never reaches the publish (discovery skips it, #176);
			// publishPeerProject refuses one regardless.
			if kind == "no-origin" && !strings.Contains(out.String(), "NOT updated on its origin") {
				t.Errorf("missing warning:\n%s", out.String())
			}
			if after := peerState(t, peer, rel); after != before {
				t.Error("the peer checkout was touched")
			}
			if origin != "" && strings.Contains(originFile(t, origin, rel), "- [x]") {
				t.Error("a brain was written")
			}
		})
	}
}

// --no-project skips the project update entirely (#308): no tick here, none on
// the peer's origin.
func TestNoProjectSkipsTickAndPeerWrite(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	peer := initFleetRepo(t, parent, "nous")
	rel := seedPeerProject(t, peer, "ariadne#42")
	origin := withOrigin(t, peer)
	local := seedPeerProject(t, filepath.Join(parent, "ariadne"), "ariadne#42")
	tip := strings.TrimSpace(testfix.Capture(t, origin, "rev-parse", "main"))
	f := closeFlagsFor42(issuesDir)
	f.NoProject = true
	var out bytes.Buffer
	if err := runClose(&out, &out, f); err != nil {
		t.Fatalf("runClose: %v\n%s", err, out.String())
	}
	if now := strings.TrimSpace(testfix.Capture(t, origin, "rev-parse", "main")); now != tip {
		t.Error("--no-project still published to the peer's origin")
	}
	if raw, _ := os.ReadFile(filepath.Join(parent, "ariadne", local)); strings.Contains(string(raw), "- [x]") {
		t.Error("--no-project still ticked this repository's project")
	}
	if !strings.Contains(out.String(), "--no-project: project update skipped") {
		t.Errorf("missing acknowledgement:\n%s", out.String())
	}
	_ = rel
}

// A peer project is not part of the review snapshot: it is re-derived on the
// peer's origin at publish time, so a TL editing it mid-review cannot void a
// SHIP (kink 55). This repository's projects still are.
func TestCloseReviewSnapshotExcludesPeerProjects(t *testing.T) {
	_, _ = fleetCloseFixture(t)
	res := closeResult{projectEdits: []projectEdit{
		{path: "/fleet/ops/workshop/projects/p.md", repoDir: "/fleet/ops", oldText: "x", peer: true},
		{path: "/fleet/ariadne/workshop/projects/q.md", repoDir: "/fleet/ariadne", oldText: "y"},
	}}
	snap, err := captureCloseReviewSnapshot(res, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, a := range snap.prepared.artifacts {
		paths = append(paths, a.path)
	}
	joined := strings.Join(paths, ",")
	if strings.Contains(joined, "ops/workshop/projects/p.md") || !strings.Contains(joined, "ariadne/workshop/projects/q.md") {
		t.Errorf("snapshot artifacts %v: want this repo's project only", paths)
	}
}

// TestClose_CurrentRepoProjectNotAutoCommitted: a project in the CLOSING repo
// rides the normal close commit — the peer publish path must not touch it.
func TestClose_CurrentRepoProjectNotAutoCommitted(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	repo := filepath.Join(parent, "ariadne")
	rel := seedPeerProject(t, repo, "ariadne#42")

	var out bytes.Buffer
	if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
		t.Fatalf("runClose: %v\n%s", err, out.String())
	}

	if subject := gitIn(t, repo, "log", "-1", "--pretty=%s"); strings.Contains(subject, "close-time update") {
		t.Errorf("current repo must not be auto-committed by the peer publish path, HEAD is %q", subject)
	}
	status := gitIn(t, repo, "status", "--porcelain")
	if !strings.Contains(status, " M "+rel) {
		t.Errorf("current repo project edit should sit uncommitted for the close commit:\n%s", status)
	}
	if strings.Contains(out.String(), "NOT committed in ariadne") {
		t.Errorf("peer publish path reported the current repo: %q", out.String())
	}
}

// publishPeerProject's own refusal of a brain (defence in depth behind
// discovery's skip).
func TestPublishPeerProjectRefusesABrain(t *testing.T) {
	parent := t.TempDir()
	peer := initFleetRepo(t, parent, "brain")
	rel := seedPeerProject(t, peer, "ariadne#42")
	if err := os.MkdirAll(filepath.Join(peer, ".brain"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(peer, ".brain", "config.md"), []byte("brain\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	e := projectEdit{path: filepath.Join(peer, rel), repoDir: peer, peer: true, in: projectEditInput{RepoName: "ariadne", IssueStr: "42"}}
	if err := publishPeerProject(context.Background(), e, "ariadne#42"); err == nil || !strings.Contains(err.Error(), "brain") {
		t.Fatalf("brain: %v", err)
	}
}

// BR-7: a brain's legacy data/project home keeps its pre-#308 path — the tick
// is written in place for nous to commit — and is never published to origin.
func TestClose_BrainLegacyProjectIsWrittenInPlace(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	brain := initFleetRepo(t, parent, "brain")
	if err := os.MkdirAll(filepath.Join(brain, ".brain"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(brain, ".brain", "config.md"), []byte("brain\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rel := filepath.Join("data", "project", "metis.md")
	if err := os.MkdirAll(filepath.Join(brain, filepath.Dir(rel)), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nstatus: executing\n---\n\n# metis\n\n## tasks\n\n- [ ] [ariadne#42] the thing\n"
	if err := os.WriteFile(filepath.Join(brain, rel), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, brain, "add", ".")
	gitIn(t, brain, "commit", "-q", "-m", "brain: seed")
	origin := withOrigin(t, brain)
	head := gitIn(t, brain, "rev-parse", "HEAD")

	var out bytes.Buffer
	if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
		t.Fatalf("runClose: %v\n%s", err, out.String())
	}
	if raw, _ := os.ReadFile(filepath.Join(brain, rel)); !strings.Contains(string(raw), "- [x] [ariadne#42]") {
		t.Fatalf("brain legacy project not ticked in place:\n%s\n%s", raw, out.String())
	}
	if gitIn(t, brain, "rev-parse", "HEAD") != head {
		t.Error("close committed into the brain")
	}
	if strings.Contains(originFile(t, origin, rel), "- [x]") {
		t.Error("close published into the brain's origin")
	}
	if strings.Contains(out.String(), "NOT updated on its origin") {
		t.Errorf("a brain legacy project was treated as a peer:\n%s", out.String())
	}
}
