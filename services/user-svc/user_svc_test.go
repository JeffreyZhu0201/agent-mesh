package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"agentmesh/user-svc/model"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, testDB.AutoMigrate(&model.User{}, &model.Tenant{}))
	db = testDB

	tenant := model.Tenant{ID: 1, Name: "Default", Code: "default", Status: 1}
	require.NoError(t, testDB.Create(&tenant).Error)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/user/register", handleRegister)
	r.POST("/api/user/login", handleLogin)

	auth := r.Group("/api/user")
	auth.Use(authMiddleware())
	auth.GET("/info", handleUserInfo)

	admin := r.Group("/api/admin")
	admin.Use(authMiddleware(), platformAdminOnly())
	admin.GET("/tenants", handleListTenants)
	admin.POST("/tenants", handleCreateTenant)
	admin.PUT("/tenants/:id", handleUpdateTenant)
	admin.DELETE("/tenants/:id", handleDeleteTenant)
	admin.GET("/users", handleListAllUsers)
	admin.POST("/users", handleCreateUser)
	admin.PUT("/users/:id", handleUpdateUser)
	admin.DELETE("/users/:id", handleDeleteUser)

	tenant := r.Group("/api/tenant")
	tenant.Use(authMiddleware(), tenantAdminOnly())
	tenant.GET("/users", handleListTenantUsers)
	tenant.POST("/users", handleCreateTenantUser)
	tenant.PUT("/users/:id", handleUpdateTenantUser)
	tenant.DELETE("/users/:id", handleDeleteTenantUser)

	return r
}

func createToken(t *testing.T, user model.User) string {
	t.Helper()
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		TenantID: user.TenantID,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	require.NoError(t, err)
	return tokenStr
}

