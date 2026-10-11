package reviewbox

import (
	"os"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/testgate"
)

func TestMain(m *testing.M) {
	testgate.Enter() // one machine-wide test slot, and the -parallel cap (#324)
	testfix.PreferRealGit()
	os.Exit(m.Run())
}
