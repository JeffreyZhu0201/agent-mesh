# AgentMesh 基础框架重构 — 设计文档

> 定稿日期：2026-06-09
> 目标：去掉 LLM Agent 部分，保留用户认证基座 + 插件系统，搭建可快速二次开发的后台框架

---

## 一、背景与目标

当前 AgentMesh 内置了 AI 对话（agent-svc），但用户希望将其作为**通用后台开发框架**使用，无需处理认证、多租户等通用能力，只需往插件系统里塞业务逻辑即可。

参考若伊（Ruoyi）框架的设计哲学：认证/租户/权限/插件体系一次搭好，业务开发只在插件层发力。

**本次重构不做的事（Out of Scope）**
- 删除 agent-svc 代码（保留，docker-compose 保留，标记为可选）
- AI 对话相关前端页面（Chat 插件暂时移除或标记为可选）
- 菜单/路由动态配置
- 字典/配置管理

---

## 二、设计决策汇总

| 维度 | 选择 |
|------|------|
| 多租户实现 | 共享表 + `tenant_id` 字段隔离（若伊方案） |
| 权限模型 | 固定三元角色：`platform_admin` / `admin` / `user` / `viewer` |
| 管理员体系 | 平台超级管理员（tenant_id=0）+ 租户管理员（tenant_id=N, role=admin）两条线 |
| 后台入口 | 独立 `admin` app（:3001）给平台管理员；`main` app（:3000）给租户用户 |
| 用户创建 | 自住注册（默认 viewer） + 管理员提升权限 |
| 后台功能范围 | 用户 CRUD + 租户 CRUD + 插件开关（精简） |
| 租户内管控 | 租户管理员只能操作用户，不能操作其他 admin |

---

## 三、整体架构

```
┌──────────────────────────────────────────────────────┐
│                      浏览器                           │
│                                                      │
│   ┌──────────────────┐   ┌──────────────────────┐   │
│   │   main app       │   │   admin app          │   │
│   │  localhost:3000  │   │  localhost:3001      │   │
│   │  租户用户工作台   │   │  平台超级管理员后台   │   │
│   └────────┬─────────┘   └──────────┬───────────┘   │
│            │                         │               │
│            └──────────┬──────────────┘               │
│                       │ HTTP / JWT                    │
│              ┌────────▼────────┐                     │
│              │  API Gateway    │                     │
│              │   :8080         │                     │
│              │  CORS · 路由    │                     │
│              └──┬────┬────┬───┘                      │
│         ┌───────┘    │    └───────┐                  │
│   ┌─────▼────┐ ┌─────▼────┐ ┌─────▼────┐            │
│   │user-svc │ │plugin-svc│ │ agent-svc│            │
│   │ :8081   │ │ :8083    │ │ :8082    │            │
│   │ 认证    │ │ 插件管理 │ │ AI对话   │            │
│   │ 用户CRUD│ │租户开关  │ │(可选/保留)│            │
│   └────┬────┘ └────┬────┘ └────┬────┘              │
│        │           │           │                    │
│        └───────────┼───────────┘                    │
│                    │                                │
│            ┌───────▼───────┐                        │
│            │    MySQL 8    │                        │
│            │  agentmesh   │                        │
│            └──────────────┘                        │
└──────────────────────────────────────────────────────┘
```

---

## 四、数据库设计

### 4.1 租户表 `tenants`

平台级表，记录所有租户。tenant_id=0 保留给平台管理员。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AUTO_INCREMENT | |
| name | VARCHAR(128) NOT NULL | 租户显示名称 |
| code | VARCHAR(64) UNIQUE NOT NULL | 租户代码（URL/标识用） |
| status | TINYINT DEFAULT 1 | 1=正常 0=禁用 |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP ON UPDATE | |

### 4.2 用户表 `users`

所有用户（平台管理员 + 租户用户）共用一张表。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AUTO_INCREMENT | |
| username | VARCHAR(64) UNIQUE NOT NULL | 登录名 |
| password_hash | VARCHAR(255) NOT NULL | bcrypt 加密 |
| email | VARCHAR(128) | |
| tenant_id | BIGINT NOT NULL DEFAULT 0 | 0=平台管理员 |
| role | VARCHAR(32) DEFAULT 'viewer' | platform_admin / admin / user / viewer |
| status | TINYINT DEFAULT 1 | 1=正常 0=禁用 |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP ON UPDATE | |

**索引**：`INDEX idx_tenant (tenant_id)`

**初始化**：平台超级管理员通过 SQL 种子创建（username=admin，tenant_id=0，role=platform_admin）。

### 4.3 插件注册表 `plugins`

跨租户共享一份插件清单。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AUTO_INCREMENT | |
| plugin_key | VARCHAR(64) UNIQUE NOT NULL | 唯一标识，如 "dashboard" |
| name | VARCHAR(128) NOT NULL | 显示名 |
| version | VARCHAR(32) | 版本号 |
| type | VARCHAR(32) | page / agent / integration |
| description | TEXT | |
| author | VARCHAR(128) | |
| builtin | BOOLEAN DEFAULT TRUE | 内置 vs 用户安装 |
| created_at | TIMESTAMP | |

### 4.4 租户插件开关 `tenant_plugins`

