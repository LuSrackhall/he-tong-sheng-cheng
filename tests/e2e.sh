#!/bin/bash
# ============================================================================
# E2E API 测试脚本 — 资产租赁与催缴管理系统
# 使用前确保服务已启动：JWT_SECRET=testsecret ./server
# ============================================================================
set -euo pipefail

BASE_URL="http://localhost:8080"
TOKEN=""
PASS_COUNT=0
FAIL_COUNT=0
TEST_COUNT=0

# ── 颜色 ──────────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

# ── 工具函数 ──────────────────────────────────────────────────────────────────

log_pass() { PASS_COUNT=$((PASS_COUNT + 1)); TEST_COUNT=$((TEST_COUNT + 1)); echo -e "  ${GREEN}PASS${NC} $1"; }
log_fail() { FAIL_COUNT=$((FAIL_COUNT + 1)); TEST_COUNT=$((TEST_COUNT + 1)); echo -e "  ${RED}FAIL${NC} $1"; }
log_info() { echo -e "${YELLOW}[$1]${NC}"; }

# 发送 HTTP 请求并返回响应体
# 用法: api METHOD PATH [BODY] [EXTRA_CURL_ARGS...]
api() {
    local method="$1" path="$2" body="${3:-}" extra_args=("${@:4}")
    local url="${BASE_URL}${path}"
    local curl_args=(-s -w "\n%{http_code}" -X "$method" "$url")

    if [[ -n "$TOKEN" ]]; then
        curl_args+=(-H "Authorization: Bearer $TOKEN")
    fi

    if [[ -n "$body" ]]; then
        curl_args+=(-H "Content-Type: application/json" -d "$body")
    fi

    curl_args+=("${extra_args[@]}")
    curl "${curl_args[@]}"
}

# 提取 HTTP 状态码
get_status() {
    echo "$1" | tail -1
}

# 提取响应体（去掉最后一行状态码）
get_body() {
    echo "$1" | sed '$d'
}

# 从 JSON 中提取字段值（使用 python3 作为 fallback）
json_field() {
    local json="$1" field="$2"
    echo "$json" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d${field})" 2>/dev/null || echo ""
}

# 断言状态码
assert_status() {
    local desc="$1" expected="$2" actual="$3"
    if [[ "$actual" == "$expected" ]]; then
        log_pass "$desc (HTTP $actual)"
    else
        log_fail "$desc — 期望 HTTP $expected, 实际 HTTP $actual"
    fi
}

# 断言 JSON 字段等于期望值
assert_field() {
    local desc="$1" json="$2" field="$3" expected="$4"
    local actual
    actual=$(json_field "$json" "$field")
    if [[ "$actual" == "$expected" ]]; then
        log_pass "$desc ($field == $expected)"
    else
        log_fail "$desc — $field 期望 '$expected', 实际 '$actual'"
    fi
}

# 断言 JSON 字段非空
assert_not_empty() {
    local desc="$1" json="$2" field="$3"
    local actual
    actual=$(json_field "$json" "$field")
    if [[ -n "$actual" && "$actual" != "None" ]]; then
        log_pass "$desc ($field 非空)"
    else
        log_fail "$desc — $field 为空"
    fi
}

# ── 全局状态变量 ──────────────────────────────────────────────────────────────
ASSET_ID=""
TENANT_ID=""
CONTRACT_ID=""
PAYMENT_ID=""
TEMPLATE_ID=""
USER_ID=""
RECEIPT_BOOK_ID=""

# ── 测试：认证流程 ────────────────────────────────────────────────────────────

