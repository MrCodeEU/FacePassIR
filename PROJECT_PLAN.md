# FacePass - Project Plan & Setup Guide

## System Requirements

### Runtime Requirements
- Linux distribution (Ubuntu 20.04+, Fedora 35+, Arch, etc.)
- Webcam/camera with V4L2 support
- PAM-enabled system
- 2GB+ RAM (for face recognition models)

### Development Requirements
- Go 1.21 or higher
- GCC/G++ compiler (for CGO)
- dlib 19.24+ (C++ library for go-face)
- pkg-config
- PAM development headers (libpam0g-dev on Debian/Ubuntu)
- Git

### Optional
- OpenCV 4.x (for gocv fallback)
- CUDA (for GPU acceleration, if needed)

---

## Project Architecture

```
facepass/
├── cmd/
│   ├── facepass/              # Main CLI tool
│   │   └── main.go            # Enrollment, management, testing
│   └── facepass-pam/          # PAM module executable
│       └── main.go            # Called by PAM during auth
├── pkg/
│   ├── recognition/           # Face recognition engine
│   │   ├── detector.go        # Face detection
│   │   ├── recognizer.go      # Face recognition/comparison
│   │   └── models.go          # Model loading/management
│   ├── storage/               # Face data persistence
│   │   ├── storage.go         # Interface & implementation
│   │   ├── encryption.go      # Encrypt face embeddings
│   │   └── user.go            # User face data structure
│   ├── config/                # Configuration management
│   │   ├── config.go          # Config loading/parsing
│   │   └── defaults.go        # Default configuration
│   ├── pam/                   # PAM integration
│   │   ├── protocol.go        # PAM protocol implementation
│   │   └── auth.go            # Authentication logic
│   └── camera/                # Camera interface
│       ├── camera.go          # V4L2 camera access
│       └── capture.go         # Frame capture logic
├── configs/
│   └── facepass.yaml          # Default configuration file
├── models/                     # Pre-trained models directory
│   └── .gitkeep
├── scripts/
│   ├── install.sh             # System installation script
│   ├── uninstall.sh           # Removal script
│   └── setup-dev.sh           # Development setup
├── pam-config/
│   └── facepass                # PAM configuration snippet
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── LICENSE
```

---

## Implementation Phases

### Phase 1: Project Setup & Foundation (Day 1)
- [ ] Initialize Go module
- [ ] Set up project structure
- [ ] Install dependencies (go-face, dlib)
- [ ] Create Makefile for build automation
- [ ] Basic configuration system
- [ ] Logging infrastructure

### Phase 2: Face Recognition Core (Days 2-3)
- [ ] Integrate go-face library
- [ ] Implement face detection
- [ ] Implement face recognition (embedding generation)
- [ ] Test face comparison/matching
- [ ] Add confidence threshold configuration
- [ ] Handle multiple faces in frame
- [ ] Multi-angle face capture (5-7 angles)
- [ ] Face embedding averaging/storage for multiple angles

### Phase 2.5: IR Camera Integration (Day 4)
- [ ] Detect linux-enable-ir-emitter availability
- [ ] Auto-enable/disable IR emitter for capture
- [ ] Test with IR camera (grayscale image processing)
- [ ] Fallback to regular camera if IR not available

### Phase 2.6: Anti-Spoofing - Tier 1 (Day 4-5)
- [ ] Blink detection using eye aspect ratio (EAR)
- [ ] Multi-frame capture and consistency checking
- [ ] Detect static images (no movement between frames)
- [ ] Basic liveness score calculation

### Phase 3: Camera Integration (Day 3)
- [ ] Camera device detection
- [ ] Frame capture from V4L2 device
- [ ] Image preprocessing
- [ ] Error handling for camera issues

### Phase 4: Storage System (Days 4-5)
- [ ] Design face data schema (username -> embeddings)
- [ ] Implement file-based storage (JSON or binary)
- [ ] Encryption for stored embeddings
- [ ] User enrollment data management
- [ ] Storage location: /var/lib/facepass/ or ~/.local/share/facepass/

### Phase 5: CLI Tool (Days 5-6)
- [ ] `facepass enroll <username>` - Enroll new face
- [ ] `facepass test <username>` - Test recognition
- [ ] `facepass remove <username>` - Remove face data
- [ ] `facepass list` - List enrolled users
- [ ] `facepass config` - Show/edit configuration
- [ ] Interactive enrollment with feedback

