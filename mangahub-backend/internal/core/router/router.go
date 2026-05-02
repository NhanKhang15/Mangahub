package router

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"mangahub-backend/internal/core/external"
	"mangahub-backend/internal/core/middleware"

	artistController "mangahub-backend/internal/modules/artist/controller"
	artistService "mangahub-backend/internal/modules/artist/service"

	mangaController "mangahub-backend/internal/modules/manga/controller"
	mangaService "mangahub-backend/internal/modules/manga/service"

	progressController "mangahub-backend/internal/modules/progress/controller"
	progressService "mangahub-backend/internal/modules/progress/service"

	crudController "mangahub-backend/internal/modules/crud/controller"
)

type Deps struct {
	MongoClient *mongo.Client
	MangaSvc    *mangaService.Service
	ArtistSvc   *artistService.Service
	ProgressSvc *progressService.Service
	Aggregator  *external.Aggregator
	AdminToken  string
}

func NewRouter(env string, d Deps) *gin.Engine {
	if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), middleware.CORS())

	healthH := NewHealthHandler(d.MongoClient)
	mangaH := mangaController.NewMangaHandler(d.MangaSvc)
	artistH := artistController.NewArtistHandler(d.ArtistSvc, d.MangaSvc)
	statsH := crudController.NewStatsHandler(d.MangaSvc)
	progressH := progressController.NewProgressHandler(d.ProgressSvc)

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

	if d.Aggregator != nil {
		adminH := crudController.NewAdminHandler(d.Aggregator, d.MangaSvc)
		admin := r.Group("/admin", middleware.RequireAdmin(d.AdminToken))
		{
			admin.POST("/import", adminH.Import)
		}
	}

	return r
}
