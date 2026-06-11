package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// =============================================================================
// plugin-svc 插件管理服务
// =============================================================================
// 职责：
//   - 维护插件注册表（plugins），记录系统中所有可用的插件
//   - 维护租户级插件开关（tenant_plugins），支持插件启用/禁用
//   - 提供三个 API 端点给前端和管理后台使用
//
// 数据库设计：
//   plugins         — 插件注册表，所有租户共享一份
//     plugin_key: 唯一标识，如 "dashboard"
//     builtin:    是否内置插件（内置不能删除）
//
//   tenant_plugins — 租户级启用状态（可选记录）
//     无记录 = 默认启用（enabled=true）
//     有记录 = 以记录的 enabled 为准
//
// JWT 上下文：
//   本服务不验证 JWT 签名（由 api-gateway 统一验证）
//   直接从请求头 X-Tenant-ID 取 tenant_id 进行过滤
// =============================================================================

var configFile = flag.String("f", "plugin-svc.yaml", "the config file")

// db is the shared database handle used by handlers and tests.
var db *gorm.DB

// jwtSecret 与 user-svc 保持一致，用于解析 JWT Token 中的租户上下文
// 注意：生产环境应由 api-gateway 统一验证 JWT，本服务只解析 tenant_id
var jwtSecret = []byte("agentmesh-secret-key-change-in-production")

// Config 服务配置结构，对应 plugin-svc.yaml
type Config struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

// Plugin 插件注册表模型
// 所有租户共享同一套插件定义，plugin_key 是唯一标识
type Plugin struct {
	ID          uint   `gorm:"primaryKey"`
	PluginKey   string `gorm:"column:plugin_key;size:64;uniqueIndex;not null"` // 插件唯一标识，如 "chat"、"dashboard"
	Name        string `gorm:"size:255;not null"`                              // 插件显示名称
	Version     string `gorm:"size:32;not null"`                                // 版本号
	Type        string `gorm:"size:32"`                                         // 插件类型，如 "page"、"widget"
	Description string `gorm:"type:text"`                                       // 插件描述
	Author      string `gorm:"size:255"`                                        // 作者
	Builtin     bool   `gorm:"default:true"`                                   // 是否内置插件（内置不可删除）
	CreatedAt   time.Time
}

// TenantPlugin 租户级插件启用状态模型
// 用于覆盖插件的默认启用状态，实现租户粒度的插件管理
//
// 语义规则：
//   - 表中无该租户的记录 → 插件默认启用（enabled=true）
//   - 表中有该租户的记录 → 以记录的 enabled 为准
type TenantPlugin struct {
	ID         uint      `gorm:"primaryKey"`
	TenantID   uint      `gorm:"column:tenant_id;uniqueIndex:uq_tenant_plugin;not null"` // 租户 ID
	PluginKey  string    `gorm:"column:plugin_key;size:64;uniqueIndex:uq_tenant_plugin;not null"`
	Enabled    bool      // 插件启用状态；无记录时应用层默认启用
	UpdatedAt  time.Time
}

// Claims JWT Token 中的用户声明结构
// 本服务主要使用 TenantID 进行数据隔离
type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	TenantID uint   `json:"tenantId"` // 租户 ID，用于插件数据隔离
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// getDB 初始化数据库连接
// 优先级：环境变量 > 默认值（本地开发环境）
//
// 环境变量：
//   DB_HOST     数据库主机，默认 localhost
//   DB_PORT     数据库端口，默认 3306
//   DB_USER     数据库用户名，默认 root
//   DB_PASSWORD 数据库密码，默认 root123
//   DB_NAME     数据库名，默认 agentmesh
func getDB() *gorm.DB {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}
	pass := os.Getenv("DB_PASSWORD")
	if pass == "" {
		pass = "root123"
	}
	name := os.Getenv("DB_NAME")
	if name == "" {
		name = "agentmesh"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	return db
}

// migrate 自动迁移数据库表结构
// 会根据模型定义创建/更新 plugins 和 tenant_plugins 表
func migrate(db *gorm.DB) {
	db.AutoMigrate(&Plugin{}, &TenantPlugin{})
}

// seedPlugins 初始化内置插件数据
// 只在插件不存在时才插入，确保多次启动不会重复创建
//
// 内置插件：
//   - chat:      AI 对话插件，支持多模型切换
//   - dashboard: 数据分析仪表板插件
func seedPlugins(db *gorm.DB) {
	builtins := []Plugin{
		{
			PluginKey:   "chat",
			Name:        "Chat",
			Version:     "0.1.0",
			Type:        "page",
			Description: "Conversational AI chat with multi-model support",
			Author:      "AgentMesh",
			Builtin:     true,
		},
		{
			PluginKey:   "dashboard",
			Name:        "Dashboard",
			Version:     "0.1.0",
			Type:        "page",
			Description: "Platform overview and analytics dashboard",
			Author:      "AgentMesh",
			Builtin:     true,
		},
	}

	for _, p := range builtins {
		var existing Plugin
		db.Where("plugin_key = ?", p.PluginKey).First(&existing)
		if existing.ID == 0 {
			db.Create(&p)
		}
	}
}

