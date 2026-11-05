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
    rm -rf ~/.hat-test ~/.claude-profiles-test
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
    export HAT_CONFIG_DIR="$TEST_HOME/.hat-test"
    export HAT_PROFILES_DIR="$TEST_HOME/.claude-profiles-test"

    echo "Test HOME: $TEST_HOME"
}

# Build the binary
build_binary() {
    test_section "Building hat binary"

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

    if ./hat init | grep -q "initialized successfully"; then
        pass "Init command executed"
    else
        fail "Init command failed"
        return 1
    fi

    if [ -d "$HOME/.hat" ]; then
        pass "Config directory created"
    else
        fail "Config directory not created"
    fi

    if [ -d "$HOME/.claude-profiles" ]; then
        pass "Profiles directory created"
    else
        fail "Profiles directory not created"
    fi

    if [ -f "$HOME/.hat/config.yaml" ]; then
        pass "Config file created"
    else
        fail "Config file not created"
    fi
}

# Test create command
test_create() {
    test_section "Testing create command"

    if ./hat create work "Work profile" | grep -q "created successfully"; then
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
    if ./hat create personal "Personal projects" | grep -q "created successfully"; then
        pass "Create personal profile"
    else
        fail "Failed to create personal profile"
    fi
}

# Test list command
test_list() {
    test_section "Testing list command"

    output=$(./hat list)

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

    if ./hat work --no-run | grep -q "Switched to profile: work"; then
        pass "Switch to work profile"
    else
        fail "Failed to switch to work profile"
        return 1
    fi

    # Verify current profile
    if ./hat current | grep -q "Profile: work"; then
        pass "Current profile is work"
    else
        fail "Current profile not set correctly"
    fi
}

