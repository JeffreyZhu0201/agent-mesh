<div align="center">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go" alt="Go" />
  <img src="https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react" alt="React" />
  <img src="https://img.shields.io/badge/TypeScript-5-3178C6?style=for-the-badge&logo=typescript" alt="TypeScript" />
  <img src="https://img.shields.io/badge/MUI-5-007FFF?style=for-the-badge&logo=mui" alt="MUI" />
  <img src="https://img.shields.io/badge/MySQL-8.0-4479A1?style=for-the-badge&logo=mysql" alt="MySQL" />
  <img src="https://img.shields.io/badge/Docker-compose-2496ED?style=for-the-badge&logo=docker" alt="Docker Compose" />
  <br/>
  <img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="License" />
</div>

<h1 align="center">🚀 AgentMesh</h1>

<p align="center">
  <strong>SaaS 多租户后台开发框架</strong><br/>
  基于 Go + React 的全栈框架，内置多租户认证、RBAC 权限、插件系统
</p>

---

## ✨ 特性

| 特性 | 说明 |
|------|------|
| 🏢 **多租户** | 共享表隔离（tenant_id），JWT 鉴权贯穿全链路 |
| 🔐 **RBAC 权限** | 平台管理员 / 管理员 / 普通用户 / 访客 四级角色 |
| 🧩 **插件系统** | 前端模块注册 + 后端启用控制，租户级开关 |
| 🚀 **快速开发** | 面向 SaaS 后台的业务框架，登录即用、插件即插即用 |
| 🎨 **现代化 UI** | React 18 + MUI + TypeScript，响应式布局 |
| 🐳 **一键部署** | Docker Compose 编排，3 个微服务 + MySQL |

---

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                            🌐 浏览器                                    │
│                     React 18 + MUI + Zustand                            │
│              ┌──────────────┐  ┌──────────────┐                        │
│              │   Main App   │  │  Admin App   │                        │
│              │  (用户工作台) │  │  (管理后台)   │                        │
│              └──────┬───────┘  └──────┬───────┘                        │
├─────────────────────┼─────────────────┼────────────────────────────────┤
│                     │   HTTP / JSON   │                                │
│              ┌──────▼─────────────────▼──────┐                         │
│              │      🚪 API Gateway          │                         │
│              │    Gin Reverse Proxy         │                         │
│              │    :8080 · CORS · 路由转发    │                         │
│              └──┬───────┬───────────┘                              │
│                 │       │           │                                      │
│          ┌──────▼──┐ ┌──▼────┐                                     │
│          │ User   │ │ Plugin│                                     │
│          │ Service│ │Service│                                     │
│          │ :8081  │ │ :8083 │                                     │
│          │ 认证   │ │插件管理│                                     │
│          │ 用户   │ │租户开关│                                     │
│          │ RBAC   │ │       │                                     │
│              │         │                                              │
│              └─────────┼───────────┘                                  │
│                        │                                              │
│              ┌─────────▼──────────┐                                   │
│              │    🗄️ MySQL 8.0   │                                   │
│              │      agentmesh     │                                   │
│              └────────────────────┘                                   │
└─────────────────────────────────────────────────────────────────────────┘
```

### 微服务矩阵

| 服务 | 端口 | 职责 | 技术亮点 |
|------|------|------|----------|
| **api-gateway** | `:8080` | 统一入口、路由转发、CORS | Gin Reverse Proxy |
| **user-svc** | `:8081` | 用户注册/登录、JWT 签发、RBAC | bcrypt + JWT |
| **plugin-svc** | `:8083` | 插件注册表、租户启用/禁用 | GORM AutoMigrate |

### 前端架构

```
web/
├── apps/
│   ├── main/         🖥️  用户工作台 — 业务插件
│   ├── admin/        ⚙️  管理后台 — 插件管理、租户配置
│   └── landing/      📄  营销官网
├── packages/
│   ├── ui/           🎨  共享 MUI 组件库
│   ├── plugin-sdk/   🧩  前端插件 SDK
│   ├── stores/       📦  Zustand 全局状态
│   ├── api/          🔌  API 客户端封装
│   ├── hooks/        🪝  通用 React Hooks（占位）
│   └── types/        📐  TypeScript 类型定义
```

---

## 🚀 快速开始

### 前置条件

- Go 1.24+
- Node.js 18+
- pnpm 9+
- Docker & Docker Compose

### 一键启动（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/your-org/agentmesh.git
cd agentmesh

# 2. 启动所有服务（MySQL + 3 个微服务）
make docker-up

# 3. 安装前端依赖并启动
cd web
pnpm install
pnpm dev

# 4. 浏览器访问
open http://localhost:3000
```

