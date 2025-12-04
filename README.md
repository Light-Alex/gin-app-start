# 参考
https://github.com/pengfeidai/gin-app-start

改动：

1. 优化zap日志打印
2. 部分代码添加注释
3. 添加config.local.yaml文件
4. 修复air工具使用说明问题
5. 修复docker-compose.yml，容器时间同步问题
6. 添加order订单模块
7. 添加redis业务逻辑



<hr>


# Gin App Start

基于 [Gin](https://github.com/gin-gonic/gin) 框架的现代化 Go Web 应用脚手架，遵循清晰的分层架构设计，支持 PostgreSQL 和 Redis。

> ⚡ **最新版本**: v2.0.0 - 已升级到 Go 1.24 和最新依赖包

## 📚 完整文档

- 📖 **[项目使用指南](docs/PROJECT_GUIDE.md)** - 详细的项目文档（推荐）
- 🔌 **[API 接口文档](docs/API_REFERENCE.md)** - 完整的 API 参考
- 🏗️ **[架构设计文档](docs/ARCHITECTURE.md)** - 技术架构深度解析

## 特性

- ✅ 清晰的分层架构（Controller -> Service -> Repository）
- ✅ PostgreSQL 数据库支持
- ✅ Redis 缓存支持
- ✅ 结构化日志（zap）
- ✅ 统一错误处理
- ✅ 统一响应格式
- ✅ 中间件支持（日志、恢复、限流、CORS）
- ✅ 优雅关闭
- ✅ 环境配置管理
- ✅ 自动数据库迁移

## 目录结构

```
gin-app-start/
├── cmd/                    # 应用程序入口
│   └── server/
│       └── main.go        # 主入口文件
├── internal/              # 私有应用程序代码
│   ├── config/           # 配置加载
│   ├── controller/       # HTTP 控制器层
│   ├── service/          # 业务逻辑层
│   ├── repository/       # 数据访问层
│   ├── model/            # 数据模型
│   ├── middleware/       # Gin 中间件
│   └── router/           # 路由配置
├── pkg/                  # 公共库代码
│   ├── database/         # 数据库连接
│   ├── logger/           # 日志处理
│   ├── errors/           # 错误处理
│   └── response/         # 统一响应格式
├── configs/              # 配置文件
│   ├── config.local.yaml
│   ├── config.dev.yaml
│   └── config.prod.yaml
├── go.mod
└── go.sum
```

## 快速开始

### 环境要求

- Go >= 1.24
- PostgreSQL >= 17
- Redis >= 7.0
- Kafka >= 4.0

### 安装依赖

```bash
go mod download
```

### 配置数据库

1. 创建 PostgreSQL 数据库：

```sql
CREATE DATABASE gin_app;
```

2. 修改配置文件 `configs/config.local.yaml`：

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: gin_app
  sslmode: disable
```

### 运行应用

```bash
# 本地环境
export SERVER_ENV=local && go run cmd/server/main.go

# 开发环境
export SERVER_ENV=dev && go run cmd/server/main.go

# 生产环境
export SERVER_ENV=prod &&  go run cmd/server/main.go
```

### 健康检查

```bash
curl http://localhost:9060/health
```

## API 文档

### 健康检查

```bash
GET /health
```

### 用户管理

#### 创建用户

**request：**
```bash
POST /api/v1/users
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "phone": "13800138000",
  "password": "password123"
}
```

**response：**
- 成功响应：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 2,
        "created_at": "2025-12-03T11:55:06.317239131+08:00",
        "update_at": "2025-12-03T11:55:06.317239221+08:00",
        "username": "testuser16",
        "email": "testuser16@example.com",
        "phone": "13800138016",
        "avatar": "",
        "status": 1
    }
}
```
- 错误响应：
```json
{
    "code": 10001,
    "message": "Parameter binding failed: Key: 'CreateUserRequest.Username' Error:Field validation for 'Username' failed on the 'required' tag",
    "data": null
}
```

#### 获取用户

