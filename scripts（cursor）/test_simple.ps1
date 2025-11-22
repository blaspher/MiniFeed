# MiniFeed 简单性能测试
# 使用方法: .\scripts\test_simple.ps1

param(
    [int]$Connections = 50,
    [int]$Duration = 10
)

Write-Host "=== MiniFeed 性能测试 ===" -ForegroundColor Cyan
Write-Host "并发: $Connections | 时长: ${Duration}s"
Write-Host ""

# 检查 bombardier
$bombPath = "D:\bombardier\bombardier.exe"
if (-not (Test-Path $bombPath)) {
    Write-Host "错误: 找不到 bombardier" -ForegroundColor Red
    Write-Host "路径: $bombPath" -ForegroundColor Yellow
    exit 1
}

# 获取 Token
Write-Host "获取 Token..." -ForegroundColor Yellow
try {
    $loginBody = '{"username":"testuser","password":"123456"}'
    $response = Invoke-RestMethod -Uri "http://localhost:8888/user/login" -Method Post -ContentType "application/json" -Body $loginBody
    $token = $response.data.token
    Write-Host "Token 获取成功" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "Token 获取失败: $_" -ForegroundColor Red
    exit 1
}

# 运行测试
Write-Host "开始测试..." -ForegroundColor Green
Write-Host ""

$cmd = "$bombPath -c $Connections -d ${Duration}s -l -H `"Authorization: Bearer $token`" http://localhost:8888/api/feed/hot?limit=10"
Invoke-Expression $cmd

Write-Host ""
Write-Host "测试完成!" -ForegroundColor Green

