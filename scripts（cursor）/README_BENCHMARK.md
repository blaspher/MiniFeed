# MiniFeed 性能测试指南

## 📋 测试工具对比

| 工具 | 平台支持 | 安装难度 | 功能特点 | 推荐场景 |
|------|---------|---------|---------|---------|
| **bombardier** | Windows/Linux/macOS | ⭐⭐⭐⭐⭐ 简单 | Go 编写，跨平台，输出清晰 | ✅ Windows 首选 |
| **wrk** | Linux/macOS/WSL | ⭐⭐⭐ 中等 | C 编写，性能强，Lua 脚本支持 | Linux/WSL 首选 |

---

## 🚀 快速开始

### 方案 1: Bombardier (推荐 Windows 用户)

#### 1. 安装 bombardier

**选项 A: 使用 Scoop (推荐)**
```powershell
# 安装 Scoop (如果未安装)
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm get.scoop.sh | iex

# 安装 bombardier
scoop install bombardier
```

**选项 B: 手动下载**
1. 访问 https://github.com/codesenberg/bombardier/releases
2. 下载 `bombardier-windows-amd64.exe`
3. 重命名为 `bombardier.exe` 并添加到 PATH

**选项 C: 使用 Go 安装**
```powershell
go install github.com/codesenberg/bombardier@latest
```

#### 2. 运行测试

```powershell
# 进入项目目录
cd D:\GolangCode\minifeed

# 确保服务已启动
# go run cmd/server/main.go

# 执行测试脚本
powershell -ExecutionPolicy Bypass -File .\scripts\benchmark_bombardier.ps1
```

---

### 方案 2: wrk (推荐 Linux/WSL 用户)

#### 1. 安装 wrk

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install wrk
```

**macOS:**
```bash
brew install wrk
```

**WSL (Windows):**
```bash
# 在 WSL 终端中
sudo apt-get install wrk
```

#### 2. 运行测试

```bash
# 进入项目目录
cd /mnt/d/GolangCode/minifeed  # WSL 路径

# 确保服务已启动
# go run cmd/server/main.go

# 添加执行权限
chmod +x scripts/benchmark_wrk.sh

# 执行测试
./scripts/benchmark_wrk.sh
```

---

## 📊 测试场景说明

脚本包含 **4 个测试场景**，逐步增加负载：

| 场景 | 并发数 | 持续时间 | 目的 |
|------|--------|---------|------|
| 测试 1 | 10 | 10s | 基准性能测试（轻负载） |
| 测试 2 | 50 | 30s | 常规业务负载 |
| 测试 3 | 100 | 30s | 高峰期负载 |
| 测试 4 | 200 | 30s | 极限压力测试 |

---

## 📈 如何解读测试结果

### Bombardier 输出示例

```
Statistics        Avg      Stdev        Max
  Reqs/sec      5234.21    1243.56   12451.23
  Latency        2.34ms     1.12ms    45.67ms
  HTTP codes:
    1xx - 0, 2xx - 157026, 3xx - 0, 4xx - 0, 5xx - 0
  Throughput:     1.23MB/s
```

**关键指标：**
- **Reqs/sec (QPS)**：每秒请求数，越高越好
  - 优秀：> 5000 QPS
  - 良好：1000-5000 QPS
  - 一般：< 1000 QPS

- **Latency (延迟)**：请求响应时间
  - 优秀：< 10ms (P99)
  - 良好：10-50ms (P99)
  - 一般：> 50ms (P99)

- **HTTP codes**：状态码分布
  - 2xx：成功请求（应为 100%）
  - 5xx：服务器错误（应为 0）

### wrk 输出示例

```
Running 30s test @ http://localhost:8888/api/feed/hot
  4 threads and 50 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     2.34ms    1.12ms   45.67ms   91.23%
    Req/Sec     1.31k   234.45     2.45k    78.34%
  Latency Distribution
     50%    2.12ms
     75%    2.89ms
     90%    3.67ms
     99%    8.45ms
  157026 requests in 30.02s, 45.23MB read
Requests/sec:   5234.21
Transfer/sec:      1.51MB
```

**关键指标：**
- **Latency Distribution**：延迟分位数
  - P50 (中位数)：50% 请求的延迟
  - P99：99% 请求的延迟（关注长尾）

- **Requests/sec**：总 QPS

---

## 🔍 对比测试（验证优化效果）

### 测试热门接口缓存效果

#### 1. 测试带缓存的性能
```bash
# 正常运行（有缓存）
bombardier -c 100 -d 30s -l \
  -H "Authorization: Bearer TOKEN" \
  http://localhost:8888/api/feed/hot?limit=10
```

#### 2. 测试无缓存性能（模拟缓存失效）
```bash
# 先清空 Redis 缓存
redis-cli FLUSHDB

# 立即测试（缓存未命中）
bombardier -c 100 -d 30s -l \
  -H "Authorization: Bearer TOKEN" \
  http://localhost:8888/api/feed/hot?limit=10
