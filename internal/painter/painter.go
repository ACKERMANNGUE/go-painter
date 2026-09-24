package painter

import (
	"fmt"
	"image"
	"math"
	rand "math/rand/v2"

	"github.com/ACKERMANNGUE/go-painter/internal/gradient"
	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"github.com/ACKERMANNGUE/go-painter/internal/render"
)

type Painter struct {
	Config PainterConfig
	rng    *rand.Rand
}

func New(config PainterConfig, seed uint64) (*Painter, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &Painter{
		Config: config,
		rng:    rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)),
	}, nil
}

func (p *Painter) Paint(source image.Image) (*image.RGBA, error) {
	bounds := source.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return nil, fmt.Errorf("source image is empty")
	}

	canvas := render.NewCanvas(bounds.Dx(), bounds.Dy(), model.ColorF{R: 1, G: 1, B: 1, A: 1})

	for _, brushSize := range p.Config.BrushSizes {
		radius := max(1, brushSize/4)
		sigma := math.Max(0.8, float64(radius)*0.75)
		reference := imageutil.GaussianBlur(source, radius, sigma)
		gray := gradient.ToGrayscale(reference)
		field := gradient.Sobel(gray)

		strokes := GenerateStrokes(reference, canvas, field, brushSize, p.Config, p.rng)
		shuffleStrokes(strokes, p.rng)

		for _, stroke := range strokes {
			if p.Config.UseCurvedStrokes {
				points := DrawStroke(
					stroke.Position,
					field,
					stroke.Length,
					math.Max(1, stroke.Width*0.35),
					p.Config.CurveSmoothing,
				)
				render.DrawCurvedStroke(canvas, model.CurvedStroke{
					Points:  points,
					Width:   stroke.Width,
					Color:   stroke.Color,
					Opacity: stroke.Opacity,
				})
			} else {
				render.DrawStroke(canvas, stroke)
			}
		}
	}

	return canvas, nil
}
