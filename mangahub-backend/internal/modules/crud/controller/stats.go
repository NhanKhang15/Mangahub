package controller

import (
	"mangahub-backend/internal/modules/crud/dto"

	"mangahub-backend/internal/core/response"

	"net/http"

	"github.com/gin-gonic/gin"

	mangaService "mangahub-backend/internal/modules/manga/service"
)

type StatsHandler struct {
	mangaSvc *mangaService.Service
}

func NewStatsHandler(m *mangaService.Service) *StatsHandler { return &StatsHandler{mangaSvc: m} }



func (h *StatsHandler) Popular(c *gin.Context) {
	var q dto.LimitQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	items, err := h.mangaSvc.Popular(c.Request.Context(), q.Limit)
	if err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *StatsHandler) Trending(c *gin.Context) {
	var q dto.LimitQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	items, err := h.mangaSvc.Trending(c.Request.Context(), q.Limit)
	if err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
