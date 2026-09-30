package render

import (
	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"image"
)

const MAX_RGB float64 = 255.0

func Blend(destination, source model.ColorF, alpha float64) model.ColorF {
	alpha = imageutil.Clamp(alpha*source.A, 0, 1)
	inverse := 1 - alpha

	return model.ColorF{
		R: destination.R*inverse + source.R*alpha,
		G: destination.G*inverse + source.G*alpha,
		B: destination.B*inverse + source.B*alpha,
		A: 1,
	}
}

func BlendPixel(canvas *image.RGBA, x int, y int, c model.ColorF, opacity float64) {
	bounds := canvas.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}
	blendPixelAtOffset(canvas, canvas.PixOffset(x, y), c, opacity)
}

func blendPixelAtOffset(canvas *image.RGBA, offset int, c model.ColorF, opacity float64) {
	destinationAlpha := canvas.Pix[offset+3]
	if destinationAlpha == 0 {
		return
	}
	if destinationAlpha == 255 {
		destination := model.ColorF{
			R: float64(canvas.Pix[offset]) / MAX_RGB,
			G: float64(canvas.Pix[offset+1]) / MAX_RGB,
			B: float64(canvas.Pix[offset+2]) / MAX_RGB,
			A: 1,
		}
		pixel := imageutil.ToNRGBA(Blend(destination, c, opacity))
		canvas.Pix[offset] = pixel.R
		canvas.Pix[offset+1] = pixel.G
		canvas.Pix[offset+2] = pixel.B
		return
	}

	alpha := float64(destinationAlpha) / MAX_RGB
	destination := model.ColorF{
		R: float64(canvas.Pix[offset]) / MAX_RGB / alpha,
		G: float64(canvas.Pix[offset+1]) / MAX_RGB / alpha,
		B: float64(canvas.Pix[offset+2]) / MAX_RGB / alpha,
		A: 1,
	}
	result := Blend(destination, c, opacity)
	pixel := imageutil.ToNRGBA(result)
	canvas.Pix[offset] = premultiply(pixel.R, destinationAlpha)
	canvas.Pix[offset+1] = premultiply(pixel.G, destinationAlpha)
	canvas.Pix[offset+2] = premultiply(pixel.B, destinationAlpha)
}

func premultiply(channel, alpha uint8) uint8 {
	return uint8((uint16(channel)*uint16(alpha) + 127) / 255)
}
