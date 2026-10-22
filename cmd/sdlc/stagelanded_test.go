package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

const landedStageCard = "workshop/issue-cards/000001-procedure.md"

// This fixture retains a real stage lease across production durable landing:
// completion, archive, branch deletion and the original caller's return to rest.
// All Git transport goes to a private bare server.
type landedStageFixture struct {
	roots   []string
	remote  string
	gh      *landingFakeGH
	tracker *gitx.TrunkFile
	head    string
}

func newLandedStageFixture(t *testing.T, strategy string) *landedStageFixture {
	t.Helper()
	return newLandedStageFixtureWithMirror(t, strategy, true)
}

func newLandedStageFixtureWithMirror(t *testing.T, strategy string, mirrored bool) *landedStageFixture {
	t.Helper()
	roots, remote, gh := landingFixture(t, 1)
	gh.strategy = strategy
	root := roots[1]
	env, err := openTrackerAt(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	tf, err := gitx.NewTrunkFileContext(context.Background(), root, "upstream", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	working := []byte("---\nid: 000001\nstatus: working\nupdated: 2026-09-23\n---\n\n# Procedure\n\n## Problem\nx\n")
	if _, err := tf.Bootstrap(map[string][]byte{tracker.ManifestPath: tracker.ManifestBytes(), landedStageCard: working}, "bootstrap staged landing", func(gitx.BootstrapResult) error { return nil }); err != nil {
		t.Fatal(err)
	}
	markCutoverOn(t, roots[2], "upstream", "main")
	git(t, root, "fetch", "-q", "upstream")
	git(t, roots[0], "merge", "-q", "--ff-only", "upstream/main")
	git(t, root, "rebase", "-q", "upstream/main")
	git(t, root, "push", "-q", "--force", "upstream", landingTestBranch)
	gh.pr.HeadOID, gh.pr.BaseOID = procedureHead(t, root), git(t, root, "rev-parse", "upstream/main")
	card := []byte("---\nid: 000001\nstatus: codecomplete\nactual_hours: 1\nupdated: 2026-09-23\n---\n\n# Procedure\n\n## Problem\nx\n")
	card, err = issue.SetCardCompletion(card, issue.Completion{Token: "close-staged-landing", Repository: env.target.Repository, ReviewedHEAD: gh.pr.HeadOID, EvidenceCommit: gh.pr.HeadOID})
	if err != nil {
		t.Fatal(err)
	}
	if err := tf.UpdateMany("close staged issue", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: map[string][]byte{landedStageCard: card}, ExactBytes: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if mirrored {
		// A real close publishes its binding, then commits that card's mirror.
		// The evidence commit stays an ancestor of this later PR head.
		details, err := issue.DetailsFromCard(card, "sha1", "2026-09-23")
		if err != nil {
			t.Fatal(err)
		}
		procedureWrite(t, root, procedureIssue, string(details))
		git(t, root, "add", "--", procedureIssue)
		git(t, root, "commit", "-qm", "#1: mirror closed card")
		git(t, root, "push", "-q", "upstream", landingTestBranch)
		gh.pr.HeadOID = procedureHead(t, root)
	}
	git(t, root, "remote", "add", "origin", "https://github.com/test/repo.git")
	if out, err := runStageTest(t, root, ":0", false); err != nil {
		t.Fatalf("stage before landing: %v\n%s", err, out)
	}
	head := gh.pr.HeadOID
	if err := runMerge(io.Discard, io.Discard, landingFlags()); err != nil {
		t.Fatal("durable landing", err)
	}
	if got := git(t, root, "branch", "--show-current"); got != "main-slot1" {
		t.Fatalf("caller not resting: %s", got)
	}
	if got := git(t, root, "branch", "--list", landingTestBranch); got != "" {
		t.Fatalf("landed branch still exists: %s", got)
	}
	if got := git(t, remote, "ls-tree", "-r", "--name-only", "main"); !strings.Contains(got, "workshop/history/issues/000001-procedure.md") {
		t.Fatalf("landing did not archive details: %s", got)
	}
	raw, err := tf.Read(landedStageCard)
	if err != nil {
		t.Fatal(err)
	}
	binding, present, err := issue.CardCompletion(raw)
	if err != nil || !present || binding.LandedCommit == "" || !strings.Contains(string(raw), "status: done") {
		t.Fatalf("landing did not complete card: %v\n%s", err, raw)
	}
	lease, err := fleet.ReadStageLease(execGitRunner{}, roots[0])
	if err != nil || lease.Slot != "sample:1" || lease.Branch != landingTestBranch || lease.Commit != head || lease.Holder != "#1" {
		t.Fatalf("original stage binding: %+v, %v", lease, err)
	}
	return &landedStageFixture{roots: roots, remote: remote, gh: gh, tracker: tf, head: head}
}

func TestStageClearAfterDurableLanding(t *testing.T) {
	f := newLandedStageFixture(t, "merge")
	stagedLandingBuild(t, f.roots[0])
	out, err := runUnstageTest(t, f.roots[1], ":0")
	if err != nil {
		t.Fatalf("original landed holder must clear from rest without --abandoned: %v\n%s", err, out)
	}
	if got := git(t, f.roots[0], "branch", "--show-current"); got != "main" {
		t.Fatalf("target not restored: %s", got)
	}
	if got := procedureHead(t, f.roots[0]); got != git(t, f.remote, "rev-parse", "main") {
		t.Fatalf("target not at freshly fetched main: %s", got)
	}
	if lease, err := fleet.ReadStageLease(execGitRunner{}, f.roots[0]); err != nil || lease != (fleet.StageLease{}) {
		t.Fatalf("lease retained: %+v, %v", lease, err)
	}
	if stamp, err := os.ReadFile(filepath.Join(f.roots[0], "bin", "tool")); err != nil || strings.TrimSpace(string(stamp)) != procedureHead(t, f.roots[0]) {
		t.Fatalf("clear did not rebuild at its exact main pin: %q, %v", stamp, err)
	}
}

// Like stageFreshFixture, model binary build info while exercising real make.
func stagedLandingBuild(t *testing.T, root string) {
	t.Helper()
	procedureWrite(t, root, "Makefile", "build:\n\tmkdir -p bin && git rev-parse HEAD > bin/tool\n")
	previous := readBuildInfo
	t.Cleanup(func() { readBuildInfo = previous })
	readBuildInfo = func(path string) (*debug.BuildInfo, error) {
		stamp, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return &debug.BuildInfo{Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: strings.TrimSpace(string(stamp))},
			{Key: "vcs.modified", Value: "false"},
		}}, nil
	}
}

func (f *landedStageFixture) changeCard(t *testing.T, change func([]byte) ([]byte, error)) {
	t.Helper()
	raw, err := f.tracker.Read(landedStageCard)
	if err != nil {
		t.Fatal(err)
	}
	next, err := change(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.tracker.UpdateMany("change stage holder evidence", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: map[string][]byte{landedStageCard: next}, ExactBytes: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestStageClearLandedHolderDryRunAndRestingRetry(t *testing.T) {
	for _, resting := range []bool{false, true} {
		t.Run(map[bool]string{false: "detached", true: "already resting"}[resting], func(t *testing.T) {
			f := newLandedStageFixture(t, "merge")
			if resting {
				git(t, f.roots[0], "switch", "-q", "main")
			}
			stagedLandingBuild(t, f.roots[0])
			before := slotState(t, f.roots[0])
			lease, err := fleet.ReadStageLease(execGitRunner{}, f.roots[0])
			if err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			if err := runUnstage(context.Background(), f.roots[1], ":0", "", true, &out, &out); err != nil {
				t.Fatalf("dry run: %v\n%s", err, &out)
			}
			after, err := fleet.ReadStageLease(execGitRunner{}, f.roots[0])
			if !strings.Contains(out.String(), "Would") || slotState(t, f.roots[0]) != before || err != nil || after != lease {
				t.Fatalf("dry run mutated stage: %v\n%s", err, &out)
			}
			if _, err := os.Stat(filepath.Join(f.roots[0], "bin")); !os.IsNotExist(err) {
				t.Fatalf("dry run rebuilt binaries: %v", err)
			}
			if out, err := runUnstageTest(t, f.roots[1], ":0"); err != nil {
				t.Fatalf("clear: %v\n%s", err, out)
			}
		})
	}
}

func TestStageClearLandedHolderMergeStrategies(t *testing.T) {
	for _, strategy := range []string{"squash", "rebase"} {
		t.Run(strategy, func(t *testing.T) {
			f := newLandedStageFixture(t, strategy)
			if out, err := runUnstageTest(t, f.roots[1], ":0"); err != nil {
				t.Fatalf("clear after %s: %v\n%s", strategy, err, out)
			}
		})
	}
}

// A train car may have no useful raw mergeCommit; its exact head still
// reaches main through the train. Exercise the existing landedCommit fallback.
func TestStageClearLandedCarRawMergeVariants(t *testing.T) {
	for _, raw := range []string{"missing", "unrelated", "unknown object"} {
		t.Run(raw, func(t *testing.T) {
			f := newLandedStageFixture(t, "merge")
			f.gh.pr.MergeOID = ""
			if raw == "unknown object" {
				f.gh.pr.MergeOID = strings.Repeat("a", 40)
			}
			if raw == "unrelated" {
				f.gh.pr.MergeOID = git(t, f.roots[2], "commit-tree", git(t, f.roots[2], "rev-parse", "HEAD^{tree}"), "-m", "unrelated raw merge")
			}
			if out, err := runUnstageTest(t, f.roots[1], ":0"); err != nil {
				t.Fatalf("integrated car head must suffice: %v\n%s", err, out)
			}
		})
	}
}

func TestStageClearLandedHolderGenericArchive(t *testing.T) {
	f := newLandedStageFixture(t, "merge")
	// A done-card sweep has no per-PR trailers. Preserve the exact durable
	// archive transition and current card generation while removing provenance.
	git(t, f.roots[2], "fetch", "-q", "upstream")
	git(t, f.roots[2], "merge", "-q", "--ff-only", "upstream/main")
	git(t, f.roots[2], "commit", "-q", "--amend", "-m", "archive completed issue family")
	git(t, f.roots[2], "push", "-q", "--force", "upstream", "HEAD:main")
	if trailers := git(t, f.roots[2], "show", "-s", "--format=%(trailers:key=Landing-PR,valueonly)", "HEAD"); trailers != "" {
		t.Fatalf("generic archive fixture retained PR provenance: %s", trailers)
	}
	if out, err := runUnstageTest(t, f.roots[1], ":0"); err != nil {
		t.Fatalf("generic done-card archive must authorize the same landed holder: %v\n%s", err, out)
	}
}

func TestStageClearLandedHolderRefusesUnverifiedEvidence(t *testing.T) {
	for _, scenario := range []string{
		"other slot", "other issue", "missing slot", "replaced lease",
		"working", "codecomplete", "missing completion", "replaced completion",
		"wrong repository", "unmerged PR", "missing PR", "ambiguous PR",
		"unrelated staged commit", "unreachable landing", "missing archive",
		"changed archive", "unmirrored archive", "GitHub unavailable",
	} {
		t.Run(scenario, func(t *testing.T) {
			f := newLandedStageFixtureWithMirror(t, "merge", scenario != "unmirrored archive")
			caller := f.roots[1]
			switch scenario {
			case "other slot":
				caller = f.roots[2]
			case "other issue":
				git(t, caller, "switch", "-q", "-c", "000002-other")
			case "missing slot", "replaced lease", "unrelated staged commit":
				lease, err := fleet.ReadStageLease(execGitRunner{}, f.roots[0])
				if err != nil {
					t.Fatal(err)
				}
				switch scenario {
				case "missing slot":
					lease.Slot = ""
				case "replaced lease":
					lease.Holder, lease.Branch, lease.Slot = "#2", "000002-other", "sample:2"
				case "unrelated staged commit":
					lease.Commit = git(t, caller, "commit-tree", git(t, caller, "rev-parse", "HEAD^{tree}"), "-m", "unrelated staged object")
					git(t, f.roots[0], "checkout", "-q", "--detach", lease.Commit)
				}
				if err := fleet.WriteStageLease(execGitRunner{}, f.roots[0], lease); err != nil {
					t.Fatal(err)
				}
			case "working", "codecomplete":
				f.changeCard(t, func(raw []byte) ([]byte, error) { return issue.SetCardField(raw, "status", scenario) })
			case "missing completion", "replaced completion", "wrong repository", "unreachable landing":
				f.changeCard(t, func(raw []byte) ([]byte, error) {
					b, _, err := issue.CardCompletion(raw)
					if err != nil {
						return nil, err
					}
					switch scenario {
					case "missing completion":
						return []byte("---\nid: 000001\nstatus: done\nupdated: 2026-09-23\n---\n# Procedure\n"), nil
					case "replaced completion":
						b.Token += "-new"
					case "wrong repository":
						old := "repository: " + b.Repository
						if !bytes.Contains(raw, []byte(old)) {
							return nil, fmt.Errorf("fixture missing exact repository field %q", old)
						}
						// Remote tracker bytes are untrusted; bypass the authoring
						// setter's immutability protection to test reader refusal.
						return bytes.Replace(raw, []byte(old), []byte("repository: github.com/other/repo"), 1), nil
					case "unreachable landing":
						old := "landed_commit: " + b.LandedCommit
						if !bytes.Contains(raw, []byte(old)) {
							return nil, fmt.Errorf("fixture missing exact landed commit field %q", old)
						}
						unreachable := git(t, caller, "commit-tree", git(t, caller, "rev-parse", "HEAD^{tree}"), "-m", "unreachable landing")
						return bytes.Replace(raw, []byte(old), []byte("landed_commit: "+unreachable), 1), nil
					}
					return issue.SetCardCompletion(raw, b)
				})
			case "unmerged PR":
				f.gh.pr.State = "OPEN"
			case "missing PR":
				swapLandedGH(t, landedFakeGH{prs: map[string][]landingPR{}})
			case "ambiguous PR":
				duplicate := f.gh.pr
				duplicate.Number++
				f.gh.older = []landingPR{duplicate}
			case "GitHub unavailable":
				f.gh.queryErr = errors.New("private fake unavailable")
			case "missing archive", "changed archive":
				git(t, f.roots[2], "fetch", "-q", "upstream")
				git(t, f.roots[2], "merge", "-q", "--ff-only", "upstream/main")
				p := "workshop/history/issues/000001-procedure.md"
				if scenario == "missing archive" {
					git(t, f.roots[2], "rm", "-q", p)
				} else {
					procedureWrite(t, f.roots[2], p, "corrupted archive\n")
					git(t, f.roots[2], "add", p)
				}
				git(t, f.roots[2], "commit", "-qm", "damage archive authority")
				git(t, f.roots[2], "push", "-q", "upstream", "HEAD:main")
			}
			stagedLandingBuild(t, f.roots[0])
			before := slotState(t, f.roots[0])
			lease, err := fleet.ReadStageLease(execGitRunner{}, f.roots[0])
			if err != nil {
				t.Fatal(err)
			}
			out, err := runUnstageTest(t, caller, ":0")
			if err == nil {
				t.Fatalf("unverified authority cleared stage:\n%s", out)
			}
			after, readErr := fleet.ReadStageLease(execGitRunner{}, f.roots[0])
			if slotState(t, f.roots[0]) != before || readErr != nil || after != lease {
				t.Fatalf("refusal mutated stage: %v\n%s", err, out)
			}
			if _, err := os.Stat(filepath.Join(f.roots[0], "bin")); !os.IsNotExist(err) {
				t.Fatalf("refusal rebuilt binaries: %v", err)
			}
		})
	}
}
