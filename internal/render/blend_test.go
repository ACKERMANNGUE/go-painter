package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
)

func TestBlendPixelMatchesImageAPIs(t *testing.T) {
	bounds := image.Rect(2, 3, 6, 7)
	actualParent := image.NewRGBA(image.Rect(0, 0, 10, 10))
	expectedParent := image.NewRGBA(image.Rect(0, 0, 10, 10))
	actual := actualParent.SubImage(bounds).(*image.RGBA)
	expected := expectedParent.SubImage(bounds).(*image.RGBA)

	initialPixels := []struct {
		x, y  int
		color color.RGBA
	}{
		{x: 2, y: 3, color: color.RGBA{R: 30, G: 50, B: 10, A: 100}},
		{x: 4, y: 5, color: color.RGBA{R: 40, G: 20, B: 60, A: 120}},
		{x: 5, y: 6, color: color.RGBA{R: 10, G: 10, B: 10, A: 40}},
	}
	for _, pixel := range initialPixels {
		actualParent.SetRGBA(pixel.x, pixel.y, pixel.color)
		expectedParent.SetRGBA(pixel.x, pixel.y, pixel.color)
	}

	calls := []struct {
		x, y    int
		color   model.ColorF
		opacity float64
	}{
		{x: 2, y: 3, color: model.ColorF{R: 1, G: 0.1, B: 0.4, A: 0.7}, opacity: 0.6},
		{x: 4, y: 5, color: model.ColorF{R: 0.2, G: 0.8, B: 0.3, A: 0.5}, opacity: 1.3},
		{x: 5, y: 6, color: model.ColorF{R: 0.1, G: 0.9, B: 1, A: 1}, opacity: -0.3},
		{x: 1, y: 3, color: model.ColorF{R: 1, G: 1, B: 1, A: 1}, opacity: 1},
		{x: 6, y: 7, color: model.ColorF{R: 1, G: 1, B: 1, A: 1}, opacity: 1},
	}

	for _, call := range calls {
		BlendPixel(actual, call.x, call.y, call.color, call.opacity)
		blendPixelWithImageAPIs(expected, call.x, call.y, call.color, call.opacity)
	}

	if string(actualParent.Pix) != string(expectedParent.Pix) {
		t.Fatal("direct pixel access differs from the image API behavior")
	}
}

func blendPixelWithImageAPIs(canvas *image.RGBA, x, y int, source model.ColorF, opacity float64) {
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
	canvas.Set(x, y, imageutil.ToNRGBA(Blend(destination, source, opacity)))
}
