package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/vocab"
)

type projectStatusFlags struct{ Slug, ProjectsDir string }

var projectIssueLookupFn = lookupIssueMeta

func newProjectStatusCmd() *cobra.Command {
	f := projectStatusFlags{}
	cmd := &cobra.Command{Use: "status", Short: "Render the derived project board", Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProjectStatus(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		}}
	cmd.Flags().StringVar(&f.Slug, "slug", "", "project slug")
	cmd.Flags().StringVar(&f.ProjectsDir, "projects-dir", defaultProjectsDir(), "directory holding project files")
	_ = cmd.MarkFlagRequired("slug")
	return cmd
}

func runProjectStatus(stdout, _ io.Writer, f *projectStatusFlags) error {
	path, err := projectdoc.ResolvePath(f.ProjectsDir, f.Slug)
	if err != nil {
		return err
	}
	d, err := readProject(path)
	if err != nil {
		return err
	}
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	b, err := computeBoard(d, func(ref string) (issueMeta, error) { return projectIssueLookupFn(ref, root) })
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, renderBoard(b, projectTodayFn()))
	return nil
}

type issueMeta struct {
	Status                     string
	EstimateHours, ActualHours float64
	ActualAvailable, ActualNA  bool
	Deps                       []string
}

type boardRow struct {
	RefText, Title, IssueStatus string
	Ticked                      bool
	RemainingHours              float64
	Warning                     string
}

type board struct {
	Name, Status            string
	Rows                    []boardRow
	Done, Total             int
	RemainingHours          float64
	Deadline, PlannedFinish string
	Frontier, Blocked       []string
	Threads                 [][]string
	LastRetro               string
}

func computeBoard(d *projectdoc.Doc, lookup func(string) (issueMeta, error)) (board, error) {
	b := board{Name: d.FM("name"), Status: d.FM("status"), Deadline: d.FM("deadline"), PlannedFinish: d.FM("planned_finish"), Total: len(d.Tasks), LastRetro: projectdoc.LatestRetroDate(d)}
	metas := map[string]issueMeta{}
	for _, task := range d.Tasks {
		row := boardRow{RefText: task.RefText, Title: task.Title, Ticked: task.State == 'x'}
		if row.Ticked {
			b.Done++
		}
		if task.RefText != "" {
			meta, err := lookup(task.RefText)
			if err != nil {
				row.IssueStatus = "unresolved"
				row.Warning = err.Error()
			} else {
				metas[task.RefText] = meta
				row.IssueStatus = meta.Status
				if row.Ticked && !vocab.Issue().IsTerminal(meta.Status) {
					row.Warning = "task ticked but issue is not terminal"
				}
				if !row.Ticked && !vocab.Issue().IsTerminal(meta.Status) {
					row.RemainingHours = meta.EstimateHours
					b.RemainingHours += meta.EstimateHours
				}
			}
		}
		b.Rows = append(b.Rows, row)
	}

	unfinished := map[string]bool{}
	order := []string{}
	for _, row := range b.Rows {
		if !row.Ticked && row.RefText != "" {
			unfinished[row.RefText] = true
			order = append(order, row.RefText)
		}
	}
	graph := map[string]map[string]bool{}
	for _, ref := range order {
		graph[ref] = map[string]bool{}
	}
	for _, ref := range order {
		meta, ok := metas[ref]
		if !ok {
			continue
		}
		blocked := false
		for _, dep := range meta.Deps {
			if unfinished[dep] {
				graph[ref][dep] = true
				graph[dep][ref] = true
			}
			depMeta, ok := metas[dep]
			if !ok {
				var err error
				depMeta, err = lookup(dep)
				ok = err == nil
				if ok {
					metas[dep] = depMeta
				}
			}
			if !ok || !vocab.Issue().IsTerminal(depMeta.Status) {
				blocked = true
			}
		}
		if blocked || meta.Status == "blocked" {
			b.Blocked = append(b.Blocked, ref)
		} else if !vocab.Issue().IsTerminal(meta.Status) {
			b.Frontier = append(b.Frontier, ref)
		}
	}
	b.Threads = connectedThreads(order, graph)
	return b, nil
}

