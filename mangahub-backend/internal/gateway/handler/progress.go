package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mangahub-backend/internal/domain"
	"mangahub-backend/internal/gateway/middleware"
	"mangahub-backend/internal/progress"
)

type ProgressHandler struct {
	svc *progress.Service
}

func NewProgressHandler(s *progress.Service) *ProgressHandler { return &ProgressHandler{svc: s} }

type UpsertProgressInput struct {
	Status         string  `json:"status"          binding:"required,oneof=reading completed plan_to_read dropped"`
	CurrentChapter int     `json:"current_chapter" binding:"gte=0"`
	Rating         float64 `json:"rating"          binding:"omitempty,gte=1,lte=10"`
}

type ListProgressQuery struct {
	Page   int    `form:"page,default=1"   binding:"gte=1"`
	Limit  int    `form:"limit,default=20" binding:"gte=1,lte=100"`
	Status string `form:"status"           binding:"omitempty,oneof=reading completed plan_to_read dropped"`
}

func (h *ProgressHandler) List(c *gin.Context) {
	uid := middleware.UserID(c)
	var q ListProgressQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), uid, q.Status, q.Page, q.Limit)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "page": q.Page, "limit": q.Limit, "total": total})
}

func (h *ProgressHandler) Upsert(c *gin.Context) {
	uid := middleware.UserID(c)
	mangaID, ok := ParseObjectID(c, "mangaId")
	if !ok {
		return
	}
	var in UpsertProgressInput
	if err := c.ShouldBindJSON(&in); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}
	out, err := h.svc.Upsert(c.Request.Context(), &domain.ReadingProgress{
		UserID:         uid,
		MangaID:        mangaID,
		Status:         in.Status,
		CurrentChapter: in.CurrentChapter,
		Rating:         in.Rating,
	})
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ProgressHandler) Delete(c *gin.Context) {
	uid := middleware.UserID(c)
	mangaID, ok := ParseObjectID(c, "mangaId")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uid, mangaID); err != nil {
		RespondDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProgressHandler) Stats(c *gin.Context) {
	uid := middleware.UserID(c)
	stats, err := h.svc.Stats(c.Request.Context(), uid)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}
