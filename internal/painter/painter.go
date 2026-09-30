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

	canvas := render.NewCanvasFromImage(source)
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

func generateStrokeBatch(buffers *model.PaintBuffer, brushSize int, config PainterConfig, rng *rand.Rand, onProgress func(int, int)) {
	buffers.StrokeBatch = generateStrokes(strokeGenerationRequest{
		source: buffers.Reference, canvas: buffers.Canvas, field: buffers.Field,
		brushSize: brushSize, config: config, rng: rng, batch: buffers.StrokeBatch,
		onProgress: onProgress,
	})
}

func renderStrokeBatch(buffers *model.PaintBuffer, config PainterConfig, onStroke func() error, onProgress func(int, int)) error {
	for index, stroke := range buffers.StrokeBatch {
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
		if onStroke != nil {
			if err := onStroke(); err != nil {
				return fmt.Errorf("capture drawing step: %w", err)
			}
		}
		if onProgress != nil {
			onProgress(index+1, len(buffers.StrokeBatch))
		}
	}
	return nil
}

func clearStrokeBatch(buffers *model.PaintBuffer) {
	buffers.StrokeBatch = buffers.StrokeBatch[:0]
	buffers.CurvePoints = buffers.CurvePoints[:0]
}

func (p *Painter) Paint(source image.Image, buffers *model.PaintBuffer) (*image.RGBA, error) {
	return p.PaintWithStrokeObserver(source, buffers, nil)
}

func (p *Painter) PaintWithStrokeObserver(source image.Image, buffers *model.PaintBuffer, onStroke func() error) (*image.RGBA, error) {
	return p.PaintWithProgress(source, buffers, onStroke, nil)
}

func (p *Painter) PaintWithProgress(source image.Image, buffers *model.PaintBuffer, onStroke func() error, onProgress func(PaintProgress)) (*image.RGBA, error) {
	reference := buildReferenceImage(source, p, buffers)
	buildGradientField(reference, p, buffers)
	if onStroke != nil {
		if err := onStroke(); err != nil {
			return nil, fmt.Errorf("capture initial drawing step: %w", err)
		}
	}

	for index, brushSize := range p.Config.BrushSizes {
		pass := index + 1
		passes := len(p.Config.BrushSizes)
		var generationProgress func(int, int)
		var renderingProgress func(int, int)
		if onProgress != nil {
			generationProgress = func(completed, total int) {
				onProgress(PaintProgress{Phase: PaintPhaseGenerate, Pass: pass, Passes: passes, Completed: completed, Total: total})
			}
			renderingProgress = func(completed, total int) {
				onProgress(PaintProgress{Phase: PaintPhaseRender, Pass: pass, Passes: passes, Completed: completed, Total: total})
			}
		}
		generateStrokeBatch(buffers, brushSize, p.Config, p.rng, generationProgress)
		shuffleStrokes(buffers.StrokeBatch, p.rng)
		if err := renderStrokeBatch(buffers, p.Config, onStroke, renderingProgress); err != nil {
			return nil, err
		}
		clearStrokeBatch(buffers)
	}

	return buffers.Canvas, nil
}
