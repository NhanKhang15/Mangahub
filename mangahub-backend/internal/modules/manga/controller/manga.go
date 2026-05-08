package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mangahub-backend/internal/core/response"
	"mangahub-backend/internal/gateway/grpcclient"
	mangaGrpc "mangahub-backend/internal/modules/manga/grpcserver"
	mangaModel "mangahub-backend/internal/modules/manga/model"

	"mangahub-backend/internal/modules/manga/dto"
	catalogpb "mangahub-backend/proto/catalogpb"
)

type MangaHandler struct {
	client catalogpb.MangaCatalogClient
}

func NewMangaHandler(c catalogpb.MangaCatalogClient) *MangaHandler {
	return &MangaHandler{client: c}
}

func (h *MangaHandler) List(c *gin.Context) {
	var q dto.ListMangaQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	tags := splitCSV(q.Tags)
	resp, err := h.client.ListManga(c.Request.Context(), &catalogpb.ListMangaRequest{
		Page:  int32(q.Page),
		Limit: int32(q.Limit),
		Genre: q.Genre,
		Tags:  tags,
		Q:     q.Q,
		Sort:  q.Sort,
	})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  protoMangasToModels(resp.GetItems()),
		"page":  q.Page,
		"limit": q.Limit,
		"total": resp.GetTotal(),
	})
}

func (h *MangaHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.RespondError(c, http.StatusBadRequest, "INVALID_ID", "missing manga id", nil)
		return
	}
	resp, err := h.client.GetManga(c.Request.Context(), &catalogpb.GetMangaRequest{Id: id})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	m, convErr := mangaGrpc.MangaFromProto(resp)
	if convErr != nil {
		response.RespondError(c, http.StatusInternalServerError, "INTERNAL", convErr.Error(), nil)
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
	if !validateHexList(c, "artist_ids", in.ArtistIDs) {
		return
	}
	if !validateHexList(c, "author_ids", in.AuthorIDs) {
		return
	}

	pm := &catalogpb.Manga{
		Title:       in.Title,
		AltTitles:   in.AltTitles,
		ArtistIds:   in.ArtistIDs,
		AuthorIds:   in.AuthorIDs,
		Description: in.Description,
		Status:      in.Status,
		Genres:      in.Genres,
		Tags:        in.Tags,
		Chapters:    int32(in.Chapters),
		Rating:      in.Rating,
		CoverUrl:    in.CoverURL,
		Popularity:  int32(in.Popularity),
	}
	resp, err := h.client.CreateManga(c.Request.Context(), &catalogpb.CreateMangaRequest{Manga: pm})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	m, convErr := mangaGrpc.MangaFromProto(resp)
	if convErr != nil {
		response.RespondError(c, http.StatusInternalServerError, "INTERNAL", convErr.Error(), nil)
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *MangaHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.RespondError(c, http.StatusBadRequest, "INVALID_ID", "missing manga id", nil)
		return
	}
	var in dto.UpdateMangaInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}

	patch := &catalogpb.MangaPatch{}
	hasField := false
	if in.Title != nil {
		patch.Title, patch.TitleSet = *in.Title, true
		hasField = true
	}
	if in.AltTitles != nil {
		patch.AltTitles, patch.AltTitlesSet = *in.AltTitles, true
		hasField = true
	}
	if in.ArtistIDs != nil {
		if !validateHexList(c, "artist_ids", *in.ArtistIDs) {
			return
		}
		patch.ArtistIds, patch.ArtistIdsSet = *in.ArtistIDs, true
		hasField = true
	}
	if in.AuthorIDs != nil {
		if !validateHexList(c, "author_ids", *in.AuthorIDs) {
			return
		}
		patch.AuthorIds, patch.AuthorIdsSet = *in.AuthorIDs, true
		hasField = true
	}
	if in.Description != nil {
		patch.Description, patch.DescriptionSet = *in.Description, true
		hasField = true
	}
	if in.Status != nil {
		patch.Status, patch.StatusSet = *in.Status, true
		hasField = true
	}
	if in.Genres != nil {
		patch.Genres, patch.GenresSet = *in.Genres, true
		hasField = true
	}
	if in.Tags != nil {
		patch.Tags, patch.TagsSet = *in.Tags, true
		hasField = true
	}
	if in.Chapters != nil {
		patch.Chapters, patch.ChaptersSet = int32(*in.Chapters), true
		hasField = true
	}
	if in.Rating != nil {
		patch.Rating, patch.RatingSet = *in.Rating, true
		hasField = true
	}
	if in.CoverURL != nil {
		patch.CoverUrl, patch.CoverUrlSet = *in.CoverURL, true
		hasField = true
	}
	if in.Popularity != nil {
		patch.Popularity, patch.PopularitySet = int32(*in.Popularity), true
		hasField = true
	}
	if !hasField {
		response.RespondError(c, http.StatusBadRequest, "NO_FIELDS", "request body has no updatable fields", nil)
		return
	}

	resp, err := h.client.UpdateManga(c.Request.Context(), &catalogpb.UpdateMangaRequest{
		Id:    id,
		Patch: patch,
	})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	m, convErr := mangaGrpc.MangaFromProto(resp)
	if convErr != nil {
		response.RespondError(c, http.StatusInternalServerError, "INTERNAL", convErr.Error(), nil)
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *MangaHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.RespondError(c, http.StatusBadRequest, "INVALID_ID", "missing manga id", nil)
		return
	}
	_, err := h.client.DeleteManga(c.Request.Context(), &catalogpb.DeleteMangaRequest{Id: id})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	c.Status(http.StatusNoContent)
}

func splitCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	out := []string{}
	for _, t := range strings.Split(raw, ",") {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func validateHexList(c *gin.Context, field string, ids []string) bool {
	for _, s := range ids {
		if _, err := mangaGrpc.ValidateHex(s); err != nil {
			response.RespondError(c, http.StatusBadRequest, "INVALID_ID", "invalid object id in "+field, map[string]any{"value": s})
			return false
		}
	}
	return true
}

func protoMangasToModels(items []*catalogpb.Manga) []*mangaModel.Manga {
	out := make([]*mangaModel.Manga, 0, len(items))
	for _, p := range items {
		m, err := mangaGrpc.MangaFromProto(p)
		if err != nil {
			continue
		}
		out = append(out, m)
	}
	return out
}
