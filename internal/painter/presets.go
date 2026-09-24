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
	},
	"impressionist": {
		BrushSizes:       []int{40, 24, 12, 6},
		ErrorThreshold:   0.09,
		StrokeLength:     1.7,
		Opacity:          0.72,
		Randomness:       0.28,
		UseCurvedStrokes: true,
		CurveSmoothing:   0.28,
	},
	"rough": {
		BrushSizes:       []int{48, 24, 12},
		ErrorThreshold:   0.11,
		StrokeLength:     2.4,
		Opacity:          0.88,
		Randomness:       0.35,
		UseCurvedStrokes: false,
		CurveSmoothing:   0.3,
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
	return nil
}
