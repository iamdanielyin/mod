#!/bin/bash

# Token 缓存测试脚本
# 这个脚本演示如何测试 BigCache、BadgerDB、Redis 三种缓存策略

echo "=== Token 缓存测试脚本 ==="
echo

# 基础 URL
BASE_URL="http://localhost:3000/services"

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 函数：打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 函数：检查服务是否启动
check_service() {
    print_info "检查服务是否启动..."
    if curl -s "$BASE_URL/docs" > /dev/null; then
        print_success "服务已启动"
        return 0
    else
        print_error "服务未启动，请先启动应用"
        return 1
    fi
}

# 函数：用户登录并获取 Token
login_and_get_token() {
    local username=$1
    print_info "用户 $username 登录中..."

    response=$(curl -s -X POST "$BASE_URL/basic_login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"test123\"}")

    if echo "$response" | grep -q '"success":true'; then
        token=$(echo "$response" | grep -o '"token":"[^"]*"' | sed 's/"token":"\([^"]*\)"/\1/')
        print_success "登录成功，Token: $token"
        echo "$token"
    else
        print_error "登录失败: $response"
        echo ""
    fi
}

# 函数：验证 Token
verify_token() {
    local token=$1
    print_info "验证 Token: $token"

    response=$(curl -s -X POST "$BASE_URL/token_verify_test" \
        -H "Content-Type: application/json" \
        -H "mod-token: $token" \
        -d "{\"user_id\":\"test_user\"}")

    if echo "$response" | grep -q '"success":true'; then
        print_success "Token 验证成功"
        return 0
    else
        print_error "Token 验证失败: $response"
        return 1
    fi
}

# 函数：查询 Token 信息
query_token() {
    local token=$1
    print_info "查询 Token 信息: $token"

    response=$(curl -s -X POST "$BASE_URL/token_query_test" \
        -H "Content-Type: application/json" \
        -d "{\"token\":\"$token\"}")

    if echo "$response" | grep -q '"valid":true'; then
        print_success "Token 查询成功"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        print_warning "Token 不存在或已过期"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    fi
}

# 函数：删除 Token（登出）
logout_token() {
    local token=$1
    print_info "删除 Token (登出): $token"

    response=$(curl -s -X POST "$BASE_URL/token_logout_test" \
        -H "Content-Type: application/json" \
        -d "{\"token\":\"$token\"}")

    if echo "$response" | grep -q '"success":true'; then
        print_success "Token 删除成功"
    else
        print_error "Token 删除失败: $response"
    fi
}

# 函数：批量测试 Token
batch_test_tokens() {
    local count=$1
    print_info "批量创建 $count 个 Token..."

    response=$(curl -s -X POST "$BASE_URL/token_batch_test" \
        -H "Content-Type: application/json" \
        -d "{\"count\":$count}")

    if echo "$response" | grep -q '"total_created"'; then
        created=$(echo "$response" | grep -o '"total_created":[0-9]*' | sed 's/"total_created":\([0-9]*\)/\1/')
        errors=$(echo "$response" | grep -o '"total_errors":[0-9]*' | sed 's/"total_errors":\([0-9]*\)/\1/')
        print_success "批量测试完成: 成功创建 $created 个, 失败 $errors 个"

        # 返回第一个 Token 用于后续测试
        first_token=$(echo "$response" | grep -o '"tokens":\["[^"]*"' | sed 's/"tokens":\["\([^"]*\)"/\1/')
        echo "$first_token"
    else
        print_error "批量测试失败: $response"
        echo ""
    fi
}

# 主测试流程
main() {
    echo "开始 Token 缓存测试..."
    echo

    # 检查服务状态
    if ! check_service; then
        exit 1
    fi

    echo
    echo "=== 基础功能测试 ==="

    # 1. 登录获取 Token
    token=$(login_and_get_token "testuser")
    if [ -z "$token" ]; then
        exit 1
    fi

    echo

    # 2. 验证 Token
    verify_token "$token"

    echo

    # 3. 查询 Token 信息
    query_token "$token"

    echo

    # 4. 再次验证确保 Token 仍然有效
    verify_token "$token"

    echo

    # 5. 删除 Token
    logout_token "$token"

    echo

    # 6. 尝试验证已删除的 Token
    print_info "验证已删除的 Token（应该失败）..."
    verify_token "$token"

    echo
    echo "=== 批量测试 ==="

    # 7. 批量创建 Token
    batch_token=$(batch_test_tokens 5)

    if [ -n "$batch_token" ]; then
        echo

        # 8. 验证批量创建的 Token
        verify_token "$batch_token"

        echo

        # 9. 查询批量创建的 Token
        query_token "$batch_token"

        echo

        # 10. 删除批量创建的 Token
        logout_token "$batch_token"
    fi

    echo
    print_success "Token 缓存测试完成！"
    echo
    print_info "测试说明："
    echo "1. 如果所有测试都通过，说明当前配置的缓存策略工作正常"
    echo "2. 可以修改 mod.yml 中的 token.validation.cache_strategy 来测试不同缓存"
    echo "3. 支持的策略: bigcache (内存), badger (本地持久化), redis (远程缓存)"
    echo "4. 查看应用日志可以看到详细的缓存操作信息"
}

# 显示帮助信息
show_help() {
    echo "Token 缓存测试脚本"
    echo
    echo "用法: $0 [选项]"
    echo
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  test           运行完整测试（默认）"
    echo "  login USER     只测试用户登录"
    echo "  verify TOKEN   只测试 Token 验证"
    echo "  query TOKEN    只测试 Token 查询"
    echo "  logout TOKEN   只测试 Token 删除"
    echo "  batch COUNT    只测试批量创建"
    echo
    echo "示例:"
    echo "  $0                    # 运行完整测试"
    echo "  $0 login testuser     # 测试用户登录"
    echo "  $0 batch 10           # 批量创建10个Token"
}

# 参数处理
case "${1:-test}" in
    "test"|"")
        main
        ;;
    "login")
        check_service && login_and_get_token "${2:-testuser}"
        ;;
    "verify")
        if [ -z "$2" ]; then
            print_error "请提供 Token"
            exit 1
        fi
        check_service && verify_token "$2"
        ;;
    "query")
        if [ -z "$2" ]; then
            print_error "请提供 Token"
            exit 1
        fi
        check_service && query_token "$2"
        ;;
    "logout")
        if [ -z "$2" ]; then
            print_error "请提供 Token"
            exit 1
        fi
        check_service && logout_token "$2"
        ;;
    "batch")
        count="${2:-5}"
        check_service && batch_test_tokens "$count"
        ;;
    "-h"|"--help"|"help")
        show_help
        ;;
    *)
        print_error "未知选项: $1"
        show_help
        exit 1
        ;;
esac