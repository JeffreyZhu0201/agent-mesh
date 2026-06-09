# AgentMesh 开发指南

欢迎使用 AgentMesh！本指南将帮助您快速上手并深入了解这个多租户 SaaS 后台开发框架。

---

## 1. 项目概述

### 什么是 AgentMesh

AgentMesh 是一个现代化的多租户 SaaS 后台开发框架，采用 Go 微服务架构 + React 前端，帮助开发者快速构建企业级 AI 应用平台。

**核心能力：**
- **用户认证**：JWT-based 认证，支持登录/注册/登出
- **RBAC 权限体系**：基于角色的访问控制，四种角色灵活权限管理
- **多租户支持**：共享表方案，tenant_id 隔离，数据架构简单高效
- **插件系统**：前端插件热插拔，后端插件可启用/禁用

### 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go + Gin + go-zero |
| 前端 | React 18 + TypeScript + Vite + MUI |
| AI | Eino (LLM 编排) |
| 数据库 | MySQL 8.0 |
| 认证 | JWT |
| 状态管理 | Zustand |

### 与其他框架的区别

AgentMesh 参考了若伊（RuoYi）等成熟框架的设计思想，但在以下方面进行了现代化改造：

- **微服务架构**：解耦为独立服务（api-gateway、user-svc、plugin-svc），而非单体
- **插件系统**：前端插件自注册机制，后端插件可动态启用/禁用
- **前后端分离**：标准 REST API，支持多端接入
- **现代化技术栈**：Go 1.24+ / React 18 / TypeScript / Vite

---

## 2. 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              浏览器 (Browser)                                │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           API Gateway (端口 8080)                           │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        JWT 验证 / 租户上下文注入                       │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                       │                                     │
│                    ┌─────────────────┼─────────────────┐                   │
│                    ▼                 ▼                 ▼                   │
│            ┌──────────────┐  ┌──────────────┐  ┌──────────────┐            │
│            │  user-svc    │  │  plugin-svc  │  │  agent-svc   │            │
│            │  (端口 8081) │  │  (端口 8083) │  │  (端口 8082) │            │
│            │              │  │              │  │   (规划中)   │            │
│            │  - 用户管理   │  │  - 插件管理   │  │              │            │
│            │  - 租户管理   │  │  - 启用/禁用  │  │  - AI 对话   │            │
│            │  - JWT 签发  │  │              │  │  - 流式响应  │            │
│            └──────────────┘  └──────────────┘  └──────────────┘            │
│                    │                 │                                     │
└────────────────────┼─────────────────┼─────────────────────────────────────┘
                     │                 │
                     ▼                 ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           MySQL (端口 3306)                                  │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  users / tenants / plugins / tenant_plugins                          │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 各服务职责

| 服务 | 端口 | 职责 |
|------|------|------|
| **api-gateway** | 8080 | HTTP 入口，JWT 验证，路由转发，租户上下文注入 |
| **user-svc** | 8081 | 用户认证，JWT 签发，RBAC，租户管理 |
| **plugin-svc** | 8083 | 插件注册表管理，租户级插件启用/禁用 |
| **agent-svc** | 8082 | AI 对话能力（规划中） |

### 前端目录结构

```
web/
├── apps/
│   ├── landing/          # 营销落地页
│   ├── main/            # 用户主应用
│   │   └── src/
│   │       ├── plugins/ # 插件目录（chat 等）
│   │       ├── pages/   # 页面组件
│   │       ├── api/     # API 客户端
│   │       └── App.tsx  # 应用入口
│   └── admin/           # 管理后台
│       └── src/
│           ├── pages/
│           │   ├── UserManagement.tsx
│           │   ├── TenantManagement.tsx
│           │   └── PluginManagement.tsx
│           └── layouts/
└── packages/
    ├── plugin-sdk/       # 插件 SDK（注册/路由/菜单）
    ├── stores/          # Zustand 状态管理
    ├── api/             # API 封装
    └── ui/              # 共享 MUI 组件
```

### 请求流程

```
1. 用户登录
   POST /api/user/login
   → user-svc 验证密码，签发 JWT
   → 返回 token

2. 后续请求携带 Token
   GET /api/user/info
   Header: Authorization: Bearer <token>

3. API Gateway 验证流程
   → 解析 JWT，验证签名
   → 注入 X-User-ID / X-Tenant-ID / X-User-Role 到请求头
   → 转发给对应微服务

4. 微服务处理
   → 从请求头获取租户上下文
   → 执行业务逻辑
   → 返回响应
```

---

## 3. 快速开始

### 启动步骤

**方式一：Docker 启动（推荐）**

```bash
# 克隆项目后，在项目根目录执行
make docker-up

# 查看日志
docker-compose logs -f

# 停止服务
make docker-down
```

**方式二：本地开发**

```bash
# 后端服务
cd services/user-svc && go run .
cd services/plugin-svc && go run .
cd services/api-gateway && go run .

# 前端开发
cd web
npm install
npm run dev
```

### 访问地址

| 应用 | 地址 | 说明 |
|------|------|------|
| 用户主应用 | http://localhost:3000 | 普通用户使用 |
| 管理后台 | http://localhost:3001 | 管理员使用 |

### 默认账户

| 角色 | 用户名 | 密码 | 说明 |
|------|--------|------|------|
| 平台超级管理员 | admin | admin123 | tenant_id=0，可管理所有租户和用户 |

