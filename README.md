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
8. 新增sesson认证机制
9. 新增角色：admin、普通用户
10. 统一日志打印和错误码



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
├── cmd/                             # 应用程序入口
│   └── server/
│       └── main.go                  # 主入口文件，应用启动和初始化
├── internal/                        # 私有应用程序代码
│   ├── code/                        # 错误码定义和多语言错误消息
│   │   ├── code.go                  # 错误码常量定义
│   │   ├── zh-cn.go                 # 中文错误消息
│   │   └── en-us.go                 # 英文错误消息
│   ├── common/                      # 通用工具和常量
│   │   ├── constant.go              # 全局常量定义
│   │   ├── context.go               # 上下文管理工具
│   │   └── error.go                 # 错误处理工具
│   ├── config/                      # 配置管理
│   │   └── config.go                # 配置加载和解析
│   ├── controller/                  # HTTP 控制器层（处理请求和响应）
│   │   ├── health_controller.go     # 健康检查控制器
│   │   ├── user_controller.go       # 用户管理控制器
│   │   └── order_controller.go      # 订单管理控制器
│   ├── dto/                         # 数据传输对象（Data Transfer Objects）
│   │   ├── user_dto.go              # 用户相关DTO
│   │   └── order_dto.go             # 订单相关DTO
│   ├── interceptor/                 # 拦截器（GRPC/中间件）
│   │   ├── interceptor.go           # 通用拦截器
│   │   └── session_auth.go          # Session认证拦截器
│   ├── middleware/                  # Gin中间件
│   │   ├── cors.go                  # 跨域中间件
│   │   ├── logger.go                # 日志中间件
│   │   ├── rate_limit.go            # 限流中间件
│   │   └── recovery.go              # 异常恢复中间件
│   ├── model/                       # 数据模型定义
│   │   ├── user.go                  # 用户数据模型
│   │   └── order.go                 # 订单数据模型
│   ├── redis/                       # Redis业务逻辑层
│   │   └── redis_repository.go      # Redis数据访问实现
│   ├── repository/                  # 数据访问层（数据库操作）
│   │   ├── base_repository.go       # 基础仓储接口
│   │   ├── user_repository.go       # 用户数据访问
│   │   └── order_repository.go      # 订单数据访问
│   ├── router/                      # 路由配置
│   │   └── router.go                # 路由注册和中间件配置
│   ├── service/                     # 业务逻辑层
│   │   ├── user_service.go          # 用户业务逻辑
│   │   └── order_service.go         # 订单业务逻辑
│   └── validation/                  # 数据验证
│       └── validation.go            # 验证器实现
├── pkg/                             # 公共库代码（可被外部项目引用）
│   ├── color/                       # 终端颜色输出工具
│   │   └── string_*.go              # 平台相关的字符串颜色处理
│   ├── database/                    # 数据库连接管理
│   │   ├── postgres.go              # PostgreSQL连接初始化
│   │   ├── redis.go                 # Redis连接初始化
│   │   └── sql_plugin.go            # SQL插件支持
│   ├── errors/                      # 统一错误处理
│   │   ├── err.go                   # 业务错误定义和工具
│   │   └── err_test.go              # 错误处理单元测试
│   ├── logger/                      # 日志处理
│   │   └── logger.go                # Zap日志封装
│   ├── response/                    # 统一响应格式
│   │   └── response.go              # HTTP响应封装
│   ├── timeutil/                    # 时间工具
│   │   ├── timeutil.go              # 时间处理工具函数
│   │   └── timeutil_test.go         # 时间工具单元测试
│   ├── trace/                       # 链路追踪工具
│   │   ├── trace.go                 # 追踪功能实现
│   │   ├── debug.go                 # 调试工具
│   │   ├── dialog.go                # 对话工具
│   │   ├── sql.go                   # SQL追踪
│   │   └── redis.go                 # Redis追踪
│   └── utils/                       # 通用工具函数
│       ├── crypto.go                # 加密解密工具
│       └── utils.go                 # 杂项工具函数
├── configs/                         # 配置文件目录
│   ├── config.local.yaml            # 本地开发环境配置
│   ├── config.dev.yaml              # 开发环境配置
│   └── config.prod.yaml             # 生产环境配置
├── docs/                            # 项目文档
│   ├── PROJECT_GUIDE.md             # 项目使用指南
│   ├── API_REFERENCE.md             # API接口文档
│   ├── ARCHITECTURE.md              # 架构设计文档
│   ├── docs.go                      # Swagger文档生成
│   ├── swagger.json                 # Swagger JSON规范
│   └── swagger.yaml                 # Swagger YAML规范
├── .gitignore                       # Git忽略文件配置
├── .vscode/                         # VSCode配置
├── docker-compose.yml               # Docker编排配置
├── Dockerfile                       # Docker镜像构建文件
├── go.mod                           # Go模块依赖定义
├── go.sum                           # Go模块依赖校验和
├── Makefile                         # Make构建脚本
└── README.md                        # 项目说明文档
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
  "username": "Tim",
  "email": "Tim@example.com",
  "phone": "13800178333",
  "password": "123456"
}
```

**response：**
- 成功响应：
```json
{
    "id": 8,
    "created_at": "2026-01-08T11:28:49.432732891+08:00",
    "update_at": "2026-01-08T11:28:49.432732941+08:00",
    "username": "Tim",
    "email": "Tim@example.com",
    "phone": "13800178333",
    "avatar": "",
    "status": 1
}
```
- 错误响应：
```json
{
    "code": 20201,
    "message": "创建管理员失败"
}
```

#### 用户登录
**request：**
```bash
POST /api/v1/users/login
Content-Type: application/json

