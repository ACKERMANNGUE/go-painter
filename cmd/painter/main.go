package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ACKERMANNGUE/go-painter/internal/imageutil"
	"github.com/ACKERMANNGUE/go-painter/internal/model"
	"github.com/ACKERMANNGUE/go-painter/internal/painter"
)

func main() {
	inputPath := flag.String("input", "input/input.jpg", "Path to the input PNG or JPEG image")
	outputPath := flag.String("output", "output/painting.png", "Path to the output PNG image")
	styleName := flag.String("style", "oil", "Painting preset: oil, impressionist, or rough")
	seed := flag.Uint64("seed", 1, "Deterministic random seed used for stroke placement")
	listStyles := flag.Bool("list-styles", false, "Print available style names and exit")
	workers := flag.Int("workers", 8, "Number of workers to use")
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

	config, ok := painter.GetPreset(*styleName)
	if !ok {
		fmt.Fprintf(os.Stderr, "error: unknown style %q; available styles: %s\n", *styleName, strings.Join(painter.PresetNames(), ", "))
		os.Exit(2)
	}
	config.Workers = *workers

	source, err := imageutil.LoadImage(*inputPath)
	if err != nil {
		fatal(err)
	}

	engine, err := painter.New(config, *seed)
	if err != nil {
		fatal(fmt.Errorf("create painter: %w", err))
	}

	var buffers model.PaintBuffer
	started := time.Now()
	result, err := engine.Paint(source, &buffers)
	if err != nil {
		fatal(fmt.Errorf("paint image: %w", err))
	}
	if err := imageutil.SavePNG(*outputPath, result); err != nil {
		fatal(err)
	}

	fmt.Printf("Painted %s -> %s using style=%s seed=%d in %s\n",
		*inputPath, *outputPath, *styleName, *seed, time.Since(started).Round(time.Millisecond))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
