package gitx

import (
	"os"
	"path/filepath"
)

// IsBrainRepo reports whether repoTop is a "brain" capture repo. Brain is
// defined canonically by the presence of `.brain/config.md` (AGENTS §1), NOT by
// a repo basename — a brain checked out under any name is still a brain. The
// spine guard, `sdlc migrate`, and cross-repo project discovery all key off this
// single predicate so the "what is a brain" rule has one source (ARCH-DRY,
// #171 M2 review). repoTop is the repo's top-level directory.
func IsBrainRepo(repoTop string) bool {
	_, err := os.Stat(filepath.Join(repoTop, ".brain", "config.md"))
	return err == nil
}
