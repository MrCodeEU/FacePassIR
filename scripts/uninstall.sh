#!/bin/bash
# FacePass Uninstallation Script
# Removes FacePass from the system

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Installation paths
INSTALL_PATH="/usr/local/bin"
CONFIG_PATH="/etc/facepass"
DATA_PATH="/var/lib/facepass"
MODEL_PATH="/usr/share/facepass"
LOG_FILE="/var/log/facepass.log"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  FacePass Uninstallation${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check for root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Confirm uninstallation
echo -e "${YELLOW}This will remove FacePass from your system.${NC}"
echo ""
echo "The following will be removed:"
echo "  - Binaries: $INSTALL_PATH/facepass, $INSTALL_PATH/facepass-pam"
echo "  - Configuration: $CONFIG_PATH"
echo "  - Models: $MODEL_PATH"
echo "  - Log file: $LOG_FILE"
echo ""
echo -e "${YELLOW}Face data in $DATA_PATH will be PRESERVED by default.${NC}"
echo ""

read -p "Continue with uninstallation? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Uninstallation cancelled."
    exit 0
fi

echo ""

# Step 1: Remove binaries
echo -e "${BLUE}[1/5] Removing binaries...${NC}"
if [ -f "$INSTALL_PATH/facepass" ]; then
    rm -f "$INSTALL_PATH/facepass"
    echo -e "${GREEN}  ✓ Removed $INSTALL_PATH/facepass${NC}"
fi
if [ -f "$INSTALL_PATH/facepass-pam" ]; then
    rm -f "$INSTALL_PATH/facepass-pam"
    echo -e "${GREEN}  ✓ Removed $INSTALL_PATH/facepass-pam${NC}"
fi

# Step 2: Remove configuration
echo -e "${BLUE}[2/5] Removing configuration...${NC}"
if [ -d "$CONFIG_PATH" ]; then
    rm -rf "$CONFIG_PATH"
    echo -e "${GREEN}  ✓ Removed $CONFIG_PATH${NC}"
else
    echo -e "${YELLOW}  ! Configuration directory not found${NC}"
fi

# Step 3: Remove models
echo -e "${BLUE}[3/5] Removing models...${NC}"
if [ -d "$MODEL_PATH" ]; then
    rm -rf "$MODEL_PATH"
    echo -e "${GREEN}  ✓ Removed $MODEL_PATH${NC}"
else
    echo -e "${YELLOW}  ! Models directory not found${NC}"
fi

# Step 4: Remove log file
echo -e "${BLUE}[4/5] Removing log file...${NC}"
if [ -f "$LOG_FILE" ]; then
    rm -f "$LOG_FILE"
    echo -e "${GREEN}  ✓ Removed $LOG_FILE${NC}"
else
    echo -e "${YELLOW}  ! Log file not found${NC}"
fi

# Step 5: Handle face data
echo -e "${BLUE}[5/5] Face data...${NC}"
if [ -d "$DATA_PATH" ]; then
    echo -e "${YELLOW}  Face data preserved at: $DATA_PATH${NC}"
    echo ""
    read -p "  Remove face data as well? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$DATA_PATH"
        echo -e "${GREEN}  ✓ Removed $DATA_PATH${NC}"
    else
        echo -e "${YELLOW}  ! Face data preserved${NC}"
    fi
else
    echo -e "${YELLOW}  ! Face data directory not found${NC}"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  PAM Configuration Reminder${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${RED}IMPORTANT: If you enabled FacePass in PAM, you must manually${NC}"
echo -e "${RED}remove it from your PAM configuration!${NC}"
echo ""
echo "Edit the following files and remove FacePass lines:"
echo "  - /etc/pam.d/common-auth (Ubuntu/Debian)"
echo "  - /etc/pam.d/system-auth (Fedora/Arch)"
echo "  - /etc/pam.d/sudo (if configured for sudo only)"
echo ""
echo "Look for and remove lines containing: facepass-pam"
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}  Uninstallation Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
