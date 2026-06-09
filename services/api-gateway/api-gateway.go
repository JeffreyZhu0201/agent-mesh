package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

// getEnv 获取环境变量，如果未设置则返回默认值
// 用于支持 Docker 环境变量配置，同时兼容本地开发
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func main() {
	r := gin.Default()

	// ============================================
	// CORS 中间件：解决浏览器跨域访问问题
	// 为什么需要：前端运行在 localhost:3000，后端 API 在 8080 端口
	// 浏览器同源策略会阻止前端直接请求后端，需要此中间件放行
	// Access-Control-Allow-Origin: 允许所有来源（生产环境应限制具体域名）
	// Access-Control-Allow-Methods: 允许的 HTTP 方法
	// Access-Control-Allow-Headers: 允许的请求头（包括 Authorization 和 X-Tenant-ID 用于认证和租户识别）
	// OPTIONS 请求是预检请求，直接返回 204 表示放行
	// ============================================
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Tenant-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// ============================================
	// 健康检查接口：用于 Docker 健康检测和负载均衡器探测
	// 返回 {"status": "ok"} 表示服务正常运行
	// ============================================
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ============================================
	// 路由配置：反向代理到后端微服务
	//
	// 环境变量说明：
	//   - USER_SVC_URL:   用户服务地址（默认 http://user-svc:8081）
	//   - PLUGIN_SVC_URL: 插件服务地址（默认 http://plugin-svc:8083）
	//
	// Docker 环境：使用服务名（如 user-svc）作为主机名
	// 本地开发：可设置为 http://localhost:8081 等地址
	//
	// 路由规则：
	//   /api/user/*   -> 用户服务（处理用户注册、登录、认证等）
	//   /api/plugin/* -> 插件服务（处理插件注册、管理等）
	// ============================================

	// 用户服务代理：处理 /api/user/* 请求
	userSvcAddr := getEnv("USER_SVC_URL", "http://user-svc:8081")
	pluginSvcAddr := getEnv("PLUGIN_SVC_URL", "http://plugin-svc:8083")

	userSvcURL, _ := url.Parse(userSvcAddr)
	pluginSvcURL, _ := url.Parse(pluginSvcAddr)

	// 使用 r.Any 匹配所有 HTTP 方法（GET、POST、PUT、DELETE 等）
	r.Any("/api/user/*path", createProxy(userSvcURL, "/api/user"))
	r.Any("/api/plugin/*path", createProxy(pluginSvcURL, "/api/plugin"))

	fmt.Printf("Starting API Gateway at port 8080...\n")
	fmt.Printf("  user-svc: %s\n", userSvcAddr)
	fmt.Printf("  plugin-svc: %s\n", pluginSvcAddr)
	r.Run(":8080")
}

// ============================================
// createProxy 创建反向代理处理器
//
// 工作原理：
// 1. 使用 httputil.NewSingleHostReverseProxy 创建反向代理实例
// 2. 代理会根据 target URL 自动处理请求转发
// 3. 关键步骤：路径重写
//    - Gin 的 /*path 会捕获通配符路径（如 /api/user/list 中的 /list）
//    - 将其与 basePath 拼接成目标路径（如 /api/user/list）
//    - 更新请求的 URL.Path、RawPath 和 Host
// 4. 代理将修改后的请求转发到目标服务
//
// 参数说明：
//   target:   目标服务的 URL（包含协议、主机名、端口）
//   basePath: 目标服务的基础路径（如 /api/user）
//
// 示例：
//   客户端请求：GET /api/user/list
//   - /*path 捕获到：/list
//   - targetPath = /api/user + /list = /api/user/list
//   - 代理转发到：http://user-svc:8081/api/user/list
// ============================================
func createProxy(target *url.URL, basePath string) gin.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		// 获取 Gin 路由中 /*path 捕获的通配符路径
		path := c.Param("path")
		// 将通配符路径拼接到 basePath 上，重写目标请求路径
		targetPath := basePath + path

		c.Request.URL.Path = targetPath
		c.Request.URL.RawPath = targetPath
		c.Request.Host = target.Host

		// 使用反向代理处理请求：自动处理 TCP 连接、HTTP 协议转换等
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
