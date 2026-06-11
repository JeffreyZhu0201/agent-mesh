package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddlewarePublicPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleware("secret")

	r := gin.New()
	r.GET("/health", mw.Handler(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	mw := NewAuthMiddleware(secret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		UserID:   "1",
		TenantID: "2",
		Username: "alice",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	r := gin.New()
	r.GET("/protected", mw.Handler(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"userId":   GetUserID(c),
			"tenantId": GetTenantID(c),
			"username": GetUsername(c),
			"role":     GetRole(c),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddlewareMissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleware("secret")

	r := gin.New()
	r.GET("/protected", mw.Handler(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestContextGettersReturnEmptyWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	require.Empty(t, GetUserID(c))
	require.Empty(t, GetTenantID(c))
	require.Empty(t, GetUsername(c))
	require.Empty(t, GetRole(c))
}