### 验证服务运行

```bash
# 健康检查
curl http://localhost:8080/health

# 测试登录
curl -X POST http://localhost:8081/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 响应示例
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "userId": 1,
      "username": "admin",
      "tenantId": 0,
      "role": "platform_admin"
    }
  }
}
```

---

## 4. 核心概念

### 4.1 多租户设计

AgentMesh 采用**共享表方案**实现多租户隔离。

#### 数据结构

所有租户共用同一套表，通过 `tenant_id` 字段区分：

```sql
-- tenants 表（租户注册表）
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,      -- 租户 ID，如 1, 2, 3...
    name VARCHAR(255),           -- 租户名称
    code VARCHAR(63) UNIQUE,     -- 租户代码（URL 友好）
    status VARCHAR(20),         -- active/suspended/deleted
    created_at TIMESTAMP
);

-- users 表（用户表，所有租户共享）
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255),
    email VARCHAR(255),
    tenant_id INTEGER REFERENCES tenants(id),  -- 租户 ID
    role VARCHAR(50),        -- platform_admin/admin/user/viewer
    status VARCHAR(20),
    created_at TIMESTAMP
);
```

#### tenant_id 语义

| tenant_id 值 | 含义 | 说明 |
|--------------|------|------|
| 0 | 平台超级管理员 | 不属于任何租户，拥有最高权限 |
| 1, 2, 3... | 普通租户 | 租户内的用户从此 ID 开始 |

#### 为什么用共享表？

| 方案 | 优点 | 缺点 |
|------|------|------|
| **共享表（AgentMesh）** | 运维简单，数据量小时性能好，跨租户查询方便 | 需要注意 SQL 注入防护，需手动过滤 tenant_id |
| Schema 隔离 | 数据隔离彻底 | 创建数据库成本高，跨租户查询复杂 |
| 行级隔离 | 灵活 | 实现复杂，容易遗漏 WHERE 条件 |

AgentMesh 选择共享表是因为：
1. **运维简单**：一个数据库实例，一个备份，一个迁移脚本
2. **成本低**：不需要为每个租户创建独立 Schema
3. **够用**：大多数 SaaS 应用数据量级不需要 Schema 隔离

### 4.2 RBAC 权限体系

AgentMesh 实现了一个简单的四角色 RBAC 系统。

#### 角色定义

| 角色 | 英文标识 | 说明 |
|------|----------|------|
| 平台管理员 | platform_admin | 系统级管理员，可管理所有租户和用户 |
| 租户管理员 | admin | 租户级管理员，可管理本租户内的用户 |
| 普通用户 | user | 租户内的普通成员 |
| 访客 | viewer | 注册新用户的默认角色，只有查看权限 |

#### 权限层级

```
platform_admin (tenant_id=0)
    │
    ├── 可管理所有租户 (CRUD)
    ├── 可管理所有用户 (CRUD)
    └── 可启用/禁用插件 (跨租户)
           │
           ▼
    admin (tenant_id>0)
    │
    ├── 可管理本租户用户 (CRUD)
    ├── 可启用/禁用本租户插件
    └── 不可跨租户操作
           │
           ▼
    user / viewer
    │
    └── 只有查看权限
```

#### 代码中的权限检查

```go
// user-svc.go 中的权限中间件

// 平台管理员检查
func requirePlatformAdmin(c *gin.Context) bool {
    role, _ := c.Get("role")
    return role == "platform_admin"
}

// 租户管理员检查
func requireTenantAdmin(c *gin.Context) bool {
    role, _ := c.Get("role")
    return role == "admin"
}
```

### 4.3 JWT 认证流程

#### 完整流程

```
┌────────┐                    ┌──────────────┐                    ┌──────────┐
│ 用户   │                    │  API Gateway │                    │ user-svc  │
└───┬────┘                    └──────┬───────┘                    └────┬─────┘
    │                                │                                 │
    │  1. POST /api/user/login       │                                 │
    │  {username, password}          │                                 │
    │ ─────────────────────────────► │                                 │
    │                                │  2. 验证用户名密码                │
    │                                │ ─────────────────────────────► │
    │                                │                                 │
    │                                │  3. 返回用户信息                  │
    │                                │ ◄────────────────────────────── │
    │                                │                                 │
    │  4. 返回 JWT Token              │                                 │
    │ ◄───────────────────────────── │                                 │
    │                                │                                 │
    │  5. 后续请求带 Token            │                                 │
    │  Authorization: Bearer xxx     │                                 │
    │ ─────────────────────────────► │                                 │
    │                                │  6. 验证 JWT，注入上下文          │
    │                                │  7. 转发请求 + 租户头             │
    │                                │ ─────────────────────────────► │
    │                                │                                 │
    │  8. 返回业务数据               │                                 │
    │ ◄───────────────────────────── │                                 │
```

#### JWT Payload 结构

```json
{
  "userId": 1,
  "username": "admin",
  "tenantId": 0,
  "role": "platform_admin",
  "exp": 1750000000,
  "iat": 1749913600
}
```

#### API Gateway 的作用

API Gateway 是所有请求的入口，负责：

1. **JWT 验证**：验证 Token 签名，检查是否过期
2. **上下文注入**：将用户信息注入到请求头中
   - `X-User-ID`: 用户 ID
   - `X-Tenant-ID`: 租户 ID
   - `X-User-Role`: 用户角色
