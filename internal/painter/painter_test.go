package painter

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/ACKERMANNGUE/go-painter/internal/model"
)

func TestPaintPreservesSourceAlpha(t *testing.T) {
	source := alphaTestImage()
	config := PainterConfig{
		BrushSizes:       []int{4},
		ErrorThreshold:   0,
		StrokeLength:     1,
		Opacity:          0.8,
		Randomness:       0.1,
		UseCurvedStrokes: false,
		CurveSmoothing:   0.3,
		BlurStrength:     0.5,
		Workers:          1,
	}
	engine, err := New(config, 42)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	result, err := engine.Paint(source, &model.PaintBuffer{})
	if err != nil {
		t.Fatalf("Paint() error = %v", err)
	}
	assertSameAlpha(t, source, result)
}

func TestRegionErrorCachedMatchesDirectCalculation(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 5, 4))
	canvas := image.NewRGBA(image.Rect(0, 0, 5, 4))
	for y := 0; y < source.Bounds().Dy(); y++ {
		for x := 0; x < source.Bounds().Dx(); x++ {
			source.SetRGBA(x, y, color.RGBA{R: uint8(x * 31), G: uint8(y * 47), B: 120, A: 255})
			canvas.SetRGBA(x, y, color.RGBA{R: uint8(x * 29), G: uint8(y * 43), B: 100, A: 255})
		}
	}

	got := RegionErrorCached(BuildErrorMap(source, canvas), 2, 2, 1)
	want := RegionError(source, canvas, 2, 2, 1)
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("RegionErrorCached() = %.9f, want %.9f", got, want)
	}
}

func TestPaintWithStrokeObserverCapturesInitialAndRenderedCanvases(t *testing.T) {
	source := alphaTestImage()
	config := PainterConfig{
		BrushSizes:       []int{4},
		ErrorThreshold:   0,
		StrokeLength:     1,
		Opacity:          0.8,
		Randomness:       0.1,
		UseCurvedStrokes: false,
		CurveSmoothing:   0.3,
		BlurStrength:     0.5,
		Workers:          1,
	}
	engine, err := New(config, 42)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	var buffers model.PaintBuffer
	var callbackCount int
	var initialFrame, latestFrame []byte
	result, err := engine.PaintWithStrokeObserver(source, &buffers, func() error {
		callbackCount++
		latestFrame = append(latestFrame[:0], buffers.Canvas.Pix...)
		if callbackCount == 1 {
			initialFrame = append([]byte(nil), latestFrame...)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("PaintWithStrokeObserver() error = %v", err)
	}
	if callbackCount < 2 {
		t.Fatalf("observer called %d times, want initial canvas and at least one stroke", callbackCount)
	}
	if bytes.Equal(initialFrame, latestFrame) {
		t.Fatal("initial and final observed canvases are identical")
	}
	if !bytes.Equal(latestFrame, result.Pix) {
		t.Fatal("last observed canvas does not match the final painting")
	}
}

func alphaTestImage() *image.NRGBA {
	source := image.NewNRGBA(image.Rect(3, 5, 19, 21))
	for y := 0; y < source.Bounds().Dy(); y++ {
		for x := 0; x < source.Bounds().Dx(); x++ {
			alpha := uint8(255)
			if x < 4 || y < 4 {
				alpha = 0
			} else if x == 4 || y == 4 {
				alpha = 96
			}
			source.SetNRGBA(source.Bounds().Min.X+x, source.Bounds().Min.Y+y, color.NRGBA{
				R: uint8(x * 13),
				G: uint8(y * 11),
				B: 120,
				A: alpha,
			})
		}
	}
	return source
}

func assertSameAlpha(t *testing.T, source image.Image, result image.Image) {
	t.Helper()
	canvas, ok := result.(*image.RGBA)
	if !ok {
		t.Fatalf("Paint() returned %T, want *image.RGBA", result)
	}
	for y := 0; y < source.Bounds().Dy(); y++ {
		for x := 0; x < source.Bounds().Dx(); x++ {
			_, _, _, sourceAlpha := source.At(source.Bounds().Min.X+x, source.Bounds().Min.Y+y).RGBA()
			_, _, _, resultAlpha := result.At(x, y).RGBA()
			if resultAlpha != sourceAlpha {
				t.Fatalf("alpha at (%d, %d) = %d, want %d", x, y, resultAlpha>>8, sourceAlpha>>8)
			}
			if sourceAlpha == 0 {
				offset := canvas.PixOffset(x, y)
				if canvas.Pix[offset] != 0 || canvas.Pix[offset+1] != 0 || canvas.Pix[offset+2] != 0 {
					t.Fatalf("transparent pixel at (%d, %d) retains RGB data", x, y)
				}
			}
		}
	}
}
