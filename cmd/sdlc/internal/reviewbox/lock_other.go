//go:build !unix

package reviewbox

import (
	"errors"
	"os"
)

func lockExclusive(*os.File, bool) error {
	return errors.New("review checkouts need flock (unix only)")
}
