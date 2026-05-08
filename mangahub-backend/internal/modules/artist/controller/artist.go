package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mangahub-backend/internal/core/response"
	"mangahub-backend/internal/gateway/grpcclient"
	artistGrpc "mangahub-backend/internal/modules/artist/grpcserver"
	artistModel "mangahub-backend/internal/modules/artist/model"

	"mangahub-backend/internal/modules/artist/dto"
	artistpb "mangahub-backend/proto/artistpb"
)

type ArtistHandler struct {
	client artistpb.ArtistClient
}

func NewArtistHandler(c artistpb.ArtistClient) *ArtistHandler {
	return &ArtistHandler{client: c}
}

func (h *ArtistHandler) List(c *gin.Context) {
	var q dto.ListArtistQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	resp, err := h.client.ListArtist(c.Request.Context(), &artistpb.ListArtistRequest{
		Page:  int32(q.Page),
		Limit: int32(q.Limit),
		Q:     q.Q,
	})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  protoArtistsToModels(resp.GetItems()),
		"page":  q.Page,
		"limit": q.Limit,
		"total": resp.GetTotal(),
	})
}

func (h *ArtistHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.RespondError(c, http.StatusBadRequest, "INVALID_ID", "missing artist id", nil)
		return
	}
	resp, err := h.client.GetArtist(c.Request.Context(), &artistpb.GetArtistRequest{Id: id})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	a, convErr := artistGrpc.ArtistFromProto(resp)
	if convErr != nil {
		response.RespondError(c, http.StatusInternalServerError, "INTERNAL", convErr.Error(), nil)
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
	resp, err := h.client.CreateArtist(c.Request.Context(), &artistpb.CreateArtistRequest{
		Artist: &artistpb.ArtistEntity{
			Name: in.Name,
			Role: in.Role,
			Bio:  in.Bio,
		},
	})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	a, convErr := artistGrpc.ArtistFromProto(resp)
	if convErr != nil {
		response.RespondError(c, http.StatusInternalServerError, "INTERNAL", convErr.Error(), nil)
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *ArtistHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.RespondError(c, http.StatusBadRequest, "INVALID_ID", "missing artist id", nil)
		return
	}
	var in dto.UpdateArtistInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}
	patch := &artistpb.ArtistPatch{}
	hasField := false
	if in.Name != nil {
		patch.Name, patch.NameSet = *in.Name, true
		hasField = true
	}
	if in.Role != nil {
		patch.Role, patch.RoleSet = *in.Role, true
		hasField = true
	}
	if in.Bio != nil {
		patch.Bio, patch.BioSet = *in.Bio, true
		hasField = true
	}
	if !hasField {
		response.RespondError(c, http.StatusBadRequest, "NO_FIELDS", "request body has no updatable fields", nil)
		return
	}
	resp, err := h.client.UpdateArtist(c.Request.Context(), &artistpb.UpdateArtistRequest{
		Id:    id,
		Patch: patch,
	})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	a, convErr := artistGrpc.ArtistFromProto(resp)
	if convErr != nil {
		response.RespondError(c, http.StatusInternalServerError, "INTERNAL", convErr.Error(), nil)
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *ArtistHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.RespondError(c, http.StatusBadRequest, "INVALID_ID", "missing artist id", nil)
		return
	}
	_, err := h.client.DeleteArtist(c.Request.Context(), &artistpb.DeleteArtistRequest{Id: id})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	c.Status(http.StatusNoContent)
}

// ListMangaByArtist hits artist-svc, which in turn calls catalog-svc internally.
// The gateway never touches the manga collection on this path.
func (h *ArtistHandler) ListMangaByArtist(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.RespondError(c, http.StatusBadRequest, "INVALID_ID", "missing artist id", nil)
		return
	}
	var q dto.ListArtistMangaQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	resp, err := h.client.ListArtistManga(c.Request.Context(), &artistpb.ListArtistMangaRequest{
		ArtistId: id,
		Page:     int32(q.Page),
		Limit:    int32(q.Limit),
	})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	items := make([]gin.H, 0, len(resp.GetItems()))
	for _, m := range resp.GetItems() {
		items = append(items, gin.H{
			"id":         m.GetId(),
			"title":      m.GetTitle(),
			"cover_url":  m.GetCoverUrl(),
			"status":     m.GetStatus(),
			"chapters":   m.GetChapters(),
			"rating":     m.GetRating(),
			"popularity": m.GetPopularity(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"page":  q.Page,
		"limit": q.Limit,
		"total": resp.GetTotal(),
	})
}

func protoArtistsToModels(items []*artistpb.ArtistEntity) []*artistModel.Artist {
	out := make([]*artistModel.Artist, 0, len(items))
	for _, p := range items {
		a, err := artistGrpc.ArtistFromProto(p)
		if err != nil {
			continue
		}
		out = append(out, a)
	}
	return out
}
