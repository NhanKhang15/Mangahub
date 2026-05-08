package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mangahub-backend/internal/core/middleware"
	"mangahub-backend/internal/core/response"
	"mangahub-backend/internal/gateway/grpcclient"
	"mangahub-backend/internal/platform/grpcinterceptor"

	prefspb "mangahub-backend/proto/prefspb"
)

type PrefsHandler struct {
	client prefspb.UserPreferencesClient
}

func NewPrefsHandler(c prefspb.UserPreferencesClient) *PrefsHandler {
	return &PrefsHandler{client: c}
}

type subscribeBody struct {
	Room string `json:"room" binding:"required"`
}

type updatePrefsBody struct {
	FavoriteGenres *[]string `json:"favorite_genres"`
	Language       *string   `json:"language"`
	NSFW           *bool     `json:"nsfw"`
}

func (h *PrefsHandler) ListSubscriptions(c *gin.Context) {
	uid := middleware.UserID(c).Hex()
	ctx := grpcinterceptor.AppendUserID(c.Request.Context(), uid)
	resp, err := h.client.ListSubscriptions(ctx, &prefspb.ListSubscriptionsRequest{UserId: uid})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	items := make([]gin.H, 0, len(resp.GetItems()))
	for _, s := range resp.GetItems() {
		row := gin.H{
			"id":      s.GetId(),
			"user_id": s.GetUserId(),
			"room":    s.GetRoom(),
		}
		if ts := s.GetCreatedAt(); ts != nil {
			row["created_at"] = ts.AsTime()
		}
		items = append(items, row)
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *PrefsHandler) Subscribe(c *gin.Context) {
	uid := middleware.UserID(c).Hex()
	var in subscribeBody
	if err := c.ShouldBindJSON(&in); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}
	ctx := grpcinterceptor.AppendUserID(c.Request.Context(), uid)
	resp, err := h.client.Subscribe(ctx, &prefspb.SubscribeRequest{UserId: uid, Room: in.Room})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	row := gin.H{
		"id":      resp.GetId(),
		"user_id": resp.GetUserId(),
		"room":    resp.GetRoom(),
	}
	if ts := resp.GetCreatedAt(); ts != nil {
		row["created_at"] = ts.AsTime()
	}
	c.JSON(http.StatusOK, row)
}

func (h *PrefsHandler) Unsubscribe(c *gin.Context) {
	uid := middleware.UserID(c).Hex()
	room := c.Param("room")
	if room == "" {
		response.RespondError(c, http.StatusBadRequest, "INVALID_ROOM", "missing room", nil)
		return
	}
	ctx := grpcinterceptor.AppendUserID(c.Request.Context(), uid)
	_, err := h.client.Unsubscribe(ctx, &prefspb.UnsubscribeRequest{UserId: uid, Room: room})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *PrefsHandler) GetPreferences(c *gin.Context) {
	uid := middleware.UserID(c).Hex()
	ctx := grpcinterceptor.AppendUserID(c.Request.Context(), uid)
	resp, err := h.client.GetPreferences(ctx, &prefspb.GetPreferencesRequest{UserId: uid})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	c.JSON(http.StatusOK, prefsResponse(resp))
}

func (h *PrefsHandler) UpdatePreferences(c *gin.Context) {
	uid := middleware.UserID(c).Hex()
	var in updatePrefsBody
	if err := c.ShouldBindJSON(&in); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}
	patch := &prefspb.PreferencesPatch{}
	hasField := false
	if in.FavoriteGenres != nil {
		patch.FavoriteGenres, patch.FavoriteGenresSet = *in.FavoriteGenres, true
		hasField = true
	}
	if in.Language != nil {
		patch.Language, patch.LanguageSet = *in.Language, true
		hasField = true
	}
	if in.NSFW != nil {
		patch.Nsfw, patch.NsfwSet = *in.NSFW, true
		hasField = true
	}
	if !hasField {
		response.RespondError(c, http.StatusBadRequest, "NO_FIELDS", "request body has no updatable fields", nil)
		return
	}
	ctx := grpcinterceptor.AppendUserID(c.Request.Context(), uid)
	resp, err := h.client.UpdatePreferences(ctx, &prefspb.UpdatePreferencesRequest{UserId: uid, Patch: patch})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	c.JSON(http.StatusOK, prefsResponse(resp))
}

func prefsResponse(p *prefspb.Preferences) gin.H {
	row := gin.H{
		"user_id":         p.GetUserId(),
		"favorite_genres": p.GetFavoriteGenres(),
		"language":        p.GetLanguage(),
		"nsfw":            p.GetNsfw(),
	}
	if ts := p.GetUpdatedAt(); ts != nil {
		row["updated_at"] = ts.AsTime()
	}
	return row
}
