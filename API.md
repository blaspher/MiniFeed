# API 接口文档（MiniFeed）

统一返回格式：
```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```
`code != 0` 表示业务错误。

基础地址默认 `http://localhost:8888`，需要鉴权的接口在 Header 携带：
```
Authorization: Bearer <JWT_TOKEN>
```

## 用户

- 注册 `POST /user/register`（无需鉴权）  
  ```bash
  curl -X POST http://localhost:8888/user/register \
    -H "Content-Type: application/json" \
    -d '{"username":"alice","password":"123456"}'
  ```

- 登录 `POST /user/login`（无需鉴权）  
  ```bash
  curl -X POST http://localhost:8888/user/login \
    -H "Content-Type: application/json" \
    -d '{"username":"alice","password":"123456"}'
  ```
  响应中 `token` 即 JWT。

- 搜索用户 `GET /api/users/search?keyword=al`（鉴权）  
  ```bash
  curl "http://localhost:8888/api/users/search?keyword=al" \
    -H "Authorization: Bearer <JWT_TOKEN>"
  ```

- 当前用户信息 `GET /api/me`（鉴权）  
  ```bash
  curl http://localhost:8888/api/me \
    -H "Authorization: Bearer <JWT_TOKEN>"
  ```

## 帖子 / Feed

- 发帖 `POST /api/post`（鉴权）  
  ```bash
  curl -X POST http://localhost:8888/api/post \
    -H "Authorization: Bearer <JWT_TOKEN>" \
    -H "Content-Type: application/json" \
    -d '{"content":"hello world","image_url":"https://example.com/img.png"}'
  ```

- 点赞/取消点赞 `POST /api/post/:id/like`（鉴权）  
  ```bash
  curl -X POST http://localhost:8888/api/post/1/like \
    -H "Authorization: Bearer <JWT_TOKEN>"
  ```

- 公共流（按时间，游标分页）`GET /posts?limit=10&cursor=<last_id>`（公开）  
  ```bash
  curl "http://localhost:8888/posts?limit=10"
  ```

- 关注流 Pull 模式 `GET /api/feed/pull?limit=10&cursor=<last_id>`（鉴权）  
  ```bash
  curl "http://localhost:8888/api/feed/pull?limit=10" \
    -H "Authorization: Bearer <JWT_TOKEN>"
  ```

- Inbox 推模式 `GET /api/feed/push?limit=10&cursor=<cursor>`（鉴权）  
  ```bash
  curl "http://localhost:8888/api/feed/push?limit=10" \
    -H "Authorization: Bearer <JWT_TOKEN>"
  ```

- 热门流 `GET /api/feed/hot?limit=10`（鉴权）  
  ```bash
  curl "http://localhost:8888/api/feed/hot?limit=10" \
    -H "Authorization: Bearer <JWT_TOKEN>"
  ```

## 关注

- 关注 `POST /api/follow/:id`（鉴权）  
  ```bash
  curl -X POST http://localhost:8888/api/follow/2 \
    -H "Authorization: Bearer <JWT_TOKEN>"
  ```

- 取关 `POST /api/unfollow/:id`（鉴权）  
  ```bash
  curl -X POST http://localhost:8888/api/unfollow/2 \
    -H "Authorization: Bearer <JWT_TOKEN>"
  ```

- 关注列表 `GET /api/following`（鉴权）  
  ```bash
  curl http://localhost:8888/api/following \
    -H "Authorization: Bearer <JWT_TOKEN>"
  ```

- 粉丝列表 `GET /api/followers`（鉴权）  
  ```bash
  curl http://localhost:8888/api/followers \
    -H "Authorization: Bearer <JWT_TOKEN>"
  ```

## 监控

- Prometheus 指标 `GET /metrics`（公开）  
  ```bash
  curl http://localhost:8888/metrics
  ```


