package transcripts

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
)

// SessionEvidence pins source identity and the exact bytes discovered for replay.
type SessionEvidence struct {
	Harness, ID, CWD, ParentID, File, Digest string
	Started                                  time.Time
	Spawned                                  bool
	Length                                   int64
}

// SourceWindow describes one owned slice and its continuous engagement origin.
// Roots include historical paths retained by relocation within that engagement.
type SourceWindow struct {
	Start, End, Began time.Time
	Roots             []string
}

func normalizedPath(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}

func hasRoot(roots []string, cwd string) bool {
	if cwd == "" {
		return false
	}
	for _, root := range roots {
		if normalizedPath(root) == normalizedPath(cwd) {
			return true
		}
	}
	return false
}

// EligibleSources verifies explicit lineage before admitting spawned sessions.
// Session age never excludes a persistent root author, but child dispatch must
// occur in the same continuous engagement, with its parent admitted then.
func EligibleSources(sessions []SessionEvidence, windows []SourceWindow) (map[string][]activetime.ActivityWindow, []string, error) {
	out := make(map[string][]activetime.ActivityWindow)
	byID := make(map[string]SessionEvidence)
	key := func(s SessionEvidence) string { return s.Harness + ":" + s.ID }
	for _, s := range sessions {
		if s.ID != "" {
			if prior, ok := byID[key(s)]; ok && prior.File != s.File {
				return nil, nil, fmt.Errorf("duplicate session identity %s", key(s))
			}
			byID[key(s)] = s
		}
	}
	for _, w := range windows {
		if w.Start.IsZero() || !w.End.After(w.Start) || w.Began.IsZero() || w.Began.After(w.Start) {
			return nil, nil, fmt.Errorf("invalid source ownership window")
		}
	}
	var admitted func(SessionEvidence, time.Time, time.Time, map[string]bool) (bool, error)
	admitted = func(s SessionEvidence, at, began time.Time, seen map[string]bool) (bool, error) {
		if seen[key(s)] {
			return false, fmt.Errorf("session lineage cycle at %s", key(s))
		}
		if !s.Started.IsZero() && s.Started.After(at) {
			return false, nil
		}
		pathOK := false
		for _, w := range windows {
			if w.Began.Equal(began) && !at.Before(w.Start) && at.Before(w.End) && hasRoot(w.Roots, s.CWD) {
				pathOK = true
				break
			}
		}
		if !pathOK {
			return false, nil
		}
		if !s.Spawned {
			return true, nil
		}
		if s.Started.IsZero() {
			return false, fmt.Errorf("spawned session %s lacks dispatch timestamp", key(s))
		}
		if s.Started.Before(began) {
			return false, nil
		}
		parent, ok := byID[s.Harness+":"+s.ParentID]
		if !ok || s.ParentID == "" {
			return false, fmt.Errorf("spawned session %s missing parent %s", key(s), s.ParentID)
		}
		seen[key(s)] = true
		defer delete(seen, key(s))
		return admitted(parent, s.Started, began, seen)
	}
	var warnings []string
	for _, s := range sessions {
		for _, w := range windows {
			if !hasRoot(w.Roots, s.CWD) || (!s.Started.IsZero() && !s.Started.Before(w.End)) {
				continue
			}
			at := w.Start
			if s.Started.After(at) {
				at = s.Started
			}
			ok, err := admitted(s, at, w.Began, map[string]bool{})
			if err != nil {
				return nil, warnings, err
			}
			if ok {
				out[s.File] = append(out[s.File], activetime.ActivityWindow{Start: at, End: w.End})
			}
		}
		if len(out[s.File]) == 0 {
			warnings = append(warnings, fmt.Sprintf("excluded %s session %s: no verified ownership provenance", s.Harness, s.ID))
		}
	}
	return out, warnings, nil
}
