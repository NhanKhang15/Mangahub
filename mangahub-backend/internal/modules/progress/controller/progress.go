package controller

import (
	"mangahub-backend/internal/modules/progress/dto"

progressModel "mangahub-backend/internal/modules/progress/model"
	"mangahub-backend/internal/core/response"

	"net/http"

	"github.com/gin-gonic/gin"

	"mangahub-backend/internal/core/middleware"



	progressService "mangahub-backend/internal/modules/progress/service"
)

type ProgressHandler struct {
	svc *progressService.Service
}

func NewProgressHandler(s *progressService.Service) *ProgressHandler { return &ProgressHandler{svc: s} }


func (h *ProgressHandler) List(c *gin.Context) {
	uid := middleware.UserID(c)
	var q dto.ListProgressQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), uid, q.Status, q.Page, q.Limit)
	if err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "page": q.Page, "limit": q.Limit, "total": total})
}

func (h *ProgressHandler) Upsert(c *gin.Context) {
	uid := middleware.UserID(c)
	mangaID, ok := response.ParseObjectID(c, "mangaId")
	if !ok {
		return
	}
	var in dto.UpsertProgressInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}
	out, err := h.svc.Upsert(c.Request.Context(), &progressModel.ReadingProgress{
		UserID:         uid,
		MangaID:        mangaID,
		Status:         in.Status,
		CurrentChapter: in.CurrentChapter,
		Rating:         in.Rating,
	})
	if err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ProgressHandler) Delete(c *gin.Context) {
	uid := middleware.UserID(c)
	mangaID, ok := response.ParseObjectID(c, "mangaId")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uid, mangaID); err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProgressHandler) Stats(c *gin.Context) {
	uid := middleware.UserID(c)
	stats, err := h.svc.Stats(c.Request.Context(), uid)
	if err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}
