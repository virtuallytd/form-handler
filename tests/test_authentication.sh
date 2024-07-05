#!/bin/bash

# Define server URL for authentication test
SERVER_URL="http://localhost:8080/login"

echo "Testing authentication..."

# Test for valid credentials
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST $SERVER_URL \
    -F "username=admin" \
    -F "password=password")

# Verify if the response status code is 302 (Found), indicating a successful login
if [ "$response" -eq 302 ]; then
    echo "Authentication Test (valid credentials): Passed"
else
    echo "Authentication Test (valid credentials): Failed"
fi

# Test for invalid credentials
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST $SERVER_URL \
    -F "username=admin" \
    -F "password=wrongpassword")

# Verify if the response status code is 401 (Unauthorized)
if [ "$response" -eq 401 ]; then
    echo "Authentication Test (invalid credentials): Passed"
else
    echo "Authentication Test (invalid credentials): Failed"
fi