### Phase 6: PAM Module (Days 7-9)
- [ ] Understand PAM authentication flow
- [ ] Create PAM executable (facepass-pam)
- [ ] Implement PAM protocol (read/write to PAM)
- [ ] Integration with face recognition
- [ ] Timeout handling (fallback to password)
- [ ] Logging for authentication attempts
- [ ] Security hardening (run as appropriate user)

### Phase 7: Configuration System (Day 9)
- [ ] YAML-based configuration
- [ ] Camera device selection
- [ ] Confidence threshold tuning
- [ ] Timeout settings
- [ ] Model path configuration
- [ ] Storage location configuration

### Phase 8: Installation & Integration (Days 10-11)
- [ ] Installation script (copy binaries, set permissions)
- [ ] PAM configuration file creation
- [ ] Model download/installation
- [ ] Systemd service (if needed)
- [ ] Uninstall script

### Phase 9: Testing & Refinement (Days 12-14)
- [ ] Unit tests for core functions
- [ ] Integration tests for PAM flow
- [ ] Test on multiple Linux distros
- [ ] Performance optimization
- [ ] Security audit
- [ ] Edge case handling (no camera, poor lighting, etc.)
- [ ] Anti-spoofing tests (photos, videos, screens)

### Phase 9.5: Advanced Anti-Spoofing (Days 15-16, Optional)
- [ ] Challenge-response system (random head movements)
- [ ] Micro-movement detection
- [ ] IR reflection analysis (if IR camera available)
- [ ] Texture analysis (moire patterns, pixelation)
- [ ] Configurable liveness detection levels (basic/strict/paranoid)

### Phase 10: Documentation (Day 17-18)
- [ ] README with installation instructions
- [ ] Usage guide
- [ ] Configuration reference
- [ ] Troubleshooting guide
- [ ] Security considerations document

---

## Development Setup Commands

### 1. Install System Dependencies

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install -y build-essential pkg-config
sudo apt install -y libdlib-dev libblas-dev liblapack-dev
sudo apt install -y libpam0g-dev
sudo apt install -y v4l-utils  # For camera testing
```

**Fedora:**
```bash
sudo dnf groupinstall "Development Tools"
sudo dnf install -y dlib-devel blas-devel lapack-devel
sudo dnf install -y pam-devel
sudo dnf install -y v4l-utils
```

**Arch:**
```bash
sudo pacman -S base-devel dlib blas lapack pam v4l-utils
```

### 2. Install Go (if not already installed)
```bash
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 3. Create Project Structure
```bash
mkdir -p facepass
cd facepass

# Create directory structure
mkdir -p cmd/{facepass,facepass-pam}
mkdir -p pkg/{recognition,storage,config,pam,camera}
mkdir -p configs models scripts pam-config

# Initialize Go module
go mod init github.com/yourusername/facepass
```

### 4. Install Go Dependencies
```bash
# Primary face recognition library
go get -u github.com/Kagami/go-face

# Configuration management
go get -u gopkg.in/yaml.v3

# Logging
go get -u github.com/sirupsen/logrus

# Encryption
go get -u golang.org/x/crypto/nacl/secretbox

# (Optional) gocv as fallback
# go get -u gocv.io/x/gocv
```

### 5. Create Initial Files

**Create Makefile:**
```makefile
.PHONY: all build clean install uninstall test

BINARY_CLI=facepass
BINARY_PAM=facepass-pam
INSTALL_PATH=/usr/local/bin
PAM_PATH=/usr/lib/security
CONFIG_PATH=/etc/facepass
DATA_PATH=/var/lib/facepass

all: build

build:
	@echo "Building FacePass..."
	go build -o bin/$(BINARY_CLI) ./cmd/facepass
	go build -o bin/$(BINARY_PAM) ./cmd/facepass-pam
	@echo "Build complete!"

clean:
	rm -rf bin/
	go clean

install: build
	@echo "Installing FacePass..."
	sudo install -m 755 bin/$(BINARY_CLI) $(INSTALL_PATH)/
	sudo install -m 755 bin/$(BINARY_PAM) $(INSTALL_PATH)/
	sudo mkdir -p $(CONFIG_PATH)
	sudo mkdir -p $(DATA_PATH)
	sudo cp configs/facepass.yaml $(CONFIG_PATH)/
	sudo chmod 644 $(CONFIG_PATH)/facepass.yaml
	sudo chmod 700 $(DATA_PATH)
	@echo "Installation complete!"
	@echo "Run 'sudo make install-pam' to enable PAM integration"

install-pam:
	@echo "Installing PAM configuration..."
	sudo cp pam-config/facepass /etc/pam.d/
	@echo "PAM configuration installed!"
	@echo "Edit /etc/pam.d/common-auth or /etc/pam.d/system-auth to enable FacePass"

uninstall:
	@echo "Uninstalling FacePass..."
	sudo rm -f $(INSTALL_PATH)/$(BINARY_CLI)
	sudo rm -f $(INSTALL_PATH)/$(BINARY_PAM)
	sudo rm -rf $(CONFIG_PATH)
	@echo "Face data preserved in $(DATA_PATH)"
	@echo "Run 'sudo rm -rf $(DATA_PATH)' to remove all data"

test:
	go test -v ./...

run-cli:
	go run ./cmd/facepass

run-pam:
	go run ./cmd/facepass-pam
```

