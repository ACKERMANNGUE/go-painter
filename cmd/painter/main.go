package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"github.com/ACKERMANNGUE/go-painter/internal/painter"
	"github.com/ACKERMANNGUE/go-painter/internal/progress"
	"github.com/ACKERMANNGUE/go-painter/internal/video"
)

func main() {
	inputPath := flag.String("input", "input/input.jpg", "Path to the input PNG or JPEG image")
	outputPath := flag.String("output", "output/painting.png", "Path to the output PNG image")
	styleName := flag.String("style", "oil-sharp", "Painting preset: oil, impressionist, or rough")
	seed := flag.Uint64("seed", 1, "Deterministic random seed used for stroke placement")
	listStyles := flag.Bool("list-styles", false, "Print available style names and exit")
	workers := flag.Int("workers", runtime.GOMAXPROCS(0), "Number of workers to use")
	showDrawingSteps := flag.Bool("show-drawing-steps", false, "Export an MP4 showing each brush stroke")
	debug := flag.Bool("debug", false, "Export grayscale, blur, Sobel, and error heatmap images")
	flag.Parse()

	if *listStyles {
		fmt.Println(strings.Join(painter.PresetNames(), "\n"))
		return
	}
	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "error: -input is required")
		flag.Usage()
		os.Exit(2)
	}

	config, err := loadConfig(*styleName, *workers)
	if err != nil {
		fatal(err)
	}
	fmt.Fprintf(os.Stderr, "Parameters: input=%q output=%q style=%s seed=%d workers=%d show-drawing-steps=%t debug=%t config=%+v\n",
		*inputPath, *outputPath, *styleName, *seed, *workers, *showDrawingSteps, *debug, config)

	started := time.Now()
	if err := paintImage(*inputPath, *outputPath, config, *seed, *showDrawingSteps, *debug); err != nil {
		fatal(err)
	}

	fmt.Printf("Painted %s -> %s using style=%s seed=%d in %s\n",
		*inputPath, *outputPath, *styleName, *seed, time.Since(started).Round(time.Millisecond))
}

func loadConfig(styleName string, workers int) (painter.PainterConfig, error) {
	if workers < 1 {
		return painter.PainterConfig{}, fmt.Errorf("workers must be at least 1")
	}
	config, ok := painter.GetPreset(styleName)
	if !ok {
		return painter.PainterConfig{}, fmt.Errorf("unknown style %q; available styles: %s", styleName, strings.Join(painter.PresetNames(), ", "))
	}
	config.Workers = workers
	return config, nil
}

func paintImage(inputPath, outputPath string, config painter.PainterConfig, seed uint64, showDrawingSteps, debug bool) error {
	source, err := imageutil.LoadImage(inputPath)
	if err != nil {
		return err
	}
	engine, err := painter.New(config, seed)
	if err != nil {
		return fmt.Errorf("create painter: %w", err)
	}
	var buffers model.PaintBuffer
	if showDrawingSteps {
		if err := paintImageWithDrawingSteps(engine, source, outputPath, &buffers); err != nil {
			return err
		}
	} else {
		result, err := engine.Paint(source, &buffers)
		if err != nil {
			return fmt.Errorf("paint image: %w", err)
		}
		if err := imageutil.SavePNG(outputPath, result); err != nil {
			return err
		}
	}
	if debug {
		return painter.SaveDebugImages(outputPath, source, &buffers)
	}
	return nil
}

func paintImageWithDrawingSteps(engine *painter.Painter, source image.Image, outputPath string, buffers *model.PaintBuffer) error {
	drawingStepsPath := drawingStepsOutputPath(outputPath)
	recorder, err := video.NewDrawingStepsRecorder(drawingStepsPath, source.Bounds().Dx(), source.Bounds().Dy())
	if err != nil {
		return err
	}
	defer recorder.Abort()
	fmt.Fprintf(os.Stderr, "Drawing steps output: %s (%d fps, one frame per stroke)\n", drawingStepsPath, video.DrawingStepsFrameRate)

	drawingProgress := drawingStepsProgress{}
	defer drawingProgress.finish()
	result, err := engine.PaintWithProgress(source, buffers, func() error {
		return recorder.WriteFrame(buffers.Canvas)
	}, drawingProgress.update)
	if err != nil {
		return fmt.Errorf("paint image: %w", err)
	}
	if err := imageutil.SavePNG(outputPath, result); err != nil {
		return err
	}
	if err := recorder.Close(); err != nil {
		return err
	}
	fmt.Printf("Drawing steps video: %s\n", drawingStepsPath)
	return nil
}

func drawingStepsOutputPath(outputPath string) string {
	extension := filepath.Ext(outputPath)
	return strings.TrimSuffix(outputPath, extension) + "_drawing_steps.mp4"
}

type drawingStepsProgress struct {
	phase       string
	pass        int
	started     time.Time
	lastPercent int
	active      bool
}

func (p *drawingStepsProgress) update(update painter.PaintProgress) {
	if update.Total <= 0 {
		return
	}
	if !p.active || p.phase != update.Phase || p.pass != update.Pass {
		p.phase = update.Phase
		p.pass = update.Pass
		p.started = time.Now()
		p.lastPercent = -1
		p.active = true
	}
	percentage := update.Completed * 100 / update.Total
	if percentage == p.lastPercent && update.Completed < update.Total {
		return
	}
	p.lastPercent = percentage
	phaseName := "Generating strokes"
	if update.Phase == painter.PaintPhaseRender {
		phaseName = "Rendering strokes"
	}
	label := fmt.Sprintf("Drawing pass %d/%d: %s", update.Pass, update.Passes, phaseName)
	progress.Write(os.Stderr, label, update.Completed, update.Total, time.Since(p.started))
	if update.Completed >= update.Total {
		fmt.Fprintln(os.Stderr)
		p.active = false
	}
}

func (p *drawingStepsProgress) finish() {
	if p.active {
		fmt.Fprintln(os.Stderr)
		p.active = false
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
