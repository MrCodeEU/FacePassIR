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

## System Dependencies

Install the required system libraries (dlib, blas, lapack, pam) for your OS.

### Ubuntu / Debian

```bash
sudo apt update && sudo apt install -y \
    build-essential \
    pkg-config \
    libdlib-dev \
    libblas-dev \
    liblapack-dev \
    libpam0g-dev \
    libjpeg-dev \
    v4l-utils \
    ffmpeg \
    git
```

### Fedora / RHEL

```bash
sudo dnf install -y \
    @development-tools \
    pkgconf-pkg-config \
    dlib-devel \
    blas-devel \
    lapack-devel \
    pam-devel \
    libjpeg-turbo-devel \
    v4l-utils \
    git

# Optional: Install ffmpeg for camera testing if not present
# Note: Fedora often comes with ffmpeg-free. If you have conflicts, skip this.
# sudo dnf install -y ffmpeg
```

### Verify Installation

```bash
pkg-config --modversion dlib
```

---

## Project Setup

```bash
# Clone the repository (if not already done)
# git clone https://github.com/MrCodeEU/FacePassIR.git
# cd FacePassIR

# Download Go dependencies
go mod download

# Download required models
make build
./bin/facepass download-models
```

---


## Test Your Setup

```bash
# Build the project
make build

# Run it
./bin/facepass version

# Expected output:
# FacePass v0.X.0
# Face Recognition Authentication for Linux
#
# FacePass v0.X.0

# Test CLI interface
./bin/facepass
./bin/facepass enroll testuser
./bin/facepass list
```

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
