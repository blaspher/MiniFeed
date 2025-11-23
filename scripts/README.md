# MiniFeed 性能测试脚本使用指南

## 📁 脚本文件说明

| 文件 | 用途 | 推荐场景 |
|------|------|---------|
| `test_simple.ps1` | 单场景快速测试 | ✅ 日常测试、快速验证 |
| `test_full.ps1` | 完整 4 场景测试 | 📊 完整性能评估 |
| `README_BENCHMARK.md` | 详细使用文档 | 📚 安装和详细说明 |

---

## 🚀 快速开始

### 前提条件

1. **服务已启动**
   ```powershell
   # 在一个终端运行
   cd D:\GolangCode\minifeed
   go run cmd/server/main.go
   ```

2. **测试用户已创建**
   ```powershell
   # 创建 testuser/123456 账号
   Invoke-RestMethod -Uri "http://localhost:8888/user/register" -Method Post -ContentType "application/json" -Body '{"username":"testuser","password":"123456"}'
   ```

3. **bombardier 已安装**
   - 已下载到：`D:\bombardier\bombardier.exe`
   - 或者已添加到 PATH

---

## 📝 使用方法

### 方式 1: 快速测试（推荐）

```powershell
# 默认参数（50 并发，10 秒）
.\scripts\test_simple.ps1

# 自定义参数
.\scripts\test_simple.ps1 -Connections 100 -Duration 30
```

### 方式 2: 完整测试

```powershell
# 运行 4 个场景测试（10/50/100/200 并发）
.\scripts\test_full.ps1
```

### 方式 3: 手动测试（最可靠）

```powershell
# 获取 Token 并测试
$r = Invoke-RestMethod -Uri "http://localhost:8888/user/login" -Method Post -ContentType "application/json" -Body '{"username":"testuser","password":"123456"}'
& "D:\bombardier\bombardier.exe" -c 50 -d 10s -l -H "Authorization: Bearer $($r.data.token)" "http://localhost:8888/api/feed/hot?limit=10"
```

---

## 🔧 常见问题

### Q1: 脚本提示 "找不到 bombardier"

**解决方案：**
```powershell
# 方式 A: 修改脚本中的路径
# 编辑 test_simple.ps1 或 test_full.ps1
# 将 $bombPath = "D:\bombardier\bombardier.exe" 改为你的实际路径

# 方式 B: 添加到 PATH 后重启 PowerShell
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";D:\bombardier", "User")
```

### Q2: Token 获取失败

**检查清单：**
- [ ] 服务是否正常运行（访问 http://localhost:8888）
- [ ] testuser 账号是否存在
- [ ] 数据库是否正常连接

### Q3: 脚本无法执行

```powershell
# 允许脚本执行
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

---

## 📊 测试结果解读

### 关键指标

| 指标 | 含义 | 目标值 |
|------|------|--------|
| **Reqs/sec** | QPS（每秒请求数） | > 3000 |
| **Latency (Avg)** | 平均延迟 | < 20ms |
| **P99** | 99% 请求延迟 | < 50ms |
| **HTTP 2xx** | 成功请求比例 | 100% |

### 性能评级

- 🟢 **优秀**: QPS > 5000, P99 < 20ms
- 🟡 **良好**: QPS 3000-5000, P99 20-50ms
- 🔴 **需优化**: QPS < 3000, P99 > 50ms

---

## 🎯 验证缓存效果

```powershell
# 1. 测试有缓存性能
.\scripts\test_simple.ps1 -Connections 100 -Duration 30

# 2. 清空 Redis
redis-cli FLUSHDB

# 3. 测试无缓存性能（立即执行）
.\scripts\test_simple.ps1 -Connections 100 -Duration 30

# 对比两次结果
```

**预期：** 有缓存的 QPS 应该是无缓存的 5-10 倍

---

## 📈 监控指标

测试期间查看：
```powershell
# Prometheus 指标
start http://localhost:8888/metrics

# 查看请求总数
curl http://localhost:8888/metrics | Select-String "http_requests_total"
```

---

## 💡 性能优化建议

根据测试结果：

1. **QPS < 3000**: 检查数据库查询、缓存命中率
2. **延迟 > 50ms**: 检查网络、数据库连接池
3. **错误率 > 0**: 查看日志排查具体错误

---

## 📞 技术支持

如有问题，检查：
1. 服务日志
2. Redis 状态：`redis-cli ping`
3. MySQL 状态：`mysql -u root -p -e "SELECT 1"`

