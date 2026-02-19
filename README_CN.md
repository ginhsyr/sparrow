# Sparrow

Sparrow 是一个基于 Gin + GORM 的轻量论坛后端服务。  
当前实现了用户注册登录、帖子发布与查询、点赞能力，以及基于 SSE 的通知推送。

## 技术栈

- Go `1.24`
- Gin（`github.com/gin-gonic/gin`）
- GORM + PostgreSQL
- JWT（`github.com/golang-jwt/jwt/v5`）
- Zap 日志
- Docker / Docker Compose

## 功能特性

- 用户注册、登录（登录后签发 JWT）
- 基于角色的鉴权中间件
- 帖子创建与查询（支持按需加载正文和编辑历史）
- 帖子点赞
- SSE 实时通知订阅接口
- 启动时自动建表/迁移
- `post_likes` 哈希分区表（按 `post_id`，默认 64 个分区）

## 运行要求

- Go `1.24+`
- PostgreSQL `15+`（或直接使用 Docker Compose）
- 项目根目录存在可用的 `.env`

## 本地启动

1. 复制环境变量模板：

```bash
cp example.env .env
```

2. 按需修改 `.env`（尤其是 `JWT_SIGNING_KEY`，长度必须不少于 32）。

3. 启动 PostgreSQL：

```bash
docker compose up -d db
```

4. 启动服务：

```bash
go run .
```

默认访问地址：`http://localhost:8025`

## Docker Compose 一键启动

```bash
docker compose up --build
```

- API 地址：`http://localhost:8025`
- PostgreSQL 地址：`localhost:5432`

## 环境变量说明

| 变量名 | 必填 | 说明 |
| --- | --- | --- |
| `PORT` | 是 | HTTP 服务端口（示例值 `8025`） |
| `DB_HOST` | 是 | PostgreSQL 主机（Docker 内通常为 `db`，本地常用 `127.0.0.1`） |
| `DB_USER` | 是 | PostgreSQL 用户名 |
| `DB_PASSWORD` | 是 | PostgreSQL 密码 |
| `DB_NAME` | 是 | PostgreSQL 数据库名 |
| `DB_PORT` | 是 | PostgreSQL 端口 |
| `JWT_SIGNING_KEY` | 是 | JWT 签名密钥，至少 32 个字符 |
| `LOG_LEVEL` | 是 | Zap 日志级别（如 `debug`、`info`、`warn`、`error`） |

## API 概览

基础前缀：`/api/v1`

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/ping` | 否 | 健康检查，返回 `"pong"` |
| `GET` | `/user/:id` | 否 | 按用户 ID 查询 |
| `POST` | `/auth/register` | 否 | 注册 |
| `POST` | `/auth/login` | 否 | 登录并返回 JWT |
| `GET` | `/posts/:id` | 否 | 按帖子 ID 查询（支持查询参数） |
| `POST` | `/posts` | Bearer Token + 角色 | 发帖 |
| `POST` | `/posts/:id/like` | Bearer Token + 角色 | 点赞 |
| `GET` | `/subscribe/notify` | Bearer Token + 角色 | SSE 订阅通知 |

说明：旧接口 `POST /posts/like` 已移除。

### `GET /posts/:id` 查询参数

- `includeContent`（`true/false`，默认 `true`）
- `includeEdits`（`true/false`，默认 `false`）
- `editsLimit`（`1-200`，默认 `20`，仅在 `includeEdits=true` 时生效）

### 请求参数校验

- `POST /auth/register`
- `nickname`：必填，不能全空白，最大长度 `20`
- `realName`：必填，不能全空白，最大长度 `20`
- `email`：必填，不能全空白，必须为合法邮箱，最大长度 `254`
- `password`：必填，长度 `8-72`，且必须包含大写字母 + 小写字母 + 数字
- `birthday`：可选，Unix 时间戳（秒）

- `POST /auth/login`
- `email`：必填，不能全空白，必须为合法邮箱，最大长度 `254`
- `password`：必填，不能全空白

- `POST /posts`
- `content`：必填，不能全空白，最大长度 `2000`

校验失败统一返回格式：

```json
{
  "error": "invalid request parameters",
  "details": [
    {
      "field": "email",
      "message": "email must be a valid email"
    }
  ]
}
```

### 受保护接口请求头

```http
Authorization: Bearer <token>
```

受保护接口允许的角色：`Users`、`Admin`。

## 请求示例

注册：

```bash
curl -X POST http://localhost:8025/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "nickname":"jack",
    "realName":"Jack Sparrow",
    "email":"jack@example.com",
    "password":"StrongPass123",
    "birthday":946684800
  }'
```

登录：

```bash
curl -X POST http://localhost:8025/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email":"jack@example.com",
    "password":"StrongPass123"
  }'
```

发帖：

```bash
curl -X POST http://localhost:8025/api/v1/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"content":"hello sparrow"}'
```

点赞：

```bash
curl -X POST http://localhost:8025/api/v1/posts/1/like \
  -H "Authorization: Bearer <token>"
```

点赞响应（首次点赞，`201 Created`）：

```json
{
  "postID": 1,
  "userID": 1001,
  "liked": true
}
```

点赞响应（重复点赞，`200 OK`）：

```json
{
  "postID": 1,
  "userID": 1001,
  "liked": false,
  "message": "already liked"
}
```

SSE 订阅：

```bash
curl -N http://localhost:8025/api/v1/subscribe/notify \
  -H "Authorization: Bearer <token>"
```

## 开发命令

构建：

```bash
go build -o bin/sparrow
```

格式化与静态检查：

```bash
go fmt ./...
go vet ./...
```

测试：

```bash
go test ./...
```

## 目录结构

```text
.
├── main.go
├── configs/
├── internal/
│   ├── handler/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   ├── router/
│   ├── service/
│   └── utils/
├── tests/
├── Dockerfile
└── docker-compose.yml
```

## 说明

- 应用启动时会读取 `.env`；若缺失会直接退出。
- 日志文件输出到 `logs/sparrow.log` 与 `logs/zap.log`。
- Docker Compose 场景下 `DB_HOST` 应设置为 `db`；本地直连数据库通常使用 `127.0.0.1`。
