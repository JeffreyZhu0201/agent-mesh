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

var (
	db        *gorm.DB
	jwtSecret = []byte("agentmesh-secret-key-change-in-production")
)

// JWT claims
type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	TenantID uint   `json:"tenantId"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Response helpers
func jsonSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func jsonError(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"success": false, "error": msg})
}

func jsonSuccessMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg})
}

// Permission helpers
func requirePlatformAdmin(c *gin.Context) bool {
	role, _ := c.Get("role")
	return role == "platform_admin"
}

func requireTenantAdmin(c *gin.Context) bool {
	role, _ := c.Get("role")
	return role == "admin"
}

func getTenantID(c *gin.Context) uint {
	tenantId, _ := c.Get("tenantId")
	return tenantId.(uint)
}

func getUserID(c *gin.Context) uint {
	userId, _ := c.Get("userId")
	return userId.(uint)
}

// JWT middleware
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			jsonError(c, http.StatusUnauthorized, "missing or invalid authorization header")
			c.Abort()
			return
		}

		tokenStr := authHeader[7:]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			jsonError(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("tenantId", claims.TenantID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// Platform admin only middleware
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

// Tenant admin only middleware
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

// ============ Public Endpoints ============

// POST /api/user/register
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

	if len(body.Password) < 6 {
		jsonError(c, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	// Check if username already exists
	var existing model.User
	if err := db.Where("username = ?", body.Username).First(&existing).Error; err == nil {
		jsonError(c, http.StatusConflict, "username already exists")
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Create user with tenant_id=1, role="viewer"
	user := model.User{
		Username:     body.Username,
		PasswordHash: string(hash),
		Email:        body.Email,
		TenantID:     1,
		Role:         model.RoleViewer,
		Status:       1,
	}

	if err := db.Create(&user).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to create user")
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

// POST /api/user/login
func handleLogin(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	var user model.User
	if err := db.Where("username = ?", body.Username).First(&user).Error; err != nil {
		jsonError(c, http.StatusUnauthorized, "invalid username or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		jsonError(c, http.StatusUnauthorized, "invalid username or password")
		return
	}

	if user.Status != 1 {
		jsonError(c, http.StatusForbidden, "account is disabled")
		return
	}

	// Generate JWT
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		TenantID: user.TenantID,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

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

// ============ Authenticated Endpoints ============

// GET /api/user/info
func handleUserInfo(c *gin.Context) {
	userId := getUserID(c)

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

// ============ Platform Admin Endpoints (/api/admin/*) ============

// GET /api/admin/tenants
func handleListTenants(c *gin.Context) {
	var tenants []model.Tenant
	if err := db.Order("id asc").Find(&tenants).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to list tenants")
		return
	}
	jsonSuccess(c, tenants)
}

// POST /api/admin/tenants
func handleCreateTenant(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// Check if code already exists
	var existing model.Tenant
	if err := db.Where("code = ?", body.Code).First(&existing).Error; err == nil {
		jsonError(c, http.StatusConflict, "tenant code already exists")
		return
	}

	tenant := model.Tenant{
		Name:   body.Name,
		Code:   body.Code,
		Status: 1,
	}

	if err := db.Create(&tenant).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to create tenant")
		return
	}

	// Create first admin user for the tenant
	adminUsername := body.Code + "_admin"
	var adminExisting model.User
	if err := db.Where("username = ?", adminUsername).First(&adminExisting).Error; err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		adminUser := model.User{
			Username:     adminUsername,
			PasswordHash: string(hash),
			Email:        "",
			TenantID:     tenant.ID,
			Role:         model.RoleAdmin,
			Status:       1,
		}
		db.Create(&adminUser)
	}

	jsonSuccess(c, tenant)
}

// PUT /api/admin/tenants/:id
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

	var tenant model.Tenant
	if err := db.First(&tenant, id).Error; err != nil {
		jsonError(c, http.StatusNotFound, "tenant not found")
		return
	}

	// Check code uniqueness if changed
	if body.Code != "" && body.Code != tenant.Code {
		var existing model.Tenant
		if err := db.Where("code = ? AND id != ?", body.Code, id).First(&existing).Error; err == nil {
			jsonError(c, http.StatusConflict, "tenant code already exists")
			return
		}
		tenant.Code = body.Code
	}

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

// DELETE /api/admin/tenants/:id
func handleDeleteTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid tenant id")
		return
	}

	if id == 1 {
		jsonError(c, http.StatusForbidden, "cannot delete default tenant")
		return
	}

	// Delete all users for this tenant
	if err := db.Where("tenant_id = ?", id).Delete(&model.User{}).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to delete tenant users")
		return
	}

	// Delete the tenant
	if err := db.Delete(&model.Tenant{}, id).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to delete tenant")
		return
	}

	jsonSuccessMsg(c, "tenant deleted")
}

// GET /api/admin/users
func handleListAllUsers(c *gin.Context) {
	query := db.Model(&model.User{})

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

// POST /api/admin/users
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

	if body.TenantID == 0 {
		body.TenantID = 1
	}
	if body.Role == "" {
		body.Role = model.RoleViewer
	}

	// Check username uniqueness
	var existing model.User
	if err := db.Where("username = ?", body.Username).First(&existing).Error; err == nil {
		jsonError(c, http.StatusConflict, "username already exists")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

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

// PUT /api/admin/users/:id
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

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		jsonError(c, http.StatusNotFound, "user not found")
		return
	}

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

// DELETE /api/admin/users/:id
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

// ============ Tenant Admin Endpoints (/api/tenant/*) ============

// GET /api/tenant/users
func handleListTenantUsers(c *gin.Context) {
	tenantId := getTenantID(c)

	var users []model.User
	if err := db.Where("tenant_id = ?", tenantId).Order("id asc").Find(&users).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to list users")
		return
	}

	jsonSuccess(c, users)
}

// POST /api/tenant/users
func handleCreateTenantUser(c *gin.Context) {
	tenantId := getTenantID(c)

	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// Check username uniqueness
	var existing model.User
	if err := db.Where("username = ?", body.Username).First(&existing).Error; err == nil {
		jsonError(c, http.StatusConflict, "username already exists")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := model.User{
		Username:     body.Username,
		PasswordHash: string(hash),
		Email:        body.Email,
		TenantID:     tenantId,
		Role:         model.RoleUser, // NOT admin
		Status:       1,
	}

	if err := db.Create(&user).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	jsonSuccess(c, user)
}

// PUT /api/tenant/users/:id
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

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		jsonError(c, http.StatusNotFound, "user not found")
		return
	}

	// Ensure user belongs to this tenant
	if user.TenantID != tenantId {
		jsonError(c, http.StatusForbidden, "cannot modify user from another tenant")
		return
	}

	// CANNOT modify users with role="admin"
	if user.Role == model.RoleAdmin {
		jsonError(c, http.StatusForbidden, "cannot modify tenant admin users")
		return
	}

	if body.Email != "" {
		user.Email = body.Email
	}
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

// DELETE /api/tenant/users/:id
func handleDeleteTenantUser(c *gin.Context) {
	tenantId := getTenantID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		jsonError(c, http.StatusNotFound, "user not found")
		return
	}

	// Ensure user belongs to this tenant
	if user.TenantID != tenantId {
		jsonError(c, http.StatusForbidden, "cannot delete user from another tenant")
		return
	}

	// CANNOT delete users with role="admin"
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

// ============ Seed Data ============

func seedData() {
	// Create default tenant if not exists
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

	// Create platform admin if not exists
	var adminUser model.User
	if err := db.Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		adminUser = model.User{
			Username:     "admin",
			PasswordHash: string(hash),
			Email:        "admin@agentmesh.io",
			TenantID:     0,
			Role:         model.RolePlatformAdmin,
			Status:       1,
		}
		db.Create(&adminUser)
		fmt.Println("Seeded platform admin (username=admin, password=admin123)")
	}
}

// ============ Main ============

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "server port")
	flag.Parse()

	// Database connection
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

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connected to MySQL database")

	// Auto migrate
	if err := db.AutoMigrate(&model.User{}, &model.Tenant{}); err != nil {
		fmt.Printf("Failed to auto migrate: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Auto migration complete")

	// Seed data
	seedData()

	// Gin setup
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public routes
	r.POST("/api/user/register", handleRegister)
	r.POST("/api/user/login", handleLogin)

	// Authenticated routes
	auth := r.Group("/api/user")
	auth.Use(authMiddleware())
	{
		auth.GET("/info", handleUserInfo)
	}

	// Platform admin routes
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

	// Tenant admin routes
	tenant := r.Group("/api/tenant")
	tenant.Use(authMiddleware(), tenantAdminOnly())
	{
		tenant.GET("/users", handleListTenantUsers)
		tenant.POST("/users", handleCreateTenantUser)
		tenant.PUT("/users/:id", handleUpdateTenantUser)
		tenant.DELETE("/users/:id", handleDeleteTenantUser)
	}

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("User service listening on %s\n", addr)
	if err := r.Run(addr); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
