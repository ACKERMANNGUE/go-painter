package video

import (
	"strings"
	"testing"
)

func TestEstimateFrameStorage(t *testing.T) {
	paintedBytes, totalBytes, err := estimateFrameStorage(100)
	if err != nil {
		t.Fatalf("estimateFrameStorage returned an error: %v", err)
	}
	if paintedBytes != 120 || totalBytes != 220 {
		t.Fatalf("estimateFrameStorage(100) = (%d, %d), want (120, 220)", paintedBytes, totalBytes)
	}

	if _, _, err := estimateFrameStorage(^uint64(0)); err == nil {
		t.Fatal("estimateFrameStorage should report overflow")
	}
}

func TestFormatStorageSize(t *testing.T) {
	if got, want := formatStorageSize(2048), "2.05 Ko (0.000002 Go)"; got != want {
		t.Fatalf("formatStorageSize(2048) = %q, want %q", got, want)
	}
}

func TestCheckFrameSpace(t *testing.T) {
	if err := checkFrameSpace("/dev/test", 99, 80, 100, 180); err == nil || !strings.Contains(err.Error(), "ABORT: disk /dev/test can't handle") {
		t.Fatalf("checkFrameSpace error = %v, want an ABORT error for /dev/test", err)
	}
	if err := checkFrameSpace("/dev/test", 100, 80, 100, 180); err != nil {
		t.Fatalf("checkFrameSpace returned an unexpected error: %v", err)
	}
}

func TestValidFrameRate(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
		ok    bool
	}{
		{name: "fraction", value: "30000/1001", want: "30000/1001", ok: true},
		{name: "integer", value: "24", want: "24", ok: true},
		{name: "decimal", value: "29.97", want: "29.97", ok: true},
		{name: "zero denominator", value: "30/0"},
		{name: "zero rate", value: "0"},
		{name: "invalid", value: "N/A"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := validFrameRate(test.value)
			if got != test.want || ok != test.ok {
				t.Fatalf("validFrameRate(%q) = (%q, %t), want (%q, %t)", test.value, got, ok, test.want, test.ok)
			}
		})
	}
}

func TestOutputPixelFormat(t *testing.T) {
	for _, test := range []struct {
		name          string
		width, height int
		want          string
	}{
		{name: "even dimensions", width: 1920, height: 1080, want: "yuv420p"},
		{name: "odd width", width: 1919, height: 1080, want: "yuv444p"},
		{name: "odd height", width: 1920, height: 1079, want: "yuv444p"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := outputPixelFormat(test.width, test.height); got != test.want {
				t.Fatalf("outputPixelFormat(%d, %d) = %q, want %q", test.width, test.height, got, test.want)
			}
		})
	}
}
