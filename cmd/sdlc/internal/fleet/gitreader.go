package fleet

// GitReader is the narrow Git boundary shared by fleet vantage normalization,
// fact collection, and inventory. The caller supplies the directory so every
// response is interpreted against an explicit worktree vantage.
//
// The package-main Cobra shell adapts execGitRunner.GitInDir to this interface;
// keeping the interface here avoids making the internal package depend on main.
type GitReader interface {
	GitInDir(dir string, args ...string) ([]byte, error)
}
