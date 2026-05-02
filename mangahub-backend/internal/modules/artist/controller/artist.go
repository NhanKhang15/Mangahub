package controller

import (
	"mangahub-backend/internal/modules/artist/dto"

artistModel "mangahub-backend/internal/modules/artist/model"
	"mangahub-backend/internal/core/response"

	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"




	artistService "mangahub-backend/internal/modules/artist/service"
	mangaService "mangahub-backend/internal/modules/manga/service"
)

type ArtistHandler struct {
	svc       *artistService.Service
	mangaSvc  *mangaService.Service
}

func NewArtistHandler(s *artistService.Service, m *mangaService.Service) *ArtistHandler {
	return &ArtistHandler{svc: s, mangaSvc: m}
}


func (h *ArtistHandler) List(c *gin.Context) {
	var q dto.ListArtistQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), q.Q, q.Page, q.Limit)
	if err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "page": q.Page, "limit": q.Limit, "total": total})
}

func (h *ArtistHandler) Get(c *gin.Context) {
	id, ok := response.ParseObjectID(c, "id")
	if !ok {
		return
	}
	a, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *ArtistHandler) Create(c *gin.Context) {
	var in dto.CreateArtistInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}
	a := &artistModel.Artist{Name: in.Name, Role: in.Role, Bio: in.Bio}
	out, err := h.svc.Create(c.Request.Context(), a)
	if err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *ArtistHandler) Update(c *gin.Context) {
	id, ok := response.ParseObjectID(c, "id")
	if !ok {
		return
	}
	var in dto.UpdateArtistInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}
	set := bson.M{}
	if in.Name != nil {
		set["name"] = *in.Name
	}
	if in.Role != nil {
		set["role"] = *in.Role
	}
	if in.Bio != nil {
		set["bio"] = *in.Bio
	}
	if len(set) == 0 {
		response.RespondError(c, http.StatusBadRequest, "NO_FIELDS", "request body has no updatable fields", nil)
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, set)
	if err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ArtistHandler) Delete(c *gin.Context) {
	id, ok := response.ParseObjectID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}


func (h *ArtistHandler) ListMangaByArtist(c *gin.Context) {
	id, ok := response.ParseObjectID(c, "id")
	if !ok {
		return
	}
	var q dto.ListArtistMangaQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	items, total, err := h.mangaSvc.ListByArtist(c.Request.Context(), id, q.Page, q.Limit)
	if err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "page": q.Page, "limit": q.Limit, "total": total})
}
