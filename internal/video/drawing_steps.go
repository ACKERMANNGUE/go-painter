package video

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const DrawingStepsFrameRate = 120

type DrawingStepsRecorder struct {
	command *exec.Cmd
	input   io.WriteCloser
	output  string
	stderr  bytes.Buffer
	frame   []byte
	width   int
	height  int
	closed  bool
}

func NewDrawingStepsRecorder(output string, width, height int) (*DrawingStepsRecorder, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("drawing steps dimensions must be positive, got %dx%d", width, height)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return nil, fmt.Errorf("create drawing steps directory: %w", err)
	}

	pixelFormat := "yuv420p"
	if width%2 != 0 || height%2 != 0 {
		pixelFormat = "yuv444p"
	}
	command := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-y",
		"-f", "rawvideo", "-pixel_format", "rgba", "-video_size", fmt.Sprintf("%dx%d", width, height),
		"-framerate", fmt.Sprint(DrawingStepsFrameRate), "-i", "pipe:0",
		"-an", "-c:v", "libx264", "-crf", "18", "-pix_fmt", pixelFormat, output)
	recorder := &DrawingStepsRecorder{
		command: command,
		output:  output,
		width:   width,
		height:  height,
		frame:   make([]byte, width*height*4),
	}
	command.Stderr = &recorder.stderr
	input, err := command.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("open ffmpeg input for drawing steps: %w", err)
	}
	recorder.input = input
	if err := command.Start(); err != nil {
		_ = input.Close()
		return nil, fmt.Errorf("start ffmpeg for drawing steps: %w", err)
	}
	return recorder, nil
}

func (r *DrawingStepsRecorder) WriteFrame(canvas *image.RGBA) error {
	if r.closed {
		return fmt.Errorf("drawing steps recorder is closed")
	}
	if canvas.Bounds().Dx() != r.width || canvas.Bounds().Dy() != r.height {
		return fmt.Errorf("drawing steps frame is %dx%d, want %dx%d", canvas.Bounds().Dx(), canvas.Bounds().Dy(), r.width, r.height)
	}
	compositeFrameOnWhite(canvas, r.frame)
	return r.writeFrame()
}

func compositeFrameOnWhite(canvas *image.RGBA, frame []byte) {
	width, height := canvas.Bounds().Dx(), canvas.Bounds().Dy()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sourceOffset := canvas.PixOffset(canvas.Bounds().Min.X+x, canvas.Bounds().Min.Y+y)
			outputOffset := (y*width + x) * 4
			alpha := uint16(canvas.Pix[sourceOffset+3])
			white := uint16(255) - alpha
			for channel := 0; channel < 3; channel++ {
				value := uint16(canvas.Pix[sourceOffset+channel]) + white
				if value > 255 {
					value = 255
				}
				frame[outputOffset+channel] = byte(value)
			}
			frame[outputOffset+3] = 255
		}
	}
}

func (r *DrawingStepsRecorder) writeFrame() error {
	for frame := r.frame; len(frame) > 0; {
		written, err := r.input.Write(frame)
		if err != nil {
			return fmt.Errorf("write drawing steps frame to ffmpeg: %w: %s", err, strings.TrimSpace(r.stderr.String()))
		}
		if written == 0 {
			return io.ErrShortWrite
		}
		frame = frame[written:]
	}
	return nil
}

func (r *DrawingStepsRecorder) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	inputErr := r.input.Close()
	commandErr := r.command.Wait()
	if commandErr != nil {
		_ = os.Remove(r.output)
		return fmt.Errorf("encode drawing steps video: %w: %s", commandErr, strings.TrimSpace(r.stderr.String()))
	}
	if inputErr != nil {
		_ = os.Remove(r.output)
		return fmt.Errorf("close ffmpeg input for drawing steps: %w", inputErr)
	}
	return nil
}

func (r *DrawingStepsRecorder) Abort() {
	if r.closed {
		return
	}
	r.closed = true
	_ = r.input.Close()
	if r.command.Process != nil {
		_ = r.command.Process.Kill()
		_ = r.command.Wait()
	}
	_ = os.Remove(r.output)
}
