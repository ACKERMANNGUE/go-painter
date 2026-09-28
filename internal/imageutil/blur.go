package imageutil

import (
	"image"
	"math"

	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"github.com/ACKERMANNGUE/go-painter/internal/parallel"
)

func GaussianKernel(radius int, sigma float64) []float64 {
	if radius < 1 {
		return []float64{1}
	}
	if sigma <= 0 {
		sigma = float64(radius) / 2
		if sigma <= 0 {
			sigma = 1
		}
	}

	kernel := make([]float64, radius*2+1)
	total := 0.0

	for x := -radius; x <= radius; x++ {
		value := math.Exp(-(math.Pow(float64(x), 2) / (2 * math.Pow(sigma, 2))))
		kernel[x+radius] = value
		total += value
	}

	for i := 0; i < len(kernel); i++ {
		kernel[i] /= total
	}

	return kernel
}

func GaussianBlur(source image.Image, radius int, sigma float64, workers int) image.Image {
	kernel := GaussianKernel(radius, sigma)

	temporary := ApplyHorizontalBlur(source, kernel, workers)
	blurred := ApplyVerticalBlur(temporary, kernel, workers)
	return blurred
}

func ApplyHorizontalBlur(source image.Image, kernel []float64, workers int) image.Image {
	bounds := source.Bounds()
	blurred := image.NewRGBA(bounds)
	radius := len(kernel) / 2

	parallel.ForEachRow(source.Bounds().Dy(), workers, func(y int) {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			blurred.Set(x, y, ToNRGBA(applyKernelAt(source, x, y, kernel, radius, true)))
		}
	})

	return blurred
}

func ApplyVerticalBlur(source image.Image, kernel []float64, workers int) image.Image {
	bounds := source.Bounds()
	blurred := image.NewRGBA(bounds)
	radius := len(kernel) / 2

	parallel.ForEachRow(source.Bounds().Dy(), workers, func(y int) {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			blurred.Set(x, y, ToNRGBA(applyKernelAt(source, x, y, kernel, radius, false)))
		}
	})

	return blurred
}

func applyKernelAt(source image.Image, x int, y int, kernel []float64, radius int, horizontal bool) model.ColorF {
	accumulated := model.ColorF{}
	alpha := 0.0
	for offset := -radius; offset <= radius; offset++ {
		sampleX := x
		sampleY := y
		if horizontal {
			sampleX = x + offset
		} else {
			sampleY = y + offset
		}

		weight := kernel[offset+radius]
		sample := SampleColor(source, sampleX, sampleY)
		weightedAlpha := sample.A * weight
		accumulated.R += sample.R * weightedAlpha
		accumulated.G += sample.G * weightedAlpha
		accumulated.B += sample.B * weightedAlpha
		alpha += weightedAlpha
	}

	if alpha == 0 {
		return model.ColorF{}
	}
	accumulated.R /= alpha
	accumulated.G /= alpha
	accumulated.B /= alpha
	accumulated.A = alpha
	return accumulated
}
