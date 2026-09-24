package painter

import (
	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"image"
	"math"
)

func ColorDistance(a ColorF, b ColorF) float64 {
	distanceR := a.R - b.R
	distanceG := a.G - b.G
	distanceB := a.B - b.B

	return math.Sqrt(math.Pow(distanceR, 2) + math.Pow(distanceG, 2) + math.Pow(distanceB, 2))
}

func RegionError(source image.Image, canvas image.Image, centerX int, centerY int, radius int) float64 {
	totalError := 0.0
	pixelCount := 0

	for y := centerY - radius; y < centerY+radius; y++ {
		for x := centerX - radius; x < centerX+radius; x++ {
			if y >= source.Bounds().Dy() || y < 0 || x >= source.Bounds().Dx() || x < 0 {
				continue
			}

			sourceColor := imageutil.SampleColor(source, x, y)
			canvasColor := imageutil.SampleColor(canvas, x, y)
			totalError += ColorDistance(sourceColor, canvasColor)
			pixelCount++
		}
	}

	if pixelCount == 0 {
		return 0.0
	}

	return totalError / float64(pixelCount)
}
