package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

type usageAuthority struct {
	Repository string
	Until      time.Time
	Windows    []transcripts.SourceWindow
	Label      string
}

// Usage is local and read-only: it reads the already-fetched tracker snapshot,
// never fetches or rewrites it. The snapshot pin is printed as provenance.
func resolveUsageAuthority(ctx context.Context, root string, id int, now time.Time) (usageAuthority, error) {
	out := usageAuthority{Until: now}
	repository, e := resolveUsageRepository(ctx, root)
	if e != nil {
		return out, e
	}
	out.Repository = repository
	repo, e := recordsRepository(ctx, root)
	if e != nil {
		return out, e
	}
	if repo == nil {
		legacy, err := legacyUsageAuthority(ctx, root, id, now)
		legacy.Repository = out.Repository
		return legacy, err
	}
	snap, ok, e := repo.LocalSnapshot()
	if e != nil {
		return out, e
	}
	if !ok {
		tracked, err := localTrackerDeclared(root)
		if err != nil {
			return out, err
		}
		if !tracked {
			legacy, err := legacyUsageAuthority(ctx, root, id, now)
			legacy.Repository = out.Repository
			return legacy, err
		}
		return out, fmt.Errorf("usage issue attribution needs a local tracker snapshot; run sdlc issue list to refresh it")
	}
	rec, ok := snap.Card(fmt.Sprintf("%06d", id))
	if !ok {
		return out, fmt.Errorf("usage issue #%d has no local tracker card", id)
	}
	completion, done, e := issue.CardCompletion(rec.Raw)
	if e != nil {
		return out, e
	}
	pin := snap.Ref()
	if done {
		if completion.Repository != out.Repository {
			return out, fmt.Errorf("usage completion belongs to another repository")
		}
		stamp, e := ownedActualGit(ctx, root, "show", "-s", "--format=%cI", completion.EvidenceCommit)
		if e != nil {
			return out, e
		}
		out.Until, e = time.Parse(time.RFC3339, stamp)
		if e != nil {
			return out, e
		}
		pin, e = ownedCompletionTrackerPin(ctx, root, pin, rec.Path, completion)
		if e != nil {
			return out, e
		}
	}
	out.Windows, e = ownershipSourceWindows(ctx, repo, pin, rec.Path, out.Repository, out.Until, "")
	out.Label = "ownership intervals; local tracker " + pin
	return out, e
}

// resolveUsageRepository is the shared read-only identity contract for ownership
// and reviewer diagnostics. Publication identity spans clones; repositories with
// no declared publication target retain their canonical Git-common identity.
func resolveUsageRepository(ctx context.Context, root string) (string, error) {
	identity, err := resolveWorkspace(root)
	if err != nil {
		return "", err
	}
	resting := "main"
	if identity.RestingBranch != nil {
		resting = *identity.RestingBranch
	}
	target, err := gitx.ResolvePublicationTarget(ctx, root, resting)
	if err != nil {
		return identity.RepoIdentity, nil
	}
	return target.Repository, nil
}

func legacyUsageAuthority(ctx context.Context, root string, id int, now time.Time) (usageAuthority, error) {
	head, e := ownedActualGit(ctx, root, "rev-parse", "--verify", "HEAD^{commit}")
	if e != nil {
		return usageAuthority{}, e
	}
	_, start, _, e := gitx.CommitWindowAt(ctx, root, head, strconv.Itoa(id))
	if e != nil {
		return usageAuthority{}, e
	}
	since, e := time.Parse(time.RFC3339, start)
	if e != nil {
		return usageAuthority{}, fmt.Errorf("legacy usage has no issue window: %w", e)
	}
	return usageAuthority{Until: now, Windows: []transcripts.SourceWindow{{Start: since, End: now, Began: since, Roots: []string{root}}}, Label: "legacy no-tracker repository cwd heuristic"}, nil
}
func usageFleetSlots(ctx context.Context, root string) (map[string]string, []string, error) {
	v, e := fleet.NormalizeVantage(execGitRunner{}, root)
	if e != nil {
		return nil, nil, e
	}
	inv, e := fleet.CollectInventory(ctx, v.FleetRoot, fleet.InventoryOptions{Git: execGitRunner{}, LookupIssues: func(string, string) ([]fleet.IssueRecord, error) { return nil, nil }, LookupClaims: func(string) fleet.RepoClaims { return fleet.RepoClaims{} }})
	if e != nil {
		return nil, nil, e
	}
	slots := map[string]string{}
	roots := map[string]bool{}
	for _, row := range inv.Rows {
		roots[row.RepoRoot] = true
	}
	for _, slot := range inv.Slots {
		for _, m := range slot.Members {
			key := filepath.Clean(m.Path)
			if old := slots[key]; old != "" && old != slot.Address {
				slots[key] = "unknown-slot"
			} else {
				slots[key] = slot.Address
			}
		}
	}
	var repos []string
	for r := range roots {
		if r != "" {
			repos = append(repos, r)
		}
	}
	sort.Strings(repos)
	return slots, repos, nil
}
func usageSessionSlots(sessions []transcripts.SessionEvidence, roots map[string]string) map[string]string {
	byID := map[string]transcripts.SessionEvidence{}
	bad := map[string]bool{}
	for _, s := range sessions {
		key := s.Harness + ":" + s.ID
		if bad[key] {
			continue
		}
		if old, ok := byID[key]; ok {
			if old.CWD != s.CWD || old.ParentID != s.ParentID || old.Spawned != s.Spawned || !old.Started.Equal(s.Started) {
				delete(byID, key)
				bad[key] = true
			}
		} else {
			byID[key] = s
		}
	}
	result := map[string]string{}
	visiting := map[string]bool{}
	var visit func(string) string
	visit = func(key string) string {
		if v, ok := result[key]; ok {
			return v
		}
		if visiting[key] {
			return "unknown-slot"
		}
		visiting[key] = true
		s, ok := byID[key]
		slot := "unknown-slot"
		if ok {
			if s.Spawned && s.ParentID != "" {
				slot = visit(s.Harness + ":" + s.ParentID)
			} else if v := roots[filepath.Clean(s.CWD)]; v != "" {
				slot = v
			}
		}
		visiting[key] = false
		result[key] = slot
		return slot
	}
	for k := range byID {
		visit(k)
	}
	for k := range bad {
		result[k] = "unknown-slot"
	}
	return result
}
func usageWorkerWindows(ctx context.Context, sessions []transcripts.SessionEvidence, a usageAuthority) (map[string][]activetime.ActivityWindow, error) {
	w, _, e := transcripts.EligibleSources(sessions, a.Windows)
	return w, e
}
func usageWindowContains(ws []activetime.ActivityWindow, t time.Time) bool {
	for _, w := range ws {
		if !t.Before(w.Start) && t.Before(w.End) {
			return true
		}
	}
	return false
}
func usageRoot() (string, error) {
	cwd, e := os.Getwd()
	if e != nil {
		return "", e
	}
	id, e := resolveWorkspace(cwd)
	if e != nil {
		return "", e
	}
	return id.WorktreeRoot, nil
}
