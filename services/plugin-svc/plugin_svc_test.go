package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPluginTestDB(t *testing.T) {
	t.Helper()
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	migrate(testDB)
	db = testDB
	seedPlugins(db)
}

func createPluginToken(t *testing.T, tenantID uint) string {
	t.Helper()
	claims := &Claims{
		UserID:   1,
		Username: "tester",
		TenantID: tenantID,
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	require.NoError(t, err)
	return tokenStr
}

func setupPluginRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerRoutes(r)
	return r
}

func TestHealthEndpoint(t *testing.T) {
	setupPluginTestDB(t)
	r := setupPluginRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestPluginListDefaultsToEnabled(t *testing.T) {
	setupPluginTestDB(t)
	r := setupPluginRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/plugin/list", nil)
	req.Header.Set("Authorization", "Bearer "+createPluginToken(t, 1))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	plugins := resp["plugins"].([]interface{})
	require.Len(t, plugins, 2)
}

func TestPluginToggleDisablesPlugin(t *testing.T) {
	setupPluginTestDB(t)
	r := setupPluginRouter()
	token := createPluginToken(t, 1)

	body, _ := json.Marshal(map[string]interface{}{
		"pluginKey": "dashboard",
		"enabled":   false,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/plugin/admin/toggle", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/plugin/list", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	plugins := resp["plugins"].([]interface{})
	require.Len(t, plugins, 1)
}

func TestPluginAdminListShowsDisabledState(t *testing.T) {
	setupPluginTestDB(t)
	r := setupPluginRouter()
	token := createPluginToken(t, 1)

	body, _ := json.Marshal(map[string]interface{}{
		"pluginKey": "chat",
		"enabled":   false,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/plugin/admin/toggle", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/plugin/admin/list", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	plugins := resp["plugins"].([]interface{})
	require.Len(t, plugins, 2)

	disabledCount := 0
	for _, item := range plugins {
		plugin := item.(map[string]interface{})
		if plugin["key"] == "chat" {
			require.False(t, plugin["enabled"].(bool))
			disabledCount++
		}
	}
	require.Equal(t, 1, disabledCount)
}

func TestPluginToggleNotFound(t *testing.T) {
	setupPluginTestDB(t)
	r := setupPluginRouter()

	body, _ := json.Marshal(map[string]interface{}{
		"pluginKey": "missing",
		"enabled":   false,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/plugin/admin/toggle", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createPluginToken(t, 1))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	setupPluginTestDB(t)
	r := setupPluginRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/plugin/list", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