test_auth() {
    log_info "1. 认证流程"

    # 1.1 缺少认证头
    local old_token="$TOKEN"
    TOKEN=""
    local resp
    resp=$(api GET "/api/auth/me")
    assert_status "1.1 无 token 访问受保护接口" "401" "$(get_status "$resp")"

    # 1.2 错误密码登录
    resp=$(api POST "/api/auth/login" '{"username":"admin","password":"wrongpass"}')
    assert_status "1.2 错误密码登录" "401" "$(get_status "$resp")"

    # 1.3 正确登录
    resp=$(api POST "/api/auth/login" '{"username":"admin","password":"admin123"}')
    local status
    status=$(get_status "$resp")
    local body
    body=$(get_body "$resp")
    assert_status "1.3 正确登录" "200" "$status"
    TOKEN=$(json_field "$body" "['token']")
    assert_not_empty "1.3 返回 token" "$body" "['token']"
    assert_field "1.3 用户角色" "$body" "['user']['role']" "admin"

    # 1.4 /auth/me 验证
    resp=$(api GET "/api/auth/me")
    body=$(get_body "$resp")
    assert_status "1.4 /auth/me" "200" "$(get_status "$resp")"
    assert_field "1.4 用户名" "$body" "['username']" "admin"

    # 1.5 伪造 token
    local fake_token="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOjEsInVzZXJuYW1lIjoiYWRtaW4iLCJyb2xlIjoiYWRtaW4ifQ.invalid"
    TOKEN="$fake_token"
    resp=$(api GET "/api/auth/me")
    assert_status "1.5 伪造 token 访问" "401" "$(get_status "$resp")"

    TOKEN="$old_token"
}

# ── 测试：管理用户 ────────────────────────────────────────────────────────────

test_admin_users() {
    log_info "2. 管理用户"

    # 2.1 创建操作员用户
    local resp
    resp=$(api POST "/api/admin/users" '{"username":"operator1","password":"pass123","role":"operator"}')
    local body
    body=$(get_body "$resp")
    assert_status "2.1 创建操作员用户" "201" "$(get_status "$resp")"
    USER_ID=$(json_field "$body" "['id']")
    assert_field "2.1 角色" "$body" "['role']" "operator"

    # 2.2 重复用户名
    resp=$(api POST "/api/admin/users" '{"username":"operator1","password":"pass456"}')
    assert_status "2.2 重复用户名" "409" "$(get_status "$resp")"

    # 2.3 操作员无法访问管理接口
    local op_resp
    op_resp=$(api POST "/api/auth/login" '{"username":"operator1","password":"pass123"}')
    local op_token
    op_token=$(json_field "$(get_body "$op_resp")" "['token']")
    local saved_token="$TOKEN"
    TOKEN="$op_token"
    resp=$(api GET "/api/admin/users")
    assert_status "2.3 操作员访问管理接口" "403" "$(get_status "$resp")"
    TOKEN="$saved_token"

    # 2.4 列出用户
    resp=$(api GET "/api/admin/users")
    assert_status "2.4 列出用户" "200" "$(get_status "$resp")"

    # 2.5 删除用户
    resp=$(api DELETE "/api/admin/users/$USER_ID")
    assert_status "2.5 删除操作员" "200" "$(get_status "$resp")"
}

# ── 测试：资产 CRUD ──────────────────────────────────────────────────────────

