package gradient

import (
	"image"

	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"github.com/ACKERMANNGUE/go-painter/internal/parallel"
)

func ToGrayscale(img image.Image) model.GrayImage {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	maxValue := 65535.0
	redIntensity := 0.299
	greenIntensity := 0.587
	blueIntensity := 0.114

	grayImage := model.GrayImage{
		Width:  width,
		Height: height,
		Data:   make([]float64, width*height),
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			grayValue := redIntensity*float64(r)/maxValue + greenIntensity*float64(g)/maxValue + blueIntensity*float64(b)/maxValue
			grayImage.Data[y*width+x] = grayValue
		}
	}

	return grayImage
}

func ToGrayscaleParallel(img image.Image, workers int) model.GrayImage {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	grayImage := model.GrayImage{
		Width:  width,
		Height: height,
		Data:   make([]float64, width*height),
	}

	parallel.ForEachRange(height, workers, func(rowRange model.RangeRow) {
		GrayifyRange(img, &grayImage, rowRange)
	})

	return grayImage
}

func GrayifyRange(source image.Image, grayImage *model.GrayImage, rangeRow model.RangeRow) {
	maxValue := 65535.0
	redIntensity := 0.299
	greenIntensity := 0.587
	blueIntensity := 0.114
	width := source.Bounds().Dx()

	for y := rangeRow.Start; y < rangeRow.End; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := source.At(x, y).RGBA()
			grayValue := redIntensity*float64(r)/maxValue + greenIntensity*float64(g)/maxValue + blueIntensity*float64(b)/maxValue
			grayImage.Data[y*width+x] = grayValue
		}
	}
}
