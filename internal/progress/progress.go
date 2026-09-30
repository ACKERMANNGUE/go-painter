package progress

import (
	"fmt"
	"io"
	"strings"
	"time"
)

func Write(output io.Writer, label string, completed, total int, elapsed time.Duration) {
	if total <= 0 {
		return
	}
	if completed < 0 {
		completed = 0
	} else if completed > total {
		completed = total
	}

	const barWidth = 24
	percentage := completed * 100 / total
	filled := percentage * barWidth / 100
	bar := strings.Repeat("#", filled) + strings.Repeat("-", barWidth-filled)
	eta := time.Duration(0)
	if completed > 0 {
		eta = time.Duration(float64(elapsed) * float64(total-completed) / float64(completed))
	}
	fmt.Fprintf(output, "\r%s [%s] %3d%% (%d/%d) elapsed=%s ETA~%s",
		label, bar, percentage, completed, total, elapsed.Round(time.Second), eta.Round(time.Second))
}