test_asset_crud() {
    log_info "3. 资产 CRUD"

    # 3.1 创建资产
    local resp
    resp=$(api POST "/api/assets" '{"name":"测试商铺A","assetType":"shop","description":"一楼临街商铺"}')
    local body
    body=$(get_body "$resp")
    assert_status "3.1 创建资产" "201" "$(get_status "$resp")"
    ASSET_ID=$(json_field "$body" "['id']")
    assert_field "3.1 资产名" "$body" "['name']" "测试商铺A"
    assert_field "3.1 资产类型" "$body" "['assetType']" "shop"
    assert_field "3.1 初始状态" "$body" "['status']" "idle"

    # 3.2 缺少必填字段
    resp=$(api POST "/api/assets" '{"assetType":"shop"}')
    assert_status "3.2 缺少资产名" "400" "$(get_status "$resp")"

    # 3.3 获取单个资产
    resp=$(api GET "/api/assets/$ASSET_ID")
    body=$(get_body "$resp")
    assert_status "3.3 获取资产" "200" "$(get_status "$resp")"
    assert_field "3.3 资产名" "$body" "['name']" "测试商铺A"

    # 3.4 更新资产
    resp=$(api PATCH "/api/assets/$ASSET_ID" '{"description":"已更新描述","extraFields":"{\"面积\":\"50㎡\"}"}')
    body=$(get_body "$resp")
    assert_status "3.4 更新资产" "200" "$(get_status "$resp")"
    assert_field "3.4 更新后描述" "$body" "['description']" "已更新描述"

    # 3.5 列表与搜索
    resp=$(api POST "/api/assets" '{"name":"测试仓库B","assetType":"warehouse"}')
    resp=$(api GET "/api/assets?search=商铺&type=shop")
    body=$(get_body "$resp")
    assert_status "3.5 搜索资产" "200" "$(get_status "$resp")"
    local total
    total=$(json_field "$body" "['total']")
    if [[ "$total" -ge 1 ]]; then
        log_pass "3.5 搜索结果总数 >= 1 (total=$total)"
    else
        log_fail "3.5 搜索结果为空"
    fi

    # 3.6 按类型筛选
    resp=$(api GET "/api/assets?type=warehouse")
    body=$(get_body "$resp")
    assert_status "3.6 按类型筛选" "200" "$(get_status "$resp")"

    # 3.7 不存在的资产
    resp=$(api GET "/api/assets/99999")
    assert_status "3.7 不存在的资产" "404" "$(get_status "$resp")"
}

# ── 测试：租户 CRUD ──────────────────────────────────────────────────────────

test_tenant_crud() {
    log_info "4. 租户 CRUD"

    # 4.1 创建租户
    local resp
    resp=$(api POST "/api/tenants" '{"name":"张三","phone":"13800138000","idCard":"110101199001011234"}')
    local body
    body=$(get_body "$resp")
    assert_status "4.1 创建租户" "201" "$(get_status "$resp")"
    TENANT_ID=$(json_field "$body" "['id']")
    assert_field "4.1 租户名" "$body" "['name']" "张三"
    assert_field "4.1 手机号" "$body" "['phone']" "13800138000"

    # 4.2 获取租户
    resp=$(api GET "/api/tenants/$TENANT_ID")
    body=$(get_body "$resp")
    assert_status "4.2 获取租户" "200" "$(get_status "$resp")"

    # 4.3 更新租户
    resp=$(api PATCH "/api/tenants/$TENANT_ID" '{"phone":"13900139000"}')
    body=$(get_body "$resp")
    assert_status "4.3 更新手机号" "200" "$(get_status "$resp")"
    assert_field "4.3 新手机号" "$body" "['phone']" "13900139000"

    # 4.4 搜索租户
    resp=$(api GET "/api/tenants?search=张三")
    body=$(get_body "$resp")
    assert_status "4.4 搜索租户" "200" "$(get_status "$resp")"
    total=$(json_field "$body" "['total']")
    if [[ "$total" -ge 1 ]]; then
        log_pass "4.4 搜索结果 >= 1"
    else
        log_fail "4.4 搜索结果为空"
    fi

    # 4.5 创建第二个租户（用于后续测试）
    resp=$(api POST "/api/tenants" '{"name":"李四","phone":"13700137000"}')
    assert_status "4.5 创建第二个租户" "201" "$(get_status "$resp")"
}

# ── 测试：合同生命周期 ────────────────────────────────────────────────────────

