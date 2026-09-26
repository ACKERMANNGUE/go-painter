package imageutil

import (
	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"image"
	"image/color"
	"math"
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
	imageX, imageY := bounds.Min.X+x, bounds.Min.Y+y
	switch img := img.(type) {
	case *image.RGBA:
		offset := img.PixOffset(imageX, imageY)
		alpha := img.Pix[offset+3]
		if alpha == 0 {
			return model.ColorF{}
		}
		return model.ColorF{
			R: float64(unpremultiplyByte(img.Pix[offset], alpha)) / MAX_RGB,
			G: float64(unpremultiplyByte(img.Pix[offset+1], alpha)) / MAX_RGB,
			B: float64(unpremultiplyByte(img.Pix[offset+2], alpha)) / MAX_RGB,
			A: float64(alpha) / MAX_RGB,
		}
	case *image.NRGBA:
		offset := img.PixOffset(imageX, imageY)
		return model.ColorF{
			R: float64(img.Pix[offset]) / MAX_RGB,
			G: float64(img.Pix[offset+1]) / MAX_RGB,
			B: float64(img.Pix[offset+2]) / MAX_RGB,
			A: float64(img.Pix[offset+3]) / MAX_RGB,
		}
	default:
		c := color.NRGBAModel.Convert(img.At(imageX, imageY)).(color.NRGBA)
		return model.ColorF{
			R: float64(c.R) / MAX_RGB,
			G: float64(c.G) / MAX_RGB,
			B: float64(c.B) / MAX_RGB,
			A: float64(c.A) / MAX_RGB,
		}
	}
}

// unpremultiplyByte takes a color channel value and an alpha value, and returns the unpremultiplied color channel value
func unpremultiplyByte(channel, alpha uint8) uint8 {
	return uint8((uint32(channel) * 0xffff / uint32(alpha)) >> 8)
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
