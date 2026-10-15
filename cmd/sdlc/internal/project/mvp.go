package project

import "fmt"

// MVP is mvp_scope split by landing (#373): the one definition of "MVP done"
// shared by the mvp-landed guard, `project status` and `project close`.
type MVP struct {
	Landed, Open []string
	Errors       []error // a ref whose status could not be read; never counted as landed
}

// Done reports a non-empty scope with every ref landed and every ref read.
func (m MVP) Done() bool {
	return len(m.Open) == 0 && len(m.Errors) == 0 && len(m.Landed) > 0
}

// MVPProgress classifies each ref through the terminal-status lookup.
func MVPProgress(scope []string, terminal func(ref string) (bool, error)) MVP {
	var m MVP
	for _, ref := range scope {
		done, err := terminal(ref)
		switch {
		case err != nil:
			m.Errors = append(m.Errors, fmt.Errorf("mvp_scope %s: %w", ref, err))
		case done:
			m.Landed = append(m.Landed, ref)
		default:
			m.Open = append(m.Open, ref)
		}
	}
	return m
}
