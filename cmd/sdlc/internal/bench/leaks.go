package bench

import (
	"path/filepath"
	"strings"
)

// ToolEvent is a normalized tool invocation, never its prose/output. Adapters
// must emit opaque events for unsupported tool invocations rather than omit them.
// CWD and paths are recorded evidence; this pure scanner cannot resolve symlinks.
type ToolEvent struct {
	ID          string   `json:"id,omitempty"`
	Kind        string   `json:"kind"` // read, exec, opaque
	CWD         string   `json:"cwd,omitempty"`
	Paths       []string `json:"paths,omitempty"`
	Argv        []string `json:"argv,omitempty"`    // only literal, decoded execution argv
	Command     string   `json:"command,omitempty"` // shell source, never treated as literal argv
	Unsupported string   `json:"unsupported,omitempty"`
}

// LeakScope contains canonical roots established by the IO boundary. Complete
// means the bound transcript prefix and adapter covered every tool invocation.
type LeakScope struct {
	PreparedRoots  []string
	ProtectedRoots []string
	PrivateOrigin  string
	Complete       bool
}

type LeakFinding struct {
	EventID    string `json:"event_id,omitempty"`
	EventIndex int    `json:"event_index"`
	Kind       string `json:"kind"`
	Evidence   string `json:"evidence"`
}

type LeakCoverage struct {
	Events         int    `json:"events"`
	Supported      int    `json:"supported"`
	Unknown        int    `json:"unknown"`
	Operands       int    `json:"operands"`
	PrivateFetches int    `json:"private_fetches"`
	Complete       bool   `json:"complete"`
	Basis          string `json:"basis"`
}

type LeakReport struct {
	Status   string        `json:"status"`
	Findings []LeakFinding `json:"findings,omitempty"`
	Coverage LeakCoverage  `json:"coverage"`
	Issues   []string      `json:"issues,omitempty"`
}

// ScanLeakEvents diagnoses observed access, not secrecy. It visits each event
// once and each operand once; root comparisons are bounded by scope validation.
// All evidence is bounded, and no filesystem, shell or network IO is performed.
func ScanLeakEvents(scope LeakScope, events []ToolEvent) LeakReport {
	report := LeakReport{Status: "unknown", Coverage: LeakCoverage{Complete: scope.Complete, Basis: "lexical recorded tool operands; no symlink, host-read or network isolation proof"}}
	if !validLeakScope(scope) {
		report.Issues = []string{"invalid canonical leak scope"}
		return report
	}
	for index, event := range events {
		report.Coverage.Events++
		unknown := event.Unsupported != ""
		finding := func(kind, evidence string) {
			if len(evidence) > 512 {
				evidence = evidence[:512]
			}
			id := event.ID
			if len(id) > 128 {
				id = id[:128]
			}
			report.Findings = append(report.Findings, LeakFinding{EventID: id, EventIndex: index, Kind: kind, Evidence: evidence})
		}
		read := func(operand, cwd string) {
			report.Coverage.Operands++
			resolved, ok := leakPath(operand, cwd)
			if !ok {
				unknown = true
				return
			}
			if anyLeakRoot(scope.PreparedRoots, resolved) {
				return
			}
			kind := "outside-read"
			if anyLeakRoot(scope.ProtectedRoots, resolved) {
				kind = "protected-read"
			}
			finding(kind, resolved)
		}
		// Reject oversized decoded values before constructing joined command text.
		size := len(event.Command) + len(event.CWD) + len(event.ID) + len(event.Unsupported)
		for _, v := range event.Argv {
			size += len(v)
		}
		for _, v := range event.Paths {
			size += len(v)
		}
		if size > 1<<20 || len(event.Argv)+len(event.Paths) > 10001 {
			report.Coverage.Unknown++
			report.Coverage.Complete = false
			report.Issues = append(report.Issues, "oversized tool event omitted")
			continue
		}
		switch event.Kind {
		case "read":
			if len(event.Paths) == 0 {
				unknown = true
			}
			for _, p := range event.Paths {
				read(p, event.CWD)
			}
		case "exec":
			argv := event.Argv
			text := event.Command
			if text != "" || len(argv) == 0 {
				unknown = true
			}
			if len(argv) > 0 {
				if text != "" {
					text += " "
				}
				text += strings.Join(argv, " ")
				executable := filepath.Base(argv[0])
				switch executable {
				case "git":
					args, cwd, ok := leakGitArgs(argv[1:], event.CWD)
					if !ok || len(args) == 0 {
						unknown = true
						break
					}
					if args[0] == "fetch" || args[0] == "clone" || args[0] == "pull" {
						if event.Command == "" && safeFetchArgs(args) && exactPrivateFetch(args[1], cwd, scope.PrivateOrigin) {
							report.Coverage.PrivateFetches++
						} else {
							finding("git-transfer", strings.Join(argv, " "))
						}
					} else {
						unknown = true
					}
				case "gh":
					finding("github-access", strings.Join(argv, " "))
				case "cat", "head", "tail", "less", "more":
					if len(argv) == 1 {
						unknown = true
					}
					for _, p := range argv[1:] {
						if strings.HasPrefix(p, "-") {
							unknown = true
							continue
						}
						read(p, event.CWD)
					}
				default:
					unknown = true
				}
			}
			lower := strings.ToLower(text)
			if strings.Contains(lower, "github.com") || strings.Contains(lower, "githubusercontent.com") {
				finding("github-access", text)
			}
			if unknown {
				// Shell source and wrappers are only suspicious text, never parsed proof.
				for _, root := range scope.ProtectedRoots {
					if strings.Contains(text, root) {
						finding("protected-reference", root)
					}
				}
				if strings.Contains(lower, "fetch") || strings.Contains(lower, "clone") || strings.Contains(lower, "git pull") {
					finding("git-transfer", text)
				}
			}
		default:
			unknown = true
		}
		if unknown {
			report.Coverage.Unknown++
		} else {
			report.Coverage.Supported++
		}
	}
	if len(report.Findings) > 0 {
		report.Status = "possibly-leaked"
	} else if scope.Complete && report.Coverage.Unknown == 0 && report.Coverage.Events > 0 {
		report.Status = "no-observed-leak"
	}
	return report
}

