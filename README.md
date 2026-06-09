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

<p align="center">
  <img src="./chat-welcome.png" width="600" alt="Chat Welcome Screen" />
</p>

---

## ✨ 特性

| 特性 | 说明 |
|------|------|
| 🏢 **多租户** | Schema 级租户隔离（shared-table），JWT 鉴权贯穿全链路 |
| 🔐 **RBAC 权限** | 平台管理员 / 管理员 / 普通用户 / 访客 四级角色 |
| 🧩 **插件系统** | 前端模块注册 + 后端启用控制，租户级开关，LLM 为可选插件 |
| 🚀 **快速开发** | 面向 SaaS 后台的业务框架，登录即用、插件即插即用 |
| 🎨 **现代化 UI** | React 18 + MUI + TypeScript，响应式布局 |
| 🐳 **一键部署** | Docker Compose 编排，4 个微服务 + MySQL |

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
│              └──┬───────┬───────┬───────────┘                         │
│                 │       │       │                                      │
│          ┌──────▼──┐ ┌──▼────┐ ┌▼──────────┐                          │
│          │ User   │ │ Agent │ │ Plugin    │                          │
│          │ Service│ │Service│ │ Service   │                          │
│          │ :8081  │ │:8082  │ │ :8083     │                          │
│          │ 认证   │ │ AI    │ │ 插件管理   │                          │
│          │ 用户   │ │ (可选) │ │ 租户开关  │                          │
│          │ RBAC   │ │ 对话  │ │           │                          │
│              │         │           │                                  │
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
| **user-svc** | `:8081` | 用户注册/登录、JWT 签发 | bcrypt + JWT |
| **agent-svc** | `:8082` | AI 对话（可选插件）| Eino + LLM 流式 SSE |
| **plugin-svc** | `:8083` | 插件注册表、租户启用/禁用 | GORM AutoMigrate |

### 前端架构

```
web/
├── apps/
│   ├── main/         🖥️  用户工作台 — 聊天、Dashboard
│   ├── admin/        ⚙️  管理后台 — 插件管理、租户配置
│   └── landing/      📄  营销官网
├── packages/
│   ├── ui/           🎨  共享 MUI 组件库 (Layout, Button...)
│   ├── plugin-sdk/   🧩  前端插件 SDK (注册、路由、菜单)
│   ├── stores/       📦  Zustand 全局状态 (auth, plugins)
│   ├── api/          🔌  API 客户端封装
│   ├── hooks/        🪝  通用 React Hooks
│   └── types/        📐  TypeScript 类型定义
```

---

## 🧩 插件系统

AgentMesh 的插件系统分为两大部分：

### 后端插件服务 (`plugin-svc`)
- **`plugins` 表** — 插件注册表，启动时自动播种内置插件
- **`tenant_plugins` 表** — 租户级启用/禁用状态（多租户隔离）
- **API 端点**：`/api/plugin/list`（租户可见）、`/admin/list`（管理列表）、`/admin/toggle`（开关）
- **默认策略**：无租户配置时视为启用

### 前端插件 SDK (`@agentmesh/plugin-sdk`)
- 插件通过 `registerPlugin()` 注册路由、菜单项、图标
- 应用启动时从后端拉取租户的启用插件列表
- 未启用的插件菜单和路由自动隐藏
- 内置插件：**Dashboard**（看板）
- 可选插件：**Chat**（LLM 对话，需启用 agent-svc）

<p align="center">
  <img src="./admin-plugins.png" width="700" alt="Plugin Management" />
</p>

---

## 🚀 快速开始

### 前置条件

- Go 1.24+
- Node.js 18+
- Docker & Docker Compose
- MySQL 8.0（Docker 自动启动）

### 一键启动（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/your-org/agentmesh.git
cd agentmesh

# 2. 启动所有服务（MySQL + 4 个微服务）
make docker-up

# 3. 安装前端依赖并启动
cd web
npm install
npm run dev

# 4. 浏览器访问
open http://localhost:3000
```

### 分步启动

```bash
# 后端 — 构建并启动 Docker 服务
docker-compose up -d

# 查看启动日志
docker-compose logs -f

