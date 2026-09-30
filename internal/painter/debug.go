package painter

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"path/filepath"
	"strings"

	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
)

func SaveDebugImages(outputPath string, source image.Image, buffers *model.PaintBuffer) error {
	if buffers == nil || buffers.Reference == nil || buffers.Canvas == nil {
		return fmt.Errorf("debug images require a completed painting")
	}

	images := []struct {
		suffix string
		image  image.Image
	}{
		{suffix: "grayscale", image: grayscaleDebugImage(buffers.Gray)},
		{suffix: "blur", image: buffers.Reference},
		{suffix: "sobel", image: sobelDebugImage(buffers.Field)},
		{suffix: "error_heatmap", image: errorHeatmap(source, buffers.Canvas)},
	}
	for _, item := range images {
		path := debugImagePath(outputPath, item.suffix)
		if err := imageutil.SavePNG(path, item.image); err != nil {
			return fmt.Errorf("save %s debug image: %w", item.suffix, err)
		}
		fmt.Printf("Debug image: %s\n", path)
	}
	return nil
}

func debugImagePath(outputPath, suffix string) string {
	base := strings.TrimSuffix(outputPath, filepath.Ext(outputPath))
	return base + "_" + suffix + ".png"
}

func grayscaleDebugImage(gray model.GrayImage) *image.Gray {
	output := image.NewGray(image.Rect(0, 0, gray.Width, gray.Height))
	for index, value := range gray.Data {
		output.Pix[index] = uint8(math.Round(imageutil.Clamp(value, 0, 1) * 255))
	}
	return output
}

func sobelDebugImage(field model.GradientField) *image.Gray {
	output := image.NewGray(image.Rect(0, 0, field.Width, field.Height))
	maxMagnitude := 0.0
	for _, gradient := range field.Data {
		maxMagnitude = math.Max(maxMagnitude, gradient.Magnitude)
	}
	if maxMagnitude == 0 {
		return output
	}
	for index, gradient := range field.Data {
		intensity := imageutil.Clamp(gradient.Magnitude/maxMagnitude, 0, 1)
		output.Pix[index] = uint8(math.Round(intensity * 255))
	}
	return output
}

func errorHeatmap(source, painted image.Image) *image.NRGBA {
	errorMap := BuildErrorMap(source, painted)
	output := image.NewNRGBA(image.Rect(0, 0, errorMap.Width, errorMap.Height))
	maxDistance := math.Sqrt(3)
	for index, distance := range errorMap.Values {
		intensity := imageutil.Clamp(float64(distance)/maxDistance, 0, 1)
		red := uint8(math.Round(intensity * 255))
		green := uint8(math.Round(math.Max(0, 1-math.Abs(intensity*2-1)) * 220))
		blue := uint8(math.Round((1 - intensity) * 255))
		output.SetNRGBA(index%errorMap.Width, index/errorMap.Width, color.NRGBA{R: red, G: green, B: blue, A: 255})
	}
	return output
}
