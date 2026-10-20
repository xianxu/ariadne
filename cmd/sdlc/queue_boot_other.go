//go:build !darwin && !linux

package main

import "errors"

func queueBootID() (string, error) { return "", errors.New("machine boot identity unavailable") }
