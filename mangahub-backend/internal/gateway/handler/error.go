package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"mangahub-backend/internal/artist"
	"mangahub-backend/internal/catalog"
	"mangahub-backend/internal/progress"
)

type ErrorBody struct {
	Error   string         `json:"error"`
	Code    string         `json:"code"`
	Details map[string]any `json:"details,omitempty"`
}

func RespondError(c *gin.Context, status int, code, msg string, details map[string]any) {
	c.AbortWithStatusJSON(status, ErrorBody{Error: msg, Code: code, Details: details})
}

func RespondDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, catalog.ErrNotFound),
		errors.Is(err, artist.ErrNotFound),
		errors.Is(err, progress.ErrNotFound):
		RespondError(c, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
	default:
		RespondError(c, http.StatusInternalServerError, "INTERNAL", "internal error", map[string]any{"cause": err.Error()})
	}
}

func ParseObjectID(c *gin.Context, name string) (primitive.ObjectID, bool) {
	raw := c.Param(name)
	id, err := primitive.ObjectIDFromHex(raw)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", "invalid object id", map[string]any{"param": name, "value": raw})
		return primitive.NilObjectID, false
	}
	return id, true
}
