.PHONY: all build clean install uninstall test run-cli run-pam install-pam dev

BINARY_CLI=facepass
BINARY_PAM=facepass-pam
INSTALL_PATH=/usr/local/bin
PAM_PATH=/usr/lib/security
CONFIG_PATH=/etc/facepass
DATA_PATH=/var/lib/facepass

all: build

build:
	@echo "Building FacePass..."
	@mkdir -p bin
	go build -o bin/$(BINARY_CLI) ./cmd/facepass
	go build -o bin/$(BINARY_PAM) ./cmd/facepass-pam
	@echo "Build complete!"

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
	go test -v ./...

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
