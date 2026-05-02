package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireAdmin gates admin-only endpoints. For Phase 3 lab scope an admin
// token is supplied via the ADMIN_TOKEN env var (captured in cfg) and matched
// against the X-Admin-Token request header. Phase 2 will replace this with
// JWT-based role claims.
func RequireAdmin(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "ADMIN_TOKEN is not configured on the server",
				"code":  "ADMIN_DISABLED",
			})
			return
		}
		if c.GetHeader("X-Admin-Token") != token {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "admin token invalid or missing",
				"code":  "FORBIDDEN",
			})
			return
		}
		c.Next()
	}
}
