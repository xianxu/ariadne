package tracker

// Transfer records tracker provenance before main publication and confirms that
// publication back to the tracker before permitting local source removal.
type Transfer struct{ receipt Receipt }

func NewTransfer(spec ReceiptSpec) (Transfer, error) {
	r, err := newReceipt("transfer", spec)
	return Transfer{r}, err
}
func ResumeTransfer(r Receipt) (Transfer, error) {
	next, err := resumeReceipt(r, "transfer")
	return Transfer{next}, err
}
func (s Transfer) Receipt() Receipt { return s.receipt }
func (s Transfer) Binding() Binding { return s.receipt.binding() }
func (s Transfer) Outcome() Outcome { return s.receipt.Outcome() }
func StepTransfer(s Transfer, event Event) (Transfer, []Effect, error) {
	next, effects, err := stepOperation(s.receipt, "transfer", event)
	return Transfer{next}, effects, err
}

var transferStages = []operationStage{
	{"transfer.handoff", PublishCard},
	{"transfer.main", PublishMain},
	{"transfer.record", PublishCard},
	{"transfer.remove", RemoveSource},
}
