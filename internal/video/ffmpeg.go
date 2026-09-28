package video

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type probedVideo struct {
	Streams []struct {
		AverageFrameRate string `json:"avg_frame_rate"`
		RealFrameRate    string `json:"r_frame_rate"`
		Width            int    `json:"width"`
		Height           int    `json:"height"`
	} `json:"streams"`
}

func probeVideo(path string) (string, string, error) {
	output, err := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0",
		"-show_entries", "stream=avg_frame_rate,r_frame_rate,width,height", "-of", "json", path).CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("probe input video with ffprobe: %w: %s", err, strings.TrimSpace(string(output)))
	}
	var probed probedVideo
	if err := json.Unmarshal(output, &probed); err != nil {
		return "", "", fmt.Errorf("parse ffprobe output: %w", err)
	}
	if len(probed.Streams) == 0 {
		return "", "", fmt.Errorf("input contains no video stream")
	}
	stream := probed.Streams[0]
	for _, candidate := range []string{stream.AverageFrameRate, stream.RealFrameRate} {
		if rate, ok := validFrameRate(candidate); ok {
			return rate, outputPixelFormat(stream.Width, stream.Height), nil
		}
	}
	return "", "", fmt.Errorf("could not determine a valid frame rate for %q", path)
}

func extractVideoFrames(input, outputDir, frameRate string) ([]string, error) {
	if err := runTool("ffmpeg", "-hide_banner", "-loglevel", "error", "-i", input,
		"-map", "0:v:0", "-vf", "fps="+frameRate, "-start_number", "0",
		filepath.Join(outputDir, "frame-%08d.png")); err != nil {
		return nil, fmt.Errorf("extract video frames: %w", err)
	}
	frames, err := filepath.Glob(filepath.Join(outputDir, "frame-*.png"))
	if err != nil {
		return nil, fmt.Errorf("find extracted frames: %w", err)
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("video contains no decodable frames")
	}
	return frames, nil
}

func assembleVideo(framesDir, input, output, frameRate, pixelFormat string) error {
	outputPattern := filepath.Join(framesDir, "frame-%08d.png")
	if err := runTool("ffmpeg", "-hide_banner", "-loglevel", "error", "-y",
		"-framerate", frameRate, "-i", outputPattern, "-i", input,
		"-map", "0:v:0", "-map", "1:a?", "-c:v", "libx264", "-crf", "18",
		"-pix_fmt", pixelFormat, "-c:a", "aac", "-b:a", "192k", "-shortest", output); err != nil {
		return fmt.Errorf("assemble painted video: %w", err)
	}
	return nil
}

func outputPixelFormat(width, height int) string {
	if width%2 != 0 || height%2 != 0 {
		return "yuv444p"
	}
	return "yuv420p"
}

func validFrameRate(value string) (string, bool) {
	parts := strings.Split(value, "/")
	if len(parts) == 2 {
		numerator, numeratorErr := strconv.ParseInt(parts[0], 10, 64)
		denominator, denominatorErr := strconv.ParseInt(parts[1], 10, 64)
		if numeratorErr == nil && denominatorErr == nil && numerator > 0 && denominator > 0 {
			return fmt.Sprintf("%d/%d", numerator, denominator), true
		}
		return "", false
	}
	if len(parts) == 1 {
		rate, err := strconv.ParseFloat(value, 64)
		if err == nil && rate > 0 {
			return strconv.FormatFloat(rate, 'f', -1, 64), true
		}
	}
	return "", false
}

func runTool(name string, args ...string) error {
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(string(output)))
	}
	return nil
}
