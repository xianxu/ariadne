package qodertranscript

import (
	"path/filepath"
	"time"
)

// Metadata is complete discovery evidence. StartedAt and LastActivity describe
// real user/assistant events; bookkeeping replay cannot extend session activity.
type Metadata struct {
	SessionID    string
	CWD          string
	StartedAt    time.Time
	LastActivity time.Time
}

type config struct {
	model, effort string
	ambiguous     bool
}

// SessionState has a useful zero value. It owns immutable identity/workspace
// agreement and timestamp-indexed configuration evidence. It retains no records.
type SessionState struct {
	sessionID     string
	identityErr   error
	cwd           string
	workspaceErr  error
	started, last time.Time
	configs       map[time.Time]config
	configErr     error
}

// Step incorporates one parsed record. Contradictory identities and workspaces
// remain poisoned even after later valid replays. A workspace conflict does not
// erase the session identity needed to account unknown-slot response charges.
func (s *SessionState) Step(r Record) error {
	switch r.Type {
	case "workspace-directories", "runtime-config", "user", "assistant":
	default:
		return nil
	}
	if s.identityErr != nil {
		return s.identityErr
	}
	if r.SessionID == "" {
		s.identityErr = ErrIncomplete
		return s.identityErr
	}
	if len(r.SessionID) > MaxStringBytes {
		s.identityErr = ErrLimit
		return s.identityErr
	}
	if s.sessionID != "" && s.sessionID != r.SessionID {
		s.identityErr = ErrConflict
		return s.identityErr
	}
	s.sessionID = r.SessionID
	switch r.Type {
	case "workspace-directories":
		if s.workspaceErr != nil {
			return s.workspaceErr
		}
		var root string
		for _, v := range r.Directories {
			if v == "" || !filepath.IsAbs(v) {
				s.workspaceErr = ErrConflict
				return s.workspaceErr
			}
			v = filepath.Clean(v)
			if root != "" && root != v {
				s.workspaceErr = ErrConflict
				return s.workspaceErr
			}
			root = v
		}
		if root != "" {
			if s.cwd != "" && s.cwd != root {
				s.workspaceErr = ErrConflict
				return s.workspaceErr
			}
			s.cwd = root
		}
	case "runtime-config":
		if s.configErr != nil {
			return s.configErr
		}
		if r.Timestamp.IsZero() {
			s.configErr = ErrInvalidRecord
			return s.configErr
		}
		key := r.Timestamp.UTC()
		v := config{model: r.Model}
		if r.ReasoningEffort != nil {
			v.effort = *r.ReasoningEffort
		}
		if old, ok := s.configs[key]; ok {
			if old.ambiguous || old.model != v.model || old.effort != v.effort {
				old.ambiguous = true
				s.configs[key] = old
				return ErrConflict
			}
			return nil
		}
		if len(s.configs) >= MaxConfigEpochs {
			s.configErr = ErrLimit
			return s.configErr
		}
		if s.configs == nil {
			s.configs = make(map[time.Time]config)
		}
		s.configs[key] = v
	case "user", "assistant":
		if r.Timestamp.IsZero() {
			return ErrInvalidRecord
		}
		if s.started.IsZero() || r.Timestamp.Before(s.started) {
			s.started = r.Timestamp
		}
		if r.Timestamp.After(s.last) {
			s.last = r.Timestamp
		}
	}
	return nil
}

// Identity returns a known, uncontested session identity even if workspace or
// activity metadata is incomplete. Consumers must still check Metadata before
// assigning workspace ownership.
func (s *SessionState) Identity() (string, error) {
	if s.identityErr != nil {
		return "", s.identityErr
	}
	if s.sessionID == "" {
		return "", ErrIncomplete
	}
	return s.sessionID, nil
}

func (s *SessionState) Metadata() (Metadata, error) {
	id, e := s.Identity()
	if e != nil {
		return Metadata{}, e
	}
	if s.workspaceErr != nil {
		return Metadata{}, s.workspaceErr
	}
	m := Metadata{SessionID: id, CWD: s.cwd, StartedAt: s.started, LastActivity: s.last}
	if m.CWD == "" || m.StartedAt.IsZero() {
		return m, ErrIncomplete
	}
	return m, nil
}

// ConfigAt selects the newest observation at or before at, independent of
// arrival order. An ambiguous selected epoch cannot be repaired by replay.
// No applicable config yields empty model/effort with ErrIncomplete.
func (s *SessionState) ConfigAt(at time.Time) (string, string, error) {
	if _, e := s.Identity(); e != nil {
		return "", "", e
	}
	if s.configErr != nil {
		return "", "", s.configErr
	}
	var latest time.Time
	var found config
	for ts, c := range s.configs {
		if !ts.After(at) && (latest.IsZero() || ts.After(latest)) {
			latest = ts
			found = c
		}
	}
	if latest.IsZero() {
		return "", "", ErrIncomplete
	}
	if found.ambiguous {
		return "", "", ErrConflict
	}
	return found.model, found.effort, nil
}
