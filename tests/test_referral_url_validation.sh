#!/bin/bash

# Define server URL and invalid referer URL for referral URL validation test
SERVER_URL="http://localhost:8080/submit"
INVALID_REFERER_URL="http://invalid.url/"

echo "Testing referral URL validation..."

# Make a POST request with an invalid referer URL and check the response status code
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST $SERVER_URL \
    -H "Referer: $INVALID_REFERER_URL" \
    -F "formid=g7h8i9j0k1l2" \
    -F "email=jane.doe@example.com" \
    -F "message=Hello, this is a test message")

# Verify if the response status code is 403 (Forbidden)
if [ "$response" -eq 403 ]; then
    echo "Referral URL Validation Test: Passed"
else
    echo "Referral URL Validation Test: Failed"
fi
