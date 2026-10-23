package gitx

import (
	"fmt"
	"sort"
	"strings"
)

// MaxPathComponentBytes is the file-name limit of APFS and ext4. A pushed tree
// is checked out on other machines, so the limit is the common one, not this
// filesystem's (#566: a 331-byte name broke every parley checkout).
const MaxPathComponentBytes = 255

// CheckPathLimits refuses a path that some checkout could not create.
func CheckPathLimits(paths []string) error {
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)
	for _, p := range sorted {
		for _, c := range strings.Split(p, "/") {
			if len(c) > MaxPathComponentBytes {
				return fmt.Errorf("refusing to write %s: its name is %d bytes, over the %d-byte file-name limit of a checkout", p, len(c), MaxPathComponentBytes)
			}
		}
	}
	return nil
}
