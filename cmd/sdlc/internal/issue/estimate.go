package issue

// EstimateSection returns the `## Estimate` section body and whether it exists.
// The fenced ```estimate block the estimate package parses lives inside it.
// Delegates to the shared SectionBody helper (no bespoke regex — #117 M2 review).
func EstimateSection(body string) (string, bool) {
	return SectionBody(body, "Estimate")
}
