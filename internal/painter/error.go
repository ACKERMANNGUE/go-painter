package painter

import (
	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
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

func BuildErrorMap(source image.Image, canvas image.Image) model.ErrorMap {
	width := source.Bounds().Dx()
	height := source.Bounds().Dy()

	errorMap := model.ErrorMap{
		Width:  width,
		Height: height,
		Values: make([]float32, width*height),
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sourceColor := imageutil.SampleColor(source, x, y)
			canvasColor := imageutil.SampleColor(canvas, x, y)
			errorMap.Values[y*width+x] = float32(ColorDistance(sourceColor, canvasColor))
		}
	}

	return errorMap
}

func RegionErrorCached(errorMap model.ErrorMap, centerX int, centerY int, radius int) float64 {
	totalError := 0.0
	pixelCount := 0

	for y := centerY - radius; y < centerY+radius; y++ {
		for x := centerX - radius; x < centerX+radius; x++ {
			if y >= errorMap.Height || y < 0 || x >= errorMap.Width || x < 0 {
				continue
			}

			totalError += float64(errorMap.Values[y*errorMap.Width+x])
			pixelCount++
		}
	}

	if pixelCount == 0 {
		return 0.0
	}

	return totalError / float64(pixelCount)
}
