package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"mangahub-backend/internal/artist"
	"mangahub-backend/internal/catalog"
	"mangahub-backend/internal/domain"
)

type ArtistHandler struct {
	svc       *artist.Service
	mangaSvc  *catalog.Service
}

func NewArtistHandler(s *artist.Service, m *catalog.Service) *ArtistHandler {
	return &ArtistHandler{svc: s, mangaSvc: m}
}

type CreateArtistInput struct {
	Name string   `json:"name" binding:"required,min=1,max=200"`
	Role string   `json:"role" binding:"required,oneof=artist author both"`
	Bio  string   `json:"bio"`
}

type UpdateArtistInput struct {
	Name *string `json:"name" binding:"omitempty,min=1,max=200"`
	Role *string `json:"role" binding:"omitempty,oneof=artist author both"`
	Bio  *string `json:"bio"`
}

type ListArtistQuery struct {
	Page  int    `form:"page,default=1"   binding:"gte=1"`
	Limit int    `form:"limit,default=20" binding:"gte=1,lte=100"`
	Q     string `form:"q"`
}

func (h *ArtistHandler) List(c *gin.Context) {
	var q ListArtistQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), q.Q, q.Page, q.Limit)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "page": q.Page, "limit": q.Limit, "total": total})
}

func (h *ArtistHandler) Get(c *gin.Context) {
	id, ok := ParseObjectID(c, "id")
	if !ok {
		return
	}
	a, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *ArtistHandler) Create(c *gin.Context) {
	var in CreateArtistInput
	if err := c.ShouldBindJSON(&in); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}
	a := &domain.Artist{Name: in.Name, Role: in.Role, Bio: in.Bio}
	out, err := h.svc.Create(c.Request.Context(), a)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *ArtistHandler) Update(c *gin.Context) {
	id, ok := ParseObjectID(c, "id")
	if !ok {
		return
	}
	var in UpdateArtistInput
	if err := c.ShouldBindJSON(&in); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
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
		RespondError(c, http.StatusBadRequest, "NO_FIELDS", "request body has no updatable fields", nil)
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, set)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ArtistHandler) Delete(c *gin.Context) {
	id, ok := ParseObjectID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		RespondDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type ListArtistMangaQuery struct {
	Page  int `form:"page,default=1"   binding:"gte=1"`
	Limit int `form:"limit,default=20" binding:"gte=1,lte=100"`
}

func (h *ArtistHandler) ListMangaByArtist(c *gin.Context) {
	id, ok := ParseObjectID(c, "id")
	if !ok {
		return
	}
	var q ListArtistMangaQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	items, total, err := h.mangaSvc.ListByArtist(c.Request.Context(), id, q.Page, q.Limit)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "page": q.Page, "limit": q.Limit, "total": total})
}
