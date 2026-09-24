package render

import (
	"image"
	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
)

func NewCanvas(width int, height int, background model.ColorF) *image.RGBA {
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	backgroundPixel := imageutil.ToNRGBA(background)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			canvas.Set(x, y, backgroundPixel)
		}
	}

	return canvas
}