package gradient

import (
	"fmt"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"image"
	"sync"
)

func ToGrayscale(img image.Image) model.GrayImage {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	maxValue := 65535.0
	redIntensity := 0.299
	greenIntensity := 0.587
	blueIntensity := 0.114

	grayImage := model.GrayImage{
		Width:  width,
		Height: height,
		Data:   make([]float64, width*height),
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			grayValue := redIntensity*float64(r)/maxValue + greenIntensity*float64(g)/maxValue + blueIntensity*float64(b)/maxValue
			grayImage.Data[y*width+x] = grayValue
		}
	}

	return grayImage
}

func splitRows(height int, workers int) []model.Vec2 {
	if workers <= 0 {
		workers = 1
	}
	if workers > height {
		workers = height
	}

	rows := make([]model.Vec2, 0, workers)
	start := 0
	end := 0

	for i := 0; i < workers; i++ {
		start = i * height / workers
		end = (i + 1) * height / workers
		fmt.Printf("start %v, end %v\n", start, end)
		rows = append(rows, model.Vec2{
			X: float64(start),
			Y: float64(end),
		})
	}

	return rows
}

func checkWorkerCount(workers int, height int) int {
	if workers <= 0 {
		return 1
	}

	if workers > height {
		return height
	}

	return workers
}

func ToGrayscaleParallel(img image.Image, workers int) model.GrayImage {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	workers = checkWorkerCount(workers, height)

	grayImage := model.GrayImage{
		Width:  width,
		Height: height,
		Data:   make([]float64, width*height),
	}

	ranges := splitRows(height, workers)

	var wg sync.WaitGroup
	for i := 0; i < len(ranges); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			GrayifyRange(img, &grayImage, ranges[i])
		}(i)
	}

	wg.Wait()
	return grayImage
}

func GrayifyRange(source image.Image, grayImage *model.GrayImage, rangeRow model.Vec2) {
	maxValue := 65535.0
	redIntensity := 0.299
	greenIntensity := 0.587
	blueIntensity := 0.114
	width := source.Bounds().Dx()

	for y := int(rangeRow.X); y < int(rangeRow.Y); y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := source.At(x, y).RGBA()
			grayValue := redIntensity*float64(r)/maxValue + greenIntensity*float64(g)/maxValue + blueIntensity*float64(b)/maxValue
			grayImage.Data[y*width+x] = grayValue
		}
	}
}
