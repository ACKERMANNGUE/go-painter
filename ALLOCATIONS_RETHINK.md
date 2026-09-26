# Allocation And Memory Rethink

This document lists the areas of the project that should be reconsidered to reduce allocations, copies, memory churn, and part of the parallelism overhead.

## Priority 1

### 1. `internal/painter/painter.go`
Relevant methods:
- `Paint`

Why it should be reconsidered:
- `Paint` orchestrates the whole pipeline and accumulates several large transient allocations: RGBA canvas, blurred image, grayscale buffer, gradient field, stroke slices, and point slices for curved strokes.
- the final stroke construction appends all per-brush-size stroke lists into one global slice, so stroke memory stays alive until the end of rendering.
- in curved mode, each stroke then allocates a new point slice before rendering.

What should be reconsidered:
- process strokes in batches instead of building one global list if the pipeline allows it
- reduce how many full intermediate representations are alive at the same time
- pay attention to the preprocessing stage keeping several full-size buffers in memory

Pseudocode type:

```text
TYPE PaintBuffers STRUCT {
    Canvas ImageRGBA
    Reference ImageRGBA
    Gray GrayImage
    Field GradientField
    StrokeBatch []BrushStroke
    CurvePoints []Vec2
}

FUNC Paint(source Image, config PainterConfig, buffers *PaintBuffers) -> ImageRGBA
    reference := BuildReferenceImage(source, buffers)
    field := BuildGradientField(reference, buffers)

    FOR EACH brushSize IN config.BrushSizes
        strokeBatch := GenerateStrokeBatch(source, buffers.Canvas, field, brushSize, config, buffers)
        RenderStrokeBatch(buffers.Canvas, strokeBatch, field, config, buffers)
        ClearStrokeBatch(strokeBatch)
    END FOR

    RETURN buffers.Canvas
END FUNC
```

### 2. `internal/painter/error.go`
Relevant methods:
- `RegionError`
- `ColorDistance`

Why it should be reconsidered:
- `RegionError` is called inside the double loop of `GenerateStrokes`.
- each call rescans a full region and performs two `SampleColor` calls per pixel.
- `SampleColor` converts an image color to `ColorF` every time, so the cost accumulates very quickly.
- `ColorDistance` calls `math.Sqrt` and `math.Pow` in a hot loop, even though a more direct computation or a squared-distance threshold may be enough depending on the requirement.

What should be reconsidered:
- avoid rescanning a full region for each center if local statistics can be cached
- reduce color conversions inside the error loop
- verify whether the square root is actually needed for the threshold logic

Pseudocode type:

```text
TYPE ErrorMap STRUCT {
    Width INT
    Height INT
    Values []FLOAT64
}

FUNC BuildErrorMap(source ImageView, canvas ImageView) -> ErrorMap
    FOR EACH pixelIndex IN source
        Values[pixelIndex] = FastColorDistance(source[pixelIndex], canvas[pixelIndex])
    END FOR
    RETURN ErrorMap
END FUNC

FUNC RegionErrorCached(errorMap ErrorMap, centerX INT, centerY INT, radius INT) -> FLOAT64
    RETURN MeanWindow(errorMap, centerX, centerY, radius)
END FUNC
```

### 3. `internal/imageutil/color.go`
Relevant methods:
- `SampleColor`
- `ToNRGBA`

Why it should be reconsidered:
- these conversions are used in several very hot loops: blur, stroke generation, blending, and rendering.
- `SampleColor` uses `color.NRGBAModel.Convert(img.At(...))`, so it goes through the `image.Image` interface and a model conversion on every read.
- `ToNRGBA` repeats clamps and floating-point multiplications on every write.

What should be reconsidered:
- provide a specialized access path when the source image is already `*image.RGBA` or `*image.NRGBA`
- separate generic paths from optimized paths
- reduce float <-> uint8 conversions in per-pixel loops

Pseudocode type:

