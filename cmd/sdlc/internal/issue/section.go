package issue

import "regexp"

// SectionBody returns the body of a top-level `## <heading>` section — the text
// after `## <heading>` up to the next `## ` heading or end-of-text — and whether
// the section exists. Consolidates the `^## <H>\s*\n(.*?)(?:^## |\z)` pattern that
// the spec / done-when / estimate gates each used to spell out (#117 M2 review).
//
// checkPlan keeps its own PlanSectionRE: it needs byte offsets
// (FindStringSubmatchIndex), not just the body. The regex is compiled per call —
// section extraction runs a handful of times per command, and a package-level
// cache would need a mutex for test parallelism; not worth it. Pure.
func SectionBody(body, heading string) (string, bool) {
	re := regexp.MustCompile(`(?ms)^## ` + regexp.QuoteMeta(heading) + `\s*\n(.*?)(?:^## |\z)`)
	m := re.FindStringSubmatch(body)
	if m == nil {
		return "", false
	}
	return m[1], true
}
