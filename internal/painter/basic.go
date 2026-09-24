package painter

import (
	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"github.com/ACKERMANNGUE/go-painter/internal/render"
	"image"
	"math"
	rand "math/rand/v2"
)

func PaintBasic(source image.Image, config PainterConfig, seed uint64) *image.RGBA {
	bounds := source.Bounds()
	canvas := render.NewCanvas(bounds.Dx(), bounds.Dy(), model.ColorF{R: 1, G: 1, B: 1, A: 1})
	brushSize := 16

	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))

	for y := brushSize / 2; y < bounds.Dy(); y += brushSize {
		for x := brushSize / 2; x < bounds.Dx(); x += brushSize {
			brushSize = config.BrushSizes[int(RandomRange(0.0, float64(len(config.BrushSizes)-1), seed))]
			stroke := model.BrushStroke{
				Position: model.Vec2{X: float64(x), Y: float64(y)},
				Length:   float64(brushSize) * 1.5,
				Width:    float64(brushSize),
				Angle:    rng.Float64() * math.Pi,
				Color:    imageutil.SampleColor(source, x, y),
				Opacity:  config.Opacity,
			}
			render.DrawStroke(canvas, stroke)
		}
	}
	return canvas
}

func RandomRange(min float64, max float64, seed uint64) float64 {
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	return min + rng.Float64()*(max-min)
}
