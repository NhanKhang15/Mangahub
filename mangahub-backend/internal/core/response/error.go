package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	artistService "mangahub-backend/internal/modules/artist/service"
	mangaService "mangahub-backend/internal/modules/manga/service"
	progressService "mangahub-backend/internal/modules/progress/service"
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
	case errors.Is(err, mangaService.ErrNotFound),
		errors.Is(err, artistService.ErrNotFound),
		errors.Is(err, progressService.ErrNotFound):
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
