package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

type usageReviewBinding struct {
	Repository, Issue, Slot, Round, Attempt string
	Since, Until                            time.Time
	Conflict                                bool
	Exact                                   bool
}

func loadUsageReviewers(ctx context.Context, repos []string, issueFilter int) ([]reviewusage.Record, usage.Coverage, error) {
	var records []reviewusage.Record
	var coverage usage.Coverage
	seen := map[string]bool{}
	files := 0
	boundHit := false
	for _, root := range repos {
		if ctx.Err() != nil {
			return nil, coverage, ctx.Err()
		}
		store, e := reviewusage.NewStore(root)
		if e == nil {
			entries, e := store.List(ctx)
			if e != nil {
				coverage.Add("unavailable_reviewer_bindings", "")
			} else {
				for _, entry := range entries {
					records = append(records, entry.Record)
					if entry.Record.Phase != reviewusage.Finished {
						coverage.Add("unfinished_reviewer_bindings", "")
					}
				}
			}
		} else {
			coverage.Add("unavailable_reviewer_bindings", "")
		}
		if boundHit {
			break
		}
		for _, dir := range []string{"workshop/plans", "workshop/history/plans"} {
			e = filepath.WalkDir(filepath.Join(root, dir), func(path string, d os.DirEntry, err error) error {
				if err != nil {
					if !os.IsNotExist(err) {
						coverage.Add("unavailable_reviewer_bindings", "")
					}
					return nil
				}
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if d.IsDir() {
					return nil
				}
				name := d.Name()
				if issueFilter > 0 && !strings.HasPrefix(name, fmt.Sprintf("%06d-", issueFilter)) {
					return nil
				}
				if !strings.HasSuffix(name, "-gate.md") && !strings.HasSuffix(name, "-review.md") {
					return nil
				}
				if seen[path] {
					return nil
				}
				seen[path] = true
				files++
				if files > 4096 {
					coverage.Add("reviewer_binding_bound", "review ledgers")
					boundHit = true
					return filepath.SkipAll
				}
				info, e := d.Info()
				if e != nil || !info.Mode().IsRegular() || info.Size() > 8<<20 {
					coverage.Add("unavailable_reviewer_bindings", "")
					return nil
				}
				raw, e := os.ReadFile(path)
				if e != nil {
					coverage.Add("unavailable_reviewer_bindings", "")
					return nil
				}
				ledger, e := gatestate.ParseSidecar(string(raw))
				if e == nil {
					records = append(records, ledger.ReviewerRuns...)
					for _, round := range ledger.Rounds {
						if len(round.ReviewerRuns) == 0 && !round.NoCap {
							coverage.Add("legacy_reviewer_binding_gaps", "")
						}
						records = append(records, round.ReviewerRuns...)
					}
				}
				for _, line := range strings.Split(string(raw), "\n") {
					if strings.HasPrefix(line, "<!-- sdlc-review-usage ") && strings.HasSuffix(line, " -->") {
						var rr []reviewusage.Record
						if json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(line, "<!-- sdlc-review-usage "), " -->")), &rr) == nil {
							records = append(records, rr...)
						} else {
							coverage.Add("unavailable_reviewer_bindings", "")
						}
					}
				}
				return nil
			})
			if e != nil {
				return nil, coverage, e
			}
		}
	}
	return records, coverage, nil
}
func bindUsageReviewers(sessions []transcripts.SessionEvidence, records []reviewusage.Record, now time.Time) map[string]usageReviewBinding {
	byID := map[string]transcripts.SessionEvidence{}
	for _, s := range sessions {
		byID[s.Harness+":"+s.ID] = s
	}
	out := map[string]usageReviewBinding{}
	// The same attempt is retained independently and copied into ledgers. Collapse
	// those copies before resolving mutually exclusive issue bindings.
	best := map[string]reviewusage.Record{}
	bad := map[string]bool{}
	var conflicting []reviewusage.Record
	for _, r := range records {
		if reviewusage.Validate(r) != nil {
			continue
		}
		old, ok := best[r.Intent.ID]
		if ok && (old.Intent != r.Intent || old.SessionID != "" && r.SessionID != "" && old.SessionID != r.SessionID || old.StartedAt != "" && r.StartedAt != "" && old.StartedAt != r.StartedAt || old.FinishedAt != "" && r.FinishedAt != "" && old.FinishedAt != r.FinishedAt) {
			bad[r.Intent.ID] = true
			conflicting = append(conflicting, old, r)
			continue
		}
		if !ok || r.Phase == reviewusage.Finished || old.Phase == reviewusage.Requested && r.Phase == reviewusage.Started {
			best[r.Intent.ID] = r
		}
	}
	var selected []reviewusage.Record
	for _, r := range best {
		selected = append(selected, r)
	}
	selected = append(selected, conflicting...)
	for _, r := range selected {
		i := r.Intent
		start, e := time.Parse(time.RFC3339Nano, i.RequestedAt)
		if e != nil {
			continue
		}
		end := now
		if r.FinishedAt != "" {
			if t, e := time.Parse(time.RFC3339Nano, r.FinishedAt); e == nil {
				end = t
			}
		}
		b := usageReviewBinding{Repository: i.Repository, Issue: i.IssueID, Slot: i.Slot, Round: fmt.Sprintf("%s/%s/%d", i.IssueID, i.Recipe, i.Round), Attempt: i.ID, Since: start, Until: end, Conflict: bad[i.ID]}
		var matches []string
		uncertain := map[string]bool{}
		for key, s := range byID {
			if s.Harness != i.Agent {
				continue
			}
			if r.SessionID != "" && s.ID == r.SessionID {
				matches = append(matches, key)
			} else if r.SessionID == "" && i.DisposableRoot && filepath.Clean(s.CWD) == filepath.Clean(i.Root) && !s.Started.Before(start) && s.Started.Before(end) {
				matches = append(matches, key)
			} else if r.SessionID == "" && filepath.Clean(s.CWD) == filepath.Clean(i.Root) && !s.Started.Before(start) && s.Started.Before(end) {
				matches = append(matches, key)
				uncertain[key] = true
			}
		}
		for _, key := range matches {
			candidate := b
			candidate.Conflict = candidate.Conflict || uncertain[key]
			candidate.Exact = r.SessionID != ""
			if old, ok := out[key]; ok && old.Exact && !candidate.Exact {
				continue
			} else if ok && !old.Exact && candidate.Exact {
				out[key] = candidate
				continue
			}
			if old, ok := out[key]; ok && old.Attempt != b.Attempt {
				old.Conflict = true
				out[key] = old
			} else if old.Conflict {
				candidate.Conflict = true
				out[key] = candidate
			} else {
				out[key] = candidate
			}
		}
	}
	// Explicit bound descendants inherit dispatch identity; arbitrary root overlap
	// never supplies a reviewer binding.
	var inherit func(string, map[string]bool) (usageReviewBinding, bool)
	inherit = func(key string, vis map[string]bool) (usageReviewBinding, bool) {
		if b, ok := out[key]; ok {
			return b, true
		}
		if vis[key] {
			return usageReviewBinding{}, false
		}
		vis[key] = true
		s, ok := byID[key]
		if !ok || !s.Spawned || s.ParentID == "" {
			return usageReviewBinding{}, false
		}
		b, ok := inherit(s.Harness+":"+s.ParentID, vis)
		if ok && !s.Started.Before(b.Since) && s.Started.Before(b.Until) {
			out[key] = b
		}
		return b, ok && !s.Started.Before(b.Since) && s.Started.Before(b.Until)
	}
	for key := range byID {
		inherit(key, map[string]bool{})
	}
	return out
}