### 6. Create Initial Configuration File

**configs/facepass.yaml:**
```yaml
# FacePass Configuration

# Camera settings
camera:
  device: /dev/video0
  width: 640
  height: 480

# Recognition settings
recognition:
  confidence_threshold: 0.6  # Lower = more strict (less false positives)
  tolerance: 0.4             # Distance tolerance for face matching
  model_path: /usr/share/facepass/models

# Authentication settings
auth:
  timeout: 10                # Seconds before fallback to password
  max_attempts: 3            # Max face recognition attempts
  fallback_enabled: true     # Allow password fallback

# Storage settings
storage:
  data_dir: /var/lib/facepass
  encryption_enabled: true

# Logging
logging:
  level: info                # debug, info, warn, error
  file: /var/log/facepass.log
```

### 7. Development Workflow

**Start developing:**
```bash
# Create main CLI entry point
cat > cmd/facepass/main.go << 'EOF'
package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("FacePass - Face Recognition Authentication")
        fmt.Println("Usage: facepass [command]")
        fmt.Println("\nCommands:")
        fmt.Println("  enroll <username>   Enroll a new face")
        fmt.Println("  test <username>     Test recognition")
        fmt.Println("  remove <username>   Remove face data")
        fmt.Println("  list                List enrolled users")
        return
    }

    command := os.Args[1]
    fmt.Printf("Command: %s (not implemented yet)\n", command)
}
EOF

# Test build
make build

# Run
./bin/facepass
```

---

## Building the Project

```bash
# Development build
make build

# Run tests
make test

# Install system-wide (requires sudo)
sudo make install

# Install PAM integration (requires sudo)
sudo make install-pam
```

---

## Installation Process (End User)

```bash
# 1. Install dependencies
sudo apt install libdlib-dev libpam0g-dev  # Ubuntu/Debian

# 2. Install FacePass
sudo make install

# 3. Enroll your face
facepass enroll $USER

# 4. Test recognition
facepass test $USER

# 5. Enable PAM authentication (optional)
sudo make install-pam
# Then edit /etc/pam.d/common-auth to add FacePass
```

---

## IR Camera Integration Details

### linux-enable-ir-emitter Integration

**Detection and Control:**
```bash
# Check if IR emitter is available
$ ls /sys/bus/usb/drivers/uvcvideo/*/video4linux/video*/

# Enable IR emitter (usually done by linux-enable-ir-emitter)
$ echo 1 > /sys/class/video4linux/video0/ir_emitter

# Or use linux-enable-ir-emitter directly
$ linux-enable-ir-emitter enable
```

**Go Implementation:**
```go
type IREmitter struct {
    Available bool
    Device    string
    Enabled   bool
}

func detectIREmitter() *IREmitter {
    // Check for IR emitter sysfs entry
    devices, _ := filepath.Glob("/sys/class/video4linux/video*/ir_emitter")

    if len(devices) > 0 {
        return &IREmitter{
            Available: true,
            Device:    devices[0],
            Enabled:   false,
        }
    }

    // Alternatively, check if linux-enable-ir-emitter is installed
    _, err := exec.LookPath("linux-enable-ir-emitter")
    if err == nil {
        return &IREmitter{
            Available: true,
            Device:    "managed-by-tool",
            Enabled:   false,
        }
    }

    return &IREmitter{Available: false}
}

func (ir *IREmitter) Enable() error {
    if !ir.Available {
        return errors.New("IR emitter not available")
    }

    // Try using linux-enable-ir-emitter first
    cmd := exec.Command("linux-enable-ir-emitter", "enable")
    if err := cmd.Run(); err == nil {
        ir.Enabled = true
        return nil
    }

    // Fallback to direct sysfs control
    if ir.Device != "managed-by-tool" {
        return os.WriteFile(ir.Device, []byte("1"), 0644)
    }

    return errors.New("failed to enable IR emitter")
}

func (ir *IREmitter) Disable() error {
    if !ir.Enabled {
        return nil
    }

    cmd := exec.Command("linux-enable-ir-emitter", "disable")
    if err := cmd.Run(); err == nil {
        ir.Enabled = false
        return nil
    }

    if ir.Device != "managed-by-tool" {
        return os.WriteFile(ir.Device, []byte("0"), 0644)
    }

    return errors.New("failed to disable IR emitter")
}
```

