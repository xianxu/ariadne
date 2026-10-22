package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/pkg/workspace"
)

func planSuiteReclassification(a SuiteAttempt, repo, actor, reason string, log []byte) (suiteStartupDiagnosis, error) {
	refuse := func(reason string) (suiteStartupDiagnosis, error) {
		return suiteStartupDiagnosis{}, errors.New("local-check reclassify: " + reason)
	}
	if strings.TrimSpace(reason) == "" || len(reason) > 4096 || !utf8.ValidString(reason) || strings.ContainsRune(reason, 0) {
		return refuse("give a nonempty --reason of at most 4096 UTF-8 bytes")
	}
	address, err := workspace.ParseAddress(actor)
	if err != nil || address.String() != actor {
		return refuse("verified workspace owner or repository :0 required")
	}
	if a.Slot != "" {
		owner, err := workspace.ParseAddress(a.Slot)
		if err != nil || owner.String() != a.Slot || owner.Repo != address.Repo {
			return refuse("attempt belongs to another or unprovable workspace repository")
		}
	}
	if address.Slot != 0 && (a.Slot == "" || a.Slot != actor) {
		return refuse("recorded owner or repository :0 required")
	}
	if a.Repo != repo || a.Legacy || a.Version != 1 && a.Version != 2 {
		return refuse("requires this repository's original supported versioned attempt")
	}
	if !isTestCheck(a.Check) || a.Exit <= 0 || a.TimedOut || flakes.Kind(a.Outcome).Lands() {
		return refuse("requires a failed completed suite or affected attempt, never a pass or timeout")
	}
	if a.LogPath != filepath.Join("failed", a.ID+".log") {
		return refuse("requires the immutable failed-attempt log")
	}
	digest := sha256.Sum256(log)
	if hex.EncodeToString(digest[:]) != a.LogSHA256 {
		return refuse("raw log no longer matches the recorded SHA256")
	}
	diagnosis, ok := diagnoseSuiteStartup(string(log))
	if !ok || len(a.Failures) != 0 || len(a.Units) != 0 {
		return refuse("complete raw output does not prove startup-only failure; mixed, test-started, empty or truncated evidence cannot be corrected")
	}
	return diagnosis, nil
}
