package imageutil

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

func SavePNG(path string, img image.Image) error {
	dir := filepath.Dir(path)
	if dir != "." {
		mkdirError := os.MkdirAll(dir, 0o755)
		if  mkdirError != nil {
			return fmt.Errorf("create output directory: %w", mkdirError)
		}
	}

	file, createError := os.Create(path)
	if createError != nil {
		return fmt.Errorf("create output image %q: %w", path, createError)
	}
	defer file.Close()

	encodeError := png.Encode(file, img);
	if encodeError != nil {
		return fmt.Errorf("encode PNG %q: %w", path, encodeError)
	}

	return nil
}