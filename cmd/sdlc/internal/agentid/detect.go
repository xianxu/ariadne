package agentid

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// tailBytes bounds the transcript read: session files grow past 100 MB, and
// the latest model and effort sit in the last few records (ARCH-CONSTRAINTS).
const tailBytes = 1 << 20

// Detector finds the invoking session's identity. Home and Getenv are the IO
// seam; tests point them at a temp home and a fake environment.
type Detector struct {
	Home   string
	Getenv func(string) string
	CWD    string
	Now    func() time.Time
}

// OverrideEnv names the identity outright (agent/model/effort or "sdlc"), for
// scripts that know it and for tests that need a fixed trailer.
const OverrideEnv = "SDLC_AGENT_ID"

// Detect returns the session's identity: the agent from the same environment
// signals the judge uses, model and effort from the session's own transcript.
func (d Detector) Detect() Identity {
	if id, ok := Parse(d.Getenv(OverrideEnv)); ok {
		return id
	}
	// Qoder also exports Claude plugin variables. Its native invocation marker
	// takes precedence; Pair pane context alone never establishes invocation.
	if d.Getenv("QODER_CLI") != "" {
		return d.qoderIdentity()
	}
	env := EnvFrom(d.Getenv)
	if !env.InSession() {
		return Identity{}
	}
	id := Identity{Agent: env.Main()}
	switch id.Agent {
	case Claude:
		if sid := d.Getenv("CLAUDE_CODE_SESSION_ID"); sid != "" {
			id.Model, id.Effort = claudeFromTail(d.tail(filepath.Join(ClaudeProjects(d.Home), "*", sid+".jsonl")))
		}
		if id.Effort == "" {
			id.Effort = d.Getenv("CLAUDE_EFFORT")
		}
	case Codex:
		if tid := env.CodexThreadID; tid != "" {
			id.Model, id.Effort = codexFromTail(d.tail(filepath.Join(CodexSessions(d.Home), "*", "*", "*", "rollout-*-"+tid+".jsonl")))
		}
	}
	return id
}

// tail reads the last tailBytes of the first file matching pattern.
func (d Detector) tail(pattern string) []byte {
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 0 {
		return nil
	}
	f, err := os.Open(matches[0])
	if err != nil {
		return nil
	}
	defer f.Close()
	if info, err := f.Stat(); err == nil && info.Size() > tailBytes {
		if _, err := f.Seek(-tailBytes, io.SeekEnd); err != nil {
			return nil
		}
	}
	b, _ := io.ReadAll(io.LimitReader(f, tailBytes))
	return b
}

var current = sync.OnceValue(func() Identity {
	home, _ := os.UserHomeDir()
	return Detector{Home: home, Getenv: os.Getenv}.Detect()
})

// Current is this process's identity, detected once.
func Current() Identity { return current() }

// StampCurrent stamps a commit message with this process's identity; every
// commit sdlc makes passes its message through it.
func StampCurrent(msg string) string { return Stamp(msg, Current()) }
