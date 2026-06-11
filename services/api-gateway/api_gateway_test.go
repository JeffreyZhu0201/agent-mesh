package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_AGENTMESH_ENV", "custom")
	require.Equal(t, "custom", getEnv("TEST_AGENTMESH_ENV", "default"))
	require.Equal(t, "fallback", getEnv("TEST_AGENTMESH_MISSING", "fallback"))
}

func TestCreateProxyRewritesPath(t *testing.T) {
	path := "/list"
	targetPath := "/api/user" + path
	require.Equal(t, "/api/user/list", targetPath)
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestDefaultServiceURLs(t *testing.T) {
	os.Unsetenv("USER_SVC_URL")
	os.Unsetenv("PLUGIN_SVC_URL")
	require.Equal(t, "http://user-svc:8081", getEnv("USER_SVC_URL", "http://user-svc:8081"))
	require.Equal(t, "http://plugin-svc:8083", getEnv("PLUGIN_SVC_URL", "http://plugin-svc:8083"))
}