{
    "username": "Tim",
    "password": "123456"
}
```

**response：**
- 成功响应：
```json
{
    "avatar": "http://127.0.0.1:9060/api/v1/gin-app-start/file/",
    "email": "Tim@example.com",
    "phone": "13800178333",
    "userId": 8,
    "username": "Tim"
}
```
- 错误响应：
```json
{
    "code": 20206,
    "message": "登录失败"
}
```

#### 查询用户
**request：**
```bash
GET /api/v1/users/:id
```

**response：**
- 成功响应：
```json
{
    "id": 8,
    "created_at": "2026-01-08T11:28:49.432732+08:00",
    "update_at": "2026-01-08T11:28:49.432732+08:00",
    "username": "Tim",
    "email": "Tim@example.com",
    "phone": "13800178333",
    "avatar": "",
    "status": 1
}
```
- 错误响应：
```json
{"code":10104,"message":"签名信息错误"}
```


#### 更新用户
**request：**
```bash
PUT /api/v1/users/:id
Content-Type: application/json

{
  "email": "Tim@example.com",
  "phone": "13800178334"
}
```

**response：**
- 成功响应：
```json
{
    "id": 8,
    "created_at": "2026-01-08T11:28:49.432732+08:00",
    "update_at": "2026-01-08T14:38:36.347845976+08:00",
    "username": "Tim",
    "email": "Tim@example.com",
    "phone": "13800178334",
    "avatar": "",
    "status": 1
}
```
- 错误响应：
```json
{"code":10104,"message":"签名信息错误"}
```

#### 更改密码
**request：**
```bash
POST /api/v1/users/change_pwd
Content-Type: application/json

{
    "username": "Tim",
    "old_password": "123456",
    "new_password": "1234567"
}
```

**response：**
- 成功响应：
```json
"Change password success"
```

- 错误响应：
```json
{"code":10104,"message":"签名信息错误"}
```

#### 上传头像
**request：**
```bash
POST /api/v1/users/upload_avatar
Content-Type: multipart/form-data

