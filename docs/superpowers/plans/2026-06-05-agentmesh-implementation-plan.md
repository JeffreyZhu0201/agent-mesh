# AgentMesh 开发框架基座实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 搭建完整的 AgentMesh 开发框架底座，包含 go-zero 微服务 + React 前端三站点 + 插件系统 + LLM Chat 示例

**Architecture:** Monorepo 结构，后端 4 个 go-zero 微服务（api-gateway, user-svc, agent-svc, plugin-svc），前端 3 个站点（landing/main/admin）共享同一套组件库，插件系统支持动态加载

**Tech Stack:** go-zero, React 18, TypeScript, MUI, Eino, MySQL, Docker, Vite, Zustand

---

## Phase 1: 基础设施

### 1.1 项目脚手架搭建

**Files:**
- Create: `Makefile`
- Create: `.gitignore`
- Create: `docker-compose.yml`
- Create: `services/.gitkeep`
- Create: `web/.gitkeep`

- [ ] **Step 1: 创建 Makefile（常用开发命令）**

```makefile
# Makefile
.PHONY: help install dev build test clean docker-up docker-down

help:
	@echo "Available commands:"
	@echo "  make install    - Install dependencies"
	@echo "  make dev        - Start development mode"
	@echo "  make docker-up  - Start all services with Docker"
	@echo "  make docker-down - Stop all Docker services"
	@echo "  make test       - Run tests"
	@echo "  make build      - Build all services"

install:
	@echo "Installing dependencies..."
	cd services && go mod tidy
	cd web && npm install

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

test:
	cd services && go test ./...
	cd web && npm test

build:
	cd services && go build ./...
	cd web && npm run build
```

- [ ] **Step 2: 创建 .gitignore**

```gitignore
# Go
*.so
*.exe
*.exe~
*.dll
*.dylib
*.test
*.out
vendor/

# Node
node_modules/
dist/
.env
.env.local

# IDE
.idea/
.vscode/
*.swp
*.swo

# Docker
.docker/

# Misc
.DS_Store
Thumbs.db
```

- [ ] **Step 3: 创建 docker-compose.yml（MySQL + 所有后端服务）**

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: agentmesh-mysql
    environment:
      MYSQL_ROOT_PASSWORD: root123
      MYSQL_DATABASE: agentmesh_public
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    command: --default-authentication-plugin=mysql_native_password

  # API Gateway
  api-gateway:
    build: ./services/api-gateway
    container_name: agentmesh-api-gateway
    ports:
      - "8080:8080"
    depends_on:
      - user-svc
      - agent-svc
      - plugin-svc
    environment:
      - GRPC_USER_SVC=user-svc:8081
      - GRPC_AGENT_SVC=agent-svc:8082
      - GRPC_PLUGIN_SVC=plugin-svc:8083

  # User Service
  user-svc:
    build: ./services/user-svc
    container_name: agentmesh-user-svc
    ports:
      - "8081:8081"
    depends_on:
      - mysql
    environment:
      - MYSQL_HOST=mysql
      - MYSQL_PORT=3306
      - MYSQL_ROOT_PASSWORD=root123
      - MYSQL_DATABASE=agentmesh_public

  # Agent Service
  agent-svc:
    build: ./services/agent-svc
    container_name: agentmesh-agent-svc
    ports:
      - "8082:8082"
    depends_on:
      - mysql
    environment:
      - MYSQL_HOST=mysql
      - MYSQL_PORT=3306
      - MYSQL_ROOT_PASSWORD=root123
      - MYSQL_DATABASE=agentmesh_public

  # Plugin Service
  plugin-svc:
    build: ./services/plugin-svc
    container_name: agentmesh-plugin-svc
    ports:
      - "8083:8083"
    depends_on:
      - mysql
    environment:
      - MYSQL_HOST=mysql
      - MYSQL_PORT=3306
      - MYSQL_ROOT_PASSWORD=root123
      - MYSQL_DATABASE=agentmesh_public

volumes:
  mysql_data:
```

- [ ] **Step 4: 创建占位目录**

```bash
touch services/.gitkeep web/.gitkeep
```

- [ ] **Step 5: Commit**

```bash
git add Makefile .gitignore docker-compose.yml services/.gitkeep web/.gitkeep
git commit -m "chore: add project scaffold and docker-compose"
```

---

### 1.2 go-zero 微服务框架（4个服务）

**Files:**
- Create: `services/api-gateway`
- Create: `services/user-svc`
- Create: `services/agent-svc`
- Create: `services/plugin-svc`

#### 1.2.1 API Gateway (端口 8080)

- [ ] **Step 1: 创建 go-zero API gateway 项目结构**

```bash
cd services
mkdir -p api-gateway
cd api-gateway
go mod init agentmesh/api-gateway
```

- [ ] **Step 2: 创建 api gateway 定义文件**

Create: `services/api-gateway/api-gateway.yaml`

```yaml
Name: agentmesh-api-gateway
Host: 0.0.0.0
Port: 8080

Microservices:
  user:
    grpc: user-svc:8081
    rest: /api/user
  agent:
    grpc: agent-svc:8082
    rest: /api/agent
  plugin:
    grpc: plugin-svc:8083
    rest: /api/plugin

Middleware:
  - name: cors
  - name: tenant
    config:
      header: X-Tenant-ID
  - name: auth
    config:
      jwtSecret: ${JWT_SECRET:-agentmesh-secret-key}
```

- [ ] **Step 3: 创建 api gateway 入口文件**

Create: `services/api-gateway.go`

```go
package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/gateway"
)

var configFile = flag.String("f", "api-gateway.yaml", "config file")

func main() {
	flag.Parse()

	var gatewayConfig gateway.GatewayConfig
	conf.MustLoad(*configFile, &gatewayConfig)

	g, err := gateway.NewServer(gatewayConfig)
	if err != nil {
		panic(err)
	}

	fmt.Println("API Gateway starting on :8080")
	g.Start()
}
```

- [ ] **Step 4: Commit**

```bash
git add services/api-gateway/
git commit -m "feat: add api-gateway scaffold"
```

#### 1.2.2 User Service (端口 8081)

- [ ] **Step 1: 创建 user-svc 项目结构**

```bash
cd services
mkdir -p user-svc
cd user-svc
go mod init agentmesh/user-svc
```

- [ ] **Step 2: 创建 user-svc 配置文件**

Create: `services/user-svc/user-svc.yaml`

```yaml
Name: agentmesh-user-svc
Host: 0.0.0.0
Port: 8081

Mysql:
  Host: ${MYSQL_HOST:-localhost}
  Port: 3306
  User: root
  Password: ${MYSQL_ROOT_PASSWORD:-root123}
  Database: agentmesh_public
  MaxOpenConns: 100
  MaxIdleConns: 10

JWT:
  Secret: ${JWT_SECRET:-agentmesh-secret-key}
  ExpireHours: 24
```

- [ ] **Step 3: 创建 user 模型定义**

Create: `services/user-svc/model/user.go`

```go
package model

import (
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Email     string    `json:"email"`
	TenantID  int64     `json:"tenant_id"`
	Role      string    `json:"role"` // admin, member
	Status    int       `json:"status"` // 1=active, 0=inactive
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tenant struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Schema    string    `json:"schema"` // agentmesh_tenant_{id}
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewMysql creates mysql connection
func NewMysql(host, port, user, password, database string) sqlx.SqlConn {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user, password, host, port, database)
	return sqlx.NewMysql(dsn)
}
```

- [ ] **Step 4: 创建 user RPC 服务**

Create: `services/user-svc/internal/svc/servicecontext.go`

```go
package svc

import (
	"agentmesh/user-svc/internal/config"
	"agentmesh/user-svc/model"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	UserModel   model.UserModel
	TenantModel model.TenantModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := model.NewMysql(c.Mysql.Host, c.Mysql.Port, c.Mysql.User, c.Mysql.Password, c.Mysql.Database)
	return &ServiceContext{
		Config:      c,
		UserModel:   model.NewUserModel(sqlConn),
		TenantModel: model.NewTenantModel(sqlConn),
	}
}
```

- [ ] **Step 5: 创建 user 登录逻辑**

Create: `services/user-svc/internal/logic/loginlogic.go`

```go
package logic

