# go-painter

go-painter is a fun learning project to practice Go by turning a source image into a painting made of layered brush strokes.

## Project Goal

The main goal of this repository is to learn Go through a creative, visual challenge instead of a purely academic exercise.

This project focuses on:

- Building a non-trivial CLI app in Go.
- Structuring code into clear internal packages.
- Implementing image-processing techniques from scratch.
- Exploring deterministic generative art (same seed, same output).
- Understanding rendering trade-offs: quality, speed, and style.

## Motivation

The idea started from two inspirations:

- A ChatGPT Astra demo video that recreated an image with strokes in Paint.
- The Instagram account Edge Case Labs: https://www.instagram.com/edgecaselabs/ where Minecraft textures are repainted by hand.

Seeing both made me think: this is probably something I can code myself as a personal Go learning project.

## What the Program Does

Given an input PNG or JPEG image, the program:

1. Builds a white canvas.
2. Analyzes the source image structure (grayscale + Sobel gradients).
3. Places brush strokes where the canvas still differs from the source.
4. Orients strokes based on local image direction.
5. Blends all strokes into a final painterly PNG output.

You can choose different painting presets (`oil`, `impressionist`, `rough`) and keep output deterministic using a fixed random seed.

## Prerequisites

You need Go installed on your machine.

- Official installation guide: https://go.dev/doc/install

This project currently targets the Go version declared in `go.mod`.

## How to Run

From the repository root:

1. List available styles:

```bash
go run ./cmd/painter -list-styles
```

2. Run the painter with defaults:

```bash
go run ./cmd/painter
```

Defaults:

- Input: `input/input.jpg`
- Output: `output/painting.png`
- Style: `oil`
- Seed: `1`
- Workers: `8`

3. Run with custom options:

```bash
go run ./cmd/painter \
  -input input/my-image.jpg \
  -output output/my-painting.png \
  -style impressionist \
  -seed 42 \
  -workers 12
```

### CLI Flags

- `-input`: Path to input PNG/JPEG.
- `-output`: Path to output PNG.
- `-style`: Preset name (`oil`, `impressionist`, `rough`).
- `-seed`: Deterministic random seed for stroke placement.
- `-workers`: Number of workers for parallel grayscale conversion.
- `-list-styles`: Print style names and exit.

## Global Project Structure

```text
cmd/painter/main.go            # CLI entrypoint and flags

internal/gradient/
  grayscale.go                 # grayscale conversion (sequential + parallel)
  sobel.go                     # Sobel gradient field computation

internal/imageutil/
  blur.go                      # Gaussian blur (separable kernel)
  color.go                     # color sampling and clamping
  load.go                      # input image loading
  save.go                      # PNG output saving

internal/model/
  types.go                     # core math and rendering data structures

internal/painter/
  painter.go                   # main painter pipeline
  presets.go                   # style presets and config validation
  strokes.go                   # stroke generation, direction following, shuffling
  error.go                     # region error metric
  basic.go                     # simpler/experimental painter
  painter_benchmark_test.go    # benchmark for paint pipeline

internal/render/
  canvas.go                    # canvas initialization
  brush.go                     # drawing discs, straight and curved strokes
  blend.go                     # alpha blending
```

## Core Concepts Used

## 1) Gaussian Blur (Why it is used)

Before extracting gradients, the source image is blurred.

Why:

- Reduces high-frequency noise.
- Stabilizes edge direction estimation.
- Produces smoother, more coherent stroke flow.

Implementation note:

- Uses a separable Gaussian kernel (horizontal pass + vertical pass), which is faster than a full 2D convolution.

## 2) Grayscale Conversion (Why it is used)

Sobel works on intensity rather than color channels.

Why:

- Simplifies edge detection.
- Captures structure/luminance independently from hue.
- Enables efficient parallel processing by row ranges.

Implementation note:

- Uses weighted luminance coefficients (`0.299, 0.587, 0.114`) and includes a worker-based parallel version.

## 3) Sobel Gradient Field (Why it is used)

Sobel computes local spatial derivatives (`gx`, `gy`) and derives:

- Gradient magnitude (edge strength).
- Gradient angle (edge normal direction).

Why:

- Provides orientation cues from the image geometry.
- Helps align strokes with contours (using direction perpendicular to gradient).

Practical effect:

- Strokes follow object shapes better and feel less random.

## 4) Error-Driven Stroke Placement (Why it is used)

For each brush-size pass, the algorithm compares the source and current canvas in local regions.

Why:

- Places strokes where they are still needed.
- Avoids overspending strokes on already-good areas.
- Builds image progressively from coarse to fine detail.

Implementation note:

- Region error is based on average RGB distance over a neighborhood around each candidate point.

## 5) Multi-Scale Painting (Why it is used)

Painting is done in several brush sizes (large to small).

Why:

- Large brushes establish global masses first.
- Smaller brushes recover details later.
- Mimics traditional painting workflow and improves visual coherence.

## 6) Curved vs Straight Strokes (Why it is used)

Depending on preset:

- Curved strokes are generated by stepping through local direction and smoothing direction changes.
- Straight strokes use a single angle and length.

Why:

- Curved mode gives more organic, contour-following marks.
- Straight mode gives a rougher, bolder style.

## 7) Alpha Blending (Why it is used)

Each stroke is blended on the canvas using opacity.

Why:

- Allows gradual accumulation of paint.
- Helps merge strokes naturally instead of hard overwrites.
- Makes style presets expressive through opacity control.

## Style Presets

The project currently provides three presets:

- `oil`: balanced detail and smooth curved strokes.
- `impressionist`: larger brushes and more randomness for visible painterly texture.
- `rough`: stronger coverage, no curved stroke integration, more direct marks.

Each preset controls:

- Brush sizes
- Error threshold
- Stroke length
- Opacity
- Randomness
- Curved stroke usage and smoothing

## Determinism and Reproducibility

The random generator is seeded from the CLI `-seed` value.

This means:

- Same input + same config + same seed => same painting output.
- Easy to compare algorithm tweaks without random noise in results.

## Learning Outcomes Behind This Project

This project is intentionally educational. It is a practical way to learn:

- Go package organization and clean boundaries.
- Concurrency basics with goroutines and wait groups.
- Image processing fundamentals (blur, gradients, edge direction).
- Geometry-driven rendering with strokes.
- Reproducible procedural generation.
- Benchmarking workflow for performance awareness.

> NOTE: The project is not over yet, so I will add more stuff (such as new styles, rendering informations like progress bar or so, better concurrency, etc...)