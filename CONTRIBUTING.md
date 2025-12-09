# Contributing to FacePass

Thank you for your interest in contributing to FacePass! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Community Testing Needed](#community-testing-needed)

---

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. We expect all contributors to:

- Be respectful and considerate in all interactions
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Accept responsibility for mistakes and learn from them

---

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- Go 1.21 or higher
- GCC/G++ compiler (for CGO)
- dlib 19.24+ with development headers
- PAM development headers
- Git

### Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/facepass.git
cd facepass
git remote add upstream https://github.com/MrCodeEU/facepass.git
```

### Build and Test

```bash
# Install dependencies (Ubuntu/Debian)
sudo apt install build-essential libdlib-dev libblas-dev liblapack-dev libpam0g-dev

# Build
make build

# Run tests
make test
```

---

## How to Contribute

### Reporting Bugs

Before reporting a bug:

1. Check existing issues to avoid duplicates
2. Use the latest version of FacePass
3. Collect relevant information:
   - Operating system and version
   - Go version (`go version`)
   - Camera hardware details
   - Error messages and logs

Create a bug report with:

```markdown
## Description
[Clear description of the bug]

## Steps to Reproduce
1. [First step]
2. [Second step]
3. [...]

## Expected Behavior
[What you expected to happen]

## Actual Behavior
[What actually happened]

## Environment
- OS: [e.g., Ubuntu 22.04]
- Go version: [e.g., 1.21.0]
- Camera: [e.g., Logitech C920]
- GPU: [e.g., AMD RX 6800]

## Logs
[Relevant log output]
```

### Suggesting Features

We welcome feature suggestions! Please:

1. Check existing issues for similar suggestions
2. Explain the use case clearly
3. Consider if it fits the project scope

### Contributing Code

1. **Small fixes**: Submit a PR directly
2. **New features**: Open an issue first to discuss
3. **Breaking changes**: Require discussion and approval

---

## Development Setup

### Project Structure

```
facepass/
├── cmd/                    # Application entry points
│   ├── facepass/          # CLI application
│   └── facepass-pam/      # PAM module
├── pkg/                    # Library packages
│   ├── acceleration/      # GPU/NPU acceleration
│   ├── camera/            # Camera capture
│   ├── config/            # Configuration
│   ├── liveness/          # Liveness detection
│   ├── logging/           # Logging utilities
│   ├── pam/               # PAM authentication
│   ├── recognition/       # Face recognition
│   └── storage/           # Data storage
├── scripts/               # Build and install scripts
├── configs/               # Configuration files
└── docs/                  # Documentation
```

### Building

```bash
# Standard CPU build
make build

# GPU accelerated builds
make build-rocm      # AMD ROCm
make build-cuda      # NVIDIA CUDA
make build-openvino  # Intel OpenVINO

# Auto-detect and build
make build-accelerated
```

### Running Locally

```bash
# Run CLI
./bin/facepass --help

# Run with debug logging
FACEPASS_DEBUG=1 ./bin/facepass test $USER
```

---

## Code Style

### Go Code Style

Follow standard Go conventions:

- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use meaningful variable and function names
- Add comments for exported functions

```bash
# Format code
make fmt

# Run linter
make lint
```

### Commit Messages

Use clear, descriptive commit messages:

```
<type>: <short description>

<optional longer description>

<optional footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `test`: Adding tests
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `chore`: Maintenance tasks

Examples:
```
feat: add IR camera auto-detection

fix: resolve memory leak in face embedding storage

docs: update installation instructions for Fedora

test: add unit tests for liveness detection
```

### Code Documentation

```go
// Package recognition provides face detection and recognition functionality.
package recognition

// Recognizer handles face detection and matching operations.
// It supports multiple backends for face embedding extraction.
type Recognizer struct {
    // ...
}

// DetectFaces finds all faces in the provided image data.
// Returns a slice of Face structs containing bounding boxes and confidence scores.
// Returns an error if the image cannot be processed.
func (r *Recognizer) DetectFaces(imageData []byte) ([]Face, error) {
    // ...
}
```

---

## Testing

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# With race detection
make test-race

# Specific package
go test -v ./pkg/config/...
```

### Writing Tests

1. **Test file naming**: `*_test.go`
2. **Test function naming**: `Test<FunctionName>`
3. **Use table-driven tests** for multiple cases
4. **Target >80% coverage** for new code

```go
func TestConfigLoad(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Config
        wantErr bool
    }{
        {"valid config", "testdata/valid.yaml", &Config{...}, false},
        {"missing file", "nonexistent.yaml", nil, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := LoadConfig(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("LoadConfig() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Coverage Requirements

- **New code**: >80% coverage
- **Critical paths** (auth, encryption): >90% coverage
- **PRs must not decrease** overall coverage

---

## Pull Request Process

### Before Submitting

1. **Sync with upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks**:
   ```bash
   make fmt
   make lint
   make test
   ```

3. **Update documentation** if needed

### PR Guidelines

1. **Title**: Clear, descriptive title
2. **Description**: Explain what and why
3. **Size**: Keep PRs focused and reviewable
4. **Tests**: Include tests for new functionality
5. **Documentation**: Update if behavior changes

### PR Template

```markdown
## Description
[What does this PR do?]

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] New tests added
- [ ] Coverage maintained/improved

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No security vulnerabilities introduced
```

### Review Process

1. Maintainer reviews within 1-2 weeks
2. Address feedback in additional commits
3. Squash commits before merge if requested
4. PR merged after approval

---

## Community Testing Needed

We especially need help testing the following areas:

### GPU Acceleration

| Backend | Status | Help Needed |
|---------|--------|-------------|
| AMD ROCm | Tested | Report issues with different AMD GPUs |
| NVIDIA CUDA | Untested | Need testers with NVIDIA GPUs |
| Intel OpenVINO | Untested | Need testers with Intel hardware |

### Hardware Compatibility

- **IR Cameras**: Different laptop IR camera models
- **Webcams**: Various USB webcam brands
- **Linux Distributions**: Testing on different distros

### How to Help

1. Build with appropriate flags (`make build-cuda`, etc.)
2. Test face enrollment and recognition
3. Report issues with detailed hardware information
4. Submit PRs with fixes

---

## Questions?

- Open an issue for questions
- Check existing documentation
- Review closed issues for similar questions

Thank you for contributing to FacePass!