{
  "username": "user2",
  "file": (binary file)
}
```

**response：**
- 成功响应：
```json
"http://127.0.0.1:9060/api/v1/gin-app-start/file/63dedf56-bf03-4976-a202-4a049fd76cbe.png"
```

- 错误响应：
```json
{"code":10103,"message":"参数信息错误"}
```

#### 获取头像
```bash
GET /api/v1/users/file?username=Tim&imageName=63dedf56-bf03-4976-a202-4a049fd76cbe.png
```

**response：**
- 成功响应：
```json
展示用户头像
```

- 错误响应：
```json
{"code":10104,"message":"签名信息错误"}
```

#### 删除用户

**response：**
- 成功响应：
```json
"Deleted successfully"
```
- 错误响应：
```json
{"code":10104,"message":"签名信息错误"}
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
    "users": [
        {
            "id": 1,
            "created_at": "2025-11-26T22:01:26.823447+08:00",
            "update_at": "2025-12-03T11:58:43.290084+08:00",
            "username": "testuser15",
            "email": "newemail2@example.com",
            "phone": "13900139020",
            "avatar": "",
            "status": 1
        },
        {
            "id": 3,
            "created_at": "2025-12-05T16:13:49.914463+08:00",
            "update_at": "2025-12-05T16:13:49.914463+08:00",
            "username": "admin",
            "email": "admin6@example.com",
            "phone": "13800138022",
            "avatar": "",
            "status": 1
        },
        {
            "id": 4,
            "created_at": "2025-12-05T16:53:10.935254+08:00",
            "update_at": "2025-12-06T14:48:16.517959+08:00",
            "username": "user2",
            "email": "user4@example.com",
            "phone": "13900139024",
            "avatar": "6f36618d-75af-46a0-9462-7fa631919b97.png",
            "status": 1
        },
        {
            "id": 5,
            "created_at": "2025-12-06T15:00:58.802902+08:00",
            "update_at": "2025-12-06T15:01:26.206152+08:00",
            "username": "Jone Bob",
            "email": "user3@example.com",
            "phone": "13800138322",
            "avatar": "018e3025-dd71-4d14-9cc5-e08751a52d8f.png",
            "status": 1
        },
        {
            "id": 7,
            "created_at": "2025-12-06T15:28:13.037356+08:00",
            "update_at": "2025-12-06T15:28:13.037356+08:00",
            "username": "Bob",
            "email": "Bob@example.com",
            "phone": "13800178320",
            "avatar": "",
            "status": 1
        },
        {
            "id": 8,
            "created_at": "2026-01-08T11:28:49.432732+08:00",
            "update_at": "2026-01-08T14:53:04.488063+08:00",
            "username": "Tim",
            "email": "Tim@example.com",
            "phone": "13800178334",
            "avatar": "c4808961-71b5-45dd-ba94-84ec4ba6a36c.png",
            "status": 1
        }
    ],
    "total": 6,
    "page": 1,
    "page_size": 10
}
```
- 错误响应：
```json
{"code":10104,"message":"签名信息错误"}
```

#### 退出登录
**request：**
```bash
POST /api/v1/users/logout
Content-Type: application/json

{
    "username": "admin"
}
```

**response：**
- 成功响应：
```json
"Logout successfully"
```
- 错误响应：
```json
{"code":10104,"message":"签名信息错误"}
```

### 订单管理

#### 创建订单
**request：**
```bash
POST /api/v1/orders
Content-Type: application/json

