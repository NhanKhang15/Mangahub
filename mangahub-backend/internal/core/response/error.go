package response

import (
	"github.com/gin-gonic/gin"
)

type ErrorBody struct {
	Error   string         `json:"error"`
	Code    string         `json:"code"`
	Details map[string]any `json:"details,omitempty"`
}

func RespondError(c *gin.Context, status int, code, msg string, details map[string]any) {
	c.AbortWithStatusJSON(status, ErrorBody{Error: msg, Code: code, Details: details})
}
