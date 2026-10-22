package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
	"github.com/xianxu/ariadne/pkg/workspace"
)

const weeklyInputLimit = 64 << 20

type weeklyCardEvidence struct {
	Status     string
	CreatedAt  time.Time
	Completion *issue.Completion
}

type weeklyMergeGraph struct {
	MergeSHA, MainParent, IssueParent    string
	PR                                   int
	Branch                               string
	BranchSHAs                           map[string]bool
	EvidenceReachable, ReachableFromMain bool
}

var weeklyPullMergeSubject = regexp.MustCompile(`^Merge pull request #([1-9][0-9]*) from [^/]+/(.+)$`)

type weeklyRepoEvidence struct {
	Repo, Root, MainRef, TrackerRef string
	CollectedAt                     time.Time
	Attempts                        []SuiteAttempt
	Events                          []landingline.Event
	Cards                           map[int]weeklyCardEvidence
	Graphs                          map[string]weeklyMergeGraph // keyed by completion.LandedCommit
	RecentMainMerges                map[string]time.Time
	Warnings, CoverageUnknown       []string
}

// discoverWeeklyRepos mirrors test-stats' primary-checkout fleet walk. A slot
// checkout cannot replace its Git-verified primary in the report.
func discoverWeeklyRepos(id workspace.Identity) ([]string, error) {
	paths, err := project.FleetRepoDirs(id.FleetRoot)
	if err != nil {
		return nil, err
	}
	paths = append(paths, id.PrimaryRoot)
	byName := map[string]string{}
	for _, path := range paths {
		if path != "" && weeklyIsSDLCRepo(path) {
			byName[filepath.Base(path)] = path
		}
	}
	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)
	var out []string
	for _, name := range names {
		out = append(out, byName[name])
	}
	return out, nil
}

func weeklyIsSDLCRepo(root string) bool {
	if _, err := os.Stat(filepath.Join(root, ".brain", "config.md")); err == nil {
		return false // brain is capture-only, even if an old issue directory remains
	}
	info, err := os.Stat(filepath.Join(root, vocab.Issue().Discovery().Home))
	return err == nil && info.IsDir()
}

// weeklyReadLandingEvents is read-only; landingline.Open would create state.
func weeklyReadLandingEvents(path string) ([]landingline.Event, int, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	if info.Size() > weeklyInputLimit {
		return nil, 0, fmt.Errorf("event log exceeds 64 MiB")
	}
	r := bufio.NewReader(io.LimitReader(f, weeklyInputLimit+1))
	var events []landingline.Event
	skipped := 0
	for {
		line, readErr := r.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			var event landingline.Event
			if len(line) > weeklyInputLimit || json.Unmarshal(line, &event) != nil {
				skipped++
			} else {
				events = append(events, event)
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, skipped, readErr
		}
	}
	return events, skipped, nil
}

type weeklyLimitWriter struct {
	bytes.Buffer
	exceeded bool
}

func (w *weeklyLimitWriter) Write(p []byte) (int, error) {
	if len(p) > (8<<20)-w.Len() {
		w.exceeded = true
		return 0, fmt.Errorf("weekly Git output exceeds 8 MiB")
	}
	return w.Buffer.Write(p)
}

