package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"mangahub-backend/internal/core/response"
)

type HealthHandler struct {
	mc *mongo.Client
}

func NewHealthHandler(mc *mongo.Client) *HealthHandler { return &HealthHandler{mc: mc} }

func (h *HealthHandler) Healthz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.mc.Ping(ctx, readpref.Primary()); err != nil {
		response.RespondError(c, http.StatusServiceUnavailable, "DB_DOWN", err.Error(), nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
