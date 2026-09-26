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

func buildReferenceImage(source image.Image, p *Painter, buffers *model.PaintBuffer) image.Image {
	bounds := source.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		fmt.Errorf("source image is empty")
		return image.Black
	}

	canvas := render.NewCanvas(bounds.Dx(), bounds.Dy(), model.ColorF{R: 1, G: 1, B: 1, A: 1})
	baseBrushSize := p.Config.BrushSizes[0]
	blurScale := p.Config.BlurStrength
	if blurScale <= 0 {
		blurScale = 1
	}
	radius := max(1, int(math.Round(float64(baseBrushSize)/4.0*blurScale)))
	sigma := math.Max(0.8, float64(radius)*0.75*blurScale)
	reference := imageutil.GaussianBlur(source, radius, sigma, p.Config.Workers)

	buffers.Reference = reference
	buffers.Canvas = canvas

	return reference
}

func buildGradientField(reference image.Image, p *Painter, buffers *model.PaintBuffer) /*model.GradientField*/ {
	gray := gradient.ToGrayscaleParallel(reference, p.Config.Workers)
	field := gradient.SobelParallel(gray, p.Config.Workers)

	buffers.Gray = gray
	buffers.Field = field

	// return field
}

func generateStrokeBatch(source image.Image, buffers *model.PaintBuffer, brushSize int, config PainterConfig, rng *rand.Rand) {
	buffers.StrokeBatch = GenerateStrokes(buffers.Reference, buffers.Canvas, buffers.Field, brushSize, config, rng, buffers.StrokeBatch)
}

func renderStrokeBatch(buffers *model.PaintBuffer, config PainterConfig) {
	for _, stroke := range buffers.StrokeBatch {
		if config.UseCurvedStrokes {
			buffers.CurvePoints = DrawStroke(
				stroke.Position,
				buffers.Field,
				stroke.Length,
				math.Max(1, stroke.Width*0.35),
				config.CurveSmoothing,
				buffers.CurvePoints,
			)

			render.DrawCurvedStroke(buffers.Canvas, model.CurvedStroke{
				Points:  buffers.CurvePoints,
				Width:   stroke.Width,
				Color:   stroke.Color,
				Opacity: stroke.Opacity,
			})
		} else {
			render.DrawStroke(buffers.Canvas, stroke)
		}
	}
}

func clearStrokeBatch(buffers *model.PaintBuffer) {
	buffers.StrokeBatch = buffers.StrokeBatch[:0]
	buffers.CurvePoints = buffers.CurvePoints[:0]
}

func (p *Painter) Paint(source image.Image, buffers *model.PaintBuffer) (*image.RGBA, error) {
	reference := buildReferenceImage(source, p, buffers)
	buildGradientField(reference, p, buffers)

	for _, brushSize := range p.Config.BrushSizes {
		generateStrokeBatch(source, buffers, brushSize, p.Config, p.rng)
		shuffleStrokes(buffers.StrokeBatch, p.rng)
		renderStrokeBatch(buffers, p.Config)
		clearStrokeBatch(buffers)
	}

	return buffers.Canvas, nil
}
