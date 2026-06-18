package issue

import "regexp"

// EstimateSectionRE captures the body of the `## Estimate` section, stopping at
// the next top-level `##` heading or end-of-text. Submatch 1 is the body
// (including the fenced ```estimate block the estimate package parses). Mirrors
// PlanSectionRE / specSectionRE so the section-extraction pattern stays DRY.
var EstimateSectionRE = regexp.MustCompile(`(?ms)^## Estimate\s*\n(.*?)(?:^## |\z)`)

// EstimateSection returns the `## Estimate` section body and whether it exists.
func EstimateSection(body string) (string, bool) {
	m := EstimateSectionRE.FindStringSubmatch(body)
	if m == nil {
		return "", false
	}
	return m[1], true
}