```text
TYPE ImageView STRUCT {
    Width INT
    Height INT
    Pix []BYTE
    Stride INT
    Format PixelFormat
}

FUNC SampleColorFast(view ImageView, x INT, y INT) -> ColorF
    index := y * view.Stride + x * 4
    RETURN DecodePixel(view.Pix, index, view.Format)
END FUNC

FUNC WriteColorFast(view ImageView, x INT, y INT, color ColorF) -> VOID
    index := y * view.Stride + x * 4
    EncodePixel(view.Pix, index, color, view.Format)
END FUNC
```

### 4. `internal/render/blend.go`
Relevant methods:
- `Blend`
- `BlendPixel`

Why it should be reconsidered:
- `BlendPixel` is at the core of brush rendering and calls `RGBAAt`, builds a `ColorF`, blends, then calls `Set` with `ToNRGBA`.
- that means several conversions and several API layers for every touched pixel.
- across thousands of discs and curve points, this method quickly dominates the total cost.

What should be reconsidered:
- access the canvas pixel buffer directly instead of using `RGBAAt` and `Set`
- keep the blending logic clean but minimize conversions around it
- precompute or specialize the most common blending paths

Pseudocode type:

```text
TYPE PixelBuffer STRUCT {
    Pix []BYTE
    Stride INT
    Width INT
    Height INT
}

FUNC BlendPixelFast(buffer *PixelBuffer, x INT, y INT, source ColorF, opacity FLOAT64) -> VOID
    IF OutOfBounds(buffer, x, y) THEN
        RETURN
    END IF

    destination := LoadPixel(buffer, x, y)
    result := BlendColor(destination, source, opacity)
    StorePixel(buffer, x, y, result)
END FUNC
```

### 5. `internal/render/brush.go`
Relevant methods:
- `DrawDisc`
- `DrawStroke`
- `DrawCurvedStroke`
- `Lerp`

Why it should be reconsidered:
- `DrawDisc` performs a nested loop per disc and computes distances for every candidate pixel.
- `DrawStroke` samples the stroke as a sequence of consecutive discs.
- `DrawCurvedStroke` adds even more subdivision and discs, so the cost grows very quickly.
- `Lerp` returns a new `Vec2` at each step; that is not dramatic alone, but it accumulates on long curves.

What should be reconsidered:
- reuse a disc mask or a rasterized stamp by radius instead of recomputing the full disc every time
- limit the number of sampled points for strokes if higher density does not improve visual quality
- decouple point generation from stamping to avoid redundant work

Pseudocode type:

```text
TYPE DiscStamp STRUCT {
    Radius INT
    Offsets []PointI
}

FUNC BuildDiscStamp(radius INT) -> DiscStamp
    FOR y FROM -radius TO radius
        FOR x FROM -radius TO radius
            IF x*x + y*y <= radius*radius THEN
                AppendOffset(x, y)
            END IF
        END FOR
    END FOR
    RETURN DiscStamp
END FUNC

FUNC DrawStrokeStamped(buffer *PixelBuffer, stamp DiscStamp, path []Vec2, color ColorF, opacity FLOAT64) -> VOID
    FOR EACH point IN path
        CompositeStamp(buffer, stamp, point, color, opacity)
    END FOR
END FUNC
```

### 6. `internal/render/canvas.go`
Relevant methods:
- `NewCanvas`

Why it should be reconsidered:
- the canvas is filled pixel by pixel using `canvas.Set`.
- this is a full allocation plus a pixel-by-pixel initialization through a high-level API.
- on large images, this adds noticeable cost from the very start of the pipeline.

What should be reconsidered:
- initialize the RGBA buffer directly
- reuse buffers if multiple renders are performed back to back

Pseudocode type:

```text
TYPE CanvasPool STRUCT {
    Reusable []ImageRGBA
}

FUNC AcquireCanvas(pool *CanvasPool, width INT, height INT, background ColorF) -> ImageRGBA
    canvas := ReuseOrAllocate(pool, width, height)
    FillCanvasPixels(canvas, background)
    RETURN canvas
END FUNC
```

## Priority 2

