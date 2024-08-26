# Go source file
SRC=VeilTransfer.go

# Output binary names
OUT_LINUX=veiltransfer_linux
OUT_WINDOWS=veiltransfer_windows.exe
OUT_MACOS=veiltransfer_macos

# Default target
all: deps linux windows macos

# Ensure dependencies are installed
deps:
	go mod tidy
	go mod download

# Build for Linux
linux: deps
	GOOS=linux GOARCH=amd64 go build -o $(OUT_LINUX) $(SRC)
	strip $(OUT_LINUX)

# Build for Windows
windows: deps
	GOOS=windows GOARCH=amd64 go build -o $(OUT_WINDOWS) $(SRC)

# Build for macOS
macos: deps
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o $(OUT_MACOS) $(SRC)

# Clean the build artifacts
clean:
	rm -f $(OUT_LINUX) $(OUT_WINDOWS) $(OUT_MACOS)

# List files
list:
	ls -la

.PHONY: all deps linux windows macos clean list
