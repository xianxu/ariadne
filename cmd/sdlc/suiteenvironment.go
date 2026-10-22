package main

import (
	"fmt"
	"iter"
	"regexp"
	"strings"
)

// suiteEnvironmentError is execution evidence, never a failed code-suite result.
// Its separate log cannot enter flake attribution or prevent a same-head repair.
type suiteEnvironmentError struct {
	Check, Head, LogPath string
	Cause                error
	Diagnosis            suiteStartupDiagnosis
}

// suiteStartupDiagnosis is a confirmed startup failure, never test evidence.
type suiteStartupDiagnosis struct {
	Resource string
	Path     string
	Reason   string
}

func (e *suiteEnvironmentError) Error() string {
	detail := ""
	if e.Cause != nil {
		detail = "; " + e.Cause.Error()
	}
	check := e.Check
	if check == suiteRerunCheck {
		check = suiteCheckName
	}
	diagnosis := e.Diagnosis
	if diagnosis.Resource == "" {
		diagnosis.Resource = "Go work directory (GOTMPDIR/TMPDIR)"
	}
	where := ""
	if diagnosis.Path != "" {
		where = fmt.Sprintf(" at %q", diagnosis.Path)
	}
	if diagnosis.Reason != "" {
		where += ": " + diagnosis.Reason
	}
	revision := ""
	if e.Head != "" {
		revision = " --rev " + e.Head
	}
	return fmt.Sprintf("environment failure: could-not-run: %s%s; repair the environment and rerun sdlc local-check run --name %s%s; log %s%s", diagnosis.Resource, where, check, revision, e.LogPath, detail)
}

func (e *suiteEnvironmentError) Unwrap() error { return e.Cause }

var workDirStartup = regexp.MustCompile(`^go: creating work dir: mkdir [^\x00-\x1f\x7f]+: (permission denied|operation not permitted|no space left on device|disk quota exceeded|read-only file system|no such file or directory|not a directory|file exists|too many open files)$`)
var startupMakeExit = regexp.MustCompile(`^make(?:\[\d+\])?: \*\*\* \[test\] Error [1-9]\d*$`)
var storageStartup = regexp.MustCompile(`^(?:pattern [^\x00-\x1f\x7f]+: |go: (?:copying [^\x00-\x1f\x7f]+: |failed to initialize build cache at [^\x00-\x1f\x7f]+: |writing stat cache: ))(?:write|open|mkdir|create) ([^\x00-\x1f\x7f]+): (no space left on device|disk quota exceeded)$`)
var missingGOROOT = regexp.MustCompile(`^go: cannot find GOROOT directory: ([^\x00-\x1f\x7f]+)$`)

// The compile-before-report traceback emitted by scripts/test-shard.py. Match
// the wrapper itself, not arbitrary indented Python (which could quote tests).
// Unknown versions or incomplete wrappers remain ordinary failures.
var startupTrace = []*regexp.Regexp{
	regexp.MustCompile(`^  File "[^"\r\n]+/scripts/test-shard\.py", line \d+, in <module>$`),
	regexp.MustCompile(`^    sys\.exit\(main\(\)\)$`),
	regexp.MustCompile(`^ +[~^]+$`),
	regexp.MustCompile(`^  File "[^"\r\n]+/scripts/test-shard\.py", line \d+, in main$`),
	regexp.MustCompile(`^    subprocess\.run\(\["go", "test", "-c", "-o", str\(binary\), "\."\], cwd=main_dir, check=True\)$`),
	regexp.MustCompile(`^ +[~^]+$`),
	regexp.MustCompile(`^  File "[^"\r\n]+/subprocess\.py", line \d+, in run$`),
	regexp.MustCompile(`^    raise CalledProcessError\(retcode, process\.args,$`),
	regexp.MustCompile(`^ +output=stdout, stderr=stderr\)$`),
	regexp.MustCompile(`^subprocess\.CalledProcessError: Command '\['go', 'test', '-c', '-o', '[^'\r\n]+', '\.'\]' returned non-zero exit status [1-9]\d*\.$`),
}

