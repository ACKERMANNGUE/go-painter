package render

import (
	"image"
	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
)

func Blend(destination, source model.ColorF, alpha float64) model.ColorF {
	alpha = imageutil.Clamp(alpha * source.A, 0, 1)
	inverse := 1 - alpha

	return model.ColorF{
		R: destination.R * inverse + source.R * alpha,
		G: destination.G * inverse + source.G * alpha,
		B: destination.B * inverse + source.B * alpha,
		A: 1,
	}
}

func BlendPixel(canvas *image.RGBA, x int, y int, c model.ColorF, opacity float64) {
	if !image.Pt(x, y).In(canvas.Bounds()) {
		return
	}
	destination := imageutil.SampleColor(canvas, x, y)
	result := Blend(destination, c, opacity)
	canvas.Set(x, y, imageutil.ToNRGBA(result))
}