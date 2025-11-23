# MiniFeed 完整性能测试（4个场景）
# 使用方法: .\scripts\test_full.ps1

Write-Host "=== MiniFeed 完整性能测试 ===" -ForegroundColor Cyan
Write-Host ""

# 检查 bombardier
$bombPath = "D:\bombardier\bombardier.exe"
if (-not (Test-Path $bombPath)) {
    # 尝试从 PATH 中找
    $bombCmd = Get-Command bombardier -ErrorAction SilentlyContinue
    if ($bombCmd) {
        $bombPath = "bombardier"
    } else {
        Write-Host "错误: 找不到 bombardier" -ForegroundColor Red
        Write-Host "请先安装或检查路径" -ForegroundColor Yellow
        exit 1
    }
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

# 测试场景配置
$tests = @(
    @{Name="轻负载"; Connections=10; Duration=10},
    @{Name="中负载"; Connections=50; Duration=30},
    @{Name="高负载"; Connections=100; Duration=30},
    @{Name="极限测试"; Connections=200; Duration=30}
)

# 执行测试
for ($i = 0; $i -lt $tests.Count; $i++) {
    $test = $tests[$i]
    Write-Host "=== 测试 $($i+1): $($test.Name) ($($test.Connections) 并发, $($test.Duration)秒) ===" -ForegroundColor Magenta
    Write-Host ""
    
    $cmd = "$bombPath -c $($test.Connections) -d $($test.Duration)s -l -H `"Authorization: Bearer $token`" http://localhost:8888/api/feed/hot?limit=10"
    Invoke-Expression $cmd
    
    if ($i -lt $tests.Count - 1) {
        Write-Host ""
        Write-Host "等待 2 秒..." -ForegroundColor Gray
        Start-Sleep -Seconds 2
        Write-Host ""
    }
}

Write-Host ""
Write-Host "=== 所有测试完成！===" -ForegroundColor Green
Write-Host "查看指标: http://localhost:8888/metrics" -ForegroundColor Cyan