### 7. `internal/imageutil/blur.go`
Relevant methods:
- `GaussianKernel`
- `GaussianBlur`
- `ApplyHorizontalBlur`
- `ApplyVerticalBlur`
- `applyKernelAt`

Why it should be reconsidered:
- the blur allocates at least two full images: a temporary one and the final result.
- `applyKernelAt` is called for every pixel and creates an accumulated `ColorF` on each call.
- each kernel sample rereads a color via `SampleColor`, so there are multiple conversions and clamps per pixel.
- `GaussianKernel` rebuilds the same slice for repeated `(radius, sigma)` pairs.

What should be reconsidered:
- Gaussian kernel caching
- reusable work buffers for horizontal and vertical passes
- possibly a more compact floating-point representation for a single intermediate pass
- specialized source pixel access

Pseudocode type:

```text
TYPE BlurWorkspace STRUCT {
    Kernel []FLOAT64
    HorizontalRow []ColorF
    Temporary ImageRGBA
}

FUNC GaussianBlurWithWorkspace(source ImageView, radius INT, sigma FLOAT64, workspace *BlurWorkspace) -> ImageRGBA
    kernel := AcquireKernel(workspace, radius, sigma)
    temporary := AcquireTemporaryImage(workspace, source.Width, source.Height)

    HorizontalPass(source, temporary, kernel, workspace)
    VerticalPass(temporary, destination, kernel, workspace)

    RETURN destination
END FUNC
```

### 8. `internal/gradient/grayscale.go`
Relevant methods:
- `ToGrayscale`
- `ToGrayscaleParallel`
- `GrayifyRange`

Why it should be reconsidered:
- the grayscale pass reads the whole image and calls `RGBA()` per pixel.
- `GrayifyRange` rebuilds the same constants and repeats floating-point divisions in a very large loop.
- this stage creates a full-image `[]float64` buffer.

What should be reconsidered:
- specialized source buffer access
- more compact outputs if `float64` precision is not required everywhere
- move constants out of loops and out of goroutines where possible

Pseudocode type:

```text
TYPE GrayWorkspace STRUCT {
    Data []FLOAT32
}

FUNC ToGrayscaleFast(source ImageView, workspace *GrayWorkspace) -> GrayImageView
    invMax := CONST_FLOAT
    FOR EACH rowRange IN source
        FOR EACH pixel IN rowRange
            gray := WeightedLuminance(pixel, invMax)
            workspace.Data[pixel.Index] = gray
        END FOR
    END FOR
    RETURN GrayImageView(workspace.Data)
END FUNC
```

### 9. `internal/gradient/sobel.go`
Relevant methods:
- `Sobel`
- `SobelParallel`
- `newGradientField`

Why it should be reconsidered:
- `GradientField` stores a full structure per pixel with four `float64` values.
- this is convenient, but heavy in memory usage and cache bandwidth.
- if only part of those fields is truly needed by the final rendering, the representation may be too rich.

What should be reconsidered:
- verify whether `DX`, `DY`, `Magnitude`, and `Angle` are all needed at the same time
- possibly split storage into more compact structures or compute some values lazily
- avoid computing or storing more than what the rest of the pipeline actually uses

Pseudocode type:

```text
TYPE GradientSample STRUCT {
    DirectionX FLOAT32
    DirectionY FLOAT32
    Magnitude FLOAT32
}

TYPE GradientFieldCompact STRUCT {
    Width INT
    Height INT
    Data []GradientSample
}

FUNC SobelCompact(gray GrayImageView) -> GradientFieldCompact
    FOR EACH interiorPixel IN gray
        gradient := ComputeSobel(gray, interiorPixel)
        StoreCompactGradient(gradient)
    END FOR
    RETURN field
END FUNC
```

### 10. `internal/painter/strokes.go`
Relevant methods:
- `GenerateStrokes`
- `DrawStroke`
- `sampleGradient`

Why it should be reconsidered:
- `GenerateStrokes` builds a large `BrushStroke` slice and already stores final colors inside it.
- `DrawStroke` for curved strokes returns a new `[]Vec2` per stroke.
- `sampleGradient` is called at every path step and rereads the field with rounding each time.