func validLeakScope(scope LeakScope) bool {
	if len(scope.PreparedRoots) == 0 || len(scope.PreparedRoots) > 16 || len(scope.ProtectedRoots) > 16 {
		return false
	}
	for _, roots := range [][]string{scope.PreparedRoots, scope.ProtectedRoots} {
		for _, root := range roots {
			if len(root) > 4096 || !filepath.IsAbs(root) || filepath.Clean(root) != root || root == string(filepath.Separator) || strings.ContainsRune(root, 0) {
				return false
			}
		}
	}
	return scope.PrivateOrigin == "" || (filepath.IsAbs(scope.PrivateOrigin) && filepath.Clean(scope.PrivateOrigin) == scope.PrivateOrigin && !strings.ContainsRune(scope.PrivateOrigin, 0))
}

// Only a target and optional literal refspec are excluded. Fetch options can
// execute arbitrary upload-pack programs or override the transport identity.
func safeFetchArgs(args []string) bool {
	if len(args) < 2 || args[0] != "fetch" {
		return false
	}
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "-") || strings.ContainsAny(arg, "\x00\n\r") {
			return false
		}
	}
	return true
}

func anyLeakRoot(roots []string, path string) bool {
	for _, root := range roots {
		if path == root || strings.HasPrefix(path, root+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func leakPath(path, cwd string) (string, bool) {
	if path == "" || strings.ContainsAny(path, "\x00$`*?[]\n\r") || strings.HasPrefix(path, "~") {
		return "", false
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), true
	}
	if !filepath.IsAbs(cwd) || strings.ContainsRune(cwd, 0) {
		return "", false
	}
	return filepath.Clean(filepath.Join(cwd, path)), true
}

func leakGitArgs(args []string, cwd string) ([]string, string, bool) {
	for len(args) >= 2 && args[0] == "-C" {
		var ok bool
		cwd, ok = leakPath(args[1], cwd)
		if !ok {
			return nil, "", false
		}
		args = args[2:]
	}
	if len(args) > 0 && strings.HasPrefix(args[0], "-") {
		return nil, "", false
	}
	return args, cwd, true
}

func exactPrivateFetch(target, cwd, origin string) bool {
	if origin == "" || strings.Contains(target, ":") {
		return false
	}
	// A bare remote name could resolve via Git configuration, so it is not proof.
	if !filepath.IsAbs(target) && !strings.HasPrefix(target, "./") && !strings.HasPrefix(target, "../") {
		return false
	}
	resolved, ok := leakPath(target, cwd)
	return ok && resolved == origin
}
