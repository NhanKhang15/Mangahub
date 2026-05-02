package controller

import (
	"mangahub-backend/internal/modules/manga/dto"

mangaModel "mangahub-backend/internal/modules/manga/model"
	"mangahub-backend/internal/core/response"

	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"




	mangaService "mangahub-backend/internal/modules/manga/service"
)

type MangaHandler struct {
	svc *mangaService.Service
}

func NewMangaHandler(s *mangaService.Service) *MangaHandler { return &MangaHandler{svc: s} }


func (h *MangaHandler) List(c *gin.Context) {
	var q dto.ListMangaQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	dq := mangaModel.MangaListQuery{
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
		response.RespondDomainError(c, err)
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
	id, ok := response.ParseObjectID(c, "id")
	if !ok {
		return
	}
	m, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *MangaHandler) Create(c *gin.Context) {
	var in dto.CreateMangaInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
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

	m := &mangaModel.Manga{
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
		response.RespondDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *MangaHandler) Update(c *gin.Context) {
	id, ok := response.ParseObjectID(c, "id")
	if !ok {
		return
	}
	var in dto.UpdateMangaInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
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

func (h *MangaHandler) Delete(c *gin.Context) {
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

func parseObjectIDList(c *gin.Context, field string, raw []string) ([]primitive.ObjectID, bool) {
	if len(raw) == 0 {
		return nil, true
	}
	out := make([]primitive.ObjectID, 0, len(raw))
	for _, s := range raw {
		id, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			response.RespondError(c, http.StatusBadRequest, "INVALID_ID", "invalid object id in "+field, map[string]any{"value": s})
			return nil, false
		}
		out = append(out, id)
	}
	return out, true
}
