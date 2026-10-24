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

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

// localChecksPath is the declaration, read from the base tree (never the PR head,
// so a PR cannot declare its own substitute).
const localChecksPath = ".sdlc/local-checks.json"

const (
	localCheckDefaultTimeout = 30
	localCheckMaxTimeout     = 120
	localCheckMaxWeight      = 8 // test-lock.py clamps to the machine's slot count
	// suiteCheckName is the check every landing requires (#332): declared, or
	// the default make test.
	suiteCheckName = "suite"
	// suiteDefaultTimeout bounds the default suite: ariadne's sharded run is
	// about 6-13 minutes plus queueing behind the machine lock.
	suiteDefaultTimeout = 60
	// affectedCheckName is the affected suite's check (#473), declared by
	// affectedConfigPath rather than in local-checks.json, so an sdlc built
	// before it never reads (and refuses) its fields.
	affectedCheckName  = "affected"
	affectedConfigPath = ".sdlc/affected.json"
)

// localCheckDecl is the parsed declaration.
type localCheckDecl struct {
	Version int                        `json:"version"`
	Checks  []localCheck               `json:"checks"`
	Extra   map[string]json.RawMessage `json:"-"`
}

// localCheck is one declared local check: an argv run at the repository root of
// the head under test. Substitutes names the CI check labels (checkLabel form,
// "workflow / job") it stands in for when that job never got a runner. Required
// makes the recorded run mandatory whatever CI reports.
type localCheck struct {
	// DeclarationVersion is reader-owned capability metadata, never declared.
	DeclarationVersion int                        `json:"-"`
	FlakeRunners       flakes.UnitRunners         `json:"flake_runners,omitempty"`
	RunScope           string                     `json:"scope,omitempty"`
	Resource           string                     `json:"resource,omitempty"`
	PathsFile          string                     `json:"paths_file,omitempty"`
	Extra              map[string]json.RawMessage `json:"-"`
	Name               string                     `json:"name"`
	Command            []string                   `json:"command"`
	Substitutes        []string                   `json:"substitutes,omitempty"`
	Required           bool                       `json:"required,omitempty"`
	TimeoutMinutes     int                        `json:"timeout_minutes,omitempty"`
	// Weight is how many machine-wide suite slots the run takes at once (#346,
	// #332): ariadne's sharded suite takes 2 and runs alone. Default 1.
	Weight int `json:"weight,omitempty"`
	// KnownFlakes is the suite's policy for a failure only registered flakes
	// explain (#394 D2): "rerun" (the default) reruns just those tests,
	// "accept" records the pass with the issue link, "off" blocks. Only the
	// suite honours it; the field is the operator's (local-gates).
	KnownFlakes string `json:"known_flakes,omitempty"`
	// Scope says what a built check covers, for its status (#473: the
	// affected selection); never declared.
	Scope string `json:"-"`
	// Rerun is what a known-flake child runs besides the flaky tests (#586:
	// the affected selection's groups); nil for the suite. Never declared.
	Rerun *affectedRerun `json:"-"`
}

