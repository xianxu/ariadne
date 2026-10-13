// localchecks.go — recorded local checks (#345): the declaration and the pure
// rules around it.
//
// Local runs are first-class verification evidence; server CI guards that the
// gate configuration did not change. A repository declares its local checks in
// .sdlc/local-checks.json (under the local-gates guard, so an edit needs the
// operator's label). `sdlc local-check run` runs one on an exact head and records
// the outcome as a commit status (localStatusContext). The CI gate accepts that
// status in place of a CI job that never got a runner, and requires it for a
// check declared `required`. A CI job that ran and failed still blocks.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// localChecksPath is the declaration, read from the base tree (never the PR head,
// so a PR cannot declare its own substitute).
const localChecksPath = ".sdlc/local-checks.json"

const (
	localCheckDefaultTimeout = 30
	localCheckMaxTimeout     = 120
)

// localCheckDecl is the parsed declaration.
type localCheckDecl struct {
	Version int          `json:"version"`
	Checks  []localCheck `json:"checks"`
}

// localCheck is one declared local check: an argv run at the repository root of
// the head under test. Substitutes names the CI check labels (checkLabel form,
// "workflow / job") it stands in for when that job never got a runner. Required
// makes the recorded run mandatory whatever CI reports.
type localCheck struct {
	Name           string   `json:"name"`
	Command        []string `json:"command"`
	Substitutes    []string `json:"substitutes,omitempty"`
	Required       bool     `json:"required,omitempty"`
	TimeoutMinutes int      `json:"timeout_minutes,omitempty"`
}

// localCheckName becomes a status context and a path segment.
var localCheckName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,40}$`)

// parseLocalChecks parses the declaration strictly: unknown fields, a version
// other than 1, and malformed entries are errors, never ignored. Empty input
// (no file) is an empty declaration.
func parseLocalChecks(data []byte) (localCheckDecl, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return localCheckDecl{}, nil
	}
	bad := func(format string, a ...any) (localCheckDecl, error) {
		return localCheckDecl{}, fmt.Errorf("%s: "+format, append([]any{localChecksPath}, a...)...)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var d localCheckDecl
	if err := dec.Decode(&d); err != nil {
		return bad("%v", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return bad("trailing data after the declaration")
	}
	if d.Version != 1 {
		return bad("version %d, want 1", d.Version)
	}
	names, labels := map[string]bool{}, map[string]string{}
	for i := range d.Checks {
		c := &d.Checks[i]
		if !localCheckName.MatchString(c.Name) {
			return bad("check name %q must match %s", c.Name, localCheckName)
		}
		if names[c.Name] {
			return bad("check %q declared twice", c.Name)
		}
		names[c.Name] = true
		if len(c.Command) == 0 {
			return bad("check %q: empty command", c.Name)
		}
		for _, a := range c.Command {
			if a == "" {
				return bad("check %q: empty command argument", c.Name)
			}
		}
		for _, l := range c.Substitutes {
			if strings.TrimSpace(l) == "" {
				return bad("check %q: empty substitutes label", c.Name)
			}
			if prev, dup := labels[l]; dup {
				return bad("CI check %q substituted by both %q and %q", l, prev, c.Name)
			}
			labels[l] = c.Name
		}
		switch {
		case c.TimeoutMinutes == 0:
			c.TimeoutMinutes = localCheckDefaultTimeout
		case c.TimeoutMinutes < 0 || c.TimeoutMinutes > localCheckMaxTimeout:
			return bad("check %q: timeout_minutes %d outside 1..%d", c.Name, c.TimeoutMinutes, localCheckMaxTimeout)
		}
	}
	return d, nil
}

// substituteFor finds the local check declared to stand in for a CI check label.
func (d localCheckDecl) substituteFor(label string) (localCheck, bool) {
	for _, c := range d.Checks {
		for _, l := range c.Substitutes {
			if l == label {
				return c, true
			}
		}
	}
	return localCheck{}, false
}

func (d localCheckDecl) check(name string) (localCheck, bool) {
	for _, c := range d.Checks {
		if c.Name == name {
			return c, true
		}
	}
	return localCheck{}, false
}

// localStatusContext is the commit-status context a recorded run posts, and the
// check name the gate looks for. The one place the string is built.
func localStatusContext(name string) string { return "sdlc-local/" + name }

// jobRun is the part of GitHub's Actions job object the gate reads.
type jobRun struct {
	Status     string            `json:"status"`
	Conclusion *string           `json:"conclusion"`
	RunnerName *string           `json:"runner_name"`
	Steps      []json.RawMessage `json:"steps"`
}

// runnerless is true only for a job GitHub cancelled before any runner took it:
// completed, cancelled, no runner, zero steps (measured on pair job 115490631095).
// It reads structure, never annotation text, and a body missing the runner or
// steps fields is not runnerless: unread is never substitutable.
func (j jobRun) runnerless() bool {
	return j.Status == "completed" && j.Conclusion != nil && *j.Conclusion == "cancelled" &&
		j.Steps != nil && len(j.Steps) == 0 &&
		(j.RunnerName == nil || *j.RunnerName == "")
}

var jobLink = regexp.MustCompile(`^https://github\.com/[^/]+/[^/]+/actions/runs/[0-9]+/job/([0-9]+)$`)

// jobIDFromLink reads the Actions job id from a check's link. A commit status or
// a third-party check has no such link, so it can never count as runnerless.
func jobIDFromLink(link string) (int64, bool) {
	m := jobLink.FindStringSubmatch(link)
	if m == nil {
		return 0, false
	}
	id, err := strconv.ParseInt(m[1], 10, 64)
	return id, err == nil
}

// localRunRecord is the outcome of one local run, as recorded.
type localRunRecord struct {
	Name      string
	SHA       string
	Exit      int // -1: timed out or did not start
	TimedOut  bool
	Duration  time.Duration
	LogSHA256 string
	LogPath   string
	Host      string
	Tail      []string
}

func (r localRunRecord) passed() bool { return r.Exit == 0 && !r.TimedOut }

func (r localRunRecord) state() string {
	if r.passed() {
		return "success"
	}
	return "failure"
}

// statusDescription fits GitHub's 140-character status description.
func (r localRunRecord) statusDescription() string {
	outcome := "exit " + strconv.Itoa(r.Exit)
	if r.TimedOut {
		outcome = "timed out"
	}
	hash := r.LogSHA256
	if len(hash) > 12 {
		hash = hash[:12]
	}
	s := fmt.Sprintf("%s in %s · log sha256:%s · %s", outcome, r.Duration.Round(time.Second), hash, r.Host)
	if len(s) > 140 {
		s = s[:140]
	}
	return s
}

const (
	failureTailLines = 60
	failureTailBytes = 8 << 10
)

// failureComment renders the PR comment a failed run posts: the outcome and a
// bounded tail of the log (the full log stays local).
func (r localRunRecord) failureComment() string {
	tail := r.Tail
	if len(tail) > failureTailLines {
		tail = tail[len(tail)-failureTailLines:]
	}
	body := strings.Join(tail, "\n")
	if len(body) > failureTailBytes {
		body = body[len(body)-failureTailBytes:]
	}
	return fmt.Sprintf("**%s failed** on `%s` (%s)\n\nFull log on %s: `%s`\n\n```\n%s\n```\n",
		localStatusContext(r.Name), r.SHA, r.statusDescription(), r.Host, r.LogPath, strings.ReplaceAll(body, "```", "'''"))
}
