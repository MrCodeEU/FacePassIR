#!/bin/bash
# FacePass Accelerated Build Script
# Automatically builds with the best available acceleration

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

cd "$PROJECT_DIR"

echo -e "${BLUE}FacePass Accelerated Build${NC}"
echo ""

# Detect GPU
BACKEND=$(./scripts/detect-gpu.sh 2>/dev/null | tail -1)

echo ""
echo -e "${YELLOW}Selected backend: ${BACKEND}${NC}"
echo ""

case "$BACKEND" in
    rocm)
        echo -e "${GREEN}Building with AMD ROCm acceleration...${NC}"
        make build-rocm
        ;;
    cuda)
        echo -e "${YELLOW}Building with NVIDIA CUDA acceleration...${NC}"
        echo -e "${RED}WARNING: CUDA support needs community testing!${NC}"
        read -p "Continue? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            make build-cuda
        else
            echo "Building CPU version instead..."
            make build
        fi
        ;;
    openvino)
        echo -e "${YELLOW}Building with Intel OpenVINO acceleration...${NC}"
        echo -e "${RED}WARNING: OpenVINO support needs community testing!${NC}"
        read -p "Continue? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            make build-openvino
        else
            echo "Building CPU version instead..."
            make build
        fi
        ;;
    *)
        echo -e "${BLUE}Building CPU version...${NC}"
        make build
        ;;
esac

echo ""
echo -e "${GREEN}Build complete!${NC}"
echo ""
echo "Binaries available in ./bin/"
echo "  - facepass (CLI tool)"
echo "  - facepass-pam (PAM module)"
echo ""
echo "Next steps:"
echo "  1. Run 'sudo make install' to install system-wide"
echo "  2. Run 'facepass enroll \$USER' to enroll your face"
echo "  3. Run 'facepass test \$USER' to test recognition"