import (
	"context"
	"errors"
	"time"

	"agentmesh/user-svc/internal/svc"
	"agentmesh/user-svc/model"
	"agentmesh/user-svc/pb/user"

	"github.com/golang-jwt/jwt/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	TenantID int64  `json:"tenant_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (l *LoginLogic) Login(in *user.LoginRequest) (*user.LoginResponse, error) {
	// Find user by username
	user, err := l.svcCtx.UserModel.FindOneByUsername(in.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password (plaintext for demo, use bcrypt in production)
	if user.Password != in.Password {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT
	expireTime := time.Now().Add(24 * time.Hour)
	claims := &JWTClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(l.svcCtx.Config.JWT.Secret))
	if err != nil {
		logx.Error("failed to generate token")
		return nil, err
	}

	return &user.LoginResponse{
		Token: tokenString,
		User: &user.UserInfo{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			TenantId: user.TenantID,
			Role:     user.Role,
		},
	}, nil
}
```

- [ ] **Step 6: Commit**

```bash
git add services/user-svc/
git commit -m "feat: add user-svc with login logic"
```

#### 1.2.3 Agent Service & Plugin Service

- [ ] **Step 1: 创建 agent-svc 基础结构（与 user-svc 类似）**

Create: `services/agent-svc/agent-svc.yaml`, `services/agent-svc.go`, `services/agent-svc/internal/svc/servicecontext.go`

- [ ] **Step 2: 创建 plugin-svc 基础结构**

Create: `services/plugin-svc/plugin-svc.yaml`, `services/plugin-svc.go`, `services/plugin-svc/internal/svc/servicecontext.go`

- [ ] **Step 3: Commit**

```bash
git add services/agent-svc/ services/plugin-svc/
git commit -m "feat: add agent-svc and plugin-svc scaffolds"
```

---

### 1.3 数据库 Schema 设计

**Files:**
- Create: `services/sql/init.sql`

- [ ] **Step 1: 创建数据库初始化 SQL**

Create: `services/sql/init.sql`

```sql
-- AgentMesh Database Schema Initialization

-- Public Schema (Tenant management)
CREATE DATABASE IF NOT EXISTS agentmesh_public;
USE agentmesh_public;

-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    schema_name VARCHAR(255) NOT NULL UNIQUE,
    status INT DEFAULT 1 COMMENT '1=active, 0=inactive',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Users table (public schema for tenant admins)
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    tenant_id BIGINT COMMENT 'NULL for platform admins',
    role VARCHAR(50) NOT NULL DEFAULT 'member' COMMENT 'admin, member',
    status INT DEFAULT 1 COMMENT '1=active, 0=inactive',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_tenant_id (tenant_id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Platform admins (no tenant)
INSERT INTO users (username, password, email, role) VALUES
('admin', 'admin123', 'admin@agentmesh.io', 'platform_admin');

-- Plugins table (available plugins)
CREATE TABLE IF NOT EXISTS plugins (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL UNIQUE,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    plugin_type VARCHAR(50) NOT NULL COMMENT 'backend, frontend, full',
    backend_path VARCHAR(500) COMMENT 'path to .so file',
    frontend_entry VARCHAR(500) COMMENT 'frontend module entry',
    config_schema JSON,
    status INT DEFAULT 1 COMMENT '1=active, 0=inactive',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_type (plugin_type),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tenant Plugins (plugin installations per tenant)
CREATE TABLE IF NOT EXISTS tenant_plugins (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id BIGINT NOT NULL,
    plugin_id BIGINT NOT NULL,
    config JSON COMMENT 'tenant-specific plugin config',
    status INT DEFAULT 1 COMMENT '1=active, 0=inactive',
    installed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (plugin_id) REFERENCES plugins(id) ON DELETE CASCADE,
    UNIQUE KEY uk_tenant_plugin (tenant_id, plugin_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tenant-specific schema template (apply per tenant)
-- CREATE DATABASE IF NOT EXISTS agentmesh_tenant_{tenant_id};
```

- [ ] **Step 2: 创建租户 Schema 初始化 SQL 模板**

Create: `services/sql/tenant_schema.sql`

```sql
-- Tenant-specific Schema Template
-- Apply for each new tenant: agentmesh_tenant_{tenant_id}

USE agentmesh_tenant_{tenant_id};

-- Conversations table
CREATE TABLE IF NOT EXISTS conversations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    title VARCHAR(255),
    model VARCHAR(100),
    status INT DEFAULT 1 COMMENT '1=active, 0=archived',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant_user (tenant_id, user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    conversation_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    role VARCHAR(50) NOT NULL COMMENT 'user, assistant, system',
    content TEXT,
    model VARCHAR(100) COMMENT 'model used for this message',
    token_count INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_conversation (conversation_id),
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Plugin configs per tenant
CREATE TABLE IF NOT EXISTS plugin_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    tenant_id BIGINT NOT NULL,
    plugin_id BIGINT NOT NULL,
    config_key VARCHAR(255) NOT NULL,
    config_value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tenant_plugin_key (tenant_id, plugin_id, config_key),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **Step 3: Commit**

```bash
git add services/sql/
git commit -m "feat: add database schema initialization scripts"
```

---

### 1.4 JWT 认证流程

**Files:**
- Modify: `services/api-gateway/middleware/auth.go`
- Create: `services/api-gateway/middleware/tenant.go`

- [ ] **Step 1: 创建 JWT 认证中间件**

Create: `services/api-gateway/middleware/auth.go`

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v2"
)

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	TenantID int64  `json:"tenant_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (m *AuthMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// Public paths that don't require auth
var PublicPaths = []string{
	"/api/user/login",
	"/api/user/register",
	"/api/user/health",
	"/",
	"/app",
}

func IsPublicPath(path string) bool {
	for _, p := range PublicPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: 创建租户中间件**

Create: `services/api-gateway/middleware/tenant.go`

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type TenantMiddleware struct{}

func NewTenantMiddleware() *TenantMiddleware {
	return &TenantMiddleware{}
}

func (m *TenantMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID != "" {
			c.Set("tenant_id", tenantID)
		}
		// For public paths, use default tenant (1)
		if !IsPublicPath(c.Request.URL.Path) {
			// Auth middleware will handle tenant validation
		}
		c.Next()
	}
}

// RequireTenant validates tenant context
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("tenant_id"); !exists {
			// Set default tenant for development
			c.Set("tenant_id", "1")
		}
		c.Next()
	}
}
```

- [ ] **Step 3: 创建路由配置**

Create: `services/api-gateway/routes/routes.go`

```go
package routes

import (
	"agentmesh/api-gateway/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, authMw *middleware.AuthMiddleware, tenantMw *middleware.TenantMiddleware) {
	// Public routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "AgentMesh API Gateway"})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API routes
	api := r.Group("/api")
	{
		// User routes (some public)
		user := api.Group("/user")
		user.Use(tenantMw.Handler())
		{
			user.POST("/login", handleLogin)
			user.POST("/register", handleRegister)
			user.GET("/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "healthy"})
			})

			// Protected routes
			protected := user.Group("")
			protected.Use(authMw.Handler())
			{
				protected.GET("/profile", handleGetProfile)
				protected.PUT("/profile", handleUpdateProfile)
			}
		}

		// Agent routes (protected)
		agent := api.Group("/agent")
		agent.Use(authMw.Handler(), tenantMw.Handler())
		{
			agent.POST("/chat", handleChat)
			agent.GET("/conversations", handleListConversations)
			agent.GET("/conversations/:id", handleGetConversation)
		}

		// Plugin routes (protected)
		plugin := api.Group("/plugin")
		plugin.Use(authMw.Handler(), tenantMw.Handler())
		{
			plugin.GET("/list", handleListPlugins)
			plugin.POST("/install", handleInstallPlugin)
			plugin.DELETE("/:id", handleUninstallPlugin)
		}
	}
}
```

- [ ] **Step 4: Commit**

```bash
git add services/api-gateway/middleware/ services/api-gateway/routes/
git commit -m "feat: add JWT auth and tenant middleware"
```

---

## Phase 2: 核心功能

### 2.1 用户注册/登录 API

**Files:**
- Modify: `services/user-svc/rpc/internal/logic/loginlogic.go`
- Create: `services/user-svc/rpc/internal/logic/registerlogic.go`
- Create: `services/user-svc/rpc/pb/user.proto`

- [ ] **Step 1: 定义 user.proto**

Create: `services/user-svc/rpc/pb/user.proto`

```protobuf
syntax = "proto3";

package user;

option go_package = "./pb";

service User {
    rpc Login(LoginRequest) returns (LoginResponse);
    rpc Register(RegisterRequest) returns (RegisterResponse);
    rpc GetUser(GetUserRequest) returns (UserInfo);
    rpc UpdateUser(UpdateUserRequest) returns (UserInfo);
}

message LoginRequest {
    string username = 1;
    string password = 2;
}

message LoginResponse {
    string token = 1;
    UserInfo user = 2;
}

message RegisterRequest {
    string username = 1;
    string password = 2;
    string email = 3;
    int64 tenant_id = 4;
}

message RegisterResponse {
    int64 user_id = 1;
    string message = 2;
}

message GetUserRequest {
    int64 user_id = 1;
}

message UpdateUserRequest {
    int64 user_id = 1;
    string email = 2;
}

message UserInfo {
    int64 id = 1;
    string username = 2;
    string email = 3;
    int64 tenant_id = 4;
    string role = 5;
}
```

- [ ] **Step 2: 创建 Register 逻辑**

Create: `services/user-svc/rpc/internal/logic/registerlogic.go`

```go
package logic

import (
	"context"

	"agentmesh/user-svc/internal/svc"
	"agentmesh/user-svc/model"
	"agentmesh/user-svc/pb/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(in *user.RegisterRequest) (*user.RegisterResponse, error) {
	// Check if username exists
	existing, err := l.svcCtx.UserModel.FindOneByUsername(in.Username)
	if existing != nil && err == nil {
		return nil, ErrUserAlreadyExists
	}

	newUser := &model.User{
		Username: in.Username,
		Password: in.Password, // TODO: hash password
		Email:    in.Email,
		TenantID: in.TenantId,
		Role:     "member",
		Status:   1,
	}

	id, err := l.svcCtx.UserModel.Insert(newUser)
	if err != nil {
		logx.Error("failed to create user", logx.Field("error", err))
		return nil, err
	}

	return &user.RegisterResponse{
		UserId:  id,
		Message: "user registered successfully",
	}, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add services/user-svc/rpc/
git commit -m "feat: add user registration API"
```

---

### 2.2 租户管理 API

**Files:**
- Create: `services/user-svc/rpc/internal/logic/tenantlogic.go`
- Create: `services/user-svc/model/tenantmodel.go`

- [ ] **Step 1: 创建 TenantModel**

Create: `services/user-svc/model/tenantmodel.go`

```go
package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type TenantModel struct {
	db sqlx.SqlConn
}

func NewTenantModel(db sqlx.SqlConn) *TenantModel {
	return &TenantModel{db: db}
}

type Tenant struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Schema    string    `json:"schema"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m *TenantModel) Insert(ctx context.Context, tenant *Tenant) (int64, error) {
	query := `INSERT INTO tenants (name, schema_name, status) VALUES (?, ?, ?)`
	result, err := m.db.ExecCtx(ctx, query, tenant.Name, tenant.Schema, tenant.Status)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (m *TenantModel) FindOneByID(ctx context.Context, id int64) (*Tenant, error) {
	query := `SELECT id, name, schema_name, status, created_at, updated_at FROM tenants WHERE id = ?`
	var tenant Tenant
	err := m.db.QueryRowCtx(ctx, &tenant, query, id)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (m *TenantModel) CreateTenantSchema(ctx context.Context, tenantID int64) error {
	schemaName := fmt.Sprintf("agentmesh_tenant_%d", tenantID)

	// Create database schema
	_, err := m.db.ExecCtx(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", schemaName))
	if err != nil {
		return err
	}

	// Initialize tenant schema with tables
	conn, err := sql.Open("mysql", fmt.Sprintf("agentmesh:%s@tcp(%s)/?charset=utf8mb4&parseTime=true&loc=Local",
		os.Getenv("MYSQL_ROOT_PASSWORD"), os.Getenv("MYSQL_HOST")))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create tables in tenant schema
	tables := []string{
		fmt.Sprintf("USE %s", schemaName),
		`CREATE TABLE IF NOT EXISTS conversations (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			tenant_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			title VARCHAR(255),
			model VARCHAR(100),
			status INT DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			conversation_id BIGINT NOT NULL,
			tenant_id BIGINT NOT NULL,
			role VARCHAR(50) NOT NULL,
			content TEXT,
			model VARCHAR(100),
			token_count INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, stmt := range tables {
		_, err = conn.ExecContext(ctx, stmt)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	return nil
}
```

- [ ] **Step 2: 创建租户管理 Logic**

Create: `services/user-svc/rpc/internal/logic/tenantlogic.go`

```go
package logic

import (
	"context"
	"fmt"

	"agentmesh/user-svc/internal/svc"
	"agentmesh/user-svc/model"
	"agentmesh/user-svc/pb/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type TenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantLogic {
	return &TenantLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *TenantLogic) CreateTenant(in *user.CreateTenantRequest) (*user.CreateTenantResponse, error) {
	tenant := &model.Tenant{
		Name:   in.Name,
		Schema: fmt.Sprintf("agentmesh_tenant_%d", in.Name), // Simplified
		Status: 1,
	}

	id, err := l.svcCtx.TenantModel.Insert(l.ctx, tenant)
	if err != nil {
		logx.Error("failed to create tenant", logx.Field("error", err))
		return nil, err
	}

	// Create tenant schema
	if err := l.svcCtx.TenantModel.CreateTenantSchema(l.ctx, id); err != nil {
		logx.Error("failed to create tenant schema", logx.Field("error", err))
		// Continue anyway, schema can be created later
	}

	return &user.CreateTenantResponse{
		TenantId: id,
		Message:  "tenant created successfully",
	}, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add services/user-svc/model/tenantmodel.go services/user-svc/rpc/internal/logic/tenantlogic.go
git commit -m "feat: add tenant management API"
```

---

### 2.3 插件系统后端

**Files:**
- Create: `services/plugin-svc/plugin/plugin.go`
- Create: `services/plugin-svc/internal/logic/pluginlogic.go`

- [ ] **Step 1: 定义插件接口**

Create: `services/plugin-svc/plugin/plugin.go`

```go
package plugin

import (
	"context"
)

// PluginMetadata 插件元信息
type PluginMetadata struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Type        string `json:"type"` // backend, frontend, full
	Author      string `json:"author"`
}

// MenuItem 菜单项配置
type MenuItem struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Icon     string `json:"icon"`
	Path     string `json:"path"`
	Parent   string `json:"parent,omitempty"`
	Order    int    `json:"order"`
}

// Plugin 插件主接口
type Plugin interface {
	// Metadata 返回插件元信息
	Metadata() *PluginMetadata

	// Init 初始化插件
	Init(ctx context.Context) error

	// GetMenuItems 返回插件的菜单项
	GetMenuItems() []*MenuItem

	// Close 关闭插件，释放资源
	Close() error
}

// PluginManager 插件管理器
type PluginManager struct {
	plugins map[string]Plugin
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
	}
}

func (pm *PluginManager) Register(name string, p Plugin) error {
	if _, exists := pm.plugins[name]; exists {
		return ErrPluginAlreadyRegistered
	}
	pm.plugins[name] = p
	return nil
}

func (pm *PluginManager) Unregister(name string) error {
	p, exists := pm.plugins[name]
	if !exists {
		return ErrPluginNotFound
	}
	if err := p.Close(); err != nil {
		return err
	}
	delete(pm.plugins, name)
	return nil
}

func (pm *PluginManager) Get(name string) (Plugin, error) {
	p, exists := pm.plugins[name]
	if !exists {
		return nil, ErrPluginNotFound
	}
	return p, nil
}

func (pm *PluginManager) List() []*PluginMetadata {
	result := make([]*PluginMetadata, 0, len(pm.plugins))
	for _, p := range pm.plugins {
		result = append(result, p.Metadata())
	}
	return result
}
```

- [ ] **Step 2: 创建插件服务 Logic**

Create: `services/plugin-svc/internal/logic/pluginlogic.go`

```go
package logic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"

	"agentmesh/plugin-svc/plugin"
	"agentmesh/plugin-svc/pb/plugin"

	"github.com/zeromicro/go-zero/core/logx"
)

type PluginLogic struct {
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	manager   *plugin.PluginManager
	mu        sync.RWMutex
	pluginDir string
}

func NewPluginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PluginLogic {
	return &PluginLogic{
		ctx:       ctx,
		svcCtx:    svcCtx,
		manager:   plugin.NewPluginManager(),
		pluginDir: "/app/plugins",
	}
}

func (l *PluginLogic) LoadPlugin(in *plugin.LoadPluginRequest) (*plugin.LoadPluginResponse, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	pluginPath := filepath.Join(l.pluginDir, in.PluginName+".so")

	// Open plugin file
	p, err := plugin.Open(pluginPath)
	if err != nil {
		logx.Errorf("failed to open plugin: %s", err)
		return nil, fmt.Errorf("failed to load plugin: %w", err)
	}

	// Lookup plugin symbol
	symbol, err := p.Lookup("PluginInstance")
	if err != nil {
		return nil, fmt.Errorf("plugin symbol not found: %w", err)
	}

	// Assert to Plugin interface
	pluginInstance, ok := symbol.(plugin.Plugin)
	if !ok {
		return nil, fmt.Errorf("invalid plugin type")
	}

	// Initialize plugin
	if err := pluginInstance.Init(l.ctx); err != nil {
		return nil, fmt.Errorf("failed to init plugin: %w", err)
	}

	// Register plugin
	if err := l.manager.Register(pluginInstance.Metadata().Name, pluginInstance); err != nil {
		return nil, fmt.Errorf("failed to register plugin: %w", err)
	}

	return &plugin.LoadPluginResponse{
		Success: true,
		Message: fmt.Sprintf("plugin %s loaded successfully", in.PluginName),
	}, nil
}

func (l *PluginLogic) UnloadPlugin(in *plugin.UnloadPluginRequest) (*plugin.UnloadPluginResponse, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.manager.Unregister(in.PluginName); err != nil {
		return nil, err
	}

	return &plugin.UnloadPluginResponse{
		Success: true,
		Message: fmt.Sprintf("plugin %s unloaded", in.PluginName),
	}, nil
}

func (l *PluginLogic) ListPlugins(in *plugin.ListPluginsRequest) (*plugin.ListPluginsResponse, error) {
	plugins := l.manager.List()
	return &plugin.ListPluginsResponse{
		Plugins: plugins,
	}, nil
}

func (l *PluginLogic) GetMenuItems(in *plugin.GetMenuItemsRequest) (*plugin.GetMenuItemsResponse, error) {
	p, err := l.manager.Get(in.PluginName)
	if err != nil {
		return nil, err
	}

	return &plugin.GetMenuItemsResponse{
		Items: p.GetMenuItems(),
	}, nil
}
```

- [ ] **Step 3: 创建 Chat 插件示例**

Create: `services/plugins/chat/main.go`

```go
package main

import (
	"context"
	"fmt"

	"agentmesh/plugin-svc/plugin"
)

type ChatPlugin struct{}

func (c *ChatPlugin) Metadata() *plugin.PluginMetadata {
	return &plugin.PluginMetadata{
		Name:        "chat",
		Version:     "1.0.0",
		Description: "LLM Chat Plugin - Provides chat interface with streaming support",
		Type:        "full",
		Author:      "AgentMesh",
	}
}

func (c *ChatPlugin) Init(ctx context.Context) error {
	fmt.Println("ChatPlugin initialized")
	return nil
}

func (c *ChatPlugin) GetMenuItems() []*plugin.MenuItem {
	return []*plugin.MenuItem{
		{
			Key:   "chat",
			Label: "Chat",
			Icon:  "chat",
			Path:  "/app/chat",
			Order: 1,
		},
		{
			Key:   "chat-history",
			Label: "History",
			Icon:  "history",
			Path:  "/app/chat/history",
			Parent: "chat",
			Order: 1,
		},
	}
}

func (c *ChatPlugin) Close() error {
	return nil
}

// PluginInstance is the entry point for the plugin system
var PluginInstance plugin.Plugin = &ChatPlugin{}
```

- [ ] **Step 4: Commit**

```bash
git add services/plugin-svc/plugin/ services/plugin-svc/internal/logic/ services/plugins/chat/
git commit -m "feat: add plugin system backend with chat plugin"
```

---

### 2.4 前端插件 SDK

**Files:**
- Create: `web/packages/plugin-sdk/src/index.ts`
- Create: `web/packages/plugin-sdk/src/hooks.ts`
- Create: `web/packages/plugin-sdk/src/types.ts`

- [ ] **Step 1: 创建插件 SDK 类型定义**

Create: `web/packages/plugin-sdk/src/types.ts`

```typescript
export interface PluginMetadata {
  name: string;
  version: string;
  description: string;
  type: 'backend' | 'frontend' | 'full';
  author: string;
}

export interface MenuItem {
  key: string;
  label: string;
  icon?: string;
  path: string;
  parent?: string;
  order: number;
}

export interface Plugin {
  metadata: PluginMetadata;
  menuItems: MenuItem[];
  routes?: RouteConfig[];
  component?: React.ComponentType;
}

export interface RouteConfig {
  path: string;
  component: React.ComponentType;
  exact?: boolean;
  meta?: Record<string, unknown>;
}

export interface PluginConfig {
  name: string;
  enabled: boolean;
  config?: Record<string, unknown>;
}

export interface PluginContext {
  tenantId: string;
  userId: string;
  token: string;
  apiBase: string;
}
```

- [ ] **Step 2: 创建插件注册 Hooks**

Create: `web/packages/plugin-sdk/src/hooks.ts`

```typescript
import { useState, useCallback, useEffect } from 'react';
import type { Plugin, MenuItem, RouteConfig, PluginContext } from './types';

type PluginCallback = (plugin: Plugin) => void;

const registeredPlugins = new Map<string, Plugin>();
const menuItems: MenuItem[] = [];
const routeConfigs: RouteConfig[] = [];
const listeners: PluginCallback[] = [];

export function registerPlugin(plugin: Plugin): void {
  if (registeredPlugins.has(plugin.metadata.name)) {
    console.warn(`Plugin ${plugin.metadata.name} already registered`);
    return;
  }

  registeredPlugins.set(plugin.metadata.name, plugin);

  // Register menu items
  plugin.menuItems.forEach(item => {
    const existingIndex = menuItems.findIndex(m => m.key === item.key);
    if (existingIndex >= 0) {
      menuItems[existingIndex] = item;
    } else {
      menuItems.push(item);
    }
  });

  // Register routes
  if (plugin.routes) {
    routeConfigs.push(...plugin.routes);
  }

  // Notify listeners
  listeners.forEach(cb => cb(plugin));

  console.log(`Plugin registered: ${plugin.metadata.name}`);
}

export function unregisterPlugin(name: string): void {
  const plugin = registeredPlugins.get(name);
  if (!plugin) return;

  registeredPlugins.delete(name);

  // Remove menu items
  plugin.menuItems.forEach(item => {
    const index = menuItems.findIndex(m => m.key === item.key);
    if (index >= 0) menuItems.splice(index, 1);
  });

  // Remove routes
  plugin.routes?.forEach(route => {
    const index = routeConfigs.findIndex(r => r.path === route.path);
    if (index >= 0) routeConfigs.splice(index, 1);
  });

  console.log(`Plugin unregistered: ${name}`);
}

export function getRegisteredPlugins(): Plugin[] {
  return Array.from(registeredPlugins.values());
}

export function getMenuItems(): MenuItem[] {
  return [...menuItems].sort((a, b) => a.order - b.order);
}

export function getRouteConfigs(): RouteConfig[] {
  return [...routeConfigs];
}

export function usePlugin(name: string): Plugin | undefined {
  return registeredPlugins.get(name);
}

export function usePlugins(): Plugin[] {
  const [plugins, setPlugins] = useState<Plugin[]>([]);

  useEffect(() => {
    setPlugins(getRegisteredPlugins());

    const listener: PluginCallback = (plugin) => {
      setPlugins(getRegisteredPlugins());
    };
    listeners.push(listener);

    return () => {
      const index = listeners.indexOf(listener);
      if (index >= 0) listeners.splice(index, 1);
    };
  }, []);

  return plugins;
}

export function useMenuItems(): MenuItem[] {
  const [items, setItems] = useState<MenuItem[]>([]);

  useEffect(() => {
    setItems(getMenuItems());

    const listener: PluginCallback = () => {
      setItems(getMenuItems());
    };
    listeners.push(listener);

    return () => {
      const index = listeners.indexOf(listener);
      if (index >= 0) listeners.splice(index, 1);
    };
  }, []);

  return items;
}
```

- [ ] **Step 3: 创建插件 SDK 入口**

Create: `web/packages/plugin-sdk/src/index.ts`

```typescript
export * from './types';
export * from './hooks';

// Plugin context for API calls
let pluginContext: PluginContext | null = null;

export function setPluginContext(ctx: PluginContext): void {
  pluginContext = ctx;
}

export function getPluginContext(): PluginContext | null {
  return pluginContext;
}

// Helper function for plugins to make API calls
export async function pluginApiCall<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  if (!pluginContext) {
    throw new Error('Plugin context not set');
  }

  const url = `${pluginContext.apiBase}${endpoint}`;
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${pluginContext.token}`,
    'X-Tenant-ID': pluginContext.tenantId,
  };

  const response = await fetch(url, {
    ...options,
    headers: { ...headers, ...options?.headers },
  });

  if (!response.ok) {
    throw new Error(`API call failed: ${response.statusText}`);
  }

  return response.json();
}

// Plugin SDK version
export const SDK_VERSION = '1.0.0';
```

- [ ] **Step 4: 创建 SDK package.json**

Create: `web/packages/plugin-sdk/package.json`

```json
{
  "name": "@agentmesh/plugin-sdk",
  "version": "1.0.0",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "dev": "tsc --watch"
  },
  "dependencies": {
    "react": "^18.2.0"
  },
  "peerDependencies": {
    "react": "^18.2.0"
  }
}
```

- [ ] **Step 5: Commit**

```bash
git add web/packages/plugin-sdk/
git commit -m "feat: add frontend plugin SDK"
```

---

## Phase 3: 前端三站

### 3.1 项目脚手架

**Files:**
- Create: `web/package.json`
- Create: `web/tsconfig.json`
- Create: `web/vite.config.ts`
- Create: `web/apps/landing/src/main.tsx`
- Create: `web/apps/main/src/main.tsx`
- Create: `web/apps/admin/src/main.tsx`

- [ ] **Step 1: 创建前端 Monorepo 根配置**

Create: `web/package.json`

```json
{
  "name": "agentmesh-web",
  "version": "1.0.0",
  "private": true,
  "workspaces": [
    "apps/*",
    "packages/*"
  ],
  "scripts": {
    "dev": "turbo run dev",
    "build": "turbo run build",
    "test": "turbo run test",
    "lint": "turbo run lint",
    "clean": "turbo run clean"
  },
  "devDependencies": {
    "turbo": "^1.13.0",
    "typescript": "^5.4.0"
  }
}
```

- [ ] **Step 2: 创建 Vite 配置**

Create: `web/vite.config.ts`

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@agentmesh/ui': path.resolve(__dirname, '../packages/ui/src'),
      '@agentmesh/hooks': path.resolve(__dirname, '../packages/hooks/src'),
      '@agentmesh/stores': path.resolve(__dirname, '../packages/stores/src'),
      '@agentmesh/api': path.resolve(__dirname, '../packages/api/src'),
      '@agentmesh/plugin-sdk': path.resolve(__dirname, '../packages/plugin-sdk/src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
```

- [ ] **Step 3: 创建共享类型**

Create: `web/packages/types/src/index.ts`

```typescript
export interface User {
  id: number;
  username: string;
  email: string;
  tenantId: number;
  role: string;
}

export interface Tenant {
  id: number;
  name: string;
  schema: string;
  status: number;
}

export interface Conversation {
  id: number;
  tenantId: number;
  userId: number;
  title: string;
  model: string;
  status: number;
  createdAt: string;
  updatedAt: string;
}

export interface Message {
  id: number;
  conversationId: number;
  tenantId: number;
  role: 'user' | 'assistant' | 'system';
  content: string;
  model?: string;
  tokenCount?: number;
  createdAt: string;
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
  code: number;
}
```

- [ ] **Step 4: Commit**

```bash
git add web/
git commit -m "feat: add frontend monorepo scaffold"
```

---

### 3.2 共享组件库

**Files:**
- Create: `web/packages/ui/src/theme/index.ts`
- Create: `web/packages/ui/src/components/Layout/index.tsx`
- Create: `web/packages/ui/src/components/Button/index.tsx`

- [ ] **Step 1: 创建 MUI 主题配置**

Create: `web/packages/ui/src/theme/index.ts`

```typescript
import { createTheme, ThemeOptions } from '@mui/material/styles';

const themeOptions: ThemeOptions = {
  palette: {
    mode: 'light',
    primary: {
      main: '#6366F1', // Indigo
      light: '#818CF8',
      dark: '#4F46E5',
    },
    secondary: {
      main: '#EC4899', // Pink
      light: '#F472B6',
      dark: '#DB2777',
    },
    background: {
      default: '#FAFAFA',
      paper: '#FFFFFF',
    },
    text: {
      primary: '#1F2937',
      secondary: '#6B7280',
    },
  },
  typography: {
    fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
    h1: {
      fontSize: '2.5rem',
      fontWeight: 700,
    },
    h2: {
      fontSize: '2rem',
      fontWeight: 600,
    },
    h3: {
      fontSize: '1.5rem',
      fontWeight: 600,
    },
    body1: {
      fontSize: '1rem',
    },
    body2: {
      fontSize: '0.875rem',
    },
  },
  shape: {
    borderRadius: 8,
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 600,
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          boxShadow: '0 1px 3px 0 rgb(0 0 0 / 0.1)',
        },
      },
    },
  },
};

export const theme = createTheme(themeOptions);
```

- [ ] **Step 2: 创建 Layout 组件**

Create: `web/packages/ui/src/components/Layout/index.tsx`

```typescript
import React from 'react';
import { Outlet } from 'react-router-dom';
import {
  Box,
  Drawer,
  AppBar,
  Toolbar,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Typography,
  IconButton,
  Avatar,
  Menu,
  MenuItem,
} from '@mui/material';
import {
  Menu as MenuIcon,
  Dashboard,
  Chat,
  Settings,
  Logout,
  Person,
} from '@mui/icons-material';

const DRAWER_WIDTH = 240;

interface LayoutProps {
  menuItems?: Array<{
    label: string;
    icon: React.ReactNode;
    path: string;
  }>;
}

export function Layout({ menuItems }: LayoutProps) {
  const [mobileOpen, setMobileOpen] = React.useState(false);
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const handleMenu = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const defaultMenuItems = [
    { label: 'Dashboard', icon: <Dashboard />, path: '/app' },
    { label: 'Chat', icon: <Chat />, path: '/app/chat' },
    { label: 'Settings', icon: <Settings />, path: '/app/settings' },
  ];

  const items = menuItems || defaultMenuItems;

  const drawer = (
    <Box>
      <Toolbar>
        <Typography variant="h6" noWrap component="div">
          AgentMesh
        </Typography>
      </Toolbar>
      <List>
        {items.map((item) => (
          <ListItem key={item.path} disablePadding>
            <ListItemButton component="a" href={item.path}>
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.label} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
    </Box>
  );

  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar
        position="fixed"
        sx={{
          width: { sm: `calc(100% - ${DRAWER_WIDTH}px)` },
          ml: { sm: `${DRAWER_WIDTH}px` },
        }}
      >
        <Toolbar>
          <IconButton
            color="inherit"
            edge="start"
            onClick={handleDrawerToggle}
            sx={{ mr: 2, display: { sm: 'none' } }}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" noWrap component="div" sx={{ flexGrow: 1 }}>
            App
          </Typography>
          <IconButton onClick={handleMenu}>
            <Avatar sx={{ width: 32, height: 32 }}>U</Avatar>
          </IconButton>
          <Menu
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            onClose={handleClose}
          >
            <MenuItem onClick={handleClose}>
              <Person sx={{ mr: 1 }} /> Profile
            </MenuItem>
            <MenuItem onClick={handleClose}>
              <Logout sx={{ mr: 1 }} /> Logout
            </MenuItem>
          </Menu>
        </Toolbar>
      </AppBar>
      <Box
        component="nav"
        sx={{ width: { sm: DRAWER_WIDTH }, flexShrink: { sm: 0 } }}
      >
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={handleDrawerToggle}
          sx={{
            display: { xs: 'block', sm: 'none' },
            '& .MuiDrawer-paper': { width: DRAWER_WIDTH },
          }}
        >
          {drawer}
        </Drawer>
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: 'none', sm: 'block' },
            '& .MuiDrawer-paper': { width: DRAWER_WIDTH },
          }}
          open
        >
          {drawer}
        </Drawer>
      </Box>
      <Box
        component="main"
        sx={{ flexGrow: 1, p: 3, width: { sm: `calc(100% - ${DRAWER_WIDTH}px)` } }}
      >
        <Toolbar />
        <Outlet />
      </Box>
    </Box>
  );
}
```

- [ ] **Step 3: 创建 UI 组件库入口**

Create: `web/packages/ui/src/index.ts`

```typescript
export * from './theme';
export * from './components/Layout';
```

- [ ] **Step 4: Commit**

```bash
git add web/packages/ui/
git commit -m "feat: add shared UI component library"
```

---

### 3.3 落地页（调用 ui-ux-pro-max）

**Files:**
- Create: `web/apps/landing/src/App.tsx`
- Create: `web/apps/landing/src/pages/Home.tsx`

- [ ] **Step 1: 使用 ui-ux-pro-max 设计落地页**

**需要先调用 `ui-ux-pro-max:ui-ux-pro-max` skill 进行页面设计**

以下为设计后的落地页结构：

- [ ] **Step 2: 创建落地页 App 组件**

Create: `web/apps/landing/src/App.tsx`

```typescript
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { theme } from '@agentmesh/ui';
import { Home } from './pages/Home';

export default function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Home />
    </ThemeProvider>
  );
}
```

- [ ] **Step 3: 创建落地页主组件**

Create: `web/apps/landing/src/pages/Home.tsx`

```typescript
import React from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  Container,
  Box,
  Grid,
  Card,
  CardContent,
  CardMedia,
  Chip,
} from '@mui/material';
import {
  AutoAwesome,
  Security,
  Extension,
  Speed,
  ArrowForward,
  Star,
} from '@mui/icons-material';
import { styled } from '@mui/material/styles';

// Hero Section
const HeroSection = styled(Box)(({ theme }) => ({
  minHeight: '100vh',
  display: 'flex',
  alignItems: 'center',
  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
  color: 'white',
  padding: theme.spacing(8, 0),
}));

// Feature Card
const FeatureCard = styled(Card)(({ theme }) => ({
  height: '100%',
  transition: 'transform 0.3s',
  '&:hover': {
    transform: 'translateY(-8px)',
    boxShadow: theme.shadows[8],
  },
}));

const features = [
  {
    icon: <AutoAwesome sx={{ fontSize: 48, color: '#6366F1' }} />,
    title: 'AI Native',
    description: 'Built-in Eino integration for seamless AI agent orchestration',
  },
  {
    icon: <Extension sx={{ fontSize: 48, color: '#EC4899' }} />,
    title: 'Plugin System',
    description: 'Dynamic plugin loading for unlimited extensibility',
  },
  {
    icon: <Security sx={{ fontSize: 48, color: '#10B981' }} />,
    title: 'SaaS Ready',
    description: 'Multi-tenant architecture with schema-level isolation',
  },
  {
    icon: <Speed sx={{ fontSize: 48, color: '#F59E0B' }} />,
    title: 'High Performance',
    description: 'Go-zero microservices with gRPC communication',
  },
];

const testimonials = [
  {
    name: 'Sarah Chen',
    role: 'CTO at TechCorp',
    content: 'AgentMesh accelerated our AI product development by 3x.',
    rating: 5,
  },
  {
    name: 'Michael Park',
    role: 'Lead Developer at StartupXYZ',
    content: 'The plugin system is incredibly flexible. We built our entire product on it.',
    rating: 5,
  },
  {
    name: 'Emily Johnson',
    role: 'Product Manager at AI Labs',
    content: 'Finally, a framework that makes multi-tenant AI apps easy.',
    rating: 5,
  },
];

export function Home() {
  return (
    <Box>
      {/* Navigation */}
      <AppBar position="fixed" color="transparent" sx={{ backdropFilter: 'blur(10px)' }}>
        <Container maxWidth="lg">
          <Toolbar disableGutters>
            <Typography variant="h6" sx={{ flexGrow: 1, fontWeight: 700 }}>
              AgentMesh
            </Typography>
            <Button color="inherit" href="/app">Login</Button>
            <Button variant="contained" color="secondary" sx={{ ml: 2 }}>
              Get Started
            </Button>
          </Toolbar>
        </Container>
      </AppBar>

      {/* Hero Section */}
      <HeroSection>
        <Container maxWidth="lg">
          <Grid container spacing={6} alignItems="center">
            <Grid item xs={12} md={6}>
              <Chip label="Now in Beta" color="secondary" sx={{ mb: 2 }} />
              <Typography variant="h2" sx={{ fontWeight: 700, mb: 2 }}>
                Build AI Applications
                <br />
                <Box component="span" sx={{ color: '#FCD34D' }}>
                  10x Faster
                </Box>
              </Typography>
              <Typography variant="h5" sx={{ mb: 4, opacity: 0.9 }}>
                A comprehensive SaaS framework with go-zero microservices,
                React frontend, and Eino-powered AI agents.
              </Typography>
              <Box sx={{ display: 'flex', gap: 2 }}>
                <Button
                  variant="contained"
                  size="large"
                  endIcon={<ArrowForward />}
                  sx={{ px: 4, py: 1.5 }}
                >
                  Start Building
                </Button>
                <Button
                  variant="outlined"
                  size="large"
                  sx={{ px: 4, py: 1.5, color: 'white', borderColor: 'white' }}
                >
                  View Demo
                </Button>
              </Box>
            </Grid>
            <Grid item xs={12} md={6}>
              <Box
                component="img"
                src="/hero-illustration.svg"
                alt="AgentMesh Architecture"
                sx={{ width: '100%', borderRadius: 4 }}
              />
            </Grid>
          </Grid>
        </Container>
      </HeroSection>

      {/* Features Section */}
      <Box sx={{ py: 12, bgcolor: 'grey.50' }}>
        <Container maxWidth="lg">
          <Typography variant="h3" align="center" sx={{ mb: 2, fontWeight: 700 }}>
            Everything You Need
          </Typography>
          <Typography variant="h6" align="center" color="text.secondary" sx={{ mb: 8 }}>
            A complete framework for building production-ready AI SaaS applications
          </Typography>
          <Grid container spacing={4}>
            {features.map((feature, index) => (
              <Grid item xs={12} sm={6} md={3} key={index}>
                <FeatureCard>
                  <CardContent sx={{ textAlign: 'center', p: 4 }}>
                    <Box sx={{ mb: 2 }}>{feature.icon}</Box>
                    <Typography variant="h6" sx={{ mb: 1, fontWeight: 600 }}>
                      {feature.title}
                    </Typography>
                    <Typography color="text.secondary">
                      {feature.description}
                    </Typography>
                  </CardContent>
                </FeatureCard>
              </Grid>
            ))}
          </Grid>
        </Container>
      </Box>

      {/* Testimonials */}
      <Box sx={{ py: 12 }}>
        <Container maxWidth="lg">
          <Typography variant="h3" align="center" sx={{ mb: 8, fontWeight: 700 }}>
            Loved by Developers
          </Typography>
          <Grid container spacing={4}>
            {testimonials.map((testimonial, index) => (
              <Grid item xs={12} md={4} key={index}>
                <Card sx={{ height: '100%', p: 2 }}>
                  <CardContent>
                    <Box sx={{ display: 'flex', mb: 2 }}>
                      {Array.from({ length: testimonial.rating }).map((_, i) => (
                        <Star key={i} sx={{ color: '#FCD34D' }} />
                      ))}
                    </Box>
                    <Typography sx={{ mb: 2, fontStyle: 'italic' }}>
                      "{testimonial.content}"
                    </Typography>
                    <Typography variant="subtitle2" sx={{ fontWeight: 600 }}>
                      {testimonial.name}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {testimonial.role}
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>
            ))}
          </Grid>
        </Container>
      </Box>

      {/* CTA Section */}
      <Box sx={{ py: 12, bgcolor: 'primary.main', color: 'white' }}>
        <Container maxWidth="md" align="center">
          <Typography variant="h4" sx={{ mb: 2, fontWeight: 700 }}>
            Ready to Build?
          </Typography>
          <Typography variant="h6" sx={{ mb: 4, opacity: 0.9 }}>
            Start your AI application development today
          </Typography>
          <Button
            variant="contained"
            color="secondary"
            size="large"
            endIcon={<ArrowForward />}
            sx={{ px: 6, py: 2 }}
          >
            Get Started Free
          </Button>
        </Container>
      </Box>

      {/* Footer */}
      <Box sx={{ py: 4, bgcolor: 'grey.900', color: 'grey.400' }}>
        <Container maxWidth="lg">
          <Typography align="center">
            © 2024 AgentMesh. All rights reserved.
          </Typography>
        </Container>
      </Box>
    </Box>
  );
}
```

- [ ] **Step 4: Commit**

```bash
git add web/apps/landing/
git commit -m "feat: add landing page"
```

---

### 3.4 主站框架

**Files:**
- Create: `web/apps/main/src/App.tsx`
- Create: `web/apps/main/src/pages/Dashboard.tsx`
- Create: `web/apps/main/src/pages/PluginMarket.tsx`

- [ ] **Step 1: 创建主站 App 组件**

Create: `web/apps/main/src/App.tsx`

```typescript
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { Box } from '@mui/material';
import { theme } from '@agentmesh/ui';
import { Layout } from '@agentmesh/ui';
import { Dashboard } from './pages/Dashboard';
import { PluginMarket } from './pages/PluginMarket';
import { Chat } from './pages/Chat';
import { getMenuItems, getRouteConfigs } from '@agentmesh/plugin-sdk';

function App() {
  const menuItems = getMenuItems();
  const routeConfigs = getRouteConfigs();

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout menuItems={menuItems as any} />}>
            <Route index element={<Navigate to="/app" replace />} />
            <Route path="/app" element={<Dashboard />} />
            <Route path="/app/chat" element={<Chat />} />
            <Route path="/app/chat/:id" element={<Chat />} />
            <Route path="/app/plugins" element={<PluginMarket />} />
            {/* Dynamic plugin routes */}
            {routeConfigs.map((route) => (
              <Route
                key={route.path}
                path={route.path}
                element={<route.component />}
              />
            ))}
          </Route>
        </Routes>
      </BrowserRouter>
    </ThemeProvider>
  );
}

export default App;
```

- [ ] **Step 2: 创建 Dashboard 页面**

Create: `web/apps/main/src/pages/Dashboard.tsx`

```typescript
import {
  Box,
  Grid,
  Card,
  CardContent,
  Typography,
  Avatar,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
} from '@mui/material';
import {
  Chat,
  Extension,
  TrendingUp,
  People,
} from '@mui/icons-material';

const stats = [
  { label: 'Total Chats', value: '1,234', icon: <Chat />, color: '#6366F1' },
  { label: 'Active Plugins', value: '12', icon: <Extension />, color: '#EC4899' },
  { label: 'Messages', value: '45.2K', icon: <TrendingUp />, color: '#10B981' },
  { label: 'Team Members', value: '8', icon: <People />, color: '#F59E0B' },
];

const recentConversations = [
  { id: 1, title: 'Project Discussion', messages: 42, updated: '2 min ago' },
  { id: 2, title: 'Code Review', messages: 28, updated: '15 min ago' },
  { id: 3, title: 'Design Feedback', messages: 15, updated: '1 hour ago' },
];

export function Dashboard() {
  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 4, fontWeight: 700 }}>
        Dashboard
      </Typography>

      {/* Stats Grid */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        {stats.map((stat) => (
          <Grid item xs={12} sm={6} md={3} key={stat.label}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <Avatar
                    sx={{
                      bgcolor: `${stat.color}20`,
                      color: stat.color,
                      width: 48,
                      height: 48,
                    }}
                  >
                    {stat.icon}
                  </Avatar>
                </Box>
                <Typography variant="h4" sx={{ fontWeight: 700 }}>
                  {stat.value}
                </Typography>
                <Typography color="text.secondary">{stat.label}</Typography>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Recent Conversations */}
      <Card>
        <CardContent>
          <Typography variant="h6" sx={{ mb: 2, fontWeight: 600 }}>
            Recent Conversations
          </Typography>
          <List>
            {recentConversations.map((conv) => (
              <ListItem
                key={conv.id}
                sx={{
                  borderBottom: '1px solid',
                  borderColor: 'divider',
                  '&:last-child': { borderBottom: 'none' },
                }}
              >
                <ListItemAvatar>
                  <Avatar sx={{ bgcolor: 'primary.main' }}>
                    <Chat />
                  </Avatar>
                </ListItemAvatar>
                <ListItemText
                  primary={conv.title}
                  secondary={`${conv.messages} messages • ${conv.updated}`}
                />
              </ListItem>
            ))}
          </List>
        </CardContent>
      </Card>
    </Box>
  );
}
```

- [ ] **Step 3: 创建插件市场页面**

Create: `web/apps/main/src/pages/PluginMarket.tsx`

```typescript
import {
  Box,
  Grid,
  Card,
  CardContent,
  CardActions,
  Typography,
  Button,
  Chip,
  TextField,
  InputAdornment,
} from '@mui/material';
import { Search, Download, Extension } from '@mui/icons-material';

const plugins = [
  {
    name: 'chat',
    title: 'LLM Chat',
    description: 'Full-featured chat interface with streaming support',
    author: 'AgentMesh',
    installs: 1234,
    rating: 4.8,
    installed: true,
  },
  {
    name: 'image-gen',
    title: 'Image Generation',
    description: 'AI-powered image generation with Stable Diffusion',
    author: 'Community',
    installs: 856,
    rating: 4.5,
    installed: false,
  },
  {
    name: 'code-assist',
    title: 'Code Assistant',
    description: 'AI code completion and review',
    author: 'Community',
    installs: 2100,
    rating: 4.9,
    installed: false,
  },
];

export function PluginMarket() {
  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 4, fontWeight: 700 }}>
        Plugin Market
      </Typography>

      <TextField
        fullWidth
        placeholder="Search plugins..."
        InputProps={{
          startAdornment: (
            <InputAdornment position="start">
              <Search />
            </InputAdornment>
          ),
        }}
        sx={{ mb: 4 }}
      />

      <Grid container spacing={3}>
        {plugins.map((plugin) => (
          <Grid item xs={12} sm={6} md={4} key={plugin.name}>
            <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <Extension sx={{ mr: 1, color: 'primary.main' }} />
                  <Typography variant="h6" sx={{ fontWeight: 600 }}>
                    {plugin.title}
                  </Typography>
                </Box>
                <Typography color="text.secondary" sx={{ mb: 2 }}>
                  {plugin.description}
                </Typography>
                <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
                  <Chip label={`by ${plugin.author}`} size="small" />
                  <Chip label={`⭐ ${plugin.rating}`} size="small" />
                  <Chip label={`${plugin.installs} installs`} size="small" />
                </Box>
              </CardContent>
              <CardActions>
                {plugin.installed ? (
                  <Button size="small" disabled>
                    Installed
                  </Button>
                ) : (
                  <Button size="small" startIcon={<Download />}>
                    Install
                  </Button>
                )}
              </CardActions>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Box>
  );
}
```

- [ ] **Step 4: Commit**

```bash
git add web/apps/main/src/pages/
git commit -m "feat: add main site pages (dashboard, plugin market)"
```

---

### 3.5 后台管理框架

**Files:**
- Create: `web/apps/admin/src/App.tsx`
- Create: `web/apps/admin/src/pages/TenantManagement.tsx`
- Create: `web/apps/admin/src/pages/PluginManagement.tsx`

- [ ] **Step 1: 创建后台 App 组件**

Create: `web/apps/admin/src/App.tsx`

```typescript
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { Box } from '@mui/material';
import { theme } from '@agentmesh/ui';
import { AdminLayout } from './layouts/AdminLayout';
import { Dashboard } from './pages/Dashboard';
import { TenantManagement } from './pages/TenantManagement';
import { PluginManagement } from './pages/PluginManagement';
import { UserManagement } from './pages/UserManagement';

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<AdminLayout />}>
            <Route index element={<Navigate to="/admin" replace />} />
            <Route path="/admin" element={<Dashboard />} />
            <Route path="/admin/tenants" element={<TenantManagement />} />
            <Route path="/admin/plugins" element={<PluginManagement />} />
            <Route path="/admin/users" element={<UserManagement />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ThemeProvider>
  );
}

export default App;
```

- [ ] **Step 2: 创建 Admin Layout**

Create: `web/apps/admin/src/layouts/AdminLayout.tsx`

```typescript
import React from 'react';
import { Outlet } from 'react-router-dom';
import {
  Box,
  Drawer,
  AppBar,
  Toolbar,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Typography,
  Divider,
} from '@mui/material';
import {
  Dashboard,
  Business,
  Extension,
  People,
  Settings,
} from '@mui/icons-material';

const DRAWER_WIDTH = 240;

const menuItems = [
  { label: 'Dashboard', icon: <Dashboard />, path: '/admin' },
  { label: 'Tenant Management', icon: <Business />, path: '/admin/tenants' },
  { label: 'Plugin Management', icon: <Extension />, path: '/admin/plugins' },
  { label: 'User Management', icon: <People />, path: '/admin/users' },
];

export function AdminLayout() {
  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
        <Toolbar>
          <Typography variant="h6" noWrap sx={{ flexGrow: 1 }}>
            AgentMesh Admin
          </Typography>
        </Toolbar>
      </AppBar>
      <Drawer
        variant="permanent"
        sx={{
          width: DRAWER_WIDTH,
          flexShrink: 0,
          '& .MuiDrawer-paper': { width: DRAWER_WIDTH },
        }}
      >
        <Toolbar />
        <Box sx={{ overflow: 'auto' }}>
          <List>
            {menuItems.map((item) => (
              <ListItem key={item.path} disablePadding>
                <ListItemButton component="a" href={item.path}>
                  <ListItemIcon>{item.icon}</ListItemIcon>
                  <ListItemText primary={item.label} />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
          <Divider />
        </Box>
      </Drawer>
      <Box
        component="main"
        sx={{ flexGrow: 1, p: 3, width: `calc(100% - ${DRAWER_WIDTH}px)` }}
      >
        <Toolbar />
        <Outlet />
      </Box>
    </Box>
  );
}
```

- [ ] **Step 3: 创建租户管理页面**

Create: `web/apps/admin/src/pages/TenantManagement.tsx`

```typescript
import React, { useState } from 'react';
import {
  Box,
  Typography,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
} from '@mui/material';
import { Add, Edit, Delete } from '@mui/icons-material';

const initialTenants = [
  { id: 1, name: 'Acme Corp', schema: 'agentmesh_tenant_1', status: 'active', users: 15, created: '2024-01-15' },
  { id: 2, name: 'TechStartup', schema: 'agentmesh_tenant_2', status: 'active', users: 8, created: '2024-02-20' },
  { id: 3, name: 'Enterprise Co', schema: 'agentmesh_tenant_3', status: 'inactive', users: 42, created: '2024-03-10' },
];

export function TenantManagement() {
  const [tenants, setTenants] = useState(initialTenants);
  const [open, setOpen] = useState(false);
  const [editingTenant, setEditingTenant] = useState<typeof initialTenants[0] | null>(null);

  const handleAdd = () => {
    setEditingTenant(null);
    setOpen(true);
  };

  const handleEdit = (tenant: typeof initialTenants[0]) => {
    setEditingTenant(tenant);
    setOpen(true);
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4" sx={{ fontWeight: 700 }}>
          Tenant Management
        </Typography>
        <Button variant="contained" startIcon={<Add />} onClick={handleAdd}>
          Add Tenant
        </Button>
      </Box>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>ID</TableCell>
              <TableCell>Name</TableCell>
              <TableCell>Schema</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Users</TableCell>
              <TableCell>Created</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {tenants.map((tenant) => (
              <TableRow key={tenant.id}>
                <TableCell>{tenant.id}</TableCell>
                <TableCell>{tenant.name}</TableCell>
                <TableCell sx={{ fontFamily: 'monospace' }}>{tenant.schema}</TableCell>
                <TableCell>
                  <Chip
                    label={tenant.status}
                    color={tenant.status === 'active' ? 'success' : 'default'}
                    size="small"
                  />
                </TableCell>
                <TableCell>{tenant.users}</TableCell>
                <TableCell>{tenant.created}</TableCell>
                <TableCell>
                  <IconButton size="small" onClick={() => handleEdit(tenant)}>
                    <Edit />
                  </IconButton>
                  <IconButton size="small" color="error">
                    <Delete />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={open} onClose={() => setOpen(false)}>
        <DialogTitle>{editingTenant ? 'Edit Tenant' : 'Add Tenant'}</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            label="Tenant Name"
            margin="normal"
            defaultValue={editingTenant?.name}
          />
          <TextField
            fullWidth
            label="Schema Name"
            margin="normal"
            defaultValue={editingTenant?.schema}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={() => setOpen(false)}>
            {editingTenant ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
```

- [ ] **Step 4: Commit**

```bash
git add web/apps/admin/
git commit -m "feat: add admin site with tenant and plugin management"
```

---

## Phase 4: 示例插件

### 4.1 LLM Chat 后端集成

**Files:**
- Create: `services/agent-svc/internal/logic/chatlogic.go`
- Create: `services/agent-svc/internal/logic/conversationlogic.go`

- [ ] **Step 1: 创建 Chat Logic（Eino 集成）**

Create: `services/agent-svc/internal/logic/chatlogic.go`

```go
package logic

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"agentmesh/agent-svc/internal/svc"
	"agentmesh/agent-svc/model"
	"agentmesh/agent-svc/pb/agent"

	"github.com/zeromicro/go-zero/core/logx"
)

// EinoChatConfig Eino chat configuration
type EinoChatConfig struct {
	Model       string `json:"model"`
	MaxTokens   int    `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// ChatLogic handles chat requests with Eino
type ChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ChatLogic) Chat(stream agent.Agent_ChatServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Save user message
		userMsg := &model.Message{
			ConversationID: req.ConversationId,
			TenantID:        req.TenantId,
			Role:            "user",
			Content:         req.Content,
			Model:           req.Model,
		}
		_, err = l.svcCtx.MessageModel.Insert(l.ctx, userMsg)
		if err != nil {
			logx.Error("failed to save user message", logx.Field("error", err))
		}

		// Get AI response
		response, err := l.callEino(l.ctx, req.Content, req.Model)
		if err != nil {
			logx.Error("Eino call failed", logx.Field("error", err))
			response = "Sorry, I encountered an error."
		}

		// Save assistant message
		assistantMsg := &model.Message{
			ConversationID: req.ConversationId,
			TenantID:        req.TenantId,
			Role:            "assistant",
			Content:         response,
			Model:           req.Model,
		}
		_, err = l.svcCtx.MessageModel.Insert(l.ctx, assistantMsg)
		if err != nil {
			logx.Error("failed to save assistant message", logx.Field("error", err))
		}

		// Send response
		if err := stream.Send(&agent.ChatResponse{
			Content:   response,
			Model:     req.Model,
			Timestamp: "",
		}); err != nil {
			return err
		}
	}
}

// callEino calls Eino for chat completion
// In production, this would use the actual Eino SDK
func (l *ChatLogic) callEino(ctx context.Context, prompt, modelName string) (string, error) {
	// TODO: Integrate with actual Eino SDK
	// For now, return a mock response

	switch {
	case strings.Contains(strings.ToLower(modelName), "gpt"):
		return "This is a GPT response to: " + prompt, nil
	case strings.Contains(strings.ToLower(modelName), "claude"):
		return "This is a Claude response to: " + prompt, nil
	default:
		return "This is a default response to: " + prompt, nil
	}
}

// ChatStream supports streaming responses
func (l *ChatLogic) ChatStream(in *agent.ChatRequest, stream agent.Agent_ChatStreamServer) error {
	response, err := l.callEino(l.ctx, in.Content, in.Model)
	if err != nil {
		return err
	}

	// Stream word by word for demo
	words := strings.Split(response, " ")
	for _, word := range words {
		if err := stream.Send(&agent.ChatResponse{
			Content: word + " ",
		}); err != nil {
			return err
		}
	}

	return nil
}
```

- [ ] **Step 2: 创建 Conversation Logic**

Create: `services/agent-svc/internal/logic/conversationlogic.go`

```go
package logic

import (
	"context"

	"agentmesh/agent-svc/internal/svc"
	"agentmesh/agent-svc/model"
	"agentmesh/agent-svc/pb/agent"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConversationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConversationLogic {
	return &ConversationLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ConversationLogic) CreateConversation(in *agent.CreateConversationRequest) (*agent.Conversation, error) {
	conv := &model.Conversation{
		TenantID: in.TenantId,
		UserID:   in.UserId,
		Title:    in.Title,
		Model:    in.Model,
		Status:   1,
	}

	id, err := l.svcCtx.ConversationModel.Insert(l.ctx, conv)
	if err != nil {
		logx.Error("failed to create conversation", logx.Field("error", err))
		return nil, err
	}

	return &agent.Conversation{
		Id:    id,
		Title: in.Title,
		Model: in.Model,
	}, nil
}

func (l *ConversationLogic) ListConversations(in *agent.ListConversationsRequest) (*agent.ListConversationsResponse, error) {
	convs, err := l.svcCtx.ConversationModel.FindAllByTenantAndUser(l.ctx, in.TenantId, in.UserId)
	if err != nil {
		return nil, err
	}

	var result []*agent.Conversation
	for _, c := range convs {
		result = append(result, &agent.Conversation{
			Id:        c.ID,
			Title:     c.Title,
			Model:     c.Model,
			Status:    int32(c.Status),
			CreatedAt: c.CreatedAt.Unix(),
		})
	}

	return &agent.ListConversationsResponse{
		Conversations: result,
	}, nil
}

func (l *ConversationLogic) GetMessages(in *agent.GetMessagesRequest) (*agent.GetMessagesResponse, error) {
	msgs, err := l.svcCtx.MessageModel.FindAllByConversation(l.ctx, in.ConversationId)
	if err != nil {
		return nil, err
	}

	var result []*agent.Message
	for _, m := range msgs {
		result = append(result, &agent.Message{
			Id:             m.ID,
			ConversationId: m.ConversationID,
			Role:            m.Role,
			Content:         m.Content,
			Model:           m.Model,
			CreatedAt:       m.CreatedAt.Unix(),
		})
	}

	return &agent.GetMessagesResponse{
		Messages: result,
	}, nil
}
```

- [ ] **Step 3: 创建 Message Model**

Create: `services/agent-svc/model/messagemodel.go`

```go
package model

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type MessageModel struct {
	db sqlx.SqlConn
}

func NewMessageModel(db sqlx.SqlConn) *MessageModel {
	return &MessageModel{db: db}
}

type Message struct {
	ID             int64  `json:"id"`
	ConversationID int64  `json:"conversation_id"`
	TenantID       int64  `json:"tenant_id"`
	Role           string `json:"role"` // user, assistant, system
	Content        string `json:"content"`
	Model          string `json:"model,omitempty"`
	TokenCount     int    `json:"token_count,omitempty"`
	CreatedAt      int64  `json:"created_at"`
}

func (m *MessageModel) Insert(ctx context.Context, msg *Message) (int64, error) {
	query := `INSERT INTO messages (conversation_id, tenant_id, role, content, model, token_count) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := m.db.ExecCtx(ctx, query, msg.ConversationID, msg.TenantID, msg.Role, msg.Content, msg.Model, msg.TokenCount)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (m *MessageModel) FindAllByConversation(ctx context.Context, conversationID int64) ([]*Message, error) {
	query := `SELECT id, conversation_id, tenant_id, role, content, model, token_count, created_at FROM messages WHERE conversation_id = ? ORDER BY created_at ASC`
	var msgs []*Message
	rows, err := m.db.QueryCtx(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.TenantID, &msg.Role, &msg.Content, &msg.Model, &msg.TokenCount, &msg.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, &msg)
	}

	return msgs, nil
}
```

- [ ] **Step 4: Commit**

```bash
git add services/agent-svc/internal/logic/ services/agent-svc/model/
git commit -m "feat: add chat backend with conversation management"
```

---

### 4.2 Chat UI 开发（调用 ui-ux-pro-max）

**Files:**
- Create: `web/apps/main/src/pages/Chat.tsx`
- Create: `web/packages/components/src/ChatInput.tsx`
- Create: `web/packages/components/src/MessageBubble.tsx`

- [ ] **Step 1: 使用 ui-ux-pro-max 设计 Chat UI**

**需要先调用 `ui-ux-pro-max:ui-ux-pro-max` skill 进行聊天界面设计**

以下为设计后的 Chat 页面结构：

- [ ] **Step 2: 创建 Chat 页面**

Create: `web/apps/main/src/pages/Chat.tsx`

```typescript
import React, { useState, useRef, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  IconButton,
  Menu,
  MenuItem,
  Select,
  FormControl,
  InputLabel,
  Fab,
  Tooltip,
} from '@mui/material';
import { Send, MoreVert, Settings, Delete } from '@mui/icons-material';
import { ChatInput } from '@agentmesh/ui';
import { MessageBubble } from '@agentmesh/ui';

interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  model?: string;
  timestamp: Date;
}

const MODELS = [
  { value: 'gpt-4', label: 'GPT-4' },
  { value: 'gpt-3.5-turbo', label: 'GPT-3.5 Turbo' },
  { value: 'claude-3-opus', label: 'Claude 3 Opus' },
  { value: 'claude-3-sonnet', label: 'Claude 3 Sonnet' },
];

export function Chat() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      role: 'assistant',
      content: 'Hello! I\'m your AI assistant. How can I help you today?',
      model: 'gpt-4',
      timestamp: new Date(),
    },
  ]);
  const [selectedModel, setSelectedModel] = useState('gpt-4');
  const [isTyping, setIsTyping] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = async (content: string) => {
    if (!content.trim()) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content,
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setIsTyping(true);

    // Simulate AI response
    setTimeout(() => {
      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: `This is a response from ${selectedModel}. In production, this would be a real streaming response from the AI model.`,
        model: selectedModel,
        timestamp: new Date(),
      };
      setMessages((prev) => [...prev, assistantMessage]);
      setIsTyping(false);
    }, 1500);
  };

  return (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Header */}
      <Box
        sx={{
          p: 2,
          borderBottom: 1,
          borderColor: 'divider',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Typography variant="h6" sx={{ fontWeight: 600 }}>
          Chat
        </Typography>
        <FormControl size="small" sx={{ minWidth: 180 }}>
          <InputLabel>Model</InputLabel>
          <Select
            value={selectedModel}
            label="Model"
            onChange={(e) => setSelectedModel(e.target.value)}
          >
            {MODELS.map((model) => (
              <MenuItem key={model.value} value={model.value}>
                {model.label}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </Box>

      {/* Messages */}
      <Box
        sx={{
          flexGrow: 1,
          overflow: 'auto',
          p: 2,
          display: 'flex',
          flexDirection: 'column',
          gap: 2,
        }}
      >
        {messages.map((message) => (
          <MessageBubble
            key={message.id}
            message={message.content}
            role={message.role}
            model={message.model}
            timestamp={message.timestamp}
          />
        ))}
        {isTyping && (
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              gap: 1,
              p: 2,
            }}
          >
            <Typography color="text.secondary" sx={{ fontStyle: 'italic' }}>
              {selectedModel} is typing...
            </Typography>
          </Box>
        )}
        <div ref={messagesEndRef} />
      </Box>

      {/* Input */}
      <Box sx={{ p: 2, borderTop: 1, borderColor: 'divider' }}>
        <ChatInput onSend={handleSend} disabled={isTyping} />
      </Box>
    </Box>
  );
}
```

- [ ] **Step 3: 创建 ChatInput 组件**

Create: `web/packages/ui/src/components/ChatInput/index.tsx`

```typescript
import React, { useState } from 'react';
import {
  Box,
  IconButton,
  TextField,
  Paper,
  Tooltip,
} from '@mui/material';
import { Send, AttachFile } from '@mui/icons-material';

interface ChatInputProps {
  onSend: (message: string) => void;
  disabled?: boolean;
  placeholder?: string;
}

export function ChatInput({
  onSend,
  disabled = false,
  placeholder = 'Type a message...',
}: ChatInputProps) {
  const [value, setValue] = useState('');

  const handleSend = () => {
    if (value.trim() && !disabled) {
      onSend(value);
      setValue('');
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <Paper
      variant="outlined"
      sx={{
        display: 'flex',
        alignItems: 'center',
        p: 1,
        gap: 1,
      }}
    >
      <Tooltip title="Attach file">
        <IconButton size="small" disabled={disabled}>
          <AttachFile />
        </IconButton>
      </Tooltip>
      <TextField
        fullWidth
        multiline
        maxRows={4}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyPress={handleKeyPress}
        placeholder={placeholder}
        disabled={disabled}
        variant="standard"
        InputProps={{
          disableUnderline: true,
          sx: { fontSize: '1rem' },
        }}
      />
      <IconButton
        color="primary"
        onClick={handleSend}
        disabled={disabled || !value.trim()}
      >
        <Send />
      </IconButton>
    </Paper>
  );
}
```

- [ ] **Step 4: 创建 MessageBubble 组件**

Create: `web/packages/ui/src/components/MessageBubble/index.tsx`

```typescript
import React from 'react';
import { Box, Typography, Chip, Avatar } from '@mui/material';
import { Person, SmartToy } from '@mui/icons-material';

interface MessageBubbleProps {
  message: string;
  role: 'user' | 'assistant' | 'system';
  model?: string;
  timestamp: Date;
}

export function MessageBubble({
  message,
  role,
  model,
  timestamp,
}: MessageBubbleProps) {
  const isUser = role === 'user';

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: isUser ? 'row-reverse' : 'row',
        gap: 1.5,
        alignItems: 'flex-start',
      }}
    >
      <Avatar
        sx={{
          bgcolor: isUser ? 'primary.main' : 'secondary.main',
          width: 36,
          height: 36,
        }}
      >
        {isUser ? <Person /> : <SmartToy />}
      </Avatar>

      <Box sx={{ maxWidth: '70%' }}>
        <Box
          sx={{
            bgcolor: isUser ? 'primary.main' : 'grey.100',
            color: isUser ? 'white' : 'text.primary',
            borderRadius: 2,
            p: 2,
            borderTopLeftRadius: isUser ? 16 : 4,
            borderTopRightRadius: isUser ? 4 : 16,
          }}
        >
          <Typography variant="body1" sx={{ whiteSpace: 'pre-wrap' }}>
            {message}
          </Typography>
        </Box>
        <Box
          sx={{
            display: 'flex',
            gap: 1,
            mt: 0.5,
            justifyContent: isUser ? 'flex-end' : 'flex-start',
          }}
        >
          {model && (
            <Chip
              label={model}
              size="small"
              sx={{
                height: 20,
                fontSize: '0.7rem',
                bgcolor: 'grey.200',
              }}
            />
          )}
          <Typography variant="caption" color="text.secondary">
            {timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </Typography>
        </Box>
      </Box>
    </Box>
  );
}
```

- [ ] **Step 5: Commit**

```bash
git add web/apps/main/src/pages/Chat.tsx web/packages/ui/src/components/ChatInput/ web/packages/ui/src/components/MessageBubble/
git commit -m "feat: add chat UI with streaming support"
```

---

### 4.3 对话历史

**Files:**
- Modify: `web/apps/main/src/pages/Chat.tsx`（添加侧边栏对话列表）

- [ ] **Step 1: 添加对话列表侧边栏**

Create: `web/apps/main/src/components/ChatSidebar/index.tsx`

```typescript
import React from 'react';
import {
  Box,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Typography,
  IconButton,
  Divider,
  TextField,
  InputAdornment,
} from '@mui/material';
import { Search, Add, Chat, Delete } from '@mui/icons-material';

interface Conversation {
  id: string;
  title: string;
  updatedAt: Date;
  messageCount: number;
}

interface ChatSidebarProps {
  conversations: Conversation[];
  selectedId?: string;
  onSelect: (id: string) => void;
  onNew: () => void;
}

export function ChatSidebar({
  conversations,
  selectedId,
  onSelect,
  onNew,
}: ChatSidebarProps) {
  return (
    <Box sx={{ width: 280, borderRight: 1, borderColor: 'divider', height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Header */}
      <Box sx={{ p: 2 }}>
        <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>
          Conversations
        </Typography>
        <IconButton fullWidth variant="outlined" onClick={onNew} sx={{ borderStyle: 'dashed' }}>
          <Add /> New Chat
        </IconButton>
      </Box>

      {/* Search */}
      <Box sx={{ px: 2, pb: 1 }}>
        <TextField
          fullWidth
          size="small"
          placeholder="Search..."
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <Search sx={{ fontSize: 20 }} />
              </InputAdornment>
            ),
          }}
        />
      </Box>

      {/* Conversation List */}
      <List sx={{ flexGrow: 1, overflow: 'auto' }}>
        {conversations.map((conv) => (
          <ListItem
            key={conv.id}
            disablePadding
            secondaryAction={
              <IconButton size="small" edge="end">
                <Delete fontSize="small" />
              </IconButton>
            }
          >
            <ListItemButton
              selected={selectedId === conv.id}
              onClick={() => onSelect(conv.id)}
              sx={{ borderRadius: 1, mx: 1 }}
            >
              <Chat sx={{ mr: 1.5, fontSize: 20, color: 'text.secondary' }} />
              <ListItemText
                primary={conv.title}
                secondary={`${conv.messageCount} messages`}
                primaryTypographyProps={{
                  variant: 'body2',
                  fontWeight: selectedId === conv.id ? 600 : 400,
                  noWrap: true,
                }}
                secondaryTypographyProps={{
                  variant: 'caption',
                }}
              />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
    </Box>
  );
}
```

- [ ] **Step 2: 更新 Chat 页面集成侧边栏**

```typescript
// In Chat.tsx, add state and integrate ChatSidebar
const [conversations] = useState<Conversation[]>([
  { id: '1', title: 'Project Discussion', updatedAt: new Date(), messageCount: 12 },
  { id: '2', title: 'Code Review', updatedAt: new Date(), messageCount: 8 },
]);

// In the return statement, wrap with a Box that has display: flex
<Box sx={{ display: 'flex', height: '100%' }}>
  <ChatSidebar
    conversations={conversations}
    selectedId={selectedConversationId}
    onSelect={handleSelectConversation}
    onNew={handleNewChat}
  />
  <Box sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column' }}>
    {/* existing chat content */}
  </Box>
</Box>
```

- [ ] **Step 3: Commit**

```bash
git add web/apps/main/src/components/ChatSidebar/
git commit -m "feat: add conversation history sidebar"
```

---

## 执行计划总结

**任务数量：** 约 25 个任务
**预计时间：** 4 个阶段逐步完成

| Phase | 任务数 | 主要内容 |
|-------|--------|----------|
| Phase 1 | 6 | 基础设施（脚手架、微服务框架、数据库、JWT） |
| Phase 2 | 4 | 核心功能（用户/租户管理、插件后端、前端SDK） |
| Phase 3 | 5 | 前端三站（脚手架、组件库、落地页、主站、后台） |
| Phase 4 | 4 | 示例插件（Chat后端、Chat UI、流式输出、对话历史） |

**Git 分支建议：**
```bash
git checkout -b feature/infrastructure
git checkout -b feature/core-features
git checkout -b feature/frontend-sites
git checkout -b feature/llm-chat-plugin
```
