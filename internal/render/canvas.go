package render

import (
	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"image"
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

func NewCanvasFromImage(source image.Image) *image.RGBA {
	bounds := source.Bounds()
	canvas := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			_, _, _, alpha := source.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			alphaByte := uint8(alpha >> 8)
			offset := canvas.PixOffset(x, y)
			canvas.Pix[offset] = alphaByte
			canvas.Pix[offset+1] = alphaByte
			canvas.Pix[offset+2] = alphaByte
			canvas.Pix[offset+3] = alphaByte
		}
	}
	return canvas
}
