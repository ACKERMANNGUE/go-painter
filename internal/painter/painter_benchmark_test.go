package painter

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"testing"

	"github.com/ACKERMANNGUE/go-painter/internal/model"
)

func BenchmarkPaint(b *testing.B) {
	source := benchmarkImage(1280, 1280)
	var buffers model.PaintBuffer
	baseConfig, ok := GetPreset("oil")
	if !ok {
		b.Fatal("oil preset is missing")
	}

	workerCounts := uniqueWorkerCounts([]int{1, 2, 4, 8, 16, 32, 64, runtime.GOMAXPROCS(0)})
	for _, workers := range workerCounts {
		b.Run(benchmarkName(workers), func(b *testing.B) {
			config := baseConfig
			config.Workers = workers
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				engine, err := New(config, 1)
				if err != nil {
					b.Fatal(err)
				}
				if _, err := engine.Paint(source, &buffers); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func benchmarkImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8((x * 255) / max(1, width-1)),
				G: uint8((y * 255) / max(1, height-1)),
				B: uint8(((x + y) * 255) / max(1, width+height-2)),
				A: 255,
			})
		}
	}
	return img
}

func benchmarkName(workers int) string {
	return fmt.Sprintf("workers_%d", workers)
}

func uniqueWorkerCounts(values []int) []int {
	seen := make(map[int]bool)
	result := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
