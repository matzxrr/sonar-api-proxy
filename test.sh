#!/bin/bash
# Script to test the sonar-api-proxy endpoints

API_URL=http://localhost:8080

# Set colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print success/error messages
success() {
    echo -e "${GREEN}SUCCESS: $1${NC}"
}

error() {
    echo -e "${RED}ERROR: $1${NC}"
}

# Test health endpoint
echo "Testing health endpoint..."
response=$(curl -s -o /dev/null -w "%{http_code}" ${API_URL}/health)
if [ "$response" -eq 200 ]; then
    success "Health endpoint returned 200 OK"
else
    error "Health endpoint returned $response"
fi

# Test contact-us endpoint
echo -e "\nTesting contact-us endpoint..."
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{
        "first_name": "John",
        "last_name": "Doe",
        "phone": "123-456-7890",
        "email": "matzxrr@gmail.com",
        "street_address": "123 Main St",
        "city": "Anytown",
        "state": "CA",
        "reason": "General Inquiry",
        "message": "This is a test message for the contact form."
    }' \
    ${API_URL}/api/v1/contact-us)

if [ "$response" -eq 201 ]; then
    success "Contact-us endpoint returned 201 Created"
else
    error "Contact-us endpoint returned $response"
fi

# Test report-an-outage endpoint
echo -e "\nTesting report-an-outage endpoint..."
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{
        "first_name": "Jane",
        "last_name": "Smith",
        "phone": "987-654-3210",
        "email": "matzxrr@gmail.com",
        "street_address": "456 Oak Ave",
        "city": "Somewhere",
        "state": "NY",
        "reason": "Internet Outage",
        "message": "My internet has been down since 2pm today."
    }' \
    ${API_URL}/api/v1/report-an-outage)

if [ "$response" -eq 201 ]; then
    success "Report-an-outage endpoint returned 201 Created"
else
    error "Report-an-outage endpoint returned $response"
fi

# Test sign-up endpoint
echo -e "\nTesting sign-up endpoint..."
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{
        "first_name": "Robert",
        "last_name": "Johnson",
        "phone": "555-123-4567",
        "email": "matzxrr@gmail.com",
        "street_address": "789 Pine St",
        "service": "Fiber Internet",
        "message": "I would like to sign up for your fiber internet service."
    }' \
    ${API_URL}/api/v1/sign-up)

if [ "$response" -eq 201 ]; then
    success "Sign-up endpoint returned 201 Created"
else
    error "Sign-up endpoint returned $response"
fi

# Test voip-support endpoint
echo -e "\nTesting voip-support endpoint..."
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{
        "first_name": "Alice",
        "last_name": "Brown",
        "email": "matzxrr@gmail.com",
        "street_address": "101 Maple Dr",
        "city": "Elsewhere",
        "state": "TX",
        "message": "My VoIP phone is not connecting properly."
    }' \
    ${API_URL}/api/v1/voip-support)

if [ "$response" -eq 201 ]; then
    success "Voip-support endpoint returned 201 Created"
else
    error "Voip-support endpoint returned $response"
fi

echo -e "\nAPI testing complete!"
