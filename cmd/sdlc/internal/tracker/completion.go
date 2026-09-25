package tracker

// Completion binds card transitions to durable branch evidence and its exact
// landing. Done publication and archival have separate recoverable stages.
type Completion struct{ receipt Receipt }

func NewCompletion(spec ReceiptSpec) (Completion, error) {
	r, err := newReceipt("completion", spec)
	return Completion{r}, err
}
func ResumeCompletion(r Receipt) (Completion, error) {
	next, err := resumeReceipt(r, "completion")
	return Completion{next}, err
}
func (s Completion) Receipt() Receipt { return s.receipt }
func (s Completion) Binding() Binding { return s.receipt.binding() }
func (s Completion) Outcome() Outcome { return s.receipt.Outcome() }
func StepCompletion(s Completion, event Event) (Completion, []Effect, error) {
	next, effects, err := stepOperation(s.receipt, "completion", event)
	return Completion{next}, effects, err
}

var completionStages = []operationStage{
	{"completion.evidence", WriteEvidence},
	{"completion.codecomplete", PublishCard},
	{"completion.landing", ObserveLanding},
	{"completion.done", PublishCard},
	{"completion.archive", ArchiveDetails},
}
