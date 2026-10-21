package landingline

import "fmt"

// Impact records whether the candidate changed a failing test's package.
// Unknown is deliberately the zero value: missing evidence proves no impact.
type Impact int

const (
	ImpactUnknown Impact = iota
	ImpactUntouched
	ImpactAffected
)

// SplitFailure is one exact failing test with registration and impact evidence.
// The caller owns test-name grammar, registration lookup and package analysis.
type SplitFailure struct {
	Test            string
	RegisteredIssue string
	Impact          Impact
}

// AttributionKind describes whether a split has evidence against its candidate.
type AttributionKind string

const (
	AttributionAttributable AttributionKind = "attributable"
	AttributionUnattributed AttributionKind = "unattributed"
	AttributionUnverified   AttributionKind = "unverified"
)

// SplitAttribution explains a split without granting any passing-suite authority.
type SplitAttribution struct {
	Kind   AttributionKind
	Reason string
}

// ClassifySplit classifies exact failures using only supplied evidence. It scans
// once, retaining a lexical representative of each category for stable reasons
// without sorting or mutating the caller's facts. Malformed facts and global
// uncertainty veto attribution; a valid affected failure can coexist with an
// unrelated failure whose impact is unknown.
func ClassifySplit(failures []SplitFailure, unverified string) SplitAttribution {
	if unverified != "" {
		return SplitAttribution{Kind: AttributionUnverified, Reason: "unverified: " + unverified}
	}
	if len(failures) == 0 {
		return SplitAttribution{Kind: AttributionUnverified, Reason: "unverified: no exact failing tests"}
	}
	var malformed, affected, registered, untouched, unknown string
	selectFirst := func(current *string, candidate string) {
		if *current == "" || candidate < *current {
			*current = candidate
		}
	}
	for _, failure := range failures {
		if failure.Test == "" {
			selectFirst(&malformed, "empty failing test name")
			continue
		}
		if failure.Impact < ImpactUnknown || failure.Impact > ImpactAffected {
			selectFirst(&malformed, fmt.Sprintf("invalid impact %d for %s", failure.Impact, failure.Test))
			continue
		}
		if failure.RegisteredIssue != "" {
			selectFirst(&registered, "registered flake "+failure.Test+" (issue #"+failure.RegisteredIssue+")")
			continue
		}
		switch failure.Impact {
		case ImpactAffected:
			selectFirst(&affected, failure.Test)
		case ImpactUntouched:
			selectFirst(&untouched, failure.Test)
		case ImpactUnknown:
			selectFirst(&unknown, failure.Test)
		}
	}
	if malformed != "" {
		return SplitAttribution{Kind: AttributionUnverified, Reason: "unverified: " + malformed}
	}
	if affected != "" {
		return SplitAttribution{Kind: AttributionAttributable, Reason: "attributable: affected package " + affected}
	}
	if unknown != "" {
		return SplitAttribution{Kind: AttributionUnverified, Reason: "unverified: unknown package impact " + unknown}
	}
	reason := registered
	if untouched != "" {
		if reason != "" {
			reason += "; "
		}
		reason += "untouched package " + untouched
	}
	return SplitAttribution{Kind: AttributionUnattributed, Reason: "unattributed: " + reason}
}
