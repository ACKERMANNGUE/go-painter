package gradient

import (
	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"image"
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
