package painter

import (
	"fmt"
	"sort"
)

var presets = map[string]PainterConfig{
	"oil": {
		BrushSizes:       []int{32, 16, 8, 4},
		ErrorThreshold:   0.075,
		StrokeLength:     2.1,
		Opacity:          0.78,
		Randomness:       0.18,
		UseCurvedStrokes: true,
		CurveSmoothing:   0.35,
		BlurStrength:     0.7,
	},
	"oil-sharp": {
		BrushSizes:       []int{32, 16, 8, 4},
		ErrorThreshold:   0.08,
		StrokeLength:     2.2,
		Opacity:          0.8,
		Randomness:       0.16,
		UseCurvedStrokes: true,
		CurveSmoothing:   0.28,
		BlurStrength:     0.5,
	},
	"oil-soft": {
		BrushSizes:       []int{36, 20, 10, 5},
		ErrorThreshold:   0.07,
		StrokeLength:     2.5,
		Opacity:          0.82,
		Randomness:       0.22,
		UseCurvedStrokes: true,
		CurveSmoothing:   0.42,
		BlurStrength:     0.9,
	},
	"dry": {
		BrushSizes:       []int{22, 12, 6, 3},
		ErrorThreshold:   0.1,
		StrokeLength:     1.4,
		Opacity:          0.6,
		Randomness:       0.3,
		UseCurvedStrokes: true,
		CurveSmoothing:   0.2,
		BlurStrength:     0.35,
	},
	"detail": {
		BrushSizes:       []int{18, 10, 6, 3},
		ErrorThreshold:   0.06,
		StrokeLength:     1.8,
		Opacity:          0.7,
		Randomness:       0.12,
		UseCurvedStrokes: true,
		CurveSmoothing:   0.18,
		BlurStrength:     0.3,
	},
	"impressionist": {
		BrushSizes:       []int{40, 24, 12, 6},
		ErrorThreshold:   0.09,
		StrokeLength:     1.7,
		Opacity:          0.72,
		Randomness:       0.28,
		UseCurvedStrokes: true,
		CurveSmoothing:   0.28,
		BlurStrength:     0.8,
	},
	"rough": {
		BrushSizes:       []int{48, 24, 12},
		ErrorThreshold:   0.11,
		StrokeLength:     2.4,
		Opacity:          0.88,
		Randomness:       0.35,
		UseCurvedStrokes: false,
		CurveSmoothing:   0.3,
		BlurStrength:     0.9,
	},
	"sketch": {
		BrushSizes:       []int{30, 16, 8},
		ErrorThreshold:   0.13,
		StrokeLength:     2.8,
		Opacity:          0.95,
		Randomness:       0.12,
		UseCurvedStrokes: false,
		CurveSmoothing:   0.15,
		BlurStrength:     0.25,
	},
}

func GetPreset(name string) (PainterConfig, bool) {
	config, ok := presets[name]
	if !ok {
		return PainterConfig{}, false
	}
	config.BrushSizes = append([]int(nil), config.BrushSizes...)
	return config, true
}

func PresetNames() []string {
	names := make([]string, 0, len(presets))
	for name := range presets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c PainterConfig) Validate() error {
	if len(c.BrushSizes) == 0 {
		return fmt.Errorf("at least one brush size is required")
	}
	for _, size := range c.BrushSizes {
		if size <= 0 {
			return fmt.Errorf("brush sizes must be positive, got %d", size)
		}
	}
	if c.ErrorThreshold < 0 {
		return fmt.Errorf("error threshold cannot be negative")
	}
	if c.StrokeLength <= 0 {
		return fmt.Errorf("stroke length must be positive")
	}
	if c.Opacity <= 0 || c.Opacity > 1 {
		return fmt.Errorf("opacity must be in (0, 1]")
	}
	if c.Randomness < 0 {
		return fmt.Errorf("randomness cannot be negative")
	}
	if c.CurveSmoothing < 0 || c.CurveSmoothing > 1 {
		return fmt.Errorf("curve smoothing must be in [0, 1]")
	}
	if c.BlurStrength <= 0 {
		return fmt.Errorf("blur strength must be positive")
	}
	return nil
}
