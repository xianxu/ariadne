package bench

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// Task is the immutable, replayable definition of one benchmark experiment,
// frozen from a live issue. It never changes after `sdlc bench freeze`, so any
// agent can be replayed against the exact same starting conditions even after
// main has advanced. Pure: ParseTask/RenderTask are inverses with no IO.
type Task struct {
	ID          string // slug, usually "<issue>-<short-slug>"; matches the filename
	Repo        string
	SourceIssue string
	BaseSHA     string // immutable commit agents branch from (reproducibility anchor)
	Created     string // ISO date
	Spec        string // the issue's ## Spec, copied verbatim
	Setup       []string
	Rubric      Rubric
}

// taskConfig is the JSON payload embedded in the ## Config section. Keeping
// setup+rubric in one block (vs frontmatter) lets them carry structure without a
// YAML dependency — the repo is stdlib + cobra only.
type taskConfig struct {
	Setup  []string `json:"setup"`
	Rubric Rubric   `json:"rubric"`
}

// RenderTask serializes a Task to its on-disk markdown form.
func RenderTask(t Task) string {
	fm := strings.Join([]string{
		"type: benchmark-task",
		"id: " + t.ID,
		"repo: " + t.Repo,
		"source_issue: " + t.SourceIssue,
		"base_sha: " + t.BaseSHA,
		"created: " + t.Created,
	}, "\n")
	cfg, _ := json.MarshalIndent(taskConfig{Setup: t.Setup, Rubric: t.Rubric}, "", "  ")
	body := fmt.Sprintf("# Benchmark task: %s\n\n## Spec\n\n%s\n\n## Config\n\n```json\n%s\n```\n",
		t.ID, strings.TrimRight(t.Spec, "\n"), string(cfg))
	return issue.Compose(fm, body)
}

// ParseTask is the inverse of RenderTask.
func ParseTask(text string) (Task, error) {
	fm, body, err := issue.Parse(text)
	if err != nil {
		return Task{}, err
	}
	get := func(k string) string { v, _ := issue.GetField(fm, k); return v }
	t := Task{
		ID:          get("id"),
		Repo:        get("repo"),
		SourceIssue: get("source_issue"),
		BaseSHA:     get("base_sha"),
		Created:     get("created"),
	}
	if spec, ok := issue.SectionBody(body, "Spec"); ok {
		t.Spec = strings.TrimSpace(spec)
	}
	// Scope extraction to the ## Config section — a verbatim spec may itself
	// contain a ```json fence, and extractJSONBlock returns the FIRST one.
	cfgBody, ok := issue.SectionBody(body, "Config")
	if !ok {
		return Task{}, fmt.Errorf("benchmark-task %s: missing ## Config section", t.ID)
	}
	raw, ok := extractJSONBlock(cfgBody)
	if !ok {
		return Task{}, fmt.Errorf("benchmark-task %s: missing ## Config json block", t.ID)
	}
	var cfg taskConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return Task{}, fmt.Errorf("benchmark-task %s: bad config json: %w", t.ID, err)
	}
	t.Setup, t.Rubric = cfg.Setup, cfg.Rubric
	return t, nil
}

// extractJSONBlock returns the content of the first ```json fenced block in s.
// Callers scope s to a single section so "first" is unambiguous (shared by Task
// and RunRecord parsing — one block-parser, DRY).
func extractJSONBlock(s string) (string, bool) {
	const open = "```json"
	i := strings.Index(s, open)
	if i < 0 {
		return "", false
	}
	rest := s[i+len(open):]
	j := strings.Index(rest, "```")
	if j < 0 {
		return "", false
	}
	return strings.TrimSpace(rest[:j]), true
}
