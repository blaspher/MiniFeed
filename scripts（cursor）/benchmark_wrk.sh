#!/bin/bash
# ========================================
# MiniFeed 性能测试脚本 (wrk)
# 适用于 Linux/macOS/WSL
# ========================================

set -e

BASE_URL="http://localhost:8888"
TOKEN=""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${CYAN}=== MiniFeed Performance Test (wrk) ===${NC}\n"

# 检查 wrk 是否安装
check_wrk() {
    if ! command -v wrk &> /dev/null; then
        echo -e "${RED}❌ 未检测到 wrk，请先安装:${NC}"
        echo ""
        echo "  Ubuntu/Debian: sudo apt-get install wrk"
        echo "  macOS: brew install wrk"
        echo "  源码安装: https://github.com/wg/wrk"
        echo ""
        exit 1
    fi
    echo -e "${GREEN}✓ wrk 已安装${NC}\n"
}

# 获取 Token
get_token() {
    echo -e "${YELLOW}[1/3] 获取认证 Token...${NC}"
    
    # 尝试注册
    RANDOM_USER="test_user_$RANDOM"
    REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/user/register" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$RANDOM_USER\",\"password\":\"test123456\"}")
    
    TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    
    # 如果注册失败，尝试登录
    if [ -z "$TOKEN" ]; then
        echo -e "${YELLOW}  注册失败，尝试登录已有用户...${NC}"
        LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/user/login" \
            -H "Content-Type: application/json" \
            -d '{"username":"testuser","password":"123456"}')
        
        TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    fi
    
    if [ -z "$TOKEN" ]; then
        echo -e "${RED}❌ 获取 Token 失败，请检查服务是否启动${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ Token 获取成功${NC}"
}

# 预热缓存
warmup_cache() {
    echo -e "\n${YELLOW}[2/3] 预热热门接口缓存...${NC}"
    
    for i in {1..5}; do
        curl -s "$BASE_URL/api/feed/hot?limit=10" \
            -H "Authorization: Bearer $TOKEN" > /dev/null
        echo -e "  预热请求 $i/5 完成"
        sleep 0.2
    done
    
    echo -e "${GREEN}✓ 缓存预热完成${NC}"
}

# 创建 wrk Lua 脚本
create_lua_script() {
    cat > /tmp/wrk_auth.lua << EOF
wrk.method = "GET"
wrk.headers["Authorization"] = "Bearer $TOKEN"
wrk.headers["Content-Type"] = "application/json"
EOF
}

# 执行性能测试
run_tests() {
    echo -e "\n${YELLOW}[3/3] 开始性能测试...${NC}"
    echo -e "${CYAN}================================================${NC}\n"
    
    # 测试场景 1: 轻负载
    echo -e "${MAGENTA}【测试 1】轻负载 - 4 线程 / 10 并发 / 10 秒${NC}"
    wrk -t4 -c10 -d10s --latency \
        -s /tmp/wrk_auth.lua \
        "$BASE_URL/api/feed/hot?limit=10"
    
    sleep 2
    
    # 测试场景 2: 中负载
    echo -e "\n${MAGENTA}【测试 2】中负载 - 4 线程 / 50 并发 / 30 秒${NC}"
    wrk -t4 -c50 -d30s --latency \
        -s /tmp/wrk_auth.lua \
        "$BASE_URL/api/feed/hot?limit=10"
    
    sleep 2
    
    # 测试场景 3: 高负载
    echo -e "\n${MAGENTA}【测试 3】高负载 - 8 线程 / 100 并发 / 30 秒${NC}"
    wrk -t8 -c100 -d30s --latency \
        -s /tmp/wrk_auth.lua \
        "$BASE_URL/api/feed/hot?limit=10"
    
    sleep 2
    
    # 测试场景 4: 极限测试
    echo -e "\n${MAGENTA}【测试 4】极限测试 - 8 线程 / 200 并发 / 30 秒${NC}"
    wrk -t8 -c200 -d30s --latency \
        -s /tmp/wrk_auth.lua \
        "$BASE_URL/api/feed/hot?limit=10"
    
    # 清理临时文件
    rm -f /tmp/wrk_auth.lua
}

# 主流程
main() {
    check_wrk
    get_token
    warmup_cache
    create_lua_script
    run_tests
    
    echo -e "\n${CYAN}================================================${NC}"
    echo -e "${GREEN}✓ 性能测试完成！${NC}"
    echo -e "\n${YELLOW}查看详细指标: http://localhost:8888/metrics${NC}"
    echo -e "${CYAN}================================================${NC}\n"
}

main

