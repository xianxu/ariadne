package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// Observe the real pushed refs and consume exact-head green CI before a
// publication advances main. The embedded fake still owns merges and statuses;
// this wrapper never copies checks or suite results to a different head.
type trainIssueGH struct {
	*trainFakeGH
	observeMu                  sync.Mutex
	ciHeads                    map[string]bool
	firstCIHead, publishedMain string
	readinessCalls             int
	once                       sync.Once
	publish                    func(context.Context) (string, error)
	publishErr                 error
	readinessMode              string
	afterReadiness             func(context.Context) (string, error)
	readinessOnce              sync.Once
	lateMain                   string
	lateErr                    error
}

func (g *trainIssueGH) currentPR(ctx context.Context, repo string, n int) (landingPR, bool, error) {
	g.trainFakeGH.mu.Lock()
	var branch string
	for name, pr := range g.prs {
		if pr.Number == n {
			branch = name
			break
		}
	}
	g.trainFakeGH.mu.Unlock()
	if branch == "" {
		return landingPR{}, false, fmt.Errorf("unknown PR %d", n)
	}
	prs, err := g.trainFakeGH.LandingPRs(ctx, repo, branch)
	if err != nil || len(prs) != 1 {
		return landingPR{}, false, fmt.Errorf("read PR %d: %w", n, err)
	}
	return prs[0], strings.HasPrefix(branch, trainBranchPrefix), nil
}

func (g *trainIssueGH) PRHeadOID(ctx context.Context, repo string, n int) (string, error) {
	pr, _, err := g.currentPR(ctx, repo, n)
	return pr.HeadOID, err
}

func (g *trainIssueGH) PRChecks(ctx context.Context, repo string, n int) ([]prCheck, error) {
	pr, train, err := g.currentPR(ctx, repo, n)
	if err != nil {
		return nil, err
	}
	checks, err := g.ciFakeGH.PRChecks(ctx, repo, n)
	if err != nil || !train {
		return checks, err
	}
	g.observeMu.Lock()
	g.ciHeads[pr.HeadOID] = true
	g.observeMu.Unlock()
	g.once.Do(func() {
		// This response is already green on the old train head. A real CAS
		// publication moves main before the conductor consumes that response.
		moved, publishErr := g.publish(ctx)
		g.observeMu.Lock()
		g.firstCIHead, g.publishedMain, g.publishErr = pr.HeadOID, moved, publishErr
		g.observeMu.Unlock()
	})
	g.observeMu.Lock()
	publishErr := g.publishErr
	g.observeMu.Unlock()
	if publishErr != nil {
		return nil, publishErr
	}
	return checks, nil
}

func (g *trainIssueGH) LandingReadiness(ctx context.Context, repo string, n int) (landingReadiness, error) {
	pr, _, err := g.currentPR(ctx, repo, n)
	if err != nil {
		return landingReadiness{}, err
	}
	head, err := trainGitOut(g.remote, "rev-parse", "refs/heads/"+pr.HeadRef)
	if err != nil {
		return landingReadiness{}, err
	}
	base, err := trainGitOut(g.remote, "rev-parse", "refs/heads/main")
	if err != nil {
		return landingReadiness{}, err
	}
	state := "CLEAN"
	if _, err := trainGitOut(g.remote, "merge-tree", "--write-tree", base, head); err != nil {
		state = "DIRTY"
	}
	result := landingReadiness{Repo: pr.Repo, Number: pr.Number, State: pr.State, HeadRef: pr.HeadRef, HeadOID: head, BaseRef: "main", BaseTargetOID: base, MergeStateStatus: state}
	g.observeMu.Lock()
	g.readinessCalls++
	mode := g.readinessMode
	g.observeMu.Unlock()
	switch mode {
	case "unknown":
		result.MergeStateStatus = "UNKNOWN"
	case "stale":
		result.BaseTargetOID = pr.BaseOID
	case "unavailable":
		return landingReadiness{}, errors.New("readiness unavailable in fixture")
	case "identity":
		result.HeadOID = strings.Repeat("9", 40)
	}
	if g.afterReadiness != nil && strings.HasPrefix(pr.HeadRef, trainBranchPrefix) {
		g.readinessOnce.Do(func() {
			late, lateErr := g.afterReadiness(ctx)
			g.observeMu.Lock()
			g.lateMain, g.lateErr = late, lateErr
			g.observeMu.Unlock()
		})
		g.observeMu.Lock()
		lateErr := g.lateErr
		g.observeMu.Unlock()
		if lateErr != nil {
			return landingReadiness{}, lateErr
		}
	}
	return result, nil
}

