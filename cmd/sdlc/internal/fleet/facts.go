package fleet

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CollectFacts measures one canonical worktree through GitReader. It is total:
// a failed command is represented on the returned facts rather than silently
// becoming an empty value or a zero count.
func CollectFacts(git GitReader, worktreeRoot string) MeasuredFacts {
	facts := MeasuredFacts{}
	if git == nil {
		facts.Error = "git reader is nil"
		return facts
	}

	headOut, err := git.GitInDir(worktreeRoot, "rev-parse", "HEAD")
	if err != nil {
		facts.Error = gitFactFailure("rev-parse HEAD", err, headOut)
		return facts
	}
	facts.Head = gitLine(headOut)
	if facts.Head == "" {
		facts.Error = "git rev-parse HEAD returned an empty SHA"
		return facts
	}

	timeOut, err := git.GitInDir(worktreeRoot, "show", "-s", "--format=%cI", "HEAD")
	if err != nil {
		facts.Error = gitFactFailure("show -s --format=%cI HEAD", err, timeOut)
		return facts
	}
	facts.CommitTimestamp = gitLine(timeOut)
	if _, err := time.Parse(time.RFC3339, facts.CommitTimestamp); err != nil {
		facts.Error = fmt.Sprintf("git show -s --format=%%cI HEAD returned invalid timestamp %q: %v", facts.CommitTimestamp, err)
		return facts
	}

	statusOut, err := git.GitInDir(worktreeRoot, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		facts.Error = gitFactFailure("status --porcelain=v1 -z --untracked-files=all", err, statusOut)
		return facts
	}
	dirtyCount, err := countStatusEntries(statusOut)
	if err != nil {
		facts.Error = fmt.Sprintf("parse git status --porcelain=v1 -z: %v", err)
		return facts
	}
	facts.DirtyCount = &dirtyCount
	facts.Available = true

	facts.collectBase(git, worktreeRoot)
	return facts
}

func (facts *MeasuredFacts) collectBase(git GitReader, worktreeRoot string) {
	var probeFailures []string
	for _, ref := range []string{"origin/main", "main"} {
		out, err := git.GitInDir(worktreeRoot, "rev-parse", "--verify", "--quiet", ref)
		if err != nil {
			if !missingRefError(err) {
				facts.BaseError = gitFactFailure("rev-parse --verify --quiet "+ref, err, out)
				return
			}
			probeFailures = append(probeFailures, gitFactFailure("rev-parse --verify --quiet "+ref, err, out))
			continue
		}
		if gitLine(out) == "" {
			probeFailures = append(probeFailures, "git rev-parse --verify --quiet "+ref+" returned an empty SHA")
			continue
		}
		facts.BaseRef = ref
		facts.collectDivergence(git, worktreeRoot)
		return
	}
	facts.BaseError = "no base reference available: " + strings.Join(probeFailures, "; ")
}

func (facts *MeasuredFacts) collectDivergence(git GitReader, worktreeRoot string) {
	args := []string{"rev-list", "--left-right", "--count", facts.BaseRef + "...HEAD"}
	out, err := git.GitInDir(worktreeRoot, args...)
	if err != nil {
		facts.BaseError = gitFactFailure(strings.Join(args, " "), err, out)
		return
	}
	behind, ahead, err := parseDivergence(out)
	if err != nil {
		facts.BaseError = fmt.Sprintf("parse git rev-list --left-right --count %s...HEAD: %v", facts.BaseRef, err)
		return
	}
	facts.BaseAvailable = true
	facts.Behind = &behind // <base>...HEAD: left side is behind.
	facts.Ahead = &ahead   // <base>...HEAD: right side is ahead.
}

func countStatusEntries(porcelain []byte) (int, error) {
	if len(porcelain) == 0 {
		return 0, nil
	}
	if porcelain[len(porcelain)-1] != 0 {
		return 0, fmt.Errorf("status stream is missing final NUL terminator")
	}
	fields := bytes.Split(porcelain, []byte{0})
	count := 0
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		if len(field) == 0 {
			if i == len(fields)-1 {
				continue
			}
			return 0, fmt.Errorf("empty status field %d", i+1)
		}
		if len(field) < 4 || field[2] != ' ' || !validStatusCode(string(field[:2])) {
			return 0, fmt.Errorf("field %d is not XY+path status", i+1)
		}
		if field[0] == 0 || field[1] == 0 || len(field[3:]) == 0 {
			return 0, fmt.Errorf("field %d is malformed", i+1)
		}
		count++
		if field[0] == 'R' || field[0] == 'C' || field[1] == 'R' || field[1] == 'C' {
			i++
			if i >= len(fields) || len(fields[i]) == 0 {
				return 0, fmt.Errorf("field %d rename/copy is missing source path", i)
			}
		}
	}
	return count, nil
}

func parseDivergence(output []byte) (behind, ahead int, err error) {
	if len(output) < 4 || output[len(output)-1] != '\n' {
		return 0, 0, fmt.Errorf("expected digits<TAB>digits<LF>, got %q", output)
	}
	fields := bytes.Split(output[:len(output)-1], []byte{'\t'})
	if len(fields) != 2 || !decimalBytes(fields[0]) || !decimalBytes(fields[1]) {
		return 0, 0, fmt.Errorf("expected digits<TAB>digits<LF>, got %q", output)
	}
	behind, err = strconv.Atoi(string(fields[0]))
	if err != nil || behind < 0 {
		return 0, 0, fmt.Errorf("invalid behind count %q", fields[0])
	}
	ahead, err = strconv.Atoi(string(fields[1]))
	if err != nil || ahead < 0 {
		return 0, 0, fmt.Errorf("invalid ahead count %q", fields[1])
	}
	return behind, ahead, nil
}

func decimalBytes(value []byte) bool {
	if len(value) == 0 {
		return false
	}
	for _, b := range value {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func validStatusCode(code string) bool {
	switch code {
	case " A", " M", " T", " D",
		"M ", "MM", "MT", "MD",
		"T ", "TM", "TT", "TD",
		"A ", "AM", "AT", "AD",
		"D ",
		"R ", "RM", "RT", "RD",
		"C ", "CM", "CT", "CD",
		" R", " C",
		"DD", "AU", "UD", "UA", "DU", "AA", "UU",
		"??":
		return true
	default:
		return false
	}
}

type exitCoder interface{ ExitCode() int }

func missingRefError(err error) bool {
	var coded exitCoder
	return errors.As(err, &coded) && coded.ExitCode() == 1
}

func gitLine(output []byte) string {
	return strings.TrimSuffix(string(output), "\n")
}

func gitFactFailure(command string, err error, output []byte) string {
	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Sprintf("git %s: %v", command, err)
	}
	return fmt.Sprintf("git %s: %v: %s", command, err, message)
}
