#!/bin/bash
# Integration test for YAML import/include functionality
# Tests the import feature end-to-end with real spec files

# Note: Not using set -e because test binary failures should not exit the script
set -o pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Temp directory for test files
TEST_DIR=$(mktemp -d -t platform-spec-import-test-XXXXXX)
trap "rm -rf $TEST_DIR" EXIT

# Binary path
BINARY="../dist/platform-spec"
if [ ! -f "$BINARY" ]; then
    echo -e "${RED}✗ Binary not found at $BINARY${NC}"
    echo "Run 'make build' first"
    exit 1
fi

echo "========================================"
echo "Platform-Spec Import Integration Tests"
echo "========================================"
echo ""
echo "Test directory: $TEST_DIR"
echo ""

# Helper functions
pass_test() {
    echo -e "${GREEN}✓${NC} $1"
    ((TESTS_PASSED++))
}

fail_test() {
    echo -e "${RED}✗${NC} $1"
    if [ -n "$2" ]; then
        echo -e "  ${RED}Error: $2${NC}"
    fi
    ((TESTS_FAILED++))
}

# Test 1: Basic single import
echo "Test 1: Basic single import"
echo "----------------------------"

cat > "$TEST_DIR/base.yaml" << 'EOF'
version: "1.0"
tests:
  packages:
    - name: "base package"
      packages: [bash]
      state: present
EOF

cat > "$TEST_DIR/main.yaml" << 'EOF'
version: "1.0"
imports:
  - base.yaml
tests:
  files:
    - name: "main file"
      path: /tmp
      type: directory
EOF

# Run test and capture output (|| true allows test failures without exiting script)
OUTPUT=$($BINARY test local "$TEST_DIR/main.yaml" 2>&1 || true)

# Check that both tests ran (we don't care if they passed/failed)
if echo "$OUTPUT" | grep -q "base package" && echo "$OUTPUT" | grep -q "main file"; then
    pass_test "Basic import works - both tests executed"
else
    fail_test "Basic import failed - not all tests found"
fi
echo ""

# Test 2: Multiple imports
echo "Test 2: Multiple imports"
echo "------------------------"

cat > "$TEST_DIR/import1.yaml" << 'EOF'
version: "1.0"
tests:
  packages:
    - name: "import1"
      packages: [curl]
EOF

cat > "$TEST_DIR/import2.yaml" << 'EOF'
version: "1.0"
tests:
  files:
    - name: "import2"
      path: /etc
      type: directory
EOF

cat > "$TEST_DIR/multi.yaml" << 'EOF'
version: "1.0"
imports:
  - import1.yaml
  - import2.yaml
tests:
  users:
    - name: "main user"
      user: root
EOF

OUTPUT=$($BINARY test local "$TEST_DIR/multi.yaml" 2>&1 || true)
if echo "$OUTPUT" | grep -q "import1" && echo "$OUTPUT" | grep -q "import2" && echo "$OUTPUT" | grep -q "main user"; then
    pass_test "Multiple imports work - all 3 tests found"
else
    fail_test "Multiple imports failed" "$OUTPUT"
fi
echo ""

# Test 3: Nested imports (A imports B, B imports C)
echo "Test 3: Nested imports"
echo "----------------------"

cat > "$TEST_DIR/level-c.yaml" << 'EOF'
version: "1.0"
tests:
  packages:
    - name: "level-c"
      packages: [vim]
EOF

cat > "$TEST_DIR/level-b.yaml" << 'EOF'
version: "1.0"
imports:
  - level-c.yaml
tests:
  packages:
    - name: "level-b"
      packages: [git]
EOF

cat > "$TEST_DIR/level-a.yaml" << 'EOF'
version: "1.0"
imports:
  - level-b.yaml
tests:
  packages:
    - name: "level-a"
      packages: [curl]
EOF

