# ========================================
# MiniFeed 快速性能测试 (单次测试)
# ========================================

param(
    [int]$Connections = 50,
    [int]$Duration = 10,
    [string]$Url = "http://localhost:8888/api/feed/hot?limit=10"
)

Write-Host "=== MiniFeed 快速性能测试 ===" -ForegroundColor Cyan
Write-Host "并发数: $Connections | 持续时间: ${Duration}s" -ForegroundColor Yellow

# 检查 bombardier
if (-not (Get-Command bombardier -ErrorAction SilentlyContinue)) {
    Write-Host ""
    Write-Host "未安装 bombardier，安装方法:" -ForegroundColor Red
    Write-Host "  scoop install bombardier" -ForegroundColor Yellow
    Write-Host "或查看: scripts/README_BENCHMARK.md" -ForegroundColor Yellow
    Write-Host ""
    exit 1
}

# 获取 Token
Write-Host ""
Write-Host "获取 Token..." -ForegroundColor Yellow
try {
    $loginBody = @{
        username = "testuser"
        password = "123456"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "http://localhost:8888/user/login" -Method Post -ContentType "application/json" -Body $loginBody -ErrorAction Stop
    
    $token = $response.data.token
    Write-Host "Token 获取成功" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "获取 Token 失败，请确保服务已启动且有 testuser/123456 账号" -ForegroundColor Red
    Write-Host "错误: $_" -ForegroundColor Red
    Write-Host ""
    exit 1
}

# 执行测试
Write-Host "开始测试..." -ForegroundColor Green
Write-Host ""
bombardier -c $Connections -d "${Duration}s" -l -H "Authorization: Bearer $token" $Url

Write-Host ""
Write-Host "测试完成！" -ForegroundColor Green
Write-Host "查看 Prometheus 指标: http://localhost:8888/metrics" -ForegroundColor Cyan
Write-Host ""
