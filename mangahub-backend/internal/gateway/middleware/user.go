package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ctxUserID = "user_id"

// RequireUser is a placeholder until Phase 2 introduces JWT.
// It expects header "X-User-ID: <hex objectid>" so that /me/* endpoints can
// be tested before authentication is wired in.
func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("X-User-ID")
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing X-User-ID header (placeholder until JWT in Phase 2)",
				"code":  "UNAUTHORIZED",
			})
			return
		}
		id, err := primitive.ObjectIDFromHex(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "X-User-ID is not a valid ObjectID hex",
				"code":  "UNAUTHORIZED",
			})
			return
		}
		c.Set(ctxUserID, id)
		c.Next()
	}
}

func UserID(c *gin.Context) primitive.ObjectID {
	v, _ := c.Get(ctxUserID)
	id, _ := v.(primitive.ObjectID)
	return id
}
