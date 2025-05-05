#!/bin/bash
# Simple script to test only the contact-us endpoint once

API_URL=http://localhost:8080

# Set colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "Testing contact-us endpoint..."
response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{
        "first_name": "Test",
        "last_name": "User",
        "phone": "555-1234",
        "email": "matzxrr@gmail.com",
        "street_address": "123 Test St",
        "city": "Testville",
        "state": "TX",
        "reason": "API Test",
        "message": "Single test request"
    }' \
    ${API_URL}/api/v1/contact-us)

status=$?

if [ $status -eq 0 ]; then
    echo -e "${GREEN}Request completed${NC}"
    echo "Response:"
    echo "$response"
else
    echo -e "${RED}Request failed with status $status${NC}"
fi
