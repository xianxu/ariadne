package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Interruption is execution evidence outside SuiteAttempt and flake/main
// classification. Resume permits a manual retry; it does not restart this run.
type queueInterruptedError struct {
	ID, Reason, LogPath string
	Cause               error
}

func (e *queueInterruptedError) Error() string {
	s := "queue interrupted: " + e.Reason + "; rerun after the matching pause owners resume"
	if e.LogPath != "" {
		s += "; log " + e.LogPath
	}
	if e.Cause != nil {
		s += "; " + e.Cause.Error()
	}
	return s
}
func (e *queueInterruptedError) Unwrap() error { return e.Cause }
func queueInterruptionReason(err error) string {
	var interruption *queueInterruptedError
	if errors.As(err, &interruption) {
		return interruption.Reason
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err.Error()
	}
	return ""
}

func retainQueueInterruption(logDir, sha, check, raw string, e *queueInterruptedError) error {
	path := filepath.Join(logDir, sha, check+".interrupted.log")
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err == nil {
		err = os.Rename(raw, path)
	}
	if err != nil {
		e.LogPath = raw
		e.Cause = errors.Join(e.Cause, fmt.Errorf("retain interruption log: %w", err))
		return e
	}
	e.LogPath = path
	pruneLocalCheckLogs(logDir, localCheckLogKeep)
	return e
}