OUTPUT=$($BINARY test local "$TEST_DIR/level-a.yaml" 2>&1 || true)
if echo "$OUTPUT" | grep -q "level-c" && echo "$OUTPUT" | grep -q "level-b" && echo "$OUTPUT" | grep -q "level-a"; then
    pass_test "Nested imports work - all 3 levels found"
else
    fail_test "Nested imports failed" "$OUTPUT"
fi
echo ""

# Test 4: Circular import detection (should fail)
echo "Test 4: Circular import detection"
echo "----------------------------------"

cat > "$TEST_DIR/circular-a.yaml" << 'EOF'
version: "1.0"
imports:
  - circular-b.yaml
tests:
  packages:
    - name: "circular-a"
      packages: [curl]
EOF

cat > "$TEST_DIR/circular-b.yaml" << 'EOF'
version: "1.0"
imports:
  - circular-a.yaml
tests:
  packages:
    - name: "circular-b"
      packages: [git]
EOF

OUTPUT=$($BINARY test local "$TEST_DIR/circular-a.yaml" 2>&1 || true)
if echo "$OUTPUT" | grep -q "circular import"; then
    pass_test "Circular import correctly detected and rejected"
else
    fail_test "Circular import should have been detected"
fi
echo ""

# Test 5: Subdirectory imports
echo "Test 5: Subdirectory imports"
echo "----------------------------"

mkdir -p "$TEST_DIR/common"
cat > "$TEST_DIR/common/baseline.yaml" << 'EOF'
version: "1.0"
tests:
  packages:
    - name: "baseline"
      packages: [bash]
EOF

cat > "$TEST_DIR/subdir-main.yaml" << 'EOF'
version: "1.0"
imports:
  - common/baseline.yaml
tests:
  files:
    - name: "subdir-main"
      path: /tmp
EOF

OUTPUT=$($BINARY test local "$TEST_DIR/subdir-main.yaml" 2>&1 || true)
if echo "$OUTPUT" | grep -q "baseline" && echo "$OUTPUT" | grep -q "subdir-main"; then
    pass_test "Subdirectory import works"
else
    fail_test "Subdirectory import failed" "$OUTPUT"
fi
echo ""

# Test 6: Absolute path import
echo "Test 6: Absolute path import"
echo "----------------------------"

cat > "$TEST_DIR/absolute-base.yaml" << 'EOF'
version: "1.0"
tests:
  packages:
    - name: "absolute-base"
      packages: [bash]
EOF

cat > "$TEST_DIR/absolute-main.yaml" << EOF
version: "1.0"
imports:
  - $TEST_DIR/absolute-base.yaml
tests:
  files:
    - name: "absolute-main"
      path: /tmp
EOF

OUTPUT=$($BINARY test local "$TEST_DIR/absolute-main.yaml" 2>&1 || true)
if echo "$OUTPUT" | grep -q "absolute-base" && echo "$OUTPUT" | grep -q "absolute-main"; then
    pass_test "Absolute path import works"
else
    fail_test "Absolute path import failed" "$OUTPUT"
fi
echo ""

# Test 7: Metadata and tag merging
echo "Test 7: Metadata and tag merging"
echo "--------------------------------"

cat > "$TEST_DIR/tagged-base.yaml" << 'EOF'
version: "1.0"
metadata:
  name: "Base Spec"
  tags: ["baseline", "security"]
tests:
  packages:
    - name: "tagged-base"
      packages: [bash]
EOF

cat > "$TEST_DIR/tagged-main.yaml" << 'EOF'
version: "1.0"
imports:
  - tagged-base.yaml
metadata:
  name: "Main Spec"
  tags: ["web", "security"]
tests:
  files:
    - name: "tagged-main"
      path: /tmp
EOF

# For this test, we need to check the actual merged spec
# We'll parse the output to verify tag merging occurred
OUTPUT=$($BINARY test local "$TEST_DIR/tagged-main.yaml" 2>&1 || true)
# The spec name should be "Main Spec" (from main file)
if echo "$OUTPUT" | grep -q "Main Spec"; then
    pass_test "Metadata merging works - main spec name used"