# 前端 — 开发模式
cd web
npm install
npm run dev
```

---

## 💻 开发指南

### 常用命令

```bash
make help          # 查看所有命令
make install       # 安装所有 Go 服务依赖
make build         # 构建服务二进制
make docker-up     # Docker Compose 启动
make docker-down   # 停止服务
make test          # 运行测试
```

### 前端开发

```bash
cd web
npm install        # 安装依赖
npm run dev        # 开发服务器 → localhost:3000
npm run build      # 生产构建
npm run lint       # 代码检查
```

### 认证流程

```
1. POST /api/user/register  →  创建账户（bcrypt 加密密码）
2. POST /api/user/login     →  获取 JWT Token
3. 前端存储 Token 至 localStorage（通过 Zustand authStore）
4. 后续请求携带 Authorization: Bearer <token>
5. 各微服务通过 JWT 中间件解析 userId / tenantId / role
```

### AI 对话配置（可选）

`agent-svc` 为可选插件，支持多种 LLM 后端，通过环境变量配置：

```bash
# 方式一：Anthropic Claude
ANTHROPIC_API_KEY=sk-xxx
ANTHROPIC_BASE_URL=https://api.anthropic.com
ANTHROPIC_MODEL=claude-3-5-sonnet-20241022

# 方式二：OpenAI 兼容接口（如 Volcengine Ark）
LLM_API_KEY=xxx
LLM_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
LLM_MODEL=ep-xxxx-yyyyy
```

<p align="center">
  <img src="./chat-with-code.png" width="700" alt="Chat with Code Highlighting" />
</p>

---

## 📁 项目结构

```
AgentMesh/
├── services/                 # 🎯 Go 微服务
│   ├── api-gateway/         #  API 网关（Gin Reverse Proxy）
│   ├── user-svc/            #  用户认证服务
│   ├── agent-svc/           #  AI 对话服务（可选插件）
│   └── plugin-svc/          #  插件管理服务
├── web/                     # 🎨 前端 Monorepo
│   ├── apps/
│   │   ├── main/            #  主应用（Vite + React）
│   │   ├── admin/           #  管理后台
│   │   └── landing/         #  营销官网
│   └── packages/
│       ├── ui/              #  共享 UI 组件库
│       ├── plugin-sdk/      #  插件开发 SDK
│       ├── stores/          #  Zustand 状态管理
│       ├── api/             #  API 客户端
│       ├── hooks/           #  通用 Hooks
│       └── types/           #  TypeScript 类型
├── docker-compose.yml       # 🐳 Docker 编排
├── Makefile                  # 🛠️ 构建脚本
└── CLAUDE.md                # 📝 AI 辅助开发指南
```

---

## 🛠️ 技术栈

| 层 | 技术 | 用途 |
|----|------|------|
| 后端语言 | Go 1.24+ | 高性能微服务 |
| **Web 框架** | Gin | HTTP 路由、中间件 |
| **AI 框架** | Eino（可选）| LLM 调用编排（插件） |
| **数据库** | MySQL 8.0 + GORM | 持久化存储 |
| **认证** | JWT + bcrypt | 安全鉴权 |
| **前端** | React 18 + TypeScript | 用户界面 |
| **UI 库** | MUI 5 | 组件系统 |
| **状态管理** | Zustand | 全局状态 |
| **构建** | Vite 5 | 前端打包 |
| **容器化** | Docker Compose | 服务编排 |

---

## 🔌 API 概览

| 方法 | 路径 | 说明 | 微服务 |
|------|------|------|--------|
| `POST` | `/api/user/register` | 用户注册 | user-svc |
| `POST` | `/api/user/login` | 用户登录 → JWT | user-svc |
| `GET` | `/api/user/info` | 获取用户信息 🔐 | user-svc |
| `POST` | `/api/agent/chat` | AI 对话（SSE 流式）🔐 | agent-svc |
| `GET` | `/api/agent/conversations` | 会话历史列表 🔐 | agent-svc |
| `GET` | `/api/plugin/list` | 租户可见插件列表 🔐 | plugin-svc |
| `GET` | `/api/plugin/admin/list` | 全部插件（含状态）🔐 | plugin-svc |
| `POST` | `/api/plugin/admin/toggle` | 切换插件启用状态 🔐 | plugin-svc |

> 🔐 = 需要 `Authorization: Bearer <JWT>` 头

---

## 📄 许可

[MIT License](LICENSE)

---

<p align="center">
  用 ❤️ 构建 · AgentMesh · SaaS 多租户后台开发框架
</p>