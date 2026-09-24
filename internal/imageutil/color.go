package imageutil

import (
	"image"
	"image/color"
	"math"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
)

const MAX_RGB = 255.0

func SampleColor(img image.Image, x, y int) model.ColorF {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width == 0 || height == 0 {
		return model.ColorF{}
	}

	x = clampInt(x, 0, width-1)
	y = clampInt(y, 0, height-1)
	c := color.NRGBAModel.Convert(img.At(bounds.Min.X+x, bounds.Min.Y+y)).(color.NRGBA)

	return model.ColorF{
		R: float64(c.R) / MAX_RGB,
		G: float64(c.G) / MAX_RGB,
		B: float64(c.B) / MAX_RGB,
		A: float64(c.A) / MAX_RGB,
	}
}

func ToNRGBA(c model.ColorF) color.NRGBA {
	return color.NRGBA{
		R: uint8(math.Round(Clamp(c.R, 0, 1) * MAX_RGB)),
		G: uint8(math.Round(Clamp(c.G, 0, 1) * MAX_RGB)),
		B: uint8(math.Round(Clamp(c.B, 0, 1) * MAX_RGB)),
		A: uint8(math.Round(Clamp(c.A, 0, 1) * MAX_RGB)),
	}
}

func Clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func clampInt(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
