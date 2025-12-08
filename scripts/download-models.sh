#!/bin/bash
# Download dlib face recognition models for FacePass

set -e

MODEL_DIR="${1:-$HOME/.local/share/facepass/models}"

echo "FacePass Model Downloader"
echo "========================="
echo ""
echo "Downloading models to: $MODEL_DIR"
echo ""

# Create directory
mkdir -p "$MODEL_DIR"
cd "$MODEL_DIR"

# Download shape predictor
if [ ! -f "shape_predictor_5_face_landmarks.dat" ]; then
    echo "Downloading shape_predictor_5_face_landmarks.dat..."
    wget -q --show-progress "http://dlib.net/files/shape_predictor_5_face_landmarks.dat.bz2"
    bunzip2 shape_predictor_5_face_landmarks.dat.bz2
    echo "Done!"
else
    echo "shape_predictor_5_face_landmarks.dat already exists, skipping."
fi

# Download face recognition model
if [ ! -f "dlib_face_recognition_resnet_model_v1.dat" ]; then
    echo "Downloading dlib_face_recognition_resnet_model_v1.dat..."
    wget -q --show-progress "http://dlib.net/files/dlib_face_recognition_resnet_model_v1.dat.bz2"
    bunzip2 dlib_face_recognition_resnet_model_v1.dat.bz2
    echo "Done!"
else
    echo "dlib_face_recognition_resnet_model_v1.dat already exists, skipping."
fi

# Optional: CNN face detector (more accurate but slower)
if [ "$2" == "--with-cnn" ]; then
    if [ ! -f "mmod_human_face_detector.dat" ]; then
        echo "Downloading mmod_human_face_detector.dat (CNN detector)..."
        wget -q --show-progress "http://dlib.net/files/mmod_human_face_detector.dat.bz2"
        bunzip2 mmod_human_face_detector.dat.bz2
        echo "Done!"
    else
        echo "mmod_human_face_detector.dat already exists, skipping."
    fi
fi

echo ""
echo "Models downloaded successfully to: $MODEL_DIR"
echo ""
echo "Files:"
ls -lh "$MODEL_DIR"/*.dat 2>/dev/null || echo "No model files found"
