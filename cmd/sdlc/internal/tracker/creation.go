package tracker

// Creation reserves a card, then publishes its details to main (#308): an
// issue is filed on origin in one operation, and no checkout is written. Its
// stage is private; after reservation no event can return to ID allocation.
type Creation struct{ receipt Receipt }

func NewCreation(spec ReceiptSpec) (Creation, error) {
	r, err := newReceipt("creation", spec)
	return Creation{r}, err
}
func ResumeCreation(r Receipt) (Creation, error) {
	next, err := resumeReceipt(r, "creation")
	return Creation{next}, err
}
func (s Creation) Receipt() Receipt { return s.receipt }
func (s Creation) Binding() Binding { return s.receipt.binding() }
func (s Creation) Outcome() Outcome { return s.receipt.Outcome() }
func StepCreation(s Creation, event Event) (Creation, []Effect, error) {
	next, effects, err := stepOperation(s.receipt, "creation", event)
	return Creation{next}, effects, err
}

var creationStages = []operationStage{
	{"creation.reserve", PublishCard},
	{"creation.main", PublishMain},
}

// LegacyCreationDetail reports a creation receipt an older binary left
// mid-way through its retired local stage (creation.detail, before #308): its
// pinned candidate is the detail blob itself, which no main publication can
// be. Such a receipt can never drive creation.main; its card is reserved, and
// `sdlc issue publish` finishes the creation.
func (r Receipt) LegacyCreationDetail() bool {
	return r.wire.Operation == "creation" && r.wire.Stage == 1 &&
		r.wire.CandidateOID != "" && r.wire.CandidateOID == r.wire.Spec.SourceBlob
}
