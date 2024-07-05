#!/bin/bash

# Define server URL, referer URL, and origin for input sanitization test
SERVER_URL="http://localhost:8080/submit"
REFERER_URL="http://127.0.0.1:8000/"
ORIGIN="http://127.0.0.1:8000"

echo "Testing input sanitization..."

# Make a POST request with a script injection attempt and capture the response
response=$(curl -s -w "%{http_code}" -X POST $SERVER_URL \
    -H "Referer: $REFERER_URL" \
    -H "Origin: $ORIGIN" \
    --form-string "formid=g7h8i9j0k1l2" \
    --form-string "email=jane.doe@example.com" \
    --form-string "message=<script>alert('XSS');</script>" -o response_output.txt)

# Read the response body
response_body=$(cat response_output.txt)

# Verify if the response status code is 200 (OK) and if the message field was sanitized
if [ "$response" -eq 200 ]; then
    echo "Sanitization Test: Passed"
    echo "Response Body: $response_body"
    
    # Check if the script tag is not present in the response
    if [[ "$response_body" != *"<script>alert('XSS');</script>"* ]]; then
        echo "Input Sanitization: Passed"
    else
        echo "Input Sanitization: Failed"
    fi
else
    echo "Sanitization Test: Failed"
    echo "HTTP Status Code: $response"
    echo "Response Body: $response_body"
fi

# Clean up
rm response_output.txt