// Count actual successful train-head pushes, including a redundant same-head
// publication. PR creation counts alone would miss reintegration pushes.
type trainIssuePushes struct {
	gitRunner
	mu    sync.Mutex
	heads []string
}

func (r *trainIssuePushes) GitInDir(dir string, args ...string) ([]byte, error) {
	out, err := r.gitRunner.GitInDir(dir, args...)
	if err == nil {
		for _, arg := range args {
			source, _, found := strings.Cut(arg, ":refs/heads/"+trainBranchPrefix)
			if found && landingOIDValid(source) {
				r.mu.Lock()
				r.heads = append(r.heads, source)
				r.mu.Unlock()
			}
		}
	}
	return out, err
}

func trainIssuePublication(ctx context.Context, root string, files map[string][]byte, deletes ...string) (string, error) {
	tf, err := gitx.NewTrunkFileContext(ctx, root, "origin", "main")
	if err != nil {
		return "", err
	}
	if err := tf.UpdateMany("issue: publish records while train CI completes", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: files, Delete: deletes, ExactBytes: true}, nil
	}); err != nil {
		return "", err
	}
	view, err := tf.Snapshot()
	if err != nil {
		return "", err
	}
	return view.Ref(), nil
}

func TestLandingTrainIssueOnlyMainKeepsVerifiedHead(t *testing.T) {
	for _, kind := range []string{"details", "cards", "details and cards"} {
		t.Run(kind, func(t *testing.T) {
			runTrainIssueMovement(t, kind, "", false)
		})
	}
}

func TestLandingTrainIssueOnlyMainControls(t *testing.T) {
	for _, kind := range []string{"code", "configuration", "unknown", "stale", "unavailable", "identity", "code added and reverted", "late-main", "late-code", "watched issue"} {
		t.Run(kind, func(t *testing.T) {
			mode := ""
			if slices.Contains([]string{"unknown", "stale", "unavailable", "identity"}, kind) {
				mode = kind
			}
			runTrainIssueMovement(t, kind, mode, false)
		})
	}
}

func TestLandingTrainIssueOnlyLateDuplicateRefuses(t *testing.T) {
	runTrainIssueMovement(t, "duplicate", "", true)
}

