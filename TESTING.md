# FacePass Testing Guide

This document describes the testing strategy, workflow, and procedures for the FacePass project.

## Table of Contents

- [Testing Strategy](#testing-strategy)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Test Categories](#test-categories)
- [Writing Tests](#writing-tests)
- [CI/CD Integration](#cicd-integration)
- [Manual Testing](#manual-testing)
- [Troubleshooting](#troubleshooting)

---

## Testing Strategy

FacePass uses a multi-layered testing approach:

1. **Unit Tests** - Test individual functions and methods in isolation
2. **Integration Tests** - Test component interactions
3. **Mock Tests** - Test components with hardware dependencies using mocks
4. **Manual Tests** - Test actual face recognition with real hardware

### Coverage Goals

| Package | Target Coverage | Notes |
|---------|----------------|-------|
| `pkg/config` | >90% | Pure logic, easy to test |
| `pkg/logging` | >80% | Wrapper functions |
| `pkg/storage` | >85% | File I/O with encryption |
| `pkg/recognition` | >70% | Hardware dependent, uses mocks |
| `pkg/camera` | >60% | Hardware dependent, uses mocks |
| `pkg/liveness` | >85% | Algorithm testing |
| `pkg/pam` | >75% | Integration focused |
| `pkg/acceleration` | >80% | Detection logic |
| **Overall** | **>80%** | Project target |

---

## Running Tests

### Quick Start

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with verbose output
make test-verbose

# Run specific package tests
go test -v ./pkg/config/...
go test -v ./pkg/storage/...
```

### Test Commands

```bash
# Run all unit tests
go test ./...

# Run with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run with race detection
go test -race ./...

# Run specific test
go test -v -run TestConfigLoad ./pkg/config/

# Run tests matching pattern
go test -v -run "Test.*Encryption" ./pkg/storage/

# Benchmark tests
go test -bench=. ./pkg/recognition/
```

### Using the Test Script

```bash
# Run full test suite
./scripts/run-tests.sh

# Run with options
./scripts/run-tests.sh --coverage    # Generate coverage report
./scripts/run-tests.sh --race        # Enable race detection
./scripts/run-tests.sh --verbose     # Verbose output
./scripts/run-tests.sh --package config  # Test specific package
```

---

## Test Coverage

### Viewing Coverage

```bash
# Generate coverage report
make test-coverage

# View in browser
open coverage.html

# Console coverage summary
go test -cover ./...
```

### Coverage Requirements

- **Pull Requests**: Must maintain or improve overall coverage
- **New Code**: Should have >80% coverage
- **Critical Paths**: Authentication and encryption must have >90% coverage

### Excluding from Coverage

Some code is intentionally excluded from coverage:
- Hardware interaction code (requires real devices)
- Main entry points (`cmd/*/main.go`)
- Generated code

---

## Test Categories

### Unit Tests (`*_test.go`)

Test individual functions in isolation:

```go
func TestConfigLoad(t *testing.T) {
    // Test loading configuration from file
}

func TestEncryptDecrypt(t *testing.T) {
    // Test encryption round-trip
}
```

### Integration Tests (`*_integration_test.go`)

Test component interactions:

```go
// +build integration

func TestStorageWithEncryption(t *testing.T) {
    // Test storage with real encryption
}
```

Run integration tests:
```bash
go test -tags=integration ./...
```

### Mock Tests

Tests using mock interfaces for hardware:

```go
func TestRecognizerWithMock(t *testing.T) {
    mock := &MockRecognizer{}
    // Test with mock
}
```

### Benchmark Tests

Performance benchmarks:

```go
func BenchmarkEuclideanDistance(b *testing.B) {
    for i := 0; i < b.N; i++ {
        EuclideanDistance(d1, d2)
    }
}
```

---

## Writing Tests

### Test File Structure

```
pkg/
├── config/
│   ├── config.go
│   └── config_test.go      # Unit tests
├── storage/
│   ├── storage.go
│   ├── storage_test.go     # Unit tests
│   └── testdata/           # Test fixtures
│       └── test_config.yaml
```

### Test Naming Convention

```go
// Function tests: Test<FunctionName>
func TestConfigLoad(t *testing.T) {}

// Method tests: Test<Type>_<Method>
func TestFileStorage_SaveUser(t *testing.T) {}

// Sub-tests for variations
func TestConfigLoad(t *testing.T) {
    t.Run("valid config", func(t *testing.T) {})
    t.Run("missing file", func(t *testing.T) {})
    t.Run("invalid yaml", func(t *testing.T) {})
}
```

### Table-Driven Tests

```go
func TestEuclideanDistance(t *testing.T) {
    tests := []struct {
        name     string
        d1, d2   Descriptor
        expected float64
    }{
        {"identical", d1, d1, 0.0},
        {"different", d1, d2, 0.5},
        {"empty", Descriptor{}, Descriptor{}, 0.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := EuclideanDistance(tt.d1, tt.d2)
            if result != tt.expected {
                t.Errorf("expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

### Using testify

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
    result, err := DoSomething()

    require.NoError(t, err)           // Fails test immediately
    assert.Equal(t, expected, result) // Reports but continues
    assert.NotNil(t, result)
    assert.Len(t, result, 5)
}
```

### Mocking

```go
// Define interface
type Recognizer interface {
    DetectFaces([]byte) ([]Face, error)
}

// Create mock
type MockRecognizer struct {
    DetectFacesFunc func([]byte) ([]Face, error)
}

func (m *MockRecognizer) DetectFaces(data []byte) ([]Face, error) {
    return m.DetectFacesFunc(data)
}

// Use in tests
func TestWithMock(t *testing.T) {
    mock := &MockRecognizer{
        DetectFacesFunc: func(data []byte) ([]Face, error) {
            return []Face{{Confidence: 0.9}}, nil
        },
    }
    // Test with mock
}
```

---

## CI/CD Integration

### GitHub Actions Workflow

Create `.github/workflows/test.yml`:

```yaml
name: Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.21'

    - name: Install dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y libdlib-dev libblas-dev liblapack-dev

    - name: Run tests
      run: make test-coverage

    - name: Check coverage
      run: |
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$COVERAGE < 80" | bc -l) )); then
          echo "Coverage $COVERAGE% is below 80%"
          exit 1
        fi

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out
```

### Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Run tests before commit

echo "Running tests..."
make test

if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

echo "Tests passed!"
```

---

## Manual Testing

### Face Recognition Testing

1. **Enrollment Test**
   ```bash
   # Enroll with test user
   facepass enroll testuser

   # Verify enrollment
   facepass list
   ```

2. **Recognition Test**
   ```bash
   # Test recognition
   facepass test testuser

   # Expected: "Face recognized" or clear error message
   ```

3. **Liveness Test**
   ```bash
   # Test with photo (should fail)
   # Hold phone with photo to camera
   facepass test testuser
   # Expected: "Liveness check failed"
   ```

### PAM Testing

**WARNING**: Test PAM in a VM or with a backup authentication method!

1. **Safe Testing Setup**
   ```bash
   # Keep a root shell open
   sudo -i

   # In another terminal, test PAM
   sudo -u testuser facepass test testuser
   ```

2. **PAM Module Test**
   ```bash
   # Test PAM module directly
   PAM_USER=testuser /usr/local/bin/facepass-pam
   echo $?  # 0=success, 1=failed, 2=fallback
   ```

### Camera Testing

```bash
# List cameras
facepass cameras

# Test capture
facepass test-camera /dev/video0

# Test IR emitter
facepass test-ir
```

---

## Troubleshooting

### Common Test Failures

**"dlib not found"**
```bash
# Tests requiring dlib will be skipped if not installed
# Install dlib or use mock tests
sudo apt install libdlib-dev
```

**"Camera not found"**
```bash
# Camera tests are skipped without hardware
# Use mock tests for CI
go test -tags=mock ./pkg/camera/
```

**"Permission denied"**
```bash
# Some tests need elevated permissions
sudo go test ./pkg/pam/...
```

### Debug Mode

```bash
# Run tests with debug logging
FACEPASS_DEBUG=1 go test -v ./...

# Run single test with tracing
go test -v -run TestSpecific ./pkg/config/ 2>&1 | tee test.log
```

### Flaky Tests

If tests are flaky (intermittent failures):

1. Run with `-count` to repeat:
   ```bash
   go test -count=10 -v ./pkg/liveness/
   ```

2. Check for race conditions:
   ```bash
   go test -race ./...
   ```

3. Check for timing issues in liveness tests

---

## Test Data

### Test Fixtures

Test data is stored in `testdata/` directories:

```
pkg/config/testdata/
├── valid_config.yaml
├── invalid_config.yaml
└── minimal_config.yaml

pkg/storage/testdata/
├── test_user.json
└── encrypted_user.enc
```

### Generating Test Data

```bash
# Generate test face embeddings
go run ./tools/gen-test-data/

# Generate test images (requires camera)
go run ./tools/capture-test-images/
```

---

## Contributing Tests

When adding new features:

1. Write tests BEFORE implementation (TDD encouraged)
2. Ensure >80% coverage for new code
3. Add both positive and negative test cases
4. Update this document if adding new test patterns

See [CONTRIBUTING.md](CONTRIBUTING.md) for more details.
