package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// sectionBudget is how long a section may go without a write. A zero budget
// checks only that the section is stamped (Needs you is event-driven and may
// rightly stay unchanged for hours); landing also makes it stale once the
// fleet lands something newer than its stamp. Operator instructions is the
// operator's and has no row.
type sectionBudget struct {
	heading string
	budget  time.Duration
	landing bool
}

// missingSection is in couch's refusal to read a section the file lacks
// (pair sessionsync/file.go); TestSessionSyncConformance pins it.
const missingSection = "heading line, found 0"

var sectionBudgets = []sectionBudget{
	{"Needs you", 0, false},
	{"Quota", 12 * time.Hour, false},
	{"Threads", 2 * time.Hour, true},
	{"Slots", 2 * time.Hour, true},
}

// Landing is the fleet's newest first-parent merge on origin/main among
// the SDLC repositories next to the status file's.
type Landing struct {
	Repo   string
	Commit string
	At     time.Time
}

// staleLines reports each budgeted section that is unstamped, past its
// budget, or older than the newest landing. A section the file lacks is not
// reported; a failed read is.
func staleLines(d deps, file string) []string {
	now := d.clock()
	land, landErr := d.newestLanding(file)
	var out []string
	if landErr != nil {
		out = append(out, "check: newest landing unknown: "+landErr.Error())
	}
	for _, b := range sectionBudgets {
		body, _, err := d.ss.ReadSection(file, b.heading)
		if err != nil {
			if !strings.Contains(err.Error(), missingSection) {
				out = append(out, fmt.Sprintf("check: ## %s unread: %v", b.heading, err))
			}
			continue
		}
		at, _ := splitStamp(body)
		var why []string
		switch {
		case at.IsZero():
			why = append(why, "unstamped")
		default:
			if age := now.Sub(at); b.budget > 0 && age > b.budget {
				why = append(why, fmt.Sprintf("%s old, budget %s", hm(age), hm(b.budget)))
			}
			if b.landing && land.At.Truncate(time.Minute).After(at) { // the stamp has minutes only
				why = append(why, fmt.Sprintf("%s merge %.12s landed %s", land.Repo, land.Commit, land.At.Format("01-02 15:04")))
			}
		}
		if len(why) > 0 {
			stamp := ""
			if !at.IsZero() {
				stamp = "updated " + at.Format("01-02 15:04") + "; "
			}
			out = append(out, fmt.Sprintf("stale: ## %s (%s%s)", b.heading, stamp, strings.Join(why, "; ")))
		}
	}
	return out
}

// printStale ends a status verb's output with the stale sections, so every
// event the TL handles shows what has gone stale.
func printStale(d deps, out io.Writer, file string) {
	lines := staleLines(d, file)
	if len(lines) == 0 {
		fmt.Fprintln(out, "fresh: every budgeted section")
		return
	}
	fmt.Fprintln(out, strings.Join(lines, "\n"))
}

// hm renders a duration as 3h05m, 2h or 45m.
func hm(d time.Duration) string {
	h, m := int(d.Hours()), int(d.Minutes())%60
	switch {
	case h == 0:
		return fmt.Sprintf("%dm", m)
	case m == 0:
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%02dm", h, m)
}

// newestLanding reads the landing seam; without one there is no landing (main
// wires gitNewestLanding).
func (d deps) newestLanding(file string) (Landing, error) {
	if d.landing == nil {
		return Landing{}, nil
	}
	return d.landing(file)
}

// gitNewestLanding reads origin/main's latest first-parent merge, from local refs (no
// fetch), in the local hub's explicitly claimed SDLC repositories (or every
// sibling SDLC repository when no local .hub exists): one
// with workshop/issues/ and no .brain/ (a brain's auto-commits are captures,
// not landings; sdlc's spine guard draws the same line). sdlc merges PRs with
// gh pr merge --merge. The first-parent restriction excludes issue-branch
// integrations; direct-main filing, publish, archive and tick commits are skipped.
func gitNewestLanding(file string) (Landing, error) {
	scope, err := readHubScope(filepath.Dir(file))
	if err != nil {
		return Landing{}, err
	}
	top, err := git(filepath.Dir(file), "rev-parse", "--show-toplevel")
	if err != nil {
		return Landing{}, err
	}
	parent := filepath.Dir(top)
	var repos []string
	if scope.enabled {
		repos = scope.repoNames()
	} else {
		entries, err := os.ReadDir(parent)
		if err != nil {
			return Landing{}, err
		}
		for _, e := range entries {
			repos = append(repos, e.Name())
		}
	}
	var newest Landing
	for _, repo := range repos {
		dir := filepath.Join(parent, repo)
		if !isDir(filepath.Join(dir, "workshop", "issues")) || isDir(filepath.Join(dir, ".brain")) {
			continue
		}
		ct, err := git(dir, "log", "--first-parent", "--merges", "-1", "--format=%ct %H", "refs/remotes/origin/main")
		if err != nil {
			continue // no origin/main: nothing lands there
		}
		fields := strings.Fields(ct)
		if len(fields) != 2 {
			continue // no merge: no landing
		}
		sec, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			continue
		}
		if at := time.Unix(sec, 0); at.After(newest.At) {
			newest = Landing{Repo: repo, Commit: fields[1], At: at}
		}
	}
	return newest, nil
}

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

func git(dir string, args ...string) (string, error) {
	out, err := gitWithInput(dir, nil, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Keep byte-delimited Git output intact for callers using --stdin and -z.
func gitWithInput(dir string, input io.Reader, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Stdin = input
	out, err := cmd.Output()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok && len(exit.Stderr) > 0 {
			return string(out), fmt.Errorf("git %s in %s: %w: %s", strings.Join(args, " "), dir, err, strings.TrimSpace(string(exit.Stderr)))
		}
		return string(out), fmt.Errorf("git %s in %s: %w", strings.Join(args, " "), dir, err)
	}
	return string(out), nil
}
