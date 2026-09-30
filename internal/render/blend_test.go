package render

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/ACKERMANNGUE/go-painter/internal/model"
)

func TestBlendPixelPreservesAlphaMask(t *testing.T) {
	source := image.NewNRGBA(image.Rect(4, 7, 8, 8))
	alphas := []uint8{0, 64, 128, 255}
	for x, alpha := range alphas {
		source.SetNRGBA(4+x, 7, color.NRGBA{R: 90, G: 80, B: 70, A: alpha})
	}
	canvas := NewCanvasFromImage(source)
	initialColors := make([]color.RGBA, len(alphas))
	for x, alpha := range alphas {
		initialColors[x] = canvas.RGBAAt(x, 0)
		want := color.RGBA{R: alpha, G: alpha, B: alpha, A: alpha}
		if initialColors[x] != want {
			t.Errorf("initial pixel %d = %#v, want white with source alpha %#v", x, initialColors[x], want)
		}
	}

	paint := model.ColorF{R: 0.2, G: 0.8, B: 0.4, A: 1}
	for x := range alphas {
		BlendPixel(canvas, x, 0, paint, 1)
	}
	BlendPixel(canvas, 1, 0, model.ColorF{R: 1, G: 0, B: 0, A: 1}, 1)
	BlendPixel(canvas, 8, 0, paint, 1)

	for x, wantAlpha := range alphas {
		got := canvas.RGBAAt(x, 0)
		if got.A != wantAlpha {
			t.Errorf("pixel %d alpha = %d, want %d", x, got.A, wantAlpha)
		}
		if wantAlpha == 0 && got != initialColors[x] {
			t.Errorf("transparent pixel changed from %#v to %#v", initialColors[x], got)
		}
	}
}

func TestDrawDiscMatchesPixelBlendingAtCanvasEdges(t *testing.T) {
	bounds := image.Rect(3, 4, 8, 9)
	centers := []model.Vec2{{X: 3.2, Y: 4.4}, {X: 6.8, Y: 7.5}, {X: 1, Y: 1}, {X: 5.5, Y: 5.5}}
	paint := model.ColorF{R: 0.2, G: 0.8, B: 0.4, A: 1}

	for _, center := range centers {
		got := opaqueCanvas(bounds)
		want := opaqueCanvas(bounds)
		DrawDisc(got, center, 2.3, paint, 0.65)
		drawDiscReference(want, center, 2.3, paint, 0.65)
		if !bytes.Equal(got.Pix, want.Pix) {
			t.Errorf("DrawDisc(%+v) differs from per-pixel blending", center)
		}
	}
}

func opaqueCanvas(bounds image.Rectangle) *image.RGBA {
	canvas := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			canvas.SetRGBA(x, y, color.RGBA{R: 180, G: 120, B: 90, A: 255})
		}
	}
	return canvas
}

func drawDiscReference(canvas *image.RGBA, center model.Vec2, radius float64, paint model.ColorF, opacity float64) {
	minX := int(math.Floor(center.X - radius))
	maxX := int(math.Ceil(center.X + radius))
	minY := int(math.Floor(center.Y - radius))
	maxY := int(math.Ceil(center.Y + radius))
	radiusSquared := math.Pow(radius, 2)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x) - center.X
			dy := float64(y) - center.Y
			if math.Pow(dx, 2)+math.Pow(dy, 2) <= radiusSquared {
				BlendPixel(canvas, x, y, paint, opacity)
			}
		}
	}
}
