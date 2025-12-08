# FacePass Quick Start Guide

Get started with FacePass development in 5 minutes!

---

## Prerequisites Check

```bash
# Check Go installation
go version  # Should be 1.21+

# Check GCC
gcc --version

# Check if you have required libraries
pkg-config --exists dlib && echo "dlib: ✓" || echo "dlib: ✗ INSTALL NEEDED"
```

---

## One-Command Setup (Ubuntu/Debian)

```bash
# Install all dependencies
sudo apt update && sudo apt install -y \
    build-essential \
    pkg-config \
    libdlib-dev \
    libblas-dev \
    liblapack-dev \
    libpam0g-dev \
    v4l-utils \
    git

# Verify installation
pkg-config --modversion dlib
```

---

## Create Project

```bash
# Create project directory
mkdir -p ~/projects/facepass
cd ~/projects/facepass

# Initialize Go module
go mod init github.com/yourusername/facepass

# Create directory structure
mkdir -p cmd/{facepass,facepass-pam}
mkdir -p pkg/{recognition,storage,config,pam,camera,liveness}
mkdir -p configs models scripts pam-config

# Install Go dependencies
go get -u github.com/Kagami/go-face
go get -u gopkg.in/yaml.v3
go get -u github.com/sirupsen/logrus
go get -u golang.org/x/crypto/nacl/secretbox
```

---

## Create Your First File

```bash
cat > cmd/facepass/main.go << 'EOF'
package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("FacePass v0.1.0")
    fmt.Println("Face Recognition Authentication for Linux\n")

    if len(os.Args) < 2 {
        printUsage()
        return
    }

    command := os.Args[1]

    switch command {
    case "enroll":
        if len(os.Args) < 3 {
            fmt.Println("Error: username required")
            fmt.Println("Usage: facepass enroll <username>")
            return
        }
        fmt.Printf("Enrolling user: %s (not implemented yet)\n", os.Args[2])

    case "test":
        if len(os.Args) < 3 {
            fmt.Println("Error: username required")
            fmt.Println("Usage: facepass test <username>")
            return
        }
        fmt.Printf("Testing recognition for: %s (not implemented yet)\n", os.Args[2])

    case "list":
        fmt.Println("Enrolled users: (not implemented yet)")

    case "remove":
        if len(os.Args) < 3 {
            fmt.Println("Error: username required")
            fmt.Println("Usage: facepass remove <username>")
            return
        }
        fmt.Printf("Removing user: %s (not implemented yet)\n", os.Args[2])

    case "version":
        fmt.Println("FacePass v0.1.0")

    default:
        fmt.Printf("Unknown command: %s\n", command)
        printUsage()
    }
}

func printUsage() {
    fmt.Println("Usage: facepass [command]")
    fmt.Println("\nCommands:")
    fmt.Println("  enroll <username>   Enroll a new face")
    fmt.Println("  test <username>     Test face recognition")
    fmt.Println("  remove <username>   Remove face data")
    fmt.Println("  list                List enrolled users")
    fmt.Println("  version             Show version")
    fmt.Println("\nExamples:")
    fmt.Println("  facepass enroll john")
    fmt.Println("  facepass test john")
}
EOF
```

---

## Create Makefile

```bash
cat > Makefile << 'EOF'
.PHONY: all build clean test run install

BINARY_CLI=facepass
BINARY_PAM=facepass-pam

all: build

build:
	@echo "Building FacePass..."
	@mkdir -p bin
	go build -o bin/$(BINARY_CLI) ./cmd/facepass
	@echo "✓ Build complete: bin/$(BINARY_CLI)"

clean:
	rm -rf bin/
	go clean

test:
	go test -v ./...

run:
	go run ./cmd/facepass

install: build
	@echo "Installing FacePass..."
	sudo install -m 755 bin/$(BINARY_CLI) /usr/local/bin/
	@echo "✓ Installed to /usr/local/bin/$(BINARY_CLI)"

dev:
	@echo "Starting development mode..."
	@echo "Watching for changes..."
	@while true; do \
		make build && ./bin/$(BINARY_CLI) version; \
		inotifywait -qre close_write .; \
	done
EOF
```

---

## Create Config File

```bash
cat > configs/facepass.yaml << 'EOF'
# FacePass Configuration

camera:
  device: /dev/video0
  width: 640
  height: 480
  fps: 30
  prefer_ir: true

recognition:
  confidence_threshold: 0.6
  tolerance: 0.4
  model_path: ~/.local/share/facepass/models

liveness_detection:
  level: standard  # basic, standard, strict, paranoid
  blink_required: true
  consistency_check: true
  challenge_response: false
  min_liveness_score: 0.7

auth:
  timeout: 10
  max_attempts: 3
  fallback_enabled: true

storage:
  data_dir: ~/.local/share/facepass
  encryption_enabled: true

logging:
  level: info
  file: ~/.local/share/facepass/facepass.log
EOF
```

---

## Test Your Setup

```bash
# Build the project
make build

# Run it
./bin/facepass version

# Expected output:
# FacePass v0.1.0
# Face Recognition Authentication for Linux
#
# FacePass v0.1.0

# Test CLI interface
./bin/facepass
./bin/facepass enroll testuser
./bin/facepass list
```

---

## Next Steps

1. **Implement Camera Module** (`pkg/camera/camera.go`)
   - Open V4L2 device
   - Capture frames
   - IR emitter detection/control

2. **Implement Recognition Module** (`pkg/recognition/recognizer.go`)
   - Load go-face
   - Detect faces
   - Generate embeddings
   - Compare faces

3. **Implement Storage Module** (`pkg/storage/storage.go`)
   - Save/load face embeddings
   - Encryption
   - Per-user storage

4. **Implement Liveness Detection** (`pkg/liveness/detector.go`)
   - Blink detection
   - Multi-frame consistency
   - Challenge-response

5. **Wire Everything Together**
   - Update `cmd/facepass/main.go` with actual implementations
   - Add enrollment flow
   - Add authentication flow

---

## Useful Commands During Development

```bash
# Quick rebuild and test
make build && ./bin/facepass enroll $USER

# Run tests
make test

# Check camera devices
v4l2-ctl --list-devices

# Test camera capture
ffplay /dev/video0

# Monitor logs (once logging is implemented)
tail -f ~/.local/share/facepass/facepass.log

# Clean and rebuild
make clean && make build
```

---

## Development Tips

- **Start simple:** Get basic face detection working first, then add features
- **Test incrementally:** Test each module independently before integration
- **Use sample images:** Test with static images before using camera
- **Log everything:** Add verbose logging during development
- **Handle errors:** Always check errors, especially for camera/file operations

---

## Troubleshooting

### "cannot find package github.com/Kagami/go-face"
```bash
go get -u github.com/Kagami/go-face
```

### "dlib not found"
```bash
sudo apt install libdlib-dev  # Ubuntu/Debian
sudo dnf install dlib-devel   # Fedora
```

### "camera permission denied"
```bash
sudo usermod -a -G video $USER
# Logout and login again
```

### "go-face build fails"
```bash
# Make sure you have C++ compiler and dlib
sudo apt install build-essential libdlib-dev libblas-dev liblapack-dev
```

---

## Ready to Code!

You now have a working FacePass project structure. Start implementing the modules in this order:

1. `pkg/camera/` - Camera access
2. `pkg/recognition/` - Face detection and recognition
3. `pkg/storage/` - Data persistence
4. `pkg/liveness/` - Anti-spoofing
5. `pkg/pam/` - PAM integration (last)

Good luck!
