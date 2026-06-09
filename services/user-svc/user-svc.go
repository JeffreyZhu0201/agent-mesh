// Package main - 用户认证服务
//
// 职责：用户注册/登录/JWT签发 + RBAC权限控制 + 多租户管理
//
// 架构说明：
//   - 所有用户共用 users 表，通过 tenant_id 字段区分租户（共享表模式）
//   - tenant_id=0 表示平台超级管理员（platform_admin），可管理所有租户
//   - tenant_id>0 表示租户管理员（admin）或普通用户（viewer/user）
//   - JWT payload 包含 userId, username, tenantId, role，用于身份认证和权限判断
//   - 所有需要认证的请求通过 authMiddleware() 解析 JWT 并注入上下文
//   - 权限层级：platform_admin > admin > viewer/user
//
// 数据库表：
//   users(id, username, password_hash, email, tenant_id, role, status)
//   tenants(id, name, code, status)
//
// 角色说明：
//   - platform_admin：平台超级管理员，tenant_id=0，可管理所有租户和用户
//   - admin：租户管理员，可管理本租户内的普通用户
//   - user/viewer：普通用户，只有查看权限
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"agentmesh/user-svc/model"
)

// ============ 全局变量 ============
//
// db: 全局数据库连接，整个服务共享同一个连接池
// jwtSecret: JWT签名密钥，用于验证和签发token
//
// 安全注意：生产环境中 jwtSecret 必须通过环境变量或密钥管理服务注入，
// 禁止硬编码在代码中。当前默认值仅用于本地开发。
var (
	db        *gorm.DB
	jwtSecret = []byte("agentmesh-secret-key-change-in-production")
)

// ============ JWT Claims 结构体 ============
//
// JWT Claims 是令牌的载荷，包含了用户身份信息
//
// 字段说明：
//   - UserID: 用户唯一标识，用于标识是哪个用户
//   - Username: 用户名，用于显示，不用于权限判断
//   - TenantID: 租户ID，0表示平台管理员，非0表示租户
//   - Role: 角色，用于RBAC权限判断
//
// JWT工作流程：
//   1. 用户登录时，服务端验证用户名密码，查询用户信息
//   2. 服务端创建Claims，包含用户身份信息
//   3. 使用jwtSecret对Claims进行签名，生成token返回给客户端
//   4. 客户端后续请求时在Header中携带 token
//   5. 服务端authMiddleware验证token签名，确认用户身份
type Claims struct {
	UserID   uint   `json:"userId"`   // 用户ID
	Username string `json:"username"` // 用户名
	TenantID uint   `json:"tenantId"` // 租户ID，0=平台管理员
	Role     string `json:"role"`     // 角色：platform_admin, admin, user, viewer
	jwt.RegisteredClaims              // JWT标准声明，包含过期时间等
}

// ============ 响应辅助函数 ============
//
// 统一API响应格式：
//   成功：{"success": true, "data": ...}
//   失败：{"success": false, "error": "错误信息"}
//
// 为什么统一格式？前端可以基于success字段判断请求是否成功，
// 不需要每个接口单独处理错误展示逻辑。

// jsonSuccess - 返回成功响应
// 用于需要返回数据结构的成功响应
// c: Gin上下文
// data: 要返回的数据，可以是任意类型
func jsonSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// jsonError - 返回错误响应
// 用于所有失败场景，包括参数错误、权限不足、资源不存在等
// c: Gin上下文
// code: HTTP状态码
// msg: 错误信息，应简洁明了，告诉用户或开发者出了什么问题
func jsonError(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"success": false, "error": msg})
}

// jsonSuccessMsg - 返回成功消息
// 用于不需要返回数据，只需要确认操作成功的响应（如删除操作）
// c: Gin上下文
// msg: 成功消息
func jsonSuccessMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg})
}

// ============ 权限辅助函数 ============
//
// RBAC（基于角色的访问控制）实现：
//   - 通过 gin.Context 中的 "role" 键值判断权限
//   - authMiddleware 解析JWT后会设置 role 到上下文
//   - 各handler通过 requirePlatformAdmin/requireTenantAdmin 检查权限

