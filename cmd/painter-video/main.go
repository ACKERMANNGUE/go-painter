package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ACKERMANNGUE/go-painter/internal/painter"
	"github.com/ACKERMANNGUE/go-painter/internal/video"
)

func main() {
	inputPath := flag.String("input", "", "Source video path")
	outputPath := flag.String("output", "output/painting.mp4", "Destination video path")
	presetName := flag.String("preset", "oil-sharp", "Painting preset")
	seed := flag.Uint64("seed", 1, "Deterministic random seed")
	workers := flag.Int("workers", 8, "Number of image-processing workers")
	listPresets := flag.Bool("list-presets", false, "Print available presets and exit")
	flag.Parse()

	if *listPresets {
		fmt.Println(strings.Join(painter.PresetNames(), "\n"))
		return
	}
	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "error: -input is required")
		flag.Usage()
		os.Exit(2)
	}

	config, err := loadConfig(*presetName, *workers)
	if err != nil {
		fatal(err, 2)
	}
	fmt.Fprintf(os.Stderr, "Parameters: input=%q output=%q preset=%s seed=%d workers=%d config=%+v\n",
		*inputPath, *outputPath, *presetName, *seed, *workers, config)

	started := time.Now()
	if err := video.Process(*inputPath, *outputPath, config, *seed); err != nil {
		fatal(err, 1)
	}
	fmt.Printf("Painted video %s -> %s (preset=%s seed=%d, %s)\n",
		*inputPath, *outputPath, *presetName, *seed, time.Since(started).Round(time.Millisecond))
}

func loadConfig(name string, workers int) (painter.PainterConfig, error) {
	if workers < 1 {
		return painter.PainterConfig{}, fmt.Errorf("workers must be at least 1")
	}
	config, ok := painter.GetPreset(name)
	if !ok {
		return painter.PainterConfig{}, fmt.Errorf("unknown preset %q; available presets: %s", name, strings.Join(painter.PresetNames(), ", "))
	}
	config.Workers = workers
	return config, nil
}

func fatal(err error, code int) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(code)
}
