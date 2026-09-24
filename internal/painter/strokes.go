package painter

import (
	"image"
	"math"
	rand "math/rand/v2"

	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
)

func GenerateStrokes(
	source image.Image,
	canvas image.Image,
	field model.GradientField,
	brushSize int,
	config PainterConfig,
	rng *rand.Rand,
) []model.BrushStroke {
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	step := brushSize
	if step < 1 {
		step = 1
	}
	radius := brushSize / 2
	if radius < 1 {
		radius = 1
	}

	strokes := make([]model.BrushStroke, 0, (width/step+1)*(height/step+1)/2)
	for y := step / 2; y < height; y += step {
		for x := step / 2; x < width; x += step {
			if RegionError(source, canvas, x, y, radius) < config.ErrorThreshold {
				continue
			}

			jitter := float64(brushSize) * config.Randomness
			sx := clampFloat(float64(x)+randomRange(rng, -jitter, jitter), 0, float64(width-1))
			sy := clampFloat(float64(y)+randomRange(rng, -jitter, jitter), 0, float64(height-1))
			sampleX, sampleY := int(math.Round(sx)), int(math.Round(sy))
			g := field.Data[int(sampleX+source.Bounds().Dx()*sampleY)]

			angle := g.Angle + math.Pi/2
			if g.Magnitude < 0.015 {
				// Flat regions do not have a stable contour direction, so a small
				// randomized angle avoids large blocks of perfectly parallel marks.
				angle = rng.Float64() * math.Pi
			}
			angle += randomRange(rng, -0.18, 0.18) * config.Randomness

			lengthVariation := 1 + randomRange(rng, -0.20, 0.20)*config.Randomness
			widthVariation := 1 + randomRange(rng, -0.14, 0.14)*config.Randomness
			stroke := model.BrushStroke{
				Position: model.Vec2{X: sx, Y: sy},
				Length:   float64(brushSize) * config.StrokeLength * lengthVariation,
				Width:    math.Max(1, float64(brushSize)*0.72*widthVariation),
				Angle:    angle,
				Color:    imageutil.SampleColor(source, sampleX, sampleY),
				Opacity:  config.Opacity,
			}
			strokes = append(strokes, stroke)
		}
	}
	return strokes
}

func DrawStroke(
	start model.Vec2,
	field model.GradientField,
	length float64,
	stepLength float64,
	smoothing float64,
) []model.Vec2 {
	if length <= 0 {
		return []model.Vec2{start}
	}
	if stepLength <= 0 {
		stepLength = 1
	}
	smoothing = clampFloat(smoothing, 0, 1)

	steps := int(math.Max(1, math.Ceil(length/stepLength)))
	points := make([]model.Vec2, 0, steps+1)
	points = append(points, start)

	current := start
	previousDirection := model.Vec2{X: 1, Y: 0}
	hasPreviousDirection := false

	for i := 0; i < steps; i++ {
		gradientSample, ok := sampleGradient(field, current)
		if !ok {
			break
		}

		// The local stroke direction follows image isophotes, i.e. it is
		// perpendicular to the image gradient.
		direction := model.Vec2{X: -gradientSample.DY, Y: gradientSample.DX}
		directionMagnitude := math.Hypot(direction.X, direction.Y)
		if directionMagnitude < 1e-8 {
			if !hasPreviousDirection {
				break
			}
			direction = previousDirection
		} else {
			direction.X /= directionMagnitude
			direction.Y /= directionMagnitude
		}

		if hasPreviousDirection {
			direction.X = previousDirection.X*smoothing + direction.X*(1-smoothing)
			direction.Y = previousDirection.Y*smoothing + direction.Y*(1-smoothing)
			blendedMagnitude := math.Hypot(direction.X, direction.Y)
			if blendedMagnitude > 1e-8 {
				direction.X /= blendedMagnitude
				direction.Y /= blendedMagnitude
			}
		}

		next := model.Vec2{X: current.X + direction.X*stepLength, Y: current.Y + direction.Y*stepLength}
		if next.X < 0 || next.Y < 0 || next.X >= float64(field.Width) || next.Y >= float64(field.Height) {
			break
		}

		points = append(points, next)
		current = next
		previousDirection = direction
		hasPreviousDirection = true
	}

	if len(points) == 1 {
		fallback := model.Vec2{X: clampFloat(start.X+stepLength, 0, float64(field.Width-1)), Y: start.Y}
		points = append(points, fallback)
	}

	return points
}

func shuffleStrokes(strokes []model.BrushStroke, rng *rand.Rand) {
	rng.Shuffle(len(strokes), func(i, j int) {
		strokes[i], strokes[j] = strokes[j], strokes[i]
	})
}

func randomRange(rng *rand.Rand, minimum, maximum float64) float64 {
	return minimum + rng.Float64()*(maximum-minimum)
}

func clampFloat(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func sampleGradient(field model.GradientField, position model.Vec2) (model.Gradient, bool) {
	x := int(math.Round(position.X))
	y := int(math.Round(position.Y))
	if x < 0 || y < 0 || x >= field.Width || y >= field.Height {
		return model.Gradient{}, false
	}
	return field.Data[y*field.Width+x], true
}
