package issue

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// One operation validates a repository's 100 already-current active details.
// Each issue has its own card/baseline and roughly 4 KiB of branch-local design
// and log text; setup and blob hashing are outside the timed operation.
func BenchmarkRefreshMirrorHundredActiveDetails(b *testing.B) {
	type input struct{ card, details []byte }
	inputs := make([]input, 100)
	var totalBytes int64
	for i := range inputs {
		raw := strings.Replace(cardDetailFixture, "000252", fmt.Sprintf("%06d", i+1), 1)
		raw = strings.Replace(raw, "status: open", "status: working\nstarted: 2026-09-25T14:00:00-07:00", 1)
		raw = strings.Replace(raw, "Keep  these bytes.", strings.Repeat("The branch owns implementation requirements and acceptance evidence.\n", 20), 1)
		raw += strings.Repeat("- Investigated the current behavior, recorded the decision, and verified the resulting change.\n", 25)
		card, details, err := SplitCard([]byte(raw))
		if err != nil {
			b.Fatal(err)
		}
		inputs[i] = input{card, details}
		totalBytes += int64(len(details))
	}
	b.ReportAllocs()
	b.SetBytes(totalBytes)
	b.ResetTimer()
	for range b.N {
		for _, in := range inputs {
			got, err := RefreshMirror(in.details, in.card, in.card)
			if err != nil || !bytes.Equal(got, in.details) {
				b.Fatalf("current mirror changed: %v", err)
			}
		}
	}
	b.ReportMetric(100, "details/op")
}