**request：**
```bash
GET /api/v1/users/:id
```
**response：**
- 成功响应：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 1,
        "created_at": "2025-11-26T22:01:26.823447+08:00",
        "update_at": "2025-11-26T22:01:26.823447+08:00",
        "username": "testuser15",
        "email": "testuser15@example.com",
        "phone": "13800138015",
        "avatar": "",
        "status": 1
    }
}
```
- 错误响应：
```json
{
    "code": 10002,
    "message": "User not found",
    "data": null
}
```

#### 更新用户
**request：**
```bash
PUT /api/v1/users/:id
Content-Type: application/json

{
  "email": "newemail@example.com",
  "phone": "13900139000"
}
```

**response：**
- 成功响应：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 1,
        "created_at": "2025-11-26T22:01:26.823447+08:00",
        "update_at": "2025-12-03T11:58:43.290084569+08:00",
        "username": "testuser15",
        "email": "newemail2@example.com",
        "phone": "13900139020",
        "avatar": "",
        "status": 1
    }
}
```
- 错误响应：
```json
{
    "code": 10002,
    "message": "User not found",
    "data": null
}
```

#### 删除用户
**request：**
```bash
DELETE /api/v1/users/:id
```

**response：**
- 成功响应：
```json
{
    "code": 0,
    "message": "Deleted successfully",
    "data": null
}
```

#### 用户列表
**request：**
```bash
GET /api/v1/users?page=1&page_size=10
```
**response：**
- 成功响应：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [
            {
                "id": 1,
                "created_at": "2025-11-26T22:01:26.823447+08:00",
                "update_at": "2025-12-03T11:58:43.290084+08:00",
                "username": "testuser15",
                "email": "newemail2@example.com",
                "phone": "13900139020",
                "avatar": "",
                "status": 1
            }
        ],
        "total": 1,
        "page": 1,
        "page_size": 10
    }
}
```

### 订单管理

#### 创建订单
**request：**
```bash
POST /api/v1/orders
Content-Type: application/json

{
  "user_id": 1,
  "total_price": 200.00,
  "description": "Good product"
}
```
**response：**
- 成功响应：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 6,
        "order_number": "EC20251203974341",
        "created_at": "2025-12-03T12:03:58.328230483+08:00",
        "update_at": "2025-12-03T12:03:58.328230563+08:00",
        "user_id": 5,
        "total_price": 50,
        "description": "Bad!",
        "status": 1
    }
}
```
- 错误响应：
```json
{
    "code": 10001,
    "message": "Parameter binding failed: invalid character ',' looking for beginning of value",
    "data": null
}
```
#### 获取订单
**request：**
```bash
GET /api/v1/orders/:order_number
```
**response：**
- 成功响应：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 2,
        "order_number": "EC20251202659066",
        "created_at": "2025-12-02T23:06:46.861499+08:00",
        "update_at": "2025-12-02T23:06:46.861499+08:00",
        "user_id": 2,
        "total_price": 100.99,
        "description": "want more!",
        "status": 1
    }
}
```
- 错误响应：
```json
{
    "code": 10029,
    "message": "Set empty cache",
    "data": null
}
```

#### 更新订单
**request：**
```bash
PUT /api/v1/orders/:order_number
Content-Type: application/json

{
  "description": "Order for John Doe",
  "status": 1,
  "total_price": 99.99
}
```

**response：**
- 成功响应：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 2,
        "order_number": "EC20251202659066",
        "created_at": "2025-12-02T23:06:46.861499+08:00",
        "update_at": "2025-12-03T12:06:36.147182487+08:00",
        "user_id": 2,
        "total_price": 40,
        "description": "Bad product!!!",
        "status": 1
    }
}
```
- 错误响应：
```json
{
    "code": 10023,
    "message": "Order not found",
    "data": null
}
```

#### 删除订单
**request：**
```bash
DELETE /api/v1/orders/:order_number
```
**response：**
- 成功响应：
```json
{
    "code": 0,
    "message": "Deleted successfully",
    "data": null
}
```
- 错误响应：
```json
{
    "code": 10023,
    "message": "Order not found",
    "data": null
}
```

