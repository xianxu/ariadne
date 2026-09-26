// issuerecord.go — the composed issue reader every consumer uses (#252): card
// fields from the tracker, detail fields from this checkout.
package main

import (
	"context"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// loadIssueRecords composes the tracker's cards with the details in issuesDir,
// for the repository that contains issuesDir (not the process's current one). A
// directory outside any repository, or a repository whose publication remote has
// no tracker, reads its details as the whole record (pre-migration). Gates that
// authorize a write pass tracker.Fresh; read-only views pass PreferFresh and
// label a stale result.
func loadIssueRecords(ctx context.Context, issuesDir string, mode tracker.FetchMode) (tracker.Records, error) {
	abs, err := filepath.Abs(issuesDir)
	if err != nil {
		return tracker.Records{}, err
	}
	root := ""
	for dir := abs; ; dir = filepath.Dir(dir) {
		if out, err := gitx.RunGit("-C", dir, "rev-parse", "--show-toplevel"); err == nil {
			root = strings.TrimSpace(string(out))
			break
		}
		if filepath.Dir(dir) == dir {
			break // not inside a repository (or the directory does not exist yet)
		}
	}
	if root == "" {
		return tracker.LoadRecords(commandContext(ctx), nil, abs, mode)
	}
	return loadIssueRecordsAt(ctx, root, abs, mode)
}

// loadIssueRecordsAt reads the repository rooted at root (issuesDir relative to it).
func loadIssueRecordsAt(ctx context.Context, root, issuesDir string, mode tracker.FetchMode) (tracker.Records, error) {
	if !filepath.IsAbs(issuesDir) {
		issuesDir = filepath.Join(root, issuesDir)
	}
	scope, _ := commandContext(ctx).Value(recordsScopeKey{}).(*recordsScope)
	key := root + "\x00" + issuesDir
	if rs, ok := scope.get(key, mode); ok {
		return rs, nil
	}
	repo, err := recordsRepository(ctx, root)
	if err != nil {
		return tracker.Records{}, err
	}
	rs, err := tracker.LoadRecords(commandContext(ctx), repo, issuesDir, mode)
	if err == nil {
		scope.put(key, rs)
	}
	return rs, err
}

// recordsScope holds one command's composed issue views (#252, the one-fetch
// operating envelope): every consumer in a verb reads the same fetched cards.
// A stale read never satisfies a fresh request, and a card write invalidates
// the scope, so a later step sees the verb's own writes.
type recordsScope struct {
	mu    sync.Mutex
	cache map[string]tracker.Records
}

type recordsScopeKey struct{}

// withIssueRecordsScope gives a command its records scope; without one (a test
// calling a verb function directly) every read loads afresh.
func withIssueRecordsScope(ctx context.Context) context.Context {
	return context.WithValue(ctx, recordsScopeKey{}, &recordsScope{cache: map[string]tracker.Records{}})
}

func (s *recordsScope) get(key string, mode tracker.FetchMode) (tracker.Records, bool) {
	if s == nil {
		return tracker.Records{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rs, ok := s.cache[key]
	if ok && rs.Stale && mode == tracker.Fresh {
		return tracker.Records{}, false
	}
	return rs, ok
}

func (s *recordsScope) put(key string, rs tracker.Records) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = rs
}

// invalidateIssueRecords drops the command's cached views after a card write.
func invalidateIssueRecords(ctx context.Context) {
	if s, _ := commandContext(ctx).Value(recordsScopeKey{}).(*recordsScope); s != nil {
		s.mu.Lock()
		s.cache = map[string]tracker.Records{}
		s.mu.Unlock()
	}
}

// recordsRepository opens the tracker of the checkout at root.
func recordsRepository(ctx context.Context, root string) (*tracker.Repository, error) {
	return tracker.RepositoryForCheckout(commandContext(ctx), root)
}
