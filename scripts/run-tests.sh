#!/bin/bash
# FacePass Test Runner
# Comprehensive test execution script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

# Default options
COVERAGE=false
RACE=false
VERBOSE=false
PACKAGE=""
INTEGRATION=false
BENCHMARK=false
MIN_COVERAGE=80

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --coverage|-c)
            COVERAGE=true
            shift
            ;;
        --race|-r)
            RACE=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --package|-p)
            PACKAGE="$2"
            shift 2
            ;;
        --integration|-i)
            INTEGRATION=true
            shift
            ;;
        --benchmark|-b)
            BENCHMARK=true
            shift
            ;;
        --min-coverage)
            MIN_COVERAGE="$2"
            shift 2
            ;;
        --help|-h)
            echo "FacePass Test Runner"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -c, --coverage      Generate coverage report"
            echo "  -r, --race          Enable race detection"
            echo "  -v, --verbose       Verbose output"
            echo "  -p, --package PKG   Test specific package"
            echo "  -i, --integration   Run integration tests"
            echo "  -b, --benchmark     Run benchmarks"
            echo "  --min-coverage N    Minimum coverage percentage (default: 80)"
            echo "  -h, --help          Show this help"
            echo ""
            echo "Examples:"
            echo "  $0                          # Run all tests"
            echo "  $0 --coverage               # Run with coverage"
            echo "  $0 --package config         # Test config package only"
            echo "  $0 --coverage --race        # Coverage + race detection"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  FacePass Test Runner${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Build test command
TEST_CMD="go test"
TEST_ARGS=""
TEST_PATH="./..."

if [ -n "$PACKAGE" ]; then
    TEST_PATH="./pkg/$PACKAGE/..."
    echo -e "${YELLOW}Testing package: $PACKAGE${NC}"
fi

if [ "$VERBOSE" = true ]; then
    TEST_ARGS="$TEST_ARGS -v"
fi

if [ "$RACE" = true ]; then
    TEST_ARGS="$TEST_ARGS -race"
    echo -e "${YELLOW}Race detection: enabled${NC}"
fi

if [ "$COVERAGE" = true ]; then
    TEST_ARGS="$TEST_ARGS -coverprofile=coverage.out -covermode=atomic"
    echo -e "${YELLOW}Coverage: enabled${NC}"
fi

if [ "$INTEGRATION" = true ]; then
    TEST_ARGS="$TEST_ARGS -tags=integration"
    echo -e "${YELLOW}Integration tests: enabled${NC}"
fi

echo ""
echo -e "${BLUE}Running tests...${NC}"
echo ""

# Run tests
if $TEST_CMD $TEST_ARGS $TEST_PATH; then
    echo ""
    echo -e "${GREEN}✓ All tests passed!${NC}"
else
    echo ""
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi

# Run benchmarks if requested
if [ "$BENCHMARK" = true ]; then
    echo ""
    echo -e "${BLUE}Running benchmarks...${NC}"
    go test -bench=. -benchmem $TEST_PATH
fi

# Process coverage
if [ "$COVERAGE" = true ] && [ -f coverage.out ]; then
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  Coverage Report${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""

    # Show coverage summary
    go tool cover -func=coverage.out | tail -20

    # Calculate total coverage
    TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

    echo ""
    echo -e "${BLUE}----------------------------------------${NC}"
    echo -e "Total Coverage: ${YELLOW}${TOTAL_COVERAGE}%${NC}"
    echo -e "${BLUE}----------------------------------------${NC}"

    # Check minimum coverage
    if (( $(echo "$TOTAL_COVERAGE < $MIN_COVERAGE" | bc -l) )); then
        echo ""
        echo -e "${RED}✗ Coverage ${TOTAL_COVERAGE}% is below minimum ${MIN_COVERAGE}%${NC}"
        exit 1
    else
        echo -e "${GREEN}✓ Coverage meets minimum requirement${NC}"
    fi

    # Generate HTML report
    go tool cover -html=coverage.out -o coverage.html
    echo ""
    echo -e "${GREEN}HTML coverage report: coverage.html${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Test Run Complete${NC}"
echo -e "${GREEN}========================================${NC}"
