package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

type usageReviewerScope struct {
	Repository string
	Issue      int
}

// loadUsageReviewers loads role-exclusion identities across all issues. scope
// qualifies legacy-gap diagnostics by repository and issue only; included totals
// are filtered after binding. Its zero value selects fleet diagnostics.
func loadUsageReviewers(ctx context.Context, repos []string, scope usageReviewerScope) ([]reviewusage.Record, usage.Coverage, error) {
	var records []reviewusage.Record
	var coverage usage.Coverage
	if scope.Issue > 0 && scope.Repository == "" {
		return nil, coverage, fmt.Errorf("reviewer diagnostic scope requires repository identity")
	}
	seen := map[string]bool{}
	files := 0
	boundHit := false
	for _, root := range repos {
		var diagnosticRepository string
		diagnosticRepositoryResolved := false
		if ctx.Err() != nil {
			return nil, coverage, ctx.Err()
		}
		store, e := reviewusage.NewStore(root)
		if e == nil {
			entries, e := store.AccountingEvidence(ctx)
			if e != nil {
				coverage.Add("unavailable_reviewer_bindings", "")
			}
			for _, entry := range entries {
				records = append(records, entry.Record)
				if entry.Record.Phase != reviewusage.Finished {
					coverage.Add("unfinished_reviewer_bindings", "")
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
				retained, evidenceCoverage := decodeReviewerAccountingArtifact(string(raw))
				if strings.HasSuffix(name, "-gate.md") && !strings.HasPrefix(strings.TrimSpace(string(raw)), "---") {
					evidenceCoverage.Add("unavailable_reviewer_bindings", "reviewer ledger missing frontmatter")
				}
				records = append(records, retained...)
				// Exclusive role selection needs every issue's retained identity,
				// including reviewers executing inside an owned worker checkout.
				// Legacy rounds without identity only diagnose the selected issue
				// (or the whole fleet). Unreadable evidence remains visible for
				// every issue because it may hide a required role exclusion.
				if scope.Issue > 0 {
					selected := strings.HasPrefix(name, fmt.Sprintf("%06d-", scope.Issue))
					if selected && evidenceCoverage.Counts["legacy_reviewer_binding_gaps"] > 0 {
						if !diagnosticRepositoryResolved {
							var identityErr error
							diagnosticRepository, identityErr = resolveUsageRepository(ctx, root)
							diagnosticRepositoryResolved = true
							if identityErr != nil {
								coverage.Add("unavailable_reviewer_bindings", "reviewer diagnostic repository unavailable")
							}
						}
						selected = diagnosticRepository == scope.Repository
					}
					if !selected {
						delete(evidenceCoverage.Counts, "legacy_reviewer_binding_gaps")
						delete(evidenceCoverage.Examples, "legacy_reviewer_binding_gaps")
					}
				}
				coverage.Merge(evidenceCoverage)
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
			// Invalid accounting is not worker evidence. Retain any usable
			// session/root identity as a quarantined binding below.
			conflicting = append(conflicting, r)
			bad[r.Intent.ID] = true
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
		start, startErr := time.Parse(time.RFC3339Nano, i.RequestedAt)
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
			} else if r.SessionID == "" && startErr == nil && filepath.IsAbs(i.Root) && i.DisposableRoot && filepath.Clean(s.CWD) == filepath.Clean(i.Root) && !s.Started.Before(start) && s.Started.Before(end) {
				matches = append(matches, key)
			} else if r.SessionID == "" && startErr == nil && filepath.IsAbs(i.Root) && filepath.Clean(s.CWD) == filepath.Clean(i.Root) && !s.Started.Before(start) && s.Started.Before(end) {
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
