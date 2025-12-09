.PHONY: all build clean install uninstall test run-cli run-pam install-pam dev
.PHONY: build-rocm build-cuda build-openvino build-accelerated detect-gpu download-models

BINARY_CLI=facepass
BINARY_PAM=facepass-pam
INSTALL_PATH=/usr/local/bin
PAM_PATH=/usr/lib/security
CONFIG_PATH=/etc/facepass
DATA_PATH=/var/lib/facepass
MODEL_PATH=/usr/share/facepass/models

# Build tags for acceleration
ROCM_TAGS=-tags rocm
CUDA_TAGS=-tags cuda
OPENVINO_TAGS=-tags openvino

all: build

# Standard CPU build (always works)
build:
	@echo "Building FacePass (CPU)..."
	@mkdir -p bin
	go build -o bin/$(BINARY_CLI) ./cmd/facepass
	go build -o bin/$(BINARY_PAM) ./cmd/facepass-pam
	@echo "Build complete!"

# AMD ROCm accelerated build (tested and supported)
build-rocm:
	@echo "Building FacePass with AMD ROCm acceleration..."
	@echo "Note: Requires ROCm and ONNX Runtime ROCm package installed"
	@mkdir -p bin
	CGO_ENABLED=1 go build $(ROCM_TAGS) -o bin/$(BINARY_CLI) ./cmd/facepass
	CGO_ENABLED=1 go build $(ROCM_TAGS) -o bin/$(BINARY_PAM) ./cmd/facepass-pam
	@echo "ROCm build complete!"

# NVIDIA CUDA accelerated build (needs community testing)
build-cuda:
	@echo "Building FacePass with NVIDIA CUDA acceleration..."
	@echo "WARNING: CUDA support has not been tested by maintainers!"
	@echo "Please report issues at: https://github.com/MrCodeEU/facepass/issues"
	@mkdir -p bin
	CGO_ENABLED=1 go build $(CUDA_TAGS) -o bin/$(BINARY_CLI) ./cmd/facepass
	CGO_ENABLED=1 go build $(CUDA_TAGS) -o bin/$(BINARY_PAM) ./cmd/facepass-pam
	@echo "CUDA build complete!"

# Intel OpenVINO accelerated build (needs community testing)
build-openvino:
	@echo "Building FacePass with Intel OpenVINO acceleration..."
	@echo "WARNING: OpenVINO support has not been tested by maintainers!"
	@echo "Please report issues at: https://github.com/MrCodeEU/facepass/issues"
	@mkdir -p bin
	CGO_ENABLED=1 go build $(OPENVINO_TAGS) -o bin/$(BINARY_CLI) ./cmd/facepass
	CGO_ENABLED=1 go build $(OPENVINO_TAGS) -o bin/$(BINARY_PAM) ./cmd/facepass-pam
	@echo "OpenVINO build complete!"

# Auto-detect GPU and build with appropriate acceleration
build-accelerated:
	@echo "Auto-detecting GPU acceleration..."
	@./scripts/detect-gpu.sh && ./scripts/build-accelerated.sh

# Detect available GPU/NPU
detect-gpu:
	@./scripts/detect-gpu.sh

# Download models (including ONNX models for acceleration)
download-models:
	@./scripts/download-models.sh

download-models-onnx:
	@./scripts/download-models.sh --onnx

clean:
	rm -rf bin/
	go clean

install: build
	@echo "Installing FacePass..."
	sudo install -m 755 bin/$(BINARY_CLI) $(INSTALL_PATH)/
	sudo install -m 755 bin/$(BINARY_PAM) $(INSTALL_PATH)/
	sudo mkdir -p $(CONFIG_PATH)
	sudo mkdir -p $(DATA_PATH)
	sudo cp configs/facepass.yaml $(CONFIG_PATH)/
	sudo chmod 644 $(CONFIG_PATH)/facepass.yaml
	sudo chmod 700 $(DATA_PATH)
	@echo "Installation complete!"
	@echo "Run 'sudo make install-pam' to enable PAM integration"

install-pam:
	@echo "Installing PAM configuration..."
	sudo cp pam-config/facepass /etc/pam.d/
	@echo "PAM configuration installed!"
	@echo "Edit /etc/pam.d/common-auth or /etc/pam.d/system-auth to enable FacePass"

uninstall:
	@echo "Uninstalling FacePass..."
	sudo rm -f $(INSTALL_PATH)/$(BINARY_CLI)
	sudo rm -f $(INSTALL_PATH)/$(BINARY_PAM)
	sudo rm -rf $(CONFIG_PATH)
	@echo "Face data preserved in $(DATA_PATH)"
	@echo "Run 'sudo rm -rf $(DATA_PATH)' to remove all data"

test:
	@echo "Running tests..."
	go test ./...

test-verbose:
	@echo "Running tests (verbose)..."
	go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@./scripts/run-tests.sh --coverage

test-race:
	@echo "Running tests with race detection..."
	go test -race ./...

test-all:
	@echo "Running full test suite..."
	@./scripts/run-tests.sh --coverage --race --verbose

test-integration:
	@echo "Running integration tests..."
	go test -tags=integration -v ./...

test-benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

coverage-html:
	@echo "Generating HTML coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

run-cli:
	go run ./cmd/facepass

run-pam:
	go run ./cmd/facepass-pam

dev:
	@echo "Starting development mode..."
	@echo "Watching for changes..."
	@while true; do \
		make build && ./bin/$(BINARY_CLI) version; \
		inotifywait -qre close_write .; \
	done