test_contract_lifecycle() {
    log_info "5. 合同生命周期"

    # 5.1 创建合同
    local resp
    resp=$(api POST "/api/contracts" "{
        \"assetId\": $ASSET_ID,
        \"tenantId\": $TENANT_ID,
        \"startDate\": \"2026-01-01\",
        \"endDate\": \"2026-12-31\",
        \"monthlyRent\": 1000,
        \"deposit\": 2000,
        \"notes\": \"测试合同\"
    }")
    local body
    body=$(get_body "$resp")
    assert_status "5.1 创建合同" "201" "$(get_status "$resp")"
    CONTRACT_ID=$(json_field "$body" "['id']")
    assert_field "5.1 初始状态" "$body" "['status']" "arrears"
    assert_field "5.1 初始已收" "$body" "['totalReceived']" "0.0"
    assert_not_empty "5.1 应收总额非空" "$body" "['totalReceivable']"

    # 5.2 自动计算应收总额验证
    # 2026-01-01 到 2026-12-31: 11个日历月 + 30天 = 12000
    # totalReceivable = 11 * 1000 + 30 * (1000/30) = 12000
    local tr
    tr=$(json_field "$body" "['totalReceivable']")
    local expected="12000"
    if python3 -c "exit(0 if abs(float('$tr') - float('$expected')) < 0.01 else 1)" 2>/dev/null; then
        log_pass "5.2 自动计算应收总额 ≈ 12000 (实际=$tr)"
    else
        log_fail "5.2 应收总额: 期望 ≈ $expected, 实际 $tr"
    fi

    # 5.3 重复合同检测
    resp=$(api POST "/api/contracts" "{
        \"assetId\": $ASSET_ID,
        \"tenantId\": $TENANT_ID,
        \"startDate\": \"2026-06-01\",
        \"endDate\": \"2027-06-01\",
        \"monthlyRent\": 1000
    }")
    assert_status "5.3 重复合同（时间重叠）" "409" "$(get_status "$resp")"

    # 5.4 日期无效
    resp=$(api POST "/api/contracts" "{
        \"assetId\": $ASSET_ID,
        \"tenantId\": $TENANT_ID,
        \"startDate\": \"2026-12-31\",
        \"endDate\": \"2026-01-01\",
        \"monthlyRent\": 1000
    }")
    assert_status "5.4 结束日期早于开始日期" "400" "$(get_status "$resp")"

    # 5.5 获取合同详情（含嵌套 asset 和 tenant）
    resp=$(api GET "/api/contracts/$CONTRACT_ID")
    body=$(get_body "$resp")
    assert_status "5.5 获取合同详情" "200" "$(get_status "$resp")"
    local asset_name
    asset_name=$(json_field "$body" "['asset']['name']")
    if [[ "$asset_name" == "测试商铺A" ]]; then
        log_pass "5.5 嵌套资产正确 (name=$asset_name)"
    else
        log_fail "5.5 嵌套资产: 期望 '测试商铺A', 实际 '$asset_name'"
    fi

    # 5.6 按状态筛选
    resp=$(api GET "/api/contracts?status=active")
    body=$(get_body "$resp")
    assert_status "5.6 按状态筛选 active" "200" "$(get_status "$resp")"
    total=$(json_field "$body" "['total']")
    if [[ "$total" -ge 1 ]]; then
        log_pass "5.6 筛选结果 >= 1"
    else
        log_fail "5.6 筛选结果为空"
    fi

    # 5.7 按名称搜索
    resp=$(api GET "/api/contracts?search=张三")
    assert_status "5.7 按租户名搜索" "200" "$(get_status "$resp")"

    # 5.8 更新合同备注
    resp=$(api PATCH "/api/contracts/$CONTRACT_ID" '{"notes":"更新后的备注"}')
    body=$(get_body "$resp")
    assert_status "5.8 更新合同" "200" "$(get_status "$resp")"
    assert_field "5.8 备注" "$body" "['notes']" "更新后的备注"
}

# ── 测试：收款流程 ────────────────────────────────────────────────────────────

