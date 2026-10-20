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
	Check, LogPath string
	Cause          error
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
	return fmt.Sprintf("environment failure: Go could not create its work directory (GOTMPDIR/TMPDIR); repair temp-directory access and rerun sdlc local-check run --name %s; log %s%s", check, e.LogPath, detail)
}

func (e *suiteEnvironmentError) Unwrap() error { return e.Cause }

var workDirStartup = regexp.MustCompile(`^go: creating work dir: mkdir [^\x00-\x1f\x7f]+: (permission denied|operation not permitted|no space left on device|read-only file system|no such file or directory|not a directory|file exists|too many open files)$`)
var startupMakeExit = regexp.MustCompile(`^make(?:\[\d+\])?: \*\*\* \[test\] Error [1-9]\d*$`)

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

func suiteEnvironmentLog(log string) bool { return suiteEnvironmentLines(strings.Lines(log)) }

// One pass, bounded lines and fixed grammar work per line. Taking a sequence
// lets the N/2N test count actual input visits without timing a shared machine.
func suiteEnvironmentLines(lines iter.Seq[string]) bool {
	seen, trace := false, -1
	for raw := range lines {
		if !strings.HasSuffix(raw, "\n") || len(raw) > suiteLogLineLimit+1 {
			return false
		}
		line := strings.TrimSuffix(raw, "\n")
		if trace >= 0 {
			if !startupTrace[trace].MatchString(line) {
				return false
			}
			trace++
			if trace == len(startupTrace) {
				trace = -1
			}
			continue
		}
		switch {
		case workDirStartup.MatchString(line):
			seen = true
		case seen && line == "Traceback (most recent call last):":
			trace = 0
		case seen && startupMakeExit.MatchString(line):
		default:
			return false
		}
	}
	return seen && trace == -1
}