默认平台管理员：`admin` / `admin123`（首次部署后请修改）

---

## 💻 开发指南

### 常用命令

```bash
make help           # 查看所有命令
make install        # 安装 Go 服务依赖
make build          # 构建服务二进制到 bin/
make docker-up      # Docker Compose 启动
make docker-down    # 停止服务
make test           # 运行全部 Go 单元测试
make test-coverage  # Go + 前端覆盖率报告
```

### 前端开发

```bash
cd web
pnpm install        # 安装依赖
pnpm dev            # 开发服务器 → localhost:3000
pnpm build          # 生产构建
pnpm test           # 运行 Vitest 单元测试
pnpm test:coverage  # 前端覆盖率（目标 ≥50%）
```

### 认证流程

```
1. POST /api/user/register  →  创建账户（bcrypt 加密密码）
2. POST /api/user/login     →  获取 JWT Token（响应格式：{ success, data: { token, user } }）
3. 前端存储 Token 至 localStorage（Zustand authStore）
4. 后续请求携带 Authorization: Bearer <token>
5. 各微服务通过 JWT 中间件解析 userId / tenantId / role
```

---

## 🧪 测试

项目包含 Go 后端测试和前端 Vitest 测试，核心模块覆盖率 ≥50%。

| 模块 | 测试文件 | 覆盖率 |
|------|----------|--------|
| user-svc | `user_svc_test.go` | ~52% |
| plugin-svc | `plugin_svc_test.go` | ~64% |
| plugin-svc/plugin | `plugin_test.go` | ~95% |
| api-gateway/middleware | `auth_test.go` | ~64% |
| @agentmesh/stores | `*.test.ts` | ~94% |
| @agentmesh/plugin-sdk | `hooks.test.ts` | ~45% |
| @agentmesh/api | `auth.test.ts` | ~67% |

```bash
# 后端测试
make test

# 前端测试
cd web && pnpm test

# 完整覆盖率
make test-coverage
```

---

## 📁 项目结构

```
AgentMesh/
├── services/                 # Go 微服务
│   ├── api-gateway/         # API 网关（Gin 反向代理）
│   ├── user-svc/            # 用户认证 + RBAC
│   ├── plugin-svc/          # 插件管理
│   ├── plugins/chat/      # 示例插件（未接入 docker-compose）
│   └── sql/                 # 数据库初始化脚本
├── web/                     # 前端 Monorepo（pnpm + Turbo）
│   ├── apps/main|admin|landing
│   └── packages/ui|stores|plugin-sdk|api|hooks|types
├── docs/                    # 设计文档与规划（本地）
├── docker-compose.yml
├── Makefile
└── LICENSE
```

---

## 🧩 插件系统

### 后端 (`plugin-svc`)
- `plugins` 表 — 插件注册表，启动时自动播种内置插件
- `tenant_plugins` 表 — 租户级启用/禁用（无记录 = 默认启用）
- API：`/api/plugin/list`、`/api/plugin/admin/list`、`/api/plugin/admin/toggle`

### 前端 (`@agentmesh/plugin-sdk`)
- `registerPlugin()` 注册路由与菜单
- 启动时从后端拉取租户启用插件列表
- 未启用插件的菜单和路由自动隐藏

---

## 🔌 API 概览

| 方法 | 路径 | 说明 | 微服务 |
|------|------|------|--------|
| `POST` | `/api/user/register` | 用户注册 | user-svc |
| `POST` | `/api/user/login` | 用户登录 → JWT | user-svc |
| `GET` | `/api/user/info` | 获取用户信息 🔐 | user-svc |
| `GET` | `/api/plugin/list` | 租户可见插件列表 🔐 | plugin-svc |
| `GET` | `/api/plugin/admin/list` | 全部插件（含状态）🔐 | plugin-svc |
| `POST` | `/api/plugin/admin/toggle` | 切换插件启用状态 🔐 | plugin-svc |

> 🔐 = 需要 `Authorization: Bearer <JWT>` 头

---

## 🛠️ 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.24+ · Gin · GORM · JWT · bcrypt |
| 数据库 | MySQL 8.0 |
| 前端 | React 18 · TypeScript · Vite · MUI · Zustand |
| 包管理 | pnpm workspaces · Turbo |
| 测试 | Go testing + testify · Vitest |
| 部署 | Docker Compose |

---

## 📄 许可

[MIT License](LICENSE)

---

<p align="center">
  用 ❤️ 构建 · AgentMesh · SaaS 多租户后台开发框架
</p>