// authMiddleware JWT 认证中间件
//
// 设计说明：
//   - 本服务信任 api-gateway 传来的 Authorization header
//   - api-gateway 已在入口处验证 JWT 签名并注入 X-Tenant-ID
//   - 这里只解析 JWT 获取 tenant_id 和用户信息
//
// 注意：生产环境建议移除此中间件，直接由 api-gateway 验证
//       或者使用内部 RPC 调用获取租户上下文
func authMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization header"})
		c.Abort()
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization header"})
		c.Abort()
		return
	}

	tokenStr := parts[1]
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		c.Abort()
		return
	}

	// 将解析出的用户信息存入 Gin Context，供后续 handler 使用
	c.Set("userId", claims.UserID)
	c.Set("tenantId", claims.TenantID)
	c.Set("role", claims.Role)
	c.Next()
}

// =============================================================================
// main 函数入口
// =============================================================================
// 启动流程：
//   1. 解析命令行参数（-f 指定配置文件）
//   2. 连接数据库并自动迁移表结构
//   3. 初始化内置插件数据
//   4. 注册路由并启动 HTTP 服务（端口 8083）
//
// 路由设计：
//   GET  /health           — 健康检查（无需认证）
//   GET  /api/plugin/list  — 获取当前租户已启用的插件列表（前端用）
//   GET  /api/plugin/admin/list  — 获取所有插件及启用状态（管理后台用）
//   POST /api/plugin/admin/toggle — 启用/禁用插件（管理后台用）
// =============================================================================
func main() {
	flag.Parse()

	db = getDB()
	migrate(db)
	seedPlugins(db)

	r := gin.Default()
	registerRoutes(r)

	fmt.Println("Starting Plugin Service at port 8083...")
	r.Run(":8083")
}

// tenantPluginRow 租户插件开关记录（查询结果）
type tenantPluginRow struct {
	PluginKey string `gorm:"column:plugin_key"`
	Enabled   bool   `gorm:"column:enabled"`
}

// tenantPluginMaps 租户插件状态：disabled=明确禁用，explicit=有显式记录，enabled=显式启用值
type tenantPluginMaps struct {
	disabled map[string]bool
	explicit map[string]bool
	enabled  map[string]bool
}

// loadTenantPluginMaps 加载指定租户的插件开关状态
func loadTenantPluginMaps(db *gorm.DB, tenantID uint) tenantPluginMaps {
	var rows []tenantPluginRow
	db.Table("tenant_plugins").Select("plugin_key, enabled").
		Where("tenant_id = ?", tenantID).Scan(&rows)

	result := tenantPluginMaps{
		disabled: make(map[string]bool),
		explicit: make(map[string]bool),
		enabled:  make(map[string]bool),
	}
	for _, row := range rows {
		result.explicit[row.PluginKey] = true
		result.enabled[row.PluginKey] = row.Enabled
		if !row.Enabled {
			result.disabled[row.PluginKey] = true
		}
	}
	return result
}

// pluginToJSON 将 Plugin 模型转为 API 响应字段
func pluginToJSON(p Plugin, enabled bool) gin.H {
	return gin.H{
		"key": p.PluginKey, "name": p.Name, "version": p.Version,
		"type": p.Type, "description": p.Description, "author": p.Author,
		"enabled": enabled,
	}
}

func registerRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authed := r.Group("/api/plugin", authMiddleware)
	{
		authed.GET("/list", func(c *gin.Context) {
			tenantID, _ := c.Get("tenantId")
			var allPlugins []Plugin
			db.Find(&allPlugins)

			state := loadTenantPluginMaps(db, tenantID.(uint))
			result := make([]gin.H, 0, len(allPlugins))
			for _, p := range allPlugins {
				if state.disabled[p.PluginKey] {
					continue
				}
				result = append(result, pluginToJSON(p, true))
			}
			c.JSON(http.StatusOK, gin.H{"plugins": result})
		})

		authed.GET("/admin/list", func(c *gin.Context) {
			tenantID, _ := c.Get("tenantId")
			var allPlugins []Plugin
			db.Find(&allPlugins)

			state := loadTenantPluginMaps(db, tenantID.(uint))
			result := make([]gin.H, 0, len(allPlugins))
			for _, p := range allPlugins {
				enabled := true
				if state.explicit[p.PluginKey] {
					enabled = state.enabled[p.PluginKey]
				}
				result = append(result, pluginToJSON(p, enabled))
			}
			c.JSON(http.StatusOK, gin.H{"plugins": result})
		})

		authed.POST("/admin/toggle", func(c *gin.Context) {
			tenantID, _ := c.Get("tenantId")

			var req struct {
				PluginKey string `json:"pluginKey" binding:"required"`
				Enabled   bool   `json:"enabled"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
				return
			}

			var plugin Plugin
			if err := db.Where("plugin_key = ?", req.PluginKey).First(&plugin).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"message": "plugin not found"})
				return
			}

			var tp TenantPlugin
			result := db.Where("tenant_id = ? AND plugin_key = ?", tenantID, req.PluginKey).First(&tp)
			record := TenantPlugin{
				TenantID:  tenantID.(uint),
				PluginKey: req.PluginKey,
				Enabled:   req.Enabled,
				UpdatedAt: time.Now(),
			}
			if result.Error != nil {
				if err := db.Select("TenantID", "PluginKey", "Enabled", "UpdatedAt").Create(&record).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to save plugin state"})
					return
				}
			} else {
				tp.Enabled = req.Enabled
				tp.UpdatedAt = time.Now()
				if err := db.Save(&tp).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update plugin state"})
					return
				}
			}

			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})
	}
}