### IR Camera Benefits
- **Better low-light performance** - Works in complete darkness
- **Consistent lighting** - IR LED provides uniform illumination
- **Anti-spoofing** - IR reflection patterns differ between real faces and photos/screens
- **Privacy** - IR images are less identifiable than regular photos
- **Depth information** - Some IR cameras provide depth data

### Camera Device Selection
```yaml
camera:
  # Prefer IR camera if available
  prefer_ir: true

  # Device paths (will auto-detect if not specified)
  ir_device: /dev/video2      # Usually video2 for IR on laptops
  rgb_device: /dev/video0     # Regular camera fallback

  # IR emitter control
  ir_emitter_enabled: true
  ir_emitter_tool: linux-enable-ir-emitter  # or 'sysfs'
```

### Enrollment with IR Camera
```go
func enrollWithIR(username string) error {
    // Enable IR emitter
    ir := detectIREmitter()
    if ir.Available {
        ir.Enable()
        defer ir.Disable()

        // Use IR camera device
        camera, err := openCamera("/dev/video2") // IR camera
    } else {
        // Fallback to regular camera
        camera, err := openCamera("/dev/video0")
    }

    // Capture multiple angles
    angles := []string{"front", "left", "right", "up", "down"}
    embeddings := [][]float32{}

    for _, angle := range angles {
        frame := captureFrame(camera)
        embedding := extractFaceEmbedding(frame)
        embeddings = append(embeddings, embedding)
    }

    // Store with encryption
    return storeFaceData(username, embeddings)
}
```

---

## Anti-Spoofing Technical Implementation

### Blink Detection (Tier 1)
```go
// Eye Aspect Ratio (EAR) formula
// EAR = (||p2-p6|| + ||p3-p5||) / (2 * ||p1-p4||)
// Where p1-p6 are eye landmarks
// Blink detected when EAR drops below threshold (~0.2)

func detectBlink(frames []Frame) bool {
    earHistory := []float64{}
    for _, frame := range frames {
        ear := calculateEyeAspectRatio(frame.landmarks)
        earHistory = append(earHistory, ear)
    }

    // Detect significant EAR drop (blink)
    return hasSignificantDrop(earHistory, 0.2)
}
```

### Multi-Frame Consistency (Tier 1)
```go
// Capture 5-10 frames over 2 seconds
// Compare embeddings - should be very similar
// Static photo will have identical embeddings
// Real face will have slight variations

func checkFrameConsistency(embeddings [][]float32) bool {
    // Calculate variance between embeddings
    variance := calculateVariance(embeddings)

    // Too low = static image (photo)
    // Too high = different person or poor capture
    return variance > MIN_VARIANCE && variance < MAX_VARIANCE
}
```

### Challenge-Response (Tier 2)
```go
type Challenge struct {
    Action string  // "turn_left", "turn_right", "look_up", "look_down"
    Angle  float64
}

func performChallenge() bool {
    challenge := randomChallenge()
    showPrompt(challenge.Action) // "Turn your head left"

    frames := captureFrames(30) // Capture 1 second
    movement := analyzeHeadMovement(frames)

    return movement.Direction == challenge.Action &&
           movement.Angle >= challenge.Angle * 0.8
}
```

### IR Reflection Analysis (Tier 3)
```go
// Screens and photos reflect IR differently than skin
// Look for uniform reflection patterns (screens)
// Real skin has varied IR absorption/reflection

func analyzeIRReflection(irFrame []byte) float64 {
    // Calculate histogram of IR intensities
    histogram := calculateHistogram(irFrame)

    // Real skin: varied distribution
    // Screen: peaks at specific intensities (backlight)
    // Photo: more uniform than real skin

    return calculateLivenessScore(histogram)
}
```

