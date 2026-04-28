package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mangahub-backend/internal/catalog"
)

type StatsHandler struct {
	mangaSvc *catalog.Service
}

func NewStatsHandler(m *catalog.Service) *StatsHandler { return &StatsHandler{mangaSvc: m} }

type LimitQuery struct {
	Limit int `form:"limit,default=10" binding:"gte=1,lte=100"`
}

func (h *StatsHandler) Popular(c *gin.Context) {
	var q LimitQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	items, err := h.mangaSvc.Popular(c.Request.Context(), q.Limit)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *StatsHandler) Trending(c *gin.Context) {
	var q LimitQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	items, err := h.mangaSvc.Trending(c.Request.Context(), q.Limit)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
