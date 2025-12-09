# FacePass - Face Recognition PAM Authentication for Linux

FacePass is a secure face recognition authentication module for Linux that integrates with PAM (Pluggable Authentication Modules). It provides a Windows Hello-like experience for Linux systems, with support for IR cameras, liveness detection, and GPU acceleration.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)
![Platform](https://img.shields.io/badge/platform-linux-green.svg)

## Features

- **Face Recognition Authentication** - Authenticate using your face instead of passwords
- **Multi-Angle Enrollment** - Capture 5 angles for robust recognition
- **Liveness Detection** - Prevent photo/video spoofing attacks
  - Blink detection
  - Frame consistency checking
  - Micro-movement detection
  - Challenge-response system
- **IR Camera Support** - Works with infrared cameras (linux-enable-ir-emitter integration)
- **Encrypted Storage** - Face embeddings encrypted at rest using NaCl
- **GPU Acceleration** - Optional acceleration via:
  - AMD ROCm (tested and supported)
  - NVIDIA CUDA (community testing needed)
  - Intel OpenVINO (community testing needed)
- **Password Fallback** - Automatic fallback to password on timeout/failure

## Quick Start

### Prerequisites

```bash
# Ubuntu/Debian
sudo apt install build-essential libdlib-dev libblas-dev liblapack-dev libpam0g-dev v4l-utils ffmpeg

# Fedora
sudo dnf install dlib-devel blas-devel lapack-devel pam-devel v4l-utils ffmpeg

# Arch
sudo pacman -S dlib blas lapack pam v4l-utils ffmpeg
```

### Installation

```bash
# Clone the repository
git clone https://github.com/MrCodeEU/facepass.git
cd facepass

# Build
make build

# Install (requires sudo)
sudo ./scripts/install.sh
```

### Enroll Your Face

```bash
# Enroll with 5 angles
facepass enroll $USER

# Test recognition
facepass test $USER
```

### Enable PAM (Optional)

See [PAM Configuration](#pam-configuration) for detailed instructions.

## Documentation

- [PROJECT_PLAN.md](PROJECT_PLAN.md) - Full project plan and implementation phases
- [QUICK_START.md](QUICK_START.md) - 5-minute quick start guide
- [ARCHITECTURE.md](ARCHITECTURE.md) - Technical architecture documentation
- [TESTING.md](TESTING.md) - Testing guide and workflow
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [SECURITY.md](SECURITY.md) - Security policy and reporting

## Building

### Standard Build (CPU)

```bash
make build
```

### GPU Accelerated Builds

```bash
# AMD ROCm (recommended for AMD GPUs)
make build-rocm

# NVIDIA CUDA (needs community testing)
make build-cuda

# Intel OpenVINO (needs community testing)
make build-openvino

# Auto-detect and build
make build-accelerated

# Check available GPUs
make detect-gpu
```

### Build Requirements

- Go 1.21 or higher
- GCC/G++ compiler (for CGO)
- dlib 19.24+ with development headers
- PAM development headers

## Usage

### CLI Commands

```bash
# Face enrollment
facepass enroll <username>       # Enroll with 5 angles
facepass add-face <username>     # Add more angles to existing enrollment

# Testing
facepass test <username>         # Test face recognition

# Management
facepass list                    # List enrolled users
facepass remove <username>       # Remove user enrollment
facepass cameras                 # List available cameras

# Configuration
facepass config                  # Show current configuration
facepass version                 # Show version information
```

### Configuration

Configuration file: `/etc/facepass/facepass.yaml` or `~/.config/facepass/facepass.yaml`

```yaml
# Camera settings
camera:
  device: /dev/video0
  width: 640
  height: 480
  prefer_ir: true

# Recognition settings
recognition:
  confidence_threshold: 0.6
  tolerance: 0.4
  model_path: ~/.local/share/facepass/models

# Liveness detection
liveness_detection:
  level: standard  # basic, standard, strict, paranoid
  blink_required: true
  min_liveness_score: 0.7

# Authentication
auth:
  timeout: 10
  max_attempts: 3
  fallback_enabled: true

# Storage
storage:
  data_dir: ~/.local/share/facepass
  encryption_enabled: true

# GPU Acceleration
acceleration:
  backend: auto  # auto, cpu, rocm, cuda, openvino
  fallback_to_cpu: true
```

## PAM Configuration

> **WARNING**: Test thoroughly before enabling system-wide PAM. Always keep a root terminal open. Incorrect configuration can lock you out!

### Ubuntu/Debian

Edit `/etc/pam.d/common-auth` and add BEFORE `pam_unix.so`:

```
auth    [success=2 default=ignore]    pam_exec.so expose_authtok /usr/local/bin/facepass-pam
```

### Fedora/RHEL/CentOS

Edit `/etc/pam.d/system-auth` and add BEFORE `pam_unix.so`:

```
auth    sufficient    pam_exec.so expose_authtok /usr/local/bin/facepass-pam
```

### Arch Linux

Edit `/etc/pam.d/system-auth` and add BEFORE `pam_unix.so`:

```
auth    sufficient    pam_exec.so expose_authtok /usr/local/bin/facepass-pam
```

### For sudo Only (Safer Testing)

Edit `/etc/pam.d/sudo`:

```
auth    sufficient    pam_exec.so expose_authtok /usr/local/bin/facepass-pam
```

## Security

### Liveness Detection Levels

| Level | Checks | Use Case |
|-------|--------|----------|
| basic | Blink + consistency | Low-security environments |
| standard | + movement detection | Normal desktop use |
| strict | + challenge-response, IR analysis | Secure workstations |
| paranoid | All checks + texture analysis | High-security environments |

### Encryption

- Face embeddings encrypted with NaCl secretbox (XSalsa20 + Poly1305)
- Machine-specific key derivation (data tied to hardware)
- Secure storage with 0700 permissions

### Anti-Spoofing Protection

- **Photo attacks**: Blink detection, movement analysis
- **Video attacks**: Frame consistency checking, micro-movements
- **Screen attacks**: Texture/moire pattern analysis (strict+)
- **IR reflection**: Analysis for IR cameras

## GPU Acceleration

### AMD ROCm (Tested and Supported)

```bash
# Install ROCm: https://rocm.docs.amd.com/
make build-rocm
```

### NVIDIA CUDA

```bash
# Install CUDA: https://developer.nvidia.com/cuda-downloads
make build-cuda
```

> **Note**: CUDA support needs community testing. Please report issues!

### Intel OpenVINO

```bash
# Install OpenVINO: https://docs.openvino.ai/
make build-openvino
```

> **Note**: OpenVINO support needs community testing. Please report issues!

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run with race detection
make test-race

# Run benchmarks
make test-benchmark
```

See [TESTING.md](TESTING.md) for detailed testing documentation.

## Troubleshooting

### Camera not found

```bash
# List available cameras
facepass cameras
v4l2-ctl --list-devices

# Check permissions
ls -la /dev/video*
sudo usermod -aG video $USER
```

### Face not recognized

1. Ensure good lighting
2. Re-enroll: `facepass enroll <username>`
3. Lower tolerance in config (e.g., 0.5)
4. Add more angles: `facepass add-face <username>`

### Liveness check failing

1. Blink clearly when prompted
2. Ensure face is centered
3. Lower liveness level in config
4. Remove reflective glasses

### PAM Issues

1. Keep a root terminal open when testing
2. Test manually: `PAM_USER=$USER /usr/local/bin/facepass-pam`
3. Check logs: `tail -f /var/log/facepass.log`

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Community Testing Needed

We especially need help testing:
- NVIDIA CUDA acceleration
- Intel OpenVINO acceleration
- Various Linux distributions
- Different camera hardware

## Security

Please report security vulnerabilities according to our [SECURITY.md](SECURITY.md) policy.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [dlib](http://dlib.net/) - Face recognition library
- [go-face](https://github.com/Kagami/go-face) - Go bindings for dlib
- [linux-enable-ir-emitter](https://github.com/EmixamPP/linux-enable-ir-emitter) - IR camera support
- [logrus](https://github.com/sirupsen/logrus) - Logging library
