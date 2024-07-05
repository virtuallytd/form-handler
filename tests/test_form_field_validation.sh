#!/bin/bash

# Define server URL, referer URL, and origin for form field validation test
SERVER_URL="http://localhost:8080/submit"
REFERER_URL="http://127.0.0.1:8000/"
ORIGIN="http://127.0.0.1:8000"

echo "Testing form field validation..."

# Test for missing required fields
response=$(curl -s -w "%{http_code}" -X POST $SERVER_URL \
    -H "Referer: $REFERER_URL" \
    -H "Origin: $ORIGIN" \
    --form-string "formid=g7h8i9j0k1l2" \
    --form-string "email=" \
    --form-string "message=" -o response_output_missing.txt)

# Read the response body for missing fields test
response_body_missing=$(cat response_output_missing.txt)
response_code_missing=$response

# Verify if the response status code is 400 (Bad Request) for missing fields
if [ "$response_code_missing" -eq 400 ]; then
    echo "Field Validation Test (missing fields): Passed"
else
    echo "Field Validation Test (missing fields): Failed"
    echo "HTTP Status Code: $response_code_missing"
    echo "Response Body: $response_body_missing"
fi

# Test for exceeding max length
response=$(curl -s -w "%{http_code}" -X POST $SERVER_URL \
    -H "Referer: $REFERER_URL" \
    -H "Origin: $ORIGIN" \
    --form-string "formid=g7h8i9j0k1l2" \
    --form-string "email=jane.doe@example.com" \
    --form-string "message=$(printf '%*s' 501)" -o response_output_max_length.txt)

# Read the response body for max length test
response_body_max_length=$(cat response_output_max_length.txt)
response_code_max_length=$response

# Verify if the response status code is 400 (Bad Request) for exceeding max length
if [ "$response_code_max_length" -eq 400 ]; then
    echo "Field Validation Test (max length): Passed"
else
    echo "Field Validation Test (max length): Failed"
    echo "HTTP Status Code: $response_code_max_length"
    echo "Response Body: $response_body_max_length"
fi

# Clean up
rm response_output_missing.txt
rm response_output_max_length.txt
