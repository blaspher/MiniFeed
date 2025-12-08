# MiniFeed：基于 Go 的高并发内容 Feed 流系统

🧩 1. 项目简介  
MiniFeed 是一个仿微博/小红书/抖音底层能力的 Feed 流系统，基于 Go + MySQL + Redis 实现，包含用户、关注、发布动态、点赞、高性能 Feed 流模块，支持拉模式和推模式。  
适合作为：移动互联网后台实习项目 / Go Web 入门实战 / 内容推荐系统入门 Demo。

🚀 2. 技术栈  
- Go（Gin / Gorm）  
- MySQL 8  
- Redis（ZSet / Set / String）  
- JWT 权限认证  
- Docker Compose（本地一键起环境）  
- Prometheus 监控

🎯 3. 功能点  
- 用户注册登录（JWT）  
- 发布动态（图文）  
- 关注 / 取关  
- 点赞（Redis + MySQL，异步落库）  
- Feed 流查询（拉模式）  
- Redis Inbox（推模式）  
- 热门动态缓存（定时刷新 + 双删）  
- 游标分页（cursor）

🧱 4. 系统架构图  
后端层次：API（Gin）→ Service → DAO（Gorm）→ MySQL / Redis；定时任务同步点赞与热门榜单；Prometheus 暴露指标。  
目录参考：`cmd/server`（入口）+ `internal/{api,service,dao,cron,metrics,middleware,model,config}` + `pkg/jwt`。

🗄 5. 数据库表（简要）  
- users：id, username, password_hash, created_at  
- posts：id, user_id, content, like_count, created_at  
- follows：follower_id, followee_id, created_at  
建表 SQL 可参考 `internal/model` 自动迁移生成的结构。

🔥 6. 如何运行  
1) 准备环境：Docker、Docker Compose、Go 1.22+（本机调试用）。  
2) 配置环境变量（示例）：  
   - `MYSQL_DSN=user:pass@tcp(mysql:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local`  
   - `REDIS_ADDR=redis:6379`  
   - `JWT_SECRET=your-jwt-secret`  
3) 启动（推荐容器化）：  
   - 一键脚本：  
     - Windows: `.\scripts\start.ps1`  
     - Linux/macOS: `bash scripts/start.sh`  
   - 或直接执行：  
     ```bash
     docker-compose up -d --build
     docker-compose ps
     ```  
   访问接口 `http://localhost:8888`，测试页面 `http://localhost:8888/test.html`，监控 `http://localhost:8888/metrics`。  
4) 本地直跑（需自备 MySQL/Redis）：  
   ```bash
   go mod tidy
   go run cmd/server/main.go
   ```


