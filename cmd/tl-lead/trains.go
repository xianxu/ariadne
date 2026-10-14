package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xianxu/ariadne/pkg/markdoc"
	"github.com/xianxu/ariadne/pkg/trains"
)

// readConcurrency bounds the parallel `sdlc issue show` calls one render makes
// (each is a process; a project has tens of cars, a fleet tens of slots).
const readConcurrency = 8

// boundedEach runs f(0..n-1) with at most readConcurrency running at once and
// returns when all have finished.
func boundedEach(n int, f func(i int)) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, readConcurrency)
	for i := range n {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() { <-sem; wg.Done() }()
			f(i)
		}()
	}
	wg.Wait()
}

// gatherTrains renders a project's trains block from sources: the project's
// trains table, sdlc's observation of each car, and each claimant checkout's
// issue file for the bar. Nothing is typed by hand, so a rendered line can't
// drift from the truth. The TL pastes it into the project's `### Project`
// thread.
func gatherTrains(sd SDLC, projectPath, defaultRepo string) (string, error) {
	raw, err := os.ReadFile(projectPath)
	if err != nil {
		return "", err
	}
	section, ok := markdoc.SectionBody(string(raw), "Trains")
	if !ok {
		return "", fmt.Errorf("%s has no `## Trains` section", projectPath)
	}
	plan := trains.ParseTrains(section, defaultRepo)
	if len(plan.Trains) == 0 {
		return "", fmt.Errorf("%s: no `| Train | Cars, in order | Slot |` rows under `## Trains`", projectPath)
	}
	seen := map[trains.IssueRef]bool{}
	var refs []trains.IssueRef
	for _, t := range plan.Trains {
		for _, r := range t.Cars {
			if !seen[r] {
				seen[r] = true
				refs = append(refs, r)
			}
		}
	}
	return renderTrains(plan, readCars(sd, filepath.Dir(projectPath), refs)), nil
}

// readCars observes every car, at most readConcurrency at once. A car whose
// read fails carries the reason, so it renders `?`. Each repository's checkout
// is resolved once, before the reads start.
func readCars(sd SDLC, vantage string, refs []trains.IssueRef) map[trains.IssueRef]CarState {
	repoPath := map[string]string{}
	repoErr := map[string]string{}
	for _, r := range refs {
		if _, seen := repoPath[r.Repo]; seen {
			continue
		}
		if _, seen := repoErr[r.Repo]; seen {
			continue
		}
		if r.Repo == "" {
			repoErr[r.Repo] = "a bare #N with no repository (pass --default-repo)"
			continue
		}
		p, err := sd.RepoRoot(vantage, r.Repo)
		if err != nil {
			repoErr[r.Repo] = err.Error()
			continue
		}
		repoPath[r.Repo] = p
	}
	cars := make(map[trains.IssueRef]CarState, len(refs))
	var mu sync.Mutex
	boundedEach(len(refs), func(i int) {
		ref := refs[i]
		var c CarState
		if why, bad := repoErr[ref.Repo]; bad {
			c = CarState{Err: why}
		} else {
			c = readCar(sd, ref, repoPath[ref.Repo])
		}
		mu.Lock()
		cars[ref] = c
		mu.Unlock()
	})
	return cars
}

func readCar(sd SDLC, ref trains.IssueRef, repoPath string) CarState {
	o, err := sd.IssueShow(repoPath, ref.ID)
	if err != nil {
		return CarState{Err: err.Error()}
	}
	c := CarState{Status: o.Status, Claimant: o.Claimant}
	if o.ClaimantPath != "" {
		c.Bar = issueBar(o.ClaimantPath, ref.ID)
	} else {
		c.Bar = "no checkout"
	}
	return c
}

// issueBar reads issue id's details file in a checkout and renders its bar.
// The directory is listed and matched by name (never globbed: a checkout path
// may hold glob metacharacters).
func issueBar(checkout string, id int) string {
	dir := filepath.Join(checkout, "workshop", "issues")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "no issue file"
		}
		return "? " + err.Error()
	}
	prefix := fmt.Sprintf("%06d-", id)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) && strings.HasSuffix(e.Name(), ".md") {
			raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				return "? " + err.Error()
			}
			return planBar(string(raw))
		}
	}
	return "no issue file"
}
