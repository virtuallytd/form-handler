#!/bin/bash

# Define server URL, referer URL, and origin for rate limiting test
SERVER_URL="http://localhost:8080/submit"
REFERER_URL="http://127.0.0.1:8000/"
ORIGIN="http://127.0.0.1:8000"

echo "Testing rate limiting..."

# Send multiple requests to trigger rate limiting
for i in {1..6}; do
  response=$(curl -s -o /dev/null -w "%{http_code}" -X POST $SERVER_URL \
    -H "Referer: $REFERER_URL" \
    -H "Origin: $ORIGIN" \
    -F "formid=g7h8i9j0k1l2" \
    -F "email=jane.doe@example.com" \
    -F "message=Hello, this is a test message for rate limiting")
  echo "Request $i: HTTP $response"
done
