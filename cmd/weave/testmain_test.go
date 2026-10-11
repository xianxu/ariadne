package main

import (
	"os"
	"testing"

	"github.com/xianxu/ariadne/pkg/testgate"
)

func TestMain(m *testing.M) {
	testgate.Enter() // one machine-wide test slot, and the -parallel cap (#324)
	os.Exit(m.Run())
}