{
  "username": "Bob",
  "total_price": 200.00,
  "description": "Good product"
}
```
**response：**
- 成功响应：
```json
{
    "id": 7,
    "order_number": "EC20260108124133",
    "created_at": "2026-01-08T14:58:00.803234908+08:00",
    "update_at": "2026-01-08T14:58:00.803234958+08:00",
    "user_id": 3,
    "username": "user2",
    "total_price": 50,
    "description": "Good quality",
    "status": 1
}
```
- 错误响应：
```json
{"code":10104,"message":"签名信息错误"}
```

#### 获取订单
**request：**
```bash
GET /api/v1/orders/search?order_number=EC20251202659066&username=Bob
```
**response：**
- 成功响应：
```json
{
    "id": 1,
    "order_number": "EC20251206344246",
    "created_at": "2025-12-06T15:40:04.018367+08:00",
    "update_at": "2025-12-06T15:44:10.473489+08:00",
    "user_id": 7,
    "username": "Bob",
    "total_price": 40,
    "description": "Bad product!!!",
    "status": 1
}
```
- 错误响应：
```json
{"code":10104,"message":"签名信息错误"}
```

#### 更新订单
**request：**
```bash
PUT /api/v1/orders/:order_number
Content-Type: application/json

{
  "username": "Bob",
  "order_number": "EC20251206344246",
  "total_price": 44,
  "description": "Bad product!!!",
  "status": 0
}
```

**response：**
- 成功响应：
```json
{
    "id": 1,
    "order_number": "EC20251206344246",
    "created_at": "2025-12-06T15:40:04.018367+08:00",
    "update_at": "2026-01-08T15:12:34.054476758+08:00",
    "user_id": 7,
    "username": "Bob",
    "total_price": 44,
    "description": "Bad product!!!",
    "status": 1
}
```
- 错误响应：
```json
{"code":20503,"message":"更新订单失败"}
```

#### 删除订单
**request：**
```bash
DELETE /api/v1/orders
Content-Type: application/json

{
  "order_number": "EC20251206344246"
  "username": "Bob"
}
```
**response：**
- 成功响应：
```json
{
    "username": "Bob",
    "order_number": "EC20251206344246"
}
```
- 错误响应：
```json
{"code":20504,"message":"删除订单失败"}
```

#### 订单列表
**request：**
```bash
GET /api/v1/orders?username=Bob
```
**response：**
- 成功响应：
```json
{
    "orders": [
        {
            "id": 3,
            "order_number": "EC20251206794733",
            "created_at": "2025-12-06T15:45:08.447049+08:00",
            "update_at": "2025-12-06T15:45:08.447049+08:00",
            "user_id": 7,
            "username": "Bob",
            "total_price": 65,
            "description": "Good",
            "status": 1
        },
        {
            "id": 4,
            "order_number": "EC20251206169258",
            "created_at": "2025-12-06T15:45:17.993545+08:00",
            "update_at": "2025-12-06T15:45:17.993545+08:00",
            "user_id": 7,
            "username": "Bob",
            "total_price": 99,
            "description": "Very Good!!!",
            "status": 1
        }
    ],
    "total": 2
}
```
- 错误响应：
```json
{"code":10104,"message":"签名信息错误"}
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

### 语言配置
```yaml
language:
  local: zh-CN  # 错误信息的显示语言，可选项：zh-CN、en-US
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

### 日志配置
```yaml
log:
  level: info # 日志级别，可选值：debug, info, warn, error, panic, fatal
  file_path: /var/log/gin-app/app.log # 日志文件路径
  max_size: 100 # 最大日志文件大小为100M
  max_age: 30   # 最大日志文件保存时间为30天
```

### 文件上传配置
```yaml
file:
  dir_name: 'public/file/' # 文件上传目录
  url_prefix: 'http://127.0.0.1:9060/api/v1/gin-app-start/file/' # 文件上传URL前缀
  max_size: 8388608 # 最大文件上传大小为8M
```

### 会话配置
```yaml
session:
  use_redis: true   # 是否使用Redis存储会话, 默认为false
  name: 'mysession' # 会话名称
  size: 10          # 会话大小, 默认为10
  key: gin-session  # 会话键名
  max_age: 120      # 会话过期时间, 默认为120秒
  path: /           # 会话路径, 默认为"/"
  domain: ""        # 会话域名, 默认为""
  http_only: true   # 是否仅通过HTTP访问会话, 默认为true
  secure: false     # 是否仅通过HTTPS访问会话, 默认为false
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