else
    fail_test "Metadata merging failed - wrong spec name"
fi
echo ""

# Test 8: Import ordering (imported tests first)
echo "Test 8: Import test execution order"
echo "------------------------------------"

cat > "$TEST_DIR/order-import.yaml" << 'EOF'
version: "1.0"
tests:
  packages:
    - name: "IMPORT-TEST"
      packages: [bash]
EOF

cat > "$TEST_DIR/order-main.yaml" << 'EOF'
version: "1.0"
imports:
  - order-import.yaml
tests:
  packages:
    - name: "MAIN-TEST"
      packages: [curl]
EOF

OUTPUT=$($BINARY test local "$TEST_DIR/order-main.yaml" 2>&1 || true)
# Check if IMPORT-TEST appears before MAIN-TEST in output
IMPORT_LINE=$(echo "$OUTPUT" | grep -n "IMPORT-TEST" | cut -d: -f1 | head -1)
MAIN_LINE=$(echo "$OUTPUT" | grep -n "MAIN-TEST" | cut -d: -f1 | head -1)

if [ -n "$IMPORT_LINE" ] && [ -n "$MAIN_LINE" ] && [ "$IMPORT_LINE" -lt "$MAIN_LINE" ]; then
    pass_test "Import test ordering correct - imported tests execute first"
else
    fail_test "Import test ordering wrong" "IMPORT at line $IMPORT_LINE, MAIN at line $MAIN_LINE"
fi
echo ""

# Test 9: Nonexistent import file (should fail)
echo "Test 9: Nonexistent import file"
echo "--------------------------------"

cat > "$TEST_DIR/missing-import.yaml" << 'EOF'
version: "1.0"
imports:
  - nonexistent.yaml
tests:
  packages:
    - name: "test"
      packages: [bash]
EOF

OUTPUT=$($BINARY test local "$TEST_DIR/missing-import.yaml" 2>&1 || true)
if echo "$OUTPUT" | grep -q "failed to import\|no such file"; then
    pass_test "Missing import file correctly detected"
else
    fail_test "Should have failed with missing import file"
fi
echo ""

# Test 10: All test types merge correctly
echo "Test 10: All test types merge"
echo "------------------------------"

cat > "$TEST_DIR/all-types-import.yaml" << 'EOF'
version: "1.0"
tests:
  packages:
    - name: "import-package"
      packages: [bash]
  files:
    - name: "import-file"
      path: /tmp
  services:
    - name: "import-service"
      service: sshd
      state: running
EOF

cat > "$TEST_DIR/all-types-main.yaml" << 'EOF'
version: "1.0"
imports:
  - all-types-import.yaml
tests:
  users:
    - name: "main-user"
      user: root
  groups:
    - name: "main-group"
      groups: [wheel]
      state: present
EOF

OUTPUT=$($BINARY test local "$TEST_DIR/all-types-main.yaml" 2>&1 || true)
FOUND_COUNT=0
echo "$OUTPUT" | grep -q "import-package" && ((FOUND_COUNT++))
echo "$OUTPUT" | grep -q "import-file" && ((FOUND_COUNT++))
echo "$OUTPUT" | grep -q "import-service" && ((FOUND_COUNT++))
echo "$OUTPUT" | grep -q "main-user" && ((FOUND_COUNT++))
echo "$OUTPUT" | grep -q "main-group" && ((FOUND_COUNT++))

if [ "$FOUND_COUNT" -eq 5 ]; then
    pass_test "All test types merged correctly - found all 5 tests"
else
    fail_test "Test type merging incomplete - only found $FOUND_COUNT/5 tests"
fi
echo ""

# Summary
echo "========================================"
echo "Test Summary"
echo "========================================"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
    exit 1
else
    echo -e "Failed: 0"
    echo ""
    echo -e "${GREEN}✅ All import integration tests passed!${NC}"
    exit 0
fi
