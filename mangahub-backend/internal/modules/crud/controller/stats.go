package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mangahub-backend/internal/core/response"
	"mangahub-backend/internal/gateway/grpcclient"
	mangaGrpc "mangahub-backend/internal/modules/manga/grpcserver"
	mangaModel "mangahub-backend/internal/modules/manga/model"

	"mangahub-backend/internal/modules/crud/dto"
	catalogpb "mangahub-backend/proto/catalogpb"
)

type StatsHandler struct {
	client catalogpb.MangaCatalogClient
}

func NewStatsHandler(c catalogpb.MangaCatalogClient) *StatsHandler {
	return &StatsHandler{client: c}
}

func (h *StatsHandler) Popular(c *gin.Context) {
	var q dto.LimitQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	resp, err := h.client.GetPopularManga(c.Request.Context(), &catalogpb.GetPopularMangaRequest{Limit: int32(q.Limit)})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": protoMangasToModels(resp.GetItems())})
}

func (h *StatsHandler) Trending(c *gin.Context) {
	var q dto.LimitQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	resp, err := h.client.GetTrendingManga(c.Request.Context(), &catalogpb.GetTrendingMangaRequest{Limit: int32(q.Limit)})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": protoMangasToModels(resp.GetItems())})
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
