#!/usr/bin/env bash
# Integration tests for fiss-lint install.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_SCRIPT="${SCRIPT_DIR}/install.sh"
TEST_ROOT="$(mktemp -d -t 'fiss-lint-test-XXXXXX')"

cleanup() {
  rm -rf "$TEST_ROOT"
}
trap cleanup EXIT INT TERM

echo "=== Running install.sh integration tests ==="
echo "Test root directory: $TEST_ROOT"

# Test 1: Default installation of latest version into custom directory
echo ""
echo "--- Test 1: Install latest version into custom directory ---"
TEST1_DIR="${TEST_ROOT}/test1"
INSTALL_DIR="$TEST1_DIR" "$INSTALL_SCRIPT"

if [ ! -x "${TEST1_DIR}/fiss-lint" ]; then
  echo "FAIL: ${TEST1_DIR}/fiss-lint not found or not executable"
  exit 1
fi
OUTPUT1=$("${TEST1_DIR}/fiss-lint" --version)
echo "Output: $OUTPUT1"
if ! echo "$OUTPUT1" | grep -q "fiss-lint version"; then
  echo "FAIL: Output does not contain 'fiss-lint version'"
  exit 1
fi
echo "PASS: Test 1 passed."

# Test 2: Install specific version via positional argument ($1)
echo ""
echo "--- Test 2: Install specific version v1.0.0 via positional argument ---"
TEST2_DIR="${TEST_ROOT}/test2"
INSTALL_DIR="$TEST2_DIR" "$INSTALL_SCRIPT" v1.0.0

if [ ! -x "${TEST2_DIR}/fiss-lint" ]; then
  echo "FAIL: ${TEST2_DIR}/fiss-lint not found or not executable"
  exit 1
fi
OUTPUT2=$("${TEST2_DIR}/fiss-lint" --version)
echo "Output: $OUTPUT2"
if ! echo "$OUTPUT2" | grep -q "fiss-lint version"; then
  echo "FAIL: Output does not contain 'fiss-lint version'"
  exit 1
fi
echo "PASS: Test 2 passed."

# Test 3: Install specific version via VERSION env variable without 'v' prefix
echo ""
echo "--- Test 3: Install version '1.0.0' via VERSION env var (auto-prefix 'v') ---"
TEST3_DIR="${TEST_ROOT}/test3"
VERSION="1.0.0" INSTALL_DIR="$TEST3_DIR" "$INSTALL_SCRIPT"

if [ ! -x "${TEST3_DIR}/fiss-lint" ]; then
  echo "FAIL: ${TEST3_DIR}/fiss-lint not found or not executable"
  exit 1
fi
OUTPUT3=$("${TEST3_DIR}/fiss-lint" --version)
echo "Output: $OUTPUT3"
if ! echo "$OUTPUT3" | grep -q "fiss-lint version"; then
  echo "FAIL: Output does not contain 'fiss-lint version'"
  exit 1
fi
echo "PASS: Test 3 passed."

# Test 4: Simulated pipe execution via bash (curl ... | bash)
echo ""
echo "--- Test 4: Simulated pipe execution (cat install.sh | bash) ---"
TEST4_DIR="${TEST_ROOT}/test4"
INSTALL_DIR="$TEST4_DIR" bash < "$INSTALL_SCRIPT"

if [ ! -x "${TEST4_DIR}/fiss-lint" ]; then
  echo "FAIL: ${TEST4_DIR}/fiss-lint not found or not executable"
  exit 1
fi
OUTPUT4=$("${TEST4_DIR}/fiss-lint" --version)
echo "Output: $OUTPUT4"
if ! echo "$OUTPUT4" | grep -q "fiss-lint version"; then
  echo "FAIL: Output does not contain 'fiss-lint version'"
  exit 1
fi
echo "PASS: Test 4 passed."

# Test 5: Negative scenario - non-existent release version
echo ""
echo "--- Test 5: Error handling on non-existent version ---"
TEST5_DIR="${TEST_ROOT}/test5"
set +e
VERSION="v99.99.99" INSTALL_DIR="$TEST5_DIR" "$INSTALL_SCRIPT" >/dev/null 2>&1
EXIT_CODE=$?
set -e
if [ "$EXIT_CODE" -eq 0 ]; then
  echo "FAIL: Expected non-zero exit code for non-existent version, got 0"
  exit 1
fi
if [ -d "$TEST5_DIR" ] && [ -f "${TEST5_DIR}/fiss-lint" ]; then
  echo "FAIL: Binary should not be installed on error"
  exit 1
fi
echo "PASS: Test 5 passed (exit code $EXIT_CODE on non-existent version)."

# Test 6: Negative scenario - invalid version string format
echo ""
echo "--- Test 6: Error handling on invalid version format ---"
TEST6_DIR="${TEST_ROOT}/test6"
set +e
VERSION="invalid/version!" INSTALL_DIR="$TEST6_DIR" "$INSTALL_SCRIPT" >/dev/null 2>&1
EXIT_CODE=$?
set -e
if [ "$EXIT_CODE" -eq 0 ]; then
  echo "FAIL: Expected non-zero exit code for invalid version format, got 0"
  exit 1
fi
echo "PASS: Test 6 passed (exit code $EXIT_CODE on invalid version format)."

echo ""
echo "=== ALL INSTALLATION TESTS PASSED SUCCESSFULLY ==="