3. **路由转发**：将请求转发到对应的微服务

```go
// 中间件代码示例
c.Set("user_id", claims.UserID)
c.Set("tenant_id", claims.TenantID)
c.Set("username", claims.Username)
c.Set("role", claims.Role)
```

#### 各微服务的行为

微服务**不再重复验证 JWT**，而是从请求头中获取已验证的上下文：

```go
// plugin-svc.go 中的做法
func authMiddleware(c *gin.Context) {
    // 直接信任 api-gateway 注入的 header
    tenantID := c.GetHeader("X-Tenant-ID")
    // ...处理请求
}
```

### 4.4 插件系统

#### 前端插件注册机制

前端插件通过 `registerPlugin()` 自注册，导入即生效：

```typescript
// web/apps/main/src/plugins/chat/index.ts

import { registerPlugin } from '@agentmesh/plugin-sdk';
import ChatPage from './ChatPage';

const chatPlugin = {
  metadata: {
    name: 'chat',
    version: '0.1.0',
    description: 'AI 对话插件',
  },
  menuItems: [
    {
      key: 'chat',
      label: 'Chat',
      icon: 'C',
      path: '/app/chat',
      order: 10,
    },
  ],
  routes: [
    {
      path: '/app/chat',
      component: ChatPage,
      exact: true,
    },
  ],
};

// 导入即自注册
registerPlugin(chatPlugin);
```

#### 插件注册中心

所有插件在 `plugins/index.ts` 中统一导入：

```typescript
// web/apps/main/src/plugins/index.ts
// 导入即注册
import './chat';  // Chat 插件已注册
```

#### 后端插件管理

后端插件通过 `tenant_plugins` 表控制启用状态：

```sql
-- plugins 表：所有可用插件
CREATE TABLE plugins (
    id SERIAL PRIMARY KEY,
    plugin_key VARCHAR(64) UNIQUE,  -- 插件唯一标识
    name VARCHAR(255),
    version VARCHAR(32),
    builtin BOOLEAN                  -- 是否内置（不可删除）
);

-- tenant_plugins 表：租户级插件开关
CREATE TABLE tenant_plugins (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER,
    plugin_key VARCHAR(64),
    enabled BOOLEAN DEFAULT TRUE,
    UNIQUE(tenant_id, plugin_key)
);
```

#### 启用/禁用逻辑

| tenant_plugins 表记录 | 插件状态 |
|------------------------|----------|
| 无记录 | 默认启用 |
| enabled=true | 启用 |
| enabled=false | 禁用 |

租户管理员在 Admin 后台可以开启/关闭本租户的插件。

#### 插件工作流程

```
1. 平台管理员在 Admin 后台启用插件
   POST /api/plugin/admin/toggle
   { "pluginKey": "chat", "enabled": true }

2. 前端获取已启用插件列表
   GET /api/plugin/list
   → 返回 ["chat", "dashboard"]

3. 前端根据列表渲染菜单
   → Chat 和 Dashboard 菜单显示

4. 用户点击菜单，访问插件页面
   GET /app/chat
```

---

## 5. 数据库设计

### users 表（用户表）

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,              -- 用户 ID
    username VARCHAR(100) NOT NULL,     -- 用户名（唯一）
    password_hash VARCHAR(255) NOT NULL,-- 密码（bcrypt 加密）
    email VARCHAR(255),                  -- 邮箱（唯一）
    tenant_id INTEGER DEFAULT 0,       -- 租户 ID（0=平台管理员）
    role VARCHAR(50) DEFAULT 'viewer',  -- 角色
    status INTEGER DEFAULT 1,           -- 状态（1=启用，0=禁用）
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- 索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_role ON users(role);
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键，自增 |
| username | VARCHAR(100) | 用户名，唯一 |
| password_hash | VARCHAR(255) | bcrypt 加密后的密码 |
| email | VARCHAR(255) | 邮箱地址 |
| tenant_id | INTEGER | 租户 ID，0=平台管理员 |
| role | VARCHAR(50) | 角色：platform_admin/admin/user/viewer |
| status | INTEGER | 1=启用，0=禁用 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### tenants 表（租户表）

