package reviewusage

import (
	"os"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/testgate"
)

func TestMain(m *testing.M) {
	testgate.Enter()
	testfix.PreferRealGit()
	cleanup := testfix.QuietBackgroundGit()
	code := m.Run()
	cleanup()
	os.Exit(code)
}
