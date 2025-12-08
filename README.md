# FacePass

Face Recognition Authentication for Linux using PAM.

## Overview

FacePass is a face recognition authentication system for Linux that integrates with PAM (Pluggable Authentication Modules). It supports IR cameras, multi-angle enrollment, and anti-spoofing measures.

## Features

- **Face Recognition Authentication** - Log in with your face
- **IR Camera Support** - Works with infrared cameras for better low-light performance
- **Multi-Angle Enrollment** - Captures 5-7 angles for improved accuracy
- **Anti-Spoofing** - Liveness detection to prevent photo/video attacks
- **PAM Integration** - Works with any PAM-enabled application (login, sudo, etc.)
- **Password Fallback** - Gracefully falls back to password if face recognition fails
- **Encrypted Storage** - Face embeddings are encrypted at rest

## Quick Start

See [QUICK_START.md](QUICK_START.md) for a 5-minute setup guide.

## Documentation

- [PROJECT_PLAN.md](PROJECT_PLAN.md) - Full project plan and setup guide
- [QUICK_START.md](QUICK_START.md) - Quick start guide
- [ARCHITECTURE.md](ARCHITECTURE.md) - Technical architecture

## Requirements

### Runtime
- Linux (Ubuntu 20.04+, Fedora 35+, Arch)
- Webcam/camera with V4L2 support
- PAM-enabled system
- 2GB+ RAM

### Development
- Go 1.21+
- GCC/G++ compiler
- dlib 19.24+
- PAM development headers

## Installation

```bash
# Install dependencies (Ubuntu/Debian)
sudo apt install -y build-essential pkg-config libdlib-dev libblas-dev liblapack-dev libpam0g-dev v4l-utils

# Build
make build

# Install
sudo make install

# Enroll your face
facepass enroll $USER

# Test recognition
facepass test $USER
```

## Usage

```bash
# Enroll a new user (captures 5-7 angles)
facepass enroll <username>

# Add more face angles to existing enrollment
facepass add-face <username>

# Test face recognition
facepass test <username>

# List enrolled users
facepass list

# Remove a user's face data
facepass remove <username>

# Show configuration
facepass config
```

## Configuration

Configuration file: `/etc/facepass/facepass.yaml` or `~/.config/facepass/facepass.yaml`

See [configs/facepass.yaml](configs/facepass.yaml) for all options.

## License

MIT License
