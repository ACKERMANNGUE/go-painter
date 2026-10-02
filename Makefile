.DEFAULT_GOAL := help

GO ?= go
BIN_DIR ?= bin
ARGS ?=
GO_FILES := $(shell find cmd internal -type f -name '*.go')

.PHONY: all help build install test test-race coverage vet fmt fmt-check check benchmark benchmark-quick run run-folder run-video list-styles list-presets check-video-deps clean

all: check build

help:
	@printf '%s\n' \
		'Available targets:' \
		'  all                Run checks and build all command binaries' \
		'  build              Build all command binaries into $(BIN_DIR)' \
		'  install            Install all command binaries' \
		'  test               Run all Go tests' \
		'  test-race          Run all Go tests with the race detector' \
		'  coverage           Run all Go tests with coverage enabled' \
		'  vet                Run go vet on all packages' \
		'  fmt                Format all Go source files' \
		'  fmt-check          Check Go formatting without changing files' \
		'  check              Run formatting checks, vet, and tests' \
		'  benchmark          Run painter benchmarks' \
		'  benchmark-quick    Run the one-worker painter benchmark once' \
		'  run                Run the image painter; pass options with ARGS' \
		'  run-folder         Run the folder painter; pass options with ARGS' \
		'  run-video          Run the video painter; pass options with ARGS' \
		'  list-styles        Print image painter styles' \
		'  list-presets       Print batch painter presets' \
		'  check-video-deps   Check for ffmpeg and ffprobe' \
		'  clean              Remove binaries built by this Makefile'

build: $(BIN_DIR)/painter $(BIN_DIR)/painter-video $(BIN_DIR)/painter-folder

$(BIN_DIR)/painter: $(GO_FILES) go.mod
	@mkdir -p "$(BIN_DIR)"
	$(GO) build -o "$@" ./cmd/painter

$(BIN_DIR)/painter-video: $(GO_FILES) go.mod
	@mkdir -p "$(BIN_DIR)"
	$(GO) build -o "$@" ./cmd/painter-video

$(BIN_DIR)/painter-folder: $(GO_FILES) go.mod
	@mkdir -p "$(BIN_DIR)"
	$(GO) build -o "$@" ./cmd/painter-folder

install:
	$(GO) install ./cmd/...

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

coverage:
	$(GO) test -cover ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w $(GO_FILES)

fmt-check:
	@files="$$(gofmt -l $(GO_FILES))"; \
	if [ -n "$$files" ]; then \
		printf 'Files need gofmt:\n%s\n' "$$files"; \
		exit 1; \
	fi
	@printf '%s\n' 'All Go files are formatted.'

check: fmt-check vet test

benchmark:
	$(GO) test ./internal/painter -run '^$$' -bench . -benchmem

benchmark-quick:
	$(GO) test ./internal/painter -run '^$$' -bench '^BenchmarkPaint/workers_1$$' -benchmem -benchtime=1x

run:
	$(GO) run ./cmd/painter $(ARGS)

run-folder:
	$(GO) run ./cmd/painter-folder $(ARGS)

run-video:
	$(GO) run ./cmd/painter-video $(ARGS)

list-styles:
	$(GO) run ./cmd/painter -list-styles

list-presets:
	$(GO) run ./cmd/painter-folder -list-presets

check-video-deps:
	@command -v ffmpeg >/dev/null 2>&1 || { printf '%s\n' 'ffmpeg is required but was not found on PATH.'; exit 1; }
	@command -v ffprobe >/dev/null 2>&1 || { printf '%s\n' 'ffprobe is required but was not found on PATH.'; exit 1; }
	@printf '%s\n' 'ffmpeg and ffprobe are available.'

clean:
	rm -rf -- "$(BIN_DIR)"