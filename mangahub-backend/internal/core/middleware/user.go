package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ctxUserID = "user_id"

type TokenVerifier interface {
	VerifyToken(tokenString string) (string, error)
}

func RequireUser(verifier TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing Authorization header",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid Authorization header format",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		tokenString := parts[1]
		userIDStr, err := verifier.VerifyToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		id, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user id in token",
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