# Test current command
test_current() {
    test_section "Testing current command"

    output=$(./hat current)

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

# Test rule command
test_rule() {
    test_section "Testing rule command"

    RULE_DIR=$(mktemp -d)
    HAT_BIN="$PWD/hat"

    output=$(./hat rule add "$RULE_DIR" work 2>&1)
    if echo "$output" | grep -q "Rule added"; then
        pass "Rule add"
    else
        fail "Rule add failed"
    fi

    output=$(./hat rule add "$RULE_DIR" ghost 2>&1)
    if echo "$output" | grep -q "does not exist"; then
        pass "Rule add rejects unknown profile"
    else
        fail "Rule add accepted unknown profile"
    fi

    output=$(./hat rule list 2>&1)
    if echo "$output" | grep -q "$RULE_DIR -> work"; then
        pass "Rule list shows rule"
    else
        fail "Rule list missing rule"
    fi

    output=$(cd "$RULE_DIR" && "$HAT_BIN" rule list 2>&1)
    if echo "$output" | grep -q "▶"; then
        pass "Rule list marks matching rule"
    else
        fail "Rule list did not mark matching rule"
    fi

    output=$(cd "$RULE_DIR" && "$HAT_BIN" --no-run 2>&1)
    if echo "$output" | grep -q "Using profile work"; then
        pass "Bare hat uses rule"
    else
        fail "Bare hat ignored rule"
    fi

    output=$(./hat rule rm "$RULE_DIR" 2>&1)
    if echo "$output" | grep -q "Rule removed"; then
        pass "Rule rm"
    else
        fail "Rule rm failed"
    fi

    output=$(./hat rule list 2>&1)
    if echo "$output" | grep -q "No rules"; then
        pass "Rule list empty after rm"
    else
        fail "Rule list not empty after rm"
    fi

    rm -rf "$RULE_DIR"
}

test_shared_and_plugins() {
    test_section "Testing shared layer and plugins"

    mkdir -p "$HOME/.hat/shared/skills/demo"
    echo "demo" > "$HOME/.hat/shared/skills/demo/SKILL.md"

    ./hat shared link work skills > /dev/null 2>&1
    if [ -L "$HOME/.claude-profiles/work/skills/demo" ]; then
        pass "shared link creates symlink"
    else
        fail "shared link did not create symlink"
    fi

    rm -rf "$HOME/.hat/shared/skills/demo"
    ./hat shared sync work > /dev/null 2>&1
    if [ ! -e "$HOME/.claude-profiles/work/skills/demo" ] && [ ! -L "$HOME/.claude-profiles/work/skills/demo" ]; then
        pass "shared sync prunes dangling symlink"
    else
        fail "shared sync left dangling symlink"
    fi

    ./hat shared unlink work skills > /dev/null 2>&1
    if ! grep -q "\"shared\"" "$HOME/.claude-profiles/work/.metadata.json"; then
        pass "shared unlink removes opt-in"
    else
        fail "shared unlink did not remove opt-in"
    fi

    if ./hat plugins enable work demo@market > /dev/null 2>&1 && ./hat plugins work | grep -q "demo@market  *yes"; then
        pass "plugins enable"
    else
        fail "plugins enable failed"
    fi

    if ./hat plugins disable work demo@market > /dev/null 2>&1 && ./hat plugins work | grep -q "demo@market  *no"; then
        pass "plugins disable"
    else
        fail "plugins disable failed"
    fi
}

# Test delete command
test_delete() {
    test_section "Testing delete command"

    # Try to delete current profile (should fail)
    output=$(echo "y" | ./hat delete work 2>&1)
    if echo "$output" | grep -qi "cannot delete.*current"; then
        pass "Cannot delete current profile"
    else
        fail "Should not allow deleting current profile"
    fi

    # Switch to different profile first
    ./hat personal --no-run > /dev/null 2>&1

    # Delete work profile
    if echo "y" | ./hat delete work 2>&1 | grep -q "deleted successfully"; then
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

    if ./hat help | grep -q "Claude Code profile switcher"; then
        pass "Help command works"
    else
        fail "Help command failed"
    fi

    if ./hat version | grep -q "Version:"; then
        pass "Version command works"
    else
        fail "Version command failed"
    fi
}

# Test error handling
test_error_handling() {
    test_section "Testing error handling"

    # Try to create profile with invalid name
    output1=$(./hat create "invalid name" 2>&1)
    if echo "$output1" | grep -qi "invalid.*profile.*name"; then
        pass "Rejects invalid profile names"
    else
        fail "Should reject invalid profile names"
    fi

    # Try to delete non-existent profile
    output2=$(echo "y" | ./hat delete nonexistent 2>&1)
    if echo "$output2" | grep -qi "does not exist"; then
        pass "Handles non-existent profile deletion"
    else
        fail "Should handle non-existent profile"
    fi

    # Try to switch to non-existent profile
    output3=$(./hat nonexistent --no-run 2>&1)
    if echo "$output3" | grep -qi "does not exist"; then
        pass "Handles non-existent profile switch"
    else
        fail "Should handle non-existent profile switch"
    fi
}

# Test flag passthrough to Claude
test_flag_passthrough() {
    test_section "Testing flag passthrough to Claude"

    # Create a mock claude script that echoes its arguments
    MOCK_DIR=$(mktemp -d)
    MOCK_CLAUDE="$MOCK_DIR/claude"
    cat > "$MOCK_CLAUDE" << 'SCRIPT'
#!/bin/bash
echo "CLAUDE_ARGS: $@"
echo "CLAUDE_CONFIG_DIR: $CLAUDE_CONFIG_DIR"
SCRIPT
    chmod +x "$MOCK_CLAUDE"

    # Prepend mock dir to PATH so our mock claude is found first
    OLD_PATH="$PATH"
    export PATH="$MOCK_DIR:$PATH"

    # Test 1: Single flag passthrough
    output1=$(./hat personal --continue 2>&1)
    if echo "$output1" | grep -q "CLAUDE_ARGS:.*--continue"; then
        pass "Single Claude flag passed through"
    else
        fail "Single Claude flag not passed through"
        echo "Output was: $output1"
    fi

    # Test 2: Multiple flags passthrough
    output2=$(./hat personal --continue --verbose 2>&1)
    if echo "$output2" | grep -q "CLAUDE_ARGS:.*--continue" && echo "$output2" | grep -q "CLAUDE_ARGS:.*--verbose"; then
        pass "Multiple Claude flags passed through"
    else
        fail "Multiple Claude flags not passed through"
        echo "Output was: $output2"
    fi

    # Test 3: CLAUDE_CONFIG_DIR is set correctly
    if echo "$output2" | grep -q "CLAUDE_CONFIG_DIR:.*personal"; then
        pass "CLAUDE_CONFIG_DIR set correctly"
    else
        fail "CLAUDE_CONFIG_DIR not set correctly"
        echo "Output was: $output2"
    fi

    # Test 4: Flags with values
    output3=$(./hat personal --model gpt-4 --temperature 0.5 2>&1)
    if echo "$output3" | grep -q "CLAUDE_ARGS:.*--model.*gpt-4" && echo "$output3" | grep -q "CLAUDE_ARGS:.*--temperature.*0.5"; then
        pass "Flags with values passed through"
    else
        fail "Flags with values not passed through"
        echo "Output was: $output3"
    fi

    # Restore PATH and cleanup
    export PATH="$OLD_PATH"
    rm -rf "$MOCK_DIR"
}

# Test import command
test_import() {
    test_section "Testing import command"

    # Create a temporary source directory with Claude files
    IMPORT_SOURCE=$(mktemp -d)
    echo '{"token":"test123"}' > "$IMPORT_SOURCE/.claude.json"
    echo '{"theme":"dark"}' > "$IMPORT_SOURCE/settings.json"
    echo "custom data" > "$IMPORT_SOURCE/custom.txt"

    mkdir -p "$IMPORT_SOURCE/skills/demo" "$IMPORT_SOURCE/projects/p"
    echo "skill" > "$IMPORT_SOURCE/skills/demo/SKILL.md"
    echo "history" > "$IMPORT_SOURCE/projects/p/session.jsonl"

    # Test basic import with confirmation
    if echo -e "y\ny" | ./hat create imported --import-from "$IMPORT_SOURCE" --description "Imported profile" 2>/dev/null; then
        if [ -d "$HOME/.claude-profiles/imported" ]; then
            pass "Import command created profile directory"
        else
            fail "Import did not create profile directory"
        fi

        # Verify files were copied
        if [ -f "$HOME/.claude-profiles/imported/.claude.json" ]; then
            pass "Import copied .claude.json"
        else
            fail "Import did not copy .claude.json"
        fi

        if [ -f "$HOME/.claude-profiles/imported/settings.json" ]; then
            pass "Import copied settings.json"
        else
            fail "Import did not copy settings.json"
        fi

        if [ -f "$HOME/.claude-profiles/imported/custom.txt" ]; then
            pass "Import copied custom files"
        else
            fail "Import did not copy custom files"
        fi

        # Verify metadata was created
        if [ -f "$HOME/.claude-profiles/imported/.metadata.json" ]; then
            pass "Import created metadata file"
        else
            fail "Import did not create metadata file"
        fi

        if [ -f "$HOME/.claude-profiles/imported/skills/demo/SKILL.md" ]; then
            pass "Import copied subdirectories"
        else
            fail "Import should copy subdirectories"
        fi

        if [ ! -e "$HOME/.claude-profiles/imported/projects" ]; then
            pass "Import excluded runtime directories"
        else
            fail "Import should exclude runtime directories"
        fi
    else
        fail "Import command failed"
    fi

    # Test import with missing source file
    output=$(./hat create failed-import --import-from /nonexistent/path 2>&1)
    if echo "$output" | grep -qi "does not exist"; then
        pass "Import handles non-existent source path"
    else
        fail "Import should reject non-existent source"
    fi

    # Cleanup
    rm -rf "$IMPORT_SOURCE"
}

# Main test execution
main() {
    echo -e "${YELLOW}hat End-to-End Tests${NC}"
    echo "================================"

    setup_test_env
    build_binary
    test_init
    test_create
    test_list
    test_switch
    test_current
    test_shared_and_plugins
    test_rule
    test_delete
    test_import
    test_flag_passthrough
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
