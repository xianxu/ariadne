package weavefs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GoModEditor is weave's ONE exec seam: it mutates a repo's root go.mod for the
// `tool` intent's owner self-walk (ports ensure_go_tool_dependency's self-walk
// branch — `go mod edit -tool <module>/<path>`). Injected into plan.Apply so the
// planner stays pure (ARCH-PURE) and Apply is testable with a fake editor that
// records the call instead of shelling out.
type GoModEditor interface {
	// AddTool declares toolImportPath as a `tool` directive in the go.mod at
	// gomodPath (the module-qualified import path, e.g.
	// github.com/xianxu/ariadne/cmd/sdlc). Idempotent — `go mod edit -tool` is a
	// no-op when the directive already exists. Implementations also ensure the
	// go directive is >= 1.24 first (the version that introduced `tool`), porting
	// ensure_go_directive_24.
	AddTool(gomodPath, toolImportPath string) error
}

// OSGoMod is the production GoModEditor: it shells out to the `go` toolchain,
// matching setup.sh's `( cd "$TARGET_DIR" && go mod edit -tool ... )`. Its zero
// value is ready to use.
type OSGoMod struct{}

// AddTool runs `go mod edit -tool <toolImportPath>` in gomodPath's directory,
// bumping the go directive to 1.24 first if it is older (the `tool` directive
// needs Go 1.24). Ports ensure_go_tool_dependency's self-walk branch +
// ensure_go_directive_24 (ARCH-DRY). A missing `go` on PATH surfaces as the
// exec error, the same observable failure setup.sh guarded with a skip.
func (OSGoMod) AddTool(gomodPath, toolImportPath string) error {
	dir := filepath.Dir(gomodPath)
	if err := ensureGoDirective24(dir, gomodPath); err != nil {
		return err
	}
	cmd := exec.Command("go", "mod", "edit", "-tool", toolImportPath)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod edit -tool %s in %s: %w: %s", toolImportPath, dir, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ensureGoDirective24 bumps the go directive in the go.mod at gomodPath to 1.24
// when it is older — `go mod edit -tool` requires it. Ports
// ensure_go_directive_24: read the current `go` line, and only when it is < 1.24
// run `go mod edit -go=1.24`. No-op when the directive is absent or already
// recent (so a fresh-enough module is never rewritten).
func ensureGoDirective24(dir, gomodPath string) error {
	data, err := os.ReadFile(gomodPath)
	if err != nil {
		return fmt.Errorf("read go.mod %s: %w", gomodPath, err)
	}
	cur := goDirective(string(data))
	if cur == "" || goAtLeast124(cur) {
		return nil
	}
	cmd := exec.Command("go", "mod", "edit", "-go=1.24")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod edit -go=1.24 in %s: %w: %s", dir, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// goDirective extracts the version from the first `go ` line of go.mod content
// (awk '/^go / {print $2; exit}'). Pure. "" when absent.
func goDirective(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "go ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	}
	return ""
}

// goAtLeast124 reports whether the go directive version (e.g. "1.26", "1.23.4")
// is >= 1.24. Ports ensure_go_directive_24's major/minor comparison. Pure.
func goAtLeast124(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return false
	}
	major := atoiOrZero(parts[0])
	minor := atoiOrZero(parts[1])
	if major > 1 {
		return true
	}
	return major == 1 && minor >= 24
}

// atoiOrZero parses a non-negative integer prefix, returning 0 on any junk —
// matching the shell's tolerant arithmetic comparison.
func atoiOrZero(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// ensure OSGoMod satisfies GoModEditor at compile time.
var _ GoModEditor = OSGoMod{}
