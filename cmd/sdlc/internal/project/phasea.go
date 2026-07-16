package project

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	phaseALineRE  = regexp.MustCompile(`(?m)^\*\*phase-a:\*\*\s*(.*?)\s*$`)
	phaseAValueRE = regexp.MustCompile(`^((?:\d+(?:\.\d+)?|\.\d+))h$`)
)

// ParsePhaseA distinguishes an absent legacy estimate from a valid estimate
// and malformed/non-positive input. Consumers may explicitly bypass absence;
// malformed input always fails closed.
func ParsePhaseA(estimateBody string) (hours float64, present bool, err error) {
	line := phaseALineRE.FindStringSubmatch(estimateBody)
	if line == nil {
		return 0, false, nil
	}
	value := strings.TrimSpace(line[1])
	m := phaseAValueRE.FindStringSubmatch(value)
	if m == nil {
		return 0, true, fmt.Errorf("invalid phase-a %q; expected **phase-a:** <positive N>h", value)
	}
	hours, err = strconv.ParseFloat(m[1], 64)
	if err != nil || hours <= 0 {
		return 0, true, fmt.Errorf("invalid phase-a %q; hours must be positive", value)
	}
	return hours, true, nil
}
