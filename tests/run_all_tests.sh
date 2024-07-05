#!/bin/bash

# Run all security tests
echo "Running all security tests..."

# Array of test scripts
tests=(
    "test_input_sanitization.sh"
    "test_rate_limiting.sh"
    "test_referral_url_validation.sh"
    "test_cors_validation.sh"
    "test_form_field_validation.sh"
    "test_authentication.sh"
)

# Execute each test script
for test in "${tests[@]}"; do
    echo "Running $test..."
    bash "./tests/$test"
    echo ""
done

echo "All tests completed."