func runTrainIssueMovement(t *testing.T, kind, mode string, validate bool) {
	t.Helper()
	s := newTrainSetup(t, "train")
	if validate {
		gateInstancesByReading(t)
		s.validate = true
	}
	trace := ""
	if kind == "watched issue" {
		trace = filepath.Join(t.TempDir(), "landing-check-heads")
		configureTrainIssueWatch(t, s, trace)
	}
	secondRoot := s.pushCar(t, "000002-second", "second", 2, 2)
	if validate {
		// Synthetic trainCar normally supplies only Log. This control needs both
		// real instance/section gates to pass before publishing the late duplicate.
		completeTrainIssueSections(t, s, s.first, landingTestBranch, procedureIssue, 1)
		completeTrainIssueSections(t, s, secondRoot, "000002-second", "workshop/issues/000002-second.md", 2)
	}
	s.gh.trainChecks = []prCheck{{Name: "train-build", Bucket: "pass"}}
	files := map[string][]byte{}
	detail := "workshop/issues/000700-filed.md"
	card := "workshop/issue-cards/000701-filed.md"
	switch kind {
	case "cards":
		files[card] = []byte("---\nid: 000701\nstatus: open\n---\n# Filed card\n")
	case "details and cards":
		files[detail] = []byte("---\nid: 000700\nstatus: open\n---\n# Filed details\n")
		files[card] = []byte("---\nid: 000701\nstatus: open\n---\n# Filed card\n")
	case "code":
		files["late-main.go"] = []byte("package main\n")
	case "configuration":
		files[landingConfigPath] = []byte("{\"train_interval_minutes\": 30}\n")
	case "duplicate":
		files["workshop/issues/000002-late-duplicate.md"] = []byte("---\nid: 000002\nstatus: open\nupdated: 2026-10-21\n---\n# Duplicate filed after CI\n\n## Spec\nDuplicate control.\n\n## Done when\n- ID is unique.\n\n## Plan\n- [ ] Verify.\n\n## Log\n")
	default:
		files[detail] = []byte("---\nid: 000700\nstatus: open\n---\n# Filed details\n")
	}
	wrapper := &trainIssueGH{trainFakeGH: s.gh, ciHeads: map[string]bool{}, readinessMode: mode, publish: func(ctx context.Context) (string, error) {
		if kind == "code added and reverted" {
			if _, err := trainIssuePublication(ctx, s.server, map[string][]byte{"reverted-main.go": []byte("package main\n")}); err != nil {
				return "", err
			}
			return trainIssuePublication(ctx, s.server, files, "reverted-main.go")
		}
		return trainIssuePublication(ctx, s.server, files)
	}}
	lateFiles := map[string][]byte{}
	if kind == "late-main" || kind == "late-code" {
		path := "workshop/issues/000702-after-snapshot.md"
		body := "---\nid: 000702\nstatus: open\n---\n# Main advanced after readiness\n"
		if kind == "late-code" {
			path, body = "after-snapshot.go", "package main\n"
		}
		lateFiles[path] = []byte(body)
		wrapper.afterReadiness = func(ctx context.Context) (string, error) { return trainIssuePublication(ctx, s.server, lateFiles) }
	}
	ghClient = wrapper
	pushes := &trainIssuePushes{gitRunner: mergeRunner}
	originalRunner := mergeRunner
	mergeRunner = pushes
	t.Cleanup(func() { mergeRunner = originalRunner })
	one := s.merge(s.first)
	waitQueued(t, s.line, 1, []chan error{one})
	two := s.waitingCar(t, "000002-second", 2, 2)
	waitVouched(t, s.line, 2, []chan error{one, two})
	s.blocker.Release()
	err := finished(t, one)
	// Always consume the waiter before inspecting mutable fake state. A refused
	// train settles the other car too; this is not a second merge invocation.
	waiterErr := awaitLandingCompletion(t, two, "second car did not settle")
	wrapper.observeMu.Lock()
	original, moved, publicationErr, readinessCalls := wrapper.firstCIHead, wrapper.publishedMain, wrapper.publishErr, wrapper.readinessCalls
	lateMain, lateErr := wrapper.lateMain, wrapper.lateErr
	ciHeads := make([]string, 0, len(wrapper.ciHeads))
	for head := range wrapper.ciHeads {
		ciHeads = append(ciHeads, head)
	}
	wrapper.observeMu.Unlock()
	pushes.mu.Lock()
	publications := append([]string(nil), pushes.heads...)
	pushes.mu.Unlock()
	if original == "" || moved == "" || publicationErr != nil {
		t.Fatalf("CI publication boundary was not exercised: original=%s moved=%s err=%v merge=%v", original, moved, publicationErr, err)
	}
	if kind == "identity" || kind == "duplicate" {
		want := "identity"
		if kind == "duplicate" {
			want = "duplicate"
		}
		if err == nil || !strings.Contains(err.Error(), want) || s.gh.merges != 0 {
			t.Fatalf("%s did not refuse before server merge: err=%v merges=%d waiter=%v", kind, err, s.gh.merges, waiterErr)
		}
		if len(publications) != 1 || publications[0] != original {
			t.Fatalf("refusal changed verified train head: %v original=%s", publications, original)
		}
		return
	}
	if err != nil {
		t.Fatalf("train merge: %v; waiter=%v", err, waiterErr)
	}
	var settled *landingline.Settled
	if !errors.As(waiterErr, &settled) || settled.Outcome.Kind != landingline.OutcomeLanded {
		t.Fatalf("second car not landed: %v", waiterErr)
	}
	if s.gh.merges != 1 {
		t.Fatalf("server merges=%d want1", s.gh.merges)
	}
	positive := slices.Contains([]string{"details", "cards", "details and cards"}, kind)
	if positive {
		if len(publications) != 1 || publications[0] != original || len(ciHeads) != 1 || ciHeads[0] != original {
			t.Fatalf("issue-only movement restarted publication/CI: pushes=%v CIheads=%v original=%s", publications, ciHeads, original)
		}
		if readinessCalls == 0 {
			t.Fatal("issue-only fast landing did not obtain fresh readiness")
		}
	} else {
		if len(publications) != 2 || publications[0] != original || publications[1] == original || len(ciHeads) != 2 {
			t.Fatalf("%s reused old-head CI: pushes=%v CIheads=%v original=%s", kind, publications, ciHeads, original)
		}
	}
	cars := []string{s.heads[landingTestBranch], s.heads["000002-second"]}
	suites := slices.DeleteFunc(s.gh.suiteHeads(), func(head string) bool { return slices.Contains(cars, head) })
	wantSuites := 1
	if kind == "code" || kind == "configuration" || kind == "late-code" {
		wantSuites = 2
	}
	if len(suites) != wantSuites {
		t.Fatalf("%s suites=%v want%d", kind, suites, wantSuites)
	}
	main := git(t, s.gh.remote, "rev-parse", "refs/heads/main")
	if kind == "late-main" || kind == "late-code" {
		if lateMain == "" || lateErr != nil {
			t.Fatalf("late-main snapshot race was not exercised: %s %v", lateMain, lateErr)
		}
		moved = lateMain
		for path, body := range lateFiles {
			files[path] = body
		}
	}
	if kind == "watched issue" {
		raw, err := os.ReadFile(trace)
		heads := strings.Fields(string(raw))
		if err != nil || len(heads) != 2 || heads[0] != original || heads[1] != publications[1] {
			t.Fatalf("watched issue did not rerun selected check on the integrated head: %q err=%v publications=%v", raw, err, publications)
		}
	}
	if kind == "code added and reverted" {
		if !gitSucceeds(s.gh.remote, "cat-file", "-e", moved+"^:reverted-main.go") || gitSucceeds(s.gh.remote, "cat-file", "-e", moved+":reverted-main.go") {
			t.Fatal("fixture did not add and revert code before the records-only final delta")
		}
	}
	for _, ancestor := range append(cars, moved) {
		if !gitSucceeds(s.gh.remote, "merge-base", "--is-ancestor", ancestor, main) {
			t.Fatalf("final main omitted %s", ancestor)
		}
	}
	for path, body := range files {
		if got := git(t, s.gh.remote, "show", main+":"+path); got != strings.TrimSpace(string(body)) {
			t.Fatalf("final main lost published %s: %q", path, got)
		}
	}
}