### Texture Analysis (Tier 3)
```go
// Detect moire patterns from LCD screens
// Detect pixelation from printed photos

func analyzeTexture(frame Image) bool {
    // FFT to detect regular patterns (screen refresh, pixels)
    fft := fastFourierTransform(frame)

    // Screens show regular frequency patterns
    hasScreenPattern := detectRegularPatterns(fft)

    // High-frequency analysis for pixelation
    hasPixelation := detectPixelation(frame)

    return !hasScreenPattern && !hasPixelation
}
```

### Combined Liveness Score
```go
type LivenessResult struct {
    Blinked       bool
    Consistent    bool
    ChallengePass bool
    IRScore       float64
    TextureOK     bool
    TotalScore    float64
}

func calculateLiveness(config LivenessConfig) LivenessResult {
    result := LivenessResult{}

    // Tier 1 (always enabled)
    result.Blinked = detectBlink(capturedFrames)
    result.Consistent = checkFrameConsistency(embeddings)

    // Tier 2 (if enabled in config)
    if config.EnableChallenge {
        result.ChallengePass = performChallenge()
    }

    // Tier 3 (if IR camera available)
    if config.HasIRCamera {
        result.IRScore = analyzeIRReflection(irFrames)
        result.TextureOK = analyzeTexture(frames[0])
    }

    // Calculate total score (0.0 - 1.0)
    result.TotalScore = weightedAverage(result, config.Weights)

    return result
}
```

### Configuration Levels
```yaml
liveness_detection:
  level: strict  # basic, standard, strict, paranoid

  # basic: blink + consistency
  # standard: + challenge-response
  # strict: + IR analysis (if available)
  # paranoid: all checks + manual review flag

  blink_required: true
  consistency_check: true
  challenge_response: false  # Enable for higher security
  ir_analysis: true           # Auto-enabled if IR camera detected
  texture_analysis: true

  # Thresholds
  min_liveness_score: 0.7     # 0.0 - 1.0
  max_authentication_time: 10 # seconds
```

---

## Design Decisions (Finalized)

1. **Model Storage:** Download on first run + cache for development (avoid re-downloading on reinstall)

2. **Multi-user:** Per-user storage in `~/.local/share/facepass/` (more flexible, better permissions)

3. **Fallback behavior:** Configurable timeout (default 10s), then PAM takes over with password prompt

4. **Security:** Encrypt face embeddings at rest (mandatory)

5. **Camera permissions:** PAM runs as root, so camera access should work fine

6. **Multiple angles per user:**
   - Enroll 5-7 angles during initial setup (front, left, right, up, down)
   - Support `facepass add-face <username>` to add more angles later
   - Integrate with `linux-enable-ir-emitter` for IR camera support
   - IR cameras preferred for better low-light and liveness detection

7. **Anti-spoofing (Liveness Detection):**
   - Tier 1: Blink detection + multi-frame consistency
   - Tier 2: Challenge-response (random head movements)
   - Tier 3: IR-specific analysis (reflection patterns, texture analysis)
   - Tier 4: Depth sensing (if hardware supports)

---

## Additional Implementation Notes

### IR Camera Support
- Detect `linux-enable-ir-emitter` availability
- Auto-enable IR emitter during enrollment/auth
- IR images work with dlib (grayscale processing)
- Better anti-spoofing capabilities

### Multi-Angle Enrollment Flow
```
$ facepass enroll username

Starting enrollment for 'username'...
Please ensure good lighting and face the camera.

[1/5] Look directly at camera... ✓ Captured
[2/5] Turn head slightly left... ✓ Captured
[3/5] Turn head slightly right... ✓ Captured
[4/5] Tilt head slightly up... ✓ Captured
[5/5] Tilt head slightly down... ✓ Captured

Enrollment complete! 5 angles captured.
Testing recognition... ✓ Success

You can add more angles later with: facepass add-face username
```

### Liveness Detection Strategy
1. Capture 5-10 frames over 2 seconds
2. Detect at least one blink (eye aspect ratio tracking)
3. Optional: Challenge-response (random head turn direction)
4. Analyze frame consistency and micro-movements
5. IR-specific checks if IR camera detected

---

## Ready to Start Implementation!

All design decisions are finalized. You can begin development with a clear roadmap.
