package parallel

import (
	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"sync"
)

func checkWorkersCount(workers int, height int) int {
	if workers <= 0 {
		return 1
	}
	if workers > height {
		return height
	}
	return workers
}

func ForEachRow(height int, workers int, fn func(int)) {
	if height <= 0 {
		return
	}

	checkWorkersCount(workers, height)

	jobs := make(chan int)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := range jobs {
				fn(y)
			}
		}()
	}

	for y := 0; y < height; y++ {
		jobs <- y
	}
	close(jobs)

	wg.Wait()
}

func ForEachRange(height int, workers int, fn func(model.RangeRow)) {
	ranges := SplitRows(height, workers)
	if len(ranges) == 0 {
		return
	}

	jobs := make(chan model.RangeRow)

	var wg sync.WaitGroup
	for i := 0; i < len(ranges); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rowRange := range jobs {
				fn(rowRange)
			}
		}()
	}

	for _, rowRange := range ranges {
		jobs <- rowRange
	}
	close(jobs)

	wg.Wait()

}
