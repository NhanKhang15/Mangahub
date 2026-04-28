package gateway

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"mangahub-backend/internal/artist"
	"mangahub-backend/internal/catalog"
	"mangahub-backend/internal/gateway/handler"
	"mangahub-backend/internal/gateway/middleware"
	"mangahub-backend/internal/progress"
)

type Deps struct {
	MongoClient *mongo.Client
	MangaSvc    *catalog.Service
	ArtistSvc   *artist.Service
	ProgressSvc *progress.Service
}

func NewRouter(env string, d Deps) *gin.Engine {
	if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), middleware.CORS())

	healthH := handler.NewHealthHandler(d.MongoClient)
	mangaH := handler.NewMangaHandler(d.MangaSvc)
	artistH := handler.NewArtistHandler(d.ArtistSvc, d.MangaSvc)
	statsH := handler.NewStatsHandler(d.MangaSvc)
	progressH := handler.NewProgressHandler(d.ProgressSvc)

	r.GET("/healthz", healthH.Healthz)

	manga := r.Group("/manga")
	{
		manga.GET("", mangaH.List)
		manga.GET("/:id", mangaH.Get)
		manga.POST("", mangaH.Create)
		manga.PUT("/:id", mangaH.Update)
		manga.DELETE("/:id", mangaH.Delete)
	}

	artists := r.Group("/artists")
	{
		artists.GET("", artistH.List)
		artists.GET("/:id", artistH.Get)
		artists.POST("", artistH.Create)
		artists.PUT("/:id", artistH.Update)
		artists.DELETE("/:id", artistH.Delete)
		artists.GET("/:id/manga", artistH.ListMangaByArtist)
	}

	stats := r.Group("/stats")
	{
		stats.GET("/popular", statsH.Popular)
		stats.GET("/trending", statsH.Trending)
	}

	me := r.Group("/me", middleware.RequireUser())
	{
		me.GET("/reading", progressH.List)
		me.PUT("/reading/:mangaId", progressH.Upsert)
		me.DELETE("/reading/:mangaId", progressH.Delete)
		me.GET("/stats", progressH.Stats)
	}

	return r
}
