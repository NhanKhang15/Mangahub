package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mangahub-backend/internal/core/external"
	"mangahub-backend/internal/core/response"
	"mangahub-backend/internal/core/ws"
	mangaGrpc "mangahub-backend/internal/modules/manga/grpcserver"
	mangaModel "mangahub-backend/internal/modules/manga/model"

	"mangahub-backend/internal/modules/crud/dto"
	catalogpb "mangahub-backend/proto/catalogpb"
)

type AdminHandler struct {
	agg    *external.Aggregator
	client catalogpb.MangaCatalogClient
	hub    *ws.Hub
}

func NewAdminHandler(agg *external.Aggregator, client catalogpb.MangaCatalogClient, hub *ws.Hub) *AdminHandler {
	return &AdminHandler{agg: agg, client: client, hub: hub}
}

func (h *AdminHandler) Import(c *gin.Context) {
	var q dto.ImportQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	ctx := c.Request.Context()

	items, err := h.fetch(ctx, q)
	if err != nil {
		response.RespondError(c, http.StatusBadGateway, "UPSTREAM_ERROR", err.Error(), nil)
		return
	}

	res := dto.ImportResult{Source: q.Source, Query: q.Q, Fetched: len(items)}
	for _, m := range items {
		resp, err := h.client.UpsertMangaByExternalIDs(ctx, &catalogpb.UpsertMangaByExternalIDsRequest{
			Manga: mangaGrpc.MangaToProto(m),
		})
		if err != nil {
			res.Skipped++
			continue
		}
		switch resp.GetAction() {
		case "insert":
			res.Inserted++
		case "update":
			res.Updated++
		}
	}
	c.JSON(http.StatusOK, res)
}

func (h *AdminHandler) Notify(c *gin.Context) {
	var req struct {
		UserID  string `json:"user_id" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}

	h.hub.SendDirect(&ws.DirectMessage{
		UserID:  req.UserID,
		Type:    "notification",
		Content: req.Content,
	})

	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}

func (h *AdminHandler) fetch(ctx context.Context, q dto.ImportQuery) ([]*mangaModel.Manga, error) {
	switch q.Source {
	case "all":
		return h.agg.Search(ctx, q.Q, q.Page)
	case "mangadex":
		if h.agg.MangaDex == nil {
			return nil, errSourceDisabled("mangadex")
		}
		ents, err := h.agg.MangaDex.SearchManga(ctx, q.Q, q.Page)
		if err != nil {
			return nil, err
		}
		return external.MergeEntities(ents, nil, nil), nil
	case "myanimelist":
		if h.agg.MAL == nil {
			return nil, errSourceDisabled("myanimelist")
		}
		ents, err := h.agg.MAL.SearchManga(ctx, q.Q, q.Page)
		if err != nil {
			return nil, err
		}
		return external.MergeEntities(nil, ents, nil), nil
	case "anilist":
		if h.agg.AniList == nil {
			return nil, errSourceDisabled("anilist")
		}
		ents, err := h.agg.AniList.SearchManga(ctx, q.Q, q.Page)
		if err != nil {
			return nil, err
		}
		return external.MergeEntities(nil, nil, ents), nil
	}
	return nil, nil
}

func errSourceDisabled(name string) error {
	return errors.New("source disabled (missing credentials): " + name)
}
