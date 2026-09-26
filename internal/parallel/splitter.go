package parallel

import (
	"github.com/ACKERMANNGUE/go-painter/internal/model"
)

func SplitRows(height int, workers int) []model.RangeRow {
	if workers <= 0 {
		workers = 1
	}
	if workers > height {
		workers = height
	}

	rows := make([]model.RangeRow, 0, workers)
	start := 0
	end := 0

	for i := 0; i < workers; i++ {
		start = i * height / workers
		end = (i + 1) * height / workers
		rows = append(rows, model.RangeRow{
			Start: start,
			End:   end,
		})
	}

	return rows
}
