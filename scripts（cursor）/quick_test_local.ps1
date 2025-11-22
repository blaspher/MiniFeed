# ========================================
# MiniFeed 快速性能测试 (使用本地 bombardier)
# ========================================

param(
    [int]$Connections = 50,
    [int]$Duration = 10,
    [string]$Url = "http://localhost:8888/api/feed/hot?limit=10",
    [string]$BombardierPath = "D:\bombardier\bombardier-windows-amd64.exe"
)

Write-Host "=== MiniFeed 快速性能测试 ===" -ForegroundColor Cyan
Write-Host "并发数: $Connections | 持续时间: ${Duration}s" -ForegroundColor Yellow

# 检查 bombardier 文件是否存在
if (-not (Test-Path $BombardierPath)) {
    Write-Host "`n❌ 未找到 bombardier 文件: $BombardierPath" -ForegroundColor Red
    Write-Host "请检查路径是否正确`n" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ 找到 bombardier: $BombardierPath`n" -ForegroundColor Green

# 获取 Token
Write-Host "获取 Token..." -ForegroundColor Yellow
try {
    $loginBody = @{
        username = "testuser"
        password = "123456"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "http://localhost:8888/user/login" `
        -Method Post `
        -ContentType "application/json" `
        -Body $loginBody -ErrorAction Stop
    
    $token = $response.data.token
    Write-Host "✓ Token 获取成功`n" -ForegroundColor Green
} catch {
    Write-Host "❌ 获取 Token 失败，请确保服务已启动且有 testuser/123456 账号" -ForegroundColor Red
    Write-Host "错误: $_`n" -ForegroundColor Red
    exit 1
}

# 执行测试
Write-Host "开始测试...`n" -ForegroundColor Green
& $BombardierPath -c $Connections -d "${Duration}s" -l `
    -H "Authorization: Bearer $token" `
    $Url

Write-Host "`n✓ 测试完成！" -ForegroundColor Green
Write-Host "查看 Prometheus 指标: http://localhost:8888/metrics`n" -ForegroundColor Cyan

