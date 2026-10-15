package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// seedLinkedProject commits the ops layout in peer: the record lives at
// teams/t/projects/<slug>/project.md, linked from workshop/projects/<slug>.md.
// Returns the real file's repo-relative path.
func seedLinkedProject(t *testing.T, peer, slug, frontmatter, rows string) string {
	t.Helper()
	realRel := filepath.Join("teams", "t", "projects", slug, "project.md")
	if err := os.MkdirAll(filepath.Join(peer, filepath.Dir(realRel)), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: " + slug + "\nstatus: executing\n" + frontmatter + "---\n\n# " + slug + "\n\n## Breakdown\n\n" + rows
	if err := os.WriteFile(filepath.Join(peer, realRel), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(peer, "workshop", "projects"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join("..", "..", realRel), filepath.Join(peer, "workshop", "projects", slug+".md")); err != nil {
		t.Fatal(err)
	}
	gitIn(t, peer, "add", ".")
	gitIn(t, peer, "commit", "-q", "-m", "project: seed "+slug)
	return realRel
}

func originHead(t *testing.T, origin string) string {
	t.Helper()
	return strings.TrimSpace(testfix.Capture(t, origin, "rev-parse", "main"))
}

// #373: a hand-run tick writes every live project citing the issue, in the
// issue's own repo and in a peer (a TL-owned ops record behind a link), each on
// its origin main, and touches no checkout. A rerun publishes nothing; a dry
// run writes nothing.
func TestProjectTick_PublishesEveryCitingProjectOnOrigin(t *testing.T) {
	parent, _ := fleetCloseFixture(t)
	own := filepath.Join(parent, "ariadne")
	ownRel := seedPeerProject(t, own, "ariadne#42")
	ownOrigin := withOrigin(t, own)
	ops := initFleetRepo(t, parent, "ops")
	opsRel := seedLinkedProject(t, ops, "robust", "tl: xian-xu@ops:0\n",
		"- [ ] the thing [ariadne#42]\n- [-] dropped half [ariadne#42 M2]\n- [ ] other [ariadne#43]\n")
	opsOrigin := withOrigin(t, ops)
	opsBefore := peerState(t, ops, opsRel)
	ownHeadBefore := gitIn(t, own, "rev-parse", "HEAD")

	var out bytes.Buffer
	dry := &projectTickFlags{Issue: "ariadne#42", PR: "7", DryRun: true}
	if err := runProjectTick(context.Background(), &out, &out, dry); err != nil {
		t.Fatalf("dry run: %v\n%s", err, out.String())
	}
	if got := out.String(); !strings.Contains(got, "Would publish to ops's origin main: "+filepath.ToSlash(opsRel)) || !strings.Contains(got, "Would publish to ariadne's origin main") {
		t.Errorf("dry run output:\n%s", got)
	}
	if strings.Contains(originFile(t, opsOrigin, opsRel), "[x]") {
		t.Fatal("dry run wrote origin")
	}

	out.Reset()
	if err := runProjectTick(context.Background(), &out, &out, &projectTickFlags{Issue: "ariadne#42", PR: "7"}); err != nil {
		t.Fatalf("tick: %v\n%s", err, out.String())
	}
	opsGot := originFile(t, opsOrigin, opsRel)
	for _, want := range []string{"- [x] the thing [ariadne#42] (PR #7)", "- [-] dropped half [ariadne#42 M2]\n", "- [ ] other [ariadne#43]\n"} {
		if !strings.Contains(opsGot, want) {
			t.Errorf("ops origin lacks %q:\n%s", want, opsGot)
		}
	}
	if got := originFile(t, ownOrigin, ownRel); !strings.Contains(got, "- [x] [ariadne#42] the thing (PR #7)") {
		t.Errorf("own repo's origin not ticked:\n%s", got)
	}
	subject := strings.TrimSpace(testfix.Capture(t, opsOrigin, "log", "-1", "--format=%s", "main"))
	if subject != "project: tick ariadne#42 landed (PR #7)" {
		t.Errorf("ops origin subject %q", subject)
	}
	if after := peerState(t, ops, opsRel); after != opsBefore {
		t.Errorf("the ops checkout was touched:\nbefore %q\nafter  %q", opsBefore, after)
	}
	if gitIn(t, own, "rev-parse", "HEAD") != ownHeadBefore {
		t.Error("the own checkout gained a commit")
	}

	heads := originHead(t, opsOrigin) + originHead(t, ownOrigin)
	out.Reset()
	if err := runProjectTick(context.Background(), &out, &out, &projectTickFlags{Issue: "ariadne#42", PR: "7"}); err != nil {
		t.Fatalf("rerun: %v\n%s", err, out.String())
	}
	if originHead(t, opsOrigin)+originHead(t, ownOrigin) != heads {
		t.Errorf("rerun published again:\n%s", out.String())
	}
}

// #373: merge's hook honours each project's landing_tick policy. A TL-owned
// project is not written by default and the merge prints the TL's command; one
// that opts in, and one without tl:, are ticked.
func TestTickOnLanding_TLOwnedIsOptIn(t *testing.T) {
	parent, _ := fleetCloseFixture(t)
	own := filepath.Join(parent, "ariadne")
	ops := initFleetRepo(t, parent, "ops")
	offRel := seedLinkedProject(t, ops, "tlowned", "tl: xian-xu@ops:0\n", "- [ ] the thing [ariadne#42]\n")
	onRel := seedLinkedProject(t, ops, "optedin", "tl: xian-xu@ops:0\nlanding_tick: true\n", "- [ ] the thing [ariadne#42]\n")
	opsOrigin := withOrigin(t, ops)
	nous := initFleetRepo(t, parent, "nous")
	nousRel := seedPeerProject(t, nous, "ariadne#42")
	nousOrigin := withOrigin(t, nous)

	var out bytes.Buffer
	tickOnLanding(context.Background(), &out, &out, own, "000042-x", "7")
	if got := originFile(t, opsOrigin, offRel); strings.Contains(got, "[x]") {
		t.Errorf("TL-owned project written without opting in:\n%s", got)
	}
	if !strings.Contains(out.String(), "sdlc project tick --issue ariadne#42 --pr 7") {
		t.Errorf("merge did not print the TL's command:\n%s", out.String())
	}
	if got := originFile(t, opsOrigin, onRel); !strings.Contains(got, "- [x] the thing [ariadne#42] (PR #7)") {
		t.Errorf("opted-in project not ticked:\n%s\n%s", got, out.String())
	}
	if got := originFile(t, nousOrigin, nousRel); !strings.Contains(got, "- [x] [ariadne#42] the thing (PR #7)") {
		t.Errorf("project without tl: not ticked:\n%s", got)
	}
}

// A failed publish warns with the retry and never fails the landing; the
// hand-run reports it as an error.
func TestTickOnLanding_FailureWarnsWithRetry(t *testing.T) {
	parent, _ := fleetCloseFixture(t)
	own := filepath.Join(parent, "ariadne")
	nous := initFleetRepo(t, parent, "nous")
	seedPeerProject(t, nous, "ariadne#42") // no origin: the publish cannot happen

	var out bytes.Buffer
	tickOnLanding(context.Background(), &out, &out, own, "000042-x", "7")
	if got := out.String(); !strings.Contains(got, "NOT ticked") || !strings.Contains(got, "retry: sdlc project tick --issue ariadne#42 --pr 7") {
		t.Errorf("want a warning naming the retry:\n%s", got)
	}
	out.Reset()
	if err := runProjectTick(context.Background(), &out, &out, &projectTickFlags{Issue: "ariadne#42", PR: "7"}); err == nil {
		t.Errorf("hand-run tick should fail when a project was not ticked:\n%s", out.String())
	}
	if err := runProjectTick(context.Background(), &out, &out, &projectTickFlags{Issue: "ariadne#42", PR: "x7"}); err == nil {
		t.Error("a non-numeric --pr should refuse")
	}
	tickOnLanding(context.Background(), &out, &out, own, "no-id-branch", "7") // no id: nothing, no panic
}

// #373 Done-when: landing an issue ticks it in a project in another repo. The
// durable merge of sample#1 from slot 1 publishes the tick to ops's origin.
func TestLandingTicksPeerProject(t *testing.T) {
	roots, _, _ := landingFixture(t, 1)
	fleet := filepath.Dir(roots[0])
	ops := initFleetRepo(t, fleet, "ops")
	rel := seedLinkedProject(t, ops, "p", "", "- [ ] the procedure [sample#1]\n")
	origin := withOrigin(t, ops)

	var out bytes.Buffer
	if err := runMerge(&out, &out, landingFlags()); err != nil {
		t.Fatalf("merge: %v\n%s", err, out.String())
	}
	if got := originFile(t, origin, rel); !strings.Contains(got, "- [x] the procedure [sample#1] (PR #1)") {
		t.Errorf("landing did not tick ops's project:\n%s\n%s", got, out.String())
	}
}

// #373: the legacy (non-slot) merge path ticks too, after its archive step.
func TestRunMerge_LegacyLandingTicksPeerProject(t *testing.T) {
	dir := tempRepo(t)
	git(t, dir, "branch", "-m", "feature", "000042-x")
	git(t, dir, "push", "-q", "origin", "000042-x")
	ops := initFleetRepo(t, filepath.Dir(dir), "ops")
	rel := seedLinkedProject(t, ops, "p", "", "- [ ] the thing ["+filepath.Base(dir)+"#42]\n")
	origin := withOrigin(t, ops)

	swapMergeDeps(t, &e2eGH{openPR: "42"}, nil)
	var out bytes.Buffer
	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	if msg, died := expectDie(t, func() { runMerge(&out, &out, f) }); died {
		t.Fatalf("merge died: %s\n%s", msg, out.String())
	}
	if got := originFile(t, origin, rel); !strings.Contains(got, "- [x] the thing ["+filepath.Base(dir)+"#42] (PR #42)") {
		t.Errorf("legacy landing did not tick ops's project:\n%s\n%s", got, out.String())
	}
}