var _ landingReadinessGH = (*trainIssueGH)(nil)

// Complete only the fixture's missing body sections; the real duplicate-ID
// gate remains active, and the refreshed branch head gets its own branch suite.
func completeTrainIssueSections(t *testing.T, s *trainSetup, root, branch, detail string, number int) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(detail))
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	for _, section := range []string{"Spec", "Plan", "Done when"} {
		if !strings.Contains(body, "## "+section+"\n") {
			content := "Fixture landing contract."
			if section == "Plan" {
				content = "- [x] Verify this car participates in the train."
			} else if section == "Done when" {
				content = "- The train retains both cars and unique issue IDs."
			}
			body += "\n## " + section + "\n\n" + content + "\n"
		}
	}
	if body != string(raw) {
		procedureWrite(t, root, detail, body)
		git(t, root, "add", detail)
		git(t, root, "commit", "-qm", "complete fixture issue sections")
		git(t, root, "push", "-q", "upstream", branch)
	}
	s.addCar(t, branch, number)
}

func configureTrainIssueWatch(t *testing.T, s *trainSetup, trace string) {
	t.Helper()
	decl := localCheckDecl{Version: 2, Checks: []localCheck{
		{Name: "suite", Required: true, Command: []string{"sh", "-c", "echo suite passed"}},
		{Name: "vm", RunScope: "landing", Resource: "runner", Required: true, PathsFile: "issue-paths.json", Command: []string{"sh", "-c", `git rev-parse HEAD >> "$1"`, "fixture", trace}},
	}}
	raw, err := json.Marshal(decl)
	if err != nil {
		t.Fatal(err)
	}
	procedureWrite(t, s.server, localChecksPath, string(raw))
	procedureWrite(t, s.server, "issue-paths.json", `{"version":1,"paths":["workshop/issues/**"]}`)
	git(t, s.server, "add", ".")
	git(t, s.server, "commit", "-qm", "watch issue documentation with a real landing check")
	git(t, s.server, "push", "-q", "origin", "main")
}

// This exact path is a tracker cutover configuration, not a Markdown record.
// Exercise the train's history-proof seam without installing an invalid tracker
// marker in a legacy fixture and changing unrelated post-merge cleanup behavior.
func TestLandingTrainIssueOnlyTrackerConfigRequiresIntegration(t *testing.T) {
	g := &movementCountingGit{out: "workshop/issue-tracker.json\x00"}
	only, err := mainMovementIssueOnly(g, "fixture", strings.Repeat("a", 40), strings.Repeat("b", 40))
	if err != nil || only || g.calls != 1 {
		t.Fatalf("tracker configuration admitted by record history proof: only=%v err=%v calls=%d", only, err, g.calls)
	}
}
