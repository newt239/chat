#!/bin/bash

# バックエンドAPIの手動テストスクリプト
# 使用方法: ./scripts/test_endpoints.sh [BASE_URL]

BASE_URL=${1:-"http://localhost:8080"}
echo "Testing API endpoints at: $BASE_URL"

# 色付きの出力用
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# テスト結果のカウンター
PASSED=0
FAILED=0

# テスト関数
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local description=$5
    
    echo -n "Testing $description... "
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -w "%{http_code}" -X $method "$BASE_URL$endpoint")
    fi
    
    # レスポンスコードを取得
    status_code="${response: -3}"
    body="${response%???}"
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC} (Expected: $expected_status, Got: $status_code)"
        echo "Response: $body"
        ((FAILED++))
    fi
}

# テスト用のデータ
REGISTER_DATA='{"email":"test@example.com","password":"password123","displayName":"Test User"}'
LOGIN_DATA='{"email":"test@example.com","password":"password123"}'
REFRESH_DATA='{"refreshToken":"test-refresh-token"}'
LOGOUT_DATA='{"refreshToken":"test-refresh-token"}'

WORKSPACE_DATA='{"name":"Test Workspace","description":"Test Description"}'
WORKSPACE_UPDATE_DATA='{"name":"Updated Workspace","description":"Updated Description"}'

CHANNEL_DATA='{"name":"test-channel","description":"Test Channel"}'

MESSAGE_DATA='{"content":"Hello, World!","type":"text"}'

echo "Starting API endpoint tests..."
echo "=================================="

# ヘルスチェック
test_endpoint "GET" "/healthz" "" "200" "Health check"

# 認証エンドポイント
test_endpoint "POST" "/api/auth/register" "$REGISTER_DATA" "201" "User registration"
test_endpoint "POST" "/api/auth/login" "$LOGIN_DATA" "200" "User login"
test_endpoint "POST" "/api/auth/refresh" "$REFRESH_DATA" "200" "Token refresh"
test_endpoint "POST" "/api/auth/logout" "$LOGOUT_DATA" "200" "User logout"

# ワークスペースエンドポイント（認証が必要）
echo "Note: Workspace endpoints require authentication. Testing without auth will return 401."
test_endpoint "GET" "/api/workspaces" "" "401" "List workspaces (unauthorized)"
test_endpoint "POST" "/api/workspaces" "$WORKSPACE_DATA" "401" "Create workspace (unauthorized)"
test_endpoint "GET" "/api/workspaces/test-id" "" "401" "Get workspace (unauthorized)"
test_endpoint "PATCH" "/api/workspaces/test-id" "$WORKSPACE_UPDATE_DATA" "401" "Update workspace (unauthorized)"
test_endpoint "DELETE" "/api/workspaces/test-id" "" "401" "Delete workspace (unauthorized)"

# チャンネルエンドポイント（認証が必要）
test_endpoint "GET" "/api/workspaces/test-id/channels" "" "401" "List channels (unauthorized)"
test_endpoint "POST" "/api/workspaces/test-id/channels" "$CHANNEL_DATA" "401" "Create channel (unauthorized)"

# メッセージエンドポイント（認証が必要）
test_endpoint "GET" "/api/channels/test-id/messages" "" "401" "List messages (unauthorized)"
test_endpoint "POST" "/api/channels/test-id/messages" "$MESSAGE_DATA" "401" "Create message (unauthorized)"

# 既読状態エンドポイント（認証が必要）
test_endpoint "GET" "/api/channels/test-id/unread_count" "" "401" "Get unread count (unauthorized)"
test_endpoint "POST" "/api/channels/test-id/reads" "" "401" "Update read state (unauthorized)"

# リアクションエンドポイント（認証が必要）
test_endpoint "GET" "/api/messages/test-id/reactions" "" "401" "List reactions (unauthorized)"
test_endpoint "POST" "/api/messages/test-id/reactions" '{"emoji":"👍"}' "401" "Add reaction (unauthorized)"
test_endpoint "DELETE" "/api/messages/test-id/reactions/👍" "" "401" "Remove reaction (unauthorized)"

# ユーザーグループエンドポイント（認証が必要）
test_endpoint "POST" "/api/user-groups" '{"name":"Test Group","description":"Test Description"}' "401" "Create user group (unauthorized)"
test_endpoint "GET" "/api/user-groups" "" "401" "List user groups (unauthorized)"
test_endpoint "GET" "/api/user-groups/test-id" "" "401" "Get user group (unauthorized)"
test_endpoint "PATCH" "/api/user-groups/test-id" '{"name":"Updated Group"}' "401" "Update user group (unauthorized)"
test_endpoint "DELETE" "/api/user-groups/test-id" "" "401" "Delete user group (unauthorized)"

# リンクエンドポイント（認証が必要）
test_endpoint "POST" "/api/links/fetch-ogp" '{"url":"https://example.com"}' "401" "Fetch OGP (unauthorized)"

echo "=================================="
echo "Test Results:"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi
