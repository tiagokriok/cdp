#!/bin/bash
set -uo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0

# Clean up function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test environment...${NC}"
    rm -rf ~/.cdp-test ~/.claude-profiles-test
    unset HOME
}

# Set up trap to clean up on exit
trap cleanup EXIT INT TERM

# Helper functions
pass() {
    echo -e "${GREEN}✓ $1${NC}"
    ((TESTS_PASSED++))
    ((TESTS_RUN++))
}

fail() {
    echo -e "${RED}✗ $1${NC}"
    ((TESTS_RUN++))
}

test_section() {
    echo -e "\n${YELLOW}=== $1 ===${NC}"
}

# Set up test environment
setup_test_env() {
    test_section "Setting up test environment"

    # Create temporary home directory
    TEST_HOME=$(mktemp -d)
    export HOME="$TEST_HOME"

    # Override config locations
    export CDP_CONFIG_DIR="$TEST_HOME/.cdp-test"
    export CDP_PROFILES_DIR="$TEST_HOME/.claude-profiles-test"

    echo "Test HOME: $TEST_HOME"
}

# Build the binary
build_binary() {
    test_section "Building CDP binary"

    if make build; then
        pass "Binary built successfully"
    else
        fail "Failed to build binary"
        exit 1
    fi
}

# Test init command
test_init() {
    test_section "Testing init command"

    if ./cdp init | grep -q "initialized successfully"; then
        pass "Init command executed"
    else
        fail "Init command failed"
        return 1
    fi

    if [ -d "$HOME/.cdp" ]; then
        pass "Config directory created"
    else
        fail "Config directory not created"
    fi

    if [ -d "$HOME/.claude-profiles" ]; then
        pass "Profiles directory created"
    else
        fail "Profiles directory not created"
    fi

    if [ -f "$HOME/.cdp/config.yaml" ]; then
        pass "Config file created"
    else
        fail "Config file not created"
    fi
}

# Test create command
test_create() {
    test_section "Testing create command"

    if ./cdp create work "Work profile" | grep -q "created successfully"; then
        pass "Create work profile"
    else
        fail "Failed to create work profile"
        return 1
    fi

    if [ -d "$HOME/.claude-profiles/work" ]; then
        pass "Work profile directory created"
    else
        fail "Work profile directory not created"
    fi

    if [ -f "$HOME/.claude-profiles/work/.metadata.json" ]; then
        pass "Profile metadata created"
    else
        fail "Profile metadata not created"
    fi

    # Create second profile
    if ./cdp create personal "Personal projects" | grep -q "created successfully"; then
        pass "Create personal profile"
    else
        fail "Failed to create personal profile"
    fi
}

# Test list command
test_list() {
    test_section "Testing list command"

    output=$(./cdp list)

    if echo "$output" | grep -q "work"; then
        pass "Work profile listed"
    else
        fail "Work profile not in list"
    fi

    if echo "$output" | grep -q "personal"; then
        pass "Personal profile listed"
    else
        fail "Personal profile not in list"
    fi

    if echo "$output" | grep -q "Found 2 profile"; then
        pass "Correct profile count"
    else
        fail "Incorrect profile count"
    fi
}

# Test switch command
test_switch() {
    test_section "Testing switch command"

    if ./cdp work --no-run | grep -q "Switched to profile: work"; then
        pass "Switch to work profile"
    else
        fail "Failed to switch to work profile"
        return 1
    fi

    # Verify current profile
    if ./cdp current | grep -q "Profile: work"; then
        pass "Current profile is work"
    else
        fail "Current profile not set correctly"
    fi
}

# Test current command
test_current() {
    test_section "Testing current command"

    output=$(./cdp current)

    if echo "$output" | grep -q "Profile: work"; then
        pass "Current command shows correct profile"
    else
        fail "Current command failed"
    fi

    if echo "$output" | grep -q "Work profile"; then
        pass "Current command shows description"
    else
        fail "Current command missing description"
    fi
}

# Test delete command
test_delete() {
    test_section "Testing delete command"

    # Try to delete current profile (should fail)
    output=$(echo "y" | ./cdp delete work 2>&1)
    if echo "$output" | grep -qi "cannot delete.*current"; then
        pass "Cannot delete current profile"
    else
        fail "Should not allow deleting current profile"
    fi

    # Switch to different profile first
    ./cdp personal --no-run > /dev/null 2>&1

    # Delete work profile
    if echo "y" | ./cdp delete work 2>&1 | grep -q "deleted successfully"; then
        pass "Delete work profile"
    else
        fail "Failed to delete work profile"
    fi

    if [ ! -d "$HOME/.claude-profiles/work" ]; then
        pass "Work profile directory removed"
    else
        fail "Work profile directory still exists"
    fi
}

# Test help and version
test_help_version() {
    test_section "Testing help and version commands"

    if ./cdp help | grep -q "Claude Profile Switcher"; then
        pass "Help command works"
    else
        fail "Help command failed"
    fi

    if ./cdp version | grep -q "Version:"; then
        pass "Version command works"
    else
        fail "Version command failed"
    fi
}

# Test error handling
test_error_handling() {
    test_section "Testing error handling"

    # Try to create profile with invalid name
    output1=$(./cdp create "invalid name" 2>&1)
    if echo "$output1" | grep -qi "invalid.*profile.*name"; then
        pass "Rejects invalid profile names"
    else
        fail "Should reject invalid profile names"
    fi

    # Try to delete non-existent profile
    output2=$(echo "y" | ./cdp delete nonexistent 2>&1)
    if echo "$output2" | grep -qi "does not exist"; then
        pass "Handles non-existent profile deletion"
    else
        fail "Should handle non-existent profile"
    fi

    # Try to switch to non-existent profile
    output3=$(./cdp nonexistent --no-run 2>&1)
    if echo "$output3" | grep -qi "does not exist"; then
        pass "Handles non-existent profile switch"
    else
        fail "Should handle non-existent profile switch"
    fi
}

# Main test execution
main() {
    echo -e "${YELLOW}CDP End-to-End Tests${NC}"
    echo "================================"

    setup_test_env
    build_binary
    test_init
    test_create
    test_list
    test_switch
    test_current
    test_delete
    test_help_version
    test_error_handling

    # Print summary
    echo -e "\n${YELLOW}================================${NC}"
    echo -e "Tests run: $TESTS_RUN"
    echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Tests failed: ${RED}$((TESTS_RUN - TESTS_PASSED))${NC}"

    if [ $TESTS_PASSED -eq $TESTS_RUN ]; then
        echo -e "\n${GREEN}All tests passed!${NC}"
        exit 0
    else
        echo -e "\n${RED}Some tests failed!${NC}"
        exit 1
    fi
}

main "$@"
