package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TenantMiddleware handles tenant extraction from headers
type TenantMiddleware struct {
	DefaultTenantID string
}

// NewTenantMiddleware creates a new TenantMiddleware instance
func NewTenantMiddleware(defaultTenantID string) *TenantMiddleware {
	return &TenantMiddleware{
		DefaultTenantID: defaultTenantID,
	}
}

// Handler returns a gin.HandlerFunc for tenant extraction
func (t *TenantMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")

		// Use default tenant if not provided (for development)
		if tenantID == "" {
			tenantID = t.DefaultTenantID
		}

		// Set tenant_id in gin.Context
		c.Set("tenant_id", tenantID)

		c.Next()
	}
}

// RequireTenant returns a middleware that requires a valid tenant ID
func (t *TenantMiddleware) RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")

		if tenantID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "X-Tenant-ID header required",
			})
			return
		}

		// Set tenant_id in gin.Context
		c.Set("tenant_id", tenantID)

		c.Next()
	}
}
