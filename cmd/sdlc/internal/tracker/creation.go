package tracker

// Creation reserves a card before materializing its branch-local detail. Its
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
	{"creation.detail", MaterializeDetail},
}