// requirePlatformAdmin - 检查是否为平台管理员
// 平台管理员拥有最高权限，可以管理所有租户和用户
// tenant_id=0 且 role=platform_admin 才是平台管理员
// 返回：true=是平台管理员，false=不是
func requirePlatformAdmin(c *gin.Context) bool {
	role, _ := c.Get("role")
	return role == "platform_admin"
}

// requireTenantAdmin - 检查是否为租户管理员
// 租户管理员可以管理本租户内的用户，但不能跨租户操作
// 需要 tenant_id>0 且 role=admin
// 返回：true=是租户管理员，false=不是
func requireTenantAdmin(c *gin.Context) bool {
	role, _ := c.Get("role")
	return role == "admin"
}

// getTenantID - 从上下文获取当前用户的租户ID
// 用于租户隔离场景，确保用户只能操作本租户的数据
// 平台管理员的tenant_id为0
func getTenantID(c *gin.Context) uint {
	tenantId, _ := c.Get("tenantId")
	return tenantId.(uint)
}

// getUserID - 从上下文获取当前用户的ID
// 用于记录操作日志、关联用户数据等场景
func getUserID(c *gin.Context) uint {
	userId, _ := c.Get("userId")
	return userId.(uint)
}

// ============ JWT认证中间件 ============
//
// authMiddleware 是所有需要认证的接口的前置拦截器
//
// 工作流程（逐行）：
//   1. 从请求Header中获取 Authorization 字段
//   2. 检查格式：必须是 "Bearer <token>" 结构
//   3. 提取token字符串
//   4. 使用jwtSecret验证token签名
//   5. 验证通过后，从Claims中提取用户信息
//   6. 将用户信息设置到 gin.Context 中，供后续handler使用
//   7. 调用 c.Next() 继续处理请求
//
// 失败场景：
//   - Header为空或格式不对：返回401 missing or invalid authorization header
//   - token签名无效或已过期：返回401 invalid or expired token
//
// curl 示例：
//   curl -H "Authorization: Bearer <token>" http://localhost:8081/api/user/info
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 第1步：获取Authorization Header
		authHeader := c.GetHeader("Authorization")
		// 第2步：验证格式，必须是 "Bearer " 开头
		if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			jsonError(c, http.StatusUnauthorized, "missing or invalid authorization header")
			c.Abort() // 阻止后续handler执行
			return
		}

		// 第3步：提取token字符串（去掉 "Bearer " 前缀）
		tokenStr := authHeader[7:]
		claims := &Claims{}

		// 第4步：解析并验证token
		// jwt.ParseWithClaims 会：
		//   - 验证签名是否正确（防止篡改）
		//   - 验证是否过期（ExpiresAt）
		//   - 解析Claims到我们定义的结构体
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			// 验证签名算法，必须是HMAC（HS256/HS384/HS512）
			// 这是为了防止攻击者使用其他算法伪造token
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return jwtSecret, nil // 返回签名密钥
		})

		// 第5步：检查解析结果
		if err != nil || !token.Valid {
			jsonError(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		// 第6步：将用户信息注入上下文，供后续handler使用
		// 这些信息来自JWT token，是可信的（因为已经验证过签名）
		c.Set("userId", claims.UserID)
		c.Set("tenantId", claims.TenantID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		// 第7步：继续处理请求
		c.Next()
	}
}

// ============ 平台管理员权限中间件 ============
//
// platformAdminOnly 是平台管理员专用接口的拦截器
// 必须在 authMiddleware 之后使用，因为它依赖上下文中的 role 信息
//
// 权限判断逻辑：
//   - 检查 c.Get("role") 是否为 "platform_admin"
//   - 只有 tenant_id=0 且 role=platform_admin 的用户能通过
//
// curl 示例：
//   curl -H "Authorization: Bearer <platform_admin_token>" http://localhost:8081/api/admin/tenants
func platformAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requirePlatformAdmin(c) {
			jsonError(c, http.StatusForbidden, "platform admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}

// ============ 租户管理员权限中间件 ============
//
// tenantAdminOnly 是租户管理员专用接口的拦截器
// 必须在 authMiddleware 之后使用
//
// 权限判断逻辑：
//   - 检查 c.Get("role") 是否为 "admin"
//   - 注意：平台管理员（platform_admin）也能通过此检查
//   - 因为平台管理员理论上可以协助管理任何租户
//
// curl 示例：
//   curl -H "Authorization: Bearer <tenant_admin_token>" http://localhost:8081/api/tenant/users
func tenantAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requireTenantAdmin(c) {
			jsonError(c, http.StatusForbidden, "tenant admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}

// ============ 公开接口（无需认证） ============
//
// 公开接口不需要Authorization Header
// 用于：用户注册、用户登录

// POST /api/user/register - 用户注册
//
// 功能：创建新用户账号
//
// 请求体：
//   - username: 用户名（必填，唯一）
//   - password: 密码（必填，最少6字符）
//   - email: 邮箱（可选）
//
// 业务逻辑：
//   1. 验证请求参数
//   2. 检查用户名是否已被占用
//   3. 使用bcrypt对密码进行哈希加密（不可逆）
//   4. 创建用户，默认租户ID=1，角色为viewer
//
// 注意：新注册用户默认是普通用户（viewer），不是管理员
//
// curl 示例：
//   curl -X POST http://localhost:8081/api/user/register \
//     -H "Content-Type: application/json" \
//     -d '{"username":"testuser","password":"123456","email":"test@example.com"}'
func handleRegister(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// 密码长度验证
	if len(body.Password) < 6 {
		jsonError(c, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	// 检查用户名唯一性
	var existing model.User
	if err := db.Where("username = ?", body.Username).First(&existing).Error; err == nil {
		// err == nil 表示查询成功，即找到了同名用户
		jsonError(c, http.StatusConflict, "username already exists")
		return
	}

	// 密码加密：bcrypt是单向哈希，无法解密
	// 即使数据库泄露，攻击者也无法获取用户原始密码
	// DefaultCost=10，计算时间约300ms，能有效防止暴力破解
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// 创建用户记录
	// tenant_id=1：默认加入第一个租户（共享表模式）
	// role=viewer：默认角色是查看者，不是管理员
	user := model.User{
		Username:     body.Username,
		PasswordHash: string(hash),
		Email:        body.Email,
		TenantID:     1,
		Role:         model.RoleViewer,
		Status:       1, // 1=启用，0=禁用
	}

	if err := db.Create(&user).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	// 返回创建的用户信息（不包含密码）
	jsonSuccess(c, gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"tenantId":  user.TenantID,
		"role":      user.Role,
		"status":    user.Status,
		"createdAt": user.CreatedAt,
	})
}

// POST /api/user/login - 用户登录
//
// 功能：验证用户凭证，签发JWT token
//
// 请求体：
//   - username: 用户名（必填）
//   - password: 密码（必填）
//
// 业务逻辑：
//   1. 验证请求参数
//   2. 根据用户名查询用户
//   3. 使用bcrypt验证密码
//   4. 检查账号状态是否启用
//   5. 生成JWT token返回
//
// JWT token有效期：24小时
// 包含信息：userId, username, tenantId, role
//
// curl 示例：
//   curl -X POST http://localhost:8081/api/user/login \
//     -H "Content-Type: application/json" \
//     -d '{"username":"admin","password":"admin123"}'
func handleLogin(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// 查询用户
	var user model.User
	if err := db.Where("username = ?", body.Username).First(&user).Error; err != nil {
		// 用户不存在，但为了防止用户名枚举攻击，返回相同的错误信息
		jsonError(c, http.StatusUnauthorized, "invalid username or password")
		return
	}

	// 验证密码：bcrypt.CompareHashAndPassword 比较哈希值
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		// 密码错误，同样返回相同信息
		jsonError(c, http.StatusUnauthorized, "invalid username or password")
		return
	}

	// 检查账号状态：1=启用，0=禁用
	if user.Status != 1 {
		jsonError(c, http.StatusForbidden, "account is disabled")
		return
	}

	// 构造JWT Claims
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		TenantID: user.TenantID,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24小时后过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),                       // 签发时间
		},
	}

	// 签发token：使用HS256算法，需要jwtSecret密钥
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// 返回token和用户基本信息
	jsonSuccess(c, gin.H{
		"token": tokenStr,
		"user": gin.H{
			"userId":   user.ID,
			"username": user.Username,
			"tenantId": user.TenantID,
			"role":     user.Role,
		},
	})
}

// ============ 需要认证的接口 ============
//
// 需要携带有效的JWT token访问
// 通过 authMiddleware 进行身份验证

// GET /api/user/info - 获取当前用户信息
//
// 功能：获取已登录用户的基本信息
// 需要认证：是的（authMiddleware）
//
// curl 示例：
//   curl -H "Authorization: Bearer <token>" http://localhost:8081/api/user/info
func handleUserInfo(c *gin.Context) {
	userId := getUserID(c) // 从JWT上下文获取当前用户ID

	var user model.User
	if err := db.First(&user, userId).Error; err != nil {
		jsonError(c, http.StatusNotFound, "user not found")
		return
	}

	jsonSuccess(c, gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"tenantId":  user.TenantID,
		"role":      user.Role,
		"status":    user.Status,
		"createdAt": user.CreatedAt,
	})
}

// ============ 平台管理员接口 (/api/admin/*) ============
//
// 平台管理员专用接口，用于管理所有租户和用户
// 需要权限：platform_admin（authMiddleware + platformAdminOnly）
//
// 权限说明：
//   - 平台管理员的 tenant_id=0，role=platform_admin
//   - 可以创建、修改、删除租户
//   - 可以查看和管理所有租户下的用户

// GET /api/admin/tenants - 获取所有租户列表
//
// 功能：列出系统中所有租户
// 权限：platform_admin
//
// curl 示例：
//   curl -H "Authorization: Bearer <platform_admin_token>" http://localhost:8081/api/admin/tenants
func handleListTenants(c *gin.Context) {
	var tenants []model.Tenant
	if err := db.Order("id asc").Find(&tenants).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to list tenants")
		return
	}
	jsonSuccess(c, tenants)
}