func TestHandleRegisterAndLogin(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	registerBody, _ := json.Marshal(map[string]string{
		"username": "alice",
		"password": "secret1",
		"email":    "alice@example.com",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	loginBody, _ := json.Marshal(map[string]string{
		"username": "alice",
		"password": "secret1",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.True(t, resp["success"].(bool))
	data := resp["data"].(map[string]interface{})
	require.NotEmpty(t, data["token"])
}

func TestHandleRegisterDuplicateUsername(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	body, _ := json.Marshal(map[string]string{
		"username": "bob",
		"password": "secret1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestHandleRegisterShortPassword(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	body, _ := json.Marshal(map[string]string{
		"username": "carol",
		"password": "123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleLoginInvalidCredentials(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := model.User{
		Username:     "admin",
		PasswordHash: string(hash),
		TenantID:     0,
		Role:         model.RolePlatformAdmin,
		Status:       1,
	}
	require.NoError(t, db.Create(&user).Error)

	body, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "wrong",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleUserInfoRequiresAuth(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/user/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleUserInfoWithToken(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	hash, _ := bcrypt.GenerateFromPassword([]byte("secret1"), bcrypt.DefaultCost)
	user := model.User{
		Username:     "dave",
		PasswordHash: string(hash),
		Email:        "dave@example.com",
		TenantID:     1,
		Role:         model.RoleViewer,
		Status:       1,
	}
	require.NoError(t, db.Create(&user).Error)

	token := createToken(t, user)
	req := httptest.NewRequest(http.MethodGet, "/api/user/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestPlatformAdminOnly(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	viewer := model.User{Username: "viewer", TenantID: 1, Role: model.RoleViewer, Status: 1}
	require.NoError(t, db.Create(&viewer).Error)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/tenants", nil)
	req.Header.Set("Authorization", "Bearer "+createToken(t, viewer))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	admin := model.User{Username: "platform", TenantID: 0, Role: model.RolePlatformAdmin, Status: 1}
	require.NoError(t, db.Create(&admin).Error)

	req = httptest.NewRequest(http.MethodGet, "/api/admin/tenants", nil)
	req.Header.Set("Authorization", "Bearer "+createToken(t, admin))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePlatformAdminAndTenantAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("role", model.RolePlatformAdmin)
	require.True(t, requirePlatformAdmin(c))

	c.Set("role", model.RoleAdmin)
	require.True(t, requireTenantAdmin(c))
	require.False(t, requirePlatformAdmin(c))
}

func TestHandleLoginDisabledAccount(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	hash, _ := bcrypt.GenerateFromPassword([]byte("secret1"), bcrypt.DefaultCost)
	user := model.User{
		Username:     "disabled",
		PasswordHash: string(hash),
		TenantID:     1,
		Role:         model.RoleViewer,
		Status:       1,
	}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Model(&user).Update("status", 0).Error)

	body, _ := json.Marshal(map[string]string{
		"username": "disabled",
		"password": "secret1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandleCreateTenantAsPlatformAdmin(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	admin := model.User{Username: "platform", TenantID: 0, Role: model.RolePlatformAdmin, Status: 1}
	require.NoError(t, db.Create(&admin).Error)
	token := createToken(t, admin)

	body, _ := json.Marshal(map[string]string{
		"name": "ACME",
		"code": "acme",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/tenants", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestSeedDataCreatesDefaultTenantAndAdmin(t *testing.T) {
	setupTestDB(t)
	seedData()

	var tenant model.Tenant
	require.NoError(t, db.Where("code = ?", "default").First(&tenant).Error)

	var admin model.User
	require.NoError(t, db.Where("username = ?", "admin").First(&admin).Error)
	require.Equal(t, model.RolePlatformAdmin, admin.Role)
}

func TestAuthMiddlewareRejectsInvalidToken(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/user/info", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTenantIDAndUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("tenantId", uint(9))
	c.Set("userId", uint(42))
	require.Equal(t, uint(9), getTenantID(c))
	require.Equal(t, uint(42), getUserID(c))
}

func createPlatformAdmin(t *testing.T) (model.User, string) {
	t.Helper()
	admin := model.User{Username: "platform", TenantID: 0, Role: model.RolePlatformAdmin, Status: 1}
	require.NoError(t, db.Create(&admin).Error)
	return admin, createToken(t, admin)
}

func createTenantAdmin(t *testing.T, tenantID uint) (model.User, string) {
	t.Helper()
	admin := model.User{Username: "tenantadmin", TenantID: tenantID, Role: model.RoleAdmin, Status: 1}
	require.NoError(t, db.Create(&admin).Error)
	return admin, createToken(t, admin)
}

func TestHandleListAllUsersAndCreateUser(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()
	_, token := createPlatformAdmin(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	body, _ := json.Marshal(map[string]interface{}{
		"username": "newuser",
		"password": "secret1",
		"tenantId": 1,
		"role":     model.RoleViewer,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDeleteTenantProtectsDefault(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()
	_, token := createPlatformAdmin(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/tenants/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandleDeleteNonDefaultTenant(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()
	_, token := createPlatformAdmin(t)

	tenant := model.Tenant{Name: "Temp", Code: "temp", Status: 1}
	require.NoError(t, db.Create(&tenant).Error)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/tenants/"+strconv.FormatUint(uint64(tenant.ID), 10), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestTenantAdminUserCRUD(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()
	_, token := createTenantAdmin(t, 1)

	req := httptest.NewRequest(http.MethodGet, "/api/tenant/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	body, _ := json.Marshal(map[string]string{
		"username": "employee",
		"password": "secret1",
		"email":    "emp@example.com",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/tenant/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var user model.User
	require.NoError(t, db.Where("username = ?", "employee").First(&user).Error)

	updateBody, _ := json.Marshal(map[string]string{"email": "updated@example.com"})
	req = httptest.NewRequest(http.MethodPut, "/api/tenant/users/"+strconv.FormatUint(uint64(user.ID), 10), bytes.NewReader(updateBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodDelete, "/api/tenant/users/"+strconv.FormatUint(uint64(user.ID), 10), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestHandleUpdateUserAsPlatformAdmin(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()
	_, token := createPlatformAdmin(t)

	user := model.User{Username: "target", TenantID: 1, Role: model.RoleViewer, Status: 1, Email: "old@example.com"}
	require.NoError(t, db.Create(&user).Error)

	body, _ := json.Marshal(map[string]string{"email": "new@example.com", "role": model.RoleUser})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/"+strconv.FormatUint(uint64(user.ID), 10), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestHandleUpdateTenant(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()
	_, token := createPlatformAdmin(t)

	body, _ := json.Marshal(map[string]string{"name": "Updated Default"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/tenants/1", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestTenantAdminCannotModifyOtherTenantUser(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()
	_, token := createTenantAdmin(t, 1)

	otherTenant := model.Tenant{Name: "Other", Code: "other", Status: 1}
	require.NoError(t, db.Create(&otherTenant).Error)
	otherUser := model.User{Username: "outsider", TenantID: otherTenant.ID, Role: model.RoleUser, Status: 1}
	require.NoError(t, db.Create(&otherUser).Error)

	body, _ := json.Marshal(map[string]string{"email": "hack@example.com"})
	req := httptest.NewRequest(http.MethodPut, "/api/tenant/users/"+strconv.FormatUint(uint64(otherUser.ID), 10), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}
