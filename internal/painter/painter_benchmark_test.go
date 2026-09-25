package painter

import (
	"image"
	"image/color"
	"testing"
)

func BenchmarkPaintSequential(b *testing.B) {
	source := benchmarkImage(128, 128)
	config, exists := GetPreset("oil")

	if exists != true {
		b.Fatal("Selected preset [oil] doesn't exist....")
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		engine, errorNew := New(config, 1)
		if errorNew != nil {
			b.Fatal(errorNew)
		}

		_, errorPaint := engine.Paint(source)

		if errorPaint != nil {
			b.Fatal(errorPaint)
		}
	}
}

func benchmarkImage(width int, height int) *image.RGBA {
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
