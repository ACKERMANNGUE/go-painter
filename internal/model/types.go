package model

type Vec2 struct {
	X float64
	Y float64
}

type ColorF struct {
	R float64
	G float64
	B float64
	A float64
}

type Gradient struct {
	DX        float64
	DY        float64
	Magnitude float64
	Angle     float64
}

type BrushStroke struct {
	Position Vec2
	Length   float64
	Width    float64
	Color    ColorF
	Opacity  float64
	Angle    float64
}

type CurvedStroke struct {
	Points  []Vec2
	Width   float64
	Color   ColorF
	Opacity float64
}

type GrayImage struct {
	Width  int
	Height int
	Data   []float64
}

type GradientField struct {
	Width  int
	Height int
	Data   []Gradient
}

type RangeRow struct {
	Start int
	End   int
}
