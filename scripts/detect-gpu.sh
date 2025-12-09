#!/bin/bash
# FacePass GPU/NPU Detection Script
# Detects available acceleration hardware and recommends build options

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  FacePass GPU/NPU Detection${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Track what we find
FOUND_AMD=false
FOUND_NVIDIA=false
FOUND_INTEL=false
FOUND_NPU=false

AMD_DEVICE=""
NVIDIA_DEVICE=""
INTEL_DEVICE=""

# Check for AMD GPU (ROCm)
echo -e "${YELLOW}Checking for AMD GPU...${NC}"
if command -v rocm-smi &> /dev/null; then
    AMD_DEVICE=$(rocm-smi --showproductname 2>/dev/null | grep -i "GPU" | head -1 || echo "")
    if [ -n "$AMD_DEVICE" ]; then
        FOUND_AMD=true
        echo -e "${GREEN}  ✓ Found AMD GPU: ${AMD_DEVICE}${NC}"
    fi
elif command -v rocminfo &> /dev/null; then
    AMD_DEVICE=$(rocminfo 2>/dev/null | grep -i "gfx" | head -1 || echo "")
    if [ -n "$AMD_DEVICE" ]; then
        FOUND_AMD=true
        echo -e "${GREEN}  ✓ Found AMD GPU: ${AMD_DEVICE}${NC}"
    fi
else
    # Check via sysfs
    for card in /sys/class/drm/card*/device/vendor; do
        if [ -f "$card" ]; then
            vendor=$(cat "$card" 2>/dev/null)
            if [ "$vendor" = "0x1002" ]; then
                FOUND_AMD=true
                device_path=$(dirname "$card")
                AMD_DEVICE=$(cat "${device_path}/device" 2>/dev/null || echo "Unknown AMD GPU")
                echo -e "${GREEN}  ✓ Found AMD GPU (device: ${AMD_DEVICE})${NC}"
                break
            fi
        fi
    done
fi

if [ "$FOUND_AMD" = false ]; then
    echo -e "  ✗ No AMD GPU detected"
fi

# Check for NVIDIA GPU (CUDA)
echo -e "${YELLOW}Checking for NVIDIA GPU...${NC}"
if command -v nvidia-smi &> /dev/null; then
    NVIDIA_DEVICE=$(nvidia-smi --query-gpu=name --format=csv,noheader 2>/dev/null | head -1 || echo "")
    if [ -n "$NVIDIA_DEVICE" ]; then
        FOUND_NVIDIA=true
        NVIDIA_VERSION=$(nvidia-smi --query-gpu=driver_version --format=csv,noheader 2>/dev/null | head -1 || echo "unknown")
        echo -e "${GREEN}  ✓ Found NVIDIA GPU: ${NVIDIA_DEVICE} (Driver: ${NVIDIA_VERSION})${NC}"
    fi
else
    # Check via sysfs
    for card in /sys/class/drm/card*/device/vendor; do
        if [ -f "$card" ]; then
            vendor=$(cat "$card" 2>/dev/null)
            if [ "$vendor" = "0x10de" ]; then
                FOUND_NVIDIA=true
                NVIDIA_DEVICE="NVIDIA GPU (nvidia-smi not installed)"
                echo -e "${GREEN}  ✓ Found NVIDIA GPU (install nvidia-utils for details)${NC}"
                break
            fi
        fi
    done
fi

if [ "$FOUND_NVIDIA" = false ]; then
    echo -e "  ✗ No NVIDIA GPU detected"
fi

# Check for Intel GPU/NPU (OpenVINO)
echo -e "${YELLOW}Checking for Intel GPU/NPU...${NC}"

# Check for Intel GPU
for card in /sys/class/drm/card*/device/vendor; do
    if [ -f "$card" ]; then
        vendor=$(cat "$card" 2>/dev/null)
        if [ "$vendor" = "0x8086" ]; then
            FOUND_INTEL=true
            device_path=$(dirname "$card")
            INTEL_DEVICE=$(cat "${device_path}/device" 2>/dev/null || echo "Unknown Intel GPU")
            echo -e "${GREEN}  ✓ Found Intel GPU (device: ${INTEL_DEVICE})${NC}"
            break
        fi
    fi
done

# Check for Intel NPU
if [ -e /dev/accel/accel0 ]; then
    FOUND_NPU=true
    echo -e "${GREEN}  ✓ Found Intel NPU (/dev/accel/accel0)${NC}"
fi

if [ "$FOUND_INTEL" = false ] && [ "$FOUND_NPU" = false ]; then
    echo -e "  ✗ No Intel GPU/NPU detected"
fi

# Check for installed acceleration libraries
echo ""
echo -e "${YELLOW}Checking acceleration libraries...${NC}"

ROCM_INSTALLED=false
CUDA_INSTALLED=false
OPENVINO_INSTALLED=false

# ROCm
if [ -d "/opt/rocm" ] || [ -n "$ROCM_PATH" ]; then
    ROCM_INSTALLED=true
    ROCM_VERSION=$(cat /opt/rocm/.info/version 2>/dev/null || echo "unknown")
    echo -e "${GREEN}  ✓ ROCm installed (version: ${ROCM_VERSION})${NC}"
else
    echo -e "  ✗ ROCm not installed"
fi

# CUDA
if [ -d "/usr/local/cuda" ] || command -v nvcc &> /dev/null; then
    CUDA_INSTALLED=true
    CUDA_VERSION=$(nvcc --version 2>/dev/null | grep "release" | sed 's/.*release \([0-9.]*\).*/\1/' || echo "unknown")
    echo -e "${GREEN}  ✓ CUDA installed (version: ${CUDA_VERSION})${NC}"
else
    echo -e "  ✗ CUDA not installed"
fi

# OpenVINO
if [ -d "/opt/intel/openvino" ] || [ -n "$INTEL_OPENVINO_DIR" ]; then
    OPENVINO_INSTALLED=true
    OPENVINO_PATH="${INTEL_OPENVINO_DIR:-/opt/intel/openvino}"
    OPENVINO_VERSION=$(cat "${OPENVINO_PATH}/version.txt" 2>/dev/null || echo "unknown")
    echo -e "${GREEN}  ✓ OpenVINO installed (version: ${OPENVINO_VERSION})${NC}"
else
    echo -e "  ✗ OpenVINO not installed"
fi

# Recommendations
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Recommendations${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

RECOMMENDED=""

if [ "$FOUND_AMD" = true ] && [ "$ROCM_INSTALLED" = true ]; then
    echo -e "${GREEN}✓ AMD ROCm acceleration available (RECOMMENDED - tested and supported)${NC}"
    echo -e "  Build with: ${YELLOW}make build-rocm${NC}"
    RECOMMENDED="rocm"
elif [ "$FOUND_AMD" = true ]; then
    echo -e "${YELLOW}! AMD GPU found but ROCm not installed${NC}"
    echo -e "  Install ROCm: https://rocm.docs.amd.com/en/latest/deploy/linux/installer/install.html"
fi

if [ "$FOUND_NVIDIA" = true ] && [ "$CUDA_INSTALLED" = true ]; then
    echo -e "${YELLOW}! NVIDIA CUDA acceleration available (needs community testing)${NC}"
    echo -e "  Build with: ${YELLOW}make build-cuda${NC}"
    echo -e "  ${RED}WARNING: CUDA support has not been tested by maintainers${NC}"
    if [ -z "$RECOMMENDED" ]; then
        RECOMMENDED="cuda"
    fi
elif [ "$FOUND_NVIDIA" = true ]; then
    echo -e "${YELLOW}! NVIDIA GPU found but CUDA toolkit not installed${NC}"
    echo -e "  Install CUDA: https://developer.nvidia.com/cuda-downloads"
fi

if ([ "$FOUND_INTEL" = true ] || [ "$FOUND_NPU" = true ]) && [ "$OPENVINO_INSTALLED" = true ]; then
    echo -e "${YELLOW}! Intel OpenVINO acceleration available (needs community testing)${NC}"
    echo -e "  Build with: ${YELLOW}make build-openvino${NC}"
    echo -e "  ${RED}WARNING: OpenVINO support has not been tested by maintainers${NC}"
    if [ -z "$RECOMMENDED" ]; then
        RECOMMENDED="openvino"
    fi
elif [ "$FOUND_INTEL" = true ] || [ "$FOUND_NPU" = true ]; then
    echo -e "${YELLOW}! Intel GPU/NPU found but OpenVINO not installed${NC}"
    echo -e "  Install OpenVINO: https://docs.openvino.ai/latest/openvino_docs_install_guides_overview.html"
fi

if [ -z "$RECOMMENDED" ]; then
    echo -e "${BLUE}No GPU acceleration available. Using CPU build.${NC}"
    echo -e "  Build with: ${YELLOW}make build${NC}"
    RECOMMENDED="cpu"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo ""

# Export for use by other scripts
export FACEPASS_RECOMMENDED_BACKEND="$RECOMMENDED"
export FACEPASS_HAS_AMD="$FOUND_AMD"
export FACEPASS_HAS_NVIDIA="$FOUND_NVIDIA"
export FACEPASS_HAS_INTEL="$FOUND_INTEL"
export FACEPASS_HAS_NPU="$FOUND_NPU"

# Output recommendation for scripting
echo "$RECOMMENDED"
