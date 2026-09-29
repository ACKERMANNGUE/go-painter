package render

import (
	"image"
	"image/color"
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
