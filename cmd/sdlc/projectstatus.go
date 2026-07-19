package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/vocab"
)

type projectStatusFlags struct{ Slug, ProjectsDir, BrainDir string }

var projectIssueLookupFn = lookupIssueMeta

func newProjectStatusCmd() *cobra.Command {
	f := projectStatusFlags{}
	cmd := &cobra.Command{Use: "status", Short: "Render the derived project board", Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProjectStatus(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		}}
	cmd.Flags().StringVar(&f.Slug, "slug", "", "project slug")
	cmd.Flags().StringVar(&f.ProjectsDir, "projects-dir", defaultProjectsDir(), "directory holding project files")
	cmd.Flags().StringVar(&f.BrainDir, "brain-dir", "../brain", "brain root holding the throughput baseline (for the forecast line)")
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
	if line := forecastLine(path, f.BrainDir, projectTodayFn()); line != "" {
		fmt.Fprintln(stdout, line)
	}
	return nil
}

type issueMeta struct {
	Identity                   string
	Status                     string
	EstimateHours, ActualHours float64
	ActualAvailable, ActualNA  bool
	Deps                       []string
}

type boardRow struct {
	RefText, Title, IssueStatus string
	Identity                    string
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
	metadata, err := d.Metadata()
	if err != nil {
		return board{}, err
	}
	b := board{Name: metadata.Name, Status: metadata.Status, Deadline: metadata.Deadline, PlannedFinish: metadata.PlannedFinish, Total: len(d.Tasks), LastRetro: projectdoc.LatestRetroDate(d)}
	metas := map[string]issueMeta{}
	display := map[string]string{}
	refIDs := map[string]string{}
	for _, task := range d.Tasks {
		row := boardRow{RefText: task.RefText, Title: task.Title, Ticked: task.State == 'x'}
		if row.Ticked {
			b.Done++
		}
		if task.RefText != "" {
			meta, err := lookup(task.RefText)
			if err != nil {
				row.Identity = "raw:" + task.RefText
				row.IssueStatus = "unresolved"
				row.Warning = err.Error()
			} else {
				row.Identity = valueOr(meta.Identity, "raw:"+task.RefText)
				_, duplicate := display[row.Identity]
				if !duplicate {
					display[row.Identity] = task.RefText
				}
				refIDs[task.RefText] = row.Identity
				metas[row.Identity] = meta
				row.IssueStatus = meta.Status
				if row.Ticked && !vocab.Issue().IsTerminal(meta.Status) {
					row.Warning = "task ticked but issue is not terminal"
				}
				if duplicate && row.Warning == "" {
					row.Warning = "duplicate logical issue reference (first " + display[row.Identity] + ")"
				}
			}
		}
		b.Rows = append(b.Rows, row)
	}

	unfinished := map[string]bool{}
	order := []string{}
	for _, row := range b.Rows {
		if !row.Ticked && row.RefText != "" {
			id := valueOr(row.Identity, "raw:"+row.RefText)
			if !unfinished[id] {
				unfinished[id] = true
				order = append(order, id)
				if display[id] == "" {
					display[id] = row.RefText
				}
			}
		}
	}
	graph := map[string]map[string]bool{}
	for _, ref := range order {
		graph[ref] = map[string]bool{}
		if meta, ok := metas[ref]; ok && !vocab.Issue().IsTerminal(meta.Status) {
			b.RemainingHours += meta.EstimateHours
		}
	}
	for _, id := range order {
		meta, ok := metas[id]
		if !ok {
			continue
		}
		blocked := false
		for _, dep := range meta.Deps {
			depID := refIDs[dep]
			depMeta, ok := metas[depID]
			if !ok {
				var err error
				depMeta, err = lookup(dep)
				ok = err == nil
				if ok {
					depID = valueOr(depMeta.Identity, "raw:"+dep)
					metas[depID] = depMeta
					refIDs[dep] = depID
				}
			}
			if ok && unfinished[depID] {
				graph[id][depID] = true
				graph[depID][id] = true
			}
			if !ok || !vocab.Issue().IsTerminal(depMeta.Status) {
				blocked = true
			}
		}
		if blocked || vocab.Issue().IsEventTarget(meta.Status, "block") {
			b.Blocked = append(b.Blocked, display[id])
		} else if !vocab.Issue().IsTerminal(meta.Status) {
			b.Frontier = append(b.Frontier, display[id])
		}
	}
	for _, component := range connectedThreads(order, graph) {
		thread := make([]string, len(component))
		for i, id := range component {
			thread[i] = display[id]
		}
		b.Threads = append(b.Threads, thread)
	}
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
	lastRetro := valueOr(b.LastRetro, "-")
	if retroDate, err := time.Parse("2006-01-02", b.LastRetro); err == nil {
		if now, err := time.Parse("2006-01-02", today); err == nil {
			lastRetro = fmt.Sprintf("%s (%d days ago)", b.LastRetro, int(now.Sub(retroDate).Hours()/24))
		}
	}
	fmt.Fprintf(&s, "last retro: %s\n", lastRetro)
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
	decoded, err := projectdoc.DecodeMetadata(fm)
	if err != nil {
		return issueMeta{}, fmt.Errorf("%s: %w", refText, err)
	}
	meta := issueMeta{Identity: canonicalRepoIssueIdentity(repoDir, ref.ID), Status: decoded.Status, Deps: decoded.Deps}
	meta.EstimateHours, _, _, err = projectdoc.NumberValue(decoded.EstimateHours, "estimate_hours")
	if err != nil {
		return issueMeta{}, fmt.Errorf("%s has %w", refText, err)
	}
	meta.ActualHours, meta.ActualAvailable, meta.ActualNA, err = projectdoc.NumberValue(decoded.ActualHours, "actual_hours")
	if err != nil {
		return issueMeta{}, fmt.Errorf("%s has %w", refText, err)
	}
	return meta, nil
}