// knownFlakePolicy is the check's parsed policy; parseLocalChecks has
// already refused an invalid value.
func (c localCheck) knownFlakePolicy() flakes.Policy {
	p, _ := flakes.ParsePolicy(c.KnownFlakes)
	return p
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
	if isV2LocalChecks(data) {
		return parseLocalChecksV2(data)
	}
	if isV3LocalChecks(data) {
		return parseLocalChecksV3(data)
	}
	var legacy map[string]json.RawMessage
	if err := json.Unmarshal(data, &legacy); err == nil {
		var checks []map[string]json.RawMessage
		if json.Unmarshal(legacy["checks"], &checks) == nil {
			for _, c := range checks {
				for _, k := range []string{"scope", "resource", "paths_file"} {
					for key := range c {
						if strings.EqualFold(key, k) {
							return bad("unknown field %q in version 1", key)
						}
					}
				}
			}
		}
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
	return validateLocalChecks(d)
}

func validateLocalChecks(d localCheckDecl) (localCheckDecl, error) {
	bad := func(format string, a ...any) (localCheckDecl, error) {
		return localCheckDecl{}, fmt.Errorf(localChecksPath+": "+format, a...)
	}
	names, labels := map[string]bool{}, map[string]string{}
	for i := range d.Checks {
		c := &d.Checks[i]
		c.DeclarationVersion = d.Version
		if len(c.FlakeRunners) != 0 && (d.Version != 3 || c.Name != suiteCheckName) {
			return bad("flake_runners requires a version 3 suite declaration")
		}
		if d.Version == 1 && (c.RunScope != "" || c.Resource != "" || c.PathsFile != "") {
			return bad("version 1 does not support landing fields")
		}
		if !localCheckName.MatchString(c.Name) {
			return bad("check name %q must match %s", c.Name, localCheckName)
		}
		if names[c.Name] {
			return bad("check %q declared twice", c.Name)
		}
		if c.Name == suiteRerunCheck {
			return bad("check name %q is reserved for known-flake reruns (#394)", c.Name)
		}
		if c.Name == affectedCheckName {
			return bad("check name %q is reserved: the affected suite is declared in %s (#473)", c.Name, affectedConfigPath)
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
		case c.Weight == 0:
			c.Weight = 1
		case c.Weight < 0 || c.Weight > localCheckMaxWeight:
			return bad("check %q: weight %d outside 1..%d", c.Name, c.Weight, localCheckMaxWeight)
		}
		if _, err := flakes.ParsePolicy(c.KnownFlakes); err != nil {
			return bad("check %q: %v", c.Name, err)
		}
		if c.KnownFlakes != "" && c.Name != suiteCheckName {
			return bad("check %q: known_flakes applies to the %q check only", c.Name, suiteCheckName)
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

// defaultSuiteCheck is the suite a repository runs when it declares none: its
// make test, at the weight its Makefile gives (#332).
func defaultSuiteCheck(weight int) localCheck {
	if weight < 1 {
		weight = 1
	}
	return localCheck{Name: suiteCheckName, Command: []string{"make", "test"}, Required: true,
		TimeoutMinutes: suiteDefaultTimeout, Weight: min(weight, localCheckMaxWeight)}
}

// effectiveLocalChecks is the declaration plus the default suite when the
// declaration has none and the repository has a make test target (#332). A
// declared suite overrides execution settings, but cannot waive enforcement.
func effectiveLocalChecks(d localCheckDecl, hasMakeTest bool, weight int) localCheckDecl {
	if _, declared := d.check(suiteCheckName); declared {
		out := d
		out.Checks = append([]localCheck(nil), d.Checks...)
		for i := range out.Checks {
			if out.Checks[i].Name == suiteCheckName {
				out.Checks[i].Required = true
			}
		}
		return out
	}
	if !hasMakeTest {
		return d
	}
	out := d
	if out.Version == 0 {
		out.Version = 1
	}
	out.Checks = append(append([]localCheck(nil), d.Checks...), defaultSuiteCheck(weight))
	return out
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
func localStatusContext(name string) string { return localStatusPrefix + name }

const localStatusPrefix = "sdlc-local/"

// localStatusContextValid reports whether s is a context localStatusContext builds.
func localStatusContextValid(s string) bool {
	name, ok := strings.CutPrefix(s, localStatusPrefix)
	return ok && localCheckName.MatchString(name)
}

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
	ResumeParent, ResumeChild string
	Name                      string
	SHA                       string
	Exit                      int // -1: timed out or did not start
	TimedOut                  bool
	Duration                  time.Duration
	LogSHA256                 string
	LogPath                   string
	Host                      string
	Tail                      []string
	// Outcome is the known-flake gate's decision on a failed suite (#394);
	// "" when no gate ran. A known-flake landing passes despite its exit.
	Outcome     flakes.Kind
	KnownFlakes []flakes.Entry
	Scope       string // what the run covered (#473), leading its status
	// The managed execution's identity and timing (#561), for train
	// measurements: the saved attempt (when one was written), the time
	// queued for a suite slot, and when the command finished (zero when it
	// never reported).
	AttemptID string
	QueueWait time.Duration
	Finished  time.Time
	// Reused marks a saved pass posted again rather than run (#431).
	Reused bool
}

// knownFlakeOutcome is the record's gate decision, for revalidation.
func (r localRunRecord) knownFlakeOutcome() flakes.Outcome {
	return flakes.Outcome{Kind: r.Outcome, Known: r.KnownFlakes}
}

func (r localRunRecord) passed() bool {
	if r.Outcome != "" && r.Outcome != flakes.Passed {
		return r.Outcome == flakes.KnownFlakePassed || r.Outcome == flakes.KnownFlakeAccept
	}
	return r.Exit == 0 && !r.TimedOut
}

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
	switch issues := "#" + strings.Join(r.knownFlakeOutcome().Issues(), ",#"); r.Outcome {
	case flakes.KnownFlakePassed:
		s = "pass · known flake " + issues + " rerun · " + s
	case flakes.KnownFlakeAccept:
		s = "pass · known flake " + issues + " accepted · " + s
	case "", flakes.Passed:
	default:
		s = string(r.Outcome) + " · " + s
	}
	if r.Scope != "" {
		s = r.Scope + " · " + s
	}
	if r.ResumeParent != "" {
		s = fmt.Sprintf("resume %.12s/%.12s · %s", r.ResumeParent, r.ResumeChild, s)
	}
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
