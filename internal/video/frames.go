package video

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"github.com/ACKERMANNGUE/go-painter/internal/painter"
	"github.com/ACKERMANNGUE/go-painter/internal/progress"
)

func paintFrames(frames []string, outputDir string, config painter.PainterConfig, seed uint64) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create painted frame directory: %w", err)
	}
	engine, err := painter.New(config, seed)
	if err != nil {
		return fmt.Errorf("create painter: %w", err)
	}
	sort.Strings(frames)
	var buffers model.PaintBuffer
	started := time.Now()
	defer fmt.Fprintln(os.Stderr)
	for index, framePath := range frames {
		source, err := imageutil.LoadImage(framePath)
		if err != nil {
			return err
		}
		result, err := engine.Paint(source, &buffers)
		if err != nil {
			return fmt.Errorf("paint frame %q: %w", framePath, err)
		}
		outputPath := filepath.Join(outputDir, filepath.Base(framePath))
		if err := imageutil.SavePNG(outputPath, result); err != nil {
			return fmt.Errorf("save painted frame %q: %w", outputPath, err)
		}
		writeFrameProgress(os.Stderr, index+1, len(frames), time.Since(started))
	}
	return nil
}

func writeFrameProgress(output io.Writer, completed, total int, elapsed time.Duration) {
	progress.Write(output, "Painting frames", completed, total, elapsed)
}