func weeklyGit(ctx context.Context, root string, args ...string) ([]byte, error) {
	if ctx == nil {
		return nil, fmt.Errorf("weekly Git requires the caller's context")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = time.Second
	var out weeklyLimitWriter
	var stderr weeklyLimitWriter
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	_ = processgroup.Terminate(cmd, true)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if out.exceeded || stderr.exceeded {
		return nil, fmt.Errorf("weekly Git output exceeds 8 MiB")
	}
	if err != nil {
		classified := gitx.ClassifyNetworkFailure(args, stderr.Bytes(), err)
		if classified != err {
			return nil, fmt.Errorf("git %s: %w", args[0], classified)
		}
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return out.Bytes(), nil
}

func readWeeklyEvidence(ctx context.Context, repo, root string, window WeeklyWindow) weeklyRepoEvidence {
	e := weeklyRepoEvidence{Repo: repo, Root: root, Cards: map[int]weeklyCardEvidence{}, Graphs: map[string]weeklyMergeGraph{}, RecentMainMerges: map[string]time.Time{}}
	commonOut, err := weeklyGit(ctx, root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		e.CoverageUnknown = append(e.CoverageUnknown, "Git common directory: "+err.Error())
		return e
	}
	common := strings.TrimSpace(string(commonOut))
	_, fetchErr := weeklyGit(ctx, root, "fetch", "--no-tags", "origin", "main")
	if fetchErr != nil {
		e.CoverageUnknown = append(e.CoverageUnknown, "fetch main: "+fetchErr.Error())
	}
	mainOut, err := weeklyGit(ctx, root, "rev-parse", "FETCH_HEAD")
	if err != nil {
		e.CoverageUnknown = append(e.CoverageUnknown, "main ref: "+err.Error())
	} else if fetchErr == nil {
		e.MainRef = strings.TrimSpace(string(mainOut))
	}
	if e.MainRef != "" {
		mergeOut, err := weeklyGit(ctx, root, "log", "--merges", "--since="+window.Start.Format(time.RFC3339), "--until="+window.End.Format(time.RFC3339), "--format=%H%x09%cI", e.MainRef)
		if err != nil {
			e.CoverageUnknown = append(e.CoverageUnknown, "main merge inventory: "+err.Error())
		} else {
			for _, line := range strings.Split(strings.TrimSpace(string(mergeOut)), "\n") {
				if line == "" {
					continue
				}
				fields := strings.Split(line, "\t")
				if len(fields) != 2 {
					e.CoverageUnknown = append(e.CoverageUnknown, "malformed main merge inventory")
					continue
				}
				at, parseErr := time.Parse(time.RFC3339, fields[1])
				if parseErr != nil {
					e.CoverageUnknown = append(e.CoverageUnknown, "malformed main merge time: "+fields[0])
					continue
				}
				if window.Contains(at) {
					e.RecentMainMerges[fields[0]] = at
				}
			}
		}
	}
	rootStats := filepath.Join(common, "sdlc", "test-stats")
	snapshot, err := readSuiteEvidence(rootStats, testStatsRecordLimit)
	attempts := snapshot.Attempts
	warnings := append(snapshot.Warnings, snapshot.coverageNotes("")...)
	if err != nil {
		e.CoverageUnknown = append(e.CoverageUnknown, "suite attempts: "+err.Error())
	} else {
		for _, a := range attempts {
			if a.Check != suiteRerunCheck { // a known-flake rerun (#394) is not a full suite; other checks stay as before
				e.Attempts = append(e.Attempts, a)
			}
		}
		for _, warning := range warnings {
			e.Warnings = append(e.Warnings, "suite attempts: "+warning)
			e.CoverageUnknown = append(e.CoverageUnknown, "suite attempts: "+warning)
		}
	}
	known := snapshot.knownSuiteHeads()
	for i := range e.Attempts {
		e.Attempts[i].Repo = repo
		if isSuiteRun(e.Attempts[i]) {
			known[e.Attempts[i].SHA] = true
		}
	}
	legacy, legacyWarnings := readLegacySuiteLogs(common, repo, known)
	e.Attempts = append(e.Attempts, legacy...)
	for _, warning := range legacyWarnings {
		e.Warnings = append(e.Warnings, warning)
		e.CoverageUnknown = append(e.CoverageUnknown, warning)
	}
	var skipped int
	e.Events, skipped, err = weeklyReadLandingEvents(filepath.Join(common, "sdlc", "landing", landingline.EventsFile))
	if err != nil {
		e.CoverageUnknown = append(e.CoverageUnknown, "landing events: "+err.Error())
	} else {
		if skipped > 0 {
			e.CoverageUnknown = append(e.CoverageUnknown, fmt.Sprintf("landing events: %d malformed lines", skipped))
		}
	}
	landedInLog := map[int]bool{}
	for _, event := range e.Events {
		if event.Kind == landingline.EventLanded && window.Contains(event.At) {
			landedInLog[event.Issue] = true
		}
	}
	tr, err := tracker.NewRepository(ctx, root, "origin")
	if err != nil {
		e.CoverageUnknown = append(e.CoverageUnknown, "tracker: "+err.Error())
		return e
	}
	tr.BoundRemote(30 * time.Second)
	snap, err := tr.Snapshot()
	if err != nil {
		e.CoverageUnknown = append(e.CoverageUnknown, "tracker snapshot: "+err.Error())
		return e
	}
	e.TrackerRef = snap.Ref()
	for _, unreadable := range snap.Unreadable() {
		e.CoverageUnknown = append(e.CoverageUnknown, "tracker card "+unreadable.ID+": "+unreadable.Err.Error())
	}
	for _, record := range snap.Records() {
		id, err := strconv.Atoi(record.ID)
		if err != nil {
			e.CoverageUnknown = append(e.CoverageUnknown, "tracker card invalid ID "+record.ID)
			continue
		}
		status, _ := issue.GetField(record.Card.Frontmatter, "status")
		card := weeklyCardEvidence{Status: status}
		created, _ := issue.GetField(record.Card.Frontmatter, "created")
		card.CreatedAt, _ = time.Parse("2006-01-02", created)
		completion, ok, err := issue.CardCompletion(record.Raw)
		if err != nil {
			if landedInLog[id] {
				e.CoverageUnknown = append(e.CoverageUnknown, fmt.Sprintf("tracker card #%d completion: %v", id, err))
			}
		} else if ok {
			card.Completion = &completion
			if completion.LandedCommit != "" && landedInLog[id] {
				if _, inWindow := e.RecentMainMerges[completion.LandedCommit]; !inWindow {
					e.Cards[id] = card
					continue
				}
				if _, exists := e.Graphs[completion.LandedCommit]; !exists {
					graph, err := weeklyReadGraph(ctx, root, e.MainRef, completion)
					if err != nil {
						e.CoverageUnknown = append(e.CoverageUnknown, fmt.Sprintf("tracker card #%d Git graph: %v", id, err))
					} else {
						e.Graphs[completion.LandedCommit] = graph
					}
				}
			}
		}
		e.Cards[id] = card
	}
	return e
}

func weeklyReadGraph(ctx context.Context, root, mainRef string, c issue.Completion) (weeklyMergeGraph, error) {
	g := weeklyMergeGraph{MergeSHA: c.LandedCommit, BranchSHAs: map[string]bool{}}
	parentsOut, err := weeklyGit(ctx, root, "rev-list", "--parents", "-n", "1", c.LandedCommit)
	if err != nil {
		return g, err
	}
	parents := strings.Fields(string(parentsOut))
	if len(parents) != 3 || parents[0] != c.LandedCommit {
		return g, fmt.Errorf("landing %s has no two-parent merge", c.LandedCommit)
	}
	g.MainParent, g.IssueParent = parents[1], parents[2]
	subjectOut, err := weeklyGit(ctx, root, "show", "-s", "--format=%s", c.LandedCommit)
	if err != nil {
		return g, err
	}
	identity := weeklyPullMergeSubject.FindStringSubmatch(strings.TrimSpace(string(subjectOut)))
	if len(identity) != 3 {
		return g, fmt.Errorf("landing %s has no GitHub PR/branch identity", c.LandedCommit)
	}
	g.PR, err = strconv.Atoi(identity[1])
	if err != nil {
		return g, fmt.Errorf("landing %s has invalid PR: %w", c.LandedCommit, err)
	}
	g.Branch = identity[2]
	issueOut, err := weeklyGit(ctx, root, "rev-list", g.IssueParent, "--not", g.MainParent)
	if err != nil {
		return g, err
	}
	for _, sha := range strings.Fields(string(issueOut)) {
		g.BranchSHAs[sha] = true
	}
	if !g.BranchSHAs[c.ReviewedHEAD] {
		return g, fmt.Errorf("reviewed HEAD not on issue side of landing")
	}
	if !g.BranchSHAs[c.EvidenceCommit] {
		return g, fmt.Errorf("evidence commit not on issue side of landing")
	}
	_, err = weeklyGit(ctx, root, "merge-base", "--is-ancestor", c.EvidenceCommit, g.IssueParent)
	if err != nil {
		return g, fmt.Errorf("evidence commit not reachable from issue parent: %w", err)
	}
	g.EvidenceReachable = true
	if mainRef == "" {
		return g, fmt.Errorf("main ref unavailable")
	}
	_, err = weeklyGit(ctx, root, "merge-base", "--is-ancestor", c.LandedCommit, mainRef)
	if err != nil {
		return g, fmt.Errorf("landing not reachable from main: %w", err)
	}
	g.ReachableFromMain = true
	return g, nil
}
