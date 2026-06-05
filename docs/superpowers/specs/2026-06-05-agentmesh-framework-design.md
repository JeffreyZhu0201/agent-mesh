# AgentMesh 开发框架基座设计

## 1. 项目概述

**项目名称：** AgentMesh
**项目类型：** SaaS 多租户 AI 应用开发框架
**核心功能：** 提供一套完整的后端微服务 + 前端三站点的开发框架底座，内置插件系统，方便后续快速开发 AI 应用
**目标用户：** 开发团队，需要快速搭建 AI SaaS 产品的技术负责人

---

## 2. 技术栈

| 层级 | 技术选型 | 说明 |
|------|----------|------|
| 后端框架 | go-zero | 微服务 RPC 框架 |
| 前端框架 | React 18 + TypeScript | |
| UI 组件库 | Material UI (MUI) | |
| AI 智能体 | Eino | LLM编排框架 |
| 数据库 | MySQL | SaaS Schema 级别隔离 |
| 部署 | Docker + docker-compose | 本地开发 |
| API 风格 | REST (对外) + gRPC (内部) | |
| 前端构建 | Vite | |
| 状态管理 | Zustand | |

---

## 3. 项目结构

```
AgentMesh/
├── docker-compose.yml
├── Makefile
├── docs/
│   └── superpowers/specs/
├── services/                    # 后端微服务
│   ├── api-gateway/             # 端口 8080
│   ├── user-svc/                # 端口 8081
│   ├── agent-svc/               # 端口 8082
│   └── plugin-svc/              # 端口 8083
└── web/                        # 前端 Monorepo
    ├── apps/
    │   ├── landing/             # 落地页 /
    │   ├── main/                # 主站 /app
    │   └── admin/               # 后台 /admin
    ├── packages/
    │   ├── ui/                  # 共享 MUI 组件库
    │   ├── hooks/               # 共享 React Hooks
    │   ├── stores/              # 状态管理
    │   └── api/                 # API 客户端
    └── plugins/
        └── chat/                # LLM Chat 示例插件
```

---

## 4. 后端微服务设计

### 4.1 服务列表

| 服务 | 端口 | 数据库 Schema | 核心功能 |
|------|------|---------------|----------|
| api-gateway | 8080 | - | 路由、鉴权、限流、多租户中间件 |
| user-svc | 8081 | agentmesh_public, agentmesh_tenant_{id} | 用户注册登录、JWT、租户管理 |
| agent-svc | 8082 | agentmesh_tenant_{id} | Eino 编排、LLM 调用 |
| plugin-svc | 8083 | agentmesh_tenant_{id} | 插件注册、动态加载、生命周期管理 |

### 4.2 服务间通信

- 前端 → 网关：HTTP REST
- 网关 → 后端服务：gRPC
- 后端服务间：gRPC

### 4.3 多租户方案

**Schema 级别隔离**，数据库命名：`agentmesh_tenant_{tenant_id}`

**请求流程：**
1. 用户登录 → user-svc 验证 → 返回 JWT（含 tenant_id）
2. 请求带 Authorization + X-Tenant-ID header
3. api-gateway 中间件解析租户 → 路由到对应服务
4. 各 service 通过 tenant 中间件自动切换 DataSource

---

## 5. 插件系统架构

### 5.1 后端插件（Go Plugin）

- 插件以 `.so` 动态库形式存在
- 通过 `plugin-svc` 的 `plugin.Open()` 动态加载/卸载
- 通过 gRPC 与 agent-svc 通信

**插件接口：**
```go
type Plugin interface {
    Name() string
    Version() string
    Init(ctx context.Context) error
    RegisterRoutes(r *Router)
    GetMenuItem() *MenuItem
}
```

### 5.2 前端插件（React Module）

- 通过 `registerPlugin()` 注册菜单和路由
- 独立 bundle，按需加载
- 通过 `@agentmesh/plugin-sdk` 与后端通信

---

## 6. 前端三站设计

### 6.1 站点列表

| 站点 | 路径 | 内容 |
|------|------|------|
| 落地页 | `/` | Hero、功能介绍、定价、FAQ、CTA |
| 主站 | `/app` | 插件市场、聊天页面、用户面板 |
| 后台 | `/admin` | 租户管理、插件管理、平台配置 |

### 6.2 共享基础设施

- 统一认证（JWT）
- 统一 MUI 主题系统
- 共享 Layout 组件（侧边栏、顶栏）
- 共享 API 客户端
- 前端插件系统

### 6.3 前端设计规范

**UI 设计时使用 `ui-ux-pro-max` skill 进行设计，确保高质量界面输出。**

---

## 7. LLM Chat 示例插件

作为第一个示例插件，提供：
- 对话列表管理
- 流式输出聊天界面
- 模型选择（OpenAI / Claude / 本地模型）
- 对话历史持久化

---

## 8. Git 分支策略

```
main
└── develop
    ├── feature/plugin-system
    ├── feature/user-auth
    ├── feature/llm-chat-plugin
    └── ...
```

---

## 9. 实施阶段

### Phase 1: 基础设施
- [ ] 项目脚手架搭建（Monorepo + Docker）
- [ ] 微服务框架搭建（go-zero 4个服务）
- [ ] 数据库 Schema 设计
- [ ] JWT 认证流程

### Phase 2: 核心功能
- [ ] 用户注册/登录 API
- [ ] 租户管理 API
- [ ] 插件系统后端
- [ ] 前端插件 SDK

### Phase 3: 前端三站
- [ ] 共享组件库搭建
- [ ] 落地页开发（调用 ui-ux-pro-max）
- [ ] 主站框架 + 插件市场
- [ ] 后台管理框架

### Phase 4: 示例插件
- [ ] LLM Chat 后端集成
- [ ] Chat UI 开发（调用 ui-ux-pro-max）
- [ ] 流式输出实现
- [ ] 对话历史

---

## 10. 设计决策

1. **Monorepo 管理前后端** — 方便代码共享和版本管理
2. **Schema 隔离而非行级隔离** — 更好的数据安全和性能
3. **动态插件加载** — 支持不停机更新插件
4. **gRPC 内部通信** — 高性能微服务间调用
5. **前端三站同一代码库** — 共享认证和组件，减少维护成本