// POST /api/admin/tenants - 创建新租户
//
// 功能：在系统中创建一个新的租户
// 权限：platform_admin
//
// 请求体：
//   - name: 租户名称（必填）
//   - code: 租户代码（必填，唯一，用于URL和标识）
//
// 业务逻辑：
//   1. 验证请求参数
//   2. 检查code唯一性
//   3. 创建租户记录
//   4. 自动创建该租户的第一个管理员账号
//      - 用户名：{code}_admin（如 "acme_admin"）
//      - 密码：password（临时密码，生产环境应通知用户修改）
//      - 角色：admin
//
// curl 示例：
//   curl -X POST -H "Authorization: Bearer <platform_admin_token>" \
//     -H "Content-Type: application/json" \
//     -d '{"name":"ACME Corp","code":"acme"}' \
//     http://localhost:8081/api/admin/tenants
func handleCreateTenant(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// 检查code唯一性
	var existing model.Tenant
	if err := db.Where("code = ?", body.Code).First(&existing).Error; err == nil {
		jsonError(c, http.StatusConflict, "tenant code already exists")
		return
	}

	// 创建租户
	tenant := model.Tenant{
		Name:   body.Name,
		Code:   body.Code,
		Status: 1, // 启用状态
	}

	if err := db.Create(&tenant).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to create tenant")
		return
	}

	// 为新租户创建第一个管理员账号
	// 约定：用户名 = {code}_admin，密码 = password
	adminUsername := body.Code + "_admin"
	var adminExisting model.User
	if err := db.Where("username = ?", adminUsername).First(&adminExisting).Error; err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		adminUser := model.User{
			Username:     adminUsername,
			PasswordHash: string(hash),
			Email:        "",
			TenantID:     tenant.ID, // 关联到新创建的租户
			Role:         model.RoleAdmin,
			Status:       1,
		}
		db.Create(&adminUser)
	}

	jsonSuccess(c, tenant)
}

