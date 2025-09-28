#!/bin/bash

# Encryption Example Test Script
# This script demonstrates the encryption and signature verification functionality

echo "=== Encryption Example Test Script ==="
echo "Testing service encryption and signature verification..."
echo

# Base URL
BASE_URL="http://localhost:8080"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo_colored() {
    echo -e "${1}${2}${NC}"
}

# Function to make API requests
api_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local headers=$4

    echo_colored $BLUE "Request: $method $endpoint"
    if [ ! -z "$data" ]; then
        echo_colored $YELLOW "Data: $data"
    fi
    if [ ! -z "$headers" ]; then
        echo_colored $YELLOW "Headers: $headers"
    fi

    if [ ! -z "$headers" ]; then
        response=$(curl -s -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -H "$headers" \
            -d "$data")
    else
        response=$(curl -s -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi

    echo_colored $GREEN "Response:"
    echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
    echo
    echo "$response"
}

# Check if server is running
echo_colored $BLUE "Checking if server is running..."
if ! curl -s "$BASE_URL/services/docs" > /dev/null; then
    echo_colored $RED "Server is not running! Please start the server first:"
    echo_colored $YELLOW "cd /Users/danielyin/Projects/github.com/mod/examples"
    echo_colored $YELLOW "MOD_PATH=encryption_mod.yml go run encryption_example.go"
    exit 1
fi
echo_colored $GREEN "Server is running!"
echo

# Test 1: Get public user info (no encryption required)
echo_colored $BLUE "=== Test 1: Get Public User Info (No Encryption) ==="
echo_colored $YELLOW "This service is in the whitelist, so no encryption is required."
api_request "POST" "/services/get-public-user" '{"id":"1"}'

# Test 2: Try to access encrypted service with plain JSON (should fail)
echo_colored $BLUE "=== Test 2: Try to Access Encrypted Service with Plain JSON (Should Fail) ==="
echo_colored $YELLOW "This should fail because the service requires encryption."
api_request "POST" "/services/get-user" '{"id":"1"}'

# Note: For encrypted services, we would need to implement client-side encryption
# This is more complex and would require a proper client implementation
echo_colored $BLUE "=== Test 3: Encrypted Service Test ==="
echo_colored $YELLOW "Note: Testing encrypted services requires client-side encryption implementation."
echo_colored $YELLOW "The request body needs to be encrypted and signed according to the EncryptedRequest format:"
echo_colored $YELLOW "{"
echo_colored $YELLOW "  \"data\": \"<base64-encoded-encrypted-data>\","
echo_colored $YELLOW "  \"signature\": \"<base64-encoded-signature>\","
echo_colored $YELLOW "  \"algorithm\": \"AES256-GCM\","
echo_colored $YELLOW "  \"mode\": \"symmetric\""
echo_colored $YELLOW "}"
echo

echo_colored $YELLOW "To properly test encryption, you would need to:"
echo_colored $YELLOW "1. Encrypt the JSON payload using AES256-GCM with the configured key"
echo_colored $YELLOW "2. Create an HMAC-SHA256 signature of the encrypted data"
echo_colored $YELLOW "3. Base64 encode both the encrypted data and signature"
echo_colored $YELLOW "4. Send the request in the EncryptedRequest format"
echo

echo_colored $BLUE "=== Test 4: Create User (Encrypted Service) ==="
echo_colored $YELLOW "This would also require proper encryption implementation."
echo_colored $YELLOW "Example encrypted request structure for creating a user:"
echo_colored $YELLOW "Original data: {\"name\":\"John Doe\",\"email\":\"john@example.com\",\"age\":30,\"role\":\"user\",\"salary\":60000,\"password\":\"secret123\"}"
echo_colored $YELLOW "After encryption and signing, it becomes an EncryptedRequest format."
echo

echo_colored $GREEN "=== Encryption Testing Information ==="
echo_colored $YELLOW "Configuration details:"
echo_colored $YELLOW "  - Global encryption: enabled"
echo_colored $YELLOW "  - Encryption mode: symmetric"
echo_colored $YELLOW "  - Encryption algorithm: AES256-GCM"
echo_colored $YELLOW "  - Signature algorithm: HMAC-SHA256"
echo_colored $YELLOW "  - Encrypted services: create-user, get-user"
echo_colored $YELLOW "  - Public services: get-public-user (in whitelist)"
echo_colored $YELLOW ""
echo_colored $YELLOW "To implement a proper client:"
echo_colored $YELLOW "1. Use the same encryption key from config (base64 decode it first)"
echo_colored $YELLOW "2. Use the same signature key from config (base64 decode it first)"
echo_colored $YELLOW "3. Implement AES256-GCM encryption"
echo_colored $YELLOW "4. Implement HMAC-SHA256 signature generation"
echo_colored $YELLOW "5. Follow the EncryptedRequest format"
echo_colored $YELLOW ""
echo_colored $YELLOW "Visit http://localhost:8080/services/docs for API documentation."

echo_colored $GREEN "=== Encryption Testing Complete ==="