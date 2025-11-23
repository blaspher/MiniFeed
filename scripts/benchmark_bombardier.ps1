# ========================================
# MiniFeed 性能测试脚本 (Bombardier)
# ========================================

Write-Host "=== MiniFeed Performance Test (Bombardier) ===" -ForegroundColor Cyan

# 配置
$BASE_URL = "http://localhost:8888"
$TOKEN = ""  # 请先运行 Get-Token 获取 token

# 获取 Token 函数
function Get-Token {
    Write-Host "`n[1/3] 注册测试用户..." -ForegroundColor Yellow
    $registerBody = @{
        username = "test_user_$(Get-Random -Minimum 1000 -Maximum 9999)"
        password = "test123456"
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$BASE_URL/user/register" `
            -Method Post `
            -ContentType "application/json" `
            -Body $registerBody
        
        if ($response.data.token) {
            Write-Host "✓ 注册成功: $($response.data.username)" -ForegroundColor Green
            return $response.data.token
        }
    } catch {
        Write-Host "注册失败，尝试登录已有用户..." -ForegroundColor Yellow
    }

    # 如果注册失败，使用默认用户登录
    Write-Host "`n[1/3] 登录测试用户..." -ForegroundColor Yellow
    $loginBody = @{
        username = "testuser"
        password = "123456"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$BASE_URL/user/login" `
        -Method Post `
        -ContentType "application/json" `
        -Body $loginBody
    
    Write-Host "✓ 登录成功" -ForegroundColor Green
    return $response.data.token
}

# 预热缓存函数
function Warmup-Cache {
    param($token)
    
    Write-Host "`n[2/3] 预热热门接口缓存..." -ForegroundColor Yellow
    
    for ($i = 1; $i -le 5; $i++) {
        try {
            Invoke-RestMethod -Uri "$BASE_URL/api/feed/hot?limit=10" `
                -Headers @{ Authorization = "Bearer $token" } `
                -ErrorAction SilentlyContinue | Out-Null
            Write-Host "  预热请求 $i/5 完成" -ForegroundColor Gray
        } catch {
            Write-Host "  预热请求 $i/5 失败" -ForegroundColor Red
        }
        Start-Sleep -Milliseconds 200
    }
    
    Write-Host "✓ 缓存预热完成" -ForegroundColor Green
}

# 检查 bombardier 是否安装
function Test-Bombardier {
    try {
        $null = Get-Command bombardier -ErrorAction Stop
        return $true
    } catch {
        return $false
    }
}

# 主测试逻辑
Write-Host "`n检查 bombardier 安装状态..." -ForegroundColor Yellow

if (-not (Test-Bombardier)) {
    Write-Host "❌ 未检测到 bombardier，正在安装..." -ForegroundColor Red
    Write-Host "`n请选择安装方式:" -ForegroundColor Cyan
    Write-Host "  1. 使用 Scoop 安装 (推荐): scoop install bombardier"
    Write-Host "  2. 手动下载: https://github.com/codesenberg/bombardier/releases"
    Write-Host "  3. 使用 Go 安装: go install github.com/codesenberg/bombardier@latest"
    Write-Host "`n安装后请重新运行此脚本`n" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ bombardier 已安装`n" -ForegroundColor Green

# 获取 Token
if (-not $TOKEN) {
    $TOKEN = Get-Token
}

# 预热缓存
Warmup-Cache -token $TOKEN

# 执行性能测试
Write-Host "`n[3/3] 开始性能测试..." -ForegroundColor Yellow
Write-Host "================================================" -ForegroundColor Cyan

# 测试场景 1: 轻负载测试
Write-Host "`n【测试 1】轻负载 - 10 并发 / 10 秒" -ForegroundColor Magenta
bombardier -c 10 -d 10s -l `
    -H "Authorization: Bearer $TOKEN" `
    "$BASE_URL/api/feed/hot?limit=10"

Start-Sleep -Seconds 2

# 测试场景 2: 中负载测试
Write-Host "`n【测试 2】中负载 - 50 并发 / 30 秒" -ForegroundColor Magenta
bombardier -c 50 -d 30s -l `
    -H "Authorization: Bearer $TOKEN" `
    "$BASE_URL/api/feed/hot?limit=10"

Start-Sleep -Seconds 2

# 测试场景 3: 高负载测试
Write-Host "`n【测试 3】高负载 - 100 并发 / 30 秒" -ForegroundColor Magenta
bombardier -c 100 -d 30s -l `
    -H "Authorization: Bearer $TOKEN" `
    "$BASE_URL/api/feed/hot?limit=10"

Start-Sleep -Seconds 2

# 测试场景 4: 极限测试
Write-Host "`n【测试 4】极限测试 - 200 并发 / 30 秒" -ForegroundColor Magenta
bombardier -c 200 -d 30s -l `
    -H "Authorization: Bearer $TOKEN" `
    "$BASE_URL/api/feed/hot?limit=10"

Write-Host "`n================================================" -ForegroundColor Cyan
Write-Host "✓ 性能测试完成！" -ForegroundColor Green
Write-Host "`n查看详细指标: http://localhost:8888/metrics" -ForegroundColor Yellow
Write-Host "================================================`n" -ForegroundColor Cyan

