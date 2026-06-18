package estimate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// estimateFenceRE captures the body inside a ```estimate fenced block: submatch 1
// is everything between the opening fence line and the closing ``` line.
var estimateFenceRE = regexp.MustCompile("(?ms)^```estimate[^\n]*\n(.*?)^```")

// itemRE matches `item: <slug>  design=<f> impl=<f>`.
var itemRE = regexp.MustCompile(`^item:\s*(\S+)\s+design=(\S+)\s+impl=(\S+)\s*$`)

// ParseBlock extracts and parses the ```estimate fenced block from a `## Estimate`
// section body. familiarity defaults to 1.0 and design-buffer to 0.30 when the
// line is absent. Pure: a string→Block function with no IO. Semantic validation
// (known model/primitives, reconciliation) lives in Check, not here.
func ParseBlock(section string) (Block, error) {
	m := estimateFenceRE.FindStringSubmatch(section)
	if m == nil {
		return Block{}, fmt.Errorf("no ```estimate fenced block found in the ## Estimate section")
	}
	b := Block{Familiarity: 1.0, DesignBuffer: 0.30}
	sawTotal := false
	for _, raw := range strings.Split(m[1], "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if im := itemRE.FindStringSubmatch(line); im != nil {
			d, err := strconv.ParseFloat(im[2], 64)
			if err != nil {
				return Block{}, fmt.Errorf("item %q: design=%q is not a number", im[1], im[2])
			}
			i, err := strconv.ParseFloat(im[3], 64)
			if err != nil {
				return Block{}, fmt.Errorf("item %q: impl=%q is not a number", im[1], im[3])
			}
			b.Items = append(b.Items, Item{Slug: im[1], Design: d, Impl: i})
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			return Block{}, fmt.Errorf("unparseable estimate line: %q", line)
		}
		key, val = strings.TrimSpace(key), strings.TrimSpace(val)
		switch key {
		case "model":
			b.Model = val
		case "familiarity":
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return Block{}, fmt.Errorf("familiarity %q is not a number", val)
			}
			b.Familiarity = f
		case "design-buffer":
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return Block{}, fmt.Errorf("design-buffer %q is not a number", val)
			}
			b.DesignBuffer = f
		case "total":
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return Block{}, fmt.Errorf("total %q is not a number", val)
			}
			b.Total, sawTotal = f, true
		default:
			return Block{}, fmt.Errorf("unknown estimate field %q", key)
		}
	}
	if !sawTotal {
		return Block{}, fmt.Errorf("estimate block missing the required `total:` line")
	}
	return b, nil
}
