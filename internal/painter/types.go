package painter

import (
	"github.com/ACKERMANNGUE/go-painter/internal/model"
)

type Vec2 = model.Vec2
type ColorF = model.ColorF
type BrushStroke = model.BrushStroke
type CurvedStroke = model.CurvedStroke
type Gradient = model.Gradient

type PainterConfig struct {
	BrushSizes       []int
	ErrorThreshold   float64
	StrokeLength     float64
	Opacity          float64
	Randomness       float64
	UseCurvedStrokes bool
	CurveSmoothing   float64
	Workers          int
}
