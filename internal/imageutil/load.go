package imageutil

import (
	"fmt"
	"image"
	_ "image/jpeg" // Register the JPEG decoder for image.Decode.
	_ "image/png"  // Register the PNG decoder for image.Decode.
	"os"
)

func LoadImage(path string) (image.Image, error) {
	file, openError := os.Open(path)

	if openError != nil {
		return nil, fmt.Errorf("open input image %q: %w", path, openError)
	}

	defer file.Close()

	img, _, decodeError := image.Decode(file)

	if decodeError != nil {
		return nil, fmt.Errorf("decode input image %q: %w", path, decodeError)
	}

	return img, nil
}