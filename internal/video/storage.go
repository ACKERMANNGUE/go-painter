package video

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func frameDirectorySize(directory string) (uint64, error) {
	var total uint64
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		size := uint64(info.Size())
		if total > ^uint64(0)-size {
			return fmt.Errorf("frame size total overflows uint64")
		}
		total += size
		return nil
	})
	if err != nil {
		return 0, err
	}
	return total, nil
}

func estimateFrameStorage(inputBytes uint64) (paintedBytes, totalBytes uint64, err error) {
	margin := inputBytes / 5
	if inputBytes%5 != 0 {
		margin++
	}
	if inputBytes > ^uint64(0)-margin {
		return 0, 0, fmt.Errorf("painted frame size estimate overflows uint64")
	}
	paintedBytes = inputBytes + margin
	if inputBytes > ^uint64(0)-paintedBytes {
		return 0, 0, fmt.Errorf("total temporary frame size estimate overflows uint64")
	}
	return paintedBytes, inputBytes + paintedBytes, nil
}

func availableDiskSpace(path string) (uint64, string, error) {
	output, err := exec.Command("df", "-Pk", path).Output()
	if err != nil {
		return 0, "", fmt.Errorf("run df: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return 0, "", fmt.Errorf("unexpected df output")
	}
	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) < 4 {
		return 0, "", fmt.Errorf("unexpected df output: %q", lines[len(lines)-1])
	}
	availableBlocks, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("parse available blocks from df: %w", err)
	}
	if availableBlocks > ^uint64(0)/1024 {
		return 0, "", fmt.Errorf("available disk size overflows uint64")
	}
	return availableBlocks * 1024, fields[0], nil
}

func gigaBytes(bytes uint64) float64 {
	return float64(bytes) / 1_000_000_000
}

func formatStorageSize(bytes uint64) string {
	switch {
	case bytes >= 1_000_000_000:
		return fmt.Sprintf("%.2f Go", gigaBytes(bytes))
	case bytes >= 1_000_000:
		return fmt.Sprintf("%.2f Mo (%.4f Go)", float64(bytes)/1_000_000, gigaBytes(bytes))
	case bytes >= 1_000:
		return fmt.Sprintf("%.2f Ko (%.6f Go)", float64(bytes)/1_000, gigaBytes(bytes))
	default:
		return fmt.Sprintf("%d octets (%.9f Go)", bytes, gigaBytes(bytes))
	}
}

func checkFrameSpace(disk string, availableBytes, inputFrameBytes, paintedEstimate, totalEstimate uint64) error {
	if availableBytes >= paintedEstimate {
		return nil
	}
	return fmt.Errorf("ABORT: disk %s can't handle the estimated total frame storage of %s; %s is used by extracted frames and only %s is free for the estimated %s of painted frames",
		disk, formatStorageSize(totalEstimate), formatStorageSize(inputFrameBytes),
		formatStorageSize(availableBytes), formatStorageSize(paintedEstimate))
}