```

#### 3. 对比指标
记录以下数据：

| 场景 | QPS | P50 延迟 | P99 延迟 | 错误率 |
|------|-----|---------|---------|--------|
| 有缓存 | ? | ? | ? | ? |
| 无缓存 | ? | ? | ? | ? |

**预期结果：**
- 有缓存 QPS 应该显著高于无缓存（5-10倍）
- 有缓存 P99 延迟应该 < 10ms
- 无缓存 P99 延迟可能 > 50ms

---

## 📊 Prometheus 监控指标

测试期间可以访问 Prometheus 指标端点：

```bash
curl http://localhost:8888/metrics
```

**关键指标：**

1. **请求总数**
```
http_requests_total{method="GET",path="/api/feed/hot",status="200"} 157026
```

2. **请求延迟分布**
```
http_request_duration_seconds_bucket{method="GET",path="/api/feed/hot",le="0.005"} 145234
http_request_duration_seconds_bucket{method="GET",path="/api/feed/hot",le="0.01"} 156234
http_request_duration_seconds_bucket{method="GET",path="/api/feed/hot",le="0.05"} 156989
```

---

## 🎯 性能优化目标

### 推荐基准（Redis 缓存命中场景）

| 指标 | 目标值 | 说明 |
|------|--------|------|
| **QPS** | > 3000 | 单机 3000+ QPS |
| **P50 延迟** | < 5ms | 中位数延迟 |
| **P99 延迟** | < 20ms | 99% 请求延迟 |
| **错误率** | 0% | 无 5xx 错误 |
| **吞吐量** | > 1MB/s | 数据传输速率 |

### 如果性能不达标

**常见问题排查：**

1. **QPS 过低**
   - 检查数据库连接池配置
   - 检查 Redis 连接是否正常
   - 查看 CPU/内存占用

2. **延迟过高**
   - 检查是否频繁查询数据库（缓存未命中）
   - 查看慢查询日志
   - 检查网络延迟

3. **错误率高**
   - 查看服务日志
   - 检查数据库连接数是否耗尽
   - 检查 Redis 连接数

---

## 🛠️ 手动测试命令

如果脚本无法运行，可以手动执行：

### Bombardier 手动命令

```powershell
# 获取 Token (先登录)
$response = Invoke-RestMethod -Uri "http://localhost:8888/user/login" `
    -Method Post `
    -ContentType "application/json" `
    -Body '{"username":"testuser","password":"123456"}'
$token = $response.data.token

# 执行测试
bombardier -c 100 -d 30s -l `
    -H "Authorization: Bearer $token" `
    "http://localhost:8888/api/feed/hot?limit=10"
```

### wrk 手动命令

```bash
# 获取 Token
TOKEN=$(curl -s -X POST http://localhost:8888/user/login \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"123456"}' \
    | grep -o '"token":"[^"]*' | cut -d'"' -f4)

# 创建 Lua 脚本
cat > auth.lua << EOF
wrk.headers["Authorization"] = "Bearer $TOKEN"
EOF

# 执行测试
wrk -t4 -c100 -d30s --latency \
    -s auth.lua \
    "http://localhost:8888/api/feed/hot?limit=10"
```

---

## 📝 测试报告模板

记录测试结果：

```markdown
# MiniFeed 性能测试报告

## 测试环境
- 操作系统: Windows 10
- CPU: ?
- 内存: ?
- Go 版本: go version
- Redis 版本: redis-server --version

## 测试结果

### 场景 1: 轻负载 (10 并发)
- QPS: ?
- P50 延迟: ?
- P99 延迟: ?
- 错误率: ?

### 场景 2: 中负载 (50 并发)
- QPS: ?
- P50 延迟: ?
- P99 延迟: ?
- 错误率: ?

### 场景 3: 高负载 (100 并发)
- QPS: ?
- P50 延迟: ?
- P99 延迟: ?
- 错误率: ?

### 场景 4: 极限测试 (200 并发)
- QPS: ?
- P50 延迟: ?
- P99 延迟: ?
- 错误率: ?

## 缓存效果对比
| 场景 | QPS | P99 延迟 | 提升倍数 |
|------|-----|---------|---------|
| 无缓存 | ? | ? | - |
| 有缓存 | ? | ? | ?x |

## 结论
（总结性能表现和优化建议）
```

---

## 🆘 常见问题

### Q1: Token 获取失败？
**A:** 确保服务已启动，可以手动访问 `http://localhost:8888/user/login` 测试

### Q2: 所有请求返回 401？
**A:** Token 过期或无效，重新运行脚本获取新 Token

### Q3: 性能远低于预期？
**A:** 检查以下项目：
- Redis 是否正常运行
- 数据库连接池配置
- 测试机器资源占用情况

### Q4: Windows 执行脚本提示权限错误？
**A:** 以管理员身份运行 PowerShell，并执行：
```powershell
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

---

## 📚 参考资料

- [bombardier GitHub](https://github.com/codesenberg/bombardier)
- [wrk GitHub](https://github.com/wg/wrk)
- [Prometheus 指标说明](https://prometheus.io/docs/concepts/metric_types/)

