#!/bin/bash

# Test script for postfix_hook

# Function to run a test case
run_test() {
    local test_name="$1"
    local sender="$2"
    local recipient="$3"
    local input="$4"
    local expected_exit_code="$5"

    echo "Running test: $test_name"
    echo "Sender: $sender"
    echo "Recipient: $recipient"
    echo "Input:"
    echo "$input"
    echo "Expected exit code: $expected_exit_code"

    # Run the postfix_hook and capture its output and exit code
    output=$(echo "$input" | ./postfix_hook -f "$sender" "$recipient" 2>&1)
    exit_code=$?

    echo "Actual exit code: $exit_code"
    echo "Output:"
    echo "$output"

    if [ $exit_code -eq "$expected_exit_code" ]; then
        echo "Test passed!"
    else
        echo "Test failed. Expected exit code $expected_exit_code, got $exit_code"
    fi
    echo "----------------------------------------"
}

# Test case 1: Valid email
run_test "Valid email" "sender@example.com" "recipient@example.com" "Subject: Test Email
From: sender@example.com
To: recipient@example.com

This is a test email." 0

# Test case 2: Missing arguments
run_test "Missing arguments" "" "" "" 64

# Test case 3: Invalid sender email
run_test "Invalid sender email" "invalid-email" "recipient@example.com" "Subject: Test Email
From: invalid-email
To: recipient@example.com

This is a test email with an invalid sender." 75

# Test case 4: Email with potential spam content
run_test "Potential spam content" "sender@example.com" "recipient@example.com" "Subject: Make Money Fast!!!
From: sender@example.com
To: recipient@example.com

Get rich quick! This is definitely not spam!" 69


echo "All tests completed."
