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

var configFile = flag.String("f", "plugin-svc.yaml", "the config file")
var jwtSecret = []byte("agentmesh-secret-key-change-in-production") // same as user-svc

type Config struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

type Plugin struct {
	ID          uint   `gorm:"primaryKey"`
	PluginKey   string `gorm:"column:plugin_key;size:64;uniqueIndex;not null"`
	Name        string `gorm:"size:255;not null"`
	Version     string `gorm:"size:32;not null"`
	Type        string `gorm:"size:32"`
	Description string `gorm:"type:text"`
	Author      string `gorm:"size:255"`
	Builtin     bool   `gorm:"default:true"`
	CreatedAt   time.Time
}

type TenantPlugin struct {
	ID         uint   `gorm:"primaryKey"`
	TenantID   uint   `gorm:"column:tenant_id;uniqueIndex:uq_tenant_plugin;not null"`
	PluginKey  string `gorm:"column:plugin_key;size:64;uniqueIndex:uq_tenant_plugin;not null"`
	Enabled    bool   `gorm:"default:true"`
	UpdatedAt  time.Time
}

type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	TenantID uint   `json:"tenantId"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

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

func migrate(db *gorm.DB) {
	db.AutoMigrate(&Plugin{}, &TenantPlugin{})
}

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

	c.Set("userId", claims.UserID)
	c.Set("tenantId", claims.TenantID)
	c.Set("role", claims.Role)
	c.Next()
}

func main() {
	flag.Parse()

	db := getDB()
	migrate(db)
	seedPlugins(db)

	r := gin.Default()

	// (CORS is handled by the api-gateway; setting it here too causes duplicate headers.)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authed := r.Group("/api/plugin", authMiddleware)
	{
		// Public list: only enabled plugins for the tenant
		authed.GET("/list", func(c *gin.Context) {
			tenantID, _ := c.Get("tenantId")

			var allPlugins []Plugin
			db.Find(&allPlugins)

			// Build a set of explicitly disabled plugins for this tenant
			type tp struct {
				PluginKey string `gorm:"column:plugin_key"`
				Enabled   bool   `gorm:"column:enabled"`
			}
			var tps []tp
			db.Table("tenant_plugins").Select("plugin_key, enabled").
				Where("tenant_id = ?", tenantID).Scan(&tps)

			disabled := make(map[string]bool)
			for _, t := range tps {
				if !t.Enabled {
					disabled[t.PluginKey] = true
				}
			}

			// Return only enabled plugins
			result := make([]gin.H, 0, len(allPlugins))
			for _, p := range allPlugins {
				if disabled[p.PluginKey] {
					continue
				}
				result = append(result, gin.H{
					"key":         p.PluginKey,
					"name":        p.Name,
					"version":     p.Version,
					"type":        p.Type,
					"description": p.Description,
					"author":      p.Author,
					"enabled":     true,
				})
			}

			c.JSON(http.StatusOK, gin.H{"plugins": result})
		})

		// Admin list: all plugins with enable/disable state
		authed.GET("/admin/list", func(c *gin.Context) {
			tenantID, _ := c.Get("tenantId")

			var allPlugins []Plugin
			db.Find(&allPlugins)

			// Fetch tenant enable state
			type tp struct {
				PluginKey string `gorm:"column:plugin_key"`
				Enabled   bool   `gorm:"column:enabled"`
			}
			var tps []tp
			db.Table("tenant_plugins").Select("plugin_key, enabled").
				Where("tenant_id = ?", tenantID).Scan(&tps)

			state := make(map[string]bool)
			explicit := make(map[string]bool)
			for _, t := range tps {
				state[t.PluginKey] = t.Enabled
				explicit[t.PluginKey] = true
			}

			result := make([]gin.H, 0, len(allPlugins))
			for _, p := range allPlugins {
				enabled := true // default
				if explicit[p.PluginKey] {
					enabled = state[p.PluginKey]
				}
				result = append(result, gin.H{
					"key":         p.PluginKey,
					"name":        p.Name,
					"version":     p.Version,
					"type":        p.Type,
					"description": p.Description,
					"author":      p.Author,
					"enabled":     enabled,
				})
			}

			c.JSON(http.StatusOK, gin.H{"plugins": result})
		})

		// Admin toggle: flip enabled state for (tenant, plugin)
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

			// Verify plugin exists
			var plugin Plugin
			if err := db.Where("plugin_key = ?", req.PluginKey).First(&plugin).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"message": "plugin not found"})
				return
			}

			// Upsert the tenant_plugin row
			var tp TenantPlugin
			db.Where("tenant_id = ? AND plugin_key = ?", tenantID, req.PluginKey).First(&tp)
			tp.TenantID = tenantID.(uint)
			tp.PluginKey = req.PluginKey
			tp.Enabled = req.Enabled
			tp.UpdatedAt = time.Now()

			if tp.ID == 0 {
				db.Create(&tp)
			} else {
				db.Save(&tp)
			}

			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})
	}

	fmt.Println("Starting Plugin Service at port 8083...")
	r.Run(":8083")
}
