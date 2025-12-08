#!/bin/bash
# FacePass Installation Script

set -e

INSTALL_PATH="/usr/local/bin"
CONFIG_PATH="/etc/facepass"
DATA_PATH="/var/lib/facepass"

echo "FacePass Installer"
echo "=================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (sudo ./install.sh)"
    exit 1
fi

# Check for required binaries
if [ ! -f "bin/facepass" ] || [ ! -f "bin/facepass-pam" ]; then
    echo "Error: Binaries not found. Please run 'make build' first."
    exit 1
fi

echo "Installing binaries..."
install -m 755 bin/facepass "$INSTALL_PATH/"
install -m 755 bin/facepass-pam "$INSTALL_PATH/"

echo "Creating configuration directory..."
mkdir -p "$CONFIG_PATH"
cp configs/facepass.yaml "$CONFIG_PATH/"
chmod 644 "$CONFIG_PATH/facepass.yaml"

echo "Creating data directory..."
mkdir -p "$DATA_PATH"
chmod 700 "$DATA_PATH"

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "  1. Enroll your face: facepass enroll \$USER"
echo "  2. Test recognition: facepass test \$USER"
echo "  3. Enable PAM (optional): sudo ./scripts/install-pam.sh"
echo ""