What should be reconsidered:
- streaming or batch-based generation instead of one global accumulation
- reusing a point buffer for curved strokes
- deciding whether color should be captured at generation time or at render time

Pseudocode type:

```text
TYPE StrokeBuffers STRUCT {
    Strokes []BrushStroke
    CurvePoints []Vec2
}

FUNC GenerateStrokeBatch(source ImageView, canvas ImageView, field GradientField, brushSize INT, config PainterConfig, buffers *StrokeBuffers) -> []BrushStroke
    ResetSlice(buffers.Strokes)
    FOR EACH candidateCell
        IF RegionNeedsStroke(...) THEN
            stroke := BuildStrokeDescriptor(...)
            AppendStroke(buffers.Strokes, stroke)
        END IF
    END FOR
    RETURN buffers.Strokes
END FUNC

FUNC BuildCurvePoints(start Vec2, field GradientField, buffers *StrokeBuffers) -> []Vec2
    ResetSlice(buffers.CurvePoints)
    AppendStart(buffers.CurvePoints, start)
    FOR EACH step
        AppendPoint(buffers.CurvePoints, nextPoint)
    END FOR
    RETURN buffers.CurvePoints
END FUNC
```

## Priority 3

### 11. `internal/parallel/rows.go`
Relevant methods:
- `ForEachRow`
- `ForEachRange`

Why it should be reconsidered:
- each parallel stage recreates its channel and goroutines.
- for large images this is not the main allocator compared with image-sized buffers, but it still adds noise and scheduling overhead.
- the impact becomes visible when many small passes are chained together or when the work per job is too small.

What should be reconsidered:
- a reusable worker pool
- work units large enough to amortize channel overhead
- using `ForEachRange` for passes whose body is heavy enough

Pseudocode type:

```text
TYPE WorkerPool STRUCT {
    Workers INT
    Queue JobQueue
}

TYPE RowJob STRUCT {
    StartY INT
    EndY INT
    Fn FUNC(INT) -> VOID
}

FUNC SubmitRows(pool *WorkerPool, height INT, chunkSize INT, fn FUNC(INT) -> VOID) -> VOID
    FOR yStart FROM 0 TO height STEP chunkSize
        Enqueue(pool.Queue, RowJob{StartY: yStart, EndY: Min(yStart + chunkSize, height), Fn: fn})
    END FOR
    WaitAll(pool)
END FUNC
```

### 12. `internal/painter/basic.go`
Relevant methods:
- `PaintBasic`
- `RandomRange`

Why it should be reconsidered:
- this is not necessarily the main hot path of `BenchmarkPaint`, but `RandomRange` recreates an RNG on every call.
- if this path remains useful for teaching or comparisons, it should be cleaned up to stay consistent with the rest of the project.

What should be reconsidered:
- share the random source instead of re-instantiating a generator in a loop

Pseudocode type:

```text
TYPE RandomSource STRUCT {
    State RNG
}

FUNC NextRange(rng *RandomSource, minimum FLOAT64, maximum FLOAT64) -> FLOAT64
    RETURN minimum + NextUnitFloat(rng) * (maximum - minimum)
END FUNC
```

## Recommended Work Order

1. Rework `RegionError` and the color conversions used inside its loops.
2. Rework `BlendPixel` and brush rendering (`DrawDisc`, `DrawStroke`, `DrawCurvedStroke`).
3. Rework the blur pipeline to limit intermediate images and per-pixel conversions.
4. Rework global accumulation in `Paint` and `GenerateStrokes` to avoid keeping everything alive in memory at once.
5. Only after that, refine the parallel helpers and the `GrayImage` / `GradientField` representations.

## Verification Checklist

- does each image pass avoid generic `image.Image` conversions in hot loops?
- is each full-image-sized buffer truly necessary?
- is each slice built per stroke or per batch reusable?
- does each pixel write still go through `Set` / `RGBAAt` when direct access would be possible?
- can each region or mask recomputation be cached?
- does each parallel stage perform enough work to amortize channels and goroutines?