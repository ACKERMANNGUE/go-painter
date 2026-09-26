package imageutil

import (
	"image"
	"image/color"
	"testing"
)

type genericImage struct {
	image.Image
}

func TestSampleColorMatchesGenericConversion(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(4, 7, 6, 9))
	rgba.SetRGBA(4, 7, color.RGBA{R: 64, G: 32, B: 1, A: 128})
	rgba.SetRGBA(5, 7, color.RGBA{R: 0, G: 0, B: 0, A: 0})

	nrgba := image.NewNRGBA(image.Rect(4, 7, 6, 9))
	nrgba.SetNRGBA(4, 7, color.NRGBA{R: 210, G: 75, B: 32, A: 96})
	nrgba.SetNRGBA(5, 7, color.NRGBA{R: 80, G: 40, B: 20, A: 0})

	gray := image.NewGray(image.Rect(4, 7, 6, 9))
	gray.SetGray(4, 7, color.Gray{Y: 127})

	images := []struct {
		name string
		img  image.Image
	}{
		{name: "RGBA", img: rgba},
		{name: "NRGBA", img: nrgba},
		{name: "generic", img: gray},
	}
	coordinates := [][2]int{{0, 0}, {1, 0}, {0, 1}, {-1, 0}, {5, 5}}

	for _, testImage := range images {
		for _, point := range coordinates {
			x, y := point[0], point[1]
			got := SampleColor(testImage.img, x, y)
			want := SampleColor(genericImage{Image: testImage.img}, x, y)
			if got != want {
				t.Errorf("SampleColor(%s, %d, %d) = %#v, want %#v", testImage.name, x, y, got, want)
			}
		}
	}
}
