package progress

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {
	var output bytes.Buffer
	Write(&output, "Painting frames", 2, 4, 2*time.Second)

	for _, expected := range []string{"Painting frames", "50%", "(2/4)", "ETA~2s"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("progress output %q does not contain %q", output.String(), expected)
		}
	}
}
