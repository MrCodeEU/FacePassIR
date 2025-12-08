#!/bin/bash
# FacePass Uninstallation Script

set -e

INSTALL_PATH="/usr/local/bin"
CONFIG_PATH="/etc/facepass"
DATA_PATH="/var/lib/facepass"

echo "FacePass Uninstaller"
echo "===================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (sudo ./uninstall.sh)"
    exit 1
fi

echo "Removing binaries..."
rm -f "$INSTALL_PATH/facepass"
rm -f "$INSTALL_PATH/facepass-pam"

echo "Removing configuration..."
rm -rf "$CONFIG_PATH"

echo ""
echo "Uninstallation complete!"
echo ""
echo "Note: Face data has been preserved in $DATA_PATH"
echo "To remove all data, run: sudo rm -rf $DATA_PATH"
echo ""
echo "If you enabled PAM integration, please manually remove"
echo "the FacePass entries from /etc/pam.d/common-auth"
echo ""