// PUT /api/admin/tenants/:id - 更新租户信息
//
// 功能：修改租户名称、代码或状态
// 权限：platform_admin
//
// 请求体：
//   - name: 租户名称（可选）
//   - code: 租户代码（可选）
//   - status: 状态（可选，1=启用，0=禁用）
//
// curl 示例：
//   curl -X PUT -H "Authorization: Bearer <platform_admin_token>" \
//     -H "Content-Type: application/json" \
//     -d '{"name":"ACME Corporation Updated"}' \
//     http://localhost:8081/api/admin/tenants/1
func handleUpdateTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid tenant id")
		return
	}

	var body struct {
		Name   string `json:"name"`
		Code   string `json:"code"`
		Status *int   `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// 查询要更新的租户
	var tenant model.Tenant
	if err := db.First(&tenant, id).Error; err != nil {
		jsonError(c, http.StatusNotFound, "tenant not found")
		return
	}

	// 检查code唯一性（如果要修改code）
	if body.Code != "" && body.Code != tenant.Code {
		var existing model.Tenant
		if err := db.Where("code = ? AND id != ?", body.Code, id).First(&existing).Error; err == nil {
			jsonError(c, http.StatusConflict, "tenant code already exists")
			return
		}
		tenant.Code = body.Code
	}

	// 更新可选字段
	if body.Name != "" {
		tenant.Name = body.Name
	}
	if body.Status != nil {
		tenant.Status = *body.Status
	}

	if err := db.Save(&tenant).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to update tenant")
		return
	}

	jsonSuccess(c, tenant)
}

// DELETE /api/admin/tenants/:id - 删除租户
//
// 功能：删除租户及其所有用户数据
// 权限：platform_admin
//
// 重要：此操作会同时删除该租户下的所有用户，数据不可恢复！
// 防护：id=1（默认租户）不能删除
//
// curl 示例：
//   curl -X DELETE -H "Authorization: Bearer <platform_admin_token>" \
//     http://localhost:8081/api/admin/tenants/2
func handleDeleteTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid tenant id")
		return
	}

	// 防护：不能删除默认租户（id=1）
	if id == 1 {
		jsonError(c, http.StatusForbidden, "cannot delete default tenant")
		return
	}

	// 先删除该租户下的所有用户
	if err := db.Where("tenant_id = ?", id).Delete(&model.User{}).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to delete tenant users")
		return
	}

	// 再删除租户本身
	if err := db.Delete(&model.Tenant{}, id).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to delete tenant")
		return
	}

	jsonSuccessMsg(c, "tenant deleted")
}

// GET /api/admin/users - 获取所有用户列表
//
// 功能：查看系统中所有用户（跨所有租户）
// 权限：platform_admin
// 查询参数：
//   - tenant_id: 可选，按租户筛选
//
// curl 示例：
//   # 查看所有用户
//   curl -H "Authorization: Bearer <platform_admin_token>" http://localhost:8081/api/admin/users
//
//   # 只看租户2的用户
//   curl -H "Authorization: Bearer <platform_admin_token>" "http://localhost:8081/api/admin/users?tenant_id=2"
func handleListAllUsers(c *gin.Context) {
	query := db.Model(&model.User{})

	// 支持按租户筛选
	if tenantIdStr := c.Query("tenant_id"); tenantIdStr != "" {
		tenantId, err := strconv.ParseUint(tenantIdStr, 10, 32)
		if err == nil {
			query = query.Where("tenant_id = ?", tenantId)
		}
	}

	var users []model.User
	if err := query.Order("id asc").Find(&users).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to list users")
		return
	}

	jsonSuccess(c, users)
}

// POST /api/admin/users - 创建用户（跨租户）
//
// 功能：平台管理员可以为任意租户创建用户
// 权限：platform_admin
//
// 请求体：
//   - username: 用户名（必填）
//   - password: 密码（必填）
//   - email: 邮箱（可选）
//   - tenantId: 租户ID（可选，默认=1）
//   - role: 角色（可选，默认=viewer）
//
// curl 示例：
//   curl -X POST -H "Authorization: Bearer <platform_admin_token>" \
//     -H "Content-Type: application/json" \
//     -d '{"username":"john","password":"secret","tenantId":2,"role":"viewer"}' \
//     http://localhost:8081/api/admin/users
func handleCreateUser(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
		TenantID uint   `json:"tenantId"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// 默认值处理
	if body.TenantID == 0 {
		body.TenantID = 1
	}
	if body.Role == "" {
		body.Role = model.RoleViewer
	}

	// 检查用户名唯一性
	var existing model.User
	if err := db.Where("username = ?", body.Username).First(&existing).Error; err == nil {
		jsonError(c, http.StatusConflict, "username already exists")
		return
	}

	// 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// 创建用户
	user := model.User{
		Username:     body.Username,
		PasswordHash: string(hash),
		Email:        body.Email,
		TenantID:     body.TenantID,
		Role:         body.Role,
		Status:       1,
	}

	if err := db.Create(&user).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	jsonSuccess(c, user)
}