func connectedThreads(order []string, graph map[string]map[string]bool) [][]string {
	seen := map[string]bool{}
	var out [][]string
	for _, start := range order {
		if seen[start] {
			continue
		}
		queue := []string{start}
		seen[start] = true
		var component []string
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			component = append(component, cur)
			neighbors := make([]string, 0, len(graph[cur]))
			for n := range graph[cur] {
				neighbors = append(neighbors, n)
			}
			sort.Strings(neighbors)
			for _, n := range neighbors {
				if !seen[n] {
					seen[n] = true
					queue = append(queue, n)
				}
			}
		}
		out = append(out, component)
	}
	return out
}

func renderBoard(b board, today string) string {
	var s strings.Builder
	fmt.Fprintf(&s, "%s — %s\n", b.Name, b.Status)
	deadline := b.Deadline
	if deadline != "" {
		if d, err := time.Parse("2006-01-02", deadline); err == nil {
			if now, err := time.Parse("2006-01-02", today); err == nil {
				deadline = fmt.Sprintf("%s (%d days)", deadline, int(d.Sub(now).Hours()/24))
			}
		}
	}
	fmt.Fprintf(&s, "deadline: %s · planned finish: %s\n", valueOr(deadline, "-"), valueOr(b.PlannedFinish, "-"))
	fmt.Fprintf(&s, "progress: %d/%d done · Σ remaining ≈ %gh\n", b.Done, b.Total, b.RemainingHours)
	fmt.Fprintf(&s, "frontier: %s\n", valueOr(strings.Join(b.Frontier, ", "), "-"))
	fmt.Fprintf(&s, "blocked: %s\n", valueOr(strings.Join(b.Blocked, ", "), "-"))
	parts := make([]string, len(b.Threads))
	for i, thread := range b.Threads {
		parts[i] = "[" + strings.Join(thread, ", ") + "]"
	}
	fmt.Fprintf(&s, "threads: %d — %s\n", len(b.Threads), strings.Join(parts, " / "))
	fmt.Fprintf(&s, "last retro: %s\n", valueOr(b.LastRetro, "-"))
	for _, row := range b.Rows {
		if row.Warning != "" {
			fmt.Fprintf(&s, "warning: %s — %s\n", valueOr(row.RefText, row.Title), row.Warning)
		}
	}
	return s.String()
}

func lookupIssueMeta(refText, currentRepoRoot string) (issueMeta, error) {
	ref, err := parseRef(refText)
	if err != nil {
		return issueMeta{}, err
	}
	if ref.GitHub {
		return issueMeta{}, fmt.Errorf("GitHub ref %q has no local issue metadata", refText)
	}
	repoDir, err := resolveRepoDir(ref, currentRepoRoot)
	if err != nil {
		return issueMeta{}, err
	}
	disc := vocab.Issue().Discovery()
	path, err := locateIssueFile(filepath.Join(repoDir, disc.Home), ref.ID)
	if err != nil {
		archive := filepath.Join(repoDir, vocab.ArchiveSubdir(disc.Archive, vocab.ArchiveIssues))
		path, err = locateIssueFile(archive, ref.ID)
		if err != nil {
			return issueMeta{}, fmt.Errorf("resolve %s: %w", refText, err)
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return issueMeta{}, err
	}
	fm, _, err := issue.Parse(string(raw))
	if err != nil {
		return issueMeta{}, err
	}
	meta := issueMeta{}
	meta.Status, _ = issue.GetField(fm, "status")
	if value, _ := issue.GetField(fm, "estimate_hours"); value != "" {
		meta.EstimateHours, _ = strconv.ParseFloat(value, 64)
	}
	if value, _ := issue.GetField(fm, "actual_hours"); value != "" {
		if value == "N/A" {
			meta.ActualNA = true
		} else {
			meta.ActualHours, err = strconv.ParseFloat(value, 64)
			if err != nil {
				return issueMeta{}, fmt.Errorf("%s has invalid actual_hours %q", refText, value)
			}
			meta.ActualAvailable = true
		}
	}
	if value, _ := issue.GetField(fm, "deps"); value != "" {
		meta.Deps = parseInlineRefs(value)
	}
	return meta, nil
}

func parseInlineRefs(value string) []string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(strings.TrimSuffix(value, "]"), "[")
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if ref := strings.TrimSpace(part); ref != "" {
			out = append(out, ref)
		}
	}
	return out
}
