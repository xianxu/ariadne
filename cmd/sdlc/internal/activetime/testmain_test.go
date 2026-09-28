package activetime

import (
	"os"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestMain(m *testing.M) {
	testfix.PreferRealGit()
	os.Exit(m.Run())
}