```sql
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,              -- 租户 ID
    name VARCHAR(255) NOT NULL,         -- 租户名称
    code VARCHAR(63) UNIQUE,           -- 租户代码（URL 友好）
    status INTEGER DEFAULT 1,           -- 状态（1=启用，0=禁用）
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE INDEX idx_tenants_code ON tenants(code);
CREATE INDEX idx_tenants_status ON tenants(status);
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键，租户唯一标识 |
| name | VARCHAR(255) | 租户显示名称 |
| code | VARCHAR(63) | 租户代码，唯一，用于 URL |
| status | INTEGER | 1=启用，0=禁用 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### plugins 表（插件注册表）

```sql
CREATE TABLE plugins (
    id SERIAL PRIMARY KEY,
    plugin_key VARCHAR(64) UNIQUE,      -- 插件唯一标识
    name VARCHAR(255) NOT NULL,         -- 插件显示名称
    version VARCHAR(32) NOT NULL,       -- 版本号
    type VARCHAR(32),                   -- 插件类型
    description TEXT,                   -- 插件描述
    author VARCHAR(255),                 -- 作者
    builtin BOOLEAN DEFAULT TRUE,       -- 是否内置插件
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE INDEX idx_plugins_key ON plugins(plugin_key);
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| plugin_key | VARCHAR(64) | 插件唯一标识，如 "chat"、"dashboard" |
| name | VARCHAR(255) | 插件显示名称 |
| version | VARCHAR(32) | 版本号 |
| type | VARCHAR(32) | 类型：如 "page"、"widget" |
| description | TEXT | 描述 |
| author | VARCHAR(255) | 作者 |
| builtin | BOOLEAN | 是否内置（内置不可删除） |

### tenant_plugins 表（租户插件关联表）

```sql
CREATE TABLE tenant_plugins (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,         -- 租户 ID
    plugin_key VARCHAR(64) NOT NULL,     -- 插件标识
    enabled BOOLEAN DEFAULT TRUE,      -- 是否启用
    updated_at TIMESTAMP,
    UNIQUE(tenant_id, plugin_key)
);

CREATE INDEX idx_tenant_plugins_tenant ON tenant_plugins(tenant_id);
CREATE INDEX idx_tenant_plugins_plugin ON tenant_plugins(plugin_key);
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| tenant_id | INTEGER | 租户 ID |
| plugin_key | VARCHAR(64) | 插件唯一标识 |
| enabled | BOOLEAN | 是否启用（无记录=默认启用） |

---

## 6. API 参考

### 响应格式规范

所有 API 响应统一格式：

```json
// 成功响应
{
  "success": true,
  "data": { ... }
}

// 成功消息（无 data）
{
  "success": true,
  "message": "操作成功"
}

// 错误响应
{
  "success": false,
  "error": "错误信息"
}
```

### 公开接口

#### POST /api/user/register - 用户注册

| 项目 | 说明 |
|------|------|
| 认证 | 不需要 |
| 方法 | POST |
| Content-Type | application/json |

**请求体：**

```json
{
  "username": "testuser",
  "password": "123456",
  "email": "test@example.com"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，唯一 |
| password | string | 是 | 密码，最少 6 字符 |
| email | string | 否 | 邮箱 |

**响应示例：**

```json
{
  "success": true,
  "data": {
    "id": 2,
    "username": "testuser",
    "email": "test@example.com",
    "tenantId": 1,
    "role": "viewer",
    "status": 1,
    "createdAt": "2026-06-09T10:00:00Z"
  }
}
```

---

#### POST /api/user/login - 用户登录

| 项目 | 说明 |
|------|------|
| 认证 | 不需要 |
| 方法 | POST |
| Content-Type | application/json |

**请求体：**

```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "userId": 1,
      "username": "admin",
      "tenantId": 0,
      "role": "platform_admin"
    }
  }
}
```

---

### 登录用户接口

#### GET /api/user/info - 获取当前用户信息

| 项目 | 说明 |
|------|------|
| 认证 | 需要（Bearer Token） |
| 方法 | GET |
| 权限 | 任意登录用户 |

**请求示例：**

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8081/api/user/info
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@agentmesh.io",
    "tenantId": 0,
    "role": "platform_admin",
    "status": 1,
    "createdAt": "2026-06-01T00:00:00Z"
  }
}
```

---

### 平台管理员接口

> 注意：以下接口需要 `role=platform_admin`

#### GET /api/admin/tenants - 获取所有租户列表

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | GET |
| 权限 | platform_admin |

**请求示例：**

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8081/api/admin/tenants
```

**响应示例：**

```json
{
  "success": true,
  "data": [
    { "id": 1, "name": "Default Tenant", "code": "default", "status": 1 },
    { "id": 2, "name": "ACME Corp", "code": "acme", "status": 1 }
  ]
}
```

---

#### POST /api/admin/tenants - 创建租户

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | POST |
| 权限 | platform_admin |

**请求体：**

