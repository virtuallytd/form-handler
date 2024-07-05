#!/bin/bash

# Define server URL and invalid origin for CORS validation test
SERVER_URL="http://localhost:8080/submit"
INVALID_ORIGIN="http://invalid.origin"

echo "Testing CORS validation..."

# Make a POST request with an invalid origin and check the response status code
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST $SERVER_URL \
    -H "Origin: $INVALID_ORIGIN" \
    -F "formid=g7h8i9j0k1l2" \
    -F "email=jane.doe@example.com" \
    -F "message=Hello, this is a test message")

# Verify if the response status code is 403 (Forbidden)
if [ "$response" -eq 403 ]; then
    echo "CORS Validation Test: Passed"
else
    echo "CORS Validation Test: Failed"
fi