test_payment_flow() {
    log_info "6. 收款流程"

    # 6.1 第一笔收款
    local resp
    resp=$(api POST "/api/contracts/$CONTRACT_ID/payments" '{"amount":3000,"paidAt":"2026-01-15","notes":"首期租金"}')
    local body
    body=$(get_body "$resp")
    assert_status "6.1 第一笔收款" "201" "$(get_status "$resp")"
    local shortfall
    shortfall=$(json_field "$body" "['shortfall']")
    if python3 -c "exit(0 if float('$shortfall') > 0 else 1)" 2>/dev/null; then
        log_pass "6.1 shortfall > 0 (shortfall=$shortfall)"
    else
        log_fail "6.1 shortfall 应 > 0, 实际=$shortfall"
    fi

    # 6.2 验证合同 totalReceived 已更新
    resp=$(api GET "/api/contracts/$CONTRACT_ID")
    body=$(get_body "$resp")
    assert_field "6.2 totalReceived" "$body" "['totalReceived']" "3000"

    # 6.3 第二笔收款
    resp=$(api POST "/api/contracts/$CONTRACT_ID/payments" '{"amount":2000,"paidAt":"2026-02-15"}')
    body=$(get_body "$resp")
    assert_status "6.3 第二笔收款" "201" "$(get_status "$resp")"

    # 6.4 验证累计总额
    resp=$(api GET "/api/contracts/$CONTRACT_ID")
    body=$(get_body "$resp")
    assert_field "6.4 累计已收 5000" "$body" "['totalReceived']" "5000"

    # 6.5 查看收款记录
    resp=$(api GET "/api/contracts/$CONTRACT_ID/payments")
    body=$(get_body "$resp")
    assert_status "6.5 查看收款记录" "200" "$(get_status "$resp")"
    local count
    count=$(echo "$body" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
    if [[ "$count" == "2" ]]; then
        log_pass "6.5 收款记录数 = 2"
    else
        log_fail "6.5 收款记录数: 期望 2, 实际 $count"
    fi

    # 6.6 无效收款金额
    resp=$(api POST "/api/contracts/$CONTRACT_ID/payments" '{"amount":-100}')
    assert_status "6.6 负数金额拒绝" "400" "$(get_status "$resp")"

    # 6.7 不存在的合同
    resp=$(api POST "/api/contracts/99999/payments" '{"amount":1000}')
    assert_status "6.7 不存在的合同" "404" "$(get_status "$resp")"

    # 6.8 全额付清
    local tr
    tr=$(json_field "$body" "['totalReceivable']")
    resp=$(api POST "/api/contracts/$CONTRACT_ID/payments" "{\"amount\": $tr, \"paidAt\":\"2026-03-01\"}")
    body=$(get_body "$resp")
    assert_status "6.8 全额付清" "201" "$(get_status "$resp")"
    shortfall=$(json_field "$body" "['shortfall']")

    # 6.9 验证合同状态变更为 paidup
    resp=$(api GET "/api/contracts/$CONTRACT_ID")
    body=$(get_body "$resp")
    assert_field "6.9 状态变为 paidup" "$body" "['status']" "paidup"

    local shortfall_int
    shortfall_int=$(python3 -c "print(int(float('$shortfall')))" 2>/dev/null || echo "?")
    if [[ "$shortfall_int" == "0" ]]; then
        log_pass "6.8 shortfall = 0"
    else
        log_fail "6.8 shortfall: 期望 0, 实际 $shortfall"
    fi
}

# ── 测试：催缴分级 ────────────────────────────────────────────────────────────

test_arrears_classification() {
    log_info "7. 催缴分级"

    # 创建一个未付清的合同用于催缴测试
    # 先创建新资产和新租户
    local resp
    resp=$(api POST "/api/assets" '{"name":"催缴测试商铺","assetType":"shop"}')
    local arrears_asset_id
    arrears_asset_id=$(json_field "$(get_body "$resp")" "['id']")

    resp=$(api POST "/api/tenants" '{"name":"催缴测试租户","phone":"13600136000"}')
    local arrears_tenant_id
    arrears_tenant_id=$(json_field "$(get_body "$resp")" "['id']")

    # 7.1 创建一个逾期合同（已过期，部分付款）
    # 合同期: 2025-01-01 到 2025-12-31，已过期
    resp=$(api POST "/api/contracts" "{
        \"assetId\": $arrears_asset_id,
        \"tenantId\": $arrears_tenant_id,
        \"startDate\": \"2025-01-01\",
        \"endDate\": \"2025-06-30\",
        \"monthlyRent\": 1000
    }")
    local expired_contract_id
    expired_contract_id=$(json_field "$(get_body "$resp")" "['id']")

    # 收部分款（只付了2个月）
    resp=$(api POST "/api/contracts/$expired_contract_id/payments" '{"amount":2000,"paidAt":"2025-03-01"}')

    # 7.2 查询催缴清单
    resp=$(api GET "/api/arrears")
    local body
    body=$(get_body "$resp")
    assert_status "7.2 查询催缴清单" "200" "$(get_status "$resp")"

    # 验证催缴清单包含刚创建的合同
    local found
    found=$(echo "$body" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for item in data:
    if item.get('id') == $expired_contract_id:
        print(item.get('arrearsLevel', 0))
        break
else:
    print('not_found')
" 2>/dev/null || echo "error")

    if [[ "$found" == "5" ]]; then
        log_pass "7.3 已过期合同催缴级别 = 5 (已到期欠费追缴)"
    elif [[ "$found" == "not_found" ]]; then
        log_fail "7.3 催缴清单中未找到过期合同"
    else
        log_pass "7.3 过期合同催缴级别 = $found (5=追缴/3=逾期)"
    fi

    # 7.4 创建一个即将到期的合同
    # 使用30天后的日期作为结束日期（到期预警）
    local today_plus_15
    today_plus_15=$(python3 -c "from datetime import datetime,timedelta; print((datetime.now()+timedelta(days=15)).strftime('%Y-%m-%d'))" 2>/dev/null)
    local next_year
    next_year=$(python3 -c "from datetime import datetime,timedelta; print((datetime.now()+timedelta(days=395)).strftime('%Y-%m-%d'))" 2>/dev/null)

    if [[ -n "$today_plus_15" && -n "$next_year" ]]; then
        resp=$(api POST "/api/assets" '{"name":"到期预警商铺","assetType":"shop"}')
        local exp_asset_id
        exp_asset_id=$(json_field "$(get_body "$resp")" "['id']")
        resp=$(api POST "/api/tenants" '{"name":"到期预警租户"}')
        local exp_tenant_id
        exp_tenant_id=$(json_field "$(get_body "$resp")" "['id']")

        resp=$(api POST "/api/contracts" "{
            \"assetId\": $exp_asset_id,
            \"tenantId\": $exp_tenant_id,
            \"startDate\": \"2026-01-01\",
            \"endDate\": \"$today_plus_15\",
            \"monthlyRent\": 2000
        }")
        local exp_contract_id
        exp_contract_id=$(json_field "$(get_body "$resp")" "['id']")

        # 查询催缴
        resp=$(api GET "/api/arrears")
        body=$(get_body "$resp")
        found=$(echo "$body" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for item in data:
    if item.get('id') == $exp_contract_id:
        print(item.get('arrearsLevel', 0))
        break
else:
    print('not_found')
" 2>/dev/null || echo "error")

        if [[ "$found" == "4" ]]; then
            log_pass "7.4 即将到期合同催缴级别 = 4 (到期预警)"
        elif [[ "$found" == "not_found" ]]; then
            log_fail "7.4 催缴清单中未找到即将到期合同"
        else
            log_info "7.4 即将到期合同催缴级别 = $found"
        fi
    else
        log_info "7.4 跳过（日期计算失败）"
    fi
}

# ── 测试：模板管理 ────────────────────────────────────────────────────────────

test_template_management() {
    log_info "8. 模板管理"

    # 8.1 创建模板
    local resp
    resp=$(api POST "/api/templates" '{"name":"标准租赁合同模板"}')
    local body
    body=$(get_body "$resp")
    assert_status "8.1 创建模板" "201" "$(get_status "$resp")"
    TEMPLATE_ID=$(json_field "$body" "['id']")
    assert_field "8.1 模板名" "$body" "['name']" "标准租赁合同模板"
    assert_field "8.1 初始未验证" "$body" "['validated']" "False"

    # 8.2 列出模板
    resp=$(api GET "/api/templates")
    assert_status "8.2 列出模板" "200" "$(get_status "$resp")"

    # 8.3 更新字段映射
    local field_map
    field_map=$(python3 -c "import json; print(json.dumps(json.dumps({'startDate':'开始日期','endDate':'结束日期','monthlyRent':'月租金','tenantName':'租户名称','assetName':'资产名称'})))" 2>/dev/null)
    local active_fields
    active_fields=$(python3 -c "import json; print(json.dumps(json.dumps({'startDate':True,'endDate':True,'monthlyRent':True,'tenantName':True,'assetName':True})))" 2>/dev/null)
    resp=$(api PATCH "/api/templates/$TEMPLATE_ID" "{\"fieldMap\": $field_map, \"activeFields\": $active_fields}")
    body=$(get_body "$resp")
    assert_status "8.3 更新字段映射" "200" "$(get_status "$resp")"

    # 8.4 缺少必填字段的映射
    local bad_field_map
    bad_field_map=$(python3 -c "import json; print(json.dumps(json.dumps({'startDate':'开始日期'})))" 2>/dev/null)
    resp=$(api PATCH "/api/templates/$TEMPLATE_ID" "{\"fieldMap\": $bad_field_map}")
    assert_status "8.4 缺少必填字段" "400" "$(get_status "$resp")"

    # 8.5 删除模板（未被合同引用）
    local resp2
    resp2=$(api POST "/api/templates" '{"name":"临时模板"}')
    local tmp_id
    tmp_id=$(json_field "$(get_body "$resp2")" "['id']")
    resp=$(api DELETE "/api/templates/$tmp_id")
    assert_status "8.5 删除未使用模板" "200" "$(get_status "$resp")"
}

# ── 测试：收据簿 ─────────────────────────────────────────────────────────────

test_receipt_books() {
    log_info "9. 收据簿管理"

    # 9.1 创建收据簿
    local resp
    resp=$(api POST "/api/receipt-books" '{"prefix":"INV-2026","startNum":1,"totalPages":100}')
    local body
    body=$(get_body "$resp")
    assert_status "9.1 创建收据簿" "201" "$(get_status "$resp")"
    RECEIPT_BOOK_ID=$(json_field "$body" "['id']")
    assert_field "9.1 currentNum" "$body" "['currentNum']" "1"
    assert_field "9.1 status" "$body" "['status']" "active"

    # 9.2 列出收据簿
    resp=$(api GET "/api/receipt-books")
    body=$(get_body "$resp")
    assert_status "9.2 列出收据簿" "200" "$(get_status "$resp")"
    total=$(json_field "$body" "['total']")
    if [[ "$total" -ge 1 ]]; then
        log_pass "9.2 收据簿数量 >= 1"
    else
        log_fail "9.2 收据簿列表为空"
    fi
}

# ── 运行所有测试 ──────────────────────────────────────────────────────────────

run_all() {
    echo ""
    echo "============================================"
    echo " E2E API 测试 — 资产租赁与催缴管理系统"
    echo "============================================"
    echo ""

    test_auth
    echo ""
    test_admin_users
    echo ""
    test_asset_crud
    echo ""
    test_tenant_crud
    echo ""
    test_contract_lifecycle
    echo ""
    test_payment_flow
    echo ""
    test_arrears_classification
    echo ""
    test_template_management
    echo ""
    test_receipt_books

    echo ""
    echo "============================================"
    echo " 测试结果汇总"
    echo "============================================"
    echo -e " 总计: $TEST_COUNT"
    echo -e " ${GREEN}通过: $PASS_COUNT${NC}"
    echo -e " ${RED}失败: $FAIL_COUNT${NC}"
    echo "============================================"
    echo ""

    if [[ $FAIL_COUNT -gt 0 ]]; then
        exit 1
    fi
}

# ── 入口 ──────────────────────────────────────────────────────────────────────

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
    echo "用法: $0"
    echo ""
    echo "E2E API 测试脚本，需要先启动服务:"
    echo "  JWT_SECRET=testsecret ./server"
    echo ""
    echo "然后运行:"
    echo "  bash $0"
    exit 0
fi

# 检查服务是否启动
if ! curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/auth/login" -X POST \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin123"}' >/dev/null 2>&1; then
    echo -e "${RED}错误: 无法连接到 $BASE_URL${NC}"
    echo "请先启动服务: JWT_SECRET=testsecret ./server"
    exit 1
fi

run_all