```json
{
  "name": "ACME Corporation",
  "code": "acme"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 租户名称 |
| code | string | 是 | 租户代码，唯一 |

**响应示例：**

```json
{
  "success": true,
  "data": {
    "id": 2,
    "name": "ACME Corporation",
    "code": "acme",
    "status": 1
  }
}
```

---

#### PUT /api/admin/tenants/:id - 更新租户

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | PUT |
| 权限 | platform_admin |

**请求体：**

```json
{
  "name": "ACME Corp Updated",
  "status": 1
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 租户名称 |
| code | string | 否 | 租户代码 |
| status | integer | 否 | 1=启用，0=禁用 |

---

#### DELETE /api/admin/tenants/:id - 删除租户

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | DELETE |
| 权限 | platform_admin |
| 注意 | id=1（默认租户）不可删除 |

**请求示例：**

```bash
curl -X DELETE -H "Authorization: Bearer <token>" http://localhost:8081/api/admin/tenants/2
```

**响应示例：**

```json
{
  "success": true,
  "message": "tenant deleted"
}
```

---

#### GET /api/admin/users - 获取所有用户列表

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | GET |
| 权限 | platform_admin |
| 查询参数 | tenant_id（可选） |

**请求示例：**

```bash
# 所有用户
curl -H "Authorization: Bearer <token>" http://localhost:8081/api/admin/users

# 租户 1 的用户
curl -H "Authorization: Bearer <token>" "http://localhost:8081/api/admin/users?tenant_id=1"
```

---

#### POST /api/admin/users - 创建用户（跨租户）

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | POST |
| 权限 | platform_admin |

**请求体：**

```json
{
  "username": "john",
  "password": "secret",
  "email": "john@example.com",
  "tenantId": 2,
  "role": "viewer"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |
| email | string | 否 | 邮箱 |
| tenantId | integer | 否 | 租户 ID，默认 1 |
| role | string | 否 | 角色，默认 viewer |

---

#### PUT /api/admin/users/:id - 更新任意用户

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | PUT |
| 权限 | platform_admin |

**请求体：**

```json
{
  "email": "newemail@example.com",
  "role": "admin",
  "status": 1
}
```

---

#### DELETE /api/admin/users/:id - 删除任意用户

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | DELETE |
| 权限 | platform_admin |

---

### 租户管理员接口

> 注意：以下接口需要 `role=admin`，且只能操作本租户数据

#### GET /api/tenant/users - 获取本租户用户列表

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | GET |
| 权限 | admin（自动过滤本租户） |

**请求示例：**

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8081/api/tenant/users
```

---

#### POST /api/tenant/users - 在本租户创建用户

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | POST |
| 权限 | admin |
| 注意 | 新用户自动归属当前租户，角色默认 user |

**请求体：**

```json
{
  "username": "employee1",
  "password": "welcome123",
  "email": "emp@company.com"
}
```

---

#### PUT /api/tenant/users/:id - 更新本租户用户

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | PUT |
| 权限 | admin |
| 安全限制 | 不能修改 admin 角色用户，不能提权 |

---

#### DELETE /api/tenant/users/:id - 删除本租户用户

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | DELETE |
| 权限 | admin |
| 安全限制 | 不能删除 admin 角色用户 |

---

### 插件接口

#### GET /api/plugin/list - 获取当前租户已启用的插件

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | GET |
| 权限 | 任意登录用户 |

**请求示例：**

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8083/api/plugin/list
```

**响应示例：**

```json
{
  "plugins": [
    {
      "key": "chat",
      "name": "Chat",
      "version": "0.1.0",
      "type": "page",
      "description": "Conversational AI chat",
      "enabled": true
    }
  ]
}
```

---

#### GET /api/plugin/admin/list - 获取所有插件状态（平台管理员）

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | GET |
| 权限 | platform_admin |

**响应示例：**

```json
{
  "plugins": [
    { "key": "chat", "name": "Chat", "enabled": true },
    { "key": "dashboard", "name": "Dashboard", "enabled": false }
  ]
}
```

---

#### POST /api/plugin/admin/toggle - 启用/禁用插件

| 项目 | 说明 |
|------|------|
| 认证 | 需要 |
| 方法 | POST |
| 权限 | platform_admin |

**请求体：**

```json
{
  "pluginKey": "dashboard",
  "enabled": false
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pluginKey | string | 是 | 插件唯一标识 |
| enabled | boolean | 是 | true=启用，false=禁用 |

**响应示例：**

```json
{
  "message": "ok"
}
```

---

## 7. 新增业务插件开发指南

### 7.1 创建前端插件

#### 步骤 1：创建插件目录

```
web/apps/main/src/plugins/
└── my-plugin/           # 新建插件目录
    ├── index.ts         # 插件注册文件
    ├── MyPluginPage.tsx # 插件页面组件
    └── components/      # 可选：子组件目录
```

#### 步骤 2：编写插件代码

```typescript
// web/apps/main/src/plugins/my-plugin/index.ts

import { registerPlugin } from '@agentmesh/plugin-sdk';
import MyPluginPage from './MyPluginPage';

// 定义插件
const myPlugin = {
  metadata: {
    name: 'my-plugin',
    version: '0.1.0',
    description: '我的自定义插件',
    type: 'page',
    author: 'Developer',
  },
  menuItems: [
    {
      key: 'my-plugin',       // 唯一标识
      label: '我的插件',       // 菜单显示名称
      icon: 'M',              // 图标（可以是 MUI 图标名）
      path: '/app/my-plugin', // 路由路径
      order: 20,              // 菜单排序（数字越小越靠前）
    },
  ],
  routes: [
    {
      path: '/app/my-plugin',
      component: MyPluginPage,
      exact: true,
    },
  ],
  component: MyPluginPage,
};

// 自注册
registerPlugin(myPlugin);

export default myPlugin;
```

```tsx
// web/apps/main/src/plugins/my-plugin/MyPluginPage.tsx

import React from 'react';
import { Box, Typography } from '@mui/material';

const MyPluginPage: React.FC = () => {
  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4">我的插件页面</Typography>
      <Typography sx={{ mt: 2 }}>
        这里可以编写任意业务逻辑
      </Typography>
    </Box>
  );
};

export default MyPluginPage;
```

#### 步骤 3：在注册中心导入

```typescript
// web/apps/main/src/plugins/index.ts

// 导入即注册
import './chat';        // 已有插件
import './my-plugin';   // 新增插件（添加这一行）
```

#### 步骤 4：在后端注册插件（可选）

如果插件需要后端数据交互，需要在 plugin-svc 中注册：

```go
// services/plugin-svc/plugin-svc.go 的 seedPlugins() 中添加

{
    PluginKey:   "my-plugin",
    Name:        "My Plugin",
    Version:     "0.1.0",
    Type:        "page",
    Description: "我的自定义插件",
    Author:      "Developer",
    Builtin:     true,
},
```

### 7.2 创建后端服务

如果新插件需要独立的后端服务（如独立的 API），需要创建新的微服务：

#### 步骤 1：创建服务目录

```
services/
└── my-plugin-svc/
    ├── my-plugin-svc.go    # 主入口
    ├── handler.go          # 请求处理
    ├── model/              # 数据模型
    │   └── mydata.go
    └── Dockerfile          # 构建文件
```

#### 步骤 2：编写服务代码

参考 `user-svc` 的模式：

```go
// services/my-plugin-svc/my-plugin-svc.go

package main

import (
    "fmt"
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // 公开接口
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // 需要认证的接口
    auth := r.Group("/api/my-plugin")
    auth.Use(authMiddleware())  // 复用或编写认证中间件
    {
        auth.GET("/data", handleGetData)
        auth.POST("/data", handleCreateData)
    }

    fmt.Println("My Plugin Service listening on :8084")
    r.Run(":8084")
}

// 认证中间件（信任 api-gateway 注入的 header）
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从 api-gateway 注入的 header 获取上下文
        tenantID := c.GetHeader("X-Tenant-ID")
        userID := c.GetHeader("X-User-ID")
        role := c.GetHeader("X-User-Role")

        if tenantID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }

        c.Set("tenantId", tenantID)
        c.Set("userId", userID)
        c.Set("role", role)
        c.Next()
    }
}
```

#### 步骤 3：在 docker-compose.yml 中添加服务

```yaml
services:
  # ... 现有服务 ...

  my-plugin-svc:
    build:
      context: ./services/my-plugin-svc
      dockerfile: Dockerfile
    container_name: agentmesh-my-plugin-svc
    ports:
      - "8084:8084"
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_USER=root
      - DB_PASSWORD=root123
      - DB_NAME=agentmesh
    depends_on:
      mysql:
        condition: service_healthy
    networks:
      - agentmesh-network
```

#### 步骤 4：在 API Gateway 添加路由

```go
// services/api-gateway/routes/routes.go

func RegisterRoutes(r *gin.Engine, authMW *middleware.AuthMiddleware, tenantMW *middleware.TenantMiddleware) {
    // ... 现有路由 ...

    // 新增插件服务路由
    myPlugin := r.Group("/api/my-plugin")
    myPlugin.Use(authMW.Handler())
    myPlugin.Use(tenantMW.Handler())
    {
        myPlugin.GET("/data", proxyToMyPluginSvc)
        myPlugin.POST("/data", proxyToMyPluginSvc)
    }
}
```

### 7.3 插件权限控制

AgentMesh 的插件启用/禁用完全由 `tenant_plugins` 表控制，不需要修改前端代码。

**启用插件：**

```bash
curl -X POST -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"pluginKey": "my-plugin", "enabled": true}' \
  http://localhost:8083/api/plugin/admin/toggle
```

**禁用插件：**

```bash
curl -X POST -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"pluginKey": "my-plugin", "enabled": false}' \
  http://localhost:8083/api/plugin/admin/toggle
```

**前端自动响应：**

- 插件启用后，`/api/plugin/list` 会返回该插件，前端自动渲染菜单
- 插件禁用后，`/api/plugin/list` 不再返回该插件，菜单自动消失

---

## 8. 目录结构

```
AgentMesh/
├── services/                    # 后端微服务
│   ├── api-gateway/            # API 网关
│   │   ├── api-gateway.go      # 主入口
│   │   ├── middleware/         # 中间件
│   │   │   ├── auth.go         # JWT 认证
│   │   │   └── tenant.go       # 租户上下文
│   │   └── routes/             # 路由定义
│   │       └── routes.go
│   │
│   ├── user-svc/               # 用户认证服务
│   │   ├── user-svc.go         # 主入口（1250 行，详细注释）
│   │   ├── model/              # 数据模型
│   │   │   ├── user.go         # User 模型
│   │   │   └── tenant.go       # Tenant 模型
│   │   └── Dockerfile
│   │
│   ├── plugin-svc/             # 插件管理服务
│   │   ├── plugin-svc.go       # 主入口
│   │   └── plugin/             # 插件接口定义
│   │       └── plugin.go
│   │
│   └── sql/                    # SQL 脚本
│       ├── init.sql            # 数据库初始化
│       └── tenant_schema.sql   # 租户 schema（预留）
│
├── web/                        # React 前端
│   ├── apps/
│   │   ├── landing/            # 营销落地页
│   │   │   └── src/pages/Home.tsx
│   │   │
│   │   ├── main/              # 用户主应用
│   │   │   └── src/
│   │   │       ├── App.tsx            # 应用入口
│   │   │       ├── main.tsx           # React 挂载
│   │   │       ├── plugins/           # 插件目录
│   │   │       │   ├── index.ts       # 插件注册中心
│   │   │       │   └── chat/          # Chat 插件示例
│   │   │       │       ├── index.ts
│   │   │       │       └── ChatPage.tsx
│   │   │       ├── pages/             # 页面组件
│   │   │       │   ├── Login.tsx
│   │   │       │   ├── Register.tsx
│   │   │       │   └── Dashboard.tsx
│   │   │       ├── api/               # API 客户端
│   │   │       └── components/         # 共享组件
│   │   │
│   │   └── admin/              # 管理后台
│   │       └── src/
│   │           ├── App.tsx
│   │           ├── layouts/AdminLayout.tsx
│   │           └── pages/
│   │               ├── UserManagement.tsx
│   │               ├── TenantManagement.tsx
│   │               └── PluginManagement.tsx
│   │
│   └── packages/                # 共享包
│       ├── plugin-sdk/         # 插件 SDK
│       │   └── src/
│       │       ├── index.ts    # 导出 registerPlugin 等
│       │       ├── types.ts    # 类型定义
│       │       └── hooks.ts    # React hooks
│       │
│       ├── stores/             # Zustand 状态
│       │   └── src/
│       │       ├── authStore.ts
│       │       └── pluginsStore.ts
│       │
│       ├── api/                # API 封装
│       │   └── src/auth.ts
│       │
│       └── ui/                 # MUI 组件库
│           └── src/components/
│               ├── Button/
│               ├── Layout/
│               └── ChatSidebar/
│
├── docs/                       # 文档
│   └── development-guide.md   # 本文档
│
├── docker-compose.yml          # Docker 编排
├── Makefile                   # 开发命令
├── CLAUDE.md                  # Claude Code 指南
└── README.md                  # 项目说明
```

### 各目录用途说明

| 目录 | 用途 |
|------|------|
| `services/api-gateway/` | HTTP 入口，JWT 验证，路由转发 |
| `services/user-svc/` | 用户认证，JWT 签发，RBAC，租户管理 |
| `services/plugin-svc/` | 插件注册表，租户级插件开关 |
| `web/apps/main/src/plugins/` | 前端插件目录 |
| `web/packages/plugin-sdk/` | 插件 SDK，提供 registerPlugin 等 API |
| `web/apps/admin/src/pages/` | 管理后台页面 |

---

## 9. 常见问题

### 如何添加新角色？

目前 AgentMesh 硬编码了四个角色（platform_admin/admin/user/viewer）。如果要添加新角色，需要修改以下位置：

**1. 修改数据库角色约束：**

```sql
ALTER TABLE users
ADD CONSTRAINT chk_role
CHECK (role IN ('platform_admin', 'admin', 'user', 'viewer', 'your_new_role'));
```

**2. 修改后端角色常量：**

```go
// services/user-svc/model/user.go

const (
    RolePlatformAdmin = "platform_admin"
    RoleAdmin         = "admin"
    RoleUser          = "user"
    RoleViewer        = "viewer"
    RoleYourNewRole    = "your_new_role"  // 新增
)
```

**3. 添加权限检查中间件：**

```go
// services/user-svc/user-svc.go

func requireYourNewRole() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, _ := c.Get("role")
        if role != "your_new_role" {
            jsonError(c, http.StatusForbidden, "your_new_role access required")
            c.Abort()
            return
        }
        c.Next()
    }
}
```

**4. 在路由中应用：**

```go
yourNewRole := r.Group("/api/your-feature")
yourNewRole.Use(authMiddleware(), requireYourNewRole())
{
    yourNewRole.GET("/data", handleGetData)
}
```

### 如何修改 JWT 过期时间？

JWT 过期时间在 `user-svc.go` 的 `handleLogin` 函数中设置：

```go
// services/user-svc/user-svc.go 第 411 行

claims := &Claims{
    // ...
    RegisteredClaims: jwt.RegisteredClaims{
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 改成 7 * 24 * time.Hour 即可延期到 7 天
        IssuedAt:  jwt.NewNumericDate(time.Now()),
    },
}
```

**常用过期时间设置：**

| 过期时间 | 代码 |
|----------|------|
| 24 小时 | `24 * time.Hour` |
| 7 天 | `7 * 24 * time.Hour` |
| 30 天 | `30 * 24 * time.Hour` |
| 永不过期（不推荐） | 删除 `ExpiresAt` 字段 |

### 如何添加新的 API 端点？

**1. 在 user-svc 添加新接口：**

```go
// services/user-svc/user-svc.go

// 在 main() 函数中注册路由
func main() {
    // ...
    r := gin.New()
    r.Use(gin.Logger())
    r.Use(gin.Recovery())

    // 公开接口
    r.POST("/api/user/register", handleRegister)
    r.POST("/api/user/login", handleLogin)

    // 新增接口示例：POST /api/user/change-password
    r.POST("/api/user/change-password", handleChangePassword)  // 添加这行

    // 需要认证的接口
    auth := r.Group("/api/user")
    auth.Use(authMiddleware())
    {
        auth.GET("/info", handleUserInfo)
        // 新增需要认证的接口
        auth.GET("/settings", handleGetSettings)
    }
}

// 新增 Handler
func handleChangePassword(c *gin.Context) {
    // 获取当前用户 ID
    userID := getUserID(c)

    var body struct {
        OldPassword string `json:"oldPassword"`
        NewPassword string `json:"newPassword"`
    }
    if err := c.ShouldBindJSON(&body); err != nil {
        jsonError(c, http.StatusBadRequest, "invalid request")
        return
    }

    // 验证旧密码
    // ... 业务逻辑

    jsonSuccessMsg(c, "password changed")
}
```

**2. 如果需要权限控制，添加中间件：**

```go
// 新路由使用权限中间件
admin := r.Group("/api/admin")
admin.Use(authMiddleware(), platformAdminOnly())
{
    // 现有路由...
    admin.POST("/new-endpoint", handleNewEndpoint)
}
```

### 如何连接外部数据库？

**方式一：修改环境变量（推荐）**

```bash
# 启动服务时指定
DB_HOST=your-mysql-host \
DB_PORT=3306 \
DB_USER=your-user \
DB_PASSWORD=your-password \
DB_NAME=agentmesh \
go run .
```

**方式二：修改 docker-compose.yml**

```yaml
services:
  user-svc:
    # ...
    environment:
      - DB_HOST=mysql                    # 改成你的 MySQL 地址
      - DB_PORT=3306
      - DB_USER=root
      - DB_PASSWORD=root123
      - DB_NAME=agentmesh
```

**方式三：使用 .env 文件**

```bash
# 在服务目录创建 .env 文件
DB_HOST=your-mysql-host
DB_PORT=3306
DB_USER=your-user
DB_PASSWORD=your-password
DB_NAME=agentmesh
```

### 如何部署到生产环境？

**1. 构建镜像**

```bash
# 构建所有服务镜像
make build

# 或手动构建
cd services/api-gateway && docker build -t agentmesh-api-gateway .
cd services/user-svc && docker build -t agentmesh-user-svc .
cd services/plugin-svc && docker build -t agentmesh-plugin-svc .
```

**2. 配置生产环境变量**

```yaml
# docker-compose.prod.yml

services:
  api-gateway:
    image: agentmesh-api-gateway:latest
    environment:
      - GIN_MODE=release
      - DB_HOST=mysql-prod
      - DB_PORT=3306
      - DB_USER=prod_user
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=agentmesh
    # JWT 密钥必须通过环境变量注入！
    - JWT_SECRET=${JWT_SECRET}
```

**3. 安全检查清单**

- [ ] 修改默认的 `jwtSecret`（`agentmesh-secret-key-change-in-production`）
- [ ] 修改数据库密码（`root123`）
- [ ] 使用 HTTPS（配置反向代理 Nginx/Caddy）
- [ ] 启用防火墙，只开放必要端口
- [ ] 配置日志收集（ELK/Loki）
- [ ] 配置监控（Prometheus + Grafana）

**4. 使用 Nginx 反向代理**

```nginx
server {
    listen 443 ssl;
    server_name api.agentmesh.io;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

---

## 10. 下一步

### 推荐的学习路径

1. **入门（第一天）**
   - 阅读本文档的第 1-4 章，理解核心概念
   - 本地启动服务，登录管理后台
   - 尝试创建租户、用户、启用/禁用插件

2. **熟悉代码（第二三天）**
   - 阅读 `user-svc.go`（1250 行，包含详细注释）
   - 阅读 `plugin-svc.go` 理解插件机制
   - 阅读 `api-gateway/middleware/auth.go` 理解 JWT 流程

3. **实践开发（第四五天）**
   - 按第 7 章指南开发一个简单的前端插件
   - 尝试添加一个新的 API 端点
   - 理解前后端数据交互

4. **深入理解（一周后）**
   - 研究 plugin-sdk 的注册机制
   - 理解多租户数据隔离的 SQL 写法
   - 学习如何扩展 RBAC 权限

### 如何深入阅读代码

| 文件 | 行数 | 推荐阅读顺序 | 说明 |
|------|------|-------------|------|
| `services/user-svc/user-svc.go` | 1250 | **必须阅读** | 包含完整的中文注释，是学习最佳材料 |
| `services/plugin-svc/plugin-svc.go` | 448 | 第二 | 插件管理逻辑，相对独立 |
| `services/api-gateway/middleware/auth.go` | 137 | 第三 | JWT 验证流程 |
| `web/packages/plugin-sdk/src/hooks.ts` | - | 第四 | 插件注册机制 |

### 贡献代码的流程

1. **Fork 仓库**（如果是开源贡献）

2. **创建功能分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **编写代码并测试**
   ```bash
   # 本地测试
   make test

   # 构建验证
   make build
   ```

4. **提交代码**
   ```bash
   git add -A
   git commit -m "feat: 添加新功能"
   ```

5. **推送到远程**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **创建 Pull Request**
   - 描述改动内容和动机
   - 关联相关 Issue

---

## 附录

### A. 环境变量参考

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `DB_HOST` | localhost | MySQL 主机 |
| `DB_PORT` | 3306 | MySQL 端口 |
| `DB_USER` | root | 数据库用户 |
| `DB_PASSWORD` | root123 | 数据库密码 |
| `DB_NAME` | agentmesh | 数据库名 |
| `GIN_MODE` | debug | Gin 运行模式（debug/release） |

### B. 端口映射

| 服务 | 容器内端口 | 主机端口 | URL |
|------|-----------|---------|-----|
| api-gateway | 8080 | 8080 | http://localhost:8080 |
| user-svc | 8081 | 8081 | http://localhost:8081 |
| plugin-svc | 8083 | 8083 | http://localhost:8083 |
| MySQL | 3306 | 3306 | localhost:3306 |

### C. 联系方式

- **GitHub Issues**: https://github.com/your-org/agentmesh/issues
- **文档**: https://docs.agentmesh.io

---

*本文档最后更新于 2026/06/09*
