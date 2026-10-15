// markdoc.go — the fence, section and plan scanners live in pkg/markdoc (#350),
// so a binary outside cmd/sdlc (tl-lead's progress bar) counts exactly what the
// gates count. These names forward to it; no logic lives here.
package issue

import "github.com/xianxu/ariadne/pkg/markdoc"

type UnterminatedPolicy = markdoc.UnterminatedPolicy

const (
	UnterminatedIsProse  = markdoc.UnterminatedIsProse
	UnterminatedIsFenced = markdoc.UnterminatedIsFenced

	MilestoneTagPattern = markdoc.MilestoneTagPattern
)

var (
	PlanItemRE      = markdoc.PlanItemRE
	PlanUncheckedRE = markdoc.PlanUncheckedRE

	FenceSpans            = markdoc.FenceSpans
	ScanMarkdownLines     = markdoc.ScanMarkdownLines
	StripFenced           = markdoc.StripFenced
	FindLineOutsideFences = markdoc.FindLineOutsideFences

	SectionBody              = markdoc.SectionBody
	SectionLineBounds        = markdoc.SectionLineBounds
	PlanSectionBody          = markdoc.PlanSectionBody
	PlanItemsBody            = markdoc.PlanItemsBody
	SectionByteBounds        = markdoc.SectionByteBounds
	SectionHeadingByteOffset = markdoc.SectionHeadingByteOffset
	HasSection               = markdoc.HasSection

	CountPlanItems        = markdoc.CountPlanItems
	TickMilestone         = markdoc.TickMilestone
	MilestonesInPlanOrder = markdoc.MilestonesInPlanOrder
	TickedMilestones      = markdoc.TickedMilestones
)
