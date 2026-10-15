// Package agentid is the one record of which coding agent, model and effort
// made a commit (#367): the `Agent: <agent>/<model>/<effort>` trailer. Values
// are the names the agents' own transcripts record (claude-opus-5-5,
// gpt-6-astra, medium), so commits join `sdlc usage` (#365) without a mapping.
//
// ARCH-PURE: rendering, parsing and stamping are pure; detect.go is the thin
// seam that reads the environment and the session's transcript tail.
package agentid

import (
	"regexp"
	"strings"
)

// TrailerKey is the git trailer every agent-authored or sdlc-made commit carries.
const TrailerKey = "Agent"

// Human is the trailer value when no agent session was found: the operator
// ran sdlc, or committed, in a plain shell (#387).
const Human = "human"

// Legacy is the value #367 wrote for Human; history still carries it, so Parse
// reads it as Human.
const Legacy = "sdlc"

// Identity is the agent behind a commit. An empty Agent means no agent session
// was found (an operator in a plain shell); it renders as Human.
type Identity struct {
	Agent, Model, Effort string
}

// String renders the trailer value: agent/model/effort, "-" for an unknown part.
func (id Identity) String() string {
	if id.Agent == "" {
		return Human
	}
	return strings.Join([]string{id.Agent, dash(id.Model), dash(id.Effort)}, "/")
}

// Trailer is the full trailer line.
func (id Identity) Trailer() string { return TrailerKey + ": " + id.String() }

// Parse is String's inverse, for readers grouping commits by agent (#365).
// Human and Legacy both parse to the empty (human) Identity.
func Parse(value string) (Identity, bool) {
	value = strings.TrimSpace(value)
	if value == Human || value == Legacy {
		return Identity{}, true
	}
	parts := strings.Split(value, "/")
	if len(parts) != 3 || parts[0] == "" || parts[0] == "-" {
		return Identity{}, false
	}
	return Identity{Agent: parts[0], Model: undash(parts[1]), Effort: undash(parts[2])}, true
}

var trailerLine = regexp.MustCompile(`^[A-Za-z0-9-]+: `)

// Stamp adds the Agent trailer to a commit message, joining an existing
// trailer block (Review-Verdict:, Tracker-Operation:, Co-Authored-By: …) or
// opening one. A message that already names an agent is returned unchanged.
func Stamp(msg string, id Identity) string {
	body := strings.TrimRight(msg, "\n")
	paras := strings.Split(body, "\n\n")
	last := paras[len(paras)-1]
	inBlock := len(paras) > 1
	for _, line := range strings.Split(last, "\n") {
		if strings.HasPrefix(line, TrailerKey+": ") {
			return msg
		}
		if !trailerLine.MatchString(line) {
			inBlock = false
		}
	}
	if inBlock {
		return body + "\n" + id.Trailer() + "\n"
	}
	return body + "\n\n" + id.Trailer() + "\n"
}

// Commit is one commit as a window report sees it.
type Commit struct {
	SHA, Subject string
	Agents       []string // values of its Agent trailers
	Merge        bool
}

// MissingAgent is the non-merge commits that carry no Agent trailer.
func MissingAgent(commits []Commit) []Commit {
	var out []Commit
	for _, c := range commits {
		if !c.Merge && len(c.Agents) == 0 {
			out = append(out, c)
		}
	}
	return out
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func undash(s string) string {
	if s == "-" {
		return ""
	}
	return s
}

// LogFormat is the `git log --format` whose output ParseLog reads: hash,
// parent hashes, subject and the Agent trailer values, one record per commit.
const LogFormat = "%H%x1f%P%x1f%s%x1f%(trailers:key=" + TrailerKey + ",valueonly,separator=%x1e)%x1d"

// ParseLog reads `git log --format=LogFormat` output into commits.
func ParseLog(out string) []Commit {
	var commits []Commit
	for _, rec := range strings.Split(out, "\x1d") {
		f := strings.Split(strings.TrimLeft(rec, "\n"), "\x1f")
		if len(f) != 4 {
			continue
		}
		c := Commit{SHA: f[0], Subject: f[2], Merge: len(strings.Fields(f[1])) > 1}
		for _, v := range strings.Split(f[3], "\x1e") {
			if v = strings.TrimSpace(v); v != "" {
				c.Agents = append(c.Agents, v)
			}
		}
		commits = append(commits, c)
	}
	return commits
}
