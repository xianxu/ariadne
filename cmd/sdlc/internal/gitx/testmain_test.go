package gitx

import (
	"os"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/testgate"
)

func TestMain(m *testing.M) {
	testgate.Enter() // one machine-wide test slot, and the -parallel cap (#324)
	testfix.PreferRealGit()
	testfix.PinAgent()                              // #367
	cleanupQuietGit := testfix.QuietBackgroundGit() // pushes into bare origins (#302)
	code := m.Run()
	cleanupQuietGit()
	os.Exit(code)
}