// PUT /api/admin/users/:id - 更新任意用户
//
// 功能：平台管理员可以更新任意租户的用户信息
// 权限：platform_admin
//
// 请求体：
//   - email: 邮箱
//   - role: 角色
//   - status: 状态
//
// curl 示例：
//   curl -X PUT -H "Authorization: Bearer <platform_admin_token>" \
//     -H "Content-Type: application/json" \
//     -d '{"role":"admin"}' \
//     http://localhost:8081/api/admin/users/5
func handleUpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var body struct {
		Email  string `json:"email"`
		Role   string `json:"role"`
		Status *int   `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// 查询用户
	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		jsonError(c, http.StatusNotFound, "user not found")
		return
	}

	// 更新字段
	if body.Email != "" {
		user.Email = body.Email
	}
	if body.Role != "" {
		user.Role = body.Role
	}
	if body.Status != nil {
		user.Status = *body.Status
	}

	if err := db.Save(&user).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to update user")
		return
	}

	jsonSuccess(c, user)
}

// DELETE /api/admin/users/:id - 删除任意用户
//
// 功能：平台管理员可以删除任意租户的用户
// 权限：platform_admin
//
// curl 示例：
//   curl -X DELETE -H "Authorization: Bearer <platform_admin_token>" \
//     http://localhost:8081/api/admin/users/5
func handleDeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := db.Delete(&model.User{}, id).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to delete user")
		return
	}

	jsonSuccessMsg(c, "user deleted")
}

// ============ 租户管理员接口 (/api/tenant/*) ============
//
// 租户管理员专用接口，用于管理本租户内的用户
// 需要权限：admin（authMiddleware + tenantAdminOnly）
//
// 多租户隔离：
//   - 每个租户管理员只能操作本租户（tenant_id匹配）的用户
//   - 通过 getTenantID(c) 获取当前用户所属租户，自动过滤

// GET /api/tenant/users - 获取本租户用户列表
//
// 功能：列出当前租户下的所有用户
// 权限：admin（本租户）
// 自动过滤：只返回 tenant_id = 当前用户所属租户 的用户
//
// curl 示例：
//   curl -H "Authorization: Bearer <tenant_admin_token>" http://localhost:8081/api/tenant/users
func handleListTenantUsers(c *gin.Context) {
	tenantId := getTenantID(c) // 自动获取当前用户的租户ID

	var users []model.User
	if err := db.Where("tenant_id = ?", tenantId).Order("id asc").Find(&users).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to list users")
		return
	}

	jsonSuccess(c, users)
}

// POST /api/tenant/users - 在本租户创建用户
//
// 功能：租户管理员为自己所属租户创建新用户
// 权限：admin（本租户）
//
// 注意：
//   - 新用户自动归属到当前租户（tenant_id自动设置）
//   - 新用户角色默认是 user，不是 admin
//
// curl 示例：
//   curl -X POST -H "Authorization: Bearer <tenant_admin_token>" \
//     -H "Content-Type: application/json" \
//     -d '{"username":"employee1","password":"welcome123","email":"emp@company.com"}' \
//     http://localhost:8081/api/tenant/users
func handleCreateTenantUser(c *gin.Context) {
	tenantId := getTenantID(c) // 自动使用当前用户所属的租户

	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// 检查用户名唯一性（全局）
	var existing model.User
	if err := db.Where("username = ?", body.Username).First(&existing).Error; err == nil {
		jsonError(c, http.StatusConflict, "username already exists")
		return
	}

	// 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// 创建用户，自动绑定到当前租户
	user := model.User{
		Username:     body.Username,
		PasswordHash: string(hash),
		Email:        body.Email,
		TenantID:     tenantId,
		Role:         model.RoleUser, // 注意：是user，不是admin
		Status:       1,
	}

	if err := db.Create(&user).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	jsonSuccess(c, user)
}

// PUT /api/tenant/users/:id - 更新本租户用户
//
// 功能：修改本租户下用户的信息
// 权限：admin（本租户）
//
// 安全限制：
//   - 只能修改本租户的用户（tenant_id必须匹配）
//   - 不能修改 admin 角色的用户（防止权限提升）
//   - 不能将用户角色改为 admin（防止提权）
//
// curl 示例：
//   curl -X PUT -H "Authorization: Bearer <tenant_admin_token>" \
//     -H "Content-Type: application/json" \
//     -d '{"email":"newemail@company.com","role":"viewer"}' \
//     http://localhost:8081/api/tenant/users/10
func handleUpdateTenantUser(c *gin.Context) {
	tenantId := getTenantID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var body struct {
		Email  string `json:"email"`
		Role   string `json:"role"`
		Status *int   `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// 查询用户
	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		jsonError(c, http.StatusNotFound, "user not found")
		return
	}

	// 【多租户隔离】检查用户是否属于本租户
	if user.TenantID != tenantId {
		jsonError(c, http.StatusForbidden, "cannot modify user from another tenant")
		return
	}

	// 【安全限制】不能修改 admin 角色的用户
	if user.Role == model.RoleAdmin {
		jsonError(c, http.StatusForbidden, "cannot modify tenant admin users")
		return
	}

	// 更新字段
	if body.Email != "" {
		user.Email = body.Email
	}
	// 【安全限制】不能将角色改为 admin
	if body.Role != "" && body.Role != model.RoleAdmin {
		user.Role = body.Role
	}
	if body.Status != nil {
		user.Status = *body.Status
	}

	if err := db.Save(&user).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to update user")
		return
	}

	jsonSuccess(c, user)
}

