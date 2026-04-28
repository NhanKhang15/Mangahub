package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type HealthHandler struct {
	mongo *mongo.Client
}

func NewHealthHandler(m *mongo.Client) *HealthHandler { return &HealthHandler{mongo: m} }

func (h *HealthHandler) Healthz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := "ok"
	deps := gin.H{"mongo": "ok"}
	if err := h.mongo.Ping(ctx, nil); err != nil {
		status = "degraded"
		deps["mongo"] = err.Error()
	}

	httpCode := http.StatusOK
	if status != "ok" {
		httpCode = http.StatusServiceUnavailable
	}
	c.JSON(httpCode, gin.H{"status": status, "deps": deps})
}
