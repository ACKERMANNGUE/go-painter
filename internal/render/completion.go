package render

import (
	"image"

	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
)

func CompleteUnpaintedPixels(canvas *image.RGBA, source image.Image) int {
	bounds := canvas.Bounds()
	corrected := 0
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			offset := canvas.PixOffset(bounds.Min.X+x, bounds.Min.Y+y)
			alpha := canvas.Pix[offset+3]
			if alpha == 0 || !isUntouchedPixel(canvas.Pix[offset], canvas.Pix[offset+1], canvas.Pix[offset+2], alpha) {
				continue
			}

			sourceColor := imageutil.SampleColor(source, x, y)
			if sourceColor.R == 1 && sourceColor.G == 1 && sourceColor.B == 1 {
				continue // the source itself is white here, so there is nothing to fix
			}

			straight := imageutil.ToNRGBA(sourceColor)
			canvas.Pix[offset] = premultiply(straight.R, alpha)
			canvas.Pix[offset+1] = premultiply(straight.G, alpha)
			canvas.Pix[offset+2] = premultiply(straight.B, alpha)
			corrected++
		}
	}
	return corrected
}

func isUntouchedPixel(r, g, b, alpha uint8) bool {
	return r == alpha && g == alpha && b == alpha
}