// DELETE /api/tenant/users/:id - 删除本租户用户
//
// 功能：删除本租户下的用户
// 权限：admin（本租户）
//
// 安全限制：
//   - 只能删除本租户的用户
//   - 不能删除 admin 角色的用户
//
// curl 示例：
//   curl -X DELETE -H "Authorization: Bearer <tenant_admin_token>" \
//     http://localhost:8081/api/tenant/users/10
func handleDeleteTenantUser(c *gin.Context) {
	tenantId := getTenantID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	// 查询用户
	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		jsonError(c, http.StatusNotFound, "user not found")
		return
	}

	// 【多租户隔离】检查用户是否属于本租户
	if user.TenantID != tenantId {
		jsonError(c, http.StatusForbidden, "cannot delete user from another tenant")
		return
	}

	// 【安全限制】不能删除 admin 角色的用户
	if user.Role == model.RoleAdmin {
		jsonError(c, http.StatusForbidden, "cannot delete tenant admin users")
		return
	}

	if err := db.Delete(&model.User{}, id).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to delete user")
		return
	}

	jsonSuccessMsg(c, "user deleted")
}

// ============ 初始化数据（种子数据） ============
//
// seedData 用于服务启动时初始化必要的默认数据
// 目的：确保系统有一个可用的平台管理员和默认租户
//
// 创建的数据：
//   1. 默认租户（id=1, code=default）
//      - 所有新注册用户默认属于此租户
//      - 防止新用户因为没有租户而无法使用系统
//
//   2. 平台超级管理员（username=admin）
//      - tenant_id=0，表示不属于任何租户
//      - role=platform_admin，拥有最高权限
//      - 用于管理所有租户和用户
//      - 密码：admin123（生产环境必须修改！）
//
// 安全注意：
//   - 平台管理员的密码是默认密码，仅用于首次部署
//   - 生产环境部署后必须立即修改为强密码
func seedData() {
	// 创建默认租户（如果不存在）
	var defaultTenant model.Tenant
	if err := db.Where("id = ?", 1).First(&defaultTenant).Error; err != nil {
		defaultTenant = model.Tenant{
			ID:     1,
			Name:   "Default Tenant",
			Code:   "default",
			Status: 1,
		}
		db.Create(&defaultTenant)
		fmt.Println("Seeded default tenant (id=1, code=default)")
	}

	// 创建平台管理员（如果不存在）
	var adminUser model.User
	if err := db.Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		adminUser = model.User{
			Username:     "admin",
			PasswordHash: string(hash),
			Email:        "admin@agentmesh.io",
			TenantID:     0,                    // 0=平台管理员，不属于任何租户
			Role:         model.RolePlatformAdmin, // 最高权限
			Status:       1,
		}
		db.Create(&adminUser)
		fmt.Println("Seeded platform admin (username=admin, password=admin123)")
	}
}

