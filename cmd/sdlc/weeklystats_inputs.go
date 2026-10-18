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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/workspace"
)

const weeklyInputLimit = 64 << 20

type weeklyCardEvidence struct {
	Status     string
	Completion *issue.Completion
}

type weeklyMergeGraph struct {
	MergeSHA, MainParent, IssueParent    string
	BranchSHAs                           map[string]bool
	EvidenceReachable, ReachableFromMain bool
}

type weeklyRepoEvidence struct {
	Repo, Root, MainRef, TrackerRef string
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
		if path != "" {
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
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = time.Second
	var out weeklyLimitWriter
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	_ = processgroup.Terminate(cmd, true)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if out.exceeded {
		return nil, fmt.Errorf("weekly Git output exceeds 8 MiB")
	}
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return out.Bytes(), nil
}

func readWeeklyEvidence(ctx context.Context, repo, root string) weeklyRepoEvidence {
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
		mergeOut, err := weeklyGit(ctx, root, "log", "--merges", "--format=%H%x09%cI", e.MainRef)
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
				e.RecentMainMerges[fields[0]] = at
			}
		}
	}
	rootStats := filepath.Join(common, "sdlc", "test-stats")
	attempts, warnings, err := readSuiteAttempts(rootStats, testStatsRecordLimit)
	if err != nil {
		e.CoverageUnknown = append(e.CoverageUnknown, "suite attempts: "+err.Error())
	} else {
		e.Attempts = attempts
		for _, warning := range warnings {
			e.Warnings = append(e.Warnings, "suite attempts: "+warning)
			e.CoverageUnknown = append(e.CoverageUnknown, "suite attempts: "+warning)
		}
	}
	known := map[string]bool{}
	for i := range e.Attempts {
		e.Attempts[i].Repo = repo
		if e.Attempts[i].Check == suiteCheckName {
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
		if event.Kind == landingline.EventLanded {
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
		completion, ok, err := issue.CardCompletion(record.Raw)
		if err != nil {
			e.CoverageUnknown = append(e.CoverageUnknown, fmt.Sprintf("tracker card #%d completion: %v", id, err))
		} else if ok {
			card.Completion = &completion
			if completion.LandedCommit != "" && landedInLog[id] {
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
