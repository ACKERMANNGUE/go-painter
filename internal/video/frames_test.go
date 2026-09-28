package video

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestWriteFrameProgress(t *testing.T) {
	var output bytes.Buffer
	writeFrameProgress(&output, 2, 4, 2*time.Second)

	for _, expected := range []string{"50%", "(2/4)", "ETA~2s"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("progress output %q does not contain %q", output.String(), expected)
		}
	}
}
