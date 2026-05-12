package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"mangahub-backend/internal/core/middleware"
	"mangahub-backend/internal/core/response"
	"mangahub-backend/internal/gateway/grpcclient"
	"mangahub-backend/internal/gateway/notifier"
	progressGrpc "mangahub-backend/internal/modules/progress/grpcserver"
	progressModel "mangahub-backend/internal/modules/progress/model"
	"mangahub-backend/internal/platform/grpcinterceptor"

	"mangahub-backend/internal/modules/progress/dto"
	progresspb "mangahub-backend/proto/progresspb"
)

type ProgressHandler struct {
	client   progresspb.ReadingProgressClient
	notifier *notifier.Client
}

func NewProgressHandler(c progresspb.ReadingProgressClient, n *notifier.Client) *ProgressHandler {
	return &ProgressHandler{client: c, notifier: n}
}

func (h *ProgressHandler) List(c *gin.Context) {
	uid := middleware.UserID(c)
	var q dto.ListProgressQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_QUERY", err.Error(), nil)
		return
	}
	ctx := grpcinterceptor.AppendUserID(c.Request.Context(), uid.Hex())
	resp, err := h.client.ListUserProgress(ctx, &progresspb.ListUserProgressRequest{
		UserId: uid.Hex(),
		Status: q.Status,
		Page:   int32(q.Page),
		Limit:  int32(q.Limit),
	})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  protoProgressesToModels(resp.GetItems()),
		"page":  q.Page,
		"limit": q.Limit,
		"total": resp.GetTotal(),
	})
}

func (h *ProgressHandler) Upsert(c *gin.Context) {
	uid := middleware.UserID(c)
	mangaIDRaw := c.Param("mangaId")
	if mangaIDRaw == "" {
		response.RespondError(c, http.StatusBadRequest, "INVALID_ID", "missing mangaId", nil)
		return
	}
	var in dto.UpsertProgressInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.RespondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error(), nil)
		return
	}
	ctx := grpcinterceptor.AppendUserID(c.Request.Context(), uid.Hex())
	resp, err := h.client.UpsertProgress(ctx, &progresspb.UpsertProgressRequest{
		UserId:         uid.Hex(),
		MangaId:        mangaIDRaw,
		Status:         in.Status,
		CurrentChapter: int32(in.CurrentChapter),
		Rating:         in.Rating,
	})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	model, convErr := protoProgressToModel(resp)
	if convErr != nil {
		response.RespondError(c, http.StatusInternalServerError, "INTERNAL", convErr.Error(), nil)
		return
	}

	// Fan out to tcp-svc so other devices subscribed to this user_id see the
	// new progress in real time. Fire-and-forget — never block the HTTP reply.
	if h.notifier != nil {
		go func(ev notifier.ProgressEvent) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			h.notifier.PublishProgress(ctx, ev)
		}(notifier.ProgressEvent{
			UserID:  uid.Hex(),
			MangaID: mangaIDRaw,
			Chapter: in.CurrentChapter,
			Status:  in.Status,
		})
	}

	c.JSON(http.StatusOK, model)
}

func (h *ProgressHandler) Delete(c *gin.Context) {
	uid := middleware.UserID(c)
	mangaIDRaw := c.Param("mangaId")
	if mangaIDRaw == "" {
		response.RespondError(c, http.StatusBadRequest, "INVALID_ID", "missing mangaId", nil)
		return
	}
	ctx := grpcinterceptor.AppendUserID(c.Request.Context(), uid.Hex())
	_, err := h.client.DeleteProgress(ctx, &progresspb.DeleteProgressRequest{
		UserId:  uid.Hex(),
		MangaId: mangaIDRaw,
	})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProgressHandler) Stats(c *gin.Context) {
	uid := middleware.UserID(c)
	ctx := grpcinterceptor.AppendUserID(c.Request.Context(), uid.Hex())
	resp, err := h.client.Stats(ctx, &progresspb.StatsRequest{UserId: uid.Hex()})
	if grpcclient.RespondGRPCError(c, err) {
		return
	}
	stats := make(map[string]int, len(resp.GetBuckets()))
	for k, v := range resp.GetBuckets() {
		stats[k] = int(v)
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func protoProgressesToModels(items []*progresspb.Progress) []*progressModel.ReadingProgress {
	out := make([]*progressModel.ReadingProgress, 0, len(items))
	for _, p := range items {
		m, err := protoProgressToModel(p)
		if err != nil {
			continue
		}
		out = append(out, m)
	}
	return out
}

// protoProgressToModel mirrors progressGrpc.ProgressToProto in reverse so the
// gateway can reuse the existing model type for JSON encoding (preserving the
// REST contract from before the gRPC split).
func protoProgressToModel(p *progresspb.Progress) (*progressModel.ReadingProgress, error) {
	if p == nil {
		return &progressModel.ReadingProgress{}, nil
	}
	out := &progressModel.ReadingProgress{
		Status:         p.GetStatus(),
		CurrentChapter: int(p.GetCurrentChapter()),
		Rating:         p.GetRating(),
	}
	if id := p.GetId(); id != "" {
		oid, err := progressGrpc.ValidateHex(id)
		if err != nil {
			return nil, err
		}
		out.ID = oid
	}
	if uid := p.GetUserId(); uid != "" {
		oid, err := progressGrpc.ValidateHex(uid)
		if err != nil {
			return nil, err
		}
		out.UserID = oid
	}
	if mid := p.GetMangaId(); mid != "" {
		oid, err := progressGrpc.ValidateHex(mid)
		if err != nil {
			return nil, err
		}
		out.MangaID = oid
	}
	if ts := p.GetLastReadAt(); ts != nil {
		out.LastReadAt = ts.AsTime()
	}
	return out, nil
}
