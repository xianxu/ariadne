package reviewbox

import (
	"os"
	"path/filepath"
	"testing"
)

// A new checkout goes to the RAM disk when it has room; a sweep covers both
// roots, so a checkout made on either is reclaimed and none is moved (#468).
func TestRootsPreferScratch(t *testing.T) {
	cache, err := os.UserCacheDir()
	if err != nil {
		t.Skip(err)
	}
	disk := filepath.Join(cache, "sdlc", "review-checkouts")
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", "")

	ram := t.TempDir()
	t.Setenv("WF_TEST_SCRATCH", ram)
	create, all, note, err := Roots()
	want := filepath.Join(ram, "review-checkouts")
	if err != nil || note != "" || create != want || len(all) != 2 || all[0] != want || all[1] != disk {
		t.Fatalf("scratch: create=%q all=%q note=%q err=%v", create, all, note, err)
	}

	t.Setenv("WF_TEST_SCRATCH", "off")
	create, all, _, _ = Roots()
	if create != disk || len(all) != 1 {
		t.Fatalf("off: create=%q all=%q", create, all)
	}

	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", "/x/only")
	t.Setenv("WF_TEST_SCRATCH", ram)
	create, all, _, _ = Roots()
	if create != "/x/only" || len(all) != 1 {
		t.Fatalf("explicit: create=%q all=%q", create, all)
	}
}
