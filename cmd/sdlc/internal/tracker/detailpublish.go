package tracker

import (
	"fmt"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// DetailPublication prepares the main candidate that publishes an issue's
// first details: exactly these bytes at the destination, refused when main
// already has the file (creation is complete there, and another thread may
// own it). Shared by creation (#308) and the transfer that finishes a
// pre-#308 creation, so both first publications refuse identically.
func DetailPublication(main *gitx.TrunkFile, destination string, detail []byte, msg string) (gitx.Candidate, error) {
	return main.PrepareCandidate(msg, func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
		present, err := view.Exists(destination)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		if present {
			return gitx.TrunkWrite{}, fmt.Errorf("%w: %s", ErrDestinationExists, destination)
		}
		return gitx.TrunkWrite{Write: map[string][]byte{destination: detail}, ExactBytes: true}, nil
	})
}
