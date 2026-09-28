package video

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ACKERMANNGUE/go-painter/internal/painter"
)

func Process(inputPath, outputPath string, config painter.PainterConfig, seed uint64) error {
	input, err := filepath.Abs(inputPath)
	if err != nil {
		return fmt.Errorf("resolve input video: %w", err)
	}
	output, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolve output video: %w", err)
	}
	if input == output {
		return fmt.Errorf("input and output video paths must be different")
	}
	if info, err := os.Stat(input); err != nil {
		return fmt.Errorf("stat input video: %w", err)
	} else if !info.Mode().IsRegular() {
		return fmt.Errorf("input %q is not a regular file", inputPath)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	frameRate, pixelFormat, err := probeVideo(input)
	if err != nil {
		return err
	}
	workingDir, err := os.MkdirTemp("", "go-painter-video-*")
	if err != nil {
		return fmt.Errorf("create temporary video workspace: %w", err)
	}
	defer os.RemoveAll(workingDir)

	inputFrames := filepath.Join(workingDir, "input-frames")
	outputFrames := filepath.Join(workingDir, "painted-frames")
	if err := os.MkdirAll(inputFrames, 0o755); err != nil {
		return fmt.Errorf("create frame directory: %w", err)
	}
	frames, err := extractVideoFrames(input, inputFrames, frameRate)
	if err != nil {
		return err
	}
	inputFrameBytes, err := frameDirectorySize(inputFrames)
	if err != nil {
		return fmt.Errorf("measure extracted frames: %w", err)
	}
	outputFrameEstimate, totalFrameEstimate, err := estimateFrameStorage(inputFrameBytes)
	if err != nil {
		return fmt.Errorf("estimate temporary frame storage: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Frames created: %d\n", len(frames))
	fmt.Fprintf(os.Stderr, "Temporary frames folder: %s\n", workingDir)
	fmt.Fprintf(os.Stderr, "Extracted frames: %s; estimated total frame storage: %s (includes 20%% output margin)\n",
		formatStorageSize(inputFrameBytes), formatStorageSize(totalFrameEstimate))
	availableBytes, disk, err := availableDiskSpace(workingDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: cannot check free space for %s: %v\n", workingDir, err)
	} else {
		fmt.Fprintf(os.Stderr, "Free space on %s: %s\n", disk, formatStorageSize(availableBytes))
		if err := checkFrameSpace(disk, availableBytes, inputFrameBytes, outputFrameEstimate, totalFrameEstimate); err != nil {
			return err
		}
	}
	if err := paintFrames(frames, outputFrames, config, seed); err != nil {
		return err
	}
	return assembleVideo(outputFrames, input, output, frameRate, pixelFormat)
}
