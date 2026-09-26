package gradient

import (
	"math"

	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"github.com/ACKERMANNGUE/go-painter/internal/parallel"
)

var SOBEL_X = [3][3]float64{
	{-1, 0, 1},
	{-2, 0, 2},
	{-1, 0, 1},
}

var SOBEL_Y = [3][3]float64{
	{-1, -2, -1},
	{0, 0, 0},
	{1, 2, 1},
}

func Sobel(source model.GrayImage) model.GradientField {
	field := newGradientField(source)

	for y := 1; y < source.Height-1; y++ {
		for x := 1; x < source.Width-1; x++ {
			gx := 0.0
			gy := 0.0

			for kernelY := -1; kernelY <= 1; kernelY++ {
				for kernelX := -1; kernelX <= 1; kernelX++ {
					pixel := source.Data[(y+kernelY)*source.Width+(x+kernelX)]
					gx += pixel * SOBEL_X[kernelY+1][kernelX+1]
					gy += pixel * SOBEL_Y[kernelY+1][kernelX+1]
				}
			}

			magnitude := math.Sqrt(math.Pow(gx, 2) + math.Pow(gy, 2))
			angle := math.Atan2(gy, gx)
			field.Data[y*source.Width+x] = model.Gradient{DX: gx, DY: gy, Magnitude: magnitude, Angle: angle}
		}
	}

	return field
}

func newGradientField(source model.GrayImage) model.GradientField {
	return model.GradientField{
		Width:  source.Width,
		Height: source.Height,
		Data:   make([]model.Gradient, source.Height*source.Width),
	}
}

func SobelParallel(source model.GrayImage, workers int) model.GradientField {
	field := newGradientField(source)

	parallel.ForEachRow(source.Height, workers, func(y int) {
		if y == 0 || y == source.Height-1 {
			return
		}

		for x := 1; x < source.Width-1; x++ {
			gx := 0.0
			gy := 0.0

			for kernelY := -1; kernelY <= 1; kernelY++ {
				for kernelX := -1; kernelX <= 1; kernelX++ {
					pixel := source.Data[(y+kernelY)*source.Width+(x+kernelX)]
					gx += pixel * SOBEL_X[kernelY+1][kernelX+1]
					gy += pixel * SOBEL_Y[kernelY+1][kernelX+1]
				}
			}

			magnitude := math.Sqrt(math.Pow(gx, 2) + math.Pow(gy, 2))
			angle := math.Atan2(gy, gx)
			field.Data[y*source.Width+x] = model.Gradient{DX: gx, DY: gy, Magnitude: magnitude, Angle: angle}
		}
	})

	return field
}
