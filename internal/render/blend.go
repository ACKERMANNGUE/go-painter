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
	if !image.Pt(x, y).In(canvas.Bounds()) {
		return
	}
	pixel := canvas.RGBAAt(x, y)
	destination := model.ColorF{
		R: float64(pixel.R) / MAX_RGB,
		G: float64(pixel.G) / MAX_RGB,
		B: float64(pixel.B) / MAX_RGB,
		A: 1,
	}
	result := Blend(destination, c, opacity)
	canvas.Set(x, y, imageutil.ToNRGBA(result))
}
