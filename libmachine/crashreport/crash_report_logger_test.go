package crashreport

import (
	"testing"

	"github.com/rancher/machine/libmachine/log"
)

func TestLoggerPrintfForwardsArguments(t *testing.T) {
	(&logger{}).Printf("%s=%d", "checkpoint", 3)
	history := log.History()
	if len(history) == 0 || history[len(history)-1] != "checkpoint=3" {
		t.Fatalf("formatted diagnostic: %v", history)
	}
}
