package render

import (
	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"image"
	"math"
)

func DrawDisc(canvas *image.RGBA, center model.Vec2, radius float64, color model.ColorF, opacity float64) {
	if radius <= 0 {
		return
	}

	minX := int(math.Floor(center.X - radius))
	maxX := int(math.Ceil(center.X + radius))
	minY := int(math.Floor(center.Y - radius))
	maxY := int(math.Ceil(center.Y + radius))

	bounds := canvas.Bounds()
	minX = max(minX, bounds.Min.X)
	maxX = min(maxX, bounds.Max.X-1)
	minY = max(minY, bounds.Min.Y)
	maxY = min(maxY, bounds.Max.Y-1)
	if minX > maxX || minY > maxY {
		return
	}

	radiusSquared := radius * radius

	for y := minY; y <= maxY; y++ {
		offset := canvas.PixOffset(minX, y)
		for x := minX; x <= maxX; x++ {
			dx := float64(x) - center.X
			dy := float64(y) - center.Y
			distanceSquared := dx*dx + dy*dy
			if distanceSquared <= radiusSquared {
				blendPixelAtOffset(canvas, offset, color, opacity)
			}
			offset += 4
		}
	}
}

func Lerp(a model.Vec2, b model.Vec2, t float64) model.Vec2 {
	return model.Vec2{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
	}
}

func DrawStroke(canvas *image.RGBA, stroke model.BrushStroke) {
	directionX := math.Cos(stroke.Angle)
	directionY := math.Sin(stroke.Angle)
	halfLength := stroke.Length / 2.0

	start := model.Vec2{
		X: stroke.Position.X - directionX*halfLength,
		Y: stroke.Position.Y - directionY*halfLength}

	end := model.Vec2{
		X: stroke.Position.X + directionX*halfLength,
		Y: stroke.Position.Y + directionY*halfLength}

	steps := math.Ceil(stroke.Length)

	for i := 0; i < int(steps); i++ {
		t := float64(i) / steps
		point := Lerp(start, end, t)
		DrawDisc(canvas, point, stroke.Width/2, stroke.Color, stroke.Opacity)
	}
}

func DrawCurvedStroke(canvas *image.RGBA, stroke model.CurvedStroke) {
	if len(stroke.Points) == 0 || stroke.Width <= 0 {
		return
	}
	if len(stroke.Points) == 1 {
		DrawDisc(canvas, stroke.Points[0], stroke.Width/2, stroke.Color, stroke.Opacity)
		return
	}

	for i := 0; i < len(stroke.Points)-1; i++ {
		a := stroke.Points[i]
		b := stroke.Points[i+1]
		length := math.Hypot(b.X-a.X, b.Y-a.Y)
		stepLength := math.Max(1, stroke.Width*0.25)
		steps := int(math.Ceil(length / stepLength))
		if steps < 1 {
			steps = 1
		}
		for j := 0; j <= steps; j++ {
			t := float64(j) / float64(steps)
			DrawDisc(canvas, Lerp(a, b, t), stroke.Width/2, stroke.Color, stroke.Opacity)
		}
	}
}
