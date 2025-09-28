#!/bin/bash

# JWT Example Test Script
# This script demonstrates the JWT authentication flow

echo "=== JWT Example Test Script ==="
echo "Testing JWT authentication and authorization..."
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
    echo_colored $YELLOW "MOD_PATH=jwt_mod.yml go run jwt_example.go"
    exit 1
fi
echo_colored $GREEN "Server is running!"
echo

# Test 1: Login with admin user
echo_colored $BLUE "=== Test 1: Admin Login ==="
LOGIN_RESPONSE=$(api_request "POST" "/services/login" '{"username":"admin","password":"admin123"}')
ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['data']['token']['access_token'])" 2>/dev/null)

if [ ! -z "$ADMIN_TOKEN" ]; then
    echo_colored $GREEN "✅ Admin login successful!"
    echo_colored $YELLOW "Admin Token: ${ADMIN_TOKEN:0:50}..."
else
    echo_colored $RED "❌ Admin login failed!"
fi
echo

# Test 2: Login with regular user
echo_colored $BLUE "=== Test 2: User Login ==="
USER_LOGIN_RESPONSE=$(api_request "POST" "/services/login" '{"username":"user","password":"user123"}')
USER_TOKEN=$(echo "$USER_LOGIN_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['data']['token']['access_token'])" 2>/dev/null)

if [ ! -z "$USER_TOKEN" ]; then
    echo_colored $GREEN "✅ User login successful!"
    echo_colored $YELLOW "User Token: ${USER_TOKEN:0:50}..."
else
    echo_colored $RED "❌ User login failed!"
fi
echo

# Test 3: Access user info with admin token
echo_colored $BLUE "=== Test 3: Get Admin User Info ==="
api_request "POST" "/services/userinfo" '{}' "Authorization: Bearer $ADMIN_TOKEN"

# Test 4: Access user info with user token
echo_colored $BLUE "=== Test 4: Get User Info ==="
api_request "POST" "/services/userinfo" '{}' "Authorization: Bearer $USER_TOKEN"

# Test 5: Access protected data with admin token
echo_colored $BLUE "=== Test 5: Get Protected Data (Admin) ==="
api_request "POST" "/services/protected-data" '{}' "Authorization: Bearer $ADMIN_TOKEN"

# Test 6: Access protected data with user token
echo_colored $BLUE "=== Test 6: Get Protected Data (User) ==="
api_request "POST" "/services/protected-data" '{}' "Authorization: Bearer $USER_TOKEN"

# Test 7: Try to access admin-only endpoint with admin token
echo_colored $BLUE "=== Test 7: Access Admin Data (Admin Token) ==="
api_request "POST" "/admin/data" '{}' "Authorization: Bearer $ADMIN_TOKEN"

# Test 8: Try to access admin-only endpoint with user token (should fail)
echo_colored $BLUE "=== Test 8: Access Admin Data (User Token - Should Fail) ==="
api_request "POST" "/admin/data" '{}' "Authorization: Bearer $USER_TOKEN"

# Test 9: Try to access protected endpoint without token (should fail)
echo_colored $BLUE "=== Test 9: Access Protected Data (No Token - Should Fail) ==="
api_request "POST" "/services/protected-data" '{}'

# Test 10: Refresh token
echo_colored $BLUE "=== Test 10: Refresh Token ==="
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['data']['token']['refresh_token'])" 2>/dev/null)
if [ ! -z "$REFRESH_TOKEN" ]; then
    echo_colored $YELLOW "Using refresh token: ${REFRESH_TOKEN:0:50}..."
    REFRESH_RESPONSE=$(api_request "POST" "/services/refresh" "{\"refresh_token\":\"$REFRESH_TOKEN\"}")
    NEW_TOKEN=$(echo "$REFRESH_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['data']['access_token'])" 2>/dev/null)
    if [ ! -z "$NEW_TOKEN" ]; then
        echo_colored $GREEN "✅ Token refresh successful!"
        echo_colored $YELLOW "New Token: ${NEW_TOKEN:0:50}..."
    else
        echo_colored $RED "❌ Token refresh failed!"
    fi
else
    echo_colored $RED "❌ No refresh token available!"
fi
echo

# Test 11: Logout
echo_colored $BLUE "=== Test 11: Logout ==="
api_request "POST" "/services/logout" '{}' "Authorization: Bearer $ADMIN_TOKEN"

# Test 12: Try to use token after logout (should fail)
echo_colored $BLUE "=== Test 12: Try to Access After Logout (Should Fail) ==="
api_request "POST" "/services/userinfo" '{}' "Authorization: Bearer $ADMIN_TOKEN"

# Test 13: Invalid login credentials
echo_colored $BLUE "=== Test 13: Invalid Login (Should Fail) ==="
api_request "POST" "/services/login" '{"username":"invalid","password":"invalid"}'

echo_colored $GREEN "=== JWT Testing Complete ==="
echo_colored $YELLOW "Check the responses above to verify JWT functionality."
echo_colored $YELLOW "Visit http://localhost:8080/services/docs for API documentation."