租户级插件启用状态，无记录=默认可用（enabled=true）。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AUTO_INCREMENT | |
| tenant_id | BIGINT NOT NULL | |
| plugin_key | VARCHAR(64) NOT NULL | |
| enabled | BOOLEAN DEFAULT TRUE | |
| updated_at | TIMESTAMP ON UPDATE | |
| UNIQUE KEY | (tenant_id, plugin_key) | |

---

## 五、API 设计

### 5.1 user-svc（:8081）

#### 认证（公开）
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/user/register` | 自主注册，默认 tenant_id=1（第一个租户），role=viewer |
| POST | `/api/user/login` | 登录，返回 JWT |

#### 用户信息（登录用户）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/user/info` | 当前用户信息 🔐 |

#### 平台管理员接口（role=platform_admin）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/tenants` | 租户列表 |
| POST | `/api/admin/tenants` | 创建租户 |
| PUT | `/api/admin/tenants/:id` | 编辑租户 |
| DELETE | `/api/admin/tenants/:id` | 删除租户 |
| GET | `/api/admin/users` | 所有用户（可加 ?tenant_id=N 筛选） |
| POST | `/api/admin/users` | 创建任意用户 |
| PUT | `/api/admin/users/:id` | 编辑任意用户 |
| DELETE | `/api/admin/users/:id` | 删除任意用户 |

#### 租户管理员接口（role=admin，tenant_id 匹配）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/tenant/users` | 本租户用户列表 |
| POST | `/api/tenant/users` | 本租户创建用户（role 最高到 user，不能创建 admin）|
| PUT | `/api/tenant/users/:id` | 编辑本租户用户（不能操作 role=admin 的用户）|
| DELETE | `/api/tenant/users/:id` | 删除本租户用户（不能删除 role=admin 的用户）|

### 5.2 plugin-svc（:8083，保持不变）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/plugin/list` | 租户可见插件 🔐 |
| GET | `/api/plugin/admin/list` | 全部插件 + 启用状态 🔐 |
| POST | `/api/plugin/admin/toggle` | 切换插件状态 🔐 |

---

## 六、认证与权限链条

```
用户登录
    ↓
user-svc 验证 bcrypt 密码
    ↓
签发 JWT（payload: userId, username, tenantId, role）
    ↓
请求 → api-gateway
    ↓
AuthMiddleware 解析 JWT
    ↓
注入请求上下文：
  X-User-ID:    claims.userId
  X-Tenant-ID:  claims.tenantId
  X-User-Role:  claims.role
    ↓
各微服务 handler 读取 c.Get("tenantId") / c.Get("userId") / c.Get("role")
```

**接口权限矩阵：**

| 接口前缀 | 允许角色 |
|---------|---------|
| `/api/user/*`（不含 login/register）| any authenticated |
| `/api/admin/*` | platform_admin |
| `/api/tenant/*` | admin（且 tenantId 匹配）|
| `/api/plugin/*` | any authenticated |

---

## 七、前端变化

### 7.1 admin app（localhost:3001）

平台超级管理员专用后台，功能：

- **仪表盘**：系统概览（租户数、用户数、插件状态）
- **租户管理**：CRUD 租户（创建租户后自动创建 tenant_id 对应记录）
- **用户管理**：CRUD 所有用户（分平台用户/租户用户两个 Tab）
- **插件管理**：所有插件列表 + 租户级开关

登录：`/app/login` → 输入平台管理员账号

### 7.2 main app（localhost:3000）

租户用户工作台。注册后默认 viewer 角色，管理员可提升为 admin/user。

暂时移除 Chat 相关页面（agent-svc 保留为可选依赖）。

### 7.3 插件 SDK 不变

`registerPlugin()` 体系保持不变，以后开发新业务插件只需：
1. 在 `main/src/plugins/` 注册路由和菜单
2. 后端按需新增 service（类似 plugin-svc）
3. 前端 SDK 自动过滤未启用的插件

---

## 八、实施顺序

1. **数据库迁移** — user-svc 接入 MySQL，建表（tenants/users/plugins/tenant_plugins），迁移现有内存数据到 MySQL
2. **user-svc 重构** — 实现平台/租户管理员 API，JWT 中间件，bcrypt 鉴权
3. **admin app 重构** — UserManagement / TenantManagement 接真实 API，修复现有 mock 数据
4. **main app** — 注册流程不变（转为 viewer），移除 Chat 页面入口
5. **docker-compose** — 保留 agent-svc（标记注释说明可选），清理无用的环境变量
6. **README 更新** — 同步最新架构

---

## 九、初始数据

```sql
-- 平台超级管理员（系统初始化）
INSERT INTO users (username, password_hash, email, tenant_id, role, status)
VALUES ('admin', '$2a...', 'admin@agentmesh.io', 0, 'platform_admin', 1);

-- 默认租户（用户自主注册默认加入此租户）
INSERT INTO tenants (name, code, status) VALUES ('Default Tenant', 'default', 1);
```

---

## 十、验收标准

- [ ] 用户可自主注册，默认 viewer 角色，默认加入 default 租户
- [ ] 平台管理员（admin）可创建/编辑/删除任意租户
- [ ] 平台管理员可查看/编辑/删除所有用户
- [ ] 租户管理员可查看/创建/编辑/删除本租户内 role!=admin 的用户
- [ ] 租户管理员无法操作本租户内 role=admin 的用户
- [ ] 插件开关对租户内用户生效（未启用插件的菜单/路由不可见）
- [ ] JWT 过期/无效时自动跳转登录页
- [ ] admin 和 main app 独立部署在不同端口，各司其职
