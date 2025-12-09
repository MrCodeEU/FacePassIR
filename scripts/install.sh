#!/bin/bash
# FacePass Installation Script
# Installs FacePass face recognition PAM module

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
MODEL_PATH="/usr/share/facepass/models"
LOG_PATH="/var/log"
PAM_CONFIG_PATH="/etc/pam.d"

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  FacePass Installation${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check for root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Check for binaries
if [ ! -f "$PROJECT_DIR/bin/facepass" ] || [ ! -f "$PROJECT_DIR/bin/facepass-pam" ]; then
    echo -e "${RED}Error: Binaries not found. Please build first:${NC}"
    echo "  make build"
    echo "  # or for GPU acceleration:"
    echo "  make build-rocm    # AMD"
    echo "  make build-cuda    # NVIDIA"
    echo "  make build-openvino # Intel"
    exit 1
fi

# Detect distribution
if [ -f /etc/os-release ]; then
    . /etc/os-release
    DISTRO=$ID
    DISTRO_VERSION=$VERSION_ID
else
    DISTRO="unknown"
fi

echo -e "${YELLOW}Detected distribution: ${DISTRO} ${DISTRO_VERSION}${NC}"
echo ""

# Step 1: Install binaries
echo -e "${BLUE}[1/6] Installing binaries...${NC}"
install -m 755 "$PROJECT_DIR/bin/facepass" "$INSTALL_PATH/"
install -m 755 "$PROJECT_DIR/bin/facepass-pam" "$INSTALL_PATH/"

# Set capabilities for camera access (alternative to setuid)
if command -v setcap &> /dev/null; then
    setcap cap_dac_override+ep "$INSTALL_PATH/facepass-pam" 2>/dev/null || true
fi

echo -e "${GREEN}  ✓ Binaries installed to ${INSTALL_PATH}${NC}"

# Step 2: Create directories
echo -e "${BLUE}[2/6] Creating directories...${NC}"
mkdir -p "$CONFIG_PATH"
mkdir -p "$DATA_PATH"
mkdir -p "$DATA_PATH/users"
mkdir -p "$MODEL_PATH"

# Set permissions
chmod 755 "$CONFIG_PATH"
chmod 700 "$DATA_PATH"
chmod 700 "$DATA_PATH/users"
chmod 755 "$MODEL_PATH"

echo -e "${GREEN}  ✓ Directories created${NC}"

# Step 3: Install configuration
echo -e "${BLUE}[3/6] Installing configuration...${NC}"
if [ ! -f "$CONFIG_PATH/facepass.yaml" ]; then
    cp "$PROJECT_DIR/configs/facepass.yaml" "$CONFIG_PATH/"
    chmod 644 "$CONFIG_PATH/facepass.yaml"
    echo -e "${GREEN}  ✓ Configuration installed${NC}"
else
    echo -e "${YELLOW}  ! Configuration already exists, skipping${NC}"
fi

# Step 4: Download models if not present
echo -e "${BLUE}[4/6] Checking face recognition models...${NC}"
MODELS_NEEDED=false
if [ ! -f "$MODEL_PATH/shape_predictor_5_face_landmarks.dat" ]; then
    MODELS_NEEDED=true
fi
if [ ! -f "$MODEL_PATH/dlib_face_recognition_resnet_model_v1.dat" ]; then
    MODELS_NEEDED=true
fi

if [ "$MODELS_NEEDED" = true ]; then
    echo -e "${YELLOW}  Downloading face recognition models...${NC}"
    echo -e "${YELLOW}  This may take a few minutes...${NC}"

    if [ -f "$PROJECT_DIR/scripts/download-models.sh" ]; then
        MODEL_DIR="$MODEL_PATH" "$PROJECT_DIR/scripts/download-models.sh"
        echo -e "${GREEN}  ✓ Models downloaded${NC}"
    else
        echo -e "${RED}  ! Model download script not found${NC}"
        echo -e "${YELLOW}  Please run: facepass download-models${NC}"
    fi
else
    echo -e "${GREEN}  ✓ Models already present${NC}"
fi

# Step 5: Create log file
echo -e "${BLUE}[5/6] Setting up logging...${NC}"
touch "$LOG_PATH/facepass.log"
chmod 640 "$LOG_PATH/facepass.log"
echo -e "${GREEN}  ✓ Log file created${NC}"

# Step 6: PAM configuration info
echo -e "${BLUE}[6/6] PAM configuration...${NC}"
echo -e "${YELLOW}  PAM must be configured manually for safety${NC}"
echo ""

# Print PAM setup instructions based on distribution
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  PAM Setup Instructions${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

case "$DISTRO" in
    ubuntu|debian|linuxmint|pop)
        echo -e "${YELLOW}For Ubuntu/Debian-based systems:${NC}"
        echo ""
        echo "Edit /etc/pam.d/common-auth and add this line BEFORE pam_unix.so:"
        echo ""
        echo -e "${GREEN}auth    [success=2 default=ignore]    pam_exec.so expose_authtok /usr/local/bin/facepass-pam${NC}"
        echo ""
        echo "Or for sudo only, edit /etc/pam.d/sudo:"
        echo ""
        echo -e "${GREEN}auth    sufficient    pam_exec.so expose_authtok /usr/local/bin/facepass-pam${NC}"
        ;;
    fedora|rhel|centos|rocky|alma)
        echo -e "${YELLOW}For Fedora/RHEL-based systems:${NC}"
        echo ""
        echo "Edit /etc/pam.d/system-auth and add this line BEFORE pam_unix.so:"
        echo ""
        echo -e "${GREEN}auth    sufficient    pam_exec.so expose_authtok /usr/local/bin/facepass-pam${NC}"
        echo ""
        echo "You may also need to run:"
        echo "  sudo authselect create-profile facepass -b sssd"
        ;;
    arch|manjaro|endeavouros)
        echo -e "${YELLOW}For Arch-based systems:${NC}"
        echo ""
        echo "Edit /etc/pam.d/system-auth and add this line BEFORE pam_unix.so:"
        echo ""
        echo -e "${GREEN}auth    sufficient    pam_exec.so expose_authtok /usr/local/bin/facepass-pam${NC}"
        ;;
    opensuse*|suse*)
        echo -e "${YELLOW}For openSUSE/SUSE systems:${NC}"
        echo ""
        echo "Edit /etc/pam.d/common-auth and add this line BEFORE pam_unix.so:"
        echo ""
        echo -e "${GREEN}auth    sufficient    pam_exec.so expose_authtok /usr/local/bin/facepass-pam${NC}"
        ;;
    *)
        echo -e "${YELLOW}For your system:${NC}"
        echo ""
        echo "Add this line to your PAM configuration (usually /etc/pam.d/system-auth"
        echo "or /etc/pam.d/common-auth) BEFORE pam_unix.so:"
        echo ""
        echo -e "${GREEN}auth    sufficient    pam_exec.so expose_authtok /usr/local/bin/facepass-pam${NC}"
        ;;
esac

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}  Installation Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Next steps:"
echo "  1. Enroll your face: ${YELLOW}facepass enroll \$USER${NC}"
echo "  2. Test recognition: ${YELLOW}facepass test \$USER${NC}"
echo "  3. Configure PAM (see instructions above)"
echo ""
echo -e "${RED}WARNING: Test face recognition thoroughly before enabling PAM!${NC}"
echo -e "${RED}Incorrect PAM configuration can lock you out of your system.${NC}"
echo ""
echo "For help: https://github.com/MrCodeEU/facepass"
