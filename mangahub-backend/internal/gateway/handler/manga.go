package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"mangahub-backend/internal/catalog"
	"mangahub-backend/internal/domain"
)

type MangaHandler struct {
	svc *catalog.Service
}

func NewMangaHandler(s *catalog.Service) *MangaHandler { return &MangaHandler{svc: s} }

type CreateMangaInput struct {
	Title       string   `json:"title"        binding:"required,min=1,max=300"`
	AltTitles   []string `json:"alt_titles"`
	ArtistIDs   []string `json:"artist_ids"`
	AuthorIDs   []string `json:"author_ids"`
	Description string   `json:"description"`
	Status      string   `json:"status"       binding:"required,oneof=ongoing completed hiatus"`
	Genres      []string `json:"genres"       binding:"required,min=1,dive,min=1"`
	Tags        []string `json:"tags"`
	Chapters    int      `json:"chapters"     binding:"gte=0"`
	Rating      float64  `json:"rating"       binding:"gte=1,lte=10"`
	CoverURL    string   `json:"cover_url"`
	Popularity  int      `json:"popularity"   binding:"gte=0"`
}

type UpdateMangaInput struct {
	Title       *string   `json:"title"        binding:"omitempty,min=1,max=300"`
	AltTitles   *[]string `json:"alt_titles"`
	ArtistIDs   *[]string `json:"artist_ids"`
	AuthorIDs   *[]string `json:"author_ids"`
	Description *string   `json:"description"`
	Status      *string   `json:"status"       binding:"omitempty,oneof=ongoing completed hiatus"`
	Genres      *[]string `json:"genres"       binding:"omitempty,min=1"`
	Tags        *[]string `json:"tags"`
	Chapters    *int      `json:"chapters"     binding:"omitempty,gte=0"`
	Rating      *float64  `json:"rating"       binding:"omitempty,gte=1,lte=10"`
	CoverURL    *string   `json:"cover_url"`
	Popularity  *int      `json:"popularity"   binding:"omitempty,gte=0"`
}

type ListMangaQuery struct {
	Page  int    `form:"page,default=1"   binding:"gte=1"`
	Limit int    `form:"limit,default=20" binding:"gte=1,lte=100"`
	Genre string `form:"genre"`
	Tags  string `form:"tags"`  // comma-separated
	Q     string `form:"q"`
	Sort  string `form:"sort"`
}

func (h *MangaHandler) List(c *gin.Context) {
	var q ListMangaQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	dq := domain.MangaListQuery{
		Page:  q.Page,
		Limit: q.Limit,
		Genre: q.Genre,
		Q:     q.Q,
		Sort:  q.Sort,
	}
	if q.Tags != "" {
		for _, t := range strings.Split(q.Tags, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				dq.Tags = append(dq.Tags, t)
			}
		}
	}

	items, total, err := h.svc.List(c.Request.Context(), dq)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"page":  q.Page,
		"limit": q.Limit,
		"total": total,
	})
}

func (h *MangaHandler) Get(c *gin.Context) {
	id, ok := ParseObjectID(c, "id")
	if !ok {
		return
	}
	m, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *MangaHandler) Create(c *gin.Context) {
	var in CreateMangaInput
	if err := c.ShouldBindJSON(&in); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}

	artistIDs, ok := parseObjectIDList(c, "artist_ids", in.ArtistIDs)
	if !ok {
		return
	}
	authorIDs, ok := parseObjectIDList(c, "author_ids", in.AuthorIDs)
	if !ok {
		return
	}

	m := &domain.Manga{
		Title:       in.Title,
		AltTitles:   in.AltTitles,
		ArtistIDs:   artistIDs,
		AuthorIDs:   authorIDs,
		Description: in.Description,
		Status:      in.Status,
		Genres:      in.Genres,
		Tags:        in.Tags,
		Chapters:    in.Chapters,
		Rating:      in.Rating,
		CoverURL:    in.CoverURL,
		Popularity:  in.Popularity,
	}
	out, err := h.svc.Create(c.Request.Context(), m)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *MangaHandler) Update(c *gin.Context) {
	id, ok := ParseObjectID(c, "id")
	if !ok {
		return
	}
	var in UpdateMangaInput
	if err := c.ShouldBindJSON(&in); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}

	set := bson.M{}
	if in.Title != nil {
		set["title"] = *in.Title
	}
	if in.AltTitles != nil {
		set["alt_titles"] = *in.AltTitles
	}
	if in.ArtistIDs != nil {
		ids, ok := parseObjectIDList(c, "artist_ids", *in.ArtistIDs)
		if !ok {
			return
		}
		set["artist_ids"] = ids
	}
	if in.AuthorIDs != nil {
		ids, ok := parseObjectIDList(c, "author_ids", *in.AuthorIDs)
		if !ok {
			return
		}
		set["author_ids"] = ids
	}
	if in.Description != nil {
		set["description"] = *in.Description
	}
	if in.Status != nil {
		set["status"] = *in.Status
	}
	if in.Genres != nil {
		set["genres"] = *in.Genres
	}
	if in.Tags != nil {
		set["tags"] = *in.Tags
	}
	if in.Chapters != nil {
		set["chapters"] = *in.Chapters
	}
	if in.Rating != nil {
		set["rating"] = *in.Rating
	}
	if in.CoverURL != nil {
		set["cover_url"] = *in.CoverURL
	}
	if in.Popularity != nil {
		set["popularity"] = *in.Popularity
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

func (h *MangaHandler) Delete(c *gin.Context) {
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

func parseObjectIDList(c *gin.Context, field string, raw []string) ([]primitive.ObjectID, bool) {
	if len(raw) == 0 {
		return nil, true
	}
	out := make([]primitive.ObjectID, 0, len(raw))
	for _, s := range raw {
		id, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "INVALID_ID", "invalid object id in "+field, map[string]any{"value": s})
			return nil, false
		}
		out = append(out, id)
	}
	return out, true
}