// ============ 服务入口 ============
//
// main 函数是服务的启动入口
// 启动顺序：
//   1. 解析命令行参数（-port）
//   2. 连接数据库
//   3. 自动迁移（创建表结构）
//   4. 初始化种子数据
//   5. 配置路由和中间件
//   6. 启动HTTP服务器
//
// 路由分组：
//   - GET  /health              - 健康检查（公开）
//   - POST /api/user/register   - 用户注册（公开）
//   - POST /api/user/login      - 用户登录（公开）
//   - GET  /api/user/info       - 当前用户信息（需认证）
//   - /api/admin/*              - 平台管理员接口（需platform_admin）
//   - /api/tenant/*             - 租户管理员接口（需admin）
func main() {
	// 第1步：解析命令行参数
	var port int
	flag.IntVar(&port, "port", 8081, "server port")
	flag.Parse()

	// 第2步：连接数据库
	// 环境变量说明：
	//   DB_HOST: 数据库地址（默认localhost）
	//   DB_PORT: 数据库端口（默认3306）
	//   DB_USER: 数据库用户名（默认root）
	//   DB_PASSWORD: 数据库密码（默认root123）
	//   DB_NAME: 数据库名（默认agentmesh）
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}
	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = "root123"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "agentmesh"
	}

	// 构建DSN（Data Source Name）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connected to MySQL database")

	// 第3步：自动迁移
	// AutoMigrate 会自动创建或更新表结构
	// 首次运行创建 users 和 tenants 表
	// 后续运行添加新字段（如果模型有变化）
	if err := db.AutoMigrate(&model.User{}, &model.Tenant{}); err != nil {
		fmt.Printf("Failed to auto migrate: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Auto migration complete")

	// 第4步：初始化种子数据
	seedData()

	// 第5步：配置Gin路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())    // 请求日志
	r.Use(gin.Recovery()) // panic恢复

	// 健康检查接口（无需认证）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 公开接口（无需认证）
	r.POST("/api/user/register", handleRegister)
	r.POST("/api/user/login", handleLogin)

	// 需要认证的接口
	auth := r.Group("/api/user")
	auth.Use(authMiddleware())
	{
		auth.GET("/info", handleUserInfo)
	}

	// 平台管理员接口（需要platform_admin权限）
	admin := r.Group("/api/admin")
	admin.Use(authMiddleware(), platformAdminOnly())
	{
		admin.GET("/tenants", handleListTenants)
		admin.POST("/tenants", handleCreateTenant)
		admin.PUT("/tenants/:id", handleUpdateTenant)
		admin.DELETE("/tenants/:id", handleDeleteTenant)
		admin.GET("/users", handleListAllUsers)
		admin.POST("/users", handleCreateUser)
		admin.PUT("/users/:id", handleUpdateUser)
		admin.DELETE("/users/:id", handleDeleteUser)
	}

	// 租户管理员接口（需要admin权限）
	tenant := r.Group("/api/tenant")
	tenant.Use(authMiddleware(), tenantAdminOnly())
	{
		tenant.GET("/users", handleListTenantUsers)
		tenant.POST("/users", handleCreateTenantUser)
		tenant.PUT("/users/:id", handleUpdateTenantUser)
		tenant.DELETE("/users/:id", handleDeleteTenantUser)
	}

	// 第6步：启动HTTP服务器
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("User service listening on %s\n", addr)
	if err := r.Run(addr); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