#### 订单列表
**request：**
```bash
GET /api/v1/orders?page=1&page_size=10
```
**response：**
- 成功响应：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [
            {
                "id": 3,
                "order_number": "EC20251202905265",
                "created_at": "2025-12-02T23:52:01.054351+08:00",
                "update_at": "2025-12-02T23:52:01.054351+08:00",
                "user_id": 2,
                "total_price": 100.99,
                "description": "want more!",
                "status": 1
            },
            {
                "id": 4,
                "order_number": "EC20251202658812",
                "created_at": "2025-12-02T23:52:36.459454+08:00",
                "update_at": "2025-12-02T23:52:36.459454+08:00",
                "user_id": 3,
                "total_price": 50,
                "description": "Bad!",
                "status": 1
            },
            {
                "id": 5,
                "order_number": "EC20251203141066",
                "created_at": "2025-12-03T00:00:26.47771+08:00",
                "update_at": "2025-12-03T00:00:26.477711+08:00",
                "user_id": 4,
                "total_price": 66,
                "description": "Good!",
                "status": 1
            },
            {
                "id": 6,
                "order_number": "EC20251203974341",
                "created_at": "2025-12-03T12:03:58.32823+08:00",
                "update_at": "2025-12-03T12:03:58.32823+08:00",
                "user_id": 5,
                "total_price": 50,
                "description": "Bad!",
                "status": 1
            }
        ],
        "total": 4,
        "page": 1,
        "page_size": 10
    }
}
```

## 配置说明

### 服务器配置

```yaml
server:
  port: 9060              # 服务端口
  mode: debug             # 运行模式: debug/release/test
  read_timeout: 60        # 读超时（秒）
  write_timeout: 60       # 写超时（秒）
  limit_num: 100          # 限流数（每秒请求数）
```

### 数据库配置

```yaml
database:
  host: localhost         # 数据库主机
  port: 5432             # 数据库端口
  user: postgres         # 数据库用户
  password: postgres     # 数据库密码
  dbname: gin_app        # 数据库名
  sslmode: disable       # SSL模式
  max_idle_conns: 10     # 最大空闲连接数
  max_open_conns: 100    # 最大打开连接数
  max_lifetime: 3600     # 连接最大生命周期（秒）
  log_level: info        # 日志级别
  auto_migrate: true     # 自动迁移
```

### Redis配置

```yaml
redis:
  addr: localhost:6379   # Redis地址
  password: ""           # Redis密码
  db: 0                  # Redis数据库
  pool_size: 10          # 连接池大小
  min_idle_conns: 5      # 最小空闲连接数
  max_retries: 3         # 最大重试次数
```

## Docker 部署

### 构建镜像

```bash
docker build -t gin-app-start .
```

### 运行容器

```bash
docker run -d \
  -p 9060:9060 \
  -e SERVER_ENV=prod \
  -e DB_HOST=postgres \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres \
  -e DB_NAME=gin_app \
  -e REDIS_ADDR=redis:6379 \
  -e REDIS_PASSWORD="" \
  gin-app-start
```

## 开发指南

### 添加新的 API

1. 在 `internal/model` 中定义数据模型
2. 在 `internal/repository` 中实现数据访问层
3. 在 `internal/service` 中实现业务逻辑
4. 在 `internal/controller` 中实现控制器
5. 在 `internal/router` 中注册路由

### 错误处理

使用 `pkg/errors` 包定义和处理业务错误：

```go
import "gin-app-start/pkg/errors"

// 使用预定义错误
return errors.ErrUserNotFound

// 创建新错误
return errors.NewBusinessError(10001, "自定义错误消息")

// 包装错误
return errors.WrapBusinessError(10001, "操作失败", err)
```

### 日志记录

使用 `pkg/logger` 包记录日志：

```go
import (
    "gin-app-start/pkg/logger"
    "go.uber.org/zap"
)

logger.Info("操作成功", 
    zap.String("username", username),
    zap.Uint("user_id", userID),
)

logger.Error("操作失败", 
    zap.Error(err),
)
```

## 许可证

MIT License
