#!/bin/bash
# FacePass Development Setup Script

set -e

echo "FacePass Development Setup"
echo "=========================="

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    echo "Could not detect OS"
    exit 1
fi

echo "Detected OS: $OS"

# Install dependencies based on OS
case $OS in
    ubuntu|debian)
        echo "Installing dependencies for Ubuntu/Debian..."
        sudo apt update
        sudo apt install -y \
            build-essential \
            pkg-config \
            libdlib-dev \
            libblas-dev \
            liblapack-dev \
            libpam0g-dev \
            v4l-utils \
            git
        ;;
    fedora)
        echo "Installing dependencies for Fedora..."
        sudo dnf groupinstall -y "Development Tools"
        sudo dnf install -y \
            dlib-devel \
            blas-devel \
            lapack-devel \
            pam-devel \
            v4l-utils
        ;;
    arch)
        echo "Installing dependencies for Arch..."
        sudo pacman -S --noconfirm \
            base-devel \
            dlib \
            blas \
            lapack \
            pam \
            v4l-utils
        ;;
    *)
        echo "Unsupported OS: $OS"
        echo "Please install the following packages manually:"
        echo "  - build-essential / base-devel"
        echo "  - dlib development libraries"
        echo "  - BLAS and LAPACK"
        echo "  - PAM development headers"
        echo "  - v4l-utils"
        exit 1
        ;;
esac

# Check Go installation
if ! command -v go &> /dev/null; then
    echo ""
    echo "Go is not installed. Please install Go 1.21+ from:"
    echo "  https://go.dev/dl/"
    echo ""
    echo "Or run:"
    echo "  wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz"
    echo "  sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz"
    echo "  echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.bashrc"
    echo "  source ~/.bashrc"
    exit 1
fi

echo ""
echo "Go version: $(go version)"

# Install Go dependencies
echo ""
echo "Installing Go dependencies..."
go mod download

# Verify dlib
echo ""
echo "Verifying dlib installation..."
if pkg-config --exists dlib; then
    echo "dlib version: $(pkg-config --modversion dlib)"
else
    echo "Warning: dlib not found via pkg-config"
    echo "The build may fail. Please ensure dlib is properly installed."
fi

# Create local data directories
echo ""
echo "Creating local data directories..."
mkdir -p ~/.local/share/facepass/{models,users,logs}

echo ""
echo "Development setup complete!"
echo ""
echo "Next steps:"
echo "  1. Build: make build"
echo "  2. Run: ./bin/facepass version"
echo "  3. Test: make test"
echo ""