// The go-list-before-compile wrapper recorded during #508's cache outage.
var startupListTrace = []*regexp.Regexp{
	startupTrace[0], startupTrace[1], startupTrace[2], startupTrace[3],
	regexp.MustCompile(`^    others = \[p for p in git_list\(root\) if p != MAIN_PKG\]$`),
	startupTrace[2],
	regexp.MustCompile(`^  File "[^"\r\n]+/scripts/test-shard\.py", line \d+, in git_list$`),
	regexp.MustCompile(`^    out = subprocess\.check_output\(\["go", "list", "\./\.\.\."\], cwd=root, text=True\)$`),
	regexp.MustCompile(`^  File "[^"\r\n]+/subprocess\.py", line \d+, in check_output$`),
	regexp.MustCompile(`^    return run\(\*popenargs, stdout=PIPE, timeout=timeout, check=True,$`),
	startupTrace[2],
	regexp.MustCompile(`^ +\*\*kwargs\)\.stdout$`),
	startupTrace[2], startupTrace[6], startupTrace[7], startupTrace[8],
	regexp.MustCompile(`^subprocess\.CalledProcessError: Command '\['go', 'list', '\./\.\.\.'\]' returned non-zero exit status [1-9]\d*\.$`),
}

func suiteEnvironmentLog(log string) bool {
	_, ok := diagnoseSuiteStartup(log)
	return ok
}

func diagnoseSuiteStartup(log string) (suiteStartupDiagnosis, bool) {
	return diagnoseSuiteStartupLines(strings.Lines(log))
}

func diagnoseSuiteStartupLine(line string) (suiteStartupDiagnosis, bool) {
	if m := workDirStartup.FindStringSubmatch(line); m != nil {
		path := strings.TrimSuffix(strings.TrimPrefix(line, "go: creating work dir: mkdir "), ": "+m[1])
		return suiteStartupDiagnosis{Resource: "Go work directory (GOTMPDIR/TMPDIR)", Path: path, Reason: m[1]}, true
	}
	if m := storageStartup.FindStringSubmatch(line); m != nil {
		return suiteStartupDiagnosis{Resource: "Go cache/build storage", Path: m[1], Reason: m[2]}, true
	}
	if m := missingGOROOT.FindStringSubmatch(line); m != nil {
		return suiteStartupDiagnosis{Resource: "Go toolchain", Path: m[1], Reason: "GOROOT is missing"}, true
	}
	return suiteStartupDiagnosis{}, false
}

// One pass, bounded lines and fixed grammar work per line. Taking a sequence
// lets the N/2N test count actual input visits without timing a shared machine.
func suiteEnvironmentLines(lines iter.Seq[string]) bool {
	_, ok := diagnoseSuiteStartupLines(lines)
	return ok
}

func diagnoseSuiteStartupLines(lines iter.Seq[string]) (suiteStartupDiagnosis, bool) {
	var diagnosis suiteStartupDiagnosis
	seen, tracing := false, false
	traces := [2][]*regexp.Regexp{startupTrace, startupListTrace}
	var positions [2]int
	for raw := range lines {
		if !strings.HasSuffix(raw, "\n") || len(raw) > suiteLogLineLimit+1 {
			return suiteStartupDiagnosis{}, false
		}
		line := strings.TrimSuffix(raw, "\n")
		if tracing {
			alive, complete := false, false
			for i, pos := range positions {
				if pos < 0 {
					continue
				}
				if !traces[i][pos].MatchString(line) {
					positions[i] = -1
					continue
				}
				positions[i]++
				alive = true
				complete = complete || positions[i] == len(traces[i])
			}
			if !alive {
				return suiteStartupDiagnosis{}, false
			}
			tracing = !complete
			continue
		}
		if d, ok := diagnoseSuiteStartupLine(line); ok {
			if !seen {
				diagnosis = d
			}
			seen = true
			continue
		}
		switch {
		case seen && line == "Traceback (most recent call last):":
			tracing = true
			positions = [2]int{}
		case seen && startupMakeExit.MatchString(line):
		default:
			return suiteStartupDiagnosis{}, false
		}
	}
	return diagnosis, seen && !tracing
}
